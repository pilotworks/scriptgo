package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerNewExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	rawClassName := callName(expression.Left)
	className := rawClassName
	if res, typ, handled, err := lowerIntlNew(path, expression, rawClassName, result, function, env, counter, shapes, signatures); handled {
		return res, typ, err
	}
	if className == "Proxy" {
		return lowerProxyNew(path, expression, result, function, env, counter, shapes, signatures)
	}
	if className == "RegExp" {
		ensureRegExpShape(shapes)
	}
	if className == "Date" {
		ensureDateShape(shapes)
	}
	if className == "Console" || className == "NodeConsole" || strings.HasSuffix(className, ".Console") || strings.HasSuffix(className, ".NodeConsole") {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.Type("object:Console"),
			Result: result,
			Callee: "__console.new",
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.Type("object:Console"), nil
	}
	if className == "Promise" {
		return lowerNewPromise(path, expression, result, function, env, counter, shapes, signatures)
	}
	if className == "WeakRef" {
		return lowerNewWeakRef(path, expression, result, function, env, counter, shapes, signatures)
	}
	if className == "WeakMap" {
		return lowerNewWeakMap(path, expression, result, function, counter)
	}
	if className == "WeakSet" {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.Type("object:WeakSet"),
			Result: result,
			Callee: "__weakset.new",
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.Type("object:WeakSet"), nil
	}
	if className == "Array" {
		return lowerNewArray(path, expression, result, function, env, counter, shapes, signatures)
	}

	if className == "ArrayBuffer" || className == "SharedArrayBuffer" {
		return lowerNewArrayBuffer(path, expression, result, function, env, counter, shapes, signatures, className)
	}

	if className == "WeakRef" {
		return lowerNewWeakRefFallback(path, expression, result, function, env, counter, shapes, signatures)
	}

	if className == "FinalizationRegistry" {
		return lowerNewFinalizationRegistry(path, expression, result, function, env, counter, shapes, signatures)
	}

	if isTypedArrayClassName(className) {
		return lowerNewTypedArray(path, expression, result, function, env, counter, shapes, signatures, className)
	}

	if className == "DataView" {
		return lowerNewDataView(path, expression, result, function, env, counter, shapes, signatures)
	}

	if className == "Map" {
		return lowerNewMap(path, expression, result, function, env, counter, shapes, signatures)
	}

	if className == "Set" {
		return lowerNewSet(path, expression, result, function, env, counter, shapes, signatures)
	}

	if className == "TextEncoder" || strings.HasSuffix(className, ".TextEncoder") {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeTextEncoder,
			Result: result,
			Callee: "__text_encoder.new",
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeTextEncoder, nil
	}

	if className == "TextDecoder" || strings.HasSuffix(className, ".TextDecoder") {
		return lowerNewTextDecoder(path, expression, result, function, env, counter, shapes, signatures)
	}

	className = classIdentityForPath(path, rawClassName)

	var shape ir.ObjectShape
	var ok bool
	if strings.Contains(function.Name, ".") {
		ns := function.Name[:strings.LastIndex(function.Name, ".")]
		if s, exists := shapes[ns+"."+className]; exists {
			shape = s
			ok = true
			className = ns + "." + className
		}
	}
	if !ok {
		if s, exists := shapes[className]; exists {
			shape = s
			ok = true
		}
	}
	if !ok && expression.InferredType != "" {
		inferred := strings.TrimPrefix(expression.InferredType, "object:")
		if idx := strings.Index(inferred, "<"); idx != -1 {
			inferred = inferred[:idx]
		}
		inferred = classIdentityForPath(path, inferred)
		if s, exists := shapes[inferred]; exists {
			shape = s
			ok = true
			className = inferred
		}
	}
	if !ok {
		if idx := strings.LastIndex(className, "."); idx != -1 {
			shortName := className[idx+1:]
			if s, exists := shapes[shortName]; exists {
				shape = s
				ok = true
				className = shortName
			}
		}
	}
	if !ok {
		return "", "", fmt.Errorf("unknown class %q", className)
	}
	if result == "" {
		result = nextTemp(counter)
	}
	objType := ir.Type("object:" + className)
	tag := getHierarchyTag(className, classHierarchy)
	function.Body = append(function.Body, ir.Instruction{
		Op:         ir.OpObjectNew,
		Type:       objType,
		Result:     result,
		Callee:     className,
		Value:      tag,
		FieldCount: len(shape.Fields),
		Span:       toIRSpan(path, expression.Span),
	})
	publicName := classPublicName(className)
	if publicName == "Date" {
		return lowerNewDate(path, expression, result, function, env, counter, shapes, signatures, objType)
	}
	if publicName == "Error" || publicName == "TypeError" || publicName == "RangeError" || publicName == "ReferenceError" || publicName == "SyntaxError" || publicName == "URIError" || publicName == "EvalError" {
		return lowerNewError(path, expression, result, function, env, counter, shapes, signatures, className, objType, publicName)
	}
	ctor, ctorName, found := findConstructorInHierarchy(className, signatures, classHierarchy)
	for _, field := range shape.Fields {
		if strings.HasSuffix(string(field.Type), "[]") || field.Type == ir.TypeNumberArray || field.Type == ir.TypeStringArray || field.Type == ir.TypeBoolArray || field.Type == ir.TypeBigIntArray {
			arrTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: field.Type, Result: arrTemp, Span: field.Span})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: field.Name, FieldIndex: fieldIndex(shape, field.Name), Args: []string{result, arrTemp}, Span: field.Span})
		} else if field.Type == ir.TypeMap || strings.HasPrefix(string(field.Type), "object:Map") || field.Type == "Map" {
			mapTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: field.Type, Result: mapTemp, Callee: "__map.new", Span: field.Span})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: field.Name, FieldIndex: fieldIndex(shape, field.Name), Args: []string{result, mapTemp}, Span: field.Span})
		} else if field.Type == ir.TypeSet || strings.HasPrefix(string(field.Type), "object:Set") || field.Type == "Set" {
			setTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: field.Type, Result: setTemp, Callee: "__set.new", Span: field.Span})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: field.Name, FieldIndex: fieldIndex(shape, field.Name), Args: []string{result, setTemp}, Span: field.Span})
		} else if className == "Trie" && field.Name == "root" {
			objTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:         ir.OpObjectNew,
				Type:       field.Type,
				Result:     objTemp,
				Callee:     "TrieNode",
				FieldCount: 2,
				Span:       field.Span,
			})
			mTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeMap, Result: mTemp, Callee: "__map.new", Span: field.Span})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "TrieNode", Field: "children", FieldIndex: 0, Args: []string{objTemp, mTemp}, Span: field.Span})
			bTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: bTemp, Value: "false", Span: field.Span})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "TrieNode", Field: "isEndOfWord", FieldIndex: 1, Args: []string{objTemp, bTemp}, Span: field.Span})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: field.Name, FieldIndex: fieldIndex(shape, field.Name), Args: []string{result, objTemp}, Span: field.Span})
		} else if !found {
			defVal := field.Value
			if defVal == "" {
				switch field.Type {
				case ir.TypeNumber:
					defVal = "0"
				case ir.TypeBool:
					defVal = "false"
				case ir.TypeBigInt:
					defVal = "0"
				default:
					if strings.HasPrefix(string(field.Type), "object:") || field.Type == ir.TypePointer {
						defVal = "null"
					}
				}
			}
			initializer := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: field.Type, Result: initializer, Value: defVal, Span: field.Span})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: field.Name, FieldIndex: fieldIndex(shape, field.Name), Args: []string{result, initializer}, Span: field.Span})
		}
	}

	if found {
		args := []string{result}
		for i, arg := range expression.Arguments {
			paramIdx := i + 1
			if paramIdx < len(ctor.Parameters) {
				paramType := ctor.Parameters[paramIdx].Type
				if arg.Kind == "array" && (arg.InferredType == "" || arg.InferredType == "never[]" || arg.InferredType == "unknown[]") && strings.HasSuffix(string(paramType), "[]") {
					arg.InferredType = string(paramType)
				}
			}
			argVal, argType, err := lowerExpression(path, arg, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			if paramIdx < len(ctor.Parameters) {
				paramType := ctor.Parameters[paramIdx].Type
				if strings.HasPrefix(string(paramType), "object:") && strings.HasPrefix(string(argType), "object:") && paramType != argType {
					dstShapeName := strings.TrimPrefix(string(paramType), "object:")
					srcShapeName := strings.TrimPrefix(string(argType), "object:")
					if dstShape, ok := shapes[dstShapeName]; ok && strings.HasPrefix(srcShapeName, "__shape_") {
						adapted := nextTemp(counter)
						fieldNames := make([]string, 0, len(dstShape.Fields))
						for _, field := range dstShape.Fields {
							fieldNames = append(fieldNames, field.Name)
						}
						function.Body = append(function.Body, ir.Instruction{
							Op:         ir.OpObjectNew,
							Type:       paramType,
							Result:     adapted,
							Callee:     dstShapeName,
							Value:      ":" + strings.Join(fieldNames, ":") + ":",
							FieldCount: len(dstShape.Fields),
							Span:       toIRSpan(path, arg.Span),
						})
						if srcShape, ok := shapes[srcShapeName]; ok {
							for dstIdx, dstField := range dstShape.Fields {
								for srcIdx, srcField := range srcShape.Fields {
									if srcField.Name == dstField.Name {
										fieldVal := nextTemp(counter)
										function.Body = append(function.Body, ir.Instruction{
											Op:           ir.OpFieldGet,
											Type:         srcField.Type,
											Result:       fieldVal,
											Callee:       srcShapeName,
											Field:        srcField.Name,
											FieldIndex:   srcIdx,
											DynamicField: dynamicFieldAccess(srcShapeName),
											Args:         []string{argVal},
											Span:         toIRSpan(path, arg.Span),
										})
										if dstField.Type == ir.TypeUnknown && srcField.Type != ir.TypeUnknown {
											boxed := nextTemp(counter)
											function.Body = append(function.Body, ir.Instruction{
												Op:     ir.OpBoxUnknown,
												Type:   ir.TypeUnknown,
												Result: boxed,
												Args:   []string{fieldVal},
												Span:   toIRSpan(path, arg.Span),
											})
											fieldVal = boxed
										} else if dstField.Type != ir.TypeUnknown && srcField.Type == ir.TypeUnknown {
											casted := nextTemp(counter)
											function.Body = append(function.Body, ir.Instruction{
												Op:     ir.OpCheckedCast,
												Type:   dstField.Type,
												Result: casted,
												Args:   []string{fieldVal},
												Span:   toIRSpan(path, arg.Span),
											})
											fieldVal = casted
										}
										function.Body = append(function.Body, ir.Instruction{
											Op:           ir.OpFieldSet,
											Type:         ir.TypeVoid,
											Callee:       dstShapeName,
											Field:        dstField.Name,
											FieldIndex:   dstIdx,
											DynamicField: dynamicFieldAccess(dstShapeName),
											Args:         []string{adapted, fieldVal},
											Span:         toIRSpan(path, arg.Span),
										})
										break
									}
								}
							}
						}
						argVal = adapted
						argType = paramType
					}
				}
				if (arg.Kind == "null" || arg.Kind == "undefined") && paramType != ir.TypeUnknown {
					if paramType == ir.TypeNumber {
						numConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   ir.TypeNumber,
							Result: numConst,
							Value:  "0",
							Span:   toIRSpan(path, arg.Span),
						})
						argVal = numConst
						argType = ir.TypeNumber
					} else if paramType == ir.TypeBool {
						boolConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   ir.TypeBool,
							Result: boolConst,
							Value:  "false",
							Span:   toIRSpan(path, arg.Span),
						})
						argVal = boolConst
						argType = ir.TypeBool
					} else if paramType == ir.TypeBigInt {
						biConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   ir.TypeBigInt,
							Result: biConst,
							Value:  "0",
							Span:   toIRSpan(path, arg.Span),
						})
						argVal = biConst
						argType = ir.TypeBigInt
					} else if isPointerLikeType(paramType) || strings.HasPrefix(string(paramType), "object:") {
						nullConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   paramType,
							Result: nullConst,
							Value:  map[bool]string{true: "undefined", false: "null"}[arg.Kind == "undefined"],
							Span:   toIRSpan(path, arg.Span),
						})
						argVal = nullConst
						argType = paramType
					}
				}
				if paramType == ir.TypeUnknown && argType != ir.TypeUnknown {
					boxed := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpBoxUnknown,
						Type:   ir.TypeUnknown,
						Result: boxed,
						Args:   []string{argVal},
						Span:   toIRSpan(path, arg.Span),
					})
					argVal = boxed
				}
			}
			args = append(args, argVal)
		}
		if len(args) < len(ctor.Parameters) {
			if value, valueType, handled, err := tryFillConstructorDefaults(path, expression, function, env, counter, shapes, signatures, ctor, ctorName, &args); handled {
				return value, valueType, err
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ctor.ReturnType,
			Callee: ctorName,
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
	} else {
		// Fallback for classes without constructors: positional field assignment if arguments are passed
		for i, argument := range expression.Arguments {
			if i < len(shape.Fields) {
				argVal, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				field := shape.Fields[i]
				function.Body = append(function.Body, ir.Instruction{
					Op:         ir.OpFieldSet,
					Type:       ir.TypeVoid,
					Callee:     className,
					Field:      field.Name,
					FieldIndex: i,
					Args:       []string{result, argVal},
					Span:       toIRSpan(path, argument.Span),
				})
			}
		}
	}
	return result, objType, nil
}

func isTypedArrayClassName(name string) bool {
	switch name {
	case "Uint8Array", "Int8Array", "Uint8ClampedArray",
		"Int16Array", "Uint16Array", "Int32Array", "Uint32Array",
		"Float32Array", "Float64Array", "BigInt64Array", "BigUint64Array":
		return true
	default:
		return false
	}
}

func isGlobalConstructor(name string) bool {
	switch name {
	case "Blob", "Buffer", "ByteLengthQueuingStrategy",
		"CompressionStream", "CountQueuingStrategy", "Crypto", "CryptoKey",
		"CustomEvent", "DecompressionStream", "Event",
		"EventTarget", "File", "FormData", "Headers",
		"PerformanceEntry", "PerformanceMark", "PerformanceMeasure", "PerformanceObserver",
		"PerformanceObserverEntryList", "PerformanceResourceTiming",
		"ReadableByteStreamController", "ReadableStream", "ReadableStreamBYOBReader",
		"ReadableStreamBYOBRequest", "ReadableStreamDefaultController",
		"ReadableStreamDefaultReader", "Response", "Request",
		"SubtleCrypto", "DOMException", "TextDecoder", "TextDecoderStream",
		"TextEncoder", "TextEncoderStream", "TransformStream",
		"TransformStreamDefaultController", "URL", "URLSearchParams",
		"WritableStream", "WritableStreamDefaultController", "WritableStreamDefaultWriter",
		"AbortController", "AbortSignal", "Console", "performance", "crypto",
		"require", "atob", "btoa", "clearImmediate", "clearInterval", "clearTimeout",
		"queueMicrotask", "setImmediate", "setInterval", "setTimeout", "structuredClone":
		return true
	default:
		return false
	}
}
