package lowering

import (
	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerStringReceiverMethod(
	path string,
	expression *typescriptgo.SyntaxExpression,
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
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeStringArray, Result: result, Callee: "__string.matchAll", Args: []string{receiver, srcVal, flagsVal}, Span: toIRSpan(path, expression.Span)})
				return result, ir.TypeStringArray, true, nil
			default:
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: "__string.search", Args: []string{receiver, srcVal, flagsVal}, Span: toIRSpan(path, expression.Span)})
				return result, ir.TypeNumber, true, nil
			}
		}
	}
	if methodName == "replace" && len(expression.Arguments) >= 2 {
		arg0Val, arg0Typ, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		if arg0Typ == "object:RegExp" {
			arg1Val, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
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
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__string.replace_regex", Args: []string{receiver, srcVal, flagsVal, arg1Val}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeString, true, nil
		}
	}
	if methodName == "split" && len(expression.Arguments) > 0 {
		arg0Val, arg0Typ, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		sepVal := arg0Val
		if arg0Typ == "object:RegExp" {
			sepVal = nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: sepVal, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{arg0Val}, Span: toIRSpan(path, expression.Span)})
		}
		splitArgs := []string{receiver, sepVal}
		if len(expression.Arguments) > 1 {
			limVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			splitArgs = append(splitArgs, limVal)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		env[result] = ir.TypeStringArray
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeStringArray, Result: result, Callee: "__string.split", Args: splitArgs, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeStringArray, true, nil
	}
	args := []string{receiver}
	for _, argument := range expression.Arguments {
		value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		args = append(args, value)
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
