package lowering

import (
	"fmt"
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
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
		if function.ReturnType != ir.TypeVoid {
			return ir.Function{}, fmt.Errorf("function %q does not return %s", function.Name, function.ReturnType)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: function.Span})
	}
	return function, nil
}

func lowerStatement(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	switch statement.Kind {
	case "variable":
		if statement.Expression == nil {
			return fmt.Errorf("variable %q has no initializer", statement.Name)
		}
		declaredType := toIRType(statement.Type)
		if statement.Type == "" && statement.InferredType != "" {
			declaredType = toIRType(statement.InferredType)
		}
		if statement.Expression.Kind == "identifier" {
			srcType, ok := env[statement.Expression.Text]
			if !ok {
				global, isGlobal := builtinGlobal(statement.Expression.Text)
				if isGlobal {
					env[statement.Name] = global.Type
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: global.Type, Result: statement.Name, Value: global.Value, Span: toIRSpan(path, statement.Span)})
					return nil
				}
				return fmt.Errorf("unknown identifier %q", statement.Expression.Text)
			}
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
		value, typ, err := lowerExpression(path, statement.Expression, statement.Name, function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if value != statement.Name {
			return fmt.Errorf("variable %q produced unnamed value %q", statement.Name, value)
		}
		env[statement.Name] = typ
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
		value, typ, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
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
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBreak, Type: ir.TypeVoid, Span: toIRSpan(path, statement.Span)})
	case "continue":
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpContinue, Type: ir.TypeVoid, Span: toIRSpan(path, statement.Span)})
	case "dowhile":
		return lowerDoWhile(path, statement, function, env, counter, shapes, signatures)
	case "forof":
		return lowerForOf(path, statement, function, env, counter, shapes, signatures)
	case "forin":
		return lowerForIn(path, statement, function, env, counter, shapes, signatures)
	case "switch":
		return lowerSwitch(path, statement, function, env, counter, shapes, signatures)
	case "index_set":
		arrVal, arrType, err := lowerExpression(path, statement.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		idxVal, _, err := lowerExpression(path, statement.Right, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		val, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if (arrType == ir.TypeNumberArray && valType != ir.TypeNumber) || (arrType == ir.TypeStringArray && valType != ir.TypeString) {
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
		val, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
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
		val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
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
	for name, typ := range parentEnv {
		env[name] = typ
	}
	for _, statement := range statements {
		if err := lowerStatement(path, statement, &branch, env, counter, shapes, signatures); err != nil {
			return nil, err
		}
	}
	return branch.Body, nil
}

func toIRType(value string) ir.Type {
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
	if strings.HasPrefix(value, "object:") {
		return ir.Type(value)
	}
	switch value {
	case "number":
		return ir.TypeNumber
	case "bigint":
		return ir.TypeBigInt
	case "bigint[]":
		return ir.TypeBigIntArray
	case "RegExp":
		return ir.Type("object:RegExp")
	case "string", "null", "undefined":
		return ir.TypeString
	case "bool", "boolean":
		return ir.TypeBool
	case "number[]":
		return ir.TypeNumberArray
	case "string[]":
		return ir.TypeStringArray
	case "closure", "function":
		return ir.TypeClosure
	case "unknown":
		return ir.TypeUnknown
	case "void", "any", "":
		return ir.TypeVoid
	default:
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
		elemType := toIRType(strings.TrimSpace(part))
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
	return fmt.Errorf("%s:%d+%d: %w", path, span.Start, span.Length, err)
}

func statementAlwaysReturns(stmt typescriptgo.SyntaxStatement) bool {
	switch stmt.Kind {
	case "return", "throw":
		return true
	case "block":
		for _, s := range stmt.Body {
			if statementAlwaysReturns(s) {
				return true
			}
		}
		return false
	case "if":
		if len(stmt.Then) == 0 || len(stmt.Else) == 0 {
			return false
		}
		thenReturns := false
		for _, s := range stmt.Then {
			if statementAlwaysReturns(s) {
				thenReturns = true
				break
			}
		}
		elseReturns := false
		for _, s := range stmt.Else {
			if statementAlwaysReturns(s) {
				elseReturns = true
				break
			}
		}
		return thenReturns && elseReturns
	case "switch":
		hasDefault := false
		fallthroughReturns := false
		for i := len(stmt.Cases) - 1; i >= 0; i-- {
			c := stmt.Cases[i]
			if c.Expression == nil {
				hasDefault = true
			}
			caseReturns := false
			for _, s := range c.Statements {
				if statementAlwaysReturns(s) {
					caseReturns = true
					break
				}
			}
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

