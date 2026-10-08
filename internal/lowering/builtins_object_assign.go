package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerObjectAssignIntrinsic(m map[string]BuiltinIntrinsic) {
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
}
