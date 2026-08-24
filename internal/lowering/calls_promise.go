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
	if callee == "Promise.try" && len(expression.Arguments) > 0 {
		fnVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		if result == "" {
			result = nextTemp(counter)
		}
		promType := ir.Type("object:Promise<unknown>")
		if expression.InferredType != "" && strings.HasPrefix(expression.InferredType, "Promise<") && strings.HasSuffix(expression.InferredType, ">") {
			inner := strings.TrimSuffix(strings.TrimPrefix(expression.InferredType, "Promise<"), ">")
			promType = ir.Type("object:Promise<" + string(toIRType(inner)) + ">")
		} else if strings.Contains(expression.InferredType, "number") {
			promType = ir.Type("object:Promise<number>")
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   promType,
			Result: result,
			Callee: "__async.promise_try",
			Args:   []string{fnVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, promType, true, nil
	}
	if callee == "Promise.resolve" || callee == "Promise.reject" {
		var argVal string
		var argType ir.Type = ir.TypeVoid
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
		if arrExpr.Kind == "array" {
			var resArgs []string
			var resElemType ir.Type = ir.TypeUnknown
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
		return resObj, ir.Type("object:" + shapeName), true, nil
	}
	return "", "", false, nil
}
