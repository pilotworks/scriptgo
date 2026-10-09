package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerArraySpreadElement appends the elements of a spread element
// (`[...xs]`) to the array result of type arrType: a tuple element by
// element, an array with an index loop, and any other iterable (a string,
// Map, Set, typed array, generator, iterator, or class with Symbol.iterator)
// through for..of.
func lowerArraySpreadElement(path string, elem *frontend.SyntaxExpression, result string, arrType ir.Type, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	spreadVal, spreadType, err := lowerExpression(path, elem.Left, "", function, env, counter, shapes, signatures)
	if err != nil {
		return err
	}
	shapeName := strings.TrimPrefix(string(spreadType), "object:")
	if shape, ok := lookupObjectShape(shapeName, shapes); ok && isTupleShape(shape) {
		// A tuple has a fixed length: push each element in order.
		for index, field := range shape.Fields {
			item := nextTemp(counter)
			pushed := nextTemp(counter)
			function.Body = append(function.Body,
				ir.Instruction{Op: ir.OpFieldGet, Type: field.Type, Result: item, Callee: shapeName, Field: field.Name, FieldIndex: index, Args: []string{spreadVal}, Span: toIRSpan(path, elem.Span)},
				ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: pushed, Callee: "__array.push", Args: []string{result, item}, Span: toIRSpan(path, elem.Span)},
			)
		}
		return nil
	}
	if !strings.HasSuffix(string(spreadType), "[]") {
		return lowerIterableSpreadElement(path, elem, spreadVal, spreadType, result, arrType, function, env, counter, shapes, signatures)
	}
	idxVar := nextTemp(counter)
	lenVar := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: idxVar, Value: "0", Span: toIRSpan(path, elem.Span)})
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: lenVar, Callee: "__array.length", Args: []string{spreadVal}, Span: toIRSpan(path, elem.Span)})
	condVar := nextTemp(counter)
	var condBody []ir.Instruction
	condBody = append(condBody, ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: condVar, Operator: "<", Args: []string{idxVar, lenVar}, Span: toIRSpan(path, elem.Span)})
	var loopBody []ir.Instruction
	itemVar := nextTemp(counter)
	itemType := arrayElementType(arrType)
	if spreadType != "" && strings.HasSuffix(string(spreadType), "[]") {
		itemType = arrayElementType(spreadType)
	}
	loopBody = append(loopBody, ir.Instruction{Op: ir.OpIndex, Type: itemType, Result: itemVar, Args: []string{spreadVal, idxVar}, Span: toIRSpan(path, elem.Span)})
	pushRes := nextTemp(counter)
	loopBody = append(loopBody, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: pushRes, Callee: "__array.push", Args: []string{result, itemVar}, Span: toIRSpan(path, elem.Span)})
	oneConst := nextTemp(counter)
	loopBody = append(loopBody, ir.Instruction{Op: ir.OpConst, Type: ir.TypeNumber, Result: oneConst, Value: "1", Span: toIRSpan(path, elem.Span)})
	nextIdx := nextTemp(counter)
	loopBody = append(loopBody, ir.Instruction{Op: ir.OpBinary, Type: ir.TypeNumber, Result: nextIdx, Operator: "+", Args: []string{idxVar, oneConst}, Span: toIRSpan(path, elem.Span)})
	loopBody = append(loopBody, ir.Instruction{Op: ir.OpAssign, Type: ir.TypeNumber, Result: idxVar, Args: []string{nextIdx}, Span: toIRSpan(path, elem.Span)})

	function.Body = append(function.Body, ir.Instruction{
		Op:   ir.OpWhile,
		Type: ir.TypeVoid,
		Cond: condBody,
		Args: []string{condVar},
		Body: loopBody,
		Span: toIRSpan(path, elem.Span),
	})
	return nil
}

// lowerIterableSpreadElement pushes every value of the iterable spreadVal onto result with the for..of lowering, which owns the iteration
// protocol: `for (const item of spreadVal) result.push(item)`.
func lowerIterableSpreadElement(path string, elem *frontend.SyntaxExpression, spreadVal string, spreadType ir.Type, result string, arrType ir.Type, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) error {
	env[spreadVal] = spreadType
	env[result] = arrType
	item := "__spread_item_" + nextTemp(counter)
	receiver := &frontend.SyntaxExpression{Span: elem.Span, Kind: "identifier", Text: result, InferredType: string(arrType)}
	push := frontend.SyntaxStatement{
		Span: elem.Span,
		Kind: "expression",
		Expression: &frontend.SyntaxExpression{
			Span:         elem.Span,
			Kind:         "call",
			Left:         &frontend.SyntaxExpression{Span: elem.Span, Kind: "property", Text: "push", Left: receiver},
			Arguments:    []*frontend.SyntaxExpression{{Span: elem.Span, Kind: "identifier", Text: item}},
			InferredType: "number",
		},
	}
	return lowerForOf(path, frontend.SyntaxStatement{
		Span:       elem.Span,
		Kind:       "forof",
		Name:       item,
		Expression: &frontend.SyntaxExpression{Span: elem.Left.Span, Kind: "identifier", Text: spreadVal, InferredType: elem.Left.InferredType},
		Body:       []frontend.SyntaxStatement{push},
	}, function, env, counter, shapes, signatures)
}
