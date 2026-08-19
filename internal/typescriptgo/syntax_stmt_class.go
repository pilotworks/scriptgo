package typescriptgo

import "github.com/microsoft/typescript-go/internal/ast"

func syntaxClassDeclaration(node *ast.Node, span SourceSpan) (SyntaxStatement, bool) {
	classDecl := node.AsClassDeclaration()
	class := &SyntaxClass{
		Span:       span,
		Name:       node.Name().Text(),
		IsAbstract: ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract),
	}
	if classDecl.HeritageClauses != nil {
		for _, clause := range classDecl.HeritageClauses.Nodes {
			hc := clause.AsHeritageClause()
			if hc == nil {
				continue
			}
			if hc.Token == ast.KindExtendsKeyword && hc.Types != nil {
				for _, t := range hc.Types.Nodes {
					if t.Kind == ast.KindExpressionWithTypeArguments {
						exprNode := t.AsExpressionWithTypeArguments()
						if exprNode != nil && exprNode.Expression != nil {
							class.Extends = exprNode.Expression.Text()
						}
					} else if t.Kind == ast.KindTypeReference {
						class.Extends = t.AsTypeReferenceNode().TypeName.Text()
					}
				}
			} else if hc.Token == ast.KindImplementsKeyword && hc.Types != nil {
				for _, t := range hc.Types.Nodes {
					if t.Kind == ast.KindExpressionWithTypeArguments {
						exprNode := t.AsExpressionWithTypeArguments()
						if exprNode != nil && exprNode.Expression != nil {
							class.Implements = append(class.Implements, exprNode.Expression.Text())
						}
					} else if t.Kind == ast.KindTypeReference {
						class.Implements = append(class.Implements, t.AsTypeReferenceNode().TypeName.Text())
					}
				}
			}
		}
	}
	for _, member := range node.Members() {
		switch member.Kind {
		case ast.KindPropertyDeclaration:
			property := member.AsPropertyDeclaration()
			class.Fields = append(class.Fields, SyntaxField{
				Span:        sourceSpan(member),
				Name:        property.Name().Text(),
				Type:        syntaxType(property.Type),
				Initializer: syntaxExpression(property.Initializer),
				IsStatic:    ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic),
				IsPrivate:   ast.HasSyntacticModifier(member, ast.ModifierFlagsPrivate),
				IsReadonly:  ast.HasSyntacticModifier(member, ast.ModifierFlagsReadonly),
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
				IsStatic:   ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic),
				IsAbstract: ast.HasSyntacticModifier(member, ast.ModifierFlagsAbstract),
				Kind:       "method",
			})
		case ast.KindGetAccessor:
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
				Body:       body,
				IsStatic:   ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic),
				Kind:       "get",
			})
		case ast.KindSetAccessor:
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
				Type:       "void",
				Parameters: params,
				Body:       body,
				IsStatic:   ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic),
				Kind:       "set",
			})
		default:
			class.Fields = append(class.Fields, SyntaxField{Span: sourceSpan(member), Name: member.Kind.String()})
		}
	}
	return SyntaxStatement{Span: span, Kind: "class", Name: class.Name, Class: class}, true
}
