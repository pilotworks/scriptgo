package lowering

import (
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
)

// typedFilledArray rewrites `new Array(length).fill(value)` and
// `Array(length).fill(value)`, which TypeScript types as any[], into
// `new Array<T>(length).fill(value)` with T the widened type of value: every
// element of the result is that value. It returns nil for any other call.
func typedFilledArray(expression *frontend.SyntaxExpression) *frontend.SyntaxExpression {
	if expression == nil || expression.Kind != "call" || len(expression.Arguments) == 0 {
		return nil
	}
	fill := expression.Left
	if fill == nil || fill.Kind != "property" || fill.Text != "fill" || fill.Left == nil {
		return nil
	}
	construct := fill.Left
	if (construct.Kind != "new" && construct.Kind != "call") || callName(construct.Left) != "Array" ||
		len(construct.Arguments) != 1 || len(construct.TypeArguments) > 0 {
		return nil
	}
	element := widenedLiteralType(expression.Arguments[0].InferredType)
	if element == "" || element == "any" || element == "unknown" {
		return nil
	}
	arrayType := element + "[]"
	if strings.ContainsAny(element, "|&") {
		arrayType = "(" + element + ")[]"
	}
	typedConstruct := *construct
	typedConstruct.Kind = "new"
	typedConstruct.TypeArguments = []string{element}
	typedConstruct.InferredType = arrayType
	typedFill := *fill
	typedFill.Left = &typedConstruct
	typed := *expression
	typed.Left = &typedFill
	typed.InferredType = arrayType
	return &typed
}

// widenedLiteralType is the primitive type of a literal type ("0" is number,
// "\"x\"" is string, "true" is boolean, "1n" is bigint); other types are
// returned unchanged.
func widenedLiteralType(typ string) string {
	typ = strings.TrimSpace(typ)
	switch {
	case typ == "true" || typ == "false":
		return "boolean"
	case len(typ) >= 2 && (typ[0] == '"' || typ[0] == '\'' || typ[0] == '`'):
		return "string"
	case strings.HasSuffix(typ, "n") && isIntegerLiteral(strings.TrimSuffix(typ, "n")):
		return "bigint"
	}
	if _, err := strconv.ParseFloat(typ, 64); err == nil {
		return "number"
	}
	return typ
}

func isIntegerLiteral(text string) bool {
	_, err := strconv.ParseInt(strings.TrimPrefix(text, "-"), 10, 64)
	return err == nil
}
