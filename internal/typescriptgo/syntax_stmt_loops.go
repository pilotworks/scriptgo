package typescriptgo

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

func syntaxForInStatement(node *ast.Node, chk *checker.Checker, span SourceSpan) (SyntaxStatement, bool) {
	forIn := node.AsForInOrOfStatement()
	var varName, varType, varInferredType string
	if forIn.Initializer != nil && forIn.Initializer.Kind == ast.KindVariableDeclarationList {
		decls := forIn.Initializer.AsVariableDeclarationList().Declarations.Nodes
		if len(decls) == 1 {
			nameNode := decls[0].Name()
			if nameNode.Kind == ast.KindIdentifier {
				varName = nameNode.Text()
			}
			varType = syntaxType(decls[0].Type())
			varInferredType = resolveInferredType(chk, decls[0].Name())
			if varType == "" {
				varType = varInferredType
			}
		}
	}
	return SyntaxStatement{
		Span:         span,
		Kind:         "forin",
		Name:         varName,
		Type:         varType,
		InferredType: varInferredType,
		Expression:   syntaxExpression(forIn.Expression, chk),
		Body:         syntaxBlockStatements(forIn.Statement, chk),
	}, true
}

func syntaxForOfStatement(node *ast.Node, chk *checker.Checker, span SourceSpan) (SyntaxStatement, bool) {
	forOf := node.AsForInOrOfStatement()
	kind := "forof"
	if forOf.AwaitModifier != nil {
		kind = "forawaitof"
	}
	var varName, varType, varInferredType string
	var bindingStmts []SyntaxStatement
	if forOf.Initializer != nil && forOf.Initializer.Kind == ast.KindVariableDeclarationList {
		decls := forOf.Initializer.AsVariableDeclarationList().Declarations.Nodes
		if len(decls) == 1 {
			nameNode := decls[0].Name()
			if nameNode.Kind == ast.KindIdentifier {
				varName = nameNode.Text()
				varType = syntaxType(decls[0].Type())
				varInferredType = resolveInferredType(chk, decls[0].Name())
				if varType == "" {
					varType = varInferredType
				}
			} else if nameNode.Kind == ast.KindObjectBindingPattern || nameNode.Kind == ast.KindArrayBindingPattern {
				tempItemVar := fmt.Sprintf("__forof_destruct_%d", nameNode.Pos())
				varName = tempItemVar
				varInferredType = resolveInferredType(chk, nameNode)
				varType = varInferredType
				c := 0
				bindingStmts = flattenDestructuring(nameNode, &SyntaxExpression{
					Span: sourceSpan(nameNode),
					Kind: "identifier",
					Text: tempItemVar,
				}, chk, &c)
			}
		}
	}
	bodyStmts := syntaxBlockStatements(forOf.Statement, chk)
	if len(bindingStmts) > 0 {
		bodyStmts = append(bindingStmts, bodyStmts...)
	}
	return SyntaxStatement{
		Span:         span,
		Kind:         kind,
		Name:         varName,
		Type:         varType,
		InferredType: varInferredType,
		Expression:   syntaxExpression(forOf.Expression, chk),
		Body:         bodyStmts,
	}, true
}

func syntaxForStatement(node *ast.Node, chk *checker.Checker, span SourceSpan) (SyntaxStatement, bool) {
	forNode := node.AsForStatement()
	var bodyStatements []SyntaxStatement
	bodyStatements = append(bodyStatements, syntaxBlockStatements(forNode.Statement, chk)...)
	var stepStatements []SyntaxStatement
	if forNode.Incrementor != nil {
		incExpr := syntaxExpression(forNode.Incrementor, chk)
		for _, sub := range flattenCommaExpressions(incExpr) {
			if sub != nil && sub.Kind == "binary" && isAssignmentOperator(sub.Operator) {
				if sub.Left != nil && sub.Left.Kind == "identifier" {
					valExpr, _ := desugarAssignment(sub)
					stepStatements = append(stepStatements, SyntaxStatement{
						Span:       sub.Span,
						Kind:       "assign",
						Name:       sub.Left.Text,
						Expression: valExpr,
					})
					continue
				}
			}
			stepStatements = append(stepStatements, SyntaxStatement{
				Span:       sub.Span,
				Kind:       "expression",
				Expression: sub,
			})
		}
	}
	whileStmt := SyntaxStatement{
		Span:       span,
		Kind:       "while",
		Expression: syntaxExpression(forNode.Condition, chk),
		Body:       bodyStatements,
		Step:       stepStatements,
	}
	if forNode.Initializer != nil {
		if forNode.Initializer.Kind == ast.KindVariableDeclarationList {
			decls := forNode.Initializer.AsVariableDeclarationList().Declarations.Nodes
			initStmt, ok := syntaxVariableDeclarations(decls, sourceSpan(forNode.Initializer), chk, "var")
			if ok {
				var initStmts []SyntaxStatement
				if initStmt.Kind == "block" {
					initStmts = append(initStmts, initStmt.Body...)
				} else {
					initStmts = append(initStmts, initStmt)
				}
				initStmts = append(initStmts, whileStmt)
				return SyntaxStatement{
					Span: span,
					Kind: "block",
					Body: initStmts,
				}, true
			}
		} else {
			initExpr := syntaxExpression(forNode.Initializer, chk)
			var initStmts []SyntaxStatement
			for _, sub := range flattenCommaExpressions(initExpr) {
				if sub != nil && sub.Kind == "binary" && isAssignmentOperator(sub.Operator) {
					if sub.Left != nil && sub.Left.Kind == "identifier" {
						valExpr, _ := desugarAssignment(sub)
						initStmts = append(initStmts, SyntaxStatement{
							Span:       sub.Span,
							Kind:       "assign",
							Name:       sub.Left.Text,
							Expression: valExpr,
						})
						continue
					}
				}
				initStmts = append(initStmts, SyntaxStatement{
					Span:       sub.Span,
					Kind:       "expression",
					Expression: sub,
				})
			}
			initStmts = append(initStmts, whileStmt)
			return SyntaxStatement{
				Span: span,
				Kind: "block",
				Body: initStmts,
			}, true
		}
	}
	return whileStmt, true
}
