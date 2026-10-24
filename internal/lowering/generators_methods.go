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
		body[index] = replaceThisInStatement(statement, "", generatorSelf)
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

		ParameterBindings: method.ParameterBindings,
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
