package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerSuperConstructorCall(path string, expression *frontend.SyntaxExpression, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	thisType, ok := env["this"]
	if !ok {
		return "", "", fmt.Errorf("super() can only be used inside a class constructor")
	}
	currentClass := strings.TrimPrefix(string(thisType), "object:")
	meta := classHierarchy[currentClass]
	if meta.Extends == "" {
		return "", "", fmt.Errorf("super() called in class %q with no base class", currentClass)
	}
	if isBuiltinErrorClass(meta.Extends) {
		// super(message, options) on a native error base initializes the
		// inherited Error fields; the instance name is the base's name.
		msgVal := nextTemp(counter)
		if len(expression.Arguments) > 0 {
			value, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			msgVal = value
		} else {
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeString, Result: msgVal, Value: "", Span: toIRSpan(path, expression.Span),
			})
		}
		causeVal, err := lowerErrorCause(path, expression, function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		initializeErrorFields(path, expression.Span, function, counter, currentClass, "this", msgVal, causeVal, meta.Extends)
		return "", ir.TypeVoid, nil
	}
	if meta.Extends == "DOMException" {
		if len(expression.Arguments) > 0 {
			msgVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: currentClass, Field: "message", FieldIndex: 0, Args: []string{"this", msgVal}, Span: toIRSpan(path, expression.Span),
			})
		}
		return "", ir.TypeVoid, nil
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
	for i := len(args); i < len(ctor.Parameters); i++ {
		paramType := ctor.Parameters[i].Type
		defConst := nextTemp(counter)
		defVal := "0"
		if paramType == ir.TypeBool {
			defVal = "false"
		} else if paramType == ir.TypeString {
			defVal = ""
		} else if isPointerLikeType(paramType) || strings.HasPrefix(string(paramType), "object:") {
			defVal = "null"
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   paramType,
			Result: defConst,
			Value:  defVal,
			Span:   toIRSpan(path, expression.Span),
		})
		args = append(args, defConst)
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

func lowerSuperMethodCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	thisType, ok := env["this"]
	if !ok {
		return "", "", fmt.Errorf("super.method() can only be used inside a class method")
	}
	currentClass := strings.TrimPrefix(string(thisType), "object:")
	meta := classHierarchy[currentClass]
	if meta.Extends == "" {
		return "", "", fmt.Errorf("super.%s called in class %q with no base class", expression.Left.Text, currentClass)
	}
	target, mangled, found := findImplementationInHierarchy(meta.Extends, expression.Left.Text, signatures, classHierarchy)
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
