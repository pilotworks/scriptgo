package lowering

import (
	"fmt"
	"maps"
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerFunction(path string, statement typescriptgo.SyntaxStatement, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (ir.Function, error) {
	savedUsingScopes := usingScopeStack
	usingScopeStack = nil
	defer func() {
		usingScopeStack = savedUsingScopes
	}()

	retType := statement.Type
	if retType == "" && statement.InferredType != "" {
		retType = statement.InferredType
	}
	function := ir.Function{Name: statement.Name, Span: toIRSpan(path, statement.Span), ReturnType: toIRType(retType)}
	if function.ReturnType == "" {
		function.ReturnType = ir.TypeVoid
	}
	env := map[string]ir.Type{}
	if resolvedType, ok := asyncResolvedReturnType(retType); ok {
		env[asyncResolvedReturnTypeEnvKey] = resolvedType
	}
	for _, parameter := range statement.Parameters {
		pType := parameter.Type
		if pType == "" && parameter.InferredType != "" {
			pType = parameter.InferredType
		}
		typ := toIRType(pType)
		if parameter.Rest {
			if typ == "" || typ == ir.TypeUnknown {
				if pType == "number[]" {
					typ = ir.TypeNumberArray
				} else {
					typ = ir.TypeStringArray
				}
			}
		}
		if typ == "" {
			return ir.Function{}, fmt.Errorf("parameter %q has unsupported type %q", parameter.Name, parameter.Type)
		}
		function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: typ})
		if typ == ir.TypeObject && pType != "" && pType != "object" {
			env[parameter.Name] = ir.Type("object:" + pType)
		} else {
			env[parameter.Name] = typ
		}
		// Keep the source-level declaration alongside the storage type. Mixed
		// unions may use unknown storage until a control-flow guard narrows them.
		env["__decl_str."+parameter.Name] = ir.Type(pType)
		env["__param."+parameter.Name] = typ
		env["__storage_type."+parameter.Name] = typ
		fnSig := parameter.Type
		if fnSig == "" || fnSig == "closure" {
			fnSig = parameter.InferredType
		}
		if strings.Contains(fnSig, "=>") {
			retStr := extractTopLevelReturnType(fnSig)
			env[parameter.Name+".retType"] = toIRType(retStr)
		}
	}
	counter := 0
	returned := false
	for _, bodyStatement := range statement.Body {
		if err := lowerStatement(path, bodyStatement, &function, env, &counter, shapes, signatures); err != nil {
			return ir.Function{}, sourceError(path, bodyStatement.Span, err)
		}
		if statementAlwaysReturns(bodyStatement) {
			returned = true
		}
	}
	if !returned {
		if function.ReturnType == ir.TypeVoid {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: function.Span})
		} else if strings.HasPrefix(string(function.ReturnType), "object:Promise") {
			prom := appendResolvedPromiseUndefined(&function, &counter, function.Span)
			function.Body = append(function.Body, ir.Instruction{
				Op:   ir.OpReturn,
				Type: function.ReturnType,
				Args: []string{prom},
				Span: function.Span,
			})
		} else {
			defVal := ""
			if function.ReturnType == ir.TypeNumber {
				defVal = "0"
			} else if function.ReturnType == ir.TypeBool {
				defVal = "false"
			} else if strings.HasPrefix(string(function.ReturnType), "object:") || function.ReturnType == "ptr" {
				defVal = "0"
			}
			defTemp := nextTemp(&counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   function.ReturnType,
				Result: defTemp,
				Value:  defVal,
				Span:   function.Span,
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:   ir.OpReturn,
				Type: function.ReturnType,
				Args: []string{defTemp},
				Span: function.Span,
			})
		}
	}
	return function, nil
}

func lowerStatement(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	switch statement.Kind {
	case "empty":
		return nil
	case "variable", "using", "await_using":
		varResultName := statement.Name
		if _, isShadowed := env[statement.Name]; isShadowed {
			varResultName = fmt.Sprintf("%s$%d", statement.Name, *counter)
			*counter++
			env["__ident."+statement.Name] = ir.Type(varResultName)
		}
		localType := toIRType(statement.Type)
		if localType == "" {
			localType = toIRType(statement.InferredType)
		}
		if localType == "" {
			localType = ir.TypeUnknown
		}
		function.Locals = append(function.Locals, ir.Parameter{Name: varResultName, Type: localType})
		if statement.Expression == nil {
			if statement.Kind == "using" || statement.Kind == "await_using" {
				return fmt.Errorf("resource %q has no initializer", statement.Name)
			}
			typeStr := statement.Type
			if typeStr == "" {
				typeStr = statement.InferredType
			}
			declaredType := toIRType(typeStr)
			if declaredType == "" {
				declaredType = ir.TypeUnknown
			}
			env[varResultName] = declaredType
			env[statement.Name] = declaredType
			env["__storage_type."+varResultName] = declaredType
			env["__decl_str."+varResultName] = ir.Type(typeStr)
			var zeroVal string
			hasUndef := strings.Contains(typeStr, "undefined") || strings.Contains(typeStr, "void")
			switch declaredType {
			case ir.TypeNumber:
				zeroVal = "0"
			case ir.TypeBool:
				zeroVal = "false"
			case ir.TypeBigInt:
				zeroVal = "0"
			case ir.TypeString:
				if hasUndef {
					zeroVal = "undefined"
				} else {
					zeroVal = ""
				}
			case ir.TypeUnknown:
				zeroVal = "undefined"
			default:
				if hasUndef {
					zeroVal = "undefined"
				} else {
					zeroVal = "null"
				}
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   declaredType,
				Result: varResultName,
				Value:  zeroVal,
				Span:   toIRSpan(path, statement.Span),
			})
			return nil
		}
		inProgressVars[statement.Name] = true
		defer delete(inProgressVars, statement.Name)
		declaredType := toIRType(statement.Type)
		if statement.Type == "" && statement.InferredType != "" {
			declaredType = toIRType(statement.InferredType)
		}
		if statement.Expression.Kind == "identifier" || statement.Expression.Kind == "this" {
			identText := statement.Expression.Text
			if statement.Expression.Kind == "this" {
				identText = "this"
			}
			if mangled, hasMangled := env["__ident."+identText]; hasMangled {
				identText = string(mangled)
			}
			srcType, ok := env[identText]
			if ok && (declaredType == "" || declaredType == srcType || declaredType == ir.TypeUnknown) {
				if declaredType == ir.TypeUnknown {
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpBoxUnknown,
						Type:   ir.TypeUnknown,
						Result: varResultName,
						Args:   []string{identText},
						Span:   toIRSpan(path, statement.Span),
					})
					env[varResultName] = ir.TypeUnknown
					env[statement.Name] = ir.TypeUnknown
					return nil
				}
				env[varResultName] = srcType
				env[statement.Name] = srcType
				switch srcType {
				case ir.TypeNumber:
					zeroConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, statement.Span)})
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: varResultName, Operator: "+", Args: []string{identText, zeroConst}, Span: toIRSpan(path, statement.Span)})
				case ir.TypeString:
					emptyStr := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: emptyStr, Value: "", Span: toIRSpan(path, statement.Span)})
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: varResultName, Operator: "+", Args: []string{identText, emptyStr}, Span: toIRSpan(path, statement.Span)})
				case ir.TypeBool:
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: varResultName, Operator: "||", Args: []string{identText, identText}, Span: toIRSpan(path, statement.Span)})
				default:
					trueConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: trueConst, Value: "true", Span: toIRSpan(path, statement.Span)})
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: srcType, Result: varResultName, Args: []string{trueConst, identText, identText}, Span: toIRSpan(path, statement.Span)})
				}
				return nil
			} else if !ok {
				global, isGlobal := builtinGlobal(statement.Expression.Text)
				if isGlobal {
					env[varResultName] = global.Type
					env[statement.Name] = global.Type
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: global.Type, Result: varResultName, Value: global.Value, Span: toIRSpan(path, statement.Span)})
					return nil
				}
				return fmt.Errorf("unknown identifier %q", statement.Expression.Text)
			}
		}
		if declaredType == ir.TypeUnknown {
			if statement.Expression != nil && statement.Expression.Kind == "null" {
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeUnknown,
					Result: varResultName,
					Value:  "null",
					Span:   toIRSpan(path, statement.Span),
				})
				env[varResultName] = ir.TypeUnknown
				env[statement.Name] = ir.TypeUnknown
				return nil
			}
			if statement.Expression != nil && statement.Expression.Kind == "undefined" {
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeUnknown,
					Result: varResultName,
					Value:  "undefined",
					Span:   toIRSpan(path, statement.Span),
				})
				env[varResultName] = ir.TypeUnknown
				env[statement.Name] = ir.TypeUnknown
				return nil
			}
			value, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
			if err != nil {
				return err
			}
			if valType == ir.TypeUnknown {
				if value != varResultName {
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   ir.TypeUnknown,
						Result: varResultName,
						Value:  "undefined",
						Span:   toIRSpan(path, statement.Span),
					})
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpAssign,
						Type:   ir.TypeUnknown,
						Result: varResultName,
						Args:   []string{value},
						Span:   toIRSpan(path, statement.Span),
					})
				}
			} else {
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpBoxUnknown,
					Type:   ir.TypeUnknown,
					Result: varResultName,
					Args:   []string{value},
					Span:   toIRSpan(path, statement.Span),
				})
			}
			env[varResultName] = ir.TypeUnknown
			env[statement.Name] = ir.TypeUnknown
			return nil
		}
		if (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && declaredType != "" && declaredType != ir.TypeVoid && declaredType != ir.TypeUnknown {
			defaultVal := "0"
			if declaredType == ir.TypeNumber {
				defaultVal = "NaN"
			} else if declaredType == ir.TypeBool {
				defaultVal = "false"
			} else if declaredType == ir.TypeString {
				if statement.Expression.Kind == "undefined" {
					defaultVal = "undefined"
				} else {
					defaultVal = "null"
				}
			} else if strings.HasPrefix(string(declaredType), "object:") || declaredType == ir.TypePointer {
				if statement.Expression.Kind == "undefined" {
					defaultVal = "undefined"
				} else {
					defaultVal = "null"
				}
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   declaredType,
				Result: varResultName,
				Value:  defaultVal,
				Span:   toIRSpan(path, statement.Span),
			})
			env[varResultName] = declaredType
			env[statement.Name] = declaredType
			return nil
		}
		if statement.Expression != nil && (statement.Expression.Kind == "array" || (statement.Expression.Kind == "new" && callName(statement.Expression.Left) == "Array")) && statement.Type != "" {
			if toIRType(statement.Type) == ir.TypeUnknownArray {
				statement.Expression.InferredType = "unknown[]"
			} else if _, isTup := tupleFields(statement.Type); isTup {
				statement.Expression.InferredType = statement.Type
			} else if strings.HasSuffix(statement.Type, "[]") || statement.Expression.InferredType == "" || statement.Expression.InferredType == "any[]" || statement.Expression.InferredType == "never[]" || statement.Expression.InferredType == "unknown[]" {
				statement.Expression.InferredType = statement.Type
			}
		}
		if statement.Expression != nil && statement.Expression.Kind == "object_literal" && statement.Type != "" {
			statement.Expression.InferredType = statement.Type
		}
		if statement.Expression != nil && (statement.Expression.Kind == "function" || statement.Expression.Function != nil || strings.Contains(statement.Type, "=>") || strings.Contains(statement.InferredType, "=>")) {
			env[varResultName] = ir.TypeClosure
			env[statement.Name] = ir.TypeClosure
			fnSig := statement.Type
			if fnSig == "" || fnSig == "closure" {
				fnSig = statement.InferredType
			}
			if strings.Contains(fnSig, "=>") {
				retStr := extractTopLevelReturnType(fnSig)
				env[varResultName+".retType"] = toIRType(retStr)
				env[statement.Name+".retType"] = toIRType(retStr)
			}
		}
		if statement.Type != "" {
			env["__decl_str."+varResultName] = ir.Type(statement.Type)
			env["__decl_str."+statement.Name] = ir.Type(statement.Type)
		}
		// Intrinsics may honor the requested declaration name and emit an
		// assignment while lowering (for example JSON.stringify). Predeclare a
		// known inferred slot so that assignment validation sees the declaration.
		if declaredType == "" {
			// The checker may omit an expression's return type for a builtin
			// call. Use boxed storage during lowering; the actual value type is
			// installed immediately after the expression is lowered.
			declaredType = ir.TypeUnknown
		}
		env[varResultName] = declaredType
		env[statement.Name] = declaredType
		env["__storage_type."+varResultName] = declaredType
		value, valType, err := lowerExpression(path, statement.Expression, varResultName, function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if declaredType != "" && valType == ir.TypeUnknown && declaredType != ir.TypeUnknown && !isOptionalChainExpr(statement.Expression) {
			if value == varResultName && statement.Expression.Kind != "conditional" {
				tempVal := nextTemp(counter)
				for i := len(function.Body) - 1; i >= 0; i-- {
					if function.Body[i].Result == varResultName {
						function.Body[i].Result = tempVal
						break
					}
				}
				value = tempVal
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCheckedCast,
				Type:   declaredType,
				Result: varResultName,
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
			valType = declaredType
		}
		typ := valType
		if declaredType != "" && declaredType != ir.TypeUnknown {
			if !(strings.HasPrefix(string(valType), "object:") && strings.HasPrefix(string(declaredType), "object:") && !strings.Contains(string(declaredType), "shape_") && !strings.Contains(string(declaredType), "{")) {
				typ = declaredType
			}
		}
		env[varResultName] = typ
		env[statement.Name] = typ
		if typ == ir.TypeClosure {
			fnSig := statement.Type
			if fnSig == "" || fnSig == "closure" {
				fnSig = statement.InferredType
			}
			if strings.Contains(fnSig, "=>") {
				retStr := extractTopLevelReturnType(fnSig)
				env[varResultName+".retType"] = toIRType(retStr)
				env[statement.Name+".retType"] = toIRType(retStr)
			}
		}
		if statement.Kind == "using" {
			recordUsingResource(varResultName, typ, false, statement.Span)
		} else if statement.Kind == "await_using" {
			recordUsingResource(varResultName, typ, true, statement.Span)
		}
		return nil
	case "expression":
		if statement.Expression == nil {
			return fmt.Errorf("empty expression")
		}
		_, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		return err
	case "return":
		if statement.Expression == nil {
			span := toIRSpan(path, statement.Span)
			if strings.HasPrefix(string(function.ReturnType), "object:Promise") {
				// A bare return in an async function fulfills its promise with
				// undefined; it does not return a void LLVM value directly.
				promiseVal := appendResolvedPromiseUndefined(function, counter, span)
				emitAllActiveUsingScopes(path, function, counter, shapes, signatures)
				bodyLenBeforeFinally := len(function.Body)
				if err := lowerActiveReturnFinally(path, function, env, counter, shapes, signatures); err != nil {
					return err
				}
				if len(function.Body) == bodyLenBeforeFinally || function.Body[len(function.Body)-1].Op != ir.OpReturn {
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: function.ReturnType, Args: []string{promiseVal}, Span: span})
				}
				return nil
			}
			emitAllActiveUsingScopes(path, function, counter, shapes, signatures)
			bodyLenBeforeFinally := len(function.Body)
			if err := lowerActiveReturnFinally(path, function, env, counter, shapes, signatures); err != nil {
				return err
			}
			if len(function.Body) == bodyLenBeforeFinally || function.Body[len(function.Body)-1].Op != ir.OpReturn {
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: span})
			}
			return nil
		}
		if (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && function.ReturnType != "" && function.ReturnType != ir.TypeVoid && function.ReturnType != ir.TypeUnknown {
			res := nextTemp(counter)
			if strings.HasPrefix(string(function.ReturnType), "object:Promise") {
				span := toIRSpan(path, statement.Span)
				prom := ""
				if statement.Expression.Kind == "null" {
					prom = appendResolvedPromiseNull(function, counter, span)
				} else {
					prom = appendResolvedPromiseUndefined(function, counter, span)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: function.ReturnType, Args: []string{prom}, Span: span})
				return nil
			}
			defaultVal := "0"
			if function.ReturnType == ir.TypeBool {
				defaultVal = "false"
			} else if function.ReturnType == ir.TypeNumber {
				defaultVal = "0"
			} else if statement.Expression.Kind == "undefined" {
				defaultVal = "undefined"
			} else if statement.Expression.Kind == "null" || strings.HasPrefix(string(function.ReturnType), "object:") || isPointerLikeType(function.ReturnType) || function.ReturnType == ir.TypeString {
				defaultVal = "null"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   function.ReturnType,
				Result: res,
				Value:  defaultVal,
				Span:   toIRSpan(path, statement.Span),
			})
			bodyLenBeforeFinally := len(function.Body)
			if err := lowerActiveReturnFinally(path, function, env, counter, shapes, signatures); err != nil {
				return err
			}
			if len(function.Body) > bodyLenBeforeFinally && function.Body[len(function.Body)-1].Op == ir.OpReturn {
				return nil
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: function.ReturnType, Args: []string{res}, Span: toIRSpan(path, statement.Span)})
			return nil
		}
		returnExpression := statement.Expression
		expectedReturnType := function.ReturnType
		if resolvedType := env[asyncResolvedReturnTypeEnvKey]; resolvedType != "" {
			expectedReturnType = resolvedType
		}
		if returnExpression.Kind == "object_literal" && strings.HasPrefix(string(expectedReturnType), "object:") && !strings.Contains(string(expectedReturnType), "|") {
			cleanRet := strings.TrimPrefix(string(expectedReturnType), "object:")
			if aliased, ok := typeAliasesIndex[cleanRet]; !ok || !strings.Contains(aliased, "|") {
				// Contextual typing is local to this lowering pass. The frontend AST
				// can be reused by specializations and repeated Lower calls.
				returnExpression = cloneAndSubstituteExpr(returnExpression, nil)
				returnExpression.InferredType = string(expectedReturnType)
			}
		} else if returnExpression.Kind == "array" && strings.HasPrefix(string(expectedReturnType), "object:") {
			// A tuple return type may contain nullable or otherwise heterogeneous
			// elements. Contextualize the array so both branches use one slot layout.
			shapeName := strings.TrimPrefix(string(expectedReturnType), "object:")
			if isTupleShapeName(shapeName) {
				returnExpression = cloneAndSubstituteExpr(returnExpression, nil)
				returnExpression.InferredType = string(expectedReturnType)
			}
		}
		value, typ, err := lowerExpression(path, returnExpression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if strings.HasPrefix(string(function.ReturnType), "object:Promise") && !strings.HasPrefix(string(typ), "object:Promise") {
			prom := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.Type("object:Promise"),
				Result: prom,
				Callee: "__async.promise_resolve",
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
			value = prom
			typ = function.ReturnType
		}
		if function.ReturnType == ir.TypeUnknown && typ != ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: boxed,
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
			value = boxed
			typ = ir.TypeUnknown
		} else if function.ReturnType != "" && function.ReturnType != ir.TypeVoid && function.ReturnType != ir.TypeUnknown && typ == ir.TypeUnknown {
			castVal := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCheckedCast,
				Type:   function.ReturnType,
				Result: castVal,
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
			value = castVal
			typ = function.ReturnType
		}
		if strings.HasPrefix(string(function.ReturnType), "object:Promise") && !strings.HasPrefix(string(typ), "object:Promise") {
			prom := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   function.ReturnType,
				Result: prom,
				Callee: "__async.promise_resolve",
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
			value = prom
			typ = function.ReturnType
		} else if strings.HasPrefix(string(function.ReturnType), "object:") && !strings.Contains(string(function.ReturnType), "{") && strings.HasPrefix(string(typ), "object:") {
			typ = function.ReturnType
		} else if function.ReturnType == "ptr" && (typ == ir.TypeString || isPointerLikeType(typ)) {
			typ = "ptr"
		} else if function.ReturnType == ir.TypeString && typ == "ptr" {
			typ = ir.TypeString
		} else if function.ReturnType != "" && function.ReturnType != ir.TypeVoid && isPointerLikeType(function.ReturnType) && (typ == "ptr" || isPointerLikeType(typ)) {
			typ = function.ReturnType
		} else if function.ReturnType == "" || function.ReturnType == ir.TypeVoid {
			function.ReturnType = typ
			if signatures != nil {
				if sig, ok := signatures[function.Name]; ok {
					sig.ReturnType = typ
					signatures[function.Name] = sig
				}
			}
		}
		if len(activeReturnFinallyStack) > 0 && value != "" && typ != ir.TypeVoid {
			savedVal := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   typ,
				Result: savedVal,
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
			value = savedVal
		}
		emitAllActiveUsingScopes(path, function, counter, shapes, signatures)
		bodyLenBeforeFinally := len(function.Body)
		if err := lowerActiveReturnFinally(path, function, env, counter, shapes, signatures); err != nil {
			return err
		}
		if len(function.Body) > bodyLenBeforeFinally && function.Body[len(function.Body)-1].Op == ir.OpReturn {
			return nil
		}
		if typ == ir.TypeVoid || value == "" {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: toIRSpan(path, statement.Span)})
		} else {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: typ, Args: []string{value}, Span: toIRSpan(path, statement.Span)})
		}
	case "block":
		blockEnv := maps.Clone(env)
		pushUsingScope()
		for _, s := range statement.Body {
			if err := lowerStatement(path, s, function, blockEnv, counter, shapes, signatures); err != nil {
				return err
			}
		}
		popAndEmitUsingScope(path, function, counter, shapes, signatures)
	case "namespace":
		for _, s := range statement.Body {
			if s.Kind == "variable" || s.Kind == "using" || s.Kind == "await_using" {
				varCopy := s
				varCopy.Name = statement.Name + "." + s.Name
				if err := lowerStatement(path, varCopy, function, env, counter, shapes, signatures); err != nil {
					return err
				}
			} else {
				if err := lowerStatement(path, s, function, env, counter, shapes, signatures); err != nil {
					return err
				}
			}
		}
	case "assign":
		targetVarName := statement.Name
		if mangled, hasMangled := env["__ident."+statement.Name]; hasMangled {
			targetVarName = string(mangled)
		}
		varType, ok := env[targetVarName]
		if !ok {
			if topVar, isTop := topLevelVars[statement.Name]; isTop {
				vType := topVar.Type
				if vType == "" && topVar.InferredType != "" {
					vType = topVar.InferredType
				}
				varType = toIRType(vType)
				if varType == "" {
					varType = ir.TypeNumber
				}
				env[targetVarName] = varType
			} else {
				return fmt.Errorf("assignment to unknown variable %q", statement.Name)
			}
		}
		if (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && varType != ir.TypeUnknown {
			defaultVal := "0"
			if statement.Expression.Kind == "undefined" {
				if isPointerLikeType(varType) || strings.HasPrefix(string(varType), "object:") || varType == ir.TypeString {
					defaultVal = "undefined"
				} else if varType == ir.TypeNumber {
					defaultVal = "NaN"
				} else if varType == ir.TypeBool {
					defaultVal = "false"
				}
			} else {
				if isPointerLikeType(varType) || strings.HasPrefix(string(varType), "object:") || varType == ir.TypeString {
					defaultVal = "null"
				} else if varType == ir.TypeNumber {
					defaultVal = "NaN"
				} else if varType == ir.TypeBool {
					defaultVal = "false"
				}
			}
			tmp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   varType,
				Result: tmp,
				Value:  defaultVal,
				Span:   toIRSpan(path, statement.Span),
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   varType,
				Result: targetVarName,
				Args:   []string{tmp},
				Span:   toIRSpan(path, statement.Span),
			})
			return nil
		}
		if statement.Expression != nil && (statement.Expression.Kind == "array" || (statement.Expression.Kind == "new" && callName(statement.Expression.Left) == "Array")) {
			targetType := string(varType)
			if strings.HasSuffix(targetType, "[]") || statement.Expression.InferredType == "" || statement.Expression.InferredType == "any[]" || statement.Expression.InferredType == "never[]" || statement.Expression.InferredType == "unknown[]" {
				statement.Expression.InferredType = targetType
			}
		}
		value, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		storageType := env["__storage_type."+targetVarName]
		if storageType == "" {
			storageType = env["__storage_type."+statement.Name]
		}
		if storageType == "" {
			if topVar, isTop := topLevelVars[statement.Name]; isTop {
				vType := topVar.Type
				if vType == "" && topVar.InferredType != "" {
					vType = topVar.InferredType
				}
				storageType = toIRType(vType)
			}
		}
		if (varType == ir.TypeUnknown || storageType == ir.TypeUnknown) && valType != ir.TypeUnknown {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: targetVarName,
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
			return nil
		}
		if valType != varType {
			if (strings.HasPrefix(string(valType), "object:") || valType == ir.TypeObject) && (strings.HasPrefix(string(varType), "object:") || varType == ir.TypeObject) {
				// Polymorphic object assignment
			} else if valType == ir.TypeUnknown {
				unboxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCheckedCast,
					Type:   varType,
					Result: unboxed,
					Args:   []string{value},
					Span:   toIRSpan(path, statement.Span),
				})
				value = unboxed
			} else if isPointerLikeType(varType) && (valType == ir.TypeVoid || valType == ir.TypePointer) {
				// Assigning undefined or null to a pointer-like variable
			} else if strings.HasPrefix(statement.Name, "__destruct_") || strings.HasPrefix(targetVarName, "__destruct_") {
				// Destructuring temporary assignment under out-of-bounds guard
			} else {
				return fmt.Errorf("assignment type mismatch for %q: %s := %s", statement.Name, varType, valType)
			}
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpAssign, Type: varType, Result: targetVarName, Args: []string{value}, Span: toIRSpan(path, statement.Span)})
	case "while":
		return lowerWhile(path, statement, function, env, counter, shapes, signatures)
	case "if":
		return lowerIf(path, statement, function, env, counter, shapes, signatures)
	case "break":
		if err := lowerActiveBreakFinally(path, function, env, counter, shapes, signatures); err != nil {
			return err
		}
		if len(function.Body) > 0 && function.Body[len(function.Body)-1].Op == ir.OpReturn {
			return nil
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBreak, Type: ir.TypeVoid, Value: statement.Name, Span: toIRSpan(path, statement.Span)})
	case "continue":
		if err := lowerActiveBreakFinally(path, function, env, counter, shapes, signatures); err != nil {
			return err
		}
		if len(function.Body) > 0 && function.Body[len(function.Body)-1].Op == ir.OpReturn {
			return nil
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpContinue, Type: ir.TypeVoid, Value: statement.Name, Span: toIRSpan(path, statement.Span)})
	case "dowhile":
		return lowerDoWhile(path, statement, function, env, counter, shapes, signatures)
	case "forof", "forawaitof":
		return lowerForOf(path, statement, function, env, counter, shapes, signatures)
	case "forin":
		return lowerForIn(path, statement, function, env, counter, shapes, signatures)
	case "label":
		return lowerLabel(path, statement, function, env, counter, shapes, signatures)
	case "switch":
		return lowerSwitch(path, statement, function, env, counter, shapes, signatures)
	case "function":
		_, typ, err := lowerClosureExpression(path, &statement, statement.Name, function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		env[statement.Name] = typ
		return nil
	case "index_set":
		arrVal, arrType, err := lowerExpression(path, statement.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if after, ok := strings.CutPrefix(string(arrType), "object:"); ok {
			shapeName := after
			if shape, ok := shapes[shapeName]; ok {
				if statement.Right != nil && statement.Right.Kind == "number" {
					fieldIdx, _ := strconv.Atoi(statement.Right.Text)
					if fieldIdx >= 0 && fieldIdx < len(shape.Fields) {
						field := shape.Fields[fieldIdx]
						val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
						if err != nil {
							return err
						}
						function.Body = append(function.Body, ir.Instruction{
							Op:           ir.OpFieldSet,
							Type:         ir.TypeVoid,
							Callee:       shapeName,
							Field:        field.Name,
							FieldIndex:   fieldIdx,
							DynamicField: dynamicFieldAccess(shapeName),
							Args:         []string{arrVal, val},
							Span:         toIRSpan(path, statement.Span),
						})
						return nil
					}
				}
				if statement.Right != nil && statement.Right.Kind == "string" {
					propName := statement.Right.Text
					for idx, field := range shape.Fields {
						if field.Name == propName {
							val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
							if err != nil {
								return err
							}
							function.Body = append(function.Body, ir.Instruction{
								Op:           ir.OpFieldSet,
								Type:         ir.TypeVoid,
								Callee:       shapeName,
								Field:        field.Name,
								FieldIndex:   idx,
								DynamicField: dynamicFieldAccess(shapeName),
								Args:         []string{arrVal, val},
								Span:         toIRSpan(path, statement.Span),
							})
							return nil
						}
					}
				}
			}
		}
		// Dictionary-like objects use the runtime key instead of a fixed shape
		// field. Keep the array path below strict about numeric indexes.
		if arrType == ir.TypeObject || strings.HasPrefix(string(arrType), "object:") || arrType == ir.TypeUnknown {
			idxVal, idxType, err := lowerExpression(path, statement.Right, "", function, env, counter, shapes, signatures)
			if err != nil {
				return err
			}
			if idxType == ir.TypeString {
				val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
				if err != nil {
					return err
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeVoid,
					Callee: "__object.set_prop",
					Args:   []string{arrVal, idxVal, val},
					Span:   toIRSpan(path, statement.Span),
				})
				return nil
			}
		}
		if arrType == ir.TypeString {
			return fmt.Errorf("cannot assign to read-only string index")
		}
		idxVal, idxType, err := lowerExpression(path, statement.Right, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if idxType != ir.TypeNumber {
			return fmt.Errorf("array index_set requires number index, got %s", idxType)
		}
		var expectedElemType ir.Type
		if arrType == ir.TypeBigInt64Array || arrType == ir.TypeBigUint64Array {
			expectedElemType = ir.TypeBigInt
		} else if isNumberTypedArray(arrType) || arrType == ir.TypeNumberArray {
			expectedElemType = ir.TypeNumber
		} else if arrType == ir.TypeStringArray {
			expectedElemType = ir.TypeString
		} else if arrType == ir.TypeBoolArray || arrType == "boolean[]" || arrType == "bool[]" {
			expectedElemType = ir.TypeBool
		} else if before, ok := strings.CutSuffix(string(arrType), "[]"); ok {
			elemName := before
			if elemName == "boolean" {
				expectedElemType = ir.TypeBool
			} else {
				expectedElemType = ir.Type(elemName)
			}
		} else {
			return fmt.Errorf("array index_set requires an array, got %s", arrType)
		}
		if statement.Expression != nil && (statement.Expression.InferredType == "" || statement.Expression.InferredType == "never[]" || statement.Expression.InferredType == "unknown[]") && expectedElemType != "" {
			statement.Expression.InferredType = string(expectedElemType)
		}
		val, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if valType == ir.TypeVoid || (statement.Expression != nil && (statement.Expression.Kind == "undefined" || statement.Expression.Kind == "null" || (statement.Expression.Kind == "identifier" && statement.Expression.Text == "undefined"))) {
			zeroVal := nextTemp(counter)
			switch expectedElemType {
			case ir.TypeNumber:
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroVal, Value: "0", Span: toIRSpan(path, statement.Span)})
			case ir.TypeBool:
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: zeroVal, Value: "false", Span: toIRSpan(path, statement.Span)})
			case ir.TypeString:
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: zeroVal, Value: "", Span: toIRSpan(path, statement.Span)})
			default:
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: expectedElemType, Result: zeroVal, Value: "null", Span: toIRSpan(path, statement.Span)})
			}
			val = zeroVal
			valType = expectedElemType
		}
		if expectedElemType == ir.TypeUnknown && valType != ir.TypeUnknown {
			boxedVal := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: boxedVal,
				Args:   []string{val},
				Span:   toIRSpan(path, statement.Span),
			})
			env[boxedVal] = ir.TypeUnknown
			val = boxedVal
			valType = ir.TypeUnknown
		}
		if expectedElemType != "" && valType != expectedElemType && valType != ir.TypeUnknown && expectedElemType != ir.TypeUnknown {
			if !strings.HasSuffix(string(expectedElemType), "[]") || (valType != "never[]" && valType != "object:never[]" && valType != ir.TypeObject && valType != "never" && valType != "unknown[]") {
				return fmt.Errorf("array index_set type mismatch: %s cannot be assigned to %s", valType, arrType)
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpIndexSet,
			Type: ir.TypeVoid,
			Args: []string{arrVal, idxVal, val},
			Span: toIRSpan(path, statement.Span),
		})
	case "field_set":
		if statement.Left != nil && statement.Left.Kind == "identifier" {
			className := statement.Left.Text
			if meta, isClass := classHierarchy[className]; isClass {
				// Check static setter
				if _, setterName, ok := findSetterInHierarchy(className, statement.Name, signatures, classHierarchy); ok {
					val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
					if err != nil {
						return err
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: setterName,
						Args:   []string{val},
						Span:   toIRSpan(path, statement.Span),
					})
					return nil
				}
				// Static field assignment
				if _, isStatic := meta.Statics[statement.Name]; isStatic {
					staticVar := className + "_" + statement.Name
					val, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
					if err != nil {
						return err
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpAssign,
						Type:   valType,
						Result: staticVar,
						Args:   []string{val},
						Span:   toIRSpan(path, statement.Span),
					})
					return nil
				}
			}
		}

		objVal, objType, err := lowerExpression(path, statement.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if (strings.HasSuffix(string(objType), "[]") || objType == ir.TypeNumberArray || objType == ir.TypeStringArray || objType == ir.TypeBoolArray || objType == ir.TypeBigIntArray) && statement.Name == "length" {
			val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
			if err != nil {
				return err
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeVoid,
				Callee: "__array.set_length",
				Args:   []string{objVal, val},
				Span:   toIRSpan(path, statement.Span),
			})
			return nil
		}
		className := strings.TrimPrefix(string(objType), string(ir.TypeObject)+":")
		if className == "this" || className == "" {
			if statement.Left != nil && statement.Left.Text != "" {
				if t, inEnv := env[statement.Left.Text]; inEnv && string(t) != "" && string(t) != "this" {
					className = strings.TrimPrefix(string(t), string(ir.TypeObject)+":")
				}
			}
			if className == "this" || className == "" {
				if t, inEnv := env["this"]; inEnv && string(t) != "this" && string(t) != "object:this" {
					className = strings.TrimPrefix(string(t), string(ir.TypeObject)+":")
				} else if function != nil && strings.Contains(function.Name, "_") && !strings.HasPrefix(function.Name, "__closure_") {
					className = strings.Split(function.Name, "_")[0]
				}
			}
			if className == "this" || className == "" {
				for sName, s := range shapes {
					if fieldIndex(s, statement.Name) >= 0 {
						className = sName
						break
					}
				}
			}
		}

		// Check instance setter in hierarchy
		if _, setterName, ok := findSetterInHierarchy(className, statement.Name, signatures, classHierarchy); ok {
			val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
			if err != nil {
				return err
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeVoid,
				Callee: setterName,
				Args:   []string{objVal, val},
				Span:   toIRSpan(path, statement.Span),
			})
			return nil
		}

		shape, ok := shapes[className]
		if !ok {
			if s, exists := registeredShapes[className]; exists {
				shape = s
				shapes[className] = s
				ok = true
			} else if s, exists := anonymousShapes[className]; exists {
				shape = s
				shapes[className] = s
				ok = true
			}
		}
		if !ok {
			return fmt.Errorf("field set on unknown object shape %q", className)
		}
		fIndex := fieldIndex(shape, statement.Name)
		if fIndex < 0 {
			return fmt.Errorf("unknown field %q on object shape %q", statement.Name, className)
		}
		if statement.Expression != nil && (statement.Expression.Kind == "array" || (statement.Expression.Kind == "new" && callName(statement.Expression.Left) == "Array")) {
			targetFieldType := string(shape.Fields[fIndex].Type)
			if strings.HasSuffix(targetFieldType, "[]") || statement.Expression.InferredType == "" || statement.Expression.InferredType == "any[]" || statement.Expression.InferredType == "never[]" || statement.Expression.InferredType == "unknown[]" || statement.Expression.InferredType == "void[]" {
				statement.Expression.InferredType = targetFieldType
			}
		}
		var val string
		var valType ir.Type
		if statement.Expression != nil && (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && (isPointerLikeType(shape.Fields[fIndex].Type) || shape.Fields[fIndex].Type == ir.TypePointer) {
			val = nextTemp(counter)
			valType = shape.Fields[fIndex].Type
			value := "null"
			if statement.Expression.Kind == "undefined" {
				value = "undefined"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   valType,
				Result: val,
				Value:  value,
				Span:   toIRSpan(path, statement.Span),
			})
		} else {
			var err error
			val, valType, err = lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
			if err != nil {
				return err
			}
		}
		if valType != shape.Fields[fIndex].Type {
			if strings.HasSuffix(string(shape.Fields[fIndex].Type), "[]") && (valType == "void[]" || valType == "never[]" || valType == "unknown[]") {
				// allowed array assignment
			} else if valType == ir.TypePointer && (isPointerLikeType(shape.Fields[fIndex].Type) || shape.Fields[fIndex].Type == ir.TypePointer) {
				// allowed null pointer assignment to pointer-like field
			} else if (strings.HasPrefix(string(valType), "object:") || valType == ir.TypeObject) && (strings.HasPrefix(string(shape.Fields[fIndex].Type), "object:") || shape.Fields[fIndex].Type == ir.TypeObject) {
				// allowed object assignment
			} else if isSubtype(string(valType), string(shape.Fields[fIndex].Type)) {
				// allowed subtype/interface implementation assignment
			} else if shape.Fields[fIndex].Type == ir.TypeUnknown {
				boxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpBoxUnknown,
					Type:   ir.TypeUnknown,
					Result: boxed,
					Args:   []string{val},
					Span:   toIRSpan(path, statement.Span),
				})
				val = boxed
			} else if valType == ir.TypeUnknown {
				unboxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCheckedCast,
					Type:   shape.Fields[fIndex].Type,
					Result: unboxed,
					Args:   []string{val},
					Span:   toIRSpan(path, statement.Span),
				})
				val = unboxed
			} else {
				return fmt.Errorf("field set type mismatch for %q: %s := %s", statement.Name, shape.Fields[fIndex].Type, valType)
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:           ir.OpFieldSet,
			Type:         ir.TypeVoid,
			Callee:       className,
			Field:        statement.Name,
			FieldIndex:   fIndex,
			DynamicField: dynamicFieldAccess(className),
			Args:         []string{objVal, val},
			Span:         toIRSpan(path, statement.Span),
		})
	case "class":
		return nil
	case "throw":
		val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		bodyLenBeforeFinally := len(function.Body)
		if err := lowerActiveThrowFinally(path, function, env, counter, shapes, signatures); err != nil {
			return err
		}
		if len(function.Body) > bodyLenBeforeFinally && function.Body[len(function.Body)-1].Op == ir.OpReturn {
			return nil
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpThrow,
			Type: ir.TypeVoid,
			Args: []string{val},
			Span: toIRSpan(path, statement.Span),
		})
	case "try":
		return lowerTry(path, statement, function, env, counter, shapes, signatures)
	case "debugger":
		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpDebugger,
			Type: ir.TypeVoid,
			Span: toIRSpan(path, statement.Span),
		})
	case "import_alias":
		if statement.Name != "" && statement.Type != "" {
			env["__ident."+statement.Name] = ir.Type(statement.Type)
			if origType, ok := env[statement.Type]; ok {
				env[statement.Name] = origType
			}
		}
		return nil
	case "export_alias", "module", "enum":
		return nil
	default:
		return fmt.Errorf("unsupported statement %q", statement.Kind)
	}
	return nil
}

func lowerActiveReturnFinally(path string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	savedStack := activeReturnFinallyStack
	activeReturnFinallyStack = nil
	defer func() {
		activeReturnFinallyStack = savedStack
	}()

	for i := len(savedStack) - 1; i >= 0; i-- {
		activeReturnFinallyStack = savedStack[:i]
		for _, finStmt := range savedStack[i] {
			if err := lowerStatement(path, finStmt, function, env, counter, shapes, signatures); err != nil {
				return err
			}
			if len(function.Body) > 0 && function.Body[len(function.Body)-1].Op == ir.OpReturn {
				return nil
			}
		}
	}
	return nil
}

func lowerActiveThrowFinally(path string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	savedStack := activeThrowFinallyStack
	activeThrowFinallyStack = nil
	defer func() {
		activeThrowFinallyStack = savedStack
	}()

	for i := len(savedStack) - 1; i >= 0; i-- {
		activeThrowFinallyStack = savedStack[:i]
		for _, finStmt := range savedStack[i] {
			if err := lowerStatement(path, finStmt, function, env, counter, shapes, signatures); err != nil {
				return err
			}
			if len(function.Body) > 0 && function.Body[len(function.Body)-1].Op == ir.OpReturn {
				return nil
			}
		}
	}
	return nil
}

func lowerActiveBreakFinally(path string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	minDepth := 0
	if len(loopFinallyScopeStack) > 0 {
		minDepth = loopFinallyScopeStack[len(loopFinallyScopeStack)-1]
	}
	savedStack := activeReturnFinallyStack
	activeReturnFinallyStack = nil
	defer func() {
		activeReturnFinallyStack = savedStack
	}()

	for i := len(savedStack) - 1; i >= minDepth; i-- {
		activeReturnFinallyStack = savedStack[:i]
		for _, finStmt := range savedStack[i] {
			if err := lowerStatement(path, finStmt, function, env, counter, shapes, signatures); err != nil {
				return err
			}
			if len(function.Body) > 0 && function.Body[len(function.Body)-1].Op == ir.OpReturn {
				return nil
			}
		}
	}
	return nil
}

func lowerBranch(path string, statements []typescriptgo.SyntaxStatement, returnType ir.Type, parentEnv map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) ([]ir.Instruction, error) {
	branch := ir.Function{Name: "branch", ReturnType: returnType}
	env := make(map[string]ir.Type, len(parentEnv))
	maps.Copy(env, parentEnv)
	for _, statement := range statements {
		if err := lowerStatement(path, statement, &branch, env, counter, shapes, signatures); err != nil {
			return nil, err
		}
	}
	return branch.Body, nil
}

func isOptionalChainExpr(expr *typescriptgo.SyntaxExpression) bool {
	if expr == nil {
		return false
	}
	if expr.Kind == "optional_call" || expr.Kind == "optional_property" || expr.Kind == "optional_index" {
		return true
	}
	return isOptionalChainExpr(expr.Left) || isOptionalChainExpr(expr.Right)
}
