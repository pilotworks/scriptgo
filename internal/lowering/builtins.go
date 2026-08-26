package lowering

import (
	"fmt"
	"slices"
	"strings"

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

func parseAnonymousObjectShape(shapeName string) (ir.ObjectShape, bool) {
	clean := strings.TrimSpace(shapeName)
	clean = strings.TrimPrefix(clean, "object:")
	if !strings.HasPrefix(clean, "{") || !strings.HasSuffix(clean, "}") {
		return ir.ObjectShape{}, false
	}
	inner := strings.Trim(clean, "{}")
	var fields []ir.Field
	rawParts := strings.FieldsFunc(inner, func(r rune) bool {
		return r == ';' || r == ','
	})
	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		fName := strings.TrimSpace(kv[0])
		fType := strings.TrimSpace(kv[1])
		fields = append(fields, ir.Field{
			Name: fName,
			Type: toIRType(fType),
		})
	}
	if len(fields) == 0 {
		return ir.ObjectShape{}, false
	}
	return ir.ObjectShape{
		Name:   shapeName,
		Fields: fields,
	}, true
}

func lowerJSONStringify(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) != 1 {
		return "", "", fmt.Errorf("JSON.stringify expects exactly 1 argument")
	}
	argVal, argType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}
	shapeName := string(argType)
	if after, ok := strings.CutPrefix(string(argType), "object:"); ok {
		shapeName = after
	}
	if shape, ok := call.Shapes[shapeName]; ok {
		return lowerJSONStringifyObject(call, argVal, shape)
	}
	if parsedShape, ok := parseAnonymousObjectShape(shapeName); ok {
		call.Shapes[shapeName] = parsedShape
		return lowerJSONStringifyObject(call, argVal, parsedShape)
	}
	if strings.HasPrefix(string(argType), "object:") {
		return "", "", fmt.Errorf("unknown shape %q for JSON.stringify", shapeName)
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
	case ir.TypeUnknown:
		callee = "__json.stringify_unknown"
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
	register([]string{"btoa"}, CategoryWebCompat, "__web.btoa", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"atob"}, CategoryWebCompat, "__web.atob", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	register([]string{"performance.now"}, CategoryWebCompat, "__performance.now", nil, ir.TypeNumber, 0, 0)
	register([]string{"queueMicrotask"}, CategoryWebCompat, "__async.queueMicrotask", []ir.Type{ir.TypeClosure}, ir.TypeVoid, 1, 1)
	register([]string{"Promise.resolve"}, CategoryECMAScript, "__async.promise_resolve", []ir.Type{ir.TypeNumber}, ir.Type("object:Promise"), 1, 1)
	register([]string{"Promise.all"}, CategoryECMAScript, "__async.promise_all", []ir.Type{ir.TypeObject}, ir.Type("object:Promise"), 1, 1)
	register([]string{"setTimeout", "__scriptgo.setTimeout", "timers.setTimeout"}, CategoryWebCompat, "__timers.setTimeout", []ir.Type{ir.TypeClosure, ir.TypeNumber}, ir.TypeNumber, 1, 2)
	register([]string{"clearTimeout", "__scriptgo.clearTimeout", "timers.clearTimeout"}, CategoryWebCompat, "__timers.clearTimeout", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	register([]string{"setInterval", "__scriptgo.setInterval", "timers.setInterval"}, CategoryWebCompat, "__timers.setInterval", []ir.Type{ir.TypeClosure, ir.TypeNumber}, ir.TypeNumber, 1, 2)
	register([]string{"clearInterval", "__scriptgo.clearInterval", "timers.clearInterval"}, CategoryWebCompat, "__timers.clearInterval", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	register([]string{"setImmediate", "__scriptgo.setImmediate", "timers.setImmediate"}, CategoryWebCompat, "__timers.setImmediate", []ir.Type{ir.TypeClosure}, ir.TypeNumber, 1, 1)
	register([]string{"clearImmediate", "__scriptgo.clearImmediate", "timers.clearImmediate"}, CategoryWebCompat, "__timers.clearImmediate", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)

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

	register([]string{"__scriptgo.readFileSync"}, CategoryNodeModule, "__fs.readFileSync", []ir.Type{ir.TypeString, ir.TypeString}, ir.TypeString, 1, 2)
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
	register([]string{"__scriptgo.execSync"}, CategoryNodeModule, "__child_process.execSync", []ir.Type{ir.TypeString, ir.TypeString, ir.TypeString}, ir.TypeString, 1, 3)
	register([]string{"__scriptgo.spawnSync"}, CategoryNodeModule, "__child_process.spawnSync", []ir.Type{ir.TypeString, ir.TypeStringArray, ir.TypeString, ir.TypeString}, ir.Type("object:SpawnSyncReturns"), 1, 4)
	register([]string{"__scriptgo.fetchSync"}, CategoryNodeModule, "__http.fetchSync", []ir.Type{ir.TypeString, ir.TypeString, ir.TypeStringArray, ir.TypeString}, ir.Type("object:FetchResponseData"), 1, 4)
	register([]string{"__stream.getDefaultHighWaterMark", "__scriptgo.streamGetDefaultHighWaterMark"}, CategoryNodeModule, "__stream.getDefaultHighWaterMark", []ir.Type{ir.TypeBool}, ir.TypeNumber, 0, 1)
	register([]string{"__stream.setDefaultHighWaterMark", "__scriptgo.streamSetDefaultHighWaterMark"}, CategoryNodeModule, "__stream.setDefaultHighWaterMark", []ir.Type{ir.TypeBool, ir.TypeNumber}, ir.TypeVoid, 2, 2)

	registerObjectIntrinsics(m)
	registerBufferIntrinsics(m)
	registerArrayBuiltins(m)
	registerReflectIntrinsics(m)
	registerIteratorBuiltins(m)

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
