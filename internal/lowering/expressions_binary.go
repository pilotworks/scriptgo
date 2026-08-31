package lowering

import (
	"fmt"
	"maps"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerBinaryExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if expression.Operator == "&&" {
		leftVal, leftTyp, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		// Logical operators return one of their operands. The boolean-only
		// short-circuit path is valid only when the whole expression is known to
		// be boolean; otherwise `true && value` must preserve `value`'s type.
		if leftTyp == ir.TypeBool && logicalResultIsBool(expression) {
			res := result
			if res == "" {
				res = nextTemp(counter)
			}
			env[res] = ir.TypeBool
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeBool,
				Result: res,
				Value:  "false",
				Span:   toIRSpan(path, expression.Span),
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   ir.TypeBool,
				Result: res,
				Args:   []string{leftVal},
				Span:   toIRSpan(path, expression.Span),
			})
			thenBlock := ir.Function{Name: "then", ReturnType: function.ReturnType}
			thenEnv := make(map[string]ir.Type, len(env))
			maps.Copy(thenEnv, env)
			rightVal, _, err := lowerExpression(path, expression.Right, "", &thenBlock, thenEnv, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			thenBlock.Body = append(thenBlock.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   ir.TypeBool,
				Result: res,
				Args:   []string{rightVal},
				Span:   toIRSpan(path, expression.Span),
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:   ir.OpIf,
				Type: ir.TypeVoid,
				Args: []string{leftVal},
				Then: thenBlock.Body,
				Span: toIRSpan(path, expression.Span),
			})
			return res, ir.TypeBool, nil
		}
	}
	if expression.Operator == "||" {
		leftVal, leftTyp, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if leftTyp == ir.TypeBool && logicalResultIsBool(expression) {
			res := result
			if res == "" {
				res = nextTemp(counter)
			}
			env[res] = ir.TypeBool
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeBool,
				Result: res,
				Value:  "false",
				Span:   toIRSpan(path, expression.Span),
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   ir.TypeBool,
				Result: res,
				Args:   []string{leftVal},
				Span:   toIRSpan(path, expression.Span),
			})
			elseBlock := ir.Function{Name: "else", ReturnType: function.ReturnType}
			elseEnv := make(map[string]ir.Type, len(env))
			maps.Copy(elseEnv, env)
			rightVal, _, err := lowerExpression(path, expression.Right, "", &elseBlock, elseEnv, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			elseBlock.Body = append(elseBlock.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   ir.TypeBool,
				Result: res,
				Args:   []string{rightVal},
				Span:   toIRSpan(path, expression.Span),
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:   ir.OpIf,
				Type: ir.TypeVoid,
				Args: []string{leftVal},
				Else: elseBlock.Body,
				Span: toIRSpan(path, expression.Span),
			})
			return res, ir.TypeBool, nil
		}
	}
	if expression.Operator == "??" {
		if expression.Left != nil && (expression.Left.Kind == "null" || expression.Left.Kind == "undefined") {
			return lowerExpression(path, expression.Right, result, function, env, counter, shapes, signatures)
		}
		if expression.Right != nil && (expression.Right.Kind == "null" || expression.Right.Kind == "undefined") {
			return lowerExpression(path, expression.Left, result, function, env, counter, shapes, signatures)
		}
		leftVal, leftTyp, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if leftTyp == ir.TypePointer || leftTyp == "ptr" || leftTyp == ir.TypeVoid {
			return lowerExpression(path, expression.Right, result, function, env, counter, shapes, signatures)
		}
		outTyp := leftTyp
		if expression.InferredType != "" {
			infIR := toIRType(expression.InferredType)
			if infIR != "" && infIR != ir.TypeVoid {
				outTyp = infIR
			}
		}
		res := result
		if res == "" || outTyp == ir.TypeUnknown {
			res = nextTemp(counter)
		}
		env[res] = outTyp

		initLeft := leftVal
		if outTyp == ir.TypeUnknown && leftTyp != ir.TypeUnknown {
			boxed := nextTemp(counter)
			env[boxed] = ir.TypeUnknown
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{leftVal}, Span: toIRSpan(path, expression.Span)})
			initLeft = boxed
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCheckedCast,
			Type:   outTyp,
			Result: res,
			Args:   []string{initLeft},
			Span:   toIRSpan(path, expression.Span),
		})

		var cond string
		if leftTyp == ir.TypeNumber {
			cmpNaN := nextTemp(counter)
			env[cmpNaN] = ir.TypeBool
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpNaN, Operator: "==", Args: []string{leftVal, leftVal}, Span: toIRSpan(path, expression.Span)})
			cond = cmpNaN
		} else if leftTyp == ir.TypeBool {
			trueConst := nextTemp(counter)
			env[trueConst] = ir.TypeBool
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: trueConst, Value: "true", Span: toIRSpan(path, expression.Span)})
			cond = trueConst
		} else if leftTyp == ir.TypeUnknown {
			nullConst := nextTemp(counter)
			env[nullConst] = ir.TypeUnknown
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeUnknown, Result: nullConst, Value: "null", Span: toIRSpan(path, expression.Span)})
			cmpNull := nextTemp(counter)
			env[cmpNull] = ir.TypeBool
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpNull, Operator: "!=", Args: []string{leftVal, nullConst}, Span: toIRSpan(path, expression.Span)})

			undefConst := nextTemp(counter)
			env[undefConst] = ir.TypeUnknown
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeUnknown, Result: undefConst, Value: "undefined", Span: toIRSpan(path, expression.Span)})
			cmpUndef := nextTemp(counter)
			env[cmpUndef] = ir.TypeBool
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpUndef, Operator: "!=", Args: []string{leftVal, undefConst}, Span: toIRSpan(path, expression.Span)})

			condTemp := nextTemp(counter)
			env[condTemp] = ir.TypeBool
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: condTemp, Operator: "&&", Args: []string{cmpNull, cmpUndef}, Span: toIRSpan(path, expression.Span)})
			cond = condTemp
		} else {
			nullConst := nextTemp(counter)
			env[nullConst] = leftTyp
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: leftTyp, Result: nullConst, Value: "null", Span: toIRSpan(path, expression.Span)})
			cmpNull := nextTemp(counter)
			env[cmpNull] = ir.TypeBool
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpNull, Operator: "!=", Args: []string{leftVal, nullConst}, Span: toIRSpan(path, expression.Span)})

			undefConst := nextTemp(counter)
			env[undefConst] = leftTyp
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: leftTyp, Result: undefConst, Value: "undefined", Span: toIRSpan(path, expression.Span)})
			cmpUndef := nextTemp(counter)
			env[cmpUndef] = ir.TypeBool
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpUndef, Operator: "!=", Args: []string{leftVal, undefConst}, Span: toIRSpan(path, expression.Span)})

			condTemp := nextTemp(counter)
			env[condTemp] = ir.TypeBool
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: condTemp, Operator: "&&", Args: []string{cmpNull, cmpUndef}, Span: toIRSpan(path, expression.Span)})
			cond = condTemp
		}

		elseBlock := ir.Function{Name: "nullish_fallback", ReturnType: function.ReturnType}
		elseEnv := make(map[string]ir.Type, len(env))
		maps.Copy(elseEnv, env)
		if expression.Right != nil && (expression.Right.InferredType == "" || expression.Right.InferredType == "{}" || expression.Right.InferredType == "never[]" || expression.Right.InferredType == "unknown[]" || expression.Right.InferredType == "[]") {
			if strings.HasPrefix(string(outTyp), "object:") || strings.HasSuffix(string(outTyp), "[]") {
				expression.Right.InferredType = string(outTyp)
			}
		}
		rightVal, rightTyp, err := lowerExpression(path, expression.Right, "", &elseBlock, elseEnv, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		finalRight := rightVal
		if outTyp == ir.TypeUnknown && rightTyp != ir.TypeUnknown {
			boxed := nextTemp(counter)
			elseEnv[boxed] = ir.TypeUnknown
			elseBlock.Body = append(elseBlock.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{rightVal}, Span: toIRSpan(path, expression.Span)})
			finalRight = boxed
		} else if strings.HasPrefix(string(outTyp), "object:") && rightTyp != outTyp {
			castTemp := nextTemp(counter)
			elseEnv[castTemp] = outTyp
			elseBlock.Body = append(elseBlock.Body, ir.Instruction{
				Op:     ir.OpCheckedCast,
				Type:   outTyp,
				Result: castTemp,
				Args:   []string{rightVal},
				Span:   toIRSpan(path, expression.Span),
			})
			finalRight = castTemp
		}
		elseBlock.Body = append(elseBlock.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   outTyp,
			Result: res,
			Args:   []string{finalRight},
			Span:   toIRSpan(path, expression.Span),
		})

		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpIf,
			Type: ir.TypeVoid,
			Args: []string{cond},
			Else: elseBlock.Body,
			Span: toIRSpan(path, expression.Span),
		})
		return res, outTyp, nil
	}
	if expression.Operator == "instanceof" {
		left, _, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		targetClass := callName(expression.Right)
		if targetClass == "" && expression.Right != nil && (expression.Right.Kind == "identifier" || expression.Right.Kind == "type") {
			targetClass = expression.Right.Text
		}
		if targetClass == "" {
			return "", "", fmt.Errorf("instanceof requires a class identifier on the right")
		}
		if idx := strings.LastIndex(targetClass, "."); idx != -1 {
			if _, exists := classHierarchy[targetClass]; !exists {
				shortName := targetClass[idx+1:]
				if _, exists2 := classHierarchy[shortName]; exists2 {
					targetClass = shortName
				}
			}
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpInstanceOf,
			Type:   ir.TypeBool,
			Result: result,
			Args:   []string{left},
			Value:  targetClass,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeBool, nil
	}
	if expression.Operator == "in" {
		return lowerInExpression(path, expression, result, function, env, counter, shapes, signatures)
	}
	if expression.Operator == "," {
		_, _, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		return lowerExpression(path, expression.Right, result, function, env, counter, shapes, signatures)
	}
	if expression.Operator == "=" || (strings.HasSuffix(expression.Operator, "=") && expression.Operator != "==" && expression.Operator != "===" && expression.Operator != "!=" && expression.Operator != "!==" && expression.Operator != "<=" && expression.Operator != ">=") {
		if expression.Left != nil && expression.Left.Kind == "identifier" {
			varName := expression.Left.Text
			varType, ok := env[varName]
			if !ok {
				if topVar, isTop := topLevelVars[varName]; isTop {
					varType = toIRType(topVar.Type)
					if varType == "" {
						varType = toIRType(topVar.InferredType)
					}
					if varType == "" {
						varType = ir.TypeNumber
					}
					ok = true
				}
			}
			if !ok {
				return "", "", fmt.Errorf("assignment to unknown variable %q", varName)
			}
			rhsExpr := expression.Right
			if expression.Operator != "=" {
				baseOp := strings.TrimSuffix(expression.Operator, "=")
				rhsExpr = &typescriptgo.SyntaxExpression{
					Span:         expression.Span,
					Kind:         "binary",
					Operator:     baseOp,
					Left:         expression.Left,
					Right:        expression.Right,
					InferredType: expression.InferredType,
				}
			}
			val, valType, err := lowerExpression(path, rhsExpr, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			if varType == ir.TypeUnknown && valType != ir.TypeUnknown {
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpBoxUnknown,
					Type:   ir.TypeUnknown,
					Result: varName,
					Args:   []string{val},
					Span:   toIRSpan(path, expression.Span),
				})
				if result != "" {
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpAssign,
						Type:   varType,
						Result: result,
						Args:   []string{varName},
						Span:   toIRSpan(path, expression.Span),
					})
				}
				return result, varType, nil
			}
			if valType != varType && varType != ir.TypeUnknown {
				if (strings.HasPrefix(string(valType), "object:") || valType == ir.TypeObject) && (strings.HasPrefix(string(varType), "object:") || varType == ir.TypeObject) {
					// Polymorphic object assignment
				} else if isPointerLikeType(varType) && (valType == ir.TypeVoid || valType == ir.TypePointer) {
					// Assigning undefined or null to a pointer-like variable
				} else {
					return "", "", fmt.Errorf("assignment type mismatch for %q: %s := %s", varName, varType, valType)
				}
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   varType,
				Result: varName,
				Args:   []string{val},
				Span:   toIRSpan(path, expression.Span),
			})
			if result != "" {
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpAssign,
					Type:   varType,
					Result: result,
					Args:   []string{val},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, varType, nil
			}
			return val, varType, nil
		}
	}
	left, leftType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	right, rightType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if leftType == ir.TypeVoid && rightType == ir.TypeVoid {
		if result == "" {
			result = nextTemp(counter)
		}
		val := "true"
		if expression.Operator == "!==" || expression.Operator == "!=" {
			val = "false"
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeBool,
			Result: result,
			Value:  val,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeBool, nil
	}
	if leftType != rightType {
		if (expression.Operator == "!==" || expression.Operator == "!=") && leftType == ir.TypeNumber && expression.Right != nil && (expression.Right.Kind == "undefined" || expression.Right.Kind == "null") {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:       ir.OpCompare,
				Type:     ir.TypeBool,
				Result:   result,
				Operator: "==",
				Args:     []string{left, left},
				Span:     toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, nil
		}
		if (expression.Operator == "===" || expression.Operator == "==") && leftType == ir.TypeNumber && expression.Right != nil && (expression.Right.Kind == "undefined" || expression.Right.Kind == "null") {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:       ir.OpCompare,
				Type:     ir.TypeBool,
				Result:   result,
				Operator: "!=",
				Args:     []string{left, left},
				Span:     toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, nil
		}
		if (expression.Operator == "!==" || expression.Operator == "!=") && rightType == ir.TypeNumber && expression.Left != nil && (expression.Left.Kind == "undefined" || expression.Left.Kind == "null") {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:       ir.OpCompare,
				Type:     ir.TypeBool,
				Result:   result,
				Operator: "==",
				Args:     []string{right, right},
				Span:     toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, nil
		}
		if (expression.Operator == "===" || expression.Operator == "==") && rightType == ir.TypeNumber && expression.Left != nil && (expression.Left.Kind == "undefined" || expression.Left.Kind == "null") {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:       ir.OpCompare,
				Type:     ir.TypeBool,
				Result:   result,
				Operator: "!=",
				Args:     []string{right, right},
				Span:     toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, nil
		}
		if isComparison(expression.Operator) && (leftType == ir.TypeBool || rightType == ir.TypeBool) && (leftType == ir.TypeVoid || rightType == ir.TypeVoid || (expression.Left != nil && (expression.Left.Kind == "undefined" || expression.Left.Kind == "null")) || (expression.Right != nil && (expression.Right.Kind == "undefined" || expression.Right.Kind == "null"))) {
			if result == "" {
				result = nextTemp(counter)
			}
			isNot := (expression.Operator == "!==" || expression.Operator == "!=")
			constVal := "false"
			if isNot {
				constVal = "true"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeBool,
				Result: result,
				Value:  constVal,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, nil
		}
		if isComparison(expression.Operator) && ((leftType == ir.TypeVoid && (rightType == ir.TypePointer || rightType == "ptr")) || (rightType == ir.TypeVoid && (leftType == ir.TypePointer || leftType == "ptr"))) {
			if result == "" {
				result = nextTemp(counter)
			}
			val := "false"
			if expression.Operator == "!==" {
				val = "true"
			} else if expression.Operator == "!=" {
				val = "false"
			} else if expression.Operator == "==" {
				val = "true"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeBool,
				Result: result,
				Value:  val,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, nil
		}
		if isComparison(expression.Operator) && (expression.Right != nil && (expression.Right.Kind == "null" || expression.Right.Kind == "undefined")) && isPointerLikeType(leftType) {
			right = nextTemp(counter)
			rightType = leftType
			nullVal := "null"
			if expression.Right.Kind == "undefined" {
				nullVal = "undefined"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   leftType,
				Result: right,
				Value:  nullVal,
				Span:   toIRSpan(path, expression.Right.Span),
			})
		} else if isComparison(expression.Operator) && (expression.Left != nil && (expression.Left.Kind == "null" || expression.Left.Kind == "undefined")) && isPointerLikeType(rightType) {
			left = nextTemp(counter)
			leftType = rightType
			nullVal := "null"
			if expression.Left.Kind == "undefined" {
				nullVal = "undefined"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   rightType,
				Result: left,
				Value:  nullVal,
				Span:   toIRSpan(path, expression.Left.Span),
			})
		} else if isComparison(expression.Operator) && (leftType == ir.TypeUnknown || rightType == ir.TypeUnknown) {
			if leftType != ir.TypeUnknown {
				boxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{left}, Span: toIRSpan(path, expression.Span)})
				left = boxed
				leftType = ir.TypeUnknown
			}
			if rightType != ir.TypeUnknown {
				boxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{right}, Span: toIRSpan(path, expression.Span)})
				right = boxed
				rightType = ir.TypeUnknown
			}
		} else if expression.Operator == "+" && (leftType == ir.TypeString || rightType == ir.TypeString) {
			if leftType != ir.TypeString {
				strTemp := nextTemp(counter)
				callee := "__string.fromNumber"
				if leftType == ir.TypeBool {
					callee = "__string.fromBool"
				} else if leftType == ir.TypeBigInt {
					callee = "__string.fromBigInt"
				} else if leftType == ir.TypeUnknown {
					callee = "__string.fromUnknown"
				} else if leftType == ir.TypeObject || strings.HasPrefix(string(leftType), "object:") || leftType == ir.TypePointer {
					callee = "__string.fromObject"
				} else if leftType != ir.TypeNumber {
					return "", "", fmt.Errorf("operator %q does not support %s and %s", expression.Operator, leftType, rightType)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: strTemp, Callee: callee, Args: []string{left}, Span: toIRSpan(path, expression.Span)})
				left = strTemp
				leftType = ir.TypeString
			}
			if rightType != ir.TypeString {
				strTemp := nextTemp(counter)
				callee := "__string.fromNumber"
				if rightType == ir.TypeBool {
					callee = "__string.fromBool"
				} else if rightType == ir.TypeBigInt {
					callee = "__string.fromBigInt"
				} else if rightType == ir.TypeUnknown {
					callee = "__string.fromUnknown"
				} else if rightType == ir.TypeObject || strings.HasPrefix(string(rightType), "object:") || rightType == ir.TypePointer {
					callee = "__string.fromObject"
				} else if rightType != ir.TypeNumber {
					return "", "", fmt.Errorf("operator %q does not support %s and %s", expression.Operator, leftType, rightType)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: strTemp, Callee: callee, Args: []string{right}, Span: toIRSpan(path, expression.Span)})
				right = strTemp
				rightType = ir.TypeString
			}
		} else if (expression.Operator == "||" || expression.Operator == "&&") && (leftType == ir.TypeUnknown || rightType == ir.TypeUnknown) {
			if leftType != ir.TypeUnknown {
				boxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{left}, Span: toIRSpan(path, expression.Span)})
				left = boxed
				leftType = ir.TypeUnknown
			}
			if rightType != ir.TypeUnknown {
				boxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{right}, Span: toIRSpan(path, expression.Span)})
				right = boxed
				rightType = ir.TypeUnknown
			}
		} else if (expression.Operator == "||" || expression.Operator == "&&") && isPointerLikeType(leftType) {
			if rightType == "never[]" || rightType == "unknown[]" {
				rightType = leftType
			} else if isPointerLikeType(rightType) {
				if leftType != rightType && !isSubtype(string(rightType), string(leftType)) && !isSubtype(string(leftType), string(rightType)) {
					boxedLeft := nextTemp(counter)
					env[boxedLeft] = ir.TypeUnknown
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxedLeft, Args: []string{left}, Span: toIRSpan(path, expression.Span)})
					left = boxedLeft
					leftType = ir.TypeUnknown

					boxedRight := nextTemp(counter)
					env[boxedRight] = ir.TypeUnknown
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxedRight, Args: []string{right}, Span: toIRSpan(path, expression.Span)})
					right = boxedRight
					rightType = ir.TypeUnknown
				} else if isSubtype(string(rightType), string(leftType)) {
					rightType = leftType
				}
			} else if rightType == ir.TypeBool {
				boolTemp := nextTemp(counter)
				nullConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: leftType, Result: nullConst, Value: "null", Span: toIRSpan(path, expression.Span)})
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: boolTemp, Operator: "!=", Args: []string{left, nullConst}, Span: toIRSpan(path, expression.Span)})
				left = boolTemp
				leftType = ir.TypeBool
			} else {
				return "", "", fmt.Errorf("operator %q does not support %s and %s", expression.Operator, leftType, rightType)
			}
		} else if (expression.Operator == "||" || expression.Operator == "&&") && isPointerLikeType(rightType) && leftType == ir.TypeBool {
			boolTemp := nextTemp(counter)
			nullConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: rightType, Result: nullConst, Value: "null", Span: toIRSpan(path, expression.Span)})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: boolTemp, Operator: "!=", Args: []string{right, nullConst}, Span: toIRSpan(path, expression.Span)})
			right = boolTemp
			rightType = ir.TypeBool
		} else {
			return "", "", fmt.Errorf("operator %q does not support %s and %s", expression.Operator, leftType, rightType)
		}
	}
	if leftType == ir.TypeBool {
		if expression.Operator == "&&" || expression.Operator == "||" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeBool, nil
		}
		if isComparison(expression.Operator) {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeBool, nil
		}
		return "", "", fmt.Errorf("operator %q does not support bool operands", expression.Operator)
	}
	if isPointerLikeType(leftType) || leftType == ir.TypeSymbol || leftType == ir.TypeClosure || leftType == ir.TypeUnknown {
		if isComparison(expression.Operator) {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeBool, nil
		}
		if expression.Operator == "||" || expression.Operator == "&&" {
			if leftType == ir.TypeUnknown && rightType != ir.TypeUnknown {
				boxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{right}, Span: toIRSpan(path, expression.Span)})
				right = boxed
			} else if rightType == ir.TypeUnknown && leftType != ir.TypeUnknown {
				boxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{left}, Span: toIRSpan(path, expression.Span)})
				left = boxed
				leftType = ir.TypeUnknown
			}
			var cond string
			if leftType == ir.TypeUnknown {
				condTemp := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeBool,
					Result: condTemp,
					Callee: "__scriptgo.is_truthy",
					Args:   []string{left},
					Span:   toIRSpan(path, expression.Span),
				})
				cond = condTemp
			} else {
				nullConst := nextTemp(counter)
				nullVal := "null"
				if leftType == ir.TypeString {
					nullVal = ""
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: leftType, Result: nullConst, Value: nullVal, Span: toIRSpan(path, expression.Span)})
				cmpNull := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpNull, Operator: "!=", Args: []string{left, nullConst}, Span: toIRSpan(path, expression.Span)})
				cond = cmpNull
				if leftType == ir.TypeString {
					undefConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: undefConst, Value: "undefined", Span: toIRSpan(path, expression.Span)})
					cmpUndef := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpUndef, Operator: "!=", Args: []string{left, undefConst}, Span: toIRSpan(path, expression.Span)})
					condTemp := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: condTemp, Operator: "&&", Args: []string{cmpNull, cmpUndef}, Span: toIRSpan(path, expression.Span)})
					cond = condTemp
				}
			}
			if result == "" {
				result = nextTemp(counter)
			}
			selType := leftType
			selLeft := left
			selRight := right
			if leftType != rightType {
				if leftType != ir.TypeUnknown {
					bLeft := nextTemp(counter)
					env[bLeft] = ir.TypeUnknown
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: bLeft, Args: []string{left}, Span: toIRSpan(path, expression.Span)})
					selLeft = bLeft
				}
				if rightType != ir.TypeUnknown {
					bRight := nextTemp(counter)
					env[bRight] = ir.TypeUnknown
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: bRight, Args: []string{right}, Span: toIRSpan(path, expression.Span)})
					selRight = bRight
				}
				selType = ir.TypeUnknown
			}
			env[result] = selType
			if expression.Operator == "||" {
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: selType, Result: result, Args: []string{cond, selLeft, selRight}, Span: toIRSpan(path, expression.Span)})
			} else {
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: selType, Result: result, Args: []string{cond, selRight, selLeft}, Span: toIRSpan(path, expression.Span)})
			}
			return result, selType, nil
		}
		if leftType == ir.TypeString && expression.Operator == "+" {
			// continue to binary +
		} else {
			return "", "", fmt.Errorf("operator %q does not support %s operands", expression.Operator, leftType)
		}
	}
	if leftType != ir.TypeNumber && leftType != ir.TypeString && leftType != ir.TypeBigInt {
		return "", "", fmt.Errorf("operator %q does not support %s and %s", expression.Operator, leftType, rightType)
	}
	if leftType == ir.TypeString {
		if expression.Operator != "+" && !isComparison(expression.Operator) {
			return "", "", fmt.Errorf("operator %q does not support string operands", expression.Operator)
		}
	}
	if isComparison(expression.Operator) {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeBool, nil
	}
	if result == "" {
		result = nextTemp(counter)
	}
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: leftType, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
	return result, leftType, nil
}

func logicalResultIsBool(expression *typescriptgo.SyntaxExpression) bool {
	if expression == nil {
		return false
	}
	if expression.InferredType != "" {
		return toIRType(expression.InferredType) == ir.TypeBool
	}
	if expression.Right == nil {
		return false
	}
	if expression.Right.Kind == "bool" {
		return true
	}
	if expression.Right.Kind == "binary" && isComparison(expression.Right.Operator) {
		return true
	}
	return false
}

func isPointerLikeType(typ ir.Type) bool {
	return strings.HasPrefix(string(typ), "object:") ||
		strings.HasSuffix(string(typ), "[]") ||
		typ == ir.TypeObject ||
		typ == ir.TypeClosure ||
		typ == ir.TypeString ||
		typ == ir.TypeMap ||
		typ == ir.TypeSet ||
		typ == ir.TypeArrayBuffer ||
		typ == ir.TypeDataView ||
		typ == ir.TypeTextEncoder ||
		typ == ir.TypeTextDecoder ||
		typ == ir.TypeBuffer ||
		typ == ir.TypeUint8Array ||
		typ == ir.TypeInt8Array ||
		typ == ir.TypeUint8ClampedArray ||
		typ == ir.TypeInt16Array ||
		typ == ir.TypeUint16Array ||
		typ == ir.TypeInt32Array ||
		typ == ir.TypeUint32Array ||
		typ == ir.TypeFloat32Array ||
		typ == ir.TypeFloat64Array ||
		typ == ir.TypeBigInt64Array ||
		typ == ir.TypeBigUint64Array ||
		typ == ir.TypePointer ||
		typ == "ptr"
}
