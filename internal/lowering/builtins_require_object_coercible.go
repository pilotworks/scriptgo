package lowering

import (
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// registerRequireObjectCoercibleIntrinsic registers the RequireObjectCoercible
// step the frontend emits before reading from a destructuring source:
// destructuring null or undefined throws a TypeError.
func registerRequireObjectCoercibleIntrinsic(m map[string]BuiltinIntrinsic) {
	registerCustomIntrinsic(m, []string{"__scriptgo.requireObjectCoercible"}, CategoryECMAScript, "", ir.TypeVoid, 1, 1, lowerRequireObjectCoercible)
}

// lowerRequireObjectCoercible emits `if (source == null) throw TypeError`
// when the source's storage can hold null or undefined; primitive and closure
// storage cannot, so the check folds away statically. The TypeError is raised
// by the runtime unwinder, like other runtime-detected failures, so enclosing
// try/catch/finally blocks observe it the same way.
func lowerRequireObjectCoercible(call IntrinsicCall, _ BuiltinIntrinsic) (string, ir.Type, error) {
	source := call.Expression.Arguments[0]
	if source == nil || source.Kind != "identifier" || !storageMayBeNullish(call.Env[source.Text]) {
		return "", ir.TypeVoid, nil
	}
	span := call.Expression.Span
	isNullish := &frontend.SyntaxExpression{Span: span, Kind: "binary", Operator: "==", Left: source, Right: &frontend.SyntaxExpression{Span: span, Kind: "null", Text: "null", InferredType: "null"}}
	condition, _, err := call.LowerExpression(call.Path, isNullish, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}
	irSpan := toIRSpan(call.Path, span)
	message := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:   ir.OpIf,
		Type: ir.TypeVoid,
		Args: []string{condition},
		Then: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeString, Result: message, Value: "TypeError: Cannot destructure a null or undefined value.", StringLiteral: true, Span: irSpan},
			{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__error.throw", Args: []string{message}, Span: irSpan},
		},
		Span: irSpan,
	})
	return "", ir.TypeVoid, nil
}

// storageMayBeNullish reports IR storage that can hold null or undefined.
func storageMayBeNullish(typ ir.Type) bool {
	switch typ {
	case "", ir.TypeNumber, ir.TypeString, ir.TypeBool, ir.TypeBigInt, ir.TypeSymbol, ir.TypeClosure:
		return false
	}
	return true
}
