package lowering

import (
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerPromiseStaticCall(
	path string,
	callee string,
	expression *typescriptgo.SyntaxExpression,
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
	if callee == "Promise.all" && len(expression.Arguments) > 0 {
		arrExpr := expression.Arguments[0]
		if arrExpr.Kind != "array" {
			arrVal, _, err := lowerExpression(path, arrExpr, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			if result == "" {
				result = nextTemp(counter)
			}
			promiseType := toIRType(expression.InferredType)
			if promiseType == "" || promiseType == ir.TypeUnknown {
				promiseType = ir.Type("object:Promise")
			}
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpCall, Type: promiseType, Result: result,
				Callee: "__async.promise_all", Args: []string{arrVal},
				Span: toIRSpan(path, expression.Span),
			})
			return result, promiseType, true, nil
		}
		if arrExpr.Kind == "array" {
			var resArgs []string
			resElemType := ir.TypeUnknown
			for _, elem := range arrExpr.Arguments {
				promVal, promType, err := lowerExpression(path, elem, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
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
			return result, promRetType, true, nil
		} else {
			arrVal, arrType, err := lowerExpression(path, arrExpr, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			elemType := ir.TypeNumber
			if strings.HasSuffix(string(arrType), "[]") {
				inner := strings.TrimSuffix(string(arrType), "[]")
				if strings.HasPrefix(inner, "object:Promise_") {
					elemType = toIRType(strings.TrimPrefix(inner, "object:Promise_"))
				} else if strings.HasPrefix(inner, "object:Promise<") {
					elemType = toIRType(strings.TrimSuffix(strings.TrimPrefix(inner, "object:Promise<"), ">"))
				}
			}
			resArrType := ir.Type(string(elemType) + "[]")
			switch elemType {
			case ir.TypeNumber:
				resArrType = ir.TypeNumberArray
			case ir.TypeString:
				resArrType = ir.TypeStringArray
			}
			lenTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: lenTemp,
				Callee: "__array.length",
				Args:   []string{arrVal},
				Span:   toIRSpan(path, expression.Span),
			})
			resArr := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpArray,
				Type:   resArrType,
				Result: resArr,
				Args:   []string{},
				Span:   toIRSpan(path, expression.Span),
			})
			idxTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeNumber,
				Result: idxTemp,
				Value:  "0",
				Span:   toIRSpan(path, expression.Span),
			})
			env[idxTemp] = ir.TypeNumber

			condFn := ir.Function{Name: "cond", ReturnType: ir.TypeBool}
			cmpRes := nextTemp(counter)
			condFn.Body = append(condFn.Body, ir.Instruction{
				Op:       ir.OpCompare,
				Type:     ir.TypeBool,
				Result:   cmpRes,
				Operator: "<",
				Args:     []string{idxTemp, lenTemp},
				Span:     toIRSpan(path, expression.Span),
			})

			bodyFn := ir.Function{Name: "body", ReturnType: function.ReturnType}
			elemProm := nextTemp(counter)
			bodyFn.Body = append(bodyFn.Body, ir.Instruction{
				Op:     ir.OpIndex,
				Type:   ir.Type("object:Promise"),
				Result: elemProm,
				Args:   []string{arrVal, idxTemp},
				Span:   toIRSpan(path, expression.Span),
			})
			awaitedElem := nextTemp(counter)
			bodyFn.Body = append(bodyFn.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   elemType,
				Result: awaitedElem,
				Callee: "__async.await",
				Args:   []string{elemProm},
				Span:   toIRSpan(path, expression.Span),
			})
			bodyFn.Body = append(bodyFn.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: nextTemp(counter),
				Callee: "__array.push",
				Args:   []string{resArr, awaitedElem},
				Span:   toIRSpan(path, expression.Span),
			})

			stepFn := ir.Function{Name: "step", ReturnType: function.ReturnType}
			oneVal := nextTemp(counter)
			stepFn.Body = append(stepFn.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeNumber,
				Result: oneVal,
				Value:  "1",
				Span:   toIRSpan(path, expression.Span),
			})
			incVal := nextTemp(counter)
			stepFn.Body = append(stepFn.Body, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     ir.TypeNumber,
				Result:   incVal,
				Operator: "+",
				Args:     []string{idxTemp, oneVal},
				Span:     toIRSpan(path, expression.Span),
			})
			stepFn.Body = append(stepFn.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   ir.TypeNumber,
				Result: idxTemp,
				Args:   []string{incVal},
				Span:   toIRSpan(path, expression.Span),
			})

			function.Body = append(function.Body, ir.Instruction{
				Op:   ir.OpWhile,
				Type: ir.TypeVoid,
				Args: []string{cmpRes},
				Cond: condFn.Body,
				Body: bodyFn.Body,
				Step: stepFn.Body,
				Span: toIRSpan(path, expression.Span),
			})

			if result == "" {
				result = nextTemp(counter)
			}
			promRetType := ir.Type("object:Promise<" + string(resArrType) + ">")
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   promRetType,
				Result: result,
				Callee: "__async.promise_resolve",
				Args:   []string{resArr},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, promRetType, true, nil
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
					return "", "", true, err
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
			return result, promRetType, true, nil
		}
	}
	if callee == "Promise.any" && len(expression.Arguments) > 0 {
		arrExpr := expression.Arguments[0]
		if arrExpr.Kind == "array" && len(arrExpr.Arguments) > 0 {
			elem := arrExpr.Arguments[0]
			promVal, promType, err := lowerExpression(path, elem, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
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
			promRetType := ir.Type("object:Promise<" + string(elemType) + ">")
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   promRetType,
				Result: result,
				Callee: "__async.promise_resolve",
				Args:   []string{awaitedVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, promRetType, true, nil
		}
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
