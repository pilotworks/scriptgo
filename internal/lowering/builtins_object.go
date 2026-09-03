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
			if !hasTargetShape && targetType != "" {
				if fields, ok := anonymousObjectFields(targetClass, nil); ok {
					anonName := anonymousShapeName(fields)
					targetShape = ir.ObjectShape{Name: anonName, Fields: fields}
					call.Shapes[anonName] = targetShape
					hasTargetShape = true
				}
			}

			type srcInfo struct {
				val   string
				typ   ir.Type
				shape ir.ObjectShape
			}
			var sources []srcInfo
			for i := 1; i < len(call.Expression.Arguments); i++ {
				sVal, sType, err := call.LowerExpression(call.Path, call.Expression.Arguments[i], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				sClass := strings.TrimPrefix(string(sType), "object:")
				sShape, ok := call.Shapes[sClass]
				if !ok && sType != "" {
					if fields, ok := anonymousObjectFields(sClass, nil); ok {
						anonName := anonymousShapeName(fields)
						sShape = ir.ObjectShape{Name: anonName, Fields: fields}
						call.Shapes[anonName] = sShape
					}
				}
				sources = append(sources, srcInfo{val: sVal, typ: sType, shape: sShape})
			}

			if hasTargetShape {
				for _, src := range sources {
					if len(src.shape.Fields) > 0 {
						for sIdx, srcField := range src.shape.Fields {
							for tIdx, targetField := range targetShape.Fields {
								if targetField.Name == srcField.Name {
									fieldVal := nextTemp(call.Counter)
									call.Function.Body = append(call.Function.Body, ir.Instruction{
										Op:         ir.OpFieldGet,
										Type:       srcField.Type,
										Result:     fieldVal,
										Args:       []string{src.val},
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
			}

			inferredType := call.Expression.InferredType
			mergedClass := strings.TrimPrefix(inferredType, "object:")
			mergedShape, hasMergedShape := call.Shapes[mergedClass]
			if !hasMergedShape && inferredType != "" {
				if fields, ok := anonymousObjectFields(inferredType, nil); ok {
					anonName := anonymousShapeName(fields)
					mergedShape = ir.ObjectShape{Name: anonName, Fields: fields}
					call.Shapes[anonName] = mergedShape
					hasMergedShape = true
				}
			}

			if hasMergedShape && len(mergedShape.Fields) > len(targetShape.Fields) {
				resObj := call.Result
				if resObj == "" {
					resObj = nextTemp(call.Counter)
				}
				var mFieldNames []string
				for _, f := range mergedShape.Fields {
					mFieldNames = append(mFieldNames, f.Name)
				}
				mTag := ":" + strings.Join(mFieldNames, ":") + ":"
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:         ir.OpObjectNew,
					Type:       ir.Type("object:" + mergedShape.Name),
					Result:     resObj,
					Callee:     mergedShape.Name,
					Value:      mTag,
					FieldCount: len(mergedShape.Fields),
					Span:       toIRSpan(call.Path, call.Expression.Span),
				})
				for mIdx, mField := range mergedShape.Fields {
					var foundVal string
					for sIdx := len(sources) - 1; sIdx >= 0; sIdx-- {
						src := sources[sIdx]
						for fIdx, f := range src.shape.Fields {
							if f.Name == mField.Name {
								fVal := nextTemp(call.Counter)
								call.Function.Body = append(call.Function.Body, ir.Instruction{
									Op:         ir.OpFieldGet,
									Type:       f.Type,
									Result:     fVal,
									Callee:     src.shape.Name,
									Args:       []string{src.val},
									Field:      f.Name,
									FieldIndex: fIdx,
									Span:       toIRSpan(call.Path, call.Expression.Span),
								})
								foundVal = fVal
								break
							}
						}
						if foundVal != "" {
							break
						}
					}
					if foundVal == "" && hasTargetShape {
						for fIdx, f := range targetShape.Fields {
							if f.Name == mField.Name {
								fVal := nextTemp(call.Counter)
								call.Function.Body = append(call.Function.Body, ir.Instruction{
									Op:         ir.OpFieldGet,
									Type:       f.Type,
									Result:     fVal,
									Callee:     targetShape.Name,
									Args:       []string{targetVal},
									Field:      f.Name,
									FieldIndex: fIdx,
									Span:       toIRSpan(call.Path, call.Expression.Span),
								})
								foundVal = fVal
								break
							}
						}
					}
					if foundVal != "" {
						call.Function.Body = append(call.Function.Body, ir.Instruction{
							Op:         ir.OpFieldSet,
							Type:       ir.TypeVoid,
							Callee:     mergedShape.Name,
							Args:       []string{resObj, foundVal},
							Field:      mField.Name,
							FieldIndex: mIdx,
							Span:       toIRSpan(call.Path, call.Expression.Span),
						})
					}
				}
				return resObj, ir.Type("object:" + mergedShape.Name), nil
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

	m["Object.fromEntries"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.fromEntries",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			entriesVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			inferredType := call.Expression.InferredType
			if inferredType == "" || inferredType == "any" || inferredType == "{ [k: string]: any; }" || strings.Contains(inferredType, "[k: string]") {
				if declStr, ok := call.Env["__decl_str."+call.Result]; ok && declStr != "" {
					inferredType = string(declStr)
				} else if tgtType, ok := call.Env[call.Result]; ok && strings.HasPrefix(string(tgtType), "object:") {
					inferredType = string(tgtType)
				}
			}
			resClass := strings.TrimPrefix(inferredType, "object:")
			resShape, hasResShape := call.Shapes[resClass]
			if !hasResShape && inferredType != "" {
				if fields, ok := anonymousObjectFields(inferredType, nil); ok {
					anonName := anonymousShapeName(fields)
					resShape = ir.ObjectShape{Name: anonName, Fields: fields}
					call.Shapes[anonName] = resShape
					hasResShape = true
				}
			}
			if !hasResShape && len(call.Expression.Arguments) > 0 {
				arg0 := call.Expression.Arguments[0]
				if arg0.Kind == "array" && len(arg0.Arguments) > 0 {
					var fields []ir.Field
					seen := map[string]bool{}
					for _, elem := range arg0.Arguments {
						if elem.Kind == "array" && len(elem.Arguments) >= 2 {
							keyExpr := elem.Arguments[0]
							valExpr := elem.Arguments[1]
							if keyExpr.Kind == "string" || keyExpr.Kind == "literal" {
								k := strings.Trim(keyExpr.Text, "\"'`")
								if !seen[k] {
									seen[k] = true
									valType := toIRType(valExpr.InferredType)
									if valType == "" {
										valType = ir.TypeUnknown
									}
									fields = append(fields, ir.Field{Name: k, Type: valType})
								}
							}
						}
					}
					if len(fields) > 0 {
						anonName := anonymousShapeName(fields)
						resShape = ir.ObjectShape{Name: anonName, Fields: fields}
						call.Shapes[anonName] = resShape
						hasResShape = true
					}
				}
			}
			if hasResShape && len(resShape.Fields) > 0 {
				resObj := call.Result
				if resObj == "" {
					resObj = nextTemp(call.Counter)
				}
				if call.Result != "" {
					call.Env[call.Result] = ir.Type("object:" + resShape.Name)
				}
				var tagNames []string
				for _, f := range resShape.Fields {
					tagNames = append(tagNames, f.Name)
				}
				typeTag := ":" + strings.Join(tagNames, ":") + ":"
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:         ir.OpObjectNew,
					Type:       ir.Type("object:" + resShape.Name),
					Result:     resObj,
					Callee:     resShape.Name,
					Value:      typeTag,
					FieldCount: len(resShape.Fields),
					Span:       toIRSpan(call.Path, call.Expression.Span),
				})
				lenRes := nextTemp(call.Counter)
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeNumber,
					Result: lenRes,
					Callee: "__array.length",
					Args:   []string{entriesVal},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				idxVar := nextTemp(call.Counter)
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeNumber,
					Result: idxVar,
					Value:  "0",
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				condRes := nextTemp(call.Counter)
				condBlock := []ir.Instruction{
					{
						Op:       ir.OpCompare,
						Type:     ir.TypeBool,
						Result:   condRes,
						Operator: "<",
						Args:     []string{idxVar, lenRes},
						Span:     toIRSpan(call.Path, call.Expression.Span),
					},
				}
				bodyBlock := []ir.Instruction{}
				tupleVal := nextTemp(call.Counter)
				bodyBlock = append(bodyBlock, ir.Instruction{
					Op:     ir.OpIndex,
					Type:   ir.TypePointer,
					Result: tupleVal,
					Args:   []string{entriesVal, idxVar},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				oneIdx := nextTemp(call.Counter)
				bodyBlock = append(bodyBlock, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: oneIdx, Value: "1", Span: toIRSpan(call.Path, call.Expression.Span)})

				keyVal := nextTemp(call.Counter)
				bodyBlock = append(bodyBlock, ir.Instruction{
					Op:         ir.OpFieldGet,
					Type:       ir.TypeString,
					Result:     keyVal,
					Callee:     "",
					Field:      "0",
					FieldIndex: 0,
					Args:       []string{tupleVal},
					Span:       toIRSpan(call.Path, call.Expression.Span),
				})

				for fIdx, f := range resShape.Fields {
					fNameConst := nextTemp(call.Counter)
					bodyBlock = append(bodyBlock, ir.Instruction{
						Op:     ir.OpConst,
						Type:   ir.TypeString,
						Result: fNameConst,
						Value:  f.Name,
						Span:   toIRSpan(call.Path, call.Expression.Span),
					})
					isMatch := nextTemp(call.Counter)
					bodyBlock = append(bodyBlock, ir.Instruction{
						Op:       ir.OpCompare,
						Type:     ir.TypeBool,
						Result:   isMatch,
						Operator: "==",
						Args:     []string{keyVal, fNameConst},
						Span:     toIRSpan(call.Path, call.Expression.Span),
					})
					valGot := nextTemp(call.Counter)
					thenInsts := []ir.Instruction{
						{
							Op:         ir.OpFieldGet,
							Type:       f.Type,
							Result:     valGot,
							Callee:     "",
							Field:      "1",
							FieldIndex: 1,
							Args:       []string{tupleVal},
							Span:       toIRSpan(call.Path, call.Expression.Span),
						},
						{
							Op:         ir.OpFieldSet,
							Type:       ir.TypeVoid,
							Callee:     resShape.Name,
							Args:       []string{resObj, valGot},
							Field:      f.Name,
							FieldIndex: fIdx,
							Span:       toIRSpan(call.Path, call.Expression.Span),
						},
					}
					bodyBlock = append(bodyBlock, ir.Instruction{
						Op:   ir.OpIf,
						Type: ir.TypeVoid,
						Args: []string{isMatch},
						Then: thenInsts,
						Span: toIRSpan(call.Path, call.Expression.Span),
					})
				}

				nextIdx := nextTemp(call.Counter)
				bodyBlock = append(bodyBlock,
					ir.Instruction{Op: ir.OpBinary, Type: ir.TypeNumber, Result: nextIdx, Operator: "+", Args: []string{idxVar, oneIdx}, Span: toIRSpan(call.Path, call.Expression.Span)},
					ir.Instruction{Op: ir.OpAssign, Type: ir.TypeNumber, Result: idxVar, Args: []string{nextIdx}, Span: toIRSpan(call.Path, call.Expression.Span)},
				)

				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:   ir.OpWhile,
					Type: ir.TypeVoid,
					Args: []string{condRes},
					Cond: condBlock,
					Body: bodyBlock,
					Span: toIRSpan(call.Path, call.Expression.Span),
				})
				return resObj, ir.Type("object:" + resShape.Name), nil
			}

			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.Type("object:Record"),
				Result: result,
				Callee: "__object.fromEntries",
				Args:   []string{entriesVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.Type("object:Record"), nil
		},
	}

	m["Object.groupBy"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.groupBy",
		MinArgs:  2,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			itemsVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			cbVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.Type("object:Record"),
				Result: result,
				Callee: "__object.groupBy",
				Args:   []string{itemsVal, cbVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.Type("object:Record"), nil
		},
	}

	registerSimpleObj := func(names []string, callee string, retType ir.Type, minArgs, maxArgs int) {
		for _, name := range names {
			n := name
			m[n] = BuiltinIntrinsic{
				Category: CategoryECMAScript,
				Name:     n,
				MinArgs:  minArgs,
				MaxArgs:  maxArgs,
				Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
					var args []string
					rType := retType
					for i, arg := range call.Expression.Arguments {
						val, typ, err := call.LowerExpression(call.Path, arg, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
						if err != nil {
							return "", "", err
						}
						if i == 0 && retType == "" {
							rType = typ
						}
						args = append(args, val)
					}
					result := call.Result
					if result == "" {
						result = nextTemp(call.Counter)
					}
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   rType,
						Result: result,
						Callee: callee,
						Args:   args,
						Span:   toIRSpan(call.Path, call.Expression.Span),
					})
					return result, rType, nil
				},
			}
		}
	}

	registerSimpleObj([]string{"Object.create"}, "__object.create", ir.TypeObject, 1, 2)
	registerSimpleObj([]string{"Object.freeze"}, "__object.freeze", "", 1, 1)
	registerSimpleObj([]string{"Object.seal"}, "__object.seal", "", 1, 1)
	registerSimpleObj([]string{"Object.preventExtensions"}, "__object.preventExtensions", "", 1, 1)
	registerSimpleObj([]string{"Object.isFrozen"}, "__object.isFrozen", ir.TypeBool, 1, 1)
	registerSimpleObj([]string{"Object.isSealed"}, "__object.isSealed", ir.TypeBool, 1, 1)
	registerSimpleObj([]string{"Object.isExtensible"}, "__object.isExtensible", ir.TypeBool, 1, 1)
	m["Object.getOwnPropertyNames"] = m["Object.keys"]
	m["Object.getOwnPropertySymbols"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.getOwnPropertySymbols",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			_, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:         ir.OpArray,
				Type:       ir.Type("symbol[]"),
				Result:     result,
				FieldCount: 0,
				Span:       toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.Type("symbol[]"), nil
		},
	}
	registerSimpleObj([]string{"Object.getOwnPropertyDescriptor"}, "__object.getOwnPropertyDescriptor", ir.TypeObject, 2, 2)
	registerSimpleObj([]string{"Object.getOwnPropertyDescriptors"}, "__object.getOwnPropertyDescriptors", ir.TypeObject, 1, 1)
	registerSimpleObj([]string{"Object.getPrototypeOf"}, "__object.getPrototypeOf", ir.TypeObject, 1, 1)
	registerSimpleObj([]string{"Object.setPrototypeOf"}, "__object.setPrototypeOf", "", 2, 2)
	registerSimpleObj([]string{"Object.defineProperty"}, "__object.defineProperty", "", 3, 3)
	registerSimpleObj([]string{"Object.defineProperties"}, "__object.defineProperties", "", 2, 2)
}
