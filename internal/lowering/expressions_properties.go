package lowering

import (
	"fmt"
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerPropertyExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	// 1. Check built-in global constants (e.g. Math.PI, Number.MAX_VALUE, Symbol.iterator)
	if expression.Left != nil && expression.Left.Kind == "identifier" {
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
	}

	// 2. Check AST-level module and top-level constants (e.g. fs.constants.F_OK, os.EOL, buffer.constants.MAX_LENGTH)
	propertyPath := extractPropertyPath(expression)
	if len(propertyPath) >= 2 {
		firstIdent := propertyPath[0]
		if _, inEnv := env[firstIdent]; !inEnv {
			if val, valType, ok := resolveASTConstantPath(propertyPath); ok {
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   valType,
					Result: result,
					Value:  val,
					Span:   toIRSpan(path, expression.Span),
				})
				return result, valType, nil
			}
			if len(propertyPath) == 2 {
				if sig, ok := signatures[expression.Text]; ok {
					if result == "" {
						result = nextTemp(counter)
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:     ir.OpClosure,
						Type:   ir.TypeClosure,
						Result: result,
						Callee: sig.Name,
						Span:   toIRSpan(path, expression.Span),
					})
					return result, ir.TypeClosure, nil
				}
			}
		}
	}

	if expression.Left != nil && expression.Left.Kind == "identifier" {
		propKey := expression.Left.Text + "." + expression.Text

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

		// 2.5. Check function reflection (.length and .name)
		if fn, isFunc := signatures[expression.Left.Text]; isFunc {
			if expression.Text == "length" {
				arity := len(fn.Parameters)
				if defaults, hasDefaults := defaultParamsIndex[fn.Name]; hasDefaults {
					for i := 0; i < len(fn.Parameters); i++ {
						if _, isDefault := defaults[i]; isDefault {
							arity = i
							break
						}
					}
				}
				if restParamsIndex[fn.Name] && arity == len(fn.Parameters) && arity > 0 {
					arity = len(fn.Parameters) - 1
				}
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeNumber,
					Result: result,
					Value:  strconv.Itoa(arity),
					Span:   toIRSpan(path, expression.Span),
				})
				return result, ir.TypeNumber, nil
			}
			if expression.Text == "name" {
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeString,
					Result: result,
					Value:  fn.Name,
					Span:   toIRSpan(path, expression.Span),
				})
				return result, ir.TypeString, nil
			}
		}

		// 3. Check static fields in class hierarchy
		if meta, ok := classHierarchy[expression.Left.Text]; ok {
			if staticField, isStatic := meta.Statics[expression.Text]; isStatic {
				staticVar := expression.Left.Text + "_" + expression.Text
				typ := toIRType(staticField.Type)
				if typ == "" {
					typ = ir.TypeNumber
				}
				return staticVar, typ, nil
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
		if (expression.Left.Text == "process" || expression.Left.Text == "__scriptgo") && expression.Text == "version" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__process.version", Args: nil, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeString, nil
		}
		if (expression.Left.Text == "process" || expression.Left.Text == "__scriptgo") && expression.Text == "pid" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: "__process.pid", Args: nil, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeNumber, nil
		}
		if (expression.Left.Text == "process" || expression.Left.Text == "__scriptgo") && expression.Text == "ppid" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: "__process.ppid", Args: nil, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeNumber, nil
		}
		if (expression.Left.Text == "process" || expression.Left.Text == "__scriptgo") && expression.Text == "platform" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__os.platform", Args: nil, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeString, nil
		}
		if (expression.Left.Text == "process" || expression.Left.Text == "__scriptgo") && expression.Text == "arch" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__os.arch", Args: nil, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeString, nil
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

	if expression.Left != nil && expression.Left.Kind == "property" && expression.Left.Left != nil && expression.Left.Left.Kind == "identifier" && expression.Left.Left.Text == "process" && expression.Left.Text == "versions" {
		if expression.Text == "scriptgo" {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__process.version", Args: nil, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeString, nil
		}
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
		if expression.Text == "buffer" || expression.Text == "parent" {
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

	if (objectType == ir.TypeString || objectType == ir.TypeNumberArray || objectType == ir.TypeStringArray || objectType == ir.TypeBoolArray || objectType == ir.TypeBigIntArray || strings.HasSuffix(string(objectType), "[]") || objectType == ir.Type("object:Array") || isTupleShapeName(strings.TrimPrefix(string(objectType), "object:"))) && expression.Text == "length" {
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
				Op:         ir.OpFieldGet,
				Type:       ir.TypeNumber,
				Result:     result,
				Callee:     "RegExp",
				Field:      "lastIndex",
				FieldIndex: 2,
				Args:       []string{object},
				Span:       toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
	}

	if expression.Text == "length" {
		if objectType == ir.TypeString {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: result,
				Callee: "__string.length",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
		if strings.HasSuffix(string(objectType), "[]") || objectType == ir.TypeStringArray || objectType == ir.TypeNumberArray || objectType == ir.TypeBoolArray || objectType == ir.Type("symbol[]") {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: result,
				Callee: "__array.length",
				Args:   []string{object},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
		if strings.HasPrefix(string(objectType), "object:") {
			shapeName := strings.TrimPrefix(string(objectType), "object:")
			if s, ok := anonymousShapes[shapeName]; ok && len(s.Fields) > 0 && s.Fields[0].Name == "0" {
				if result == "" {
					result = nextTemp(counter)
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpConst,
					Type:   ir.TypeNumber,
					Result: result,
					Value:  strconv.Itoa(len(s.Fields)),
					Span:   toIRSpan(path, expression.Span),
				})
				return result, ir.TypeNumber, nil
			}
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

	if className == "string" || objectType == ir.TypeString {
		if expression.Text == "message" {
			if result != "" && result != object {
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCheckedCast,
					Type:   ir.TypeString,
					Result: result,
					Args:   []string{object},
					Span:   toIRSpan(path, expression.Span),
				})
				return result, ir.TypeString, nil
			}
			return object, ir.TypeString, nil
		}
	}

	if className == "this" || className == "" || className == "object" {
		if expression.Left != nil && expression.Left.Text != "" {
			if varStmt, inTop := topLevelVars[expression.Left.Text]; inTop && varStmt.Type != "" {
				className = varStmt.Type
			} else if t, inEnv := env[expression.Left.Text]; inEnv && string(t) != "" && string(t) != "this" && string(t) != "object" {
				className = strings.TrimPrefix(string(t), "object:")
			}
		}
		if className == "this" || className == "" {
			if t, inEnv := env["this"]; inEnv && string(t) != "this" && string(t) != "object:this" {
				className = strings.TrimPrefix(string(t), "object:")
			} else if function != nil && strings.Contains(function.Name, "_") && !strings.HasPrefix(function.Name, "__closure_") {
				className = strings.Split(function.Name, "_")[0]
			}
		}
		if className == "this" || className == "" {
			for sName, s := range shapes {
				if fieldIndex(s, expression.Text) >= 0 {
					className = sName
					break
				}
			}
		}
	}
	isUnionAlias := false
	if typeAliasesIndex != nil && typeAliasesIndex[className] != "" && strings.Contains(typeAliasesIndex[className], "|") {
		isUnionAlias = true
	}
	if !isUnionAlias && (className == "" || className == "Record" || strings.HasPrefix(className, "Record_") || strings.HasPrefix(className, "Record<") || strings.HasPrefix(className, "Partial_") || objectType == ir.TypeObject || objectType == ir.TypeUnknown) {
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
		retType := ir.TypeUnknown
		if expression.InferredType != "" {
			if inferred := toIRType(expression.InferredType); inferred != "" {
				retType = inferred
			}
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
	shape, ok := shapes[className]
	if !ok {
		if s, exists := registeredShapes[className]; exists {
			shape = s
			shapes[className] = s
			ok = true
		} else if s, exists := anonymousShapes[className]; exists {
			shape = s
			shapes[className] = s
			ok = true
		} else if aliased, hasAlias := typeAliasesIndex[className]; hasAlias {
			if fields, ok2 := anonymousObjectFields(aliased, nil); ok2 {
				shape = ir.ObjectShape{Name: className, Fields: fields}
				shapes[className] = shape
				ok = true
			}
		} else if fields, ok2 := anonymousObjectFields(className, nil); ok2 {
			shape = ir.ObjectShape{Name: className, Fields: fields}
			shapes[className] = shape
			ok = true
		} else if strings.Contains(className, "__") || strings.Contains(className, "_") {
			baseName := strings.Split(className, "__")[0]
			if !strings.Contains(className, "__") {
				baseName = strings.Split(className, "_")[0]
			}
			if s, exists := shapes[baseName]; exists {
				shape = s
				ok = true
			} else if s, exists := registeredShapes[baseName]; exists {
				shape = s
				ok = true
			} else if baseName == "Partial" || baseName == "Required" || baseName == "Readonly" {
				inner := strings.TrimPrefix(className, baseName+"__")
				if s, exists := shapes[inner]; exists {
					shape = s
					ok = true
				} else if s, exists := registeredShapes[inner]; exists {
					shape = s
					ok = true
				}
			}
		}
		if !ok {
			for name, s := range shapes {
				if (strings.HasPrefix(name, className+"__") || strings.HasPrefix(name, className+"_")) && fieldIndex(s, expression.Text) >= 0 {
					shape = s
					ok = true
					break
				}
			}
		}
	}
	if !ok {
		interStr := className
		if typeAliasesIndex != nil && typeAliasesIndex[className] != "" {
			interStr = typeAliasesIndex[className]
		}
		if strings.Contains(interStr, "&") || (typeAliasesIndex != nil && typeAliasesIndex[className] != "") {
			if fields, okF := resolveShapeFields(interStr, shapes); okF {
				shape = ir.ObjectShape{Name: className, Fields: fields}
				shapes[className] = shape
				shapes[interStr] = shape
				ok = true
			}
		}
	}
	if !ok {
		unionStr := className
		if typeAliasesIndex != nil && typeAliasesIndex[className] != "" {
			unionStr = typeAliasesIndex[className]
		}
		if strings.Contains(unionStr, "|") {
			for _, m := range splitTopLevelUnion(unionStr) {
				cleanM := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(m), "object:"))
				var s ir.ObjectShape
				var okS bool
				if s, okS = shapes[cleanM]; !okS {
					s, okS = registeredShapes[cleanM]
				}
				if !okS {
					if fields, okF := anonymousObjectFields(cleanM, nil); okF {
						name := anonymousShapeName(fields)
						s = ir.ObjectShape{Name: name, Fields: fields}
						shapes[name] = s
						okS = true
					}
				}
				if okS && fieldIndex(s, expression.Text) >= 0 {
					shape = s
					className = s.Name
					ok = true
					break
				}
			}
		}
	}
	if !ok {
		return "", "", fmt.Errorf("unknown object shape %q for property %q", className, expression.Text)
	}
	if fieldIndex(shape, expression.Text) < 0 {
		var matchedShape *ir.ObjectShape
		unionStr := className
		if typeAliasesIndex != nil && typeAliasesIndex[className] != "" {
			unionStr = typeAliasesIndex[className]
		}
		if strings.Contains(unionStr, "|") {
			for _, m := range splitTopLevelUnion(unionStr) {
				cleanM := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(m), "object:"))
				var s ir.ObjectShape
				var okS bool
				if s, okS = shapes[cleanM]; !okS {
					s, okS = registeredShapes[cleanM]
				}
				if !okS {
					if fields, okF := anonymousObjectFields(cleanM, nil); okF {
						name := anonymousShapeName(fields)
						s = ir.ObjectShape{Name: name, Fields: fields}
						shapes[name] = s
						okS = true
					}
				}
				if okS && fieldIndex(s, expression.Text) >= 0 {
					matchedShape = &s
					break
				}
			}
		}
		if matchedShape != nil {
			shape = *matchedShape
			className = shape.Name
		}
	}
	for _, field := range shape.Fields {
		if field.Name != expression.Text {
			continue
		}
		fType := field.Type
		if (fType == ir.TypeVoid || fType == "") && expression.InferredType != "" {
			inferred := toIRType(expression.InferredType)
			if inferred != "" && inferred != ir.TypeUnknown && inferred != ir.TypeVoid {
				fType = inferred
			}
		}
		if result == "" {
			result = nextTemp(counter)
		}
		if expression.Kind == "optional_property" {
			if fType == ir.TypeNumber || fType == ir.TypeBool {
				fType = ir.TypeUnknown
			}
			initVal := "undefined"
			if fType != ir.TypeString && fType != ir.TypeUnknown {
				initVal = "null"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   fType,
				Result: result,
				Value:  initVal,
				Span:   toIRSpan(path, expression.Span),
			})
			cond, err := coerceToBool(path, object, objectType, function, counter, expression.Span)
			if err == nil {
				thenFn := &ir.Function{}
				if fType == ir.TypeUnknown && field.Type != ir.TypeUnknown {
					rawRes := nextTemp(counter)
					thenFn.Body = append(thenFn.Body, ir.Instruction{
						Op:         ir.OpFieldGet,
						Type:       field.Type,
						Result:     rawRes,
						Callee:     className,
						Field:      field.Name,
						FieldIndex: fieldIndex(shape, field.Name),
						Args:       []string{object},
						Span:       toIRSpan(path, expression.Span),
					})
					thenFn.Body = append(thenFn.Body, ir.Instruction{
						Op:     ir.OpBoxUnknown,
						Type:   ir.TypeUnknown,
						Result: result,
						Args:   []string{rawRes},
						Span:   toIRSpan(path, expression.Span),
					})
				} else {
					thenFn.Body = append(thenFn.Body, ir.Instruction{
						Op:         ir.OpFieldGet,
						Type:       fType,
						Result:     result,
						Callee:     className,
						Field:      field.Name,
						FieldIndex: fieldIndex(shape, field.Name),
						Args:       []string{object},
						Span:       toIRSpan(path, expression.Span),
					})
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:   ir.OpIf,
					Type: ir.TypeVoid,
					Args: []string{cond},
					Then: thenFn.Body,
					Span: toIRSpan(path, expression.Span),
				})
				return result, fType, nil
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:         ir.OpFieldGet,
			Type:       fType,
			Result:     result,
			Callee:     className,
			Field:      field.Name,
			FieldIndex: fieldIndex(shape, field.Name),
			Args:       []string{object},
			Span:       toIRSpan(path, expression.Span),
		})
		return result, fType, nil
	}
	if strings.HasPrefix(className, "__shape_") {
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
		retType := ir.TypeUnknown
		if expression.InferredType != "" {
			if inferred := toIRType(expression.InferredType); inferred != "" {
				retType = inferred
			}
		}
		if expression.Kind == "optional_property" {
			initVal := "undefined"
			switch retType {
			case ir.TypeBool:
				initVal = "false"
			case ir.TypeNumber:
				initVal = "NaN"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   retType,
				Result: result,
				Value:  initVal,
				Span:   toIRSpan(path, expression.Span),
			})
			cond, err := coerceToBool(path, object, objectType, function, counter, expression.Span)
			if err == nil {
				thenFn := &ir.Function{}
				thenFn.Body = append(thenFn.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   retType,
					Result: result,
					Callee: "__object.get_prop",
					Args:   []string{object, propNameConst},
					Span:   toIRSpan(path, expression.Span),
				})
				function.Body = append(function.Body, ir.Instruction{
					Op:   ir.OpIf,
					Type: ir.TypeVoid,
					Args: []string{cond},
					Then: thenFn.Body,
					Span: toIRSpan(path, expression.Span),
				})
				return result, retType, nil
			}
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
	if expression.Kind == "optional_property" {
		if result == "" {
			result = nextTemp(counter)
		}
		retType := ir.TypeUnknown
		valStr := "undefined"
		if expression.InferredType != "" {
			inferred := toIRType(expression.InferredType)
			if inferred != "" && inferred != ir.TypeNumber && inferred != ir.TypeBool {
				retType = inferred
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   retType,
			Result: result,
			Value:  valStr,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, retType, nil
	}
	fieldNames := make([]string, 0, len(shape.Fields))
	for _, f := range shape.Fields {
		fieldNames = append(fieldNames, f.Name)
	}
	return "", "", fmt.Errorf("unknown field %q on object %q (fields: %v)", expression.Text, className, fieldNames)
}

func resolveShapeFields(name string, shapes map[string]ir.ObjectShape) ([]ir.Field, bool) {
	clean := strings.TrimSpace(strings.TrimPrefix(name, "object:"))
	if s, ok := shapes[clean]; ok {
		return s.Fields, true
	}
	if s, ok := registeredShapes[clean]; ok {
		return s.Fields, true
	}
	if fields, ok := anonymousObjectFields(clean, nil); ok {
		return fields, true
	}
	if typeAliasesIndex != nil && typeAliasesIndex[clean] != "" {
		aliased := typeAliasesIndex[clean]
		if strings.Contains(aliased, "&") {
			var fields []ir.Field
			for _, sub := range strings.Split(aliased, "&") {
				if subFields, ok := resolveShapeFields(sub, shapes); ok {
					fields = append(fields, subFields...)
				}
			}
			if len(fields) > 0 {
				return fields, true
			}
		}
		return resolveShapeFields(aliased, shapes)
	}
	if strings.Contains(clean, "&") {
		var fields []ir.Field
		for _, sub := range strings.Split(clean, "&") {
			if subFields, ok := resolveShapeFields(sub, shapes); ok {
				fields = append(fields, subFields...)
			}
		}
		if len(fields) > 0 {
			return fields, true
		}
	}
	return nil, false
}
