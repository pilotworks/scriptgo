package lowering

import (
	"fmt"
	"sort"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

var extraFunctions []ir.Function
var closureCounter int

func findFreeVariables(fn *typescriptgo.SyntaxStatement, outerEnv map[string]ir.Type, selfName string) []string {
	params := map[string]bool{}
	for _, p := range fn.Parameters {
		params[p.Name] = true
	}
	locals := map[string]bool{}
	var used []string

	var collectExpr func(e *typescriptgo.SyntaxExpression)
	var collectStmt func(s typescriptgo.SyntaxStatement)

	collectExpr = func(e *typescriptgo.SyntaxExpression) {
		if e == nil {
			return
		}
		if e.Kind == "identifier" || e.Kind == "this" {
			name := e.Text
			if e.Kind == "this" {
				name = "this"
			}
			if !params[name] && !locals[name] && name != fn.Name && name != selfName {
				if _, ok := outerEnv[name]; ok {
					used = append(used, name)
				}
			}
		}
		if e.Left != nil {
			collectExpr(e.Left)
		}
		if e.Right != nil {
			collectExpr(e.Right)
		}
		for _, arg := range e.Arguments {
			collectExpr(arg)
		}
		if e.WhenTrue != nil {
			collectExpr(e.WhenTrue)
		}
		if e.WhenFalse != nil {
			collectExpr(e.WhenFalse)
		}
		if e.Function != nil {
			collectStmt(*e.Function)
		}
	}

	collectStmt = func(s typescriptgo.SyntaxStatement) {
		if s.Kind == "variable" {
			locals[s.Name] = true
			if s.Expression != nil {
				collectExpr(s.Expression)
			}
		} else if s.Kind == "assign" || s.Kind == "field_set" || s.Kind == "index_set" {
			if !params[s.Name] && !locals[s.Name] {
				if _, ok := outerEnv[s.Name]; ok {
					used = append(used, s.Name)
				}
			}
			if s.Left != nil {
				collectExpr(s.Left)
			}
			if s.Right != nil {
				collectExpr(s.Right)
			}
			if s.Expression != nil {
				collectExpr(s.Expression)
			}
		} else if s.Kind == "return" || s.Kind == "throw" || s.Kind == "expression" {
			if s.Expression != nil {
				collectExpr(s.Expression)
			}
		} else if s.Kind == "if" {
			if s.Expression != nil {
				collectExpr(s.Expression)
			}
			for _, st := range s.Then {
				collectStmt(st)
			}
			for _, st := range s.Else {
				collectStmt(st)
			}
		} else if s.Kind == "while" || s.Kind == "dowhile" {
			if s.Expression != nil {
				collectExpr(s.Expression)
			}
			for _, st := range s.Body {
				collectStmt(st)
			}
		} else if s.Kind == "for" || s.Kind == "forof" || s.Kind == "forawaitof" || s.Kind == "forin" || s.Kind == "for_of" || s.Kind == "for_await_of" || s.Kind == "for_in" {
			if s.Name != "" {
				locals[s.Name] = true
			}
			if s.Expression != nil {
				collectExpr(s.Expression)
			}
			for _, st := range s.Body {
				collectStmt(st)
			}
			for _, st := range s.Step {
				collectStmt(st)
			}
		} else if s.Kind == "try" {
			for _, st := range s.Body {
				collectStmt(st)
			}
			for _, st := range s.Catch {
				collectStmt(st)
			}
			for _, st := range s.Finally {
				collectStmt(st)
			}
		} else if s.Kind == "block" {
			for _, st := range s.Body {
				collectStmt(st)
			}
		} else if s.Kind == "function" || s.Kind == "arrow_function" {
			oldLocals := make(map[string]bool, len(locals))
			for k, v := range locals {
				oldLocals[k] = v
			}
			for _, p := range s.Parameters {
				locals[p.Name] = true
			}
			for _, st := range s.Body {
				collectStmt(st)
			}
			locals = oldLocals
		}
	}

	for _, stmt := range fn.Body {
		collectStmt(stmt)
	}

	// Deduplicate and sort
	seen := map[string]bool{}
	var result []string
	for _, u := range used {
		if !seen[u] {
			seen[u] = true
			result = append(result, u)
		}
	}
	sort.Strings(result)
	return result
}

func lowerClosureExpression(
	path string,
	fnStmt *typescriptgo.SyntaxStatement,
	result string,
	callerFn *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, error) {
	if fnStmt == nil {
		return "", "", fmt.Errorf("closure syntax statement is nil")
	}
	closureCounter++
	closureName := fmt.Sprintf("__closure_%d", closureCounter)

	capturedVars := findFreeVariables(fnStmt, env, result)

	fnType := fnStmt.Type
	if strings.Contains(fnType, "=>") {
		fnType = extractTopLevelReturnType(fnType)
	}
	// Create closure function
	targetFn := ir.Function{
		Name:       closureName,
		Span:       toIRSpan(path, fnStmt.Span),
		ReturnType: toIRType(fnType),
	}
	if isReturningClosure(*fnStmt) {
		targetFn.ReturnType = ir.TypeClosure
	} else if (targetFn.ReturnType == "" || targetFn.ReturnType == ir.TypeVoid) && fnStmt.InferredType != "" {
		inferred := fnStmt.InferredType
		if strings.Contains(inferred, "=>") {
			inferred = extractTopLevelReturnType(inferred)
		}
		targetFn.ReturnType = toIRType(inferred)
	}
	if (fnStmt.IsAsync || fnStmt.Kind == "async_function") && (targetFn.ReturnType == "" || targetFn.ReturnType == ir.TypeVoid) {
		targetFn.ReturnType = ir.Type("object:Promise")
	}
	if targetFn.ReturnType == "" {
		targetFn.ReturnType = ir.TypeVoid
	}

	// Always add __env_ctx as first parameter
	targetFn.Parameters = append(targetFn.Parameters, ir.Parameter{
		Name: "__env_ctx",
		Type: ir.Type("ptr"),
	})

	closureEnv := map[string]ir.Type{
		"__env_ctx": ir.Type("ptr"),
	}

	for pIdx := 0; pIdx < 4; pIdx++ {
		var pName string
		var pType ir.Type
		if pIdx < len(fnStmt.Parameters) {
			p := fnStmt.Parameters[pIdx]
			pName = p.Name
			pType = toIRType(p.Type)
			if (pType == "" || pType == ir.TypeVoid) && p.InferredType != "" {
				pType = toIRType(p.InferredType)
			}
			if pType == "" || pType == ir.TypeVoid {
				pType = ir.TypeUnknown
			}
		} else {
			pName = fmt.Sprintf("__unused_arg_%d", pIdx)
			pType = ir.TypeUnknown
		}
		targetFn.Parameters = append(targetFn.Parameters, ir.Parameter{
			Name: pName + "$raw",
			Type: ir.TypeUnknown,
		})
		closureEnv[pName+"$raw"] = ir.TypeUnknown
		closureEnv[pName] = pType
		closureEnv["__param."+pName] = pType
	}

	for _, capVar := range capturedVars {
		if capType, ok := env[capVar]; ok {
			closureEnv[capVar] = capType
		}
		if retType, ok := env[capVar+".retType"]; ok {
			closureEnv[capVar+".retType"] = retType
		}
	}

	if targetFn.ReturnType == ir.TypeVoid || targetFn.ReturnType == "" {
		if len(fnStmt.Body) == 1 && fnStmt.Body[0].Kind == "return" && fnStmt.Body[0].Expression != nil {
			expr := fnStmt.Body[0].Expression
			if expr.InferredType != "" {
				targetFn.ReturnType = toIRType(expr.InferredType)
			}
			if targetFn.ReturnType == ir.TypeVoid || targetFn.ReturnType == "" {
				switch expr.Kind {
				case "string", "template":
					targetFn.ReturnType = ir.TypeString
				case "bool", "compare":
					targetFn.ReturnType = ir.TypeBool
				case "binary":
					if isComparison(expr.Operator) || expr.Operator == "&&" || expr.Operator == "||" {
						targetFn.ReturnType = ir.TypeBool
					} else if expr.Operator == "+" && ((expr.Left != nil && expr.Left.Kind == "string") || (expr.Right != nil && expr.Right.Kind == "string")) {
						targetFn.ReturnType = ir.TypeString
					} else {
						targetFn.ReturnType = ir.TypeNumber
					}
				case "number":
					targetFn.ReturnType = ir.TypeNumber
				case "call":
					if expr.Left != nil {
						calleeName := callName(expr.Left)
						if sig, ok := signatures[calleeName]; ok {
							targetFn.ReturnType = sig.ReturnType
						} else if retT, ok := env[calleeName+".retType"]; ok {
							targetFn.ReturnType = retT
						} else if retT, ok := closureEnv[calleeName+".retType"]; ok {
							targetFn.ReturnType = retT
						}
					}
				default:
					targetFn.ReturnType = ir.TypeVoid
				}
			}
		}
	}

	// In closure body, captured variables are available from closureEnv
	for _, capVar := range capturedVars {
		closureEnv[capVar] = env[capVar]
		targetFn.Captured = append(targetFn.Captured, ir.Parameter{
			Name: capVar,
			Type: env[capVar],
		})
	}
	if fnStmt.Name != "" {
		closureEnv[fnStmt.Name] = ir.TypeClosure
		closureEnv[fnStmt.Name+".retType"] = targetFn.ReturnType
		signatures[fnStmt.Name] = targetFn
	}

	closureBodyCounter := 0
	if fnStmt.IsGenerator || fnStmt.Kind == "generator_function" || fnStmt.Kind == "async_generator_function" {
		yieldType := ir.TypeNumber
		if strings.Contains(fnStmt.Type, "<") && strings.HasSuffix(fnStmt.Type, ">") {
			idx := strings.Index(fnStmt.Type, "<")
			inner := fnStmt.Type[idx+1 : len(fnStmt.Type)-1]
			parts := splitTypeArguments(inner)
			if len(parts) > 0 {
				yieldType = toIRType(parts[0])
			}
		} else if strings.Contains(fnStmt.InferredType, "<") && strings.HasSuffix(fnStmt.InferredType, ">") {
			idx := strings.Index(fnStmt.InferredType, "<")
			inner := fnStmt.InferredType[idx+1 : len(fnStmt.InferredType)-1]
			parts := splitTypeArguments(inner)
			if len(parts) > 0 {
				yieldType = toIRType(parts[0])
			}
		}

		genClassName := "Generator_" + closureName
		targetFn.ReturnType = ir.Type("object:" + genClassName)

		if _, exists := shapes[genClassName]; !exists {
			shapes[genClassName] = ir.ObjectShape{
				Name: genClassName,
				Span: toIRSpan(path, fnStmt.Span),
				Fields: []ir.Field{
					{Name: "__state", Type: ir.TypeNumber, Value: "0", Span: toIRSpan(path, fnStmt.Span)},
					{Name: "__done", Type: ir.TypeBool, Value: "false", Span: toIRSpan(path, fnStmt.Span)},
					{Name: "__value", Type: yieldType, Span: toIRSpan(path, fnStmt.Span)},
					{Name: "__items", Type: ir.Type(string(yieldType) + "[]"), Span: toIRSpan(path, fnStmt.Span)},
				},
			}
		}

		itemsTemp := nextTemp(&closureBodyCounter)
		targetFn.Body = append(targetFn.Body, ir.Instruction{Op: ir.OpArray, Type: ir.Type(string(yieldType) + "[]"), Result: itemsTemp, Span: targetFn.Span})
		closureEnv["__items"] = ir.Type(string(yieldType) + "[]")
		closureEnv[itemsTemp] = ir.Type(string(yieldType) + "[]")

		rewrittenBody := rewriteYieldsToPush(fnStmt.Body, itemsTemp)
		for _, bodyStatement := range rewrittenBody {
			_ = lowerStatement(path, bodyStatement, &targetFn, closureEnv, &closureBodyCounter, shapes, signatures)
		}

		stateZero := nextTemp(&closureBodyCounter)
		targetFn.Body = append(targetFn.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: stateZero, Value: "0", Span: targetFn.Span})
		doneFalse := nextTemp(&closureBodyCounter)
		targetFn.Body = append(targetFn.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: doneFalse, Value: "false", Span: targetFn.Span})
		defVal := nextTemp(&closureBodyCounter)
		targetFn.Body = append(targetFn.Body, ir.Instruction{Op: ir.OpConst, Type: yieldType, Result: defVal, Value: "0", Span: targetFn.Span})

		genObj := nextTemp(&closureBodyCounter)
		targetFn.Body = append(targetFn.Body, ir.Instruction{
			Op:         ir.OpObjectNew,
			Type:       ir.Type("object:" + genClassName),
			Result:     genObj,
			Callee:     genClassName,
			FieldCount: 4,
			Args:       []string{stateZero, doneFalse, defVal, itemsTemp},
			Span:       targetFn.Span,
		})
		for fIdx, fName := range []string{"__state", "__done", "__value", "__items"} {
			targetFn.Body = append(targetFn.Body, ir.Instruction{
				Op:         ir.OpFieldSet,
				Type:       ir.TypeVoid,
				Callee:     genClassName,
				Field:      fName,
				FieldIndex: fIdx,
				Args:       []string{genObj, []string{stateZero, doneFalse, defVal, itemsTemp}[fIdx]},
				Span:       targetFn.Span,
			})
		}
		targetFn.Body = append(targetFn.Body, ir.Instruction{
			Op:   ir.OpReturn,
			Type: ir.Type("object:" + genClassName),
			Args: []string{genObj},
			Span: targetFn.Span,
		})
	} else {
		if fnStmt.Name != "" {
			closureEnv[fnStmt.Name] = ir.TypeClosure
			closureEnv[fnStmt.Name+".retType"] = targetFn.ReturnType
			closureEnv[fnStmt.Name+".closureTarget"] = ir.Type(closureName)
		}
		if result != "" {
			closureEnv[result] = ir.TypeClosure
			closureEnv[result+".retType"] = targetFn.ReturnType
			closureEnv[result+".closureTarget"] = ir.Type(closureName)
		}
		for _, p := range fnStmt.Parameters {
			typ := closureEnv[p.Name]
			targetFn.Body = append(targetFn.Body, ir.Instruction{
				Op:     ir.OpCheckedCast,
				Type:   typ,
				Result: p.Name,
				Args:   []string{p.Name + "$raw"},
				Span:   targetFn.Span,
			})
		}
		returned := false
		for _, bodyStatement := range fnStmt.Body {
			if err := lowerStatement(path, bodyStatement, &targetFn, closureEnv, &closureBodyCounter, shapes, signatures); err != nil {
				return "", "", sourceError(path, bodyStatement.Span, err)
			}
			if statementAlwaysReturns(bodyStatement) {
				returned = true
			}
		}
		if !returned {
			if strings.HasPrefix(string(targetFn.ReturnType), "object:Promise") {
				prom := appendResolvedPromiseUndefined(&targetFn, &closureBodyCounter, targetFn.Span)
				targetFn.Body = append(targetFn.Body, ir.Instruction{Op: ir.OpReturn, Type: targetFn.ReturnType, Args: []string{prom}, Span: targetFn.Span})
			} else if targetFn.ReturnType != ir.TypeVoid {
				targetFn.Body = append(targetFn.Body, ir.Instruction{Op: ir.OpReturn, Type: targetFn.ReturnType, Span: targetFn.Span})
			} else {
				targetFn.Body = append(targetFn.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: targetFn.Span})
			}
		}
	}

	extraFunctions = append(extraFunctions, targetFn)
	signatures[closureName] = targetFn

	if result == "" {
		result = nextTemp(counter)
	}
	signatures[result] = targetFn
	env[result+".retType"] = targetFn.ReturnType

	callerFn.Body = append(callerFn.Body, ir.Instruction{
		Op:     ir.OpClosure,
		Type:   ir.TypeClosure,
		Result: result,
		Callee: closureName,
		Args:   capturedVars,
		Span:   toIRSpan(path, fnStmt.Span),
	})

	return result, ir.TypeClosure, nil
}

func ensureFunctionClosureTrampoline(path string, sig ir.Function, signatures map[string]ir.Function) string {
	if strings.HasPrefix(sig.Name, "__closure_") {
		return sig.Name
	}
	trampolineName := "__closure_trampoline_" + sig.Name
	if _, exists := signatures[trampolineName]; exists {
		return trampolineName
	}

	trampolineFn := ir.Function{
		Name:       trampolineName,
		ReturnType: sig.ReturnType,
		Span:       sig.Span,
	}

	// Closure ABI parameters: (__env_ctx: ptr, param0$raw: unknown, param1$raw: unknown, param2$raw: unknown, param3$raw: unknown)
	trampolineFn.Parameters = append(trampolineFn.Parameters, ir.Parameter{
		Name: "__env_ctx",
		Type: ir.TypePointer,
	})

	var callArgs []string
	counter := 0
	for i, param := range sig.Parameters {
		rawName := fmt.Sprintf("arg_%d$raw", i)
		trampolineFn.Parameters = append(trampolineFn.Parameters, ir.Parameter{
			Name: rawName,
			Type: ir.TypeUnknown,
		})
		unboxedName := fmt.Sprintf("arg_%d", i)
		trampolineFn.Body = append(trampolineFn.Body, ir.Instruction{
			Op:     ir.OpCheckedCast,
			Type:   param.Type,
			Result: unboxedName,
			Args:   []string{rawName},
			Span:   sig.Span,
		})
		callArgs = append(callArgs, unboxedName)
		counter++
	}

	// Fill remaining up to 4 raw args if sig has fewer than 4 parameters
	for i := len(sig.Parameters); i < 4; i++ {
		rawName := fmt.Sprintf("__unused_arg_%d$raw", i)
		trampolineFn.Parameters = append(trampolineFn.Parameters, ir.Parameter{
			Name: rawName,
			Type: ir.TypeUnknown,
		})
	}

	if sig.ReturnType != ir.TypeVoid {
		retVal := fmt.Sprintf("ret_%d", counter)
		trampolineFn.Body = append(trampolineFn.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   sig.ReturnType,
			Result: retVal,
			Callee: sig.Name,
			Args:   callArgs,
			Span:   sig.Span,
		})
		trampolineFn.Body = append(trampolineFn.Body, ir.Instruction{
			Op:   ir.OpReturn,
			Type: sig.ReturnType,
			Args: []string{retVal},
			Span: sig.Span,
		})
	} else {
		trampolineFn.Body = append(trampolineFn.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeVoid,
			Callee: sig.Name,
			Args:   callArgs,
			Span:   sig.Span,
		})
		trampolineFn.Body = append(trampolineFn.Body, ir.Instruction{
			Op:   ir.OpReturn,
			Type: ir.TypeVoid,
			Span: sig.Span,
		})
	}

	extraFunctions = append(extraFunctions, trampolineFn)
	signatures[trampolineName] = trampolineFn
	return trampolineName
}
