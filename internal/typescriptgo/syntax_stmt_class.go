package typescriptgo

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

func syntaxClassDeclaration(node *ast.Node, span SourceSpan, chk *checker.Checker) (SyntaxStatement, bool) {
	name := ""
	if node.Name() != nil {
		name = node.Name().Text()
	}
	var heritageClauses *ast.NodeList
	if node.Kind == ast.KindClassDeclaration {
		classDecl := node.AsClassDeclaration()
		if classDecl != nil {
			heritageClauses = classDecl.HeritageClauses
		}
	} else if node.Kind == ast.KindClassExpression {
		clsExpr := node.AsClassExpression()
		if clsExpr != nil {
			heritageClauses = clsExpr.HeritageClauses
		}
	}
	class := &SyntaxClass{
		Span:           span,
		Name:           name,
		TypeParameters: syntaxTypeParameters(node.TypeParameters()),
		IsAbstract:     ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract),
		Decorators:     syntaxDecorators(node, chk, name, nil, ""),
	}
	if heritageClauses != nil {
		for _, clause := range heritageClauses.Nodes {
			hc := clause.AsHeritageClause()
			if hc == nil {
				continue
			}
			if hc.Token == ast.KindExtendsKeyword && hc.Types != nil {
				for _, t := range hc.Types.Nodes {
					switch t.Kind {
					case ast.KindExpressionWithTypeArguments:
						exprNode := t.AsExpressionWithTypeArguments()
						if exprNode != nil && exprNode.Expression != nil {
							class.Extends = syntaxMemberName(exprNode.Expression)
						}
					case ast.KindTypeReference:
						ref := t.AsTypeReferenceNode()
						if ref != nil && ref.TypeName != nil {
							class.Extends = syntaxMemberName(ref.TypeName)
						}
					}
				}
			} else if hc.Token == ast.KindImplementsKeyword && hc.Types != nil {
				for _, t := range hc.Types.Nodes {
					if t.Kind == ast.KindExpressionWithTypeArguments {
						exprNode := t.AsExpressionWithTypeArguments()
						if exprNode != nil && exprNode.Expression != nil {
							class.Implements = append(class.Implements, syntaxMemberName(exprNode.Expression))
						}
					} else if t.Kind == ast.KindTypeReference {
						ref := t.AsTypeReferenceNode()
						if ref != nil && ref.TypeName != nil {
							class.Implements = append(class.Implements, syntaxMemberName(ref.TypeName))
						}
					}
				}
			}
		}
	}
	for _, member := range node.Members() {
		switch member.Kind {
		case ast.KindPropertyDeclaration:
			property := member.AsPropertyDeclaration()
			name := syntaxMemberName(property.Name())
			fType := syntaxType(property.Type)
			inferredFType := resolveInferredType(chk, property.Name())
			if inferredFType == "" {
				inferredFType = resolveInferredType(chk, member)
			}
			if fType == "" && inferredFType != "" {
				fType = inferredFType
			}
			isStatic := ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic)
			isAccessor := ast.HasSyntacticModifier(member, ast.ModifierFlagsAccessor)
			fDecs := syntaxDecorators(member, chk, fType, nil, "")
			if isAccessor {
				storageName := "__accessor_storage_" + name
				class.Fields = append(class.Fields, SyntaxField{
					Span:         sourceSpan(member),
					Name:         storageName,
					Type:         fType,
					InferredType: inferredFType,
					Initializer:  syntaxExpression(property.Initializer, chk),
					Decorators:   fDecs,
					IsStatic:     isStatic,
					IsPrivate:    true,
				})
				getterBody := []SyntaxStatement{
					{
						Span: sourceSpan(member),
						Kind: "return",
						Expression: &SyntaxExpression{
							Span: sourceSpan(member),
							Kind: "property",
							Text: storageName,
							Left: &SyntaxExpression{Span: sourceSpan(member), Kind: "identifier", Text: "this"},
						},
					},
				}
				class.Methods = append(class.Methods, SyntaxMethod{
					Span:         sourceSpan(member),
					Name:         name,
					Type:         fType,
					InferredType: inferredFType,
					Decorators:   fDecs,
					Body:         getterBody,
					IsStatic:     isStatic,
					Kind:         "get",
				})
				setterBody := []SyntaxStatement{
					{
						Span: sourceSpan(member),
						Kind: "field_set",
						Name: storageName,
						Left: &SyntaxExpression{
							Span: sourceSpan(member),
							Kind: "identifier",
							Text: "this",
						},
						Expression: &SyntaxExpression{
							Span: sourceSpan(member),
							Kind: "identifier",
							Text: "value",
						},
					},
				}
				class.Methods = append(class.Methods, SyntaxMethod{
					Span:         sourceSpan(member),
					Name:         name,
					Type:         "void",
					InferredType: "void",
					Parameters: []SyntaxParameter{
						{
							Span:         sourceSpan(member),
							Name:         "value",
							Type:         fType,
							InferredType: inferredFType,
						},
					},
					Decorators: fDecs,
					Body:       setterBody,
					IsStatic:   isStatic,
					Kind:       "set",
				})
			} else {
				class.Fields = append(class.Fields, SyntaxField{
					Span:         sourceSpan(member),
					Name:         syntaxMemberName(property.Name()),
					Type:         fType,
					InferredType: inferredFType,
					Initializer:  syntaxExpression(property.Initializer, chk),
					Decorators:   fDecs,
					IsStatic:     ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic),
					IsPrivate:    ast.HasSyntacticModifier(member, ast.ModifierFlagsPrivate),
					IsReadonly:   ast.HasSyntacticModifier(member, ast.ModifierFlagsReadonly),
				})
			}
		case ast.KindConstructor:
			var params []SyntaxParameter
			var paramPropStmts []SyntaxStatement
			var bindingStmts []SyntaxStatement
			for pIdx, p := range member.Parameters() {
				pType := syntaxType(p.Type())
				inferredPType := resolveInferredType(chk, p.Name())
				if inferredPType == "" {
					inferredPType = resolveInferredType(chk, p)
				}
				if pType == "" && inferredPType != "" {
					pType = inferredPType
				}
				pName, binds := extractParameterBinding(p, pIdx, chk)
				if len(binds) > 0 {
					bindingStmts = append(bindingStmts, binds...)
				}
				pNode := p.AsNode()
				pDecs := syntaxDecorators(pNode, chk, pType, nil, "")
				params = append(params, SyntaxParameter{
					Span:         parameterSpan(p),
					Name:         pName,
					Type:         pType,
					InferredType: inferredPType,
					Decorators:   pDecs,
					Rest:         p.AsParameterDeclaration().DotDotDotToken != nil,
					Optional:     p.AsParameterDeclaration().QuestionToken != nil,
					Initializer:  syntaxExpression(p.Initializer(), chk),
				})
				isParamProp := ast.HasSyntacticModifier(p, ast.ModifierFlagsPublic|ast.ModifierFlagsPrivate|ast.ModifierFlagsProtected|ast.ModifierFlagsReadonly)
				if isParamProp && pName != "" {
					class.Fields = append(class.Fields, SyntaxField{
						Span:         parameterSpan(p),
						Name:         pName,
						Type:         pType,
						InferredType: inferredPType,
						Decorators:   pDecs,
						IsPrivate:    ast.HasSyntacticModifier(p, ast.ModifierFlagsPrivate),
						IsReadonly:   ast.HasSyntacticModifier(p, ast.ModifierFlagsReadonly),
					})
					paramPropStmts = append(paramPropStmts, SyntaxStatement{
						Span: parameterSpan(p),
						Kind: "field_set",
						Left: &SyntaxExpression{
							Span: parameterSpan(p),
							Kind: "identifier",
							Text: "this",
						},
						Name: pName,
						Expression: &SyntaxExpression{
							Span: parameterSpan(p),
							Kind: "identifier",
							Text: pName,
						},
					})
				}
			}
			var body []SyntaxStatement
			if b := member.Body(); b != nil {
				for _, s := range b.Statements() {
					if converted, ok := syntaxStatement(s, chk); ok {
						body = append(body, converted)
					}
				}
			}
			if len(bindingStmts) > 0 {
				insertIdx := 0
				if len(body) > 0 && body[0].Kind == "expression" && body[0].Expression != nil && body[0].Expression.Kind == "call" && body[0].Expression.Left != nil && body[0].Expression.Left.Text == "super" {
					insertIdx = 1
				}
				newBody := make([]SyntaxStatement, 0, len(body)+len(bindingStmts))
				newBody = append(newBody, body[:insertIdx]...)
				newBody = append(newBody, bindingStmts...)
				newBody = append(newBody, body[insertIdx:]...)
				body = newBody
			}
			if len(paramPropStmts) > 0 {
				insertIdx := 0
				if len(body) > 0 && body[0].Kind == "expression" && body[0].Expression != nil && body[0].Expression.Kind == "call" && body[0].Expression.Left != nil && body[0].Expression.Left.Text == "super" {
					insertIdx = 1
				}
				newBody := make([]SyntaxStatement, 0, len(body)+len(paramPropStmts))
				newBody = append(newBody, body[:insertIdx]...)
				newBody = append(newBody, paramPropStmts...)
				newBody = append(newBody, body[insertIdx:]...)
				body = newBody
			}
			class.Constructor = &SyntaxConstructor{
				Span:       sourceSpan(member),
				Parameters: params,
				Body:       body,
			}
		case ast.KindMethodDeclaration:
			var params []SyntaxParameter
			var pTypes []string
			var bindingStmts []SyntaxStatement
			for pIdx, p := range member.Parameters() {
				pType := syntaxType(p.Type())
				inferredPType := resolveInferredType(chk, p.Name())
				if inferredPType == "" {
					inferredPType = resolveInferredType(chk, p)
				}
				if pType == "" && inferredPType != "" {
					pType = inferredPType
				}
				pTypes = append(pTypes, pType)
				pName, binds := extractParameterBinding(p, pIdx, chk)
				if len(binds) > 0 {
					bindingStmts = append(bindingStmts, binds...)
				}
				pNode := p.AsNode()
				pDecs := syntaxDecorators(pNode, chk, pType, nil, "")
				params = append(params, SyntaxParameter{
					Span:         parameterSpan(p),
					Name:         pName,
					Type:         pType,
					InferredType: inferredPType,
					Decorators:   pDecs,
					Rest:         p.AsParameterDeclaration().DotDotDotToken != nil,
					Optional:     p.AsParameterDeclaration().QuestionToken != nil,
					Initializer:  syntaxExpression(p.Initializer(), chk),
				})
			}
			var body []SyntaxStatement
			if b := member.Body(); b != nil {
				body = []SyntaxStatement{}
				if len(bindingStmts) > 0 {
					body = append(body, bindingStmts...)
				}
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
			isGen := member.BodyData() != nil && member.BodyData().AsteriskToken != nil
			isAsync := ast.HasSyntacticModifier(member, ast.ModifierFlagsAsync)
			mDecs := syntaxDecorators(member, chk, "", pTypes, mType)
			class.Methods = append(class.Methods, SyntaxMethod{
				Span:           sourceSpan(member),
				Name:           syntaxMemberName(member.Name()),
				Type:           mType,
				InferredType:   inferredMType,
				TypeParameters: syntaxTypeParameters(member.TypeParameters()),
				Parameters:     params,
				Decorators:     mDecs,
				Body:           body,
				IsStatic:       ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic),
				IsAbstract:     ast.HasSyntacticModifier(member, ast.ModifierFlagsAbstract),
				Kind:           "method",
				IsGenerator:    isGen,
				IsAsync:        isAsync,
			})
		case ast.KindGetAccessor:
			var body []SyntaxStatement
			if b := member.Body(); b != nil {
				body = []SyntaxStatement{}
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
			mDecs := syntaxDecorators(member, chk, mType, nil, "")
			class.Methods = append(class.Methods, SyntaxMethod{
				Span:         sourceSpan(member),
				Name:         syntaxMemberName(member.Name()),
				Type:         mType,
				InferredType: inferredMType,
				Decorators:   mDecs,
				Body:         body,
				IsStatic:     ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic),
				Kind:         "get",
			})
		case ast.KindSetAccessor:
			var params []SyntaxParameter
			var pTypes []string
			var bindingStmts []SyntaxStatement
			for pIdx, p := range member.Parameters() {
				pType := syntaxType(p.Type())
				inferredPType := resolveInferredType(chk, p.Name())
				if inferredPType == "" {
					inferredPType = resolveInferredType(chk, p)
				}
				if pType == "" && inferredPType != "" {
					pType = inferredPType
				}
				pTypes = append(pTypes, pType)
				pName, binds := extractParameterBinding(p, pIdx, chk)
				if len(binds) > 0 {
					bindingStmts = append(bindingStmts, binds...)
				}
				pNode := p.AsNode()
				pDecs := syntaxDecorators(pNode, chk, pType, nil, "")
				params = append(params, SyntaxParameter{
					Span:         parameterSpan(p),
					Name:         pName,
					Type:         pType,
					InferredType: inferredPType,
					Decorators:   pDecs,
					Rest:         p.AsParameterDeclaration().DotDotDotToken != nil,
					Optional:     p.AsParameterDeclaration().QuestionToken != nil,
					Initializer:  syntaxExpression(p.Initializer(), chk),
				})
			}
			var body []SyntaxStatement
			if b := member.Body(); b != nil {
				body = []SyntaxStatement{}
				if len(bindingStmts) > 0 {
					body = append(body, bindingStmts...)
				}
				for _, s := range b.Statements() {
					if converted, ok := syntaxStatement(s, chk); ok {
						body = append(body, converted)
					}
				}
			}
			mDecs := syntaxDecorators(member, chk, "", pTypes, "void")
			class.Methods = append(class.Methods, SyntaxMethod{
				Span:         sourceSpan(member),
				Name:         syntaxMemberName(member.Name()),
				Type:         "void",
				InferredType: "void",
				Parameters:   params,
				Decorators:   mDecs,
				Body:         body,
				IsStatic:     ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic),
				Kind:         "set",
			})
		case ast.KindClassStaticBlockDeclaration:
			staticBlock := member.AsClassStaticBlockDeclaration()
			var body []SyntaxStatement
			if staticBlock.Body != nil {
				for _, s := range staticBlock.Body.Statements() {
					if converted, ok := syntaxStatement(s, chk); ok {
						converted = replaceThisWithClassStmt(converted, class.Name)
						body = append(body, converted)
					}
				}
			}
			class.StaticBlocks = append(class.StaticBlocks, body)
		default:
			class.Fields = append(class.Fields, SyntaxField{Span: sourceSpan(member), Name: member.Kind.String()})
		}
	}
	return SyntaxStatement{Span: span, Kind: "class", Name: class.Name, Class: class}, true
}

func syntaxDecorators(node *ast.Node, chk *checker.Checker, designType string, paramTypes []string, returnType string) []SyntaxDecorator {
	if node == nil {
		return nil
	}
	var decNodes []*ast.Node
	if mods := node.Modifiers(); mods != nil {
		for _, mod := range mods.Nodes {
			if mod.Kind == ast.KindDecorator {
				decNodes = append(decNodes, mod)
			}
		}
	}
	if len(decNodes) == 0 && node.Decorators() != nil {
		decNodes = append(decNodes, node.Decorators()...)
	}
	if len(decNodes) == 0 {
		return nil
	}
	var result []SyntaxDecorator
	for _, d := range decNodes {
		dec := d.AsDecorator()
		if dec == nil {
			continue
		}
		expr := syntaxExpression(dec.Expression, chk)
		name := ""
		var args []*SyntaxExpression
		if expr != nil {
			if expr.Kind == "identifier" {
				name = expr.Text
			} else if expr.Kind == "call" {
				if expr.Left != nil {
					name = expr.Left.Text
				}
				args = expr.Arguments
			} else if expr.Kind == "property" {
				name = expr.Text
			}
		}
		result = append(result, SyntaxDecorator{
			Span:       sourceSpan(d),
			Name:       name,
			Expression: expr,
			Arguments:  args,
			DesignType: designType,
			ParamTypes: paramTypes,
			ReturnType: returnType,
		})
	}
	return result
}

func replaceThisWithClassStmt(stmt SyntaxStatement, className string) SyntaxStatement {
	if stmt.Left != nil {
		stmt.Left = replaceThisWithClassExpr(stmt.Left, className)
	}
	if stmt.Right != nil {
		stmt.Right = replaceThisWithClassExpr(stmt.Right, className)
	}
	if stmt.Expression != nil {
		stmt.Expression = replaceThisWithClassExpr(stmt.Expression, className)
	}
	for i := range stmt.Body {
		stmt.Body[i] = replaceThisWithClassStmt(stmt.Body[i], className)
	}
	for i := range stmt.Then {
		stmt.Then[i] = replaceThisWithClassStmt(stmt.Then[i], className)
	}
	for i := range stmt.Else {
		stmt.Else[i] = replaceThisWithClassStmt(stmt.Else[i], className)
	}
	for i := range stmt.Catch {
		stmt.Catch[i] = replaceThisWithClassStmt(stmt.Catch[i], className)
	}
	for i := range stmt.Finally {
		stmt.Finally[i] = replaceThisWithClassStmt(stmt.Finally[i], className)
	}
	for i := range stmt.Step {
		stmt.Step[i] = replaceThisWithClassStmt(stmt.Step[i], className)
	}
	for i := range stmt.Cases {
		if stmt.Cases[i].Expression != nil {
			stmt.Cases[i].Expression = replaceThisWithClassExpr(stmt.Cases[i].Expression, className)
		}
		for j := range stmt.Cases[i].Statements {
			stmt.Cases[i].Statements[j] = replaceThisWithClassStmt(stmt.Cases[i].Statements[j], className)
		}
	}
	return stmt
}

func replaceThisWithClassExpr(expr *SyntaxExpression, className string) *SyntaxExpression {
	if expr == nil {
		return nil
	}
	if expr.Kind == "identifier" && expr.Text == "this" {
		expr.Text = className
	}
	if expr.Left != nil {
		expr.Left = replaceThisWithClassExpr(expr.Left, className)
	}
	if expr.Right != nil {
		expr.Right = replaceThisWithClassExpr(expr.Right, className)
	}
	if expr.WhenTrue != nil {
		expr.WhenTrue = replaceThisWithClassExpr(expr.WhenTrue, className)
	}
	if expr.WhenFalse != nil {
		expr.WhenFalse = replaceThisWithClassExpr(expr.WhenFalse, className)
	}
	for i := range expr.Arguments {
		expr.Arguments[i] = replaceThisWithClassExpr(expr.Arguments[i], className)
	}
	return expr
}

func syntaxMemberName(nameNode *ast.Node) string {
	if nameNode == nil {
		return ""
	}
	switch nameNode.Kind {
	case ast.KindComputedPropertyName:
		if expr := nameNode.Expression(); expr != nil {
			if expr.Kind == ast.KindPropertyAccessExpression {
				return syntaxMemberName(expr.Expression()) + "." + syntaxMemberName(expr.Name())
			} else if expr.Kind == ast.KindIdentifier {
				return expr.Text()
			}
		}
		return "[computed]"
	case ast.KindPropertyAccessExpression:
		return syntaxMemberName(nameNode.Expression()) + "." + syntaxMemberName(nameNode.Name())
	case ast.KindQualifiedName:
		qn := nameNode.AsQualifiedName()
		if qn != nil {
			return syntaxMemberName(qn.Left) + "." + syntaxMemberName(qn.Right)
		}
		return ""
	case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral:
		return nameNode.Text()
	default:
		return ""
	}
}

