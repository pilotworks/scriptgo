package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerProxyNew(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if len(expression.Arguments) != 2 {
		return "", "", fmt.Errorf("Proxy constructor requires target and handler arguments")
	}
	target, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	handler, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if result == "" {
		result = nextTemp(counter)
	}
	retType := ir.TypeUnknown
	if expression.InferredType != "" {
		retType = toIRType(expression.InferredType)
		if retType == "" {
			retType = ir.TypeUnknown
		}
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   retType,
		Result: result,
		Callee: "__dynamic.new_proxy",
		Args:   []string{target, handler},
		Span:   toIRSpan(path, expression.Span),
	})
	return result, retType, nil
}
