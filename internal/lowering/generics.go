package lowering

import (
	"maps"
	"strings"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
)

var (
	currGenericFuncs   map[string]typescriptgo.SyntaxStatement
	currGenericClasses map[string]typescriptgo.SyntaxClass
	currGenericMethods map[string]typescriptgo.SyntaxMethod
	currFuncTypes      map[string]string
	currUsedMethods    map[string]bool
)

// SpecializeGenerics monomorphizes generic functions and classes based on
// concrete type arguments at call sites, instantiations, and type annotations.
func SpecializeGenerics(program frontend.Program) (frontend.Program, error) {
	genericFuncs := map[string]typescriptgo.SyntaxStatement{}
	genericClasses := map[string]typescriptgo.SyntaxClass{}
	genericClassKinds := map[string]string{}
	genericMethods := map[string]typescriptgo.SyntaxMethod{}
	funcTypes := map[string]string{}
	usedMethods := map[string]bool{}

	currGenericFuncs = genericFuncs
	currGenericClasses = genericClasses
	currGenericMethods = genericMethods
	currFuncTypes = funcTypes
	currUsedMethods = usedMethods
	defer func() {
		currGenericFuncs = nil
		currGenericClasses = nil
		currGenericMethods = nil
		currFuncTypes = nil
		currUsedMethods = nil
	}()

	for _, file := range program.Files {
		if strings.HasSuffix(file.FileName, ".d.ts") {
			continue
		}
		for _, statement := range file.Syntax.Statements {
			if statement.Kind == "function" && len(statement.TypeParameters) > 0 {
				genericFuncs[statement.Name] = statement
			} else if (statement.Kind == "class" || statement.Kind == "interface" || statement.Kind == "type_alias") && statement.Class != nil {
				if len(statement.Class.TypeParameters) > 0 {
					genericClasses[statement.Class.Name] = *statement.Class
					genericClassKinds[statement.Class.Name] = statement.Kind
				}
				for _, m := range statement.Class.Methods {
					if len(m.TypeParameters) > 0 && len(m.Body) > 0 {
						if m.IsStatic {
							genericMethods[statement.Class.Name+".static."+m.Name] = m
						} else {
							genericMethods[statement.Class.Name+"."+m.Name] = m
						}
					}
				}
			} else if statement.Kind == "variable" && statement.Expression != nil && statement.Expression.Kind == "arrow_function" && statement.Expression.Function != nil && len(statement.Expression.Function.TypeParameters) > 0 {
				fn := *statement.Expression.Function
				fn.Name = statement.Name
				genericFuncs[statement.Name] = fn
			}
			if (statement.Kind == "function" || statement.Kind == "generator_function" || statement.Kind == "async_function" || statement.Kind == "async_generator_function" || statement.IsGenerator || statement.IsAsync) && statement.Name != "" {
				t := statement.Type
				if t == "" {
					t = statement.InferredType
				}
				funcTypes[statement.Name] = t
			}
		}
	}

	if len(genericFuncs) == 0 && len(genericClasses) == 0 && len(genericMethods) == 0 {
		return normalizeGenericArrayTypes(program), nil
	}

	specializedFuncs := map[string]bool{}
	specializedClasses := map[string]bool{}
	specializedMethods := map[string]bool{}

	funcInstances := map[string][]typescriptgo.SyntaxStatement{}
	classInstances := map[string][]typescriptgo.SyntaxStatement{}
	methodInstances := map[string][]typescriptgo.SyntaxMethod{}

	// Helper to request function specialization
	var requestFuncSpec func(name string, typeArgs []string, originFile string) string
	var requestClassSpec func(name string, typeArgs []string, originFile string) string
	var requestMethodSpec func(className, methodName string, typeArgs []string) string

	requestFuncSpec = func(name string, typeArgs []string, originFile string) string {
		if originFile == "" {
			for _, f := range program.Files {
				if !strings.HasSuffix(f.FileName, ".d.ts") {
					originFile = f.FileName
					break
				}
			}
		}
		fnTemplate, ok := genericFuncs[name]
		if !ok || len(typeArgs) != len(fnTemplate.TypeParameters) {
			return name
		}
		mangled := mangleGenericName(name, typeArgs)
		if specializedFuncs[mangled] {
			return mangled
		}
		specializedFuncs[mangled] = true

		subst := make(map[string]string, len(typeArgs))
		for i, param := range fnTemplate.TypeParameters {
			subst[param] = typeArgs[i]
		}

		specFn := cloneStatement(fnTemplate)
		specFn.Name = mangled
		specFn.TypeParameters = nil
		specFn.Type = substituteType(specFn.Type, subst)
		specFn.InferredType = substituteType(specFn.InferredType, subst)
		for i := range specFn.Parameters {
			specFn.Parameters[i].Type = substituteType(specFn.Parameters[i].Type, subst)
			specFn.Parameters[i].InferredType = substituteType(specFn.Parameters[i].InferredType, subst)
			if specFn.Parameters[i].Initializer != nil {
				specFn.Parameters[i].Initializer = cloneAndSubstituteExpr(specFn.Parameters[i].Initializer, subst)
			}
		}
		for i := range specFn.Body {
			specFn.Body[i] = cloneAndSubstituteStmt(specFn.Body[i], subst)
		}

		funcInstances[originFile] = append(funcInstances[originFile], specFn)

		// Scan specialized function body for other generic dependencies
		specEnv := map[string]string{}
		for _, p := range specFn.Parameters {
			if p.Type != "" {
				specEnv[p.Name] = p.Type
				scanTypeForGenerics(p.Type, originFile, genericClasses, requestClassSpec)
			}
		}
		for _, stmt := range specFn.Body {
			scanAndSpecializeStmt(stmt, originFile, specEnv, funcTypes, genericFuncs, genericClasses, genericMethods, requestFuncSpec, requestClassSpec, requestMethodSpec)
		}

		return mangled
	}

	requestMethodSpec = func(className, methodName string, typeArgs []string) string {
		key := className + "." + methodName
		mTemplate, ok := genericMethods[key]
		if !ok {
			key = className + ".static." + methodName
			mTemplate, ok = genericMethods[key]
		}
		if !ok && strings.Contains(className, "__") {
			baseName := className[:strings.Index(className, "__")]
			mTemplate, ok = genericMethods[baseName+"."+methodName]
			if !ok {
				mTemplate, ok = genericMethods[baseName+".static."+methodName]
			}
		}
		if !ok || len(typeArgs) != len(mTemplate.TypeParameters) {
			return methodName
		}
		mangled := mangleGenericName(methodName, typeArgs)
		fullMangled := className + "." + mangled
		if specializedMethods[fullMangled] {
			return mangled
		}
		specializedMethods[fullMangled] = true

		subst := make(map[string]string, len(typeArgs))
		for i, param := range mTemplate.TypeParameters {
			subst[param] = typeArgs[i]
		}

		specM := cloneMethod(mTemplate)
		specM.Name = mangled
		specM.TypeParameters = nil
		specM.Type = substituteType(specM.Type, subst)
		specM.InferredType = substituteType(specM.InferredType, subst)
		for i := range specM.Parameters {
			specM.Parameters[i].Type = substituteType(specM.Parameters[i].Type, subst)
			specM.Parameters[i].InferredType = substituteType(specM.Parameters[i].InferredType, subst)
			if specM.Parameters[i].Initializer != nil {
				specM.Parameters[i].Initializer = cloneAndSubstituteExpr(specM.Parameters[i].Initializer, subst)
			}
		}
		for i := range specM.Body {
			specM.Body[i] = cloneAndSubstituteStmt(specM.Body[i], subst)
		}

		methodInstances[className] = append(methodInstances[className], specM)
		if strings.Contains(className, "__") {
			baseName := className[:strings.Index(className, "__")]
			methodInstances[baseName] = append(methodInstances[baseName], specM)
		}
		return mangled
	}

	requestClassSpec = func(name string, typeArgs []string, originFile string) string {
		if originFile == "" {
			for _, f := range program.Files {
				if !strings.HasSuffix(f.FileName, ".d.ts") {
					originFile = f.FileName
					break
				}
			}
		}
		clsTemplate, ok := genericClasses[name]
		if !ok || len(typeArgs) != len(clsTemplate.TypeParameters) {
			return name
		}
		mangled := mangleGenericName(name, typeArgs)
		if specializedClasses[mangled] {
			return mangled
		}
		specializedClasses[mangled] = true

		subst := make(map[string]string, len(typeArgs))
		for i, param := range clsTemplate.TypeParameters {
			subst[param] = typeArgs[i]
		}

		specCls := cloneClass(clsTemplate)
		specCls.Name = mangled
		specCls.TypeParameters = nil
		specCls.Extends = substituteType(specCls.Extends, subst)
		for i := range specCls.Fields {
			specCls.Fields[i].Type = substituteType(specCls.Fields[i].Type, subst)
			specCls.Fields[i].InferredType = substituteType(specCls.Fields[i].InferredType, subst)
			if specCls.Fields[i].Initializer != nil {
				specCls.Fields[i].Initializer = cloneAndSubstituteExpr(specCls.Fields[i].Initializer, subst)
			}
		}
		if specCls.Constructor != nil {
			for i := range specCls.Constructor.Parameters {
				specCls.Constructor.Parameters[i].Type = substituteType(specCls.Constructor.Parameters[i].Type, subst)
				specCls.Constructor.Parameters[i].InferredType = substituteType(specCls.Constructor.Parameters[i].InferredType, subst)
				if specCls.Constructor.Parameters[i].Initializer != nil {
					specCls.Constructor.Parameters[i].Initializer = cloneAndSubstituteExpr(specCls.Constructor.Parameters[i].Initializer, subst)
				}
			}
			for i := range specCls.Constructor.Body {
				specCls.Constructor.Body[i] = cloneAndSubstituteStmt(specCls.Constructor.Body[i], subst)
			}
		}
		var concreteMethods []typescriptgo.SyntaxMethod
		for i := range specCls.Methods {
			if len(specCls.Methods[i].TypeParameters) > 0 {
				specCls.Methods[i].Type = substituteType(specCls.Methods[i].Type, subst)
				specCls.Methods[i].InferredType = substituteType(specCls.Methods[i].InferredType, subst)
				for j := range specCls.Methods[i].Parameters {
					specCls.Methods[i].Parameters[j].Type = substituteType(specCls.Methods[i].Parameters[j].Type, subst)
					specCls.Methods[i].Parameters[j].InferredType = substituteType(specCls.Methods[i].Parameters[j].InferredType, subst)
				}
				if specCls.Methods[i].IsStatic {
					genericMethods[specCls.Name+".static."+specCls.Methods[i].Name] = specCls.Methods[i]
				} else {
					genericMethods[specCls.Name+"."+specCls.Methods[i].Name] = specCls.Methods[i]
				}
				continue
			}
			mOrig := clsTemplate.Methods[i]
			isSelfExpanding := strings.HasPrefix(mOrig.Type, clsTemplate.Name+"<") && strings.Contains(mOrig.Type, "[]")
			if isSelfExpanding && currUsedMethods != nil && !currUsedMethods[clsTemplate.Name+"."+mOrig.Name] && !currUsedMethods[mOrig.Name] {
				continue
			}

			specCls.Methods[i].Type = substituteType(specCls.Methods[i].Type, subst)
			specCls.Methods[i].InferredType = substituteType(specCls.Methods[i].InferredType, subst)
			for j := range specCls.Methods[i].Parameters {
				specCls.Methods[i].Parameters[j].Type = substituteType(specCls.Methods[i].Parameters[j].Type, subst)
				specCls.Methods[i].Parameters[j].InferredType = substituteType(specCls.Methods[i].Parameters[j].InferredType, subst)
				if specCls.Methods[i].Parameters[j].Initializer != nil {
					specCls.Methods[i].Parameters[j].Initializer = cloneAndSubstituteExpr(specCls.Methods[i].Parameters[j].Initializer, subst)
				}
			}
			for j := range specCls.Methods[i].Body {
				specCls.Methods[i].Body[j] = cloneAndSubstituteStmt(specCls.Methods[i].Body[j], subst)
			}
			concreteMethods = append(concreteMethods, specCls.Methods[i])
		}
		specCls.Methods = concreteMethods

		kind := genericClassKinds[name]
		if kind == "" {
			kind = "class"
		}
		classInstances[originFile] = append(classInstances[originFile], typescriptgo.SyntaxStatement{
			Span:  clsTemplate.Span,
			Kind:  kind,
			Name:  mangled,
			Class: &specCls,
		})

		// Scan specialized class members for generic dependencies
		specEnv := map[string]string{}
		for _, f := range specCls.Fields {
			if f.Type != "" {
				scanTypeForGenerics(f.Type, originFile, genericClasses, requestClassSpec)
			}
		}
		if specCls.Constructor != nil {
			for _, p := range specCls.Constructor.Parameters {
				if p.Type != "" {
					specEnv[p.Name] = p.Type
					scanTypeForGenerics(p.Type, originFile, genericClasses, requestClassSpec)
				}
			}
			for _, stmt := range specCls.Constructor.Body {
				scanAndSpecializeStmt(stmt, originFile, specEnv, funcTypes, genericFuncs, genericClasses, genericMethods, requestFuncSpec, requestClassSpec, requestMethodSpec)
			}
		}
		for _, m := range specCls.Methods {
			if len(m.TypeParameters) > 0 {
				continue
			}
			mEnv := make(map[string]string, len(specEnv)+len(m.Parameters)+1)
			maps.Copy(mEnv, specEnv)
			mEnv["this"] = mangled
			for _, p := range m.Parameters {
				if p.Type != "" {
					mEnv[p.Name] = p.Type
					scanTypeForGenerics(p.Type, originFile, genericClasses, requestClassSpec)
				}
			}
			for _, stmt := range m.Body {
				scanAndSpecializeStmt(stmt, originFile, mEnv, funcTypes, genericFuncs, genericClasses, genericMethods, requestFuncSpec, requestClassSpec, requestMethodSpec)
			}
		}

		return mangled
	}

	// First pass: scan all type annotations and instantiations across files
	for _, file := range program.Files {
		var localEnv = map[string]string{}
		for _, stmt := range file.Syntax.Statements {
			scanAndSpecializeStmt(stmt, file.FileName, localEnv, funcTypes, genericFuncs, genericClasses, genericMethods, requestFuncSpec, requestClassSpec, requestMethodSpec)
		}
	}

	// Rebuild program files with specialized instances replacing generic templates
	newFiles := make([]typescriptgo.SourceFile, 0, len(program.Files))
	for _, file := range program.Files {
		var newStmts []typescriptgo.SyntaxStatement
		var fileEnv = map[string]string{}
		for _, stmt := range file.Syntax.Statements {
			if stmt.Kind == "function" && len(stmt.TypeParameters) > 0 {
				continue // Skip generic function template
			}
			if (stmt.Kind == "class" || stmt.Kind == "interface" || stmt.Kind == "type_alias") && stmt.Class != nil && len(stmt.Class.TypeParameters) > 0 {
				continue // Skip generic class/interface/type_alias template
			}
			if stmt.Kind == "variable" && stmt.Expression != nil && stmt.Expression.Kind == "arrow_function" && stmt.Expression.Function != nil && len(stmt.Expression.Function.TypeParameters) > 0 {
				continue // Skip generic arrow function template
			}
			rewritten := rewriteStatementTypes(stmt, fileEnv, genericFuncs, genericClasses, genericMethods, requestFuncSpec, requestClassSpec, requestMethodSpec, file.FileName)
			if rewritten.Class != nil {
				baseName := rewritten.Class.Name
				if idx := strings.Index(baseName, "__"); idx != -1 {
					baseName = baseName[:idx]
				}
				seen := map[string]bool{}
				for _, m := range rewritten.Class.Methods {
					seen[m.Name] = true
				}
				for _, ms := range [][]typescriptgo.SyntaxMethod{methodInstances[rewritten.Class.Name], methodInstances[baseName]} {
					for _, m := range ms {
						if !seen[m.Name] {
							seen[m.Name] = true
							rewritten.Class.Methods = append(rewritten.Class.Methods, rewriteMethod(m, fileEnv, genericFuncs, genericClasses, genericMethods, requestFuncSpec, requestClassSpec, requestMethodSpec, file.FileName))
						}
					}
				}
			}
			newStmts = append(newStmts, rewritten)
		}
		// Append specialized class instances then function instances
		for i := 0; i < len(classInstances[file.FileName]); i++ {
			clsStmt := classInstances[file.FileName][i]
			clsStmt = rewriteStatementTypes(clsStmt, fileEnv, genericFuncs, genericClasses, genericMethods, requestFuncSpec, requestClassSpec, requestMethodSpec, file.FileName)
			if clsStmt.Class != nil {
				baseName := clsStmt.Class.Name
				if idx := strings.Index(baseName, "__"); idx != -1 {
					baseName = baseName[:idx]
				}
				seen := map[string]bool{}
				for _, m := range clsStmt.Class.Methods {
					seen[m.Name] = true
				}
				for _, ms := range [][]typescriptgo.SyntaxMethod{methodInstances[clsStmt.Class.Name], methodInstances[baseName]} {
					for _, m := range ms {
						if !seen[m.Name] {
							seen[m.Name] = true
							clsStmt.Class.Methods = append(clsStmt.Class.Methods, rewriteMethod(m, fileEnv, genericFuncs, genericClasses, genericMethods, requestFuncSpec, requestClassSpec, requestMethodSpec, file.FileName))
						}
					}
				}
			}
			newStmts = append(newStmts, clsStmt)
		}
		for i := 0; i < len(funcInstances[file.FileName]); i++ {
			fnStmt := funcInstances[file.FileName][i]
			rewrittenFn := rewriteStatementTypes(fnStmt, fileEnv, genericFuncs, genericClasses, genericMethods, requestFuncSpec, requestClassSpec, requestMethodSpec, file.FileName)
			newStmts = append(newStmts, rewrittenFn)
		}

		for sIdx := range newStmts {
			if newStmts[sIdx].Class != nil {
				cls := newStmts[sIdx].Class
				baseName := cls.Name
				if idx := strings.Index(baseName, "__"); idx != -1 {
					baseName = baseName[:idx]
				}
				seen := map[string]bool{}
				for _, m := range cls.Methods {
					seen[m.Name] = true
				}
				for _, ms := range [][]typescriptgo.SyntaxMethod{methodInstances[cls.Name], methodInstances[baseName]} {
					for _, m := range ms {
						if !seen[m.Name] {
							seen[m.Name] = true
							cls.Methods = append(cls.Methods, rewriteMethod(m, fileEnv, genericFuncs, genericClasses, genericMethods, requestFuncSpec, requestClassSpec, requestMethodSpec, file.FileName))
						}
					}
				}
			}
		}

		fileCopy := file
		fileCopy.Syntax.Statements = newStmts
		newFiles = append(newFiles, fileCopy)
	}

	program.Files = newFiles
	return normalizeGenericArrayTypes(program), nil
}

func scanAndSpecializeStmt(stmt typescriptgo.SyntaxStatement, fileName string, env map[string]string, funcTypes map[string]string, genericFuncs map[string]typescriptgo.SyntaxStatement, genericClasses map[string]typescriptgo.SyntaxClass, genericMethods map[string]typescriptgo.SyntaxMethod, reqFn func(string, []string, string) string, reqCls func(string, []string, string) string, reqMethod func(string, string, []string) string) {
	if stmt.Kind == "function" && len(stmt.TypeParameters) > 0 {
		return
	}
	if stmt.Kind == "class" && stmt.Class != nil && len(stmt.Class.TypeParameters) > 0 {
		return
	}
	if stmt.Type != "" {
		scanTypeForGenerics(stmt.Type, fileName, genericClasses, reqCls)
	}
	if stmt.Kind == "variable" && stmt.Name != "" {
		if stmt.Type != "" {
			env[stmt.Name] = stmt.Type
		} else if stmt.Expression != nil {
			env[stmt.Name] = inferExprType(stmt.Expression, env, funcTypes)
		}
	}
	if stmt.Expression != nil {
		scanAndSpecializeExpr(stmt.Expression, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	if stmt.Left != nil {
		scanAndSpecializeExpr(stmt.Left, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	if stmt.Right != nil {
		scanAndSpecializeExpr(stmt.Right, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
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
	for _, s := range stmt.Catch {
		scanAndSpecializeStmt(s, fileName, env, funcTypes, genericFuncs, genericClasses, genericMethods, reqFn, reqCls, reqMethod)
	}
	for _, s := range stmt.Finally {
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
					cleanT := strings.TrimPrefix(t, "object:")
					if idx := strings.Index(cleanT, "<"); idx != -1 {
						cleanT = cleanT[:idx]
					}
					if idx := strings.Index(cleanT, "__"); idx != -1 {
						cleanT = cleanT[:idx]
					}
					clsName = cleanT
				}
			} else {
				isInstance = true
				t := inferExprType(expr.Left.Left, env, funcTypes)
				if t != "" {
					cleanT := strings.TrimPrefix(t, "object:")
					if idx := strings.Index(cleanT, "<"); idx != -1 {
						cleanT = cleanT[:idx]
					}
					if idx := strings.Index(cleanT, "__"); idx != -1 {
						cleanT = cleanT[:idx]
					}
					clsName = cleanT
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
			if mTemplate, ok := genericMethods[lookupKey]; ok {
				typeArgs := expr.TypeArguments
				if len(typeArgs) == 0 {
					typeArgs = inferTypeArgsForMethod(mTemplate, mTemplate.TypeParameters, expr.Arguments, env, funcTypes)
				}
				if len(typeArgs) == len(mTemplate.TypeParameters) {
					reqMethod(clsName, methodName, typeArgs)
				}
			}
			if !isInstance && expr.Left.Left.Kind == "identifier" {
				if clsTemplate, ok := genericClasses[clsName]; ok {
					typeArgs := expr.TypeArguments
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
		scanTypeForGenerics(clean[:len(clean)-2], fileName, genericClasses, reqCls)
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
