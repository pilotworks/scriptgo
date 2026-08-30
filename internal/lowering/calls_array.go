package lowering

import (
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerTypedArrayReceiverMethod(
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
	if !isTypedArrayType(receiverType) {
		return "", "", false, nil
	}
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
				return "", "", true, err
			}
			beginVal = b
		}
		if len(expression.Arguments) > 1 {
			e, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
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
		return result, receiverType, true, nil
	}
	if methodName == "set" && len(expression.Arguments) > 0 {
		srcVal, srcType, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		offsetVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeNumber, Result: offsetVal, Value: "0", Span: toIRSpan(path, expression.Span),
		})
		if len(expression.Arguments) > 1 {
			off, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			offsetVal = off
		}
		callee := "__typedarray.set"
		if expression.Arguments[0].Kind == "array" || srcType == ir.TypeNumberArray || strings.HasSuffix(string(srcType), "[]") {
			callee = "__typedarray.set_array"
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeVoid,
			Callee: callee,
			Args:   []string{receiver, srcVal, offsetVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return "", ir.TypeVoid, true, nil
	}
	if methodName == "fill" && len(expression.Arguments) > 0 {
		valVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
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
				return "", "", true, err
			}
			startVal = s
		}
		if len(expression.Arguments) > 2 {
			e, _, err := lowerExpression(path, expression.Arguments[2], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
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
		return result, receiverType, true, nil
	}
	return "", "", false, nil
}

func lowerArrayReceiverMethod(
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
	shapeName := strings.TrimPrefix(string(receiverType), "object:")
	isTuple := isTupleShapeName(shapeName)
	isArr := receiverType == ir.TypeNumberArray || receiverType == ir.TypeStringArray || receiverType == ir.TypeBoolArray || receiverType == ir.TypeBigIntArray || strings.HasSuffix(string(receiverType), "[]") || receiverType == "object:Array" || receiverType == "Array" || isTuple
	if !isArr || !isArrayMethod(methodName) {
		return "", "", false, nil
	}
	if isTuple && methodName == "slice" {
		var srcShape ir.ObjectShape
		if s, ok := shapes[shapeName]; ok {
			srcShape = s
		} else if s, ok := registeredShapes[shapeName]; ok {
			srcShape = s
		} else if s, ok := anonymousShapes[shapeName]; ok {
			srcShape = s
		}
		if len(srcShape.Fields) > 0 {
			startIdx := 0
			if len(expression.Arguments) > 0 {
				if n, err := strconv.Atoi(expression.Arguments[0].Text); err == nil {
					startIdx = n
				}
			}
			endIdx := len(srcShape.Fields)
			if len(expression.Arguments) > 1 {
				if n, err := strconv.Atoi(expression.Arguments[1].Text); err == nil {
					endIdx = n
				}
			}
			if startIdx < 0 {
				startIdx = 0
			}
			if endIdx > len(srcShape.Fields) {
				endIdx = len(srcShape.Fields)
			}
			var resFields []ir.Field
			for i := startIdx; i < endIdx; i++ {
				resFields = append(resFields, ir.Field{
					Name: strconv.Itoa(len(resFields)),
					Type: srcShape.Fields[i].Type,
				})
			}
			resShapeName := anonymousShapeName(resFields)
			resShape := ir.ObjectShape{
				Name:   resShapeName,
				Fields: resFields,
			}
			shapes[resShapeName] = resShape
			anonymousShapes[resShapeName] = resShape
			registeredShapes[resShapeName] = resShape
			resType := ir.Type("object:" + resShapeName)
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:         ir.OpObjectNew,
				Type:       resType,
				Result:     result,
				Callee:     resShapeName,
				FieldCount: len(resFields),
				Span:       toIRSpan(path, expression.Span),
			})
			for newIdx, oldIdx := 0, startIdx; oldIdx < endIdx; newIdx, oldIdx = newIdx+1, oldIdx+1 {
				fVal := nextTemp(counter)
				fType := srcShape.Fields[oldIdx].Type
				function.Body = append(function.Body, ir.Instruction{
					Op:         ir.OpFieldGet,
					Type:       fType,
					Result:     fVal,
					Callee:     shapeName,
					Field:      srcShape.Fields[oldIdx].Name,
					FieldIndex: oldIdx,
					Args:       []string{receiver},
					Span:       toIRSpan(path, expression.Span),
				})
				function.Body = append(function.Body, ir.Instruction{
					Op:         ir.OpFieldSet,
					Type:       ir.TypeVoid,
					Callee:     resShapeName,
					Field:      resFields[newIdx].Name,
					FieldIndex: newIdx,
					Args:       []string{result, fVal},
					Span:       toIRSpan(path, expression.Span),
				})
			}
			return result, resType, true, nil
		}
	}
	args := []string{receiver}
	elemType := arrayElementType(receiverType)
	// The native array ABI appends/prepends one value per call. Expand the
	// variadic JavaScript push/unshift operation here while preserving argument
	// evaluation order and returning the length from the final operation.
	if methodName == "push" || methodName == "unshift" {
		if len(expression.Arguments) == 0 {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpCall, Type: ir.TypeNumber, Result: result,
				Callee: "__array.length", Args: []string{receiver},
				Span: toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, true, nil
		}
		lastResult := ""
		for index, argument := range expression.Arguments {
			if argument.Kind == "array" && (argument.InferredType == "" || argument.InferredType == "never[]" || argument.InferredType == "unknown[]") && elemType != "" {
				argument.InferredType = string(elemType)
			}
			value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			callResult := nextTemp(counter)
			if index == len(expression.Arguments)-1 && result != "" {
				callResult = result
			}
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpCall, Type: ir.TypeNumber, Result: callResult,
				Callee: "__array." + methodName, Args: []string{receiver, value},
				Span: toIRSpan(path, expression.Span),
			})
			lastResult = callResult
		}
		return lastResult, ir.TypeNumber, true, nil
	}
	for _, argument := range expression.Arguments {
		value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		args = append(args, value)
	}
	if result == "" {
		result = nextTemp(counter)
	}
	returnType := ir.TypeNumber
	switch methodName {
	case "map", "flatMap":
		if isTuple {
			returnType = tupleCommonElementType(shapeName)
		} else {
			returnType = receiverType
		}
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
		if isTuple {
			returnType = tupleCommonElementType(shapeName)
		} else {
			returnType = receiverType
		}
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
	callee := "__array." + methodName
	if methodName == "flatMap" && len(args) > 1 {
		if cbRet, ok := env[args[1]+".retType"]; ok && cbRet == ir.TypeNumber {
			callee = "__array.flatMap_scalar"
		}
	}
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: callee, Args: args, Span: toIRSpan(path, expression.Span)})
	return result, returnType, true, nil
}
