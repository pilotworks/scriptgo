package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerBinaryExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if expression.Operator == "??" {
		leftVal, leftTyp, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		rightVal, rightTyp, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if (expression.Right.Kind == "null" || expression.Right.Kind == "undefined") && rightTyp != leftTyp {
			rightTyp = leftTyp
			nullVal := "null"
			switch leftTyp {
			case ir.TypeNumber:
				nullVal = "0"
			case ir.TypeBool:
				nullVal = "false"
			}
			rightVal = nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: leftTyp, Result: rightVal, Value: nullVal, Span: toIRSpan(path, expression.Right.Span)})
		}
		if leftTyp == ir.TypeVoid || leftTyp == ir.TypePointer || leftTyp == "ptr" || (expression.Left != nil && (expression.Left.Kind == "null" || expression.Left.Kind == "undefined")) {
			if result != "" && result != rightVal {
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCheckedCast,
					Type:   rightTyp,
					Result: result,
					Args:   []string{rightVal},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, rightTyp, nil
			}
			return rightVal, rightTyp, nil
		}
		if rightTyp == ir.TypeVoid || (expression.Right != nil && (expression.Right.Kind == "null" || expression.Right.Kind == "undefined")) {
			if result != "" && result != leftVal {
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCheckedCast,
					Type:   leftTyp,
					Result: result,
					Args:   []string{leftVal},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, leftTyp, nil
			}
			return leftVal, leftTyp, nil
		}
		if leftTyp != rightTyp {
			if strings.HasPrefix(string(leftTyp), "object:") && strings.HasPrefix(string(rightTyp), "object:") {
				castTemp := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCheckedCast,
					Type:   leftTyp,
					Result: castTemp,
					Args:   []string{rightVal},
					Span:   toIRSpan(path, expression.Span),
				})
				rightVal = castTemp
			} else {
				return "", "", fmt.Errorf("operator ?? does not support %s and %s", leftTyp, rightTyp)
			}
		}
		nullConst := nextTemp(counter)
		nullVal := "null"
		switch leftTyp {
		case ir.TypeNumber:
			nullVal = "0"
		case ir.TypeBool:
			nullVal = "false"
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: leftTyp, Result: nullConst, Value: nullVal, Span: toIRSpan(path, expression.Span)})
		cmpNull := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpNull, Operator: "!=", Args: []string{leftVal, nullConst}, Span: toIRSpan(path, expression.Span)})

		cond := cmpNull
		if leftTyp == ir.TypeString {
			undefConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: undefConst, Value: "undefined", Span: toIRSpan(path, expression.Span)})
			cmpUndef := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpUndef, Operator: "!=", Args: []string{leftVal, undefConst}, Span: toIRSpan(path, expression.Span)})

			condTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: condTemp, Operator: "&&", Args: []string{cmpNull, cmpUndef}, Span: toIRSpan(path, expression.Span)})
			cond = condTemp
		}

		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: leftTyp, Result: result, Args: []string{cond, leftVal, rightVal}, Span: toIRSpan(path, expression.Span)})
		return result, leftTyp, nil
	}
	if expression.Operator == "instanceof" {
		left, _, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		targetClass := ""
		if expression.Right != nil && (expression.Right.Kind == "identifier" || expression.Right.Kind == "type") {
			targetClass = expression.Right.Text
		} else {
			return "", "", fmt.Errorf("instanceof requires a class identifier on the right")
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpInstanceOf,
			Type:   ir.TypeBool,
			Result: result,
			Args:   []string{left},
			Value:  targetClass,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeBool, nil
	}
	if expression.Operator == "in" {
		return lowerInExpression(path, expression, result, function, env, counter, shapes, signatures)
	}
	if expression.Operator == "," {
		_, _, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		return lowerExpression(path, expression.Right, result, function, env, counter, shapes, signatures)
	}
	if expression.Operator == "=" {
		if expression.Left != nil && expression.Left.Kind == "identifier" {
			varName := expression.Left.Text
			varType, ok := env[varName]
			if !ok {
				return "", "", fmt.Errorf("assignment to unknown variable %q", varName)
			}
			val, valType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			if valType != varType && varType != ir.TypeUnknown {
				return "", "", fmt.Errorf("assignment type mismatch for %q: %s := %s", varName, varType, valType)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   varType,
				Result: varName,
				Args:   []string{val},
				Span:   toIRSpan(path, expression.Span),
			})
			if result != "" {
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpAssign,
					Type:   varType,
					Result: result,
					Args:   []string{val},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, varType, nil
			}
			return val, varType, nil
		}
	}
	left, leftType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	right, rightType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if leftType == ir.TypeVoid && rightType == ir.TypeVoid {
		if result == "" {
			result = nextTemp(counter)
		}
		val := "true"
		if expression.Operator == "!==" || expression.Operator == "!=" {
			val = "false"
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeBool,
			Result: result,
			Value:  val,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeBool, nil
	}
	if leftType != rightType {
		if isComparison(expression.Operator) && (expression.Right != nil && (expression.Right.Kind == "null" || expression.Right.Kind == "undefined")) && isPointerLikeType(leftType) {
			right = nextTemp(counter)
			rightType = leftType
			nullVal := "null"
			if expression.Right.Kind == "undefined" && leftType == ir.TypeString {
				nullVal = "undefined"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   leftType,
				Result: right,
				Value:  nullVal,
				Span:   toIRSpan(path, expression.Right.Span),
			})
		} else if isComparison(expression.Operator) && (expression.Left != nil && (expression.Left.Kind == "null" || expression.Left.Kind == "undefined")) && isPointerLikeType(rightType) {
			left = nextTemp(counter)
			leftType = rightType
			nullVal := "null"
			if expression.Left.Kind == "undefined" && rightType == ir.TypeString {
				nullVal = "undefined"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   rightType,
				Result: left,
				Value:  nullVal,
				Span:   toIRSpan(path, expression.Left.Span),
			})
		} else if isComparison(expression.Operator) && (expression.Operator == "===" || expression.Operator == "!==" || expression.Operator == "==" || expression.Operator == "!=" || leftType == ir.TypeUnknown || rightType == ir.TypeUnknown) {
			if leftType != ir.TypeUnknown {
				boxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{left}, Span: toIRSpan(path, expression.Span)})
				left = boxed
				leftType = ir.TypeUnknown
			}
			if rightType != ir.TypeUnknown {
				boxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{right}, Span: toIRSpan(path, expression.Span)})
				right = boxed
				rightType = ir.TypeUnknown
			}
		} else if expression.Operator == "+" && (leftType == ir.TypeString || rightType == ir.TypeString) {
			if leftType != ir.TypeString {
				strTemp := nextTemp(counter)
				callee := "__string.fromNumber"
				if leftType == ir.TypeBool {
					callee = "__string.fromBool"
				} else if leftType == ir.TypeBigInt {
					callee = "__string.fromBigInt"
				} else if leftType != ir.TypeNumber {
					return "", "", fmt.Errorf("operator %q does not support %s and %s", expression.Operator, leftType, rightType)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: strTemp, Callee: callee, Args: []string{left}, Span: toIRSpan(path, expression.Span)})
				left = strTemp
				leftType = ir.TypeString
			}
			if rightType != ir.TypeString {
				strTemp := nextTemp(counter)
				callee := "__string.fromNumber"
				if rightType == ir.TypeBool {
					callee = "__string.fromBool"
				} else if rightType == ir.TypeBigInt {
					callee = "__string.fromBigInt"
				} else if rightType != ir.TypeNumber {
					return "", "", fmt.Errorf("operator %q does not support %s and %s", expression.Operator, leftType, rightType)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: strTemp, Callee: callee, Args: []string{right}, Span: toIRSpan(path, expression.Span)})
				right = strTemp
				rightType = ir.TypeString
			}
		} else {
			return "", "", fmt.Errorf("operator %q does not support %s and %s", expression.Operator, leftType, rightType)
		}
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
	if (isPointerLikeType(leftType) && leftType != ir.TypeString) || leftType == ir.TypeSymbol || leftType == ir.TypeClosure || leftType == ir.TypeUnknown {
		if isComparison(expression.Operator) {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeBool, nil
		}
		return "", "", fmt.Errorf("operator %q does not support %s operands", expression.Operator, leftType)
	}
	if leftType != ir.TypeNumber && leftType != ir.TypeString && leftType != ir.TypeBigInt {
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
}

func isPointerLikeType(typ ir.Type) bool {
	return strings.HasPrefix(string(typ), "object:") ||
		strings.HasSuffix(string(typ), "[]") ||
		typ == ir.TypeObject ||
		typ == ir.TypeClosure ||
		typ == ir.TypeString ||
		typ == ir.TypeMap ||
		typ == ir.TypeSet ||
		typ == ir.TypeArrayBuffer ||
		typ == ir.TypeDataView ||
		typ == ir.TypeTextEncoder ||
		typ == ir.TypeTextDecoder ||
		typ == ir.TypeBuffer ||
		typ == ir.TypePointer ||
		typ == "ptr"
}
