package lowering

import (
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func callName(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil {
		return ""
	}
	if expression.Kind == "identifier" {
		return expression.Text
	}
	if expression.Kind == "as" {
		return callName(expression.Left)
	}
	if expression.Kind == "property" || expression.Kind == "optional_property" {
		leftName := callName(expression.Left)
		if leftName != "" {
			return leftName + "." + expression.Text
		}
	}
	return ""
}

func isStringMethod(name string) bool {
	switch name {
	case "split", "indexOf", "lastIndexOf", "slice", "startsWith", "endsWith", "trim", "trimStart", "trimEnd", "trimLeft", "trimRight", "replace", "replaceAll", "substring", "substr", "charAt", "at", "charCodeAt", "includes", "toLowerCase", "toUpperCase", "toLocaleLowerCase", "toLocaleUpperCase", "repeat", "padStart", "padEnd", "concat", "match", "matchAll", "search", "codePointAt", "isWellFormed", "toWellFormed", "localeCompare", "normalize", "valueOf", "toString", "anchor", "big", "blink", "bold", "fixed", "fontcolor", "fontsize", "italics", "link", "small", "strike", "sub", "sup":
		return true
	default:
		return false
	}
}

func isArrayMethod(name string) bool {
	switch name {
	case "push", "pop", "slice", "indexOf", "lastIndexOf", "includes", "join", "reverse", "concat", "shift", "unshift", "splice", "at", "map", "filter", "forEach", "reduce", "reduceRight", "find", "findLast", "some", "every", "findIndex", "findLastIndex", "fill", "toReversed", "toSorted", "toSpliced", "with", "sort", "copyWithin", "toString", "toLocaleString", "flat", "flatMap", "entries", "keys", "values":
		return true
	default:
		return false
	}
}

func stringMethod(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil || (expression.Kind != "property" && expression.Kind != "optional_property") || expression.Left == nil {
		return ""
	}
	if isStringMethod(expression.Text) {
		return expression.Text
	}
	return ""
}

func arrayMethod(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil || (expression.Kind != "property" && expression.Kind != "optional_property") || expression.Left == nil {
		return ""
	}
	if isArrayMethod(expression.Text) {
		return expression.Text
	}
	return ""
}

func resolveMapTypes(expr *typescriptgo.SyntaxExpression, env map[string]ir.Type) (string, string) {
	if expr == nil {
		return "", ""
	}
	target := expr
	if target.Kind == "property" && target.Left != nil && (target.Text == "get" || target.Text == "forEach" || target.Text == "set" || target.Text == "has" || target.Text == "delete") {
		target = target.Left
	}
	rawType := target.InferredType
	if target.Kind == "property" {
		propName := target.Text
		for _, meta := range classHierarchy {
			for _, f := range meta.Fields {
				if f.Name == propName && strings.HasPrefix(f.Type, "Map<") {
					rawType = f.Type
					break
				}
			}
			if rawType != "" {
				break
			}
		}
	} else if target.Kind == "identifier" {
		if t, ok := env[target.Text]; ok && strings.Contains(string(t), "<") {
			rawType = string(t)
		}
	}
	if strings.Contains(rawType, "<") && strings.HasSuffix(rawType, ">") {
		idx := strings.Index(rawType, "<")
		inner := rawType[idx+1 : len(rawType)-1]
		parts := splitTypeArguments(inner)
		if len(parts) >= 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}
	return "", ""
}
