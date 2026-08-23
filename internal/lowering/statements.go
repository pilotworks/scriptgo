package lowering

import (
	"fmt"
	"maps"
	"slices"
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
			if pType == "number[]" {
				typ = ir.TypeNumberArray
			} else {
				typ = ir.TypeStringArray
			}
		}
		if typ == "" {
			return ir.Function{}, fmt.Errorf("parameter %q has unsupported type %q", parameter.Name, parameter.Type)
		}
		function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: typ})
		env[parameter.Name] = typ
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
	case "variable":
		if statement.Expression == nil {
			return fmt.Errorf("variable %q has no initializer", statement.Name)
		}
		inProgressVars[statement.Name] = true
		defer delete(inProgressVars, statement.Name)
		declaredType := toIRType(statement.Type)
		if statement.Type == "" && statement.InferredType != "" {
			declaredType = toIRType(statement.InferredType)
		}
		if statement.Expression.Kind == "identifier" {
			srcType, ok := env[statement.Expression.Text]
			if ok && (declaredType == "" || declaredType == srcType || declaredType == ir.TypeUnknown) {
				if declaredType == ir.TypeUnknown {
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpBoxUnknown,
						Type:   ir.TypeUnknown,
						Result: statement.Name,
						Args:   []string{statement.Expression.Text},
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
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: statement.Name, Operator: "+", Args: []string{statement.Expression.Text, zeroConst}, Span: toIRSpan(path, statement.Span)})
				case ir.TypeString:
					emptyStr := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: emptyStr, Value: "", Span: toIRSpan(path, statement.Span)})
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: statement.Name, Operator: "+", Args: []string{statement.Expression.Text, emptyStr}, Span: toIRSpan(path, statement.Span)})
				case ir.TypeBool:
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: statement.Name, Operator: "||", Args: []string{statement.Expression.Text, statement.Expression.Text}, Span: toIRSpan(path, statement.Span)})
				default:
					trueConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: trueConst, Value: "true", Span: toIRSpan(path, statement.Span)})
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: srcType, Result: statement.Name, Args: []string{trueConst, statement.Expression.Text, statement.Expression.Text}, Span: toIRSpan(path, statement.Span)})
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
		if (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && declaredType != "" && declaredType != ir.TypeVoid && declaredType != ir.TypeString && declaredType != ir.TypeUnknown {
			defaultVal := "0"
			if declaredType == ir.TypeBool {
				defaultVal = "false"
			} else if strings.HasPrefix(string(declaredType), "object:") {
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
		value, typ, err := lowerExpression(path, statement.Expression, statement.Name, function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if value != statement.Name {
			return fmt.Errorf("variable %q produced unnamed value %q", statement.Name, value)
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
		if (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && function.ReturnType != "" && function.ReturnType != ir.TypeVoid && function.ReturnType != ir.TypeString && function.ReturnType != ir.TypeUnknown {
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
			} else if strings.HasPrefix(string(function.ReturnType), "object:") {
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
		if strings.HasPrefix(string(function.ReturnType), "object:") && strings.HasPrefix(string(typ), "object:") {
			typ = function.ReturnType
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: typ, Args: []string{value}, Span: toIRSpan(path, statement.Span)})
	case "block":
		for _, s := range statement.Body {
			if err := lowerStatement(path, s, function, env, counter, shapes, signatures); err != nil {
				return err
			}
		}
	case "assign":
		varType, ok := env[statement.Name]
		if !ok {
			return fmt.Errorf("assignment to unknown variable %q", statement.Name)
		}
		if (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && varType != ir.TypeString && varType != ir.TypeUnknown {
			defaultVal := "0"
			if varType == ir.TypeBool {
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
			return fmt.Errorf("assignment type mismatch for %q: %s := %s", statement.Name, varType, valType)
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
					if _, exists := env[staticVar]; !exists {
						_, valType, err := lowerExpression(path, statement.Expression, staticVar, function, env, counter, shapes, signatures)
						if err != nil {
							return err
						}
						env[staticVar] = valType
						return nil
					}
					val, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
					if err != nil {
						return err
					}
					env[staticVar] = valType
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
		className := strings.TrimPrefix(string(objType), string(ir.TypeObject)+":")

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
		if statement.Expression != nil && statement.Expression.Kind == "array" && (statement.Expression.InferredType == "" || statement.Expression.InferredType == "never[]" || statement.Expression.InferredType == "any[]" || statement.Expression.InferredType == "unknown[]") {
			statement.Expression.InferredType = string(shape.Fields[fIndex].Type)
		}
		var val string
		var valType ir.Type
		if statement.Expression != nil && (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && strings.HasPrefix(string(shape.Fields[fIndex].Type), "object:") {
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
			return fmt.Errorf("field set type mismatch for %q: %s := %s", statement.Name, shape.Fields[fIndex].Type, valType)
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
		val, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if strings.HasPrefix(string(valType), "object:Error") || valType == "object:TypeError" || valType == "object:RangeError" || valType == "object:SyntaxError" {
			msgVal := nextTemp(counter)
			className := strings.TrimPrefix(string(valType), "object:")
			function.Body = append(function.Body, ir.Instruction{
				Op:         ir.OpFieldGet,
				Type:       ir.TypeString,
				Result:     msgVal,
				Callee:     className,
				Field:      "message",
				FieldIndex: 0,
				Args:       []string{val},
				Span:       toIRSpan(path, statement.Span),
			})
			val = msgVal
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

var typeAliasesIndex = map[string]string{}

func mangleGenericTypeString(t string) string {
	t = strings.TrimSpace(t)
	if strings.HasSuffix(t, "[]") {
		elem := strings.TrimSuffix(t, "[]")
		return mangleGenericTypeString(elem) + "_arr"
	}
	if strings.Contains(t, "<") && strings.HasSuffix(t, ">") {
		idx := strings.Index(t, "<")
		base := t[:idx]
		inner := t[idx+1 : len(t)-1]
		typeArgs := splitTypeArguments(inner)
		return mangleGenericName(base, typeArgs)
	}
	return t
}

func toIRType(value string) ir.Type {
	value = strings.TrimSpace(value)
	if aliased, ok := typeAliasesIndex[value]; ok && aliased != value {
		return toIRType(aliased)
	}
	base := value
	if idx := strings.Index(base, "__"); idx != -1 {
		base = base[:idx]
	}
	if idx := strings.Index(base, "<"); idx != -1 {
		base = base[:idx]
	}
	if aliased, ok := typeAliasesIndex[base]; ok && aliased != base {
		if strings.Contains(aliased, "=>") {
			return ir.TypeClosure
		}
	}
	if strings.Contains(value, "=>") {
		return ir.TypeClosure
	}
	if strings.HasSuffix(value, "_arr") {
		elem := strings.TrimSuffix(value, "_arr")
		return toIRType(elem + "[]")
	}
	if strings.HasSuffix(value, "[]") {
		elem := strings.TrimSuffix(value, "[]")
		elemType := toIRType(elem)
		switch elemType {
		case ir.TypeNumber:
			return ir.TypeNumberArray
		case ir.TypeString:
			return ir.TypeStringArray
		case ir.TypeBool:
			return ir.TypeBoolArray
		case ir.TypeBigInt:
			return ir.TypeBigIntArray
		default:
			return ir.Type(string(elemType) + "[]")
		}
	}
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) || (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return ir.TypeString
	}
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return ir.TypeNumber
	}
	if value == "true" || value == "false" {
		return ir.TypeBool
	}
	if strings.Contains(value, "|") {
		parts := strings.Split(value, "|")
		var nonNullish []string
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "null" && trimmed != "undefined" && trimmed != "void" && trimmed != "" {
				nonNullish = append(nonNullish, trimmed)
			}
		}
		if len(nonNullish) == 0 {
			return ir.TypeVoid
		}
		return toIRType(nonNullish[0])
	}
	if strings.Contains(value, "<") && strings.HasSuffix(value, ">") {
		clean := strings.TrimPrefix(value, "object:")
		idx := strings.Index(clean, "<")
		base := clean[:idx]
		inner := clean[idx+1 : len(clean)-1]
		if base == "Promise" {
			return ir.Type("object:Promise_" + mangleGenericTypeString(inner))
		}
		if base == "Array" {
			return ir.Type(mangleGenericTypeString(inner) + "[]")
		}
		if base == "Map" {
			return ir.TypeMap
		}
		if base == "Set" {
			return ir.TypeSet
		}
		typeArgs := splitTypeArguments(inner)
		return ir.Type("object:" + mangleGenericName(base, typeArgs))
	}
	if strings.HasPrefix(value, "object:") {
		trimmed := strings.TrimPrefix(value, "object:")
		switch trimmed {
		case "Buffer":
			return ir.TypeBuffer
		case "Uint8Array":
			return ir.TypeUint8Array
		case "Int8Array":
			return ir.TypeInt8Array
		case "Uint8ClampedArray":
			return ir.TypeUint8ClampedArray
		case "Int16Array":
			return ir.TypeInt16Array
		case "Uint16Array":
			return ir.TypeUint16Array
		case "Int32Array":
			return ir.TypeInt32Array
		case "Uint32Array":
			return ir.TypeUint32Array
		case "Float32Array":
			return ir.TypeFloat32Array
		case "Float64Array":
			return ir.TypeFloat64Array
		case "BigInt64Array":
			return ir.TypeBigInt64Array
		case "BigUint64Array":
			return ir.TypeBigUint64Array
		case "DataView":
			return ir.TypeDataView
		case "ArrayBuffer":
			return ir.TypeArrayBuffer
		case "Map":
			return ir.TypeMap
		case "Set":
			return ir.TypeSet
		case "TextEncoder":
			return ir.TypeTextEncoder
		case "TextDecoder":
			return ir.TypeTextDecoder
		case "RegExp":
			return ir.Type("object:RegExp")
		}
		return ir.Type(value)
	}
	switch value {
	case "number":
		return ir.TypeNumber
	case "bigint":
		return ir.TypeBigInt
	case "bigint[]":
		return ir.TypeBigIntArray
	case "symbol":
		return ir.TypeSymbol
	case "symbol[]":
		return ir.TypeSymbolArray
	case "RegExp":
		return ir.Type("object:RegExp")
	case "string", "null", "undefined":
		return ir.TypeString
	case "bool", "boolean":
		return ir.TypeBool
	case "bool[]", "boolean[]":
		return ir.TypeBoolArray
	case "number[]":
		return ir.TypeNumberArray
	case "string[]", "TemplateStringsArray":
		return ir.TypeStringArray
	case "closure", "function":
		return ir.TypeClosure
	case "unknown":
		return ir.TypeUnknown
	case "Uint8Array":
		return ir.TypeUint8Array
	case "Int8Array":
		return ir.TypeInt8Array
	case "Uint8ClampedArray":
		return ir.TypeUint8ClampedArray
	case "Int16Array":
		return ir.TypeInt16Array
	case "Uint16Array":
		return ir.TypeUint16Array
	case "Int32Array":
		return ir.TypeInt32Array
	case "Uint32Array":
		return ir.TypeUint32Array
	case "Float32Array":
		return ir.TypeFloat32Array
	case "Float64Array":
		return ir.TypeFloat64Array
	case "BigInt64Array":
		return ir.TypeBigInt64Array
	case "BigUint64Array":
		return ir.TypeBigUint64Array
	case "DataView":
		return ir.TypeDataView
	case "ArrayBuffer":
		return ir.TypeArrayBuffer
	case "Map":
		return ir.TypeMap
	case "Set":
		return ir.TypeSet
	case "TextEncoder":
		return ir.TypeTextEncoder
	case "TextDecoder":
		return ir.TypeTextDecoder
	case "Buffer":
		return ir.TypeBuffer
	case "void", "any", "":
		return ir.TypeVoid
	default:
		if strings.HasPrefix(value, "Map<") && strings.HasSuffix(value, ">") {
			return ir.TypeMap
		}
		if strings.HasPrefix(value, "Set<") && strings.HasSuffix(value, ">") {
			return ir.TypeSet
		}
		if before, ok := strings.CutSuffix(value, "[]"); ok {
			elem := before
			return ir.Type(string(toIRType(elem)) + "[]")
		}
		if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			if fields, ok := tupleFields(value); ok {
				return ir.Type("object:" + anonymousShapeName(fields))
			}
		}
		if strings.Contains(value, "=>") || strings.HasPrefix(value, "(") || strings.HasPrefix(value, "Function") {
			return ir.TypeClosure
		}
		return ir.Type("object:" + value)
	}
}

func tupleFields(typeStr string) ([]ir.Field, bool) {
	if !strings.HasPrefix(typeStr, "[") || !strings.HasSuffix(typeStr, "]") {
		return nil, false
	}
	inner := strings.TrimSpace(typeStr[1 : len(typeStr)-1])
	if inner == "" {
		return nil, false
	}
	parts := strings.Split(inner, ",")
	var fields []ir.Field
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		if idx := strings.Index(trimmed, ":"); idx != -1 {
			trimmed = strings.TrimSpace(trimmed[idx+1:])
		}
		elemType := toIRType(trimmed)
		fields = append(fields, ir.Field{
			Name: strconv.Itoa(i),
			Type: elemType,
		})
	}
	return fields, true
}

func toIRSpan(path string, span typescriptgo.SourceSpan) ir.SourceSpan {
	return ir.SourceSpan{Path: path, Offset: span.Start, Length: span.Length}
}

func fieldIndex(shape ir.ObjectShape, name string) int {
	for index, field := range shape.Fields {
		if field.Name == name {
			return index
		}
	}
	return -1
}

func sourceError(path string, span typescriptgo.SourceSpan, err error) error {
	if err == nil {
		return nil
	}
	return &DiagnosticError{Path: path, Offset: span.Start, Length: span.Length, Err: err}
}

type DiagnosticError struct {
	Path   string
	Offset int
	Length int
	Err    error
}

func (e *DiagnosticError) Error() string {
	return e.Err.Error()
}

func isBigIntTypedArray(t ir.Type) bool {
	switch t {
	case ir.TypeBigInt64Array, ir.TypeBigUint64Array:
		return true
	default:
		return false
	}
}

func isNumberTypedArray(t ir.Type) bool {
	switch t {
	case ir.TypeBuffer, ir.TypeInt8Array, ir.TypeUint8Array, ir.TypeUint8ClampedArray,
		ir.TypeInt16Array, ir.TypeUint16Array, ir.TypeInt32Array, ir.TypeUint32Array,
		ir.TypeFloat32Array, ir.TypeFloat64Array:
		return true
	default:
		return false
	}
}

func isTypedArrayType(t ir.Type) bool {
	switch t {
	case ir.TypeBuffer, ir.TypeInt8Array, ir.TypeUint8Array, ir.TypeUint8ClampedArray,
		ir.TypeInt16Array, ir.TypeUint16Array, ir.TypeInt32Array, ir.TypeUint32Array,
		ir.TypeFloat32Array, ir.TypeFloat64Array, ir.TypeBigInt64Array, ir.TypeBigUint64Array:
		return true
	default:
		return false
	}
}

func isMapType(t ir.Type) bool {
	return t == ir.TypeMap
}

func isSetType(t ir.Type) bool {
	return t == ir.TypeSet
}

func statementAlwaysReturns(stmt typescriptgo.SyntaxStatement) bool {
	switch stmt.Kind {
	case "return", "throw":
		return true
	case "block":
		return slices.ContainsFunc(stmt.Body, statementAlwaysReturns)
	case "if":
		if len(stmt.Then) == 0 || len(stmt.Else) == 0 {
			return false
		}
		thenReturns := slices.ContainsFunc(stmt.Then, statementAlwaysReturns)
		elseReturns := slices.ContainsFunc(stmt.Else, statementAlwaysReturns)
		return thenReturns && elseReturns
	case "switch":
		hasDefault := false
		fallthroughReturns := false
		for i := len(stmt.Cases) - 1; i >= 0; i-- {
			c := stmt.Cases[i]
			if c.Expression == nil {
				hasDefault = true
			}
			caseReturns := slices.ContainsFunc(c.Statements, statementAlwaysReturns)
			if len(c.Statements) == 0 {
				caseReturns = fallthroughReturns
			}
			if !caseReturns {
				return false
			}
			fallthroughReturns = caseReturns
		}
		return hasDefault
	default:
		return false
	}
}
