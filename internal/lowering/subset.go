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
	case "property":
		if expression.Left != nil && expression.Left.Kind == "identifier" {
			return nil
		}
		return subsetError(fileName, expression.Span, "nested property access")
	case "binary":
		if err := validateExpression(fileName, expression.Left); err != nil {
			return err
		}
		return validateExpression(fileName, expression.Right)
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
