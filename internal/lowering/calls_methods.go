package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// isPromiseReceiverType reports receivers whose then/catch are
// Promise.prototype methods; other objects may define methods with those names.
func isPromiseReceiverType(receiverType ir.Type) bool {
	return strings.HasPrefix(string(receiverType), "object:Promise")
}

func lowerPromiseThenCatchCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, methodName string, receiver string) (string, ir.Type, error) {
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
		Value:  string(env[args[1]+".retType"]),
		Span:   toIRSpan(path, expression.Span),
	})
	return result, ir.Type("object:Promise"), nil
}

func lowerArrayBufferSliceCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, receiver string) (string, ir.Type, error) {
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

func lowerNumberFormatCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, methodName string, receiver string) (string, ir.Type, error) {
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

func lowerHasOwnPropertyCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, receiver string) (string, ir.Type, error) {
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
