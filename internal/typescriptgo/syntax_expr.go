package typescriptgo

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
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
	case ast.KindBigIntLiteral:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "bigint", Text: strings.TrimSuffix(node.Text(), "n"), InferredType: "bigint"}
	case ast.KindRegularExpressionLiteral:
		text := node.Text()
		lastSlash := strings.LastIndex(text, "/")
		pattern := ""
		flags := ""
		if lastSlash > 0 {
			pattern = text[1:lastSlash]
			flags = text[lastSlash+1:]
		} else {
			pattern = text
		}
		return &SyntaxExpression{
			Span:         sourceSpan(node),
			Kind:         "regex",
			Text:         pattern,
			Operator:     flags,
			InferredType: "object:RegExp",
		}
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
		if node.Text() == "undefined" {
			return &SyntaxExpression{Span: sourceSpan(node), Kind: "undefined", Text: "undefined", InferredType: "undefined"}
		}
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "identifier", Text: node.Text()}
	case ast.KindThisKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "identifier", Text: "this"}
	case ast.KindSuperKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "identifier", Text: "super"}
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
		return syntaxExpression(node.Expression(), chk)
	case ast.KindVoidExpression:
		voidExpr := node.AsVoidExpression()
		inner := syntaxExpression(voidExpr.Expression, chk)
		return &SyntaxExpression{
			Span:         sourceSpan(node),
			Kind:         "unary",
			Operator:     "void",
			Left:         inner,
			InferredType: "undefined",
		}
	case ast.KindAsExpression:
		asExpr := node.AsAsExpression()
		targetType := syntaxType(asExpr.Type)
		inner := syntaxExpression(asExpr.Expression, chk)
		return &SyntaxExpression{
			Span:         sourceSpan(node),
			Kind:         "as",
			Left:         inner,
			Text:         targetType,
			InferredType: targetType,
		}
	case ast.KindTypeAssertionExpression:
		taExpr := node.AsTypeAssertion()
		targetType := syntaxType(taExpr.Type)
		inner := syntaxExpression(taExpr.Expression, chk)
		return &SyntaxExpression{
			Span:         sourceSpan(node),
			Kind:         "as",
			Left:         inner,
			Text:         targetType,
			InferredType: targetType,
		}
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
		kind := "call"
		if callExpr.QuestionDotToken != nil {
			kind = "optional_call"
		}
		result := &SyntaxExpression{
			Span:          sourceSpan(node),
			Kind:          kind,
			Left:          syntaxExpression(callExpr.Expression, chk),
			TypeArguments: syntaxTypeArguments(callExpr.TypeArguments),
		}
		for _, argument := range node.Arguments() {
			result.Arguments = append(result.Arguments, syntaxExpression(argument, chk))
		}
		return result
	case ast.KindTaggedTemplateExpression:
		tagged := node.AsTaggedTemplateExpression()
		result := &SyntaxExpression{
			Span:          sourceSpan(node),
			Kind:          "tagged_template",
			Left:          syntaxExpression(tagged.Tag, chk),
			TypeArguments: syntaxTypeArguments(tagged.TypeArguments),
		}
		stringsArray := &SyntaxExpression{
			Span:         sourceSpan(tagged.Template),
			Kind:         "array",
			InferredType: "string[]",
		}
		if tagged.Template.Kind == ast.KindNoSubstitutionTemplateLiteral {
			stringsArray.Arguments = append(stringsArray.Arguments, &SyntaxExpression{
				Span:         sourceSpan(tagged.Template),
				Kind:         "string",
				Text:         tagged.Template.Text(),
				InferredType: "string",
			})
			result.Arguments = append(result.Arguments, stringsArray)
		} else if tagged.Template.Kind == ast.KindTemplateExpression {
			tpl := tagged.Template.AsTemplateExpression()
			headText := ""
			if tpl.Head != nil {
				headText = tpl.Head.Text()
			}
			stringsArray.Arguments = append(stringsArray.Arguments, &SyntaxExpression{
				Span:         sourceSpan(tpl.Head),
				Kind:         "string",
				Text:         headText,
				InferredType: "string",
			})
			var rawValues []*SyntaxExpression
			if spans := tpl.TemplateSpans; spans != nil {
				for _, spanNode := range spans.Nodes {
					span := spanNode.AsTemplateSpan()
					if span.Expression != nil {
						rawValues = append(rawValues, syntaxExpression(span.Expression, chk))
					}
					litText := ""
					if span.Literal != nil {
						litText = span.Literal.Text()
					}
					stringsArray.Arguments = append(stringsArray.Arguments, &SyntaxExpression{
						Span:         sourceSpan(span.Literal),
						Kind:         "string",
						Text:         litText,
						InferredType: "string",
					})
				}
			}
			result.Arguments = append(result.Arguments, stringsArray)
			result.Arguments = append(result.Arguments, rawValues...)
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
	case ast.KindPostfixUnaryExpression:
		postfix := node.AsPostfixUnaryExpression()
		return &SyntaxExpression{
			Span:         sourceSpan(node),
			Kind:         "postfix_unary",
			Operator:     prefixUnaryOperator(postfix.Operator),
			Left:         syntaxExpression(postfix.Operand, chk),
			InferredType: "number",
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
		arrType := resolveInferredType(chk, node)
		elemType := ""
		if before, ok := strings.CutSuffix(arrType, "[]"); ok {
			elemType = before
		}
		if elements := node.AsArrayLiteralExpression().Elements; elements != nil {
			for _, element := range elements.Nodes {
				elemExpr := syntaxExpression(element, chk)
				if elemExpr != nil && elemType != "" && !strings.HasPrefix(elemType, "{") && (elemExpr.InferredType == "" || strings.HasPrefix(elemExpr.InferredType, "{")) {
					elemExpr.InferredType = elemType
				}
				result.Arguments = append(result.Arguments, elemExpr)
			}
		}
		result.InferredType = arrType
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
		result := &SyntaxExpression{Span: sourceSpan(node), Kind: "object_literal", InferredType: resolveInferredType(chk, node)}
		if properties := objLit.Properties; properties != nil {
			for _, propNode := range properties.Nodes {
				switch propNode.Kind {
				case ast.KindPropertyAssignment:
					prop := propNode.AsPropertyAssignment()
					result.Arguments = append(result.Arguments, &SyntaxExpression{
						Span:         sourceSpan(propNode),
						Kind:         "property_assignment",
						Text:         prop.Name().Text(),
						Left:         syntaxExpression(prop.Initializer, chk),
						InferredType: resolveInferredType(chk, propNode),
					})
				case ast.KindShorthandPropertyAssignment:
					prop := propNode.AsShorthandPropertyAssignment()
					name := prop.Name().Text()
					result.Arguments = append(result.Arguments, &SyntaxExpression{
						Span:         sourceSpan(propNode),
						Kind:         "property_assignment",
						Text:         name,
						Left:         &SyntaxExpression{Span: sourceSpan(propNode), Kind: "identifier", Text: name, InferredType: resolveInferredType(chk, propNode)},
						InferredType: resolveInferredType(chk, propNode),
					})
				case ast.KindSpreadAssignment:
					spread := propNode.AsSpreadAssignment()
					result.Arguments = append(result.Arguments, &SyntaxExpression{
						Span:         sourceSpan(propNode),
						Kind:         "spread",
						Left:         syntaxExpression(spread.Expression, chk),
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
		var bindingStmts []SyntaxStatement
		for pIdx, parameter := range node.Parameters() {
			pType := syntaxType(parameter.Type())
			inferredPType := resolveInferredType(chk, parameter.Name())
			if inferredPType == "" {
				inferredPType = resolveInferredType(chk, parameter)
			}
			nameNode := parameter.Name()
			var pName string
			if nameNode != nil && nameNode.Kind == ast.KindIdentifier {
				pName = nameNode.Text()
			} else if nameNode != nil && (nameNode.Kind == ast.KindArrayBindingPattern || nameNode.Kind == ast.KindObjectBindingPattern) {
				pName = fmt.Sprintf("__param_%d", pIdx)
				c := 0
				bindingStmts = append(bindingStmts, flattenDestructuring(nameNode, &SyntaxExpression{
					Span: sourceSpan(nameNode),
					Kind: "identifier",
					Text: pName,
				}, chk, &c)...)
			} else if nameNode != nil {
				pName = fmt.Sprintf("__param_%d", pIdx)
			}
			params = append(params, SyntaxParameter{
				Span:         parameterSpan(parameter),
				Name:         pName,
				Type:         pType,
				InferredType: inferredPType,
				Rest:         parameter.AsParameterDeclaration().DotDotDotToken != nil,
				Optional:     parameter.AsParameterDeclaration().QuestionToken != nil,
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
		if len(bindingStmts) > 0 {
			body = append(bindingStmts, body...)
		}
		fnType := syntaxType(node.Type())
		inferredRetType := resolveFunctionReturnType(chk, node)
		if fnType == "" && inferredRetType != "" {
			fnType = inferredRetType
		}
		isGen := false
		if node.Kind == ast.KindFunctionExpression && node.AsFunctionExpression() != nil && node.AsFunctionExpression().AsteriskToken != nil {
			isGen = true
		}
		isAsync := ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
		kind := "function"
		if isGen && isAsync {
			kind = "async_generator_function"
		} else if isGen {
			kind = "generator_function"
		} else if isAsync {
			kind = "async_function"
		}
		fnStmt := &SyntaxStatement{
			Span:           span,
			Kind:           kind,
			Name:           name,
			Type:           fnType,
			InferredType:   inferredRetType,
			TypeParameters: syntaxTypeParameters(node.TypeParameters()),
			Parameters:     params,
			Body:           body,
			IsGenerator:    isGen,
			IsAsync:        isAsync,
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
	case ast.KindYieldExpression:
		yieldNode := node.AsYieldExpression()
		kind := "yield"
		if yieldNode.AsteriskToken != nil {
			kind = "yield_star"
		}
		var inner *SyntaxExpression
		if yieldNode.Expression != nil {
			inner = syntaxExpression(yieldNode.Expression, chk)
		}
		return &SyntaxExpression{
			Span: sourceSpan(node),
			Kind: kind,
			Left: inner,
		}
	default:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "unsupported", Text: node.Kind.String()}
	}
}

func binaryOperator(kind string) string {
	kind = strings.TrimPrefix(kind, "Kind")
	operators := map[string]string{
		"PlusToken":                              "+",
		"MinusToken":                             "-",
		"AsteriskToken":                          "*",
		"SlashToken":                             "/",
		"PercentToken":                           "%",
		"AsteriskAsteriskToken":                  "**",
		"AmpersandToken":                         "&",
		"BarToken":                               "|",
		"CaretToken":                             "^",
		"LessThanLessThanToken":                  "<<",
		"GreaterThanGreaterThanToken":            ">>",
		"GreaterThanGreaterThanGreaterThanToken": ">>>",
		"EqualsToken":                            "=",
		"PlusEqualsToken":                        "+=",
		"MinusEqualsToken":                       "-=",
		"AsteriskEqualsToken":                    "*=",
		"SlashEqualsToken":                       "/=",
		"PercentEqualsToken":                     "%=",
		"AsteriskAsteriskEqualsToken":            "**=",
		"AmpersandEqualsToken":                   "&=",
		"BarEqualsToken":                         "|=",
		"CaretEqualsToken":                       "^=",
		"LessThanLessThanEqualsToken":            "<<=",
		"GreaterThanGreaterThanEqualsToken":      ">>=",
		"GreaterThanGreaterThanGreaterThanEqualsToken": ">>>=",
		"AmpersandAmpersandEqualsToken":                "&&=",
		"BarBarEqualsToken":                            "||=",
		"QuestionQuestionEqualsToken":                  "??=",
		"EqualsEqualsToken":                            "==",
		"EqualsEqualsEqualsToken":                      "===",
		"ExclamationEqualsToken":                       "!=",
		"ExclamationEqualsEqualsToken":                 "!==",
		"LessThanToken":                                "<",
		"LessThanEqualsToken":                          "<=",
		"GreaterThanToken":                             ">",
		"GreaterThanEqualsToken":                       ">=",
		"AmpersandAmpersandToken":                      "&&",
		"BarBarToken":                                  "||",
		"QuestionQuestionToken":                        "??",
		"InstanceOfKeyword":                            "instanceof",
		"InKeyword":                                    "in",
		"CommaToken":                                   ",",
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
	case ast.KindPlusPlusToken:
		return "++"
	case ast.KindMinusMinusToken:
		return "--"
	default:
		return kind.String()
	}
}
