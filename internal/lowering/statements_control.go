package lowering

import (
	"fmt"
	"maps"
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func coerceToBool(path string, value string, valType ir.Type, function *ir.Function, counter *int, span typescriptgo.SourceSpan) (string, error) {
	if valType == ir.TypeBool {
		return value, nil
	}
	if valType == ir.TypeNumber {
		zeroConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, span)})
		boolRes := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: boolRes, Operator: "!=", Args: []string{value, zeroConst}, Span: toIRSpan(path, span)})
		return boolRes, nil
	}
	if valType == ir.TypeString {
		nullConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: nullConst, Value: "null", Span: toIRSpan(path, span)})
		nonNull := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: nonNull, Operator: "!=", Args: []string{value, nullConst}, Span: toIRSpan(path, span)})

		emptyConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: emptyConst, Value: `""`, Span: toIRSpan(path, span)})
		nonEmpty := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: nonEmpty, Operator: "!=", Args: []string{value, emptyConst}, Span: toIRSpan(path, span)})

		boolRes := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: boolRes, Operator: "&&", Args: []string{nonNull, nonEmpty}, Span: toIRSpan(path, span)})
		return boolRes, nil
	}
	if valType == ir.TypeUnknown {
		boolRes := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeBool,
			Result: boolRes,
			Callee: "__scriptgo.is_truthy",
			Args:   []string{value},
			Span:   toIRSpan(path, span),
		})
		return boolRes, nil
	}
	if valType == ir.TypeObject || strings.HasPrefix(string(valType), "object:") || strings.HasSuffix(string(valType), "[]") || valType == ir.TypeNumberArray || valType == ir.TypeStringArray || valType == ir.TypeBoolArray || valType == ir.TypeBigIntArray || valType == ir.TypeMap || valType == ir.TypeSet || valType == ir.TypeBuffer || valType == ir.TypeUint8Array || isPointerLikeType(valType) || valType == ir.TypeClosure || valType == ir.TypePointer {
		nullConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: valType, Result: nullConst, Value: "null", Span: toIRSpan(path, span)})
		boolRes := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: boolRes, Operator: "!=", Args: []string{value, nullConst}, Span: toIRSpan(path, span)})
		return boolRes, nil
	}
	return "", fmt.Errorf("cannot coerce %s to boolean condition", valType)
}

func lowerIf(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	condition, typ, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	condition, err = coerceToBool(path, condition, typ, function, counter, statement.Span)
	if err != nil {
		return fmt.Errorf("if condition: %w", err)
	}
	thenEnv := make(map[string]ir.Type, len(env))
	maps.Copy(thenEnv, env)
	elseEnv := make(map[string]ir.Type, len(env))
	maps.Copy(elseEnv, env)

	if statement.Expression != nil && statement.Expression.Kind == "binary" && statement.Expression.Operator == "in" && statement.Expression.Left != nil && (statement.Expression.Left.Kind == "string" || statement.Expression.Left.Kind == "literal") && statement.Expression.Right != nil && statement.Expression.Right.Kind == "identifier" {
		varName := statement.Expression.Right.Text
		propName := strings.Trim(statement.Expression.Left.Text, "\"'`")
		for typeName, typeDef := range typeAliasesIndex {
			if fields, ok := anonymousObjectFields(typeDef, nil); ok {
				hasProp := false
				for _, f := range fields {
					if f.Name == propName {
						hasProp = true
						break
					}
				}
				if hasProp {
					thenEnv[varName] = toIRType(typeName)
				} else {
					elseEnv[varName] = toIRType(typeName)
				}
			}
		}
	} else if statement.Expression != nil && statement.Expression.Kind == "binary" && statement.Expression.Operator == "instanceof" && statement.Expression.Left != nil && statement.Expression.Left.Kind == "identifier" && statement.Expression.Right != nil && (statement.Expression.Right.Kind == "identifier" || statement.Expression.Right.Kind == "type") {
		varName := statement.Expression.Left.Text
		className := statement.Expression.Right.Text
		thenEnv[varName] = toIRType(className)
		if currType, ok := env[varName]; ok {
			cleanCurr := strings.TrimPrefix(string(currType), "object:")
			if strings.Contains(cleanCurr, "|") {
				var remaining []string
				for _, part := range strings.Split(cleanCurr, "|") {
					trimmed := strings.TrimSpace(part)
					if trimmed != className && trimmed != "object:"+className {
						remaining = append(remaining, trimmed)
					}
				}
				if len(remaining) > 0 {
					elseEnv[varName] = toIRType(strings.Join(remaining, " | "))
				}
			}
		}
	} else if statement.Expression != nil && statement.Expression.Kind == "binary" && (statement.Expression.Operator == "===" || statement.Expression.Operator == "==") {
		left := statement.Expression.Left
		right := statement.Expression.Right
		if left != nil && left.Kind == "typeof" && left.Left != nil && left.Left.Kind == "identifier" && right != nil && (right.Kind == "string" || right.Kind == "literal") {
			varName := left.Left.Text
			valStr := strings.Trim(right.Text, "\"'`")
			switch valStr {
			case "number":
				thenEnv[varName] = ir.TypeNumber
			case "string":
				thenEnv[varName] = ir.TypeString
			case "boolean":
				thenEnv[varName] = ir.TypeBool
			case "bigint":
				thenEnv[varName] = ir.TypeBigInt
			}
		} else if right != nil && right.Kind == "typeof" && right.Left != nil && right.Left.Kind == "identifier" && left != nil && (left.Kind == "string" || left.Kind == "literal") {
			varName := right.Left.Text
			valStr := strings.Trim(left.Text, "\"'`")
			switch valStr {
			case "number":
				thenEnv[varName] = ir.TypeNumber
			case "string":
				thenEnv[varName] = ir.TypeString
			case "boolean":
				thenEnv[varName] = ir.TypeBool
			case "bigint":
				thenEnv[varName] = ir.TypeBigInt
			}
		}
		var propAccess *typescriptgo.SyntaxExpression
		var literalVal *typescriptgo.SyntaxExpression
		if left != nil && (left.Kind == "property" || left.Kind == "member") && right != nil && (right.Kind == "string" || right.Kind == "literal") {
			propAccess = left
			literalVal = right
		} else if right != nil && (right.Kind == "property" || right.Kind == "member") && left != nil && (left.Kind == "string" || left.Kind == "literal") {
			propAccess = right
			literalVal = left
		}
		if propAccess != nil && propAccess.Left != nil && propAccess.Left.Kind == "identifier" && literalVal != nil {
			varName := propAccess.Left.Text
			valStr := literalVal.Text
			if currType, ok := env[varName]; ok {
				matched := findMatchingDiscriminatedType(valStr, string(currType), shapes)
				if matched != "" {
					thenEnv[varName] = ir.Type("object:" + matched)
				}
			}
		}
	}

	thenBody, err := lowerBranch(path, statement.Then, function.ReturnType, thenEnv, counter, shapes, signatures)
	if err != nil {
		return err
	}
	elseBody, err := lowerBranch(path, statement.Else, function.ReturnType, elseEnv, counter, shapes, signatures)
	if err != nil {
		return err
	}
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpIf, Type: ir.TypeVoid, Args: []string{condition}, Then: thenBody, Else: elseBody, Span: toIRSpan(path, statement.Span)})
	return nil
}

func lowerWhile(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	condFunc := ir.Function{Name: "cond", ReturnType: ir.TypeBool}
	condVal, condType, err := lowerExpression(path, statement.Expression, "", &condFunc, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	condVal, err = coerceToBool(path, condVal, condType, &condFunc, counter, statement.Span)
	if err != nil {
		return fmt.Errorf("while condition: %w", err)
	}
	bodyInstructions, err := lowerBranch(path, statement.Body, function.ReturnType, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	var stepInstructions []ir.Instruction
	if len(statement.Step) > 0 {
		stepInstructions, err = lowerBranch(path, statement.Step, function.ReturnType, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:    ir.OpWhile,
		Type:  ir.TypeVoid,
		Value: statement.Label,
		Args:  []string{condVal},
		Cond:  condFunc.Body,
		Body:  bodyInstructions,
		Step:  stepInstructions,
		Span:  toIRSpan(path, statement.Span),
	})
	return nil
}

func lowerDoWhile(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	condFunc := ir.Function{Name: "cond", ReturnType: ir.TypeBool}
	condVal, condType, err := lowerExpression(path, statement.Expression, "", &condFunc, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	condVal, err = coerceToBool(path, condVal, condType, &condFunc, counter, statement.Span)
	if err != nil {
		return fmt.Errorf("do-while condition: %w", err)
	}
	bodyInstructions, err := lowerBranch(path, statement.Body, function.ReturnType, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	var stepInstructions []ir.Instruction
	if len(statement.Step) > 0 {
		stepInstructions, err = lowerBranch(path, statement.Step, function.ReturnType, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:    ir.OpDoWhile,
		Type:  ir.TypeVoid,
		Value: statement.Label,
		Args:  []string{condVal},
		Cond:  condFunc.Body,
		Body:  bodyInstructions,
		Step:  stepInstructions,
		Span:  toIRSpan(path, statement.Span),
	})
	return nil
}

func lowerForOf(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	arrVal, arrType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	if (strings.Contains(string(arrType), "Generator") || strings.Contains(string(arrType), "Iterator")) && !strings.Contains(string(arrType), "MapIterator") && !strings.Contains(string(arrType), "SetIterator") {
		shapeName := strings.TrimPrefix(string(arrType), "object:")
		nextFn := shapeName + "_next"
		targetNext, hasNext := signatures[nextFn]
		valType := ir.TypeNumber
		resShapeName := ""
		if hasNext {
			resShapeName = strings.TrimPrefix(string(targetNext.ReturnType), "object:")
			if resShape, ok := shapes[resShapeName]; ok && len(resShape.Fields) > 1 {
				valType = resShape.Fields[1].Type
			}
		} else {
			if statement.Expression != nil && strings.Contains(statement.Expression.InferredType, "<") && strings.HasSuffix(strings.TrimSpace(statement.Expression.InferredType), ">") {
				inferred := strings.TrimSpace(statement.Expression.InferredType)
				idx := strings.Index(inferred, "<")
				inner := inferred[idx+1 : len(inferred)-1]
				parts := splitTypeArguments(inner)
				if len(parts) > 0 {
					valType = toIRType(parts[0])
				}
			} else if strings.Contains(string(arrType), "<") && strings.HasSuffix(string(arrType), ">") {
				idx := strings.Index(string(arrType), "<")
				inner := string(arrType)[idx+1 : len(string(arrType))-1]
				parts := splitTypeArguments(inner)
				if len(parts) > 0 {
					valType = toIRType(parts[0])
				}
			}
			nextFn = "__generator.next"
			resShapeName = fmt.Sprintf("IteratorResult_%s", valType)
			if _, exists := shapes[resShapeName]; !exists {
				shapes[resShapeName] = ir.ObjectShape{
					Name: resShapeName,
					Span: toIRSpan(path, statement.Span),
					Fields: []ir.Field{
						{Name: "done", Type: ir.TypeBool, Span: toIRSpan(path, statement.Span)},
						{Name: "value", Type: valType, Span: toIRSpan(path, statement.Span)},
					},
				}
			}
		}

		condFunc := ir.Function{Name: "cond", ReturnType: ir.TypeBool}
		condConst := nextTemp(counter)
		condFunc.Body = append(condFunc.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: condConst, Value: "true", Span: toIRSpan(path, statement.Span)})

		bodyEnv := make(map[string]ir.Type, len(env)+2)
		maps.Copy(bodyEnv, env)
		bodyBranch := ir.Function{Name: "body", ReturnType: function.ReturnType}

		resVal := nextTemp(counter)
		retType := ir.Type("object:" + resShapeName)
		if hasNext {
			retType = targetNext.ReturnType
		}
		bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   retType,
			Result: resVal,
			Callee: nextFn,
			Args:   []string{arrVal},
			Span:   toIRSpan(path, statement.Span),
		})

		doneVal := nextTemp(counter)
		bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
			Op:         ir.OpFieldGet,
			Type:       ir.TypeBool,
			Result:     doneVal,
			Callee:     resShapeName,
			Field:      "done",
			FieldIndex: 0,
			Args:       []string{resVal},
			Span:       toIRSpan(path, statement.Span),
		})

		bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
			Op:   ir.OpIf,
			Type: ir.TypeVoid,
			Args: []string{doneVal},
			Then: []ir.Instruction{
				{Op: ir.OpBreak, Type: ir.TypeVoid, Span: toIRSpan(path, statement.Span)},
			},
			Span: toIRSpan(path, statement.Span),
		})

		valVal := nextTemp(counter)
		bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
			Op:         ir.OpFieldGet,
			Type:       valType,
			Result:     valVal,
			Callee:     resShapeName,
			Field:      "value",
			FieldIndex: 1,
			Args:       []string{resVal},
			Span:       toIRSpan(path, statement.Span),
		})
		bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   valType,
			Result: statement.Name,
			Args:   []string{valVal},
			Span:   toIRSpan(path, statement.Span),
		})
		bodyEnv[statement.Name] = valType

		for _, bodyStmt := range statement.Body {
			if err := lowerStatement(path, bodyStmt, &bodyBranch, bodyEnv, counter, shapes, signatures); err != nil {
				return err
			}
		}

		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpWhile,
			Type: ir.TypeVoid,
			Args: []string{condConst},
			Cond: condFunc.Body,
			Body: bodyBranch.Body,
			Span: toIRSpan(path, statement.Span),
		})
		return nil
	}

	if shapeName, ok := strings.CutPrefix(string(arrType), "object:"); ok {
		var s ir.ObjectShape
		var found bool
		if s, found = shapes[shapeName]; !found {
			s, found = anonymousShapes[shapeName]
		}
		if found && len(s.Fields) > 0 {
			isTuple := true
			for i, f := range s.Fields {
				if f.Name != strconv.Itoa(i) {
					isTuple = false
					break
				}
			}
			if isTuple {
				for i, f := range s.Fields {
					function.Body = append(function.Body, ir.Instruction{
						Op:         ir.OpFieldGet,
						Type:       f.Type,
						Result:     statement.Name,
						Callee:     shapeName,
						Field:      f.Name,
						FieldIndex: i,
						Args:       []string{arrVal},
						Span:       toIRSpan(path, statement.Span),
					})
					iterEnv := make(map[string]ir.Type, len(env)+1)
					for k, v := range env {
						iterEnv[k] = v
					}
					iterEnv[statement.Name] = f.Type
					for _, bodyStmt := range statement.Body {
						if err := lowerStatement(path, bodyStmt, function, iterEnv, counter, shapes, signatures); err != nil {
							return err
						}
					}
				}
				return nil
			}
		}
	}

	isString := (arrType == ir.TypeString)
	var elemType ir.Type
	if isString {
		elemType = ir.TypeString
	} else if strings.Contains(string(arrType), "MapIterator") || strings.Contains(string(arrType), "SetIterator") {
		if after, ok := strings.CutPrefix(string(arrType), "object:MapIterator__"); ok {
			clean := after
			if strings.HasPrefix(clean, "[") && strings.HasSuffix(clean, "]") {
				inner := clean[1 : len(clean)-1]
				parts := strings.Split(inner, "_")
				var fields []ir.Field
				for i, p := range parts {
					fields = append(fields, ir.Field{
						Name: strconv.Itoa(i),
						Type: toIRType(p),
					})
				}
				name := anonymousShapeName(fields)
				registerAnonymousShape(name, fields)
				elemType = ir.Type("object:" + name)
			} else {
				elemType = toIRType(clean)
			}
		} else if after, ok := strings.CutPrefix(string(arrType), "object:SetIterator__"); ok {
			elemType = toIRType(after)
		} else if statement.Expression != nil && strings.Contains(statement.Expression.InferredType, "<") && strings.HasSuffix(strings.TrimSpace(statement.Expression.InferredType), ">") {
			inferred := strings.TrimSpace(statement.Expression.InferredType)
			idx := strings.Index(inferred, "<")
			inner := inferred[idx+1 : len(inferred)-1]
			parts := splitTypeArguments(inner)
			if len(parts) > 0 {
				elemType = toIRType(parts[0])
			}
		} else {
			elemType = ir.TypeUnknown
		}
	} else if before, ok := strings.CutSuffix(string(arrType), "[]"); ok {
		elemType = ir.Type(before)
	} else if arrType == ir.TypeStringArray {
		elemType = ir.TypeString
	} else if arrType == ir.TypeNumberArray {
		elemType = ir.TypeNumber
	} else {
		return fmt.Errorf("for...of requires iterable array or string, got %s", arrType)
	}

	idxName := fmt.Sprintf("__i_%d", *counter)
	lenName := fmt.Sprintf("__len_%d", *counter)
	*counter++
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: idxName, Value: "0", Span: toIRSpan(path, statement.Span)})
	env[idxName] = ir.TypeNumber
	if isString {
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: lenName, Callee: "__string.length", Args: []string{arrVal}, Span: toIRSpan(path, statement.Span)})
	} else {
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: lenName, Callee: "__array.length", Args: []string{arrVal}, Span: toIRSpan(path, statement.Span)})
	}
	env[lenName] = ir.TypeNumber

	condFunc := ir.Function{Name: "cond", ReturnType: ir.TypeBool}
	condCmp := fmt.Sprintf("__cmp_%d", *counter)
	*counter++
	condFunc.Body = append(condFunc.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: condCmp, Operator: "<", Args: []string{idxName, lenName}, Span: toIRSpan(path, statement.Span)})

	bodyEnv := make(map[string]ir.Type, len(env)+2)
	maps.Copy(bodyEnv, env)
	bodyBranch := ir.Function{Name: "body", ReturnType: function.ReturnType}
	if statement.Kind == "forawaitof" {
		if isString {
			charVal := nextTemp(counter)
			bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: charVal, Callee: "__string.charAt", Args: []string{arrVal, idxName}, Span: toIRSpan(path, statement.Span)})
			bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpAssign, Type: ir.TypeString, Result: statement.Name, Args: []string{charVal}, Span: toIRSpan(path, statement.Span)})
			bodyEnv[statement.Name] = ir.TypeString
		} else if strings.HasPrefix(string(elemType), "object:Promise") || elemType == "object:Promise" {
			rawPromise := nextTemp(counter)
			bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpIndex, Type: elemType, Result: rawPromise, Args: []string{arrVal, idxName}, Span: toIRSpan(path, statement.Span)})
			bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: statement.Name,
				Callee: "__async.await",
				Args:   []string{rawPromise},
				Span:   toIRSpan(path, statement.Span),
			})
			bodyEnv[statement.Name] = ir.TypeNumber
		} else {
			rawVal := nextTemp(counter)
			bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpIndex, Type: elemType, Result: rawVal, Args: []string{arrVal, idxName}, Span: toIRSpan(path, statement.Span)})
			bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   elemType,
				Result: statement.Name,
				Args:   []string{rawVal},
				Span:   toIRSpan(path, statement.Span),
			})
			bodyEnv[statement.Name] = elemType
		}
	} else {
		if isString {
			bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: statement.Name, Callee: "__string.charAt", Args: []string{arrVal, idxName}, Span: toIRSpan(path, statement.Span)})
		} else {
			bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpIndex, Type: elemType, Result: statement.Name, Args: []string{arrVal, idxName}, Span: toIRSpan(path, statement.Span)})
		}
		bodyEnv[statement.Name] = elemType
	}

	for _, bodyStmt := range statement.Body {
		if err := lowerStatement(path, bodyStmt, &bodyBranch, bodyEnv, counter, shapes, signatures); err != nil {
			return err
		}
	}

	stepBranch := ir.Function{Name: "step", ReturnType: function.ReturnType}
	incVal := fmt.Sprintf("__inc_%d", *counter)
	oneVal := fmt.Sprintf("__one_%d", *counter)
	*counter++
	stepBranch.Body = append(stepBranch.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: oneVal, Value: "1", Span: toIRSpan(path, statement.Span)})
	stepBranch.Body = append(stepBranch.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeNumber, Operator: "+", Result: incVal, Args: []string{idxName, oneVal}, Span: toIRSpan(path, statement.Span)})
	stepBranch.Body = append(stepBranch.Body, ir.Instruction{Op: ir.OpAssign, Type: ir.TypeNumber, Result: idxName, Args: []string{incVal}, Span: toIRSpan(path, statement.Span)})

	function.Body = append(function.Body, ir.Instruction{
		Op:    ir.OpWhile,
		Type:  ir.TypeVoid,
		Value: statement.Label,
		Args:  []string{condCmp},
		Cond:  condFunc.Body,
		Body:  bodyBranch.Body,
		Step:  stepBranch.Body,
		Span:  toIRSpan(path, statement.Span),
	})
	return nil
}

func lowerForIn(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	objVal, objType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	if strings.HasSuffix(string(objType), "[]") || objType == ir.TypeNumberArray || objType == ir.TypeStringArray {
		idxName := fmt.Sprintf("__i_%d", *counter)
		lenName := fmt.Sprintf("__len_%d", *counter)
		*counter++
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: idxName, Value: "0", Span: toIRSpan(path, statement.Span)})
		env[idxName] = ir.TypeNumber
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: lenName, Callee: "__array.length", Args: []string{objVal}, Span: toIRSpan(path, statement.Span)})
		env[lenName] = ir.TypeNumber

		condFunc := ir.Function{Name: "cond", ReturnType: ir.TypeBool}
		condCmp := fmt.Sprintf("__cmp_%d", *counter)
		*counter++
		condFunc.Body = append(condFunc.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: condCmp, Operator: "<", Args: []string{idxName, lenName}, Span: toIRSpan(path, statement.Span)})

		bodyEnv := make(map[string]ir.Type, len(env)+2)
		maps.Copy(bodyEnv, env)
		bodyBranch := ir.Function{Name: "body", ReturnType: function.ReturnType}
		keyType := toIRType(statement.Type)
		if keyType == "" && statement.InferredType != "" {
			keyType = toIRType(statement.InferredType)
		}
		if keyType == ir.TypeNumber {
			bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeNumber, Operator: "+", Result: statement.Name, Args: []string{idxName, idxName}, Span: toIRSpan(path, statement.Span)})
			bodyEnv[statement.Name] = ir.TypeNumber
		} else {
			bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: statement.Name, Callee: "__string.fromNumber", Args: []string{idxName}, Span: toIRSpan(path, statement.Span)})
			bodyEnv[statement.Name] = ir.TypeString
		}

		for _, bodyStmt := range statement.Body {
			if err := lowerStatement(path, bodyStmt, &bodyBranch, bodyEnv, counter, shapes, signatures); err != nil {
				return err
			}
		}

		stepBranch := ir.Function{Name: "step", ReturnType: function.ReturnType}
		incVal := fmt.Sprintf("__inc_%d", *counter)
		oneVal := fmt.Sprintf("__one_%d", *counter)
		*counter++
		stepBranch.Body = append(stepBranch.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: oneVal, Value: "1", Span: toIRSpan(path, statement.Span)})
		stepBranch.Body = append(stepBranch.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeNumber, Operator: "+", Result: incVal, Args: []string{idxName, oneVal}, Span: toIRSpan(path, statement.Span)})
		stepBranch.Body = append(stepBranch.Body, ir.Instruction{Op: ir.OpAssign, Type: ir.TypeNumber, Result: idxName, Args: []string{incVal}, Span: toIRSpan(path, statement.Span)})

		function.Body = append(function.Body, ir.Instruction{
			Op:    ir.OpWhile,
			Type:  ir.TypeVoid,
			Value: statement.Label,
			Args:  []string{condCmp},
			Cond:  condFunc.Body,
			Body:  bodyBranch.Body,
			Step:  stepBranch.Body,
			Span:  toIRSpan(path, statement.Span),
		})
		return nil
	} else if after, ok := strings.CutPrefix(string(objType), "object:"); ok {
		shapeName := after
		shape, ok := shapes[shapeName]
		if !ok {
			if s, exists := anonymousShapes[shapeName]; exists {
				shape = s
				ok = true
			} else if s, exists := registeredShapes[shapeName]; exists {
				shape = s
				ok = true
			} else if aliased, exists := typeAliasesIndex[shapeName]; exists && aliased != shapeName {
				cleanAliased := strings.TrimPrefix(aliased, "object:")
				if s, exists2 := shapes[cleanAliased]; exists2 {
					shape = s
					ok = true
				} else if s, exists2 := anonymousShapes[cleanAliased]; exists2 {
					shape = s
					ok = true
				} else if s, exists2 := registeredShapes[cleanAliased]; exists2 {
					shape = s
					ok = true
				}
			}
		}
		if !ok {
			if statement.Expression != nil && statement.Expression.Kind == "identifier" {
				if topVar, exists := topLevelVars[statement.Expression.Text]; exists && topVar.Expression != nil && topVar.Expression.Kind == "object_literal" {
					var fields []ir.Field
					for _, prop := range topVar.Expression.Arguments {
						fields = append(fields, ir.Field{Name: prop.Text, Type: toIRType(prop.InferredType)})
					}
					shape = ir.ObjectShape{Name: shapeName, Fields: fields}
					ok = true
				}
			}
		}
		if !ok {
			return fmt.Errorf("unknown shape %q for for...in", shapeName)
		}
		for _, f := range shape.Fields {
			fieldEnv := make(map[string]ir.Type, len(env)+1)
			maps.Copy(fieldEnv, env)
			fieldEnv[statement.Name] = ir.TypeString
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeString,
				Result: statement.Name,
				Value:  f.Name,
				Span:   toIRSpan(path, statement.Span),
			})
			for _, bodyStmt := range statement.Body {
				substStmt := substituteStringIndexInStmt(bodyStmt, statement.Name, f.Name)
				if err := lowerStatement(path, substStmt, function, fieldEnv, counter, shapes, signatures); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return fmt.Errorf("for...in requires object or array, got %s", objType)
}

func substituteStringIndex(expr *typescriptgo.SyntaxExpression, varName string, stringVal string) *typescriptgo.SyntaxExpression {
	if expr == nil {
		return nil
	}
	copy := *expr
	if (copy.Kind == "index" || copy.Kind == "optional_index") && copy.Right != nil && copy.Right.Kind == "identifier" && copy.Right.Text == varName {
		copy.Right = &typescriptgo.SyntaxExpression{
			Span: copy.Right.Span,
			Kind: "string",
			Text: stringVal,
		}
	}
	if copy.Left != nil {
		copy.Left = substituteStringIndex(copy.Left, varName, stringVal)
	}
	if copy.Right != nil {
		copy.Right = substituteStringIndex(copy.Right, varName, stringVal)
	}
	if len(copy.Arguments) > 0 {
		newArgs := make([]*typescriptgo.SyntaxExpression, len(copy.Arguments))
		for i, a := range copy.Arguments {
			newArgs[i] = substituteStringIndex(a, varName, stringVal)
		}
		copy.Arguments = newArgs
	}
	return &copy
}

func substituteStringIndexInStmt(stmt typescriptgo.SyntaxStatement, varName string, stringVal string) typescriptgo.SyntaxStatement {
	copy := stmt
	if copy.Expression != nil {
		copy.Expression = substituteStringIndex(copy.Expression, varName, stringVal)
	}
	if len(copy.Body) > 0 {
		newBody := make([]typescriptgo.SyntaxStatement, len(copy.Body))
		for i, s := range copy.Body {
			newBody[i] = substituteStringIndexInStmt(s, varName, stringVal)
		}
		copy.Body = newBody
	}
	return copy
}

func lowerLabel(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	bodyInstructions, err := lowerBranch(path, statement.Body, function.ReturnType, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:    ir.OpWhile,
		Type:  ir.TypeVoid,
		Value: statement.Label,
		Body:  bodyInstructions,
		Span:  toIRSpan(path, statement.Span),
	})
	return nil
}

func lowerSwitch(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	targetVal, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	matchedVar := fmt.Sprintf("__matched_%d", *counter)
	*counter++
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: matchedVar,
		Value:  "false",
		Span:   toIRSpan(path, statement.Span),
	})
	env[matchedVar] = ir.TypeBool

	loopTrue := fmt.Sprintf("__true_%d", *counter)
	*counter++
	condFunc := ir.Function{Name: "cond", ReturnType: ir.TypeBool}
	condFunc.Body = append(condFunc.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeBool,
		Result: loopTrue,
		Value:  "true",
		Span:   toIRSpan(path, statement.Span),
	})

	switchEnv := make(map[string]ir.Type, len(env)+2)
	maps.Copy(switchEnv, env)
	switchBranch := ir.Function{Name: "switch_body", ReturnType: function.ReturnType}

	var normalCases []typescriptgo.SyntaxSwitchCase
	var defaultCase *typescriptgo.SyntaxSwitchCase
	for i := range statement.Cases {
		c := statement.Cases[i]
		if c.Expression != nil {
			normalCases = append(normalCases, c)
		} else {
			defaultCase = &statement.Cases[i]
		}
	}

	for _, c := range normalCases {
		caseVal, _, err := lowerExpression(path, c.Expression, "", &switchBranch, switchEnv, counter, shapes, signatures)
		if err != nil {
			return err
		}
		cmpResult := fmt.Sprintf("__cmp_%d", *counter)
		*counter++
		switchBranch.Body = append(switchBranch.Body, ir.Instruction{
			Op:       ir.OpCompare,
			Type:     ir.TypeBool,
			Result:   cmpResult,
			Operator: "===",
			Args:     []string{targetVal, caseVal},
			Span:     toIRSpan(path, c.Span),
		})
		switchEnv[cmpResult] = ir.TypeBool

		newMatched := fmt.Sprintf("__new_matched_%d", *counter)
		*counter++
		switchBranch.Body = append(switchBranch.Body, ir.Instruction{
			Op:       ir.OpBinary,
			Type:     ir.TypeBool,
			Result:   newMatched,
			Operator: "||",
			Args:     []string{matchedVar, cmpResult},
			Span:     toIRSpan(path, c.Span),
		})
		switchBranch.Body = append(switchBranch.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   ir.TypeBool,
			Result: matchedVar,
			Args:   []string{newMatched},
			Span:   toIRSpan(path, c.Span),
		})

		caseEnv := make(map[string]ir.Type, len(switchEnv))
		maps.Copy(caseEnv, switchEnv)
		if statement.Expression != nil && (statement.Expression.Kind == "property" || statement.Expression.Kind == "member") && statement.Expression.Left != nil && statement.Expression.Left.Kind == "identifier" && c.Expression != nil && (c.Expression.Kind == "string" || c.Expression.Kind == "literal") {
			varName := statement.Expression.Left.Text
			valStr := c.Expression.Text
			if currType, ok := switchEnv[varName]; ok {
				matched := findMatchingDiscriminatedType(valStr, string(currType), shapes)
				if matched != "" {
					caseEnv[varName] = ir.Type("object:" + matched)
				}
			}
		}

		caseStmts, err := lowerBranch(path, c.Statements, function.ReturnType, caseEnv, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if len(caseStmts) > 0 {
			switchBranch.Body = append(switchBranch.Body, ir.Instruction{
				Op:   ir.OpIf,
				Type: ir.TypeVoid,
				Args: []string{matchedVar},
				Then: caseStmts,
				Span: toIRSpan(path, c.Span),
			})
		}
	}

	if defaultCase != nil {
		defStmts, err := lowerBranch(path, defaultCase.Statements, function.ReturnType, switchEnv, counter, shapes, signatures)
		if err != nil {
			return err
		}
		switchBranch.Body = append(switchBranch.Body, defStmts...)
	}

	switchBranch.Body = append(switchBranch.Body, ir.Instruction{
		Op:   ir.OpBreak,
		Type: ir.TypeVoid,
		Span: toIRSpan(path, statement.Span),
	})

	function.Body = append(function.Body, ir.Instruction{
		Op:   ir.OpWhile,
		Type: ir.TypeVoid,
		Args: []string{loopTrue},
		Cond: condFunc.Body,
		Body: switchBranch.Body,
		Span: toIRSpan(path, statement.Span),
	})
	return nil
}

var activeReturnFinallyStack [][]typescriptgo.SyntaxStatement
var activeThrowFinallyStack [][]typescriptgo.SyntaxStatement

func lowerTry(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	if len(statement.Finally) > 0 {
		activeReturnFinallyStack = append(activeReturnFinallyStack, statement.Finally)
		defer func() {
			activeReturnFinallyStack = activeReturnFinallyStack[:len(activeReturnFinallyStack)-1]
		}()
	}

	bodyInstructions, err := lowerBranch(path, statement.Body, function.ReturnType, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	var catchInstructions []ir.Instruction
	if len(statement.Catch) > 0 {
		if len(statement.Finally) > 0 {
			activeThrowFinallyStack = append(activeThrowFinallyStack, statement.Finally)
			defer func() {
				activeThrowFinallyStack = activeThrowFinallyStack[:len(activeThrowFinallyStack)-1]
			}()
		}
		catchEnv := make(map[string]ir.Type, len(env)+1)
		maps.Copy(catchEnv, env)
		if statement.CatchVar != "" {
			catchEnv[statement.CatchVar] = ir.Type("object:Error")
		}
		catchBranch := ir.Function{Name: "catch", ReturnType: function.ReturnType}
		for _, catchStmt := range statement.Catch {
			if err := lowerStatement(path, catchStmt, &catchBranch, catchEnv, counter, shapes, signatures); err != nil {
				return err
			}
		}
		catchInstructions = catchBranch.Body
	}
	var finallyInstructions []ir.Instruction
	if len(statement.Finally) > 0 {
		finallyBranch, err := lowerBranch(path, statement.Finally, function.ReturnType, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		finallyInstructions = finallyBranch
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:       ir.OpTry,
		Type:     ir.TypeVoid,
		Body:     bodyInstructions,
		CatchVar: statement.CatchVar,
		Catch:    catchInstructions,
		Finally:  finallyInstructions,
		Span:     toIRSpan(path, statement.Span),
	})
	return nil
}

func findMatchingDiscriminatedType(targetVal string, unionType string, shapes map[string]ir.ObjectShape) string {
	cleanTarget := strings.ToLower(strings.Trim(targetVal, "\"'`"))
	cleanUnion := strings.TrimPrefix(unionType, "object:")
	if aliased, ok := typeAliasesIndex[cleanUnion]; ok {
		cleanUnion = aliased
	}
	if strings.Contains(cleanUnion, "|") {
		for _, typePart := range strings.Split(cleanUnion, "|") {
			part := strings.TrimSpace(typePart)
			part = strings.TrimPrefix(part, "object:")
			if strings.Contains(strings.ToLower(part), cleanTarget) {
				return part
			}
		}
	}
	for shapeName, s := range shapes {
		for _, f := range s.Fields {
			if strings.EqualFold(f.Name, "kind") || strings.EqualFold(f.Name, "type") || strings.EqualFold(f.Name, "tag") || strings.EqualFold(f.Name, "status") {
				cleanVal := strings.ToLower(strings.Trim(f.Value, "\"'`"))
				if cleanVal == cleanTarget {
					if cleanUnion == "" || cleanUnion == "object" || strings.HasPrefix(cleanUnion, "__shape_") || strings.Contains(cleanUnion, shapeName) {
						return shapeName
					}
				}
			}
		}
	}
	for shapeName := range shapes {
		if strings.Contains(strings.ToLower(shapeName), cleanTarget) {
			if cleanUnion != "" && cleanUnion != "object" && strings.Contains(cleanUnion, shapeName) {
				return shapeName
			}
		}
	}
	return ""
}
