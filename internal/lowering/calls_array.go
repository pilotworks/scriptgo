package lowering

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerTypedArrayReceiverMethod(
	path string,
	expression *frontend.SyntaxExpression,
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
	if methodName == "join" || methodName == "toString" {
		separator := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeString, Result: separator, Value: ",", StringLiteral: true, Span: toIRSpan(path, expression.Span),
		})
		if methodName == "join" && len(expression.Arguments) > 0 && expression.Arguments[0].Kind != "undefined" {
			value, valueType, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			if valueType != ir.TypeString {
				return "", "", true, fmt.Errorf("TypedArray.prototype.join separator must be a string")
			}
			separator = value
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__typedarray.join",
			Args: []string{receiver, separator}, Span: toIRSpan(path, expression.Span),
		})
		return result, ir.TypeString, true, nil
	}
	return "", "", false, nil
}

func lowerArrayReceiverMethod(
	path string,
	expression *frontend.SyntaxExpression,
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
	if !isTuple {
		if value, typ, handled, err := lowerArrayReduce(path, expression, receiver, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
			return value, typ, true, err
		}
		if value, typ, handled, err := lowerMismatchedArraySearch(path, expression, methodName, receiverType, result, function, env, counter, shapes, signatures); handled {
			return value, typ, true, err
		}
	}
	if methodName == "toSpliced" && len(expression.Arguments) > 2 && !isTuple {
		value, typ, err := lowerToSplicedWithItems(path, expression, receiver, receiverType, result, function, env, counter, shapes, signatures)
		return value, typ, true, err
	}
	if methodName == "toSpliced" && len(expression.Arguments) == 0 {
		// toSpliced() with no start skips nothing: it is a copy, the same as
		// toSpliced(0, 0). (A start without deleteCount removes the rest.)
		zero := func() *frontend.SyntaxExpression {
			return &frontend.SyntaxExpression{Span: expression.Span, Kind: "number", Text: "0", InferredType: "number"}
		}
		copied := *expression
		copied.Arguments = []*frontend.SyntaxExpression{zero(), zero()}
		expression = &copied
	}
	if methodName == "fill" && len(expression.Arguments) == 0 {
		// fill() fills with undefined: the omitted value argument is undefined,
		// which unboxed number storage represents as NaN.
		value := &frontend.SyntaxExpression{Span: expression.Span, Kind: "undefined", Text: "undefined", InferredType: "undefined"}
		if receiverType == ir.TypeNumberArray {
			value = &frontend.SyntaxExpression{Span: expression.Span, Kind: "identifier", Text: "NaN", InferredType: "number"}
		}
		withValue := *expression
		withValue.Arguments = []*frontend.SyntaxExpression{value}
		expression = &withValue
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
			// A tuple slice has a static result shape, so its bounds must be
			// integer literals; they are relative like Array.prototype.slice.
			length := len(srcShape.Fields)
			startIdx, endIdx := 0, length
			if len(expression.Arguments) > 0 {
				n, ok := staticIntegerArgument(expression.Arguments[0])
				if !ok {
					return "", "", true, fmt.Errorf("tuple slice start must be an integer literal")
				}
				startIdx = relativeTupleIndex(n, length)
			}
			if len(expression.Arguments) > 1 {
				n, ok := staticIntegerArgument(expression.Arguments[1])
				if !ok {
					return "", "", true, fmt.Errorf("tuple slice end must be an integer literal")
				}
				endIdx = relativeTupleIndex(n, length)
			}
			if endIdx < startIdx {
				endIdx = startIdx
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
			if argument.Kind == "spread" {
				spreadVal, spreadType, err := lowerExpression(path, argument.Left, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
				}
				idxVar := nextTemp(counter)
				lenVar := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: idxVar, Value: "0", Span: toIRSpan(path, argument.Span)})
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: lenVar, Callee: "__array.length", Args: []string{spreadVal}, Span: toIRSpan(path, argument.Span)})
				condVar := nextTemp(counter)
				var condBody []ir.Instruction
				condBody = append(condBody, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: condVar, Operator: "<", Args: []string{idxVar, lenVar}, Span: toIRSpan(path, argument.Span)})
				var loopBody []ir.Instruction
				itemVar := nextTemp(counter)
				itemType := elemType
				if itemType == "" && spreadType != "" && strings.HasSuffix(string(spreadType), "[]") {
					itemType = arrayElementType(spreadType)
				}
				loopBody = append(loopBody, ir.Instruction{Op: ir.OpIndex, Type: itemType, Result: itemVar, Args: []string{spreadVal, idxVar}, Span: toIRSpan(path, argument.Span)})
				pushRes := nextTemp(counter)
				if index == len(expression.Arguments)-1 && result != "" {
					pushRes = result
				}
				function.Body = append(function.Body, ir.Instruction{
					Op: ir.OpCall, Type: ir.TypeNumber, Result: pushRes,
					Callee: "__array.length", Args: []string{receiver},
					Span: toIRSpan(path, argument.Span),
				})
				loopBody = append(loopBody, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: pushRes, Callee: "__array." + methodName, Args: []string{receiver, itemVar}, Span: toIRSpan(path, argument.Span)})
				oneConst := nextTemp(counter)
				loopBody = append(loopBody, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: oneConst, Value: "1", Span: toIRSpan(path, argument.Span)})
				stepVar := nextTemp(counter)
				loopBody = append(loopBody, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeNumber, Result: stepVar, Operator: "+", Args: []string{idxVar, oneConst}, Span: toIRSpan(path, argument.Span)})
				loopBody = append(loopBody, ir.Instruction{Op: ir.OpAssign, Type: ir.TypeNumber, Result: idxVar, Args: []string{stepVar}, Span: toIRSpan(path, argument.Span)})
				function.Body = append(function.Body, ir.Instruction{
					Op:   ir.OpWhile,
					Type: ir.TypeVoid,
					Cond: condBody,
					Args: []string{condVar},
					Body: loopBody,
					Span: toIRSpan(path, argument.Span),
				})
				lastResult = pushRes
				continue
			}
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
	for index, argument := range expression.Arguments {
		value, typ, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		value, _, present := coerceMethodArgument(path, argument, "array", methodName, index, len(expression.Arguments), value, typ, function, counter)
		if present {
			args = append(args, value)
		}
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
	case "find", "findLast":
		// T | undefined: the element when the predicate matches, else undefined.
		returnType = ir.TypeUnknown
	case "pop", "shift", "at":
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
