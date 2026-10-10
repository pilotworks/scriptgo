package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// regExpMatchPropertiesShape holds the named properties of a RegExp match
// array (RegExpMatchArray / RegExpExecArray). The match is a string[] whose
// runtime header carries this object; see scriptgo_regex_match_properties.
var regExpMatchPropertiesShape = ir.ObjectShape{Name: "RegExpMatchProperties", Fields: []ir.Field{
	{Name: "index", Type: ir.TypeNumber},
	{Name: "input", Type: ir.TypeString},
	{Name: "groups", Type: ir.TypeObject},
}}

// isRegExpMatchType reports whether a checked type is a RegExp match array.
func isRegExpMatchType(typ string) bool {
	for _, part := range splitTopLevelUnion(typ) {
		switch strings.TrimSpace(part) {
		case "RegExpMatchArray", "RegExpExecArray":
			return true
		}
	}
	return false
}

// tryLowerRegExpMatchProperty lowers index, input and groups of a match
// array to reads of its properties object.
func tryLowerRegExpMatchProperty(path string, expression *frontend.SyntaxExpression, result *string, function *ir.Function, counter *int, object string) (string, ir.Type, bool, error) {
	if expression.Left == nil || !isRegExpMatchType(expression.Left.InferredType) {
		return "", "", false, nil
	}
	for index, field := range regExpMatchPropertiesShape.Fields {
		if field.Name != expression.Text {
			continue
		}
		span := toIRSpan(path, expression.Span)
		properties := nextTemp(counter)
		if *result == "" {
			*result = nextTemp(counter)
		}
		function.Body = append(function.Body,
			ir.Instruction{Op: ir.OpCall, Type: ir.Type("object:" + regExpMatchPropertiesShape.Name), Result: properties, Callee: "__regex.matchProperties", Args: []string{object}, Span: span},
			ir.Instruction{Op: ir.OpFieldGet, Type: field.Type, Result: *result, Callee: regExpMatchPropertiesShape.Name, Field: field.Name, FieldIndex: index, Args: []string{properties}, Span: span},
		)
		return *result, field.Type, true, nil
	}
	return "", "", false, nil
}
