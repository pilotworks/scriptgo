package lowering

import (
	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerBufferIntrinsics(m map[string]BuiltinIntrinsic) {
	m["Buffer.alloc"] = BuiltinIntrinsic{
		Category: CategoryNodeGlobal,
		Name:     "Buffer.alloc",
		MinArgs:  1,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			sizeVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			fillVal := nextTemp(call.Counter)
			hasFillVal := nextTemp(call.Counter)
			isStrFillVal := nextTemp(call.Counter)

			if len(call.Expression.Arguments) > 1 {
				fv, fType, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				fillVal = fv
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpConst, Type: ir.TypeBool, Result: hasFillVal, Value: "true", Span: toIRSpan(call.Path, call.Expression.Span),
				})
				if fType == ir.TypeString {
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op: ir.OpConst, Type: ir.TypeBool, Result: isStrFillVal, Value: "true", Span: toIRSpan(call.Path, call.Expression.Span),
					})
				} else {
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op: ir.OpConst, Type: ir.TypeBool, Result: isStrFillVal, Value: "false", Span: toIRSpan(call.Path, call.Expression.Span),
					})
				}
			} else {
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpConst, Type: ir.TypeNumber, Result: fillVal, Value: "0", Span: toIRSpan(call.Path, call.Expression.Span),
				})
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpConst, Type: ir.TypeBool, Result: hasFillVal, Value: "false", Span: toIRSpan(call.Path, call.Expression.Span),
				})
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpConst, Type: ir.TypeBool, Result: isStrFillVal, Value: "false", Span: toIRSpan(call.Path, call.Expression.Span),
				})
			}

			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBuffer,
				Result: result,
				Callee: "__buffer.alloc",
				Args:   []string{sizeVal, fillVal, hasFillVal, isStrFillVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBuffer, nil
		},
	}

	m["Buffer.allocUnsafe"] = BuiltinIntrinsic{
		Category: CategoryNodeGlobal,
		Name:     "Buffer.allocUnsafe",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			sizeVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			fillVal := nextTemp(call.Counter)
			hasFillVal := nextTemp(call.Counter)
			isStrFillVal := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeNumber, Result: fillVal, Value: "0", Span: toIRSpan(call.Path, call.Expression.Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeBool, Result: hasFillVal, Value: "false", Span: toIRSpan(call.Path, call.Expression.Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeBool, Result: isStrFillVal, Value: "false", Span: toIRSpan(call.Path, call.Expression.Span),
			})
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBuffer,
				Result: result,
				Callee: "__buffer.alloc",
				Args:   []string{sizeVal, fillVal, hasFillVal, isStrFillVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBuffer, nil
		},
	}

	m["Buffer.from"] = BuiltinIntrinsic{
		Category: CategoryNodeGlobal,
		Name:     "Buffer.from",
		MinArgs:  1,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			argVal, argType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			if argType == ir.TypeString {
				encVal := nextTemp(call.Counter)
				if len(call.Expression.Arguments) > 1 {
					ev, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
					if err != nil {
						return "", "", err
					}
					encVal = ev
				} else {
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op: ir.OpConst, Type: ir.TypeString, Result: encVal, Value: "utf8", Span: toIRSpan(call.Path, call.Expression.Span),
					})
				}
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeBuffer,
					Result: result,
					Callee: "__buffer.from_string",
					Args:   []string{argVal, encVal},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				return result, ir.TypeBuffer, nil
			}

			callee := "__buffer.from_array"
			if argType == ir.TypeArrayBuffer {
				callee = "__buffer.from_arraybuffer"
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBuffer,
				Result: result,
				Callee: callee,
				Args:   []string{argVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBuffer, nil
		},
	}

	m["Buffer.concat"] = BuiltinIntrinsic{
		Category: CategoryNodeGlobal,
		Name:     "Buffer.concat",
		MinArgs:  1,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			listVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			totLenVal := nextTemp(call.Counter)
			if len(call.Expression.Arguments) > 1 {
				tl, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				totLenVal = tl
			} else {
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpConst, Type: ir.TypeNumber, Result: totLenVal, Value: "-1", Span: toIRSpan(call.Path, call.Expression.Span),
				})
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBuffer,
				Result: result,
				Callee: "__buffer.concat",
				Args:   []string{listVal, totLenVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBuffer, nil
		},
	}

	m["Buffer.isBuffer"] = BuiltinIntrinsic{
		Category: CategoryNodeGlobal,
		Name:     "Buffer.isBuffer",
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
				Type:   ir.TypeBool,
				Result: result,
				Callee: "__buffer.isBuffer",
				Args:   []string{objVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBool, nil
		},
	}

	m["Buffer.byteLength"] = BuiltinIntrinsic{
		Category: CategoryNodeGlobal,
		Name:     "Buffer.byteLength",
		MinArgs:  1,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			strVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			encVal := nextTemp(call.Counter)
			if len(call.Expression.Arguments) > 1 {
				ev, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				encVal = ev
			} else {
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpConst, Type: ir.TypeString, Result: encVal, Value: "utf8", Span: toIRSpan(call.Path, call.Expression.Span),
				})
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: result,
				Callee: "__buffer.byteLength",
				Args:   []string{strVal, encVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeNumber, nil
		},
	}

	m["__scriptgo.bufferAlloc"] = m["Buffer.alloc"]
	m["__scriptgo.bufferAllocUnsafe"] = m["Buffer.allocUnsafe"]
	m["__scriptgo.bufferFromString"] = m["Buffer.from"]
	m["__scriptgo.bufferFromArray"] = m["Buffer.from"]
	m["__scriptgo.bufferFromArrayBuffer"] = BuiltinIntrinsic{
		Category: CategoryNodeGlobal,
		Name:     "__scriptgo.bufferFromArrayBuffer",
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
				Op: ir.OpCall, Type: ir.TypeBuffer, Result: result, Callee: "__buffer.from_arraybuffer", Args: []string{argVal}, Span: toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBuffer, nil
		},
	}
	m["__scriptgo.bufferConcat"] = m["Buffer.concat"]
	m["__scriptgo.bufferIsBuffer"] = m["Buffer.isBuffer"]
	m["__scriptgo.bufferByteLength"] = m["Buffer.byteLength"]
}
