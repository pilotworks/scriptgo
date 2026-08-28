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
		fnType := syntaxType(node.Type())
		inferredRetType := resolveFunctionReturnType(chk, node)
		if fnType == "" && inferredRetType != "" && inferredRetType != "any" && inferredRetType != "unknown" {
			fnType = inferredRetType
		}
		fnDecl := node.AsFunctionDeclaration()
		isGen := (node.BodyData() != nil && node.BodyData().AsteriskToken != nil) || (fnDecl != nil && fnDecl.AsteriskToken != nil) || strings.HasPrefix(inferredRetType, "Generator") || strings.HasPrefix(inferredRetType, "AsyncGenerator") || strings.HasPrefix(fnType, "Generator") || strings.HasPrefix(fnType, "AsyncGenerator")
		if !isGen && node.Body() != nil {
			for _, s := range node.Body().Statements() {
				if s.Kind == ast.KindExpressionStatement && s.Expression() != nil && s.Expression().Kind == ast.KindYieldExpression {
					isGen = true
					break
				}
			}
		}
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
		var bindingStmts []SyntaxStatement
		for pIdx, parameter := range node.Parameters() {
			pType := syntaxType(parameter.Type())
			inferredPType := resolveInferredType(chk, parameter.Name())
			if inferredPType == "" {
				inferredPType = resolveInferredType(chk, parameter)
			}
			pName, binds := extractParameterBinding(parameter, pIdx, chk)
			if len(binds) > 0 {
				bindingStmts = append(bindingStmts, binds...)
			}
			result.Parameters = append(result.Parameters, SyntaxParameter{
				Span:         parameterSpan(parameter),
				Name:         pName,
				Type:         pType,
				InferredType: inferredPType,
				Rest:         parameter.AsParameterDeclaration().DotDotDotToken != nil,
				Optional:     parameter.AsParameterDeclaration().QuestionToken != nil,
				Initializer:  syntaxExpression(parameter.Initializer(), chk),
			})
		}
		if len(bindingStmts) > 0 {
			result.Body = append(result.Body, bindingStmts...)
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
				initStmt, ok := syntaxVariableDeclarations(decls, sourceSpan(forNode.Initializer), chk, "var")
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
	case ast.KindEnumDeclaration:
		enumDecl := node.AsEnumDeclaration()
		enumObj := &SyntaxEnum{
			Span:    span,
			Name:    node.Name().Text(),
			IsConst: ast.HasSyntacticModifier(node, ast.ModifierFlagsConst),
		}
		var currentNumericVal float64 = 0
		enumVals := make(map[string]float64)
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
					enumVals[memName] = currentNumericVal
					currentNumericVal++
				} else if initExpr.Kind == "string" {
					m.Value = initExpr.Text
				} else if v, ok := evalConstNumberWithEnv(initExpr, enumVals); ok {
					m.Value = strconv.FormatFloat(v, 'f', -1, 64)
					enumVals[memName] = v
					currentNumericVal = v + 1
				}
				enumObj.Members = append(enumObj.Members, m)
			}
		}
		return SyntaxStatement{Span: span, Kind: "enum", Name: enumObj.Name, Enum: enumObj}, true
	case ast.KindImportDeclaration, ast.KindExportDeclaration, ast.KindExportAssignment:
		return SyntaxStatement{Span: span, Kind: "module", Type: node.Kind.String()}, true
	case ast.KindModuleDeclaration:
		modDecl := node.AsModuleDeclaration()
		name := ""
		if node.Name() != nil {
			name = strings.Trim(node.Name().Text(), "\"'")
		}
		var bodyStmts []SyntaxStatement
		if modDecl != nil && modDecl.Body != nil {
			if modBlock := modDecl.Body.AsModuleBlock(); modBlock != nil && modBlock.Statements != nil {
				for _, s := range modBlock.Statements.Nodes {
					if conv, ok := syntaxStatement(s, chk); ok {
						if conv.Kind == "class" && conv.Class != nil && !strings.Contains(conv.Class.Name, ".") {
							conv.Class.Name = name + "." + conv.Class.Name
							conv.Name = conv.Class.Name
						}
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
					var extendsList []string
					for _, t := range hc.Types.Nodes {
						if t.Kind == ast.KindTypeReference {
							extendsList = append(extendsList, syntaxType(t))
						} else if t.Kind == ast.KindExpressionWithTypeArguments {
							exprNode := t.AsExpressionWithTypeArguments()
							if exprNode != nil && exprNode.Expression != nil {
								extName := exprNode.Expression.Text()
								if exprNode.TypeArguments != nil && len(exprNode.TypeArguments.Nodes) > 0 {
									var typeArgs []string
									for _, ta := range exprNode.TypeArguments.Nodes {
										typeArgs = append(typeArgs, syntaxType(ta))
									}
									extName = exprNode.Expression.Text() + "<" + strings.Join(typeArgs, ", ") + ">"
								}
								extendsList = append(extendsList, extName)
							}
						} else {
							extendsList = append(extendsList, syntaxType(t))
						}
					}
					extendsName = strings.Join(extendsList, ", ")
				}
			}
		}
		var fields []SyntaxField
		var methods []SyntaxMethod
		var constructor *SyntaxConstructor
		members := node.Members()
		if len(members) == 0 && iface != nil && iface.Members != nil {
			members = iface.Members.Nodes
		}
		for _, member := range members {
			switch member.Kind {
			case ast.KindPropertySignature:
				fType := syntaxType(member.Type())
				inferredFType := resolveInferredType(chk, member.Name())
				if inferredFType == "" {
					inferredFType = resolveInferredType(chk, member)
				}
				pName := syntaxMemberName(member.Name())
				fields = append(fields, SyntaxField{
					Span:         sourceSpan(member),
					Name:         pName,
					Type:         fType,
					InferredType: inferredFType,
				})
			case ast.KindConstructSignature:
				mType := syntaxType(member.Type())
				var params []SyntaxParameter
				for _, p := range member.Parameters() {
					paramName := syntaxMemberName(p.Name())
					paramType := syntaxType(p.Type())
					pDecl := p.AsParameterDeclaration()
					isOpt := pDecl != nil && pDecl.QuestionToken != nil
					isRest := pDecl != nil && pDecl.DotDotDotToken != nil
					params = append(params, SyntaxParameter{
						Span:     sourceSpan(p),
						Name:     paramName,
						Type:     paramType,
						Optional: isOpt,
						Rest:     isRest,
					})
				}
				constructor = &SyntaxConstructor{
					Span:       sourceSpan(member),
					Parameters: params,
				}
				methods = append(methods, SyntaxMethod{
					Span:       sourceSpan(member),
					Name:       "constructor",
					Type:       mType,
					Parameters: params,
				})
			case ast.KindMethodSignature, ast.KindCallSignature:
				mType := syntaxType(member.Type())
				pName := syntaxMemberName(member.Name())
				var params []SyntaxParameter
				for _, p := range member.Parameters() {
					paramName := syntaxMemberName(p.Name())
					paramType := syntaxType(p.Type())
					pDecl := p.AsParameterDeclaration()
					isOpt := pDecl != nil && pDecl.QuestionToken != nil
					isRest := pDecl != nil && pDecl.DotDotDotToken != nil
					params = append(params, SyntaxParameter{
						Span:     sourceSpan(p),
						Name:     paramName,
						Type:     paramType,
						Optional: isOpt,
						Rest:     isRest,
					})
				}
				methods = append(methods, SyntaxMethod{
					Span:       sourceSpan(member),
					Name:       pName,
					Type:       mType,
					Parameters: params,
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
			Methods:        methods,
			Constructor:    constructor,
			IsAbstract:     true,
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
