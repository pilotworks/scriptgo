package lowering

import (
	"slices"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// numericMethodArguments lists, per receiver kind and method, the argument
// positions the specification converts with ToNumber/ToIntegerOrInfinity
// (indices, counts, lengths). Typed code passes numbers already; code that
// opts out of type checking may pass any primitive.
var numericMethodArguments = map[string]map[string][]int{
	"string": {
		"slice": {0, 1}, "substring": {0, 1}, "substr": {0, 1}, "charAt": {0}, "at": {0},
		"charCodeAt": {0}, "codePointAt": {0}, "repeat": {0}, "padStart": {0}, "padEnd": {0},
		"indexOf": {1}, "lastIndexOf": {1}, "includes": {1}, "startsWith": {1}, "endsWith": {1},
		"split": {1},
	},
	"array": {
		"slice": {0, 1}, "at": {0}, "indexOf": {1}, "lastIndexOf": {1}, "includes": {1},
		"fill": {1, 2}, "copyWithin": {0, 1, 2}, "with": {0}, "splice": {0, 1}, "toSpliced": {0, 1}, "flat": {0},
	},
}

// omittedWhenUndefined lists end/length positions where an undefined
// argument means "not present" (to the end of the receiver), unlike other
// index positions where ToIntegerOrInfinity(undefined) is 0.
var omittedWhenUndefined = map[string]map[string]int{
	"string": {"slice": 1, "substring": 1, "substr": 1, "split": 1},
	"array":  {"slice": 1, "fill": 2, "copyWithin": 2},
}

// coerceMethodArgument converts a numeric argument position of a string or
// array method that is not already a number: undefined is 0 (or, as the
// trailing end/length argument, dropped as not present), and other primitives
// go through ToNumber. It reports false when the argument is dropped.
func coerceMethodArgument(path string, argument *frontend.SyntaxExpression, receiverKind, method string, position, count int, value string, typ ir.Type, function *ir.Function, counter *int) (string, ir.Type, bool) {
	if receiverKind == "array" && position == 0 {
		if optional, ok := callbackMethods[method]; ok {
			if optional && typ == ir.TypeVoid {
				// sort(undefined) uses the default comparison.
				return "", "", false
			}
			if !mayBeCallable(typ) {
				return throwNotCallable(path, argument, function, counter), ir.TypeClosure, true
			}
			return value, typ, true
		}
	}
	if slices.Contains(stringMethodArguments[receiverKind][method], position) {
		if typ == ir.TypeVoid && (method == "padStart" || method == "padEnd") {
			// An undefined fill string means the default " ".
			return "", "", false
		}
		value, typ = coercePrimitiveToString(path, argument.Span, value, typ, function, counter)
		return value, typ, true
	}
	if !slices.Contains(numericMethodArguments[receiverKind][method], position) {
		return value, typ, true
	}
	if typ == ir.TypeVoid {
		if end, ok := omittedWhenUndefined[receiverKind][method]; ok && end == position && position == count-1 {
			return "", "", false
		}
		zero := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zero, Value: "0", Span: toIRSpan(path, argument.Span)})
		return zero, ir.TypeNumber, true
	}
	value, typ = coerceToNumber(path, argument.Span, value, typ, function, counter)
	return value, typ, true
}

// coerceToNumber lowers ToNumber of a primitive or boxed value. Numbers pass
// through unchanged, so typed programs emit the same IR as before.
func coerceToNumber(path string, span frontend.SourceSpan, value string, typ ir.Type, function *ir.Function, counter *int) (string, ir.Type) {
	if typ == ir.TypeNumber {
		return value, typ
	}
	irSpan := toIRSpan(path, span)
	switch {
	case typ == ir.TypeSymbol:
		// ToNumber(symbol) throws.
		message := nextTemp(counter)
		function.Body = append(function.Body,
			ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: message, Value: "TypeError: Cannot convert a Symbol value to a number", StringLiteral: true, Span: irSpan},
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__error.throw", Args: []string{message}, Span: irSpan},
		)
		nan := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: nan, Value: "NaN", Span: irSpan})
		return nan, ir.TypeNumber
	case typ == ir.TypeBool, typ == ir.TypeString, typ == ir.TypeVoid, typ == ir.TypeUnknown, typ == ir.TypeBigInt, isPointerLikeType(typ):
		// Objects convert through their string form (arrays and errors match
		// ToPrimitive); a user valueOf is not modelled, as for String(obj).
	default:
		return value, typ
	}
	number := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: number, Callee: "__number.new", Args: []string{value}, Span: irSpan})
	return number, ir.TypeNumber
}

// callbackMethods are array methods whose first argument must be callable;
// the value reports whether undefined is allowed (a default comparator).
var callbackMethods = map[string]bool{
	"map": false, "filter": false, "forEach": false, "reduce": false, "reduceRight": false,
	"find": false, "findLast": false, "findIndex": false, "findLastIndex": false,
	"some": false, "every": false, "flatMap": false, "sort": true, "toSorted": true,
}

// stringMethodArguments lists string-valued argument positions (search and
// fill strings) that apply ToString to a primitive argument.
var stringMethodArguments = map[string]map[string][]int{
	"string": {
		"indexOf": {0}, "lastIndexOf": {0}, "includes": {0}, "startsWith": {0}, "endsWith": {0},
		"padStart": {1}, "padEnd": {1}, "localeCompare": {0},
	},
}

// mayBeCallable reports storage that can hold a function: closures, boxed
// values, and object references (checked at run time by the callee).
func mayBeCallable(typ ir.Type) bool {
	switch typ {
	case ir.TypeBool, ir.TypeNumber, ir.TypeString, ir.TypeBigInt, ir.TypeSymbol, ir.TypeVoid:
		return false
	}
	return true
}

// throwNotCallable raises the TypeError a non-callable callback causes and
// returns a closure placeholder so the (unreachable) call stays well typed.
func throwNotCallable(path string, argument *frontend.SyntaxExpression, function *ir.Function, counter *int) string {
	irSpan := toIRSpan(path, argument.Span)
	message := nextTemp(counter)
	placeholder := nextTemp(counter)
	function.Body = append(function.Body,
		ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: message, Value: "TypeError: callback is not a function", StringLiteral: true, Span: irSpan},
		ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__error.throw", Args: []string{message}, Span: irSpan},
		ir.Instruction{Op: ir.OpConst, Type: ir.TypeClosure, Result: placeholder, Value: "null", Span: irSpan},
	)
	return placeholder
}

// coercePrimitiveToString lowers ToString of a primitive argument; strings
// and references pass through unchanged.
func coercePrimitiveToString(path string, span frontend.SourceSpan, value string, typ ir.Type, function *ir.Function, counter *int) (string, ir.Type) {
	irSpan := toIRSpan(path, span)
	callee := ""
	switch typ {
	case ir.TypeNumber:
		callee = "__string.fromNumber"
	case ir.TypeBool:
		callee = "__string.fromBool"
	case ir.TypeBigInt:
		callee = "__string.fromBigInt"
	case ir.TypeVoid:
		text := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: text, Value: "undefined", StringLiteral: true, Span: irSpan})
		return text, ir.TypeString
	default:
		return value, typ
	}
	text := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: text, Callee: callee, Args: []string{value}, Span: irSpan})
	return text, ir.TypeString
}
