package lowering

import (
	"strings"
	"unicode"
	"unicode/utf8"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
)

// isAllowedPropertyReceiver reports whether a given syntax kind is a valid
// receiver for property access in the native Static tier.
func isAllowedPropertyReceiver(kind string) bool {
	switch kind {
	case "identifier", "string", "number", "bool", "call", "optional_call",
		"property", "optional_property", "index", "optional_index",
		"object_literal", "array", "as", "non_null", "binary",
		"conditional", "paren", "cast", "type_assertion", "new",
		"template", "tagged_template", "await":
		return true
	default:
		return false
	}
}

// isAllowedUnionBinaryOp reports whether a binary operator is permitted on union operands
// (e.g. equality checks, narrowing with instanceof/in, logical operators).
func isAllowedUnionBinaryOp(op string) bool {
	switch op {
	case "instanceof", "in", "==", "!=", "===", "!==", "&&", "||", "??":
		return true
	default:
		return false
	}
}

// isObjectBuiltinCall reports whether an expression is a recognized Object/Reflect builtin call.
func isObjectBuiltinCall(expr *typescriptgo.SyntaxExpression) bool {
	if expr == nil || expr.Kind != "call" {
		return false
	}
	name := ""
	if expr.Left != nil {
		if expr.Left.Kind == "identifier" && expr.Text != "" {
			name = expr.Left.Text + "." + expr.Text
		} else if (expr.Left.Kind == "property" || expr.Left.Kind == "optional_property") && expr.Left.Left != nil && expr.Left.Left.Kind == "identifier" {
			name = expr.Left.Left.Text + "." + expr.Left.Text
		}
	}
	if strings.HasPrefix(name, "Reflect.") {
		return true
	}
	return name == "Object.values" || name == "Object.entries" || name == "Object.assign" || name == "Object.keys"
}

// hasStringConcatUnion reports whether an expression is a string concatenation involving an allowed union.
func hasStringConcatUnion(expression *typescriptgo.SyntaxExpression) bool {
	if expression == nil || expression.Operator != "+" {
		return false
	}
	for _, operand := range []*typescriptgo.SyntaxExpression{expression.Left, expression.Right} {
		if operand == nil || !isHeterogeneousUnion(operand.InferredType) {
			continue
		}
		for _, member := range splitTopLevelUnion(operand.InferredType) {
			if toPrimitiveCategory(member) == "string" {
				return true
			}
		}
	}
	return false
}

// formatSubsetMessage formats an unsupported feature diagnostic into standard ScriptGo prose.
func formatSubsetMessage(code SubsetCode, feature string) string {
	feature = strings.TrimSpace(feature)
	if feature == "" {
		return "unsupported feature in native subset"
	}
	if strings.Contains(feature, "in native subset") || strings.Contains(feature, "native subset") {
		return feature
	}
	if strings.HasSuffix(feature, ".") {
		return strings.TrimSuffix(feature, ".") + " in native subset."
	}
	if strings.ToLower(feature) == "any type" {
		return "The any type is not supported in native subset."
	}
	r, size := utf8.DecodeRuneInString(feature)
	capitalized := string(unicode.ToUpper(r)) + feature[size:]
	if strings.Contains(strings.ToLower(feature), "not supported") {
		return capitalized + " in native subset."
	}
	return capitalized + " is not supported in native subset."
}
