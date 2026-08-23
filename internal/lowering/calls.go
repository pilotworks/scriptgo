package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerCallExpression(
	path string,
	expression *typescriptgo.SyntaxExpression,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, error) {
	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "optional_property") && expression.Left.Left != nil {
		methodName := expression.Left.Text
		receiver, receiverType, err := lowerExpression(path, expression.Left.Left, "", function, env, counter, shapes, signatures)
		if err == nil {
			if res, typ, handled, err := lowerWeakReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
				return res, typ, err
			}
			if res, typ, handled, err := lowerRegExpReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
				return res, typ, err
			}
			if res, typ, handled, err := lowerDateReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
				return res, typ, err
			}
			if res, typ, handled, err := lowerIntlReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
				return res, typ, err
			}
			if receiverType == ir.TypeBigInt && methodName == "toString" {
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__string.fromBigInt", Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
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
				beginVal := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op: ir.OpConst, Type: ir.TypeNumber, Result: beginVal, Value: "0", Span: toIRSpan(path, expression.Span),
				})
				endVal := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeNumber,
					Result: endVal,
					Callee: "__arraybuffer.byteLength",
					Args:   []string{receiver},
					Span:   toIRSpan(path, expression.Span),
				})
				if len(expression.Arguments) > 0 {
					b, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					beginVal = b
				}
				if len(expression.Arguments) > 1 {
					e, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					endVal = e
				}
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeArrayBuffer,
					Result: result,
					Callee: "__arraybuffer.slice",
					Args:   []string{receiver, beginVal, endVal},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, ir.TypeArrayBuffer, nil
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
			if receiverType == ir.TypeString {
				if res, typ, handled, err := lowerStringReceiverMethod(path, expression, receiver, methodName, result, function, env, counter, shapes, signatures); handled {
					return res, typ, err
				}
			}
			if receiverType == ir.TypeNumber && methodName == "valueOf" {
				return receiver, ir.TypeNumber, nil
			}
			if receiverType == ir.TypeNumber && (methodName == "toFixed" || methodName == "toString" || methodName == "toExponential" || methodName == "toPrecision" || methodName == "toLocaleString") {
				args := []string{receiver}
				for _, argument := range expression.Arguments {
					value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					args = append(args, value)
				}
				if result == "" {
					result = nextTemp(counter)
				}
				env[result] = ir.TypeString
				callee := "__number." + methodName
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: callee, Args: args, Span: toIRSpan(path, expression.Span)})
				return result, ir.TypeString, nil
			}
			if res, typ, handled, err := lowerArrayReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
				return res, typ, err
			}
			if after, ok := strings.CutPrefix(string(receiverType), "object:"); ok || receiverType == ir.TypeObject {
				className := after
				if target, mangled, ok := findMethodInHierarchy(className, methodName, signatures, classHierarchy); ok {
					args := []string{receiver}
					for aIdx, argument := range expression.Arguments {
						argVal, valType, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						pIdx := aIdx + 1
						if pIdx < len(target.Parameters) && target.Parameters[pIdx].Type == ir.TypeUnknown && valType != ir.TypeUnknown {
							boxed := nextTemp(counter)
							function.Body = append(function.Body, ir.Instruction{
								Op:     ir.OpBoxUnknown,
								Type:   ir.TypeUnknown,
								Result: boxed,
								Args:   []string{argVal},
								Span:   toIRSpan(path, argument.Span),
							})
							argVal = boxed
						}
						args = append(args, argVal)
					}
					if len(args) < len(target.Parameters) {
						defaults := defaultParamsIndex[mangled]
						if defaults != nil {
							for i := len(args); i < len(target.Parameters); i++ {
								if initExpr, ok := defaults[i]; ok {
									if target.Parameters[i].Type == ir.TypeNumber && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
										numConst := nextTemp(counter)
										function.Body = append(function.Body, ir.Instruction{
											Op:     ir.OpConst,
											Type:   ir.TypeNumber,
											Result: numConst,
											Value:  "0",
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
									}
									val, valType, err := lowerExpression(path, initExpr, "", function, env, counter, shapes, signatures)
									if err != nil {
										return "", "", err
									}
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
									args = append(args, val)
								}
							}
						}
					}
					if restParamsIndex[mangled] && len(target.Parameters) > 0 && strings.HasSuffix(string(target.Parameters[len(target.Parameters)-1].Type), "[]") {
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
					if result == "" {
						result = nextTemp(counter)
					}
					retType := target.ReturnType
					if retType == "" {
						retType = ir.TypeVoid
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   retType,
						Result: result,
						Callee: mangled,
						Args:   args,
						Span:   toIRSpan(path, expression.Span),
					})
					return result, retType, nil
				}
				if methodName == "hasOwnProperty" || methodName == "propertyIsEnumerable" {
					if len(expression.Arguments) > 0 {
						propVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						if result == "" {
							result = nextTemp(counter)
						}
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpCall,
							Type:   ir.TypeBool,
							Result: result,
							Callee: "__object.hasOwn",
							Args:   []string{receiver, propVal},
							Span:   toIRSpan(path, expression.Span),
						})
						return result, ir.TypeBool, nil
					}
				}
				if methodName == "isPrototypeOf" {
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   ir.TypeBool,
						Result: result,
						Value:  "false",
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeBool, nil
				}
				if methodName == "toString" || methodName == "toLocaleString" {
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   ir.TypeString,
						Result: result,
						Value:  "[object Object]",
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeString, nil
				}
				if methodName == "valueOf" {
					return receiver, receiverType, nil
				}
			}
		}
	}

	// super(...) call in constructor
	if expression.Left != nil && expression.Left.Kind == "identifier" && expression.Left.Text == "super" {
		thisType, ok := env["this"]
		if !ok {
			return "", "", fmt.Errorf("super() can only be used inside a class constructor")
		}
		currentClass := strings.TrimPrefix(string(thisType), "object:")
		meta := classHierarchy[currentClass]
		if meta.Extends == "" {
			return "", "", fmt.Errorf("super() called in class %q with no base class", currentClass)
		}
		ctor, ctorName, found := findConstructorInHierarchy(meta.Extends, signatures, classHierarchy)
		if !found {
			if len(expression.Arguments) == 0 {
				return "", ir.TypeVoid, nil
			}
			return "", "", fmt.Errorf("super constructor not found for base class %q", meta.Extends)
		}
		args := []string{"this"}
		for _, argument := range expression.Arguments {
			val, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			args = append(args, val)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ctor.ReturnType,
			Callee: ctorName,
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return "", ir.TypeVoid, nil
	}

	// super.method(...) call
	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "member") && expression.Left.Left != nil && expression.Left.Left.Text == "super" {
		thisType, ok := env["this"]
		if !ok {
			return "", "", fmt.Errorf("super.method() can only be used inside a class method")
		}
		currentClass := strings.TrimPrefix(string(thisType), "object:")
		meta := classHierarchy[currentClass]
		if meta.Extends == "" {
			return "", "", fmt.Errorf("super.%s called in class %q with no base class", expression.Left.Text, currentClass)
		}
		target, mangled, found := findMethodInHierarchy(meta.Extends, expression.Left.Text, signatures, classHierarchy)
		if !found {
			return "", "", fmt.Errorf("super method %q not found in base class %q", expression.Left.Text, meta.Extends)
		}
		args := []string{"this"}
		for _, argument := range expression.Arguments {
			val, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			args = append(args, val)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   target.ReturnType,
			Result: result,
			Callee: mangled,
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, target.ReturnType, nil
	}

	// Static method call ClassName.method(...)
	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "member") && expression.Left.Left != nil && expression.Left.Left.Kind == "identifier" {
		className := expression.Left.Left.Text
		methodName := expression.Left.Text
		if target, mangled, found := findStaticMethodInHierarchy(className, methodName, signatures, classHierarchy); found {
			args := make([]string, 0, len(expression.Arguments))
			for aIdx, argument := range expression.Arguments {
				val, valType, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				if aIdx < len(target.Parameters) && target.Parameters[aIdx].Type == ir.TypeUnknown && valType != ir.TypeUnknown {
					boxed := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpBoxUnknown,
						Type:   ir.TypeUnknown,
						Result: boxed,
						Args:   []string{val},
						Span:   toIRSpan(path, argument.Span),
					})
					val = boxed
				}
				args = append(args, val)
			}
			if len(args) < len(target.Parameters) {
				defaults := defaultParamsIndex[mangled]
				if defaults != nil {
					for i := len(args); i < len(target.Parameters); i++ {
						if initExpr, ok := defaults[i]; ok {
							val, valType, err := lowerExpression(path, initExpr, "", function, env, counter, shapes, signatures)
							if err != nil {
								return "", "", err
							}
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
							args = append(args, val)
						}
					}
				}
			}
			if restParamsIndex[mangled] && len(target.Parameters) > 0 && strings.HasSuffix(string(target.Parameters[len(target.Parameters)-1].Type), "[]") {
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
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   target.ReturnType,
				Result: result,
				Callee: mangled,
				Args:   args,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, target.ReturnType, nil
		}
	}

	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "member") && (expression.Left.Text == "then" || expression.Left.Text == "catch") {
		receiverVal, _, err := lowerExpression(path, expression.Left.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		cbVal := ""
		if len(expression.Arguments) > 0 {
			v, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			cbVal = v
		}
		if result == "" {
			result = nextTemp(counter)
		}
		callee := "__async.promise_then"
		if expression.Left.Text == "catch" {
			callee = "__async.promise_catch"
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.Type("object:Promise"),
			Result: result,
			Callee: callee,
			Args:   []string{receiverVal, cbVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.Type("object:Promise"), nil
	}

	if expression.Left != nil && expression.Left.Kind == "arrow_function" {
		closureVal, _, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		args := make([]string, 0, len(expression.Arguments))
		for _, argument := range expression.Arguments {
			value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			args = append(args, value)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		retType := ir.TypeNumber
		if rt, ok := env[closureVal+".retType"]; ok && rt != "" {
			retType = rt
		} else if target, ok := signatures[closureVal]; ok && target.ReturnType != "" {
			retType = target.ReturnType
		} else if expression.InferredType != "" {
			retType = toIRType(expression.InferredType)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpClosureCall,
			Type:   retType,
			Result: result,
			Callee: closureVal,
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, nil
	}

	callee := callName(expression.Left)
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
		args := make([]string, 0, len(expression.Arguments))
		for _, argument := range expression.Arguments {
			value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			args = append(args, value)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeStringArray,
			Result: result,
			Callee: "__intl.get_canonical_locales",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeStringArray, nil
	}

	if calleeType, ok := env[callee]; ok && (calleeType == ir.TypeClosure || calleeType == ir.TypeUnknown || strings.HasPrefix(string(calleeType), "object:") || strings.Contains(string(calleeType), "=>")) {
		args := make([]string, 0, len(expression.Arguments))
		for _, argument := range expression.Arguments {
			value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			args = append(args, value)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		retType := ir.TypeNumber
		if rt, ok := env[callee+".retType"]; ok && rt != "" {
			retType = rt
		} else if target, ok := signatures[callee]; ok && target.ReturnType != "" {
			retType = target.ReturnType
		} else if expression.InferredType != "" {
			retType = toIRType(expression.InferredType)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpClosureCall,
			Type:   retType,
			Result: result,
			Callee: callee,
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, nil
	}

	if intrinsic, ok := builtinIntrinsic(callee); ok {
		return intrinsic.Lower(IntrinsicCall{Path: path, Expression: expression, Result: result, Function: function, Env: env, Counter: counter, Shapes: shapes, Signatures: signatures, LowerExpression: lowerExpression}, intrinsic)
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

	target, ok := signatures[callee]
	if !ok {
		if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "member" || expression.Left.Kind == "index") {
			closureVal, closureType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
			if err == nil && (closureType == ir.TypeClosure || closureType == ir.TypeUnknown || strings.HasPrefix(string(closureType), "object:")) {
				args := make([]string, 0, len(expression.Arguments))
				for _, argument := range expression.Arguments {
					value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					args = append(args, value)
				}
				if result == "" {
					result = nextTemp(counter)
				}
				retType := ir.TypeNumber
				if rt, ok := env[closureVal+".retType"]; ok && rt != "" {
					retType = rt
				} else if expression.InferredType != "" {
					retType = toIRType(expression.InferredType)
				} else if strings.HasPrefix(string(closureType), "object:Generator_") {
					retType = closureType
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpClosureCall,
					Type:   retType,
					Result: result,
					Callee: closureVal,
					Args:   args,
					Span:   toIRSpan(path, expression.Span),
				})
				return result, retType, nil
			}
		}
		return "", "", fmt.Errorf("unknown function %q", callee)
	}

	callee = target.Name
	args := make([]string, 0, len(expression.Arguments))
	for aIdx, argument := range expression.Arguments {
		value, valType, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if aIdx < len(target.Parameters) && target.Parameters[aIdx].Type == ir.TypeUnknown && valType != ir.TypeUnknown {
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
		if aIdx < len(target.Parameters) && target.Parameters[aIdx].Type == ir.TypeClosure && valType != ir.TypeClosure {
			closureSlot := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpClosure,
				Type:   ir.TypeClosure,
				Result: closureSlot,
				Callee: value,
				Args:   nil,
				Span:   toIRSpan(path, argument.Span),
			})
			value = closureSlot
		}
		args = append(args, value)
	}

	if restParamsIndex[callee] && len(target.Parameters) > 0 && (target.Parameters[len(target.Parameters)-1].Type == ir.TypeStringArray || target.Parameters[len(target.Parameters)-1].Type == ir.TypeNumberArray) {
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
		for i := len(args); i < len(target.Parameters); i++ {
			if defaults != nil {
				if initExpr, ok := defaults[i]; ok {
					if target.Parameters[i].Type == ir.TypeNumber && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
						numConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   ir.TypeNumber,
							Result: numConst,
							Value:  "0",
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
					}
					val, valType, err := lowerExpression(path, initExpr, "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
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
						closureSlot := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpClosure,
							Type:   ir.TypeClosure,
							Result: closureSlot,
							Callee: val,
							Args:   nil,
							Span:   toIRSpan(path, initExpr.Span),
						})
						val = closureSlot
					}
					args = append(args, val)
					continue
				}
			}
			if target.Parameters[i].Type == ir.TypeUnknown {
				undefConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeString,
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
					Value:  "0",
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
