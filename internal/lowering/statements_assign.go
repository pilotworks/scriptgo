package lowering

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerAssignStatement(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	targetVarName := statement.Name
	if mangled, hasMangled := env["__ident."+statement.Name]; hasMangled {
		targetVarName = string(mangled)
	}
	varType, ok := env[targetVarName]
	if !ok {
		if topVar, isTop := topLevelVars[statement.Name]; isTop {
			vType := topVar.Type
			if vType == "" && topVar.InferredType != "" {
				vType = topVar.InferredType
			}
			varType = toIRType(vType)
			if varType == "" {
				varType = ir.TypeNumber
			}
			env[targetVarName] = varType
		} else {
			return fmt.Errorf("assignment to unknown variable %q", statement.Name)
		}
	}
	if (statement.Expression.Kind == "null" || statement.Expression.Kind == "undefined") && varType != ir.TypeUnknown {
		defaultVal := "0"
		if statement.Expression.Kind == "undefined" {
			if isPointerLikeType(varType) || strings.HasPrefix(string(varType), "object:") || varType == ir.TypeString {
				defaultVal = "undefined"
			} else if varType == ir.TypeNumber {
				defaultVal = "undefined"
			} else if varType == ir.TypeBool {
				defaultVal = "false"
			}
		} else {
			if isPointerLikeType(varType) || strings.HasPrefix(string(varType), "object:") || varType == ir.TypeString {
				defaultVal = "null"
			} else if varType == ir.TypeNumber {
				defaultVal = "null"
			} else if varType == ir.TypeBool {
				defaultVal = "false"
			}
		}
		tmp := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   varType,
			Result: tmp,
			Value:  defaultVal,
			Span:   toIRSpan(path, statement.Span),
		})
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   varType,
			Result: targetVarName,
			Args:   []string{tmp},
			Span:   toIRSpan(path, statement.Span),
		})
		return nil
	}
	if statement.Expression != nil && (statement.Expression.Kind == "array" || (statement.Expression.Kind == "new" && callName(statement.Expression.Left) == "Array")) {
		targetType := string(varType)
		if strings.HasSuffix(targetType, "[]") || statement.Expression.InferredType == "" || statement.Expression.InferredType == "any[]" || statement.Expression.InferredType == "never[]" || statement.Expression.InferredType == "unknown[]" {
			statement.Expression.InferredType = targetType
		}
	}
	value, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	storageType := env["__storage_type."+targetVarName]
	if storageType == "" {
		storageType = env["__storage_type."+statement.Name]
	}
	if storageType == "" {
		if topVar, isTop := topLevelVars[statement.Name]; isTop {
			vType := topVar.Type
			if vType == "" && topVar.InferredType != "" {
				vType = topVar.InferredType
			}
			storageType = toIRType(vType)
		}
	}
	if (varType == ir.TypeUnknown || storageType == ir.TypeUnknown) && valType != ir.TypeUnknown {
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpBoxUnknown,
			Type:   ir.TypeUnknown,
			Result: targetVarName,
			Args:   []string{value},
			Span:   toIRSpan(path, statement.Span),
		})
		return nil
	}
	if valType != varType {
		if (strings.HasPrefix(string(valType), "object:") || valType == ir.TypeObject) && (strings.HasPrefix(string(varType), "object:") || varType == ir.TypeObject) {
			// Polymorphic object assignment
		} else if (varType == ir.TypeUint8Array || varType == ir.TypeBuffer) && (valType == ir.TypeUint8Array || valType == ir.TypeBuffer) {
			// Buffer extends Uint8Array and Uint8Array is binary-compatible with Buffer
		} else if valType == ir.TypeUnknown {
			unboxed := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCheckedCast,
				Type:   varType,
				Result: unboxed,
				Args:   []string{value},
				Span:   toIRSpan(path, statement.Span),
			})
			value = unboxed
		} else if isPointerLikeType(varType) && (valType == ir.TypeVoid || valType == ir.TypePointer) {
			// Assigning undefined or null to a pointer-like variable
		} else if strings.HasPrefix(statement.Name, "__destruct_") || strings.HasPrefix(targetVarName, "__destruct_") {
			// Destructuring temporary assignment under out-of-bounds guard
		} else {
			return fmt.Errorf("assignment type mismatch for %q: %s := %s", statement.Name, varType, valType)
		}
	}
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpAssign, Type: varType, Result: targetVarName, Args: []string{value}, Span: toIRSpan(path, statement.Span)})
	return nil
}

func lowerIndexSetStatement(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	if statement.Left != nil && statement.Left.Kind == "property" && statement.Left.Left != nil && statement.Left.Left.Kind == "identifier" && statement.Left.Left.Text == "process" && statement.Left.Text == "env" {
		keyVal, keyType, err := lowerExpression(path, statement.Right, "", function, env, counter, shapes, signatures)
		if err != nil || keyType != ir.TypeString {
			return fmt.Errorf("process.env requires string index")
		}
		valVal, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil || valType != ir.TypeString {
			return fmt.Errorf("process.env requires string value")
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeVoid,
			Callee: "__process.set_env",
			Args:   []string{keyVal, valVal},
			Span:   toIRSpan(path, statement.Span),
		})
		return nil
	}
	arrVal, arrType, err := lowerExpression(path, statement.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	if after, ok := strings.CutPrefix(string(arrType), "object:"); ok {
		shapeName := after
		if shape, ok := shapes[shapeName]; ok {
			if statement.Right != nil && statement.Right.Kind == "number" {
				fieldIdx, _ := strconv.Atoi(statement.Right.Text)
				if fieldIdx >= 0 && fieldIdx < len(shape.Fields) {
					field := shape.Fields[fieldIdx]
					val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
					if err != nil {
						return err
					}
					function.Body = append(function.Body, ir.Instruction{
						Op:           ir.OpFieldSet,
						Type:         ir.TypeVoid,
						Callee:       shapeName,
						Field:        field.Name,
						FieldIndex:   fieldIdx,
						DynamicField: dynamicFieldAccess(shapeName),
						Args:         []string{arrVal, val},
						Span:         toIRSpan(path, statement.Span),
					})
					return nil
				}
			}
			if statement.Right != nil && statement.Right.Kind == "string" {
				propName := statement.Right.Text
				for idx, field := range shape.Fields {
					if field.Name == propName {
						val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
						if err != nil {
							return err
						}
						function.Body = append(function.Body, ir.Instruction{
							Op:           ir.OpFieldSet,
							Type:         ir.TypeVoid,
							Callee:       shapeName,
							Field:        field.Name,
							FieldIndex:   idx,
							DynamicField: dynamicFieldAccess(shapeName),
							Args:         []string{arrVal, val},
							Span:         toIRSpan(path, statement.Span),
						})
						return nil
					}
				}
			}
		}
	}
	// Dictionary-like objects use the runtime key instead of a fixed shape
	// field. Keep the array path below strict about numeric indexes.
	if arrType == ir.TypeObject || strings.HasPrefix(string(arrType), "object:") || arrType == ir.TypeUnknown {
		idxVal, idxType, err := lowerExpression(path, statement.Right, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if idxType == ir.TypeString || idxType == ir.TypeUnknown || idxType == ir.TypeSymbol {
			if idxType == ir.TypeSymbol {
				idxVal, _ = propertyKeyValue(path, statement.Right.Span, idxVal, idxType, function, counter)
			}
			val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
			if err != nil {
				return err
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeVoid,
				Callee: "__object.set_prop",
				Args:   []string{arrVal, idxVal, val},
				Span:   toIRSpan(path, statement.Span),
			})
			return nil
		}
	}
	if arrType == ir.TypeString {
		return fmt.Errorf("cannot assign to read-only string index")
	}
	idxVal, idxType, err := lowerExpression(path, statement.Right, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	if idxType != ir.TypeNumber {
		return fmt.Errorf("array index_set requires number index, got %s", idxType)
	}
	var expectedElemType ir.Type
	if arrType == ir.TypeBigInt64Array || arrType == ir.TypeBigUint64Array {
		expectedElemType = ir.TypeBigInt
	} else if isNumberTypedArray(arrType) || arrType == ir.TypeNumberArray {
		expectedElemType = ir.TypeNumber
	} else if arrType == ir.TypeStringArray {
		expectedElemType = ir.TypeString
	} else if arrType == ir.TypeBoolArray || arrType == "boolean[]" || arrType == "bool[]" {
		expectedElemType = ir.TypeBool
	} else if before, ok := strings.CutSuffix(string(arrType), "[]"); ok {
		elemName := before
		if elemName == "boolean" {
			expectedElemType = ir.TypeBool
		} else {
			expectedElemType = ir.Type(elemName)
		}
	} else {
		return fmt.Errorf("array index_set requires an array, got %s", arrType)
	}
	if statement.Expression != nil && (statement.Expression.InferredType == "" || statement.Expression.InferredType == "never[]" || statement.Expression.InferredType == "unknown[]") && expectedElemType != "" {
		statement.Expression.InferredType = string(expectedElemType)
	}
	val, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	if valType == ir.TypeVoid || (statement.Expression != nil && (statement.Expression.Kind == "undefined" || statement.Expression.Kind == "null" || (statement.Expression.Kind == "identifier" && statement.Expression.Text == "undefined"))) {
		zeroVal := nextTemp(counter)
		switch expectedElemType {
		case ir.TypeNumber:
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroVal, Value: "0", Span: toIRSpan(path, statement.Span)})
		case ir.TypeBool:
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: zeroVal, Value: "false", Span: toIRSpan(path, statement.Span)})
		case ir.TypeString:
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: zeroVal, Value: "", Span: toIRSpan(path, statement.Span)})
		default:
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: expectedElemType, Result: zeroVal, Value: "null", Span: toIRSpan(path, statement.Span)})
		}
		val = zeroVal
		valType = expectedElemType
	}
	if expectedElemType == ir.TypeUnknown && valType != ir.TypeUnknown {
		boxedVal := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpBoxUnknown,
			Type:   ir.TypeUnknown,
			Result: boxedVal,
			Args:   []string{val},
			Span:   toIRSpan(path, statement.Span),
		})
		env[boxedVal] = ir.TypeUnknown
		val = boxedVal
		valType = ir.TypeUnknown
	}
	if expectedElemType != "" && valType != expectedElemType && valType != ir.TypeUnknown && expectedElemType != ir.TypeUnknown {
		if !strings.HasSuffix(string(expectedElemType), "[]") || (valType != "never[]" && valType != "object:never[]" && valType != ir.TypeObject && valType != "never" && valType != "unknown[]") {
			return fmt.Errorf("array index_set type mismatch: %s cannot be assigned to %s", valType, arrType)
		}
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:   ir.OpIndexSet,
		Type: ir.TypeVoid,
		Args: []string{arrVal, idxVal, val},
		Span: toIRSpan(path, statement.Span),
	})
	return nil
}
