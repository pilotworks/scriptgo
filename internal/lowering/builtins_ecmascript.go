package lowering

import (
	"fmt"
)

import (
	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerDateIntrinsics(m map[string]BuiltinIntrinsic) {
	// Date globals (Category 1: ECMAScript)
	registerCallIntrinsic(m, []string{"Date.now", "__date.now"}, CategoryECMAScript, "__date.now", nil, ir.TypeNumber, 0, 0)
	registerCallIntrinsic(m, []string{"Date.parse", "__date.parse"}, CategoryECMAScript, "__date.parse", []ir.Type{ir.TypeString}, ir.TypeNumber, 1, 1)
	m["Date.UTC"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Date.UTC",
		MinArgs:  1,
		MaxArgs:  7,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			var argVals []string
			for _, arg := range call.Expression.Arguments {
				v, _, err := call.LowerExpression(call.Path, arg, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				argVals = append(argVals, v)
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: result,
				Callee: "__date.UTC",
				Args:   argVals,
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeNumber, nil
		},
	}
	m["__date.UTC"] = m["Date.UTC"]
}

func registerBigIntIntrinsics(m map[string]BuiltinIntrinsic) {
	m["BigInt"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "BigInt",
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
			if argType == ir.TypeNumber {
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpCall, Type: ir.TypeBigInt, Result: result, Callee: "__bigint.fromNumber", Args: []string{argVal}, Span: toIRSpan(call.Path, call.Expression.Span),
				})
				return result, ir.TypeBigInt, nil
			}
			if argType == ir.TypeString {
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpCall, Type: ir.TypeBigInt, Result: result, Callee: "__bigint.fromString", Args: []string{argVal}, Span: toIRSpan(call.Path, call.Expression.Span),
				})
				return result, ir.TypeBigInt, nil
			}
			if argType == ir.TypeBigInt {
				return argVal, ir.TypeBigInt, nil
			}
			return "", "", fmt.Errorf("BigInt does not support %s", argType)
		},
	}
	m["BigInt.asIntN"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "BigInt.asIntN",
		MinArgs:  2,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			bitsVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			intVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBigInt,
				Result: result,
				Callee: "__bigint.asIntN",
				Args:   []string{bitsVal, intVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBigInt, nil
		},
	}
	m["BigInt.asUintN"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "BigInt.asUintN",
		MinArgs:  2,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			bitsVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			intVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[1], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBigInt,
				Result: result,
				Callee: "__bigint.asUintN",
				Args:   []string{bitsVal, intVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBigInt, nil
		},
	}
}

func registerNumberStringIntrinsics(m map[string]BuiltinIntrinsic) {
	// Number & Global functions (Category 1: ECMAScript)
	registerCallIntrinsic(m, []string{"parseInt", "Number.parseInt"}, CategoryECMAScript, "__number.parseInt", []ir.Type{ir.TypeString, ir.TypeNumber}, ir.TypeNumber, 1, 2)
	registerCallIntrinsic(m, []string{"parseFloat", "Number.parseFloat"}, CategoryECMAScript, "__number.parseFloat", []ir.Type{ir.TypeString}, ir.TypeNumber, 1, 1)
	registerCallIntrinsic(m, []string{"isNaN", "Number.isNaN"}, CategoryECMAScript, "__number.isNaN", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	registerCallIntrinsic(m, []string{"isFinite", "Number.isFinite"}, CategoryECMAScript, "__number.isFinite", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	registerCallIntrinsic(m, []string{"Number.isInteger"}, CategoryECMAScript, "__number.isInteger", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	registerCallIntrinsic(m, []string{"Number.isSafeInteger"}, CategoryECMAScript, "__number.isSafeInteger", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	registerCallIntrinsic(m, []string{"String.fromCodePoint"}, CategoryECMAScript, "__string.fromCodePoint", []ir.Type{ir.TypeNumber}, ir.TypeString, 1, 1)
	registerCallIntrinsic(m, []string{"String.fromCharCode"}, CategoryECMAScript, "__string.fromCharCode", []ir.Type{ir.TypeNumber}, ir.TypeString, 0, -1)
	registerCallIntrinsic(m, []string{"String.raw"}, CategoryECMAScript, "__string.raw", nil, ir.TypeString, 1, -1)
	registerCallIntrinsic(m, []string{"encodeURIComponent"}, CategoryECMAScript, "__string.encodeURIComponent", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	registerCallIntrinsic(m, []string{"decodeURIComponent"}, CategoryECMAScript, "__string.decodeURIComponent", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	registerCallIntrinsic(m, []string{"encodeURI"}, CategoryECMAScript, "__string.encodeURI", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	registerCallIntrinsic(m, []string{"decodeURI"}, CategoryECMAScript, "__string.decodeURI", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	registerCallIntrinsic(m, []string{"Number"}, CategoryECMAScript, "__number.new", nil, ir.TypeNumber, 0, 1)
	registerCustomIntrinsic(m, []string{"String"}, CategoryECMAScript, "__string.new", ir.TypeString, 0, 1, lowerStringConversion)
	registerCustomIntrinsic(m, []string{"Object"}, CategoryECMAScript, "__object.new", ir.TypeObject, 0, 1, lowerObjectConversion)
	// JSON.parse is a dynamic boundary: its result is any in TypeScript and
	// must remain boxed until a checked use narrows it.
	registerCallIntrinsic(m, []string{"JSON.parse"}, CategoryECMAScript, "__json.parse_unknown", []ir.Type{ir.TypeString}, ir.TypeUnknown, 1, 1)
	m["JSON.stringify"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "JSON.stringify",
		MinArgs:  1,
		MaxArgs:  1,
		Lower:    lowerJSONStringify,
	}
}

func registerMathIntrinsics(m map[string]BuiltinIntrinsic) {
	// Math functions (Category 1: ECMAScript)
	math1 := []string{"abs", "ceil", "floor", "trunc", "sqrt", "cbrt", "round", "fround", "sin", "cos", "tan", "asin", "acos", "atan", "sinh", "cosh", "tanh", "asinh", "acosh", "atanh", "log", "log2", "log10", "log1p", "exp", "expm1", "sign", "clz32"}
	for _, fn := range math1 {
		name := "Math." + fn
		m[name] = BuiltinIntrinsic{Category: CategoryECMAScript, Name: name, ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerCall("__"+name, ir.TypeNumber)}
	}
	math2 := []string{"pow", "atan2", "hypot", "imul"}
	for _, fn := range math2 {
		name := "Math." + fn
		m[name] = BuiltinIntrinsic{Category: CategoryECMAScript, Name: name, ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 2, MaxArgs: 2, Lower: lowerCall("__"+name, ir.TypeNumber)}
	}
	m["Math.min"] = BuiltinIntrinsic{Category: CategoryECMAScript, Name: "Math.min", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 0, MaxArgs: -1, Lower: lowerMathMinMax("__Math.min")}
	m["Math.max"] = BuiltinIntrinsic{Category: CategoryECMAScript, Name: "Math.max", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 0, MaxArgs: -1, Lower: lowerMathMinMax("__Math.max")}
	m["Math.random"] = BuiltinIntrinsic{Category: CategoryECMAScript, Name: "Math.random", ArgumentTypes: nil, MinArgs: 0, MaxArgs: 0, Lower: lowerCall("__Math.random", ir.TypeNumber)}
}
