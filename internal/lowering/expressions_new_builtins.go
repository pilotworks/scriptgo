package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerNewPromise(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if result == "" {
		result = nextTemp(counter)
	}
	promType := ir.Type("object:Promise<unknown>")
	if inferred := strings.TrimPrefix(expression.InferredType, "object:"); strings.HasPrefix(inferred, "Promise<") && strings.HasSuffix(inferred, ">") {
		inner := strings.TrimSuffix(strings.TrimPrefix(inferred, "Promise<"), ">")
		if resolved := toIRType(inner); resolved != "" {
			promType = ir.Type("object:Promise<" + string(resolved) + ">")
		}
	}
	if len(expression.Arguments) != 1 {
		return "", "", fmt.Errorf("promise constructor requires exactly one executor")
	}
	executor, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   promType,
		Result: result,
		Callee: "__async.promise_construct",
		Args:   []string{executor},
		Span:   toIRSpan(path, expression.Span),
	})
	return result, promType, nil
}

func lowerNewWeakRef(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	var targetArg string
	targetType := ir.TypeObject
	if len(expression.Arguments) > 0 {
		v, t, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		targetArg = v
		if t != "" {
			targetType = t
		}
	}
	if result == "" {
		result = nextTemp(counter)
	}
	resType := ir.Type("object:WeakRef<" + string(targetType) + ">")
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   resType,
		Result: result,
		Callee: "__weakref.new",
		Args:   []string{targetArg},
		Span:   toIRSpan(path, expression.Span),
	})
	return result, resType, nil
}

func lowerNewWeakMap(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, counter *int) (string, ir.Type, error) {
	if result == "" {
		result = nextTemp(counter)
	}
	// Preserve generic arguments on the handle so the backend can retain
	// the value tag across the nullable WeakMap.get ABI.
	resType := ir.Type("object:WeakMap")
	inferred := strings.TrimPrefix(expression.InferredType, "object:")
	if strings.HasPrefix(inferred, "WeakMap<") && strings.HasSuffix(inferred, ">") {
		resType = ir.Type("object:" + inferred)
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   resType,
		Result: result,
		Callee: "__weakmap.new",
		Span:   toIRSpan(path, expression.Span),
	})
	return result, resType, nil
}

func lowerNewArray(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if len(expression.Arguments) == 1 {
		lenVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		retType := ir.TypeNumberArray
		if expression.InferredType != "" {
			inferred := toIRType(expression.InferredType)
			if strings.HasSuffix(string(inferred), "[]") {
				retType = inferred
			}
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   retType,
			Result: result,
			Callee: "__array.new_length",
			Args:   []string{lenVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, nil
	}
	var args []string
	elemType := ir.TypeNumber
	for _, argExpr := range expression.Arguments {
		argVal, aType, err := lowerExpression(path, argExpr, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		args = append(args, argVal)
		if aType != "" {
			elemType = aType
		}
	}
	retType := ir.Type(string(elemType) + "[]")
	if elemType == ir.TypeNumber {
		retType = ir.TypeNumberArray
	} else if elemType == ir.TypeString {
		retType = ir.TypeStringArray
	} else if elemType == ir.TypeBool {
		retType = ir.TypeBoolArray
	} else if elemType == ir.TypeBigInt {
		retType = ir.TypeBigIntArray
	}
	if result == "" {
		result = nextTemp(counter)
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpArray,
		Type:   retType,
		Result: result,
		Args:   args,
		Span:   toIRSpan(path, expression.Span),
	})
	return result, retType, nil
}

func lowerNewWeakRefFallback(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	targetVal := "null"
	if len(expression.Arguments) > 0 {
		v, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		targetVal = v
	}
	if result == "" {
		result = nextTemp(counter)
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeObject,
		Result: result,
		Callee: "__weakref.new",
		Args:   []string{targetVal},
		Span:   toIRSpan(path, expression.Span),
	})
	return result, ir.TypeObject, nil
}

func lowerNewFinalizationRegistry(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	cbVal := "null"
	if len(expression.Arguments) > 0 {
		v, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		cbVal = v
	}
	if result == "" {
		result = nextTemp(counter)
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.Type("object:FinalizationRegistry"),
		Result: result,
		Callee: "__finalization_registry.new",
		Args:   []string{cbVal},
		Span:   toIRSpan(path, expression.Span),
	})
	return result, ir.Type("object:FinalizationRegistry"), nil
}

func lowerNewMap(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if result == "" {
		result = nextTemp(counter)
	}
	if len(expression.Arguments) == 0 {
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeMap,
			Result: result,
			Callee: "__map.new",
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeMap, nil
	}
	arg0Val, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeMap,
		Result: result,
		Callee: "__map.new_entries",
		Args:   []string{arg0Val},
		Span:   toIRSpan(path, expression.Span),
	})
	return result, ir.TypeMap, nil
}

func lowerNewSet(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if result == "" {
		result = nextTemp(counter)
	}
	if len(expression.Arguments) == 0 {
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeSet,
			Result: result,
			Callee: "__set.new",
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeSet, nil
	}
	arg0Val, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeSet,
		Result: result,
		Callee: "__set.new_values",
		Args:   []string{arg0Val},
		Span:   toIRSpan(path, expression.Span),
	})
	return result, ir.TypeSet, nil
}

func lowerNewDate(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, objType ir.Type) (string, ir.Type, error) {
	timeVal := nextTemp(counter)
	if len(expression.Arguments) == 0 {
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpCall, Type: ir.TypeNumber, Result: timeVal, Callee: "__date.now", Span: toIRSpan(path, expression.Span),
		})
	} else {
		argVal, argType, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if argType == ir.TypeNumber {
			timeVal = argVal
		} else if argType == ir.TypeString {
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpCall, Type: ir.TypeNumber, Result: timeVal, Callee: "__date.parse", Args: []string{argVal}, Span: toIRSpan(path, expression.Span),
			})
		} else {
			timeVal = argVal
		}
	}
	function.Body = append(function.Body, ir.Instruction{
		Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{result, timeVal}, Span: toIRSpan(path, expression.Span),
	})
	return result, objType, nil
}
