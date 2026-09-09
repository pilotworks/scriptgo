package lowering

import (
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
)

// normalizeDynamicAnyProgram maps TypeScript's unchecked any type onto the
// existing boxed unknown representation for Dynamic lowering. The frontend
// program is cloned so compatibility diagnostics retain the original types.
func normalizeDynamicAnyProgram(program frontend.Program) frontend.Program {
	program = cloneCompatibilityProgram(program)
	for i := range program.Files {
		for j := range program.Files[i].Symbols {
			program.Files[i].Symbols[j].Type = normalizeDynamicAnyType(program.Files[i].Symbols[j].Type)
		}
		program.Files[i].Syntax.Statements = normalizeDynamicAnyStatements(program.Files[i].Syntax.Statements)
	}
	return program
}

func normalizeDynamicAnyType(typ string) string {
	if typ == "" {
		return typ
	}
	var out strings.Builder
	for i := 0; i < len(typ); {
		if isTypeWordByte(typ[i]) {
			start := i
			i++
			for i < len(typ) && isTypeWordByte(typ[i]) {
				i++
			}
			word := typ[start:i]
			if strings.EqualFold(word, "any") || strings.EqualFold(word, "anykeyword") || strings.EqualFold(word, "kindanykeyword") {
				word = "unknown"
			}
			out.WriteString(word)
			continue
		}
		out.WriteByte(typ[i])
		i++
	}
	return out.String()
}

func isTypeWordByte(b byte) bool {
	return b == '_' || b == '$' || b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

func normalizeDynamicAnyStatements(statements []typescriptgo.SyntaxStatement) []typescriptgo.SyntaxStatement {
	for i := range statements {
		statement := &statements[i]
		statement.Type = normalizeDynamicAnyType(statement.Type)
		statement.InferredType = normalizeDynamicAnyType(statement.InferredType)
		statement.CatchVarType = normalizeDynamicAnyType(statement.CatchVarType)
		statement.Expression = normalizeDynamicAnyExpression(statement.Expression)
		statement.Left = normalizeDynamicAnyExpression(statement.Left)
		statement.Right = normalizeDynamicAnyExpression(statement.Right)
		statement.Parameters = normalizeDynamicAnyParameters(statement.Parameters)
		statement.Body = normalizeDynamicAnyStatements(statement.Body)
		statement.Step = normalizeDynamicAnyStatements(statement.Step)
		statement.Then = normalizeDynamicAnyStatements(statement.Then)
		statement.Else = normalizeDynamicAnyStatements(statement.Else)
		statement.Catch = normalizeDynamicAnyStatements(statement.Catch)
		statement.Finally = normalizeDynamicAnyStatements(statement.Finally)
		for j := range statement.Cases {
			statement.Cases[j].Expression = normalizeDynamicAnyExpression(statement.Cases[j].Expression)
			statement.Cases[j].Statements = normalizeDynamicAnyStatements(statement.Cases[j].Statements)
		}
		statement.Class = normalizeDynamicAnyClass(statement.Class)
		if statement.Enum != nil {
			for j := range statement.Enum.Members {
				statement.Enum.Members[j].Initializer = normalizeDynamicAnyExpression(statement.Enum.Members[j].Initializer)
			}
		}
	}
	return statements
}

func normalizeDynamicAnyParameters(parameters []typescriptgo.SyntaxParameter) []typescriptgo.SyntaxParameter {
	for i := range parameters {
		parameter := &parameters[i]
		parameter.Type = normalizeDynamicAnyType(parameter.Type)
		parameter.InferredType = normalizeDynamicAnyType(parameter.InferredType)
		parameter.Initializer = normalizeDynamicAnyExpression(parameter.Initializer)
		parameter.Decorators = normalizeDynamicAnyDecorators(parameter.Decorators)
	}
	return parameters
}

func normalizeDynamicAnyClass(class *typescriptgo.SyntaxClass) *typescriptgo.SyntaxClass {
	if class == nil {
		return nil
	}
	for i := range class.Decorators {
		class.Decorators[i] = normalizeDynamicAnyDecorator(class.Decorators[i])
	}
	for i := range class.Fields {
		field := &class.Fields[i]
		field.Type = normalizeDynamicAnyType(field.Type)
		field.InferredType = normalizeDynamicAnyType(field.InferredType)
		field.Initializer = normalizeDynamicAnyExpression(field.Initializer)
		field.Decorators = normalizeDynamicAnyDecorators(field.Decorators)
	}
	if class.Constructor != nil {
		class.Constructor.Parameters = normalizeDynamicAnyParameters(class.Constructor.Parameters)
		class.Constructor.Body = normalizeDynamicAnyStatements(class.Constructor.Body)
	}
	for i := range class.Methods {
		method := &class.Methods[i]
		method.Type = normalizeDynamicAnyType(method.Type)
		method.InferredType = normalizeDynamicAnyType(method.InferredType)
		method.Parameters = normalizeDynamicAnyParameters(method.Parameters)
		method.Decorators = normalizeDynamicAnyDecorators(method.Decorators)
		method.Body = normalizeDynamicAnyStatements(method.Body)
	}
	for i := range class.StaticBlocks {
		class.StaticBlocks[i] = normalizeDynamicAnyStatements(class.StaticBlocks[i])
	}
	for i := range class.StaticElements {
		element := &class.StaticElements[i]
		if element.Field != nil {
			element.Field.Type = normalizeDynamicAnyType(element.Field.Type)
			element.Field.InferredType = normalizeDynamicAnyType(element.Field.InferredType)
			element.Field.Initializer = normalizeDynamicAnyExpression(element.Field.Initializer)
			element.Field.Decorators = normalizeDynamicAnyDecorators(element.Field.Decorators)
		}
		element.Statements = normalizeDynamicAnyStatements(element.Statements)
	}
	return class
}

func normalizeDynamicAnyDecorators(decorators []typescriptgo.SyntaxDecorator) []typescriptgo.SyntaxDecorator {
	for i := range decorators {
		decorators[i] = normalizeDynamicAnyDecorator(decorators[i])
	}
	return decorators
}

func normalizeDynamicAnyDecorator(decorator typescriptgo.SyntaxDecorator) typescriptgo.SyntaxDecorator {
	decorator.DesignType = normalizeDynamicAnyType(decorator.DesignType)
	decorator.ReturnType = normalizeDynamicAnyType(decorator.ReturnType)
	for i := range decorator.ParamTypes {
		decorator.ParamTypes[i] = normalizeDynamicAnyType(decorator.ParamTypes[i])
	}
	decorator.Expression = normalizeDynamicAnyExpression(decorator.Expression)
	for i := range decorator.Arguments {
		decorator.Arguments[i] = normalizeDynamicAnyExpression(decorator.Arguments[i])
	}
	return decorator
}

func normalizeDynamicAnyExpression(expression *typescriptgo.SyntaxExpression) *typescriptgo.SyntaxExpression {
	if expression == nil {
		return nil
	}
	expression.InferredType = normalizeDynamicAnyType(expression.InferredType)
	if expression.Kind == "as" || expression.Kind == "cast" || expression.Kind == "type_assertion" {
		expression.Text = normalizeDynamicAnyType(expression.Text)
	}
	for i := range expression.TypeArguments {
		expression.TypeArguments[i] = normalizeDynamicAnyType(expression.TypeArguments[i])
	}
	expression.Left = normalizeDynamicAnyExpression(expression.Left)
	expression.Right = normalizeDynamicAnyExpression(expression.Right)
	expression.WhenTrue = normalizeDynamicAnyExpression(expression.WhenTrue)
	expression.WhenFalse = normalizeDynamicAnyExpression(expression.WhenFalse)
	for i := range expression.Arguments {
		expression.Arguments[i] = normalizeDynamicAnyExpression(expression.Arguments[i])
	}
	if expression.Function != nil {
		function := normalizeDynamicAnyStatements([]typescriptgo.SyntaxStatement{*expression.Function})
		expression.Function = &function[0]
	}
	return expression
}
