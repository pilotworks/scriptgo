package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerUnaryExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if expression.Operator == "++" || expression.Operator == "--" {
		return lowerUpdateLValue(path, expression.Left, expression.Operator, false, result, function, env, counter, shapes, signatures, expression.Span)
	}

	value, valType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if expression.Operator == "void" {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeVoid,
			Result: result,
			Value:  "undefined",
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeVoid, nil
	}
	if expression.Operator == "!" {
		boolVal, err := coerceToBool(path, value, valType, function, counter, expression.Span)
		if err != nil {
			return "", "", err
		}
		if result == "" {
			result = nextTemp(counter)
		}
		falseConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: falseConst, Value: "false", Span: toIRSpan(path, expression.Span)})
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: result, Operator: "==", Args: []string{boolVal, falseConst}, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeBool, nil
	}
	if expression.Operator == "-" {
		if valType != ir.TypeNumber && valType != ir.TypeBigInt {
			return "", "", fmt.Errorf("unary - requires a number or bigint operand")
		}
		if result == "" {
			result = nextTemp(counter)
		}
		zeroConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: valType, Result: zeroConst, Value: "0", Span: toIRSpan(path, expression.Span)})
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: valType, Result: result, Operator: "-", Args: []string{zeroConst, value}, Span: toIRSpan(path, expression.Span)})
		return result, valType, nil
	}
	if expression.Operator == "+" {
		if valType == ir.TypeString {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: result,
				Callee: "__number.parseFloat",
				Args:   []string{value},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
		if valType != ir.TypeNumber {
			return "", "", fmt.Errorf("unary + requires a number operand")
		}
		if result != "" && result != value {
			zeroConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, expression.Span)})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeNumber, Result: result, Operator: "+", Args: []string{value, zeroConst}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeNumber, nil
		}
		return value, ir.TypeNumber, nil
	}
	if expression.Operator == "~" {
		if valType != ir.TypeNumber && valType != ir.TypeBigInt {
			return "", "", fmt.Errorf("unary ~ requires a number or bigint operand")
		}
		if result == "" {
			result = nextTemp(counter)
		}
		minusOneConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: valType, Result: minusOneConst, Value: "-1", Span: toIRSpan(path, expression.Span)})
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: valType, Result: result, Operator: "^", Args: []string{value, minusOneConst}, Span: toIRSpan(path, expression.Span)})
		return result, valType, nil
	}
	return "", "", fmt.Errorf("unsupported unary operator %q", expression.Operator)
}

func lowerPostfixUnaryExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if expression.Operator == "++" || expression.Operator == "--" {
		return lowerUpdateLValue(path, expression.Left, expression.Operator, true, result, function, env, counter, shapes, signatures, expression.Span)
	}
	return "", "", fmt.Errorf("unsupported postfix operator %q", expression.Operator)
}

func lowerUpdateLValue(path string, lvalue *typescriptgo.SyntaxExpression, op string, isPostfix bool, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, span typescriptgo.SourceSpan) (string, ir.Type, error) {
	if lvalue == nil {
		return "", "", fmt.Errorf("invalid operand for %s", op)
	}

	binOp := "+"
	if op == "--" {
		binOp = "-"
	}

	switch lvalue.Kind {
	case "identifier":
		varName := lvalue.Text
		varType, ok := env[varName]
		if !ok {
			// Check static field
			for clsName, meta := range classHierarchy {
				if _, isStatic := meta.Statics[varName]; isStatic {
					varName = clsName + "_" + varName
					varType, ok = env[varName]
					break
				}
			}
		}
		if !ok {
			return "", "", fmt.Errorf("unknown identifier %q for %s", lvalue.Text, op)
		}
		if varType != ir.TypeNumber && varType != ir.TypeBigInt {
			return "", "", fmt.Errorf("operator %s requires a number or bigint variable, got %s", op, varType)
		}

		oneConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   varType,
			Result: oneConst,
			Value:  "1",
			Span:   toIRSpan(path, span),
		})

		zeroConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   varType,
			Result: zeroConst,
			Value:  "0",
			Span:   toIRSpan(path, span),
		})

		if isPostfix {
			oldVal := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     varType,
				Result:   oldVal,
				Operator: "+",
				Args:     []string{varName, zeroConst},
				Span:     toIRSpan(path, span),
			})

			newVal := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     varType,
				Result:   newVal,
				Operator: binOp,
				Args:     []string{varName, oneConst},
				Span:     toIRSpan(path, span),
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   varType,
				Result: varName,
				Args:   []string{newVal},
				Span:   toIRSpan(path, span),
			})

			if result != "" && result != oldVal {
				function.Body = append(function.Body, ir.Instruction{
					Op:       ir.OpBinary,
					Type:     varType,
					Result:   result,
					Operator: "+",
					Args:     []string{oldVal, zeroConst},
					Span:     toIRSpan(path, span),
				})
				return result, varType, nil
			}
			return oldVal, varType, nil
		}

		// Prefix
		newVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:       ir.OpBinary,
			Type:     varType,
			Result:   newVal,
			Operator: binOp,
			Args:     []string{varName, oneConst},
			Span:     toIRSpan(path, span),
		})
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   varType,
			Result: varName,
			Args:   []string{newVal},
			Span:   toIRSpan(path, span),
		})

		if result != "" && result != newVal {
			function.Body = append(function.Body, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     varType,
				Result:   result,
				Operator: "+",
				Args:     []string{newVal, zeroConst},
				Span:     toIRSpan(path, span),
			})
			return result, varType, nil
		}
		return newVal, varType, nil

	case "property":
		if lvalue.Left != nil && lvalue.Left.Kind == "identifier" {
			if meta, isClass := classHierarchy[lvalue.Left.Text]; isClass {
				if _, isStatic := meta.Statics[lvalue.Text]; isStatic {
					staticVar := lvalue.Left.Text + "_" + lvalue.Text
					varType, ok := env[staticVar]
					if !ok {
						varType = ir.TypeNumber
						env[staticVar] = varType
					}
					oneConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   varType,
						Result: oneConst,
						Value:  "1",
						Span:   toIRSpan(path, span),
					})

					zeroConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   varType,
						Result: zeroConst,
						Value:  "0",
						Span:   toIRSpan(path, span),
					})

					if isPostfix {
						oldVal := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:       ir.OpBinary,
							Type:     varType,
							Result:   oldVal,
							Operator: "+",
							Args:     []string{staticVar, zeroConst},
							Span:     toIRSpan(path, span),
						})

						newVal := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{
							Op:       ir.OpBinary,
							Type:     varType,
							Result:   newVal,
							Operator: binOp,
							Args:     []string{staticVar, oneConst},
							Span:     toIRSpan(path, span),
						})
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpAssign,
							Type:   varType,
							Result: staticVar,
							Args:   []string{newVal},
							Span:   toIRSpan(path, span),
						})

						if result != "" && result != oldVal {
							function.Body = append(function.Body, ir.Instruction{
								Op:       ir.OpBinary,
								Type:     varType,
								Result:   result,
								Operator: "+",
								Args:     []string{oldVal, zeroConst},
								Span:     toIRSpan(path, span),
							})
							return result, varType, nil
						}
						return oldVal, varType, nil
					}

					// Prefix
					newVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op:       ir.OpBinary,
						Type:     varType,
						Result:   newVal,
						Operator: binOp,
						Args:     []string{staticVar, oneConst},
						Span:     toIRSpan(path, span),
					})
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpAssign,
						Type:   varType,
						Result: staticVar,
						Args:   []string{newVal},
						Span:   toIRSpan(path, span),
					})

					if result != "" && result != newVal {
						function.Body = append(function.Body, ir.Instruction{
							Op:       ir.OpBinary,
							Type:     varType,
							Result:   result,
							Operator: "+",
							Args:     []string{newVal, zeroConst},
							Span:     toIRSpan(path, span),
						})
						return result, varType, nil
					}
					return newVal, varType, nil
				}
			}
		}
		objVal, objType, err := lowerExpression(path, lvalue.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		className := strings.TrimPrefix(string(objType), "object:")
		shape, ok := shapes[className]
		if !ok {
			return "", "", fmt.Errorf("unknown object shape %q for %s", className, op)
		}
		var fieldType ir.Type
		found := false
		for _, f := range shape.Fields {
			if f.Name == lvalue.Text {
				fieldType = f.Type
				found = true
				break
			}
		}
		if !found {
			return "", "", fmt.Errorf("unknown field %q on object %q", lvalue.Text, className)
		}
		if fieldType != ir.TypeNumber && fieldType != ir.TypeBigInt {
			return "", "", fmt.Errorf("operator %s requires a number or bigint field, got %s", op, fieldType)
		}

		fIdx := fieldIndex(shape, lvalue.Text)
		currentVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:         ir.OpFieldGet,
			Type:       fieldType,
			Result:     currentVal,
			Callee:     className,
			Field:      lvalue.Text,
			FieldIndex: fIdx,
			Args:       []string{objVal},
			Span:       toIRSpan(path, span),
		})

		oneConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   fieldType,
			Result: oneConst,
			Value:  "1",
			Span:   toIRSpan(path, span),
		})

		zeroConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   fieldType,
			Result: zeroConst,
			Value:  "0",
			Span:   toIRSpan(path, span),
		})

		newVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:       ir.OpBinary,
			Type:     fieldType,
			Result:   newVal,
			Operator: binOp,
			Args:     []string{currentVal, oneConst},
			Span:     toIRSpan(path, span),
		})

		function.Body = append(function.Body, ir.Instruction{
			Op:         ir.OpFieldSet,
			Type:       ir.TypeVoid,
			Callee:     className,
			Field:      lvalue.Text,
			FieldIndex: fIdx,
			Args:       []string{objVal, newVal},
			Span:       toIRSpan(path, span),
		})

		retVal := newVal
		if isPostfix {
			retVal = currentVal
		}
		if result != "" && result != retVal {
			function.Body = append(function.Body, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     fieldType,
				Result:   result,
				Operator: "+",
				Args:     []string{retVal, zeroConst},
				Span:     toIRSpan(path, span),
			})
			return result, fieldType, nil
		}
		return retVal, fieldType, nil

	case "index":
		arrVal, arrType, err := lowerExpression(path, lvalue.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		idxVal, _, err := lowerExpression(path, lvalue.Right, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if arrType != ir.TypeNumberArray && arrType != "number[]" {
			return "", "", fmt.Errorf("operator %s requires a number array, got %s", op, arrType)
		}

		currentVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpIndex,
			Type:   ir.TypeNumber,
			Result: currentVal,
			Args:   []string{arrVal, idxVal},
			Span:   toIRSpan(path, span),
		})

		oneConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeNumber,
			Result: oneConst,
			Value:  "1",
			Span:   toIRSpan(path, span),
		})

		zeroConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeNumber,
			Result: zeroConst,
			Value:  "0",
			Span:   toIRSpan(path, span),
		})

		newVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:       ir.OpBinary,
			Type:     ir.TypeNumber,
			Result:   newVal,
			Operator: binOp,
			Args:     []string{currentVal, oneConst},
			Span:     toIRSpan(path, span),
		})

		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpIndexSet,
			Type: ir.TypeVoid,
			Args: []string{arrVal, idxVal, newVal},
			Span: toIRSpan(path, span),
		})

		retVal := newVal
		if isPostfix {
			retVal = currentVal
		}
		if result != "" && result != retVal {
			function.Body = append(function.Body, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     ir.TypeNumber,
				Result:   result,
				Operator: "+",
				Args:     []string{retVal, zeroConst},
				Span:     toIRSpan(path, span),
			})
			return result, ir.TypeNumber, nil
		}
		return retVal, ir.TypeNumber, nil

	default:
		return "", "", fmt.Errorf("invalid target %q for operator %s", lvalue.Kind, op)
	}
}

func lowerInExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	rightVal, rightType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}

	if result == "" {
		result = nextTemp(counter)
	}

	// 1. Array check: idx in arr
	if strings.HasSuffix(string(rightType), "[]") || rightType == ir.TypeNumberArray || rightType == ir.TypeStringArray {
		leftVal, leftType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if leftType != ir.TypeNumber {
			return "", "", fmt.Errorf("operator \"in\" with array operand requires numeric index, got %s", leftType)
		}
		zeroConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, expression.Span)})
		cmpGeZero := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpGeZero, Operator: ">=", Args: []string{leftVal, zeroConst}, Span: toIRSpan(path, expression.Span)})

		lenVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: lenVal, Callee: "__array.length", Args: []string{rightVal}, Span: toIRSpan(path, expression.Span)})

		cmpLtLen := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpLtLen, Operator: "<", Args: []string{leftVal, lenVal}, Span: toIRSpan(path, expression.Span)})

		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: result, Operator: "&&", Args: []string{cmpGeZero, cmpLtLen}, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeBool, nil
	}

	// 2. Object shape / Class check: "prop" in obj
	if rightType == ir.TypeObject || rightType == ir.TypeUnknown {
		if expression.Left != nil && (expression.Left.Kind == "string" || expression.Left.Kind == "literal") {
			fieldName := strings.Trim(expression.Left.Text, "\"'`")
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpInstanceOf,
				Type:   ir.TypeBool,
				Result: result,
				Value:  fieldName,
				Args:   []string{rightVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, nil
		}
	}
	if after, ok := strings.CutPrefix(string(rightType), "object:"); ok {
		className := after
		shape, ok := shapes[className]
		if !ok {
			if s, exists := anonymousShapes[className]; exists {
				shape = s
				ok = true
			} else if s, exists := registeredShapes[className]; exists {
				shape = s
				ok = true
			}
		}
		if !ok {
			if expression.Left != nil && (expression.Left.Kind == "string" || expression.Left.Kind == "literal") {
				fieldName := strings.Trim(expression.Left.Text, "\"'`")
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpInstanceOf,
					Type:   ir.TypeBool,
					Result: result,
					Value:  fieldName,
					Args:   []string{rightVal},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, ir.TypeBool, nil
			}
			return "", "", fmt.Errorf("unknown object shape %q for \"in\" operator", className)
		}

		// Static string literal check
		if expression.Left != nil && (expression.Left.Kind == "string" || expression.Left.Kind == "literal") {
			fieldName := strings.Trim(expression.Left.Text, "\"'`")
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpInstanceOf,
				Type:   ir.TypeBool,
				Result: result,
				Value:  fieldName,
				Args:   []string{rightVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, nil
		}

		// Dynamic string expression check
		leftVal, leftType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if leftType != ir.TypeString {
			return "", "", fmt.Errorf("operator \"in\" requires string key for object, got %s", leftType)
		}

		if len(shape.Fields) == 0 {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: result, Value: "false", Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeBool, nil
		}

		var lastCond string
		for _, f := range shape.Fields {
			fConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: fConst, Value: f.Name, Span: toIRSpan(path, expression.Span)})
			cmpTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpTemp, Operator: "==", Args: []string{leftVal, fConst}, Span: toIRSpan(path, expression.Span)})
			if lastCond == "" {
				lastCond = cmpTemp
			} else {
				orTemp := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: orTemp, Operator: "||", Args: []string{lastCond, cmpTemp}, Span: toIRSpan(path, expression.Span)})
				lastCond = orTemp
			}
		}

		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: result, Operator: "||", Args: []string{lastCond, lastCond}, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeBool, nil
	}

	return "", "", fmt.Errorf("operator \"in\" requires object or array, got %s", rightType)
}
