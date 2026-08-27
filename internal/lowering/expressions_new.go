package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerNewExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	className := callName(expression.Left)
	if res, typ, handled, err := lowerIntlNew(path, expression, className, result, function, env, counter, shapes, signatures); handled {
		return res, typ, err
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
		if result == "" {
			result = nextTemp(counter)
		}
		promType := ir.Type("object:Promise<unknown>")
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   promType,
			Result: result,
			Callee: "__async.promise_create",
			Args:   []string{},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, promType, nil
	}
	if className == "WeakRef" {
		var targetArg string
		var targetType ir.Type = ir.TypeObject
		if len(expression.Arguments) > 0 {
			v, t, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			targetArg = v
			if t != "" {
				targetType = t
			}
		}
		if result == "" {
			result = nextTemp(counter)
		}
		resType := ir.Type("object:WeakRef<" + string(targetType) + ">")
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   resType,
			Result: result,
			Callee: "__weakref.new",
			Args:   []string{targetArg},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, resType, nil
	}
	if className == "WeakMap" {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.Type("object:WeakMap"),
			Result: result,
			Callee: "__weakmap.new",
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.Type("object:WeakMap"), nil
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
		if len(expression.Arguments) == 1 {
			lenVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			retType := ir.TypeNumberArray
			if expression.InferredType != "" {
				inferred := toIRType(expression.InferredType)
				if strings.HasSuffix(string(inferred), "[]") {
					retType = inferred
				}
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   retType,
				Result: result,
				Callee: "__array.new_length",
				Args:   []string{lenVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, retType, nil
		}
		var args []string
		elemType := ir.TypeNumber
		for _, argExpr := range expression.Arguments {
			argVal, aType, err := lowerExpression(path, argExpr, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			args = append(args, argVal)
			if aType != "" {
				elemType = aType
			}
		}
		retType := ir.Type(string(elemType) + "[]")
		if elemType == ir.TypeNumber {
			retType = ir.TypeNumberArray
		} else if elemType == ir.TypeString {
			retType = ir.TypeStringArray
		} else if elemType == ir.TypeBool {
			retType = ir.TypeBoolArray
		} else if elemType == ir.TypeBigInt {
			retType = ir.TypeBigIntArray
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpArray,
			Type:   retType,
			Result: result,
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, nil
	}

	if className == "ArrayBuffer" || className == "SharedArrayBuffer" {
		byteLenVal := ""
		if len(expression.Arguments) > 0 {
			v, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			byteLenVal = v
		} else {
			zeroConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, expression.Span),
			})
			byteLenVal = zeroConst
		}
		if result == "" {
			result = nextTemp(counter)
		}
		callee := "__arraybuffer.new"
		if className == "SharedArrayBuffer" {
			callee = "__atomics.sharedArrayBufferNew"
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeArrayBuffer,
			Result: result,
			Callee: callee,
			Args:   []string{byteLenVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeArrayBuffer, nil
	}

	if className == "WeakRef" {
		targetVal := "null"
		if len(expression.Arguments) > 0 {
			v, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			targetVal = v
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeObject,
			Result: result,
			Callee: "__weakref.new",
			Args:   []string{targetVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeObject, nil
	}

	if className == "FinalizationRegistry" {
		cbVal := "null"
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
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.Type("object:FinalizationRegistry"),
			Result: result,
			Callee: "__finalization_registry.new",
			Args:   []string{cbVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.Type("object:FinalizationRegistry"), nil
	}

	if isTypedArrayClassName(className) {
		targetType := ir.Type(className)
		if len(expression.Arguments) == 0 {
			zeroConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, expression.Span),
			})
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   targetType,
				Result: result,
				Callee: "__typedarray.new_length",
				Value:  className,
				Args:   []string{zeroConst},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, targetType, nil
		}
		arg0Val, arg0Type, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if result == "" {
			result = nextTemp(counter)
		}
		if arg0Type == ir.TypeNumber {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   targetType,
				Result: result,
				Callee: "__typedarray.new_length",
				Value:  className,
				Args:   []string{arg0Val},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, targetType, nil
		}
		if arg0Type == ir.TypeArrayBuffer {
			byteOffsetVal := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeNumber, Result: byteOffsetVal, Value: "0", Span: toIRSpan(path, expression.Span),
			})
			lengthVal := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeNumber, Result: lengthVal, Value: "0", Span: toIRSpan(path, expression.Span),
			})
			if len(expression.Arguments) > 1 {
				bo, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				byteOffsetVal = bo
			}
			if len(expression.Arguments) > 2 {
				l, _, err := lowerExpression(path, expression.Arguments[2], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				lengthVal = l
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   targetType,
				Result: result,
				Callee: "__typedarray.new_buffer",
				Value:  className,
				Args:   []string{arg0Val, byteOffsetVal, lengthVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, targetType, nil
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   targetType,
			Result: result,
			Callee: "__typedarray.new_array",
			Value:  className,
			Args:   []string{arg0Val},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, targetType, nil
	}

	if className == "DataView" {
		if len(expression.Arguments) == 0 {
			return "", "", fmt.Errorf("DataView constructor requires at least 1 argument")
		}
		bufVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		byteOffsetVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeNumber, Result: byteOffsetVal, Value: "0", Span: toIRSpan(path, expression.Span),
		})
		byteLenVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeNumber, Result: byteLenVal, Value: "0", Span: toIRSpan(path, expression.Span),
		})
		if len(expression.Arguments) > 1 {
			bo, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			byteOffsetVal = bo
		}
		if len(expression.Arguments) > 2 {
			bl, _, err := lowerExpression(path, expression.Arguments[2], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			byteLenVal = bl
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeDataView,
			Result: result,
			Callee: "__dataview.new",
			Args:   []string{bufVal, byteOffsetVal, byteLenVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeDataView, nil
	}

	if className == "Map" {
		if result == "" {
			result = nextTemp(counter)
		}
		if len(expression.Arguments) == 0 {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeMap,
				Result: result,
				Callee: "__map.new",
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeMap, nil
		}
		arg0Val, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeMap,
			Result: result,
			Callee: "__map.new_entries",
			Args:   []string{arg0Val},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeMap, nil
	}

	if className == "Set" {
		if result == "" {
			result = nextTemp(counter)
		}
		if len(expression.Arguments) == 0 {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeSet,
				Result: result,
				Callee: "__set.new",
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeSet, nil
		}
		arg0Val, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeSet,
			Result: result,
			Callee: "__set.new_values",
			Args:   []string{arg0Val},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeSet, nil
	}

	if className == "TextEncoder" {
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

	if className == "TextDecoder" {
		if result == "" {
			result = nextTemp(counter)
		}
		var args []string
		if len(expression.Arguments) > 0 {
			labelVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			args = append(args, labelVal)
			if len(expression.Arguments) > 1 {
				optsVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				args = append(args, optsVal)
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeTextDecoder,
			Result: result,
			Callee: "__text_decoder.new",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeTextDecoder, nil
	}

	shape, ok := shapes[className]
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
	if !ok && expression.InferredType != "" {
		inferred := strings.TrimPrefix(expression.InferredType, "object:")
		if s, exists := shapes[inferred]; exists {
			shape = s
			ok = true
			className = inferred
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
	if className == "Date" {
		timeVal := nextTemp(counter)
		if len(expression.Arguments) == 0 {
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpCall, Type: ir.TypeNumber, Result: timeVal, Callee: "__date.now", Span: toIRSpan(path, expression.Span),
			})
		} else {
			argVal, argType, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			if argType == ir.TypeNumber {
				timeVal = argVal
			} else if argType == ir.TypeString {
				function.Body = append(function.Body, ir.Instruction{
					Op: ir.OpCall, Type: ir.TypeNumber, Result: timeVal, Callee: "__date.parse", Args: []string{argVal}, Span: toIRSpan(path, expression.Span),
				})
			} else {
				timeVal = argVal
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{result, timeVal}, Span: toIRSpan(path, expression.Span),
		})
		return result, objType, nil
	}
	if className == "Error" || className == "TypeError" || className == "RangeError" || className == "SyntaxError" {
		msgVal := nextTemp(counter)
		if len(expression.Arguments) > 0 {
			mv, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			msgVal = mv
		} else {
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeString, Result: msgVal, Value: "", Span: toIRSpan(path, expression.Span),
			})
		}
		nameVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeString, Result: nameVal, Value: className, Span: toIRSpan(path, expression.Span),
		})
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "message", FieldIndex: 0, Args: []string{result, msgVal}, Span: toIRSpan(path, expression.Span),
		})
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "name", FieldIndex: 1, Args: []string{result, nameVal}, Span: toIRSpan(path, expression.Span),
		})
		stackVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeString, Result: stackVal, Value: className + ": " + path, Span: toIRSpan(path, expression.Span),
		})
		causeVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeString, Result: causeVal, Value: "", Span: toIRSpan(path, expression.Span),
		})
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "stack", FieldIndex: 2, Args: []string{result, stackVal}, Span: toIRSpan(path, expression.Span),
		})
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "cause", FieldIndex: 3, Args: []string{result, causeVal}, Span: toIRSpan(path, expression.Span),
		})
		return result, objType, nil
	}
	if className == "SuppressedError" {
		errVal := nextTemp(counter)
		suppVal := nextTemp(counter)
		msgVal := nextTemp(counter)
		if len(expression.Arguments) > 0 {
			ev, _, _ := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			errVal = ev
		} else {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: errVal, Value: "", Span: toIRSpan(path, expression.Span)})
		}
		if len(expression.Arguments) > 1 {
			sv, _, _ := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			suppVal = sv
		} else {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: suppVal, Value: "", Span: toIRSpan(path, expression.Span)})
		}
		if len(expression.Arguments) > 2 {
			mv, _, _ := lowerExpression(path, expression.Arguments[2], "", function, env, counter, shapes, signatures)
			msgVal = mv
		} else {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: msgVal, Value: "", Span: toIRSpan(path, expression.Span)})
		}
		nameVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: nameVal, Value: "SuppressedError", Span: toIRSpan(path, expression.Span)})
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "message", FieldIndex: 0, Args: []string{result, msgVal}, Span: toIRSpan(path, expression.Span)})
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "name", FieldIndex: 1, Args: []string{result, nameVal}, Span: toIRSpan(path, expression.Span)})
		stackVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: stackVal, Value: "SuppressedError: " + path, Span: toIRSpan(path, expression.Span)})
		causeVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: causeVal, Value: "", Span: toIRSpan(path, expression.Span)})
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "stack", FieldIndex: 2, Args: []string{result, stackVal}, Span: toIRSpan(path, expression.Span)})
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "cause", FieldIndex: 3, Args: []string{result, causeVal}, Span: toIRSpan(path, expression.Span)})
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "error", FieldIndex: 4, Args: []string{result, errVal}, Span: toIRSpan(path, expression.Span)})
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "suppressed", FieldIndex: 5, Args: []string{result, suppVal}, Span: toIRSpan(path, expression.Span)})
		return result, objType, nil
	}
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
		} else {
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

	// Call constructor if present
	if ctor, ctorName, found := findConstructorInHierarchy(className, signatures, classHierarchy); found {
		args := []string{result}
		for i, arg := range expression.Arguments {
			argVal, argType, err := lowerExpression(path, arg, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			paramIdx := i + 1
			if paramIdx < len(ctor.Parameters) {
				paramType := ctor.Parameters[paramIdx].Type
				if strings.HasPrefix(string(paramType), "object:") && strings.HasPrefix(string(argType), "object:") && paramType != argType {
					dstShapeName := strings.TrimPrefix(string(paramType), "object:")
					srcShapeName := strings.TrimPrefix(string(argType), "object:")
					if dstShape, ok := shapes[dstShapeName]; ok && strings.HasPrefix(srcShapeName, "__shape_") {
						adapted := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:         ir.OpObjectNew,
							Type:       paramType,
							Result:     adapted,
							Callee:     dstShapeName,
							FieldCount: len(dstShape.Fields),
							Span:       toIRSpan(path, arg.Span),
						})
						if srcShape, ok := shapes[srcShapeName]; ok {
							for dstIdx, dstField := range dstShape.Fields {
								for srcIdx, srcField := range srcShape.Fields {
									if srcField.Name == dstField.Name {
										fieldVal := nextTemp(counter)
										function.Body = append(function.Body, ir.Instruction{
											Op:         ir.OpFieldGet,
											Type:       dstField.Type,
											Result:     fieldVal,
											Callee:     srcShapeName,
											Field:      srcField.Name,
											FieldIndex: srcIdx,
											Args:       []string{argVal},
											Span:       toIRSpan(path, arg.Span),
										})
										function.Body = append(function.Body, ir.Instruction{
											Op:         ir.OpFieldSet,
											Type:       ir.TypeVoid,
											Callee:     dstShapeName,
											Field:      dstField.Name,
											FieldIndex: dstIdx,
											Args:       []string{adapted, fieldVal},
											Span:       toIRSpan(path, arg.Span),
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
							Value:  "null",
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
			defaults := defaultParamsIndex[ctorName]
			if defaults == nil {
				defaults = defaultParamsIndex[strings.Split(ctorName, "__")[0]]
			}
			for i := len(args); i < len(ctor.Parameters); i++ {
				var val string
				var valType ir.Type
				paramType := ctor.Parameters[i].Type
				if defaults != nil && defaults[i] != nil {
					defExpr := defaults[i]
					if paramType == ir.TypeNumber && (defExpr.Kind == "undefined" || defExpr.Kind == "null") {
						numConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   ir.TypeNumber,
							Result: numConst,
							Value:  "0",
							Span:   toIRSpan(path, defExpr.Span),
						})
						val = numConst
						valType = ir.TypeNumber
					} else if paramType == ir.TypeBool && (defExpr.Kind == "undefined" || defExpr.Kind == "null") {
						boolConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   ir.TypeBool,
							Result: boolConst,
							Value:  "false",
							Span:   toIRSpan(path, defExpr.Span),
						})
						val = boolConst
						valType = ir.TypeBool
					} else if paramType == ir.TypeBigInt && (defExpr.Kind == "undefined" || defExpr.Kind == "null") {
						biConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   ir.TypeBigInt,
							Result: biConst,
							Value:  "0",
							Span:   toIRSpan(path, defExpr.Span),
						})
						val = biConst
						valType = ir.TypeBigInt
					} else if (strings.HasPrefix(string(paramType), "object:") || isPointerLikeType(paramType)) && (defExpr.Kind == "null" || defExpr.Kind == "undefined") {
						nullConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   paramType,
							Result: nullConst,
							Value:  "null",
							Span:   toIRSpan(path, defExpr.Span),
						})
						val = nullConst
						valType = paramType
					} else {
						v, vt, err := lowerExpression(path, defExpr, "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						val = v
						valType = vt
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
