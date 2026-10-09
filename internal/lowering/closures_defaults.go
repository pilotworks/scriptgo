package lowering

import "github.com/pilotworks/scriptgo/internal/frontend"

// parameterDefaultStatement is `if (p$raw === undefined) p = <initializer>`
// for a closure parameter: the default applies exactly when the boxed
// argument is undefined (missing or passed explicitly).
func parameterDefaultStatement(parameter frontend.SyntaxParameter) frontend.SyntaxStatement {
	span := parameter.Span
	initializer := parameter.Initializer
	if declared := parameterTypeText(parameter); initializer.Kind == "array" && declared != "" {
		// An array literal default takes the parameter's type (a tuple for
		// an array pattern), as an initializer of a typed variable does.
		typed := *initializer
		typed.InferredType = declared
		initializer = &typed
	}
	return frontend.SyntaxStatement{
		Span: span,
		Kind: "if",
		Expression: &frontend.SyntaxExpression{
			Span:     span,
			Kind:     "binary",
			Operator: "===",
			Left:     &frontend.SyntaxExpression{Span: span, Kind: "identifier", Text: parameter.Name + "$raw", InferredType: "unknown"},
			Right:    &frontend.SyntaxExpression{Span: span, Kind: "undefined", Text: "undefined", InferredType: "undefined"},
		},
		Then: []frontend.SyntaxStatement{{
			Span:       span,
			Kind:       "assign",
			Name:       parameter.Name,
			Expression: initializer,
		}},
	}
}

func parameterTypeText(parameter frontend.SyntaxParameter) string {
	if parameter.Type != "" {
		return parameter.Type
	}
	return parameter.InferredType
}
