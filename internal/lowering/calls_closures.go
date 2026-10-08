package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func tryLowerPropertyClosureCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, receiver string, receiverType ir.Type, propVal string, propType ir.Type) (string, ir.Type, error) {
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
	retType := ir.TypeNumber
	if expression.InferredType != "" {
		retType = toIRType(expression.InferredType)
	}
	op := ir.OpClosureCall
	if propType == ir.TypeDynamicFunction || (currentDynamicMode && (receiverType == ir.TypeUnknown || propType == ir.TypeUnknown)) {
		op = ir.OpDynamicFunctionCall
	}
	instruction := ir.Instruction{
		Op: op, Type: retType, Result: result, Callee: propVal,
		Args: append([]string{propVal}, args...), This: receiver,
		Span: toIRSpan(path, expression.Span),
	}
	if op == ir.OpClosureCall {
		instruction.Args = args
		instruction.This = ""
	}
	function.Body = append(function.Body, instruction)
	return result, retType, nil

}

func lowerImmediateArrowCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	closureVal, _, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
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
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpClosureCall,
		Type:   retType,
		Result: result,
		Callee: closureVal,
		Args:   args,
		Span:   toIRSpan(path, expression.Span),
	})
	return result, retType, nil

}

func lowerClosureValueCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, callee string) (string, ir.Type, error) {
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
	} else if target, ok := resolveFunctionSignature(path, callee, signatures); ok && target.ReturnType != "" {
		retType = target.ReturnType
	} else if expression.InferredType != "" {
		retType = toIRType(expression.InferredType)
	} else if topVar, ok := topLevelVars[callee]; ok {
		if isReturningClosure(topVar) {
			retType = ir.TypeClosure
		} else if topVar.Type != "" {
			t := toIRType(topVar.Type)
			if t == ir.TypeClosure || strings.HasPrefix(string(t), "object:") {
				retType = t
			}
		}
	}
	calleeName := callee
	if targetName, ok := env[callee+".closureTarget"]; ok && targetName != "" {
		calleeName = string(targetName)
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpClosureCall,
		Type:   retType,
		Result: result,
		Callee: calleeName,
		Args:   args,
		Span:   toIRSpan(path, expression.Span),
	})
	return result, retType, nil

}

func tryLowerClosurePropertyCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, closureVal string, closureType ir.Type) (string, ir.Type, error) {
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
	if expression.InferredType != "" {
		retType = toIRType(expression.InferredType)
	} else if strings.Contains(string(closureType), "=>") {
		retStr := extractTopLevelReturnType(string(closureType))
		retType = toIRType(retStr)
	} else if rt, ok := env[closureVal+".retType"]; ok && rt != "" {
		retType = rt
	} else if strings.HasPrefix(string(closureType), "object:Generator_") {
		retType = closureType
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpClosureCall,
		Type:   retType,
		Result: result,
		Callee: closureVal,
		Args:   args,
		Span:   toIRSpan(path, expression.Span),
	})
	return result, retType, nil

}
