package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerValueToString(call IntrinsicCall, val string, valType ir.Type, span frontend.SourceSpan) string {
	if valType == ir.TypeString {
		return val
	}
	if valType == ir.TypeBigInt {
		// console formatting inspects bigints with an "n" suffix (16n), unlike
		// String(16n).
		digits := nextTemp(call.Counter)
		suffix := nextTemp(call.Counter)
		inspected := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body,
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: digits, Callee: "__string.fromBigInt", Args: []string{val}, Span: toIRSpan(call.Path, span)},
			ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: suffix, Value: "n", StringLiteral: true, Span: toIRSpan(call.Path, span)},
			ir.Instruction{Op: ir.OpBinary, Type: ir.TypeString, Result: inspected, Operator: "+", Args: []string{digits, suffix}, Span: toIRSpan(call.Path, span)},
		)
		return inspected
	}
	strTemp := nextTemp(call.Counter)
	var callee string
	switch valType {
	case ir.TypeNumber:
		callee = "__string.inspectNumber"
	case ir.TypeBool:
		callee = "__string.fromBool"
	case ir.TypeBigInt:
		callee = "__string.fromBigInt"
	case ir.TypeSymbol:
		callee = "__symbol.keyFor"
	case ir.TypeNumberArray:
		callee = "__string.inspectArray"
	case ir.TypeStringArray:
		callee = "__string.inspectArray"
	case ir.TypeMap:
		callee = "__map.toString"
	case ir.TypeSet:
		callee = "__set.toString"
	case ir.TypeUnknown:
		callee = "__string.fromUnknown"
	case ir.TypeUnknownArray:
		callee = "__string.inspectArray"
	case ir.TypeObject:
		callee = "__string.inspectObject"
	case ir.TypeBuffer:
		callee = "__string.inspectBuffer"
	default:
		if strings.HasSuffix(string(valType), "[]") || valType == ir.TypeUnknownArray {
			callee = "__string.inspectArray"
		} else if strings.HasPrefix(string(valType), "object:") {
			callee = "__string.inspectObject"
		} else {
			callee = "__string.fromUnknown"
		}
	}
	if callee == "__string.fromUnknown" {
		callee = "__console.inspectUnknown"
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeString,
		Result: strTemp,
		Callee: callee,
		Args:   []string{val},
		Span:   toIRSpan(call.Path, span),
	})
	switch callee {
	case "__string.inspectArray", "__string.inspectObject", "__map.toString", "__set.toString":
		// Containers print over several lines when they do not fit on one.
		laidOut := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: laidOut, Callee: "__console.layout", Args: []string{strTemp}, Span: toIRSpan(call.Path, span)})
		return laidOut
	}
	return strTemp
}

func lowerConsoleArg(call IntrinsicCall, arg *frontend.SyntaxExpression) (string, error) {
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

func containsFormatSpecifier(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '%' {
			c := s[i+1]
			if c == 's' || c == 'd' || c == 'i' || c == 'f' || c == 'j' || c == 'o' || c == 'O' || c == '%' {
				return true
			}
		}
	}
	return false
}

func lowerFormatString(call IntrinsicCall, fmtStr string, extraArgs []*frontend.SyntaxExpression) (string, error) {
	var parts []string
	argIdx := 0
	last := 0
	for i := 0; i < len(fmtStr); i++ {
		if fmtStr[i] == '%' && i+1 < len(fmtStr) {
			spec := fmtStr[i+1]
			if spec == 's' || spec == 'd' || spec == 'i' || spec == 'f' || spec == 'j' || spec == 'o' || spec == 'O' || spec == '%' {
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
