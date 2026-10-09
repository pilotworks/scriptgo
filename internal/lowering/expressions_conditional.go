package lowering

import (
	"fmt"
	"maps"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerConditionalExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	condition, conditionType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if conditionType != ir.TypeBool {
		condition, err = coerceToBool(path, condition, conditionType, function, counter, expression.Span)
		if err != nil {
			return "", "", fmt.Errorf("conditional condition: %w", err)
		}
	}
	resSlot := result
	if resSlot == "" {
		resSlot = nextTemp(counter)
	}
	thenFn := ir.Function{Name: function.Name}
	thenEnv := make(map[string]ir.Type, len(env))
	maps.Copy(thenEnv, env)
	whenTrue, trueType, err := lowerExpression(path, expression.WhenTrue, "", &thenFn, thenEnv, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	elseFn := ir.Function{Name: function.Name}
	elseEnv := make(map[string]ir.Type, len(env))
	maps.Copy(elseEnv, env)
	whenFalse, falseType, err := lowerExpression(path, expression.WhenFalse, "", &elseFn, elseEnv, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if trueType != falseType {
		if expression.WhenFalse != nil && (expression.WhenFalse.Kind == "null" || expression.WhenFalse.Kind == "undefined") {
			whenFalse = nextTemp(counter)
			zeroVal := "0"
			if expression.WhenFalse.Kind == "undefined" {
				switch trueType {
				case ir.TypeNumber:
					zeroVal = "undefined"
				case ir.TypeBool:
					zeroVal = "false"
				default:
					zeroVal = "undefined"
				}
			} else if trueType == ir.TypeString || trueType == ir.TypeNumber || strings.HasPrefix(string(trueType), "object:") || trueType == ir.TypePointer {
				zeroVal = "null"
			}
			elseFn.Body = append(elseFn.Body, ir.Instruction{Op: ir.OpConst, Type: trueType, Result: whenFalse, Value: zeroVal, Span: toIRSpan(path, expression.WhenFalse.Span)})
		} else if expression.WhenTrue != nil && (expression.WhenTrue.Kind == "null" || expression.WhenTrue.Kind == "undefined") {
			trueType = falseType
			whenTrue = nextTemp(counter)
			zeroVal := "0"
			if expression.WhenTrue.Kind == "undefined" {
				switch falseType {
				case ir.TypeNumber:
					zeroVal = "undefined"
				case ir.TypeBool:
					zeroVal = "false"
				default:
					zeroVal = "undefined"
				}
			} else if falseType == ir.TypeString || falseType == ir.TypeNumber || strings.HasPrefix(string(falseType), "object:") || falseType == ir.TypePointer {
				zeroVal = "null"
			}
			thenFn.Body = append(thenFn.Body, ir.Instruction{Op: ir.OpConst, Type: falseType, Result: whenTrue, Value: zeroVal, Span: toIRSpan(path, expression.WhenTrue.Span)})
		} else if trueType == ir.TypeUnknown {
			boxed := nextTemp(counter)
			elseFn.Body = append(elseFn.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{whenFalse}, Span: toIRSpan(path, expression.WhenFalse.Span)})
			whenFalse = boxed
		} else if falseType == ir.TypeUnknown {
			boxed := nextTemp(counter)
			thenFn.Body = append(thenFn.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{whenTrue}, Span: toIRSpan(path, expression.WhenTrue.Span)})
			whenTrue = boxed
			trueType = ir.TypeUnknown
		} else {
			boxedTrue := nextTemp(counter)
			thenFn.Body = append(thenFn.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxedTrue, Args: []string{whenTrue}, Span: toIRSpan(path, expression.WhenTrue.Span)})
			whenTrue = boxedTrue
			boxedFalse := nextTemp(counter)
			elseFn.Body = append(elseFn.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxedFalse, Args: []string{whenFalse}, Span: toIRSpan(path, expression.WhenFalse.Span)})
			whenFalse = boxedFalse
			trueType = ir.TypeUnknown
		}
	}
	initVal := "0"
	if trueType == ir.TypeBool {
		initVal = "false"
	} else if trueType == ir.TypeString {
		initVal = ""
	} else if trueType == ir.TypeUnknown {
		initVal = "undefined"
	} else if strings.HasPrefix(string(trueType), "object:") || trueType == ir.TypePointer {
		initVal = "null"
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   trueType,
		Result: resSlot,
		Value:  initVal,
		Span:   toIRSpan(path, expression.Span),
	})
	thenFn.Body = append(thenFn.Body, ir.Instruction{
		Op:     ir.OpAssign,
		Type:   trueType,
		Result: resSlot,
		Args:   []string{whenTrue},
		Span:   toIRSpan(path, expression.Span),
	})
	elseFn.Body = append(elseFn.Body, ir.Instruction{
		Op:     ir.OpAssign,
		Type:   trueType,
		Result: resSlot,
		Args:   []string{whenFalse},
		Span:   toIRSpan(path, expression.Span),
	})
	function.Body = append(function.Body, ir.Instruction{
		Op:   ir.OpIf,
		Type: ir.TypeVoid,
		Args: []string{condition},
		Then: thenFn.Body,
		Else: elseFn.Body,
		Span: toIRSpan(path, expression.Span),
	})
	env[resSlot] = trueType
	return resSlot, trueType, nil
}

func lowerAwaitExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	val, typ, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if result == "" {
		result = nextTemp(counter)
	}
	retType := typ
	isPromise := false
	if strings.HasPrefix(string(typ), "object:Promise_") {
		isPromise = true
		// Specialized Promise types are spelled Promise_T or Promise__T.
		inner := strings.TrimPrefix(strings.TrimPrefix(string(typ), "object:Promise_"), "_")
		retType = toIRType(inner)
	} else if strings.HasPrefix(string(typ), "object:Promise<") && strings.HasSuffix(string(typ), ">") {
		isPromise = true
		inner := strings.TrimSuffix(strings.TrimPrefix(string(typ), "object:Promise<"), ">")
		retType = toIRType(inner)
	} else if typ == ir.Type("object:Promise") {
		isPromise = true
		if resolved, ok := asyncResolvedReturnType(expression.InferredType); ok {
			retType = resolved
		} else if expression.Left != nil {
			// Native dispatch may erase Promise<T> to object:Promise;
			// recover T from the operand's checked type before falling back
			// to unknown so post-await member lowering keeps its type.
			if resolved, ok := asyncResolvedReturnType(expression.Left.InferredType); ok {
				retType = resolved
			} else {
				retType = ir.TypeUnknown
			}
		} else {
			retType = ir.TypeUnknown
		}
	} else if typ == ir.TypeUnknown && expression.InferredType != "" && !strings.Contains(expression.InferredType, "Promise") {
		retType = toIRType(expression.InferredType)
	}
	if !isPromise && typ != ir.TypeUnknown {
		// JavaScript awaits ordinary values without entering the Promise runtime.
		// Preserve the requested result name so variable declarations remain SSA-valid.
		if result != val {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCheckedCast,
				Type:   retType,
				Result: result,
				Args:   []string{val},
				Span:   toIRSpan(path, expression.Span),
			})
		}
		return result, retType, nil
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   retType,
		Result: result,
		Callee: "__async.await",
		Args:   []string{val},
		Span:   toIRSpan(path, expression.Span),
	})
	return result, retType, nil
}
