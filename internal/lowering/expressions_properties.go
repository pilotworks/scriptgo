package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerPropertyExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if expression.Left != nil && expression.Left.Kind == "identifier" {
		// 1. Check built-in global constants (e.g. Math.PI, Number.MAX_VALUE)
		propKey := expression.Left.Text + "." + expression.Text
		if global, ok := builtinGlobal(propKey); ok {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   global.Type,
				Result: result,
				Value:  global.Value,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, global.Type, nil
		}

		if propKey == "process.argv" {
			if result == "" {
				result = nextTemp(counter)
			}
			elem := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeString, Result: elem, Value: "scriptgo", Span: toIRSpan(path, expression.Span),
			})
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpArray, Type: ir.TypeStringArray, Result: result, Args: []string{elem}, Span: toIRSpan(path, expression.Span),
			})
			return result, ir.TypeStringArray, nil
		}
		if propKey == "process.env" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpCall, Type: ir.TypeObject, Result: result, Callee: "__process.env_obj", Span: toIRSpan(path, expression.Span),
			})
			return result, ir.TypeObject, nil
		}

		// 2. Check static getters
		if getter, getterName, ok := findGetterInHierarchy(expression.Left.Text, expression.Text, signatures, classHierarchy); ok && getter.Parameters == nil {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   getter.ReturnType,
				Result: result,
				Callee: getterName,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, getter.ReturnType, nil
		}

		// 3. Check static fields in class hierarchy
		if meta, ok := classHierarchy[expression.Left.Text]; ok {
			if staticField, isStatic := meta.Statics[expression.Text]; isStatic {
				staticVar := expression.Left.Text + "_" + expression.Text
				if varType, exists := env[staticVar]; exists {
					return staticVar, varType, nil
				}
				if staticField.Initializer != nil && (staticField.Initializer.Kind == "number" || staticField.Initializer.Kind == "string" || staticField.Initializer.Kind == "bool") {
					typ := toIRType(staticField.Type)
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   typ,
						Result: result,
						Value:  staticField.Initializer.Text,
						Span:   toIRSpan(path, expression.Span),
					})
					return result, typ, nil
				}
			}
		}

		// 4. Check shape const fields (e.g. Enums)
		if shape, ok := shapes[expression.Left.Text]; ok {
			for _, field := range shape.Fields {
				if field.Name == expression.Text && field.Value != "" {
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpConst,
						Type:   field.Type,
						Result: result,
						Value:  field.Value,
						Span:   toIRSpan(path, expression.Span),
					})
					return result, field.Type, nil
				}
			}
		}

		if (expression.Left.Text == "process" || expression.Left.Text == "__scriptgo") && expression.Text == "argv" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeStringArray, Result: result, Callee: "__process.argv", Args: nil, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeStringArray, nil
		}
	}

	if expression.Left != nil && expression.Left.Kind == "property" && expression.Left.Left != nil && expression.Left.Left.Kind == "identifier" && expression.Left.Left.Text == "process" && expression.Left.Text == "env" {
		keyTemp := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: keyTemp, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__process.env", Args: []string{keyTemp}, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeString, nil
	}

	object, objectType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if objectType == ir.TypeArrayBuffer && expression.Text == "byteLength" {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: result,
			Callee: "__arraybuffer.byteLength",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeNumber, nil
	}

	if isMapType(objectType) {
		if expression.Text == "size" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: result,
				Callee: "__map.size",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
	}

	if isSetType(objectType) {
		if expression.Text == "size" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: result,
				Callee: "__set.size",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
	}

	if isTypedArrayType(objectType) {
		if expression.Text == "length" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: result,
				Callee: "__typedarray.length",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
		if expression.Text == "byteLength" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: result,
				Callee: "__typedarray.byteLength",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
		if expression.Text == "byteOffset" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: result,
				Callee: "__typedarray.byteOffset",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
		if expression.Text == "buffer" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeArrayBuffer,
				Result: result,
				Callee: "__typedarray.buffer",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeArrayBuffer, nil
		}
		if expression.Text == "BYTES_PER_ELEMENT" {
			elemSize := "1"
			switch objectType {
			case ir.TypeInt16Array, ir.TypeUint16Array:
				elemSize = "2"
			case ir.TypeInt32Array, ir.TypeUint32Array, ir.TypeFloat32Array:
				elemSize = "4"
			case ir.TypeFloat64Array, ir.TypeBigInt64Array, ir.TypeBigUint64Array:
				elemSize = "8"
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeNumber,
				Result: result,
				Value:  elemSize,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
	}

	if objectType == ir.TypeDataView {
		if expression.Text == "byteLength" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: result,
				Callee: "__dataview.byteLength",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
		if expression.Text == "byteOffset" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: result,
				Callee: "__dataview.byteOffset",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
		if expression.Text == "buffer" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeArrayBuffer,
				Result: result,
				Callee: "__dataview.buffer",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeArrayBuffer, nil
		}
	}

	if objectType == ir.TypeTextEncoder {
		if expression.Text == "encoding" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: result,
				Callee: "__text_encoder.encoding",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeString, nil
		}
	}

	if objectType == ir.TypeTextDecoder {
		switch expression.Text {
		case "encoding":
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: result,
				Callee: "__text_decoder.encoding",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeString, nil
		case "fatal":
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBool,
				Result: result,
				Callee: "__text_decoder.fatal",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, nil
		case "ignoreBOM":
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBool,
				Result: result,
				Callee: "__text_decoder.ignore_bom",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, nil
		}
	}

	if (objectType == ir.TypeString || objectType == ir.TypeNumberArray || objectType == ir.TypeStringArray || objectType == ir.TypeBoolArray || objectType == ir.TypeBigIntArray || strings.HasSuffix(string(objectType), "[]") || objectType == ir.Type("object:Array")) && expression.Text == "length" {
		if result == "" {
			result = nextTemp(counter)
		}
		callee := "__string.length"
		if objectType != ir.TypeString {
			callee = "__array.length"
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: callee, Args: []string{object}, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeNumber, nil
	}

	if objectType == ir.TypeSymbol && expression.Text == "description" {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeString,
			Result: result,
			Callee: "__symbol.description",
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeString, nil
	}

	if objectType == ir.Type("object:RegExp") {
		switch expression.Text {
		case "global", "ignoreCase", "multiline", "dotAll", "unicode", "sticky", "hasIndices", "unicodeSets":
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBool,
				Result: result,
				Callee: "__regexp." + expression.Text,
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeBool, nil
		case "lastIndex":
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeNumber,
				Result: result,
				Value:  "0",
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
	}

	className := strings.TrimPrefix(string(objectType), "object:")

	// Check instance getters
	if getter, getterName, ok := findGetterInHierarchy(className, expression.Text, signatures, classHierarchy); ok {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   getter.ReturnType,
			Result: result,
			Callee: getterName,
			Args:   []string{object},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, getter.ReturnType, nil
	}

	shape, ok := shapes[className]
	if !ok {
		if s, exists := anonymousShapes[className]; exists {
			shape = s
			shapes[className] = s
			ok = true
		} else if fields, ok2 := anonymousObjectFields(className); ok2 {
			shape = ir.ObjectShape{Name: className, Fields: fields}
			shapes[className] = shape
			ok = true
		}
	}
	if !ok {
		if className == "Record" || objectType == ir.TypeObject {
			propNameConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeString,
				Result: propNameConst,
				Value:  expression.Text,
				Span:   toIRSpan(path, expression.Span),
			})
			if result == "" {
				result = nextTemp(counter)
			}
			retType := ir.TypeNumber
			if expression.InferredType != "" {
				retType = toIRType(expression.InferredType)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   retType,
				Result: result,
				Callee: "__object.get_prop",
				Args:   []string{object, propNameConst},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, retType, nil
		}
		return "", "", fmt.Errorf("unknown object shape %q for property %q", className, expression.Text)
	}
	for _, field := range shape.Fields {
		if field.Name != expression.Text {
			continue
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:         ir.OpFieldGet,
			Type:       field.Type,
			Result:     result,
			Callee:     className,
			Field:      field.Name,
			FieldIndex: fieldIndex(shape, field.Name),
			Args:       []string{object},
			Span:       toIRSpan(path, expression.Span),
		})
		return result, field.Type, nil
	}
	if expression.Kind == "optional_property" || strings.HasPrefix(className, "__shape_") {
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeString,
			Result: result,
			Value:  "undefined",
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeString, nil
	}
	fieldNames := make([]string, 0, len(shape.Fields))
	for _, f := range shape.Fields {
		fieldNames = append(fieldNames, f.Name)
	}
	return "", "", fmt.Errorf("unknown field %q on object %q (fields: %v)", expression.Text, className, fieldNames)
}
