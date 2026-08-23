package lowering

import (
	"fmt"
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	switch expression.Kind {
	case "number":
		typ := ir.TypeNumber
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "bigint":
		typ := ir.TypeBigInt
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "regex":
		return lowerRegexLiteral(path, expression, result, function, env, counter, shapes, signatures)
	case "string":
		typ := ir.TypeString
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "bool":
		typ := ir.TypeBool
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "null", "undefined":
		typ := ir.TypeString
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Kind, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "array":
		if len(expression.Arguments) == 0 {
			arrType := toIRType(expression.InferredType)
			if !strings.HasSuffix(string(arrType), "[]") && arrType != ir.TypeNumberArray && arrType != ir.TypeStringArray && arrType != ir.TypeBoolArray && arrType != ir.TypeBigIntArray && arrType != ir.TypeSymbolArray {
				if varType, ok := env[result]; ok && strings.HasSuffix(string(varType), "[]") {
					arrType = varType
				} else {
					arrType = ir.TypeNumberArray
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
			isTuple := false
			if expression.InferredType != "" {
				trimmed := strings.TrimSpace(expression.InferredType)
				if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") && !strings.HasSuffix(trimmed, "[]") {
					isTuple = true
				}
			}
			if isHomogeneous && !isTuple {
				var arrType ir.Type = ir.TypeNumberArray
				if len(types) > 0 && types[0] == ir.TypeString {
					arrType = ir.TypeStringArray
				} else if len(types) > 0 && types[0] != ir.TypeNumber {
					arrType = ir.Type(string(types[0]) + "[]")
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: arrType, Result: result, Args: arguments, Span: toIRSpan(path, expression.Span)})
				return result, arrType, nil
			}

			// Heterogeneous elements -> lower as anonymous tuple object
			var fields []ir.Field
			for i, typ := range types {
				fields = append(fields, ir.Field{
					Name: strconv.Itoa(i),
					Type: typ,
					Span: toIRSpan(path, expression.Arguments[i].Span),
				})
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
				function.Body = append(function.Body, ir.Instruction{
					Op:         ir.OpFieldSet,
					Type:       ir.TypeVoid,
					Callee:     shapeName,
					Field:      field.Name,
					FieldIndex: i,
					Args:       []string{result, arguments[i]},
					Span:       toIRSpan(path, expression.Span),
				})
			}
			return result, objType, nil
		}
		var arrType ir.Type = ir.TypeNumberArray
		if expression.InferredType != "" && expression.InferredType != "never[]" && expression.InferredType != "any[]" {
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
			}
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: arrType, Result: result, Args: nil, Span: toIRSpan(path, expression.Span)})
		for _, elem := range expression.Arguments {
			if elem.Kind == "spread" {
				spreadVal, _, err := lowerExpression(path, elem.Left, "", function, env, counter, shapes, signatures)
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
				itemType := ir.TypeNumber
				if arrType == ir.TypeStringArray {
					itemType = ir.TypeString
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
	case "index", "optional_index":
		if expression.Left != nil && expression.Left.Kind == "identifier" {
			if _, isVar := env[expression.Left.Text]; !isVar {
				if shape, isShape := shapes[expression.Left.Text]; isShape {
					// Static constant index (e.g. Color[0])
					if expression.Right != nil && expression.Right.Kind == "number" {
						for _, field := range shape.Fields {
							if field.Value == expression.Right.Text {
								if result == "" {
									result = nextTemp(counter)
								}
								function.Body = append(function.Body, ir.Instruction{
									Op:     ir.OpConst,
									Type:   ir.TypeString,
									Result: result,
									Value:  field.Name,
									Span:   toIRSpan(path, expression.Span),
								})
								return result, ir.TypeString, nil
							}
						}
					}
					// Dynamic variable index on enum (e.g. Color[val])
					idxVal, idxType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
					if err == nil && (idxType == ir.TypeNumber || idxType == ir.TypeString) {
						if result == "" {
							result = nextTemp(counter)
						}
						function.Body = append(function.Body, ir.Instruction{
							Op:     ir.OpConst,
							Type:   ir.TypeString,
							Result: result,
							Value:  "",
							Span:   toIRSpan(path, expression.Span),
						})
						for _, field := range shape.Fields {
							if field.Value != "" {
								targetVal := nextTemp(counter)
								function.Body = append(function.Body, ir.Instruction{
									Op:     ir.OpConst,
									Type:   field.Type,
									Result: targetVal,
									Value:  field.Value,
									Span:   toIRSpan(path, expression.Span),
								})
								cmpRes := nextTemp(counter)
								function.Body = append(function.Body, ir.Instruction{
									Op:       ir.OpCompare,
									Type:     ir.TypeBool,
									Operator: "==",
									Result:   cmpRes,
									Args:     []string{idxVal, targetVal},
									Span:     toIRSpan(path, expression.Span),
								})
								valStr := nextTemp(counter)
								function.Body = append(function.Body, ir.Instruction{
									Op:     ir.OpConst,
									Type:   ir.TypeString,
									Result: valStr,
									Value:  field.Name,
									Span:   toIRSpan(path, expression.Span),
								})
								selectRes := nextTemp(counter)
								function.Body = append(function.Body, ir.Instruction{
									Op:     ir.OpSelect,
									Type:   ir.TypeString,
									Result: selectRes,
									Args:   []string{cmpRes, valStr, result},
									Span:   toIRSpan(path, expression.Span),
								})
								result = selectRes
							}
						}
						return result, ir.TypeString, nil
					}
				}
			}
		}
		if expression.Left != nil && expression.Left.Kind == "property" && expression.Left.Left != nil && expression.Left.Left.Kind == "identifier" && expression.Left.Left.Text == "process" && expression.Left.Text == "env" {
			keyVal, keyType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
			if err != nil || keyType != ir.TypeString {
				return "", "", fmt.Errorf("process.env requires string index")
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__process.env", Args: []string{keyVal}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeString, nil
		}
		array, arrayType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if arrayType == ir.TypeString {
			index, indexType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			if indexType != ir.TypeNumber {
				return "", "", fmt.Errorf("string indexing requires number index")
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: result,
				Callee: "__string.charAt",
				Args:   []string{array, index},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeString, nil
		}
		if after, ok := strings.CutPrefix(string(arrayType), "object:"); ok {
			shapeName := after
			if shape, ok := shapes[shapeName]; ok {
				if expression.Right != nil && expression.Right.Kind == "number" {
					fieldIdx, _ := strconv.Atoi(expression.Right.Text)
					if fieldIdx >= 0 && fieldIdx < len(shape.Fields) {
						field := shape.Fields[fieldIdx]
						if result == "" {
							result = nextTemp(counter)
						}
						function.Body = append(function.Body, ir.Instruction{
							Op:         ir.OpFieldGet,
							Type:       field.Type,
							Result:     result,
							Callee:     shapeName,
							Field:      field.Name,
							FieldIndex: fieldIdx,
							Args:       []string{array},
							Span:       toIRSpan(path, expression.Span),
						})
						return result, field.Type, nil
					} else if fieldIdx >= len(shape.Fields) && len(shape.Fields) > 0 {
						lastField := shape.Fields[len(shape.Fields)-1]
						elemType := lastField.Type
						if strings.HasSuffix(string(elemType), "[]") {
							elemType = toIRType(strings.TrimSuffix(string(elemType), "[]"))
						}
						if result == "" {
							result = nextTemp(counter)
						}
						function.Body = append(function.Body, ir.Instruction{
							Op:         ir.OpFieldGet,
							Type:       elemType,
							Result:     result,
							Callee:     shapeName,
							Field:      strconv.Itoa(fieldIdx),
							FieldIndex: fieldIdx,
							Args:       []string{array},
							Span:       toIRSpan(path, expression.Span),
						})
						return result, elemType, nil
					}
				}
				if expression.Right != nil && expression.Right.Kind == "string" {
					propName := expression.Right.Text
					for idx, field := range shape.Fields {
						if field.Name == propName {
							if result == "" {
								result = nextTemp(counter)
							}
							function.Body = append(function.Body, ir.Instruction{
								Op:         ir.OpFieldGet,
								Type:       field.Type,
								Result:     result,
								Callee:     shapeName,
								Field:      field.Name,
								FieldIndex: idx,
								Args:       []string{array},
								Span:       toIRSpan(path, expression.Span),
							})
							return result, field.Type, nil
						}
					}
				}
			}
		}
		index, indexType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if indexType != ir.TypeNumber {
			return "", "", fmt.Errorf("array indexing requires number index, got %s", indexType)
		}
		var elemType ir.Type
		if arrayType == ir.TypeBigInt64Array || arrayType == ir.TypeBigUint64Array {
			elemType = ir.TypeBigInt
		} else if isNumberTypedArray(arrayType) || arrayType == ir.TypeNumberArray {
			elemType = ir.TypeNumber
		} else if arrayType == ir.TypeStringArray {
			elemType = ir.TypeString
		} else if arrayType == ir.TypeBoolArray || arrayType == "boolean[]" || arrayType == "bool[]" {
			elemType = ir.TypeBool
		} else if before, ok := strings.CutSuffix(string(arrayType), "[]"); ok {
			elemName := before
			if elemName == "boolean" {
				elemType = ir.TypeBool
			} else {
				elemType = ir.Type(elemName)
			}
		} else {
			return "", "", fmt.Errorf("array indexing requires an array, got %s", arrayType)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpIndex, Type: elemType, Result: result, Args: []string{array, index}, Span: toIRSpan(path, expression.Span)})
		return result, elemType, nil
	case "identifier":
		typ, ok := env[expression.Text]
		if ok {
			switch expression.InferredType {
			case "number":
				typ = ir.TypeNumber
			case "bigint":
				typ = ir.TypeBigInt
			case "symbol":
				typ = ir.TypeSymbol
			case "string":
				typ = ir.TypeString
			case "bool", "boolean":
				typ = ir.TypeBool
			case "number[]":
				typ = ir.TypeNumberArray
			case "string[]":
				typ = ir.TypeStringArray
			case "Uint8Array":
				typ = ir.TypeUint8Array
			case "Int8Array":
				typ = ir.TypeInt8Array
			case "Uint8ClampedArray":
				typ = ir.TypeUint8ClampedArray
			case "Int16Array":
				typ = ir.TypeInt16Array
			case "Uint16Array":
				typ = ir.TypeUint16Array
			case "Int32Array":
				typ = ir.TypeInt32Array
			case "Uint32Array":
				typ = ir.TypeUint32Array
			case "Float32Array":
				typ = ir.TypeFloat32Array
			case "Float64Array":
				typ = ir.TypeFloat64Array
			case "BigInt64Array":
				typ = ir.TypeBigInt64Array
			case "BigUint64Array":
				typ = ir.TypeBigUint64Array
			case "DataView":
				typ = ir.TypeDataView
			case "ArrayBuffer":
				typ = ir.TypeArrayBuffer
			case "Buffer":
				typ = ir.TypeBuffer
			default:
				if expression.InferredType != "" {
					if _, isShape := shapes[expression.InferredType]; isShape {
						typ = ir.Type("object:" + expression.InferredType)
					} else if after, ok0 := strings.CutPrefix(expression.InferredType, "object:"); ok0 {
						shapeName := after
						if _, isShape := shapes[shapeName]; isShape {
							typ = ir.Type(expression.InferredType)
						}
					}
				}
			}
			if result != "" && result != expression.Text {
				if typ == ir.TypeNumber {
					zeroConst := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, expression.Span)})
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: typ, Result: result, Operator: "+", Args: []string{expression.Text, zeroConst}, Span: toIRSpan(path, expression.Span)})
					return result, typ, nil
				}
				if typ == ir.TypeString {
					emptyStr := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: emptyStr, Value: "", Span: toIRSpan(path, expression.Span)})
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: typ, Result: result, Operator: "+", Args: []string{expression.Text, emptyStr}, Span: toIRSpan(path, expression.Span)})
					return result, typ, nil
				}
				if typ == ir.TypeBool {
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: typ, Result: result, Operator: "||", Args: []string{expression.Text, expression.Text}, Span: toIRSpan(path, expression.Span)})
					return result, typ, nil
				}
				if strings.HasPrefix(string(typ), "object:") || strings.HasSuffix(string(typ), "[]") || typ == ir.TypeUnknown {
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpCheckedCast,
						Type:   typ,
						Result: result,
						Args:   []string{expression.Text},
						Span:   toIRSpan(path, expression.Span),
					})
					return result, typ, nil
				}
			}
			return expression.Text, typ, nil
		}
		if sig, ok := signatures[expression.Text]; ok {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpClosure,
				Type:   ir.TypeClosure,
				Result: result,
				Callee: sig.Name,
				Args:   nil,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeClosure, nil
		}
		if topVar, ok := topLevelVars[expression.Text]; ok && topVar.Expression != nil && !inProgressVars[expression.Text] {
			inProgressVars[expression.Text] = true
			res, typ, err := lowerExpression(path, topVar.Expression, result, function, env, counter, shapes, signatures)
			delete(inProgressVars, expression.Text)
			return res, typ, err
		}
		if inProgressVars[expression.Text] {
			return expression.Text, ir.TypeNumber, nil
		}
		global, ok := builtinGlobal(expression.Text)
		if !ok {
			return "", "", fmt.Errorf("unknown identifier %q", expression.Text)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: global.Type, Result: result, Value: global.Value, Span: toIRSpan(path, expression.Span)})
		return result, global.Type, nil
	case "unary":
		return lowerUnaryExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "postfix_unary":
		return lowerPostfixUnaryExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "as":
		targetIRType := toIRType(expression.Text)
		val, valType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if result == "" {
			result = nextTemp(counter)
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
	case "typeof":
		val, valType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			if expression.Left != nil && expression.Left.Kind == "identifier" && expression.Left.Text == "undefined" {
				valType = ir.TypeVoid
			} else {
				return "", "", err
			}
		}
		if result == "" {
			result = nextTemp(counter)
		}
		if valType == ir.TypeUnknown || valType == ir.TypeClosure || strings.HasPrefix(string(valType), "object:") || valType == "ptr" {
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
	case "binary":
		return lowerBinaryExpression(path, expression, result, function, env, counter, shapes, signatures)

	case "template":
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
				strTemp := nextTemp(counter)
				callee := "__string.fromNumber"
				if valType == ir.TypeBool {
					callee = "__string.fromBool"
				} else if valType == ir.TypeBigInt {
					callee = "__string.fromBigInt"
				} else if valType != ir.TypeNumber {
					return "", "", fmt.Errorf("template expression does not support %s in interpolation", valType)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: strTemp, Callee: callee, Args: []string{val}, Span: toIRSpan(path, arg.Span)})
				strVal = strTemp
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
	case "conditional":
		condition, conditionType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if conditionType != ir.TypeBool {
			if strings.HasPrefix(string(conditionType), "object:") || conditionType == "ptr" {
				nullConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: conditionType, Result: nullConst, Value: "0", Span: toIRSpan(path, expression.Span)})
				boolTemp := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: boolTemp, Operator: "!=", Args: []string{condition, nullConst}, Span: toIRSpan(path, expression.Span)})
				condition = boolTemp
			} else {
				return "", "", fmt.Errorf("conditional expression requires a bool condition")
			}
		}
		whenTrue, trueType, err := lowerExpression(path, expression.WhenTrue, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		whenFalse, falseType, err := lowerExpression(path, expression.WhenFalse, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if trueType != falseType {
			if expression.WhenFalse != nil && (expression.WhenFalse.Kind == "null" || expression.WhenFalse.Kind == "undefined") {
				whenFalse = nextTemp(counter)
				zeroVal := "0"
				if trueType == ir.TypeString {
					zeroVal = "null"
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: trueType, Result: whenFalse, Value: zeroVal, Span: toIRSpan(path, expression.WhenFalse.Span)})
			} else if expression.WhenTrue != nil && (expression.WhenTrue.Kind == "null" || expression.WhenTrue.Kind == "undefined") {
				trueType = falseType
				whenTrue = nextTemp(counter)
				zeroVal := "0"
				if falseType == ir.TypeString {
					zeroVal = "null"
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: falseType, Result: whenTrue, Value: zeroVal, Span: toIRSpan(path, expression.WhenTrue.Span)})
			} else {
				return "", "", fmt.Errorf("conditional branches must have the same type")
			}
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: trueType, Result: result, Args: []string{condition, whenTrue, whenFalse}, Span: toIRSpan(path, expression.Span)})
		return result, trueType, nil
	case "property", "optional_property":
		return lowerPropertyExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "object_literal":
		return lowerObjectLiteralExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "new":
		return lowerNewExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "call":
		return lowerCallExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "optional_call":
		return lowerOptionalCallExpression(path, expression, result, function, env, counter, shapes, signatures)
	case "tagged_template":
		return lowerTaggedTemplate(path, expression, result, function, env, counter, shapes, signatures)
	case "arrow_function":
		if expression.Function != nil {
			return lowerClosureExpression(path, expression.Function, result, function, env, counter, shapes, signatures)
		}
		return "", "", fmt.Errorf("arrow function expression missing body")
	case "await":
		val, typ, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if result == "" {
			result = nextTemp(counter)
		}
		retType := typ
		if strings.HasPrefix(string(typ), "object:Promise_") {
			inner := strings.TrimPrefix(string(typ), "object:Promise_")
			retType = toIRType(inner)
		} else if strings.HasPrefix(string(typ), "object:Promise<") && strings.HasSuffix(string(typ), ">") {
			inner := strings.TrimSuffix(strings.TrimPrefix(string(typ), "object:Promise<"), ">")
			retType = toIRType(inner)
		} else if expression.InferredType != "" && !strings.Contains(expression.InferredType, "Promise") {
			retType = toIRType(expression.InferredType)
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
	case "yield", "yield_star":
		if expression.Left != nil {
			return lowerExpression(path, expression.Left, result, function, env, counter, shapes, signatures)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: result, Value: "0", Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeNumber, nil
	default:
		return "", "", fmt.Errorf("unsupported expression %q", expression.Kind)
	}
}

func nextTemp(counter *int) string {
	value := "t" + strconv.Itoa(*counter)
	*counter++
	return value
}

func isComparison(operator string) bool {
	return operator == "==" || operator == "===" || operator == "!=" || operator == "!==" || operator == "<" || operator == "<=" || operator == ">" || operator == ">="
}
