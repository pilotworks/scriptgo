package lowering

import (
	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
)

// Generic specialization mutates nested normalized nodes, so compatibility
// analysis works on a private copy and remains safe to run before lowering.
func cloneCompatibilityProgram(program frontend.Program) frontend.Program {
	cloned := program
	cloned.Diagnostics = append([]typescriptgo.Diagnostic(nil), program.Diagnostics...)
	cloned.Files = append([]typescriptgo.SourceFile(nil), program.Files...)
	for i := range cloned.Files {
		file := &cloned.Files[i]
		file.Imports = append([]typescriptgo.ModuleReference(nil), file.Imports...)
		for j := range file.Imports {
			file.Imports[j].Bindings = append([]typescriptgo.ModuleBinding(nil), file.Imports[j].Bindings...)
		}
		file.Symbols = append([]typescriptgo.Symbol(nil), file.Symbols...)
		file.Syntax.Statements = cloneCompatibilityStatements(file.Syntax.Statements)
	}
	return cloned
}

func cloneCompatibilityStatements(statements []typescriptgo.SyntaxStatement) []typescriptgo.SyntaxStatement {
	cloned := append([]typescriptgo.SyntaxStatement(nil), statements...)
	for i := range cloned {
		statement := &cloned[i]
		statement.TypeParameters = append([]string(nil), statement.TypeParameters...)
		statement.Expression = cloneCompatibilityExpression(statement.Expression)
		statement.Left = cloneCompatibilityExpression(statement.Left)
		statement.Right = cloneCompatibilityExpression(statement.Right)
		statement.Parameters = cloneCompatibilityParameters(statement.Parameters)
		statement.Body = cloneCompatibilityStatements(statement.Body)
		statement.Step = cloneCompatibilityStatements(statement.Step)
		statement.Then = cloneCompatibilityStatements(statement.Then)
		statement.Else = cloneCompatibilityStatements(statement.Else)
		statement.Catch = cloneCompatibilityStatements(statement.Catch)
		statement.Finally = cloneCompatibilityStatements(statement.Finally)
		statement.Cases = append([]typescriptgo.SyntaxSwitchCase(nil), statement.Cases...)
		for j := range statement.Cases {
			statement.Cases[j].Expression = cloneCompatibilityExpression(statement.Cases[j].Expression)
			statement.Cases[j].Statements = cloneCompatibilityStatements(statement.Cases[j].Statements)
		}
		statement.Class = cloneCompatibilityClass(statement.Class)
		if statement.Enum != nil {
			enum := *statement.Enum
			enum.Members = append([]typescriptgo.SyntaxEnumMember(nil), enum.Members...)
			for j := range enum.Members {
				enum.Members[j].Initializer = cloneCompatibilityExpression(enum.Members[j].Initializer)
			}
			statement.Enum = &enum
		}
	}
	return cloned
}

func cloneCompatibilityParameters(parameters []typescriptgo.SyntaxParameter) []typescriptgo.SyntaxParameter {
	cloned := append([]typescriptgo.SyntaxParameter(nil), parameters...)
	for i := range cloned {
		cloned[i].Initializer = cloneCompatibilityExpression(cloned[i].Initializer)
		cloned[i].Decorators = cloneCompatibilityDecorators(cloned[i].Decorators)
	}
	return cloned
}

func cloneCompatibilityClass(class *typescriptgo.SyntaxClass) *typescriptgo.SyntaxClass {
	if class == nil {
		return nil
	}
	cloned := *class
	cloned.TypeParameters = append([]string(nil), class.TypeParameters...)
	cloned.Implements = append([]string(nil), class.Implements...)
	cloned.Decorators = cloneCompatibilityDecorators(class.Decorators)
	cloned.Fields = append([]typescriptgo.SyntaxField(nil), class.Fields...)
	for i := range cloned.Fields {
		cloned.Fields[i].Initializer = cloneCompatibilityExpression(cloned.Fields[i].Initializer)
		cloned.Fields[i].Decorators = cloneCompatibilityDecorators(cloned.Fields[i].Decorators)
	}
	if class.Constructor != nil {
		constructor := *class.Constructor
		constructor.Parameters = cloneCompatibilityParameters(constructor.Parameters)
		constructor.Body = cloneCompatibilityStatements(constructor.Body)
		cloned.Constructor = &constructor
	}
	cloned.Methods = append([]typescriptgo.SyntaxMethod(nil), class.Methods...)
	for i := range cloned.Methods {
		method := &cloned.Methods[i]
		method.TypeParameters = append([]string(nil), method.TypeParameters...)
		method.Parameters = cloneCompatibilityParameters(method.Parameters)
		method.Decorators = cloneCompatibilityDecorators(method.Decorators)
		method.Body = cloneCompatibilityStatements(method.Body)
	}
	cloned.StaticBlocks = make([][]typescriptgo.SyntaxStatement, len(class.StaticBlocks))
	for i := range class.StaticBlocks {
		cloned.StaticBlocks[i] = cloneCompatibilityStatements(class.StaticBlocks[i])
	}
	cloned.StaticElements = append([]typescriptgo.SyntaxStaticElement(nil), class.StaticElements...)
	for i := range cloned.StaticElements {
		element := &cloned.StaticElements[i]
		if element.Field != nil {
			field := *element.Field
			field.Initializer = cloneCompatibilityExpression(field.Initializer)
			field.Decorators = cloneCompatibilityDecorators(field.Decorators)
			element.Field = &field
		}
		element.Statements = cloneCompatibilityStatements(element.Statements)
	}
	return &cloned
}

func cloneCompatibilityDecorators(decorators []typescriptgo.SyntaxDecorator) []typescriptgo.SyntaxDecorator {
	cloned := append([]typescriptgo.SyntaxDecorator(nil), decorators...)
	for i := range cloned {
		cloned[i].Expression = cloneCompatibilityExpression(cloned[i].Expression)
		cloned[i].Arguments = cloneCompatibilityExpressions(cloned[i].Arguments)
		cloned[i].ParamTypes = append([]string(nil), cloned[i].ParamTypes...)
	}
	return cloned
}

func cloneCompatibilityExpressions(expressions []*typescriptgo.SyntaxExpression) []*typescriptgo.SyntaxExpression {
	cloned := make([]*typescriptgo.SyntaxExpression, len(expressions))
	for i := range expressions {
		cloned[i] = cloneCompatibilityExpression(expressions[i])
	}
	return cloned
}

func cloneCompatibilityExpression(expression *typescriptgo.SyntaxExpression) *typescriptgo.SyntaxExpression {
	if expression == nil {
		return nil
	}
	cloned := *expression
	cloned.Left = cloneCompatibilityExpression(expression.Left)
	cloned.Right = cloneCompatibilityExpression(expression.Right)
	cloned.Arguments = cloneCompatibilityExpressions(expression.Arguments)
	cloned.TypeArguments = append([]string(nil), expression.TypeArguments...)
	cloned.WhenTrue = cloneCompatibilityExpression(expression.WhenTrue)
	cloned.WhenFalse = cloneCompatibilityExpression(expression.WhenFalse)
	if expression.Function != nil {
		function := cloneCompatibilityStatements([]typescriptgo.SyntaxStatement{*expression.Function})
		cloned.Function = &function[0]
	}
	return &cloned
}
