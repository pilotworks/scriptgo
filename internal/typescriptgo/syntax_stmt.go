package typescriptgo

import (
	"fmt"
	"strconv"
	"strings"

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
		isUsing := (declList.Flags & ast.NodeFlagsBlockScoped) == ast.NodeFlagsUsing
		isAwaitUsing := (declList.Flags & ast.NodeFlagsBlockScoped) == ast.NodeFlagsAwaitUsing
		return syntaxVariableDeclarations(declarations, span, chk, isUsing, isAwaitUsing)
	case ast.KindFunctionDeclaration:
		fnType := syntaxType(node.Type())
		inferredRetType := resolveFunctionReturnType(chk, node)
		if fnType == "" && inferredRetType != "" {
			fnType = inferredRetType
		}
		isGen := node.BodyData() != nil && node.BodyData().AsteriskToken != nil
		isAsync := ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
		isAmbient := node.Body() == nil || ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient)
		kind := "function"
		if isAmbient {
			kind = "declare_function"
		} else if isGen {
			if isAsync {
				kind = "async_generator_function"
			} else {
				kind = "generator_function"
			}
		} else if isAsync {
			kind = "async_function"
		}
		result := SyntaxStatement{
			Span:           span,
			Kind:           kind,
			Name:           node.Name().Text(),
			Type:           fnType,
			InferredType:   inferredRetType,
			TypeParameters: syntaxTypeParameters(node.TypeParameters()),
			IsGenerator:    isGen,
			IsAsync:        isAsync,
		}
		for _, parameter := range node.Parameters() {
			pType := syntaxType(parameter.Type())
			inferredPType := resolveInferredType(chk, parameter.Name())
			if inferredPType == "" {
				inferredPType = resolveInferredType(chk, parameter)
			}
			result.Parameters = append(result.Parameters, SyntaxParameter{
				Span:         parameterSpan(parameter),
				Name:         parameter.Name().Text(),
				Type:         pType,
				InferredType: inferredPType,
				Rest:         parameter.AsParameterDeclaration().DotDotDotToken != nil,
				Optional:     parameter.AsParameterDeclaration().QuestionToken != nil,
				Initializer:  syntaxExpression(parameter.Initializer(), chk),
			})
		}
		if body := node.Body(); body != nil {
			for _, statement := range body.Statements() {
				if converted, ok := syntaxStatement(statement, chk); ok {
					result.Body = append(result.Body, converted)
				}
			}
		}
		return result, true
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
	case ast.KindForOfStatement:
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
				} else if nameNode.Kind == ast.KindObjectBindingPattern {
					tempItemVar := fmt.Sprintf("__forof_obj_%d", nameNode.Pos())
					varName = tempItemVar
					varInferredType = resolveInferredType(chk, nameNode)
					varType = varInferredType
					pattern := nameNode.AsBindingPattern()
					for _, elem := range pattern.Elements.Nodes {
						binding := elem.AsBindingElement()
						propName := binding.Name().Text()
						vName := binding.Name().Text()
						if binding.PropertyName != nil {
							propName = binding.PropertyName.Text()
						}
						propExpr := &SyntaxExpression{
							Span:         sourceSpan(elem),
							Kind:         "property",
							Text:         propName,
							InferredType: resolveInferredType(chk, elem),
							Left: &SyntaxExpression{
								Span: sourceSpan(elem),
								Kind: "identifier",
								Text: tempItemVar,
							},
						}
						bindingStmts = append(bindingStmts, SyntaxStatement{
							Span:         sourceSpan(elem),
							Kind:         "variable",
							Name:         vName,
							InferredType: resolveInferredType(chk, elem),
							Expression:   propExpr,
						})
					}
				} else if nameNode.Kind == ast.KindArrayBindingPattern {
					tempItemVar := fmt.Sprintf("__forof_arr_%d", nameNode.Pos())
					varName = tempItemVar
					varInferredType = resolveInferredType(chk, nameNode)
					varType = varInferredType
					pattern := nameNode.AsBindingPattern()
					for idx, elem := range pattern.Elements.Nodes {
						if elem.Kind == ast.KindOmittedExpression {
							continue
						}
						binding := elem.AsBindingElement()
						vName := binding.Name().Text()
						indexExpr := &SyntaxExpression{
							Span:         sourceSpan(elem),
							Kind:         "index",
							InferredType: resolveInferredType(chk, elem),
							Left:         &SyntaxExpression{Span: sourceSpan(elem), Kind: "identifier", Text: tempItemVar},
							Right:        &SyntaxExpression{Span: sourceSpan(elem), Kind: "number", Text: fmt.Sprintf("%d", idx), InferredType: "number"},
						}
						bindingStmts = append(bindingStmts, SyntaxStatement{
							Span:         sourceSpan(elem),
							Kind:         "variable",
							Name:         vName,
							InferredType: resolveInferredType(chk, elem),
							Expression:   indexExpr,
						})
					}
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
					if converted, ok := syntaxStatement(s, chk); ok {
						stmts = append(stmts, converted)
					}
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
		forNode := node.AsForStatement()
		var bodyStatements []SyntaxStatement
		bodyStatements = append(bodyStatements, syntaxBlockStatements(forNode.Statement, chk)...)
		var stepStatements []SyntaxStatement
		if forNode.Incrementor != nil {
			incExpr := syntaxExpression(forNode.Incrementor, chk)
			if incExpr != nil && incExpr.Kind == "binary" && isAssignmentOperator(incExpr.Operator) {
				if incExpr.Left != nil && incExpr.Left.Kind == "identifier" {
					valExpr, _ := desugarAssignment(incExpr)
					stepStatements = append(stepStatements, SyntaxStatement{
						Span:       sourceSpan(forNode.Incrementor),
						Kind:       "assign",
						Name:       incExpr.Left.Text,
						Expression: valExpr,
					})
				}
			} else {
				stepStatements = append(stepStatements, SyntaxStatement{
					Span:       sourceSpan(forNode.Incrementor),
					Kind:       "expression",
					Expression: incExpr,
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
				initStmt, ok := syntaxVariableDeclarations(decls, sourceSpan(forNode.Initializer), chk, false, false)
				if ok {
					return SyntaxStatement{
						Span: span,
						Kind: "block",
						Body: []SyntaxStatement{initStmt, whileStmt},
					}, true
				}
			} else {
				initExpr := syntaxExpression(forNode.Initializer, chk)
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
				initExpr := syntaxExpression(member.Initializer, chk)
				m := SyntaxEnumMember{
					Span:        sourceSpan(memberNode),
					Name:        memName,
					Initializer: initExpr,
				}
				if initExpr == nil {
					m.Value = strconv.FormatFloat(currentNumericVal, 'f', -1, 64)
					currentNumericVal++
				} else if initExpr.Kind == "string" {
					m.Value = initExpr.Text
				} else if v, ok := evalConstNumber(initExpr); ok {
					m.Value = strconv.FormatFloat(v, 'f', -1, 64)
					currentNumericVal = v + 1
				}
				enumObj.Members = append(enumObj.Members, m)
			}
		}
		return SyntaxStatement{Span: span, Kind: "enum", Name: enumObj.Name, Enum: enumObj}, true
	case ast.KindImportDeclaration, ast.KindExportDeclaration:
		return SyntaxStatement{Span: span, Kind: "module", Type: node.Kind.String()}, true
	case ast.KindModuleDeclaration:
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient) || (node.Name() != nil && (strings.HasPrefix(node.Name().Text(), "\"") || strings.HasPrefix(node.Name().Text(), "'"))) {
			return SyntaxStatement{Span: span, Kind: "module", Type: node.Kind.String()}, true
		}
		modDecl := node.AsModuleDeclaration()
		name := ""
		if node.Name() != nil {
			name = node.Name().Text()
		}
		var bodyStmts []SyntaxStatement
		if modDecl != nil && modDecl.Body != nil {
			if modBlock := modDecl.Body.AsModuleBlock(); modBlock != nil && modBlock.Statements != nil {
				for _, s := range modBlock.Statements.Nodes {
					if conv, ok := syntaxStatement(s, chk); ok {
						bodyStmts = append(bodyStmts, conv)
					}
				}
			}
		}
		return SyntaxStatement{
			Span: span,
			Kind: "namespace",
			Name: name,
			Body: bodyStmts,
		}, true
	case ast.KindInterfaceDeclaration:
		iface := node.AsInterfaceDeclaration()
		name := ""
		if node.Name() != nil {
			name = node.Name().Text()
		}
		var extendsName string
		if iface != nil && iface.HeritageClauses != nil {
			for _, clause := range iface.HeritageClauses.Nodes {
				hc := clause.AsHeritageClause()
				if hc != nil && hc.Token == ast.KindExtendsKeyword && hc.Types != nil {
					for _, t := range hc.Types.Nodes {
						if t.Kind == ast.KindTypeReference {
							extendsName = syntaxType(t)
						} else if t.Kind == ast.KindExpressionWithTypeArguments {
							exprNode := t.AsExpressionWithTypeArguments()
							if exprNode != nil && exprNode.Expression != nil {
								extendsName = exprNode.Expression.Text()
								if exprNode.TypeArguments != nil && len(exprNode.TypeArguments.Nodes) > 0 {
									var typeArgs []string
									for _, ta := range exprNode.TypeArguments.Nodes {
										typeArgs = append(typeArgs, syntaxType(ta))
									}
									extendsName = exprNode.Expression.Text() + "<" + strings.Join(typeArgs, ", ") + ">"
								}
							}
						} else {
							extendsName = syntaxType(t)
						}
					}
				}
			}
		}
		var fields []SyntaxField
		members := node.Members()
		if len(members) == 0 && iface != nil && iface.Members != nil {
			members = iface.Members.Nodes
		}
		for _, member := range members {
			if member.Kind == ast.KindPropertySignature {
				fType := syntaxType(member.Type())
				inferredFType := resolveInferredType(chk, member.Name())
				if inferredFType == "" {
					inferredFType = resolveInferredType(chk, member)
				}
				pName := ""
				if member.Name() != nil {
					pName = member.Name().Text()
				}
				fields = append(fields, SyntaxField{
					Span:         sourceSpan(member),
					Name:         pName,
					Type:         fType,
					InferredType: inferredFType,
				})
			}
		}
		tParams := syntaxTypeParameters(node.TypeParameters())
		if len(tParams) == 0 && iface != nil && iface.TypeParameters != nil {
			tParams = syntaxTypeParameters(iface.TypeParameters.Nodes)
		}
		cls := &SyntaxClass{
			Span:           span,
			Name:           name,
			TypeParameters: tParams,
			Extends:        extendsName,
			Fields:         fields,
		}
		return SyntaxStatement{Span: span, Kind: "interface", Name: name, Class: cls}, true

	case ast.KindTypeAliasDeclaration:
		alias := node.AsTypeAliasDeclaration()
		name := ""
		if node.Name() != nil {
			name = node.Name().Text()
		}
		var fields []SyntaxField
		if alias != nil && alias.Type != nil {
			if alias.Type.Kind == ast.KindTypeLiteral {
				for _, member := range alias.Type.Members() {
					if member.Kind == ast.KindPropertySignature {
						pName := ""
						if member.Name() != nil {
							pName = member.Name().Text()
						}
						fields = append(fields, SyntaxField{
							Span:         sourceSpan(member),
							Name:         pName,
							Type:         syntaxType(member.Type()),
							InferredType: resolveInferredType(chk, member),
						})
					}
				}
			} else if alias.Type.Kind == ast.KindUnionType {
				unionNode := alias.Type.AsUnionTypeNode()
				if unionNode != nil && unionNode.Types != nil {
					for _, t := range unionNode.Types.Nodes {
						if t.Kind == ast.KindTypeLiteral {
							for _, member := range t.Members() {
								if member.Kind == ast.KindPropertySignature {
									pName := ""
									if member.Name() != nil {
										pName = member.Name().Text()
									}
									fields = append(fields, SyntaxField{
										Span:         sourceSpan(member),
										Name:         pName,
										Type:         syntaxType(member.Type()),
										InferredType: resolveInferredType(chk, member),
									})
								}
							}
						}
					}
				}
			}
		}
		cls := &SyntaxClass{
			Span:           span,
			Name:           name,
			Fields:         fields,
			TypeParameters: syntaxTypeParameters(node.TypeParameters()),
		}
		tStr := ""
		if alias != nil && alias.Type != nil {
			tStr = syntaxType(alias.Type)
		}
		return SyntaxStatement{Span: span, Kind: "type_alias", Name: name, Type: tStr, Class: cls}, true
	default:
		return SyntaxStatement{Span: span, Kind: "unsupported", Type: node.Kind.String()}, true
	}
}
