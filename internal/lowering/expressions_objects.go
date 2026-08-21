package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
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

	if objectType == ir.TypeUint8Array || objectType == ir.TypeInt32Array || objectType == ir.TypeFloat64Array {
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
			if result == "" {
				result = nextTemp(counter)
			}
			val := "1"
			if objectType == ir.TypeInt32Array {
				val = "4"
			} else if objectType == ir.TypeFloat64Array {
				val = "8"
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeNumber,
				Result: result,
				Value:  val,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, nil
		}
	}

	if (objectType == ir.TypeString || objectType == ir.TypeNumberArray || objectType == ir.TypeStringArray || objectType == ir.TypeBoolArray || objectType == ir.TypeBigIntArray || strings.HasSuffix(string(objectType), "[]")) && expression.Text == "length" {
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
		return "", "", fmt.Errorf("unknown object shape %q", className)
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
	return "", "", fmt.Errorf("unknown field %q on object %q", expression.Text, className)
}

func lowerObjectLiteralExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if len(expression.Arguments) == 0 {
		return "", "", fmt.Errorf("empty object literal needs explicit type or shape")
	}
	var fields []ir.Field
	var propValues []string
	for _, prop := range expression.Arguments {
		val, valType, err := lowerExpression(path, prop.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		fields = append(fields, ir.Field{
			Name: prop.Text,
			Type: valType,
			Span: toIRSpan(path, prop.Span),
		})
		propValues = append(propValues, val)
	}
	shapeName := anonymousShapeName(fields)
	if expression.InferredType != "" {
		cleanInf := strings.TrimPrefix(expression.InferredType, "object:")
		if _, ok := shapes[cleanInf]; ok {
			shapeName = cleanInf
		}
	}
	if shapeName == anonymousShapeName(fields) {
		for name, s := range shapes {
			if !strings.HasPrefix(name, "__shape_") && len(s.Fields) == len(fields) {
				match := true
				for _, f := range fields {
					idx := fieldIndex(s, f.Name)
					if idx == -1 {
						match = false
						break
					}
				}
				if match {
					shapeName = name
					break
				}
			}
		}
	}
	if _, ok := shapes[shapeName]; !ok {
		shapes[shapeName] = ir.ObjectShape{
			Name:   shapeName,
			Span:   toIRSpan(path, expression.Span),
			Fields: fields,
		}
	}
	targetShape := shapes[shapeName]
	if result == "" {
		result = nextTemp(counter)
	}
	objType := ir.Type("object:" + shapeName)
	function.Body = append(function.Body, ir.Instruction{
		Op:         ir.OpObjectNew,
		Type:       objType,
		Result:     result,
		Callee:     shapeName,
		FieldCount: len(targetShape.Fields),
		Span:       toIRSpan(path, expression.Span),
	})
	propMap := map[string]string{}
	for i, f := range fields {
		propMap[f.Name] = propValues[i]
	}
	for i, field := range targetShape.Fields {
		if val, exists := propMap[field.Name]; exists {
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
	}
	return result, objType, nil
}

func anonymousShapeName(fields []ir.Field) string {
	var parts []string
	for _, f := range fields {
		cleanType := strings.ReplaceAll(string(f.Type), ":", "_")
		cleanType = strings.ReplaceAll(cleanType, "[]", "_arr")
		parts = append(parts, fmt.Sprintf("%s_%s", f.Name, cleanType))
	}
	return "__shape_" + strings.Join(parts, "_")
}

func ensureDateShape(shapes map[string]ir.ObjectShape) {
	if _, ok := shapes["Date"]; !ok {
		shapes["Date"] = ir.ObjectShape{
			Name: "Date",
			Fields: []ir.Field{
				{Name: "time", Type: ir.TypeNumber},
			},
		}
	}
}

func ensureErrorShape(shapes map[string]ir.ObjectShape, name string) {
	if name == "" {
		name = "Error"
	}
	if _, ok := shapes[name]; !ok {
		shapes[name] = ir.ObjectShape{
			Name: name,
			Fields: []ir.Field{
				{Name: "message", Type: ir.TypeString},
				{Name: "name", Type: ir.TypeString},
			},
		}
	}
}

func lowerNewExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	className := callName(expression.Left)
	if className == "RegExp" {
		ensureRegExpShape(shapes)
	}
	if className == "Date" {
		ensureDateShape(shapes)
	}
	if className == "Error" || className == "TypeError" || className == "RangeError" || className == "SyntaxError" {
		ensureErrorShape(shapes, className)
	}
	if className == "ArrayBuffer" {
		byteLenVal := ""
		if len(expression.Arguments) > 0 {
			v, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			byteLenVal = v
		} else {
			zeroConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, expression.Span),
			})
			byteLenVal = zeroConst
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeArrayBuffer,
			Result: result,
			Callee: "__arraybuffer.new",
			Args:   []string{byteLenVal},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeArrayBuffer, nil
	}

	if className == "Uint8Array" || className == "Int32Array" || className == "Float64Array" {
		targetType := ir.Type(className)
		if len(expression.Arguments) == 0 {
			zeroConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, expression.Span),
			})
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   targetType,
				Result: result,
				Callee: "__typedarray.new_length",
				Value:  className,
				Args:   []string{zeroConst},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, targetType, nil
		}
		arg0Val, arg0Type, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if result == "" {
			result = nextTemp(counter)
		}
		if arg0Type == ir.TypeNumber {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   targetType,
				Result: result,
				Callee: "__typedarray.new_length",
				Value:  className,
				Args:   []string{arg0Val},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, targetType, nil
		}
		if arg0Type == ir.TypeArrayBuffer {
			byteOffsetVal := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeNumber, Result: byteOffsetVal, Value: "0", Span: toIRSpan(path, expression.Span),
			})
			lengthVal := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeNumber, Result: lengthVal, Value: "0", Span: toIRSpan(path, expression.Span),
			})
			if len(expression.Arguments) > 1 {
				bo, _, err := lowerExpression(path, expression.Arguments[1], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				byteOffsetVal = bo
			}
			if len(expression.Arguments) > 2 {
				l, _, err := lowerExpression(path, expression.Arguments[2], "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				lengthVal = l
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   targetType,
				Result: result,
				Callee: "__typedarray.new_buffer",
				Value:  className,
				Args:   []string{arg0Val, byteOffsetVal, lengthVal},
				Span:   toIRSpan(path, expression.Span),
			})
			return result, targetType, nil
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   targetType,
			Result: result,
			Callee: "__typedarray.new_array",
			Value:  className,
			Args:   []string{arg0Val},
			Span:   toIRSpan(path, expression.Span),
		})
		return result, targetType, nil
	}

	shape, ok := shapes[className]
	if !ok {
		return "", "", fmt.Errorf("unknown class %q", className)
	}
	if result == "" {
		result = nextTemp(counter)
	}
	objType := ir.Type("object:" + className)
	tag := getHierarchyTag(className, classHierarchy)
	function.Body = append(function.Body, ir.Instruction{
		Op:         ir.OpObjectNew,
		Type:       objType,
		Result:     result,
		Callee:     className,
		Value:      tag,
		FieldCount: len(shape.Fields),
		Span:       toIRSpan(path, expression.Span),
	})
	if className == "Date" {
		timeVal := nextTemp(counter)
		if len(expression.Arguments) == 0 {
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpCall, Type: ir.TypeNumber, Result: timeVal, Callee: "__date.now", Span: toIRSpan(path, expression.Span),
			})
		} else {
			argVal, argType, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			if argType == ir.TypeNumber {
				timeVal = argVal
			} else if argType == ir.TypeString {
				function.Body = append(function.Body, ir.Instruction{
					Op: ir.OpCall, Type: ir.TypeNumber, Result: timeVal, Callee: "__date.parse", Args: []string{argVal}, Span: toIRSpan(path, expression.Span),
				})
			} else {
				timeVal = argVal
			}
		}
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: "Date", Field: "time", FieldIndex: 0, Args: []string{result, timeVal}, Span: toIRSpan(path, expression.Span),
		})
		return result, objType, nil
	}
	if className == "Error" || className == "TypeError" || className == "RangeError" || className == "SyntaxError" {
		msgVal := nextTemp(counter)
		if len(expression.Arguments) > 0 {
			mv, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			msgVal = mv
		} else {
			function.Body = append(function.Body, ir.Instruction{
				Op: ir.OpConst, Type: ir.TypeString, Result: msgVal, Value: "", Span: toIRSpan(path, expression.Span),
			})
		}
		nameVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpConst, Type: ir.TypeString, Result: nameVal, Value: className, Span: toIRSpan(path, expression.Span),
		})
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "message", FieldIndex: 0, Args: []string{result, msgVal}, Span: toIRSpan(path, expression.Span),
		})
		function.Body = append(function.Body, ir.Instruction{
			Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: "name", FieldIndex: 1, Args: []string{result, nameVal}, Span: toIRSpan(path, expression.Span),
		})
		return result, objType, nil
	}
	for _, field := range shape.Fields {
		if strings.HasSuffix(string(field.Type), "[]") || field.Type == ir.TypeNumberArray || field.Type == ir.TypeStringArray || field.Type == ir.TypeBoolArray || field.Type == ir.TypeBigIntArray {
			arrTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: field.Type, Result: arrTemp, Span: field.Span})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: field.Name, FieldIndex: fieldIndex(shape, field.Name), Args: []string{result, arrTemp}, Span: field.Span})
		} else {
			defVal := field.Value
			if defVal == "" {
				switch field.Type {
				case ir.TypeNumber:
					defVal = "0"
				case ir.TypeBool:
					defVal = "false"
				case ir.TypeBigInt:
					defVal = "0"
				}
			}
			initializer := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: field.Type, Result: initializer, Value: defVal, Span: field.Span})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: field.Name, FieldIndex: fieldIndex(shape, field.Name), Args: []string{result, initializer}, Span: field.Span})
		}
	}

	// Call constructor if present
	if ctor, ctorName, found := findConstructorInHierarchy(className, signatures, classHierarchy); found {
		args := []string{result}
		for _, arg := range expression.Arguments {
			argVal, _, err := lowerExpression(path, arg, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			args = append(args, argVal)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ctor.ReturnType,
			Callee: ctorName,
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
	} else {
		// Fallback for classes without constructors: positional field assignment if arguments are passed
		for i, argument := range expression.Arguments {
			if i < len(shape.Fields) {
				argVal, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				field := shape.Fields[i]
				function.Body = append(function.Body, ir.Instruction{
					Op:         ir.OpFieldSet,
					Type:       ir.TypeVoid,
					Callee:     className,
					Field:      field.Name,
					FieldIndex: i,
					Args:       []string{result, argVal},
					Span:       toIRSpan(path, argument.Span),
				})
			}
		}
	}
	return result, objType, nil
}
