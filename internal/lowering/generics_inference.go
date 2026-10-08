package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
)

func inferTypeArgsForFunc(fnTemplate frontend.SyntaxStatement, args []*frontend.SyntaxExpression, env map[string]string, funcTypes map[string]string) []string {
	return inferTypeArgsForFuncDepth(fnTemplate, args, env, funcTypes, 0)
}

func inferTypeArgsForFuncDepth(fnTemplate frontend.SyntaxStatement, args []*frontend.SyntaxExpression, env map[string]string, funcTypes map[string]string, depth int) []string {
	inferred := map[string]string{}
	for i, param := range fnTemplate.Parameters {
		if i < len(args) {
			argType := inferExprTypeDepth(args[i], env, funcTypes, depth)
			matchTypeParam(param.Type, argType, inferred)
		}
	}
	var res []string
	for _, tp := range fnTemplate.TypeParameters {
		if t, ok := inferred[tp]; ok && t != "" {
			res = append(res, t)
		} else {
			res = append(res, "number")
		}
	}
	return res
}

func inferTypeArgsForMethod(method frontend.SyntaxMethod, classTypeParams []string, args []*frontend.SyntaxExpression, env map[string]string, funcTypes map[string]string) []string {
	inferred := map[string]string{}
	for i, param := range method.Parameters {
		if i < len(args) {
			argType := inferExprType(args[i], env, funcTypes)
			matchTypeParam(param.Type, argType, inferred)
		}
	}
	var res []string
	for _, tp := range classTypeParams {
		if t, ok := inferred[tp]; ok && t != "" {
			res = append(res, t)
		} else {
			res = append(res, "number")
		}
	}
	return res
}

func inferTypeArgsForClass(clsTemplate frontend.SyntaxClass, args []*frontend.SyntaxExpression, env map[string]string, funcTypes map[string]string) []string {
	inferred := map[string]string{}
	if clsTemplate.Constructor != nil {
		for i, param := range clsTemplate.Constructor.Parameters {
			if i < len(args) {
				argType := inferExprType(args[i], env, funcTypes)
				matchTypeParam(param.Type, argType, inferred)
			}
		}
	}
	var res []string
	for _, tp := range clsTemplate.TypeParameters {
		if t, ok := inferred[tp]; ok && t != "" {
			res = append(res, t)
		} else {
			res = append(res, "number")
		}
	}
	return res
}

func matchTypeParam(paramType, argType string, inferred map[string]string) {
	if paramType == "" || argType == "" {
		return
	}
	paramType = strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(paramType, "| null"), "| undefined"))
	paramType = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(paramType, "null |"), "undefined |"))
	paramType = strings.TrimPrefix(paramType, "object:")
	argType = strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(argType, "| null"), "| undefined"))
	argType = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(argType, "null |"), "undefined |"))
	argType = strings.TrimPrefix(argType, "object:")

	if strings.Contains(argType, "__") && !strings.Contains(argType, "<") {
		idx := strings.Index(argType, "__")
		base := argType[:idx]
		if _, ok := currGenericClasses[base]; ok {
			inner := strings.ReplaceAll(argType[idx+2:], "__", ", ")
			argType = base + "<" + inner + ">"
		} else if _, ok := currGenericTypeAliases[base]; ok {
			inner := strings.ReplaceAll(argType[idx+2:], "__", ", ")
			argType = base + "<" + inner + ">"
		}
	}

	if strings.HasSuffix(paramType, "[]") && strings.HasSuffix(argType, "[]") {
		matchTypeParam(paramType[:len(paramType)-2], argType[:len(argType)-2], inferred)
		return
	}
	if strings.HasPrefix(paramType, "Array<") && strings.HasSuffix(paramType, ">") {
		innerParam := strings.TrimSuffix(strings.TrimPrefix(paramType, "Array<"), ">")
		if strings.HasSuffix(argType, "[]") {
			matchTypeParam(innerParam, argType[:len(argType)-2], inferred)
			return
		}
		if strings.HasPrefix(argType, "Array<") && strings.HasSuffix(argType, ">") {
			innerArg := strings.TrimSuffix(strings.TrimPrefix(argType, "Array<"), ">")
			matchTypeParam(innerParam, innerArg, inferred)
			return
		}
	}

	if strings.HasPrefix(paramType, "[") && strings.HasSuffix(paramType, "]") {
		pInner := paramType[1 : len(paramType)-1]
		pParts := strings.Split(pInner, ",")
		var aParts []string
		if strings.HasPrefix(argType, "[") && strings.HasSuffix(argType, "]") {
			aInner := argType[1 : len(argType)-1]
			aParts = strings.Split(aInner, ",")
		} else if strings.HasPrefix(argType, "__shape_") {
			clean := strings.TrimPrefix(argType, "__shape_")
			tokens := strings.Split(clean, "_")
			for i := 0; i < len(tokens); i += 2 {
				if i+1 < len(tokens) {
					aParts = append(aParts, tokens[i+1])
				}
			}
		}
		minLen := len(pParts)
		if len(aParts) < minLen {
			minLen = len(aParts)
		}
		for i := 0; i < minLen; i++ {
			matchTypeParam(strings.TrimSpace(pParts[i]), strings.TrimSpace(aParts[i]), inferred)
		}
		return
	}
	if strings.HasPrefix(paramType, "__shape_") {
		pClean := strings.TrimPrefix(paramType, "__shape_")
		pTokens := strings.Split(pClean, "_")
		var pParts []string
		for i := 0; i < len(pTokens); i += 2 {
			if i+1 < len(pTokens) {
				pParts = append(pParts, pTokens[i+1])
			}
		}
		var aParts []string
		if strings.HasPrefix(argType, "__shape_") {
			aClean := strings.TrimPrefix(argType, "__shape_")
			aTokens := strings.Split(aClean, "_")
			for i := 0; i < len(aTokens); i += 2 {
				if i+1 < len(aTokens) {
					aParts = append(aParts, aTokens[i+1])
				}
			}
		} else if strings.HasPrefix(argType, "[") && strings.HasSuffix(argType, "]") {
			aInner := argType[1 : len(argType)-1]
			aParts = strings.Split(aInner, ",")
		}
		minLen := len(pParts)
		if len(aParts) < minLen {
			minLen = len(aParts)
		}
		for i := 0; i < minLen; i++ {
			matchTypeParam(strings.TrimSpace(pParts[i]), strings.TrimSpace(aParts[i]), inferred)
		}
		return
	}
	if strings.Contains(paramType, "<") && strings.HasSuffix(paramType, ">") {
		pIdx := strings.Index(paramType, "<")
		pName := paramType[:pIdx]
		pParts := splitTypeArguments(paramType[pIdx+1 : len(paramType)-1])
		if alias, ok := currGenericTypeAliases[pName]; ok {
			tParams := alias.TypeParameters
			if len(tParams) == 0 && alias.Class != nil {
				tParams = alias.Class.TypeParameters
			}
			if len(pParts) == len(tParams) {
				subst := make(map[string]string, len(pParts))
				for i, tp := range tParams {
					subst[tp] = pParts[i]
				}
				expanded := substituteType(alias.Type, subst)
				matchTypeParam(expanded, argType, inferred)
				return
			}
		}
		if strings.Contains(argType, "<") && strings.HasSuffix(argType, ">") {
			aIdx := strings.Index(argType, "<")
			if paramType[:pIdx] == argType[:aIdx] {
				aParts := splitTypeArguments(argType[aIdx+1 : len(argType)-1])
				minLen := len(pParts)
				if len(aParts) < minLen {
					minLen = len(aParts)
				}
				for i := 0; i < minLen; i++ {
					matchTypeParam(pParts[i], aParts[i], inferred)
				}
				return
			}
		}
		if meta, ok := classHierarchy[argType]; ok {
			for _, imp := range meta.Implements {
				matchTypeParam(paramType, imp, inferred)
			}
			if meta.Extends != "" {
				matchTypeParam(paramType, meta.Extends, inferred)
			}
			return
		}
	}
	if strings.HasPrefix(paramType, "(") && strings.HasSuffix(paramType, ")") && strings.HasPrefix(argType, "(") && strings.HasSuffix(argType, ")") && !strings.Contains(paramType, "=>") && !strings.Contains(argType, "=>") {
		pInner := paramType[1 : len(paramType)-1]
		aInner := argType[1 : len(argType)-1]
		pParts := splitTypeArguments(pInner)
		aParts := splitTypeArguments(aInner)
		minLen := len(pParts)
		if len(aParts) < minLen {
			minLen = len(aParts)
		}
		for i := 0; i < minLen; i++ {
			pP := pParts[i]
			aP := aParts[i]
			if idx := strings.Index(pP, ":"); idx != -1 {
				pP = strings.TrimSpace(pP[idx+1:])
			}
			if idx := strings.Index(aP, ":"); idx != -1 {
				aP = strings.TrimSpace(aP[idx+1:])
			}
			matchTypeParam(pP, aP, inferred)
		}
		return
	}
	if strings.Contains(paramType, "=>") && strings.Contains(argType, "=>") {
		pIdx := strings.LastIndex(paramType, "=>")
		aIdx := strings.LastIndex(argType, "=>")
		pParamsStr := strings.TrimSpace(paramType[:pIdx])
		aParamsStr := strings.TrimSpace(argType[:aIdx])
		pRet := strings.TrimSpace(paramType[pIdx+2:])
		aRet := strings.TrimSpace(argType[aIdx+2:])
		matchTypeParam(pRet, aRet, inferred)

		pParamsStr = strings.TrimPrefix(strings.TrimSuffix(pParamsStr, ")"), "(")
		aParamsStr = strings.TrimPrefix(strings.TrimSuffix(aParamsStr, ")"), "(")
		if pParamsStr != "" && aParamsStr != "" {
			pList := strings.Split(pParamsStr, ",")
			aList := strings.Split(aParamsStr, ",")
			minLen := len(pList)
			if len(aList) < minLen {
				minLen = len(aList)
			}
			for i := 0; i < minLen; i++ {
				pItem := strings.TrimSpace(pList[i])
				aItem := strings.TrimSpace(aList[i])
				if idx := strings.Index(pItem, ":"); idx != -1 {
					pItem = strings.TrimSpace(pItem[idx+1:])
				}
				if idx := strings.Index(aItem, ":"); idx != -1 {
					aItem = strings.TrimSpace(aItem[idx+1:])
				}
				matchTypeParam(pItem, aItem, inferred)
			}
		}
		return
	}
	cleanParam := strings.TrimSpace(strings.TrimPrefix(paramType, "() => "))
	cleanArg := strings.TrimSpace(strings.TrimPrefix(argType, "() => "))
	if cleanParam != paramType || cleanArg != argType {
		matchTypeParam(cleanParam, cleanArg, inferred)
		return
	}
	if !strings.Contains(paramType, "<") && !strings.Contains(paramType, "[]") && !strings.Contains(paramType, "(") {
		inferred[paramType] = argType
	}
}

func inferExprType(expr *frontend.SyntaxExpression, env map[string]string, funcTypes map[string]string) string {
	return inferExprTypeDepth(expr, env, funcTypes, 0)
}

func inferExprTypeDepth(expr *frontend.SyntaxExpression, env map[string]string, funcTypes map[string]string, depth int) string {
	if expr == nil || depth > 8 {
		return ""
	}
	if funcTypes == nil {
		funcTypes = currFuncTypes
	}
	if expr.Kind == "this" || expr.Text == "this" || expr.InferredType == "this" {
		if t, ok := env["this"]; ok && t != "" {
			return t
		}
	}
	if expr.InferredType != "" {
		return expr.InferredType
	}
	switch expr.Kind {
	case "number":
		return "number"
	case "string", "template":
		return "string"
	case "bool":
		return "bool"
	case "arrow_function", "function":
		if expr.Function != nil {
			fn := expr.Function
			var paramTypes []string
			for _, p := range fn.Parameters {
				pt := p.Type
				if pt == "" {
					pt = p.InferredType
				}
				if pt == "" {
					pt = "unknown"
				}
				paramTypes = append(paramTypes, pt)
			}
			retType := fn.Type
			if retType == "" {
				retType = fn.InferredType
			}
			if retType == "" && len(fn.Body) > 0 {
				for _, stmt := range fn.Body {
					if stmt.Kind == "return" && stmt.Expression != nil {
						retType = inferExprTypeDepth(stmt.Expression, env, funcTypes, depth+1)
						break
					}
				}
			}
			if retType == "" {
				retType = "void"
			}
			return "(" + strings.Join(paramTypes, ", ") + ") => " + retType
		}
		return "() => void"
	case "property", "member":
		if expr.Text == "length" {
			return "number"
		}
		return ""
	case "this":
		if t, ok := env["this"]; ok && t != "" {
			return t
		}
		return ""
	case "identifier":
		if t, ok := env[expr.Text]; ok && t != "" {
			return t
		}
		if funcTypes != nil {
			if t, ok := funcTypes[expr.Text]; ok && t != "" {
				return t
			}
		}
		return ""
	case "array":
		if len(expr.Arguments) > 0 {
			elemType := inferExprTypeDepth(expr.Arguments[0], env, funcTypes, depth+1)
			if elemType != "" {
				return elemType + "[]"
			}
		}
		return "number[]"
	case "new":
		if expr.Left != nil && expr.Left.Kind == "identifier" {
			if len(expr.TypeArguments) > 0 {
				return mangleGenericName(expr.Left.Text, expr.TypeArguments)
			}
			return expr.Left.Text
		}
	case "call":
		if expr.Left != nil {
			if expr.Left.Kind == "identifier" {
				fnName := expr.Left.Text
				if fnTemplate, ok := currGenericFuncs[fnName]; ok {
					typeArgs := expr.TypeArguments
					if len(typeArgs) == 0 {
						typeArgs = inferTypeArgsForFuncDepth(fnTemplate, expr.Arguments, env, funcTypes, depth+1)
					}
					if len(typeArgs) == len(fnTemplate.TypeParameters) {
						subst := make(map[string]string, len(typeArgs))
						for i, tp := range fnTemplate.TypeParameters {
							subst[tp] = typeArgs[i]
						}
						return substituteType(fnTemplate.Type, subst)
					}
				} else if retType, ok := funcTypes[fnName]; ok && retType != "" {
					return retType
				} else if t, ok := env[fnName]; ok && t != "" {
					if strings.Contains(t, "=>") {
						return extractTopLevelReturnType(t)
					}
					return t
				}
			} else if (expr.Left.Kind == "property" || expr.Left.Kind == "member") && expr.Left.Left != nil {
				recvType := inferExprTypeDepth(expr.Left.Left, env, funcTypes, depth+1)
				cleanRecv := strings.TrimPrefix(recvType, "object:")
				var className string
				var classTypeArgs []string
				if idx := strings.Index(cleanRecv, "<"); idx != -1 && strings.HasSuffix(cleanRecv, ">") {
					className = cleanRecv[:idx]
					classTypeArgs = splitTypeArguments(cleanRecv[idx+1 : len(cleanRecv)-1])
				} else if idx := strings.Index(cleanRecv, "__"); idx != -1 {
					className = cleanRecv[:idx]
					classTypeArgs = strings.Split(cleanRecv[idx+2:], "_")
				} else {
					className = cleanRecv
				}
				methodName := expr.Left.Text

				// Check static method on generic class template
				if clsTemplate, ok := currGenericClasses[className]; ok {
					for _, m := range clsTemplate.Methods {
						if m.IsStatic && m.Name == methodName {
							typeArgs := expr.TypeArguments
							if len(typeArgs) == 0 {
								typeArgs = inferTypeArgsForMethod(m, clsTemplate.TypeParameters, expr.Arguments, env, funcTypes)
							}
							subst := make(map[string]string, len(typeArgs))
							for i, tp := range clsTemplate.TypeParameters {
								if i < len(typeArgs) {
									subst[tp] = typeArgs[i]
								}
							}
							return substituteType(m.Type, subst)
						}
					}
				}

				// Check instance method on specialized class or template
				if mTemplate, ok := currGenericMethods[className+"."+methodName]; ok {
					typeArgs := expr.TypeArguments
					if len(typeArgs) == 0 {
						typeArgs = inferTypeArgsForMethod(mTemplate, mTemplate.TypeParameters, expr.Arguments, env, funcTypes)
					}
					subst := make(map[string]string, len(typeArgs))
					for i, tp := range mTemplate.TypeParameters {
						if i < len(typeArgs) {
							subst[tp] = typeArgs[i]
						}
					}
					return substituteType(mTemplate.Type, subst)
				}

				if clsTemplate, ok := currGenericClasses[className]; ok {
					for _, m := range clsTemplate.Methods {
						if m.Name == methodName {
							subst := make(map[string]string, len(classTypeArgs))
							for i, tp := range clsTemplate.TypeParameters {
								if i < len(classTypeArgs) {
									subst[tp] = classTypeArgs[i]
								}
							}
							if len(m.TypeParameters) > 0 {
								typeArgs := expr.TypeArguments
								if len(typeArgs) == 0 {
									typeArgs = inferTypeArgsForMethod(m, m.TypeParameters, expr.Arguments, env, funcTypes)
								}
								for i, tp := range m.TypeParameters {
									if i < len(typeArgs) {
										subst[tp] = typeArgs[i]
									}
								}
							}
							return substituteType(m.Type, subst)
						}
					}
				}
			}
		}
	}
	return ""
}
