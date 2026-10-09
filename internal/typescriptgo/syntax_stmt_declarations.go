package typescriptgo

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

func syntaxFunctionDeclaration(node *ast.Node, chk *checker.Checker, span SourceSpan) (SyntaxStatement, bool) {
	fnType := syntaxType(node.Type())
	inferredRetType := resolveFunctionReturnType(chk, node)
	if fnType == "" && inferredRetType != "" && inferredRetType != "any" && inferredRetType != "unknown" {
		fnType = inferredRetType
	}
	fnDecl := node.AsFunctionDeclaration()
	// A generator is a function* declaration; a plain function may return a
	// Generator (one made by a generator it calls).
	isGen := (node.BodyData() != nil && node.BodyData().AsteriskToken != nil) || (fnDecl != nil && fnDecl.AsteriskToken != nil)
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
		DefaultExport:  ast.HasSyntacticModifier(node, ast.ModifierFlagsDefault),
		Type:           fnType,
		InferredType:   inferredRetType,
		TypeParameters: syntaxTypeParameters(node.TypeParameters()),
		IsGenerator:    isGen,
		IsAsync:        isAsync,
	}
	if node.Name() != nil {
		result.Name = node.Name().Text()
	}
	var bindingStmts []SyntaxStatement
	for pIdx, parameter := range valueParameters(node.Parameters()) {
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
			appendSyntaxStatement(&result.Body, statement, chk)
		}
	}
	return result, true
}

func syntaxEnumDeclaration(node *ast.Node, chk *checker.Checker, span SourceSpan) (SyntaxStatement, bool) {
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
}
