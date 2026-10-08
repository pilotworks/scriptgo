package lowering

import (
	"fmt"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerReflectGetPrototypeOf(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: result,
		Value:  "Object.prototype",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeString, nil
}

func lowerReflectSetPrototypeOf(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: result,
		Value:  "true",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeBool, nil
}

func lowerReflectIsExtensible(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: result,
		Value:  "true",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeBool, nil
}

func lowerReflectPreventExtensions(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: result,
		Value:  "true",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeBool, nil
}

func lowerReflectApply(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) < 3 {
		return "", "", fmt.Errorf("reflect.apply requires 3 arguments")
	}

	fnArg := call.Expression.Arguments[0]
	argsArg := call.Expression.Arguments[2]

	var directArgs []string
	if argsArg != nil && (argsArg.Kind == "array" || argsArg.Kind == "array_literal") {
		for _, el := range argsArg.Arguments {
			elVal, _, err := call.LowerExpression(call.Path, el, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			directArgs = append(directArgs, elVal)
		}
	} else if argsArg != nil {
		arrVal, _, err := call.LowerExpression(call.Path, argsArg, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if err != nil {
			return "", "", err
		}
		for i := 0; i < 4; i++ {
			idxConst := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeNumber,
				Result: idxConst,
				Value:  fmt.Sprintf("%d", i),
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			item := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpIndex,
				Type:   ir.TypeUnknown,
				Result: item,
				Args:   []string{arrVal, idxConst},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			directArgs = append(directArgs, item)
		}
	}

	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	fnName := fnArg.Text
	if sig, exists := call.Signatures[fnName]; exists {
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   sig.ReturnType,
			Result: result,
			Callee: fnName,
			Args:   directArgs,
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return result, sig.ReturnType, nil
	}

	fnVal, _, err := call.LowerExpression(call.Path, fnArg, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpClosureCall,
		Type:   ir.TypeUnknown,
		Result: result,
		Callee: fnVal,
		Args:   directArgs,
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeUnknown, nil
}

func lowerReflectConstruct(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) < 2 {
		return "", "", fmt.Errorf("reflect.construct requires at least 2 arguments")
	}

	targetArg := call.Expression.Arguments[0]
	className := targetArg.Text

	shape, exists := call.Shapes[className]
	if !exists {
		return "", "", fmt.Errorf("reflect.construct: unknown class or constructor %q", className)
	}

	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:         ir.OpObjectNew,
		Type:       ir.Type("object:" + className),
		Result:     result,
		FieldCount: len(shape.Fields),
		Span:       toIRSpan(call.Path, call.Expression.Span),
	})

	for i, field := range shape.Fields {
		defVal := field.Value
		if defVal == "" {
			switch field.Type {
			case ir.TypeNumber:
				defVal = "0"
			case ir.TypeBool:
				defVal = "false"
			case ir.TypeBigInt:
				defVal = "0"
			}
		}
		initTemp := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   field.Type,
			Result: initTemp,
			Value:  defVal,
			Span:   field.Span,
		})
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:         ir.OpFieldSet,
			Type:       ir.TypeVoid,
			Field:      field.Name,
			FieldIndex: i,
			Args:       []string{result, initTemp},
			Span:       field.Span,
		})
	}

	return result, ir.Type("object:" + className), nil
}
