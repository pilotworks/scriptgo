package lowering

import (
	"strings"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
)

// SpecializeGenerics monomorphizes generic functions and classes based on
// concrete type arguments at call sites, instantiations, and type annotations.
func SpecializeGenerics(program frontend.Program) (frontend.Program, error) {
	genericFuncs := map[string]typescriptgo.SyntaxStatement{}
	genericClasses := map[string]typescriptgo.SyntaxClass{}

	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if statement.Kind == "function" && len(statement.TypeParameters) > 0 {
				genericFuncs[statement.Name] = statement
			} else if statement.Kind == "class" && statement.Class != nil && len(statement.Class.TypeParameters) > 0 {
				genericClasses[statement.Class.Name] = *statement.Class
			} else if statement.Kind == "variable" && statement.Expression != nil && statement.Expression.Kind == "arrow_function" && statement.Expression.Function != nil && len(statement.Expression.Function.TypeParameters) > 0 {
				fn := *statement.Expression.Function
				fn.Name = statement.Name
				genericFuncs[statement.Name] = fn
			}
		}
	}

	if len(genericFuncs) == 0 && len(genericClasses) == 0 {
		return normalizeGenericArrayTypes(program), nil
	}

	specializedFuncs := map[string]bool{}
	specializedClasses := map[string]bool{}

	funcInstances := map[string][]typescriptgo.SyntaxStatement{}
	classInstances := map[string][]typescriptgo.SyntaxStatement{}

	// Helper to request function specialization
	var requestFuncSpec func(name string, typeArgs []string, originFile string) string
	var requestClassSpec func(name string, typeArgs []string, originFile string) string

	requestFuncSpec = func(name string, typeArgs []string, originFile string) string {
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
		for i := range specFn.Parameters {
			specFn.Parameters[i].Type = substituteType(specFn.Parameters[i].Type, subst)
			if specFn.Parameters[i].Initializer != nil {
				specFn.Parameters[i].Initializer = cloneAndSubstituteExpr(specFn.Parameters[i].Initializer, subst)
			}
		}
		for i := range specFn.Body {
			specFn.Body[i] = cloneAndSubstituteStmt(specFn.Body[i], subst)
		}

		funcInstances[originFile] = append(funcInstances[originFile], specFn)
		return mangled
	}

	requestClassSpec = func(name string, typeArgs []string, originFile string) string {
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
		for i := range specCls.Fields {
			specCls.Fields[i].Type = substituteType(specCls.Fields[i].Type, subst)
			if specCls.Fields[i].Initializer != nil {
				specCls.Fields[i].Initializer = cloneAndSubstituteExpr(specCls.Fields[i].Initializer, subst)
			}
		}
		if specCls.Constructor != nil {
			for i := range specCls.Constructor.Parameters {
				specCls.Constructor.Parameters[i].Type = substituteType(specCls.Constructor.Parameters[i].Type, subst)
				if specCls.Constructor.Parameters[i].Initializer != nil {
					specCls.Constructor.Parameters[i].Initializer = cloneAndSubstituteExpr(specCls.Constructor.Parameters[i].Initializer, subst)
				}
			}
			for i := range specCls.Constructor.Body {
				specCls.Constructor.Body[i] = cloneAndSubstituteStmt(specCls.Constructor.Body[i], subst)
			}
		}
		for i := range specCls.Methods {
			specCls.Methods[i].Type = substituteType(specCls.Methods[i].Type, subst)
			for j := range specCls.Methods[i].Parameters {
				specCls.Methods[i].Parameters[j].Type = substituteType(specCls.Methods[i].Parameters[j].Type, subst)
				if specCls.Methods[i].Parameters[j].Initializer != nil {
					specCls.Methods[i].Parameters[j].Initializer = cloneAndSubstituteExpr(specCls.Methods[i].Parameters[j].Initializer, subst)
				}
			}
			for j := range specCls.Methods[i].Body {
				specCls.Methods[i].Body[j] = cloneAndSubstituteStmt(specCls.Methods[i].Body[j], subst)
			}
		}

		classInstances[originFile] = append(classInstances[originFile], typescriptgo.SyntaxStatement{
			Span:  clsTemplate.Span,
			Kind:  "class",
			Name:  mangled,
			Class: &specCls,
		})
		return mangled
	}

	// First pass: scan all type annotations and instantiations across files
	for _, file := range program.Files {
		var localEnv = map[string]string{}
		for _, stmt := range file.Syntax.Statements {
			scanAndSpecializeStmt(stmt, file.FileName, localEnv, genericFuncs, genericClasses, requestFuncSpec, requestClassSpec)
		}
	}

	// Rebuild program files with specialized instances replacing generic templates
	newFiles := make([]typescriptgo.SourceFile, 0, len(program.Files))
	for _, file := range program.Files {
		var newStmts []typescriptgo.SyntaxStatement
		for _, stmt := range file.Syntax.Statements {
			if stmt.Kind == "function" && len(stmt.TypeParameters) > 0 {
				continue // Skip generic function template
			}
			if stmt.Kind == "class" && stmt.Class != nil && len(stmt.Class.TypeParameters) > 0 {
				continue // Skip generic class template
			}
			if stmt.Kind == "variable" && stmt.Expression != nil && stmt.Expression.Kind == "arrow_function" && stmt.Expression.Function != nil && len(stmt.Expression.Function.TypeParameters) > 0 {
				continue // Skip generic arrow function template
			}
			rewritten := rewriteStatementTypes(stmt, genericFuncs, genericClasses, requestFuncSpec, requestClassSpec, file.FileName)
			newStmts = append(newStmts, rewritten)
		}
		// Append specialized class instances then function instances
		newStmts = append(newStmts, classInstances[file.FileName]...)
		newStmts = append(newStmts, funcInstances[file.FileName]...)

		fileCopy := file
		fileCopy.Syntax.Statements = newStmts
		newFiles = append(newFiles, fileCopy)
	}

	program.Files = newFiles
	return normalizeGenericArrayTypes(program), nil
}

func scanAndSpecializeStmt(stmt typescriptgo.SyntaxStatement, fileName string, env map[string]string, genericFuncs map[string]typescriptgo.SyntaxStatement, genericClasses map[string]typescriptgo.SyntaxClass, reqFn func(string, []string, string) string, reqCls func(string, []string, string) string) {
	if stmt.Type != "" {
		scanTypeForGenerics(stmt.Type, fileName, genericClasses, reqCls)
	}
	if stmt.Kind == "variable" && stmt.Name != "" {
		if stmt.Type != "" {
			env[stmt.Name] = stmt.Type
		} else if stmt.Expression != nil {
			env[stmt.Name] = inferExprType(stmt.Expression, env)
		}
	}
	if stmt.Expression != nil {
		scanAndSpecializeExpr(stmt.Expression, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
	if stmt.Left != nil {
		scanAndSpecializeExpr(stmt.Left, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
	if stmt.Right != nil {
		scanAndSpecializeExpr(stmt.Right, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
	for _, p := range stmt.Parameters {
		if p.Type != "" {
			env[p.Name] = p.Type
			scanTypeForGenerics(p.Type, fileName, genericClasses, reqCls)
		}
		if p.Initializer != nil {
			scanAndSpecializeExpr(p.Initializer, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
		}
	}
	for _, s := range stmt.Body {
		scanAndSpecializeStmt(s, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
	for _, s := range stmt.Then {
		scanAndSpecializeStmt(s, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
	for _, s := range stmt.Else {
		scanAndSpecializeStmt(s, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
	for _, s := range stmt.Catch {
		scanAndSpecializeStmt(s, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
	for _, s := range stmt.Finally {
		scanAndSpecializeStmt(s, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
	if stmt.Class != nil {
		for _, f := range stmt.Class.Fields {
			if f.Type != "" {
				scanTypeForGenerics(f.Type, fileName, genericClasses, reqCls)
			}
			if f.Initializer != nil {
				scanAndSpecializeExpr(f.Initializer, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
			}
		}
		if stmt.Class.Constructor != nil {
			for _, p := range stmt.Class.Constructor.Parameters {
				if p.Type != "" {
					scanTypeForGenerics(p.Type, fileName, genericClasses, reqCls)
				}
			}
			for _, s := range stmt.Class.Constructor.Body {
				scanAndSpecializeStmt(s, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
			}
		}
		for _, m := range stmt.Class.Methods {
			if m.Type != "" {
				scanTypeForGenerics(m.Type, fileName, genericClasses, reqCls)
			}
			for _, p := range m.Parameters {
				if p.Type != "" {
					scanTypeForGenerics(p.Type, fileName, genericClasses, reqCls)
				}
			}
			for _, s := range m.Body {
				scanAndSpecializeStmt(s, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
			}
		}
	}
}

func scanAndSpecializeExpr(expr *typescriptgo.SyntaxExpression, fileName string, env map[string]string, genericFuncs map[string]typescriptgo.SyntaxStatement, genericClasses map[string]typescriptgo.SyntaxClass, reqFn func(string, []string, string) string, reqCls func(string, []string, string) string) {
	if expr == nil {
		return
	}
	if expr.Kind == "call" && expr.Left != nil && expr.Left.Kind == "identifier" {
		fnName := expr.Left.Text
		if fnTemplate, ok := genericFuncs[fnName]; ok {
			typeArgs := expr.TypeArguments
			if len(typeArgs) == 0 {
				typeArgs = inferTypeArgsForFunc(fnTemplate, expr.Arguments, env)
			}
			if len(typeArgs) == len(fnTemplate.TypeParameters) {
				reqFn(fnName, typeArgs, fileName)
			}
		}
	}
	if expr.Kind == "new" && expr.Left != nil && expr.Left.Kind == "identifier" {
		clsName := expr.Left.Text
		if clsTemplate, ok := genericClasses[clsName]; ok {
			typeArgs := expr.TypeArguments
			if len(typeArgs) == 0 && clsTemplate.Constructor != nil {
				typeArgs = inferTypeArgsForClass(clsTemplate, expr.Arguments, env)
			}
			if len(typeArgs) == len(clsTemplate.TypeParameters) {
				reqCls(clsName, typeArgs, fileName)
			}
		}
	}

	if expr.Left != nil {
		scanAndSpecializeExpr(expr.Left, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
	if expr.Right != nil {
		scanAndSpecializeExpr(expr.Right, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
	for _, arg := range expr.Arguments {
		scanAndSpecializeExpr(arg, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
	if expr.WhenTrue != nil {
		scanAndSpecializeExpr(expr.WhenTrue, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
	if expr.WhenFalse != nil {
		scanAndSpecializeExpr(expr.WhenFalse, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
	if expr.Function != nil {
		scanAndSpecializeStmt(*expr.Function, fileName, env, genericFuncs, genericClasses, reqFn, reqCls)
	}
}

func scanTypeForGenerics(typ, fileName string, genericClasses map[string]typescriptgo.SyntaxClass, reqCls func(string, []string, string) string) {
	clean := strings.TrimPrefix(typ, "object:")
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

func inferTypeArgsForFunc(fnTemplate typescriptgo.SyntaxStatement, args []*typescriptgo.SyntaxExpression, env map[string]string) []string {
	inferred := map[string]string{}
	for i, param := range fnTemplate.Parameters {
		if i < len(args) {
			argType := inferExprType(args[i], env)
			matchTypeParam(param.Type, argType, inferred)
		}
	}
	var res []string
	for _, tp := range fnTemplate.TypeParameters {
		if t, ok := inferred[tp]; ok && t != "" {
			res = append(res, t)
		} else {
			res = append(res, "number") // Fallback default
		}
	}
	return res
}

func inferTypeArgsForClass(clsTemplate typescriptgo.SyntaxClass, args []*typescriptgo.SyntaxExpression, env map[string]string) []string {
	inferred := map[string]string{}
	if clsTemplate.Constructor != nil {
		for i, param := range clsTemplate.Constructor.Parameters {
			if i < len(args) {
				argType := inferExprType(args[i], env)
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
	if !strings.Contains(paramType, "<") && !strings.Contains(paramType, "[]") {
		inferred[paramType] = argType
	}
}

func inferExprType(expr *typescriptgo.SyntaxExpression, env map[string]string) string {
	if expr == nil {
		return ""
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
		if t, ok := env[expr.Text]; ok {
			return t
		}
		return ""
	case "array":
		if len(expr.Arguments) > 0 {
			elemType := inferExprType(expr.Arguments[0], env)
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
	}
	return ""
}
