package lowering

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerArrayLiteral(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if len(expression.Arguments) == 0 {
		arrType := ir.TypeNumberArray
		if varType, ok := env[result]; ok && strings.HasSuffix(string(varType), "[]") {
			arrType = varType
		} else if expression.InferredType != "" {
			inferredIR := toIRType(expression.InferredType)
			if strings.HasSuffix(string(inferredIR), "[]") || inferredIR == ir.TypeNumberArray || inferredIR == ir.TypeStringArray || inferredIR == ir.TypeBoolArray || inferredIR == ir.TypeBigIntArray || inferredIR == ir.TypeSymbolArray {
				arrType = inferredIR
			}
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: arrType, Result: result, Args: nil, Span: toIRSpan(path, expression.Span)})
		return result, arrType, nil
	}
	if result == "" {
		result = nextTemp(counter)
	}
	hasSpread := false
	for _, elem := range expression.Arguments {
		if elem.Kind == "spread" {
			hasSpread = true
			break
		}
	}
	if !hasSpread {
		arguments := make([]string, 0, len(expression.Arguments))
		types := make([]ir.Type, 0, len(expression.Arguments))
		isHomogeneous := true
		for i, element := range expression.Arguments {
			value, typ, err := lowerExpression(path, element, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			arguments = append(arguments, value)
			types = append(types, typ)
			if i > 0 && typ != types[0] {
				isHomogeneous = false
			}
		}
		inferredTuple := false
		var trimmed string
		if expression.InferredType != "" {
			trimmed = strings.TrimSpace(expression.InferredType)
			trimmed = strings.TrimPrefix(trimmed, "readonly ")
			if strings.HasPrefix(trimmed, "Readonly<") && strings.HasSuffix(trimmed, ">") {
				trimmed = strings.TrimSuffix(strings.TrimPrefix(trimmed, "Readonly<"), ">")
				trimmed = strings.TrimSpace(trimmed)
			}
			if strings.HasSuffix(trimmed, "[]") {
				arrType := toIRType(trimmed)
				if arrType == ir.Type(string(ir.TypeVoid)+"[]") {
					// undefined[] has no unboxed element storage.
					arrType = ir.TypeUnknownArray
				}
				if arrType == ir.TypeUnknownArray {
					for idx, argName := range arguments {
						if types[idx] != ir.TypeUnknown {
							boxed := nextTemp(counter)
							function.Body = append(function.Body, ir.Instruction{
								Op:     ir.OpBoxUnknown,
								Type:   ir.TypeUnknown,
								Result: boxed,
								Args:   []string{argName},
								Span:   toIRSpan(path, expression.Span),
							})
							arguments[idx] = boxed
						}
					}
				} else {
					elemType := arrayLiteralElementType(arrType)
					for idx, argName := range arguments {
						if types[idx] == ir.TypeUnknown && elemType != ir.TypeUnknown {
							casted := nextTemp(counter)
							function.Body = append(function.Body, ir.Instruction{
								Op:     ir.OpCheckedCast,
								Type:   elemType,
								Result: casted,
								Args:   []string{argName},
								Span:   toIRSpan(path, expression.Arguments[idx].Span),
							})
							arguments[idx] = casted
						}
					}
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: arrType, Result: result, Args: arguments, Span: toIRSpan(path, expression.Span)})
				return result, arrType, nil
			} else if tFields, ok := tupleFields(trimmed); ok && len(tFields) >= 1 {
				inferredTuple = true
			} else if shape, ok := shapes[strings.TrimPrefix(trimmed, "object:")]; ok && len(shape.Fields) > 0 && shape.Fields[0].Name == "0" {
				inferredTuple = true
			}
		}
		if isHomogeneous && !inferredTuple && len(types) > 0 && types[0] == ir.TypeVoid {
			// Elements that are all undefined have no unboxed storage; keep
			// them as boxed values in an unknown[] array.
			for idx, argName := range arguments {
				boxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{argName}, Span: toIRSpan(path, expression.Arguments[idx].Span)})
				arguments[idx] = boxed
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: ir.TypeUnknownArray, Result: result, Args: arguments, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeUnknownArray, nil
		}
		if isHomogeneous && !inferredTuple {
			arrType := ir.TypeNumberArray
			if len(types) > 0 && types[0] == ir.TypeString {
				arrType = ir.TypeStringArray
			} else if len(types) > 0 && types[0] != ir.TypeNumber {
				arrType = ir.Type(string(types[0]) + "[]")
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: arrType, Result: result, Args: arguments, Span: toIRSpan(path, expression.Span)})
			return result, arrType, nil
		}

		if strings.Contains(trimmed, "...") || toIRType(trimmed) == ir.TypeUnknownArray {
			arrType := ir.TypeUnknownArray
			for idx, argName := range arguments {
				if types[idx] != ir.TypeUnknown {
					boxed := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpBoxUnknown,
						Type:   ir.TypeUnknown,
						Result: boxed,
						Args:   []string{argName},
						Span:   toIRSpan(path, expression.Arguments[idx].Span),
					})
					arguments[idx] = boxed
				}
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: arrType, Result: result, Args: arguments, Span: toIRSpan(path, expression.Span)})
			return result, arrType, nil
		}

		// Heterogeneous elements -> lower as anonymous tuple object
		var fields []ir.Field
		if tFields, ok := tupleFields(trimmed); ok && len(tFields) > 0 {
			fields = tFields
		} else if shape, ok := shapes[strings.TrimPrefix(trimmed, "object:")]; ok && len(shape.Fields) > 0 && shape.Fields[0].Name == "0" {
			fields = shape.Fields
		} else {
			for i, typ := range types {
				fields = append(fields, ir.Field{
					Name: strconv.Itoa(i),
					Type: typ,
					Span: toIRSpan(path, expression.Arguments[i].Span),
				})
			}
		}
		shapeName := anonymousShapeName(fields)
		if _, ok := shapes[shapeName]; !ok {
			shapes[shapeName] = ir.ObjectShape{
				Name:   shapeName,
				Span:   toIRSpan(path, expression.Span),
				Fields: fields,
			}
		}
		objType := ir.Type("object:" + shapeName)
		function.Body = append(function.Body, ir.Instruction{
			Op:         ir.OpObjectNew,
			Type:       objType,
			Result:     result,
			Callee:     shapeName,
			FieldCount: len(fields),
			Span:       toIRSpan(path, expression.Span),
		})
		for i, field := range fields {
			var val string
			if i < len(arguments) {
				val = arguments[i]
			} else {
				defVal := "undefined"
				if field.Type == ir.TypeNumber {
					defVal = "NaN"
				} else if field.Type == ir.TypeBool {
					defVal = "false"
				}
				defConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   field.Type,
					Result: defConst,
					Value:  defVal,
					Span:   toIRSpan(path, expression.Span),
				})
				val = defConst
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:         ir.OpFieldSet,
				Type:       ir.TypeVoid,
				Callee:     shapeName,
				Field:      field.Name,
				FieldIndex: i,
				Args:       []string{result, val},
				Span:       toIRSpan(path, expression.Span),
			})
		}
		return result, objType, nil
	}
	arrType := ir.TypeNumberArray
	if expression.InferredType != "" && expression.InferredType != "never[]" && expression.InferredType != "unknown[]" {
		inferred := toIRType(expression.InferredType)
		if strings.HasSuffix(string(inferred), "[]") {
			arrType = inferred
		}
	}
	if len(expression.Arguments) > 0 {
		firstElem := expression.Arguments[0]
		if firstElem.Kind != "spread" {
			_, typ, err := lowerExpression(path, firstElem, "", function, env, counter, shapes, signatures)
			if err == nil {
				switch typ {
				case ir.TypeString:
					arrType = ir.TypeStringArray
				case ir.TypeNumber:
					arrType = ir.TypeNumberArray
				case ir.TypeBool:
					arrType = ir.TypeBoolArray
				case ir.TypeBigInt:
					arrType = ir.TypeBigIntArray
				case ir.TypeSymbol:
					arrType = ir.TypeSymbolArray
				default:
					if strings.HasPrefix(string(typ), "object:") {
						arrType = ir.Type(string(typ) + "[]")
					}
				}
			}
		} else if firstElem.Left != nil {
			if firstElem.Left.InferredType != "" {
				inf := toIRType(firstElem.Left.InferredType)
				if strings.HasSuffix(string(inf), "[]") {
					arrType = inf
				}
			} else if firstElem.Left.Kind == "identifier" {
				if t, ok := env[firstElem.Left.Text]; ok && strings.HasSuffix(string(t), "[]") {
					arrType = t
				}
			}
		}
	}
	elemTypeStr := ""
	if strings.HasSuffix(string(arrType), "[]") {
		elemTypeStr = strings.TrimSuffix(string(arrType), "[]")
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpArray,
		Type:   arrType,
		Result: result,
		Args:   nil,
		Span:   toIRSpan(path, expression.Span),
	})
	for _, elem := range expression.Arguments {
		if elem.Kind == "object_literal" && elemTypeStr != "" && elem.InferredType == "" {
			elem.InferredType = elemTypeStr
		}
		if elem.Kind == "spread" {
			spreadVal, spreadType, err := lowerExpression(path, elem.Left, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			idxVar := nextTemp(counter)
			lenVar := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: idxVar, Value: "0", Span: toIRSpan(path, elem.Span)})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: lenVar, Callee: "__array.length", Args: []string{spreadVal}, Span: toIRSpan(path, elem.Span)})
			condVar := nextTemp(counter)
			var condBody []ir.Instruction
			condBody = append(condBody, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: condVar, Operator: "<", Args: []string{idxVar, lenVar}, Span: toIRSpan(path, elem.Span)})
			var loopBody []ir.Instruction
			itemVar := nextTemp(counter)
			itemType := arrayElementType(arrType)
			if spreadType != "" && strings.HasSuffix(string(spreadType), "[]") {
				itemType = arrayElementType(spreadType)
			}
			loopBody = append(loopBody, ir.Instruction{Op: ir.OpIndex, Type: itemType, Result: itemVar, Args: []string{spreadVal, idxVar}, Span: toIRSpan(path, elem.Span)})
			pushRes := nextTemp(counter)
			loopBody = append(loopBody, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: pushRes, Callee: "__array.push", Args: []string{result, itemVar}, Span: toIRSpan(path, elem.Span)})
			oneConst := nextTemp(counter)
			loopBody = append(loopBody, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: oneConst, Value: "1", Span: toIRSpan(path, elem.Span)})
			nextIdx := nextTemp(counter)
			loopBody = append(loopBody, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeNumber, Result: nextIdx, Operator: "+", Args: []string{idxVar, oneConst}, Span: toIRSpan(path, elem.Span)})
			loopBody = append(loopBody, ir.Instruction{Op: ir.OpAssign, Type: ir.TypeNumber, Result: idxVar, Args: []string{nextIdx}, Span: toIRSpan(path, elem.Span)})

			function.Body = append(function.Body, ir.Instruction{
				Op:   ir.OpWhile,
				Type: ir.TypeVoid,
				Cond: condBody,
				Args: []string{condVar},
				Body: loopBody,
				Span: toIRSpan(path, elem.Span),
			})
		} else {
			itemVal, _, err := lowerExpression(path, elem, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			pushRes := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: pushRes, Callee: "__array.push", Args: []string{result, itemVal}, Span: toIRSpan(path, elem.Span)})
		}
	}
	return result, arrType, nil
}

func lowerTemplateLiteral(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if len(expression.Arguments) == 0 {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: result, Value: "", Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeString, nil
	}
	var currentResult string
	for index, arg := range expression.Arguments {
		val, valType, err := lowerExpression(path, arg, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		strVal := val
		if valType != ir.TypeString {
			if arg != nil && arg.InferredType != "" && toIRType(arg.InferredType) == ir.TypeString {
				strTemp := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCheckedCast,
					Type:   ir.TypeString,
					Result: strTemp,
					Args:   []string{val},
					Span:   toIRSpan(path, arg.Span),
				})
				strVal = strTemp
			} else if strings.HasPrefix(string(valType), "object:") {
				className := strings.TrimPrefix(string(valType), "object:")
				if converted, ok := lowerClassToString(path, val, valType, arg.Span, function, counter, signatures); ok {
					strVal = converted
				} else if len(className) <= 2 || className == "T" || className == "K" || className == "V" || className == "U" || className == "A" || className == "B" {
					if arg != nil && (arg.InferredType == "number" || arg.InferredType == "bigint") {
						strTemp := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: strTemp, Callee: "__string.fromNumber", Args: []string{val}, Span: toIRSpan(path, arg.Span)})
						strVal = strTemp
					} else {
						strVal = val
					}
				} else {
					strTemp := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: strTemp, Callee: "__string.fromObject", Args: []string{val}, Span: toIRSpan(path, arg.Span)})
					strVal = strTemp
				}
			} else {
				strTemp := nextTemp(counter)
				callee := "__string.fromNumber"
				if valType == ir.TypeBool {
					callee = "__string.fromBool"
				} else if valType == ir.TypeBigInt {
					callee = "__string.fromBigInt"
				} else if valType == ir.TypeUnknown || valType == ir.TypeVoid {
					callee = "__string.fromUnknown"
				} else if valType == ir.TypeObject || strings.HasPrefix(string(valType), "object:") {
					callee = "__string.fromObject"
				} else if valType != ir.TypeNumber {
					return "", "", fmt.Errorf("template expression does not support %s in interpolation", valType)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: strTemp, Callee: callee, Args: []string{val}, Span: toIRSpan(path, arg.Span)})
				strVal = strTemp
			}
		}
		if index == 0 {
			currentResult = strVal
		} else {
			concatTemp := nextTemp(counter)
			if index == len(expression.Arguments)-1 && result != "" {
				concatTemp = result
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeString, Result: concatTemp, Operator: "+", Args: []string{currentResult, strVal}, Span: toIRSpan(path, expression.Span)})
			currentResult = concatTemp
		}
	}
	return currentResult, ir.TypeString, nil
}
