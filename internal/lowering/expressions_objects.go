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
	if (objectType == ir.TypeString || objectType == ir.TypeNumberArray || objectType == ir.TypeStringArray) && expression.Text == "length" {
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

func lowerNewExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	className := callName(expression.Left)
	if className == "RegExp" {
		ensureRegExpShape(shapes)
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
	for _, field := range shape.Fields {
		if strings.HasSuffix(string(field.Type), "[]") || field.Type == ir.TypeNumberArray || field.Type == ir.TypeStringArray || field.Type == ir.TypeBoolArray || field.Type == ir.TypeBigIntArray {
			arrTemp := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: field.Type, Result: arrTemp, Span: field.Span})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: field.Name, FieldIndex: fieldIndex(shape, field.Name), Args: []string{result, arrTemp}, Span: field.Span})
		} else if field.Value != "" {
			initializer := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: field.Type, Result: initializer, Value: field.Value, Span: field.Span})
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
