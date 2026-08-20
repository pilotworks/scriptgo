package lowering

import (
	"strings"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
)

func rewriteStatementTypes(stmt typescriptgo.SyntaxStatement, genericFuncs map[string]typescriptgo.SyntaxStatement, genericClasses map[string]typescriptgo.SyntaxClass, reqFn func(string, []string, string) string, reqCls func(string, []string, string) string, fileName string) typescriptgo.SyntaxStatement {
	res := cloneStatement(stmt)
	res.Type = rewriteTypeString(res.Type)
	for i := range res.Parameters {
		res.Parameters[i].Type = rewriteTypeString(res.Parameters[i].Type)
		if res.Parameters[i].Initializer != nil {
			res.Parameters[i].Initializer = rewriteExpr(res.Parameters[i].Initializer, genericFuncs, genericClasses, reqFn, reqCls, fileName)
		}
	}
	if res.Expression != nil {
		res.Expression = rewriteExpr(res.Expression, genericFuncs, genericClasses, reqFn, reqCls, fileName)
	}
	if res.Left != nil {
		res.Left = rewriteExpr(res.Left, genericFuncs, genericClasses, reqFn, reqCls, fileName)
	}
	if res.Right != nil {
		res.Right = rewriteExpr(res.Right, genericFuncs, genericClasses, reqFn, reqCls, fileName)
	}
	for i := range res.Body {
		res.Body[i] = rewriteStatementTypes(res.Body[i], genericFuncs, genericClasses, reqFn, reqCls, fileName)
	}
	for i := range res.Then {
		res.Then[i] = rewriteStatementTypes(res.Then[i], genericFuncs, genericClasses, reqFn, reqCls, fileName)
	}
	for i := range res.Else {
		res.Else[i] = rewriteStatementTypes(res.Else[i], genericFuncs, genericClasses, reqFn, reqCls, fileName)
	}
	for i := range res.Catch {
		res.Catch[i] = rewriteStatementTypes(res.Catch[i], genericFuncs, genericClasses, reqFn, reqCls, fileName)
	}
	for i := range res.Finally {
		res.Finally[i] = rewriteStatementTypes(res.Finally[i], genericFuncs, genericClasses, reqFn, reqCls, fileName)
	}
	if res.Class != nil {
		for i := range res.Class.Fields {
			res.Class.Fields[i].Type = rewriteTypeString(res.Class.Fields[i].Type)
			if res.Class.Fields[i].Initializer != nil {
				res.Class.Fields[i].Initializer = rewriteExpr(res.Class.Fields[i].Initializer, genericFuncs, genericClasses, reqFn, reqCls, fileName)
			}
		}
		if res.Class.Constructor != nil {
			for i := range res.Class.Constructor.Parameters {
				res.Class.Constructor.Parameters[i].Type = rewriteTypeString(res.Class.Constructor.Parameters[i].Type)
				if res.Class.Constructor.Parameters[i].Initializer != nil {
					res.Class.Constructor.Parameters[i].Initializer = rewriteExpr(res.Class.Constructor.Parameters[i].Initializer, genericFuncs, genericClasses, reqFn, reqCls, fileName)
				}
			}
			for i := range res.Class.Constructor.Body {
				res.Class.Constructor.Body[i] = rewriteStatementTypes(res.Class.Constructor.Body[i], genericFuncs, genericClasses, reqFn, reqCls, fileName)
			}
		}
		for i := range res.Class.Methods {
			res.Class.Methods[i].Type = rewriteTypeString(res.Class.Methods[i].Type)
			for j := range res.Class.Methods[i].Parameters {
				res.Class.Methods[i].Parameters[j].Type = rewriteTypeString(res.Class.Methods[i].Parameters[j].Type)
				if res.Class.Methods[i].Parameters[j].Initializer != nil {
					res.Class.Methods[i].Parameters[j].Initializer = rewriteExpr(res.Class.Methods[i].Parameters[j].Initializer, genericFuncs, genericClasses, reqFn, reqCls, fileName)
				}
			}
			for j := range res.Class.Methods[i].Body {
				res.Class.Methods[i].Body[j] = rewriteStatementTypes(res.Class.Methods[i].Body[j], genericFuncs, genericClasses, reqFn, reqCls, fileName)
			}
		}
	}
	return res
}

func rewriteExpr(expr *typescriptgo.SyntaxExpression, genericFuncs map[string]typescriptgo.SyntaxStatement, genericClasses map[string]typescriptgo.SyntaxClass, reqFn func(string, []string, string) string, reqCls func(string, []string, string) string, fileName string) *typescriptgo.SyntaxExpression {
	if expr == nil {
		return nil
	}
	res := cloneExpr(expr)
	if res.Kind == "call" && res.Left != nil && res.Left.Kind == "identifier" {
		fnName := res.Left.Text
		if fnTemplate, ok := genericFuncs[fnName]; ok {
			typeArgs := res.TypeArguments
			if len(typeArgs) == 0 {
				typeArgs = inferTypeArgsForFunc(fnTemplate, res.Arguments, nil)
			}
			if len(typeArgs) == len(fnTemplate.TypeParameters) {
				mangled := reqFn(fnName, typeArgs, fileName)
				res.Left.Text = mangled
				res.TypeArguments = nil
			}
		}
	}
	if res.Kind == "new" && res.Left != nil && res.Left.Kind == "identifier" {
		clsName := res.Left.Text
		if clsTemplate, ok := genericClasses[clsName]; ok {
			typeArgs := res.TypeArguments
			if len(typeArgs) == 0 && clsTemplate.Constructor != nil {
				typeArgs = inferTypeArgsForClass(clsTemplate, res.Arguments, nil)
			}
			if len(typeArgs) == len(clsTemplate.TypeParameters) {
				mangled := reqCls(clsName, typeArgs, fileName)
				res.Left.Text = mangled
				res.TypeArguments = nil
			}
		}
	}
	if res.Left != nil {
		res.Left = rewriteExpr(res.Left, genericFuncs, genericClasses, reqFn, reqCls, fileName)
	}
	if res.Right != nil {
		res.Right = rewriteExpr(res.Right, genericFuncs, genericClasses, reqFn, reqCls, fileName)
	}
	for i := range res.Arguments {
		res.Arguments[i] = rewriteExpr(res.Arguments[i], genericFuncs, genericClasses, reqFn, reqCls, fileName)
	}
	if res.WhenTrue != nil {
		res.WhenTrue = rewriteExpr(res.WhenTrue, genericFuncs, genericClasses, reqFn, reqCls, fileName)
	}
	if res.WhenFalse != nil {
		res.WhenFalse = rewriteExpr(res.WhenFalse, genericFuncs, genericClasses, reqFn, reqCls, fileName)
	}
	return res
}

func rewriteTypeString(typ string) string {
	if typ == "" {
		return typ
	}
	if strings.HasSuffix(typ, "[]") {
		return rewriteTypeString(typ[:len(typ)-2]) + "[]"
	}
	hasObj := strings.HasPrefix(typ, "object:")
	clean := strings.TrimPrefix(typ, "object:")
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
	}
	return stmt
}
