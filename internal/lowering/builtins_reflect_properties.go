package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerReflectGet(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) < 2 {
		return "", "", fmt.Errorf("reflect.get requires at least 2 arguments")
	}

	targetVal, targetType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	propArg := call.Expression.Arguments[1]
	propName := propArg.Text

	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	if after, ok := strings.CutPrefix(string(targetType), "object:"); ok {
		className := after
		if shape, exists := call.Shapes[className]; exists {
			for i, f := range shape.Fields {
				if f.Name == propName {
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op:         ir.OpFieldGet,
						Type:       f.Type,
						Result:     result,
						Args:       []string{targetVal},
						Field:      f.Name,
						FieldIndex: i,
						Span:       toIRSpan(call.Path, call.Expression.Span),
					})
					return result, f.Type, nil
				}
			}
		}
	}

	// Fallback to dynamic property get
	propVal, _, err := call.LowerExpression(call.Path, propArg, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeUnknown,
		Result: result,
		Callee: "__object.get_prop",
		Args:   []string{targetVal, propVal},
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeUnknown, nil
}

func lowerReflectHas(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) < 2 {
		return "", "", fmt.Errorf("reflect.has requires 2 arguments")
	}

	targetVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	propArg := call.Expression.Arguments[1]
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}

	if propArg.Kind == "string" || propArg.Kind == "literal" {
		fieldName := strings.Trim(propArg.Text, "\"'`")
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpInstanceOf,
			Type:   ir.TypeBool,
			Result: result,
			Value:  fieldName,
			Args:   []string{targetVal},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return result, ir.TypeBool, nil
	}

	propVal, _, err := call.LowerExpression(call.Path, propArg, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpInstanceOf,
		Type:   ir.TypeBool,
		Result: result,
		Value:  "",
		Args:   []string{targetVal, propVal},
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeBool, nil
}

func lowerReflectOwnKeys(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) < 1 {
		return "", "", fmt.Errorf("Reflect.ownKeys requires 1 argument")
	}

	targetVal, targetType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Env[result] = ir.TypeStringArray

	if after, ok := strings.CutPrefix(string(targetType), "object:"); ok {
		className := after
		if shape, exists := call.Shapes[className]; exists {
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:         ir.OpArray,
				Type:       ir.TypeStringArray,
				Result:     result,
				FieldCount: len(shape.Fields),
				Span:       toIRSpan(call.Path, call.Expression.Span),
			})
			for _, f := range shape.Fields {
				constName := nextTemp(call.Counter)
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeString,
					Result: constName,
					Value:  f.Name,
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				pushRes := nextTemp(call.Counter)
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeNumber,
					Result: pushRes,
					Callee: "__array.push",
					Args:   []string{result, constName},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
			}
			return result, ir.TypeStringArray, nil
		}
	}

	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeStringArray,
		Result: result,
		Callee: "__object.keys",
		Args:   []string{targetVal},
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeStringArray, nil
}
