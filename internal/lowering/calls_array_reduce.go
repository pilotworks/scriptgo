package lowering

import (
	"strconv"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerArrayReduce lowers arr.reduce(cb, init?) and reduceRight. The native
// runtime folds number[] with a number accumulator; every other accumulator
// type (strings, arrays, objects, mixed) is lowered as an explicit loop:
//
//	let acc = init (or the first visited element; empty arrays throw)
//	for each remaining index i in order: acc = cb(acc, arr[i], i, arr)
//
// The accumulator is stored as the call's checked result type.
func lowerArrayReduce(path string, expression *frontend.SyntaxExpression, receiver, methodName string, receiverType ir.Type, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, bool, error) {
	if methodName != "reduce" && methodName != "reduceRight" {
		return "", "", false, nil
	}
	if len(expression.Arguments) != 1 && len(expression.Arguments) != 2 {
		return "", "", false, nil
	}
	resultType := toIRType(expression.InferredType)
	if arrayElementType(receiverType) == ir.TypeNumber && resultType == ir.TypeNumber && reduceInitIsNumber(expression) {
		return "", "", false, nil
	}
	span := expression.Span
	bind := func(expr *frontend.SyntaxExpression) (string, ir.Type, error) {
		value, typ, err := lowerExpression(path, expr, "", function, env, counter, shapes, signatures)
		if err == nil {
			env[value] = typ
		}
		return value, typ, err
	}
	callback, callbackType, err := bind(expression.Arguments[0])
	if err != nil {
		return "", "", true, err
	}
	if !mayBeCallable(callbackType) {
		throwNotCallable(path, expression.Arguments[0], function, counter)
		if result == "" {
			result = nextTemp(counter)
		}
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeUnknown, Result: result, Value: "undefined", Span: toIRSpan(path, span)})
		return result, ir.TypeUnknown, true, nil
	}
	env[receiver] = receiverType
	ident := func(name string) *frontend.SyntaxExpression {
		return &frontend.SyntaxExpression{Span: span, Kind: "identifier", Text: name, InferredType: string(env[name])}
	}
	number := func(text string) *frontend.SyntaxExpression {
		return &frontend.SyntaxExpression{Span: span, Kind: "number", Text: text, InferredType: "number"}
	}
	binary := func(op string, left, right *frontend.SyntaxExpression) *frontend.SyntaxExpression {
		return &frontend.SyntaxExpression{Span: span, Kind: "binary", Operator: op, Left: left, Right: right}
	}
	length := &frontend.SyntaxExpression{Span: span, Kind: "property", Text: "length", Left: ident(receiver), InferredType: "number"}
	element := func(index *frontend.SyntaxExpression) *frontend.SyntaxExpression {
		return &frontend.SyntaxExpression{Span: span, Kind: "index", Left: ident(receiver), Right: index}
	}
	*counter++
	acc := "__reduce_acc_" + strconv.Itoa(*counter)
	index := "__reduce_i_" + strconv.Itoa(*counter)
	accType := expression.InferredType
	if resultType == "" || resultType == ir.TypeVoid {
		accType = "unknown"
	}
	right := methodName == "reduceRight"
	first := number("0")
	if right {
		first = binary("-", length, number("1"))
	}
	var statements []frontend.SyntaxStatement
	if len(expression.Arguments) == 2 {
		statements = append(statements,
			frontend.SyntaxStatement{Span: span, Kind: "variable", VarDeclKind: "let", Name: acc, Type: accType, InferredType: accType, Expression: expression.Arguments[1]},
			frontend.SyntaxStatement{Span: span, Kind: "variable", VarDeclKind: "let", Name: index, Type: "number", InferredType: "number", Expression: first},
		)
	} else {
		// No initial value: an empty array throws, else the first visited
		// element seeds the accumulator.
		statements = append(statements,
			frontend.SyntaxStatement{Span: span, Kind: "if", Expression: binary("===", length, number("0")), Then: []frontend.SyntaxStatement{{
				Span: span, Kind: "throw", Expression: &frontend.SyntaxExpression{Span: span, Kind: "new", Left: &frontend.SyntaxExpression{Span: span, Kind: "identifier", Text: "TypeError"},
					Arguments: []*frontend.SyntaxExpression{{Span: span, Kind: "string", Text: "Reduce of empty array with no initial value", InferredType: "string"}}},
			}}},
			frontend.SyntaxStatement{Span: span, Kind: "variable", VarDeclKind: "let", Name: index, Type: "number", InferredType: "number", Expression: first},
			frontend.SyntaxStatement{Span: span, Kind: "variable", VarDeclKind: "let", Name: acc, Type: accType, InferredType: accType, Expression: element(ident(index))},
		)
		step := "+"
		if right {
			step = "-"
		}
		statements = append(statements, frontend.SyntaxStatement{Span: span, Kind: "assign", Name: index, Expression: binary(step, ident(index), number("1"))})
	}
	condition := binary("<", ident(index), length)
	step := binary("+", ident(index), number("1"))
	if right {
		condition = binary(">=", ident(index), number("0"))
		step = binary("-", ident(index), number("1"))
	}
	invoke := &frontend.SyntaxExpression{Span: span, Kind: "call", Left: ident(callback), Arguments: []*frontend.SyntaxExpression{ident(acc), element(ident(index)), ident(index), ident(receiver)}}
	statements = append(statements, frontend.SyntaxStatement{
		Span:       span,
		Kind:       "while",
		Expression: condition,
		Body:       []frontend.SyntaxStatement{{Span: span, Kind: "assign", Name: acc, Expression: invoke}},
		Step:       []frontend.SyntaxStatement{{Span: span, Kind: "assign", Name: index, Expression: step}},
	})
	for _, statement := range statements {
		if err := lowerStatement(path, statement, function, env, counter, shapes, signatures); err != nil {
			return "", "", true, err
		}
	}
	value, typ, err := lowerExpression(path, ident(acc), result, function, env, counter, shapes, signatures)
	return value, typ, true, err
}

// reduceInitIsNumber reports a reduce call without an initial value or with
// one TypeScript types as number.
func reduceInitIsNumber(expression *frontend.SyntaxExpression) bool {
	return len(expression.Arguments) < 2 || toIRType(expression.Arguments[1].InferredType) == ir.TypeNumber
}
