package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerNewError(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, className string, objType ir.Type, publicName string) (string, ir.Type, error) {
	msgVal := nextTemp(counter)
	if len(expression.Arguments) > 0 {
		mv, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		msgVal = mv
	} else {
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeString, Result: msgVal, Value: "", Span: toIRSpan(path, expression.Span),
		})
	}
	causeVal, err := lowerErrorCause(path, expression, function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	initializeErrorFields(path, expression.Span, function, counter, className, result, msgVal, causeVal, publicName)
	return result, objType, nil
}

// lowerErrorCause returns the options.cause argument of an Error construction
// (new Error(message, { cause })), or an empty string.
func lowerErrorCause(path string, expression *frontend.SyntaxExpression, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, error) {
	if len(expression.Arguments) > 1 && expression.Arguments[1].Kind == "object_literal" {
		for _, prop := range expression.Arguments[1].Arguments {
			if prop.Text == "cause" && prop.Left != nil {
				if cv, _, err := lowerExpression(path, prop.Left, "", function, env, counter, shapes, signatures); err == nil {
					return cv, nil
				}
			}
		}
	}
	causeVal := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{
		Op: ir.OpConst, Type: ir.TypeString, Result: causeVal, Value: "", Span: toIRSpan(path, expression.Span),
	})
	return causeVal, nil
}

// initializeErrorFields stores message, name, stack, and cause on a new
// error instance (built-in or a user subclass), in the built-in Error layout.
func initializeErrorFields(path string, span frontend.SourceSpan, function *ir.Function, counter *int, className, target, msgVal, causeVal, publicName string) {
	nameVal := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{
		Op: ir.OpConst, Type: ir.TypeString, Result: nameVal, Value: publicName, Span: toIRSpan(path, span),
	})
	stackVal := nextTemp(counter)
	function.Body = append(function.Body,
		ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "message", FieldIndex: 0, Args: []string{target, msgVal}, Span: toIRSpan(path, span)},
		ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "name", FieldIndex: 1, Args: []string{target, nameVal}, Span: toIRSpan(path, span)},
		ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Callee: "__error.captureStack", Result: stackVal, Args: []string{nameVal, msgVal}, Span: toIRSpan(path, span)},
		ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "stack", FieldIndex: 2, Args: []string{target, stackVal}, Span: toIRSpan(path, span)},
		ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "cause", FieldIndex: 3, Args: []string{target, causeVal}, Span: toIRSpan(path, span)},
	)
}

// lowerNewAggregateError lowers new AggregateError(errors, message?, options?)
// into the built-in Error layout followed by the errors array.
func lowerNewAggregateError(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, className string, objType ir.Type) (string, ir.Type, error) {
	if len(expression.Arguments) == 0 {
		return "", "", fmt.Errorf("AggregateError requires an errors array")
	}
	errorsVal, errorsType, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if !strings.HasSuffix(string(errorsType), "[]") {
		return "", "", fmt.Errorf("AggregateError supports an errors array in the native subset, got %s", expression.Arguments[0].InferredType)
	}
	rest := *expression
	rest.Arguments = expression.Arguments[1:]
	value, typ, err := lowerNewError(path, &rest, result, function, env, counter, shapes, signatures, className, objType, "AggregateError")
	if err != nil {
		return "", "", err
	}
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "errors", FieldIndex: 4, Args: []string{value, errorsVal}, Span: toIRSpan(path, expression.Span)})
	return value, typ, nil
}
