package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerFunctionValueMethod lowers call, apply and bind on a function value.
// Closures in the native subset have no dynamic `this` (and strict mode bars
// `this` in a function without a `this:` parameter), so thisArg is evaluated
// but not passed: the call goes straight to the closure.
// apply spreads an array literal or tuple; a runtime-length array passes all
// of its elements through __closure.apply.
// bind without partial arguments is the function itself; with them it is a
// closure that prepends them (lowerBoundClosure).
func lowerFunctionValueMethod(path string, expression *frontend.SyntaxExpression, receiver string, receiverType ir.Type, methodName string, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, bool, error) {
	if receiverType != ir.TypeClosure || (methodName != "call" && methodName != "apply" && methodName != "bind") {
		return "", "", false, nil
	}
	span := toIRSpan(path, expression.Span)
	if methodName == "bind" {
		if len(expression.Arguments) > 0 {
			if _, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures); err != nil {
				return "", "", true, err
			}
		}
		if len(expression.Arguments) > 1 {
			return lowerBoundClosure(path, expression, receiver, result, function, env, counter, shapes, signatures)
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
	var spread string
	var err error
	if methodName == "call" {
		args, spread, err = lowerClosureCallArguments(path, tail(expression.Arguments), expression.Span, function, env, counter, shapes, signatures)
	} else if len(expression.Arguments) > 1 {
		args, spread, err = lowerApplyArguments(path, expression.Arguments[1], function, env, counter, shapes, signatures)
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
	appendClosureCall(function, counter, receiver, args, spread, retType, result, span)
	return result, retType, true, nil
}

// appendClosureCall calls closure with args, or with the elements of the
// unknown[] array spread when it is set.
func appendClosureCall(function *ir.Function, counter *int, closure string, args []string, spread string, retType ir.Type, result string, span ir.SourceSpan) {
	if spread == "" {
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpClosureCall, Type: retType, Result: result, Callee: closure, Args: args, Span: span})
		return
	}
	if retType == ir.TypeUnknown {
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeUnknown, Result: result, Callee: "__closure.apply", Args: []string{closure, spread}, Span: span})
		return
	}
	raw := nextTemp(counter)
	function.Body = append(function.Body,
		ir.Instruction{Op: ir.OpCall, Type: ir.TypeUnknown, Result: raw, Callee: "__closure.apply", Args: []string{closure, spread}, Span: span},
		ir.Instruction{Op: ir.OpCheckedCast, Type: retType, Result: result, Args: []string{raw}, Span: span},
	)
}

// lowerClosureCallArguments lowers a closure call's arguments. A call that
// spreads an array passes all of its arguments as one unknown[] array
// (spread) for __closure.apply.
func lowerClosureCallArguments(path string, arguments []*frontend.SyntaxExpression, span frontend.SourceSpan, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (args []string, spread string, err error) {
	for _, argument := range arguments {
		if argument.Kind == "spread" {
			array := &frontend.SyntaxExpression{Span: span, Kind: "array", Arguments: arguments, InferredType: "unknown[]"}
			spread, _, err = lowerExpression(path, array, "", function, env, counter, shapes, signatures)
			return nil, spread, err
		}
	}
	args, err = lowerArgumentList(path, arguments, function, env, counter, shapes, signatures)
	return args, "", err
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

// lowerApplyArguments spreads apply's argument array into call arguments. A
// runtime-length unknown[] is returned as spread instead, for the runtime to
// pass its elements.
func lowerApplyArguments(path string, argument *frontend.SyntaxExpression, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (args []string, spread string, err error) {
	if argument.Kind == "array" {
		args, spread, err = lowerClosureCallArguments(path, argument.Arguments, argument.Span, function, env, counter, shapes, signatures)
		return args, spread, err
	}
	value, typ, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
	if err != nil {
		return nil, "", err
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
		return values, "", nil
	}
	if typ != ir.TypeUnknownArray {
		return nil, "", fmt.Errorf("Function.prototype.apply needs an array literal, a tuple, or an unknown[] in the native subset, got %s", argument.InferredType)
	}
	return nil, value, nil
}

// lowerBoundClosure lowers f.bind(thisArg, ...partial) as the closure
// (...rest) => f(...partial, ...rest). f and the partial arguments are
// evaluated once, at bind time, into locals the closure captures.
func lowerBoundClosure(path string, expression *frontend.SyntaxExpression, receiver string, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, bool, error) {
	span := expression.Span
	ident := func(name string) *frontend.SyntaxExpression {
		return &frontend.SyntaxExpression{Span: span, Kind: "identifier", Text: name}
	}
	base := fmt.Sprintf("__bound_%d_%d", span.Start, *counter)
	target := base + "_target"
	function.Locals = append(function.Locals, ir.Parameter{Name: target, Type: ir.TypeClosure})
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpAssign, Type: ir.TypeClosure, Result: target, Args: []string{receiver}, Span: toIRSpan(path, span)})
	env[target] = ir.TypeClosure
	var arguments []*frontend.SyntaxExpression
	for index, argument := range expression.Arguments[1:] {
		name := fmt.Sprintf("%s_arg%d", base, index)
		if _, typ, err := lowerExpression(path, argument, name, function, env, counter, shapes, signatures); err != nil {
			return "", "", true, err
		} else {
			env[name] = typ
		}
		arguments = append(arguments, ident(name))
	}
	rest := base + "_rest"
	arguments = append(arguments, &frontend.SyntaxExpression{Span: span, Kind: "spread", Left: ident(rest)})
	returnType := extractTopLevelReturnType(expression.InferredType)
	call := &frontend.SyntaxExpression{Span: span, Kind: "call", Left: ident(target), Arguments: arguments, InferredType: returnType}
	closure := &frontend.SyntaxStatement{
		Span:       span,
		Kind:       "function",
		Type:       returnType,
		Parameters: []frontend.SyntaxParameter{{Span: span, Name: rest, Type: "unknown[]", InferredType: "unknown[]", Rest: true}},
		Body:       []frontend.SyntaxStatement{{Span: span, Kind: "return", Expression: call}},
	}
	value, typ, err := lowerClosureExpression(path, closure, result, function, env, counter, shapes, signatures)
	return value, typ, true, err
}
