package lowering

import (
	"maps"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerLogicalAndExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, bool, error) {
	leftVal, leftTyp, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", true, err
	}
	// Logical operators return one of their operands. The boolean-only
	// short-circuit path is valid only when the whole expression is known to
	// be boolean; otherwise `true && value` must preserve `value`'s type.
	if leftTyp == ir.TypeBool && logicalResultIsBool(expression) {
		res := result
		if res == "" {
			res = nextTemp(counter)
		}
		env[res] = ir.TypeBool
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeBool,
			Result: res,
			Value:  "false",
			Span:   toIRSpan(path, expression.Span),
		})
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   ir.TypeBool,
			Result: res,
			Args:   []string{leftVal},
			Span:   toIRSpan(path, expression.Span),
		})
		thenBlock := ir.Function{Name: "then", ReturnType: function.ReturnType}
		thenEnv := make(map[string]ir.Type, len(env))
		maps.Copy(thenEnv, env)
		rightVal, _, err := lowerExpression(path, expression.Right, "", &thenBlock, thenEnv, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		thenBlock.Body = append(thenBlock.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   ir.TypeBool,
			Result: res,
			Args:   []string{rightVal},
			Span:   toIRSpan(path, expression.Span),
		})
		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpIf,
			Type: ir.TypeVoid,
			Args: []string{leftVal},
			Then: thenBlock.Body,
			Span: toIRSpan(path, expression.Span),
		})
		return res, ir.TypeBool, true, nil
	}
	if leftTyp == ir.TypeVoid {
		// An undefined left operand is always falsy, so && returns it
		// without evaluating the right operand.
		return leftVal, leftTyp, true, nil
	}

	return "", "", false, nil
}

func lowerLogicalOrExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, bool, error) {
	leftVal, leftTyp, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", true, err
	}
	if leftTyp == ir.TypeBool && logicalResultIsBool(expression) {
		res := result
		if res == "" {
			res = nextTemp(counter)
		}
		env[res] = ir.TypeBool
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpConst,
			Type:   ir.TypeBool,
			Result: res,
			Value:  "false",
			Span:   toIRSpan(path, expression.Span),
		})
		function.Body = append(function.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   ir.TypeBool,
			Result: res,
			Args:   []string{leftVal},
			Span:   toIRSpan(path, expression.Span),
		})
		elseBlock := ir.Function{Name: "else", ReturnType: function.ReturnType}
		elseEnv := make(map[string]ir.Type, len(env))
		maps.Copy(elseEnv, env)
		rightVal, _, err := lowerExpression(path, expression.Right, "", &elseBlock, elseEnv, counter, shapes, signatures)
		if err != nil {
			return "", "", true, err
		}
		elseBlock.Body = append(elseBlock.Body, ir.Instruction{
			Op:     ir.OpAssign,
			Type:   ir.TypeBool,
			Result: res,
			Args:   []string{rightVal},
			Span:   toIRSpan(path, expression.Span),
		})
		function.Body = append(function.Body, ir.Instruction{
			Op:   ir.OpIf,
			Type: ir.TypeVoid,
			Args: []string{leftVal},
			Else: elseBlock.Body,
			Span: toIRSpan(path, expression.Span),
		})
		return res, ir.TypeBool, true, nil
	}
	if leftTyp == ir.TypeVoid {
		// An undefined left operand is always falsy, so || evaluates and
		// returns the right operand.
		{
			v0, v1, v2 := lowerExpression(path, expression.Right, result, function, env, counter, shapes, signatures)
			return v0, v1, true, v2
		}
	}

	return "", "", false, nil
}

func lowerNullishCoalescingExpression(path string, expression *frontend.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if expression.Left != nil && (expression.Left.Kind == "null" || expression.Left.Kind == "undefined") {
		return lowerExpression(path, expression.Right, result, function, env, counter, shapes, signatures)
	}
	if expression.Right != nil && (expression.Right.Kind == "null" || expression.Right.Kind == "undefined") {
		return lowerExpression(path, expression.Left, result, function, env, counter, shapes, signatures)
	}
	leftVal, leftTyp, err := lowerExpression(path, expression.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	if leftTyp == ir.TypePointer || leftTyp == "ptr" || leftTyp == ir.TypeVoid {
		return lowerExpression(path, expression.Right, result, function, env, counter, shapes, signatures)
	}
	outTyp := leftTyp
	if expression.InferredType != "" {
		infIR := toIRType(expression.InferredType)
		if infIR != "" && infIR != ir.TypeVoid {
			outTyp = infIR
		}
	}
	res := result
	if res == "" || outTyp == ir.TypeUnknown {
		res = nextTemp(counter)
	}
	env[res] = outTyp

	initLeft := leftVal
	if outTyp == ir.TypeUnknown && leftTyp != ir.TypeUnknown {
		boxed := nextTemp(counter)
		env[boxed] = ir.TypeUnknown
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{leftVal}, Span: toIRSpan(path, expression.Span)})
		initLeft = boxed
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCheckedCast,
		Type:   outTyp,
		Result: res,
		Args:   []string{initLeft},
		Span:   toIRSpan(path, expression.Span),
	})

	var cond string
	if leftTyp == ir.TypeNumber {
		cmpNaN := nextTemp(counter)
		env[cmpNaN] = ir.TypeBool
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpNaN, Operator: "==", Args: []string{leftVal, leftVal}, Span: toIRSpan(path, expression.Span)})
		cond = cmpNaN
	} else if leftTyp == ir.TypeBool {
		trueConst := nextTemp(counter)
		env[trueConst] = ir.TypeBool
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: trueConst, Value: "true", Span: toIRSpan(path, expression.Span)})
		cond = trueConst
	} else if leftTyp == ir.TypeUnknown {
		nullConst := nextTemp(counter)
		env[nullConst] = ir.TypeUnknown
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeUnknown, Result: nullConst, Value: "null", Span: toIRSpan(path, expression.Span)})
		cmpNull := nextTemp(counter)
		env[cmpNull] = ir.TypeBool
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpNull, Operator: "!=", Args: []string{leftVal, nullConst}, Span: toIRSpan(path, expression.Span)})

		undefConst := nextTemp(counter)
		env[undefConst] = ir.TypeUnknown
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeUnknown, Result: undefConst, Value: "undefined", Span: toIRSpan(path, expression.Span)})
		cmpUndef := nextTemp(counter)
		env[cmpUndef] = ir.TypeBool
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpUndef, Operator: "!=", Args: []string{leftVal, undefConst}, Span: toIRSpan(path, expression.Span)})

		condTemp := nextTemp(counter)
		env[condTemp] = ir.TypeBool
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: condTemp, Operator: "&&", Args: []string{cmpNull, cmpUndef}, Span: toIRSpan(path, expression.Span)})
		cond = condTemp
	} else {
		nullConst := nextTemp(counter)
		env[nullConst] = leftTyp
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: leftTyp, Result: nullConst, Value: "null", Span: toIRSpan(path, expression.Span)})
		cmpNull := nextTemp(counter)
		env[cmpNull] = ir.TypeBool
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpNull, Operator: "!=", Args: []string{leftVal, nullConst}, Span: toIRSpan(path, expression.Span)})

		undefConst := nextTemp(counter)
		env[undefConst] = leftTyp
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: leftTyp, Result: undefConst, Value: "undefined", Span: toIRSpan(path, expression.Span)})
		cmpUndef := nextTemp(counter)
		env[cmpUndef] = ir.TypeBool
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: cmpUndef, Operator: "!=", Args: []string{leftVal, undefConst}, Span: toIRSpan(path, expression.Span)})

		condTemp := nextTemp(counter)
		env[condTemp] = ir.TypeBool
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeBool, Result: condTemp, Operator: "&&", Args: []string{cmpNull, cmpUndef}, Span: toIRSpan(path, expression.Span)})
		cond = condTemp
	}

	elseBlock := ir.Function{Name: "nullish_fallback", ReturnType: function.ReturnType}
	elseEnv := make(map[string]ir.Type, len(env))
	maps.Copy(elseEnv, env)
	if expression.Right != nil && (expression.Right.InferredType == "" || expression.Right.InferredType == "{}" || expression.Right.InferredType == "never[]" || expression.Right.InferredType == "unknown[]" || expression.Right.InferredType == "[]") {
		if strings.HasPrefix(string(outTyp), "object:") || strings.HasSuffix(string(outTyp), "[]") {
			expression.Right.InferredType = string(outTyp)
		}
	}
	rightVal, rightTyp, err := lowerExpression(path, expression.Right, "", &elseBlock, elseEnv, counter, shapes, signatures)
	if err != nil {
		return "", "", err
	}
	finalRight := rightVal
	if outTyp == ir.TypeUnknown && rightTyp != ir.TypeUnknown {
		boxed := nextTemp(counter)
		elseEnv[boxed] = ir.TypeUnknown
		elseBlock.Body = append(elseBlock.Body, ir.Instruction{Op: ir.OpBoxUnknown, Type: ir.TypeUnknown, Result: boxed, Args: []string{rightVal}, Span: toIRSpan(path, expression.Span)})
		finalRight = boxed
	} else if strings.HasPrefix(string(outTyp), "object:") && rightTyp != outTyp {
		castTemp := nextTemp(counter)
		elseEnv[castTemp] = outTyp
		elseBlock.Body = append(elseBlock.Body, ir.Instruction{
			Op:     ir.OpCheckedCast,
			Type:   outTyp,
			Result: castTemp,
			Args:   []string{rightVal},
			Span:   toIRSpan(path, expression.Span),
		})
		finalRight = castTemp
	}
	elseBlock.Body = append(elseBlock.Body, ir.Instruction{
		Op:     ir.OpAssign,
		Type:   outTyp,
		Result: res,
		Args:   []string{finalRight},
		Span:   toIRSpan(path, expression.Span),
	})

	function.Body = append(function.Body, ir.Instruction{
		Op:   ir.OpIf,
		Type: ir.TypeVoid,
		Args: []string{cond},
		Else: elseBlock.Body,
		Span: toIRSpan(path, expression.Span),
	})
	return res, outTyp, nil
}
