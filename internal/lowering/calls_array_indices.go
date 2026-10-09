package lowering

import (
	"strconv"

	"github.com/pilotworks/scriptgo/internal/frontend"
)

// staticIntegerArgument reads an integer literal, including a negated one.
func staticIntegerArgument(expression *frontend.SyntaxExpression) (int, bool) {
	if expression == nil {
		return 0, false
	}
	if expression.Kind == "unary" && expression.Operator == "-" {
		n, ok := staticIntegerArgument(expression.Left)
		return -n, ok
	}
	if expression.Kind != "number" {
		return 0, false
	}
	n, err := strconv.Atoi(expression.Text)
	return n, err == nil
}

// relativeTupleIndex resolves a relative index (negative counts from the
// end) and clamps it to [0, length].
func relativeTupleIndex(index, length int) int {
	if index < 0 {
		index += length
	}
	return max(0, min(index, length))
}
