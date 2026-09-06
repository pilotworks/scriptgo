package lowering

import (
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
)

func isHeterogeneousUnion(typ string) bool {
	normalized := strings.ToLower(strings.TrimSpace(typ))
	if strings.HasSuffix(normalized, "[]") || strings.Contains(normalized, "generator") || strings.Contains(normalized, "iterator") || strings.Contains(normalized, "string | symbol") || strings.Contains(normalized, "symbol | string") || !strings.Contains(normalized, "|") {
		return false
	}
	parts := splitTopLevelUnion(normalized)
	if len(parts) <= 1 {
		if len(normalized) > 1 && normalized[0] == '(' && normalized[len(normalized)-1] == ')' {
			return isHeterogeneousUnion(strings.TrimSpace(normalized[1 : len(normalized)-1]))
		}
		return false
	}
	var nonNullish []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		trimmed = strings.TrimPrefix(trimmed, "(")
		trimmed = strings.TrimSuffix(trimmed, ")")
		if trimmed != "null" && trimmed != "undefined" && trimmed != "void" && trimmed != "" {
			nonNullish = append(nonNullish, trimmed)
		}
	}
	if len(nonNullish) <= 1 {
		return false
	}
	first := toPrimitiveCategory(nonNullish[0])
	for _, other := range nonNullish[1:] {
		if toPrimitiveCategory(other) != first {
			return true
		}
	}
	return false
}

func toPrimitiveCategory(t string) string {
	t = strings.TrimSpace(t)
	if (strings.HasPrefix(t, "\"") && strings.HasSuffix(t, "\"")) || (strings.HasPrefix(t, "'") && strings.HasSuffix(t, "'")) || t == "string" {
		return "string"
	}
	if _, err := strconv.ParseFloat(t, 64); err == nil || t == "number" {
		return "number"
	}
	if t == "boolean" || t == "bool" || t == "true" || t == "false" {
		return "bool"
	}
	if t == "bigint" {
		return "bigint"
	}
	if t == "symbol" {
		return "symbol"
	}
	if strings.HasSuffix(t, "[]") {
		return "array"
	}
	return "object"
}

func isUnknownType(typ string) bool {
	normalized := strings.ToLower(strings.TrimSpace(typ))
	return normalized == "unknown" || normalized == "unknownkeyword" || normalized == "kindunknownkeyword"
}

func isOrContainsAny(typ string) bool {
	typ = strings.ToLower(strings.TrimSpace(typ))
	if typ == "" {
		return false
	}
	if typ == "any" || typ == "anykeyword" || typ == "kindanykeyword" {
		return true
	}
	for i := 0; i < len(typ); {
		b := typ[i]
		if (b >= 'a' && b <= 'z') || b == '_' || b == '$' {
			start := i
			for i < len(typ) {
				ch := typ[i]
				if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '$' {
					i++
				} else {
					break
				}
			}
			word := typ[start:i]
			if word == "any" || word == "anykeyword" || word == "kindanykeyword" {
				return true
			}
		} else {
			i++
		}
	}
	return false
}

func validateStaticType(fileName string, span typescriptgo.SourceSpan, typ string) error {
	if isOrContainsAny(typ) {
		return subsetError(fileName, span, CodeAnyBoundary, "any type")
	}
	return nil
}

type Warning struct {
	FileName string
	Span     typescriptgo.SourceSpan
	Code     SubsetCode
	Message  string
}

func (w Warning) Format() string {
	return typescriptgo.Format(w.FileName, w.Span.Start, w.Span.Length, "warning", string(w.Code), w.Message, "")
}

var (
	warnings         []Warning
	WarnRuntimeCasts bool
)

func ClearDiagnostics() { warnings = nil }

func GetWarnings() []Warning {
	lowerMu.Lock()
	defer lowerMu.Unlock()
	return append([]Warning(nil), warnings...)
}

func recordWarning(fileName string, span typescriptgo.SourceSpan, code SubsetCode, message string) {
	warnings = append(warnings, Warning{FileName: fileName, Span: span, Code: code, Message: message})
}
