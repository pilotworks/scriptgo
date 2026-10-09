package lowering

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
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
	Expression      *frontend.SyntaxExpression
	Result          string
	Function        *ir.Function
	Env             map[string]ir.Type
	Counter         *int
	Shapes          map[string]ir.ObjectShape
	Signatures      map[string]ir.Function
	LowerExpression lowerExpressionFunc
}

type lowerExpressionFunc func(string, *frontend.SyntaxExpression, string, *ir.Function, map[string]ir.Type, *int, map[string]ir.ObjectShape, map[string]ir.Function) (string, ir.Type, error)

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
	// Node.js v22 does not expose Symbol.metadata yet. Keep the property
	// undefined until the runtime supports the proposal natively.
	"Symbol.metadata":       {Category: CategoryECMAScript, Name: "Symbol.metadata", Type: ir.TypeVoid, Value: "undefined"},
	"Symbol.replace":        {Category: CategoryECMAScript, Name: "Symbol.replace", Type: ir.TypeSymbol, Value: "Symbol.replace"},
	"Symbol.search":         {Category: CategoryECMAScript, Name: "Symbol.search", Type: ir.TypeSymbol, Value: "Symbol.search"},
	"Symbol.species":        {Category: CategoryECMAScript, Name: "Symbol.species", Type: ir.TypeSymbol, Value: "Symbol.species"},
	"Symbol.split":          {Category: CategoryECMAScript, Name: "Symbol.split", Type: ir.TypeSymbol, Value: "Symbol.split"},
	"Symbol.toPrimitive":    {Category: CategoryECMAScript, Name: "Symbol.toPrimitive", Type: ir.TypeSymbol, Value: "Symbol.toPrimitive"},
	"Symbol.toStringTag":    {Category: CategoryECMAScript, Name: "Symbol.toStringTag", Type: ir.TypeSymbol, Value: "Symbol.toStringTag"},
	"Symbol.unscopables":    {Category: CategoryECMAScript, Name: "Symbol.unscopables", Type: ir.TypeSymbol, Value: "Symbol.unscopables"},
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

func lowerRandomFill(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	if len(call.Expression.Arguments) < 1 || len(call.Expression.Arguments) > 4 {
		return "", "", fmt.Errorf("builtin %s expects between 1 and 4 argument(s)", intrinsic.Name)
	}
	buffer, bufferType, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}
	if bufferType != ir.TypeBuffer && bufferType != ir.TypeUint8Array {
		return "", "", fmt.Errorf("builtin %s requires a buffer", intrinsic.Name)
	}
	var offset, size, callback string
	for i := 1; i < len(call.Expression.Arguments); i++ {
		value, typ, lowerErr := call.LowerExpression(call.Path, call.Expression.Arguments[i], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if lowerErr != nil {
			return "", "", lowerErr
		}
		if typ == ir.TypeClosure || strings.Contains(string(typ), "=>") {
			callback = value
			break
		}
		if typ != ir.TypeNumber {
			return "", "", fmt.Errorf("builtin %s expects numeric offset/size or callback", intrinsic.Name)
		}
		if offset == "" {
			offset = value
		} else if size == "" {
			size = value
		} else {
			return "", "", fmt.Errorf("builtin %s received too many numeric arguments", intrinsic.Name)
		}
	}
	for offset == "" || size == "" {
		zero := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zero, Value: "0", Span: toIRSpan(call.Path, call.Expression.Span)})
		if offset == "" {
			offset = zero
		} else {
			size = zero
		}
	}
	if callback == "" {
		undefined := nextTemp(call.Counter)
		call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeClosure, Result: undefined, Value: "undefined", Span: toIRSpan(call.Path, call.Expression.Span)})
		callback = undefined
	}
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeBuffer, Result: result, Callee: "__crypto.randomFill", Args: []string{buffer, offset, size, callback}, Span: toIRSpan(call.Path, call.Expression.Span)})
	return result, ir.TypeBuffer, nil
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

	registerMathIntrinsics(m)
	registerNumberStringIntrinsics(m)
	registerBigIntIntrinsics(m)
	registerRegExpSymbolIntrinsics(m)
	registerStructuredCloneIntrinsic(m)
	registerDateIntrinsics(m)
	registerTypedArrayIntrinsics(m)
	registerWebTimerIntrinsics(m)
	registerAtomicsIntrinsics(m)
	registerWeakRefIntrinsics(m)
	registerNodeGlobalIntrinsics(m)
	registerNetworkIntrinsics(m)
	registerObjectIntrinsics(m)
	registerBufferIntrinsics(m)
	registerArrayBuiltins(m)
	registerReflectIntrinsics(m)
	registerRequireObjectCoercibleIntrinsic(m)
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
		if len(intrinsic.ArgumentTypes) == 1 && intrinsic.ArgumentTypes[0] == ir.TypeNumber {
			// A number-only builtin applies ToNumber to its arguments.
			value, typ = coerceToNumber(call.Path, argument.Span, value, typ, call.Function, call.Counter)
		}
		if len(intrinsic.ArgumentTypes) > 0 && !slices.Contains(intrinsic.ArgumentTypes, typ) {
			return nil, nil, fmt.Errorf("builtin %s does not support %s", intrinsic.Name, typ)
		}
		args = append(args, value)
		types = append(types, typ)
	}
	return args, types, nil
}

// registerCallIntrinsic registers aliases that lower to a direct runtime call.
func registerCallIntrinsic(m map[string]BuiltinIntrinsic, aliases []string, cat BuiltinCategory, callee string, argTypes []ir.Type, retType ir.Type, minArgs, maxArgs int) {
	for _, name := range aliases {
		m[name] = BuiltinIntrinsic{Category: cat, Name: aliases[0], ArgumentTypes: argTypes, MinArgs: minArgs, MaxArgs: maxArgs, Lower: lowerCall(callee, retType)}
	}
}

// registerCustomIntrinsic registers aliases with a custom lowering function.
func registerCustomIntrinsic(m map[string]BuiltinIntrinsic, aliases []string, cat BuiltinCategory, callee string, retType ir.Type, minArgs, maxArgs int, lower func(IntrinsicCall, BuiltinIntrinsic) (string, ir.Type, error)) {
	for _, name := range aliases {
		m[name] = BuiltinIntrinsic{Category: cat, Name: aliases[0], MinArgs: minArgs, MaxArgs: maxArgs, Lower: lower}
	}
}
