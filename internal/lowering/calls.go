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
					if _, ok := signatures[expression.Left.Text]; ok {
						isModuleNamespace = true
					}
				}
			}
		}
		if !isModuleNamespace {
			methodName := expression.Left.Text
			if (expression.Left.Kind == "index" || expression.Left.Kind == "optional_index") && expression.Left.Right != nil {
				if expression.Left.Right.Kind == "property" && expression.Left.Right.Left != nil && expression.Left.Right.Left.Text == "Symbol" {
					methodName = "Symbol." + expression.Left.Right.Text
				} else if expression.Left.Right.Kind == "string" {
					methodName = expression.Left.Right.Text
				} else if expression.Left.Right.Kind == "identifier" {
					methodName = expression.Left.Right.Text
				}
			}
			receiver, receiverType, err := lowerExpression(path, expression.Left.Left, "", function, env, counter, shapes, signatures)
			if err == nil {
				if methodName == "then" || methodName == "catch" {
					args := []string{receiver}
					for _, arg := range expression.Arguments {
						v, _, err := lowerExpression(path, arg, "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						args = append(args, v)
					}
					if result == "" {
						result = nextTemp(counter)
					}
					callee := "__async.promise_then"
					if methodName == "catch" {
						callee = "__async.promise_catch"
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.Type("object:Promise"),
						Result: result,
						Callee: callee,
						Args:   args,
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.Type("object:Promise"), nil
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
				if receiverType == ir.TypeBigInt && (methodName == "toString" || methodName == "toLocaleString") {
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__string.fromBigInt", Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
					return result, ir.TypeString, nil
				}
				if receiverType == ir.TypeBigInt && methodName == "valueOf" {
					return receiver, ir.TypeBigInt, nil
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
				if res, typ, handled, err := lowerIteratorReceiverMethod(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
					return res, typ, err
				}
				className := strings.TrimPrefix(string(receiverType), "object:")
				if className != "" && className != "number" && className != "string" && className != "bool" && className != "void" {
					if target, mangled, ok := findMethodInHierarchy(className, methodName, signatures, classHierarchy); ok {
						args := []string{receiver}
						for aIdx, argument := range expression.Arguments {
							pIdx := aIdx + 1
							if argument.Kind == "array" && (argument.InferredType == "" || argument.InferredType == "never[]" || argument.InferredType == "unknown[]") && pIdx < len(target.Parameters) && strings.HasSuffix(string(target.Parameters[pIdx].Type), "[]") {
								argument.InferredType = string(target.Parameters[pIdx].Type)
							}
							argVal, valType, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
							if err != nil {
								return "", "", err
							}
							if pIdx < len(target.Parameters) {
								argVal, valType, _ = adaptStructuralObjectArgument(path, toIRSpan(path, argument.Span), argVal, valType, target.Parameters[pIdx].Type, function, counter, shapes)
							}
							hasRest := (restParamsIndex[mangled] || restParamsIndex[strings.Split(mangled, "__")[0]]) && len(target.Parameters) > 0
							restIsUnknown := hasRest && target.Parameters[len(target.Parameters)-1].Type == ir.TypeUnknownArray
							fixed := len(target.Parameters) - 1
							needsUnknownBox := (pIdx < len(target.Parameters) && target.Parameters[pIdx].Type == ir.TypeUnknown) || (restIsUnknown && pIdx >= fixed)
							if needsUnknownBox && valType != ir.TypeUnknown {
								boxed := nextTemp(counter)
								function.Body = append(function.Body, ir.Instruction{
									Op:     ir.OpBoxUnknown,
									Type:   ir.TypeUnknown,
									Result: boxed,
									Args:   []string{argVal},
									Span:   toIRSpan(path, argument.Span),
								})
								argVal = boxed
							} else if pIdx < len(target.Parameters) && isPointerLikeType(target.Parameters[pIdx].Type) && (argument.Kind == "null" || argument.Kind == "undefined") {
								nullConst := nextTemp(counter)
								function.Body = append(function.Body, ir.Instruction{
									Op:     ir.OpConst,
									Type:   target.Parameters[pIdx].Type,
									Result: nullConst,
									Value:  "null",
									Span:   toIRSpan(path, argument.Span),
								})
								argVal = nullConst
							}
							args = append(args, argVal)
						}
						if (restParamsIndex[mangled] || restParamsIndex[strings.Split(mangled, "__")[0]]) && len(target.Parameters) > 0 && (strings.HasSuffix(string(target.Parameters[len(target.Parameters)-1].Type), "[]") || target.Parameters[len(target.Parameters)-1].Type == ir.TypeStringArray || target.Parameters[len(target.Parameters)-1].Type == ir.TypeNumberArray) {
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
							defaults := defaultParamsIndex[mangled]
							if defaults == nil {
								defaults = defaultParamsIndex[strings.Split(mangled, "__")[0]]
							}
							for i := len(args); i < len(target.Parameters); i++ {
								var val string
								var valType ir.Type
								paramType := target.Parameters[i].Type
								if defaults != nil && defaults[i] != nil {
									initExpr := defaults[i]
									if paramType == ir.TypeNumber && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
										numConst := nextTemp(counter)
										function.Body = append(function.Body, ir.Instruction{
											Op:     ir.OpConst,
											Type:   ir.TypeNumber,
											Result: numConst,
											Value:  "0",
											Span:   toIRSpan(path, initExpr.Span),
										})
										val = numConst
										valType = ir.TypeNumber
									} else if paramType == ir.TypeBool && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
										boolConst := nextTemp(counter)
										function.Body = append(function.Body, ir.Instruction{
											Op:     ir.OpConst,
											Type:   ir.TypeBool,
											Result: boolConst,
											Value:  "false",
											Span:   toIRSpan(path, initExpr.Span),
										})
										val = boolConst
										valType = ir.TypeBool
									} else if paramType == ir.TypeBigInt && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
										biConst := nextTemp(counter)
										function.Body = append(function.Body, ir.Instruction{
											Op:     ir.OpConst,
											Type:   ir.TypeBigInt,
											Result: biConst,
											Value:  "0",
											Span:   toIRSpan(path, initExpr.Span),
										})
										val = biConst
										valType = ir.TypeBigInt
									} else if (strings.HasPrefix(string(paramType), "object:") || isPointerLikeType(paramType)) && (initExpr.Kind == "null" || initExpr.Kind == "undefined") {
										valStr := "null"
										if initExpr.Kind == "undefined" {
											valStr = "undefined"
										}
										nullConst := nextTemp(counter)
										function.Body = append(function.Body, ir.Instruction{
											Op:     ir.OpConst,
											Type:   paramType,
											Result: nullConst,
											Value:  valStr,
											Span:   toIRSpan(path, initExpr.Span),
										})
										val = nullConst
										valType = paramType
									} else {
										v, vt, err := lowerExpression(path, initExpr, "", function, env, counter, shapes, signatures)
										if err == nil {
											val = v
											valType = vt
										}
									}
								}
								if val == "" {
									val = nextTemp(counter)
									valType = paramType
									defStr := "0"
									if paramType == ir.TypeBool {
										defStr = "false"
									} else if paramType == ir.TypeString {
										defStr = ""
									} else if isPointerLikeType(paramType) || strings.HasPrefix(string(paramType), "object:") {
										defStr = "null"
									}
									function.Body = append(function.Body, ir.Instruction{
										Op:     ir.OpConst,
										Type:   paramType,
										Result: val,
										Value:  defStr,
										Span:   toIRSpan(path, expression.Span),
									})
								}
								if paramType == ir.TypeUnknown && valType != ir.TypeUnknown {
									boxed := nextTemp(counter)
									function.Body = append(function.Body, ir.Instruction{
										Op:     ir.OpBoxUnknown,
										Type:   ir.TypeUnknown,
										Result: boxed,
										Args:   []string{val},
										Span:   toIRSpan(path, expression.Span),
									})
									val = boxed
								}
								args = append(args, val)
							}
						}
						if result == "" {
							result = nextTemp(counter)
						}
						retType := target.ReturnType
						if retType == "" {
							retType = ir.TypeVoid
						}
						// A single IR function represents an overloaded implementation,
						// so its ABI may be `unknown` even when TypeScript selected a
						// concrete overload for this call. Keep the call ABI intact and
						// narrow the boxed result at the call boundary.
						selectedType := ir.Type("")
						if expression.InferredType != "" && expression.InferredType != "this" && expression.InferredType != "object:this" {
							inferred := toIRType(expression.InferredType)
							if inferred != "" && inferred != ir.TypeUnknown && inferred != retType {
								selectedType = inferred
							}
						}
						callResult := result
						if selectedType != "" {
							callResult = nextTemp(counter)
						}
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpCall,
							Type:   retType,
							Result: callResult,
							Callee: mangled,
							Args:   args,
							Span:   toIRSpan(path, expression.Span),
						})
						if selectedType != "" {
							if result == "" {
								result = nextTemp(counter)
							}
							function.Body = append(function.Body, ir.Instruction{
								Op:     ir.OpCheckedCast,
								Type:   selectedType,
								Result: result,
								Args:   []string{callResult},
								Span:   toIRSpan(path, expression.Span),
							})
							return result, selectedType, nil
						}
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
					propVal, _, err := lowerPropertyExpression(path, &typescriptgo.SyntaxExpression{
						Span:         expression.Left.Span,
						Kind:         "property",
						Left:         expression.Left.Left,
						Text:         methodName,
						InferredType: "",
					}, "", function, env, counter, shapes, signatures)
					if err == nil {
						args := make([]string, 0, len(expression.Arguments))
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
						retType := ir.TypeNumber
						if expression.InferredType != "" {
							retType = toIRType(expression.InferredType)
						}
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpClosureCall,
							Type:   retType,
							Result: result,
							Callee: propVal,
							Args:   args,
							Span:   toIRSpan(path, expression.Span),
						})
						return result, retType, nil
					}
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
		if meta.Extends == "Error" || meta.Extends == "TypeError" || meta.Extends == "RangeError" || meta.Extends == "SyntaxError" || meta.Extends == "ReferenceError" || meta.Extends == "URIError" || meta.Extends == "EvalError" || meta.Extends == "DOMException" {
			if len(expression.Arguments) > 0 {
				msgVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				function.Body = append(function.Body, ir.Instruction{
					Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: currentClass, Field: "message", FieldIndex: 0, Args: []string{"this", msgVal}, Span: toIRSpan(path, expression.Span),
				})
			}
			return "", ir.TypeVoid, nil
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
		for i := len(args); i < len(ctor.Parameters); i++ {
			paramType := ctor.Parameters[i].Type
			defConst := nextTemp(counter)
			defVal := "0"
			if paramType == ir.TypeBool {
				defVal = "false"
			} else if paramType == ir.TypeString {
				defVal = ""
			} else if isPointerLikeType(paramType) || strings.HasPrefix(string(paramType), "object:") {
				defVal = "null"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   paramType,
				Result: defConst,
				Value:  defVal,
				Span:   toIRSpan(path, expression.Span),
			})
			args = append(args, defConst)
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
				args := make([]string, 0, len(expression.Arguments))
				for aIdx, argument := range expression.Arguments {
					val, valType, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					hasRest := (restParamsIndex[mangled] || restParamsIndex[strings.Split(mangled, "__")[0]]) && len(target.Parameters) > 0
					restIsUnknown := hasRest && target.Parameters[len(target.Parameters)-1].Type == ir.TypeUnknownArray
					fixed := len(target.Parameters) - 1
					needsUnknownBox := (aIdx < len(target.Parameters) && target.Parameters[aIdx].Type == ir.TypeUnknown) || (restIsUnknown && aIdx >= fixed)
					if needsUnknownBox && valType != ir.TypeUnknown {
						boxed := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpBoxUnknown,
							Type:   ir.TypeUnknown,
							Result: boxed,
							Args:   []string{val},
							Span:   toIRSpan(path, argument.Span),
						})
						val = boxed
					} else if aIdx < len(target.Parameters) && isPointerLikeType(target.Parameters[aIdx].Type) && (argument.Kind == "null" || argument.Kind == "undefined") {
						valStr := "null"
						if argument.Kind == "undefined" {
							valStr = "undefined"
						}
						nullConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   target.Parameters[aIdx].Type,
							Result: nullConst,
							Value:  valStr,
							Span:   toIRSpan(path, argument.Span),
						})
						val = nullConst
					}
					args = append(args, val)
				}
				if len(args) < len(target.Parameters) {
					defaults := defaultParamsIndex[mangled]
					if defaults == nil {
						defaults = defaultParamsIndex[strings.Split(mangled, "__")[0]]
					}
					if defaults != nil {
						for i := len(args); i < len(target.Parameters); i++ {
							if initExpr, ok := defaults[i]; ok {
								paramType := target.Parameters[i].Type
								if paramType == ir.TypeNumber && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
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
								} else if paramType == ir.TypeBool && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
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
								} else if paramType == ir.TypeBigInt && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
									biConst := nextTemp(counter)
									function.Body = append(function.Body, ir.Instruction{
										Op:     ir.OpConst,
										Type:   ir.TypeBigInt,
										Result: biConst,
										Value:  "0",
										Span:   toIRSpan(path, initExpr.Span),
									})
									args = append(args, biConst)
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
								} else if i < len(target.Parameters) && (isPointerLikeType(target.Parameters[i].Type) || strings.HasPrefix(string(target.Parameters[i].Type), "object:")) && (initExpr.Kind == "null" || initExpr.Kind == "undefined") {
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
									val = nullConst
								}
								args = append(args, val)
							}
						}
					}
				}
				if (restParamsIndex[mangled] || restParamsIndex[strings.Split(mangled, "__")[0]]) && len(target.Parameters) > 0 && strings.HasSuffix(string(target.Parameters[len(target.Parameters)-1].Type), "[]") {
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

	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "member" || expression.Left.Kind == "index" || expression.Left.Kind == "call" || expression.Left.Kind == "paren" || expression.Left.Kind == "optional_call") {
		isModuleFunc := false
		if (expression.Left.Kind == "property" || expression.Left.Kind == "optional_property") && expression.Left.Left != nil && expression.Left.Left.Kind == "identifier" {
			if _, inEnv := env[expression.Left.Left.Text]; !inEnv {
				if _, inTop := topLevelVars[expression.Left.Left.Text]; !inTop {
					if _, ok := signatures[expression.Left.Text]; ok {
						isModuleFunc = true
					}
				}
			}
		}
		if !isModuleFunc {
			closureVal, closureType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
			if err == nil && (closureType == ir.TypeClosure || closureType == "Function" || closureType == "function" || strings.Contains(string(closureType), "=>")) {
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
				} else if expression.Left.InferredType != "" && strings.Contains(expression.Left.InferredType, "=>") {
					retStr := extractTopLevelReturnType(expression.Left.InferredType)
					if parsed := toIRType(retStr); parsed != "" {
						retType = parsed
					}
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

	calleeIsClosure := false
	calleeType, hasCalleeType := env[callee]
	if hasCalleeType && (calleeType == ir.TypeClosure || calleeType == ir.TypeUnknown || calleeType == ir.TypeObject || calleeType == "Function" || calleeType == "function" || strings.Contains(string(calleeType), "=>")) {
		calleeIsClosure = true
	} else if topVar, isTop := topLevelVars[callee]; isTop && ((topVar.Expression != nil && (topVar.Expression.Kind == "arrow_function" || topVar.Expression.Kind == "function")) || strings.Contains(topVar.Type, "=>") || strings.Contains(topVar.InferredType, "=>")) {
		calleeIsClosure = true
	}

	if calleeIsClosure {
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
		} else if target, ok := resolveFunctionSignature(path, callee, signatures); ok && target.ReturnType != "" {
			retType = target.ReturnType
		} else if expression.InferredType != "" {
			retType = toIRType(expression.InferredType)
		} else if topVar, ok := topLevelVars[callee]; ok {
			if isReturningClosure(topVar) {
				retType = ir.TypeClosure
			} else if topVar.Type != "" {
				t := toIRType(topVar.Type)
				if t == ir.TypeClosure || strings.HasPrefix(string(t), "object:") {
					retType = t
				}
			}
		}
		calleeName := callee
		if targetName, ok := env[callee+".closureTarget"]; ok && targetName != "" {
			calleeName = string(targetName)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpClosureCall,
			Type:   retType,
			Result: result,
			Callee: calleeName,
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, nil
	}

	if intrinsic, ok := builtinIntrinsic(callee); ok {
		return intrinsic.Lower(IntrinsicCall{Path: path, Expression: expression, Result: result, Function: function, Env: env, Counter: counter, Shapes: shapes, Signatures: signatures, LowerExpression: lowerExpression}, intrinsic)
	}

	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "member" || expression.Left.Kind == "index" || expression.Left.Kind == "optional_index" || expression.Left.Kind == "optional_property") {
		isModuleFunc := false
		if (expression.Left.Kind == "property" || expression.Left.Kind == "optional_property") && expression.Left.Left != nil && expression.Left.Left.Kind == "identifier" {
			if _, inEnv := env[expression.Left.Left.Text]; !inEnv {
				if _, inTop := topLevelVars[expression.Left.Left.Text]; !inTop {
					if _, ok := signatures[expression.Left.Text]; ok {
						isModuleFunc = true
					}
				}
			}
		}
		if !isModuleFunc {
			closureVal, closureType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
			if err == nil && (closureType == ir.TypeClosure || closureType == ir.TypeUnknown || strings.HasPrefix(string(closureType), "object:") || strings.Contains(string(closureType), "=>")) {
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
				if expression.InferredType != "" {
					retType = toIRType(expression.InferredType)
				} else if strings.Contains(string(closureType), "=>") {
					retStr := extractTopLevelReturnType(string(closureType))
					retType = toIRType(retStr)
				} else if rt, ok := env[closureVal+".retType"]; ok && rt != "" {
					retType = rt
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
		if strings.HasPrefix(callee, "fs.promises.") {
			method := strings.TrimPrefix(callee, "fs.promises.")
			if sig, ok2 := signatures["FSPromises."+method]; ok2 {
				target = sig
				ok = true
			} else if sig, ok2 := signatures["FSPromises_"+method]; ok2 {
				target = sig
				ok = true
			}
		} else if strings.HasPrefix(callee, "promises.") {
			method := strings.TrimPrefix(callee, "promises.")
			if sig, ok2 := signatures["FSPromises."+method]; ok2 {
				target = sig
				ok = true
			} else if sig, ok2 := signatures["FSPromises_"+method]; ok2 {
				target = sig
				ok = true
			}
		} else if strings.Contains(callee, ".") {
			parts := strings.Split(callee, ".")
			funcName := parts[len(parts)-1]
			if expression.Left != nil && expression.Left.Left != nil {
				recvType := env[expression.Left.Left.Text]
				if recvType == "" && expression.Left.Left.InferredType != "" {
					recvType = toIRType(expression.Left.Left.InferredType)
				}
				if strings.HasPrefix(string(recvType), "object:") {
					cls := strings.TrimPrefix(string(recvType), "object:")
					if idx := strings.Index(cls, "<"); idx != -1 {
						cls = cls[:idx]
					}
					mangled := cls + "_" + funcName
					if sig, ok2 := signatures[mangled]; ok2 {
						target = sig
						ok = true
					}
				}
			}
			if !ok {
				if sig, ok2 := signatures[funcName]; ok2 {
					target = sig
					ok = true
				}
			}
		}
	}
	if !ok {
		return "", "", fmt.Errorf("unknown function %q", callee)
	}

	callee = target.Name
	args := make([]string, 0, len(expression.Arguments))
	paramOffset := 0
	if len(target.Parameters) > 0 && target.Parameters[0].Name == "this" {
		dummyThis := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   target.Parameters[0].Type,
			Result: dummyThis,
			Value:  "null",
			Span:   toIRSpan(path, expression.Span),
		})
		args = append(args, dummyThis)
		paramOffset = 1
	}
	for aIdx, argument := range expression.Arguments {
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
					env[args[j]] = target.Parameters[j].Type
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
		needsUnknownBox := (pIdx < len(target.Parameters) && target.Parameters[pIdx].Type == ir.TypeUnknown) || (restIsUnknown && pIdx >= fixed)
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
			if sig, isSig := signatures[value]; isSig {
				closureSlot := nextTemp(counter)
				calleeName := ensureFunctionClosureTrampoline(path, sig, signatures)
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

	if (restParamsIndex[callee] || restParamsIndex[strings.Split(callee, "__")[0]]) && len(target.Parameters) > 0 && (strings.HasSuffix(string(target.Parameters[len(target.Parameters)-1].Type), "[]") || target.Parameters[len(target.Parameters)-1].Type == ir.TypeStringArray || target.Parameters[len(target.Parameters)-1].Type == ir.TypeNumberArray) {
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
				env[args[j]] = target.Parameters[j].Type
			}
		}
		for i := len(args); i < len(target.Parameters); i++ {
			if defaults != nil {
				if initExpr, ok := defaults[i]; ok {
					initExpr = substituteParamIdentifiers(initExpr, paramMap)
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
						if sig, isSig := signatures[val]; isSig {
							closureSlot := nextTemp(counter)
							calleeName := ensureFunctionClosureTrampoline(path, sig, signatures)
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
					pName := target.Parameters[i].Name
					if pName != "" && pName != "this" {
						paramMap[pName] = val
						env[val] = target.Parameters[i].Type
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

func substituteParamIdentifiers(expr *typescriptgo.SyntaxExpression, paramMap map[string]string) *typescriptgo.SyntaxExpression {
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
		newArgs := make([]*typescriptgo.SyntaxExpression, len(copy.Arguments))
		for k, arg := range copy.Arguments {
			newArgs[k] = substituteParamIdentifiers(arg, paramMap)
		}
		copy.Arguments = newArgs
	}
	return &copy
}
