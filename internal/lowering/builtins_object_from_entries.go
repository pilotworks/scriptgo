package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerObjectFromEntriesIntrinsic(m map[string]BuiltinIntrinsic) {
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
}
