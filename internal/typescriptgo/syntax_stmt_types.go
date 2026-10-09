package typescriptgo

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

func syntaxInterfaceDeclaration(node *ast.Node, chk *checker.Checker, span SourceSpan) (SyntaxStatement, bool) {
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
			property := member.AsPropertySignatureDeclaration()
			fields = append(fields, SyntaxField{
				Span:         sourceSpan(member),
				Name:         pName,
				Type:         fType,
				InferredType: inferredFType,
				Optional:     property != nil && property.PostfixToken != nil,
			})
		case ast.KindConstructSignature:
			mType := syntaxType(member.Type())
			var params []SyntaxParameter
			for _, p := range valueParameters(member.Parameters()) {
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
			for _, p := range valueParameters(member.Parameters()) {
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
}

func syntaxTypeAliasDeclaration(node *ast.Node, chk *checker.Checker, span SourceSpan) (SyntaxStatement, bool) {
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
					pName := syntaxMemberName(member.Name())
					property := member.AsPropertySignatureDeclaration()
					fields = append(fields, SyntaxField{
						Span:         sourceSpan(member),
						Name:         pName,
						Type:         syntaxType(member.Type()),
						InferredType: resolveInferredType(chk, member),
						Optional:     property != nil && property.PostfixToken != nil,
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
								pName := syntaxMemberName(member.Name())
								property := member.AsPropertySignatureDeclaration()
								fields = append(fields, SyntaxField{
									Span:         sourceSpan(member),
									Name:         pName,
									Type:         syntaxType(member.Type()),
									InferredType: resolveInferredType(chk, member),
									Optional:     property != nil && property.PostfixToken != nil,
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
}
