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
		if stmt, ok := assignmentStatement(span, expr.Left, valExpr); ok {
			return stmt, true
		}
	}
	return SyntaxStatement{Span: span, Kind: "expression", Expression: expr}, true
}

// assignmentStatement normalizes `target = value` for an identifier, index,
// or property target into the assign/index_set/field_set statement forms.
func assignmentStatement(span SourceSpan, target, value *SyntaxExpression) (SyntaxStatement, bool) {
	switch target.Kind {
	case "identifier":
		return SyntaxStatement{Span: span, Kind: "assign", Name: target.Text, Expression: value}, true
	case "index":
		return SyntaxStatement{Span: span, Kind: "index_set", Left: target.Left, Right: target.Right, Expression: value}, true
	case "property":
		return SyntaxStatement{Span: span, Kind: "field_set", Left: target.Left, Name: target.Text, Expression: value}, true
	}
	return SyntaxStatement{}, false
}
