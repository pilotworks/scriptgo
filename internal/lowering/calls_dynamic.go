package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerDynamicImportCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, dynamic dynamicImportBinding) (string, ir.Type, error) {
	if result == "" {
		result = nextTemp(counter)
	}
	returnType := toIRType(expression.InferredType)
	if strings.Contains(expression.InferredType, "=>") {
		returnType = ir.TypeDynamicFunction
	}
	if returnType == "" || returnType == ir.TypeVoid {
		returnType = ir.TypeUnknown
	}
	args := make([]string, 0, len(expression.Arguments))
	for _, argument := range expression.Arguments {
		value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		args = append(args, value)
	}
	function.Body = append(function.Body, ir.Instruction{
		Op: ir.OpDynamicCall, Type: returnType, Result: result,
		Callee: dynamic.Path + "#" + dynamic.Export, Args: args,
		Field: dynamic.Export, FieldIndex: dynamic.Arity, Span: toIRSpan(path, expression.Span),
	})
	if strings.HasPrefix(string(returnType), "object:") || returnType == ir.TypeObject {
		env[result+".dynamic"] = ir.Type("true")
	}
	return result, returnType, nil

}

func tryLowerDynamicPropertyCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, closureVal string, closureType ir.Type) (string, ir.Type, error) {
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
	if rt, ok := env[closureVal+".retType"]; ok && rt != "" {
		retType = rt
	} else if target, ok := signatures[closureVal]; ok && target.ReturnType != "" {
		retType = target.ReturnType
	} else if expression.InferredType != "" {
		retType = toIRType(expression.InferredType)
	} else if expression.Left.InferredType != "" && strings.Contains(expression.Left.InferredType, "=>") {
		retStr := extractTopLevelReturnType(expression.Left.InferredType)
		if parsed := toIRType(retStr); parsed != "" {
			retType = parsed
		}
	}
	op := ir.OpClosureCall
	if closureType == ir.TypeDynamicFunction || (currentDynamicMode && closureType == ir.TypeUnknown) {
		var receiver string
		if expression.Left != nil && expression.Left.Left != nil {
			rec, _, receiverErr := lowerExpression(path, expression.Left.Left, "", function, env, counter, shapes, signatures)
			if receiverErr == nil {
				receiver = rec
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpDynamicFunctionCall, Type: retType, Result: result,
			Callee: closureVal, This: receiver, Args: append([]string{closureVal}, args...),
			Span: toIRSpan(path, expression.Span),
		})
		return result, retType, nil
	}
	function.Body = append(function.Body, ir.Instruction{
		Op: op, Type: retType, Result: result, Callee: closureVal, Args: args,
		Span: toIRSpan(path, expression.Span),
	})
	return result, retType, nil

}

func lowerDynamicFunctionCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, callee string) (string, ir.Type, error) {
	args := make([]string, 0, len(expression.Arguments)+1)
	args = append(args, callee)
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
	retType := toIRType(expression.InferredType)
	if retType == "" {
		retType = ir.TypeUnknown
	}
	function.Body = append(function.Body, ir.Instruction{
		Op: ir.OpDynamicFunctionCall, Type: retType, Result: result,
		Callee: callee, Args: args, Span: toIRSpan(path, expression.Span),
	})
	return result, retType, nil

}
