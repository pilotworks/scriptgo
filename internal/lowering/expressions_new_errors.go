package lowering

import (
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
	nameVal := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{
		Op: ir.OpConst, Type: ir.TypeString, Result: nameVal, Value: publicName, Span: toIRSpan(path, expression.Span),
	})
	function.Body = append(function.Body, ir.Instruction{
		Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "message", FieldIndex: 0, Args: []string{result, msgVal}, Span: toIRSpan(path, expression.Span),
	})
	function.Body = append(function.Body, ir.Instruction{
		Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "name", FieldIndex: 1, Args: []string{result, nameVal}, Span: toIRSpan(path, expression.Span),
	})
	stackVal := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{
		Op: ir.OpCall, Type: ir.TypeString, Callee: "__error.captureStack", Result: stackVal, Args: []string{nameVal, msgVal}, Span: toIRSpan(path, expression.Span),
	})
	causeVal := nextTemp(counter)
	causeFound := false
	if len(expression.Arguments) > 1 && expression.Arguments[1].Kind == "object_literal" {
		for _, prop := range expression.Arguments[1].Arguments {
			if prop.Text == "cause" && prop.Left != nil {
				cv, _, err := lowerExpression(path, prop.Left, "", function, env, counter, shapes, signatures)
				if err == nil {
					causeVal = cv
					causeFound = true
					break
				}
			}
		}
	}
	if !causeFound {
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeString, Result: causeVal, Value: "", Span: toIRSpan(path, expression.Span),
		})
	}
	function.Body = append(function.Body, ir.Instruction{
		Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "stack", FieldIndex: 2, Args: []string{result, stackVal}, Span: toIRSpan(path, expression.Span),
	})
	function.Body = append(function.Body, ir.Instruction{
		Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "cause", FieldIndex: 3, Args: []string{result, causeVal}, Span: toIRSpan(path, expression.Span),
	})
	return result, objType, nil
}
