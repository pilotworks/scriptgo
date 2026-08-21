package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerObjectIntrinsics(m map[string]BuiltinIntrinsic) {
	m["Object.is"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.is",
		MinArgs:  2,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			v1, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			v2, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBool,
				Result: result,
				Callee: "__object.is",
				Args:   []string{v1, v2},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBool, nil
		},
	}

	m["Object.hasOwn"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.hasOwn",
		MinArgs:  2,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			objVal, objType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}

			// Static check if object shape and property is a string literal
			if after, ok := strings.CutPrefix(string(objType), "object:"); ok {
				className := after
				shape, exists := call.Shapes[className]
				if exists && call.Expression.Arguments[1] != nil && call.Expression.Arguments[1].Kind == "string" {
					propName := call.Expression.Arguments[1].Text
					hasProp := false
					for _, f := range shape.Fields {
						if f.Name == propName {
							hasProp = true
							break
						}
					}
					valStr := "false"
					if hasProp {
						valStr = "true"
					}
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   ir.TypeBool,
						Result: result,
						Value:  valStr,
						Span:   toIRSpan(call.Path, call.Expression.Span),
					})
					return result, ir.TypeBool, nil
				}
			}

			propVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBool,
				Result: result,
				Callee: "__object.hasOwn",
				Args:   []string{objVal, propVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBool, nil
		},
	}

	m["Object.keys"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.keys",
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

	m["Object.assign"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.assign",
		MinArgs:  1,
		MaxArgs:  64,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			targetVal, targetType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			targetClass := strings.TrimPrefix(string(targetType), "object:")
			targetShape, hasTargetShape := call.Shapes[targetClass]

			for i := 1; i < len(call.Expression.Arguments); i++ {
				srcVal, srcType, err := call.LowerExpression(call.Path, call.Expression.Arguments[i], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				srcClass := strings.TrimPrefix(string(srcType), "object:")
				srcShape, hasSrcShape := call.Shapes[srcClass]
				if hasTargetShape && hasSrcShape {
					for sIdx, srcField := range srcShape.Fields {
						for tIdx, targetField := range targetShape.Fields {
							if targetField.Name == srcField.Name {
								fieldVal := nextTemp(call.Counter)
								call.Function.Body = append(call.Function.Body, ir.Instruction{
									Op:         ir.OpFieldGet,
									Type:       srcField.Type,
									Result:     fieldVal,
									Args:       []string{srcVal},
									Field:      srcField.Name,
									FieldIndex: sIdx,
									Span:       toIRSpan(call.Path, call.Expression.Span),
								})
								call.Function.Body = append(call.Function.Body, ir.Instruction{
									Op:         ir.OpFieldSet,
									Type:       ir.TypeVoid,
									Args:       []string{targetVal, fieldVal},
									Field:      targetField.Name,
									FieldIndex: tIdx,
									Span:       toIRSpan(call.Path, call.Expression.Span),
								})
								break
							}
						}
					}
				}
			}
			result := call.Result
			if result == "" {
				result = targetVal
			} else if result != targetVal {
				trueConst := nextTemp(call.Counter)
				call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: trueConst, Value: "true", Span: toIRSpan(call.Path, call.Expression.Span)})
				call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpSelect, Type: targetType, Result: result, Args: []string{trueConst, targetVal, targetVal}, Span: toIRSpan(call.Path, call.Expression.Span)})
			}
			return result, targetType, nil
		},
	}
}
