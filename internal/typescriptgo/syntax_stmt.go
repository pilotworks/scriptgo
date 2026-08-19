package typescriptgo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/internal/ast"
)

func syntaxStatement(node *ast.Node) (SyntaxStatement, bool) {
	span := sourceSpan(node)
	switch node.Kind {
	case ast.KindVariableStatement:
		declarations := node.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes
		return syntaxVariableDeclarations(declarations, span)
	case ast.KindFunctionDeclaration:
		result := SyntaxStatement{
			Span:           span,
			Kind:           "function",
			Name:           node.Name().Text(),
			Type:           syntaxType(node.Type()),
			TypeParameters: syntaxTypeParameters(node.TypeParameters()),
		}
		for _, parameter := range node.Parameters() {
			result.Parameters = append(result.Parameters, SyntaxParameter{
				Span:        parameterSpan(parameter),
				Name:        parameter.Name().Text(),
				Type:        syntaxType(parameter.Type()),
				Rest:        parameter.AsParameterDeclaration().DotDotDotToken != nil,
				Initializer: syntaxExpression(parameter.Initializer()),
			})
		}
		if body := node.Body(); body != nil {
			for _, statement := range body.Statements() {
				if converted, ok := syntaxStatement(statement); ok {
					result.Body = append(result.Body, converted)
				}
			}
		}
		return result, true
	case ast.KindClassDeclaration:
		return syntaxClassDeclaration(node, span)
	case ast.KindReturnStatement:
		return SyntaxStatement{Span: span, Kind: "return", Expression: syntaxExpression(node.Expression())}, true
	case ast.KindThrowStatement:
		return SyntaxStatement{Span: span, Kind: "throw", Expression: syntaxExpression(node.Expression())}, true
	case ast.KindTryStatement:
		tryNode := node.AsTryStatement()
		res := SyntaxStatement{
			Span: span,
			Kind: "try",
			Body: syntaxBlockStatements(tryNode.TryBlock),
		}
		if tryNode.CatchClause != nil {
			catchClause := tryNode.CatchClause.AsCatchClause()
			if catchClause.VariableDeclaration != nil {
				res.CatchVar = catchClause.VariableDeclaration.Name().Text()
			}
			res.Catch = syntaxBlockStatements(catchClause.Block)
		}
		if tryNode.FinallyBlock != nil {
			res.Finally = syntaxBlockStatements(tryNode.FinallyBlock)
		}
		return res, true
	case ast.KindBreakStatement:
		return SyntaxStatement{Span: span, Kind: "break"}, true
	case ast.KindContinueStatement:
		return SyntaxStatement{Span: span, Kind: "continue"}, true
	case ast.KindIfStatement:
		ifNode := node.AsIfStatement()
		result := SyntaxStatement{Span: span, Kind: "if", Expression: syntaxExpression(ifNode.Expression)}
		result.Then = syntaxBlockStatements(ifNode.ThenStatement)
		result.Else = syntaxBlockStatements(ifNode.ElseStatement)
		return result, true
	case ast.KindWhileStatement:
		whileNode := node.AsWhileStatement()
		return SyntaxStatement{
			Span:       span,
			Kind:       "while",
			Expression: syntaxExpression(whileNode.Expression),
			Body:       syntaxBlockStatements(whileNode.Statement),
		}, true
	case ast.KindDoStatement:
		doNode := node.AsDoStatement()
		return SyntaxStatement{
			Span:       span,
			Kind:       "dowhile",
			Expression: syntaxExpression(doNode.Expression),
			Body:       syntaxBlockStatements(doNode.Statement),
		}, true
	case ast.KindForOfStatement:
		forOf := node.AsForInOrOfStatement()
		var varName, varType string
		if forOf.Initializer != nil && forOf.Initializer.Kind == ast.KindVariableDeclarationList {
			decls := forOf.Initializer.AsVariableDeclarationList().Declarations.Nodes
			if len(decls) == 1 {
				varName = decls[0].Name().Text()
				varType = syntaxType(decls[0].Type())
			}
		}
		return SyntaxStatement{
			Span:       span,
			Kind:       "forof",
			Name:       varName,
			Type:       varType,
			Expression: syntaxExpression(forOf.Expression),
			Body:       syntaxBlockStatements(forOf.Statement),
		}, true
	case ast.KindSwitchStatement:
		switchNode := node.AsSwitchStatement()
		switchExpr := syntaxExpression(switchNode.Expression)
		type caseItem struct {
			expr  *SyntaxExpression
			stmts []SyntaxStatement
		}
		var cases []caseItem
		var defaultStmts []SyntaxStatement
		for _, clause := range switchNode.CaseBlock.AsNode().AsCaseBlock().Clauses.Nodes {
			caseClause := clause.AsCaseOrDefaultClause()
			if clause.Kind == ast.KindCaseClause {
				cExpr := syntaxExpression(caseClause.Expression)
				var stmts []SyntaxStatement
				if caseClause.Statements != nil {
					for _, s := range caseClause.Statements.Nodes {
						if s.Kind == ast.KindBreakStatement {
							continue
						}
						if converted, ok := syntaxStatement(s); ok {
							stmts = append(stmts, converted)
						}
					}
				}
				cases = append(cases, caseItem{expr: cExpr, stmts: stmts})
			} else if clause.Kind == ast.KindDefaultClause {
				if caseClause.Statements != nil {
					for _, s := range caseClause.Statements.Nodes {
						if s.Kind == ast.KindBreakStatement {
							continue
						}
						if converted, ok := syntaxStatement(s); ok {
							defaultStmts = append(defaultStmts, converted)
						}
					}
				}
			}
		}
		currentElse := defaultStmts
		for i := len(cases) - 2; i >= 0; i-- {
			if len(cases[i].stmts) == 0 {
				cases[i].stmts = cases[i+1].stmts
			}
		}
		for i := len(cases) - 1; i >= 0; i-- {
			c := cases[i]
			condExpr := &SyntaxExpression{
				Span:     span,
				Kind:     "binary",
				Operator: "===",
				Left:     switchExpr,
				Right:    c.expr,
			}
			ifStmt := SyntaxStatement{
				Span:       span,
				Kind:       "if",
				Expression: condExpr,
				Then:       c.stmts,
				Else:       currentElse,
			}
			currentElse = []SyntaxStatement{ifStmt}
		}
		if len(currentElse) == 1 {
			return currentElse[0], true
		}
		return SyntaxStatement{
			Span: span,
			Kind: "block",
			Body: currentElse,
		}, true
	case ast.KindForStatement:
		forNode := node.AsForStatement()
		var bodyStatements []SyntaxStatement
		bodyStatements = append(bodyStatements, syntaxBlockStatements(forNode.Statement)...)
		if forNode.Incrementor != nil {
			incExpr := syntaxExpression(forNode.Incrementor)
			if incExpr != nil && incExpr.Kind == "binary" && isAssignmentOperator(incExpr.Operator) {
				if incExpr.Left != nil && incExpr.Left.Kind == "identifier" {
					valExpr, _ := desugarAssignment(incExpr)
					bodyStatements = append(bodyStatements, SyntaxStatement{
						Span:       sourceSpan(forNode.Incrementor),
						Kind:       "assign",
						Name:       incExpr.Left.Text,
						Expression: valExpr,
					})
				}
			} else {
				bodyStatements = append(bodyStatements, SyntaxStatement{
					Span:       sourceSpan(forNode.Incrementor),
					Kind:       "expression",
					Expression: incExpr,
				})
			}
		}
		whileStmt := SyntaxStatement{
			Span:       span,
			Kind:       "while",
			Expression: syntaxExpression(forNode.Condition),
			Body:       bodyStatements,
		}
		if forNode.Initializer != nil {
			if forNode.Initializer.Kind == ast.KindVariableDeclarationList {
				decls := forNode.Initializer.AsVariableDeclarationList().Declarations.Nodes
				initStmt, ok := syntaxVariableDeclarations(decls, sourceSpan(forNode.Initializer))
				if ok {
					return SyntaxStatement{
						Span: span,
						Kind: "block",
						Body: []SyntaxStatement{initStmt, whileStmt},
					}, true
				}
			} else {
				initExpr := syntaxExpression(forNode.Initializer)
				if initExpr != nil && initExpr.Kind == "binary" && isAssignmentOperator(initExpr.Operator) {
					if initExpr.Left != nil && initExpr.Left.Kind == "identifier" {
						valExpr, _ := desugarAssignment(initExpr)
						initStmt := SyntaxStatement{
							Span:       sourceSpan(forNode.Initializer),
							Kind:       "assign",
							Name:       initExpr.Left.Text,
							Expression: valExpr,
						}
						return SyntaxStatement{
							Span: span,
							Kind: "block",
							Body: []SyntaxStatement{initStmt, whileStmt},
						}, true
					}
				}
				initStmt := SyntaxStatement{
					Span:       sourceSpan(forNode.Initializer),
					Kind:       "expression",
					Expression: initExpr,
				}
				return SyntaxStatement{
					Span: span,
					Kind: "block",
					Body: []SyntaxStatement{initStmt, whileStmt},
				}, true
			}
		}
		return whileStmt, true
	case ast.KindExpressionStatement:
		expr := syntaxExpression(node.Expression())
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
	case ast.KindEnumDeclaration:
		enumDecl := node.AsEnumDeclaration()
		enumObj := &SyntaxEnum{
			Span: span,
			Name: node.Name().Text(),
		}
		var currentNumericVal float64 = 0
		if enumDecl.Members != nil {
			for _, memberNode := range enumDecl.Members.Nodes {
				member := memberNode.AsEnumMember()
				memName := member.Name().Text()
				initExpr := syntaxExpression(member.Initializer)
				m := SyntaxEnumMember{
					Span:        sourceSpan(memberNode),
					Name:        memName,
					Initializer: initExpr,
				}
				if initExpr == nil {
					m.Value = strconv.FormatFloat(currentNumericVal, 'f', -1, 64)
					currentNumericVal++
				} else if initExpr.Kind == "number" {
					if v, err := strconv.ParseFloat(initExpr.Text, 64); err == nil {
						currentNumericVal = v + 1
						m.Value = initExpr.Text
					}
				} else if initExpr.Kind == "string" {
					m.Value = initExpr.Text
				}
				enumObj.Members = append(enumObj.Members, m)
			}
		}
		return SyntaxStatement{Span: span, Kind: "enum", Name: enumObj.Name, Enum: enumObj}, true
	case ast.KindImportDeclaration, ast.KindExportDeclaration, ast.KindModuleDeclaration:
		return SyntaxStatement{Span: span, Kind: "module", Type: node.Kind.String()}, true
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		name := ""
		if node.Name() != nil {
			name = node.Name().Text()
		}
		return SyntaxStatement{Span: span, Kind: "interface", Name: name}, true
	default:
		return SyntaxStatement{Span: span, Kind: "unsupported", Type: node.Kind.String()}, true
	}
}

func syntaxVariableDeclarations(decls []*ast.Node, span SourceSpan) (SyntaxStatement, bool) {
	var stmts []SyntaxStatement
	for _, declaration := range decls {
		nameNode := declaration.Name()
		initExpr := syntaxExpression(declaration.Initializer())
		if nameNode.AsNode().Kind == ast.KindObjectBindingPattern {
			pattern := nameNode.AsNode().AsBindingPattern()
			objVar := fmt.Sprintf("__destruct_obj_%d", nameNode.AsNode().Pos())
			if initExpr != nil && initExpr.Kind == "identifier" {
				objVar = initExpr.Text
			} else {
				stmts = append(stmts, SyntaxStatement{
					Span:       sourceSpan(nameNode.AsNode()),
					Kind:       "variable",
					Name:       objVar,
					Expression: initExpr,
				})
			}
			for _, elem := range pattern.Elements.Nodes {
				binding := elem.AsBindingElement()
				propName := binding.Name().Text()
				varName := binding.Name().Text()
				if binding.PropertyName != nil {
					propName = binding.PropertyName.Text()
				}
				propExpr := &SyntaxExpression{
					Span: sourceSpan(elem),
					Kind: "property",
					Text: propName,
					Left: &SyntaxExpression{
						Span: sourceSpan(elem),
						Kind: "identifier",
						Text: objVar,
					},
				}
				stmts = append(stmts, SyntaxStatement{
					Span:       sourceSpan(elem),
					Kind:       "variable",
					Name:       varName,
					Expression: propExpr,
				})
			}
		} else if nameNode.AsNode().Kind == ast.KindArrayBindingPattern {
			pattern := nameNode.AsNode().AsBindingPattern()
			arrVar := fmt.Sprintf("__destruct_arr_%d", nameNode.AsNode().Pos())
			if initExpr != nil && initExpr.Kind == "identifier" {
				arrVar = initExpr.Text
			} else {
				stmts = append(stmts, SyntaxStatement{
					Span:       sourceSpan(nameNode.AsNode()),
					Kind:       "variable",
					Name:       arrVar,
					Expression: initExpr,
				})
			}
			for idx, elem := range pattern.Elements.Nodes {
				if elem.Kind == ast.KindOmittedExpression {
					continue
				}
				binding := elem.AsBindingElement()
				varName := binding.Name().Text()
				if binding.DotDotDotToken != nil {
					sliceExpr := &SyntaxExpression{
						Span: sourceSpan(elem),
						Kind: "call",
						Left: &SyntaxExpression{
							Span: sourceSpan(elem),
							Kind: "property",
							Text: "slice",
							Left: &SyntaxExpression{Span: sourceSpan(elem), Kind: "identifier", Text: arrVar},
						},
						Arguments: []*SyntaxExpression{
							{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx)},
							{Span: sourceSpan(elem), Kind: "property", Text: "length", Left: &SyntaxExpression{Span: sourceSpan(elem), Kind: "identifier", Text: arrVar}},
						},
					}
					stmts = append(stmts, SyntaxStatement{
						Span:       sourceSpan(elem),
						Kind:       "variable",
						Name:       varName,
						Expression: sliceExpr,
					})
				} else {
					idxExpr := &SyntaxExpression{
						Span:  sourceSpan(elem),
						Kind:  "index",
						Left:  &SyntaxExpression{Span: sourceSpan(elem), Kind: "identifier", Text: arrVar},
						Right: &SyntaxExpression{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx)},
					}
					stmts = append(stmts, SyntaxStatement{
						Span:       sourceSpan(elem),
						Kind:       "variable",
						Name:       varName,
						Expression: idxExpr,
					})
				}
			}
		} else {
			result := SyntaxStatement{Span: sourceSpan(declaration), Kind: "variable", Name: nameNode.Text()}
			result.Type = syntaxType(declaration.Type())
			result.Expression = initExpr
			stmts = append(stmts, result)
		}
	}
	if len(stmts) == 1 {
		return stmts[0], true
	}
	return SyntaxStatement{
		Span: span,
		Kind: "block",
		Body: stmts,
	}, true
}

func syntaxBlockStatements(node *ast.Node) []SyntaxStatement {
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindBlock {
		result := make([]SyntaxStatement, 0, len(node.Statements()))
		for _, statement := range node.Statements() {
			if converted, ok := syntaxStatement(statement); ok {
				result = append(result, converted)
			}
		}
		return result
	}
	if converted, ok := syntaxStatement(node); ok {
		return []SyntaxStatement{converted}
	}
	return nil
}

func isAssignmentOperator(op string) bool {
	if op == "=" {
		return true
	}
	if strings.HasSuffix(op, "=") && op != "==" && op != "===" && op != "!=" && op != "!==" && op != "<=" && op != ">=" {
		return true
	}
	return false
}

func desugarAssignment(expr *SyntaxExpression) (*SyntaxExpression, bool) {
	if expr == nil || expr.Kind != "binary" {
		return nil, false
	}
	if expr.Operator == "=" {
		return expr.Right, true
	}
	if isAssignmentOperator(expr.Operator) {
		baseOp := strings.TrimSuffix(expr.Operator, "=")
		return &SyntaxExpression{
			Span:     expr.Span,
			Kind:     "binary",
			Operator: baseOp,
			Left:     expr.Left,
			Right:    expr.Right,
		}, true
	}
	return nil, false
}

