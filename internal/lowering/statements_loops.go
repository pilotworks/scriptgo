package lowering

import (
	"fmt"
	"maps"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerWhile(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	loopFinallyScopeStack = append(loopFinallyScopeStack, len(activeReturnFinallyStack))
	defer func() {
		loopFinallyScopeStack = loopFinallyScopeStack[:len(loopFinallyScopeStack)-1]
	}()
	condFunc := ir.Function{Name: "cond", ReturnType: ir.TypeBool}
	condVal, condType, err := lowerExpression(path, statement.Expression, "", &condFunc, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	condVal, err = coerceToBool(path, condVal, condType, &condFunc, counter, statement.Span)
	if err != nil {
		return fmt.Errorf("while condition: %w", err)
	}
	bodyInstructions, err := lowerBranch(path, statement.Body, function.ReturnType, env, function, counter, shapes, signatures)
	if err != nil {
		return err
	}
	var stepInstructions []ir.Instruction
	if len(statement.Step) > 0 {
		stepInstructions, err = lowerBranch(path, statement.Step, function.ReturnType, env, function, counter, shapes, signatures)
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

func lowerDoWhile(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	loopFinallyScopeStack = append(loopFinallyScopeStack, len(activeReturnFinallyStack))
	defer func() {
		loopFinallyScopeStack = loopFinallyScopeStack[:len(loopFinallyScopeStack)-1]
	}()
	condFunc := ir.Function{Name: "cond", ReturnType: ir.TypeBool}
	condVal, condType, err := lowerExpression(path, statement.Expression, "", &condFunc, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	condVal, err = coerceToBool(path, condVal, condType, &condFunc, counter, statement.Span)
	if err != nil {
		return fmt.Errorf("do-while condition: %w", err)
	}
	bodyInstructions, err := lowerBranch(path, statement.Body, function.ReturnType, env, function, counter, shapes, signatures)
	if err != nil {
		return err
	}
	var stepInstructions []ir.Instruction
	if len(statement.Step) > 0 {
		stepInstructions, err = lowerBranch(path, statement.Step, function.ReturnType, env, function, counter, shapes, signatures)
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

func lowerForIn(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	loopFinallyScopeStack = append(loopFinallyScopeStack, len(activeReturnFinallyStack))
	defer func() {
		loopFinallyScopeStack = loopFinallyScopeStack[:len(loopFinallyScopeStack)-1]
	}()
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
			if strings.HasPrefix(shapeName, "Record_") || strings.HasPrefix(shapeName, "Record__") || strings.HasPrefix(shapeName, "Record<") || shapeName == "Record" || strings.HasPrefix(shapeName, "Partial_") || strings.HasPrefix(shapeName, "Partial<") {
				shape = ir.ObjectShape{Name: shapeName, Fields: nil}
			} else {
				return fmt.Errorf("unknown shape %q for for...in", shapeName)
			}
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

func substituteStringIndex(expr *frontend.SyntaxExpression, varName string, stringVal string) *frontend.SyntaxExpression {
	if expr == nil {
		return nil
	}
	copy := *expr
	if (copy.Kind == "index" || copy.Kind == "optional_index") && copy.Right != nil && copy.Right.Kind == "identifier" && copy.Right.Text == varName {
		copy.Right = &frontend.SyntaxExpression{
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
		newArgs := make([]*frontend.SyntaxExpression, len(copy.Arguments))
		for i, a := range copy.Arguments {
			newArgs[i] = substituteStringIndex(a, varName, stringVal)
		}
		copy.Arguments = newArgs
	}
	return &copy
}

func substituteStringIndexInStmt(stmt frontend.SyntaxStatement, varName string, stringVal string) frontend.SyntaxStatement {
	copy := stmt
	if copy.Expression != nil {
		copy.Expression = substituteStringIndex(copy.Expression, varName, stringVal)
	}
	if len(copy.Body) > 0 {
		newBody := make([]frontend.SyntaxStatement, len(copy.Body))
		for i, s := range copy.Body {
			newBody[i] = substituteStringIndexInStmt(s, varName, stringVal)
		}
		copy.Body = newBody
	}
	return copy
}

func lowerLabel(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	bodyInstructions, err := lowerBranch(path, statement.Body, function.ReturnType, env, function, counter, shapes, signatures)
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
