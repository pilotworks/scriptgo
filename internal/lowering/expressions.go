package lowering

import (
	"fmt"
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	switch expression.Kind {
	case "number":
		typ := ir.TypeNumber
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "bigint":
		typ := ir.TypeBigInt
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "regex":
		return lowerRegexLiteral(path, expression, result, function, env, counter, shapes, signatures)
	case "string":
		typ := ir.TypeString
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "bool":
		typ := ir.TypeBool
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "null", "undefined":
		typ := ir.TypeString
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Kind, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "array":
		if len(expression.Arguments) == 0 {
			arrType := toIRType(expression.InferredType)
			if arrType != ir.TypeNumberArray && arrType != ir.TypeStringArray {
				if varType, ok := env[result]; ok && (varType == ir.TypeNumberArray || varType == ir.TypeStringArray) {
					arrType = varType
				} else {
					arrType = ir.TypeNumberArray
				}
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: arrType, Result: result, Args: nil, Span: toIRSpan(path, expression.Span)})
			return result, arrType, nil
		}
		if result == "" {
			result = nextTemp(counter)
		}
		hasSpread := false
		for _, elem := range expression.Arguments {
			if elem.Kind == "spread" {
				hasSpread = true
				break
			}
		}
		if !hasSpread {
			arguments := make([]string, 0, len(expression.Arguments))
			types := make([]ir.Type, 0, len(expression.Arguments))
			isHomogeneous := true
			for i, element := range expression.Arguments {
				value, typ, err := lowerExpression(path, element, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				arguments = append(arguments, value)
				types = append(types, typ)
				if i > 0 && typ != types[0] {
					isHomogeneous = false
				}
			}
			if isHomogeneous {
				var arrType ir.Type = ir.TypeNumberArray
				if len(types) > 0 && types[0] == ir.TypeString {
					arrType = ir.TypeStringArray
				} else if len(types) > 0 && types[0] != ir.TypeNumber {
					arrType = ir.Type(string(types[0]) + "[]")
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: arrType, Result: result, Args: arguments, Span: toIRSpan(path, expression.Span)})
				return result, arrType, nil
			}

			// Heterogeneous elements -> lower as anonymous tuple object
			var fields []ir.Field
			for i, typ := range types {
				fields = append(fields, ir.Field{
					Name: strconv.Itoa(i),
					Type: typ,
					Span: toIRSpan(path, expression.Arguments[i].Span),
				})
			}
			shapeName := anonymousShapeName(fields)
			if _, ok := shapes[shapeName]; !ok {
				shapes[shapeName] = ir.ObjectShape{
					Name:   shapeName,
					Span:   toIRSpan(path, expression.Span),
					Fields: fields,
				}
			}
			objType := ir.Type("object:" + shapeName)
			function.Body = append(function.Body, ir.Instruction{
				Op:         ir.OpObjectNew,
				Type:       objType,
				Result:     result,
				Callee:     shapeName,
				FieldCount: len(fields),
				Span:       toIRSpan(path, expression.Span),
			})
			for i, field := range fields {
				function.Body = append(function.Body, ir.Instruction{
					Op:         ir.OpFieldSet,
					Type:       ir.TypeVoid,
					Callee:     shapeName,
					Field:      field.Name,
					FieldIndex: i,
					Args:       []string{result, arguments[i]},
					Span:       toIRSpan(path, expression.Span),
				})
			}
			return result, objType, nil
		}
		var arrType ir.Type = ir.TypeNumberArray
		for _, elem := range expression.Arguments {
			if elem.Kind == "spread" {
				val, typ, err := lowerExpression(path, elem.Left, "", function, env, counter, shapes, signatures)
				if err == nil && typ == ir.TypeStringArray {
					arrType = ir.TypeStringArray
					_ = val
					break
				}
			} else {
				val, typ, err := lowerExpression(path, elem, "", function, env, counter, shapes, signatures)
				if err == nil && typ == ir.TypeString {
					arrType = ir.TypeStringArray
					_ = val
					break
				}
			}
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: arrType, Result: result, Args: nil, Span: toIRSpan(path, expression.Span)})
		for _, elem := range expression.Arguments {
			if elem.Kind == "spread" {
				spreadVal, _, err := lowerExpression(path, elem.Left, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				idxVar := nextTemp(counter)
				lenVar := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: idxVar, Value: "0", Span: toIRSpan(path, elem.Span)})
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: lenVar, Callee: "__array.length", Args: []string{spreadVal}, Span: toIRSpan(path, elem.Span)})
				condVar := nextTemp(counter)
				var condBody []ir.Instruction
				condBody = append(condBody, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: condVar, Operator: "<", Args: []string{idxVar, lenVar}, Span: toIRSpan(path, elem.Span)})
				var loopBody []ir.Instruction
				itemVar := nextTemp(counter)
				itemType := ir.TypeNumber
				if arrType == ir.TypeStringArray {
					itemType = ir.TypeString
				}
				loopBody = append(loopBody, ir.Instruction{Op: ir.OpIndex, Type: itemType, Result: itemVar, Args: []string{spreadVal, idxVar}, Span: toIRSpan(path, elem.Span)})
				pushRes := nextTemp(counter)
				loopBody = append(loopBody, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: pushRes, Callee: "__array.push", Args: []string{result, itemVar}, Span: toIRSpan(path, elem.Span)})
				oneConst := nextTemp(counter)
				loopBody = append(loopBody, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: oneConst, Value: "1", Span: toIRSpan(path, elem.Span)})
				nextIdx := nextTemp(counter)
				loopBody = append(loopBody, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeNumber, Result: nextIdx, Operator: "+", Args: []string{idxVar, oneConst}, Span: toIRSpan(path, elem.Span)})
				loopBody = append(loopBody, ir.Instruction{Op: ir.OpAssign, Type: ir.TypeNumber, Result: idxVar, Args: []string{nextIdx}, Span: toIRSpan(path, elem.Span)})

				function.Body = append(function.Body, ir.Instruction{
					Op:   ir.OpWhile,
					Type: ir.TypeVoid,
					Cond: condBody,
					Args: []string{condVar},
					Body: loopBody,
					Span: toIRSpan(path, elem.Span),
				})
			} else {
				itemVal, _, err := lowerExpression(path, elem, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				pushRes := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: pushRes, Callee: "__array.push", Args: []string{result, itemVal}, Span: toIRSpan(path, elem.Span)})
			}
		}
		return result, arrType, nil
	case "index", "optional_index":
		if expression.Left != nil && expression.Left.Kind == "identifier" {
			if shape, isShape := shapes[expression.Left.Text]; isShape {
				// Static constant index (e.g. Color[0])
				if expression.Right != nil && expression.Right.Kind == "number" {
					for _, field := range shape.Fields {
						if field.Value == expression.Right.Text {
							if result == "" {
								result = nextTemp(counter)
							}
							function.Body = append(function.Body, ir.Instruction{
								Op:     ir.OpConst,
								Type:   ir.TypeString,
								Result: result,
								Value:  field.Name,
								Span:   toIRSpan(path, expression.Span),
							})
							return result, ir.TypeString, nil
						}
					}
				}
				// Dynamic variable index on enum (e.g. Color[val])
				idxVal, idxType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
				if err == nil && (idxType == ir.TypeNumber || idxType == ir.TypeString) {
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   ir.TypeString,
						Result: result,
						Value:  "",
						Span:   toIRSpan(path, expression.Span),
					})
					for _, field := range shape.Fields {
						if field.Value != "" {
							targetVal := nextTemp(counter)
							function.Body = append(function.Body, ir.Instruction{
								Op:     ir.OpConst,
								Type:   field.Type,
								Result: targetVal,
								Value:  field.Value,
								Span:   toIRSpan(path, expression.Span),
							})
							cmpRes := nextTemp(counter)
							function.Body = append(function.Body, ir.Instruction{
								Op:       ir.OpCompare,
								Type:     ir.TypeBool,
								Operator: "==",
								Result:   cmpRes,
								Args:     []string{idxVal, targetVal},
								Span:     toIRSpan(path, expression.Span),
							})
							valStr := nextTemp(counter)
							function.Body = append(function.Body, ir.Instruction{
								Op:     ir.OpConst,
								Type:   ir.TypeString,
								Result: valStr,
								Value:  field.Name,
								Span:   toIRSpan(path, expression.Span),
							})
							selectRes := nextTemp(counter)
							function.Body = append(function.Body, ir.Instruction{
								Op:     ir.OpSelect,
								Type:   ir.TypeString,
								Result: selectRes,
								Args:   []string{cmpRes, valStr, result},
								Span:   toIRSpan(path, expression.Span),
							})
							result = selectRes
						}
					}
					return result, ir.TypeString, nil
				}
			}
		}
		if expression.Left != nil && expression.Left.Kind == "property" && expression.Left.Left != nil && expression.Left.Left.Kind == "identifier" && expression.Left.Left.Text == "process" && expression.Left.Text == "env" {
			keyVal, keyType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
			if err != nil || keyType != ir.TypeString {
				return "", "", fmt.Errorf("process.env requires string index")
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__process.env", Args: []string{keyVal}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeString, nil
		}
		array, arrayType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if strings.HasPrefix(string(arrayType), "object:") {
			shapeName := strings.TrimPrefix(string(arrayType), "object:")
			if shape, ok := shapes[shapeName]; ok {
				if expression.Right != nil && expression.Right.Kind == "number" {
					fieldIdx, _ := strconv.Atoi(expression.Right.Text)
					if fieldIdx >= 0 && fieldIdx < len(shape.Fields) {
						field := shape.Fields[fieldIdx]
						if result == "" {
							result = nextTemp(counter)
						}
						function.Body = append(function.Body, ir.Instruction{
							Op:         ir.OpFieldGet,
							Type:       field.Type,
							Result:     result,
							Callee:     shapeName,
							Field:      field.Name,
							FieldIndex: fieldIdx,
							Args:       []string{array},
							Span:       toIRSpan(path, expression.Span),
						})
						return result, field.Type, nil
					}
				}
			}
		}
		index, indexType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if (arrayType != ir.TypeNumberArray && arrayType != ir.TypeStringArray) || indexType != ir.TypeNumber {
			return "", "", fmt.Errorf("array indexing requires an array and number operands")
		}
		if result == "" {
			result = nextTemp(counter)
		}
		resultType := ir.TypeNumber
		if arrayType == ir.TypeStringArray {
			resultType = ir.TypeString
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpIndex, Type: resultType, Result: result, Args: []string{array, index}, Span: toIRSpan(path, expression.Span)})
		return result, resultType, nil
	case "identifier":
		typ, ok := env[expression.Text]
		if ok {
			switch expression.InferredType {
			case "number":
				typ = ir.TypeNumber
			case "string":
				typ = ir.TypeString
			case "bool", "boolean":
				typ = ir.TypeBool
			case "number[]":
				typ = ir.TypeNumberArray
			case "string[]":
				typ = ir.TypeStringArray
			default:
				if expression.InferredType != "" {
					if _, isShape := shapes[expression.InferredType]; isShape {
						typ = ir.Type("object:" + expression.InferredType)
					} else if strings.HasPrefix(expression.InferredType, "object:") {
						shapeName := strings.TrimPrefix(expression.InferredType, "object:")
						if _, isShape := shapes[shapeName]; isShape {
							typ = ir.Type(expression.InferredType)
						}
					}
				}
			}
			if result != "" && result != expression.Text {
				if typ == ir.TypeNumber {
					zeroConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, expression.Span)})
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: typ, Result: result, Operator: "+", Args: []string{expression.Text, zeroConst}, Span: toIRSpan(path, expression.Span)})
					return result, typ, nil
				}
				if typ == ir.TypeString {
					emptyStr := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: emptyStr, Value: "", Span: toIRSpan(path, expression.Span)})
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: typ, Result: result, Operator: "+", Args: []string{expression.Text, emptyStr}, Span: toIRSpan(path, expression.Span)})
					return result, typ, nil
				}
				if typ == ir.TypeBool {
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: typ, Result: result, Operator: "||", Args: []string{expression.Text, expression.Text}, Span: toIRSpan(path, expression.Span)})
					return result, typ, nil
				}
			}
			return expression.Text, typ, nil
		}
		if sig, ok := signatures[expression.Text]; ok {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpClosure,
				Type:   ir.TypeClosure,
				Result: result,
				Callee: sig.Name,
				Args:   nil,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeClosure, nil
		}
		global, ok := builtinGlobal(expression.Text)
		if !ok {
			return "", "", fmt.Errorf("unknown identifier %q", expression.Text)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: global.Type, Result: result, Value: global.Value, Span: toIRSpan(path, expression.Span)})
		return result, global.Type, nil
	case "unary":
		return lowerUnaryExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "postfix_unary":
		return lowerPostfixUnaryExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "as":
		val, valType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		targetIRType := toIRType(expression.Text)
		if valType == targetIRType {
			if result == "" {
				return val, valType, nil
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpAssign, Type: valType, Result: result, Args: []string{val}, Span: toIRSpan(path, expression.Span)})
			return result, valType, nil
		}
		if valType == ir.TypeUnknown || strings.Contains(string(valType), "|") {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCheckedCast,
				Type:   targetIRType,
				Result: result,
				Args:   []string{val},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, targetIRType, nil
		}
		if targetIRType == ir.TypeUnknown {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: result,
				Args:   []string{val},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeUnknown, nil
		}
		if result == "" {
			return val, targetIRType, nil
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpAssign, Type: targetIRType, Result: result, Args: []string{val}, Span: toIRSpan(path, expression.Span)})
		return result, targetIRType, nil
	case "typeof":
		val, valType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			if expression.Left != nil && expression.Left.Kind == "identifier" && expression.Left.Text == "undefined" {
				valType = ir.TypeVoid
			} else {
				return "", "", err
			}
		}
		if result == "" {
			result = nextTemp(counter)
		}
		if valType == ir.TypeUnknown {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpTypeOf,
				Type:   ir.TypeString,
				Result: result,
				Args:   []string{val},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeString, nil
		}
		var typeStr string
		if expression.Left != nil && expression.Left.Kind == "null" {
			typeStr = "object"
		} else if expression.Left != nil && expression.Left.Kind == "undefined" {
			typeStr = "undefined"
		} else {
			switch valType {
			case ir.TypeNumber:
				typeStr = "number"
			case ir.TypeBigInt:
				typeStr = "bigint"
			case ir.TypeSymbol:
				typeStr = "symbol"
			case ir.TypeString:
				typeStr = "string"
			case ir.TypeBool:
				typeStr = "boolean"
			case ir.TypeVoid:
				typeStr = "undefined"
			default:
				typeStr = "object"
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: result,
			Value:  typeStr,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeString, nil
	case "binary":
		return lowerBinaryExpression(path, expression, result, function, env, counter, shapes, signatures)

	case "template":
		if len(expression.Arguments) == 0 {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: result, Value: "", Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeString, nil
		}
		var currentResult string
		for index, arg := range expression.Arguments {
			val, valType, err := lowerExpression(path, arg, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			strVal := val
			if valType != ir.TypeString {
				strTemp := nextTemp(counter)
				callee := "__string.fromNumber"
				if valType == ir.TypeBool {
					callee = "__string.fromBool"
				} else if valType != ir.TypeNumber {
					return "", "", fmt.Errorf("template expression does not support %s in interpolation", valType)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: strTemp, Callee: callee, Args: []string{val}, Span: toIRSpan(path, arg.Span)})
				strVal = strTemp
			}
			if index == 0 {
				currentResult = strVal
			} else {
				concatTemp := nextTemp(counter)
				if index == len(expression.Arguments)-1 && result != "" {
					concatTemp = result
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeString, Result: concatTemp, Operator: "+", Args: []string{currentResult, strVal}, Span: toIRSpan(path, expression.Span)})
				currentResult = concatTemp
			}
		}
		return currentResult, ir.TypeString, nil
	case "conditional":
		condition, conditionType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if conditionType != ir.TypeBool {
			return "", "", fmt.Errorf("conditional expression requires a bool condition")
		}
		whenTrue, trueType, err := lowerExpression(path, expression.WhenTrue, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		whenFalse, falseType, err := lowerExpression(path, expression.WhenFalse, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if trueType != falseType {
			return "", "", fmt.Errorf("conditional branches must have the same type")
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: trueType, Result: result, Args: []string{condition, whenTrue, whenFalse}, Span: toIRSpan(path, expression.Span)})
		return result, trueType, nil
	case "property", "optional_property":
		return lowerPropertyExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "object_literal":
		return lowerObjectLiteralExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "new":
		return lowerNewExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "call":
		return lowerCallExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "arrow_function":
		if expression.Function != nil {
			return lowerClosureExpression(path, expression.Function, result, function, env, counter, shapes, signatures)
		}
		return "", "", fmt.Errorf("arrow function expression missing body")
	case "await":
		val, typ, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if result == "" {
			result = nextTemp(counter)
		}
		retType := ir.TypeNumber
		if typ == ir.TypeString {
			retType = ir.TypeString
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   retType,
			Result: result,
			Callee: "__async.await",
			Args:   []string{val},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, nil
	default:
		return "", "", fmt.Errorf("unsupported expression %q", expression.Kind)
	}
}

func nextTemp(counter *int) string {
	value := "t" + strconv.Itoa(*counter)
	*counter++
	return value
}

func isComparison(operator string) bool {
	return operator == "==" || operator == "===" || operator == "!=" || operator == "!==" || operator == "<" || operator == "<=" || operator == ">" || operator == ">="
}



