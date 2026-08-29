package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerWeakReceiverMethod(
	path string,
	expression *typescriptgo.SyntaxExpression,
	receiver string,
	methodName string,
	receiverType ir.Type,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, bool, error) {
	if (receiverType == "object:WeakRef" || strings.HasPrefix(string(receiverType), "object:WeakRef<") || strings.HasPrefix(string(receiverType), "object:WeakRef__") || receiverType == "WeakRef") && methodName == "deref" {
		var retType ir.Type = ir.TypeObject
		if strings.HasPrefix(string(receiverType), "object:WeakRef<") && strings.HasSuffix(string(receiverType), ">") {
			inner := strings.TrimSuffix(strings.TrimPrefix(string(receiverType), "object:WeakRef<"), ">")
			retType = toIRType(inner)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   retType,
			Result: result,
			Callee: "__weakref.deref",
			Args:   []string{receiver},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, true, nil
	}
	if receiverType == "object:WeakMap" || strings.HasPrefix(string(receiverType), "object:WeakMap<") || strings.HasPrefix(string(receiverType), "object:WeakMap__") || receiverType == "WeakMap" {
		switch methodName {
		case "set":
			if len(expression.Arguments) >= 2 {
				kVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
				}
				vVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeVoid,
					Callee: "__weakmap.set",
					Args:   []string{receiver, kVal, vVal},
					Span:   toIRSpan(path, expression.Span),
				})
				return receiver, receiverType, true, nil
			}
		case "get":
			if len(expression.Arguments) >= 1 {
				kVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
				}
				if result == "" {
					result = nextTemp(counter)
				}
				retType := ir.TypeObject
				if strings.HasPrefix(string(receiverType), "object:WeakMap<") && strings.HasSuffix(string(receiverType), ">") {
					inner := strings.TrimSuffix(strings.TrimPrefix(string(receiverType), "object:WeakMap<"), ">")
					parts := strings.Split(inner, ",")
					if len(parts) >= 2 {
						retType = toIRType(strings.TrimSpace(parts[1]))
					}
				} else if expression.InferredType != "" {
					retType = toIRType(expression.InferredType)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   retType,
					Result: result,
					Callee: "__weakmap.get",
					Args:   []string{receiver, kVal},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, retType, true, nil
			}
		case "has":
			if len(expression.Arguments) >= 1 {
				kVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
				}
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeBool,
					Result: result,
					Callee: "__weakmap.has",
					Args:   []string{receiver, kVal},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, ir.TypeBool, true, nil
			}
		case "delete":
			if len(expression.Arguments) >= 1 {
				kVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
				}
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeBool,
					Result: result,
					Callee: "__weakmap.delete",
					Args:   []string{receiver, kVal},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, ir.TypeBool, true, nil
			}
		}
	}
	if receiverType == "object:WeakSet" || strings.HasPrefix(string(receiverType), "object:WeakSet<") || strings.HasPrefix(string(receiverType), "object:WeakSet__") || receiverType == "WeakSet" {
		switch methodName {
		case "add":
			if len(expression.Arguments) >= 1 {
				vVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeVoid,
					Callee: "__weakset.add",
					Args:   []string{receiver, vVal},
					Span:   toIRSpan(path, expression.Span),
				})
				return receiver, receiverType, true, nil
			}
		case "has":
			if len(expression.Arguments) >= 1 {
				vVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
				}
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeBool,
					Result: result,
					Callee: "__weakset.has",
					Args:   []string{receiver, vVal},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, ir.TypeBool, true, nil
			}
		case "delete":
			if len(expression.Arguments) >= 1 {
				vVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
				}
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeBool,
					Result: result,
					Callee: "__weakset.delete",
					Args:   []string{receiver, vVal},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, ir.TypeBool, true, nil
			}
		}
	}
	if receiverType == "object:FinalizationRegistry" || strings.HasPrefix(string(receiverType), "object:FinalizationRegistry<") || strings.HasPrefix(string(receiverType), "object:FinalizationRegistry__") || receiverType == "FinalizationRegistry" {
		switch methodName {
		case "register":
			if len(expression.Arguments) >= 2 {
				targetVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
				}
				heldVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
				}
				tokenVal := "null"
				if len(expression.Arguments) >= 3 {
					tok, _, err := lowerExpression(path, expression.Arguments[2], "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", true, err
					}
					tokenVal = tok
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeVoid,
					Callee: "__finalization_registry.register",
					Args:   []string{receiver, targetVal, heldVal, tokenVal},
					Span:   toIRSpan(path, expression.Span),
				})
				return "", ir.TypeVoid, true, nil
			}
		case "unregister":
			if len(expression.Arguments) >= 1 {
				tokenVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", true, err
				}
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeBool,
					Result: result,
					Callee: "__finalization_registry.unregister",
					Args:   []string{receiver, tokenVal},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, ir.TypeBool, true, nil
			}
		}
	}
	return "", "", false, nil
}

func lowerMapSetReceiverMethod(
	path string,
	expression *typescriptgo.SyntaxExpression,
	receiver string,
	methodName string,
	receiverType ir.Type,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, bool, error) {
	if isMapType(receiverType) {
		switch methodName {
		case "set":
			if len(expression.Arguments) < 2 {
				return "", "", true, fmt.Errorf("Map.set requires key and value arguments")
			}
			_, valTypeStr := resolveMapTypes(expression.Left, env)
			if valTypeStr != "" && expression.Arguments[1].Kind == "array" && (expression.Arguments[1].InferredType == "" || expression.Arguments[1].InferredType == "never[]" || expression.Arguments[1].InferredType == "unknown[]") {
				expression.Arguments[1].InferredType = valTypeStr
			}
			kVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			if _, mapValType := resolveMapTypes(expression.Left, env); mapValType != "" {
				if expression.Arguments[1].InferredType == "" || expression.Arguments[1].InferredType == "unknown[]" || expression.Arguments[1].InferredType == "any[]" {
					expression.Arguments[1].InferredType = mapValType
				}
			}
			vVal, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeMap,
				Result: result,
				Callee: "__map.set",
				Args:   []string{receiver, kVal, vVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeMap, true, nil

		case "get":
			if len(expression.Arguments) < 1 {
				return "", "", true, fmt.Errorf("Map.get requires key argument")
			}
			kVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			if result == "" {
				result = nextTemp(counter)
			}
			retType := ir.TypeUnknown
			if _, valType := resolveMapTypes(expression.Left, env); valType != "" {
				retType = toIRType(valType)
			} else if expression.InferredType != "" {
				retType = toIRType(expression.InferredType)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   retType,
				Result: result,
				Callee: "__map.get",
				Args:   []string{receiver, kVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, retType, true, nil

		case "has":
			if len(expression.Arguments) < 1 {
				return "", "", true, fmt.Errorf("Map.has requires key argument")
			}
			kVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBool,
				Result: result,
				Callee: "__map.has",
				Args:   []string{receiver, kVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, true, nil

		case "delete":
			if len(expression.Arguments) < 1 {
				return "", "", true, fmt.Errorf("Map.delete requires key argument")
			}
			kVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBool,
				Result: result,
				Callee: "__map.delete",
				Args:   []string{receiver, kVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, true, nil

		case "clear":
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeVoid,
				Callee: "__map.clear",
				Args:   []string{receiver},
				Span:   toIRSpan(path, expression.Span),
			})
			return "", ir.TypeVoid, true, nil

		case "keys", "values", "entries":
			if result == "" {
				result = nextTemp(counter)
			}
			retType := ir.TypeStringArray
			if expression.InferredType != "" {
				retType = toIRType(expression.InferredType)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   retType,
				Result: result,
				Callee: "__map." + methodName,
				Args:   []string{receiver},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, retType, true, nil

		case "forEach":
			if len(expression.Arguments) < 1 {
				return "", "", true, fmt.Errorf("Map.forEach requires callback argument")
			}
			if cb := expression.Arguments[0]; cb != nil && (cb.Kind == "arrow_function" || cb.Kind == "function") && cb.Function != nil {
				keyType, valType := resolveMapTypes(expression.Left, env)
				if valType != "" && len(cb.Function.Parameters) > 0 && cb.Function.Parameters[0].Type == "" && cb.Function.Parameters[0].InferredType == "" {
					cb.Function.Parameters[0].Type = valType
				}
				if keyType != "" && len(cb.Function.Parameters) > 1 && cb.Function.Parameters[1].Type == "" && cb.Function.Parameters[1].InferredType == "" {
					cb.Function.Parameters[1].Type = keyType
				}
			}
			cbVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeVoid,
				Callee: "__map.forEach",
				Args:   []string{receiver, cbVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return "", ir.TypeVoid, true, nil
		}
	}
	if isSetType(receiverType) {
		switch methodName {
		case "add":
			if len(expression.Arguments) < 1 {
				return "", "", true, fmt.Errorf("Set.add requires value argument")
			}
			vVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeSet,
				Result: result,
				Callee: "__set.add",
				Args:   []string{receiver, vVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeSet, true, nil

		case "has":
			if len(expression.Arguments) < 1 {
				return "", "", true, fmt.Errorf("Set.has requires value argument")
			}
			vVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBool,
				Result: result,
				Callee: "__set.has",
				Args:   []string{receiver, vVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, true, nil

		case "delete":
			if len(expression.Arguments) < 1 {
				return "", "", true, fmt.Errorf("Set.delete requires value argument")
			}
			vVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBool,
				Result: result,
				Callee: "__set.delete",
				Args:   []string{receiver, vVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, true, nil

		case "clear":
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeVoid,
				Callee: "__set.clear",
				Args:   []string{receiver},
				Span:   toIRSpan(path, expression.Span),
			})
			return "", ir.TypeVoid, true, nil

		case "keys", "values", "entries":
			if result == "" {
				result = nextTemp(counter)
			}
			retType := ir.TypeNumberArray
			if expression.InferredType != "" {
				retType = toIRType(expression.InferredType)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   retType,
				Result: result,
				Callee: "__set." + methodName,
				Args:   []string{receiver},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, retType, true, nil

		case "union", "intersection", "difference", "symmetricDifference":
			if len(expression.Arguments) < 1 {
				return "", "", true, fmt.Errorf("Set.%s requires other Set argument", methodName)
			}
			otherVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeSet,
				Result: result,
				Callee: "__set." + methodName,
				Args:   []string{receiver, otherVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeSet, true, nil

		case "isSubsetOf", "isSupersetOf", "isDisjointFrom":
			if len(expression.Arguments) < 1 {
				return "", "", true, fmt.Errorf("Set.%s requires other Set argument", methodName)
			}
			otherVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBool,
				Result: result,
				Callee: "__set." + methodName,
				Args:   []string{receiver, otherVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, true, nil

		case "forEach":
			if len(expression.Arguments) < 1 {
				return "", "", true, fmt.Errorf("Set.forEach requires callback argument")
			}
			cbVal, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", true, err
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeVoid,
				Callee: "__set.forEach",
				Args:   []string{receiver, cbVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return "", ir.TypeVoid, true, nil
		}
	}
	return "", "", false, nil
}
