package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerCallExpression(
	path string,
	expression *frontend.SyntaxExpression,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, error) {
	if typed := typedFilledArray(expression); typed != nil {
		expression = typed
	}
	if expression.Left != nil && expression.Left.Kind == "identifier" {
		if dynamic, ok := dynamicImports[expression.Left.Text]; ok {
			return lowerDynamicImportCall(path, expression, result, function, env, counter, shapes, signatures, dynamic)
		}
	}
	// A non-null assertion is erased at runtime. Remove it at the call
	// boundary so `options.encode!()` follows the same closure/method path as
	// the unasserted expression while retaining the narrowed checker type.
	if expression.Left != nil && expression.Left.Kind == "non_null" && expression.Left.Left != nil {
		inner := *expression.Left.Left
		if expression.Left.InferredType != "" {
			inner.InferredType = expression.Left.InferredType
		}
		withoutAssertion := *expression
		withoutAssertion.Left = &inner
		return lowerCallExpression(path, &withoutAssertion, result, function, env, counter, shapes, signatures)
	}
	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "optional_property" || expression.Left.Kind == "index" || expression.Left.Kind == "optional_index") && expression.Left.Left != nil {
		isModuleNamespace := false
		if expression.Left.Left.Kind == "identifier" {
			qualifier := expression.Left.Left.Text
			if _, inEnv := env[qualifier]; !inEnv {
				if _, inTop := topLevelVars[qualifier]; !inTop {
					fullName := qualifier + "." + expression.Left.Text
					if _, ok := signatures[fullName]; ok {
						isModuleNamespace = true
					} else if _, ok := signatures[expression.Left.Text]; ok {
						isModuleNamespace = true
					}
				}
			}
		}
		if !isModuleNamespace {
			isIndexedCall := false
			methodName := expression.Left.Text
			if (expression.Left.Kind == "index" || expression.Left.Kind == "optional_index") && expression.Left.Right != nil {
				if expression.Left.Right.Kind == "property" && expression.Left.Right.Left != nil && expression.Left.Right.Left.Text == "Symbol" {
					methodName = "Symbol." + expression.Left.Right.Text
				} else if expression.Left.Right.Kind == "string" {
					methodName = expression.Left.Right.Text
				} else {
					isIndexedCall = true
				}
			}
			if !isIndexedCall {
				receiver, receiverType, err := lowerExpression(path, expression.Left.Left, "", function, env, counter, shapes, signatures)
				if err == nil {
					if (methodName == "then" || methodName == "catch") && isPromiseReceiverType(receiverType) {
						return lowerPromiseThenCatchCall(path, expression, result, function, env, counter, shapes, signatures, methodName, receiver)
					}
					if res, typ, handled, err := lowerWeakReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
						return res, typ, err
					}
					if res, typ, handled, err := lowerRegExpReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
						return res, typ, err
					}
					if res, typ, handled, err := lowerDateReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
						return res, typ, err
					}
					if receiverType == ir.TypeBigInt && methodName == "toLocaleString" {
						if result == "" {
							result = nextTemp(counter)
						}
						function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__string.fromBigIntLocale", Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
						return result, ir.TypeString, nil
					}
					if receiverType == ir.TypeBigInt && methodName == "toString" {
						if result == "" {
							result = nextTemp(counter)
						}
						function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__string.fromBigInt", Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
						return result, ir.TypeString, nil
					}
					if receiverType == ir.TypeBigInt && methodName == "valueOf" {
						return receiver, ir.TypeBigInt, nil
					}
					if (receiverType == ir.TypeObject || strings.HasPrefix(string(receiverType), "object:")) &&
						(strings.Contains(string(receiverType), "Error") || strings.HasSuffix(string(receiverType), "Error")) && methodName == "toString" {
						if result == "" {
							result = nextTemp(counter)
						}
						function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__string.errorToString", Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
						return result, ir.TypeString, nil
					}
					if receiverType == ir.TypeSymbol && methodName == "toString" {
						if result == "" {
							result = nextTemp(counter)
						}
						function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__symbol.toString", Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
						return result, ir.TypeString, nil
					}
					if receiverType == ir.TypeSymbol && methodName == "valueOf" {
						return receiver, ir.TypeSymbol, nil
					}
					if receiverType == ir.TypeArrayBuffer && methodName == "slice" {
						return lowerArrayBufferSliceCall(path, expression, result, function, env, counter, shapes, signatures, receiver)
					}
					if res, typ, handled, err := lowerBufferMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
						return res, typ, err
					}
					if res, typ, handled, err := lowerTypedArrayReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
						return res, typ, err
					}
					if res, typ, handled, err := lowerDataViewReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
						return res, typ, err
					}
					if res, typ, handled, err := lowerMapSetReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
						return res, typ, err
					}
					if res, typ, handled, err := lowerTextEncodingReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
						return res, typ, err
					}
					if res, typ, handled, err := lowerWeakReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
						return res, typ, err
					}
					if res, typ, handled, err := lowerIntlReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
						return res, typ, err
					}
					if receiverType == ir.TypeString {
						if res, typ, handled, err := lowerStringReceiverMethod(path, expression, receiver, methodName, result, function, env, counter, shapes, signatures); handled {
							return res, typ, err
						}
					}
					if receiverType == ir.TypeNumber && methodName == "valueOf" {
						return receiver, ir.TypeNumber, nil
					}
					if receiverType == ir.TypeNumber && (methodName == "toFixed" || methodName == "toString" || methodName == "toExponential" || methodName == "toPrecision" || methodName == "toLocaleString") {
						return lowerNumberFormatCall(path, expression, result, function, env, counter, shapes, signatures, methodName, receiver)
					}
					if res, typ, handled, err := lowerArrayReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
						return res, typ, err
					}
					if res, typ, handled, err := lowerIteratorReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
						return res, typ, err
					}
					if res, typ, handled, err := lowerFunctionValueMethod(path, expression, receiver, receiverType, methodName, result, function, env, counter, shapes, signatures); handled {
						return res, typ, err
					}
					className := strings.TrimPrefix(string(receiverType), "object:")
					className = classIdentityForPath(path, className)
					if className != "" && className != "number" && className != "string" && className != "bool" && className != "void" {
						if target, mangled, ok := findMethodInHierarchy(className, methodName, signatures, classHierarchy); ok {
							return lowerInstanceMethodCall(path, expression, result, function, env, counter, shapes, signatures, receiver, target, mangled)
						}
						if methodName == "hasOwnProperty" || methodName == "propertyIsEnumerable" {
							if len(expression.Arguments) > 0 {
								return lowerHasOwnPropertyCall(path, expression, result, function, env, counter, shapes, signatures, receiver)
							}
						}
						if methodName == "isPrototypeOf" && len(expression.Arguments) == 1 {
							value, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
							if err != nil {
								return "", "", err
							}
							if result == "" {
								result = nextTemp(counter)
							}
							function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeBool, Result: result, Callee: "__object.isPrototypeOf", Args: []string{receiver, value}, Span: toIRSpan(path, expression.Span)})
							return result, ir.TypeBool, nil
						}
						if value, typ, handled := lowerInheritedObjectMethod(path, expression, receiver, receiverType, className, methodName, result, function, counter, shapes); handled {
							return value, typ, nil
						}
						propVal, propType, err := lowerPropertyExpression(path, &frontend.SyntaxExpression{
							Span:         expression.Left.Span,
							Kind:         "property",
							Left:         expression.Left.Left,
							Text:         methodName,
							InferredType: "",
						}, "", function, env, counter, shapes, signatures)
						if err == nil {
							return tryLowerPropertyClosureCall(path, expression, result, function, env, counter, shapes, signatures, receiver, receiverType, propVal, propType)
						}
					}
				}
			}
		}
	}

	// super(...) call in constructor
	if expression.Left != nil && expression.Left.Kind == "identifier" && expression.Left.Text == "super" {
		return lowerSuperConstructorCall(path, expression, function, env, counter, shapes, signatures)
	}

	// super.method(...) call
	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "member") && expression.Left.Left != nil && expression.Left.Left.Text == "super" {
		return lowerSuperMethodCall(path, expression, result, function, env, counter, shapes, signatures)
	}

	// Static method call ClassName.method(...) or this.method(...) inside static methods
	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "member") && expression.Left.Left != nil {
		className := ""
		if expression.Left.Left.Kind == "identifier" && expression.Left.Left.Text != "this" {
			className = classIdentityForPath(path, expression.Left.Left.Text)
		} else if (expression.Left.Left.Kind == "this" || (expression.Left.Left.Kind == "identifier" && expression.Left.Left.Text == "this")) && strings.Contains(function.Name, "_static_") {
			className = strings.Split(function.Name, "_static_")[0]
		}
		if className != "" {
			methodName := expression.Left.Text
			if target, mangled, found := findStaticMethodInHierarchy(className, methodName, signatures, classHierarchy); found {
				return lowerStaticMethodCall(path, expression, result, function, env, counter, shapes, signatures, target, mangled)
			}
		}
	}

	if expression.Left != nil && expression.Left.Kind == "arrow_function" {
		return lowerImmediateArrowCall(path, expression, result, function, env, counter, shapes, signatures)
	}

	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "member" || expression.Left.Kind == "index" || expression.Left.Kind == "call" || expression.Left.Kind == "paren" || expression.Left.Kind == "optional_call" || expression.Left.Kind == "as" || expression.Left.Kind == "cast" || expression.Left.Kind == "type_assertion") {
		isModuleFunc := false
		if (expression.Left.Kind == "property" || expression.Left.Kind == "optional_property") && expression.Left.Left != nil && expression.Left.Left.Kind == "identifier" {
			if _, inEnv := env[expression.Left.Left.Text]; !inEnv {
				if _, inTop := topLevelVars[expression.Left.Left.Text]; !inTop {
					fullName := expression.Left.Left.Text + "." + expression.Left.Text
					if _, ok := signatures[fullName]; ok {
						isModuleFunc = true
					} else if _, ok := signatures[expression.Left.Text]; ok {
						isModuleFunc = true
					}
				}
			}
		}
		if !isModuleFunc {
			closureVal, closureType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
			if err == nil && (closureType == ir.TypeDynamicFunction || (currentDynamicMode && closureType == ir.TypeUnknown) || closureType == ir.TypeClosure || closureType == "Function" || closureType == "function" || strings.Contains(string(closureType), "=>")) {
				return tryLowerDynamicPropertyCall(path, expression, result, function, env, counter, shapes, signatures, closureVal, closureType)
			}
		}
	}

	callee := callName(expression.Left)
	if aliased, ok := env["__ident."+callee]; ok && aliased != "" {
		callee = string(aliased)
	}
	if callee == "gc" {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: result,
			Callee: "__gc.collect",
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeNumber, nil
	}

	if res, typ, handled, err := lowerPromiseStaticCall(path, callee, expression, result, function, env, counter, shapes, signatures); handled {
		return res, typ, err
	}

	if callee == "Intl.getCanonicalLocales" {
		return lowerIntlGetCanonicalLocalesCall(path, expression, result, function, env, counter, shapes, signatures)
	}

	calleeIsClosure := false
	calleeType, hasCalleeType := env[callee]
	if hasCalleeType && (calleeType == ir.TypeDynamicFunction || (currentDynamicMode && calleeType == ir.TypeUnknown)) {
		return lowerDynamicFunctionCall(path, expression, result, function, env, counter, shapes, signatures, callee)
	}
	if hasCalleeType && (calleeType == ir.TypeClosure || calleeType == ir.TypeUnknown || calleeType == ir.TypeObject || calleeType == "Function" || calleeType == "function" || strings.Contains(string(calleeType), "=>")) {
		calleeIsClosure = true
	} else if topVar, isTop := topLevelVars[callee]; isTop && ((topVar.Expression != nil && (topVar.Expression.Kind == "arrow_function" || topVar.Expression.Kind == "function")) || strings.Contains(topVar.Type, "=>") || strings.Contains(topVar.InferredType, "=>")) {
		calleeIsClosure = true
	}

	if calleeIsClosure {
		return lowerClosureValueCall(path, expression, result, function, env, counter, shapes, signatures, callee)
	}

	if intrinsic, ok := builtinIntrinsic(callee); ok {
		return intrinsic.Lower(IntrinsicCall{Path: path, Expression: expression, Result: result, Function: function, Env: env, Counter: counter, Shapes: shapes, Signatures: signatures, LowerExpression: lowerExpression}, intrinsic)
	}

	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "member" || expression.Left.Kind == "index" || expression.Left.Kind == "optional_index" || expression.Left.Kind == "optional_property" || expression.Left.Kind == "as" || expression.Left.Kind == "cast" || expression.Left.Kind == "type_assertion" || expression.Left.Kind == "paren") {
		isModuleFunc := false
		if (expression.Left.Kind == "property" || expression.Left.Kind == "optional_property") && expression.Left.Left != nil && expression.Left.Left.Kind == "identifier" {
			if _, inEnv := env[expression.Left.Left.Text]; !inEnv {
				if _, inTop := topLevelVars[expression.Left.Left.Text]; !inTop {
					fullName := expression.Left.Left.Text + "." + expression.Left.Text
					if _, ok := signatures[fullName]; ok {
						isModuleFunc = true
					} else if _, ok := signatures[expression.Left.Text]; ok {
						isModuleFunc = true
					}
				}
			}
		}
		if !isModuleFunc {
			closureVal, closureType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
			if err == nil && (closureType == ir.TypeClosure || closureType == ir.TypeUnknown || strings.HasPrefix(string(closureType), "object:") || strings.Contains(string(closureType), "=>")) {
				return tryLowerClosurePropertyCall(path, expression, result, function, env, counter, shapes, signatures, closureVal, closureType)
			}
		}
	}

	if callee == "" {
		leftKind := ""
		leftText := ""
		receiverKind := ""
		receiverText := ""
		if expression.Left != nil {
			leftKind = expression.Left.Kind
			leftText = expression.Left.Text
			if expression.Left.Left != nil {
				receiverKind = expression.Left.Left.Kind
				receiverText = expression.Left.Left.Text
			}
		}
		return "", "", fmt.Errorf("unsupported call target (kind: %s, leftKind: %s, leftText: %s, recvKind: %s, recvText: %s, args: %d)", expression.Kind, leftKind, leftText, receiverKind, receiverText, len(expression.Arguments))
	}

	target, ok := resolveFunctionSignature(path, callee, signatures)
	if !ok {
		target, ok = resolveFallbackCallTarget(path, expression, env, signatures, callee, target, ok)
	}
	if !ok {
		return "", "", fmt.Errorf("unknown function %q", callee)
	}

	callee = target.Name
	args := make([]string, 0, len(expression.Arguments))
	paramOffset := 0
	if len(target.Parameters) > 0 && target.Parameters[0].Name == "this" {
		var receiver string
		if expression.Left != nil {
			receiverExpression := expression.Left.Left
			if expression.Left.Kind == "property" || expression.Left.Kind == "member" || expression.Left.Kind == "optional_property" {
				var err error
				receiver, _, err = lowerExpression(path, receiverExpression, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
			}
		}
		if receiver == "" {
			return "", "", fmt.Errorf("method %q has no receiver", callee)
		}
		args = append(args, receiver)
		paramOffset = 1
	}
	restIndex := restParameterIndex(target, callee)
	arguments, restPacked, err := spreadRestArguments(expression.Arguments, expression.Span, target, restIndex, paramOffset, false)
	if err != nil {
		return "", "", err
	}
	for aIdx, argument := range arguments {
		pIdx := aIdx + paramOffset
		if argument.Kind == "array" && pIdx < len(target.Parameters) {
			paramType := target.Parameters[pIdx].Type
			shapeName := strings.TrimPrefix(string(paramType), "object:")
			if shape, ok := shapes[shapeName]; ok && len(shape.Fields) > 0 && shape.Fields[0].Name == "0" {
				argument.InferredType = string(paramType)
			} else if fields, isTup := tupleFields(string(paramType)); isTup && len(fields) > 0 {
				argument.InferredType = string(paramType)
			}
		}
		defaults := defaultParamsIndex[callee]
		if defaults == nil {
			defaults = defaultParamsIndex[strings.Split(callee, "__")[0]]
		}
		if (argument.Kind == "undefined" || (argument.Kind == "identifier" && argument.Text == "undefined")) && defaults != nil && defaults[pIdx] != nil && defaults[pIdx].Kind != "undefined" {
			paramMap := make(map[string]string)
			for j := 0; j < len(args) && j < len(target.Parameters); j++ {
				pName := target.Parameters[j].Name
				if pName != "this" && pName != "" {
					paramMap[pName] = args[j]
					if _, exists := env[args[j]]; !exists {
						env[args[j]] = target.Parameters[j].Type
					}
				}
			}
			initExpr := substituteParamIdentifiers(defaults[pIdx], paramMap)
			argument = initExpr
		}
		value, valType, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if pIdx < len(target.Parameters) {
			value, valType, _ = adaptStructuralObjectArgument(
				path,
				toIRSpan(path, argument.Span),
				value,
				valType,
				target.Parameters[pIdx].Type,
				function,
				counter,
				shapes,
			)
		}
		hasRest := (restParamsIndex[callee] || restParamsIndex[strings.Split(callee, "__")[0]]) && len(target.Parameters) > 0
		restIsUnknown := hasRest && target.Parameters[len(target.Parameters)-1].Type == ir.TypeUnknownArray
		fixed := len(target.Parameters) - 1
		needsUnknownBox := (pIdx < len(target.Parameters) && target.Parameters[pIdx].Type == ir.TypeUnknown) || (restIsUnknown && pIdx >= fixed && !restPacked)
		if needsUnknownBox && valType != ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: boxed,
				Args:   []string{value},
				Span:   toIRSpan(path, argument.Span),
			})
			value = boxed
		}
		if pIdx < len(target.Parameters) && target.Parameters[pIdx].Type == ir.TypeClosure && valType != ir.TypeClosure {
			if !strings.HasPrefix(value, "%") && valType != ir.TypeVoid && valType != "null" && valType != "ptr" {
				if sig, isSig := signatures[value]; isSig {
					closureSlot := nextTemp(counter)
					calleeName := ensureFunctionClosureTrampoline(path, sig, shapes, signatures)
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpClosure,
						Type:   ir.TypeClosure,
						Result: closureSlot,
						Callee: calleeName,
						Args:   nil,
						Span:   toIRSpan(path, argument.Span),
					})
					value = closureSlot
				}
			}
		}
		if pIdx < len(target.Parameters) && isPointerLikeType(target.Parameters[pIdx].Type) && (argument.Kind == "null" || argument.Kind == "undefined") {
			valStr := "null"
			if argument.Kind == "undefined" {
				valStr = "undefined"
			}
			constTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   target.Parameters[pIdx].Type,
				Result: constTemp,
				Value:  valStr,
				Span:   toIRSpan(path, argument.Span),
			})
			value = constTemp
		}
		args = append(args, value)
	}

	if !restPacked && (restParamsIndex[callee] || restParamsIndex[strings.Split(callee, "__")[0]]) && len(target.Parameters) > 0 && (strings.HasSuffix(string(target.Parameters[len(target.Parameters)-1].Type), "[]") || target.Parameters[len(target.Parameters)-1].Type == ir.TypeStringArray || target.Parameters[len(target.Parameters)-1].Type == ir.TypeNumberArray) {
		restType := target.Parameters[len(target.Parameters)-1].Type
		fixed := len(target.Parameters) - 1
		if len(args) >= fixed {
			restArgs := append([]string(nil), args[fixed:]...)
			args = args[:fixed]
			array := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: restType, Result: array, Args: restArgs, Span: toIRSpan(path, expression.Span)})
			args = append(args, array)
		}
	}

	if len(args) < len(target.Parameters) {
		defaults := defaultParamsIndex[callee]
		if defaults == nil {
			defaults = defaultParamsIndex[strings.Split(callee, "__")[0]]
		}
		paramMap := make(map[string]string)
		for j := 0; j < len(args) && j < len(target.Parameters); j++ {
			pName := target.Parameters[j].Name
			if pName != "this" && pName != "" {
				paramMap[pName] = args[j]
				if _, exists := env[args[j]]; !exists {
					env[args[j]] = target.Parameters[j].Type
				}
			}
		}
		for i := len(args); i < len(target.Parameters); i++ {
			if i == restIndex {
				args = append(args, emptyRestArray(path, expression.Span, target.Parameters[i].Type, function, counter))
				restPacked = true
				continue
			}
			if defaults != nil {
				if initExpr, ok := defaults[i]; ok {
					initExpr = substituteParamIdentifiers(initExpr, paramMap)
					if target.Parameters[i].Type == ir.TypeNumber && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
						numConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   ir.TypeNumber,
							Result: numConst,
							Value:  initExpr.Kind, // number storage holds undefined and null as markers
							Span:   toIRSpan(path, initExpr.Span),
						})
						args = append(args, numConst)
						continue
					} else if target.Parameters[i].Type == ir.TypeBool && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
						boolConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   ir.TypeBool,
							Result: boolConst,
							Value:  "false",
							Span:   toIRSpan(path, initExpr.Span),
						})
						args = append(args, boolConst)
						continue
					} else if isPointerLikeType(target.Parameters[i].Type) && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
						valStr := "null"
						if initExpr.Kind == "undefined" {
							valStr = "undefined"
						}
						nullConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   target.Parameters[i].Type,
							Result: nullConst,
							Value:  valStr,
							Span:   toIRSpan(path, initExpr.Span),
						})
						args = append(args, nullConst)
						continue
					}
					val, valType, err := lowerExpression(path, initExpr, "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					env[val] = valType
					if i < len(target.Parameters) && target.Parameters[i].Type == ir.TypeUnknown && valType != ir.TypeUnknown {
						boxed := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpBoxUnknown,
							Type:   ir.TypeUnknown,
							Result: boxed,
							Args:   []string{val},
							Span:   toIRSpan(path, initExpr.Span),
						})
						val = boxed
					}
					if i < len(target.Parameters) && target.Parameters[i].Type == ir.TypeClosure && valType != ir.TypeClosure {
						if !strings.HasPrefix(val, "%") && valType != ir.TypeVoid && valType != "null" && valType != "ptr" {
							if sig, isSig := signatures[val]; isSig {
								closureSlot := nextTemp(counter)
								calleeName := ensureFunctionClosureTrampoline(path, sig, shapes, signatures)
								function.Body = append(function.Body, ir.Instruction{
									Op:     ir.OpClosure,
									Type:   ir.TypeClosure,
									Result: closureSlot,
									Callee: calleeName,
									Args:   nil,
									Span:   toIRSpan(path, initExpr.Span),
								})
								val = closureSlot
							}
						}
					}
					pName := target.Parameters[i].Name
					if pName != "" && pName != "this" {
						paramMap[pName] = val
						if _, exists := env[val]; !exists {
							env[val] = target.Parameters[i].Type
						}
					}
					args = append(args, val)
					continue
				}
			}
			if target.Parameters[i].Type == ir.TypeUnknown {
				undefConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeVoid,
					Result: undefConst,
					Value:  "undefined",
					Span:   toIRSpan(path, expression.Span),
				})
				boxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpBoxUnknown,
					Type:   ir.TypeUnknown,
					Result: boxed,
					Args:   []string{undefConst},
					Span:   toIRSpan(path, expression.Span),
				})
				args = append(args, boxed)
			} else if target.Parameters[i].Type == ir.TypeNumber {
				numConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeNumber,
					Result: numConst,
					Value:  "undefined", // an omitted argument
					Span:   toIRSpan(path, expression.Span),
				})
				args = append(args, numConst)
			} else if target.Parameters[i].Type == ir.TypeBool {
				boolConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeBool,
					Result: boolConst,
					Value:  "false",
					Span:   toIRSpan(path, expression.Span),
				})
				args = append(args, boolConst)
			} else {
				undefConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeString,
					Result: undefConst,
					Value:  "undefined",
					Span:   toIRSpan(path, expression.Span),
				})
				args = append(args, undefConst)
			}
		}
	}

	if result == "" {
		result = nextTemp(counter)
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   target.ReturnType,
		Result: result,
		Callee: callee,
		Args:   args,
		Span:   toIRSpan(path, expression.Span),
	})
	return result, target.ReturnType, nil
}

func substituteParamIdentifiers(expr *frontend.SyntaxExpression, paramMap map[string]string) *frontend.SyntaxExpression {
	if expr == nil {
		return nil
	}
	copy := *expr
	if copy.Kind == "identifier" {
		if replacement, ok := paramMap[copy.Text]; ok {
			copy.Text = replacement
		}
	}
	if copy.Left != nil {
		copy.Left = substituteParamIdentifiers(copy.Left, paramMap)
	}
	if copy.Right != nil {
		copy.Right = substituteParamIdentifiers(copy.Right, paramMap)
	}
	if len(copy.Arguments) > 0 {
		newArgs := make([]*frontend.SyntaxExpression, len(copy.Arguments))
		for k, arg := range copy.Arguments {
			newArgs[k] = substituteParamIdentifiers(arg, paramMap)
		}
		copy.Arguments = newArgs
	}
	return &copy
}
