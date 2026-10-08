package lowering

import (
	"fmt"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerNewArrayBuffer(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, className string) (string, ir.Type, error) {
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

func lowerNewTypedArray(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, className string) (string, ir.Type, error) {
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
	if isTypedArrayType(arg0Type) {
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   targetType,
			Result: result,
			Callee: "__typedarray.new_typed_array",
			Value:  className,
			Args:   []string{arg0Val},
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

func lowerNewDataView(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
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

func lowerNewTextDecoder(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
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
