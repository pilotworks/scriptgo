package lowering

import (
	"fmt"
	"slices"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// BuiltinCategory specifies the standard architectural group of a built-in symbol.
//
// Dual-Surface APIs (e.g. process, Buffer, console, crypto, timers, URL):
// Symbols that exist both as auto-globals (Category 2/3) and as explicit module
// imports (Category 4 via node:process, node:buffer, node:timers, etc.) resolve
// to identical underlying native implementations.
type BuiltinCategory string

const (
	CategoryECMAScript BuiltinCategory = "ECMAScript"
	CategoryWebCompat  BuiltinCategory = "WebCompat"
	CategoryNodeGlobal BuiltinCategory = "NodeGlobal"
	CategoryNodeModule BuiltinCategory = "NodeModule"
)

// BuiltinGlobal describes a globally available value admitted by the native subset.
type BuiltinGlobal struct {
	Category BuiltinCategory
	Name     string
	Type     ir.Type
	Value    string
}

// BuiltinIntrinsic describes a small, explicitly promoted intrinsic.
type BuiltinIntrinsic struct {
	Category      BuiltinCategory
	Name          string
	ArgumentTypes []ir.Type
	MinArgs       int
	MaxArgs       int
	Lower         func(IntrinsicCall, BuiltinIntrinsic) (string, ir.Type, error)
}

type IntrinsicCall struct {
	Path            string
	Expression      *typescriptgo.SyntaxExpression
	Result          string
	Function        *ir.Function
	Env             map[string]ir.Type
	Counter         *int
	Shapes          map[string]ir.ObjectShape
	Signatures      map[string]ir.Function
	LowerExpression lowerExpressionFunc
}

type lowerExpressionFunc func(string, *typescriptgo.SyntaxExpression, string, *ir.Function, map[string]ir.Type, *int, map[string]ir.ObjectShape, map[string]ir.Function) (string, ir.Type, error)

var builtinGlobals = map[string]BuiltinGlobal{
	// Category 1: ECMAScript built-ins
	"NaN":                      {Category: CategoryECMAScript, Name: "NaN", Type: ir.TypeNumber, Value: "NaN"},
	"Infinity":                 {Category: CategoryECMAScript, Name: "Infinity", Type: ir.TypeNumber, Value: "+Inf"},
	"Math.PI":                  {Category: CategoryECMAScript, Name: "Math.PI", Type: ir.TypeNumber, Value: "3.141592653589793"},
	"Math.E":                   {Category: CategoryECMAScript, Name: "Math.E", Type: ir.TypeNumber, Value: "2.718281828459045"},
	"Math.LN2":                 {Category: CategoryECMAScript, Name: "Math.LN2", Type: ir.TypeNumber, Value: "0.6931471805599453"},
	"Math.LN10":                {Category: CategoryECMAScript, Name: "Math.LN10", Type: ir.TypeNumber, Value: "2.302585092994046"},
	"Math.LOG2E":               {Category: CategoryECMAScript, Name: "Math.LOG2E", Type: ir.TypeNumber, Value: "1.4426950408889634"},
	"Math.LOG10E":              {Category: CategoryECMAScript, Name: "Math.LOG10E", Type: ir.TypeNumber, Value: "0.4342944819032518"},
	"Math.SQRT1_2":             {Category: CategoryECMAScript, Name: "Math.SQRT1_2", Type: ir.TypeNumber, Value: "0.7071067811865476"},
	"Math.SQRT2":               {Category: CategoryECMAScript, Name: "Math.SQRT2", Type: ir.TypeNumber, Value: "1.4142135623730951"},
	"Number.MAX_SAFE_INTEGER":  {Category: CategoryECMAScript, Name: "Number.MAX_SAFE_INTEGER", Type: ir.TypeNumber, Value: "9007199254740991"},
	"Number.MIN_SAFE_INTEGER":  {Category: CategoryECMAScript, Name: "Number.MIN_SAFE_INTEGER", Type: ir.TypeNumber, Value: "-9007199254740991"},
	"Number.MAX_VALUE":         {Category: CategoryECMAScript, Name: "Number.MAX_VALUE", Type: ir.TypeNumber, Value: "1.7976931348623157e+308"},
	"Number.MIN_VALUE":         {Category: CategoryECMAScript, Name: "Number.MIN_VALUE", Type: ir.TypeNumber, Value: "5e-324"},
	"Number.EPSILON":           {Category: CategoryECMAScript, Name: "Number.EPSILON", Type: ir.TypeNumber, Value: "2.220446049250313e-16"},
	"Number.POSITIVE_INFINITY": {Category: CategoryECMAScript, Name: "Number.POSITIVE_INFINITY", Type: ir.TypeNumber, Value: "+Inf"},
	"Number.NEGATIVE_INFINITY": {Category: CategoryECMAScript, Name: "Number.NEGATIVE_INFINITY", Type: ir.TypeNumber, Value: "-Inf"},
	"Number.NaN":               {Category: CategoryECMAScript, Name: "Number.NaN", Type: ir.TypeNumber, Value: "NaN"},

	// Well-known Symbols (Category 1: ECMAScript)
	"Symbol.iterator":           {Category: CategoryECMAScript, Name: "Symbol.iterator", Type: ir.TypeSymbol, Value: "Symbol.iterator"},
	"Symbol.asyncIterator":      {Category: CategoryECMAScript, Name: "Symbol.asyncIterator", Type: ir.TypeSymbol, Value: "Symbol.asyncIterator"},
	"Symbol.dispose":            {Category: CategoryECMAScript, Name: "Symbol.dispose", Type: ir.TypeSymbol, Value: "Symbol.dispose"},
	"Symbol.asyncDispose":       {Category: CategoryECMAScript, Name: "Symbol.asyncDispose", Type: ir.TypeSymbol, Value: "Symbol.asyncDispose"},
	"Symbol.hasInstance":        {Category: CategoryECMAScript, Name: "Symbol.hasInstance", Type: ir.TypeSymbol, Value: "Symbol.hasInstance"},
	"Symbol.isConcatSpreadable": {Category: CategoryECMAScript, Name: "Symbol.isConcatSpreadable", Type: ir.TypeSymbol, Value: "Symbol.isConcatSpreadable"},
	"Symbol.match":              {Category: CategoryECMAScript, Name: "Symbol.match", Type: ir.TypeSymbol, Value: "Symbol.match"},
	"Symbol.matchAll":           {Category: CategoryECMAScript, Name: "Symbol.matchAll", Type: ir.TypeSymbol, Value: "Symbol.matchAll"},
	"Symbol.metadata":           {Category: CategoryECMAScript, Name: "Symbol.metadata", Type: ir.TypeSymbol, Value: "Symbol.metadata"},
	"Symbol.replace":            {Category: CategoryECMAScript, Name: "Symbol.replace", Type: ir.TypeSymbol, Value: "Symbol.replace"},
	"Symbol.search":             {Category: CategoryECMAScript, Name: "Symbol.search", Type: ir.TypeSymbol, Value: "Symbol.search"},
	"Symbol.species":            {Category: CategoryECMAScript, Name: "Symbol.species", Type: ir.TypeSymbol, Value: "Symbol.species"},
	"Symbol.split":              {Category: CategoryECMAScript, Name: "Symbol.split", Type: ir.TypeSymbol, Value: "Symbol.split"},
	"Symbol.toPrimitive":        {Category: CategoryECMAScript, Name: "Symbol.toPrimitive", Type: ir.TypeSymbol, Value: "Symbol.toPrimitive"},
	"Symbol.toStringTag":        {Category: CategoryECMAScript, Name: "Symbol.toStringTag", Type: ir.TypeSymbol, Value: "Symbol.toStringTag"},
	"Symbol.unscopables":        {Category: CategoryECMAScript, Name: "Symbol.unscopables", Type: ir.TypeSymbol, Value: "Symbol.unscopables"},
	"Error.stackTraceLimit":     {Category: CategoryNodeGlobal, Name: "Error.stackTraceLimit", Type: ir.TypeNumber, Value: "10"},
}

func lowerMathMinMax(callee string) func(IntrinsicCall, BuiltinIntrinsic) (string, ir.Type, error) {
	return func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
		args := call.Expression.Arguments
		if len(args) == 0 {
			val := "Infinity"
			if callee == "__Math.max" {
				val = "-Infinity"
			}
			res := call.Result
			if res == "" {
				res = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: res, Value: val, Span: toIRSpan(call.Path, call.Expression.Span)})
			return res, ir.TypeNumber, nil
		}
		if len(args) == 1 && args[0].Kind == "spread" {
			arrVal, _, err := call.LowerExpression(call.Path, args[0].Left, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			res := call.Result
			if res == "" {
				res = nextTemp(call.Counter)
			}
			call.Env[res] = ir.TypeNumber
			initVal := "Infinity"
			if callee == "__Math.max" {
				initVal = "-Infinity"
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeNumber,
				Result: res,
				Value:  initVal,
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			lenRes := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: lenRes,
				Callee: "__array.length",
				Args:   []string{arrVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			idxVar := nextTemp(call.Counter)
			call.Env[idxVar] = ir.TypeNumber
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
			elemVal := nextTemp(call.Counter)
			bodyBlock = append(bodyBlock, ir.Instruction{
				Op:     ir.OpIndex,
				Type:   ir.TypeNumber,
				Result: elemVal,
				Args:   []string{arrVal, idxVar},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			minMaxRes := nextTemp(call.Counter)
			bodyBlock = append(bodyBlock, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: minMaxRes,
				Callee: callee,
				Args:   []string{res, elemVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			bodyBlock = append(bodyBlock, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   ir.TypeNumber,
				Result: res,
				Args:   []string{minMaxRes},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			oneConst := nextTemp(call.Counter)
			bodyBlock = append(bodyBlock, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeNumber,
				Result: oneConst,
				Value:  "1",
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			nextIdx := nextTemp(call.Counter)
			bodyBlock = append(bodyBlock, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     ir.TypeNumber,
				Result:   nextIdx,
				Operator: "+",
				Args:     []string{idxVar, oneConst},
				Span:     toIRSpan(call.Path, call.Expression.Span),
			})
			bodyBlock = append(bodyBlock, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   ir.TypeNumber,
				Result: idxVar,
				Args:   []string{nextIdx},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})

			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:   ir.OpWhile,
				Type: ir.TypeVoid,
				Args: []string{condRes},
				Cond: condBlock,
				Body: bodyBlock,
				Span: toIRSpan(call.Path, call.Expression.Span),
			})
			return res, ir.TypeNumber, nil
		}
		if len(args) == 1 {
			return call.LowerExpression(call.Path, args[0], call.Result, call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		}
		currentVal, _, err := call.LowerExpression(call.Path, args[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if err != nil {
			return "", "", err
		}
		for i := 1; i < len(args); i++ {
			nextVal, _, err := call.LowerExpression(call.Path, args[i], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			res := nextTemp(call.Counter)
			if i == len(args)-1 && call.Result != "" {
				res = call.Result
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: res,
				Callee: callee,
				Args:   []string{currentVal, nextVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			currentVal = res
		}
		return currentVal, ir.TypeNumber, nil
	}
}

func lowerCall(callee string, returnType ir.Type) func(IntrinsicCall, BuiltinIntrinsic) (string, ir.Type, error) {
	return func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
		args, _, err := call.arguments(intrinsic)
		if err != nil {
			return "", "", err
		}
		result := call.Result
		if result == "" {
			result = nextTemp(call.Counter)
		}
		target := callee
		if target == "" {
			target = "__" + intrinsic.Name
		}
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   returnType,
			Result: result,
			Callee: target,
			Args:   args,
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		return result, returnType, nil
	}
}

func lowerFsReadFileSync(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) < 1 || len(call.Expression.Arguments) > 2 {
		return "", "", fmt.Errorf("builtin %s expects between 1 and 2 argument(s)", intrinsic.Name)
	}
	args := make([]string, 0, len(call.Expression.Arguments))
	for _, argument := range call.Expression.Arguments {
		value, typ, err := call.LowerExpression(call.Path, argument, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if err != nil {
			return "", "", err
		}
		if typ != ir.TypeString && typ != ir.TypeUnknown {
			return "", "", fmt.Errorf("builtin %s does not support %s", intrinsic.Name, typ)
		}
		args = append(args, value)
	}
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	returnType := ir.TypeBuffer
	if len(args) == 2 && call.Expression.Arguments[1].Kind != "undefined" {
		returnType = ir.TypeString
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   returnType,
		Result: result,
		Callee: "__fs.readFileSync",
		Args:   args,
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, returnType, nil
}

func initIntrinsics() map[string]BuiltinIntrinsic {
	m := make(map[string]BuiltinIntrinsic)

	// Math functions (Category 1: ECMAScript)
	math1 := []string{"abs", "ceil", "floor", "trunc", "sqrt", "cbrt", "round", "fround", "f16round", "sin", "cos", "tan", "asin", "acos", "atan", "sinh", "cosh", "tanh", "asinh", "acosh", "atanh", "log", "log2", "log10", "log1p", "exp", "expm1", "sign", "clz32"}
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

	// Number & Global functions (Category 1: ECMAScript)
	register := func(aliases []string, cat BuiltinCategory, callee string, argTypes []ir.Type, retType ir.Type, minArgs, maxArgs int) {
		for _, name := range aliases {
			m[name] = BuiltinIntrinsic{Category: cat, Name: aliases[0], ArgumentTypes: argTypes, MinArgs: minArgs, MaxArgs: maxArgs, Lower: lowerCall(callee, retType)}
		}
	}

	register([]string{"parseInt", "Number.parseInt"}, CategoryECMAScript, "__number.parseInt", []ir.Type{ir.TypeString, ir.TypeNumber}, ir.TypeNumber, 1, 2)
	register([]string{"parseFloat", "Number.parseFloat"}, CategoryECMAScript, "__number.parseFloat", []ir.Type{ir.TypeString}, ir.TypeNumber, 1, 1)
	register([]string{"isNaN", "Number.isNaN"}, CategoryECMAScript, "__number.isNaN", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	register([]string{"isFinite", "Number.isFinite"}, CategoryECMAScript, "__number.isFinite", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	register([]string{"Number.isInteger"}, CategoryECMAScript, "__number.isInteger", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	register([]string{"Number.isSafeInteger"}, CategoryECMAScript, "__number.isSafeInteger", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	register([]string{"String.fromCodePoint"}, CategoryECMAScript, "__string.fromCodePoint", []ir.Type{ir.TypeNumber}, ir.TypeString, 1, 1)
	register([]string{"String.fromCharCode"}, CategoryECMAScript, "__string.fromCharCode", []ir.Type{ir.TypeNumber}, ir.TypeString, 0, -1)
	register([]string{"String.raw"}, CategoryECMAScript, "__string.raw", nil, ir.TypeString, 1, -1)
	register([]string{"encodeURIComponent"}, CategoryECMAScript, "__string.encodeURIComponent", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"decodeURIComponent"}, CategoryECMAScript, "__string.decodeURIComponent", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"encodeURI"}, CategoryECMAScript, "__string.encodeURI", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"decodeURI"}, CategoryECMAScript, "__string.decodeURI", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"Number"}, CategoryECMAScript, "__number.new", nil, ir.TypeNumber, 0, 1)
	register([]string{"String"}, CategoryECMAScript, "__string.new", nil, ir.TypeString, 0, 1)
	register([]string{"Object"}, CategoryECMAScript, "__object.new", nil, ir.TypeObject, 0, 1)
	// JSON.parse is a dynamic boundary: its result is any in TypeScript and
	// must remain boxed until a checked use narrows it.
	register([]string{"JSON.parse"}, CategoryECMAScript, "__json.parse_unknown", []ir.Type{ir.TypeString}, ir.TypeUnknown, 1, 1)
	m["JSON.stringify"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "JSON.stringify",
		MinArgs:  1,
		MaxArgs:  1,
		Lower:    lowerJSONStringify,
	}
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
	m["RegExp"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "RegExp",
		MinArgs:  1,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			ensureRegExpShape(call.Shapes)
			patternVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
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
	m["RegExp.escape"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "RegExp.escape",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			strVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
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
				Callee: "__regexp.escape",
				Args:   []string{strVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeString, nil
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
				if dType == ir.TypeString {
					descVal = dv
				} else if dType == ir.TypeNumber {
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op: ir.OpCall, Type: ir.TypeString, Result: descVal, Callee: "__string.fromNumber", Args: []string{dv}, Span: toIRSpan(call.Path, call.Expression.Span),
					})
				} else {
					descVal = dv
				}
			} else {
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpConst, Type: ir.TypeString, Result: descVal, Value: "", Span: toIRSpan(call.Path, call.Expression.Span),
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

	// Date globals (Category 1: ECMAScript)
	register([]string{"Date.now", "__date.now"}, CategoryECMAScript, "__date.now", nil, ir.TypeNumber, 0, 0)
	register([]string{"Date.parse", "__date.parse"}, CategoryECMAScript, "__date.parse", []ir.Type{ir.TypeString}, ir.TypeNumber, 1, 1)
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

	// TypedArray & ArrayBuffer globals (Category 1: ECMAScript)
	m["ArrayBuffer.isView"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "ArrayBuffer.isView",
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
				Callee: "__arraybuffer.isView",
				Args:   []string{argVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBool, nil
		},
	}
	for _, name := range []string{
		"Uint8Array", "Int8Array", "Uint8ClampedArray",
		"Int16Array", "Uint16Array", "Int32Array", "Uint32Array",
		"Float32Array", "Float64Array", "BigInt64Array", "BigUint64Array",
	} {
		targetKind := ir.Type(name)
		className := name
		m[name+".from"] = BuiltinIntrinsic{
			Category: CategoryECMAScript,
			Name:     name + ".from",
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
					Type:   targetKind,
					Result: result,
					Callee: "__typedarray.new_array",
					Value:  className,
					Args:   []string{argVal},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				return result, targetKind, nil
			},
		}
	}

	// Web-compatible & Timer globals (Category 2: WebCompat & NodeGlobal)
	register([]string{"btoa", "__scriptgo.btoa", "buffer.btoa"}, CategoryWebCompat, "__web.btoa", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"atob", "__scriptgo.atob", "buffer.atob"}, CategoryWebCompat, "__web.atob", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"performance.now"}, CategoryWebCompat, "__performance.now", nil, ir.TypeNumber, 0, 0)
	register([]string{"queueMicrotask"}, CategoryWebCompat, "__async.queueMicrotask", []ir.Type{ir.TypeClosure}, ir.TypeVoid, 1, 1)
	register([]string{"Promise.resolve"}, CategoryECMAScript, "__async.promise_resolve", []ir.Type{ir.TypeNumber}, ir.Type("object:Promise"), 1, 1)
	register([]string{"Promise.all"}, CategoryECMAScript, "__async.promise_all", []ir.Type{ir.TypeObject}, ir.Type("object:Promise"), 1, 1)
	register([]string{"setTimeout", "__scriptgo.setTimeout", "timers.setTimeout"}, CategoryWebCompat, "__timers.setTimeout", []ir.Type{ir.TypeClosure, ir.TypeNumber, ir.TypeUnknownArray}, ir.TypeNumber, 1, 3)
	register([]string{"clearTimeout", "__scriptgo.clearTimeout", "timers.clearTimeout"}, CategoryWebCompat, "__timers.clearTimeout", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	register([]string{"setInterval", "__scriptgo.setInterval", "timers.setInterval"}, CategoryWebCompat, "__timers.setInterval", []ir.Type{ir.TypeClosure, ir.TypeNumber, ir.TypeUnknownArray}, ir.TypeNumber, 1, 3)
	register([]string{"clearInterval", "__scriptgo.clearInterval", "timers.clearInterval"}, CategoryWebCompat, "__timers.clearInterval", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	register([]string{"setImmediate", "__scriptgo.setImmediate", "timers.setImmediate"}, CategoryWebCompat, "__timers.setImmediate", []ir.Type{ir.TypeClosure, ir.TypeUnknownArray}, ir.TypeNumber, 1, 2)
	register([]string{"clearImmediate", "__scriptgo.clearImmediate", "timers.clearImmediate"}, CategoryWebCompat, "__timers.clearImmediate", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)

	// Atomics & SharedArrayBuffer globals (Category 1: ECMAScript)
	register([]string{"__scriptgo.sharedArrayBufferNew", "SharedArrayBuffer"}, CategoryECMAScript, "__atomics.sharedArrayBufferNew", []ir.Type{ir.TypeNumber}, ir.TypeArrayBuffer, 1, 1)
	register([]string{"__scriptgo.atomicsIsLockFree", "Atomics.isLockFree"}, CategoryECMAScript, "__atomics.isLockFree", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	register([]string{"__scriptgo.atomicsAdd", "Atomics.add"}, CategoryECMAScript, "__atomics.add", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	register([]string{"__scriptgo.atomicsSub", "Atomics.sub"}, CategoryECMAScript, "__atomics.sub", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	register([]string{"__scriptgo.atomicsAnd", "Atomics.and"}, CategoryECMAScript, "__atomics.and", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	register([]string{"__scriptgo.atomicsOr", "Atomics.or"}, CategoryECMAScript, "__atomics.or", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	register([]string{"__scriptgo.atomicsXor", "Atomics.xor"}, CategoryECMAScript, "__atomics.xor", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	register([]string{"__scriptgo.atomicsLoad", "Atomics.load"}, CategoryECMAScript, "__atomics.load", []ir.Type{ir.TypeInt32Array, ir.TypeNumber}, ir.TypeNumber, 2, 2)
	register([]string{"__scriptgo.atomicsStore", "Atomics.store"}, CategoryECMAScript, "__atomics.store", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	register([]string{"__scriptgo.atomicsExchange", "Atomics.exchange"}, CategoryECMAScript, "__atomics.exchange", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	register([]string{"__scriptgo.atomicsCompareExchange", "Atomics.compareExchange"}, CategoryECMAScript, "__atomics.compareExchange", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 4, 4)
	register([]string{"__scriptgo.atomicsWait", "Atomics.wait"}, CategoryECMAScript, "__atomics.wait", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber, ir.TypeNumber}, ir.TypeString, 3, 4)
	register([]string{"__scriptgo.atomicsNotify", "Atomics.notify"}, CategoryECMAScript, "__atomics.notify", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 2, 3)

	// WeakRef & FinalizationRegistry (Category 1: ECMAScript)
	register([]string{"__scriptgo.weakrefNew", "WeakRef"}, CategoryECMAScript, "__weakref.new", []ir.Type{ir.TypeObject}, ir.TypeObject, 1, 1)
	register([]string{"__scriptgo.weakrefDeref"}, CategoryECMAScript, "__weakref.deref", []ir.Type{ir.TypeObject}, ir.TypeObject, 1, 1)
	register([]string{"__scriptgo.finalizationRegistryNew", "FinalizationRegistry"}, CategoryECMAScript, "__finalization_registry.new", []ir.Type{ir.TypeClosure}, ir.TypeObject, 1, 1)
	register([]string{"__scriptgo.finalizationRegistryRegister"}, CategoryECMAScript, "__finalization_registry.register", []ir.Type{ir.TypeObject, ir.TypeObject, ir.TypeUnknown, ir.TypeObject}, ir.TypeVoid, 3, 4)
	register([]string{"__scriptgo.finalizationRegistryUnregister"}, CategoryECMAScript, "__finalization_registry.unregister", []ir.Type{ir.TypeObject, ir.TypeObject}, ir.TypeBool, 2, 2)

	// Node-specific globals (Category 3: NodeGlobal)
	for _, logMethod := range []string{"log", "info", "debug", "warn", "error", "dir", "dirxml", "table"} {
		name := "console." + logMethod
		m[name] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: name, ArgumentTypes: []ir.Type{
			ir.TypeNumber, ir.TypeBigInt, ir.TypeSymbol, ir.TypeString, ir.TypeBool, ir.TypeUnknown,
			ir.TypeNumberArray, ir.TypeStringArray, ir.TypeObject,
			ir.TypeUint8Array, ir.TypeInt8Array, ir.TypeUint8ClampedArray,
			ir.TypeInt16Array, ir.TypeUint16Array, ir.TypeInt32Array, ir.TypeUint32Array,
			ir.TypeFloat32Array, ir.TypeFloat64Array, ir.TypeBigInt64Array, ir.TypeBigUint64Array,
			ir.TypeDataView, ir.TypeArrayBuffer,
			ir.TypeMap, ir.TypeSet,
		}, MinArgs: 0, MaxArgs: 256, Lower: lowerPrint}
	}
	m["console.assert"] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: "console.assert", MinArgs: 0, MaxArgs: 256, Lower: lowerConsoleAssert}
	m["console.group"] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: "console.group", MinArgs: 0, MaxArgs: 256, Lower: lowerConsoleGroup}
	m["console.groupCollapsed"] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: "console.groupCollapsed", MinArgs: 0, MaxArgs: 256, Lower: lowerConsoleGroup}
	m["console.timeLog"] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: "console.timeLog", MinArgs: 0, MaxArgs: 256, Lower: lowerConsoleTimeLog}
	m["console.trace"] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: "console.trace", MinArgs: 0, MaxArgs: 256, Lower: lowerConsoleTrace}
	m["console.profile"] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: "console.profile", MinArgs: 0, MaxArgs: 256, Lower: lowerConsoleNoop}
	m["console.profileEnd"] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: "console.profileEnd", MinArgs: 0, MaxArgs: 256, Lower: lowerConsoleNoop}
	m["console.timeStamp"] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: "console.timeStamp", MinArgs: 0, MaxArgs: 256, Lower: lowerConsoleNoop}

	register([]string{"console.clear", "__console.clear"}, CategoryNodeGlobal, "__console.clear", nil, ir.TypeVoid, 0, 0)
	register([]string{"console.groupEnd", "__console.groupEnd"}, CategoryNodeGlobal, "__console.groupEnd", nil, ir.TypeVoid, 0, 0)
	register([]string{"console.count", "__console.count"}, CategoryNodeGlobal, "__console.count", []ir.Type{ir.TypeString}, ir.TypeVoid, 0, 1)
	register([]string{"console.countReset", "__console.countReset"}, CategoryNodeGlobal, "__console.countReset", []ir.Type{ir.TypeString}, ir.TypeVoid, 0, 1)
	register([]string{"console.time", "__console.time"}, CategoryNodeGlobal, "__console.time", []ir.Type{ir.TypeString}, ir.TypeVoid, 0, 1)
	register([]string{"console.timeEnd", "__console.timeEnd"}, CategoryNodeGlobal, "__console.timeEnd", []ir.Type{ir.TypeString}, ir.TypeVoid, 0, 1)
	register([]string{"process.exit", "__scriptgo.exit"}, CategoryNodeGlobal, "__process.exit", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	register([]string{"process.cwd", "__scriptgo.cwd"}, CategoryNodeGlobal, "__process.cwd", nil, ir.TypeString, 0, 0)
	register([]string{"__scriptgo.argv"}, CategoryNodeGlobal, "__process.argv", nil, ir.TypeStringArray, 0, 0)
	register([]string{"__scriptgo.pid"}, CategoryNodeGlobal, "__process.pid", nil, ir.TypeNumber, 0, 0)
	register([]string{"__scriptgo.ppid"}, CategoryNodeGlobal, "__process.ppid", nil, ir.TypeNumber, 0, 0)
	register([]string{"__scriptgo.version"}, CategoryNodeGlobal, "__process.version", nil, ir.TypeString, 0, 0)

	m["__scriptgo.readFileSync"] = BuiltinIntrinsic{
		Category: CategoryNodeModule,
		Name:     "__scriptgo.readFileSync",
		MinArgs:  1,
		MaxArgs:  2,
		Lower:    lowerFsReadFileSync,
	}
	register([]string{"__scriptgo.writeFileSync"}, CategoryNodeModule, "__fs.writeFileSync", []ir.Type{ir.TypeString}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.existsSync"}, CategoryNodeModule, "__fs.existsSync", []ir.Type{ir.TypeString}, ir.TypeBool, 1, 1)
	register([]string{"__scriptgo.unlinkSync"}, CategoryNodeModule, "__fs.unlinkSync", []ir.Type{ir.TypeString}, ir.TypeVoid, 1, 1)
	register([]string{"__scriptgo.readdirSync"}, CategoryNodeModule, "__fs.readdirSync", []ir.Type{ir.TypeString}, ir.TypeStringArray, 1, 1)
	register([]string{"__scriptgo.copyFileSync"}, CategoryNodeModule, "__fs.copyFileSync", []ir.Type{ir.TypeString, ir.TypeString}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.renameSync"}, CategoryNodeModule, "__fs.renameSync", []ir.Type{ir.TypeString, ir.TypeString}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.appendFileSync"}, CategoryNodeModule, "__fs.appendFileSync", []ir.Type{ir.TypeString, ir.TypeString}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.mkdirSync"}, CategoryNodeModule, "__fs.mkdirSync", []ir.Type{ir.TypeString, ir.TypeBool}, ir.TypeVoid, 1, 2)
	register([]string{"__scriptgo.rmSync"}, CategoryNodeModule, "__fs.rmSync", []ir.Type{ir.TypeString, ir.TypeBool, ir.TypeBool}, ir.TypeVoid, 1, 3)
	register([]string{"__scriptgo.statSync"}, CategoryNodeModule, "__fs.statSync", []ir.Type{ir.TypeString}, ir.Type("object:Stats"), 1, 1)
	register([]string{"__scriptgo.accessSync"}, CategoryNodeModule, "__fs.accessSync", []ir.Type{ir.TypeString, ir.TypeNumber}, ir.TypeBool, 1, 2)
	register([]string{"__scriptgo.chmodSync"}, CategoryNodeModule, "__fs.chmodSync", []ir.Type{ir.TypeString, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.realpathSync"}, CategoryNodeModule, "__fs.realpathSync", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"__scriptgo.truncateSync"}, CategoryNodeModule, "__fs.truncateSync", []ir.Type{ir.TypeString, ir.TypeNumber}, ir.TypeVoid, 1, 2)
	register([]string{"__scriptgo.mkdtempSync"}, CategoryNodeModule, "__fs.mkdtempSync", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"__scriptgo.openSync"}, CategoryNodeModule, "__fs.openSync", []ir.Type{ir.TypeString, ir.TypeString, ir.TypeNumber}, ir.TypeNumber, 1, 3)
	register([]string{"__scriptgo.closeSync"}, CategoryNodeModule, "__fs.closeSync", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	register([]string{"__scriptgo.readFdSync"}, CategoryNodeModule, "__fs.readFdSync", []ir.Type{ir.TypeNumber, ir.TypeBuffer, ir.TypeNumber, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 5, 5)
	register([]string{"__scriptgo.writeFdSync"}, CategoryNodeModule, "__fs.writeFdSync", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 4, 4)
	register([]string{"__scriptgo.opendirSync"}, CategoryNodeModule, "__fs.opendirSync", []ir.Type{ir.TypeString}, ir.TypeStringArray, 1, 1)
	register([]string{"__scriptgo.fstatSync"}, CategoryNodeModule, "__fs.fstatSync", []ir.Type{ir.TypeNumber}, ir.Type("object:Stats"), 1, 1)
	register([]string{"__scriptgo.statfsSync"}, CategoryNodeModule, "__fs.statfsSync", []ir.Type{ir.TypeString}, ir.Type("object:StatFs"), 1, 1)
	register([]string{"__scriptgo.chownSync"}, CategoryNodeModule, "__fs.chownSync", []ir.Type{ir.TypeString, ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.lchownSync"}, CategoryNodeModule, "__fs.lchownSync", []ir.Type{ir.TypeString, ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.fchownSync"}, CategoryNodeModule, "__fs.fchownSync", []ir.Type{ir.TypeNumber, ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.fchmodSync"}, CategoryNodeModule, "__fs.fchmodSync", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.linkSync"}, CategoryNodeModule, "__fs.linkSync", []ir.Type{ir.TypeString, ir.TypeString}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.symlinkSync"}, CategoryNodeModule, "__fs.symlinkSync", []ir.Type{ir.TypeString, ir.TypeString}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.readlinkSync"}, CategoryNodeModule, "__fs.readlinkSync", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"__scriptgo.utimesSync"}, CategoryNodeModule, "__fs.utimesSync", []ir.Type{ir.TypeString, ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.lutimesSync"}, CategoryNodeModule, "__fs.lutimesSync", []ir.Type{ir.TypeString, ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.futimesSync"}, CategoryNodeModule, "__fs.futimesSync", []ir.Type{ir.TypeNumber, ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.fsyncSync"}, CategoryNodeModule, "__fs.fsyncSync", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	register([]string{"__scriptgo.fdatasyncSync"}, CategoryNodeModule, "__fs.fdatasyncSync", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	register([]string{"__scriptgo.ftruncateSync"}, CategoryNodeModule, "__fs.ftruncateSync", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.rmdirSync"}, CategoryNodeModule, "__fs.rmdirSync", []ir.Type{ir.TypeString}, ir.TypeVoid, 1, 1)

	register([]string{"crypto.randomUUID", "__scriptgo.randomUUID", "randomUUID"}, CategoryNodeModule, "__crypto.randomUUID", nil, ir.TypeString, 0, 0)
	register([]string{"crypto.hashDigest", "__scriptgo.hashDigest", "hashDigest"}, CategoryNodeModule, "__crypto.hashDigest", []ir.Type{ir.TypeString, ir.TypeBuffer, ir.TypeUint8Array, ir.TypeUnknown, ir.TypeObject}, ir.TypeString, 2, 3)
	register([]string{"crypto.hashDigestBuffer", "__scriptgo.hashDigestBuffer", "hashDigestBuffer"}, CategoryNodeModule, "__crypto.hashDigestBuffer", []ir.Type{ir.TypeString, ir.TypeBuffer, ir.TypeUint8Array, ir.TypeUnknown, ir.TypeObject}, ir.TypeString, 2, 3)
	register([]string{"crypto.randomBytes", "__scriptgo.randomBytes", "randomBytes"}, CategoryNodeModule, "__crypto.randomBytes", []ir.Type{ir.TypeNumber}, ir.TypeBuffer, 1, 1)
	register([]string{"crypto.randomInt", "__scriptgo.randomInt", "randomInt"}, CategoryNodeModule, "__crypto.randomInt", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 2, 2)
	register([]string{"crypto.randomFill", "__scriptgo.randomFill", "randomFill"}, CategoryNodeModule, "__crypto.randomFill", []ir.Type{ir.TypeBuffer, ir.TypeNumber, ir.TypeNumber}, ir.TypeBuffer, 1, 3)
	register([]string{"crypto.timingSafeEqual", "__scriptgo.timingSafeEqual", "timingSafeEqual"}, CategoryNodeModule, "__crypto.timingSafeEqual", []ir.Type{ir.TypeBuffer, ir.TypeBuffer}, ir.TypeBool, 2, 2)
	register([]string{"crypto.hmacDigest", "__scriptgo.hmacDigest", "hmacDigest"}, CategoryNodeModule, "__crypto.hmacDigest", []ir.Type{ir.TypeString, ir.TypeString, ir.TypeString, ir.TypeString}, ir.TypeString, 3, 4)
	register([]string{"crypto.hmacDigestBuffer", "__scriptgo.hmacDigestBuffer", "hmacDigestBuffer"}, CategoryNodeModule, "__crypto.hmacDigestBuffer", []ir.Type{ir.TypeString, ir.TypeBuffer, ir.TypeBuffer, ir.TypeString}, ir.TypeString, 3, 4)
	register([]string{"crypto.pbkdf2Sync", "__scriptgo.pbkdf2Sync", "pbkdf2Sync"}, CategoryNodeModule, "__crypto.pbkdf2Sync", []ir.Type{ir.TypeString, ir.TypeString, ir.TypeNumber, ir.TypeNumber, ir.TypeString}, ir.TypeBuffer, 4, 5)
	register([]string{"crypto.hkdfSync", "__scriptgo.hkdfSync", "hkdfSync"}, CategoryNodeModule, "__crypto.hkdfSync", []ir.Type{ir.TypeString, ir.TypeString, ir.TypeString, ir.TypeString, ir.TypeNumber}, ir.TypeArrayBuffer, 5, 5)
	register([]string{"crypto.scryptSync", "__scriptgo.scryptSync", "scryptSync"}, CategoryNodeModule, "__crypto.scryptSync", []ir.Type{ir.TypeString, ir.TypeString, ir.TypeNumber}, ir.TypeBuffer, 3, 3)
	register([]string{"__scriptgo.zlibTransformString", "zlibTransformString"}, CategoryNodeModule, "__zlib.transform_string", []ir.Type{ir.TypeString, ir.TypeNumber}, ir.TypeUint8Array, 2, 2)
	register([]string{"__scriptgo.zlibTransformBuffer", "zlibTransformBuffer"}, CategoryNodeModule, "__zlib.transform_buffer", []ir.Type{ir.TypeUint8Array, ir.TypeNumber}, ir.TypeUint8Array, 2, 2)
	register([]string{"os.platform", "__scriptgo.platform", "platform"}, CategoryNodeModule, "__os.platform", nil, ir.TypeString, 0, 0)
	register([]string{"os.arch", "__scriptgo.arch", "arch"}, CategoryNodeModule, "__os.arch", nil, ir.TypeString, 0, 0)
	register([]string{"os.homedir", "__scriptgo.homedir", "homedir"}, CategoryNodeModule, "__os.homedir", nil, ir.TypeString, 0, 0)
	register([]string{"os.uptime", "__scriptgo.uptime", "uptime"}, CategoryNodeModule, "__os.uptime", nil, ir.TypeNumber, 0, 0)
	register([]string{"os.totalmem", "__scriptgo.totalmem", "totalmem"}, CategoryNodeModule, "__os.totalmem", nil, ir.TypeNumber, 0, 0)
	register([]string{"os.freemem", "__scriptgo.freemem", "freemem"}, CategoryNodeModule, "__os.freemem", nil, ir.TypeNumber, 0, 0)
	register([]string{"os.type", "__scriptgo.type", "type"}, CategoryNodeModule, "__os.type", nil, ir.TypeString, 0, 0)
	register([]string{"os.release", "__scriptgo.release", "release"}, CategoryNodeModule, "__os.release", nil, ir.TypeString, 0, 0)
	register([]string{"os.tmpdir", "__scriptgo.tmpdir", "tmpdir"}, CategoryNodeModule, "__os.tmpdir", nil, ir.TypeString, 0, 0)
	register([]string{"__scriptgo.execSync"}, CategoryNodeModule, "__child_process.execSync", []ir.Type{ir.TypeString, ir.TypeString, ir.TypeString}, ir.TypeString, 1, 3)
	register([]string{"__scriptgo.spawnSync"}, CategoryNodeModule, "__child_process.spawnSync", []ir.Type{ir.TypeString, ir.TypeStringArray, ir.TypeString, ir.TypeString}, ir.Type("object:SpawnSyncReturns"), 1, 4)
	register([]string{"__scriptgo.fetchSync"}, CategoryNodeModule, "__http.fetchSync", []ir.Type{ir.TypeString, ir.TypeString, ir.TypeStringArray, ir.TypeString}, ir.Type("object:FetchResponseData"), 1, 4)
	register([]string{"__stream.getDefaultHighWaterMark", "__scriptgo.streamGetDefaultHighWaterMark"}, CategoryNodeModule, "__stream.getDefaultHighWaterMark", []ir.Type{ir.TypeBool}, ir.TypeNumber, 0, 1)
	register([]string{"__stream.setDefaultHighWaterMark", "__scriptgo.streamSetDefaultHighWaterMark"}, CategoryNodeModule, "__stream.setDefaultHighWaterMark", []ir.Type{ir.TypeBool, ir.TypeNumber}, ir.TypeVoid, 2, 2)

	register([]string{"__scriptgo.sqliteOpen"}, CategoryNodeModule, "__sqlite.open", nil, ir.TypeObject, 1, 2)
	register([]string{"__scriptgo.sqliteClose"}, CategoryNodeModule, "__sqlite.close", nil, ir.TypeVoid, 1, 1)
	register([]string{"__scriptgo.sqliteExec"}, CategoryNodeModule, "__sqlite.exec", nil, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.sqlitePrepare"}, CategoryNodeModule, "__sqlite.prepare", nil, ir.TypeObject, 2, 2)
	register([]string{"__scriptgo.sqliteRun"}, CategoryNodeModule, "__sqlite.run", nil, ir.Type("object:StatementResult"), 1, 3)
	register([]string{"__scriptgo.sqliteGet"}, CategoryNodeModule, "__sqlite.get", nil, ir.TypeObject, 1, 3)
	register([]string{"__scriptgo.sqliteAll"}, CategoryNodeModule, "__sqlite.all", nil, ir.TypeUnknownArray, 1, 3)
	register([]string{"__scriptgo.sqliteColumns"}, CategoryNodeModule, "__sqlite.columns", nil, ir.TypeUnknownArray, 1, 1)
	register([]string{"__scriptgo.sqliteExpandedSQL"}, CategoryNodeModule, "__sqlite.expandedSQL", nil, ir.TypeString, 1, 1)
	register([]string{"__scriptgo.sqliteFinalize"}, CategoryNodeModule, "__sqlite.finalize", nil, ir.TypeVoid, 1, 1)
	register([]string{"__scriptgo.sqliteStmtConfig"}, CategoryNodeModule, "__sqlite.stmtConfig", nil, ir.TypeVoid, 5, 5)
	register([]string{"__scriptgo.sqliteEnableLoadExtension"}, CategoryNodeModule, "__sqlite.enableLoadExtension", nil, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.sqliteLoadExtension"}, CategoryNodeModule, "__sqlite.loadExtension", nil, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.sqliteCreateSession"}, CategoryNodeModule, "__sqlite.createSession", nil, ir.TypeObject, 1, 2)
	register([]string{"__scriptgo.sqliteSessionChangeset"}, CategoryNodeModule, "__sqlite.sessionChangeset", nil, ir.TypeUint8Array, 1, 1)
	register([]string{"__scriptgo.sqliteSessionPatchset"}, CategoryNodeModule, "__sqlite.sessionPatchset", nil, ir.TypeUint8Array, 1, 1)
	register([]string{"__scriptgo.sqliteSessionClose"}, CategoryNodeModule, "__sqlite.sessionClose", nil, ir.TypeVoid, 1, 1)
	register([]string{"__scriptgo.sqliteApplyChangeset"}, CategoryNodeModule, "__sqlite.applyChangeset", nil, ir.TypeBool, 2, 3)
	register([]string{"__scriptgo.sqliteLocation"}, CategoryNodeModule, "__sqlite.location", nil, ir.TypeString, 1, 2)
	register([]string{"__scriptgo.sqliteIsTransaction"}, CategoryNodeModule, "__sqlite.isTransaction", nil, ir.TypeBool, 1, 1)
	register([]string{"__scriptgo.sqliteCreateFunction"}, CategoryNodeModule, "__sqlite.createFunction", nil, ir.TypeVoid, 4, 5)
	register([]string{"__scriptgo.sqliteCreateAggregate"}, CategoryNodeModule, "__sqlite.createAggregate", nil, ir.TypeVoid, 4, 5)
	register([]string{"__scriptgo.sqliteBackup"}, CategoryNodeModule, "__sqlite.backup", nil, ir.TypeNumber, 2, 2)

	// DNS Intrinsics
	register([]string{"__scriptgo.dnsLookup"}, CategoryNodeModule, "__dns.lookup", []ir.Type{ir.TypeString, ir.TypeNumber}, ir.TypeObject, 1, 2)
	register([]string{"__scriptgo.dnsLookupService"}, CategoryNodeModule, "__dns.lookupService", []ir.Type{ir.TypeString, ir.TypeNumber}, ir.TypeObject, 2, 2)
	register([]string{"__scriptgo.dnsReverse"}, CategoryNodeModule, "__dns.reverse", []ir.Type{ir.TypeString}, ir.TypeStringArray, 1, 1)
	register([]string{"__scriptgo.dnsResolveStrings"}, CategoryNodeModule, "__dns.resolveStrings", []ir.Type{ir.TypeString, ir.TypeString}, ir.TypeStringArray, 2, 2)

	// Net TCP Intrinsics
	register([]string{"__scriptgo.netSocketCreate"}, CategoryNodeModule, "__net.socketCreate", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 0, 2)
	register([]string{"__scriptgo.netSocketConnect"}, CategoryNodeModule, "__net.socketConnect", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.netSocketWrite"}, CategoryNodeModule, "__net.socketWrite", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	register([]string{"__scriptgo.netSocketRead"}, CategoryNodeModule, "__net.socketRead", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeString, 2, 2)
	register([]string{"__scriptgo.netSocketClose"}, CategoryNodeModule, "__net.socketClose", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	register([]string{"__scriptgo.netServerListen"}, CategoryNodeModule, "__net.serverListen", []ir.Type{ir.TypeString, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	register([]string{"__scriptgo.netServerAccept"}, CategoryNodeModule, "__net.serverAccept", []ir.Type{ir.TypeNumber}, ir.TypeObject, 1, 1)

	// Dgram UDP Intrinsics
	register([]string{"__scriptgo.dgramSocketCreate"}, CategoryNodeModule, "__dgram.socketCreate", []ir.Type{ir.TypeNumber}, ir.TypeNumber, 0, 1)
	register([]string{"__scriptgo.dgramBind"}, CategoryNodeModule, "__dgram.bind", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.dgramSend"}, CategoryNodeModule, "__dgram.send", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber, ir.TypeNumber, ir.TypeString}, ir.TypeNumber, 5, 5)
	register([]string{"__scriptgo.dgramRecv"}, CategoryNodeModule, "__dgram.recv", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeObject, 2, 2)
	register([]string{"__scriptgo.dgramSetBroadcast"}, CategoryNodeModule, "__dgram.setBroadcast", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.dgramSetMulticastTTL"}, CategoryNodeModule, "__dgram.setMulticastTTL", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.dgramSetMulticastLoopback"}, CategoryNodeModule, "__dgram.setMulticastLoopback", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.dgramSetRecvBufferSize"}, CategoryNodeModule, "__dgram.setRecvBufferSize", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.dgramSetSendBufferSize"}, CategoryNodeModule, "__dgram.setSendBufferSize", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.dgramGetRecvBufferSize"}, CategoryNodeModule, "__dgram.getRecvBufferSize", []ir.Type{ir.TypeNumber}, ir.TypeNumber, 1, 1)
	register([]string{"__scriptgo.dgramGetSendBufferSize"}, CategoryNodeModule, "__dgram.getSendBufferSize", []ir.Type{ir.TypeNumber}, ir.TypeNumber, 1, 1)
	register([]string{"__scriptgo.dgramSetTTL"}, CategoryNodeModule, "__dgram.setTTL", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.dgramSetMulticastInterface"}, CategoryNodeModule, "__dgram.setMulticastInterface", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.dgramAddMembership"}, CategoryNodeModule, "__dgram.addMembership", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeString}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.dgramDropMembership"}, CategoryNodeModule, "__dgram.dropMembership", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeString}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.dgramAddSourceSpecificMembership"}, CategoryNodeModule, "__dgram.addSourceSpecificMembership", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeString, ir.TypeString}, ir.TypeVoid, 4, 4)
	register([]string{"__scriptgo.dgramDropSourceSpecificMembership"}, CategoryNodeModule, "__dgram.dropSourceSpecificMembership", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeString, ir.TypeString}, ir.TypeVoid, 4, 4)
	register([]string{"__scriptgo.dgramConnect"}, CategoryNodeModule, "__dgram.connect", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.dgramDisconnect"}, CategoryNodeModule, "__dgram.disconnect", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	register([]string{"__scriptgo.dgramClose"}, CategoryNodeModule, "__dgram.close", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)

	// TLS intrinsics. The TypeScript node:tls adapter owns option handling and
	// object/event semantics; these calls expose only the native TLS ABI.
	register([]string{"__scriptgo.tlsContextCreate"}, CategoryNodeModule, "__tls.contextCreate", []ir.Type{ir.TypeString, ir.TypeBool}, ir.TypeNumber, 7, 7)
	register([]string{"__scriptgo.tlsSocketCreate"}, CategoryNodeModule, "__tls.socketCreate", []ir.Type{ir.TypeNumber, ir.TypeBool}, ir.TypeNumber, 2, 2)
	register([]string{"__scriptgo.tlsSocketConnect"}, CategoryNodeModule, "__tls.socketConnect", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber, ir.TypeString, ir.TypeBool, ir.TypeUint8Array}, ir.TypeNumber, 6, 6)
	register([]string{"__scriptgo.tlsSocketAdopt"}, CategoryNodeModule, "__tls.socketAdopt", []ir.Type{ir.TypeNumber, ir.TypeNumber, ir.TypeString, ir.TypeBool, ir.TypeBool, ir.TypeBool, ir.TypeUint8Array}, ir.TypeNumber, 7, 7)
	register([]string{"__scriptgo.tlsSocketWrite"}, CategoryNodeModule, "__tls.socketWrite", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	register([]string{"__scriptgo.tlsSocketWriteBytes"}, CategoryNodeModule, "__tls.socketWriteBytes", []ir.Type{ir.TypeNumber, ir.TypeUint8Array, ir.TypeBuffer}, ir.TypeNumber, 2, 2)
	register([]string{"__scriptgo.tlsSocketRead"}, CategoryNodeModule, "__tls.socketRead", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeString, 2, 2)
	register([]string{"__scriptgo.tlsPairWrite"}, CategoryNodeModule, "__tls.pairWrite", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeNumber, 4, 4)
	register([]string{"__scriptgo.tlsPairWriteBytes"}, CategoryNodeModule, "__tls.pairWriteBytes", []ir.Type{ir.TypeNumber, ir.TypeUint8Array, ir.TypeBuffer}, ir.TypeNumber, 3, 3)
	register([]string{"__scriptgo.tlsPairRead"}, CategoryNodeModule, "__tls.pairRead", []ir.Type{ir.TypeNumber}, ir.TypeString, 3, 3)
	register([]string{"__scriptgo.tlsSocketClose"}, CategoryNodeModule, "__tls.socketClose", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	register([]string{"__scriptgo.tlsSocketInfo"}, CategoryNodeModule, "__tls.socketInfo", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeString, 2, 2)
	register([]string{"__scriptgo.tlsSocketNumber"}, CategoryNodeModule, "__tls.socketNumber", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeNumber, 2, 2)
	register([]string{"__scriptgo.tlsSocketBool"}, CategoryNodeModule, "__tls.socketBool", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeBool, 2, 2)
	register([]string{"__scriptgo.tlsExportKeyingMaterial"}, CategoryNodeModule, "__tls.exportKeyingMaterial", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeString, ir.TypeUint8Array}, ir.TypeString, 4, 4)
	register([]string{"__scriptgo.tlsSocketSetOption"}, CategoryNodeModule, "__tls.socketSetOption", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.tlsSocketSetServername"}, CategoryNodeModule, "__tls.socketSetServername", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.tlsSocketSetSession"}, CategoryNodeModule, "__tls.socketSetSession", []ir.Type{ir.TypeNumber, ir.TypeUint8Array}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.tlsSocketSetKeyCert"}, CategoryNodeModule, "__tls.socketSetKeyCert", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.tlsSocketRenegotiate"}, CategoryNodeModule, "__tls.socketRenegotiate", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	register([]string{"__scriptgo.tlsPairCreate"}, CategoryNodeModule, "__tls.pairCreate", []ir.Type{ir.TypeNumber, ir.TypeBool}, ir.TypeNumber, 4, 4)
	register([]string{"__scriptgo.tlsServerListen"}, CategoryNodeModule, "__tls.serverListen", []ir.Type{ir.TypeNumber, ir.TypeBool, ir.TypeString}, ir.TypeNumber, 6, 6)
	register([]string{"__scriptgo.tlsServerAccept"}, CategoryNodeModule, "__tls.serverAccept", []ir.Type{ir.TypeNumber}, ir.TypeNumber, 1, 1)
	register([]string{"__scriptgo.tlsServerClose"}, CategoryNodeModule, "__tls.serverClose", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	register([]string{"__scriptgo.tlsServerInfo"}, CategoryNodeModule, "__tls.serverInfo", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeString, 2, 2)
	register([]string{"__scriptgo.tlsServerSetContext"}, CategoryNodeModule, "__tls.serverSetContext", []ir.Type{ir.TypeNumber, ir.TypeBool}, ir.TypeVoid, 4, 4)
	register([]string{"__scriptgo.tlsServerAddContext"}, CategoryNodeModule, "__tls.serverAddContext", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeVoid, 3, 3)
	register([]string{"__scriptgo.tlsServerSetTicketKeys"}, CategoryNodeModule, "__tls.serverSetTicketKeys", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeVoid, 2, 2)
	register([]string{"__scriptgo.tlsX509ParsePem"}, CategoryNodeModule, "__tls.x509ParsePem", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"__scriptgo.tlsX509ParseBytes"}, CategoryNodeModule, "__tls.x509ParseBytes", []ir.Type{ir.TypeUint8Array, ir.TypeBuffer}, ir.TypeString, 1, 1)
	register([]string{"__scriptgo.tlsCiphers"}, CategoryNodeModule, "__tls.ciphers", nil, ir.TypeString, 0, 0)
	register([]string{"__scriptgo.tlsRootCertificates"}, CategoryNodeModule, "__tls.rootCertificates", nil, ir.TypeString, 0, 0)
	register([]string{"__scriptgo.tlsSystemCertificates"}, CategoryNodeModule, "__tls.systemCertificates", nil, ir.TypeString, 0, 0)
	register([]string{"__scriptgo.tlsExtraCertificates"}, CategoryNodeModule, "__tls.extraCertificates", nil, ir.TypeString, 0, 0)

	registerObjectIntrinsics(m)
	registerBufferIntrinsics(m)
	registerArrayBuiltins(m)
	registerReflectIntrinsics(m)
	registerIteratorBuiltins(m)

	m["Error.captureStackTrace"] = BuiltinIntrinsic{
		Category: CategoryNodeGlobal,
		Name:     "Error.captureStackTrace",
		MinArgs:  1,
		MaxArgs:  2,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeVoid,
				Result: result,
				Value:  "",
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeVoid, nil
		},
	}
	m["__Error.captureStackTrace"] = m["Error.captureStackTrace"]

	return m
}

var builtinIntrinsics = initIntrinsics()

func builtinGlobal(name string) (BuiltinGlobal, bool) {
	global, ok := builtinGlobals[name]
	return global, ok
}

func builtinIntrinsic(name string) (BuiltinIntrinsic, bool) {
	intrinsic, ok := builtinIntrinsics[name]
	return intrinsic, ok
}

// BuiltinsByCategory returns all registered globals and intrinsics for a category.
func BuiltinsByCategory(cat BuiltinCategory) ([]BuiltinGlobal, []BuiltinIntrinsic) {
	var globals []BuiltinGlobal
	for _, g := range builtinGlobals {
		if g.Category == cat {
			globals = append(globals, g)
		}
	}
	var intrinsics []BuiltinIntrinsic
	for _, i := range builtinIntrinsics {
		if i.Category == cat {
			intrinsics = append(intrinsics, i)
		}
	}
	return globals, intrinsics
}

func (call IntrinsicCall) arguments(intrinsic BuiltinIntrinsic) ([]string, []ir.Type, error) {
	if len(call.Expression.Arguments) < intrinsic.MinArgs || (intrinsic.MaxArgs >= 0 && len(call.Expression.Arguments) > intrinsic.MaxArgs) {
		return nil, nil, fmt.Errorf("builtin %s expects between %d and %d argument(s)", intrinsic.Name, intrinsic.MinArgs, intrinsic.MaxArgs)
	}
	args := make([]string, 0, len(call.Expression.Arguments))
	types := make([]ir.Type, 0, len(call.Expression.Arguments))
	for _, argument := range call.Expression.Arguments {
		value, typ, err := call.LowerExpression(call.Path, argument, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if err != nil {
			return nil, nil, err
		}
		if len(intrinsic.ArgumentTypes) > 0 && !slices.Contains(intrinsic.ArgumentTypes, typ) {
			return nil, nil, fmt.Errorf("builtin %s does not support %s", intrinsic.Name, typ)
		}
		args = append(args, value)
		types = append(types, typ)
	}
	return args, types, nil
}
