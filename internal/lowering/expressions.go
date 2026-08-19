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
			return "", "", fmt.Errorf("empty array literal needs an explicit runtime representation")
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
			var arrType ir.Type = ir.TypeNumberArray
			for i, element := range expression.Arguments {
				value, typ, err := lowerExpression(path, element, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				if i == 0 {
					if typ == ir.TypeString {
						arrType = ir.TypeStringArray
					} else if typ != ir.TypeNumber {
						return "", "", fmt.Errorf("array literal currently supports number or string elements only")
					}
				} else {
					expectedElem := ir.TypeNumber
					if arrType == ir.TypeStringArray {
						expectedElem = ir.TypeString
					}
					if typ != expectedElem {
						return "", "", fmt.Errorf("inconsistent element type in array literal")
					}
				}
				arguments = append(arguments, value)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: arrType, Result: result, Args: arguments, Span: toIRSpan(path, expression.Span)})
			return result, arrType, nil
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
		value, valType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if expression.Operator == "!" {
			if valType != ir.TypeBool {
				return "", "", fmt.Errorf("unary ! requires a bool operand")
			}
			if result == "" {
				result = nextTemp(counter)
			}
			falseConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: falseConst, Value: "false", Span: toIRSpan(path, expression.Span)})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: result, Operator: "==", Args: []string{value, falseConst}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeBool, nil
		}
		if expression.Operator == "-" {
			if valType != ir.TypeNumber {
				return "", "", fmt.Errorf("unary - requires a number operand")
			}
			if result == "" {
				result = nextTemp(counter)
			}
			zeroConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, expression.Span)})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeNumber, Result: result, Operator: "-", Args: []string{zeroConst, value}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeNumber, nil
		}
		if expression.Operator == "+" {
			if valType != ir.TypeNumber {
				return "", "", fmt.Errorf("unary + requires a number operand")
			}
			return value, ir.TypeNumber, nil
		}
		if expression.Operator == "~" {
			if valType != ir.TypeNumber {
				return "", "", fmt.Errorf("unary ~ requires a number operand")
			}
			if result == "" {
				result = nextTemp(counter)
			}
			minusOneConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: minusOneConst, Value: "-1", Span: toIRSpan(path, expression.Span)})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeNumber, Result: result, Operator: "^", Args: []string{value, minusOneConst}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeNumber, nil
		}
		return "", "", fmt.Errorf("unsupported unary operator %q", expression.Operator)
	case "binary":
		if expression.Operator == "??" {
			leftVal, leftTyp, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			rightVal, rightTyp, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			if leftTyp != rightTyp {
				return "", "", fmt.Errorf("operator ?? does not support %s and %s", leftTyp, rightTyp)
			}
			nullConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: nullConst, Value: "null", Span: toIRSpan(path, expression.Span)})
			cmpNull := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpNull, Operator: "!=", Args: []string{leftVal, nullConst}, Span: toIRSpan(path, expression.Span)})

			undefConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: undefConst, Value: "undefined", Span: toIRSpan(path, expression.Span)})
			cmpUndef := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpUndef, Operator: "!=", Args: []string{leftVal, undefConst}, Span: toIRSpan(path, expression.Span)})

			cond := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: cond, Operator: "&&", Args: []string{cmpNull, cmpUndef}, Span: toIRSpan(path, expression.Span)})

			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: leftTyp, Result: result, Args: []string{cond, leftVal, rightVal}, Span: toIRSpan(path, expression.Span)})
			return result, leftTyp, nil
		}
		left, leftType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		right, rightType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if leftType != rightType {
			return "", "", fmt.Errorf("operator %q does not support %s and %s", expression.Operator, leftType, rightType)
		}
		if leftType == ir.TypeBool {
			if expression.Operator == "&&" || expression.Operator == "||" {
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
				return result, ir.TypeBool, nil
			}
			if isComparison(expression.Operator) {
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
				return result, ir.TypeBool, nil
			}
			return "", "", fmt.Errorf("operator %q does not support bool operands", expression.Operator)
		}
		if leftType != ir.TypeNumber && leftType != ir.TypeString {
			return "", "", fmt.Errorf("operator %q does not support %s and %s", expression.Operator, leftType, rightType)
		}
		if leftType == ir.TypeString {
			if expression.Operator != "+" && !isComparison(expression.Operator) {
				return "", "", fmt.Errorf("operator %q does not support string operands", expression.Operator)
			}
		}
		if isComparison(expression.Operator) {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeBool, nil
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: leftType, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
		return result, leftType, nil
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
		if expression.Left != nil && expression.Left.Kind == "identifier" {
			if (expression.Left.Text == "process" || expression.Left.Text == "__scriptgo") && expression.Text == "argv" {
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeStringArray, Result: result, Callee: "__process.argv", Args: nil, Span: toIRSpan(path, expression.Span)})
				return result, ir.TypeStringArray, nil
			}
		}
		if expression.Left != nil && expression.Left.Kind == "property" && expression.Left.Left != nil && expression.Left.Left.Kind == "identifier" && expression.Left.Left.Text == "process" && expression.Left.Text == "env" {
			keyTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: keyTemp, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__process.env", Args: []string{keyTemp}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeString, nil
		}
		object, objectType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if (objectType == ir.TypeString || objectType == ir.TypeNumberArray || objectType == ir.TypeStringArray) && expression.Text == "length" {
			if result == "" {
				result = nextTemp(counter)
			}
			callee := "__string.length"
			if objectType != ir.TypeString {
				callee = "__array.length"
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: callee, Args: []string{object}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeNumber, nil
		}
		className := strings.TrimPrefix(string(objectType), "object:")
		shape, ok := shapes[className]
		if !ok {
			return "", "", fmt.Errorf("unknown object shape %q", className)
		}
		for _, field := range shape.Fields {
			if field.Name != expression.Text {
				continue
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: field.Type, Result: result, Callee: className, Field: field.Name, FieldIndex: fieldIndex(shape, field.Name), Args: []string{object}, Span: toIRSpan(path, expression.Span)})
			return result, field.Type, nil
		}
		return "", "", fmt.Errorf("unknown field %q on object %q", expression.Text, className)
	case "new":
		className := callName(expression.Left)
		shape, ok := shapes[className]
		if !ok {
			return "", "", fmt.Errorf("unknown class %q", className)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpObjectNew, Type: ir.Type("object:" + className), Result: result, Callee: className, FieldCount: len(shape.Fields), Span: toIRSpan(path, expression.Span)})
		for _, field := range shape.Fields {
			initializer := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: field.Type, Result: initializer, Value: field.Value, Span: field.Span})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: field.Name, FieldIndex: fieldIndex(shape, field.Name), Args: []string{result, initializer}, Span: field.Span})
		}
		for i, argument := range expression.Arguments {
			if i < len(shape.Fields) {
				argVal, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				field := shape.Fields[i]
				function.Body = append(function.Body, ir.Instruction{
					Op:         ir.OpFieldSet,
					Type:       ir.TypeVoid,
					Callee:     className,
					Field:      field.Name,
					FieldIndex: i,
					Args:       []string{result, argVal},
					Span:       toIRSpan(path, argument.Span),
				})
			}
		}
		return result, ir.Type("object:" + className), nil
	case "call":
		if expression.Left != nil && expression.Left.Kind == "property" && expression.Left.Left != nil {
			methodName := expression.Left.Text
			receiver, receiverType, err := lowerExpression(path, expression.Left.Left, "", function, env, counter, shapes, signatures)
			if err == nil {
				if receiverType == ir.TypeString && isStringMethod(methodName) {
					args := []string{receiver}
					for _, argument := range expression.Arguments {
						value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						args = append(args, value)
					}
					if result == "" {
						result = nextTemp(counter)
					}
					returnType := ir.TypeNumber
					switch methodName {
					case "slice", "trim", "replace", "substring":
						returnType = ir.TypeString
					case "startsWith", "endsWith":
						returnType = ir.TypeBool
					case "split":
						returnType = ir.TypeStringArray
					}
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: "__string." + methodName, Args: args, Span: toIRSpan(path, expression.Span)})
					return result, returnType, nil
				}
				if (receiverType == ir.TypeNumberArray || receiverType == ir.TypeStringArray) && isArrayMethod(methodName) {
					args := []string{receiver}
					for _, argument := range expression.Arguments {
						value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						args = append(args, value)
					}
					if result == "" {
						result = nextTemp(counter)
					}
					returnType := ir.TypeNumber
					switch methodName {
					case "slice":
						returnType = receiverType
					case "includes":
						returnType = ir.TypeBool
					case "pop":
						if receiverType == ir.TypeNumberArray {
							returnType = ir.TypeNumber
						} else {
							returnType = ir.TypeString
						}
					}
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: "__array." + methodName, Args: args, Span: toIRSpan(path, expression.Span)})
					return result, returnType, nil
				}
				if strings.HasPrefix(string(receiverType), "object:") {
					className := strings.TrimPrefix(string(receiverType), "object:")
					mangled := className + "_" + methodName
					if target, ok := signatures[mangled]; ok {
						args := []string{receiver}
						for _, argument := range expression.Arguments {
							argVal, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
							if err != nil {
								return "", "", err
							}
							args = append(args, argVal)
						}
						if result == "" {
							result = nextTemp(counter)
						}
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpCall,
							Type:   target.ReturnType,
							Result: result,
							Callee: mangled,
							Args:   args,
							Span:   toIRSpan(path, expression.Span),
						})
						return result, target.ReturnType, nil
					}
				}
			}
		}
		callee := callName(expression.Left)
		if intrinsic, ok := builtinIntrinsic(callee); ok {
			return intrinsic.Lower(IntrinsicCall{Path: path, Expression: expression, Result: result, Function: function, Env: env, Counter: counter, Shapes: shapes, Signatures: signatures, LowerExpression: lowerExpression}, intrinsic)
		}
		if callee == "" {
			return "", "", fmt.Errorf("unsupported call target")
		}
		args := make([]string, 0, len(expression.Arguments))
		for _, argument := range expression.Arguments {
			value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			args = append(args, value)
		}
		target, ok := signatures[callee]
		if !ok {
			return "", "", fmt.Errorf("unknown function %q", callee)
		}
		callee = target.Name
		if len(args) < len(target.Parameters) {
			defaults := defaultParamsIndex[callee]
			if defaults != nil {
				for i := len(args); i < len(target.Parameters); i++ {
					if initExpr, ok := defaults[i]; ok {
						val, _, err := lowerExpression(path, initExpr, "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						args = append(args, val)
					}
				}
			}
		}
		if len(target.Parameters) > 0 && (target.Parameters[len(target.Parameters)-1].Type == ir.TypeStringArray || target.Parameters[len(target.Parameters)-1].Type == ir.TypeNumberArray) {
			restType := target.Parameters[len(target.Parameters)-1].Type
			fixed := len(target.Parameters) - 1
			if len(args) < fixed {
				return "", "", fmt.Errorf("call to %q has too few arguments", callee)
			}
			restArgs := append([]string(nil), args[fixed:]...)
			args = args[:fixed]
			array := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: restType, Result: array, Args: restArgs, Span: toIRSpan(path, expression.Span)})
			args = append(args, array)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		returnType := target.ReturnType
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: callee, Args: args, Span: toIRSpan(path, expression.Span)})
		return result, returnType, nil
	default:
		return "", "", fmt.Errorf("unsupported expression %q", expression.Kind)
	}
}

func callName(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil {
		return ""
	}
	if expression.Kind == "identifier" {
		return expression.Text
	}
	if expression.Kind == "property" && expression.Left != nil && expression.Left.Kind == "identifier" {
		return expression.Left.Text + "." + expression.Text
	}
	return ""
}

func nextTemp(counter *int) string {
	value := "t" + strconv.Itoa(*counter)
	*counter++
	return value
}

func isComparison(operator string) bool {
	return operator == "==" || operator == "===" || operator == "!=" || operator == "!==" || operator == "<" || operator == "<=" || operator == ">" || operator == ">="
}

func isStringMethod(name string) bool {
	switch name {
	case "indexOf", "lastIndexOf", "slice", "startsWith", "endsWith", "trim", "replace", "substring", "split":
		return true
	default:
		return false
	}
}

func isArrayMethod(name string) bool {
	switch name {
	case "push", "pop", "slice", "indexOf", "includes":
		return true
	default:
		return false
	}
}

func stringMethod(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil || expression.Kind != "property" || expression.Left == nil {
		return ""
	}
	if isStringMethod(expression.Text) {
		return expression.Text
	}
	return ""
}

func arrayMethod(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil || expression.Kind != "property" || expression.Left == nil {
		return ""
	}
	if isArrayMethod(expression.Text) {
		return expression.Text
	}
	return ""
}


