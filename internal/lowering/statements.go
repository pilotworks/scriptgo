package lowering

import (
	"fmt"

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
			typ = ir.TypeStringArray
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
		if bodyStatement.Kind == "return" {
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
	case "class":
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
		return ""
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
