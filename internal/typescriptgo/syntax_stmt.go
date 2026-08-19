package typescriptgo

import (
	"fmt"

	"github.com/microsoft/typescript-go/internal/ast"
)

func syntaxStatement(node *ast.Node) (SyntaxStatement, bool) {
	span := sourceSpan(node)
	switch node.Kind {
	case ast.KindVariableStatement:
		declarations := node.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes
		return syntaxVariableDeclarations(declarations, span)
	case ast.KindFunctionDeclaration:
		result := SyntaxStatement{Span: span, Kind: "function", Name: node.Name().Text(), Type: syntaxType(node.Type())}
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
		class := &SyntaxClass{Span: span, Name: node.Name().Text()}
		for _, member := range node.Members() {
			switch member.Kind {
			case ast.KindPropertyDeclaration:
				property := member.AsPropertyDeclaration()
				class.Fields = append(class.Fields, SyntaxField{
					Span:        sourceSpan(member),
					Name:        property.Name().Text(),
					Type:        syntaxType(property.Type),
					Initializer: syntaxExpression(property.Initializer),
				})
			case ast.KindConstructor:
				var params []SyntaxParameter
				for _, p := range member.Parameters() {
					params = append(params, SyntaxParameter{
						Span:        parameterSpan(p),
						Name:        p.Name().Text(),
						Type:        syntaxType(p.Type()),
						Rest:        p.AsParameterDeclaration().DotDotDotToken != nil,
						Initializer: syntaxExpression(p.Initializer()),
					})
				}
				var body []SyntaxStatement
				if b := member.Body(); b != nil {
					for _, s := range b.Statements() {
						if converted, ok := syntaxStatement(s); ok {
							body = append(body, converted)
						}
					}
				}
				class.Constructor = &SyntaxConstructor{
					Span:       sourceSpan(member),
					Parameters: params,
					Body:       body,
				}
			case ast.KindMethodDeclaration:
				var params []SyntaxParameter
				for _, p := range member.Parameters() {
					params = append(params, SyntaxParameter{
						Span:        parameterSpan(p),
						Name:        p.Name().Text(),
						Type:        syntaxType(p.Type()),
						Rest:        p.AsParameterDeclaration().DotDotDotToken != nil,
						Initializer: syntaxExpression(p.Initializer()),
					})
				}
				var body []SyntaxStatement
				if b := member.Body(); b != nil {
					for _, s := range b.Statements() {
						if converted, ok := syntaxStatement(s); ok {
							body = append(body, converted)
						}
					}
				}
				class.Methods = append(class.Methods, SyntaxMethod{
					Span:       sourceSpan(member),
					Name:       member.Name().Text(),
					Type:       syntaxType(member.Type()),
					Parameters: params,
					Body:       body,
				})
			default:
				class.Fields = append(class.Fields, SyntaxField{Span: sourceSpan(member), Name: member.Kind.String()})
			}
		}
		return SyntaxStatement{Span: span, Kind: "class", Name: class.Name, Class: class}, true
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
			if incExpr != nil && incExpr.Kind == "binary" && (incExpr.Operator == "=" || incExpr.Operator == "+=" || incExpr.Operator == "-=") {
				if incExpr.Left != nil && incExpr.Left.Kind == "identifier" {
					valExpr := incExpr.Right
					if incExpr.Operator == "+=" {
						valExpr = &SyntaxExpression{
							Span:     incExpr.Span,
							Kind:     "binary",
							Operator: "+",
							Left:     incExpr.Left,
							Right:    incExpr.Right,
						}
					} else if incExpr.Operator == "-=" {
						valExpr = &SyntaxExpression{
							Span:     incExpr.Span,
							Kind:     "binary",
							Operator: "-",
							Left:     incExpr.Left,
							Right:    incExpr.Right,
						}
					}
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
				if initExpr != nil && initExpr.Kind == "binary" && (initExpr.Operator == "=" || initExpr.Operator == "+=" || initExpr.Operator == "-=") {
					if initExpr.Left != nil && initExpr.Left.Kind == "identifier" {
						valExpr := initExpr.Right
						if initExpr.Operator == "+=" {
							valExpr = &SyntaxExpression{
								Span:     initExpr.Span,
								Kind:     "binary",
								Operator: "+",
								Left:     initExpr.Left,
								Right:    initExpr.Right,
							}
						} else if initExpr.Operator == "-=" {
							valExpr = &SyntaxExpression{
								Span:     initExpr.Span,
								Kind:     "binary",
								Operator: "-",
								Left:     initExpr.Left,
								Right:    initExpr.Right,
							}
						}
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
		if expr != nil && expr.Kind == "binary" && (expr.Operator == "=" || expr.Operator == "+=" || expr.Operator == "-=") {
			if expr.Left != nil {
				if expr.Left.Kind == "identifier" {
					valExpr := expr.Right
					switch expr.Operator {
					case "+=":
						valExpr = &SyntaxExpression{
							Span:     expr.Span,
							Kind:     "binary",
							Operator: "+",
							Left:     expr.Left,
							Right:    expr.Right,
						}
					case "-=":
						valExpr = &SyntaxExpression{
							Span:     expr.Span,
							Kind:     "binary",
							Operator: "-",
							Left:     expr.Left,
							Right:    expr.Right,
						}
					}
					return SyntaxStatement{
						Span:       span,
						Kind:       "assign",
						Name:       expr.Left.Text,
						Expression: valExpr,
					}, true
				}
				if expr.Left.Kind == "index" {
					valExpr := expr.Right
					switch expr.Operator {
					case "+=":
						valExpr = &SyntaxExpression{
							Span:     expr.Span,
							Kind:     "binary",
							Operator: "+",
							Left:     expr.Left,
							Right:    expr.Right,
						}
					case "-=":
						valExpr = &SyntaxExpression{
							Span:     expr.Span,
							Kind:     "binary",
							Operator: "-",
							Left:     expr.Left,
							Right:    expr.Right,
						}
					}
					return SyntaxStatement{
						Span:       span,
						Kind:       "index_set",
						Left:       expr.Left.Left,
						Right:      expr.Left.Right,
						Expression: valExpr,
					}, true
				}
				if expr.Left.Kind == "property" {
					valExpr := expr.Right
					switch expr.Operator {
					case "+=":
						valExpr = &SyntaxExpression{
							Span:     expr.Span,
							Kind:     "binary",
							Operator: "+",
							Left:     expr.Left,
							Right:    expr.Right,
						}
					case "-=":
						valExpr = &SyntaxExpression{
							Span:     expr.Span,
							Kind:     "binary",
							Operator: "-",
							Left:     expr.Left,
							Right:    expr.Right,
						}
					}
					return SyntaxStatement{
						Span:       span,
						Kind:       "field_set",
						Left:       expr.Left.Left,
						Name:       expr.Left.Text,
						Expression: valExpr,
					}, true
				}
			}
		}
		return SyntaxStatement{Span: span, Kind: "expression", Expression: expr}, true
	case ast.KindImportDeclaration, ast.KindExportDeclaration, ast.KindModuleDeclaration:
		return SyntaxStatement{Span: span, Kind: "module", Type: node.Kind.String()}, true
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
