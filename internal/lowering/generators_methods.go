package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
)

// desugarGeneratorMethods turns each instance generator method of a class
// (`*items() { ... }`, `*[Symbol.iterator]() { ... }`) into a module-level
// generator function that takes the receiver as its first parameter, and the
// method into a call of it: generators are lowered as functions, and `this`
// inside the body becomes that parameter.
func desugarGeneratorMethods(program frontend.Program) frontend.Program {
	files := make([]frontend.SourceFile, len(program.Files))
	for fileIndex, file := range program.Files {
		files[fileIndex] = file
		var added []frontend.SyntaxStatement
		statements := make([]frontend.SyntaxStatement, len(file.Syntax.Statements))
		for index, statement := range file.Syntax.Statements {
			statements[index] = statement
			if statement.Class == nil || statement.Kind == "interface" || statement.Kind == "type_alias" {
				continue
			}
			class := *statement.Class
			methods := make([]frontend.SyntaxMethod, len(class.Methods))
			copy(methods, class.Methods)
			changed := false
			for methodIndex, method := range methods {
				if !method.IsGenerator || method.IsStatic || method.IsAbstract || len(method.Body) == 0 {
					continue
				}
				generator, call := generatorMethodFunction(class.Name, method)
				added = append(added, generator)
				method.IsGenerator = false
				method.IsAsync = false
				method.Body = []frontend.SyntaxStatement{{Span: method.Span, Kind: "return", Expression: call}}
				methods[methodIndex] = method
				changed = true
			}
			if changed {
				class.Methods = methods
				statement.Class = &class
				statements[index] = statement
			}
		}
		if len(added) > 0 {
			syntax := file.Syntax
			syntax.Statements = append(statements, added...)
			files[fileIndex].Syntax = syntax
		}
	}
	program.Files = files
	return program
}

// generatorMethodFunction is the generator function for a generator method
// and the call expression that replaces the method's body.
func generatorMethodFunction(className string, method frontend.SyntaxMethod) (frontend.SyntaxStatement, *frontend.SyntaxExpression) {
	name := "__generator_" + className + "_" + strings.NewReplacer(".", "_", "[", "_", "]", "_").Replace(method.Name)
	kind := "generator_function"
	if method.IsAsync {
		kind = "async_generator_function"
	}
	parameters := append([]frontend.SyntaxParameter{{Span: method.Span, Name: generatorSelf, Type: className, InferredType: className}}, method.Parameters...)
	body := make([]frontend.SyntaxStatement, len(method.Body))
	for index, statement := range method.Body {
		body[index] = replaceThisInStatement(statement)
	}
	generator := frontend.SyntaxStatement{
		Span:           method.Span,
		Kind:           kind,
		Name:           name,
		Type:           method.Type,
		InferredType:   method.InferredType,
		TypeParameters: method.TypeParameters,
		Parameters:     parameters,
		Body:           body,
		IsGenerator:    true,
		IsAsync:        method.IsAsync,
	}
	arguments := []*frontend.SyntaxExpression{{Span: method.Span, Kind: "identifier", Text: "this", InferredType: className}}
	for _, parameter := range method.Parameters {
		arguments = append(arguments, &frontend.SyntaxExpression{Span: parameter.Span, Kind: "identifier", Text: parameter.Name, InferredType: parameter.Type})
	}
	call := &frontend.SyntaxExpression{
		Span:         method.Span,
		Kind:         "call",
		Left:         &frontend.SyntaxExpression{Span: method.Span, Kind: "identifier", Text: name},
		Arguments:    arguments,
		InferredType: method.Type,
	}
	return generator, call
}

const generatorSelf = "__self"

// replaceThisInStatement and replaceThisInExpression copy a syntax tree with
// every lexical `this` (including inside arrow functions) renamed to
// generatorSelf. A `this` bound by its own call site (ThisBinding
// "caller") is left alone.
func replaceThisInStatement(statement frontend.SyntaxStatement) frontend.SyntaxStatement {
	statement.Expression = replaceThisInExpression(statement.Expression)
	statement.Left = replaceThisInExpression(statement.Left)
	statement.Right = replaceThisInExpression(statement.Right)
	statement.Body = replaceThisInStatements(statement.Body)
	statement.Step = replaceThisInStatements(statement.Step)
	statement.Then = replaceThisInStatements(statement.Then)
	statement.Else = replaceThisInStatements(statement.Else)
	statement.Catch = replaceThisInStatements(statement.Catch)
	statement.Finally = replaceThisInStatements(statement.Finally)
	if statement.Cases != nil {
		cases := make([]frontend.SyntaxSwitchCase, len(statement.Cases))
		for index, c := range statement.Cases {
			c.Expression = replaceThisInExpression(c.Expression)
			c.Statements = replaceThisInStatements(c.Statements)
			cases[index] = c
		}
		statement.Cases = cases
	}
	return statement
}

func replaceThisInStatements(statements []frontend.SyntaxStatement) []frontend.SyntaxStatement {
	if statements == nil {
		return nil
	}
	result := make([]frontend.SyntaxStatement, len(statements))
	for index, statement := range statements {
		result[index] = replaceThisInStatement(statement)
	}
	return result
}

func replaceThisInExpression(expression *frontend.SyntaxExpression) *frontend.SyntaxExpression {
	if expression == nil {
		return nil
	}
	copied := *expression
	if (copied.Kind == "identifier" || copied.Kind == "this") && copied.Text == "this" && copied.ThisBinding != "caller" {
		copied.Kind = "identifier"
		copied.Text = generatorSelf
		return &copied
	}
	copied.Left = replaceThisInExpression(copied.Left)
	copied.Right = replaceThisInExpression(copied.Right)
	copied.WhenTrue = replaceThisInExpression(copied.WhenTrue)
	copied.WhenFalse = replaceThisInExpression(copied.WhenFalse)
	if copied.Arguments != nil {
		arguments := make([]*frontend.SyntaxExpression, len(copied.Arguments))
		for index, argument := range copied.Arguments {
			arguments[index] = replaceThisInExpression(argument)
		}
		copied.Arguments = arguments
	}
	if copied.Function != nil {
		function := replaceThisInStatement(*copied.Function)
		copied.Function = &function
	}
	return &copied
}
