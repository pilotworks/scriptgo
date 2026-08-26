package lowering

import (
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerValueToString(call IntrinsicCall, val string, valType ir.Type, span typescriptgo.SourceSpan) string {
	if valType == ir.TypeString {
		return val
	}
	strTemp := nextTemp(call.Counter)
	var callee string
	switch valType {
	case ir.TypeNumber:
		callee = "__string.fromNumber"
	case ir.TypeBool:
		callee = "__string.fromBool"
	case ir.TypeBigInt:
		callee = "__string.fromBigInt"
	case ir.TypeSymbol:
		callee = "__symbol.keyFor"
	case ir.TypeNumberArray:
		callee = "__json.stringify_number_array"
	case ir.TypeStringArray:
		callee = "__json.stringify_string_array"
	case ir.TypeMap:
		callee = "__map.toString"
	case ir.TypeSet:
		callee = "__set.toString"
	case ir.TypeUnknown:
		callee = "__string.fromUnknown"
	case ir.TypeObject:
		callee = "__string.fromObject"
	default:
		if strings.HasSuffix(string(valType), "[]") {
			callee = "__json.stringify_string_array"
		} else if strings.HasPrefix(string(valType), "object:") {
			callee = "__string.fromObject"
		} else {
			callee = "__string.fromUnknown"
		}
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeString,
		Result: strTemp,
		Callee: callee,
		Args:   []string{val},
		Span:   toIRSpan(call.Path, span),
	})
	return strTemp
}

func lowerConsoleArg(call IntrinsicCall, arg *typescriptgo.SyntaxExpression) (string, error) {
	if arg.Kind == "spread" && arg.Left != nil {
		arrVal, _, err := call.LowerExpression(call.Path, arg.Left, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if err != nil {
			return "", err
		}
		spaceConst := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: spaceConst,
			Value:  " ",
			Span:   toIRSpan(call.Path, arg.Span),
		})
		joinTemp := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeString,
			Result: joinTemp,
			Callee: "__array.join",
			Args:   []string{arrVal, spaceConst},
			Span:   toIRSpan(call.Path, arg.Span),
		})
		return joinTemp, nil
	}
	val, valType, err := call.LowerExpression(call.Path, arg, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", err
	}
	return lowerValueToString(call, val, valType, arg.Span), nil
}

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
		if isTypedArrayType(argType) {
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
		if argType == ir.TypeMap {
			strTemp := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: strTemp,
				Callee: "__map.toString",
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
		if argType == ir.TypeSet {
			strTemp := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: strTemp,
				Callee: "__set.toString",
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
		if argType == ir.TypeNumberArray || argType == ir.TypeStringArray || strings.HasSuffix(string(argType), "[]") {
			strTemp := nextTemp(call.Counter)
			callee := "__json.stringify_number_array"
			if argType == ir.TypeStringArray || strings.HasSuffix(string(argType), "[]") {
				callee = "__json.stringify_string_array"
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: strTemp,
				Callee: callee,
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

func containsFormatSpecifier(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '%' {
			c := s[i+1]
			if c == 's' || c == 'd' || c == 'i' || c == 'f' || c == 'j' || c == '%' {
				return true
			}
		}
	}
	return false
}

func lowerFormatString(call IntrinsicCall, fmtStr string, extraArgs []*typescriptgo.SyntaxExpression) (string, error) {
	var parts []string
	argIdx := 0
	last := 0
	for i := 0; i < len(fmtStr); i++ {
		if fmtStr[i] == '%' && i+1 < len(fmtStr) {
			spec := fmtStr[i+1]
			if spec == 's' || spec == 'd' || spec == 'i' || spec == 'f' || spec == 'j' || spec == '%' {
				if i > last {
					litConst := nextTemp(call.Counter)
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   ir.TypeString,
						Result: litConst,
						Value:  fmtStr[last:i],
						Span:   toIRSpan(call.Path, call.Expression.Span),
					})
					parts = append(parts, litConst)
				}
				i++
				last = i + 1
				if spec == '%' {
					pctConst := nextTemp(call.Counter)
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   ir.TypeString,
						Result: pctConst,
						Value:  "%",
						Span:   toIRSpan(call.Path, call.Expression.Span),
					})
					parts = append(parts, pctConst)
				} else if argIdx < len(extraArgs) {
					arg := extraArgs[argIdx]
					argIdx++
					val, valType, err := call.LowerExpression(call.Path, arg, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
					if err != nil {
						return "", err
					}
					parts = append(parts, lowerValueToString(call, val, valType, arg.Span))
				} else {
					unfilledConst := nextTemp(call.Counter)
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   ir.TypeString,
						Result: unfilledConst,
						Value:  "%" + string(spec),
						Span:   toIRSpan(call.Path, call.Expression.Span),
					})
					parts = append(parts, unfilledConst)
				}
			}
		}
	}
	if last < len(fmtStr) {
		tailConst := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: tailConst,
			Value:  fmtStr[last:],
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		parts = append(parts, tailConst)
	}

	// Any remaining arguments after format specifiers are appended with space
	for ; argIdx < len(extraArgs); argIdx++ {
		arg := extraArgs[argIdx]
		part, err := lowerConsoleArg(call, arg)
		if err != nil {
			return "", err
		}
		spaceConst := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: spaceConst,
			Value:  " ",
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		parts = append(parts, spaceConst, part)
	}

	if len(parts) == 0 {
		emptyConst := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: emptyConst,
			Value:  "",
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return emptyConst, nil
	}

	current := parts[0]
	for i := 1; i < len(parts); i++ {
		comb := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:       ir.OpBinary,
			Type:     ir.TypeString,
			Result:   comb,
			Operator: "+",
			Args:     []string{current, parts[i]},
			Span:     toIRSpan(call.Path, call.Expression.Span),
		})
		current = comb
	}
	return current, nil
}

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
