package lowering

import (
	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerStructuredCloneIntrinsic(m map[string]BuiltinIntrinsic) {
	m["structuredClone"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "structuredClone",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			argVal, argType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   argType,
				Result: result,
				Callee: "__clone.structured",
				Args:   []string{argVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, argType, nil
		},
	}
}

func registerRegExpSymbolIntrinsics(m map[string]BuiltinIntrinsic) {
	m["RegExp"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "RegExp",
		MinArgs:  0,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			ensureRegExpShape(call.Shapes)
			var patternVal string
			patternType := ir.TypeString
			if len(call.Expression.Arguments) == 0 {
				// RegExp() matches the empty string; its source is "(?:)".
				patternVal = nextTemp(call.Counter)
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpConst, Type: ir.TypeString, Result: patternVal, Value: "(?:)", StringLiteral: true, Span: toIRSpan(call.Path, call.Expression.Span),
				})
			} else {
				value, valueType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				patternVal, patternType = value, valueType
			}
			flagsVal := nextTemp(call.Counter)
			if len(call.Expression.Arguments) > 1 {
				fv, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				flagsVal = fv
			} else {
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpConst, Type: ir.TypeString, Result: flagsVal, Value: "", Span: toIRSpan(call.Path, call.Expression.Span),
				})
			}
			if patternType == ir.TypeString {
				// An invalid pattern or flags throws SyntaxError here, as
				// the RegExp constructor does.
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__regex.validate", Args: []string{patternVal, flagsVal}, Span: toIRSpan(call.Path, call.Expression.Span),
				})
			}
			res := call.Result
			if res == "" {
				res = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpObjectNew, Type: ir.Type("object:RegExp"), Result: res, FieldCount: 3, Span: toIRSpan(call.Path, call.Expression.Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{res, patternVal}, Span: toIRSpan(call.Path, call.Expression.Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{res, flagsVal}, Span: toIRSpan(call.Path, call.Expression.Span),
			})
			zeroVal := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroVal, Value: "0", Span: toIRSpan(call.Path, call.Expression.Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "RegExp", Field: "lastIndex", FieldIndex: 2, Args: []string{res, zeroVal}, Span: toIRSpan(call.Path, call.Expression.Span),
			})
			return res, ir.Type("object:RegExp"), nil
		},
	}
	m["Symbol"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Symbol",
		MinArgs:  0,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			descVal := nextTemp(call.Counter)
			if len(call.Expression.Arguments) > 0 {
				dv, dType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				if dType == ir.TypeVoid {
					// Symbol(undefined) has no description.
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op: ir.OpConst, Type: ir.TypeString, Result: descVal, Value: "undefined", Span: toIRSpan(call.Path, call.Expression.Span),
					})
				} else if dType == ir.TypeString {
					descVal = dv
				} else if dType == ir.TypeNumber {
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op: ir.OpCall, Type: ir.TypeString, Result: descVal, Callee: "__string.fromNumber", Args: []string{dv}, Span: toIRSpan(call.Path, call.Expression.Span),
					})
				} else {
					descVal = dv
				}
			} else {
				// Symbol() has no description (the undefined sentinel).
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpConst, Type: ir.TypeString, Result: descVal, Value: "undefined", Span: toIRSpan(call.Path, call.Expression.Span),
				})
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeSymbol,
				Result: result,
				Callee: "__symbol.create",
				Args:   []string{descVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeSymbol, nil
		},
	}
	m["Symbol.for"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Symbol.for",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			keyVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeSymbol,
				Result: result,
				Callee: "__symbol.for",
				Args:   []string{keyVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeSymbol, nil
		},
	}
	m["Symbol.keyFor"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Symbol.keyFor",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			symVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: result,
				Callee: "__symbol.keyFor",
				Args:   []string{symVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeString, nil
		},
	}
}
