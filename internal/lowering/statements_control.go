package lowering

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func coerceToBool(path string, value string, valType ir.Type, function *ir.Function, counter *int, span frontend.SourceSpan) (string, error) {
	if valType == ir.TypeBool {
		return value, nil
	}
	if valType == ir.TypeNumber {
		zeroConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, span)})
		isNonZero := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: isNonZero, Operator: "!=", Args: []string{value, zeroConst}, Span: toIRSpan(path, span)})
		isNotNaN := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: isNotNaN, Operator: "==", Args: []string{value, value}, Span: toIRSpan(path, span)})
		boolRes := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: boolRes, Operator: "&&", Args: []string{isNonZero, isNotNaN}, Span: toIRSpan(path, span)})
		return boolRes, nil
	}
	if valType == ir.TypeString {
		nullConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: nullConst, Value: "null", Span: toIRSpan(path, span)})
		nonNull := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: nonNull, Operator: "!=", Args: []string{value, nullConst}, Span: toIRSpan(path, span)})

		undefConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: undefConst, Value: "undefined", Span: toIRSpan(path, span)})
		nonUndef := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: nonUndef, Operator: "!=", Args: []string{value, undefConst}, Span: toIRSpan(path, span)})

		nonNullish := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: nonNullish, Operator: "&&", Args: []string{nonNull, nonUndef}, Span: toIRSpan(path, span)})

		strLen := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeNumber,
			Result: strLen,
			Callee: "__string.length",
			Args:   []string{value},
			Span:   toIRSpan(path, span),
		})
		zeroConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, span)})
		nonEmpty := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: nonEmpty, Operator: ">", Args: []string{strLen, zeroConst}, Span: toIRSpan(path, span)})

		boolRes := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: boolRes, Operator: "&&", Args: []string{nonNullish, nonEmpty}, Span: toIRSpan(path, span)})
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
		nonNull := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: nonNull, Operator: "!=", Args: []string{value, nullConst}, Span: toIRSpan(path, span)})

		undefConst := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: valType, Result: undefConst, Value: "undefined", Span: toIRSpan(path, span)})
		nonUndef := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: nonUndef, Operator: "!=", Args: []string{value, undefConst}, Span: toIRSpan(path, span)})

		boolRes := nextTemp(counter)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: boolRes, Operator: "&&", Args: []string{nonNull, nonUndef}, Span: toIRSpan(path, span)})
		return boolRes, nil
	}
	return "", fmt.Errorf("cannot coerce %s to boolean condition", valType)
}

func lowerIf(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
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

	applyConditionNarrowing(statement.Expression, thenEnv, elseEnv, env, shapes)

	thenBody, err := lowerBranch(path, statement.Then, function.ReturnType, thenEnv, function, counter, shapes, signatures)
	if err != nil {
		return err
	}
	elseBody, err := lowerBranch(path, statement.Else, function.ReturnType, elseEnv, function, counter, shapes, signatures)
	if err != nil {
		return err
	}
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpIf, Type: ir.TypeVoid, Args: []string{condition}, Then: thenBody, Else: elseBody, Span: toIRSpan(path, statement.Span)})
	if len(statement.Else) == 0 && slices.ContainsFunc(statement.Then, statementAlwaysReturns) {
		maps.Copy(env, elseEnv)
	}
	return nil
}

func lowerSwitch(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	loopFinallyScopeStack = append(loopFinallyScopeStack, len(activeReturnFinallyStack))
	defer func() {
		loopFinallyScopeStack = loopFinallyScopeStack[:len(loopFinallyScopeStack)-1]
	}()
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

	var hasDefault bool
	var cmpResults = make([]string, len(statement.Cases))
	var caseCmps []string

	for i, c := range statement.Cases {
		if c.Expression != nil {
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
			cmpResults[i] = cmpResult
			caseCmps = append(caseCmps, cmpResult)
		} else {
			hasDefault = true
		}
	}

	var shouldRunDefault string
	if hasDefault {
		if len(caseCmps) == 0 {
			trueVal := fmt.Sprintf("__def_true_%d", *counter)
			*counter++
			switchBranch.Body = append(switchBranch.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeBool,
				Result: trueVal,
				Value:  "true",
				Span:   toIRSpan(path, statement.Span),
			})
			switchEnv[trueVal] = ir.TypeBool
			shouldRunDefault = trueVal
		} else {
			anyMatched := caseCmps[0]
			for _, cmp := range caseCmps[1:] {
				orRes := fmt.Sprintf("__any_cmp_%d", *counter)
				*counter++
				switchBranch.Body = append(switchBranch.Body, ir.Instruction{
					Op:       ir.OpBinary,
					Type:     ir.TypeBool,
					Result:   orRes,
					Operator: "||",
					Args:     []string{anyMatched, cmp},
					Span:     toIRSpan(path, statement.Span),
				})
				switchEnv[orRes] = ir.TypeBool
				anyMatched = orRes
			}
			falseConst := fmt.Sprintf("__false_%d", *counter)
			*counter++
			switchBranch.Body = append(switchBranch.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   ir.TypeBool,
				Result: falseConst,
				Value:  "false",
				Span:   toIRSpan(path, statement.Span),
			})
			switchEnv[falseConst] = ir.TypeBool

			notAny := fmt.Sprintf("__not_any_%d", *counter)
			*counter++
			switchBranch.Body = append(switchBranch.Body, ir.Instruction{
				Op:       ir.OpCompare,
				Type:     ir.TypeBool,
				Result:   notAny,
				Operator: "==",
				Args:     []string{anyMatched, falseConst},
				Span:     toIRSpan(path, statement.Span),
			})
			switchEnv[notAny] = ir.TypeBool
			shouldRunDefault = notAny
		}
	}

	for i, c := range statement.Cases {
		if c.Expression != nil {
			cmpResult := cmpResults[i]
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
		} else {
			newMatched := fmt.Sprintf("__new_matched_%d", *counter)
			*counter++
			switchBranch.Body = append(switchBranch.Body, ir.Instruction{
				Op:       ir.OpBinary,
				Type:     ir.TypeBool,
				Result:   newMatched,
				Operator: "||",
				Args:     []string{matchedVar, shouldRunDefault},
				Span:     toIRSpan(path, c.Span),
			})
			switchBranch.Body = append(switchBranch.Body, ir.Instruction{
				Op:     ir.OpAssign,
				Type:   ir.TypeBool,
				Result: matchedVar,
				Args:   []string{newMatched},
				Span:   toIRSpan(path, c.Span),
			})
		}

		caseEnv := make(map[string]ir.Type, len(switchEnv))
		maps.Copy(caseEnv, switchEnv)
		if statement.Expression != nil && (statement.Expression.Kind == "property" || statement.Expression.Kind == "member") && statement.Expression.Left != nil && statement.Expression.Left.Kind == "identifier" && c.Expression != nil && (c.Expression.Kind == "string" || c.Expression.Kind == "literal") {
			varName := statement.Expression.Left.Text
			propName := statement.Expression.Text
			if statement.Expression.Right != nil && statement.Expression.Right.Text != "" {
				propName = statement.Expression.Right.Text
			}
			valStr := c.Expression.Text
			if currType, ok := switchEnv[varName]; ok {
				matched := findMatchingDiscriminatedType(propName, valStr, string(currType), shapes)
				if matched != "" {
					caseEnv[varName] = ir.Type("object:" + matched)
				}
			}
		}

		caseStmts, err := lowerBranch(path, c.Statements, function.ReturnType, caseEnv, function, counter, shapes, signatures)
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

var activeReturnFinallyStack [][]frontend.SyntaxStatement
var activeThrowFinallyStack [][]frontend.SyntaxStatement
var loopFinallyScopeStack []int

func lowerTry(path string, statement frontend.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	if len(statement.Finally) > 0 {
		activeReturnFinallyStack = append(activeReturnFinallyStack, statement.Finally)
	}

	bodyInstructions, err := lowerBranch(path, statement.Body, function.ReturnType, env, function, counter, shapes, signatures)
	if len(statement.Finally) > 0 {
		activeReturnFinallyStack = activeReturnFinallyStack[:len(activeReturnFinallyStack)-1]
	}
	if err != nil {
		return err
	}
	var catchInstructions []ir.Instruction
	if len(statement.Catch) > 0 {
		if len(statement.Finally) > 0 {
			activeReturnFinallyStack = append(activeReturnFinallyStack, statement.Finally)
			activeThrowFinallyStack = append(activeThrowFinallyStack, statement.Finally)
		}
		catchEnv := make(map[string]ir.Type, len(env)+1)
		maps.Copy(catchEnv, env)
		if statement.CatchVar != "" {
			if statement.CatchVarType != "" {
				catchEnv[statement.CatchVar] = toIRType(statement.CatchVarType)
			} else {
				catchEnv[statement.CatchVar] = ir.Type("object:Error")
			}
		}
		catchBranch := ir.Function{Name: "catch", ReturnType: function.ReturnType}
		for _, catchStmt := range statement.Catch {
			if err := lowerStatement(path, catchStmt, &catchBranch, catchEnv, counter, shapes, signatures); err != nil {
				if len(statement.Finally) > 0 {
					activeReturnFinallyStack = activeReturnFinallyStack[:len(activeReturnFinallyStack)-1]
					activeThrowFinallyStack = activeThrowFinallyStack[:len(activeThrowFinallyStack)-1]
				}
				return err
			}
		}
		if len(statement.Finally) > 0 {
			activeReturnFinallyStack = activeReturnFinallyStack[:len(activeReturnFinallyStack)-1]
			activeThrowFinallyStack = activeThrowFinallyStack[:len(activeThrowFinallyStack)-1]
		}
		catchInstructions = catchBranch.Body
		for _, local := range catchBranch.Locals {
			found := false
			for _, existing := range function.Locals {
				if existing.Name == local.Name {
					found = true
					break
				}
			}
			if !found {
				function.Locals = append(function.Locals, local)
			}
		}
	}
	var finallyInstructions []ir.Instruction
	if len(statement.Finally) > 0 {
		finallyBranch, err := lowerBranch(path, statement.Finally, function.ReturnType, env, function, counter, shapes, signatures)
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
