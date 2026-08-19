package typescriptgo

import (
	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/checker"
)

func syntaxClassDeclaration(node *ast.Node, span SourceSpan, chk *checker.Checker) (SyntaxStatement, bool) {
	classDecl := node.AsClassDeclaration()
	class := &SyntaxClass{
		Span:           span,
		Name:           node.Name().Text(),
		TypeParameters: syntaxTypeParameters(node.TypeParameters()),
		IsAbstract:     ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract),
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
			fType := syntaxType(property.Type)
			inferredFType := resolveInferredType(chk, property.Name())
			if inferredFType == "" {
				inferredFType = resolveInferredType(chk, member)
			}
			class.Fields = append(class.Fields, SyntaxField{
				Span:         sourceSpan(member),
				Name:         property.Name().Text(),
				Type:         fType,
				InferredType: inferredFType,
				Initializer:  syntaxExpression(property.Initializer, chk),
				IsStatic:     ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic),
				IsPrivate:    ast.HasSyntacticModifier(member, ast.ModifierFlagsPrivate),
				IsReadonly:   ast.HasSyntacticModifier(member, ast.ModifierFlagsReadonly),
			})
		case ast.KindConstructor:
			var params []SyntaxParameter
			for _, p := range member.Parameters() {
				pType := syntaxType(p.Type())
				inferredPType := resolveInferredType(chk, p.Name())
				if inferredPType == "" {
					inferredPType = resolveInferredType(chk, p)
				}
				params = append(params, SyntaxParameter{
					Span:         parameterSpan(p),
					Name:         p.Name().Text(),
					Type:         pType,
					InferredType: inferredPType,
					Rest:         p.AsParameterDeclaration().DotDotDotToken != nil,
					Initializer:  syntaxExpression(p.Initializer(), chk),
				})
			}
			var body []SyntaxStatement
			if b := member.Body(); b != nil {
				for _, s := range b.Statements() {
					if converted, ok := syntaxStatement(s, chk); ok {
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
				pType := syntaxType(p.Type())
				inferredPType := resolveInferredType(chk, p.Name())
				if inferredPType == "" {
					inferredPType = resolveInferredType(chk, p)
				}
				params = append(params, SyntaxParameter{
					Span:         parameterSpan(p),
					Name:         p.Name().Text(),
					Type:         pType,
					InferredType: inferredPType,
					Rest:         p.AsParameterDeclaration().DotDotDotToken != nil,
					Initializer:  syntaxExpression(p.Initializer(), chk),
				})
			}
			var body []SyntaxStatement
			if b := member.Body(); b != nil {
				for _, s := range b.Statements() {
					if converted, ok := syntaxStatement(s, chk); ok {
						body = append(body, converted)
					}
				}
			}
			mType := syntaxType(member.Type())
			inferredMType := resolveFunctionReturnType(chk, member)
			if mType == "" && inferredMType != "" {
				mType = inferredMType
			}
			class.Methods = append(class.Methods, SyntaxMethod{
				Span:           sourceSpan(member),
				Name:           member.Name().Text(),
				Type:           mType,
				InferredType:   inferredMType,
				TypeParameters: syntaxTypeParameters(member.TypeParameters()),
				Parameters:     params,
				Body:           body,
				IsStatic:       ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic),
				IsAbstract:     ast.HasSyntacticModifier(member, ast.ModifierFlagsAbstract),
				Kind:           "method",
			})
		case ast.KindGetAccessor:
			var body []SyntaxStatement
			if b := member.Body(); b != nil {
				for _, s := range b.Statements() {
					if converted, ok := syntaxStatement(s, chk); ok {
						body = append(body, converted)
					}
				}
			}
			mType := syntaxType(member.Type())
			inferredMType := resolveFunctionReturnType(chk, member)
			if mType == "" && inferredMType != "" {
				mType = inferredMType
			}
			class.Methods = append(class.Methods, SyntaxMethod{
				Span:         sourceSpan(member),
				Name:         member.Name().Text(),
				Type:         mType,
				InferredType: inferredMType,
				Body:         body,
				IsStatic:     ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic),
				Kind:         "get",
			})
		case ast.KindSetAccessor:
			var params []SyntaxParameter
			for _, p := range member.Parameters() {
				pType := syntaxType(p.Type())
				inferredPType := resolveInferredType(chk, p.Name())
				if inferredPType == "" {
					inferredPType = resolveInferredType(chk, p)
				}
				params = append(params, SyntaxParameter{
					Span:         parameterSpan(p),
					Name:         p.Name().Text(),
					Type:         pType,
					InferredType: inferredPType,
					Rest:         p.AsParameterDeclaration().DotDotDotToken != nil,
					Initializer:  syntaxExpression(p.Initializer(), chk),
				})
			}
			var body []SyntaxStatement
			if b := member.Body(); b != nil {
				for _, s := range b.Statements() {
					if converted, ok := syntaxStatement(s, chk); ok {
						body = append(body, converted)
					}
				}
			}
			class.Methods = append(class.Methods, SyntaxMethod{
				Span:         sourceSpan(member),
				Name:         member.Name().Text(),
				Type:         "void",
				InferredType: "void",
				Parameters:   params,
				Body:         body,
				IsStatic:     ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic),
				Kind:         "set",
			})
		default:
			class.Fields = append(class.Fields, SyntaxField{Span: sourceSpan(member), Name: member.Kind.String()})
		}
	}
	return SyntaxStatement{Span: span, Kind: "class", Name: class.Name, Class: class}, true
}
