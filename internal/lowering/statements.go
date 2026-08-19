package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerFunction(path string, statement typescriptgo.SyntaxStatement, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (ir.Function, error) {
	function := ir.Function{Name: statement.Name, Span: toIRSpan(path, statement.Span), ReturnType: toIRType(statement.Type)}
	if function.ReturnType == "" {
		function.ReturnType = ir.TypeVoid
	}
	env := map[string]ir.Type{}
	for _, parameter := range statement.Parameters {
		typ := toIRType(parameter.Type)
		if parameter.Rest {
			if parameter.Type == "number[]" {
				typ = ir.TypeNumberArray
			} else {
				typ = ir.TypeStringArray
			}
		}
		if typ == "" {
			return ir.Function{}, fmt.Errorf("parameter %q has unsupported type %q", parameter.Name, parameter.Type)
		}
		function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: typ})
		env[parameter.Name] = typ
	}
	counter := 0
	returned := false
	for _, bodyStatement := range statement.Body {
		if err := lowerStatement(path, bodyStatement, &function, env, &counter, shapes, signatures); err != nil {
			return ir.Function{}, sourceError(path, bodyStatement.Span, err)
		}
		if statementAlwaysReturns(bodyStatement) {
			returned = true
		}
	}
	if !returned {
		if function.ReturnType != ir.TypeVoid {
			return ir.Function{}, fmt.Errorf("function %q does not return %s", function.Name, function.ReturnType)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: function.Span})
	}
	return function, nil
}

func lowerStatement(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	switch statement.Kind {
	case "variable":
		if statement.Expression == nil {
			return fmt.Errorf("variable %q has no initializer", statement.Name)
		}
		if statement.Expression.Kind == "identifier" {
			srcType, ok := env[statement.Expression.Text]
			if !ok {
				global, isGlobal := builtinGlobal(statement.Expression.Text)
				if isGlobal {
					env[statement.Name] = global.Type
					function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: global.Type, Result: statement.Name, Value: global.Value, Span: toIRSpan(path, statement.Span)})
					return nil
				}
				return fmt.Errorf("unknown identifier %q", statement.Expression.Text)
			}
			env[statement.Name] = srcType
			switch srcType {
			case ir.TypeNumber:
				zeroConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: zeroConst, Value: "0", Span: toIRSpan(path, statement.Span)})
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: statement.Name, Operator: "+", Args: []string{statement.Expression.Text, zeroConst}, Span: toIRSpan(path, statement.Span)})
			case ir.TypeString:
				emptyStr := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeString, Result: emptyStr, Value: "", Span: toIRSpan(path, statement.Span)})
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: statement.Name, Operator: "+", Args: []string{statement.Expression.Text, emptyStr}, Span: toIRSpan(path, statement.Span)})
			case ir.TypeBool:
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: srcType, Result: statement.Name, Operator: "||", Args: []string{statement.Expression.Text, statement.Expression.Text}, Span: toIRSpan(path, statement.Span)})
			default:
				trueConst := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: trueConst, Value: "true", Span: toIRSpan(path, statement.Span)})
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: srcType, Result: statement.Name, Args: []string{trueConst, statement.Expression.Text, statement.Expression.Text}, Span: toIRSpan(path, statement.Span)})
			}
			return nil
		}
		value, typ, err := lowerExpression(path, statement.Expression, statement.Name, function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if value != statement.Name {
			return fmt.Errorf("variable %q produced unnamed value %q", statement.Name, value)
		}
		env[statement.Name] = typ
	case "expression":
		if statement.Expression == nil {
			return fmt.Errorf("empty expression")
		}
		_, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		return err
	case "return":
		if statement.Expression == nil {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: toIRSpan(path, statement.Span)})
			return nil
		}
		value, typ, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: typ, Args: []string{value}, Span: toIRSpan(path, statement.Span)})
	case "block":
		for _, s := range statement.Body {
			if err := lowerStatement(path, s, function, env, counter, shapes, signatures); err != nil {
				return err
			}
		}
	case "assign":
		varType, ok := env[statement.Name]
		if !ok {
			return fmt.Errorf("assignment to unknown variable %q", statement.Name)
		}
		value, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if valType != varType {
			return fmt.Errorf("assignment type mismatch for %q: %s := %s", statement.Name, varType, valType)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpAssign, Type: varType, Result: statement.Name, Args: []string{value}, Span: toIRSpan(path, statement.Span)})
	case "while":
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
		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpWhile,
			Type: ir.TypeVoid,
			Args: []string{condVal},
			Cond: condFunc.Body,
			Body: bodyInstructions,
			Span: toIRSpan(path, statement.Span),
		})
	case "if":
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
	case "break":
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBreak, Type: ir.TypeVoid, Span: toIRSpan(path, statement.Span)})
	case "continue":
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpContinue, Type: ir.TypeVoid, Span: toIRSpan(path, statement.Span)})
	case "dowhile":
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
		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpDoWhile,
			Type: ir.TypeVoid,
			Args: []string{condVal},
			Cond: condFunc.Body,
			Body: bodyInstructions,
			Span: toIRSpan(path, statement.Span),
		})
	case "forof":
		arrVal, arrType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		elemType := ir.TypeNumber
		if arrType == ir.TypeStringArray {
			elemType = ir.TypeString
		}
		idxName := fmt.Sprintf("__i_%d", *counter)
		lenName := fmt.Sprintf("__len_%d", *counter)
		*counter++
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: idxName, Value: "0", Span: toIRSpan(path, statement.Span)})
		env[idxName] = ir.TypeNumber
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: lenName, Callee: "__array.length", Args: []string{arrVal}, Span: toIRSpan(path, statement.Span)})
		env[lenName] = ir.TypeNumber

		condFunc := ir.Function{Name: "cond", ReturnType: ir.TypeBool}
		condCmp := fmt.Sprintf("__cmp_%d", *counter)
		*counter++
		condFunc.Body = append(condFunc.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: condCmp, Operator: "<", Args: []string{idxName, lenName}, Span: toIRSpan(path, statement.Span)})

		bodyEnv := make(map[string]ir.Type, len(env)+2)
		for k, v := range env {
			bodyEnv[k] = v
		}
		bodyBranch := ir.Function{Name: "body", ReturnType: function.ReturnType}
		bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpIndex, Type: elemType, Result: statement.Name, Args: []string{arrVal, idxName}, Span: toIRSpan(path, statement.Span)})
		bodyEnv[statement.Name] = elemType

		for _, bodyStmt := range statement.Body {
			if err := lowerStatement(path, bodyStmt, &bodyBranch, bodyEnv, counter, shapes, signatures); err != nil {
				return err
			}
		}
		incVal := fmt.Sprintf("__inc_%d", *counter)
		oneVal := fmt.Sprintf("__one_%d", *counter)
		*counter++
		bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: oneVal, Value: "1", Span: toIRSpan(path, statement.Span)})
		bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeNumber, Operator: "+", Result: incVal, Args: []string{idxName, oneVal}, Span: toIRSpan(path, statement.Span)})
		bodyBranch.Body = append(bodyBranch.Body, ir.Instruction{Op: ir.OpAssign, Type: ir.TypeNumber, Result: idxName, Args: []string{incVal}, Span: toIRSpan(path, statement.Span)})

		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpWhile,
			Type: ir.TypeVoid,
			Args: []string{condCmp},
			Cond: condFunc.Body,
			Body: bodyBranch.Body,
			Span: toIRSpan(path, statement.Span),
		})
	case "index_set":
		arrVal, arrType, err := lowerExpression(path, statement.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		idxVal, _, err := lowerExpression(path, statement.Right, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		val, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if (arrType == ir.TypeNumberArray && valType != ir.TypeNumber) || (arrType == ir.TypeStringArray && valType != ir.TypeString) {
			return fmt.Errorf("array index_set type mismatch: %s cannot be assigned to %s", valType, arrType)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpIndexSet,
			Type: ir.TypeVoid,
			Args: []string{arrVal, idxVal, val},
			Span: toIRSpan(path, statement.Span),
		})
	case "field_set":
		objVal, objType, err := lowerExpression(path, statement.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		shapeName := strings.TrimPrefix(string(objType), string(ir.TypeObject)+":")
		shape, ok := shapes[shapeName]
		if !ok {
			return fmt.Errorf("field set on unknown object shape %q", shapeName)
		}
		fIndex := fieldIndex(shape, statement.Name)
		if fIndex < 0 {
			return fmt.Errorf("unknown field %q on object shape %q", statement.Name, shapeName)
		}
		val, valType, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		if valType != shape.Fields[fIndex].Type {
			return fmt.Errorf("field set type mismatch for %q: %s := %s", statement.Name, shape.Fields[fIndex].Type, valType)
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:         ir.OpFieldSet,
			Type:       ir.TypeVoid,
			Field:      statement.Name,
			FieldIndex: fIndex,
			Args:       []string{objVal, val},
			Span:       toIRSpan(path, statement.Span),
		})
	case "class":
		return nil
	case "throw":
		val, _, err := lowerExpression(path, statement.Expression, "", function, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpThrow,
			Type: ir.TypeVoid,
			Args: []string{val},
			Span: toIRSpan(path, statement.Span),
		})
	case "try":
		bodyInstructions, err := lowerBranch(path, statement.Body, function.ReturnType, env, counter, shapes, signatures)
		if err != nil {
			return err
		}
		var catchInstructions []ir.Instruction
		if len(statement.Catch) > 0 {
			catchEnv := make(map[string]ir.Type, len(env)+1)
			for k, v := range env {
				catchEnv[k] = v
			}
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
	case "module", "enum":
		return nil
	default:
		return fmt.Errorf("unsupported statement %q", statement.Kind)
	}
	return nil
}

func lowerBranch(path string, statements []typescriptgo.SyntaxStatement, returnType ir.Type, parentEnv map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) ([]ir.Instruction, error) {
	branch := ir.Function{Name: "branch", ReturnType: returnType}
	env := make(map[string]ir.Type, len(parentEnv))
	for name, typ := range parentEnv {
		env[name] = typ
	}
	for _, statement := range statements {
		if err := lowerStatement(path, statement, &branch, env, counter, shapes, signatures); err != nil {
			return nil, err
		}
	}
	return branch.Body, nil
}

func toIRType(value string) ir.Type {
	if strings.HasPrefix(value, "object:") {
		return ir.Type(value)
	}
	switch value {
	case "number":
		return ir.TypeNumber
	case "string":
		return ir.TypeString
	case "bool":
		return ir.TypeBool
	case "number[]":
		return ir.TypeNumberArray
	case "string[]":
		return ir.TypeStringArray
	case "void", "":
		return ir.TypeVoid
	default:
		return ir.Type("object:" + value)
	}
}

func toIRSpan(path string, span typescriptgo.SourceSpan) ir.SourceSpan {
	return ir.SourceSpan{Path: path, Offset: span.Start, Length: span.Length}
}

func fieldIndex(shape ir.ObjectShape, name string) int {
	for index, field := range shape.Fields {
		if field.Name == name {
			return index
		}
	}
	return -1
}

func sourceError(path string, span typescriptgo.SourceSpan, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s:%d+%d: %w", path, span.Start, span.Length, err)
}

func statementAlwaysReturns(stmt typescriptgo.SyntaxStatement) bool {
	switch stmt.Kind {
	case "return", "throw":
		return true
	case "block":
		for _, s := range stmt.Body {
			if statementAlwaysReturns(s) {
				return true
			}
		}
		return false
	case "if":
		if len(stmt.Then) == 0 || len(stmt.Else) == 0 {
			return false
		}
		thenReturns := false
		for _, s := range stmt.Then {
			if statementAlwaysReturns(s) {
				thenReturns = true
				break
			}
		}
		elseReturns := false
		for _, s := range stmt.Else {
			if statementAlwaysReturns(s) {
				elseReturns = true
				break
			}
		}
		return thenReturns && elseReturns
	default:
		return false
	}
}

