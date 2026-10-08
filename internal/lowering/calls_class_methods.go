package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerInstanceMethodCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, receiver string, target ir.Function, mangled string) (string, ir.Type, error) {
	args := []string{receiver}
	for aIdx, argument := range expression.Arguments {
		pIdx := aIdx + 1
		if argument.Kind == "array" && (argument.InferredType == "" || argument.InferredType == "never[]" || argument.InferredType == "unknown[]") && pIdx < len(target.Parameters) && strings.HasSuffix(string(target.Parameters[pIdx].Type), "[]") {
			argument.InferredType = string(target.Parameters[pIdx].Type)
		}
		argVal, valType, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if pIdx < len(target.Parameters) {
			argVal, valType, _ = adaptStructuralObjectArgument(path, toIRSpan(path, argument.Span), argVal, valType, target.Parameters[pIdx].Type, function, counter, shapes)
		}
		hasRest := (restParamsIndex[mangled] || restParamsIndex[strings.Split(mangled, "__")[0]]) && len(target.Parameters) > 0
		restIsUnknown := hasRest && target.Parameters[len(target.Parameters)-1].Type == ir.TypeUnknownArray
		fixed := len(target.Parameters) - 1
		needsUnknownBox := (pIdx < len(target.Parameters) && target.Parameters[pIdx].Type == ir.TypeUnknown) || (restIsUnknown && pIdx >= fixed)
		if needsUnknownBox && valType != ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: boxed,
				Args:   []string{argVal},
				Span:   toIRSpan(path, argument.Span),
			})
			argVal = boxed
		} else if pIdx < len(target.Parameters) && target.Parameters[pIdx].Type == ir.TypeString && argument.Kind == "undefined" {
			undefinedConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: undefinedConst, Value: "undefined", Span: toIRSpan(path, argument.Span)})
			argVal = undefinedConst
		} else if pIdx < len(target.Parameters) && isPointerLikeType(target.Parameters[pIdx].Type) && (argument.Kind == "null" || argument.Kind == "undefined") {
			nullConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   target.Parameters[pIdx].Type,
				Result: nullConst,
				Value:  "null",
				Span:   toIRSpan(path, argument.Span),
			})
			argVal = nullConst
		}
		args = append(args, argVal)
	}
	if (restParamsIndex[mangled] || restParamsIndex[strings.Split(mangled, "__")[0]]) && len(target.Parameters) > 0 && (strings.HasSuffix(string(target.Parameters[len(target.Parameters)-1].Type), "[]") || target.Parameters[len(target.Parameters)-1].Type == ir.TypeStringArray || target.Parameters[len(target.Parameters)-1].Type == ir.TypeNumberArray) {
		restType := target.Parameters[len(target.Parameters)-1].Type
		fixed := len(target.Parameters) - 1
		if len(args) >= fixed {
			restArgs := append([]string(nil), args[fixed:]...)
			args = args[:fixed]
			array := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: restType, Result: array, Args: restArgs, Span: toIRSpan(path, expression.Span)})
			args = append(args, array)
		}
	}
	if len(args) < len(target.Parameters) {
		defaults := defaultParamsIndex[mangled]
		if defaults == nil {
			defaults = defaultParamsIndex[strings.Split(mangled, "__")[0]]
		}
		for i := len(args); i < len(target.Parameters); i++ {
			var val string
			var valType ir.Type
			paramType := target.Parameters[i].Type
			if defaults != nil && defaults[i] != nil {
				initExpr := defaults[i]
				if paramType == ir.TypeNumber && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
					numConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   ir.TypeNumber,
						Result: numConst,
						Value:  "0",
						Span:   toIRSpan(path, initExpr.Span),
					})
					val = numConst
					valType = ir.TypeNumber
				} else if paramType == ir.TypeBool && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
					boolConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   ir.TypeBool,
						Result: boolConst,
						Value:  "false",
						Span:   toIRSpan(path, initExpr.Span),
					})
					val = boolConst
					valType = ir.TypeBool
				} else if paramType == ir.TypeBigInt && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
					biConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   ir.TypeBigInt,
						Result: biConst,
						Value:  "0",
						Span:   toIRSpan(path, initExpr.Span),
					})
					val = biConst
					valType = ir.TypeBigInt
				} else if paramType == ir.TypeString && initExpr.Kind == "undefined" {
					strConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op: ir.OpConst, Type: ir.TypeString, Result: strConst, Value: "undefined",
						Span: toIRSpan(path, initExpr.Span),
					})
					val = strConst
					valType = ir.TypeString
				} else if (strings.HasPrefix(string(paramType), "object:") || isPointerLikeType(paramType)) && (initExpr.Kind == "null" || initExpr.Kind == "undefined") {
					valStr := "null"
					if initExpr.Kind == "undefined" {
						valStr = "undefined"
					}
					nullConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   paramType,
						Result: nullConst,
						Value:  valStr,
						Span:   toIRSpan(path, initExpr.Span),
					})
					val = nullConst
					valType = paramType
				} else {
					v, vt, err := lowerExpression(path, initExpr, "", function, env, counter, shapes, signatures)
					if err == nil {
						val = v
						valType = vt
					}
				}
			}
			if val == "" {
				val = nextTemp(counter)
				valType = paramType
				defStr := "0"
				if paramType == ir.TypeBool {
					defStr = "false"
				} else if paramType == ir.TypeString {
					// Missing optional strings must retain undefined rather than
					// collapsing into the valid empty-string value.
					defStr = "undefined"
				} else if isPointerLikeType(paramType) || strings.HasPrefix(string(paramType), "object:") {
					defStr = "null"
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   paramType,
					Result: val,
					Value:  defStr,
					Span:   toIRSpan(path, expression.Span),
				})
			}
			if paramType == ir.TypeUnknown && valType != ir.TypeUnknown {
				boxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpBoxUnknown,
					Type:   ir.TypeUnknown,
					Result: boxed,
					Args:   []string{val},
					Span:   toIRSpan(path, expression.Span),
				})
				val = boxed
			}
			args = append(args, val)
		}
	}
	if result == "" {
		result = nextTemp(counter)
	}
	retType := target.ReturnType
	if retType == "" {
		retType = ir.TypeVoid
	}
	// A single IR function represents an overloaded implementation,
	// so its ABI may be `unknown` even when TypeScript selected a
	// concrete overload for this call. Keep the call ABI intact and
	// narrow the boxed result at the call boundary.
	selectedType := ir.Type("")
	if expression.InferredType != "" && expression.InferredType != "this" && expression.InferredType != "object:this" {
		inferred := toIRType(expression.InferredType)
		if inferred != "" && inferred != ir.TypeUnknown && inferred != retType {
			selectedType = inferred
		}
	}
	callResult := result
	if selectedType != "" {
		callResult = nextTemp(counter)
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   retType,
		Result: callResult,
		Callee: mangled,
		Args:   args,
		Span:   toIRSpan(path, expression.Span),
	})
	if selectedType != "" {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCheckedCast,
			Type:   selectedType,
			Result: result,
			Args:   []string{callResult},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, selectedType, nil
	}
	return result, retType, nil

}

func lowerStaticMethodCall(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, target ir.Function, mangled string) (string, ir.Type, error) {
	args := make([]string, 0, len(expression.Arguments))
	for aIdx, argument := range expression.Arguments {
		val, valType, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		hasRest := (restParamsIndex[mangled] || restParamsIndex[strings.Split(mangled, "__")[0]]) && len(target.Parameters) > 0
		restIsUnknown := hasRest && target.Parameters[len(target.Parameters)-1].Type == ir.TypeUnknownArray
		fixed := len(target.Parameters) - 1
		needsUnknownBox := (aIdx < len(target.Parameters) && target.Parameters[aIdx].Type == ir.TypeUnknown) || (restIsUnknown && aIdx >= fixed)
		if needsUnknownBox && valType != ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: boxed,
				Args:   []string{val},
				Span:   toIRSpan(path, argument.Span),
			})
			val = boxed
		} else if aIdx < len(target.Parameters) && target.Parameters[aIdx].Type == ir.TypeString && argument.Kind == "undefined" {
			undefinedConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: undefinedConst, Value: "undefined", Span: toIRSpan(path, argument.Span)})
			val = undefinedConst
		} else if aIdx < len(target.Parameters) && isPointerLikeType(target.Parameters[aIdx].Type) && (argument.Kind == "null" || argument.Kind == "undefined") {
			valStr := "null"
			if argument.Kind == "undefined" {
				valStr = "undefined"
			}
			nullConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   target.Parameters[aIdx].Type,
				Result: nullConst,
				Value:  valStr,
				Span:   toIRSpan(path, argument.Span),
			})
			val = nullConst
		}
		args = append(args, val)
	}
	if len(args) < len(target.Parameters) {
		defaults := defaultParamsIndex[mangled]
		if defaults == nil {
			defaults = defaultParamsIndex[strings.Split(mangled, "__")[0]]
		}
		if defaults != nil {
			for i := len(args); i < len(target.Parameters); i++ {
				if initExpr, ok := defaults[i]; ok {
					paramType := target.Parameters[i].Type
					if paramType == ir.TypeNumber && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
						numConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   ir.TypeNumber,
							Result: numConst,
							Value:  "0",
							Span:   toIRSpan(path, initExpr.Span),
						})
						args = append(args, numConst)
						continue
					} else if paramType == ir.TypeBool && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
						boolConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   ir.TypeBool,
							Result: boolConst,
							Value:  "false",
							Span:   toIRSpan(path, initExpr.Span),
						})
						args = append(args, boolConst)
						continue
					} else if paramType == ir.TypeBigInt && (initExpr.Kind == "undefined" || initExpr.Kind == "null") {
						biConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   ir.TypeBigInt,
							Result: biConst,
							Value:  "0",
							Span:   toIRSpan(path, initExpr.Span),
						})
						args = append(args, biConst)
						continue
					} else if paramType == ir.TypeString && initExpr.Kind == "undefined" {
						strConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op: ir.OpConst, Type: ir.TypeString, Result: strConst, Value: "undefined",
							Span: toIRSpan(path, initExpr.Span),
						})
						args = append(args, strConst)
						continue
					}
					val, valType, err := lowerExpression(path, initExpr, "", function, env, counter, shapes, signatures)
					if err != nil {
						return "", "", err
					}
					if i < len(target.Parameters) && target.Parameters[i].Type == ir.TypeUnknown && valType != ir.TypeUnknown {
						boxed := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpBoxUnknown,
							Type:   ir.TypeUnknown,
							Result: boxed,
							Args:   []string{val},
							Span:   toIRSpan(path, initExpr.Span),
						})
						val = boxed
					} else if i < len(target.Parameters) && (isPointerLikeType(target.Parameters[i].Type) || strings.HasPrefix(string(target.Parameters[i].Type), "object:")) && (initExpr.Kind == "null" || initExpr.Kind == "undefined") {
						valStr := "null"
						if initExpr.Kind == "undefined" {
							valStr = "undefined"
						}
						nullConst := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   target.Parameters[i].Type,
							Result: nullConst,
							Value:  valStr,
							Span:   toIRSpan(path, initExpr.Span),
						})
						val = nullConst
					}
					args = append(args, val)
				}
			}
		}
	}
	if (restParamsIndex[mangled] || restParamsIndex[strings.Split(mangled, "__")[0]]) && len(target.Parameters) > 0 && strings.HasSuffix(string(target.Parameters[len(target.Parameters)-1].Type), "[]") {
		restType := target.Parameters[len(target.Parameters)-1].Type
		fixed := len(target.Parameters) - 1
		if len(args) >= fixed {
			restArgs := append([]string(nil), args[fixed:]...)
			args = args[:fixed]
			array := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: restType, Result: array, Args: restArgs, Span: toIRSpan(path, expression.Span)})
			args = append(args, array)
		}
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
