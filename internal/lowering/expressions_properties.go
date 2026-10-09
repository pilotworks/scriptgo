package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerPropertyExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	className := ""
	if expression.Left != nil && expression.Left.Kind == "identifier" {
		className = classIdentityForPath(path, expression.Left.Text)
	}
	if expression.Left != nil && expression.Left.Kind == "identifier" && expression.Left.Text == "Error" && expression.Text == "stackTraceLimit" {
		// A mutable runtime setting, not a constant.
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: "__error.stackTraceLimit", Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeNumber, nil
	}
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
		if value, valueType, handled, err := tryLowerConstantPropertyPath(path, expression, &result, function, env, counter, shapes, signatures, propertyPath); handled {
			return value, valueType, err
		}
	}

	if expression.Left != nil && expression.Left.Kind == "identifier" {
		if value, valueType, handled, err := tryLowerIdentifierReceiverProperty(path, expression, &result, function, counter, shapes, signatures, className); handled {
			return value, valueType, err
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

	// ArrayBufferView is a runtime union of typed arrays and DataView. Its shared
	// fields need a tag-aware ABI; concrete typed arrays can use their faster ABI.
	if isArrayBufferViewType(objectType) && objectType != ir.TypeDataView {
		if value, valueType, handled, err := tryLowerArrayBufferViewProperty(path, expression, &result, function, counter, object, objectType); handled {
			return value, valueType, err
		}
	}

	if objectType == ir.TypeDataView {
		if value, valueType, handled, err := tryLowerDataViewProperty(path, expression, &result, function, counter, object); handled {
			return value, valueType, err
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
		if value, valueType, handled, err := tryLowerTextDecoderProperty(path, expression, &result, function, counter, object); handled {
			return value, valueType, err
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
		if value, valueType, handled, err := tryLowerRegExpProperty(path, expression, &result, function, counter, object); handled {
			return value, valueType, err
		}
	}

	if expression.Text == "length" {
		if value, valueType, handled, err := tryLowerLengthProperty(path, expression, &result, function, counter, object, objectType); handled {
			return value, valueType, err
		}
	}

	className = strings.TrimPrefix(string(objectType), "object:")
	className = classIdentityForPath(path, className)

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
		className = resolveThisPropertyClass(expression, function, env, shapes, className)
	}
	if isFunctionDeclarationReference(path, expression.Left, env, signatures) && !functionValueProperty(expression.Text) {
		return "", "", fmt.Errorf("property %q on a function declaration is not supported", expression.Text)
	}
	isUnionAlias := false
	if typeAliasesIndex != nil && typeAliasesIndex[className] != "" && strings.Contains(typeAliasesIndex[className], "|") {
		isUnionAlias = true
	}
	if !isUnionAlias && (className == "" || className == "Record" || className == "closure" || strings.HasPrefix(className, "__closure_") || strings.HasPrefix(className, "Record_") || strings.HasPrefix(className, "Record<") || strings.HasPrefix(className, "Partial_") || strings.Contains(className, "[") || objectType == ir.TypeObject || objectType == ir.TypeUnknown) {
		return lowerDynamicRecordProperty(path, expression, result, function, counter, object)
	}
	shape, ok := shapes[className]
	if !ok {
		shape, ok = resolveRegisteredPropertyShape(expression, shapes, className, shape, ok)
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
		className, shape, ok = resolveUnionAliasPropertyShape(expression, shapes, className, shape, ok)
	}
	if !ok {
		return "", "", fmt.Errorf("unknown object shape %q for property %q", className, expression.Text)
	}
	if fieldIndex(shape, expression.Text) < 0 {
		className, shape = resolveFieldOwnerShape(expression, shapes, className, shape)
	}
	for _, field := range shape.Fields {
		if field.Name != expression.Text {
			continue
		}
		fType := field.Type
		if fType == ir.TypeClosure && env[object+".dynamic"] != "" {
			fType = ir.TypeDynamicFunction
		}
		if field.Optional {
			fType = ir.TypeUnknown
		}
		if propertyPath := extractPropertyPath(expression); len(propertyPath) > 0 {
			if narrowed, ok := env[strings.Join(propertyPath, ".")]; ok && narrowed != "" {
				fType = narrowed
			}
		}
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
						Op:           ir.OpFieldGet,
						Type:         field.Type,
						Result:       rawRes,
						Callee:       className,
						Field:        field.Name,
						FieldIndex:   fieldIndex(shape, field.Name),
						DynamicField: dynamicFieldAccess(className),
						Args:         []string{object},
						Span:         toIRSpan(path, expression.Span),
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
						Op:           ir.OpFieldGet,
						Type:         fType,
						Result:       result,
						Callee:       className,
						Field:        field.Name,
						FieldIndex:   fieldIndex(shape, field.Name),
						DynamicField: dynamicFieldAccess(className),
						Args:         []string{object},
						Span:         toIRSpan(path, expression.Span),
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
			Op:           ir.OpFieldGet,
			Type:         fType,
			Result:       result,
			Callee:       className,
			Field:        field.Name,
			FieldIndex:   fieldIndex(shape, field.Name),
			DynamicField: dynamicFieldAccess(className),
			Args:         []string{object},
			Span:         toIRSpan(path, expression.Span),
		})
		return result, fType, nil
	}
	if strings.HasPrefix(className, "__shape_") {
		return lowerAnonymousShapeProperty(path, expression, result, function, counter, object, objectType)
	}
	if expression.Kind == "optional_property" {
		return lowerOptionalPropertyFallback(path, expression, result, function, counter)
	}
	fieldNames := make([]string, 0, len(shape.Fields))
	for _, f := range shape.Fields {
		fieldNames = append(fieldNames, f.Name)
	}
	return "", "", fmt.Errorf("unknown field %q on object %q (type %q in %s; fields: %v)", expression.Text, className, objectType, function.Name, fieldNames)
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

// isFunctionDeclarationReference reports an identifier that names a function
// declaration. Each such reference materializes a fresh closure-ABI
// trampoline, so it has no identity to hold properties (foo.x = 1 would be
// lost, and reading foo.x would read past the closure object).
func isFunctionDeclarationReference(path string, expression *frontend.SyntaxExpression, env map[string]ir.Type, signatures map[string]ir.Function) bool {
	if expression == nil || expression.Kind != "identifier" {
		return false
	}
	if _, isVariable := env[expression.Text]; isVariable {
		return false
	}
	_, ok := resolveFunctionSignature(path, expression.Text, signatures)
	return ok
}

// functionValueProperty reports the Function properties the native subset
// models on a function declaration reference.
func functionValueProperty(name string) bool {
	switch name {
	case "name", "length", "call", "apply", "bind", "toString":
		return true
	}
	return false
}
