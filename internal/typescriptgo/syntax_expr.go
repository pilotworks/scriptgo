package typescriptgo

import (
	"strings"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/checker"
)

func syntaxExpression(node *ast.Node, chk *checker.Checker) *SyntaxExpression {
	if node == nil {
		return nil
	}
	expr := syntaxExpressionInner(node, chk)
	if expr != nil && expr.InferredType == "" && chk != nil {
		expr.InferredType = resolveInferredType(chk, node)
	}
	return expr
}

func syntaxExpressionInner(node *ast.Node, chk *checker.Checker) *SyntaxExpression {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindNumericLiteral:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "number", Text: node.Text(), InferredType: "number"}
	case ast.KindStringLiteral:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "string", Text: node.Text(), InferredType: "string"}
	case ast.KindNullKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "null", Text: "null", InferredType: "null"}
	case ast.KindUndefinedKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "undefined", Text: "undefined", InferredType: "undefined"}
	case ast.KindTrueKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "bool", Text: "true", InferredType: "bool"}
	case ast.KindFalseKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "bool", Text: "false", InferredType: "bool"}
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "identifier", Text: node.Text()}
	case ast.KindThisKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "identifier", Text: "this"}
	case ast.KindSuperKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "identifier", Text: "super"}
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
		return syntaxExpression(node.Expression(), chk)
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		return &SyntaxExpression{
			Span:     sourceSpan(node),
			Kind:     "binary",
			Operator: binaryOperator(binary.OperatorToken.Kind.String()),
			Left:     syntaxExpression(binary.Left, chk),
			Right:    syntaxExpression(binary.Right, chk),
		}
	case ast.KindCallExpression:
		callExpr := node.AsCallExpression()
		result := &SyntaxExpression{
			Span:          sourceSpan(node),
			Kind:          "call",
			Left:          syntaxExpression(callExpr.Expression, chk),
			TypeArguments: syntaxTypeArguments(callExpr.TypeArguments),
		}
		for _, argument := range node.Arguments() {
			result.Arguments = append(result.Arguments, syntaxExpression(argument, chk))
		}
		return result
	case ast.KindTypeOfExpression:
		return &SyntaxExpression{
			Span:         sourceSpan(node),
			Kind:         "typeof",
			Left:         syntaxExpression(node.Expression(), chk),
			InferredType: "string",
		}
	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		return &SyntaxExpression{
			Span:     sourceSpan(node),
			Kind:     "unary",
			Operator: prefixUnaryOperator(prefix.Operator),
			Left:     syntaxExpression(prefix.Operand, chk),
		}
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return &SyntaxExpression{
			Span:      sourceSpan(node),
			Kind:      "conditional",
			Left:      syntaxExpression(conditional.Condition, chk),
			WhenTrue:  syntaxExpression(conditional.WhenTrue, chk),
			WhenFalse: syntaxExpression(conditional.WhenFalse, chk),
		}
	case ast.KindNoSubstitutionTemplateLiteral:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "string", Text: node.Text(), InferredType: "string"}
	case ast.KindTemplateExpression:
		template := node.AsTemplateExpression()
		result := &SyntaxExpression{Span: sourceSpan(node), Kind: "template", InferredType: "string"}
		if head := template.Head; head != nil && head.Text() != "" {
			result.Arguments = append(result.Arguments, &SyntaxExpression{Span: sourceSpan(head), Kind: "string", Text: head.Text(), InferredType: "string"})
		}
		if spans := template.TemplateSpans; spans != nil {
			for _, spanNode := range spans.Nodes {
				span := spanNode.AsTemplateSpan()
				if span.Expression != nil {
					result.Arguments = append(result.Arguments, syntaxExpression(span.Expression, chk))
				}
				if span.Literal != nil && span.Literal.Text() != "" {
					result.Arguments = append(result.Arguments, &SyntaxExpression{Span: sourceSpan(span.Literal), Kind: "string", Text: span.Literal.Text(), InferredType: "string"})
				}
			}
		}
		return result
	case ast.KindArrayLiteralExpression:
		result := &SyntaxExpression{Span: sourceSpan(node), Kind: "array"}
		if elements := node.AsArrayLiteralExpression().Elements; elements != nil {
			for _, element := range elements.Nodes {
				result.Arguments = append(result.Arguments, syntaxExpression(element, chk))
			}
		}
		return result
	case ast.KindSpreadElement:
		spread := node.AsSpreadElement()
		return &SyntaxExpression{
			Span: sourceSpan(node),
			Kind: "spread",
			Left: syntaxExpression(spread.Expression, chk),
		}
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		kind := "index"
		if element.QuestionDotToken != nil {
			kind = "optional_index"
		}
		return &SyntaxExpression{
			Span:  sourceSpan(node),
			Kind:  kind,
			Left:  syntaxExpression(element.Expression, chk),
			Right: syntaxExpression(element.ArgumentExpression, chk),
		}
	case ast.KindPropertyAccessExpression:
		prop := node.AsPropertyAccessExpression()
		kind := "property"
		if prop.QuestionDotToken != nil {
			kind = "optional_property"
		}
		return &SyntaxExpression{
			Span: sourceSpan(node),
			Kind: kind,
			Text: node.Name().Text(),
			Left: syntaxExpression(node.Expression(), chk),
		}
	case ast.KindNewExpression:
		newExpression := node.AsNewExpression()
		result := &SyntaxExpression{
			Span:          sourceSpan(node),
			Kind:          "new",
			Left:          syntaxExpression(newExpression.Expression, chk),
			TypeArguments: syntaxTypeArguments(newExpression.TypeArguments),
		}
		if arguments := newExpression.Arguments; arguments != nil {
			for _, argument := range arguments.Nodes {
				result.Arguments = append(result.Arguments, syntaxExpression(argument, chk))
			}
		}
		return result
	case ast.KindObjectLiteralExpression:
		objLit := node.AsObjectLiteralExpression()
		result := &SyntaxExpression{Span: sourceSpan(node), Kind: "object_literal"}
		if properties := objLit.Properties; properties != nil {
			for _, propNode := range properties.Nodes {
				if propNode.Kind == ast.KindPropertyAssignment {
					prop := propNode.AsPropertyAssignment()
					result.Arguments = append(result.Arguments, &SyntaxExpression{
						Span:         sourceSpan(propNode),
						Kind:         "property_assignment",
						Text:         prop.Name().Text(),
						Left:         syntaxExpression(prop.Initializer, chk),
						InferredType: resolveInferredType(chk, propNode),
					})
				} else if propNode.Kind == ast.KindShorthandPropertyAssignment {
					prop := propNode.AsShorthandPropertyAssignment()
					name := prop.Name().Text()
					result.Arguments = append(result.Arguments, &SyntaxExpression{
						Span:         sourceSpan(propNode),
						Kind:         "property_assignment",
						Text:         name,
						Left:         &SyntaxExpression{Span: sourceSpan(propNode), Kind: "identifier", Text: name, InferredType: resolveInferredType(chk, propNode)},
						InferredType: resolveInferredType(chk, propNode),
					})
				}
			}
		}
		return result
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		span := sourceSpan(node)
		var name string
		if node.Name() != nil {
			name = node.Name().Text()
		}
		var params []SyntaxParameter
		for _, parameter := range node.Parameters() {
			pType := syntaxType(parameter.Type())
			inferredPType := resolveInferredType(chk, parameter.Name())
			if inferredPType == "" {
				inferredPType = resolveInferredType(chk, parameter)
			}
			params = append(params, SyntaxParameter{
				Span:         parameterSpan(parameter),
				Name:         parameter.Name().Text(),
				Type:         pType,
				InferredType: inferredPType,
				Rest:         parameter.AsParameterDeclaration().DotDotDotToken != nil,
				Initializer:  syntaxExpression(parameter.Initializer(), chk),
			})
		}
		var body []SyntaxStatement
		if b := node.Body(); b != nil {
			if b.Kind == ast.KindBlock {
				for _, statement := range b.Statements() {
					if converted, ok := syntaxStatement(statement, chk); ok {
						body = append(body, converted)
					}
				}
			} else {
				body = append(body, SyntaxStatement{
					Span:       sourceSpan(b),
					Kind:       "return",
					Expression: syntaxExpression(b, chk),
				})
			}
		}
		fnType := syntaxType(node.Type())
		inferredRetType := resolveFunctionReturnType(chk, node)
		if fnType == "" && inferredRetType != "" {
			fnType = inferredRetType
		}
		fnStmt := &SyntaxStatement{
			Span:           span,
			Kind:           "function",
			Name:           name,
			Type:           fnType,
			InferredType:   inferredRetType,
			TypeParameters: syntaxTypeParameters(node.TypeParameters()),
			Parameters:     params,
			Body:           body,
		}
		return &SyntaxExpression{
			Span:         span,
			Kind:         "arrow_function",
			Function:     fnStmt,
			InferredType: resolveInferredType(chk, node),
		}
	case ast.KindAwaitExpression:
		awaitNode := node.AsAwaitExpression()
		return &SyntaxExpression{
			Span: sourceSpan(node),
			Kind: "await",
			Left: syntaxExpression(awaitNode.Expression, chk),
		}
	default:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "unsupported", Text: node.Kind.String()}
	}
}

func binaryOperator(kind string) string {
	kind = strings.TrimPrefix(kind, "Kind")
	operators := map[string]string{
		"PlusToken":                                   "+",
		"MinusToken":                                  "-",
		"AsteriskToken":                               "*",
		"SlashToken":                                  "/",
		"PercentToken":                                "%",
		"AsteriskAsteriskToken":                       "**",
		"AmpersandToken":                              "&",
		"BarToken":                                    "|",
		"CaretToken":                                  "^",
		"LessThanLessThanToken":                       "<<",
		"GreaterThanGreaterThanToken":                 ">>",
		"GreaterThanGreaterThanGreaterThanToken":      ">>>",
		"EqualsToken":                                 "=",
		"PlusEqualsToken":                             "+=",
		"MinusEqualsToken":                            "-=",
		"AsteriskEqualsToken":                         "*=",
		"SlashEqualsToken":                            "/=",
		"PercentEqualsToken":                          "%=",
		"AsteriskAsteriskEqualsToken":                 "**=",
		"AmpersandEqualsToken":                        "&=",
		"BarEqualsToken":                              "|=",
		"CaretEqualsToken":                            "^=",
		"LessThanLessThanEqualsToken":                 "<<=",
		"GreaterThanGreaterThanEqualsToken":           ">>=",
		"GreaterThanGreaterThanGreaterThanEqualsToken": ">>>=",
		"AmpersandAmpersandEqualsToken":               "&&=",
		"BarBarEqualsToken":                           "||=",
		"QuestionQuestionEqualsToken":                 "??=",
		"EqualsEqualsToken":                           "==",
		"EqualsEqualsEqualsToken":                     "===",
		"ExclamationEqualsToken":                      "!=",
		"ExclamationEqualsEqualsToken":                "!==",
		"LessThanToken":                               "<",
		"LessThanEqualsToken":                         "<=",
		"GreaterThanToken":                            ">",
		"GreaterThanEqualsToken":                      ">=",
		"AmpersandAmpersandToken":                     "&&",
		"BarBarToken":                                 "||",
		"QuestionQuestionToken":                       "??",
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
	case ast.KindTildeToken:
		return "~"
	default:
		return kind.String()
	}
}
