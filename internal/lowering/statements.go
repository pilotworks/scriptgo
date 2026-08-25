package lowering

import (
	"fmt"
	"maps"
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerFunction(path string, statement typescriptgo.SyntaxStatement, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (ir.Function, error) {
	retType := statement.Type
	if retType == "" && statement.InferredType != "" {
		retType = statement.InferredType
	}
	function := ir.Function{Name: statement.Name, Span: toIRSpan(path, statement.Span), ReturnType: toIRType(retType)}
	if function.ReturnType == "" {
		function.ReturnType = ir.TypeVoid
	}
	env := map[string]ir.Type{}
	for _, parameter := range statement.Parameters {
		pType := parameter.Type
		if pType == "" && parameter.InferredType != "" {
			pType = parameter.InferredType
		}
		typ := toIRType(pType)
		if parameter.Rest {
			if typ == "" || typ == ir.TypeUnknown {
				if pType == "number[]" {
					typ = ir.TypeNumberArray
				} else {
					typ = ir.TypeStringArray
				}
			}
		}
		if typ == "" {
			return ir.Function{}, fmt.Errorf("parameter %q has unsupported type %q", parameter.Name, parameter.Type)
		}
		function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: typ})
		if typ == ir.TypeObject && pType != "" && pType != "object" {
			env[parameter.Name] = ir.Type("object:" + pType)
		} else {
			env[parameter.Name] = typ
		}
		fnSig := parameter.Type
		if fnSig == "" || fnSig == "closure" {
			fnSig = parameter.InferredType
		}
		if strings.Contains(fnSig, "=>") {
			parts := strings.Split(fnSig, "=>")
			retStr := strings.TrimSpace(parts[len(parts)-1])
			env[parameter.Name+".retType"] = toIRType(retStr)
		}
	}
	counter := 0
	returned := false
	for _, bodyStatement := range statement.Body {
		if err := lowerStatement(path, bodyStatement, &function, env, &counter, shapes, signatures); err != nil {
			return ir.Function{}, sourceError(path, bodyStatement.Span, err)
		}
		if statementAlwaysReturns(bodyStatement) {
			returned = true
		}
	}
	if !returned {
		if function.ReturnType == ir.TypeVoid {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: function.Span})
		} else if strings.HasPrefix(string(function.ReturnType), "object:Promise") {
			prom := nextTemp(&counter)
			zeroVal := nextTemp(&counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeNumber,
				Result: zeroVal,
				Value:  "0",
				Span:   function.Span,
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.Type("object:Promise"),
				Result: prom,
				Callee: "__async.promise_resolve",
				Args:   []string{zeroVal},
				Span:   function.Span,
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:   ir.OpReturn,
				Type: function.ReturnType,
				Args: []string{prom},
				Span: function.Span,
			})
		} else {
			defVal := ""
			if function.ReturnType == ir.TypeNumber {
				defVal = "0"
			} else if function.ReturnType == ir.TypeBool {
				defVal = "false"
			} else if strings.HasPrefix(string(function.ReturnType), "object:") || function.ReturnType == "ptr" {
				defVal = "0"
			}
			defTemp := nextTemp(&counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   function.ReturnType,
				Result: defTemp,
				Value:  defVal,
				Span:   function.Span,
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:   ir.OpReturn,
				Type: function.ReturnType,
				Args: []string{defTemp},
				Span: function.Span,
			})
		}
	}
	return function, nil
}

func lowerStatement(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	switch statement.Kind {
	case "variable", "using", "await_using":
		if statement.Expression == nil {
			return fmt.Errorf("variable %q has no initializer", statement.Name)
		}
		inProgressVars[statement.Name] = true
		defer delete(inProgressVars, statement.Name)
		declaredType := toIRType(statement.Type)
		if statement.Type == "" && statement.InferredType != "" {
			declaredType = toIRType(statement.InferredType)
		}
		if statement.Expression.Kind == "identifier" || statement.Expression.Kind == "this" {
			identText := statement.Expression.Text
			if statement.Expression.Kind == "this" {
				identText = "this"
			}
			srcType, ok := env[identText]
			if ok && (declaredType == "" || declaredType == srcType || declaredType == ir.TypeUnknown) {
				if declaredType == ir.TypeUnknown {
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpBoxUnknown,
						Type:   ir.TypeUnknown,
						Result: statement.Name,
						Args:   []string{identText},
						Span:   toIRSpan(path, statement.Span),
					})
					env[statement.Name] = ir.TypeUnknown
					return nil
				}
				env[statement.Name] = srcType
				switch srcType {
				case ir.TypeNumber:
					zeroConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, statement.Span)})
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: statement.Name, Operator: "+", Args: []string{identText, zeroConst}, Span: toIRSpan(path, statement.Span)})
				case ir.TypeString:
					emptyStr := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: emptyStr, Value: "", Span: toIRSpan(path, statement.Span)})
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: statement.Name, Operator: "+", Args: []string{identText, emptyStr}, Span: toIRSpan(path, statement.Span)})
				case ir.TypeBool:
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: statement.Name, Operator: "||", Args: []string{identText, identText}, Span: toIRSpan(path, statement.Span)})
				default:
					trueConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: trueConst, Value: "true", Span: toIRSpan(path, statement.Span)})
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: srcType, Result: statement.Name, Args: []string{trueConst, identText, identText}, Span: toIRSpan(path, statement.Span)})
				}
				return nil
			} else if !ok {
				global, isGlobal := builtinGlobal(statement.Expression.Text)
				if isGlobal {
					env[statement.Name] = global.Type
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: global.Type, Result: statement.Name, Value: global.Value, Span: toIRSpan(path, statement.Span)})
					return nil
				}
				return fmt.Errorf("unknown identifier %q", statement.Expression.Text)
			}
		}
		if declaredType == ir.TypeUnknown {
			value, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
			if err != nil {
				return err
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: statement.Name,
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
			env[statement.Name] = ir.TypeUnknown
			return nil
		}
		if (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && declaredType != "" && declaredType != ir.TypeVoid && declaredType != ir.TypeUnknown {
			defaultVal := "0"
			if declaredType == ir.TypeNumber {
				defaultVal = "NaN"
			} else if declaredType == ir.TypeBool {
				defaultVal = "false"
			} else if declaredType == ir.TypeString {
				if statement.Expression.Kind == "undefined" {
					defaultVal = "undefined"
				} else {
					defaultVal = "null"
				}
			} else if strings.HasPrefix(string(declaredType), "object:") || declaredType == ir.TypePointer {
				defaultVal = "null"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   declaredType,
				Result: statement.Name,
				Value:  defaultVal,
				Span:   toIRSpan(path, statement.Span),
			})
			env[statement.Name] = declaredType
			return nil
		}
		if statement.Expression != nil && statement.Expression.Kind == "array" && statement.Type != "" && (statement.Expression.InferredType == "" || statement.Expression.InferredType == "never[]" || statement.Expression.InferredType == "any[]" || statement.Expression.InferredType == "unknown[]") {
			statement.Expression.InferredType = statement.Type
		}
		if statement.Expression != nil && statement.Expression.Kind == "object_literal" && statement.Type != "" && !strings.Contains(statement.Type, "|") {
			if aliased, isUnion := typeAliasesIndex[statement.Type]; !isUnion || !strings.Contains(aliased, "|") {
				statement.Expression.InferredType = statement.Type
			}
		}
		if statement.Expression != nil && (statement.Expression.Kind == "function" || statement.Expression.Function != nil || strings.Contains(statement.Type, "=>") || strings.Contains(statement.InferredType, "=>")) {
			env[statement.Name] = ir.TypeClosure
			fnSig := statement.Type
			if fnSig == "" || fnSig == "closure" {
				fnSig = statement.InferredType
			}
			if strings.Contains(fnSig, "=>") {
				parts := strings.Split(fnSig, "=>")
				retStr := strings.TrimSpace(parts[len(parts)-1])
				env[statement.Name+".retType"] = toIRType(retStr)
			}
		}
		value, typ, err := lowerExpression(path, statement.Expression, statement.Name, function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if value != statement.Name {
			return fmt.Errorf("variable %q produced unnamed value %q", statement.Name, value)
		}
		if statement.Type != "" && !strings.HasPrefix(string(typ), "object:Generator_") {
			if !strings.Contains(statement.Type, "|") {
				if declared := toIRType(statement.Type); declared != "" {
					typ = declared
				}
			}
		} else if statement.InferredType != "" && !strings.HasPrefix(string(typ), "object:Generator_") {
			if inferred := toIRType(statement.InferredType); inferred != "" {
				typ = inferred
			}
		}
		env[statement.Name] = typ
		if typ == ir.TypeClosure {
			fnSig := statement.Type
			if fnSig == "" || fnSig == "closure" {
				fnSig = statement.InferredType
			}
			if strings.Contains(fnSig, "=>") {
				parts := strings.Split(fnSig, "=>")
				retStr := strings.TrimSpace(parts[len(parts)-1])
				env[statement.Name+".retType"] = toIRType(retStr)
			}
		}
		return nil
	case "expression":
		if statement.Expression == nil {
			return fmt.Errorf("empty expression")
		}
		_, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		return err
	case "return":
		if statement.Expression == nil {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: toIRSpan(path, statement.Span)})
			return nil
		}
		if (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && function.ReturnType != "" && function.ReturnType != ir.TypeVoid && function.ReturnType != ir.TypeUnknown {
			res := nextTemp(counter)
			if strings.HasPrefix(string(function.ReturnType), "object:Promise") {
				prom := nextTemp(counter)
				zeroVal := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeNumber,
					Result: zeroVal,
					Value:  "0",
					Span:   toIRSpan(path, statement.Span),
				})
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.Type("object:Promise"),
					Result: prom,
					Callee: "__async.promise_resolve",
					Args:   []string{zeroVal},
					Span:   toIRSpan(path, statement.Span),
				})
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: function.ReturnType, Args: []string{prom}, Span: toIRSpan(path, statement.Span)})
				return nil
			}
			defaultVal := "0"
			if function.ReturnType == ir.TypeBool {
				defaultVal = "false"
			} else if function.ReturnType == ir.TypeNumber {
				defaultVal = "0"
			} else if statement.Expression.Kind == "undefined" {
				defaultVal = "undefined"
			} else if statement.Expression.Kind == "null" || strings.HasPrefix(string(function.ReturnType), "object:") || isPointerLikeType(function.ReturnType) || function.ReturnType == ir.TypeString {
				defaultVal = "null"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   function.ReturnType,
				Result: res,
				Value:  defaultVal,
				Span:   toIRSpan(path, statement.Span),
			})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: function.ReturnType, Args: []string{res}, Span: toIRSpan(path, statement.Span)})
			return nil
		}
		if statement.Expression != nil && statement.Expression.Kind == "object_literal" && function.ReturnType != "" && (statement.Expression.InferredType == "" || strings.HasPrefix(string(function.ReturnType), "object:")) && !strings.Contains(string(function.ReturnType), "|") {
			cleanRet := strings.TrimPrefix(string(function.ReturnType), "object:")
			if aliased, ok := typeAliasesIndex[cleanRet]; !ok || !strings.Contains(aliased, "|") {
				statement.Expression.InferredType = string(function.ReturnType)
			}
		}
		value, typ, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if strings.HasPrefix(string(function.ReturnType), "object:Promise") && !strings.HasPrefix(string(typ), "object:Promise") {
			prom := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.Type("object:Promise"),
				Result: prom,
				Callee: "__async.promise_resolve",
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
			value = prom
			typ = function.ReturnType
		}
		if function.ReturnType == ir.TypeUnknown && typ != ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: boxed,
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
			value = boxed
			typ = ir.TypeUnknown
		}
		if strings.HasPrefix(string(function.ReturnType), "object:Promise") && !strings.HasPrefix(string(typ), "object:Promise") {
			prom := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   function.ReturnType,
				Result: prom,
				Callee: "__async.promise_resolve",
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
			value = prom
			typ = function.ReturnType
		} else if strings.HasPrefix(string(function.ReturnType), "object:") && !strings.Contains(string(function.ReturnType), "{") && strings.HasPrefix(string(typ), "object:") {
			typ = function.ReturnType
		} else if function.ReturnType == "ptr" && (typ == ir.TypeString || isPointerLikeType(typ)) {
			typ = "ptr"
		} else if function.ReturnType == ir.TypeString && typ == "ptr" {
			typ = ir.TypeString
		} else if function.ReturnType != "" && function.ReturnType != ir.TypeVoid && isPointerLikeType(function.ReturnType) && (typ == "ptr" || isPointerLikeType(typ)) {
			typ = function.ReturnType
		} else if function.ReturnType == "" || function.ReturnType == ir.TypeVoid {
			function.ReturnType = typ
			if signatures != nil {
				if sig, ok := signatures[function.Name]; ok {
					sig.ReturnType = typ
					signatures[function.Name] = sig
				}
			}
		}
		if err := lowerActiveReturnFinally(path, function, env, counter, shapes, signatures); err != nil {
			return err
		}
		if typ == ir.TypeVoid || value == "" {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: toIRSpan(path, statement.Span)})
		} else {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: typ, Args: []string{value}, Span: toIRSpan(path, statement.Span)})
		}
	case "block":
		for _, s := range statement.Body {
			if err := lowerStatement(path, s, function, env, counter, shapes, signatures); err != nil {
				return err
			}
		}
	case "namespace":
		for _, s := range statement.Body {
			if s.Kind == "variable" || s.Kind == "using" || s.Kind == "await_using" {
				varCopy := s
				varCopy.Name = statement.Name + "." + s.Name
				if err := lowerStatement(path, varCopy, function, env, counter, shapes, signatures); err != nil {
					return err
				}
			} else {
				if err := lowerStatement(path, s, function, env, counter, shapes, signatures); err != nil {
					return err
				}
			}
		}
	case "assign":
		varType, ok := env[statement.Name]
		if !ok {
			if topVar, isTop := topLevelVars[statement.Name]; isTop {
				vType := topVar.Type
				if vType == "" && topVar.InferredType != "" {
					vType = topVar.InferredType
				}
				varType = toIRType(vType)
				if varType == "" {
					varType = ir.TypeNumber
				}
				env[statement.Name] = varType
			} else {
				return fmt.Errorf("assignment to unknown variable %q", statement.Name)
			}
		}
		if (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && varType != ir.TypeString && varType != ir.TypeUnknown {
			defaultVal := "0"
			if varType == ir.TypeNumber {
				defaultVal = "NaN"
			} else if varType == ir.TypeBool {
				defaultVal = "false"
			} else if strings.HasPrefix(string(varType), "object:") {
				defaultVal = "null"
			}
			tmp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   varType,
				Result: tmp,
				Value:  defaultVal,
				Span:   toIRSpan(path, statement.Span),
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   varType,
				Result: statement.Name,
				Args:   []string{tmp},
				Span:   toIRSpan(path, statement.Span),
			})
			return nil
		}
		if statement.Expression != nil && statement.Expression.Kind == "array" && (statement.Expression.InferredType == "" || statement.Expression.InferredType == "never[]" || statement.Expression.InferredType == "any[]" || statement.Expression.InferredType == "unknown[]") {
			statement.Expression.InferredType = string(varType)
		}
		value, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if varType == ir.TypeUnknown && valType != ir.TypeUnknown {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: statement.Name,
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
			return nil
		}
		if valType != varType {
			if strings.HasPrefix(string(valType), "object:") && strings.HasPrefix(string(varType), "object:") {
				// Polymorphic object assignment
			} else {
				return fmt.Errorf("assignment type mismatch for %q: %s := %s", statement.Name, varType, valType)
			}
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpAssign, Type: varType, Result: statement.Name, Args: []string{value}, Span: toIRSpan(path, statement.Span)})
	case "while":
		return lowerWhile(path, statement, function, env, counter, shapes, signatures)
	case "if":
		return lowerIf(path, statement, function, env, counter, shapes, signatures)
	case "break":
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBreak, Type: ir.TypeVoid, Value: statement.Name, Span: toIRSpan(path, statement.Span)})
	case "continue":
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpContinue, Type: ir.TypeVoid, Value: statement.Name, Span: toIRSpan(path, statement.Span)})
	case "dowhile":
		return lowerDoWhile(path, statement, function, env, counter, shapes, signatures)
	case "forof", "forawaitof":
		return lowerForOf(path, statement, function, env, counter, shapes, signatures)
	case "forin":
		return lowerForIn(path, statement, function, env, counter, shapes, signatures)
	case "label":
		return lowerLabel(path, statement, function, env, counter, shapes, signatures)
	case "switch":
		return lowerSwitch(path, statement, function, env, counter, shapes, signatures)
	case "function":
		_, typ, err := lowerClosureExpression(path, &statement, statement.Name, function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		env[statement.Name] = typ
		return nil
	case "index_set":
		arrVal, arrType, err := lowerExpression(path, statement.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if after, ok := strings.CutPrefix(string(arrType), "object:"); ok {
			shapeName := after
			if shape, ok := shapes[shapeName]; ok {
				if statement.Right != nil && statement.Right.Kind == "number" {
					fieldIdx, _ := strconv.Atoi(statement.Right.Text)
					if fieldIdx >= 0 && fieldIdx < len(shape.Fields) {
						field := shape.Fields[fieldIdx]
						val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
						if err != nil {
							return err
						}
						function.Body = append(function.Body, ir.Instruction{
							Op:         ir.OpFieldSet,
							Type:       ir.TypeVoid,
							Callee:     shapeName,
							Field:      field.Name,
							FieldIndex: fieldIdx,
							Args:       []string{arrVal, val},
							Span:       toIRSpan(path, statement.Span),
						})
						return nil
					}
				}
				if statement.Right != nil && statement.Right.Kind == "string" {
					propName := statement.Right.Text
					for idx, field := range shape.Fields {
						if field.Name == propName {
							val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
							if err != nil {
								return err
							}
							function.Body = append(function.Body, ir.Instruction{
								Op:         ir.OpFieldSet,
								Type:       ir.TypeVoid,
								Callee:     shapeName,
								Field:      field.Name,
								FieldIndex: idx,
								Args:       []string{arrVal, val},
								Span:       toIRSpan(path, statement.Span),
							})
							return nil
						}
					}
				}
			}
		}
		if arrType == ir.TypeString {
			return fmt.Errorf("cannot assign to read-only string index")
		}
		idxVal, idxType, err := lowerExpression(path, statement.Right, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if idxType != ir.TypeNumber {
			return fmt.Errorf("array index_set requires number index, got %s", idxType)
		}
		val, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		var expectedElemType ir.Type
		if arrType == ir.TypeBigInt64Array || arrType == ir.TypeBigUint64Array {
			expectedElemType = ir.TypeBigInt
		} else if isNumberTypedArray(arrType) || arrType == ir.TypeNumberArray {
			expectedElemType = ir.TypeNumber
		} else if arrType == ir.TypeStringArray {
			expectedElemType = ir.TypeString
		} else if arrType == ir.TypeBoolArray || arrType == "boolean[]" || arrType == "bool[]" {
			expectedElemType = ir.TypeBool
		} else if before, ok := strings.CutSuffix(string(arrType), "[]"); ok {
			elemName := before
			if elemName == "boolean" {
				expectedElemType = ir.TypeBool
			} else {
				expectedElemType = ir.Type(elemName)
			}
		} else {
			return fmt.Errorf("array index_set requires an array, got %s", arrType)
		}
		if valType == ir.TypeVoid || (statement.Expression != nil && (statement.Expression.Kind == "undefined" || statement.Expression.Kind == "null" || (statement.Expression.Kind == "identifier" && statement.Expression.Text == "undefined"))) {
			zeroVal := nextTemp(counter)
			switch expectedElemType {
			case ir.TypeNumber:
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroVal, Value: "0", Span: toIRSpan(path, statement.Span)})
			case ir.TypeBool:
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: zeroVal, Value: "false", Span: toIRSpan(path, statement.Span)})
			case ir.TypeString:
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: zeroVal, Value: "", Span: toIRSpan(path, statement.Span)})
			default:
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: expectedElemType, Result: zeroVal, Value: "null", Span: toIRSpan(path, statement.Span)})
			}
			val = zeroVal
			valType = expectedElemType
		}
		if expectedElemType != "" && valType != expectedElemType && valType != ir.TypeUnknown && expectedElemType != ir.TypeUnknown {
			return fmt.Errorf("array index_set type mismatch: %s cannot be assigned to %s", valType, arrType)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpIndexSet,
			Type: ir.TypeVoid,
			Args: []string{arrVal, idxVal, val},
			Span: toIRSpan(path, statement.Span),
		})
	case "field_set":
		if statement.Left != nil && statement.Left.Kind == "identifier" {
			className := statement.Left.Text
			if meta, isClass := classHierarchy[className]; isClass {
				// Check static setter
				if _, setterName, ok := findSetterInHierarchy(className, statement.Name, signatures, classHierarchy); ok {
					val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
					if err != nil {
						return err
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: setterName,
						Args:   []string{val},
						Span:   toIRSpan(path, statement.Span),
					})
					return nil
				}
				// Static field assignment
				if _, isStatic := meta.Statics[statement.Name]; isStatic {
					staticVar := className + "_" + statement.Name
					val, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
					if err != nil {
						return err
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpAssign,
						Type:   valType,
						Result: staticVar,
						Args:   []string{val},
						Span:   toIRSpan(path, statement.Span),
					})
					return nil
				}
			}
		}

		objVal, objType, err := lowerExpression(path, statement.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if (strings.HasSuffix(string(objType), "[]") || objType == ir.TypeNumberArray || objType == ir.TypeStringArray || objType == ir.TypeBoolArray || objType == ir.TypeBigIntArray) && statement.Name == "length" {
			val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
			if err != nil {
				return err
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeVoid,
				Callee: "__array.set_length",
				Args:   []string{objVal, val},
				Span:   toIRSpan(path, statement.Span),
			})
			return nil
		}
		className := strings.TrimPrefix(string(objType), string(ir.TypeObject)+":")
		if className == "this" || className == "" {
			if statement.Left != nil && statement.Left.Text != "" {
				if t, inEnv := env[statement.Left.Text]; inEnv && string(t) != "" && string(t) != "this" {
					className = strings.TrimPrefix(string(t), string(ir.TypeObject)+":")
				}
			}
			if className == "this" || className == "" {
				if t, inEnv := env["this"]; inEnv && string(t) != "this" && string(t) != "object:this" {
					className = strings.TrimPrefix(string(t), string(ir.TypeObject)+":")
				} else if function != nil && strings.Contains(function.Name, "_") && !strings.HasPrefix(function.Name, "__closure_") {
					className = strings.Split(function.Name, "_")[0]
				}
			}
			if className == "this" || className == "" {
				for sName, s := range shapes {
					if fieldIndex(s, statement.Name) >= 0 {
						className = sName
						break
					}
				}
			}
		}

		// Check instance setter in hierarchy
		if _, setterName, ok := findSetterInHierarchy(className, statement.Name, signatures, classHierarchy); ok {
			val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
			if err != nil {
				return err
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeVoid,
				Callee: setterName,
				Args:   []string{objVal, val},
				Span:   toIRSpan(path, statement.Span),
			})
			return nil
		}

		shape, ok := shapes[className]
		if !ok {
			return fmt.Errorf("field set on unknown object shape %q", className)
		}
		fIndex := fieldIndex(shape, statement.Name)
		if fIndex < 0 {
			return fmt.Errorf("unknown field %q on object shape %q", statement.Name, className)
		}
		if statement.Expression != nil && (statement.Expression.Kind == "array" || (statement.Expression.Kind == "new" && callName(statement.Expression.Left) == "Array")) && (statement.Expression.InferredType == "" || statement.Expression.InferredType == "never[]" || statement.Expression.InferredType == "any[]" || statement.Expression.InferredType == "unknown[]" || statement.Expression.InferredType == "void[]") {
			statement.Expression.InferredType = string(shape.Fields[fIndex].Type)
		}
		var val string
		var valType ir.Type
		if statement.Expression != nil && (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && (strings.HasPrefix(string(shape.Fields[fIndex].Type), "object:") || shape.Fields[fIndex].Type == ir.TypeClosure) {
			val = nextTemp(counter)
			valType = shape.Fields[fIndex].Type
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   valType,
				Result: val,
				Value:  "null",
				Span:   toIRSpan(path, statement.Span),
			})
		} else {
			var err error
			val, valType, err = lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
			if err != nil {
				return err
			}
		}
		if valType != shape.Fields[fIndex].Type {
			if strings.HasSuffix(string(shape.Fields[fIndex].Type), "[]") && (valType == "void[]" || valType == "never[]" || valType == "any[]" || valType == "unknown[]") {
				valType = shape.Fields[fIndex].Type
			} else {
				return fmt.Errorf("field set type mismatch for %q: %s := %s", statement.Name, shape.Fields[fIndex].Type, valType)
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:         ir.OpFieldSet,
			Type:       ir.TypeVoid,
			Field:      statement.Name,
			FieldIndex: fIndex,
			Args:       []string{objVal, val},
			Span:       toIRSpan(path, statement.Span),
		})
	case "class":
		return nil
	case "throw":
		val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if err := lowerActiveThrowFinally(path, function, env, counter, shapes, signatures); err != nil {
			return err
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpThrow,
			Type: ir.TypeVoid,
			Args: []string{val},
			Span: toIRSpan(path, statement.Span),
		})
	case "try":
		return lowerTry(path, statement, function, env, counter, shapes, signatures)
	case "debugger":
		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpDebugger,
			Type: ir.TypeVoid,
			Span: toIRSpan(path, statement.Span),
		})
	case "module", "enum":
		return nil
	default:
		return fmt.Errorf("unsupported statement %q", statement.Kind)
	}
	return nil
}

func lowerActiveReturnFinally(path string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	for i := len(activeReturnFinallyStack) - 1; i >= 0; i-- {
		for _, finStmt := range activeReturnFinallyStack[i] {
			if err := lowerStatement(path, finStmt, function, env, counter, shapes, signatures); err != nil {
				return err
			}
		}
	}
	return nil
}

func lowerActiveThrowFinally(path string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	for i := len(activeThrowFinallyStack) - 1; i >= 0; i-- {
		for _, finStmt := range activeThrowFinallyStack[i] {
			if err := lowerStatement(path, finStmt, function, env, counter, shapes, signatures); err != nil {
				return err
			}
		}
	}
	return nil
}

func lowerBranch(path string, statements []typescriptgo.SyntaxStatement, returnType ir.Type, parentEnv map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) ([]ir.Instruction, error) {
	branch := ir.Function{Name: "branch", ReturnType: returnType}
	env := make(map[string]ir.Type, len(parentEnv))
	maps.Copy(env, parentEnv)
	for _, statement := range statements {
		if err := lowerStatement(path, statement, &branch, env, counter, shapes, signatures); err != nil {
			return nil, err
		}
	}
	return branch.Body, nil
}
