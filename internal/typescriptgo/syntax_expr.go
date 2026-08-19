package typescriptgo

import (
	"strings"

	"github.com/microsoft/typescript-go/internal/ast"
)

func syntaxExpression(node *ast.Node) *SyntaxExpression {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindNumericLiteral:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "number", Text: node.Text()}
	case ast.KindStringLiteral:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "string", Text: node.Text()}
	case ast.KindNullKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "null", Text: "null"}
	case ast.KindUndefinedKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "undefined", Text: "undefined"}
	case ast.KindTrueKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "bool", Text: "true"}
	case ast.KindFalseKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "bool", Text: "false"}
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "identifier", Text: node.Text()}
	case ast.KindThisKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "identifier", Text: "this"}
	case ast.KindSuperKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "identifier", Text: "super"}
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
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
		callExpr := node.AsCallExpression()
		result := &SyntaxExpression{
			Span:          sourceSpan(node),
			Kind:          "call",
			Left:          syntaxExpression(callExpr.Expression),
			TypeArguments: syntaxTypeArguments(callExpr.TypeArguments),
		}
		for _, argument := range node.Arguments() {
			result.Arguments = append(result.Arguments, syntaxExpression(argument))
		}
		return result
	case ast.KindTypeOfExpression:
		return &SyntaxExpression{
			Span: sourceSpan(node),
			Kind: "typeof",
			Left: syntaxExpression(node.Expression()),
		}
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
	case ast.KindSpreadElement:
		spread := node.AsSpreadElement()
		return &SyntaxExpression{
			Span: sourceSpan(node),
			Kind: "spread",
			Left: syntaxExpression(spread.Expression),
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
			Left:  syntaxExpression(element.Expression),
			Right: syntaxExpression(element.ArgumentExpression),
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
			Left: syntaxExpression(node.Expression()),
		}
	case ast.KindNewExpression:
		newExpression := node.AsNewExpression()
		result := &SyntaxExpression{
			Span:          sourceSpan(node),
			Kind:          "new",
			Left:          syntaxExpression(newExpression.Expression),
			TypeArguments: syntaxTypeArguments(newExpression.TypeArguments),
		}
		if arguments := newExpression.Arguments; arguments != nil {
			for _, argument := range arguments.Nodes {
				result.Arguments = append(result.Arguments, syntaxExpression(argument))
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
						Span: sourceSpan(propNode),
						Kind: "property_assignment",
						Text: prop.Name().Text(),
						Left: syntaxExpression(prop.Initializer),
					})
				} else if propNode.Kind == ast.KindShorthandPropertyAssignment {
					prop := propNode.AsShorthandPropertyAssignment()
					name := prop.Name().Text()
					result.Arguments = append(result.Arguments, &SyntaxExpression{
						Span: sourceSpan(propNode),
						Kind: "property_assignment",
						Text: name,
						Left: &SyntaxExpression{Span: sourceSpan(propNode), Kind: "identifier", Text: name},
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
			params = append(params, SyntaxParameter{
				Span:        parameterSpan(parameter),
				Name:        parameter.Name().Text(),
				Type:        syntaxType(parameter.Type()),
				Rest:        parameter.AsParameterDeclaration().DotDotDotToken != nil,
				Initializer: syntaxExpression(parameter.Initializer()),
			})
		}
		var body []SyntaxStatement
		if b := node.Body(); b != nil {
			if b.Kind == ast.KindBlock {
				for _, statement := range b.Statements() {
					if converted, ok := syntaxStatement(statement); ok {
						body = append(body, converted)
					}
				}
			} else {
				body = append(body, SyntaxStatement{
					Span:       sourceSpan(b),
					Kind:       "return",
					Expression: syntaxExpression(b),
				})
			}
		}
		fnStmt := &SyntaxStatement{
			Span:           span,
			Kind:           "function",
			Name:           name,
			Type:           syntaxType(node.Type()),
			TypeParameters: syntaxTypeParameters(node.TypeParameters()),
			Parameters:     params,
			Body:           body,
		}
		return &SyntaxExpression{
			Span:     span,
			Kind:     "arrow_function",
			Function: fnStmt,
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
