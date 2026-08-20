package lowering

import (
	"fmt"
	"slices"
	"strings"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
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
	"Symbol.hasInstance":        {Category: CategoryECMAScript, Name: "Symbol.hasInstance", Type: ir.TypeSymbol, Value: "Symbol.hasInstance"},
	"Symbol.isConcatSpreadable": {Category: CategoryECMAScript, Name: "Symbol.isConcatSpreadable", Type: ir.TypeSymbol, Value: "Symbol.isConcatSpreadable"},
	"Symbol.match":              {Category: CategoryECMAScript, Name: "Symbol.match", Type: ir.TypeSymbol, Value: "Symbol.match"},
	"Symbol.replace":            {Category: CategoryECMAScript, Name: "Symbol.replace", Type: ir.TypeSymbol, Value: "Symbol.replace"},
	"Symbol.search":             {Category: CategoryECMAScript, Name: "Symbol.search", Type: ir.TypeSymbol, Value: "Symbol.search"},
	"Symbol.species":            {Category: CategoryECMAScript, Name: "Symbol.species", Type: ir.TypeSymbol, Value: "Symbol.species"},
	"Symbol.split":              {Category: CategoryECMAScript, Name: "Symbol.split", Type: ir.TypeSymbol, Value: "Symbol.split"},
	"Symbol.toPrimitive":        {Category: CategoryECMAScript, Name: "Symbol.toPrimitive", Type: ir.TypeSymbol, Value: "Symbol.toPrimitive"},
	"Symbol.toStringTag":        {Category: CategoryECMAScript, Name: "Symbol.toStringTag", Type: ir.TypeSymbol, Value: "Symbol.toStringTag"},
	"Symbol.unscopables":        {Category: CategoryECMAScript, Name: "Symbol.unscopables", Type: ir.TypeSymbol, Value: "Symbol.unscopables"},
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

func lowerJSONStringifyObject(call IntrinsicCall, argVal string, shape ir.ObjectShape) (string, ir.Type, error) {
	curr := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: curr,
		Value:  "{",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})

	for i, f := range shape.Fields {
		prefix := fmt.Sprintf("\"%s\":", f.Name)
		if i > 0 {
			prefix = "," + prefix
		}
		prefConst := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: prefConst,
			Value:  prefix,
			Span:   toIRSpan(call.Path, call.Expression.Span),
		})
		afterPref := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:       ir.OpBinary,
			Type:     ir.TypeString,
			Result:   afterPref,
			Operator: "+",
			Args:     []string{curr, prefConst},
			Span:     toIRSpan(call.Path, call.Expression.Span),
		})
		curr = afterPref

		fVal := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:         ir.OpFieldGet,
			Type:       f.Type,
			Result:     fVal,
			Callee:     shape.Name,
			Field:      f.Name,
			FieldIndex: i,
			Args:       []string{argVal},
			Span:       toIRSpan(call.Path, call.Expression.Span),
		})

		fStr := nextTemp(call.Counter)
		switch f.Type {
		case ir.TypeString:
			qConst := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeString, Result: qConst, Value: "\"", Span: toIRSpan(call.Path, call.Expression.Span),
			})
			q1 := nextTemp(call.Counter)
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpBinary, Type: ir.TypeString, Result: q1, Operator: "+", Args: []string{qConst, fVal}, Span: toIRSpan(call.Path, call.Expression.Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpBinary, Type: ir.TypeString, Result: fStr, Operator: "+", Args: []string{q1, qConst}, Span: toIRSpan(call.Path, call.Expression.Span),
			})
		case ir.TypeNumber:
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpCall, Type: ir.TypeString, Result: fStr, Callee: "__string.fromNumber", Args: []string{fVal}, Span: toIRSpan(call.Path, call.Expression.Span),
			})
		case ir.TypeBool:
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpCall, Type: ir.TypeString, Result: fStr, Callee: "__string.fromBool", Args: []string{fVal}, Span: toIRSpan(call.Path, call.Expression.Span),
			})
		default:
			if after, ok := strings.CutPrefix(string(f.Type), "object:"); ok {
				nestedShapeName := after
				if nestedShape, ok := call.Shapes[nestedShapeName]; ok {
					nestedStr, _, err := lowerJSONStringifyObject(call, fVal, nestedShape)
					if err != nil {
						return "", "", err
					}
					fStr = nestedStr
				} else {
					qConst := nextTemp(call.Counter)
					call.Function.Body = append(call.Function.Body, ir.Instruction{
						Op: ir.OpConst, Type: ir.TypeString, Result: qConst, Value: "\"{}\"", Span: toIRSpan(call.Path, call.Expression.Span),
					})
					fStr = qConst
				}
			} else if strings.HasSuffix(string(f.Type), "[]") || f.Type == ir.TypeNumberArray || f.Type == ir.TypeStringArray {
				callee := "__json.stringify_string_array"
				if f.Type == ir.TypeNumberArray {
					callee = "__json.stringify_number_array"
				}
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpCall, Type: ir.TypeString, Result: fStr, Callee: callee, Args: []string{fVal}, Span: toIRSpan(call.Path, call.Expression.Span),
				})
			} else {
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op: ir.OpCall, Type: ir.TypeString, Result: fStr, Callee: "__string.fromNumber", Args: []string{fVal}, Span: toIRSpan(call.Path, call.Expression.Span),
				})
			}
		}

		afterVal := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{
			Op:       ir.OpBinary,
			Type:     ir.TypeString,
			Result:   afterVal,
			Operator: "+",
			Args:     []string{curr, fStr},
			Span:     toIRSpan(call.Path, call.Expression.Span),
		})
		curr = afterVal
	}

	closeConst := nextTemp(call.Counter)
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: closeConst,
		Value:  "}",
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	res := call.Result
	if res == "" {
		res = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:       ir.OpBinary,
		Type:     ir.TypeString,
		Result:   res,
		Operator: "+",
		Args:     []string{curr, closeConst},
		Span:     toIRSpan(call.Path, call.Expression.Span),
	})
	return res, ir.TypeString, nil
}

func lowerJSONStringify(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) != 1 {
		return "", "", fmt.Errorf("JSON.stringify expects exactly 1 argument")
	}
	argVal, argType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}
	if after, ok := strings.CutPrefix(string(argType), "object:"); ok {
		shapeName := after
		shape, ok := call.Shapes[shapeName]
		if !ok {
			return "", "", fmt.Errorf("unknown shape %q for JSON.stringify", shapeName)
		}
		return lowerJSONStringifyObject(call, argVal, shape)
	}
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	var callee string
	args := []string{argVal}
	switch argType {
	case ir.TypeNumber:
		callee = "__json.stringify_number"
	case ir.TypeString:
		callee = "__json.stringify_string"
	case ir.TypeBool:
		callee = "__json.stringify_bool"
	case ir.TypeNumberArray:
		callee = "__json.stringify_number_array"
	case ir.TypeStringArray:
		callee = "__json.stringify_string_array"
	default:
		return "", "", fmt.Errorf("JSON.stringify does not support type %s", argType)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeString,
		Result: result,
		Callee: callee,
		Args:   args,
		Span:   toIRSpan(call.Path, call.Expression.Span),
	})
	return result, ir.TypeString, nil
}

func initIntrinsics() map[string]BuiltinIntrinsic {
	m := make(map[string]BuiltinIntrinsic)

	// Math functions (Category 1: ECMAScript)
	math1 := []string{"abs", "ceil", "floor", "trunc", "sqrt", "round", "sin", "cos", "tan", "atan", "log", "log2", "log10", "exp", "sign"}
	for _, fn := range math1 {
		name := "Math." + fn
		m[name] = BuiltinIntrinsic{Category: CategoryECMAScript, Name: name, ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerCall("__"+name, ir.TypeNumber)}
	}
	math2 := []string{"min", "max", "pow", "atan2", "hypot"}
	for _, fn := range math2 {
		name := "Math." + fn
		m[name] = BuiltinIntrinsic{Category: CategoryECMAScript, Name: name, ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 2, MaxArgs: 2, Lower: lowerCall("__"+name, ir.TypeNumber)}
	}
	m["Math.random"] = BuiltinIntrinsic{Category: CategoryECMAScript, Name: "Math.random", ArgumentTypes: nil, MinArgs: 0, MaxArgs: 0, Lower: lowerCall("__Math.random", ir.TypeNumber)}

	// Number & Global functions (Category 1: ECMAScript)
	register := func(aliases []string, cat BuiltinCategory, callee string, argTypes []ir.Type, retType ir.Type, minArgs, maxArgs int) {
		for _, name := range aliases {
			m[name] = BuiltinIntrinsic{Category: cat, Name: aliases[0], ArgumentTypes: argTypes, MinArgs: minArgs, MaxArgs: maxArgs, Lower: lowerCall(callee, retType)}
		}
	}

	register([]string{"parseInt", "Number.parseInt"}, CategoryECMAScript, "__number.parseInt", []ir.Type{ir.TypeString}, ir.TypeNumber, 1, 1)
	register([]string{"parseFloat", "Number.parseFloat"}, CategoryECMAScript, "__number.parseFloat", []ir.Type{ir.TypeString}, ir.TypeNumber, 1, 1)
	register([]string{"isNaN", "Number.isNaN"}, CategoryECMAScript, "__number.isNaN", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	register([]string{"isFinite", "Number.isFinite"}, CategoryECMAScript, "__number.isFinite", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	register([]string{"Number.isInteger"}, CategoryECMAScript, "__number.isInteger", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	register([]string{"JSON.parse"}, CategoryECMAScript, "__json.parse_string", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
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
				Op: ir.OpObjectNew, Type: ir.Type("object:RegExp"), Result: res, FieldCount: 2, Span: toIRSpan(call.Path, call.Expression.Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{res, patternVal}, Span: toIRSpan(call.Path, call.Expression.Span),
			})
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{res, flagsVal}, Span: toIRSpan(call.Path, call.Expression.Span),
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

	// Date globals (Category 1: ECMAScript)
	register([]string{"Date.now", "__date.now"}, CategoryECMAScript, "__date.now", nil, ir.TypeNumber, 0, 0)
	register([]string{"Date.parse", "__date.parse"}, CategoryECMAScript, "__date.parse", []ir.Type{ir.TypeString}, ir.TypeNumber, 1, 1)

	// Web-compatible globals (Category 2: WebCompat)
	register([]string{"btoa"}, CategoryWebCompat, "__web.btoa", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"atob"}, CategoryWebCompat, "__web.atob", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"performance.now"}, CategoryWebCompat, "__performance.now", nil, ir.TypeNumber, 0, 0)
	register([]string{"queueMicrotask"}, CategoryWebCompat, "__async.queueMicrotask", []ir.Type{ir.TypeClosure}, ir.TypeVoid, 1, 1)
	register([]string{"Promise.resolve"}, CategoryECMAScript, "__async.promise_resolve", []ir.Type{ir.TypeNumber}, ir.Type("object:Promise"), 1, 1)

	// Node-specific globals (Category 3: NodeGlobal)
	for _, logMethod := range []string{"log", "info", "debug", "warn", "error", "dir", "dirxml"} {
		name := "console." + logMethod
		m[name] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: name, ArgumentTypes: []ir.Type{ir.TypeNumber, ir.TypeBigInt, ir.TypeSymbol, ir.TypeString, ir.TypeBool, ir.TypeUnknown, ir.TypeNumberArray, ir.TypeStringArray, ir.TypeObject}, MinArgs: 0, MaxArgs: 256, Lower: lowerPrint}
	}
	m["console.assert"] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: "console.assert", MinArgs: 0, MaxArgs: 256, Lower: lowerConsoleAssert}
	m["console.group"] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: "console.group", MinArgs: 0, MaxArgs: 256, Lower: lowerConsoleGroup}
	m["console.groupCollapsed"] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: "console.groupCollapsed", MinArgs: 0, MaxArgs: 256, Lower: lowerConsoleGroup}
	m["console.timeLog"] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: "console.timeLog", MinArgs: 0, MaxArgs: 256, Lower: lowerConsoleTimeLog}
	m["console.trace"] = BuiltinIntrinsic{Category: CategoryNodeGlobal, Name: "console.trace", MinArgs: 0, MaxArgs: 256, Lower: lowerConsoleTrace}

	register([]string{"console.clear", "__console.clear"}, CategoryNodeGlobal, "__console.clear", nil, ir.TypeVoid, 0, 0)
	register([]string{"console.groupEnd", "__console.groupEnd"}, CategoryNodeGlobal, "__console.groupEnd", nil, ir.TypeVoid, 0, 0)
	register([]string{"console.count", "__console.count"}, CategoryNodeGlobal, "__console.count", []ir.Type{ir.TypeString}, ir.TypeVoid, 0, 1)
	register([]string{"console.countReset", "__console.countReset"}, CategoryNodeGlobal, "__console.countReset", []ir.Type{ir.TypeString}, ir.TypeVoid, 0, 1)
	register([]string{"console.time", "__console.time"}, CategoryNodeGlobal, "__console.time", []ir.Type{ir.TypeString}, ir.TypeVoid, 0, 1)
	register([]string{"console.timeEnd", "__console.timeEnd"}, CategoryNodeGlobal, "__console.timeEnd", []ir.Type{ir.TypeString}, ir.TypeVoid, 0, 1)
	register([]string{"process.exit", "__scriptgo.exit"}, CategoryNodeGlobal, "__process.exit", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	register([]string{"process.cwd", "__scriptgo.cwd"}, CategoryNodeGlobal, "__process.cwd", nil, ir.TypeString, 0, 0)

	// Node built-in modules (Category 4: NodeModule)
	register([]string{"fs.readFileSync", "__scriptgo.readFileSync", "readFileSync"}, CategoryNodeModule, "__fs.readFileSync", []ir.Type{ir.TypeString, ir.TypeString}, ir.TypeString, 1, 2)
	register([]string{"fs.writeFileSync", "__scriptgo.writeFileSync", "writeFileSync"}, CategoryNodeModule, "__fs.writeFileSync", []ir.Type{ir.TypeString}, ir.TypeVoid, 2, 2)
	register([]string{"fs.existsSync", "__scriptgo.existsSync", "existsSync"}, CategoryNodeModule, "__fs.existsSync", []ir.Type{ir.TypeString}, ir.TypeBool, 1, 1)
	register([]string{"fs.unlinkSync", "__scriptgo.unlinkSync", "unlinkSync"}, CategoryNodeModule, "__fs.unlinkSync", []ir.Type{ir.TypeString}, ir.TypeVoid, 1, 1)
	register([]string{"crypto.randomUUID", "__scriptgo.randomUUID", "randomUUID"}, CategoryNodeModule, "__crypto.randomUUID", nil, ir.TypeString, 0, 0)
	register([]string{"crypto.hashDigest", "__scriptgo.hashDigest", "hashDigest"}, CategoryNodeModule, "__crypto.hashDigest", []ir.Type{ir.TypeString, ir.TypeString, ir.TypeString}, ir.TypeString, 2, 3)
	register([]string{"os.platform", "__scriptgo.platform", "platform"}, CategoryNodeModule, "__os.platform", nil, ir.TypeString, 0, 0)
	register([]string{"os.arch", "__scriptgo.arch", "arch"}, CategoryNodeModule, "__os.arch", nil, ir.TypeString, 0, 0)
	register([]string{"os.homedir", "__scriptgo.homedir", "homedir"}, CategoryNodeModule, "__os.homedir", nil, ir.TypeString, 0, 0)
	register([]string{"os.uptime", "__scriptgo.uptime", "uptime"}, CategoryNodeModule, "__os.uptime", nil, ir.TypeNumber, 0, 0)
	register([]string{"os.totalmem", "__scriptgo.totalmem", "totalmem"}, CategoryNodeModule, "__os.totalmem", nil, ir.TypeNumber, 0, 0)
	register([]string{"os.freemem", "__scriptgo.freemem", "freemem"}, CategoryNodeModule, "__os.freemem", nil, ir.TypeNumber, 0, 0)
	register([]string{"os.type", "__scriptgo.type", "type"}, CategoryNodeModule, "__os.type", nil, ir.TypeString, 0, 0)
	register([]string{"os.release", "__scriptgo.release", "release"}, CategoryNodeModule, "__os.release", nil, ir.TypeString, 0, 0)
	register([]string{"os.tmpdir", "__scriptgo.tmpdir", "tmpdir"}, CategoryNodeModule, "__os.tmpdir", nil, ir.TypeString, 0, 0)

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
	if len(call.Expression.Arguments) < intrinsic.MinArgs || len(call.Expression.Arguments) > intrinsic.MaxArgs {
		return nil, nil, fmt.Errorf("builtin %s expects between %d and %d argument(s)", intrinsic.Name, intrinsic.MinArgs, intrinsic.MaxArgs)
	}
	args := make([]string, 0, len(call.Expression.Arguments))
	types := make([]ir.Type, 0, len(call.Expression.Arguments))
	for _, argument := range call.Expression.Arguments {
		value, typ, err := call.LowerExpression(call.Path, argument, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if err != nil {
			return nil, nil, err
		}
		accepted := slices.Contains(intrinsic.ArgumentTypes, typ)
		if !accepted {
			return nil, nil, fmt.Errorf("builtin %s does not support %s", intrinsic.Name, typ)
		}
		args = append(args, value)
		types = append(types, typ)
	}
	return args, types, nil
}
