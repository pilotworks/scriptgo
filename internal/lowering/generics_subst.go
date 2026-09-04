package lowering

import (
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
)

func isBuiltinGeneric(name string) bool {
	switch name {
	case "Map", "Set", "Promise", "AsyncGenerator", "Generator", "Iterator", "Iterable", "AsyncIterable", "AsyncIterator", "Record", "Partial", "Readonly", "Pick", "Omit", "Exclude", "Extract", "NonNullable", "ReturnType", "InstanceType", "Parameters", "ConstructorParameters":
		return true
	default:
		return false
	}
}

func mangleGenericName(name string, typeArgs []string) string {
	if len(typeArgs) == 0 {
		return name
	}
	var cleanArgs []string
	for _, arg := range typeArgs {
		clean := strings.ReplaceAll(arg, "[]", "_arr")
		clean = strings.ReplaceAll(clean, "=>", "_to_")
		clean = strings.ReplaceAll(clean, "=", "_")
		clean = strings.ReplaceAll(clean, "<", "_")
		clean = strings.ReplaceAll(clean, ">", "_")
		clean = strings.ReplaceAll(clean, ", ", "_")
		clean = strings.ReplaceAll(clean, ",", "_")
		clean = strings.ReplaceAll(clean, "object:", "")
		clean = strings.ReplaceAll(clean, "{", "")
		clean = strings.ReplaceAll(clean, "}", "")
		clean = strings.ReplaceAll(clean, "(", "")
		clean = strings.ReplaceAll(clean, ")", "")
		clean = strings.ReplaceAll(clean, "[", "")
		clean = strings.ReplaceAll(clean, "]", "")
		clean = strings.ReplaceAll(clean, ":", "_")
		clean = strings.ReplaceAll(clean, ";", "_")
		clean = strings.ReplaceAll(clean, "|", "_")
		clean = strings.ReplaceAll(clean, "&", "_")
		clean = strings.ReplaceAll(clean, "?", "_")
		clean = strings.ReplaceAll(clean, "!", "_")
		clean = strings.ReplaceAll(clean, "+", "_")
		clean = strings.ReplaceAll(clean, "-", "_")
		clean = strings.ReplaceAll(clean, "*", "_")
		clean = strings.ReplaceAll(clean, "/", "_")
		clean = strings.ReplaceAll(clean, ".", "_")
		clean = strings.ReplaceAll(clean, "\"", "")
		clean = strings.ReplaceAll(clean, "'", "")
		clean = strings.ReplaceAll(clean, " ", "_")
		for strings.Contains(clean, "__") {
			clean = strings.ReplaceAll(clean, "__", "_")
		}
		clean = strings.Trim(clean, "_")
		cleanArgs = append(cleanArgs, clean)
	}
	return name + "__" + strings.Join(cleanArgs, "_")
}

func splitTypeArguments(s string) []string {
	var res []string
	depth := 0
	start := 0
	for i, r := range s {
		if r == '<' {
			depth++
		} else if r == '>' {
			depth--
		} else if r == ',' && depth == 0 {
			res = append(res, strings.TrimSpace(s[start:i]))
			start = i + 1
		}
	}
	if start < len(s) {
		res = append(res, strings.TrimSpace(s[start:]))
	}
	return res
}

func substituteType(typ string, subst map[string]string) string {
	if typ == "" {
		return ""
	}
	if val, ok := subst[typ]; ok {
		return val
	}
	if strings.HasSuffix(typ, "[]") {
		elem := typ[:len(typ)-2]
		hadParens := false
		if strings.HasPrefix(elem, "(") && strings.HasSuffix(elem, ")") {
			hadParens = true
			elem = elem[1 : len(elem)-1]
		}
		substElem := substituteType(elem, subst)
		if hadParens || strings.Contains(substElem, "|") {
			return "(" + substElem + ")[]"
		}
		return substElem + "[]"
	}
	hasObj := strings.HasPrefix(typ, "object:")
	clean := strings.TrimPrefix(typ, "object:")
	if strings.Contains(clean, "=>") {
		depth := 0
		splitIdx := -1
	scanLoop:
		for i := 0; i < len(clean)-1; i++ {
			switch clean[i] {
			case '(', '<', '{', '[':
				depth++
			case ')', '>', '}', ']':
				if depth > 0 {
					depth--
				}
			case '=':
				if depth == 0 && clean[i+1] == '>' {
					splitIdx = i
					break scanLoop
				}
			}
		}
		if splitIdx != -1 {
			paramPart := strings.TrimSpace(clean[:splitIdx])
			retPart := strings.TrimSpace(clean[splitIdx+2:])
			substRet := substituteType(retPart, subst)
			substParams := substituteType(paramPart, subst)
			for k, v := range subst {
				substParams = strings.ReplaceAll(substParams, ": "+k, ": "+v)
				substParams = strings.ReplaceAll(substParams, ":"+k, ":"+v)
				substParams = strings.ReplaceAll(substParams, " "+k, " "+v)
				substParams = strings.ReplaceAll(substParams, "("+k, "("+v)
			}
			res := substParams + " => " + substRet
			if hasObj {
				return "object:" + res
			}
			return res
		}
	}
	if strings.Contains(clean, "|") {
		parts := strings.Split(clean, "|")
		var newParts []string
		for _, p := range parts {
			newParts = append(newParts, substituteType(strings.TrimSpace(p), subst))
		}
		res := strings.Join(newParts, " | ")
		if hasObj {
			return "object:" + res
		}
		return res
	}
	if strings.Contains(clean, "<") && strings.HasSuffix(clean, ">") {
		idx := strings.Index(clean, "<")
		name := clean[:idx]
		inner := clean[idx+1 : len(clean)-1]
		parts := splitTypeArguments(inner)
		var newParts []string
		for _, p := range parts {
			newParts = append(newParts, substituteType(p, subst))
		}
		if name == "Array" || name == "ReadonlyArray" {
			if len(newParts) == 1 {
				return newParts[0] + "[]"
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
	if strings.HasPrefix(clean, "[") && strings.HasSuffix(clean, "]") {
		inner := clean[1 : len(clean)-1]
		parts := strings.Split(inner, ",")
		var newParts []string
		for _, p := range parts {
			newParts = append(newParts, substituteType(strings.TrimSpace(p), subst))
		}
		res := "[" + strings.Join(newParts, ", ") + "]"
		if hasObj {
			return "object:" + res
		}
		return res
	}
	if strings.HasPrefix(clean, "__shape_") {
		cleanShape := strings.TrimPrefix(clean, "__shape_")
		tokens := strings.Split(cleanShape, "_")
		var newTokens []string
		for i := 0; i < len(tokens); i += 2 {
			if i+1 < len(tokens) {
				fName := tokens[i]
				fType := tokens[i+1]
				isArr := false
				if strings.HasSuffix(fType, "arr") {
					isArr = true
					fType = strings.TrimSuffix(strings.TrimSuffix(fType, "_arr"), "arr")
				}
				substType := substituteType(fType, subst)
				cleanSubst := strings.ReplaceAll(substType, ":", "_")
				cleanSubst = strings.ReplaceAll(cleanSubst, "[]", "_arr")
				if isArr && !strings.HasSuffix(cleanSubst, "_arr") {
					cleanSubst = cleanSubst + "_arr"
				}
				newTokens = append(newTokens, fName, cleanSubst)
			}
		}
		res := "__shape_" + strings.Join(newTokens, "_")
		if hasObj {
			return "object:" + res
		}
		return res
	}
	if strings.Contains(clean, "__") {
		idx := strings.Index(clean, "__")
		name := clean[:idx]
		inner := clean[idx+2:]
		isArr := false
		if strings.HasSuffix(inner, "_arr") {
			isArr = true
			inner = strings.TrimSuffix(inner, "_arr")
		}
		parts := strings.Split(inner, "_")
		var newParts []string
		for _, p := range parts {
			newParts = append(newParts, substituteType(p, subst))
		}
		mangled := mangleGenericName(name, newParts)
		if isArr {
			mangled = mangled + "[]"
		}
		if hasObj {
			return "object:" + mangled
		}
		return mangled
	}
	if val, ok := subst[clean]; ok {
		if hasObj {
			return "object:" + val
		}
		return val
	}
	return typ
}

func cloneStatement(stmt typescriptgo.SyntaxStatement) typescriptgo.SyntaxStatement {
	res := stmt
	res.Parameters = append([]typescriptgo.SyntaxParameter(nil), stmt.Parameters...)
	res.Body = append([]typescriptgo.SyntaxStatement(nil), stmt.Body...)
	res.Then = append([]typescriptgo.SyntaxStatement(nil), stmt.Then...)
	res.Else = append([]typescriptgo.SyntaxStatement(nil), stmt.Else...)
	res.Cases = make([]typescriptgo.SyntaxSwitchCase, len(stmt.Cases))
	for i, c := range stmt.Cases {
		res.Cases[i] = c
		res.Cases[i].Expression = cloneExpr(c.Expression)
		res.Cases[i].Statements = append([]typescriptgo.SyntaxStatement(nil), c.Statements...)
	}
	res.Catch = append([]typescriptgo.SyntaxStatement(nil), stmt.Catch...)
	res.Finally = append([]typescriptgo.SyntaxStatement(nil), stmt.Finally...)
	if stmt.Class != nil {
		c := cloneClass(*stmt.Class)
		res.Class = &c
	}
	return res
}

func cloneMethod(m typescriptgo.SyntaxMethod) typescriptgo.SyntaxMethod {
	res := m
	res.Parameters = append([]typescriptgo.SyntaxParameter(nil), m.Parameters...)
	res.Body = append([]typescriptgo.SyntaxStatement(nil), m.Body...)
	res.TypeParameters = append([]string(nil), m.TypeParameters...)
	return res
}

func cloneClass(cls typescriptgo.SyntaxClass) typescriptgo.SyntaxClass {
	res := cls
	res.Fields = append([]typescriptgo.SyntaxField(nil), cls.Fields...)
	res.Implements = append([]string(nil), cls.Implements...)
	if cls.Constructor != nil {
		ctor := *cls.Constructor
		ctor.Parameters = append([]typescriptgo.SyntaxParameter(nil), cls.Constructor.Parameters...)
		ctor.Body = append([]typescriptgo.SyntaxStatement(nil), cls.Constructor.Body...)
		res.Constructor = &ctor
	}
	res.Methods = make([]typescriptgo.SyntaxMethod, len(cls.Methods))
	for i, m := range cls.Methods {
		res.Methods[i] = m
		res.Methods[i].Parameters = append([]typescriptgo.SyntaxParameter(nil), m.Parameters...)
		res.Methods[i].Body = append([]typescriptgo.SyntaxStatement(nil), m.Body...)
	}
	res.StaticBlocks = make([][]typescriptgo.SyntaxStatement, len(cls.StaticBlocks))
	for i, b := range cls.StaticBlocks {
		res.StaticBlocks[i] = make([]typescriptgo.SyntaxStatement, len(b))
		for j, s := range b {
			res.StaticBlocks[i][j] = cloneStatement(s)
		}
	}
	res.StaticElements = make([]typescriptgo.SyntaxStaticElement, len(cls.StaticElements))
	for i, elem := range cls.StaticElements {
		res.StaticElements[i] = elem
		if elem.Kind == typescriptgo.StaticElementBlock {
			res.StaticElements[i].Statements = make([]typescriptgo.SyntaxStatement, len(elem.Statements))
			for j, s := range elem.Statements {
				res.StaticElements[i].Statements[j] = cloneStatement(s)
			}
		}
	}
	return res
}

func cloneExpr(expr *typescriptgo.SyntaxExpression) *typescriptgo.SyntaxExpression {
	if expr == nil {
		return nil
	}
	res := *expr
	res.Arguments = append([]*typescriptgo.SyntaxExpression(nil), expr.Arguments...)
	res.TypeArguments = append([]string(nil), expr.TypeArguments...)
	return &res
}

func cloneAndSubstituteStmt(stmt typescriptgo.SyntaxStatement, subst map[string]string) typescriptgo.SyntaxStatement {
	res := cloneStatement(stmt)
	res.Type = substituteType(res.Type, subst)
	res.InferredType = substituteType(res.InferredType, subst)
	for i := range res.Parameters {
		res.Parameters[i].Type = substituteType(res.Parameters[i].Type, subst)
		res.Parameters[i].InferredType = substituteType(res.Parameters[i].InferredType, subst)
		if res.Parameters[i].Initializer != nil {
			res.Parameters[i].Initializer = cloneAndSubstituteExpr(res.Parameters[i].Initializer, subst)
		}
	}
	if res.Expression != nil {
		res.Expression = cloneAndSubstituteExpr(res.Expression, subst)
	}
	if res.Left != nil {
		res.Left = cloneAndSubstituteExpr(res.Left, subst)
	}
	if res.Right != nil {
		res.Right = cloneAndSubstituteExpr(res.Right, subst)
	}
	for i := range res.Body {
		res.Body[i] = cloneAndSubstituteStmt(res.Body[i], subst)
	}
	for i := range res.Then {
		res.Then[i] = cloneAndSubstituteStmt(res.Then[i], subst)
	}
	for i := range res.Else {
		res.Else[i] = cloneAndSubstituteStmt(res.Else[i], subst)
	}
	for i := range res.Cases {
		res.Cases[i].Expression = cloneAndSubstituteExpr(res.Cases[i].Expression, subst)
		for j := range res.Cases[i].Statements {
			res.Cases[i].Statements[j] = cloneAndSubstituteStmt(res.Cases[i].Statements[j], subst)
		}
	}
	for i := range res.Catch {
		res.Catch[i] = cloneAndSubstituteStmt(res.Catch[i], subst)
	}
	for i := range res.Finally {
		res.Finally[i] = cloneAndSubstituteStmt(res.Finally[i], subst)
	}
	return res
}

func cloneAndSubstituteExpr(expr *typescriptgo.SyntaxExpression, subst map[string]string) *typescriptgo.SyntaxExpression {
	if expr == nil {
		return nil
	}
	res := cloneExpr(expr)
	res.InferredType = substituteType(res.InferredType, subst)
	if res.Kind == "as" {
		res.Text = substituteType(res.Text, subst)
	}
	for i := range res.TypeArguments {
		res.TypeArguments[i] = substituteType(res.TypeArguments[i], subst)
	}
	if res.Left != nil {
		res.Left = cloneAndSubstituteExpr(res.Left, subst)
	}
	if res.Right != nil {
		res.Right = cloneAndSubstituteExpr(res.Right, subst)
	}
	for i := range res.Arguments {
		res.Arguments[i] = cloneAndSubstituteExpr(res.Arguments[i], subst)
	}
	if res.WhenTrue != nil {
		res.WhenTrue = cloneAndSubstituteExpr(res.WhenTrue, subst)
	}
	if res.WhenFalse != nil {
		res.WhenFalse = cloneAndSubstituteExpr(res.WhenFalse, subst)
	}
	if res.Function != nil {
		fn := cloneAndSubstituteStmt(*res.Function, subst)
		res.Function = &fn
	}
	return res
}
