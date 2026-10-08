package typescriptgo

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

func syntaxStatement(node *ast.Node, chk *checker.Checker) (SyntaxStatement, bool) {
	span := sourceSpan(node)
	switch node.Kind {
	case ast.KindVariableStatement:
		varStmt := node.AsVariableStatement()
		declList := varStmt.DeclarationList.AsVariableDeclarationList()
		declarations := declList.Declarations.Nodes
		var varKind string
		if (declList.Flags & ast.NodeFlagsBlockScoped) == ast.NodeFlagsAwaitUsing {
			varKind = "await_using"
		} else if (declList.Flags & ast.NodeFlagsBlockScoped) == ast.NodeFlagsUsing {
			varKind = "using"
		} else if (declList.Flags & ast.NodeFlagsConst) != 0 {
			varKind = "const"
		} else if (declList.Flags & ast.NodeFlagsLet) != 0 {
			varKind = "let"
		} else {
			varKind = "var"
		}
		return syntaxVariableDeclarations(declarations, span, chk, varKind)
	case ast.KindFunctionDeclaration:
		return syntaxFunctionDeclaration(node, chk, span)

	case ast.KindClassDeclaration:
		return syntaxClassDeclaration(node, span, chk)
	case ast.KindReturnStatement:
		return SyntaxStatement{Span: span, Kind: "return", Expression: syntaxExpression(node.Expression(), chk)}, true
	case ast.KindThrowStatement:
		return SyntaxStatement{Span: span, Kind: "throw", Expression: syntaxExpression(node.Expression(), chk)}, true
	case ast.KindTryStatement:
		tryNode := node.AsTryStatement()
		res := SyntaxStatement{
			Span: span,
			Kind: "try",
			Body: syntaxBlockStatements(tryNode.TryBlock, chk),
		}
		if tryNode.CatchClause != nil {
			catchClause := tryNode.CatchClause.AsCatchClause()
			if catchClause.VariableDeclaration != nil {
				res.CatchVar = catchClause.VariableDeclaration.Name().Text()
				if catchClause.VariableDeclaration.Type() != nil {
					res.CatchVarType = syntaxType(catchClause.VariableDeclaration.Type())
				}
				res.CatchVarSpan = sourceSpan(catchClause.VariableDeclaration.Name())
			}
			res.Catch = syntaxBlockStatements(catchClause.Block, chk)
		}
		if tryNode.FinallyBlock != nil {
			res.Finally = syntaxBlockStatements(tryNode.FinallyBlock, chk)
		}
		return res, true
	case ast.KindBreakStatement:
		breakStmt := node.AsBreakStatement()
		label := ""
		if breakStmt != nil && breakStmt.Label != nil {
			label = breakStmt.Label.Text()
		}
		return SyntaxStatement{Span: span, Kind: "break", Name: label}, true
	case ast.KindDebuggerStatement:
		return SyntaxStatement{Span: span, Kind: "debugger"}, true
	case ast.KindContinueStatement:
		contStmt := node.AsContinueStatement()
		label := ""
		if contStmt != nil && contStmt.Label != nil {
			label = contStmt.Label.Text()
		}
		return SyntaxStatement{Span: span, Kind: "continue", Name: label}, true
	case ast.KindLabeledStatement:
		labeledNode := node.AsLabeledStatement()
		labelName := ""
		if labeledNode.Label != nil {
			labelName = labeledNode.Label.Text()
		}
		inner, ok := syntaxStatement(labeledNode.Statement, chk)
		if ok {
			inner.Label = labelName
			if inner.Kind == "block" && len(inner.Body) > 0 {
				for i := range inner.Body {
					if inner.Body[i].Kind == "while" || inner.Body[i].Kind == "dowhile" || inner.Body[i].Kind == "forof" || inner.Body[i].Kind == "forin" || inner.Body[i].Kind == "forawaitof" {
						inner.Body[i].Label = labelName
					}
				}
			}
			return inner, true
		}
		return SyntaxStatement{Span: span, Kind: "label", Label: labelName}, true
	case ast.KindIfStatement:
		ifNode := node.AsIfStatement()
		result := SyntaxStatement{Span: span, Kind: "if", Expression: syntaxExpression(ifNode.Expression, chk)}
		result.Then = syntaxBlockStatements(ifNode.ThenStatement, chk)
		result.Else = syntaxBlockStatements(ifNode.ElseStatement, chk)
		return result, true
	case ast.KindWhileStatement:
		whileNode := node.AsWhileStatement()
		return SyntaxStatement{
			Span:       span,
			Kind:       "while",
			Expression: syntaxExpression(whileNode.Expression, chk),
			Body:       syntaxBlockStatements(whileNode.Statement, chk),
		}, true
	case ast.KindDoStatement:
		doNode := node.AsDoStatement()
		return SyntaxStatement{
			Span:       span,
			Kind:       "dowhile",
			Expression: syntaxExpression(doNode.Expression, chk),
			Body:       syntaxBlockStatements(doNode.Statement, chk),
		}, true
	case ast.KindForInStatement:
		return syntaxForInStatement(node, chk, span)

	case ast.KindForOfStatement:
		return syntaxForOfStatement(node, chk, span)

	case ast.KindSwitchStatement:
		switchNode := node.AsSwitchStatement()
		switchExpr := syntaxExpression(switchNode.Expression, chk)
		var cases []SyntaxSwitchCase
		for _, clause := range switchNode.CaseBlock.AsNode().AsCaseBlock().Clauses.Nodes {
			caseClause := clause.AsCaseOrDefaultClause()
			var cExpr *SyntaxExpression
			if clause.Kind == ast.KindCaseClause {
				cExpr = syntaxExpression(caseClause.Expression, chk)
			}
			var stmts []SyntaxStatement
			if caseClause.Statements != nil {
				for _, s := range caseClause.Statements.Nodes {
					stmts = append(stmts, syntaxBlockStatements(s, chk)...)
				}
			}
			cases = append(cases, SyntaxSwitchCase{
				Span:       sourceSpan(clause),
				Expression: cExpr,
				Statements: stmts,
			})
		}
		return SyntaxStatement{
			Span:       span,
			Kind:       "switch",
			Expression: switchExpr,
			Cases:      cases,
		}, true
	case ast.KindForStatement:
		return syntaxForStatement(node, chk, span)

	case ast.KindExpressionStatement:
		return syntaxExpressionStatement(node, chk, span)

	case ast.KindEnumDeclaration:
		return syntaxEnumDeclaration(node, chk, span)

	case ast.KindImportDeclaration:
		return syntaxImportDeclaration(node, span)

	case ast.KindExportDeclaration:
		return syntaxExportDeclaration(node, span)

	case ast.KindExportAssignment:
		return SyntaxStatement{Span: span, Kind: "module", Type: node.Kind.String()}, true
	case ast.KindModuleDeclaration:
		return syntaxModuleDeclaration(node, chk, span)

	case ast.KindInterfaceDeclaration:
		return syntaxInterfaceDeclaration(node, chk, span)

	case ast.KindTypeAliasDeclaration:
		return syntaxTypeAliasDeclaration(node, chk, span)

	case ast.KindBlock:
		return SyntaxStatement{
			Span: span,
			Kind: "block",
			Body: syntaxBlockStatements(node, chk),
		}, true
	case ast.KindEmptyStatement:
		return SyntaxStatement{}, false
	default:
		return SyntaxStatement{Span: span, Kind: "unsupported", Type: node.Kind.String()}, true
	}
}
