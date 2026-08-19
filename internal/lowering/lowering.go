// Package lowering converts the checked frontend model into backend-independent IR.
package lowering

import (
	"fmt"
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// Lower lowers the currently supported synchronous TypeScript subset.
func Lower(program frontend.Program) (ir.Module, error) {
	if err := ValidateSubset(program); err != nil {
		return ir.Module{}, err
	}
	module := ir.Module{SourcePath: program.EntryPath, SourceFiles: make(map[string]string), StatementCount: program.StatementCount}
	for _, file := range program.Files {
		module.SourceFiles[file.FileName] = file.Source
	}
	shapes := map[string]ir.ObjectShape{}
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if statement.Kind != "class" || statement.Class == nil {
				continue
			}
			shape := ir.ObjectShape{Name: statement.Class.Name, Span: toIRSpan(file.FileName, statement.Class.Span)}
			for _, field := range statement.Class.Fields {
				shape.Fields = append(shape.Fields, ir.Field{Name: field.Name, Type: toIRType(field.Type), Value: field.Initializer.Text, Span: toIRSpan(file.FileName, field.Span)})
			}
			shapes[shape.Name] = shape
			module.Shapes = append(module.Shapes, shape)
		}
	}

	main := ir.Function{Name: "main", ReturnType: ir.TypeVoid}
	env := map[string]ir.Type{}
	signatures := buildFunctionIndex(program)
	counter := 0
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if statement.Kind == "function" {
				function, err := lowerFunction(file.FileName, statement, shapes, signatures)
				if err != nil {
					return ir.Module{}, fmt.Errorf("lower function %q: %w", statement.Name, sourceError(file.FileName, statement.Span, err))
				}
				module.Functions = append(module.Functions, function)
				continue
			}
			if statement.Kind == "module" {
				continue
			}
			if statement.Kind == "class" {
				continue
			}
			if err := lowerStatement(file.FileName, statement, &main, env, &counter, shapes, signatures); err != nil {
				return ir.Module{}, fmt.Errorf("lower %s: %w", statement.Kind, sourceError(file.FileName, statement.Span, err))
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

func lowerExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
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
	case "array":
		if len(expression.Arguments) == 0 {
			return "", "", fmt.Errorf("empty array literal needs an explicit runtime representation")
		}
		if result == "" {
			result = nextTemp(counter)
		}
		arguments := make([]string, 0, len(expression.Arguments))
		for _, element := range expression.Arguments {
			value, typ, err := lowerExpression(path, element, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			if typ != ir.TypeNumber {
				return "", "", fmt.Errorf("array literal currently supports number elements only")
			}
			arguments = append(arguments, value)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: ir.TypeNumberArray, Result: result, Args: arguments, Span: toIRSpan(path, expression.Span)})
		return result, ir.TypeNumberArray, nil
	case "index":
		array, arrayType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		index, indexType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if (arrayType != ir.TypeNumberArray && arrayType != ir.TypeStringArray) || indexType != ir.TypeNumber {
			return "", "", fmt.Errorf("array indexing requires an array and number operands")
		}
		if result == "" {
			result = nextTemp(counter)
		}
		resultType := ir.TypeNumber
		if arrayType == ir.TypeStringArray {
			resultType = ir.TypeString
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpIndex, Type: resultType, Result: result, Args: []string{array, index}, Span: toIRSpan(path, expression.Span)})
		return result, resultType, nil
	case "identifier":
		typ, ok := env[expression.Text]
		if !ok {
			return "", "", fmt.Errorf("unknown identifier %q", expression.Text)
		}
		return expression.Text, typ, nil
	case "binary":
		left, leftType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		right, rightType, err := lowerExpression(path, expression.Right, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if leftType != rightType || (leftType != ir.TypeNumber && leftType != ir.TypeString) {
			return "", "", fmt.Errorf("operator %q does not support %s and %s", expression.Operator, leftType, rightType)
		}
		if isComparison(expression.Operator) {
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
			return result, ir.TypeBool, nil
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: leftType, Result: result, Operator: expression.Operator, Args: []string{left, right}, Span: toIRSpan(path, expression.Span)})
		return result, leftType, nil
	case "conditional":
		condition, conditionType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if conditionType != ir.TypeBool {
			return "", "", fmt.Errorf("conditional expression requires a bool condition")
		}
		whenTrue, trueType, err := lowerExpression(path, expression.WhenTrue, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		whenFalse, falseType, err := lowerExpression(path, expression.WhenFalse, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		if trueType != falseType {
			return "", "", fmt.Errorf("conditional branches must have the same type")
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpSelect, Type: trueType, Result: result, Args: []string{condition, whenTrue, whenFalse}, Span: toIRSpan(path, expression.Span)})
		return result, trueType, nil
	case "property":
		if expression.Left != nil && expression.Left.Kind == "identifier" {
			objectType, ok := env[expression.Left.Text]
			if ok && (objectType == ir.TypeString || objectType == ir.TypeNumberArray || objectType == ir.TypeStringArray) && expression.Text == "length" {
				if result == "" {
					result = nextTemp(counter)
				}
				callee := "__string.length"
				if objectType != ir.TypeString {
					callee = "__array.length"
				}
				function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: callee, Args: []string{expression.Left.Text}, Span: toIRSpan(path, expression.Span)})
				return result, ir.TypeNumber, nil
			}
		}
		object, objectType, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		className := strings.TrimPrefix(string(objectType), "object:")
		shape, ok := shapes[className]
		if !ok {
			return "", "", fmt.Errorf("unknown object shape %q", className)
		}
		for _, field := range shape.Fields {
			if field.Name != expression.Text {
				continue
			}
			if result == "" {
				result = nextTemp(counter)
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldGet, Type: field.Type, Result: result, Callee: className, Field: field.Name, FieldIndex: fieldIndex(shape, field.Name), Args: []string{object}, Span: toIRSpan(path, expression.Span)})
			return result, field.Type, nil
		}
		return "", "", fmt.Errorf("unknown field %q on object %q", expression.Text, className)
	case "new":
		className := callName(expression.Left)
		shape, ok := shapes[className]
		if !ok {
			return "", "", fmt.Errorf("unknown class %q", className)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpObjectNew, Type: ir.Type("object:" + className), Result: result, Callee: className, FieldCount: len(shape.Fields), Span: toIRSpan(path, expression.Span)})
		for _, field := range shape.Fields {
			initializer := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: field.Type, Result: initializer, Value: field.Value, Span: field.Span})
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: className, Field: field.Name, FieldIndex: fieldIndex(shape, field.Name), Args: []string{result, initializer}, Span: field.Span})
		}
		return result, ir.Type("object:" + className), nil
	case "call":
		if method := stringMethod(expression.Left); method != "" {
			receiver, receiverType, err := lowerExpression(path, expression.Left.Left, "", function, env, counter, shapes, signatures)
			if err != nil || receiverType != ir.TypeString {
				return "", "", fmt.Errorf("string method %q requires a string receiver", method)
			}
			args := []string{receiver}
			for _, argument := range expression.Arguments {
				value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
				if err != nil {
					return "", "", err
				}
				args = append(args, value)
			}
			if result == "" {
				result = nextTemp(counter)
			}
			returnType := ir.TypeNumber
			if method == "slice" {
				returnType = ir.TypeString
			}
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: "__string." + method, Args: args, Span: toIRSpan(path, expression.Span)})
			return result, returnType, nil
		}
		callee := callName(expression.Left)
		if callee == "console.log" {
			if len(expression.Arguments) != 1 {
				return "", "", fmt.Errorf("console.log requires one argument")
			}
			argument, _, err := lowerExpression(path, expression.Arguments[0], "", function, env, counter, shapes, signatures)
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
			value, _, err := lowerExpression(path, argument, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			args = append(args, value)
		}
		target, ok := signatures[callee]
		if !ok {
			return "", "", fmt.Errorf("unknown function %q", callee)
		}
		callee = target.Name
		if len(target.Parameters) > 0 && target.Parameters[len(target.Parameters)-1].Type == ir.TypeStringArray {
			fixed := len(target.Parameters) - 1
			if len(args) < fixed {
				return "", "", fmt.Errorf("call to %q has too few arguments", callee)
			}
			restArgs := append([]string(nil), args[fixed:]...)
			args = args[:fixed]
			array := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: ir.TypeStringArray, Result: array, Args: restArgs, Span: toIRSpan(path, expression.Span)})
			args = append(args, array)
		}
		if result == "" {
			result = nextTemp(counter)
		}
		returnType := target.ReturnType
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: returnType, Result: result, Callee: callee, Args: args, Span: toIRSpan(path, expression.Span)})
		return result, returnType, nil
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

func isComparison(operator string) bool {
	return operator == "==" || operator == "!==" || operator == "<" || operator == "<=" || operator == ">" || operator == ">="
}

func stringMethod(expression *typescriptgo.SyntaxExpression) string {
	if expression == nil || expression.Kind != "property" || expression.Left == nil {
		return ""
	}
	switch expression.Text {
	case "lastIndexOf", "slice":
		return expression.Text
	default:
		return ""
	}
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
