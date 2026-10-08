package typescriptgo

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

func syntaxExpressionStatement(node *ast.Node, chk *checker.Checker, span SourceSpan) (SyntaxStatement, bool) {
	innerNode := node.Expression()
	for innerNode != nil && innerNode.Kind == ast.KindParenthesizedExpression {
		innerNode = innerNode.AsParenthesizedExpression().Expression
	}
	if innerNode != nil && innerNode.Kind == ast.KindBinaryExpression {
		bin := innerNode.AsBinaryExpression()
		if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
			leftNode := bin.Left
			for leftNode != nil && leftNode.Kind == ast.KindParenthesizedExpression {
				leftNode = leftNode.AsParenthesizedExpression().Expression
			}
			if leftNode != nil && (leftNode.Kind == ast.KindArrayLiteralExpression || leftNode.Kind == ast.KindObjectLiteralExpression) {
				c := 0
				stmts := flattenDestructuringAssignment(leftNode, syntaxExpression(bin.Right, chk), chk, &c)
				if len(stmts) == 1 {
					return stmts[0], true
				}
				return SyntaxStatement{
					Span: span,
					Kind: "block",
					Body: stmts,
				}, true
			}
		}
	}
	expr := syntaxExpression(node.Expression(), chk)
	if expr != nil && expr.Kind == "binary" && isAssignmentOperator(expr.Operator) && expr.Left != nil {
		valExpr, _ := desugarAssignment(expr)
		if expr.Left.Kind == "identifier" {
			return SyntaxStatement{
				Span:       span,
				Kind:       "assign",
				Name:       expr.Left.Text,
				Expression: valExpr,
			}, true
		}
		if expr.Left.Kind == "index" {
			return SyntaxStatement{
				Span:       span,
				Kind:       "index_set",
				Left:       expr.Left.Left,
				Right:      expr.Left.Right,
				Expression: valExpr,
			}, true
		}
		if expr.Left.Kind == "property" {
			return SyntaxStatement{
				Span:       span,
				Kind:       "field_set",
				Left:       expr.Left.Left,
				Name:       expr.Left.Text,
				Expression: valExpr,
			}, true
		}
	}
	return SyntaxStatement{Span: span, Kind: "expression", Expression: expr}, true
}
