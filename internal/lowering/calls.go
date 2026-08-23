package lowering

import (
	"fmt"
	"strconv"
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
			if receiverType == "object:RegExp" {
				if methodName == "test" && len(expression.Arguments) > 0 {
					argVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					srcVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: srcVal, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
					flagsVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: flagsVal, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeBool, Result: result, Callee: "__regex.test", Args: []string{srcVal, flagsVal, argVal}, Span: toIRSpan(path, expression.Span)})
					return result, ir.TypeBool, nil
				}
				if methodName == "exec" && len(expression.Arguments) > 0 {
					argVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					srcVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: srcVal, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
					flagsVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: flagsVal, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeStringArray, Result: result, Callee: "__regex.exec", Args: []string{srcVal, flagsVal, argVal}, Span: toIRSpan(path, expression.Span)})
					return result, ir.TypeStringArray, nil
				}
			}
			if receiverType == "object:Date" {
				if methodName == "getTime" || methodName == "valueOf" {
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeNumber, Result: result, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
					return result, ir.TypeNumber, nil
				}
				if methodName == "setTime" {
					argVal := "0"
					if len(expression.Arguments) > 0 {
						v, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						argVal = v
					}
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{receiver, argVal}, Span: toIRSpan(path, expression.Span)})
					return argVal, ir.TypeNumber, nil
				}
				switch methodName {
				case "setDate", "setFullYear", "setHours", "setMilliseconds", "setMinutes", "setMonth", "setSeconds",
					"setUTCDate", "setUTCFullYear", "setUTCHours", "setUTCMilliseconds", "setUTCMinutes", "setUTCMonth", "setUTCSeconds":
					timeVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeNumber, Result: timeVal, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
					var argVal string
					if len(expression.Arguments) > 0 {
						v, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						argVal = v
					} else {
						argVal = nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: argVal, Value: "0", Span: toIRSpan(path, expression.Span)})
					}
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: "__date." + methodName, Args: []string{timeVal, argVal}, Span: toIRSpan(path, expression.Span)})
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{receiver, result}, Span: toIRSpan(path, expression.Span)})
					return result, ir.TypeNumber, nil
				}
				switch methodName {
				case "getDate", "getDay", "getFullYear", "getHours", "getMilliseconds", "getMinutes", "getMonth", "getSeconds", "getTimezoneOffset",
					"getUTCDate", "getUTCDay", "getUTCFullYear", "getUTCHours", "getUTCMilliseconds", "getUTCMinutes", "getUTCMonth", "getUTCSeconds":
					timeVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeNumber, Result: timeVal, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: "__date." + methodName, Args: []string{timeVal}, Span: toIRSpan(path, expression.Span)})
					return result, ir.TypeNumber, nil
				}
				switch methodName {
				case "toISOString", "toJSON", "toString", "toDateString", "toTimeString", "toUTCString",
					"toLocaleDateString", "toLocaleString", "toLocaleTimeString", "toTemporalInstant":
					timeVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeNumber, Result: timeVal, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__date." + methodName, Args: []string{timeVal}, Span: toIRSpan(path, expression.Span)})
					return result, ir.TypeString, nil
				}
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
			if isTypedArrayType(receiverType) {
				if methodName == "subarray" || methodName == "slice" {
					beginVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op: ir.OpConst, Type: ir.TypeNumber, Result: beginVal, Value: "0", Span: toIRSpan(path, expression.Span),
					})
					endVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeNumber,
						Result: endVal,
						Callee: "__typedarray.length",
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
					callee := "__typedarray.subarray"
					if methodName == "slice" {
						callee = "__typedarray.slice"
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   receiverType,
						Result: result,
						Callee: callee,
						Args:   []string{receiver, beginVal, endVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, receiverType, nil
				}
				if methodName == "set" && len(expression.Arguments) > 0 {
					srcVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					offsetVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op: ir.OpConst, Type: ir.TypeNumber, Result: offsetVal, Value: "0", Span: toIRSpan(path, expression.Span),
					})
					if len(expression.Arguments) > 1 {
						off, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						offsetVal = off
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: "__typedarray.set",
						Args:   []string{receiver, srcVal, offsetVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return "", ir.TypeVoid, nil
				}
				if methodName == "fill" && len(expression.Arguments) > 0 {
					valVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					startVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op: ir.OpConst, Type: ir.TypeNumber, Result: startVal, Value: "0", Span: toIRSpan(path, expression.Span),
					})
					endVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeNumber,
						Result: endVal,
						Callee: "__typedarray.length",
						Args:   []string{receiver},
						Span:   toIRSpan(path, expression.Span),
					})
					if len(expression.Arguments) > 1 {
						s, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						startVal = s
					}
					if len(expression.Arguments) > 2 {
						e, _, err := lowerExpression(path, expression.Arguments[2], "", function, env, counter, shapes, signatures)
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
						Type:   receiverType,
						Result: result,
						Callee: "__typedarray.fill",
						Args:   []string{receiver, valVal, startVal, endVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, receiverType, nil
				}
			}
			if receiverType == ir.TypeDataView {
				switch methodName {
				case "getInt8", "getUint8":
					if len(expression.Arguments) < 1 {
						return "", "", fmt.Errorf("%s requires byteOffset argument", methodName)
					}
					byteOffsetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeNumber,
						Result: result,
						Callee: "__dataview." + methodName,
						Args:   []string{receiver, byteOffsetVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeNumber, nil

				case "setInt8", "setUint8":
					if len(expression.Arguments) < 2 {
						return "", "", fmt.Errorf("%s requires byteOffset and value arguments", methodName)
					}
					byteOffsetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					valVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: "__dataview." + methodName,
						Args:   []string{receiver, byteOffsetVal, valVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return "", ir.TypeVoid, nil

				case "getInt16", "getUint16", "getInt32", "getUint32", "getFloat32", "getFloat64":
					if len(expression.Arguments) < 1 {
						return "", "", fmt.Errorf("%s requires byteOffset argument", methodName)
					}
					byteOffsetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					leVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op: ir.OpConst, Type: ir.TypeBool, Result: leVal, Value: "false", Span: toIRSpan(path, expression.Span),
					})
					if len(expression.Arguments) > 1 {
						le, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						leVal = le
					}
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeNumber,
						Result: result,
						Callee: "__dataview." + methodName,
						Args:   []string{receiver, byteOffsetVal, leVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeNumber, nil

				case "setUint16", "setInt16", "setUint32", "setInt32", "setFloat32", "setFloat64":
					if len(expression.Arguments) < 2 {
						return "", "", fmt.Errorf("%s requires byteOffset and value arguments", methodName)
					}
					byteOffsetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					valVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					leVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op: ir.OpConst, Type: ir.TypeBool, Result: leVal, Value: "false", Span: toIRSpan(path, expression.Span),
					})
					if len(expression.Arguments) > 2 {
						le, _, err := lowerExpression(path, expression.Arguments[2], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						leVal = le
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: "__dataview." + methodName,
						Args:   []string{receiver, byteOffsetVal, valVal, leVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return "", ir.TypeVoid, nil

				case "getBigInt64", "getBigUint64":
					if len(expression.Arguments) < 1 {
						return "", "", fmt.Errorf("%s requires byteOffset argument", methodName)
					}
					byteOffsetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					leVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op: ir.OpConst, Type: ir.TypeBool, Result: leVal, Value: "false", Span: toIRSpan(path, expression.Span),
					})
					if len(expression.Arguments) > 1 {
						le, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						leVal = le
					}
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeBigInt,
						Result: result,
						Callee: "__dataview." + methodName,
						Args:   []string{receiver, byteOffsetVal, leVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeBigInt, nil

				case "setBigInt64", "setBigUint64":
					if len(expression.Arguments) < 2 {
						return "", "", fmt.Errorf("%s requires byteOffset and value arguments", methodName)
					}
					byteOffsetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					valVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					leVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op: ir.OpConst, Type: ir.TypeBool, Result: leVal, Value: "false", Span: toIRSpan(path, expression.Span),
					})
					if len(expression.Arguments) > 2 {
						le, _, err := lowerExpression(path, expression.Arguments[2], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						leVal = le
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: "__dataview." + methodName,
						Args:   []string{receiver, byteOffsetVal, valVal, leVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return "", ir.TypeVoid, nil
				}
			}
			if isMapType(receiverType) {
				switch methodName {
				case "set":
					if len(expression.Arguments) < 2 {
						return "", "", fmt.Errorf("Map.set requires key and value arguments")
					}
					kVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					vVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeMap,
						Result: result,
						Callee: "__map.set",
						Args:   []string{receiver, kVal, vVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeMap, nil

				case "get":
					if len(expression.Arguments) < 1 {
						return "", "", fmt.Errorf("Map.get requires key argument")
					}
					kVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					if result == "" {
						result = nextTemp(counter)
					}
					retType := ir.TypeUnknown
					if _, valType := resolveMapTypes(expression.Left, env); valType != "" {
						retType = toIRType(valType)
					} else if expression.InferredType != "" {
						retType = toIRType(expression.InferredType)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   retType,
						Result: result,
						Callee: "__map.get",
						Args:   []string{receiver, kVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, retType, nil

				case "has":
					if len(expression.Arguments) < 1 {
						return "", "", fmt.Errorf("Map.has requires key argument")
					}
					kVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
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
						Callee: "__map.has",
						Args:   []string{receiver, kVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeBool, nil

				case "delete":
					if len(expression.Arguments) < 1 {
						return "", "", fmt.Errorf("Map.delete requires key argument")
					}
					kVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
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
						Callee: "__map.delete",
						Args:   []string{receiver, kVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeBool, nil

				case "clear":
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: "__map.clear",
						Args:   []string{receiver},
						Span:   toIRSpan(path, expression.Span),
					})
					return "", ir.TypeVoid, nil

				case "keys", "values", "entries":
					if result == "" {
						result = nextTemp(counter)
					}
					retType := ir.TypeStringArray
					if expression.InferredType != "" {
						retType = toIRType(expression.InferredType)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   retType,
						Result: result,
						Callee: "__map." + methodName,
						Args:   []string{receiver},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, retType, nil

				case "forEach":
					if len(expression.Arguments) < 1 {
						return "", "", fmt.Errorf("Map.forEach requires callback argument")
					}
					if cb := expression.Arguments[0]; cb != nil && (cb.Kind == "arrow_function" || cb.Kind == "function") && cb.Function != nil {
						keyType, valType := resolveMapTypes(expression.Left, env)
						if valType != "" && len(cb.Function.Parameters) > 0 && cb.Function.Parameters[0].Type == "" && cb.Function.Parameters[0].InferredType == "" {
							cb.Function.Parameters[0].Type = valType
						}
						if keyType != "" && len(cb.Function.Parameters) > 1 && cb.Function.Parameters[1].Type == "" && cb.Function.Parameters[1].InferredType == "" {
							cb.Function.Parameters[1].Type = keyType
						}
					}
					cbVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: "__map.forEach",
						Args:   []string{receiver, cbVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return "", ir.TypeVoid, nil
				}
			}
			if isSetType(receiverType) {
				switch methodName {
				case "add":
					if len(expression.Arguments) < 1 {
						return "", "", fmt.Errorf("Set.add requires value argument")
					}
					vVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeSet,
						Result: result,
						Callee: "__set.add",
						Args:   []string{receiver, vVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeSet, nil

				case "has":
					if len(expression.Arguments) < 1 {
						return "", "", fmt.Errorf("Set.has requires value argument")
					}
					vVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
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
						Callee: "__set.has",
						Args:   []string{receiver, vVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeBool, nil

				case "delete":
					if len(expression.Arguments) < 1 {
						return "", "", fmt.Errorf("Set.delete requires value argument")
					}
					vVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
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
						Callee: "__set.delete",
						Args:   []string{receiver, vVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeBool, nil

				case "clear":
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: "__set.clear",
						Args:   []string{receiver},
						Span:   toIRSpan(path, expression.Span),
					})
					return "", ir.TypeVoid, nil

				case "keys", "values", "entries":
					if result == "" {
						result = nextTemp(counter)
					}
					retType := ir.TypeNumberArray
					if expression.InferredType != "" {
						retType = toIRType(expression.InferredType)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   retType,
						Result: result,
						Callee: "__set." + methodName,
						Args:   []string{receiver},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, retType, nil

				case "forEach":
					if len(expression.Arguments) < 1 {
						return "", "", fmt.Errorf("Set.forEach requires callback argument")
					}
					cbVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: "__set.forEach",
						Args:   []string{receiver, cbVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return "", ir.TypeVoid, nil
				}
			}
			if receiverType == ir.TypeTextEncoder {
				switch methodName {
				case "encode":
					var args []string
					if len(expression.Arguments) > 0 {
						argVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						args = append(args, argVal)
					}
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeUint8Array,
						Result: result,
						Callee: "__text_encoder.encode",
						Args:   args,
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeUint8Array, nil

				case "encodeInto":
					if len(expression.Arguments) < 2 {
						return "", "", fmt.Errorf("TextEncoder.encodeInto requires source and destination arguments")
					}
					srcVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					destVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					if result == "" {
						result = nextTemp(counter)
					}
					shapeName := ensureTextEncoderEncodeIntoResultShape(shapes)
					resType := ir.Type("object:" + shapeName)
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   resType,
						Result: result,
						Callee: "__text_encoder.encode_into",
						Args:   []string{srcVal, destVal},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, resType, nil
				}
			}
			if receiverType == ir.TypeTextDecoder {
				switch methodName {
				case "decode":
					var args []string
					args = append(args, receiver)
					if len(expression.Arguments) > 0 {
						inputVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						args = append(args, inputVal)
						if len(expression.Arguments) > 1 {
							optsVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
							if err != nil {
								return "", "", err
							}
							args = append(args, optsVal)
						}
					}
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeString,
						Result: result,
						Callee: "__text_decoder.decode",
						Args:   args,
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeString, nil
				}
			}
			if receiverType == ir.TypeString && isStringMethod(methodName) {
				if methodName == "match" || methodName == "search" || methodName == "matchAll" {
					if len(expression.Arguments) > 0 {
						argVal, argTyp, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						srcVal := argVal
						flagsVal := nextTemp(counter)
						if argTyp == "object:RegExp" {
							srcVal = nextTemp(counter)
							function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: srcVal, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{argVal}, Span: toIRSpan(path, expression.Span)})
							function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: flagsVal, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{argVal}, Span: toIRSpan(path, expression.Span)})
						} else {
							function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: flagsVal, Value: "", Span: toIRSpan(path, expression.Span)})
						}
						if result == "" {
							result = nextTemp(counter)
						}
						switch methodName {
						case "match":
							function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeStringArray, Result: result, Callee: "__string.match", Args: []string{receiver, srcVal, flagsVal}, Span: toIRSpan(path, expression.Span)})
							return result, ir.TypeStringArray, nil
						case "matchAll":
							function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeStringArray, Result: result, Callee: "__string.matchAll", Args: []string{receiver, srcVal, flagsVal}, Span: toIRSpan(path, expression.Span)})
							return result, ir.TypeStringArray, nil
						default:
							function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: "__string.search", Args: []string{receiver, srcVal, flagsVal}, Span: toIRSpan(path, expression.Span)})
							return result, ir.TypeNumber, nil
						}
					}
				}
				if methodName == "replace" && len(expression.Arguments) >= 2 {
					arg0Val, arg0Typ, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					if arg0Typ == "object:RegExp" {
						arg1Val, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						srcVal := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: srcVal, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{arg0Val}, Span: toIRSpan(path, expression.Span)})
						flagsVal := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: flagsVal, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{arg0Val}, Span: toIRSpan(path, expression.Span)})
						if result == "" {
							result = nextTemp(counter)
						}
						function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__string.replace_regex", Args: []string{receiver, srcVal, flagsVal, arg1Val}, Span: toIRSpan(path, expression.Span)})
						return result, ir.TypeString, nil
					}
				}
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
				returnType := ir.TypeNumber
				switch methodName {
				case "slice", "trim", "trimStart", "trimEnd", "trimLeft", "trimRight", "replace", "replaceAll", "substring", "substr", "charAt", "at", "toLowerCase", "toUpperCase", "toLocaleLowerCase", "toLocaleUpperCase", "repeat", "padStart", "padEnd", "concat", "toWellFormed", "normalize", "valueOf", "toString", "anchor", "big", "blink", "bold", "fixed", "fontcolor", "fontsize", "italics", "link", "small", "strike", "sub", "sup":
					returnType = ir.TypeString
				case "startsWith", "endsWith", "includes", "isWellFormed":
					returnType = ir.TypeBool
				case "split", "matchAll":
					returnType = ir.TypeStringArray
				case "indexOf", "lastIndexOf", "charCodeAt", "codePointAt", "localeCompare":
					returnType = ir.TypeNumber
				}
				if result == "" {
					result = nextTemp(counter)
				}
				env[result] = returnType
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: "__string." + methodName, Args: args, Span: toIRSpan(path, expression.Span)})
				return result, returnType, nil
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
			if (receiverType == ir.TypeNumberArray || receiverType == ir.TypeStringArray || receiverType == ir.TypeBoolArray || receiverType == ir.TypeBigIntArray || strings.HasSuffix(string(receiverType), "[]")) && isArrayMethod(methodName) {
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
				returnType := ir.TypeNumber
				switch methodName {
				case "map", "flatMap":
					returnType = receiverType
					if len(args) > 1 {
						if cbRet, ok := env[args[1]+".retType"]; ok && cbRet != "" {
							switch cbRet {
							case ir.TypeNumber:
								returnType = ir.TypeNumberArray
							case ir.TypeString:
								returnType = ir.TypeStringArray
							case ir.TypeBool:
								returnType = ir.TypeBoolArray
							case ir.TypeBigInt:
								returnType = ir.TypeBigIntArray
							default:
								returnType = ir.Type(string(cbRet) + "[]")
							}
						}
					}
				case "slice", "reverse", "concat", "splice", "filter", "fill", "toReversed", "toSorted", "toSpliced", "with", "sort", "copyWithin", "flat", "values":
					returnType = receiverType
				case "keys":
					returnType = ir.TypeNumberArray
				case "entries":
					returnType = ir.TypeStringArray
				case "includes", "some", "every":
					returnType = ir.TypeBool
				case "join", "toString", "toLocaleString":
					returnType = ir.TypeString
				case "push", "unshift", "indexOf", "lastIndexOf", "reduce", "reduceRight", "findIndex", "findLastIndex":
					returnType = ir.TypeNumber
				case "pop", "shift", "at", "find", "findLast":
					if receiverType == ir.TypeNumberArray {
						returnType = ir.TypeNumber
					} else if receiverType == ir.TypeStringArray {
						returnType = ir.TypeString
					} else if receiverType == ir.TypeBoolArray {
						returnType = ir.TypeBool
					} else if receiverType == ir.TypeBigIntArray {
						returnType = ir.TypeBigInt
					} else if before, ok := strings.CutSuffix(string(receiverType), "[]"); ok {
						elemTypeStr := before
						returnType = toIRType(elemTypeStr)
					} else {
						returnType = ir.TypeString
					}
				case "forEach":
					returnType = ir.TypeVoid
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: "__array." + methodName, Args: args, Span: toIRSpan(path, expression.Span)})
				return result, returnType, nil
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

					overrides := findOverridingSubclasses(className, methodName, classHierarchy, signatures)
					isAbstract := isAbstractMethodInHierarchy(className, methodName)
					if len(overrides) == 0 && !isAbstract {
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

					if target.ReturnType != ir.TypeVoid {
						env[result] = target.ReturnType
						defaultVal := ""
						if target.ReturnType == ir.TypeNumber {
							defaultVal = "0"
						} else if target.ReturnType == ir.TypeBool {
							defaultVal = "false"
						}
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   target.ReturnType,
							Result: result,
							Value:  defaultVal,
							Span:   toIRSpan(path, expression.Span),
						})
					}
					var elseBody []ir.Instruction
					if _, hasSig := signatures[mangled]; hasSig && !isAbstract {
						baseCallRes := nextTemp(counter)
						elseBody = append(elseBody, ir.Instruction{
							Op:     ir.OpCall,
							Type:   target.ReturnType,
							Result: baseCallRes,
							Callee: mangled,
							Args:   args,
							Span:   toIRSpan(path, expression.Span),
						})
						if target.ReturnType != ir.TypeVoid {
							elseBody = append(elseBody, ir.Instruction{
								Op:     ir.OpAssign,
								Type:   target.ReturnType,
								Result: result,
								Args:   []string{baseCallRes},
								Span:   toIRSpan(path, expression.Span),
							})
						}
					}

					currElse := elseBody
					for _, sub := range overrides {
						subMangled := sub + "_" + methodName
						instCheck := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpInstanceOf,
							Type:   ir.TypeBool,
							Result: instCheck,
							Args:   []string{receiver},
							Value:  sub,
							Span:   toIRSpan(path, expression.Span),
						})

						var thenBody []ir.Instruction
						subCallRes := nextTemp(counter)
						thenBody = append(thenBody, ir.Instruction{
							Op:     ir.OpCall,
							Type:   target.ReturnType,
							Result: subCallRes,
							Callee: subMangled,
							Args:   args,
							Span:   toIRSpan(path, expression.Span),
						})
						if target.ReturnType != ir.TypeVoid {
							thenBody = append(thenBody, ir.Instruction{
								Op:     ir.OpAssign,
								Type:   target.ReturnType,
								Result: result,
								Args:   []string{subCallRes},
								Span:   toIRSpan(path, expression.Span),
							})
						}

						ifInst := ir.Instruction{
							Op:   ir.OpIf,
							Type: ir.TypeVoid,
							Args: []string{instCheck},
							Then: thenBody,
							Else: currElse,
							Span: toIRSpan(path, expression.Span),
						}
						currElse = []ir.Instruction{ifInst}
					}
					function.Body = append(function.Body, currElse...)
					return result, target.ReturnType, nil
				}
				if methodName == "hasOwnProperty" || methodName == "propertyIsEnumerable" || methodName == "isPrototypeOf" || methodName == "toString" || methodName == "toLocaleString" || methodName == "valueOf" {
					var args []string
					args = append(args, receiver)
					for _, argument := range expression.Arguments {
						argVal, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						args = append(args, argVal)
					}
					if result == "" {
						result = nextTemp(counter)
					}
					retType := ir.TypeString
					if methodName == "hasOwnProperty" || methodName == "propertyIsEnumerable" || methodName == "isPrototypeOf" {
						retType = ir.TypeBool
					} else if methodName == "valueOf" {
						return receiver, receiverType, nil
					}
					callee := "__object." + methodName
					if methodName == "hasOwnProperty" {
						callee = "__object.hasOwn"
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   retType,
						Result: result,
						Callee: callee,
						Args:   args,
						Span:   toIRSpan(path, expression.Span),
					})
					return result, retType, nil
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
	if callee == "Promise.all" && len(expression.Arguments) > 0 {
		arrExpr := expression.Arguments[0]
		if arrExpr.Kind == "array" {
			var resArgs []string
			var resElemType ir.Type = ir.TypeUnknown
			for _, elem := range arrExpr.Arguments {
				promVal, promType, err := lowerExpression(path, elem, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				awaitedVal := nextTemp(counter)
				elemType := ir.TypeNumber
				if strings.HasPrefix(string(promType), "object:Promise_") {
					elemType = toIRType(strings.TrimPrefix(string(promType), "object:Promise_"))
				} else if strings.HasPrefix(string(promType), "object:Promise<") && strings.HasSuffix(string(promType), ">") {
					elemType = toIRType(strings.TrimSuffix(strings.TrimPrefix(string(promType), "object:Promise<"), ">"))
				} else if strings.HasPrefix(string(promType), "object:") {
					elemType = promType
				}
				resElemType = elemType
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   elemType,
					Result: awaitedVal,
					Callee: "__async.await",
					Args:   []string{promVal},
					Span:   toIRSpan(path, elem.Span),
				})
				resArgs = append(resArgs, awaitedVal)
			}
			var fields []ir.Field
			for i := range arrExpr.Arguments {
				fields = append(fields, ir.Field{
					Name: strconv.Itoa(i),
					Type: resElemType,
					Span: toIRSpan(path, expression.Span),
				})
			}
			shapeName := anonymousShapeName(fields)
			if _, ok := shapes[shapeName]; !ok {
				shapes[shapeName] = ir.ObjectShape{
					Name:   shapeName,
					Span:   toIRSpan(path, expression.Span),
					Fields: fields,
				}
			}
			objType := ir.Type("object:" + shapeName)
			arrRes := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:         ir.OpObjectNew,
				Type:       objType,
				Result:     arrRes,
				Callee:     shapeName,
				FieldCount: len(fields),
				Span:       toIRSpan(path, expression.Span),
			})
			for i, field := range fields {
				function.Body = append(function.Body, ir.Instruction{
					Op:         ir.OpFieldSet,
					Type:       ir.TypeVoid,
					Callee:     shapeName,
					Field:      field.Name,
					FieldIndex: i,
					Args:       []string{arrRes, resArgs[i]},
					Span:       toIRSpan(path, expression.Span),
				})
			}
			if result == "" {
				result = nextTemp(counter)
			}
			promRetType := ir.Type("object:Promise<" + string(objType) + ">")
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   promRetType,
				Result: result,
				Callee: "__async.promise_resolve",
				Args:   []string{arrRes},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, promRetType, nil
		}
	}
	if callee == "Promise.allSettled" && len(expression.Arguments) > 0 {
		arrExpr := expression.Arguments[0]
		if arrExpr.Kind == "array" {
			var resArgs []string
			settledShapeName := "PromiseSettledResult"
			if _, ok := shapes[settledShapeName]; !ok {
				shapes[settledShapeName] = ir.ObjectShape{
					Name: settledShapeName,
					Span: toIRSpan(path, expression.Span),
					Fields: []ir.Field{
						{Name: "status", Type: ir.TypeString, Span: toIRSpan(path, expression.Span)},
						{Name: "value", Type: ir.TypeUnknown, Span: toIRSpan(path, expression.Span)},
					},
				}
			}
			for _, elem := range arrExpr.Arguments {
				promVal, promType, err := lowerExpression(path, elem, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				awaitedVal := nextTemp(counter)
				elemType := ir.TypeNumber
				if strings.HasPrefix(string(promType), "object:Promise_") {
					elemType = toIRType(strings.TrimPrefix(string(promType), "object:Promise_"))
				} else if strings.HasPrefix(string(promType), "object:Promise<") && strings.HasSuffix(string(promType), ">") {
					elemType = toIRType(strings.TrimSuffix(strings.TrimPrefix(string(promType), "object:Promise<"), ">"))
				} else if strings.HasPrefix(string(promType), "object:") {
					elemType = promType
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   elemType,
					Result: awaitedVal,
					Callee: "__async.await",
					Args:   []string{promVal},
					Span:   toIRSpan(path, elem.Span),
				})
				statusConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeString,
					Result: statusConst,
					Value:  "fulfilled",
					Span:   toIRSpan(path, elem.Span),
				})
				itemObj := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:         ir.OpObjectNew,
					Type:       ir.Type("object:" + settledShapeName),
					Result:     itemObj,
					Callee:     settledShapeName,
					FieldCount: 2,
					Span:       toIRSpan(path, elem.Span),
				})
				function.Body = append(function.Body, ir.Instruction{
					Op:         ir.OpFieldSet,
					Type:       ir.TypeVoid,
					Callee:     settledShapeName,
					Field:      "status",
					FieldIndex: 0,
					Args:       []string{itemObj, statusConst},
					Span:       toIRSpan(path, elem.Span),
				})
				function.Body = append(function.Body, ir.Instruction{
					Op:         ir.OpFieldSet,
					Type:       ir.TypeVoid,
					Callee:     settledShapeName,
					Field:      "value",
					FieldIndex: 1,
					Args:       []string{itemObj, awaitedVal},
					Span:       toIRSpan(path, elem.Span),
				})
				resArgs = append(resArgs, itemObj)
			}
			var fields []ir.Field
			for i := range arrExpr.Arguments {
				fields = append(fields, ir.Field{
					Name: strconv.Itoa(i),
					Type: ir.Type("object:" + settledShapeName),
					Span: toIRSpan(path, expression.Span),
				})
			}
			shapeName := anonymousShapeName(fields)
			if _, ok := shapes[shapeName]; !ok {
				shapes[shapeName] = ir.ObjectShape{
					Name:   shapeName,
					Span:   toIRSpan(path, expression.Span),
					Fields: fields,
				}
			}
			objType := ir.Type("object:" + shapeName)
			arrRes := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:         ir.OpObjectNew,
				Type:       objType,
				Result:     arrRes,
				Callee:     shapeName,
				FieldCount: len(fields),
				Span:       toIRSpan(path, expression.Span),
			})
			for i, field := range fields {
				function.Body = append(function.Body, ir.Instruction{
					Op:         ir.OpFieldSet,
					Type:       ir.TypeVoid,
					Callee:     shapeName,
					Field:      field.Name,
					FieldIndex: i,
					Args:       []string{arrRes, resArgs[i]},
					Span:       toIRSpan(path, expression.Span),
				})
			}
			if result == "" {
				result = nextTemp(counter)
			}
			promRetType := ir.Type("object:Promise<" + string(objType) + ">")
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   promRetType,
				Result: result,
				Callee: "__async.promise_resolve",
				Args:   []string{arrRes},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, promRetType, nil
		}
	}
	if callee == "Promise.any" && len(expression.Arguments) > 0 {
		arrExpr := expression.Arguments[0]
		if arrExpr.Kind == "array" && len(arrExpr.Arguments) > 0 {
			elem := arrExpr.Arguments[0]
			promVal, promType, err := lowerExpression(path, elem, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			awaitedVal := nextTemp(counter)
			elemType := ir.TypeNumber
			if strings.HasPrefix(string(promType), "object:Promise_") {
				elemType = toIRType(strings.TrimPrefix(string(promType), "object:Promise_"))
			} else if strings.HasPrefix(string(promType), "object:Promise<") && strings.HasSuffix(string(promType), ">") {
				elemType = toIRType(strings.TrimSuffix(strings.TrimPrefix(string(promType), "object:Promise<"), ">"))
			} else if strings.HasPrefix(string(promType), "object:") {
				elemType = promType
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   elemType,
				Result: awaitedVal,
				Callee: "__async.await",
				Args:   []string{promVal},
				Span:   toIRSpan(path, elem.Span),
			})
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.Type("object:Promise"),
				Result: result,
				Callee: "__async.promise_resolve",
				Args:   []string{awaitedVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.Type("object:Promise"), nil
		}
	}
	if callee == "Promise.withResolvers" {
		promRes := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.Type("object:Promise"),
			Result: promRes,
			Callee: "__async.promise_create",
			Args:   nil,
			Span:   toIRSpan(path, expression.Span),
		})
		fields := []ir.Field{
			{Name: "promise", Type: ir.Type("object:Promise"), Span: toIRSpan(path, expression.Span)},
			{Name: "resolve", Type: ir.TypeClosure, Span: toIRSpan(path, expression.Span)},
			{Name: "reject", Type: ir.TypeClosure, Span: toIRSpan(path, expression.Span)},
		}
		shapeName := "PromiseWithResolvers"
		if _, ok := shapes[shapeName]; !ok {
			shapes[shapeName] = ir.ObjectShape{
				Name:   shapeName,
				Span:   toIRSpan(path, expression.Span),
				Fields: fields,
			}
		}
		resObj := result
		if resObj == "" {
			resObj = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:         ir.OpObjectNew,
			Type:       ir.Type("object:" + shapeName),
			Result:     resObj,
			Callee:     shapeName,
			FieldCount: 3,
			Span:       toIRSpan(path, expression.Span),
		})
		function.Body = append(function.Body, ir.Instruction{
			Op:         ir.OpFieldSet,
			Type:       ir.TypeVoid,
			Callee:     shapeName,
			Field:      "promise",
			FieldIndex: 0,
			Args:       []string{resObj, promRes},
			Span:       toIRSpan(path, expression.Span),
		})
		return resObj, ir.Type("object:" + shapeName), nil
	}
	if callee == "Array.fromAsync" && len(expression.Arguments) > 0 {
		srcExpr := expression.Arguments[0]
		srcVal, srcType, err := lowerExpression(path, srcExpr, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.Type("object:Promise"),
			Result: result,
			Callee: "__async.promise_resolve",
			Args:   []string{srcVal},
			Span:   toIRSpan(path, expression.Span),
		})
		_ = srcType
		return result, ir.Type("object:Promise"), nil
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
		} else if strings.HasPrefix(string(calleeType), "object:Generator_") {
			retType = calleeType
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
	for i, argument := range expression.Arguments {
		value, valType, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if i < len(target.Parameters) && target.Parameters[i].Type == ir.TypeUnknown && valType != ir.TypeUnknown {
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
		args = append(args, value)
	}
	if len(args) < len(target.Parameters) {
		defaults := defaultParamsIndex[callee]
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
					} else if i < len(target.Parameters) && strings.HasPrefix(string(target.Parameters[i].Type), "object:") && (initExpr.Kind == "null" || initExpr.Kind == "undefined") {
						nullConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   target.Parameters[i].Type,
							Result: nullConst,
							Value:  "null",
							Span:   toIRSpan(path, initExpr.Span),
						})
						val = nullConst
					}
					args = append(args, val)
				}
			}
		}
	}
	if restParamsIndex[callee] && len(target.Parameters) > 0 && (target.Parameters[len(target.Parameters)-1].Type == ir.TypeStringArray || target.Parameters[len(target.Parameters)-1].Type == ir.TypeNumberArray) {
		restType := target.Parameters[len(target.Parameters)-1].Type
		fixed := len(target.Parameters) - 1
		if len(args) < fixed {
			return "", "", fmt.Errorf("call to %q has too few arguments", callee)
		}
		restArgs := append([]string(nil), args[fixed:]...)
		args = args[:fixed]
		array := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: restType, Result: array, Args: restArgs, Span: toIRSpan(path, expression.Span)})
		args = append(args, array)
	}
	if result == "" {
		result = nextTemp(counter)
	}
	returnType := target.ReturnType
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: callee, Args: args, Span: toIRSpan(path, expression.Span)})
	return result, returnType, nil
}

func callName(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil {
		return ""
	}
	if expression.Kind == "identifier" {
		return expression.Text
	}
	if (expression.Kind == "property" || expression.Kind == "optional_property") && expression.Left != nil && expression.Left.Kind == "identifier" {
		return expression.Left.Text + "." + expression.Text
	}
	return ""
}

func isStringMethod(name string) bool {
	switch name {
	case "split", "indexOf", "lastIndexOf", "slice", "startsWith", "endsWith", "trim", "trimStart", "trimEnd", "trimLeft", "trimRight", "replace", "replaceAll", "substring", "substr", "charAt", "at", "charCodeAt", "includes", "toLowerCase", "toUpperCase", "toLocaleLowerCase", "toLocaleUpperCase", "repeat", "padStart", "padEnd", "concat", "match", "matchAll", "search", "codePointAt", "isWellFormed", "toWellFormed", "localeCompare", "normalize", "valueOf", "toString", "anchor", "big", "blink", "bold", "fixed", "fontcolor", "fontsize", "italics", "link", "small", "strike", "sub", "sup":
		return true
	default:
		return false
	}
}

func isArrayMethod(name string) bool {
	switch name {
	case "push", "pop", "slice", "indexOf", "lastIndexOf", "includes", "join", "reverse", "concat", "shift", "unshift", "splice", "at", "map", "filter", "forEach", "reduce", "reduceRight", "find", "findLast", "some", "every", "findIndex", "findLastIndex", "fill", "toReversed", "toSorted", "toSpliced", "with", "sort", "copyWithin", "toString", "toLocaleString", "flat", "flatMap", "entries", "keys", "values":
		return true
	default:
		return false
	}
}

func stringMethod(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil || (expression.Kind != "property" && expression.Kind != "optional_property") || expression.Left == nil {
		return ""
	}
	if isStringMethod(expression.Text) {
		return expression.Text
	}
	return ""
}

func arrayMethod(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil || (expression.Kind != "property" && expression.Kind != "optional_property") || expression.Left == nil {
		return ""
	}
	if isArrayMethod(expression.Text) {
		return expression.Text
	}
	return ""
}

func resolveMapTypes(expr *typescriptgo.SyntaxExpression, env map[string]ir.Type) (string, string) {
	if expr == nil {
		return "", ""
	}
	target := expr
	if target.Kind == "property" && target.Left != nil && (target.Text == "get" || target.Text == "forEach" || target.Text == "set" || target.Text == "has" || target.Text == "delete") {
		target = target.Left
	}
	rawType := target.InferredType
	if target.Kind == "property" {
		propName := target.Text
		for _, meta := range classHierarchy {
			for _, f := range meta.Fields {
				if f.Name == propName && strings.HasPrefix(f.Type, "Map<") {
					rawType = f.Type
					break
				}
			}
			if rawType != "" {
				break
			}
		}
	} else if target.Kind == "identifier" {
		if t, ok := env[target.Text]; ok {
			rawType = string(t)
		}
	}
	if strings.Contains(rawType, "<") && strings.HasSuffix(rawType, ">") {
		idx := strings.Index(rawType, "<")
		inner := rawType[idx+1 : len(rawType)-1]
		parts := splitTypeArguments(inner)
		if len(parts) >= 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}
	return "", ""
}
