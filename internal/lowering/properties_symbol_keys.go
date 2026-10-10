package lowering

import (
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// propertyKeyValue converts a property key computed at run time to the
// string the object stores it under (ToPropertyKey): a symbol becomes its
// symbol key (scriptgo_symbol_property_key, which string-key enumeration
// skips), any other primitive its ToString.
func propertyKeyValue(path string, span frontend.SourceSpan, value string, typ ir.Type, function *ir.Function, counter *int) (string, ir.Type) {
	if typ != ir.TypeSymbol {
		return coercePrimitiveToString(path, span, value, typ, function, counter)
	}
	key := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: key, Callee: "__symbol.propertyKey", Args: []string{value}, Span: toIRSpan(path, span)})
	return key, ir.TypeString
}

func hasComputedProperties(expression *frontend.SyntaxExpression) bool {
	for _, member := range expression.Arguments {
		if member.Kind == "computed_property" {
			return true
		}
	}
	return false
}

// lowerObjectLiteralWithSymbolKeys lowers an object literal with members keyed
// by a symbol value ({ [k]: v }): the other members build the object's
// layout, then each symbol-keyed member is set under its symbol key, after
// the string keys, as JavaScript enumerates them.
func lowerObjectLiteralWithSymbolKeys(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	static := *expression
	static.Arguments = nil
	var computed []*frontend.SyntaxExpression
	for _, member := range expression.Arguments {
		if member.Kind == "computed_property" {
			computed = append(computed, member)
		} else {
			static.Arguments = append(static.Arguments, member)
		}
	}
	object, objectType, err := lowerObjectLiteralExpression(path, &static, result, function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	for _, member := range computed {
		key, keyType, err := lowerExpression(path, member.Right, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		key, _ = propertyKeyValue(path, member.Right.Span, key, keyType, function, counter)
		value, _, err := lowerExpression(path, member.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__object.set_prop", Args: []string{object, key, value}, Span: toIRSpan(path, member.Span)})
	}
	return object, objectType, nil
}
