package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// promiseCombinators maps Promise.all/allSettled/any/race to the runtime
// combinator. The runtime subscribes to every input in order and settles the
// result as the specification does (first rejection for all, first
// fulfillment or an AggregateError for any, first settlement for race).
var promiseCombinators = map[string]string{
	"Promise.all":        "__async.promise_all",
	"Promise.allSettled": "__async.promise_all_settled",
	"Promise.any":        "__async.promise_any",
	"Promise.race":       "__async.promise_race",
}

// lowerPromiseCombinator lowers a combinator over an array of promises or
// values. all and allSettled resolve to the array TypeScript types the call
// with: a runtime-filled array, or for a tuple type (an array literal input)
// a tuple object allocated here so element types stay positional. The call
// records that array type in Value for the backend to lay out the elements.
func lowerPromiseCombinator(path string, expression *frontend.SyntaxExpression, callee string, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, bool, error) {
	runtimeCallee, ok := promiseCombinators[callee]
	if !ok {
		return "", "", false, nil
	}
	if len(expression.Arguments) != 1 {
		return "", "", true, fmt.Errorf("%s requires one iterable argument", callee)
	}
	argument := expression.Arguments[0]
	if argument.Kind == "array" {
		// A literal is typed as a tuple; the combinator reads its elements
		// as tagged values.
		boxed := *argument
		boxed.InferredType = "unknown[]"
		argument = &boxed
	}
	input, inputType, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", true, err
	}
	if !strings.HasSuffix(string(inputType), "[]") {
		return "", "", true, fmt.Errorf("%s supports arrays in the native subset, got %s", callee, expression.Arguments[0].InferredType)
	}
	span := toIRSpan(path, expression.Span)
	args := []string{input}
	settledArray := ""
	if callee == "Promise.all" || callee == "Promise.allSettled" {
		settled := toIRType(promiseValueType(expression.InferredType))
		shapeName := strings.TrimPrefix(string(settled), "object:")
		if shape, isShape := lookupObjectShape(shapeName, shapes); isShape && isTupleShape(shape) {
			if _, known := shapes[shapeName]; !known {
				shapes[shapeName] = shape
			}
			container := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpObjectNew, Type: settled, Result: container, Callee: shapeName, FieldCount: len(shape.Fields), Span: span})
			args = append(args, container)
		} else if strings.HasSuffix(string(settled), "[]") {
			settledArray = string(settled)
		} else {
			return "", "", true, fmt.Errorf("%s result type %q is not supported in the native subset", callee, expression.InferredType)
		}
	}
	promiseType := toIRType(expression.InferredType)
	if promiseType == "" || promiseType == ir.TypeUnknown {
		promiseType = ir.Type("object:Promise")
	}
	if result == "" {
		result = nextTemp(counter)
	}
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: promiseType, Result: result, Callee: runtimeCallee, Args: args, Value: settledArray, Span: span})
	return result, promiseType, true, nil
}

// promiseValueType returns T for a Promise<T> type string.
func promiseValueType(typ string) string {
	typ = strings.TrimSpace(typ)
	if strings.HasPrefix(typ, "Promise<") && strings.HasSuffix(typ, ">") {
		return strings.TrimSpace(typ[len("Promise<") : len(typ)-1])
	}
	return "unknown"
}
