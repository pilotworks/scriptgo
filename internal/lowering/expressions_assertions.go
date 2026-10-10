package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerAsExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	targetIRType := toIRType(expression.Text)
	val, valType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	// A function handle returned by the Dynamic engine is callable, but it
	// does not have the native closure layout. Preserve that ABI through a
	// TypeScript function assertion so later calls use the Dynamic bridge.
	if valType == ir.TypeDynamicFunction && targetIRType == ir.TypeClosure {
		targetIRType = ir.TypeDynamicFunction
	}
	if result == "" {
		result = nextTemp(counter)
	}
	if env[val+".dynamic"] != "" {
		env[result+".dynamic"] = env[val+".dynamic"]
	}
	if targetIRType == ir.TypeUnknown {
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpBoxUnknown,
			Type:   ir.TypeUnknown,
			Result: result,
			Args:   []string{val},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeUnknown, nil
	}
	srcVal := val
	if valType != ir.TypeUnknown && valType != ir.TypeDynamicFunction && !strings.HasPrefix(string(valType), "object:") && !strings.Contains(string(valType), "|") {
		boxed := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpBoxUnknown,
			Type:   ir.TypeUnknown,
			Result: boxed,
			Args:   []string{val},
			Span:   toIRSpan(path, expression.Span),
		})
		srcVal = boxed
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCheckedCast,
		Type:   targetIRType,
		Result: result,
		Args:   []string{srcVal},
		Span:   toIRSpan(path, expression.Span),
	})
	return result, targetIRType, nil
}

func lowerNonNullExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if expression.Left == nil {
		return "", "", fmt.Errorf("non-null expression is missing its operand")
	}
	targetIRType := nonNullishIRType(expression.InferredType)
	if targetIRType == ir.TypeUnknown || targetIRType == "" {
		targetIRType = nonNullishIRType(expression.Left.InferredType)
	}
	inner := expression.Left
	// Feed the narrowed type into call/property lowering so the intrinsic
	// selects a typed result ABI instead of first materializing a nullable box.
	if targetIRType != ir.TypeUnknown && targetIRType != ir.TypeVoid {
		copyExpr := *inner
		copyExpr.InferredType = string(targetIRType)
		inner = &copyExpr
	}
	val, valType, err := lowerExpression(path, inner, result, function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if targetIRType == ir.TypeUnknown || targetIRType == ir.TypeVoid || valType == targetIRType {
		return val, valType, nil
	}
	if result == "" {
		result = nextTemp(counter)
	}
	srcVal := val
	if valType != ir.TypeUnknown && !strings.HasPrefix(string(valType), "object:") && !strings.Contains(string(valType), "|") {
		boxed := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpBoxUnknown,
			Type:   ir.TypeUnknown,
			Result: boxed,
			Args:   []string{val},
			Span:   toIRSpan(path, expression.Span),
		})
		srcVal = boxed
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCheckedCast,
		Type:   targetIRType,
		Result: result,
		Args:   []string{srcVal},
		Span:   toIRSpan(path, expression.Span),
	})
	return result, targetIRType, nil
}

func lowerTypeofExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	val, valType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		callee := callName(expression.Left)
		if callee != "" && isClassIdentifier(path, callee) {
			typeStr := "function"
			if callee == "crypto" || callee == "performance" || callee == "Math" || callee == "JSON" || callee == "Intl" || callee == "Atomics" {
				typeStr = "object"
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeString,
				Result: result,
				Value:  typeStr,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeString, nil
		}
		if expression.Left != nil && (expression.Left.Kind == "property" || expression.Left.Kind == "optional_property") && expression.Left.Left != nil {
			_, targetObjType, targetErr := lowerExpression(path, expression.Left.Left, "", function, env, counter, shapes, signatures)
			if targetErr == nil {
				targetCls := strings.TrimPrefix(string(targetObjType), "object:")
				targetCls = classIdentityForPath(path, targetCls)
				if _, _, ok := findMethodInHierarchy(targetCls, expression.Left.Text, signatures, classHierarchy); ok {
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   ir.TypeString,
						Result: result,
						Value:  "function",
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeString, nil
				}
			}
		}
		if expression.Left != nil && expression.Left.Kind == "identifier" {
			name := expression.Left.Text
			if name == "undefined" {
				valType = ir.TypeVoid
			} else if isGlobalConstructor(name) || isClassIdentifier(path, name) {
				typeStr := "function"
				if name == "crypto" || name == "performance" || name == "Math" || name == "JSON" || name == "Intl" || name == "Atomics" {
					typeStr = "object"
				}
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeString,
					Result: result,
					Value:  typeStr,
					Span:   toIRSpan(path, expression.Span),
				})
				return result, ir.TypeString, nil
			} else {
				valType = ir.TypeVoid
			}
		} else {
			return "", "", err
		}
	}
	if result == "" {
		result = nextTemp(counter)
	}
	runtimeTypeOf := expression.Left != nil && (typeContainsNullish(expression.Left.InferredType) || strings.Contains(string(valType), "|"))
	if runtimeTypeOf && (isPointerLikeType(valType) || valType == ir.TypeNumber) {
		function.Body = append(function.Body, ir.Instruction{
			Op:            ir.OpTypeOf,
			Type:          ir.TypeString,
			Result:        result,
			Args:          []string{val},
			RuntimeTypeOf: true,
			Span:          toIRSpan(path, expression.Span),
		})
		return result, ir.TypeString, nil
	}
	if valType == ir.TypeUnknown || valType == ir.TypeClosure || strings.Contains(string(valType), "|") {
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpTypeOf,
			Type:   ir.TypeString,
			Result: result,
			Args:   []string{val},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeString, nil
	}
	var typeStr string
	if expression.Left != nil && expression.Left.Kind == "null" {
		typeStr = "object"
	} else if expression.Left != nil && expression.Left.Kind == "undefined" {
		typeStr = "undefined"
	} else {
		switch valType {
		case ir.TypeNumber:
			typeStr = "number"
		case ir.TypeBigInt:
			typeStr = "bigint"
		case ir.TypeSymbol:
			typeStr = "symbol"
		case ir.TypeString:
			typeStr = "string"
		case ir.TypeBool:
			typeStr = "boolean"
		case ir.TypeVoid:
			typeStr = "undefined"
		case ir.TypeClosure:
			typeStr = "function"
		default:
			typeStr = "object"
		}
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: result,
		Value:  typeStr,
		Span:   toIRSpan(path, expression.Span),
	})
	return result, ir.TypeString, nil
}
