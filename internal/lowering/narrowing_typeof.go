package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func narrowPropertyPathType(propertyPath []string, expression *frontend.SyntaxExpression, baseEnv map[string]ir.Type, shapes map[string]ir.ObjectShape, nullishKind string) ir.Type {
	if len(propertyPath) < 2 {
		return ""
	}
	current, ok := baseEnv[propertyPath[0]]
	if !ok {
		return nonNullishIRType(expression.InferredType)
	}
	for _, property := range propertyPath[1:] {
		fields, found := resolveShapeFields(string(current), shapes)
		if !found {
			return nonNullishIRType(expression.InferredType)
		}
		var field *ir.Field
		for index := range fields {
			if fields[index].Name == property {
				field = &fields[index]
				break
			}
		}
		if field == nil {
			return nonNullishIRType(expression.InferredType)
		}
		current = field.Type
	}
	if nullishKind != "" && current != ir.TypeUnknown && current != ir.TypeVoid && current != ir.TypePointer {
		return current
	}
	return nonNullishIRType(expression.InferredType)
}

func applyTypeofNarrowing(expr *frontend.SyntaxExpression, thenEnv, elseEnv, baseEnv map[string]ir.Type) bool {
	if expr == nil || expr.Kind != "binary" {
		return false
	}
	var operand, literal *frontend.SyntaxExpression
	if expr.Left != nil && expr.Left.Kind == "typeof" && expr.Left.Left != nil && expr.Left.Left.Kind == "identifier" && expr.Right != nil {
		operand, literal = expr.Left.Left, expr.Right
	} else if expr.Right != nil && expr.Right.Kind == "typeof" && expr.Right.Left != nil && expr.Right.Left.Kind == "identifier" && expr.Left != nil {
		operand, literal = expr.Right.Left, expr.Left
	} else {
		return false
	}
	if literal.Kind != "string" && literal.Kind != "literal" {
		return false
	}
	target := strings.Trim(literal.Text, "\"'`")
	if target != "number" && target != "string" && target != "boolean" && target != "bigint" && target != "symbol" && target != "function" && target != "object" && target != "undefined" {
		return false
	}
	name := operand.Text
	decl := string(baseEnv["__decl_str."+name])
	if decl == "" {
		decl = string(baseEnv[name])
	}
	decl = expandAliasForNarrowing(decl, nil)
	matched := typeofTargetIRType(target)
	if matched == "" {
		return false
	}
	keepMatch := expr.Operator == "==" || expr.Operator == "==="
	if keepMatch {
		thenEnv[name] = matched
		if remainder := excludeTypeofFromUnion(decl, target); remainder != "" {
			elseEnv[name] = remainder
		}
	} else {
		if remainder := excludeTypeofFromUnion(decl, target); remainder != "" {
			thenEnv[name] = remainder
		}
		elseEnv[name] = matched
	}
	return true
}

func typeofTargetIRType(target string) ir.Type {
	switch target {
	case "number":
		return ir.TypeNumber
	case "string":
		return ir.TypeString
	case "boolean":
		return ir.TypeBool
	case "bigint":
		return ir.TypeBigInt
	case "symbol":
		return ir.TypeSymbol
	case "function":
		return ir.TypeClosure
	case "object":
		return ir.TypeObject
	case "undefined":
		return ir.TypeUnknown
	default:
		return ""
	}
}

func expandAliasForNarrowing(value string, seen map[string]bool) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if seen == nil {
		seen = map[string]bool{}
	}
	if seen[value] {
		return value
	}
	if alias, ok := typeAliasesIndex[value]; ok && alias != value {
		seen[value] = true
		return expandAliasForNarrowing(alias, seen)
	}
	return value
}

func excludeTypeofFromUnion(decl, target string) ir.Type {
	decl = expandAliasForNarrowing(decl, nil)
	parts := splitTopLevelUnion(decl)
	if len(parts) <= 1 {
		return ""
	}
	remaining := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || typeMatchesTypeof(part, target) {
			continue
		}
		remaining = append(remaining, part)
	}
	if len(remaining) == 0 {
		return ir.TypeVoid
	}
	if len(remaining) == 1 {
		return toIRType(remaining[0])
	}
	return toIRType(strings.Join(remaining, " | "))
}

func typeMatchesTypeof(typeName, target string) bool {
	typeName = strings.TrimSpace(typeName)
	if target == "string" {
		return typeName == "string" || (strings.HasPrefix(typeName, "\"") && strings.HasSuffix(typeName, "\"")) || (strings.HasPrefix(typeName, "'") && strings.HasSuffix(typeName, "'"))
	}
	if target == "boolean" {
		return typeName == "boolean" || typeName == "bool" || typeName == "true" || typeName == "false"
	}
	if target == "object" {
		return typeName == "object" || typeName == "null" || strings.HasPrefix(typeName, "object:") || (!strings.ContainsAny(typeName, "|&") && toIRType(typeName) != ir.TypeString && toIRType(typeName) != ir.TypeNumber && toIRType(typeName) != ir.TypeBool && toIRType(typeName) != ir.TypeBigInt && toIRType(typeName) != ir.TypeSymbol && toIRType(typeName) != ir.TypeClosure)
	}
	if target == "function" {
		return strings.Contains(typeName, "=>") || toIRType(typeName) == ir.TypeClosure
	}
	return typeName == target
}

func narrowUnionTypeString(declStr string, excludeKind string) ir.Type {
	parts := splitTopLevelUnion(declStr)
	var remaining []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		switch excludeKind {
		case "undefined":
			if trimmed == "undefined" || trimmed == "void" {
				continue
			}
		case "null":
			if trimmed == "null" {
				continue
			}
		case "nullish":
			if trimmed == "undefined" || trimmed == "void" || trimmed == "null" {
				continue
			}
		}
		remaining = append(remaining, trimmed)
	}
	if len(remaining) == 0 {
		return ir.TypeVoid
	}
	if len(remaining) == 1 {
		return toIRType(remaining[0])
	}
	return toIRType(strings.Join(remaining, " | "))
}
