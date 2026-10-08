package lowering

import (
	"fmt"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerTaggedTemplate lowers tagged template expressions: tag`str1 ${expr} str2`
// Desugars into: tag(stringsArray, expr1, expr2, ...)
func lowerTaggedTemplate(
	path string,
	expression *frontend.SyntaxExpression,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, error) {
	if expression.Left == nil {
		return "", "", fmt.Errorf("tagged template missing tag function")
	}

	// Synthesize a call expression with the arguments
	callExpr := &frontend.SyntaxExpression{
		Span:          expression.Span,
		Kind:          "call",
		Left:          expression.Left,
		Arguments:     expression.Arguments,
		TypeArguments: expression.TypeArguments,
	}

	return lowerCallExpression(path, callExpr, result, function, env, counter, shapes, signatures)
}
