package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
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
	left, leftType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	right, rightType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if leftType != rightType {
		if expression.Operator == "+" && (leftType == ir.TypeString || rightType == ir.TypeString) {
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
	if leftType == ir.TypeSymbol {
		if isComparison(expression.Operator) {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeBool, nil
		}
		return "", "", fmt.Errorf("operator %q does not support symbol operands", expression.Operator)
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

