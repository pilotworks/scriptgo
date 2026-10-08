package lowering

import (
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerDynamicRecordProperty(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, counter *int, object string) (string, ir.Type, error) {
	propNameConst := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: propNameConst,
		Value:  expression.Text,
		Span:   toIRSpan(path, expression.Span),
	})
	if result == "" {
		result = nextTemp(counter)
	}
	retType := ir.TypeUnknown
	if expression.InferredType != "" {
		if inferred := toIRType(expression.InferredType); inferred != "" {
			retType = inferred
		}
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   retType,
		Result: result,
		Callee: "__object.get_prop",
		Args:   []string{object, propNameConst},
		Span:   toIRSpan(path, expression.Span),
	})
	return result, retType, nil

}

func lowerAnonymousShapeProperty(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, counter *int, object string, objectType ir.Type) (string, ir.Type, error) {
	propNameConst := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   ir.TypeString,
		Result: propNameConst,
		Value:  expression.Text,
		Span:   toIRSpan(path, expression.Span),
	})
	if result == "" {
		result = nextTemp(counter)
	}
	retType := ir.TypeUnknown
	if expression.InferredType != "" {
		if inferred := toIRType(expression.InferredType); inferred != "" {
			retType = inferred
		}
	}
	if expression.Kind == "optional_property" {
		initVal := "undefined"
		switch retType {
		case ir.TypeBool:
			initVal = "false"
		case ir.TypeNumber:
			initVal = "NaN"
		}
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   retType,
			Result: result,
			Value:  initVal,
			Span:   toIRSpan(path, expression.Span),
		})
		cond, err := coerceToBool(path, object, objectType, function, counter, expression.Span)
		if err == nil {
			thenFn := &ir.Function{}
			thenFn.Body = append(thenFn.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   retType,
				Result: result,
				Callee: "__object.get_prop",
				Args:   []string{object, propNameConst},
				Span:   toIRSpan(path, expression.Span),
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:   ir.OpIf,
				Type: ir.TypeVoid,
				Args: []string{cond},
				Then: thenFn.Body,
				Span: toIRSpan(path, expression.Span),
			})
			return result, retType, nil
		}
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   retType,
		Result: result,
		Callee: "__object.get_prop",
		Args:   []string{object, propNameConst},
		Span:   toIRSpan(path, expression.Span),
	})
	return result, retType, nil

}

func lowerOptionalPropertyFallback(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, counter *int) (string, ir.Type, error) {
	if result == "" {
		result = nextTemp(counter)
	}
	retType := ir.TypeUnknown
	valStr := "undefined"
	if expression.InferredType != "" {
		inferred := toIRType(expression.InferredType)
		if inferred != "" && inferred != ir.TypeNumber && inferred != ir.TypeBool {
			retType = inferred
		}
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   retType,
		Result: result,
		Value:  valStr,
		Span:   toIRSpan(path, expression.Span),
	})
	return result, retType, nil

}
