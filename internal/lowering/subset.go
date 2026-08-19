package lowering

import (
	"fmt"

	"github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
)

// ValidateSubset rejects checked syntax that the current native IR cannot
// represent. Keeping this gate in lowering prevents backend-specific policy.
func ValidateSubset(program frontend.Program) error {
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if err := validateStatement(file.FileName, statement); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateStatement(fileName string, statement typescriptgo.SyntaxStatement) error {
	switch statement.Kind {
	case "module":
		return nil
	case "class":
		if statement.Class == nil || len(statement.Class.Fields) == 0 {
			return subsetError(fileName, statement.Span, "empty class shape")
		}
		for _, field := range statement.Class.Fields {
			if field.Type != "number" && field.Type != "string" {
				return subsetError(fileName, field.Span, "class field type "+field.Type)
			}
			if field.Initializer != nil {
				if err := validateExpression(fileName, field.Initializer); err != nil {
					return err
				}
				if field.Initializer == nil {
					return subsetError(fileName, field.Span, "class field without initializer")
				}
				if field.Initializer.Kind != "number" && field.Initializer.Kind != "string" && field.Initializer.Kind != "bool" {
					return subsetError(fileName, field.Span, "non-literal class field initializer")
				}
			}
		}
		return nil
	case "variable":
		if statement.Expression == nil {
			return subsetError(fileName, statement.Span, "variable declaration without an initializer")
		}
		return validateExpression(fileName, statement.Expression)
	case "expression", "return":
		if statement.Expression == nil {
			return nil
		}
		return validateExpression(fileName, statement.Expression)
	case "function":
		for _, bodyStatement := range statement.Body {
			if err := validateStatement(fileName, bodyStatement); err != nil {
				return err
			}
		}
		return nil
	case "if":
		if err := validateExpression(fileName, statement.Expression); err != nil {
			return err
		}
		for _, branchStatement := range append(statement.Then, statement.Else...) {
			if err := validateStatement(fileName, branchStatement); err != nil {
				return err
			}
		}
		if len(statement.Then) == 0 || statement.Then[len(statement.Then)-1].Kind != "return" {
			return subsetError(fileName, statement.Span, "IfStatement branch without return")
		}
		if len(statement.Else) > 0 && statement.Else[len(statement.Else)-1].Kind != "return" {
			return subsetError(fileName, statement.Span, "IfStatement else branch without return")
		}
		return nil
	case "unsupported":
		return subsetError(fileName, statement.Span, statement.Type)
	default:
		return subsetError(fileName, statement.Span, statement.Kind)
	}
}

func validateExpression(fileName string, expression *typescriptgo.SyntaxExpression) error {
	switch expression.Kind {
	case "number", "string", "bool", "identifier":
		return nil
	case "array":
		if len(expression.Arguments) == 0 {
			return subsetError(fileName, expression.Span, "empty array literal")
		}
		for _, element := range expression.Arguments {
			if err := validateExpression(fileName, element); err != nil {
				return err
			}
		}
		return nil
	case "index":
		if err := validateExpression(fileName, expression.Left); err != nil {
			return err
		}
		return validateExpression(fileName, expression.Right)
	case "property":
		if expression.Left != nil && expression.Left.Kind == "identifier" {
			return nil
		}
		return subsetError(fileName, expression.Span, "nested property access")
	case "new":
		for _, argument := range expression.Arguments {
			if err := validateExpression(fileName, argument); err != nil {
				return err
			}
		}
		if callName(expression.Left) == "" {
			return subsetError(fileName, expression.Span, "dynamic constructor target")
		}
		if len(expression.Arguments) != 0 {
			return subsetError(fileName, expression.Span, "class constructors with arguments")
		}
		return nil
	case "binary":
		if err := validateExpression(fileName, expression.Left); err != nil {
			return err
		}
		return validateExpression(fileName, expression.Right)
	case "conditional":
		if err := validateExpression(fileName, expression.Left); err != nil {
			return err
		}
		if err := validateExpression(fileName, expression.WhenTrue); err != nil {
			return err
		}
		return validateExpression(fileName, expression.WhenFalse)
	case "call":
		for _, argument := range expression.Arguments {
			if err := validateExpression(fileName, argument); err != nil {
				return err
			}
		}
		if callName(expression.Left) == "" {
			return subsetError(fileName, expression.Span, "dynamic call target")
		}
		return nil
	case "unsupported":
		return subsetError(fileName, expression.Span, expression.Text)
	default:
		return subsetError(fileName, expression.Span, expression.Kind)
	}
}

func subsetError(fileName string, span typescriptgo.SourceSpan, feature string) error {
	return fmt.Errorf("native subset: %s at offset %d (length %d): unsupported feature %q", fileName, span.Start, span.Length, feature)
}
