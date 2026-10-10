package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerBinaryExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if expression.Operator == "&&" {
		if value, valueType, handled, err := lowerLogicalAndExpression(path, expression, result, function, env, counter, shapes, signatures); handled {
			return value, valueType, err
		}
	}
	if expression.Operator == "||" {
		if value, valueType, handled, err := lowerLogicalOrExpression(path, expression, result, function, env, counter, shapes, signatures); handled {
			return value, valueType, err
		}
	}
	if expression.Operator == "??" {
		return lowerNullishCoalescingExpression(path, expression, result, function, env, counter, shapes, signatures)
	}
	if expression.Operator == "instanceof" {
		return lowerInstanceofExpression(path, expression, result, function, env, counter, shapes, signatures)
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
	if expression.Operator == "=" || (strings.HasSuffix(expression.Operator, "=") && expression.Operator != "==" && expression.Operator != "===" && expression.Operator != "!=" && expression.Operator != "!==" && expression.Operator != "<=" && expression.Operator != ">=") {
		if value, valueType, handled, err := lowerAssignmentExpression(path, expression, result, function, env, counter, shapes, signatures); handled {
			return value, valueType, err
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
	// TypeScript-Go can preserve the string result type for a dynamic `+`
	// expression even when its operands come from an unknown control-flow
	// branch. JavaScript concatenation converts those operands to strings.
	if expression.Operator == "+" && (leftType == ir.TypeUnknown || rightType == ir.TypeUnknown) &&
		(leftType == ir.TypeString || rightType == ir.TypeString || toIRType(expression.InferredType) == ir.TypeString) {
		var err error
		if left, err = lowerToString(path, expression.Left.Span, left, leftType, function, counter, signatures); err != nil {
			return "", "", err
		}
		if right, err = lowerToString(path, expression.Right.Span, right, rightType, function, counter, signatures); err != nil {
			return "", "", err
		}
		leftType, rightType = ir.TypeString, ir.TypeString
	}
	coerceBoolOperandsToNumber(path, expression, function, counter, &left, &leftType, &right, &rightType)
	if leftType != rightType {
		if value, valueType, handled, err := tryLowerMixedTypeBinary(path, expression, &result, function, env, counter, signatures, &left, &leftType, &right, &rightType); handled {
			return value, valueType, err
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
	if isPointerLikeType(leftType) || leftType == ir.TypeSymbol || leftType == ir.TypeClosure || leftType == ir.TypeUnknown {
		if value, valueType, handled, err := tryLowerReferenceBinary(path, expression, &result, function, env, counter, &left, &leftType, &right, rightType); handled {
			return value, valueType, err
		}
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

func logicalResultIsBool(expression *frontend.SyntaxExpression) bool {
	if expression == nil {
		return false
	}
	if expression.InferredType != "" {
		return toIRType(expression.InferredType) == ir.TypeBool
	}
	if expression.Right == nil {
		return false
	}
	if expression.Right.Kind == "bool" {
		return true
	}
	if expression.Right.Kind == "binary" && isComparison(expression.Right.Operator) {
		return true
	}
	return false
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
		typ == ir.TypeUint8Array ||
		typ == ir.TypeInt8Array ||
		typ == ir.TypeUint8ClampedArray ||
		typ == ir.TypeInt16Array ||
		typ == ir.TypeUint16Array ||
		typ == ir.TypeInt32Array ||
		typ == ir.TypeUint32Array ||
		typ == ir.TypeFloat32Array ||
		typ == ir.TypeFloat64Array ||
		typ == ir.TypeBigInt64Array ||
		typ == ir.TypeBigUint64Array ||
		typ == ir.TypePointer ||
		typ == "ptr"
}
