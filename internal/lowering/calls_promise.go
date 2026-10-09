package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerPromiseStaticCall(
	path string,
	callee string,
	expression *frontend.SyntaxExpression,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, bool, error) {
	if callee == "Promise.resolve" || callee == "Promise.reject" {
		var argVal string
		argType := ir.TypeVoid
		if len(expression.Arguments) > 0 {
			v, t, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			argVal = v
			argType = t
		}
		if result == "" {
			result = nextTemp(counter)
		}
		promType := ir.Type("object:Promise<" + string(argType) + ">")
		calleeName := "__async.promise_resolve"
		if callee == "Promise.reject" {
			calleeName = "__async.promise_reject"
		}
		args := []string{}
		if argVal != "" {
			args = append(args, argVal)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   promType,
			Result: result,
			Callee: calleeName,
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, promType, true, nil
	}
	if value, typ, handled, err := lowerPromiseCombinator(path, expression, callee, result, function, env, counter, shapes, signatures); handled {
		return value, typ, true, err
	}
	if callee == "Promise.withResolvers" {
		promiseType := ir.Type("object:Promise")
		inferred := strings.TrimPrefix(expression.InferredType, "object:")
		if strings.HasPrefix(inferred, "PromiseWithResolvers<") && strings.HasSuffix(inferred, ">") {
			inner := strings.TrimSuffix(strings.TrimPrefix(inferred, "PromiseWithResolvers<"), ">")
			if resolved := toIRType("Promise<" + inner + ">"); resolved != "" {
				promiseType = resolved
			}
		}
		promRes := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   promiseType,
			Result: promRes,
			Callee: "__async.promise_create",
			Args:   nil,
			Span:   toIRSpan(path, expression.Span),
		})
		resolveRes := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeClosure,
			Result: resolveRes,
			Value:  "resolve",
			Callee: "__async.promise_resolver",
			Args:   []string{promRes},
			Span:   toIRSpan(path, expression.Span),
		})
		rejectRes := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeClosure,
			Result: rejectRes,
			Value:  "reject",
			Callee: "__async.promise_resolver",
			Args:   []string{promRes},
			Span:   toIRSpan(path, expression.Span),
		})
		fields := []ir.Field{
			{Name: "promise", Type: promiseType, Span: toIRSpan(path, expression.Span)},
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
		function.Body = append(function.Body,
			ir.Instruction{
				Op:         ir.OpFieldSet,
				Type:       ir.TypeVoid,
				Callee:     shapeName,
				Field:      "resolve",
				FieldIndex: 1,
				Args:       []string{resObj, resolveRes},
				Span:       toIRSpan(path, expression.Span),
			},
			ir.Instruction{
				Op:         ir.OpFieldSet,
				Type:       ir.TypeVoid,
				Callee:     shapeName,
				Field:      "reject",
				FieldIndex: 2,
				Args:       []string{resObj, rejectRes},
				Span:       toIRSpan(path, expression.Span),
			},
		)
		return resObj, ir.Type("object:" + shapeName), true, nil
	}
	return "", "", false, nil
}
