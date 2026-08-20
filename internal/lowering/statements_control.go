package lowering

import (
	"fmt"
	"maps"
	"strings"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerIf(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	condition, typ, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	if typ != ir.TypeBool {
		return fmt.Errorf("if condition must be bool")
	}
	thenBody, err := lowerBranch(path, statement.Then, function.ReturnType, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	elseBody, err := lowerBranch(path, statement.Else, function.ReturnType, env, counter, shapes, signatures)
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
	if condType != ir.TypeBool {
		return fmt.Errorf("while condition must be bool")
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
	if condType != ir.TypeBool {
		return fmt.Errorf("do-while condition must be bool")
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
	if strings.HasPrefix(string(arrType), "object:Generator_") || strings.HasPrefix(string(arrType), "object:Iterator_") {
		shapeName := strings.TrimPrefix(string(arrType), "object:")
		nextFn := shapeName + "_next"
		targetNext, hasNext := signatures[nextFn]
		if hasNext {
			resShapeName := strings.TrimPrefix(string(targetNext.ReturnType), "object:")
			resShape, ok := shapes[resShapeName]
			valType := ir.TypeNumber
			if ok && len(resShape.Fields) > 0 {
				valType = resShape.Fields[0].Type
			}

			condFunc := ir.Function{Name: "cond", ReturnType: ir.TypeBool}
			condConst := nextTemp(counter)
			condFunc.Body = append(condFunc.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: condConst, Value: "true", Span: toIRSpan(path, statement.Span)})

			bodyEnv := make(map[string]ir.Type, len(env)+2)
			maps.Copy(bodyEnv, env)
			bodyBranch := ir.Function{Name: "body", ReturnType: function.ReturnType}

			resVal := nextTemp(counter)
			bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   targetNext.ReturnType,
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
				FieldIndex: 1,
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
				FieldIndex: 0,
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
	}

	isString := (arrType == ir.TypeString)
	var elemType ir.Type
	if isString {
		elemType = ir.TypeString
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
				if err := lowerStatement(path, bodyStmt, function, fieldEnv, counter, shapes, signatures); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return fmt.Errorf("for...in requires object or array, got %s", objType)
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

		caseStmts, err := lowerBranch(path, c.Statements, function.ReturnType, switchEnv, counter, shapes, signatures)
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
		for _, stmt := range defStmts {
			switchBranch.Body = append(switchBranch.Body, stmt)
		}
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

func lowerTry(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	bodyInstructions, err := lowerBranch(path, statement.Body, function.ReturnType, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	var catchInstructions []ir.Instruction
	if len(statement.Catch) > 0 {
		catchEnv := make(map[string]ir.Type, len(env)+1)
		maps.Copy(catchEnv, env)
		if statement.CatchVar != "" {
			catchEnv[statement.CatchVar] = ir.TypeString
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
