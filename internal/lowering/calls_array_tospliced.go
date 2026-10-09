package lowering

import (
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerToSplicedWithItems lowers arr.toSpliced(start, deleteCount, ...items)
// from array operations whose index clamping matches toSpliced:
//
//	head = arr.slice(0, start)
//	rest = arr.toSpliced(start, deleteCount).slice(head.length)
//	result = head.concat([...items]).concat(rest)
//
// start and deleteCount are evaluated once, before the items, as in
// JavaScript argument order.
func lowerToSplicedWithItems(path string, expression *frontend.SyntaxExpression, receiver string, receiverType ir.Type, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	span := expression.Span
	ident := func(name string) *frontend.SyntaxExpression {
		return &frontend.SyntaxExpression{Span: span, Kind: "identifier", Text: name, InferredType: string(env[name])}
	}
	call := func(target *frontend.SyntaxExpression, method string, args ...*frontend.SyntaxExpression) *frontend.SyntaxExpression {
		return &frontend.SyntaxExpression{Span: span, Kind: "call", Left: &frontend.SyntaxExpression{Span: span, Kind: "property", Text: method, Left: target}, Arguments: args}
	}
	bind := func(expr *frontend.SyntaxExpression) (string, error) {
		value, typ, err := lowerExpression(path, expr, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", err
		}
		env[value] = typ
		return value, nil
	}
	start, err := bind(expression.Arguments[0])
	if err != nil {
		return "", "", err
	}
	deleteCount, err := bind(expression.Arguments[1])
	if err != nil {
		return "", "", err
	}
	env[receiver] = receiverType
	zero := &frontend.SyntaxExpression{Span: span, Kind: "number", Text: "0", InferredType: "number"}
	head, err := bind(call(ident(receiver), "slice", zero, ident(start)))
	if err != nil {
		return "", "", err
	}
	spliced, err := bind(call(ident(receiver), "toSpliced", ident(start), ident(deleteCount)))
	if err != nil {
		return "", "", err
	}
	headLength := &frontend.SyntaxExpression{Span: span, Kind: "property", Text: "length", Left: ident(head), InferredType: "number"}
	rest, err := bind(call(ident(spliced), "slice", headLength))
	if err != nil {
		return "", "", err
	}
	items := &frontend.SyntaxExpression{Span: span, Kind: "array", Arguments: expression.Arguments[2:], InferredType: string(receiverType)}
	return lowerExpression(path, call(call(ident(head), "concat", items), "concat", ident(rest)), result, function, env, counter, shapes, signatures)
}
