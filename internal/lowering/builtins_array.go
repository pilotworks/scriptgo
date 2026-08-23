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
			var args []string
			var retType ir.Type = ir.TypeNumberArray
			for i, argExpr := range call.Expression.Arguments {
				argVal, aType, err := call.LowerExpression(call.Path, argExpr, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				args = append(args, argVal)
				if i == 0 {
					if strings.HasSuffix(string(aType), "[]") || aType == ir.TypeNumberArray || aType == ir.TypeStringArray || aType == ir.TypeBoolArray || aType == ir.TypeBigIntArray {
						retType = aType
					} else if aType == ir.TypeString {
						retType = ir.TypeStringArray
					}
				}
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
				Args:   args,
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
			var elemType ir.Type = ir.TypeNumber
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
