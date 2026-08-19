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
			if receiverType == ir.TypeString && isStringMethod(methodName) {
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
				case "slice", "reverse", "concat", "splice":
					returnType = receiverType
				case "includes":
					returnType = ir.TypeBool
				case "join":
					returnType = ir.TypeString
				case "push", "unshift", "indexOf":
					returnType = ir.TypeNumber
				case "pop", "shift", "at":
					if receiverType == ir.TypeNumberArray {
						returnType = ir.TypeNumber
					} else {
						returnType = ir.TypeString
					}
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

	callee := callName(expression.Left)
	if intrinsic, ok := builtinIntrinsic(callee); ok {
		return intrinsic.Lower(IntrinsicCall{Path: path, Expression: expression, Result: result, Function: function, Env: env, Counter: counter, Shapes: shapes, Signatures: signatures, LowerExpression: lowerExpression}, intrinsic)
	}
	if callee == "" {
		return "", "", fmt.Errorf("unsupported call target")
	}
	args := make([]string, 0, len(expression.Arguments))
	for _, argument := range expression.Arguments {
		value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		args = append(args, value)
	}
	target, ok := signatures[callee]
	if !ok {
		return "", "", fmt.Errorf("unknown function %q", callee)
	}
	callee = target.Name
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
	if len(target.Parameters) > 0 && (target.Parameters[len(target.Parameters)-1].Type == ir.TypeStringArray || target.Parameters[len(target.Parameters)-1].Type == ir.TypeNumberArray) {
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
	case "indexOf", "lastIndexOf", "slice", "startsWith", "endsWith", "trim", "trimStart", "trimEnd", "replace", "replaceAll", "substring", "split", "charAt", "charCodeAt", "includes", "toLowerCase", "toUpperCase", "repeat", "padStart", "padEnd", "concat":
		return true
	default:
		return false
	}
}

func isArrayMethod(name string) bool {
	switch name {
	case "push", "pop", "slice", "indexOf", "includes", "join", "reverse", "concat", "shift", "unshift", "splice", "at":
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
