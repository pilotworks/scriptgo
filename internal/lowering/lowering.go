// Package lowering converts the checked frontend model into backend-independent IR.
package lowering

import (
	"fmt"
	"strconv"

	"github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// Lower lowers the currently supported synchronous TypeScript subset.
func Lower(program frontend.Program) (ir.Module, error) {
	if err := ValidateSubset(program); err != nil {
		return ir.Module{}, err
	}
	module := ir.Module{SourcePath: program.EntryPath, StatementCount: program.StatementCount}

	main := ir.Function{Name: "main", ReturnType: ir.TypeVoid}
	env := map[string]ir.Type{}
	counter := 0
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if statement.Kind == "function" {
				function, err := lowerFunction(file.FileName, statement)
				if err != nil {
					return ir.Module{}, fmt.Errorf("lower function %q: %w", statement.Name, err)
				}
				module.Functions = append(module.Functions, function)
				continue
			}
			if statement.Kind == "module" {
				continue
			}
			if err := lowerStatement(file.FileName, statement, &main, env, &counter); err != nil {
				return ir.Module{}, fmt.Errorf("lower %s: %w", statement.Kind, err)
			}
		}
	}
	main.Body = append(main.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid})
	module.Functions = append([]ir.Function{main}, module.Functions...)
	if err := module.Verify(); err != nil {
		return ir.Module{}, err
	}
	return module, nil
}

func lowerFunction(path string, statement typescriptgo.SyntaxStatement) (ir.Function, error) {
	function := ir.Function{Name: statement.Name, Span: toIRSpan(path, statement.Span), ReturnType: toIRType(statement.Type)}
	if function.ReturnType == "" {
		function.ReturnType = ir.TypeVoid
	}
	env := map[string]ir.Type{}
	for _, parameter := range statement.Parameters {
		typ := toIRType(parameter.Type)
		if typ == "" {
			return ir.Function{}, fmt.Errorf("parameter %q has unsupported type %q", parameter.Name, parameter.Type)
		}
		function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: typ})
		env[parameter.Name] = typ
	}
	counter := 0
	returned := false
	for _, bodyStatement := range statement.Body {
		if err := lowerStatement(path, bodyStatement, &function, env, &counter); err != nil {
			return ir.Function{}, err
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

func lowerStatement(path string, statement typescriptgo.SyntaxStatement, function *ir.Function, env map[string]ir.Type, counter *int) error {
	switch statement.Kind {
	case "variable":
		if statement.Expression == nil {
			return fmt.Errorf("variable %q has no initializer", statement.Name)
		}
		value, typ, err := lowerExpression(path, statement.Expression, statement.Name, function, env, counter)
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
		_, _, err := lowerExpression(path, statement.Expression, "", function, env, counter)
		return err
	case "return":
		if statement.Expression == nil {
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: toIRSpan(path, statement.Span)})
			return nil
		}
		value, typ, err := lowerExpression(path, statement.Expression, "", function, env, counter)
		if err != nil {
			return err
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: typ, Args: []string{value}, Span: toIRSpan(path, statement.Span)})
	default:
		return fmt.Errorf("unsupported statement %q", statement.Kind)
	}
	return nil
}

func lowerExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int) (string, ir.Type, error) {
	switch expression.Kind {
	case "number":
		typ := ir.TypeNumber
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "string":
		typ := ir.TypeString
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "bool":
		typ := ir.TypeBool
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: typ, Result: result, Value: expression.Text, Span: toIRSpan(path, expression.Span)})
		return result, typ, nil
	case "identifier":
		typ, ok := env[expression.Text]
		if !ok {
			return "", "", fmt.Errorf("unknown identifier %q", expression.Text)
		}
		return expression.Text, typ, nil
	case "binary":
		left, leftType, err := lowerExpression(path, expression.Left, "", function, env, counter)
		if err != nil {
			return "", "", err
		}
		right, rightType, err := lowerExpression(path, expression.Right, "", function, env, counter)
		if err != nil {
			return "", "", err
		}
		if leftType != rightType || (leftType != ir.TypeNumber && leftType != ir.TypeString) {
			return "", "", fmt.Errorf("operator %q does not support %s and %s", expression.Operator, leftType, rightType)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: leftType, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
		return result, leftType, nil
	case "property":
		return "", "", fmt.Errorf("property access %q is only supported as a call target", expression.Text)
	case "call":
		callee := callName(expression.Left)
		if callee == "console.log" {
			if len(expression.Arguments) != 1 {
				return "", "", fmt.Errorf("console.log requires one argument")
			}
			argument, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter)
			if err != nil {
				return "", "", err
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpPrint, Type: ir.TypeVoid, Callee: callee, Args: []string{argument}, Span: toIRSpan(path, expression.Span)})
			return "", ir.TypeVoid, nil
		}
		if callee == "" {
			return "", "", fmt.Errorf("unsupported call target")
		}
		args := make([]string, 0, len(expression.Arguments))
		for _, argument := range expression.Arguments {
			value, _, err := lowerExpression(path, argument, "", function, env, counter)
			if err != nil {
				return "", "", err
			}
			args = append(args, value)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: callee, Args: args, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeNumber, nil
	default:
		return "", "", fmt.Errorf("unsupported expression %q", expression.Kind)
	}
}

func callName(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil {
		return ""
	}
	if expression.Kind == "identifier" {
		return expression.Text
	}
	if expression.Kind == "property" && expression.Left != nil && expression.Left.Kind == "identifier" {
		return expression.Left.Text + "." + expression.Text
	}
	return ""
}

func nextTemp(counter *int) string {
	value := "t" + strconv.Itoa(*counter)
	*counter++
	return value
}

func toIRType(value string) ir.Type {
	switch value {
	case "number":
		return ir.TypeNumber
	case "string":
		return ir.TypeString
	case "bool":
		return ir.TypeBool
	case "void", "":
		return ir.TypeVoid
	default:
		return ""
	}
}

func toIRSpan(path string, span typescriptgo.SourceSpan) ir.SourceSpan {
	return ir.SourceSpan{Path: path, Offset: span.Start, Length: span.Length}
}
