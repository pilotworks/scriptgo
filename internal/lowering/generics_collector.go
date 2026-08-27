package lowering

import (
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
)

func scanAndSpecializeStmt(stmt typescriptgo.SyntaxStatement, fileName string, env map[string]string, funcTypes map[string]string, genericFuncs map[string]typescriptgo.SyntaxStatement, genericClasses map[string]typescriptgo.SyntaxClass, genericMethods map[string]typescriptgo.SyntaxMethod, reqFn func(string, []string, string) string, reqCls func(string, []string, string) string, reqMethod func(string, string, []string) string) {
	if stmt.Type != "" {
		scanTypeForGenerics(stmt.Type, fileName, genericClasses, reqCls)
	}
	if stmt.Kind == "variable" && stmt.Type != "" {
		env[stmt.Name] = stmt.Type
	}
	if stmt.Kind == "variable" && stmt.Expression != nil {
		if stmt.Type == "" {
			t := inferExprType(stmt.Expression, env, funcTypes)
			if t != "" {
				env[stmt.Name] = t
			}
		}
		scanAndSpecializeExpr(stmt.Expression, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	if stmt.Expression != nil {
		scanAndSpecializeExpr(stmt.Expression, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	for _, p := range stmt.Parameters {
		if p.Type != "" {
			scanTypeForGenerics(p.Type, fileName, genericClasses, reqCls)
		}
		if p.Initializer != nil {
			scanAndSpecializeExpr(p.Initializer, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
		}
	}
	for _, s := range stmt.Body {
		scanAndSpecializeStmt(s, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	for _, s := range stmt.Then {
		scanAndSpecializeStmt(s, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	for _, s := range stmt.Else {
		scanAndSpecializeStmt(s, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	if stmt.Class != nil {
		if stmt.Class.Extends != "" {
			scanTypeForGenerics(stmt.Class.Extends, fileName, genericClasses, reqCls)
		}
		for _, f := range stmt.Class.Fields {
			if f.Type != "" {
				scanTypeForGenerics(f.Type, fileName, genericClasses, reqCls)
			}
			if f.Initializer != nil {
				scanAndSpecializeExpr(f.Initializer, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
			}
		}
		if stmt.Class.Constructor != nil {
			for _, p := range stmt.Class.Constructor.Parameters {
				if p.Type != "" {
					scanTypeForGenerics(p.Type, fileName, genericClasses, reqCls)
				}
			}
			for _, s := range stmt.Class.Constructor.Body {
				scanAndSpecializeStmt(s, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
			}
		}
		for _, m := range stmt.Class.Methods {
			if len(m.TypeParameters) > 0 {
				continue
			}
			if m.Type != "" {
				scanTypeForGenerics(m.Type, fileName, genericClasses, reqCls)
			}
			for _, p := range m.Parameters {
				if p.Type != "" {
					scanTypeForGenerics(p.Type, fileName, genericClasses, reqCls)
				}
			}
			for _, s := range m.Body {
				scanAndSpecializeStmt(s, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
			}
		}
	}
}

func scanAndSpecializeExpr(expr *typescriptgo.SyntaxExpression, fileName string, env map[string]string, funcTypes map[string]string, genericFuncs map[string]typescriptgo.SyntaxStatement, genericClasses map[string]typescriptgo.SyntaxClass, genericMethods map[string]typescriptgo.SyntaxMethod, reqFn func(string, []string, string) string, reqCls func(string, []string, string) string, reqMethod func(string, string, []string) string) {
	if expr == nil {
		return
	}
	if expr.Kind == "call" && expr.Left != nil {
		if expr.Left.Kind == "identifier" {
			fnName := expr.Left.Text
			if fnTemplate, ok := genericFuncs[fnName]; ok {
				typeArgs := expr.TypeArguments
				if len(typeArgs) == 0 {
					typeArgs = inferTypeArgsForFunc(fnTemplate, expr.Arguments, env, funcTypes)
				}
				if len(typeArgs) == len(fnTemplate.TypeParameters) {
					reqFn(fnName, typeArgs, fileName)
				}
			}
		} else if (expr.Left.Kind == "property" || expr.Left.Kind == "member") && expr.Left.Left != nil {
			var clsName string
			isInstance := false
			if expr.Left.Left.Kind == "identifier" {
				ident := expr.Left.Left.Text
				clsName = ident
				if t, ok := env[ident]; ok && t != "" {
					isInstance = true
					clsName = strings.TrimPrefix(rewriteTypeString(t), "object:")
				}
			} else {
				isInstance = true
				t := inferExprType(expr.Left.Left, env, funcTypes)
				if t != "" {
					clsName = strings.TrimPrefix(rewriteTypeString(t), "object:")
				}
			}
			methodName := expr.Left.Text
			if currUsedMethods != nil {
				currUsedMethods[clsName+"."+methodName] = true
				currUsedMethods[methodName] = true
			}
			lookupKey := clsName + "." + methodName
			if !isInstance {
				lookupKey = clsName + ".static." + methodName
			}
			callTypeArgs := expr.TypeArguments
			if mTemplate, ok := genericMethods[lookupKey]; ok {
				typeArgs := callTypeArgs
				if len(typeArgs) == 0 {
					typeArgs = inferTypeArgsForMethod(mTemplate, mTemplate.TypeParameters, expr.Arguments, env, funcTypes)
				}
				if len(typeArgs) == len(mTemplate.TypeParameters) {
					reqMethod(clsName, methodName, typeArgs)
				}
			}
			if !isInstance && expr.Left.Left.Kind == "identifier" {
				if clsTemplate, ok := genericClasses[clsName]; ok {
					typeArgs := callTypeArgs
					if len(typeArgs) == 0 {
						for _, m := range clsTemplate.Methods {
							if m.IsStatic && m.Name == methodName {
								typeArgs = inferTypeArgsForMethod(m, clsTemplate.TypeParameters, expr.Arguments, env, funcTypes)
								break
							}
						}
					}
					if len(typeArgs) == 0 && clsTemplate.Constructor != nil {
						typeArgs = inferTypeArgsForClass(clsTemplate, expr.Arguments, env, funcTypes)
					}
					if len(typeArgs) == len(clsTemplate.TypeParameters) {
						reqCls(clsName, typeArgs, fileName)
					}
				}
			}
		}
	}
	if expr.Kind == "new" && expr.Left != nil && expr.Left.Kind == "identifier" {
		clsName := expr.Left.Text
		if clsTemplate, ok := genericClasses[clsName]; ok {
			typeArgs := expr.TypeArguments
			if len(typeArgs) == 0 && clsTemplate.Constructor != nil {
				typeArgs = inferTypeArgsForClass(clsTemplate, expr.Arguments, env, funcTypes)
			}
			if len(typeArgs) == len(clsTemplate.TypeParameters) {
				reqCls(clsName, typeArgs, fileName)
			}
		}
	}

	if expr.Left != nil {
		scanAndSpecializeExpr(expr.Left, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	if expr.Right != nil {
		scanAndSpecializeExpr(expr.Right, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	for _, arg := range expr.Arguments {
		scanAndSpecializeExpr(arg, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	if expr.WhenTrue != nil {
		scanAndSpecializeExpr(expr.WhenTrue, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	if expr.WhenFalse != nil {
		scanAndSpecializeExpr(expr.WhenFalse, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	if expr.Function != nil {
		scanAndSpecializeStmt(*expr.Function, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
}

func scanTypeForGenerics(typ, fileName string, genericClasses map[string]typescriptgo.SyntaxClass, reqCls func(string, []string, string) string) {
	clean := strings.TrimPrefix(typ, "object:")
	if strings.HasSuffix(clean, "[]") {
		inner := clean[:len(clean)-2]
		inner = strings.TrimPrefix(inner, "(")
		inner = strings.TrimSuffix(inner, ")")
		scanTypeForGenerics(inner, fileName, genericClasses, reqCls)
		return
	}
	if strings.Contains(clean, "|") {
		for _, part := range strings.Split(clean, "|") {
			scanTypeForGenerics(strings.TrimSpace(part), fileName, genericClasses, reqCls)
		}
		return
	}
	if strings.Contains(clean, "__") {
		idx := strings.Index(clean, "__")
		name := clean[:idx]
		inner := clean[idx+2:]
		inner = strings.TrimSuffix(inner, "_arr")
		parts := strings.Split(inner, "_")
		if _, ok := genericClasses[name]; ok {
			reqCls(name, parts, fileName)
		} else if alias, ok := currGenericTypeAliases[name]; ok {
			tParams := alias.TypeParameters
			if len(tParams) == 0 && alias.Class != nil {
				tParams = alias.Class.TypeParameters
			}
			if len(parts) == len(tParams) {
				subst := make(map[string]string, len(parts))
				for i, tp := range tParams {
					subst[tp] = parts[i]
				}
				expanded := substituteType(alias.Type, subst)
				scanTypeForGenerics(expanded, fileName, genericClasses, reqCls)
			}
		}
		for _, p := range parts {
			scanTypeForGenerics(p, fileName, genericClasses, reqCls)
		}
		return
	}
	if strings.Contains(clean, "<") && strings.HasSuffix(clean, ">") {
		idx := strings.Index(clean, "<")
		name := clean[:idx]
		inner := clean[idx+1 : len(clean)-1]
		parts := splitTypeArguments(inner)
		if _, ok := genericClasses[name]; ok {
			reqCls(name, parts, fileName)
		} else if alias, ok := currGenericTypeAliases[name]; ok {
			tParams := alias.TypeParameters
			if len(tParams) == 0 && alias.Class != nil {
				tParams = alias.Class.TypeParameters
			}
			if len(parts) == len(tParams) {
				subst := make(map[string]string, len(parts))
				for i, tp := range tParams {
					subst[tp] = parts[i]
				}
				expanded := substituteType(alias.Type, subst)
				scanTypeForGenerics(expanded, fileName, genericClasses, reqCls)
			}
		}
		for _, p := range parts {
			scanTypeForGenerics(p, fileName, genericClasses, reqCls)
		}
	}
}

func inferTypeArgsForFunc(fnTemplate typescriptgo.SyntaxStatement, args []*typescriptgo.SyntaxExpression, env map[string]string, funcTypes map[string]string) []string {
	inferred := map[string]string{}
	for i, param := range fnTemplate.Parameters {
		if i < len(args) {
			argType := inferExprType(args[i], env, funcTypes)
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

func inferTypeArgsForMethod(method typescriptgo.SyntaxMethod, classTypeParams []string, args []*typescriptgo.SyntaxExpression, env map[string]string, funcTypes map[string]string) []string {
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

func inferTypeArgsForClass(clsTemplate typescriptgo.SyntaxClass, args []*typescriptgo.SyntaxExpression, env map[string]string, funcTypes map[string]string) []string {
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
	argType = strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(argType, "| null"), "| undefined"))
	argType = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(argType, "null |"), "undefined |"))

	if strings.Contains(argType, "__") && !strings.Contains(argType, "<") {
		idx := strings.Index(argType, "__")
		base := argType[:idx]
		inner := strings.ReplaceAll(argType[idx+2:], "__", ", ")
		argType = base + "<" + inner + ">"
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
	if strings.Contains(paramType, "<") && strings.HasSuffix(paramType, ">") && strings.Contains(argType, "<") && strings.HasSuffix(argType, ">") {
		pIdx := strings.Index(paramType, "<")
		aIdx := strings.Index(argType, "<")
		if paramType[:pIdx] == argType[:aIdx] {
			pParts := splitTypeArguments(paramType[pIdx+1 : len(paramType)-1])
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
	if strings.Contains(paramType, "=>") && strings.Contains(argType, "=>") {
		pParts := strings.Split(paramType, "=>")
		aParts := strings.Split(argType, "=>")
		matchTypeParam(strings.TrimSpace(pParts[0]), strings.TrimSpace(aParts[0]), inferred)
		matchTypeParam(strings.TrimSpace(pParts[1]), strings.TrimSpace(aParts[1]), inferred)
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

func inferExprType(expr *typescriptgo.SyntaxExpression, env map[string]string, funcTypes map[string]string) string {
	if expr == nil {
		return ""
	}
	if funcTypes == nil {
		funcTypes = currFuncTypes
	}
	if expr.InferredType != "" {
		return expr.InferredType
	}
	switch expr.Kind {
	case "number":
		return "number"
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "identifier":
		if t, ok := env[expr.Text]; ok && t != "" {
			return t
		}
		if funcTypes != nil {
			if t, ok := funcTypes[expr.Text]; ok && t != "" {
				return "() => " + t
			}
		}
		return ""
	case "array":
		if len(expr.Arguments) > 0 {
			elemType := inferExprType(expr.Arguments[0], env, funcTypes)
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
						typeArgs = inferTypeArgsForFunc(fnTemplate, expr.Arguments, env, funcTypes)
					}
					if len(typeArgs) == len(fnTemplate.TypeParameters) {
						subst := make(map[string]string, len(typeArgs))
						for i, tp := range fnTemplate.TypeParameters {
							subst[tp] = typeArgs[i]
						}
						return substituteType(fnTemplate.Type, subst)
					}
				}
			} else if (expr.Left.Kind == "property" || expr.Left.Kind == "member") && expr.Left.Left != nil {
				recvType := inferExprType(expr.Left.Left, env, funcTypes)
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
