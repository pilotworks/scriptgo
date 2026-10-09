package lowering

import (
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerMismatchedArraySearch handles indexOf/lastIndexOf/includes whose
// search value has a different primitive type than the array's unboxed
// elements. Strict equality and SameValueZero never match across types, so
// the result is -1 or false; the arguments are still evaluated in order.
func lowerMismatchedArraySearch(path string, expression *frontend.SyntaxExpression, methodName string, receiverType ir.Type, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, bool, error) {
	if methodName != "indexOf" && methodName != "lastIndexOf" && methodName != "includes" {
		return "", "", false, nil
	}
	if len(expression.Arguments) == 0 || expression.Arguments[0] == nil {
		return "", "", false, nil
	}
	elemType := arrayElementType(receiverType)
	if !isUnboxedPrimitive(elemType) {
		return "", "", false, nil
	}
	searchType := toIRType(expression.Arguments[0].InferredType)
	if searchType == elemType || !(isUnboxedPrimitive(searchType) || searchType == ir.TypeVoid || searchType == ir.TypeSymbol) {
		return "", "", false, nil
	}
	for _, argument := range expression.Arguments {
		if _, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures); err != nil {
			return "", "", true, err
		}
	}
	if result == "" {
		result = nextTemp(counter)
	}
	span := toIRSpan(path, expression.Span)
	if methodName == "includes" {
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: result, Value: "false", Span: span})
		return result, ir.TypeBool, true, nil
	}
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: result, Value: "-1", Span: span})
	return result, ir.TypeNumber, true, nil
}

func isUnboxedPrimitive(typ ir.Type) bool {
	return typ == ir.TypeNumber || typ == ir.TypeString || typ == ir.TypeBool || typ == ir.TypeBigInt
}
