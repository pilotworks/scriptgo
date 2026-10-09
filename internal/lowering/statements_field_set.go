package lowering

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerFieldSetStatement(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	if statement.Left != nil && statement.Left.Kind == "property" && statement.Left.Left != nil && statement.Left.Left.Kind == "identifier" && statement.Left.Left.Text == "process" && statement.Left.Text == "env" {
		keyTemp := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: keyTemp, Value: statement.Name, Span: toIRSpan(path, statement.Span)})
		valVal, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil || valType != ir.TypeString {
			return fmt.Errorf("process.env requires string value")
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeVoid,
			Callee: "__process.set_env",
			Args:   []string{keyTemp, valVal},
			Span:   toIRSpan(path, statement.Span),
		})
		return nil
	}
	if statement.Left != nil && statement.Left.Kind == "identifier" {
		className := statement.Left.Text
		if meta, isClass := classHierarchy[className]; isClass {
			// Check static setter
			if _, setterName, ok := findSetterInHierarchy(className, statement.Name, signatures, classHierarchy); ok {
				val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
				if err != nil {
					return err
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   ir.TypeVoid,
					Callee: setterName,
					Args:   []string{val},
					Span:   toIRSpan(path, statement.Span),
				})
				return nil
			}
			// Static field assignment
			if _, isStatic := meta.Statics[statement.Name]; isStatic {
				staticVar := staticFieldGlobal(className, statement.Name)
				val, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
				if err != nil {
					return err
				}
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpAssign,
					Type:   valType,
					Result: staticVar,
					Args:   []string{val},
					Span:   toIRSpan(path, statement.Span),
				})
				return nil
			}
		}
	}

	objVal, objType, err := lowerExpression(path, statement.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	if (strings.HasSuffix(string(objType), "[]") || objType == ir.TypeNumberArray || objType == ir.TypeStringArray || objType == ir.TypeBoolArray || objType == ir.TypeBigIntArray) && statement.Name == "length" {
		val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeVoid,
			Callee: "__array.set_length",
			Args:   []string{objVal, val},
			Span:   toIRSpan(path, statement.Span),
		})
		return nil
	}
	className := strings.TrimPrefix(string(objType), string(ir.TypeObject)+":")
	if className == "this" || className == "" {
		if statement.Left != nil && statement.Left.Text != "" {
			if t, inEnv := env[statement.Left.Text]; inEnv && string(t) != "" && string(t) != "this" {
				className = strings.TrimPrefix(string(t), string(ir.TypeObject)+":")
			}
		}
		if className == "this" || className == "" {
			if t, inEnv := env["this"]; inEnv && string(t) != "this" && string(t) != "object:this" {
				className = strings.TrimPrefix(string(t), string(ir.TypeObject)+":")
			} else if function != nil && strings.Contains(function.Name, "_") && !strings.HasPrefix(function.Name, "__closure_") {
				className = strings.Split(function.Name, "_")[0]
			}
		}
		if className == "this" || className == "" {
			for _, sName := range slices.Sorted(maps.Keys(shapes)) {
				s := shapes[sName]
				if fieldIndex(s, statement.Name) >= 0 {
					className = sName
					break
				}
			}
		}
	}

	// Check instance setter in hierarchy
	if _, setterName, ok := findSetterInHierarchy(className, statement.Name, signatures, classHierarchy); ok {
		val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeVoid,
			Callee: setterName,
			Args:   []string{objVal, val},
			Span:   toIRSpan(path, statement.Span),
		})
		return nil
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
		}
	}
	if !ok {
		if className == "closure" || strings.HasPrefix(className, "__closure_") || className == "object" || className == "Record" || strings.HasPrefix(className, "Record_") || strings.HasPrefix(className, "Record<") || objType == ir.TypeObject || objType == ir.TypeUnknown {
			propNameConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeString,
				Result: propNameConst,
				Value:  statement.Name,
				Span:   toIRSpan(path, statement.Span),
			})
			val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
			if err != nil {
				return err
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeVoid,
				Callee: "__object.set_prop",
				Args:   []string{objVal, propNameConst, val},
				Span:   toIRSpan(path, statement.Span),
			})
			return nil
		}
		return fmt.Errorf("field set on unknown object shape %q", className)
	}
	fIndex := fieldIndex(shape, statement.Name)
	if fIndex < 0 {
		return fmt.Errorf("unknown field %q on object shape %q", statement.Name, className)
	}
	if statement.Expression != nil && (statement.Expression.Kind == "array" || (statement.Expression.Kind == "new" && callName(statement.Expression.Left) == "Array")) {
		targetFieldType := string(shape.Fields[fIndex].Type)
		if strings.HasSuffix(targetFieldType, "[]") || statement.Expression.InferredType == "" || statement.Expression.InferredType == "any[]" || statement.Expression.InferredType == "never[]" || statement.Expression.InferredType == "unknown[]" || statement.Expression.InferredType == "void[]" {
			statement.Expression.InferredType = targetFieldType
		}
	}
	var val string
	var valType ir.Type
	if statement.Expression != nil && (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && (isPointerLikeType(shape.Fields[fIndex].Type) || shape.Fields[fIndex].Type == ir.TypePointer) {
		val = nextTemp(counter)
		valType = shape.Fields[fIndex].Type
		value := "null"
		if statement.Expression.Kind == "undefined" {
			value = "undefined"
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   valType,
			Result: val,
			Value:  value,
			Span:   toIRSpan(path, statement.Span),
		})
	} else {
		var err error
		val, valType, err = lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
	}
	if valType != shape.Fields[fIndex].Type {
		if strings.HasSuffix(string(shape.Fields[fIndex].Type), "[]") && (valType == "void[]" || valType == "never[]" || valType == "unknown[]") {
			// allowed array assignment
		} else if valType == ir.TypePointer && (isPointerLikeType(shape.Fields[fIndex].Type) || shape.Fields[fIndex].Type == ir.TypePointer) {
			// allowed null pointer assignment to pointer-like field
		} else if (strings.HasPrefix(string(valType), "object:") || valType == ir.TypeObject) && (strings.HasPrefix(string(shape.Fields[fIndex].Type), "object:") || shape.Fields[fIndex].Type == ir.TypeObject) {
			// allowed object assignment
		} else if isSubtype(string(valType), string(shape.Fields[fIndex].Type)) {
			// allowed subtype/interface implementation assignment
		} else if shape.Fields[fIndex].Type == ir.TypeUnknown {
			boxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpBoxUnknown,
				Type:   ir.TypeUnknown,
				Result: boxed,
				Args:   []string{val},
				Span:   toIRSpan(path, statement.Span),
			})
			val = boxed
		} else if valType == ir.TypeUnknown {
			unboxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCheckedCast,
				Type:   shape.Fields[fIndex].Type,
				Result: unboxed,
				Args:   []string{val},
				Span:   toIRSpan(path, statement.Span),
			})
			val = unboxed
		} else {
			return fmt.Errorf("field set type mismatch for %q: %s := %s", statement.Name, shape.Fields[fIndex].Type, valType)
		}
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:           ir.OpFieldSet,
		Type:         ir.TypeVoid,
		Callee:       className,
		Field:        statement.Name,
		FieldIndex:   fIndex,
		DynamicField: dynamicFieldAccess(className),
		Args:         []string{objVal, val},
		Span:         toIRSpan(path, statement.Span),
	})
	return nil
}
