package lowering

import (
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
		srcVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
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
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeVoid,
			Callee: "__typedarray.set",
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
	isArr := receiverType == ir.TypeNumberArray || receiverType == ir.TypeStringArray || receiverType == ir.TypeBoolArray || receiverType == ir.TypeBigIntArray || strings.HasSuffix(string(receiverType), "[]") || receiverType == "object:Array" || receiverType == "Array"
	if !isArr || !isArrayMethod(methodName) {
		return "", "", false, nil
	}
	args := []string{receiver}
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
		returnType = receiverType
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
		returnType = receiverType
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
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: "__array." + methodName, Args: args, Span: toIRSpan(path, expression.Span)})
	return result, returnType, true, nil
}
