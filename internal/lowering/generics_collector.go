package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
)

func scanAndSpecializeStmt(stmt frontend.SyntaxStatement, fileName string, env map[string]string, funcTypes map[string]string, genericFuncs map[string]frontend.SyntaxStatement, genericClasses map[string]frontend.SyntaxClass, genericMethods map[string]frontend.SyntaxMethod, reqFn func(string, []string, string) string, reqCls func(string, []string, string) string, reqMethod func(string, string, []string) string) {
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
	} else if stmt.Expression != nil {
		scanAndSpecializeExpr(stmt.Expression, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	for _, p := range stmt.Parameters {
		if p.Type != "" {
			env[p.Name] = p.Type
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
	for _, c := range stmt.Cases {
		scanAndSpecializeExpr(c.Expression, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
		for _, s := range c.Statements {
			scanAndSpecializeStmt(s, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
		}
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
			mEnv := map[string]string{}
			for k, v := range env {
				mEnv[k] = v
			}
			clsName := stmt.Class.Name
			if clsName == "" {
				clsName = stmt.Name
			}
			mEnv["this"] = clsName
			for _, p := range m.Parameters {
				if p.Type != "" {
					scanTypeForGenerics(p.Type, fileName, genericClasses, reqCls)
					mEnv[p.Name] = p.Type
				}
			}
			for _, s := range m.Body {
				scanAndSpecializeStmt(s, fileName, mEnv, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
			}
		}
	}
}

func scanAndSpecializeExpr(expr *frontend.SyntaxExpression, fileName string, env map[string]string, funcTypes map[string]string, genericFuncs map[string]frontend.SyntaxStatement, genericClasses map[string]frontend.SyntaxClass, genericMethods map[string]frontend.SyntaxMethod, reqFn func(string, []string, string) string, reqCls func(string, []string, string) string, reqMethod func(string, string, []string) string) {
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
			baseCls := clsName
			if idx := strings.Index(clsName, "<"); idx != -1 {
				baseCls = clsName[:idx]
			} else if idx := strings.Index(clsName, "__"); idx > 0 {
				baseCls = clsName[:idx]
			}
			qualifiedCls := classIdentityForPath(fileName, clsName)
			if qualifiedCls != "" {
				if stmtClass, hasStmt := classSyntax[qualifiedCls]; hasStmt {
					hasConcrete := false
					for _, m := range stmtClass.Methods {
						if m.Name == methodName && len(m.TypeParameters) == 0 && m.IsStatic == !isInstance {
							hasConcrete = true
							break
						}
					}
					if hasConcrete {
						return
					}
				}
			}
			if stmtClass, hasStmt := classSyntax[clsName]; hasStmt {
				hasConcrete := false
				for _, m := range stmtClass.Methods {
					if m.Name == methodName && len(m.TypeParameters) == 0 && m.IsStatic == !isInstance {
						hasConcrete = true
						break
					}
				}
				if hasConcrete {
					return
				}
			}
			targetCls := qualifiedCls
			if targetCls == "" {
				targetCls = clsName
			}
			lookupKey := targetCls + "." + methodName
			if !isInstance {
				lookupKey = targetCls + ".static." + methodName
			}
			mTemplate, ok := genericMethods[lookupKey]
			if !ok && clsName != targetCls {
				lookupKey = clsName + "." + methodName
				if !isInstance {
					lookupKey = clsName + ".static." + methodName
				}
				if mTemplate, ok = genericMethods[lookupKey]; ok {
					targetCls = clsName
				}
			}
			if !ok && baseCls != "" && baseCls != clsName {
				if !isInstance {
					mTemplate, ok = genericMethods[baseCls+".static."+methodName]
				} else {
					mTemplate, ok = genericMethods[baseCls+"."+methodName]
				}
				if ok {
					targetCls = baseCls
				}
			}
			callTypeArgs := expr.TypeArguments
			if ok {
				typeArgs := callTypeArgs
				if len(typeArgs) == 0 {
					typeArgs = inferTypeArgsForMethod(mTemplate, mTemplate.TypeParameters, expr.Arguments, env, funcTypes)
				}
				if len(typeArgs) == len(mTemplate.TypeParameters) {
					reqMethod(targetCls, methodName, typeArgs)
					if baseCls != "" && baseCls != targetCls {
						reqMethod(baseCls, methodName, typeArgs)
					}
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

func scanTypeForGenerics(typ, fileName string, genericClasses map[string]frontend.SyntaxClass, reqCls func(string, []string, string) string) {
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
		if clsTemplate, ok := genericClasses[name]; ok {
			pList := parts
			if len(pList) > len(clsTemplate.TypeParameters) && len(clsTemplate.TypeParameters) > 0 {
				numParams := len(clsTemplate.TypeParameters)
				mergedLast := strings.Join(pList[numParams-1:], "_")
				pList = append(append([]string(nil), pList[:numParams-1]...), mergedLast)
			}
			reqCls(name, pList, fileName)
		} else if alias, ok := currGenericTypeAliases[name]; ok {
			tParams := alias.TypeParameters
			if len(tParams) == 0 && alias.Class != nil {
				tParams = alias.Class.TypeParameters
			}
			pList := parts
			if len(pList) > len(tParams) && len(tParams) > 0 {
				numParams := len(tParams)
				mergedLast := strings.Join(pList[numParams-1:], "_")
				pList = append(append([]string(nil), pList[:numParams-1]...), mergedLast)
			}
			if len(pList) == len(tParams) {
				subst := make(map[string]string, len(pList))
				for i, tp := range tParams {
					subst[tp] = pList[i]
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
