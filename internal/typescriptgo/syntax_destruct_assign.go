package typescriptgo

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

func flattenDestructuringAssignment(leftNode *ast.Node, initExpr *SyntaxExpression, chk *checker.Checker, counter *int) []SyntaxStatement {
	if leftNode == nil {
		return nil
	}
	var stmts []SyntaxStatement
	*counter++
	tmpVar := fmt.Sprintf("__destruct_assign_%d_%d", leftNode.Pos(), *counter)
	inferred := ""
	if initExpr != nil && initExpr.InferredType != "" && initExpr.InferredType != "void" && initExpr.InferredType != "undefined" {
		inferred = initExpr.InferredType
	} else {
		inferred = resolveInferredType(chk, leftNode)
	}
	stmts = append(stmts, SyntaxStatement{
		Span:         sourceSpan(leftNode),
		Kind:         "variable",
		Name:         tmpVar,
		Type:         inferred,
		InferredType: inferred,
		Expression:   initExpr,
	})

	if leftNode.Kind == ast.KindArrayLiteralExpression {
		arrLit := leftNode.AsArrayLiteralExpression()
		if arrLit != nil && arrLit.Elements != nil {
			for idx, elem := range arrLit.Elements.Nodes {
				if elem == nil || elem.Kind == ast.KindOmittedExpression {
					continue
				}
				if elem.Kind == ast.KindIdentifier {
					stmts = append(stmts, SyntaxStatement{
						Span: sourceSpan(elem),
						Kind: "assign",
						Name: elem.Text(),
						Expression: &SyntaxExpression{
							Span:  sourceSpan(elem),
							Kind:  "index",
							Left:  &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
							Right: &SyntaxExpression{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx)},
						},
					})
				} else if elem.Kind == ast.KindArrayLiteralExpression || elem.Kind == ast.KindObjectLiteralExpression {
					itemExpr := &SyntaxExpression{
						Span:  sourceSpan(elem),
						Kind:  "index",
						Left:  &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
						Right: &SyntaxExpression{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx)},
					}
					nested := flattenDestructuringAssignment(elem, itemExpr, chk, counter)
					stmts = append(stmts, nested...)
				}
			}
		}
	} else if leftNode.Kind == ast.KindObjectLiteralExpression {
		objLit := leftNode.AsObjectLiteralExpression()
		if objLit != nil && objLit.Properties != nil {
			for _, prop := range objLit.Properties.Nodes {
				if prop.Kind == ast.KindShorthandPropertyAssignment {
					propName := syntaxMemberName(prop.Name())
					stmts = append(stmts, SyntaxStatement{
						Span: sourceSpan(prop),
						Kind: "assign",
						Name: propName,
						Expression: &SyntaxExpression{
							Span: sourceSpan(prop),
							Kind: "property",
							Left: &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
							Text: propName,
						},
					})
				} else if prop.Kind == ast.KindPropertyAssignment {
					pAssign := prop.AsPropertyAssignment()
					propName := syntaxMemberName(pAssign.Name())
					if pAssign.Initializer != nil {
						if pAssign.Initializer.Kind == ast.KindIdentifier {
							stmts = append(stmts, SyntaxStatement{
								Span: sourceSpan(prop),
								Kind: "assign",
								Name: pAssign.Initializer.Text(),
								Expression: &SyntaxExpression{
									Span: sourceSpan(prop),
									Kind: "property",
									Left: &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
									Text: propName,
								},
							})
						} else if pAssign.Initializer.Kind == ast.KindBinaryExpression {
							bin := pAssign.Initializer.AsBinaryExpression()
							if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
								targetName := bin.Left.Text()
								defaultVal := syntaxExpression(bin.Right, chk)
								propExpr := &SyntaxExpression{
									Span: sourceSpan(prop),
									Kind: "property",
									Left: &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
									Text: propName,
								}
								stmts = append(stmts, SyntaxStatement{
									Span: sourceSpan(prop),
									Kind: "assign",
									Name: targetName,
									Expression: &SyntaxExpression{
										Span:     sourceSpan(prop),
										Kind:     "binary",
										Operator: "??",
										Left:     propExpr,
										Right:    defaultVal,
									},
								})
							}
						}
					}
				} else if prop.Kind == ast.KindBinaryExpression {
					bin := prop.AsBinaryExpression()
					if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
						propName := bin.Left.Text()
						defaultVal := syntaxExpression(bin.Right, chk)
						propExpr := &SyntaxExpression{
							Span: sourceSpan(prop),
							Kind: "property",
							Left: &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
							Text: propName,
						}
						stmts = append(stmts, SyntaxStatement{
							Span: sourceSpan(prop),
							Kind: "assign",
							Name: propName,
							Expression: &SyntaxExpression{
								Span:     sourceSpan(prop),
								Kind:     "binary",
								Operator: "??",
								Left:     propExpr,
								Right:    defaultVal,
							},
						})
					}
				}
			}
		}
	}
	return stmts
}

func countTupleElements(tupleType string) int {
	if !strings.HasPrefix(tupleType, "[") || !strings.HasSuffix(tupleType, "]") {
		return -1
	}
	inner := strings.TrimSpace(tupleType[1 : len(tupleType)-1])
	if inner == "" {
		return 0
	}
	count := 1
	depth := 0
	for _, ch := range inner {
		switch ch {
		case '[', '{', '(':
			depth++
		case ']', '}', ')':
			depth--
		case ',':
			if depth == 0 {
				count++
			}
		}
	}
	return count
}
