package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerStringReceiverMethod(
	path string,
	expression *frontend.SyntaxExpression,
	receiver string,
	methodName string,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, bool, error) {
	if !isStringMethod(methodName) {
		return "", "", false, nil
	}
	if (methodName == "toString" || methodName == "valueOf") && len(expression.Arguments) == 0 {
		return receiver, ir.TypeString, true, nil
	}
	if methodName == "match" || methodName == "search" || methodName == "matchAll" {
		if len(expression.Arguments) > 0 {
			argVal, argTyp, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			srcVal := argVal
			flagsVal := nextTemp(counter)
			if argTyp == "object:RegExp" {
				srcVal = nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: srcVal, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{argVal}, Span: toIRSpan(path, expression.Span)})
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: flagsVal, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{argVal}, Span: toIRSpan(path, expression.Span)})
			} else {
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: flagsVal, Value: "", Span: toIRSpan(path, expression.Span)})
			}
			if result == "" {
				result = nextTemp(counter)
			}
			switch methodName {
			case "match":
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeStringArray, Result: result, Callee: "__string.match", Args: []string{receiver, srcVal, flagsVal}, Span: toIRSpan(path, expression.Span)})
				return result, ir.TypeStringArray, true, nil
			case "matchAll":
				// Each element is a match array (RegExpExecArray).
				matches := ir.Type(string(ir.TypeStringArray) + "[]")
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: matches, Result: result, Callee: "__string.matchAll", Args: []string{receiver, srcVal, flagsVal}, Span: toIRSpan(path, expression.Span)})
				return result, matches, true, nil
			default:
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: "__string.search", Args: []string{receiver, srcVal, flagsVal}, Span: toIRSpan(path, expression.Span)})
				return result, ir.TypeNumber, true, nil
			}
		}
	}
	if (methodName == "replace" || methodName == "replaceAll") && len(expression.Arguments) >= 2 && isRegExpExpression(expression.Arguments[0]) {
		arg0Val, arg0Typ, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		if arg0Typ == "object:RegExp" {
			arg1Val, arg1Typ, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			srcVal := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: srcVal, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{arg0Val}, Span: toIRSpan(path, expression.Span)})
			flagsVal := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: flagsVal, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{arg0Val}, Span: toIRSpan(path, expression.Span)})
			if result == "" {
				result = nextTemp(counter)
			}
			// A function replacement is called per match; anything else is a
			// replacement string with $ patterns.
			callee := "__string.replace_regex"
			if methodName == "replaceAll" {
				callee = "__string.replaceAll_regex"
			}
			if arg1Typ == ir.TypeClosure || strings.Contains(string(arg1Typ), "=>") {
				callee += "_fn"
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: callee, Args: []string{receiver, srcVal, flagsVal, arg1Val}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeString, true, nil
		}
	}
	if methodName == "split" && len(expression.Arguments) > 0 {
		arg0Val, arg0Typ, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		sepVal := arg0Val
		callee := "__string.split"
		splitArgs := []string{receiver, sepVal}
		if arg0Typ == "object:RegExp" {
			// A RegExp separator splits at its matches, with captures.
			callee = "__string.split_regex"
			srcVal := nextTemp(counter)
			flagsVal := nextTemp(counter)
			function.Body = append(function.Body,
				ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: srcVal, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{arg0Val}, Span: toIRSpan(path, expression.Span)},
				ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: flagsVal, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{arg0Val}, Span: toIRSpan(path, expression.Span)},
			)
			splitArgs = []string{receiver, srcVal, flagsVal}
		}
		if len(expression.Arguments) > 1 {
			limVal, limType, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			limVal, _, present := coerceMethodArgument(path, expression.Arguments[1], "string", methodName, 1, len(expression.Arguments), limVal, limType, function, counter)
			if present {
				splitArgs = append(splitArgs, limVal)
			}
		}
		if result == "" {
			result = nextTemp(counter)
		}
		env[result] = ir.TypeStringArray
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeStringArray, Result: result, Callee: callee, Args: splitArgs, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeStringArray, true, nil
	}
	args := []string{receiver}
	for index, argument := range expression.Arguments {
		value, typ, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		value, _, present := coerceMethodArgument(path, argument, "string", methodName, index, len(expression.Arguments), value, typ, function, counter)
		if present {
			args = append(args, value)
		}
	}
	if result == "" {
		result = nextTemp(counter)
	}
	returnType := ir.TypeNumber
	switch methodName {
	case "slice", "trim", "trimStart", "trimEnd", "trimLeft", "trimRight", "replace", "replaceAll", "substring", "substr", "charAt", "at", "toLowerCase", "toUpperCase", "toLocaleLowerCase", "toLocaleUpperCase", "repeat", "padStart", "padEnd", "concat", "toWellFormed", "normalize", "valueOf", "toString", "anchor", "big", "blink", "bold", "fixed", "fontcolor", "fontsize", "italics", "link", "small", "strike", "sub", "sup":
		returnType = ir.TypeString
	case "startsWith", "endsWith", "includes", "isWellFormed":
		returnType = ir.TypeBool
	case "split", "matchAll":
		returnType = ir.TypeStringArray
	case "indexOf", "lastIndexOf", "charCodeAt", "codePointAt", "localeCompare":
		returnType = ir.TypeNumber
	}
	if result == "" {
		result = nextTemp(counter)
	}
	env[result] = returnType
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: "__string." + methodName, Args: args, Span: toIRSpan(path, expression.Span)})
	return result, returnType, true, nil
}

// isRegExpExpression reports whether an argument is a RegExp (a literal or a
// RegExp-typed value) from its syntax and checked type, before lowering it.
func isRegExpExpression(expression *frontend.SyntaxExpression) bool {
	if expression == nil {
		return false
	}
	return expression.Kind == "regex" || strings.TrimSpace(expression.InferredType) == "RegExp"
}
