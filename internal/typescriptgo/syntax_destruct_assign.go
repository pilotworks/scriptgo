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
				itemExpr := &SyntaxExpression{
					Span:  sourceSpan(elem),
					Kind:  "index",
					Left:  &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
					Right: &SyntaxExpression{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx)},
				}
				if elem.Kind == ast.KindSpreadElement {
					stmts = append(stmts, SyntaxStatement{Span: sourceSpan(elem), Kind: "unsupported", Type: "rest element in an array destructuring assignment"})
					continue
				}
				target := elem
				var value *SyntaxExpression = itemExpr
				if bin := assignmentWithDefault(elem); bin != nil {
					target = bin.Left
					value = &SyntaxExpression{Span: sourceSpan(elem), Kind: "binary", Operator: "??", Left: itemExpr, Right: syntaxExpression(bin.Right, chk)}
				}
				stmts = append(stmts, destructuringAssignmentTarget(target, value, chk, counter)...)
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
						propExpr := &SyntaxExpression{
							Span: sourceSpan(prop),
							Kind: "property",
							Left: &SyntaxExpression{Span: sourceSpan(leftNode), Kind: "identifier", Text: tmpVar},
							Text: propName,
						}
						target := pAssign.Initializer
						value := propExpr
						if bin := assignmentWithDefault(target); bin != nil {
							target = bin.Left
							value = &SyntaxExpression{
								Span:     sourceSpan(prop),
								Kind:     "binary",
								Operator: "??",
								Left:     propExpr,
								Right:    syntaxExpression(bin.Right, chk),
							}
						}
						stmts = append(stmts, destructuringAssignmentTarget(target, value, chk, counter)...)
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

// destructuringAssignmentTarget stores value into one assignment-pattern
// target: a nested pattern is flattened, an identifier or member expression
// is assigned, and anything else stays visible to the subset gate.
func destructuringAssignmentTarget(target *ast.Node, value *SyntaxExpression, chk *checker.Checker, counter *int) []SyntaxStatement {
	if target.Kind == ast.KindArrayLiteralExpression || target.Kind == ast.KindObjectLiteralExpression {
		return flattenDestructuringAssignment(target, value, chk, counter)
	}
	if stmt, ok := assignmentStatement(sourceSpan(target), syntaxExpression(target, chk), value); ok {
		return []SyntaxStatement{stmt}
	}
	return []SyntaxStatement{{
		Span: sourceSpan(target),
		Kind: "unsupported",
		Type: "destructuring assignment target that is not assignable",
	}}
}

// assignmentWithDefault returns `target = default` inside an assignment
// pattern, or nil for any other node.
func assignmentWithDefault(node *ast.Node) *ast.BinaryExpression {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return nil
	}
	bin := node.AsBinaryExpression()
	if bin == nil || bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindEqualsToken {
		return nil
	}
	return bin
}
