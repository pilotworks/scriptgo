package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerNewMapFromEntries lowers `new Map(entries)` for an array of [key,
// value] pairs as an empty Map filled by
// `for (const entry of entries) map.set(entry[0], entry[1])`, so every entry
// goes through the Map operations typed for its key and value.
func lowerNewMapFromEntries(path string, expression *frontend.SyntaxExpression, entries string, entriesType ir.Type, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	span := expression.Span
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeMap, Result: result, Callee: "__map.new", Span: toIRSpan(path, span)})
	env[result] = ir.TypeMap
	env[entries] = entriesType

	keyType, valueType := mapTypeArguments(expression.InferredType)
	entry := "__map_entry_" + nextTemp(counter)
	element := func(index, typ string) *frontend.SyntaxExpression {
		return &frontend.SyntaxExpression{
			Span:         span,
			Kind:         "index",
			Left:         &frontend.SyntaxExpression{Span: span, Kind: "identifier", Text: entry},
			Right:        &frontend.SyntaxExpression{Span: span, Kind: "number", Text: index, InferredType: "number"},
			InferredType: typ,
		}
	}
	receiver := &frontend.SyntaxExpression{Span: span, Kind: "identifier", Text: result, InferredType: expression.InferredType}
	set := frontend.SyntaxStatement{
		Span: span,
		Kind: "expression",
		Expression: &frontend.SyntaxExpression{
			Span:         span,
			Kind:         "call",
			Left:         &frontend.SyntaxExpression{Span: span, Kind: "property", Text: "set", Left: receiver, InferredType: expression.InferredType},
			Arguments:    []*frontend.SyntaxExpression{element("0", keyType), element("1", valueType)},
			InferredType: expression.InferredType,
		},
	}
	err := lowerForOf(path, frontend.SyntaxStatement{
		Span:       span,
		Kind:       "forof",
		Name:       entry,
		Expression: &frontend.SyntaxExpression{Span: expression.Arguments[0].Span, Kind: "identifier", Text: entries, InferredType: expression.Arguments[0].InferredType},
		Body:       []frontend.SyntaxStatement{set},
	}, function, env, counter, shapes, signatures)
	return result, ir.TypeMap, err
}

// mapTypeArguments is K and V of a `Map<K, V>` type, or empty strings.
func mapTypeArguments(typ string) (string, string) {
	typ = strings.TrimSpace(typ)
	if !strings.HasPrefix(typ, "Map<") || !strings.HasSuffix(typ, ">") {
		return "", ""
	}
	args := splitTopLevel(typ[len("Map<") : len(typ)-1])
	if len(args) != 2 {
		return "", ""
	}
	return strings.TrimSpace(args[0]), strings.TrimSpace(args[1])
}
