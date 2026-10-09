package lowering

import (
	"fmt"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerObjectConversion lowers Object() and Object(value): no argument,
// undefined, or null yields a new empty object; an object is returned as
// itself; a boxed value is converted at run time. Wrapper objects for
// primitives (Object(1n), Object("s")) are not modelled.
func lowerObjectConversion(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	emptyObject := func() (string, ir.Type, error) {
		literal := &frontend.SyntaxExpression{Span: call.Expression.Span, Kind: "object_literal"}
		return call.LowerExpression(call.Path, literal, call.Result, call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	}
	if len(call.Expression.Arguments) == 0 {
		return emptyObject()
	}
	value, typ, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}
	switch {
	case typ == ir.TypeVoid || typ == ir.TypePointer:
		return emptyObject()
	case typ == ir.TypeUnknown:
		result := call.Result
		if result == "" {
			result = nextTemp(call.Counter)
		}
		call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeUnknown, Result: result, Callee: "__object.new", Args: []string{value}, Span: toIRSpan(call.Path, call.Expression.Span)})
		return result, ir.TypeUnknown, nil
	case isPointerLikeType(typ) && typ != ir.TypeString:
		return value, typ, nil
	}
	return "", "", fmt.Errorf("Object(%s) creates a primitive wrapper object, which the native subset does not model", typ)
}
