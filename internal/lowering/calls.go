package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerCallExpression(
	path string,
	expression *typescriptgo.SyntaxExpression,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, error) {
	if expression.Left != nil && expression.Left.Kind == "property" && expression.Left.Left != nil {
		methodName := expression.Left.Text
		receiver, receiverType, err := lowerExpression(path, expression.Left.Left, "", function, env, counter, shapes, signatures)
		if err == nil {
			if receiverType == "object:RegExp" {
				if methodName == "test" && len(expression.Arguments) > 0 {
					argVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					srcVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: srcVal, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
					flagsVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: flagsVal, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeBool, Result: result, Callee: "__regex.test", Args: []string{srcVal, flagsVal, argVal}, Span: toIRSpan(path, expression.Span)})
					return result, ir.TypeBool, nil
				}
				if methodName == "exec" && len(expression.Arguments) > 0 {
					argVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					srcVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: srcVal, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
					flagsVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: flagsVal, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeStringArray, Result: result, Callee: "__regex.exec", Args: []string{srcVal, flagsVal, argVal}, Span: toIRSpan(path, expression.Span)})
					return result, ir.TypeStringArray, nil
				}
			}
			if receiverType == ir.TypeBigInt && methodName == "toString" {
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__string.fromBigInt", Args: []string{receiver}, Span: toIRSpan(path, expression.Span)})
				return result, ir.TypeString, nil
			}
			if receiverType == ir.TypeString && isStringMethod(methodName) {
				if methodName == "match" || methodName == "search" {
					if len(expression.Arguments) > 0 {
						argVal, argTyp, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
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
						if methodName == "match" {
							function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeStringArray, Result: result, Callee: "__string.match", Args: []string{receiver, srcVal, flagsVal}, Span: toIRSpan(path, expression.Span)})
							return result, ir.TypeStringArray, nil
						} else {
							function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: "__string.search", Args: []string{receiver, srcVal, flagsVal}, Span: toIRSpan(path, expression.Span)})
							return result, ir.TypeNumber, nil
						}
					}
				}
				if methodName == "replace" && len(expression.Arguments) >= 2 {
					arg0Val, arg0Typ, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					if arg0Typ == "object:RegExp" {
						arg1Val, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						srcVal := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: srcVal, Callee: "RegExp", Field: "source", FieldIndex: 0, Args: []string{arg0Val}, Span: toIRSpan(path, expression.Span)})
						flagsVal := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: ir.TypeString, Result: flagsVal, Callee: "RegExp", Field: "flags", FieldIndex: 1, Args: []string{arg0Val}, Span: toIRSpan(path, expression.Span)})
						if result == "" {
							result = nextTemp(counter)
						}
						function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__string.replace_regex", Args: []string{receiver, srcVal, flagsVal, arg1Val}, Span: toIRSpan(path, expression.Span)})
						return result, ir.TypeString, nil
					}
				}
				args := []string{receiver}
				for _, argument := range expression.Arguments {
					value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					args = append(args, value)
				}
				if result == "" {
					result = nextTemp(counter)
				}
				returnType := ir.TypeNumber
				switch methodName {
				case "slice", "trim", "trimStart", "trimEnd", "replace", "replaceAll", "substring", "charAt", "toLowerCase", "toUpperCase", "repeat", "padStart", "padEnd", "concat":
					returnType = ir.TypeString
				case "startsWith", "endsWith", "includes":
					returnType = ir.TypeBool
				case "split":
					returnType = ir.TypeStringArray
				case "indexOf", "lastIndexOf", "charCodeAt":
					returnType = ir.TypeNumber
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: "__string." + methodName, Args: args, Span: toIRSpan(path, expression.Span)})
				return result, returnType, nil
			}
			if receiverType == ir.TypeNumber && (methodName == "toFixed" || methodName == "toString") {
				args := []string{receiver}
				for _, argument := range expression.Arguments {
					value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					args = append(args, value)
				}
				if result == "" {
					result = nextTemp(counter)
				}
				callee := "__number." + methodName
				if methodName == "toString" {
					callee = "__string.fromNumber"
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: callee, Args: args, Span: toIRSpan(path, expression.Span)})
				return result, ir.TypeString, nil
			}
			if (receiverType == ir.TypeNumberArray || receiverType == ir.TypeStringArray) && isArrayMethod(methodName) {
				args := []string{receiver}
				for _, argument := range expression.Arguments {
					value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					args = append(args, value)
				}
				if result == "" {
					result = nextTemp(counter)
				}
				returnType := ir.TypeNumber
				switch methodName {
				case "slice", "reverse", "concat", "splice", "map", "filter":
					returnType = receiverType
				case "includes", "some", "every":
					returnType = ir.TypeBool
				case "join":
					returnType = ir.TypeString
				case "push", "unshift", "indexOf", "reduce":
					returnType = ir.TypeNumber
				case "pop", "shift", "at", "find":
					if receiverType == ir.TypeNumberArray {
						returnType = ir.TypeNumber
					} else {
						returnType = ir.TypeString
					}
				case "forEach":
					returnType = ir.TypeVoid
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: "__array." + methodName, Args: args, Span: toIRSpan(path, expression.Span)})
				return result, returnType, nil
			}
			if strings.HasPrefix(string(receiverType), "object:") {
				className := strings.TrimPrefix(string(receiverType), "object:")
				if target, mangled, ok := findMethodInHierarchy(className, methodName, signatures, classHierarchy); ok {
					args := []string{receiver}
					for _, argument := range expression.Arguments {
						argVal, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
						if err != nil {
							return "", "", err
						}
						args = append(args, argVal)
					}
					if result == "" {
						result = nextTemp(counter)
					}

					overrides := findOverridingSubclasses(className, methodName, classHierarchy, signatures)
					if len(overrides) == 0 {
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpCall,
							Type:   target.ReturnType,
							Result: result,
							Callee: mangled,
							Args:   args,
							Span:   toIRSpan(path, expression.Span),
						})
						return result, target.ReturnType, nil
					}

					if target.ReturnType != ir.TypeVoid {
						env[result] = target.ReturnType
					}
					var elseBody []ir.Instruction
					baseCallRes := nextTemp(counter)
					elseBody = append(elseBody, ir.Instruction{
						Op:     ir.OpCall,
						Type:   target.ReturnType,
						Result: baseCallRes,
						Callee: mangled,
						Args:   args,
						Span:   toIRSpan(path, expression.Span),
					})
					if target.ReturnType != ir.TypeVoid {
						elseBody = append(elseBody, ir.Instruction{
							Op:     ir.OpAssign,
							Type:   target.ReturnType,
							Result: result,
							Args:   []string{baseCallRes},
							Span:   toIRSpan(path, expression.Span),
						})
					}

					currElse := elseBody
					for _, sub := range overrides {
						subMangled := sub + "_" + methodName
						instCheck := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpInstanceOf,
							Type:   ir.TypeBool,
							Result: instCheck,
							Args:   []string{receiver},
							Value:  sub,
							Span:   toIRSpan(path, expression.Span),
						})

						var thenBody []ir.Instruction
						subCallRes := nextTemp(counter)
						thenBody = append(thenBody, ir.Instruction{
							Op:     ir.OpCall,
							Type:   target.ReturnType,
							Result: subCallRes,
							Callee: subMangled,
							Args:   args,
							Span:   toIRSpan(path, expression.Span),
						})
						if target.ReturnType != ir.TypeVoid {
							thenBody = append(thenBody, ir.Instruction{
								Op:     ir.OpAssign,
								Type:   target.ReturnType,
								Result: result,
								Args:   []string{subCallRes},
								Span:   toIRSpan(path, expression.Span),
							})
						}

						ifInst := ir.Instruction{
							Op:   ir.OpIf,
							Type: ir.TypeVoid,
							Args: []string{instCheck},
							Then: thenBody,
							Else: currElse,
							Span: toIRSpan(path, expression.Span),
						}
						currElse = []ir.Instruction{ifInst}
					}
					function.Body = append(function.Body, currElse...)
					return result, target.ReturnType, nil
				}
			}
		}
	}

	// super(...) call in constructor
	if expression.Left != nil && expression.Left.Kind == "identifier" && expression.Left.Text == "super" {
		thisType, ok := env["this"]
		if !ok {
			return "", "", fmt.Errorf("super() can only be used inside a class constructor")
		}
		currentClass := strings.TrimPrefix(string(thisType), "object:")
		meta := classHierarchy[currentClass]
		if meta.Extends == "" {
			return "", "", fmt.Errorf("super() called in class %q with no base class", currentClass)
		}
		ctor, ctorName, found := findConstructorInHierarchy(meta.Extends, signatures, classHierarchy)
		if !found {
			if len(expression.Arguments) == 0 {
				return "", ir.TypeVoid, nil
			}
			return "", "", fmt.Errorf("super constructor not found for base class %q", meta.Extends)
		}
		args := []string{"this"}
		for _, argument := range expression.Arguments {
			val, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			args = append(args, val)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ctor.ReturnType,
			Callee: ctorName,
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return "", ir.TypeVoid, nil
	}

	// super.method(...) call
	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "member") && expression.Left.Left != nil && expression.Left.Left.Text == "super" {
		thisType, ok := env["this"]
		if !ok {
			return "", "", fmt.Errorf("super.method() can only be used inside a class method")
		}
		currentClass := strings.TrimPrefix(string(thisType), "object:")
		meta := classHierarchy[currentClass]
		if meta.Extends == "" {
			return "", "", fmt.Errorf("super.%s called in class %q with no base class", expression.Left.Text, currentClass)
		}
		target, mangled, found := findMethodInHierarchy(meta.Extends, expression.Left.Text, signatures, classHierarchy)
		if !found {
			return "", "", fmt.Errorf("super method %q not found in base class %q", expression.Left.Text, meta.Extends)
		}
		args := []string{"this"}
		for _, argument := range expression.Arguments {
			val, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			args = append(args, val)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   target.ReturnType,
			Result: result,
			Callee: mangled,
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, target.ReturnType, nil
	}

	// Static method call ClassName.method(...)
	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "member") && expression.Left.Left != nil && expression.Left.Left.Kind == "identifier" {
		className := expression.Left.Left.Text
		methodName := expression.Left.Text
		if target, mangled, found := findMethodInHierarchy(className, methodName, signatures, classHierarchy); found && (len(target.Parameters) == 0 || target.Parameters[0].Name != "this") {
			args := make([]string, 0, len(expression.Arguments))
			for _, argument := range expression.Arguments {
				val, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				args = append(args, val)
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   target.ReturnType,
				Result: result,
				Callee: mangled,
				Args:   args,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, target.ReturnType, nil
		}
	}

	if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "member") && expression.Left.Text == "then" {
		receiverVal, _, err := lowerExpression(path, expression.Left.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		cbVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.Type("object:Promise"),
			Result: result,
			Callee: "__async.promise_then",
			Args:   []string{receiverVal, cbVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.Type("object:Promise"), nil
	}

	callee := callName(expression.Left)
	if calleeType, ok := env[callee]; ok && calleeType == ir.TypeClosure {
		args := make([]string, 0, len(expression.Arguments))
		for _, argument := range expression.Arguments {
			value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			args = append(args, value)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		retType := ir.TypeNumber
		if rt, ok := env[callee+".retType"]; ok && rt != "" {
			retType = rt
		} else if target, ok := signatures[callee]; ok && target.ReturnType != "" {
			retType = target.ReturnType
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpClosureCall,
			Type:   retType,
			Result: result,
			Callee: callee,
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, nil
	}
	if intrinsic, ok := builtinIntrinsic(callee); ok {
		return intrinsic.Lower(IntrinsicCall{Path: path, Expression: expression, Result: result, Function: function, Env: env, Counter: counter, Shapes: shapes, Signatures: signatures, LowerExpression: lowerExpression}, intrinsic)
	}
	if callee == "" {
		return "", "", fmt.Errorf("unsupported call target")
	}
	target, ok := signatures[callee]
	if !ok {
		return "", "", fmt.Errorf("unknown function %q", callee)
	}
	callee = target.Name
	args := make([]string, 0, len(expression.Arguments))
	for i, argument := range expression.Arguments {
		value, valType, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if i < len(target.Parameters) && target.Parameters[i].Type == ir.TypeUnknown && valType != ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: boxed,
				Args:   []string{value},
				Span:   toIRSpan(path, argument.Span),
			})
			value = boxed
		}
		args = append(args, value)
	}
	if len(args) < len(target.Parameters) {
		defaults := defaultParamsIndex[callee]
		if defaults != nil {
			for i := len(args); i < len(target.Parameters); i++ {
				if initExpr, ok := defaults[i]; ok {
					val, _, err := lowerExpression(path, initExpr, "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					args = append(args, val)
				}
			}
		}
	}
	if restParamsIndex[callee] && len(target.Parameters) > 0 && (target.Parameters[len(target.Parameters)-1].Type == ir.TypeStringArray || target.Parameters[len(target.Parameters)-1].Type == ir.TypeNumberArray) {
		restType := target.Parameters[len(target.Parameters)-1].Type
		fixed := len(target.Parameters) - 1
		if len(args) < fixed {
			return "", "", fmt.Errorf("call to %q has too few arguments", callee)
		}
		restArgs := append([]string(nil), args[fixed:]...)
		args = args[:fixed]
		array := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: restType, Result: array, Args: restArgs, Span: toIRSpan(path, expression.Span)})
		args = append(args, array)
	}
	if result == "" {
		result = nextTemp(counter)
	}
	returnType := target.ReturnType
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: callee, Args: args, Span: toIRSpan(path, expression.Span)})
	return result, returnType, nil
}

func callName(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil {
		return ""
	}
	if expression.Kind == "identifier" {
		return expression.Text
	}
	if expression.Kind == "property" && expression.Left != nil && expression.Left.Kind == "identifier" {
		return expression.Left.Text + "." + expression.Text
	}
	return ""
}

func isStringMethod(name string) bool {
	switch name {
	case "indexOf", "lastIndexOf", "slice", "startsWith", "endsWith", "trim", "trimStart", "trimEnd", "replace", "replaceAll", "substring", "split", "charAt", "charCodeAt", "includes", "toLowerCase", "toUpperCase", "repeat", "padStart", "padEnd", "concat", "match", "search":
		return true
	default:
		return false
	}
}

func isArrayMethod(name string) bool {
	switch name {
	case "push", "pop", "slice", "indexOf", "includes", "join", "reverse", "concat", "shift", "unshift", "splice", "at", "map", "filter", "forEach", "reduce", "find", "some", "every":
		return true
	default:
		return false
	}
}

func stringMethod(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil || expression.Kind != "property" || expression.Left == nil {
		return ""
	}
	if isStringMethod(expression.Text) {
		return expression.Text
	}
	return ""
}

func arrayMethod(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil || expression.Kind != "property" || expression.Left == nil {
		return ""
	}
	if isArrayMethod(expression.Text) {
		return expression.Text
	}
	return ""
}
