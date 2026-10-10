package lowering

import (
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerPrint(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) == 0 {
		emptyConst := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: emptyConst,
			Value:  "",
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpPrint,
			Type:   ir.TypeVoid,
			Callee: intrinsic.Name,
			Args:   []string{emptyConst},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return "", ir.TypeVoid, nil
	}

	if len(call.Expression.Arguments) == 1 {
		if call.Expression.Arguments[0].Kind == "spread" {
			strVal, err := lowerConsoleArg(call, call.Expression.Arguments[0])
			if err != nil {
				return "", "", err
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpPrint,
				Type:   ir.TypeVoid,
				Callee: intrinsic.Name,
				Args:   []string{strVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return "", ir.TypeVoid, nil
		}
		argVal, argType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if err != nil {
			return "", "", err
		}
		if isTypedArrayType(argType) && argType != ir.TypeBuffer {
			strTemp := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: strTemp,
				Callee: "__typedarray.toString",
				Args:   []string{argVal},
				Span:   toIRSpan(call.Path, call.Expression.Arguments[0].Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpPrint,
				Type:   ir.TypeVoid,
				Callee: intrinsic.Name,
				Args:   []string{strTemp},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return "", ir.TypeVoid, nil
		}
		if argType == ir.TypeDataView {
			strTemp := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: strTemp,
				Callee: "__dataview.toString",
				Args:   []string{argVal},
				Span:   toIRSpan(call.Path, call.Expression.Arguments[0].Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpPrint,
				Type:   ir.TypeVoid,
				Callee: intrinsic.Name,
				Args:   []string{strTemp},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return "", ir.TypeVoid, nil
		}
		if argType == ir.TypeArrayBuffer {
			strTemp := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: strTemp,
				Callee: "__arraybuffer.toString",
				Args:   []string{argVal},
				Span:   toIRSpan(call.Path, call.Expression.Arguments[0].Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpPrint,
				Type:   ir.TypeVoid,
				Callee: intrinsic.Name,
				Args:   []string{strTemp},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return "", ir.TypeVoid, nil
		}
		if argType == ir.TypeMap || argType == ir.TypeSet {
			// Maps and sets print their laid-out inspection.
			strTemp := lowerValueToString(call, argVal, argType, call.Expression.Arguments[0].Span)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpPrint,
				Type:   ir.TypeVoid,
				Callee: intrinsic.Name,
				Args:   []string{strTemp},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return "", ir.TypeVoid, nil
		}
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpPrint,
			Type:   ir.TypeVoid,
			Callee: intrinsic.Name,
			Args:   []string{argVal},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return "", ir.TypeVoid, nil
	}

	// Check for format string support: if args[0] is string literal with %s, %d, %i, %f, %j, %%
	if call.Expression.Arguments[0].Kind == "string" {
		fmtStr := call.Expression.Arguments[0].Text
		if containsFormatSpecifier(fmtStr) {
			formattedStr, err := lowerFormatString(call, fmtStr, call.Expression.Arguments[1:])
			if err == nil && formattedStr != "" {
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpPrint,
					Type:   ir.TypeVoid,
					Callee: intrinsic.Name,
					Args:   []string{formattedStr},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				return "", ir.TypeVoid, nil
			}
		}
	}

	var strParts []string
	for _, arg := range call.Expression.Arguments {
		part, err := lowerConsoleArg(call, arg)
		if err != nil {
			return "", "", err
		}
		strParts = append(strParts, part)
	}

	spaceTemp := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: spaceTemp,
		Value:  " ",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	current := strParts[0]
	for i := 1; i < len(strParts); i++ {
		withSpace := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:       ir.OpBinary,
			Type:     ir.TypeString,
			Result:   withSpace,
			Operator: "+",
			Args:     []string{current, spaceTemp},
			Span:     toIRSpan(call.Path, call.Expression.Span),
		})
		combined := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:       ir.OpBinary,
			Type:     ir.TypeString,
			Result:   combined,
			Operator: "+",
			Args:     []string{withSpace, strParts[i]},
			Span:     toIRSpan(call.Path, call.Expression.Span),
		})
		current = combined
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpPrint,
		Type:   ir.TypeVoid,
		Callee: intrinsic.Name,
		Args:   []string{current},
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return "", ir.TypeVoid, nil
}
