package lowering

import (
	"fmt"
	"maps"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerForOf(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	loopFinallyScopeStack = append(loopFinallyScopeStack, len(activeReturnFinallyStack))
	defer func() {
		loopFinallyScopeStack = loopFinallyScopeStack[:len(loopFinallyScopeStack)-1]
	}()
	arrVal, arrType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	isSegments := arrType == ir.Type("object:Intl.Segments")
	if isSegments {
		// Intl.Segments is iterable but is not an array-backed value.
		arrType = ir.Type("Intl.Segment[]")
	}
	isIteratorOrIterable := false
	shapeName := ""
	if strings.HasPrefix(string(arrType), "object:") {
		shapeName = strings.TrimPrefix(string(arrType), "object:")
		if _, _, ok := findMethodInHierarchy(shapeName, "next", signatures, classHierarchy); ok {
			isIteratorOrIterable = true
		} else if _, _, ok := findMethodInHierarchy(shapeName, "Symbol.iterator", signatures, classHierarchy); ok {
			isIteratorOrIterable = true
		}
	}
	if !isIteratorOrIterable && (strings.Contains(string(arrType), "Generator") || strings.Contains(string(arrType), "Iterator")) && !strings.Contains(string(arrType), "MapIterator") && !strings.Contains(string(arrType), "SetIterator") {
		isIteratorOrIterable = true
		shapeName = strings.TrimPrefix(string(arrType), "object:")
	}

	if isIteratorOrIterable {
		if fnIter, mangledIter, okIter := findMethodInHierarchy(shapeName, "Symbol.iterator", signatures, classHierarchy); okIter {
			iterRes := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   fnIter.ReturnType,
				Result: iterRes,
				Callee: mangledIter,
				Args:   []string{arrVal},
				Span:   toIRSpan(path, statement.Span),
			})
			arrVal = iterRes
			arrType = fnIter.ReturnType
			shapeName = strings.TrimPrefix(string(arrType), "object:")
		}

		nextFn := shapeName + "_next"
		targetNext, hasNext := signatures[nextFn]
		if !hasNext {
			if fn, mangled, ok := findMethodInHierarchy(shapeName, "next", signatures, classHierarchy); ok {
				targetNext = fn
				nextFn = mangled
				hasNext = true
			}
		}
		valType := ir.TypeNumber
		resShapeName := ""
		doneFieldIndex := 0
		valFieldIndex := 1
		if hasNext {
			resShapeName = strings.TrimPrefix(string(targetNext.ReturnType), "object:")
			if resShape, ok := shapes[resShapeName]; ok {
				for idx, f := range resShape.Fields {
					if f.Name == "done" {
						doneFieldIndex = idx
					} else if f.Name == "value" {
						valFieldIndex = idx
						valType = f.Type
					}
				}
			}
		} else {
			nextFn = "__generator.next"
			resShapeName = fmt.Sprintf("IteratorResult_%s", valType)
		}

		typeArg := ""
		if statement.Expression != nil && strings.Contains(statement.Expression.InferredType, "<") && strings.HasSuffix(strings.TrimSpace(statement.Expression.InferredType), ">") {
			inferred := strings.TrimSpace(statement.Expression.InferredType)
			idx := strings.Index(inferred, "<")
			inner := inferred[idx+1 : len(inferred)-1]
			parts := splitTypeArguments(inner)
			if len(parts) > 0 {
				typeArg = parts[0]
			}
		} else if strings.Contains(string(arrType), "<") && strings.HasSuffix(string(arrType), ">") {
			idx := strings.Index(string(arrType), "<")
			inner := string(arrType)[idx+1 : len(string(arrType))-1]
			parts := splitTypeArguments(inner)
			if len(parts) > 0 {
				typeArg = parts[0]
			}
		}
		if typeArg != "" {
			if specialized := toIRType(typeArg); specialized != "" {
				if valType == "object:T" || valType == "T" || strings.HasPrefix(string(valType), "object:T") || valType == ir.TypeUnknown || valType == ir.TypeNumber {
					valType = specialized
				}
			}
		}

		if !hasNext {
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
			FieldIndex: doneFieldIndex,
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
			FieldIndex: valFieldIndex,
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
	} else if isMapType(arrType) {
		keyType := ir.TypeString
		valType := ir.TypeUnknown
		if after, ok := strings.CutPrefix(string(arrType), "object:Map__"); ok {
			parts := strings.Split(after, "_")
			if len(parts) >= 2 {
				keyType = toIRType(parts[0])
				valType = toIRType(strings.Join(parts[1:], "_"))
			} else if len(parts) == 1 {
				keyType = toIRType(parts[0])
			}
		} else if statement.Expression != nil && strings.Contains(statement.Expression.InferredType, "<") {
			inferred := strings.TrimSpace(statement.Expression.InferredType)
			idx := strings.Index(inferred, "<")
			inner := inferred[idx+1 : len(inferred)-1]
			parts := splitTypeArguments(inner)
			if len(parts) >= 2 {
				keyType = toIRType(parts[0])
				valType = toIRType(parts[1])
			}
		}
		var fields []ir.Field
		fields = append(fields, ir.Field{Name: "0", Type: keyType})
		fields = append(fields, ir.Field{Name: "1", Type: valType})
		shapeName := anonymousShapeName(fields)
		registerAnonymousShape(shapeName, fields)
		elemType = ir.Type("object:" + shapeName)

		entriesRes := nextTemp(counter)
		entriesArrType := ir.Type(string(elemType) + "[]")
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   entriesArrType,
			Result: entriesRes,
			Callee: "__map.entries",
			Args:   []string{arrVal},
			Span:   toIRSpan(path, statement.Span),
		})
		env[entriesRes] = entriesArrType
		arrVal = entriesRes
	} else if isSetType(arrType) {
		elemType = ir.TypeUnknown
		if after, ok := strings.CutPrefix(string(arrType), "object:Set__"); ok {
			elemType = toIRType(after)
		} else if statement.Expression != nil && strings.Contains(statement.Expression.InferredType, "<") {
			inferred := strings.TrimSpace(statement.Expression.InferredType)
			idx := strings.Index(inferred, "<")
			inner := inferred[idx+1 : len(inferred)-1]
			parts := splitTypeArguments(inner)
			if len(parts) >= 1 {
				elemType = toIRType(parts[0])
			}
		}
		valuesRes := nextTemp(counter)
		valuesArrType := ir.Type(string(elemType) + "[]")
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   valuesArrType,
			Result: valuesRes,
			Callee: "__set.values",
			Args:   []string{arrVal},
			Span:   toIRSpan(path, statement.Span),
		})
		env[valuesRes] = valuesArrType
		arrVal = valuesRes
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
	if isSegments {
		elemType = ir.Type("object:Intl.Segment")
	}

	idxName := fmt.Sprintf("__i_%d", *counter)
	lenName := fmt.Sprintf("__len_%d", *counter)
	*counter++
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: idxName, Value: "0", Span: toIRSpan(path, statement.Span)})
	env[idxName] = ir.TypeNumber
	if isString {
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: lenName, Callee: "__string.length", Args: []string{arrVal}, Span: toIRSpan(path, statement.Span)})
	} else if isSegments {
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: lenName, Callee: "__intl.segments_length", Args: []string{arrVal}, Span: toIRSpan(path, statement.Span)})
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
		} else if isSegments {
			bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpCall, Type: elemType, Result: statement.Name, Callee: "__intl.segments_get", Args: []string{arrVal, idxName}, Span: toIRSpan(path, statement.Span)})
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
		} else if isSegments {
			bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpCall, Type: elemType, Result: statement.Name, Callee: "__intl.segments_get", Args: []string{arrVal, idxName}, Span: toIRSpan(path, statement.Span)})
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
