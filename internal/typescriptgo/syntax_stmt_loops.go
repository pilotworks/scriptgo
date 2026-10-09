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
	body := syntaxBlockStatements(forIn.Statement, chk)
	if forIn.Initializer != nil && forIn.Initializer.Kind != ast.KindVariableDeclarationList {
		varName = fmt.Sprintf("__forin_target_%d", forIn.Initializer.Pos())
		varType, varInferredType = "string", "string"
		body = append(forTargetAssignment(forIn.Initializer, varName, varType, chk), body...)
	}
	return SyntaxStatement{
		Span:         span,
		Kind:         "forin",
		Name:         varName,
		Type:         varType,
		InferredType: varInferredType,
		Expression:   syntaxExpression(forIn.Expression, chk),
		Body:         body,
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
	if forOf.Initializer != nil && forOf.Initializer.Kind != ast.KindVariableDeclarationList {
		varName = fmt.Sprintf("__forof_target_%d", forOf.Initializer.Pos())
		varInferredType = resolveIteratedElementType(chk, forOf.Expression)
		if varInferredType == "" {
			varInferredType = resolveInferredType(chk, forOf.Initializer)
		}
		varType = varInferredType
		bindingStmts = forTargetAssignment(forOf.Initializer, varName, varType, chk)
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

// forTargetAssignment assigns the loop's per-iteration value (held in temp)
// to a for-in/for-of target that is an expression rather than a declaration:
// an identifier or member access (`for (obj.key in o)`) or an assignment
// pattern (`for ([a, b] of pairs)`).
func forTargetAssignment(target *ast.Node, temp, tempType string, chk *checker.Checker) []SyntaxStatement {
	value := &SyntaxExpression{Span: sourceSpan(target), Kind: "identifier", Text: temp, InferredType: tempType}
	if target.Kind == ast.KindArrayLiteralExpression || target.Kind == ast.KindObjectLiteralExpression {
		c := 0
		return flattenDestructuringAssignment(target, value, chk, &c)
	}
	left := syntaxExpression(target, chk)
	if stmt, ok := assignmentStatement(sourceSpan(target), left, value); ok {
		return []SyntaxStatement{stmt}
	}
	// Not an assignable target (e.g. a call): keep it visible to the subset
	// gate instead of producing an iteration variable with no name.
	return []SyntaxStatement{{
		Span: sourceSpan(target),
		Kind: "unsupported",
		Type: "for-in/for-of target that is not assignable",
	}}
}

// resolveIteratedElementType returns the element type of an array being
// iterated, so an assignment-pattern target is destructured from the value's
// own type rather than the type TypeScript gives the pattern (a tuple).
func resolveIteratedElementType(chk *checker.Checker, iterable *ast.Node) string {
	if chk == nil || iterable == nil {
		return ""
	}
	t := chk.GetTypeAtLocation(iterable)
	if t == nil || !chk.IsArrayType(t) {
		return ""
	}
	element := chk.GetElementTypeOfArrayType(t)
	if element == nil {
		return ""
	}
	return normalizeInferredType(chk.TypeToString(element))
}
