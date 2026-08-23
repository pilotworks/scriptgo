package lowering

import (
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerIntlNew(
	path string,
	expression *typescriptgo.SyntaxExpression,
	className string,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, bool, error) {
	if !strings.HasPrefix(className, "Intl.") {
		return "", "", false, nil
	}

	intlType := strings.TrimPrefix(className, "Intl.")
	var args []string
	for _, arg := range expression.Arguments {
		val, _, err := lowerExpression(path, arg, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		args = append(args, val)
	}

	if result == "" {
		result = nextTemp(counter)
	}

	resType := ir.Type("object:Intl." + intlType)
	callee := "__intl." + strings.ToLower(intlType) + "_new"
	switch intlType {
	case "NumberFormat":
		callee = "__intl.number_format_new"
	case "DateTimeFormat":
		callee = "__intl.date_time_format_new"
	case "Collator":
		callee = "__intl.collator_new"
	case "Segmenter":
		callee = "__intl.segmenter_new"
	case "DisplayNames":
		callee = "__intl.display_names_new"
	case "ListFormat":
		callee = "__intl.list_format_new"
	case "RelativeTimeFormat":
		callee = "__intl.relative_time_format_new"
	case "PluralRules":
		callee = "__intl.plural_rules_new"
	}

	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   resType,
		Result: result,
		Callee: callee,
		Args:   args,
		Span:   toIRSpan(path, expression.Span),
	})
	return result, resType, true, nil
}

func lowerIntlReceiverMethod(
	path string,
	expression *typescriptgo.SyntaxExpression,
	receiver string,
	methodName string,
	receiverType ir.Type,
	result string,
	function *ir.Function,
	env map[string]ir.Type,
	counter *int,
	shapes map[string]ir.ObjectShape,
	signatures map[string]ir.Function,
) (string, ir.Type, bool, error) {
	if !strings.HasPrefix(string(receiverType), "object:Intl.") {
		return "", "", false, nil
	}

	subType := strings.TrimPrefix(string(receiverType), "object:Intl.")
	args := []string{receiver}
	for _, arg := range expression.Arguments {
		val, _, err := lowerExpression(path, arg, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		args = append(args, val)
	}

	if result == "" {
		result = nextTemp(counter)
	}

	switch subType {
	case "NumberFormat":
		if methodName == "format" {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: result,
				Callee: "__intl.number_format_format",
				Args:   args,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeString, true, nil
		}
	case "DateTimeFormat":
		if methodName == "format" {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: result,
				Callee: "__intl.date_time_format_format",
				Args:   args,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeString, true, nil
		}
	case "Collator":
		if methodName == "compare" {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeNumber,
				Result: result,
				Callee: "__intl.collator_compare",
				Args:   args,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeNumber, true, nil
		}
	case "Segmenter":
		if methodName == "segment" {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeStringArray,
				Result: result,
				Callee: "__intl.segmenter_segment",
				Args:   args,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeStringArray, true, nil
		}
	case "DisplayNames":
		if methodName == "of" {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: result,
				Callee: "__intl.display_names_of",
				Args:   args,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeString, true, nil
		}
	case "ListFormat":
		if methodName == "format" {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: result,
				Callee: "__intl.list_format_format",
				Args:   args,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeString, true, nil
		}
	case "RelativeTimeFormat":
		if methodName == "format" {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: result,
				Callee: "__intl.relative_time_format_format",
				Args:   args,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeString, true, nil
		}
	case "PluralRules":
		if methodName == "select" {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeString,
				Result: result,
				Callee: "__intl.plural_rules_select",
				Args:   args,
				Span:   toIRSpan(path, expression.Span),
			})
			return result, ir.TypeString, true, nil
		}
	}

	if methodName == "resolvedOptions" {
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeObject,
			Result: result,
			Callee: "__intl." + strings.ToLower(subType) + "_resolved_options",
			Args:   args,
			Span:   toIRSpan(path, expression.Span),
		})
		return result, ir.TypeObject, true, nil
	}

	return "", "", false, nil
}
