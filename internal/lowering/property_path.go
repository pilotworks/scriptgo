package lowering

import (
	"github.com/pilotworks/scriptgo/internal/frontend"
)

// extractPropertyPath returns the dotted chain of identifiers (e.g. ["os", "constants", "signals", "SIGINT"]).
func extractPropertyPath(expr *frontend.SyntaxExpression) []string {
	if expr == nil {
		return nil
	}
	if expr.Kind == "identifier" {
		return []string{expr.Text}
	}
	if expr.Kind == "as" {
		return extractPropertyPath(expr.Left)
	}
	if expr.Kind == "property" || expr.Kind == "optional_property" {
		leftPath := extractPropertyPath(expr.Left)
		if len(leftPath) > 0 {
			return append(leftPath, expr.Text)
		}
	}
	return nil
}
