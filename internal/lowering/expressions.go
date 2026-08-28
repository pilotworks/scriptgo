package lowering

import (
	"fmt"
	"maps"
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
	case "null":
		typ := ir.Type("ptr")
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: "null", Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "undefined":
		typ := ir.TypeVoid
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: "undefined", Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "array":
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
					}
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: arrType, Result: result, Args: arguments, Span: toIRSpan(path, expression.Span)})
					return result, arrType, nil
				} else if _, ok := tupleFields(trimmed); ok {
					inferredTuple = true
				}
			}
			if isHomogeneous && !inferredTuple {
				var arrType ir.Type = ir.TypeNumberArray
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
		var arrType ir.Type = ir.TypeNumberArray
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
		if expression.Kind == "optional_index" {
			if result == "" {
				result = nextTemp(counter)
			}
			retType := ir.TypeUnknown
			if expression.InferredType != "" {
				if t := toIRType(expression.InferredType); t != "" {
					retType = t
				}
			}
			if retType == ir.TypeNumber || retType == ir.TypeBool {
				retType = ir.TypeUnknown
			}
			initVal := "undefined"
			if retType != ir.TypeString && retType != ir.TypeUnknown {
				initVal = "null"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   retType,
				Result: result,
				Value:  initVal,
				Span:   toIRSpan(path, expression.Span),
			})

			cond, err := coerceToBool(path, array, arrayType, function, counter, expression.Span)
			if err != nil {
				return "", "", err
			}

			thenFn := &ir.Function{}
			if array != "" && arrayType != "" {
				env[array] = arrayType
			}
			elemType := ""
			if expression.InferredType != "" {
				parts := strings.Split(expression.InferredType, "|")
				for _, p := range parts {
					trimmed := strings.TrimSpace(p)
					if trimmed != "undefined" && trimmed != "null" && trimmed != "void" && trimmed != "" {
						elemType = trimmed
						break
					}
				}
			}
			standardIndex := &typescriptgo.SyntaxExpression{
				Span:         expression.Span,
				Kind:         "index",
				Left:         &typescriptgo.SyntaxExpression{Span: expression.Span, Kind: "identifier", Text: array, InferredType: string(arrayType)},
				Right:        expression.Right,
				InferredType: elemType,
			}
			idxRes, idxType, err := lowerExpression(path, standardIndex, "", thenFn, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			if retType == ir.TypeUnknown && idxType != ir.TypeUnknown {
				thenFn.Body = append(thenFn.Body, ir.Instruction{
					Op:     ir.OpBoxUnknown,
					Type:   ir.TypeUnknown,
					Result: result,
					Args:   []string{idxRes},
					Span:   toIRSpan(path, expression.Span),
				})
			} else {
				thenFn.Body = append(thenFn.Body, ir.Instruction{
					Op:     ir.OpAssign,
					Type:   idxType,
					Result: result,
					Args:   []string{idxRes},
					Span:   toIRSpan(path, expression.Span),
				})
			}

			function.Body = append(function.Body, ir.Instruction{
				Op:   ir.OpIf,
				Type: ir.TypeVoid,
				Args: []string{cond},
				Then: thenFn.Body,
				Span: toIRSpan(path, expression.Span),
			})
			return result, retType, nil
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
		if after, ok := strings.CutPrefix(string(arrayType), "object:"); ok && !strings.HasSuffix(string(arrayType), "[]") {
			shapeName := after
			shape, ok := shapes[shapeName]
			if !ok {
				if s, exists := anonymousShapes[shapeName]; exists {
					shape = s
					ok = true
				} else if s, exists := registeredShapes[shapeName]; exists {
					shape = s
					ok = true
				} else if expression.Left != nil && expression.Left.Kind == "identifier" {
					if topVar, exists := topLevelVars[expression.Left.Text]; exists && topVar.Expression != nil && topVar.Expression.Kind == "object_literal" {
						var objFields []ir.Field
						for _, p := range topVar.Expression.Arguments {
							objFields = append(objFields, ir.Field{Name: p.Text, Type: toIRType(p.InferredType)})
						}
						shape = ir.ObjectShape{Name: shapeName, Fields: objFields}
						ok = true
					}
				}
			}
			if ok {
				if expression.Right != nil && (expression.Right.Kind == "number" || expression.Right.Kind == "literal") {
					fieldIdx, err := strconv.Atoi(expression.Right.Text)
					if err == nil && fieldIdx >= 0 && fieldIdx < len(shape.Fields) {
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
			if indexType == ir.TypeString || strings.HasPrefix(string(arrayType), "object:") || arrayType == ir.TypeObject || arrayType == ir.TypeUnknown {
				if result == "" {
					result = nextTemp(counter)
				}
				retType := ir.TypeString
				if expression.InferredType != "" {
					retType = toIRType(expression.InferredType)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   retType,
					Result: result,
					Callee: "__object.get",
					Args:   []string{array, index},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, retType, nil
			}
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
		} else if arrayType == ir.TypeUnknown {
			elemType = ir.TypeUnknown
		} else if before, ok := strings.CutSuffix(string(arrayType), "[]"); ok {
			elemName := before
			if elemName == "boolean" {
				elemType = ir.TypeBool
			} else {
				elemType = ir.Type(elemName)
			}
		} else if after, ok := strings.CutPrefix(string(arrayType), "object:"); ok {
			shapeName := after
			idx := 0
			if expression.Right != nil {
				if n, err := strconv.Atoi(expression.Right.Text); err == nil {
					idx = n
				}
			}
			fType := ir.TypeUnknown
			if shape, exists := shapes[shapeName]; exists {
				if idx < len(shape.Fields) {
					fType = shape.Fields[idx].Type
				} else if len(shape.Fields) > 0 {
					fType = shape.Fields[len(shape.Fields)-1].Type
				}
			} else if shape, exists := anonymousShapes[shapeName]; exists {
				if idx < len(shape.Fields) {
					fType = shape.Fields[idx].Type
				} else if len(shape.Fields) > 0 {
					fType = shape.Fields[len(shape.Fields)-1].Type
				}
			} else {
				parts := strings.Split(shapeName, "_")
				if idx < len(parts) {
					fType = toIRType(parts[idx])
				} else if len(parts) > 0 {
					fType = toIRType(parts[len(parts)-1])
				}
			}
			if fType == "" {
				fType = ir.TypeUnknown
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:         ir.OpFieldGet,
				Type:       fType,
				Result:     result,
				Callee:     shapeName,
				Field:      strconv.Itoa(idx),
				FieldIndex: idx,
				Args:       []string{array},
				Span:       toIRSpan(path, expression.Span),
			})
			return result, fType, nil
		} else {
			return "", "", fmt.Errorf("array indexing requires an array, got %s", arrayType)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpIndex, Type: elemType, Result: result, Args: []string{array, index}, Span: toIRSpan(path, expression.Span)})
		return result, elemType, nil
	case "this":
		typ, ok := env["this"]
		if !ok {
			typ = ir.TypeObject
		}
		return "this", typ, nil
	case "identifier":
		if expression.Text == "undefined" {
			if _, inEnv := env["undefined"]; !inEnv {
				typ := ir.TypeVoid
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: "undefined", Span: toIRSpan(path, expression.Span)})
				return result, typ, nil
			}
		}
		rawEnvType, ok := env[expression.Text]
		if ok {
			typ := rawEnvType
			storageType := env["__storage_type."+expression.Text]
			if storageType == "" {
				if topVar, ok := topLevelVars[expression.Text]; ok {
					vType := topVar.Type
					if vType == "" && topVar.InferredType != "" {
						vType = topVar.InferredType
					}
					storageType = toIRType(vType)
				}
			}
			origParamType, isParam := env["__param."+expression.Text]
			isOriginallyUnknown := (rawEnvType == "" || (isParam && (origParamType == ir.TypeUnknown || origParamType == "" || strings.Contains(string(origParamType), "|"))))
			if isOriginallyUnknown {
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
				case "unknown[]":
					typ = ir.TypeUnknownArray
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
				case "ArrayBuffer", "SharedArrayBuffer":
					typ = ir.TypeArrayBuffer
				case "Buffer":
					typ = ir.TypeBuffer
				default:
					if expression.InferredType != "" && !strings.Contains(expression.InferredType, "|") && !strings.HasPrefix(string(typ), "object:Generator_") && expression.InferredType != "null" && expression.InferredType != "undefined" {
						inferredIR := toIRType(expression.InferredType)
						if inferredIR != "" && inferredIR != ir.TypeUnknown && inferredIR != ir.TypeVoid && inferredIR != ir.TypePointer {
							typ = inferredIR
						} else if _, isShape := shapes[expression.InferredType]; isShape {
							typ = ir.Type("object:" + expression.InferredType)
						}
					}
				}
			}
			if (storageType == ir.TypeUnknown && rawEnvType != ir.TypeUnknown && rawEnvType != "") || (isOriginallyUnknown && typ != ir.TypeUnknown && typ != "") {
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCheckedCast,
					Type:   typ,
					Result: result,
					Args:   []string{expression.Text},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, typ, nil
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
				if strings.HasPrefix(string(typ), "object:") || strings.HasSuffix(string(typ), "[]") {
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
			calleeName := ensureFunctionClosureTrampoline(path, sig, signatures)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpClosure,
				Type:   ir.TypeClosure,
				Result: result,
				Callee: calleeName,
				Args:   nil,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeClosure, nil
		}
		if topVar, ok := topLevelVars[expression.Text]; ok && topVar.Expression != nil && !inProgressVars[expression.Text] {
			isPrimitiveConst := topVar.VarDeclKind == "const" && (topVar.Expression.Kind == "number" || topVar.Expression.Kind == "string" || topVar.Expression.Kind == "bool" || topVar.Expression.Kind == "literal" || topVar.Expression.Kind == "null" || topVar.Expression.Kind == "undefined")
			if !isPrimitiveConst || function.Name != "main" {
				varTyp := toIRType(topVar.Type)
				if varTyp == "" {
					varTyp = toIRType(topVar.InferredType)
				}
				if varTyp == "" {
					varTyp = ir.TypeNumber
				}
				return expression.Text, varTyp, nil
			}
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
				if arg != nil && arg.InferredType != "" && toIRType(arg.InferredType) == ir.TypeString {
					strVal = val
				} else if strings.HasPrefix(string(valType), "object:") {
					className := strings.TrimPrefix(string(valType), "object:")
					if method, mangled, found := findMethodInHierarchy(className, "toString", signatures, classHierarchy); found && (method.ReturnType == ir.TypeString || method.ReturnType == "") {
						strTemp := nextTemp(counter)
						function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: strTemp, Callee: mangled, Args: []string{val}, Span: toIRSpan(path, arg.Span)})
						strVal = strTemp
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
	case "conditional":
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
				if trueType == ir.TypeString {
					zeroVal = "null"
				}
				elseFn.Body = append(elseFn.Body, ir.Instruction{Op: ir.OpConst, Type: trueType, Result: whenFalse, Value: zeroVal, Span: toIRSpan(path, expression.WhenFalse.Span)})
			} else if expression.WhenTrue != nil && (expression.WhenTrue.Kind == "null" || expression.WhenTrue.Kind == "undefined") {
				trueType = falseType
				whenTrue = nextTemp(counter)
				zeroVal := "0"
				if falseType == ir.TypeString {
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
	for {
		value := "t" + strconv.Itoa(*counter)
		*counter++
		if _, exists := topLevelVars[value]; exists {
			continue
		}
		return value
	}
}

func isComparison(operator string) bool {
	return operator == "==" || operator == "===" || operator == "!=" || operator == "!==" || operator == "<" || operator == "<=" || operator == ">" || operator == ">="
}

func arrayElementType(arrType ir.Type) ir.Type {
	switch arrType {
	case ir.TypeNumberArray:
		return ir.TypeNumber
	case ir.TypeStringArray:
		return ir.TypeString
	case ir.TypeBoolArray:
		return ir.TypeBool
	case ir.TypeBigIntArray:
		return ir.TypeBigInt
	case ir.TypeSymbolArray:
		return ir.TypeSymbol
	default:
		s := string(arrType)
		if strings.HasSuffix(s, "[]") {
			return toIRType(strings.TrimSuffix(s, "[]"))
		}
		return ir.TypeNumber
	}
}
