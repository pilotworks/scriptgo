package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerArrayBuiltins(m map[string]BuiltinIntrinsic) {
	m["Array.isArray"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Array.isArray",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			argVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
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
				Callee: "__array.isArray",
				Args:   []string{argVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBool, nil
		},
	}

	m["Array.of"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Array.of",
		MinArgs:  0,
		MaxArgs:  255,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			var args []string
			elemType := ir.TypeNumber
			for _, argExpr := range call.Expression.Arguments {
				argVal, aType, err := call.LowerExpression(call.Path, argExpr, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				args = append(args, argVal)
				if aType != "" {
					elemType = aType
				}
			}
			retType := ir.Type(string(elemType) + "[]")
			if elemType == ir.TypeNumber {
				retType = ir.TypeNumberArray
			} else if elemType == ir.TypeString {
				retType = ir.TypeStringArray
			} else if elemType == ir.TypeBool {
				retType = ir.TypeBoolArray
			} else if elemType == ir.TypeBigInt {
				retType = ir.TypeBigIntArray
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpArray,
				Type:   retType,
				Result: result,
				Args:   args,
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, retType, nil
		},
	}

	m["Array.from"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Array.from",
		MinArgs:  1,
		MaxArgs:  3,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			argVal, aType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}

			if isSetType(aType) {
				elemType := ir.TypeUnknown
				if after, ok := strings.CutPrefix(string(aType), "object:Set__"); ok {
					elemType = toIRType(after)
				} else if call.Expression.Arguments[0] != nil && strings.Contains(call.Expression.Arguments[0].InferredType, "<") {
					inferred := strings.TrimSpace(call.Expression.Arguments[0].InferredType)
					idx := strings.Index(inferred, "<")
					inner := inferred[idx+1 : len(inferred)-1]
					parts := splitTypeArguments(inner)
					if len(parts) >= 1 {
						elemType = toIRType(parts[0])
					}
				}
				valuesRes := nextTemp(call.Counter)
				valuesArrType := ir.Type(string(elemType) + "[]")
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   valuesArrType,
					Result: valuesRes,
					Callee: "__set.values",
					Args:   []string{argVal},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				call.Env[valuesRes] = valuesArrType
				argVal = valuesRes
				aType = valuesArrType
			} else if isMapType(aType) {
				keyType := ir.TypeString
				valType := ir.TypeUnknown
				if after, ok := strings.CutPrefix(string(aType), "object:Map__"); ok {
					parts := strings.Split(after, "_")
					if len(parts) >= 2 {
						keyType = toIRType(parts[0])
						valType = toIRType(strings.Join(parts[1:], "_"))
					}
				}
				var fields []ir.Field
				fields = append(fields, ir.Field{Name: "0", Type: keyType})
				fields = append(fields, ir.Field{Name: "1", Type: valType})
				entryShapeName := anonymousShapeName(fields)
				registerAnonymousShape(entryShapeName, fields)
				elemType := ir.Type("object:" + entryShapeName)
				entriesRes := nextTemp(call.Counter)
				entriesArrType := ir.Type(string(elemType) + "[]")
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   entriesArrType,
					Result: entriesRes,
					Callee: "__map.entries",
					Args:   []string{argVal},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				call.Env[entriesRes] = entriesArrType
				argVal = entriesRes
				aType = entriesArrType
			}

			shapeName := strings.TrimPrefix(string(aType), "object:")
			isIter := false
			if strings.HasPrefix(string(aType), "object:") {
				if _, _, ok := findMethodInHierarchy(shapeName, "next", call.Signatures, classHierarchy); ok {
					isIter = true
				} else if _, _, ok := findMethodInHierarchy(shapeName, "Symbol.iterator", call.Signatures, classHierarchy); ok {
					isIter = true
				}
			}
			if !isIter && (strings.Contains(string(aType), "Generator") || strings.Contains(string(aType), "Iterator")) && !strings.Contains(string(aType), "MapIterator") && !strings.Contains(string(aType), "SetIterator") {
				isIter = true
			}

			if isIter {
				if fnIter, mangledIter, okIter := findMethodInHierarchy(shapeName, "Symbol.iterator", call.Signatures, classHierarchy); okIter {
					iterRes := nextTemp(call.Counter)
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   fnIter.ReturnType,
						Result: iterRes,
						Callee: mangledIter,
						Args:   []string{argVal},
						Span:   toIRSpan(call.Path, call.Expression.Span),
					})
					argVal = iterRes
					aType = fnIter.ReturnType
					shapeName = strings.TrimPrefix(string(aType), "object:")
				}

				nextFn := shapeName + "_next"
				targetNext, hasNext := call.Signatures[nextFn]
				if !hasNext {
					if fn, mangled, ok := findMethodInHierarchy(shapeName, "next", call.Signatures, classHierarchy); ok {
						targetNext = fn
						nextFn = mangled
						hasNext = true
					}
				}

				valType := ir.TypeNumber
				resShapeName := ""
				doneFieldIndex := 0
				valFieldIndex := 1
				if hasNext {
					resShapeName = strings.TrimPrefix(string(targetNext.ReturnType), "object:")
					if resShape, ok := call.Shapes[resShapeName]; ok {
						for idx, f := range resShape.Fields {
							if f.Name == "done" {
								doneFieldIndex = idx
							} else if f.Name == "value" {
								valFieldIndex = idx
								valType = f.Type
							}
						}
					}
				}
				if (valType == ir.TypeNumber || valType == "") && call.Expression.Arguments[0] != nil {
					inferred := strings.TrimSpace(call.Expression.Arguments[0].InferredType)
					if idx := strings.Index(inferred, "<"); idx >= 0 && strings.HasSuffix(inferred, ">") {
						inner := inferred[idx+1 : len(inferred)-1]
						parts := splitTypeArguments(inner)
						if len(parts) > 0 {
							valType = toIRType(parts[0])
						}
					} else if idx := strings.Index(string(aType), "<"); idx >= 0 && strings.HasSuffix(string(aType), ">") {
						inner := string(aType)[idx+1 : len(string(aType))-1]
						parts := splitTypeArguments(inner)
						if len(parts) > 0 {
							valType = toIRType(parts[0])
						}
					}
				}
				if valType == "" {
					valType = ir.TypeUnknown
				}
				retType := ir.Type(string(valType) + "[]")
				if valType == ir.TypeString {
					retType = ir.TypeStringArray
				} else if valType == ir.TypeNumber {
					retType = ir.TypeNumberArray
				} else if valType == ir.TypeBool {
					retType = ir.TypeBoolArray
				} else if valType == ir.TypeBigInt {
					retType = ir.TypeBigIntArray
				}

				arrRes := call.Result
				if arrRes == "" {
					arrRes = nextTemp(call.Counter)
				}
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpArray,
					Type:   retType,
					Result: arrRes,
					Args:   nil,
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				call.Env[arrRes] = retType

				var mapVal string
				if len(call.Expression.Arguments) >= 2 {
					mv, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
					if err != nil {
						return "", "", err
					}
					mapVal = mv
				}

				condFunc := ir.Function{Name: "cond", ReturnType: ir.TypeBool}
				condConst := nextTemp(call.Counter)
				condFunc.Body = append(condFunc.Body, ir.Instruction{
					Op: ir.OpConst, Type: ir.TypeBool, Result: condConst, Value: "true", Span: toIRSpan(call.Path, call.Expression.Span),
				})

				bodyBranch := ir.Function{Name: "body", ReturnType: call.Function.ReturnType}
				resVal := nextTemp(call.Counter)
				retNextType := ir.Type("object:" + resShapeName)
				if hasNext {
					retNextType = targetNext.ReturnType
				}
				bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   retNextType,
					Result: resVal,
					Callee: nextFn,
					Args:   []string{argVal},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})

				doneVal := nextTemp(call.Counter)
				bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
					Op:         ir.OpFieldGet,
					Type:       ir.TypeBool,
					Result:     doneVal,
					Callee:     resShapeName,
					Field:      "done",
					FieldIndex: doneFieldIndex,
					Args:       []string{resVal},
					Span:       toIRSpan(call.Path, call.Expression.Span),
				})

				bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
					Op:   ir.OpIf,
					Type: ir.TypeVoid,
					Args: []string{doneVal},
					Then: []ir.Instruction{
						{Op: ir.OpBreak, Type: ir.TypeVoid, Span: toIRSpan(call.Path, call.Expression.Span)},
					},
					Span: toIRSpan(call.Path, call.Expression.Span),
				})

				valVal := nextTemp(call.Counter)
				bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
					Op:         ir.OpFieldGet,
					Type:       valType,
					Result:     valVal,
					Callee:     resShapeName,
					Field:      "value",
					FieldIndex: valFieldIndex,
					Args:       []string{resVal},
					Span:       toIRSpan(call.Path, call.Expression.Span),
				})

				pushVal := valVal
				if mapVal != "" {
					mappedVal := nextTemp(call.Counter)
					bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   valType,
						Result: mappedVal,
						Callee: mapVal,
						Args:   []string{valVal},
						Span:   toIRSpan(call.Path, call.Expression.Span),
					})
					pushVal = mappedVal
				}

				pushRes := nextTemp(call.Counter)
				bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeNumber,
					Result: pushRes,
					Callee: "__array.push",
					Args:   []string{arrRes, pushVal},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})

				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:   ir.OpWhile,
					Type: ir.TypeVoid,
					Args: []string{condConst},
					Cond: condFunc.Body,
					Body: bodyBranch.Body,
					Span: toIRSpan(call.Path, call.Expression.Span),
				})

				return arrRes, retType, nil
			}

			// Fast-path: Array or String slice
			retType := ir.TypeNumberArray
			if strings.HasSuffix(string(aType), "[]") || aType == ir.TypeNumberArray || aType == ir.TypeStringArray || aType == ir.TypeBoolArray || aType == ir.TypeBigIntArray {
				retType = aType
			} else if aType == ir.TypeString {
				retType = ir.TypeStringArray
			}

			if len(call.Expression.Arguments) >= 2 {
				rawArr := nextTemp(call.Counter)
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   retType,
					Result: rawArr,
					Callee: "__array.from",
					Args:   []string{argVal},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				mapVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				result := call.Result
				if result == "" {
					result = nextTemp(call.Counter)
				}
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   retType,
					Result: result,
					Callee: "__array.map",
					Args:   []string{rawArr, mapVal},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				return result, retType, nil
			}

			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   retType,
				Result: result,
				Callee: "__array.from",
				Args:   []string{argVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, retType, nil
		},
	}

	m["Array.fromAsync"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Array.fromAsync",
		MinArgs:  1,
		MaxArgs:  3,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			var args []string
			elemType := ir.TypeNumber
			for i, argExpr := range call.Expression.Arguments {
				argVal, aType, err := call.LowerExpression(call.Path, argExpr, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				args = append(args, argVal)
				if i == 0 {
					if strings.HasSuffix(string(aType), "[]") {
						elemType = toIRType(strings.TrimSuffix(string(aType), "[]"))
					} else if aType == ir.TypeString {
						elemType = ir.TypeString
					}
				}
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			retType := ir.Type("object:Promise<" + string(elemType) + "[]>")
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   retType,
				Result: result,
				Callee: "__async.array_from_async",
				Args:   args,
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, retType, nil
		},
	}
}
