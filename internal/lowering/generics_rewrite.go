package lowering

import (
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
)

func rewriteMethod(m typescriptgo.SyntaxMethod, env map[string]string, genericFuncs map[string]typescriptgo.SyntaxStatement, genericClasses map[string]typescriptgo.SyntaxClass, genericMethods map[string]typescriptgo.SyntaxMethod, reqFn func(string, []string, string) string, reqCls func(string, []string, string) string, reqMethod func(string, string, []string) string, fileName string) typescriptgo.SyntaxMethod {
	res := cloneMethod(m)
	res.Type = rewriteTypeString(res.Type)
	localEnv := map[string]string{}
	for k, v := range env {
		localEnv[k] = v
	}
	for j := range res.Parameters {
		res.Parameters[j].Type = rewriteTypeString(res.Parameters[j].Type)
		if res.Parameters[j].Type != "" {
			localEnv[res.Parameters[j].Name] = res.Parameters[j].Type
		}
		if res.Parameters[j].Initializer != nil {
			res.Parameters[j].Initializer = rewriteExpr(res.Parameters[j].Initializer, localEnv, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
		}
	}
	for j := range res.Body {
		res.Body[j] = rewriteStatementTypes(res.Body[j], localEnv, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	return res
}

func rewriteStatementTypes(stmt typescriptgo.SyntaxStatement, env map[string]string, genericFuncs map[string]typescriptgo.SyntaxStatement, genericClasses map[string]typescriptgo.SyntaxClass, genericMethods map[string]typescriptgo.SyntaxMethod, reqFn func(string, []string, string) string, reqCls func(string, []string, string) string, reqMethod func(string, string, []string) string, fileName string) typescriptgo.SyntaxStatement {
	res := cloneStatement(stmt)
	res.Type = rewriteTypeString(res.Type)
	if env == nil {
		env = map[string]string{}
	}
	if stmt.Kind == "variable" && stmt.Name != "" {
		if stmt.Type != "" {
			env[stmt.Name] = stmt.Type
		} else if stmt.Expression != nil {
			env[stmt.Name] = inferExprType(stmt.Expression, env, nil)
		}
	}
	for i := range res.Parameters {
		res.Parameters[i].Type = rewriteTypeString(res.Parameters[i].Type)
		if res.Parameters[i].Type != "" {
			env[res.Parameters[i].Name] = res.Parameters[i].Type
		}
		if res.Parameters[i].Initializer != nil {
			res.Parameters[i].Initializer = rewriteExpr(res.Parameters[i].Initializer, env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
		}
	}
	if res.Expression != nil {
		res.Expression = rewriteExpr(res.Expression, env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	if res.Left != nil {
		res.Left = rewriteExpr(res.Left, env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	if res.Right != nil {
		res.Right = rewriteExpr(res.Right, env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	for i := range res.Body {
		res.Body[i] = rewriteStatementTypes(res.Body[i], env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	for i := range res.Then {
		res.Then[i] = rewriteStatementTypes(res.Then[i], env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	for i := range res.Else {
		res.Else[i] = rewriteStatementTypes(res.Else[i], env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	for i := range res.Catch {
		res.Catch[i] = rewriteStatementTypes(res.Catch[i], env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	for i := range res.Finally {
		res.Finally[i] = rewriteStatementTypes(res.Finally[i], env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	if res.Class != nil {
		res.Class.Extends = rewriteTypeString(res.Class.Extends)
		for i := range res.Class.Fields {
			res.Class.Fields[i].Type = rewriteTypeString(res.Class.Fields[i].Type)
			if res.Class.Fields[i].Initializer != nil {
				res.Class.Fields[i].Initializer = rewriteExpr(res.Class.Fields[i].Initializer, env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
			}
		}
		if res.Class.Constructor != nil {
			ctorEnv := map[string]string{}
			for k, v := range env {
				ctorEnv[k] = v
			}
			for i := range res.Class.Constructor.Parameters {
				res.Class.Constructor.Parameters[i].Type = rewriteTypeString(res.Class.Constructor.Parameters[i].Type)
				if res.Class.Constructor.Parameters[i].Type != "" {
					ctorEnv[res.Class.Constructor.Parameters[i].Name] = res.Class.Constructor.Parameters[i].Type
				}
				if res.Class.Constructor.Parameters[i].Initializer != nil {
					res.Class.Constructor.Parameters[i].Initializer = rewriteExpr(res.Class.Constructor.Parameters[i].Initializer, ctorEnv, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
				}
			}
			for i := range res.Class.Constructor.Body {
				res.Class.Constructor.Body[i] = rewriteStatementTypes(res.Class.Constructor.Body[i], ctorEnv, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
			}
		}
		classEnv := map[string]string{}
		for k, v := range env {
			classEnv[k] = v
		}
		if res.Class != nil {
			clsName := res.Class.Name
			if clsName == "" {
				clsName = res.Name
			}
			classEnv["this"] = clsName
		}
		var nonGenericMethods []typescriptgo.SyntaxMethod
		for _, m := range res.Class.Methods {
			if len(m.TypeParameters) > 0 {
				continue // Skip generic method template
			}
			nonGenericMethods = append(nonGenericMethods, rewriteMethod(m, classEnv, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName))
		}
		res.Class.Methods = nonGenericMethods
	}
	return res
}

func rewriteExpr(expr *typescriptgo.SyntaxExpression, env map[string]string, genericFuncs map[string]typescriptgo.SyntaxStatement, genericClasses map[string]typescriptgo.SyntaxClass, genericMethods map[string]typescriptgo.SyntaxMethod, reqFn func(string, []string, string) string, reqCls func(string, []string, string) string, reqMethod func(string, string, []string) string, fileName string) *typescriptgo.SyntaxExpression {
	if expr == nil {
		return nil
	}
	res := cloneExpr(expr)
	if res.Kind == "call" && res.Left != nil {
		if res.Left.Kind == "identifier" {
			fnName := res.Left.Text
			if fnTemplate, ok := genericFuncs[fnName]; ok {
				typeArgs := res.TypeArguments
				if len(typeArgs) == 0 {
					typeArgs = inferTypeArgsForFunc(fnTemplate, res.Arguments, env, nil)
				}
				if len(typeArgs) == len(fnTemplate.TypeParameters) {
					mangled := reqFn(fnName, typeArgs, fileName)
					res.Left.Text = mangled
					res.TypeArguments = nil
				}
			}
		} else if (res.Left.Kind == "property" || res.Left.Kind == "member") && res.Left.Left != nil {
			var clsName string
			isInstance := false
			if res.Left.Left.Kind == "identifier" {
				ident := res.Left.Left.Text
				clsName = ident
				if env != nil {
					if t, ok := env[ident]; ok && t != "" {
						isInstance = true
						clsName = strings.TrimPrefix(rewriteTypeString(t), "object:")
					}
				}
			} else {
				isInstance = true
				t := inferExprType(res.Left.Left, env, nil)
				if t != "" {
					clsName = strings.TrimPrefix(rewriteTypeString(t), "object:")
				}
			}
			methodName := res.Left.Text
			baseCls := clsName
			if idx := strings.Index(clsName, "<"); idx != -1 {
				baseCls = clsName[:idx]
			} else if idx := strings.Index(clsName, "__"); idx != -1 {
				baseCls = clsName[:idx]
			}
			lookupKey := clsName + "." + methodName
			if !isInstance {
				lookupKey = clsName + ".static." + methodName
			}
			mTemplate, ok := genericMethods[lookupKey]
			if !ok {
				if !isInstance {
					mTemplate, ok = genericMethods[baseCls+".static."+methodName]
				} else {
					mTemplate, ok = genericMethods[baseCls+"."+methodName]
				}
			}
			callTypeArgs := res.TypeArguments
			if ok {
				typeArgs := callTypeArgs
				if len(typeArgs) == 0 {
					typeArgs = inferTypeArgsForMethod(mTemplate, mTemplate.TypeParameters, res.Arguments, env, nil)
				}
				if len(typeArgs) == len(mTemplate.TypeParameters) {
					mangledMethod := reqMethod(clsName, methodName, typeArgs)
					if baseCls != clsName {
						reqMethod(baseCls, methodName, typeArgs)
					}
					res.Left.Text = mangledMethod
					res.TypeArguments = nil
				}
			}
			if !isInstance && res.Left.Left.Kind == "identifier" {
				if clsTemplate, ok := genericClasses[clsName]; ok {
					typeArgs := callTypeArgs
					if len(typeArgs) == 0 {
						for _, m := range clsTemplate.Methods {
							if m.IsStatic && m.Name == methodName {
								typeArgs = inferTypeArgsForMethod(m, clsTemplate.TypeParameters, res.Arguments, env, nil)
								break
							}
						}
					}
					if len(typeArgs) == 0 && clsTemplate.Constructor != nil {
						typeArgs = inferTypeArgsForClass(clsTemplate, res.Arguments, env, nil)
					}
					if len(typeArgs) == len(clsTemplate.TypeParameters) {
						mangled := reqCls(clsName, typeArgs, fileName)
						res.Left.Left.Text = mangled
						res.TypeArguments = nil
					}
				}
			}
		}
	}
	if res.Kind == "new" && res.Left != nil && res.Left.Kind == "identifier" {
		clsName := res.Left.Text
		if clsTemplate, ok := genericClasses[clsName]; ok {
			typeArgs := res.TypeArguments
			if len(typeArgs) == 0 && clsTemplate.Constructor != nil {
				typeArgs = inferTypeArgsForClass(clsTemplate, res.Arguments, env, nil)
			}
			if len(typeArgs) == len(clsTemplate.TypeParameters) {
				mangled := reqCls(clsName, typeArgs, fileName)
				res.Left.Text = mangled
				res.TypeArguments = nil
			}
		}
	}
	if res.Left != nil {
		res.Left = rewriteExpr(res.Left, env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	if res.Right != nil {
		res.Right = rewriteExpr(res.Right, env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	for i := range res.Arguments {
		res.Arguments[i] = rewriteExpr(res.Arguments[i], env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	if res.WhenTrue != nil {
		res.WhenTrue = rewriteExpr(res.WhenTrue, env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	if res.WhenFalse != nil {
		res.WhenFalse = rewriteExpr(res.WhenFalse, env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
	}
	if res.Function != nil {
		fnStmt := rewriteStatementTypes(*res.Function, env, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod, fileName)
		res.Function = &fnStmt
	}
	return res
}

func rewriteTypeString(typ string) string {
	if typ == "" {
		return typ
	}
	if strings.HasSuffix(typ, "[]") {
		elem := typ[:len(typ)-2]
		hadParens := false
		if strings.HasPrefix(elem, "(") && strings.HasSuffix(elem, ")") {
			hadParens = true
			elem = elem[1 : len(elem)-1]
		}
		rewrittenElem := rewriteTypeString(elem)
		if hadParens || strings.Contains(rewrittenElem, "|") {
			return "(" + rewrittenElem + ")[]"
		}
		return rewrittenElem + "[]"
	}
	if strings.Contains(typ, "|") {
		var parts []string
		for _, part := range strings.Split(typ, "|") {
			parts = append(parts, rewriteTypeString(strings.TrimSpace(part)))
		}
		return strings.Join(parts, " | ")
	}
	hasObj := strings.HasPrefix(typ, "object:")
	clean := strings.TrimPrefix(typ, "object:")
	if strings.Contains(clean, "__") {
		idx := strings.Index(clean, "__")
		name := clean[:idx]
		inner := clean[idx+2:]
		inner = strings.TrimSuffix(inner, "_arr")
		parts := strings.Split(inner, "_")
		var newParts []string
		for _, p := range parts {
			newParts = append(newParts, rewriteTypeString(p))
		}
		if alias, ok := currGenericTypeAliases[name]; ok {
			tParams := alias.TypeParameters
			if len(tParams) == 0 && alias.Class != nil {
				tParams = alias.Class.TypeParameters
			}
			if len(newParts) == len(tParams) {
				subst := make(map[string]string, len(newParts))
				for i, tp := range tParams {
					subst[tp] = newParts[i]
				}
				return rewriteTypeString(substituteType(alias.Type, subst))
			}
		}
	}
	if strings.Contains(clean, "<") && strings.HasSuffix(clean, ">") {
		idx := strings.Index(clean, "<")
		name := clean[:idx]
		inner := clean[idx+1 : len(clean)-1]
		parts := splitTypeArguments(inner)
		var newParts []string
		for _, p := range parts {
			newParts = append(newParts, rewriteTypeString(p))
		}
		if name == "Array" || name == "ReadonlyArray" {
			if len(newParts) == 1 {
				return newParts[0] + "[]"
			}
		}
		switch name {
		case "Uint8Array", "Int8Array", "Uint8ClampedArray", "Int16Array", "Uint16Array", "Int32Array", "Uint32Array", "Float32Array", "Float64Array", "BigInt64Array", "BigUint64Array", "ArrayBuffer", "DataView":
			return name
		}
		if alias, ok := currGenericTypeAliases[name]; ok {
			tParams := alias.TypeParameters
			if len(tParams) == 0 && alias.Class != nil {
				tParams = alias.Class.TypeParameters
			}
			if len(newParts) == len(tParams) {
				subst := make(map[string]string, len(newParts))
				for i, tp := range tParams {
					subst[tp] = newParts[i]
				}
				return rewriteTypeString(substituteType(alias.Type, subst))
			}
		}
		if isBuiltinGeneric(name) {
			res := name + "<" + strings.Join(newParts, ", ") + ">"
			if hasObj {
				return "object:" + res
			}
			return res
		}
		mangled := mangleGenericName(name, newParts)
		if hasObj {
			return "object:" + mangled
		}
		return mangled
	}
	return typ
}

func normalizeGenericArrayTypes(program frontend.Program) frontend.Program {
	for i := range program.Files {
		for j := range program.Files[i].Syntax.Statements {
			program.Files[i].Syntax.Statements[j] = normalizeStmtArrayTypes(program.Files[i].Syntax.Statements[j])
		}
	}
	return program
}

func normalizeStmtArrayTypes(stmt typescriptgo.SyntaxStatement) typescriptgo.SyntaxStatement {
	stmt.Type = rewriteTypeString(stmt.Type)
	for i := range stmt.Parameters {
		stmt.Parameters[i].Type = rewriteTypeString(stmt.Parameters[i].Type)
	}
	for i := range stmt.Body {
		stmt.Body[i] = normalizeStmtArrayTypes(stmt.Body[i])
	}
	for i := range stmt.Then {
		stmt.Then[i] = normalizeStmtArrayTypes(stmt.Then[i])
	}
	for i := range stmt.Else {
		stmt.Else[i] = normalizeStmtArrayTypes(stmt.Else[i])
	}
	for i := range stmt.Catch {
		stmt.Catch[i] = normalizeStmtArrayTypes(stmt.Catch[i])
	}
	for i := range stmt.Finally {
		stmt.Finally[i] = normalizeStmtArrayTypes(stmt.Finally[i])
	}
	if stmt.Class != nil {
		stmt.Class.Extends = rewriteTypeString(stmt.Class.Extends)
		for i := range stmt.Class.Fields {
			stmt.Class.Fields[i].Type = rewriteTypeString(stmt.Class.Fields[i].Type)
		}
		if stmt.Class.Constructor != nil {
			for i := range stmt.Class.Constructor.Parameters {
				stmt.Class.Constructor.Parameters[i].Type = rewriteTypeString(stmt.Class.Constructor.Parameters[i].Type)
			}
			for i := range stmt.Class.Constructor.Body {
				stmt.Class.Constructor.Body[i] = normalizeStmtArrayTypes(stmt.Class.Constructor.Body[i])
			}
		}
		for i := range stmt.Class.Methods {
			stmt.Class.Methods[i].Type = rewriteTypeString(stmt.Class.Methods[i].Type)
			for j := range stmt.Class.Methods[i].Parameters {
				stmt.Class.Methods[i].Parameters[j].Type = rewriteTypeString(stmt.Class.Methods[i].Parameters[j].Type)
			}
			for j := range stmt.Class.Methods[i].Body {
				stmt.Class.Methods[i].Body[j] = normalizeStmtArrayTypes(stmt.Class.Methods[i].Body[j])
			}
		}
		for i := range stmt.Class.StaticBlocks {
			for j := range stmt.Class.StaticBlocks[i] {
				stmt.Class.StaticBlocks[i][j] = normalizeStmtArrayTypes(stmt.Class.StaticBlocks[i][j])
			}
		}
		for i := range stmt.Class.StaticElements {
			if stmt.Class.StaticElements[i].Kind == typescriptgo.StaticElementBlock {
				for j := range stmt.Class.StaticElements[i].Statements {
					stmt.Class.StaticElements[i].Statements[j] = normalizeStmtArrayTypes(stmt.Class.StaticElements[i].Statements[j])
				}
			}
		}
	}
	return stmt
}
