package lowering

import (
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerConsoleAssert(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) == 0 {
		failConst := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: failConst,
			Value:  "Assertion failed",
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpPrint,
			Type:   ir.TypeVoid,
			Callee: "console.error",
			Args:   []string{failConst},
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return "", ir.TypeVoid, nil
	}

	condVal, condType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}

	boolCond := condVal
	if condType != ir.TypeBool {
		boolTemp := nextTemp(call.Counter)
		if condType == ir.TypeNumber {
			zeroConst := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeNumber,
				Result: zeroConst,
				Value:  "0",
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:       ir.OpCompare,
				Type:     ir.TypeBool,
				Result:   boolTemp,
				Operator: "!=",
				Args:     []string{condVal, zeroConst},
				Span:     toIRSpan(call.Path, call.Expression.Span),
			})
		} else if condType == ir.TypeString {
			emptyConst := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeString,
				Result: emptyConst,
				Value:  "",
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:       ir.OpCompare,
				Type:     ir.TypeBool,
				Result:   boolTemp,
				Operator: "!=",
				Args:     []string{condVal, emptyConst},
				Span:     toIRSpan(call.Path, call.Expression.Span),
			})
		} else {
			nullConst := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   condType,
				Result: nullConst,
				Value:  "null",
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:       ir.OpCompare,
				Type:     ir.TypeBool,
				Result:   boolTemp,
				Operator: "!=",
				Args:     []string{condVal, nullConst},
				Span:     toIRSpan(call.Path, call.Expression.Span),
			})
		}
		boolCond = boolTemp
	}

	falseConst := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: falseConst,
		Value:  "false",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	notCond := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:       ir.OpCompare,
		Type:     ir.TypeBool,
		Result:   notCond,
		Operator: "==",
		Args:     []string{boolCond, falseConst},
		Span:     toIRSpan(call.Path, call.Expression.Span),
	})

	var msgResult string
	var thenInstructions []ir.Instruction
	if len(call.Expression.Arguments) <= 1 {
		failConst := nextTemp(call.Counter)
		thenInstructions = append(thenInstructions, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: failConst,
			Value:  "Assertion failed",
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		msgResult = failConst
	} else {
		prefixConst := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: prefixConst,
			Value:  "Assertion failed: ",
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		var msgParts []string
		for _, arg := range call.Expression.Arguments[1:] {
			part, err := lowerConsoleArg(call, arg)
			if err != nil {
				return "", "", err
			}
			msgParts = append(msgParts, part)
		}
		spaceConst := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: spaceConst,
			Value:  " ",
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		joined := msgParts[0]
		for i := 1; i < len(msgParts); i++ {
			withSpace := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     ir.TypeString,
				Result:   withSpace,
				Operator: "+",
				Args:     []string{joined, spaceConst},
				Span:     toIRSpan(call.Path, call.Expression.Span),
			})
			comb := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     ir.TypeString,
				Result:   comb,
				Operator: "+",
				Args:     []string{withSpace, msgParts[i]},
				Span:     toIRSpan(call.Path, call.Expression.Span),
			})
			joined = comb
		}
		combAll := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:       ir.OpBinary,
			Type:     ir.TypeString,
			Result:   combAll,
			Operator: "+",
			Args:     []string{prefixConst, joined},
			Span:     toIRSpan(call.Path, call.Expression.Span),
		})
		msgResult = combAll
	}

	thenInstructions = append(thenInstructions, ir.Instruction{
		Op:     ir.OpPrint,
		Type:   ir.TypeVoid,
		Callee: "console.error",
		Args:   []string{msgResult},
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:   ir.OpIf,
		Type: ir.TypeVoid,
		Args: []string{notCond},
		Then: thenInstructions,
		Span: toIRSpan(call.Path, call.Expression.Span),
	})
	return "", ir.TypeVoid, nil
}

func lowerConsoleGroup(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) > 0 {
		if _, _, err := lowerPrint(call, BuiltinIntrinsic{Name: "console.log"}); err != nil {
			return "", "", err
		}
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeVoid,
		Callee: "__console.group",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return "", ir.TypeVoid, nil
}

func lowerConsoleTimeLog(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	var lblArg, dataArg string
	if len(call.Expression.Arguments) > 0 {
		val, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if err != nil {
			return "", "", err
		}
		lblArg = val
	}
	if len(call.Expression.Arguments) > 1 {
		var dataParts []string
		for _, arg := range call.Expression.Arguments[1:] {
			part, err := lowerConsoleArg(call, arg)
			if err != nil {
				return "", "", err
			}
			dataParts = append(dataParts, part)
		}
		spaceConst := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: spaceConst,
			Value:  " ",
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		current := dataParts[0]
		for i := 1; i < len(dataParts); i++ {
			withSpace := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     ir.TypeString,
				Result:   withSpace,
				Operator: "+",
				Args:     []string{current, spaceConst},
				Span:     toIRSpan(call.Path, call.Expression.Span),
			})
			comb := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     ir.TypeString,
				Result:   comb,
				Operator: "+",
				Args:     []string{withSpace, dataParts[i]},
				Span:     toIRSpan(call.Path, call.Expression.Span),
			})
			current = comb
		}
		dataArg = current
	}

	var args []string
	if lblArg != "" {
		args = append(args, lblArg)
	}
	if dataArg != "" {
		args = append(args, dataArg)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeVoid,
		Callee: "__console.timeLog",
		Args:   args,
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return "", ir.TypeVoid, nil
}

func lowerConsoleTrace(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	var args []string
	if len(call.Expression.Arguments) > 0 {
		var parts []string
		for _, arg := range call.Expression.Arguments {
			part, err := lowerConsoleArg(call, arg)
			if err != nil {
				return "", "", err
			}
			parts = append(parts, part)
		}
		spaceConst := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: spaceConst,
			Value:  " ",
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		current := parts[0]
		for i := 1; i < len(parts); i++ {
			withSpace := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     ir.TypeString,
				Result:   withSpace,
				Operator: "+",
				Args:     []string{current, spaceConst},
				Span:     toIRSpan(call.Path, call.Expression.Span),
			})
			comb := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     ir.TypeString,
				Result:   comb,
				Operator: "+",
				Args:     []string{withSpace, parts[i]},
				Span:     toIRSpan(call.Path, call.Expression.Span),
			})
			current = comb
		}
		args = append(args, current)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeVoid,
		Callee: "__console.trace",
		Args:   args,
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return "", ir.TypeVoid, nil
}

func lowerConsoleNoop(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	for _, arg := range call.Expression.Arguments {
		expr := arg
		if arg.Kind == "spread" && arg.Left != nil {
			expr = arg.Left
		}
		if _, _, err := call.LowerExpression(call.Path, expr, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures); err != nil {
			return "", "", err
		}
	}
	return "", ir.TypeVoid, nil
}
