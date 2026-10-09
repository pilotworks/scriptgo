package typescriptgo

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// thisBinding classifies a `this` keyword by its this container; see
// SyntaxExpression.ThisBinding.
func thisBinding(node *ast.Node) string {
	container := ast.GetThisContainer(node, false, false)
	switch container.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		return "caller"
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		if container.Parent != nil && container.Parent.Kind == ast.KindObjectLiteralExpression {
			return "object"
		}
	}
	return ""
}
