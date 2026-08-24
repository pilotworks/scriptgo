package lowering

import (
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerIteratorBuiltins(m map[string]BuiltinIntrinsic) {
	m["Iterator.from"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Iterator.from",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			argVal, aType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			elemType := "unknown"
			if aType == ir.TypeNumberArray || aType == "number[]" {
				elemType = "number"
			} else if aType == ir.TypeStringArray || aType == "string[]" {
				elemType = "string"
			} else if aType == ir.TypeBoolArray || aType == "boolean[]" || aType == "bool[]" {
				elemType = "bool"
			} else if aType == ir.TypeBigIntArray || aType == "bigint[]" {
				elemType = "bigint"
			} else if strings.HasSuffix(string(aType), "[]") {
				elemType = strings.TrimSuffix(string(aType), "[]")
			} else if strings.HasPrefix(string(aType), "object:Generator_") {
				elemType = "number"
			}
			retType := ir.Type("object:IteratorObject<" + elemType + ">")
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   retType,
				Result: result,
				Callee: "__iterator.from",
				Args:   []string{argVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, retType, nil
		},
	}
}

func lowerIteratorReceiverMethod(
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
	isIteratorObject := strings.HasPrefix(string(receiverType), "object:IteratorObject") ||
		receiverType == ir.Type("object:IteratorObject")

	if !isIteratorObject {
		return "", "", false, nil
	}

	elemType := ir.TypeNumber
	if strings.HasPrefix(string(receiverType), "object:IteratorObject<") && strings.HasSuffix(string(receiverType), ">") {
		inner := strings.TrimSuffix(strings.TrimPrefix(string(receiverType), "object:IteratorObject<"), ">")
		if inner != "" && inner != "unknown" {
			elemType = toIRType(inner)
		}
	}

	var args []string = []string{receiver}
	for _, arg := range expression.Arguments {
		val, _, err := lowerExpression(path, arg, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		args = append(args, val)
	}

	if result == "" {
		result = nextTemp(counter)
	}

	switch methodName {
	case "map":
		retType := ir.Type("object:IteratorObject")
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   retType,
			Result: result,
			Callee: "__iterator.map",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, true, nil

	case "filter":
		retType := receiverType
		if !strings.HasPrefix(string(retType), "object:IteratorObject") {
			retType = ir.Type("object:IteratorObject<" + string(elemType) + ">")
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   retType,
			Result: result,
			Callee: "__iterator.filter",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, true, nil

	case "take":
		retType := receiverType
		if !strings.HasPrefix(string(retType), "object:IteratorObject") {
			retType = ir.Type("object:IteratorObject<" + string(elemType) + ">")
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   retType,
			Result: result,
			Callee: "__iterator.take",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, true, nil

	case "drop":
		retType := receiverType
		if !strings.HasPrefix(string(retType), "object:IteratorObject") {
			retType = ir.Type("object:IteratorObject<" + string(elemType) + ">")
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   retType,
			Result: result,
			Callee: "__iterator.drop",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, true, nil

	case "flatMap":
		retType := ir.Type("object:IteratorObject")
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   retType,
			Result: result,
			Callee: "__iterator.flat_map",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, true, nil

	case "toArray":
		arrType := ir.Type(string(elemType) + "[]")
		if elemType == ir.TypeNumber {
			arrType = ir.TypeNumberArray
		} else if elemType == ir.TypeString {
			arrType = ir.TypeStringArray
		} else if elemType == ir.TypeBool {
			arrType = ir.TypeBoolArray
		} else if elemType == ir.TypeBigInt {
			arrType = ir.TypeBigIntArray
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   arrType,
			Result: result,
			Callee: "__iterator.to_array",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, arrType, true, nil

	case "forEach":
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeVoid,
			Result: result,
			Callee: "__iterator.for_each",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeVoid, true, nil

	case "reduce":
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   elemType,
			Result: result,
			Callee: "__iterator.reduce",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, elemType, true, nil

	case "some":
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeBool,
			Result: result,
			Callee: "__iterator.some",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeBool, true, nil

	case "every":
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeBool,
			Result: result,
			Callee: "__iterator.every",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeBool, true, nil

	case "find":
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   elemType,
			Result: result,
			Callee: "__iterator.find",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, elemType, true, nil

	case "next", "return", "throw":
		resShapeName := "IteratorResult"
		if _, ok := shapes[resShapeName]; !ok {
			shapes[resShapeName] = ir.ObjectShape{
				Name: resShapeName,
				Span: toIRSpan(path, expression.Span),
				Fields: []ir.Field{
					{Name: "done", Type: ir.TypeBool, Span: toIRSpan(path, expression.Span)},
					{Name: "value", Type: elemType, Span: toIRSpan(path, expression.Span)},
				},
			}
		}
		resType := ir.Type("object:" + resShapeName)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   resType,
			Result: result,
			Callee: "__iterator." + methodName,
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, resType, true, nil
	}

	return "", "", false, nil
}
