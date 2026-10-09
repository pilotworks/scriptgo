package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// Iterator helpers (ES2025) are modelled over a materialized copy of the
// sequence: Iterator.from(array) copies the array, an IteratorObject<T> is
// stored as T[], and each helper is the matching array operation. Results
// match JavaScript for finite sequences; the difference is evaluation order
// (helpers run eagerly, not interleaved with consumption), which the parity
// report documents.

func registerIteratorBuiltins(m map[string]BuiltinIntrinsic) {
	m["Iterator.from"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Iterator.from",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			source := call.Expression.Arguments[0]
			sourceType := toIRType(source.InferredType)
			if !strings.HasSuffix(string(sourceType), "[]") {
				return "", "", fmt.Errorf("Iterator.from supports arrays in the native subset, got %s", source.InferredType)
			}
			copy := &frontend.SyntaxExpression{Span: call.Expression.Span, Kind: "call", InferredType: source.InferredType, Left: &frontend.SyntaxExpression{Span: call.Expression.Span, Kind: "property", Text: "slice", Left: source}}
			return call.LowerExpression(call.Path, copy, call.Result, call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		},
	}
}

// iteratorReceiver reports a receiver TypeScript types as an iterator object.
func iteratorReceiver(expression *frontend.SyntaxExpression) bool {
	if expression == nil || expression.Left == nil || expression.Left.Left == nil {
		return false
	}
	typ := strings.TrimSpace(expression.Left.Left.InferredType)
	return strings.HasPrefix(typ, "IteratorObject<") || strings.HasPrefix(typ, "Iterator<")
}

// lowerIteratorReceiverMethod lowers the iterator-only helpers (take, drop,
// toArray, next) on an iterator stored as an array; the helpers that share
// an array method name (map, filter, ...) are lowered as array methods.
func lowerIteratorReceiverMethod(
	path string,
	expression *frontend.SyntaxExpression,
	receiver string,
	methodName string,
	receiverType ir.Type,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, bool, error) {
	if !iteratorReceiver(expression) || !strings.HasSuffix(string(receiverType), "[]") {
		return "", "", false, nil
	}
	span := expression.Span
	if _, known := env[receiver]; !known {
		env[receiver] = receiverType
	}
	self := &frontend.SyntaxExpression{Span: span, Kind: "identifier", Text: receiver, InferredType: string(receiverType)}
	method := func(name string, args ...*frontend.SyntaxExpression) *frontend.SyntaxExpression {
		return &frontend.SyntaxExpression{Span: span, Kind: "call", InferredType: string(receiverType), Left: &frontend.SyntaxExpression{Span: span, Kind: "property", Text: name, Left: self}, Arguments: args}
	}
	lower := func(expr *frontend.SyntaxExpression) (string, ir.Type, bool, error) {
		value, typ, err := lowerExpression(path, expr, result, function, env, counter, shapes, signatures)
		return value, typ, true, err
	}
	zero := &frontend.SyntaxExpression{Span: span, Kind: "number", Text: "0", InferredType: "number"}
	switch methodName {
	case "toArray":
		return lower(method("slice"))
	case "take":
		if len(expression.Arguments) != 1 {
			return "", "", true, fmt.Errorf("Iterator take requires a limit")
		}
		return lower(method("slice", zero, expression.Arguments[0]))
	case "drop":
		if len(expression.Arguments) != 1 {
			return "", "", true, fmt.Errorf("Iterator drop requires a limit")
		}
		return lower(method("slice", expression.Arguments[0]))
	case "next":
		// { done, value }: done is read before the element is consumed.
		length := &frontend.SyntaxExpression{Span: span, Kind: "property", Text: "length", Left: self, InferredType: "number"}
		done := &frontend.SyntaxExpression{Span: span, Kind: "binary", Operator: "===", Left: length, Right: zero}
		next := &frontend.SyntaxExpression{Span: span, Kind: "object_literal", InferredType: expression.InferredType, Arguments: []*frontend.SyntaxExpression{
			{Span: span, Kind: "property_assignment", Text: "done", Left: done},
			{Span: span, Kind: "property_assignment", Text: "value", Left: method("shift")},
		}}
		return lower(next)
	}
	return "", "", true, fmt.Errorf("iterator method %q is not supported in the native subset", methodName)
}
