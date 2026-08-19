package typescriptgo

import (
	"strings"

	"github.com/microsoft/typescript-go/internal/ast"
)

func syntaxFile(file *ast.SourceFile) SyntaxFile {
	result := SyntaxFile{FileName: file.FileName()}
	if file.Statements == nil {
		return result
	}
	for _, statement := range file.Statements.Nodes {
		if converted, ok := syntaxStatement(statement); ok {
			result.Statements = append(result.Statements, converted)
		}
	}
	return result
}

func syntaxStatement(node *ast.Node) (SyntaxStatement, bool) {
	span := sourceSpan(node)
	switch node.Kind {
	case ast.KindVariableStatement:
		declarations := node.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes
		if len(declarations) != 1 {
			return SyntaxStatement{Span: span, Kind: "unsupported", Type: "multiple variable declarations"}, true
		}
		declaration := declarations[0]
		result := SyntaxStatement{Span: span, Kind: "variable", Name: declaration.Name().Text()}
		result.Type = syntaxType(declaration.Type())
		result.Expression = syntaxExpression(declaration.Initializer())
		return result, true
	case ast.KindFunctionDeclaration:
		result := SyntaxStatement{Span: span, Kind: "function", Name: node.Name().Text(), Type: syntaxType(node.Type())}
		for _, parameter := range node.Parameters() {
			result.Parameters = append(result.Parameters, SyntaxParameter{
				Span: parameterSpan(parameter),
				Name: parameter.Name().Text(),
				Type: syntaxType(parameter.Type()),
				Rest: parameter.AsParameterDeclaration().DotDotDotToken != nil,
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
						Span: parameterSpan(p),
						Name: p.Name().Text(),
						Type: syntaxType(p.Type()),
						Rest: p.AsParameterDeclaration().DotDotDotToken != nil,
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
						Span: parameterSpan(p),
						Name: p.Name().Text(),
						Type: syntaxType(p.Type()),
						Rest: p.AsParameterDeclaration().DotDotDotToken != nil,
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
		caseBlock := switchNode.CaseBlock.AsCaseBlock()
		var clauses []*ast.Node
		if caseBlock != nil && caseBlock.Clauses != nil {
			clauses = caseBlock.Clauses.Nodes
		}
		var defaultStmts []SyntaxStatement
		type caseItem struct {
			expr  *SyntaxExpression
			stmts []SyntaxStatement
		}
		var cases []caseItem
		for _, clause := range clauses {
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
				if len(decls) == 1 {
					initDecl := decls[0]
					initStmt := SyntaxStatement{
						Span:       sourceSpan(forNode.Initializer),
						Kind:       "variable",
						Name:       initDecl.Name().Text(),
						Type:       syntaxType(initDecl.Type()),
						Expression: syntaxExpression(initDecl.Initializer()),
					}
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
	case ast.KindImportDeclaration, ast.KindExportDeclaration:
		return SyntaxStatement{Span: span, Kind: "module", Type: node.Kind.String()}, true
	default:
		return SyntaxStatement{Span: span, Kind: "unsupported", Type: node.Kind.String()}, true
	}
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

func syntaxExpression(node *ast.Node) *SyntaxExpression {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindNumericLiteral:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "number", Text: node.Text()}
	case ast.KindStringLiteral:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "string", Text: node.Text()}
	case ast.KindTrueKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "bool", Text: "true"}
	case ast.KindFalseKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "bool", Text: "false"}
	case ast.KindIdentifier:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "identifier", Text: node.Text()}
	case ast.KindThisKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "identifier", Text: "this"}
	case ast.KindParenthesizedExpression:
		return syntaxExpression(node.Expression())
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		return &SyntaxExpression{
			Span:     sourceSpan(node),
			Kind:     "binary",
			Operator: binaryOperator(binary.OperatorToken.Kind.String()),
			Left:     syntaxExpression(binary.Left),
			Right:    syntaxExpression(binary.Right),
		}
	case ast.KindCallExpression:
		result := &SyntaxExpression{Span: sourceSpan(node), Kind: "call", Left: syntaxExpression(node.Expression())}
		for _, argument := range node.Arguments() {
			result.Arguments = append(result.Arguments, syntaxExpression(argument))
		}
		return result
	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		return &SyntaxExpression{
			Span:     sourceSpan(node),
			Kind:     "unary",
			Operator: prefixUnaryOperator(prefix.Operator),
			Left:     syntaxExpression(prefix.Operand),
		}
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return &SyntaxExpression{
			Span:      sourceSpan(node),
			Kind:      "conditional",
			Left:      syntaxExpression(conditional.Condition),
			WhenTrue:  syntaxExpression(conditional.WhenTrue),
			WhenFalse: syntaxExpression(conditional.WhenFalse),
		}
	case ast.KindNoSubstitutionTemplateLiteral:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "string", Text: node.Text()}
	case ast.KindTemplateExpression:
		template := node.AsTemplateExpression()
		result := &SyntaxExpression{Span: sourceSpan(node), Kind: "template"}
		if head := template.Head; head != nil && head.Text() != "" {
			result.Arguments = append(result.Arguments, &SyntaxExpression{Span: sourceSpan(head), Kind: "string", Text: head.Text()})
		}
		if spans := template.TemplateSpans; spans != nil {
			for _, spanNode := range spans.Nodes {
				span := spanNode.AsTemplateSpan()
				if span.Expression != nil {
					result.Arguments = append(result.Arguments, syntaxExpression(span.Expression))
				}
				if span.Literal != nil && span.Literal.Text() != "" {
					result.Arguments = append(result.Arguments, &SyntaxExpression{Span: sourceSpan(span.Literal), Kind: "string", Text: span.Literal.Text()})
				}
			}
		}
		return result
	case ast.KindArrayLiteralExpression:
		result := &SyntaxExpression{Span: sourceSpan(node), Kind: "array"}
		if elements := node.AsArrayLiteralExpression().Elements; elements != nil {
			for _, element := range elements.Nodes {
				result.Arguments = append(result.Arguments, syntaxExpression(element))
			}
		}
		return result
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		return &SyntaxExpression{
			Span:  sourceSpan(node),
			Kind:  "index",
			Left:  syntaxExpression(element.Expression),
			Right: syntaxExpression(element.ArgumentExpression),
		}
	case ast.KindPropertyAccessExpression:
		return &SyntaxExpression{
			Span: sourceSpan(node),
			Kind: "property",
			Text: node.Name().Text(),
			Left: syntaxExpression(node.Expression()),
		}
	case ast.KindNewExpression:
		newExpression := node.AsNewExpression()
		result := &SyntaxExpression{Span: sourceSpan(node), Kind: "new", Left: syntaxExpression(newExpression.Expression)}
		if arguments := newExpression.Arguments; arguments != nil {
			for _, argument := range arguments.Nodes {
				result.Arguments = append(result.Arguments, syntaxExpression(argument))
			}
		}
		return result
	default:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "unsupported", Text: node.Kind.String()}
	}
}

func sourceSpan(node *ast.Node) SourceSpan {
	if node == nil {
		return SourceSpan{}
	}
	return SourceSpan{Start: node.Pos(), Length: node.End() - node.Pos()}
}

func parameterSpan(node *ast.Node) SourceSpan {
	return sourceSpan(node)
}

func syntaxType(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindNumberKeyword:
		return "number"
	case ast.KindStringKeyword:
		return "string"
	case ast.KindBooleanKeyword:
		return "bool"
	case ast.KindVoidKeyword:
		return "void"
	case ast.KindArrayType:
		array := node.AsArrayTypeNode()
		return syntaxType(array.ElementType) + "[]"
	default:
		return node.Kind.String()
	}
}

func binaryOperator(kind string) string {
	kind = strings.TrimPrefix(kind, "Kind")
	operators := map[string]string{
		"PlusToken":                    "+",
		"MinusToken":                   "-",
		"AsteriskToken":                "*",
		"SlashToken":                   "/",
		"PercentToken":                 "%",
		"EqualsToken":                  "=",
		"PlusEqualsToken":              "+=",
		"MinusEqualsToken":             "-=",
		"EqualsEqualsToken":            "==",
		"EqualsEqualsEqualsToken":      "===",
		"ExclamationEqualsToken":       "!=",
		"ExclamationEqualsEqualsToken": "!==",
		"LessThanToken":                "<",
		"LessThanEqualsToken":          "<=",
		"GreaterThanToken":             ">",
		"GreaterThanEqualsToken":       ">=",
		"AmpersandAmpersandToken":      "&&",
		"BarBarToken":                  "||",
	}
	if operator, ok := operators[kind]; ok {
		return operator
	}
	return kind
}

func prefixUnaryOperator(kind ast.Kind) string {
	switch kind {
	case ast.KindExclamationToken:
		return "!"
	case ast.KindMinusToken:
		return "-"
	case ast.KindPlusToken:
		return "+"
	default:
		return kind.String()
	}
}

func statementCount(file *ast.SourceFile) int {
	if file.Statements == nil {
		return 0
	}
	return len(file.Statements.Nodes)
}

func isTypeScriptSource(fileName string) bool {
	for _, extension := range []string{".ts", ".tsx", ".mts", ".cts"} {
		if strings.HasSuffix(fileName, extension) {
			return true
		}
	}
	return false
}
