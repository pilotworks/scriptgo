package lowering

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// closureABIArgumentLimit is the number of arguments a closure call passes
// (scriptgo_closure_invoke takes four boxed arguments).
const closureABIArgumentLimit = 4

// lowerFunctionValueMethod lowers call, apply and bind on a function value.
// Closures in the native subset have no dynamic `this` (and strict mode bars
// `this` in a function without a `this:` parameter), so thisArg is evaluated
// but not passed: the call goes straight to the closure.
// apply spreads an array literal or tuple; a runtime-length array passes its
// first four elements (the closure ABI limit), missing ones as undefined.
// bind without partial arguments is the function itself.
func lowerFunctionValueMethod(path string, expression *frontend.SyntaxExpression, receiver string, receiverType ir.Type, methodName string, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, bool, error) {
	if receiverType != ir.TypeClosure || (methodName != "call" && methodName != "apply" && methodName != "bind") {
		return "", "", false, nil
	}
	span := toIRSpan(path, expression.Span)
	if methodName == "bind" {
		if len(expression.Arguments) > 1 {
			return "", "", true, fmt.Errorf("Function.prototype.bind with partial arguments is not supported in the native subset")
		}
		if len(expression.Arguments) == 1 {
			if _, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures); err != nil {
				return "", "", true, err
			}
		}
		return receiver, ir.TypeClosure, true, nil
	}
	if len(expression.Arguments) > 0 {
		// thisArg is evaluated for its effects.
		if _, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures); err != nil {
			return "", "", true, err
		}
	}
	var args []string
	var err error
	if methodName == "call" {
		args, err = lowerArgumentList(path, tail(expression.Arguments), function, env, counter, shapes, signatures)
	} else if len(expression.Arguments) > 1 {
		args, err = lowerApplyArguments(path, expression.Arguments[1], function, env, counter, shapes, signatures)
	}
	if err != nil {
		return "", "", true, err
	}
	if result == "" {
		result = nextTemp(counter)
	}
	retType := toIRType(expression.InferredType)
	if retType == "" {
		retType = ir.TypeUnknown
	}
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpClosureCall, Type: retType, Result: result, Callee: receiver, Args: args, Span: span})
	return result, retType, true, nil
}

func tail(arguments []*frontend.SyntaxExpression) []*frontend.SyntaxExpression {
	if len(arguments) == 0 {
		return nil
	}
	return arguments[1:]
}

func lowerArgumentList(path string, arguments []*frontend.SyntaxExpression, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) ([]string, error) {
	values := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

// lowerApplyArguments spreads apply's argument array into call arguments.
func lowerApplyArguments(path string, argument *frontend.SyntaxExpression, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) ([]string, error) {
	if argument.Kind == "array" {
		return lowerArgumentList(path, argument.Arguments, function, env, counter, shapes, signatures)
	}
	value, typ, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
	if err != nil {
		return nil, err
	}
	span := toIRSpan(path, argument.Span)
	shapeName := strings.TrimPrefix(string(typ), "object:")
	if shape, ok := lookupObjectShape(shapeName, shapes); ok && isTupleShape(shape) {
		values := make([]string, 0, len(shape.Fields))
		for index, field := range shape.Fields {
			element := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: field.Type, Result: element, Callee: shapeName, Field: field.Name, FieldIndex: index, Args: []string{value}, Span: span})
			values = append(values, element)
		}
		return values, nil
	}
	if typ != ir.TypeUnknownArray {
		return nil, fmt.Errorf("Function.prototype.apply needs an array literal, a tuple, or an unknown[] in the native subset, got %s", argument.InferredType)
	}
	values := make([]string, 0, closureABIArgumentLimit)
	for index := 0; index < closureABIArgumentLimit; index++ {
		position := nextTemp(counter)
		element := nextTemp(counter)
		function.Body = append(function.Body,
			ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: position, Value: strconv.Itoa(index), Span: span},
			ir.Instruction{Op: ir.OpIndex, Type: ir.TypeUnknown, Result: element, Args: []string{value, position}, Span: span},
		)
		values = append(values, element)
	}
	return values, nil
}
