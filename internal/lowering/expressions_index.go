package lowering

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerIndexExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
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
		standardIndex := &frontend.SyntaxExpression{
			Span:         expression.Span,
			Kind:         "index",
			Left:         &frontend.SyntaxExpression{Span: expression.Span, Kind: "identifier", Text: array, InferredType: string(arrayType)},
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
		if ok && isKeyedObjectShape(shapeName, shapes) && expression.Right != nil && expression.Right.Kind == "number" {
			// Positional reads are for tuples; a keyed object reads the
			// property named ToString(index).
			return lowerNumericKeyRead(path, expression, array, shapeName, result, function, env, counter, shapes, signatures)
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
			if indexType == ir.TypeSymbol {
				index, _ = propertyKeyValue(path, expression.Right.Span, index, indexType, function, counter)
			}
			retType := ir.TypeString
			if expression.InferredType != "" {
				retType = toIRType(expression.InferredType)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   retType,
				Result: result,
				Callee: "__object.get_prop",
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
		if isKeyedObjectShape(shapeName, shapes) {
			// A numeric index on a keyed object reads the property named
			// ToString(index), not the field at that position.
			return lowerNumericKeyRead(path, expression, array, shapeName, result, function, env, counter, shapes, signatures)
		}
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
}
