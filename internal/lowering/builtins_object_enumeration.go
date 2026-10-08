package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerObjectEnumerationIntrinsics(m map[string]BuiltinIntrinsic) {
	m["Object.keys"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.keys",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			objVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}

			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeStringArray,
				Result: result,
				Callee: "__object.keys",
				Args:   []string{objVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeStringArray, nil
		},
	}

	m["Object.values"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.values",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			objVal, objType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}

			if after, ok := strings.CutPrefix(string(objType), "object:"); ok {
				className := after
				shape, exists := call.Shapes[className]
				if exists {
					elemType := ir.TypeNumber
					if len(shape.Fields) > 0 {
						elemType = shape.Fields[0].Type
					}
					arrayType := ir.TypeNumberArray
					if elemType == ir.TypeString {
						arrayType = ir.TypeStringArray
					}
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op:         ir.OpArray,
						Type:       arrayType,
						Result:     result,
						FieldCount: len(shape.Fields),
						Span:       toIRSpan(call.Path, call.Expression.Span),
					})
					for i, f := range shape.Fields {
						fieldVal := nextTemp(call.Counter)
						call.Function.Body = append(call.Function.Body, ir.Instruction{
							Op:         ir.OpFieldGet,
							Type:       f.Type,
							Result:     fieldVal,
							Args:       []string{objVal},
							Field:      f.Name,
							FieldIndex: i,
							Span:       toIRSpan(call.Path, call.Expression.Span),
						})
						pushRes := nextTemp(call.Counter)
						call.Function.Body = append(call.Function.Body, ir.Instruction{
							Op:     ir.OpCall,
							Type:   ir.TypeNumber,
							Result: pushRes,
							Callee: "__array.push",
							Args:   []string{result, fieldVal},
							Span:   toIRSpan(call.Path, call.Expression.Span),
						})
					}
					return result, arrayType, nil
				}
			}

			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumberArray,
				Result: result,
				Callee: "__object.values",
				Args:   []string{objVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeNumberArray, nil
		},
	}

	m["Object.entries"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.entries",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			objVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.Type("[string,unknown][]"),
				Result: result,
				Callee: "__object.entries",
				Args:   []string{objVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.Type("[string,unknown][]"), nil
		},
	}
}
