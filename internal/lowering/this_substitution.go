package lowering

import "github.com/pilotworks/scriptgo/internal/frontend"

// replaceThisInStatement and replaceThisInExpression copy a syntax tree with
// every `this` whose ThisBinding is binding (including inside arrow
// functions) renamed to the identifier name. They do not descend into an
// object literal's methods, which bind their own `this`.
func replaceThisInStatement(statement frontend.SyntaxStatement, binding, name string) frontend.SyntaxStatement {
	statement.Expression = replaceThisInExpression(statement.Expression, binding, name)
	statement.Left = replaceThisInExpression(statement.Left, binding, name)
	statement.Right = replaceThisInExpression(statement.Right, binding, name)
	statement.Body = replaceThisInStatements(statement.Body, binding, name)
	statement.Step = replaceThisInStatements(statement.Step, binding, name)
	statement.Then = replaceThisInStatements(statement.Then, binding, name)
	statement.Else = replaceThisInStatements(statement.Else, binding, name)
	statement.Catch = replaceThisInStatements(statement.Catch, binding, name)
	statement.Finally = replaceThisInStatements(statement.Finally, binding, name)
	if statement.Cases != nil {
		cases := make([]frontend.SyntaxSwitchCase, len(statement.Cases))
		for index, c := range statement.Cases {
			c.Expression = replaceThisInExpression(c.Expression, binding, name)
			c.Statements = replaceThisInStatements(c.Statements, binding, name)
			cases[index] = c
		}
		statement.Cases = cases
	}
	return statement
}

func replaceThisInStatements(statements []frontend.SyntaxStatement, binding, name string) []frontend.SyntaxStatement {
	if statements == nil {
		return nil
	}
	result := make([]frontend.SyntaxStatement, len(statements))
	for index, statement := range statements {
		result[index] = replaceThisInStatement(statement, binding, name)
	}
	return result
}

func replaceThisInExpression(expression *frontend.SyntaxExpression, binding, name string) *frontend.SyntaxExpression {
	if expression == nil {
		return nil
	}
	copied := *expression
	if (copied.Kind == "identifier" || copied.Kind == "this") && copied.Text == "this" {
		if copied.ThisBinding == binding {
			copied.Kind = "identifier"
			copied.Text = name
		}
		return &copied
	}
	if isObjectLiteralMethod(expression) {
		return &copied
	}
	copied.Left = replaceThisInExpression(copied.Left, binding, name)
	copied.Right = replaceThisInExpression(copied.Right, binding, name)
	copied.WhenTrue = replaceThisInExpression(copied.WhenTrue, binding, name)
	copied.WhenFalse = replaceThisInExpression(copied.WhenFalse, binding, name)
	if copied.Arguments != nil {
		arguments := make([]*frontend.SyntaxExpression, len(copied.Arguments))
		for index, argument := range copied.Arguments {
			arguments[index] = replaceThisInExpression(argument, binding, name)
		}
		copied.Arguments = arguments
	}
	if copied.Function != nil {
		function := replaceThisInStatement(*copied.Function, binding, name)
		copied.Function = &function
	}
	return &copied
}

// isObjectLiteralMethod reports a method in an object literal, `m() {}`: the
// adapter lowers it to a named function value. A named function expression
// also matches; both bind their own `this`.
func isObjectLiteralMethod(expression *frontend.SyntaxExpression) bool {
	return expression.Kind == "arrow_function" && expression.Function != nil && expression.Function.Name != ""
}

// usesThis reports whether statements contain a `this` with the given
// binding outside nested object literal methods.
func usesThis(statements []frontend.SyntaxStatement, binding string) bool {
	for _, statement := range statements {
		for _, expression := range []*frontend.SyntaxExpression{statement.Expression, statement.Left, statement.Right} {
			if expressionUsesThis(expression, binding) {
				return true
			}
		}
		for _, list := range [][]frontend.SyntaxStatement{statement.Body, statement.Step, statement.Then, statement.Else, statement.Catch, statement.Finally} {
			if usesThis(list, binding) {
				return true
			}
		}
		for _, c := range statement.Cases {
			if expressionUsesThis(c.Expression, binding) || usesThis(c.Statements, binding) {
				return true
			}
		}
	}
	return false
}

func expressionUsesThis(expression *frontend.SyntaxExpression, binding string) bool {
	if expression == nil {
		return false
	}
	if (expression.Kind == "identifier" || expression.Kind == "this") && expression.Text == "this" {
		return expression.ThisBinding == binding
	}
	if isObjectLiteralMethod(expression) {
		return false
	}
	for _, nested := range append([]*frontend.SyntaxExpression{expression.Left, expression.Right, expression.WhenTrue, expression.WhenFalse}, expression.Arguments...) {
		if expressionUsesThis(nested, binding) {
			return true
		}
	}
	return expression.Function != nil && usesThis([]frontend.SyntaxStatement{*expression.Function}, binding)
}
