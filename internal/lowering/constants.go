package lowering

import (
	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// extractPropertyPath returns the dotted chain of identifiers (e.g. ["os", "constants", "signals", "SIGINT"]).
func extractPropertyPath(expr *typescriptgo.SyntaxExpression) []string {
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

// resolveASTConstantFromExpr evaluates literal expressions or traverses object literals along the remaining path.
func resolveASTConstantFromExpr(expr *typescriptgo.SyntaxExpression, remaining []string) (string, ir.Type, bool) {
	if expr == nil {
		return "", "", false
	}
	if len(remaining) == 0 {
		switch expr.Kind {
		case "number", "numeric_literal":
			return expr.Text, ir.TypeNumber, true
		case "string", "string_literal":
			return expr.Text, ir.TypeString, true
		case "boolean", "boolean_literal":
			return expr.Text, ir.TypeBool, true
		case "property", "optional_property", "identifier":
			path := extractPropertyPath(expr)
			if len(path) > 0 {
				return resolveASTConstantPath(path)
			}
		}
		return "", "", false
	}

	if expr.Kind == "object_literal" || expr.Kind == "object" {
		targetKey := remaining[0]
		for _, prop := range expr.Arguments {
			if prop != nil && prop.Text == targetKey && prop.Left != nil {
				return resolveASTConstantFromExpr(prop.Left, remaining[1:])
			}
		}
	}
	return "", "", false
}

// resolveASTConstantPath resolves a top-level constant or module constant path from the AST.
func resolveASTConstantPath(path []string) (string, ir.Type, bool) {
	if len(path) == 0 {
		return "", "", false
	}
	// 1. Direct match on top-level variable
	if topVar, ok := topLevelVars[path[0]]; ok && topVar.Expression != nil {
		if val, typ, found := resolveASTConstantFromExpr(topVar.Expression, path[1:]); found {
			return val, typ, true
		}
	}
	// 2. Module namespace prefix match (e.g. ["fs", "constants", "F_OK"] or ["path", "sep"])
	if len(path) >= 2 {
		if topVar, ok := topLevelVars[path[1]]; ok && topVar.Expression != nil {
			if val, typ, found := resolveASTConstantFromExpr(topVar.Expression, path[2:]); found {
				return val, typ, true
			}
		}
	}
	return "", "", false
}
