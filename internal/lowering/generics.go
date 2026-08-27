package lowering

import (
	"maps"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
)

var (
	currGenericFuncs       map[string]typescriptgo.SyntaxStatement
	currGenericClasses     map[string]typescriptgo.SyntaxClass
	currGenericMethods     map[string]typescriptgo.SyntaxMethod
	currGenericTypeAliases map[string]typescriptgo.SyntaxStatement
	currFuncTypes          map[string]string
	currUsedMethods        map[string]bool
)

// SpecializeGenerics monomorphizes generic functions and classes based on
// concrete type arguments at call sites, instantiations, and type annotations.
func SpecializeGenerics(program frontend.Program) (frontend.Program, error) {
	genericFuncs := map[string]typescriptgo.SyntaxStatement{}
	genericClasses := map[string]typescriptgo.SyntaxClass{}
	genericClassKinds := map[string]string{}
	genericMethods := map[string]typescriptgo.SyntaxMethod{}
	genericTypeAliases := map[string]typescriptgo.SyntaxStatement{}
	funcTypes := map[string]string{}
	usedMethods := map[string]bool{}

	currGenericFuncs = genericFuncs
	currGenericClasses = genericClasses
	currGenericMethods = genericMethods
	currGenericTypeAliases = genericTypeAliases
	currFuncTypes = funcTypes
	currUsedMethods = usedMethods
	defer func() {
		currGenericFuncs = nil
		currGenericClasses = nil
		currGenericMethods = nil
		currGenericTypeAliases = nil
		currFuncTypes = nil
		currUsedMethods = nil
	}()

	for _, file := range program.Files {
		if strings.HasSuffix(file.FileName, ".d.ts") {
			continue
		}
		for _, statement := range file.Syntax.Statements {
			tParams := statement.TypeParameters
			if len(tParams) == 0 && statement.Class != nil {
				tParams = statement.Class.TypeParameters
			}
			if statement.Kind == "type_alias" && len(tParams) > 0 {
				if statement.Class == nil || len(statement.Class.Fields) == 0 {
					genericTypeAliases[statement.Name] = statement
					continue
				}
			}
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
		for _, arg := range typeArgs {
			if len(arg) == 1 && arg >= "A" && arg <= "Z" {
				return name
			}
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
		for _, arg := range typeArgs {
			if len(arg) == 1 && arg >= "A" && arg <= "Z" {
				return methodName
			}
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

		originFile := ""
		for _, f := range program.Files {
			if !strings.HasSuffix(f.FileName, ".d.ts") {
				originFile = f.FileName
				break
			}
		}
		specEnv := map[string]string{"this": className}
		for _, p := range specM.Parameters {
			if p.Type != "" {
				specEnv[p.Name] = p.Type
				scanTypeForGenerics(p.Type, originFile, genericClasses, requestClassSpec)
			}
		}
		for _, stmt := range specM.Body {
			scanAndSpecializeStmt(stmt, originFile, specEnv, funcTypes, genericFuncs, genericClasses, genericMethods, requestFuncSpec, requestClassSpec, requestMethodSpec)
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
		for _, arg := range typeArgs {
			if len(arg) == 1 && arg >= "A" && arg <= "Z" {
				return name
			}
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
					if specCls.Methods[i].Parameters[j].Initializer != nil {
						specCls.Methods[i].Parameters[j].Initializer = cloneAndSubstituteExpr(specCls.Methods[i].Parameters[j].Initializer, subst)
					}
				}
				for j := range specCls.Methods[i].Body {
					specCls.Methods[i].Body[j] = cloneAndSubstituteStmt(specCls.Methods[i].Body[j], subst)
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
			if stmt.Class != nil && len(stmt.Class.TypeParameters) > 0 {
				continue
			}
			if (stmt.Kind == "function" || stmt.Kind == "generator_function" || stmt.Kind == "async_function") && len(stmt.TypeParameters) > 0 {
				continue
			}
			if stmt.Kind == "variable" && stmt.Expression != nil && stmt.Expression.Kind == "arrow_function" && stmt.Expression.Function != nil && len(stmt.Expression.Function.TypeParameters) > 0 {
				continue
			}
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
				classEnv := make(map[string]string, len(fileEnv)+1)
				maps.Copy(classEnv, fileEnv)
				classEnv["this"] = rewritten.Class.Name
				for _, ms := range [][]typescriptgo.SyntaxMethod{methodInstances[rewritten.Class.Name], methodInstances[baseName]} {
					for _, m := range ms {
						if !seen[m.Name] {
							seen[m.Name] = true
							rewritten.Class.Methods = append(rewritten.Class.Methods, rewriteMethod(m, classEnv, genericFuncs, genericClasses, genericMethods, requestFuncSpec, requestClassSpec, requestMethodSpec, file.FileName))
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
				classEnv := make(map[string]string, len(fileEnv)+1)
				maps.Copy(classEnv, fileEnv)
				classEnv["this"] = clsStmt.Class.Name
				for _, ms := range [][]typescriptgo.SyntaxMethod{methodInstances[clsStmt.Class.Name], methodInstances[baseName]} {
					for _, m := range ms {
						if !seen[m.Name] {
							seen[m.Name] = true
							clsStmt.Class.Methods = append(clsStmt.Class.Methods, rewriteMethod(m, classEnv, genericFuncs, genericClasses, genericMethods, requestFuncSpec, requestClassSpec, requestMethodSpec, file.FileName))
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
