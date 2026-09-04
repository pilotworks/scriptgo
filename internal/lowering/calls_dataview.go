package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerRegExpReceiverMethod(
	path string,
	expression *typescriptgo.SyntaxExpression,
	receiver string,
	methodName string,
	receiverType ir.Type,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, bool, error) {
	isRegExp := receiverType == "object:RegExp" || receiverType == "RegExp"
	if !isRegExp && receiverType == ir.TypeUnknown && expression.Left != nil && expression.Left.Left != nil && (expression.Left.Left.InferredType == "RegExp" || expression.Left.Left.InferredType == "object:RegExp") {
		isRegExp = true
		castRec := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCheckedCast,
			Type:   ir.Type("object:RegExp"),
			Result: castRec,
			Args:   []string{receiver},
			Span:   toIRSpan(path, expression.Span),
		})
		receiver = castRec
	}
	if !isRegExp {
		return "", "", false, nil
	}
	if methodName == "test" && len(expression.Arguments) > 0 {
		argVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		srcVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: srcVal, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
		flagsVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: flagsVal, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeBool, Result: result, Callee: "__regex.test", Args: []string{srcVal, flagsVal, argVal}, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeBool, true, nil
	}
	if methodName == "exec" && len(expression.Arguments) > 0 {
		argVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		srcVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: srcVal, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
		flagsVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: flagsVal, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeStringArray, Result: result, Callee: "__regex.exec_stateful", Args: []string{receiver, srcVal, flagsVal, argVal}, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeStringArray, true, nil
	}
	if methodName == "compile" {
		if len(expression.Arguments) > 0 {
			patternVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{receiver, patternVal}, Span: toIRSpan(path, expression.Span),
			})
		}
		if len(expression.Arguments) > 1 {
			flagsVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{receiver, flagsVal}, Span: toIRSpan(path, expression.Span),
			})
		}
		return receiver, receiverType, true, nil
	}
	return "", "", false, nil
}

func lowerDateReceiverMethod(
	path string,
	expression *typescriptgo.SyntaxExpression,
	receiver string,
	methodName string,
	receiverType ir.Type,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, bool, error) {
	if receiverType != "object:Date" {
		return "", "", false, nil
	}
	if methodName == "getTime" || methodName == "valueOf" {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeNumber, Result: result, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeNumber, true, nil
	}
	if methodName == "setTime" {
		argVal := "0"
		if len(expression.Arguments) > 0 {
			v, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			argVal = v
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{receiver, argVal}, Span: toIRSpan(path, expression.Span)})
		return argVal, ir.TypeNumber, true, nil
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
				return "", "", true, err
			}
			argVal = v
		} else {
			argVal = nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: argVal, Value: "0", Span: toIRSpan(path, expression.Span)})
		}
		if result == "" {
			result = nextTemp(counter)
		}
		callee := "__date." + methodName
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: callee, Args: []string{timeVal, argVal}, Span: toIRSpan(path, expression.Span)})
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{receiver, result}, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeNumber, true, nil

	case "getDate", "getDay", "getFullYear", "getHours", "getMilliseconds", "getMinutes", "getMonth", "getSeconds", "getTimezoneOffset",
		"getUTCDate", "getUTCDay", "getUTCFullYear", "getUTCHours", "getUTCMilliseconds", "getUTCMinutes", "getUTCMonth", "getUTCSeconds":
		timeVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeNumber, Result: timeVal, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
		if result == "" {
			result = nextTemp(counter)
		}
		callee := "__date." + methodName
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: callee, Args: []string{timeVal}, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeNumber, true, nil

	case "toISOString", "toJSON", "toString", "toDateString", "toTimeString", "toUTCString",
		"toLocaleDateString", "toLocaleString", "toLocaleTimeString":
		timeVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeNumber, Result: timeVal, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
		if result == "" {
			result = nextTemp(counter)
		}
		callee := "__date." + methodName
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: callee, Args: []string{timeVal}, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeString, true, nil
	}
	return "", "", false, nil
}

func lowerDataViewReceiverMethod(
	path string,
	expression *typescriptgo.SyntaxExpression,
	receiver string,
	methodName string,
	receiverType ir.Type,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, bool, error) {
	if receiverType != ir.TypeDataView {
		return "", "", false, nil
	}
	switch methodName {
	case "getInt8", "getUint8":
		if len(expression.Arguments) < 1 {
			return "", "", true, fmt.Errorf("%s requires byteOffset argument", methodName)
		}
		byteOffsetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
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
		return result, ir.TypeNumber, true, nil

	case "setInt8", "setUint8":
		if len(expression.Arguments) < 2 {
			return "", "", true, fmt.Errorf("%s requires byteOffset and value arguments", methodName)
		}
		byteOffsetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		valVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeVoid,
			Callee: "__dataview." + methodName,
			Args:   []string{receiver, byteOffsetVal, valVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return "", ir.TypeVoid, true, nil

	case "getInt16", "getUint16", "getInt32", "getUint32", "getFloat32", "getFloat64":
		if len(expression.Arguments) < 1 {
			return "", "", true, fmt.Errorf("%s requires byteOffset argument", methodName)
		}
		byteOffsetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		leVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeBool, Result: leVal, Value: "false", Span: toIRSpan(path, expression.Span),
		})
		if len(expression.Arguments) > 1 {
			le, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
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
		return result, ir.TypeNumber, true, nil

	case "setUint16", "setInt16", "setUint32", "setInt32", "setFloat32", "setFloat64":
		if len(expression.Arguments) < 2 {
			return "", "", true, fmt.Errorf("%s requires byteOffset and value arguments", methodName)
		}
		byteOffsetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		valVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		leVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeBool, Result: leVal, Value: "false", Span: toIRSpan(path, expression.Span),
		})
		if len(expression.Arguments) > 2 {
			le, _, err := lowerExpression(path, expression.Arguments[2], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
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
		return "", ir.TypeVoid, true, nil

	case "getBigInt64", "getBigUint64":
		if len(expression.Arguments) < 1 {
			return "", "", true, fmt.Errorf("%s requires byteOffset argument", methodName)
		}
		byteOffsetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		leVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeBool, Result: leVal, Value: "false", Span: toIRSpan(path, expression.Span),
		})
		if len(expression.Arguments) > 1 {
			le, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
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
		return result, ir.TypeBigInt, true, nil

	case "setBigInt64", "setBigUint64":
		if len(expression.Arguments) < 2 {
			return "", "", true, fmt.Errorf("%s requires byteOffset and value arguments", methodName)
		}
		byteOffsetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		valVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		leVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeBool, Result: leVal, Value: "false", Span: toIRSpan(path, expression.Span),
		})
		if len(expression.Arguments) > 2 {
			le, _, err := lowerExpression(path, expression.Arguments[2], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
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
		return "", ir.TypeVoid, true, nil
	}
	return "", "", false, nil
}

func lowerTextEncodingReceiverMethod(
	path string,
	expression *typescriptgo.SyntaxExpression,
	receiver string,
	methodName string,
	receiverType ir.Type,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, bool, error) {
	if receiverType == ir.TypeTextEncoder {
		switch methodName {
		case "encode":
			var args []string
			if len(expression.Arguments) > 0 {
				argVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
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
			return result, ir.TypeUint8Array, true, nil

		case "encodeInto":
			if len(expression.Arguments) < 2 {
				return "", "", true, fmt.Errorf("TextEncoder.encodeInto requires source and destination arguments")
			}
			srcVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			destVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
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
			return result, resType, true, nil
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
					return "", "", true, err
				}
				args = append(args, inputVal)
				if len(expression.Arguments) > 1 {
					optsVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", true, err
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
			return result, ir.TypeString, true, nil
		}
	}
	return "", "", false, nil
}
