package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// appendItemsGeneratorNext appends the body of next() for a generator whose
// values were collected into its __items field (at itemsIndex) when it was
// created: it returns { done: false, value: items[state++] } while values
// remain, then { done: true }. state holds the loaded __state field.
func appendItemsGeneratorNext(body []ir.Instruction, span ir.SourceSpan, genClassName string, itemsIndex int, yieldType ir.Type, resultShapeName, state string, counter *int) []ir.Instruction {
	resultType := ir.Type("object:" + resultShapeName)
	result := func(done, value string) (string, []ir.Instruction) {
		object := nextTemp(counter)
		return object, []ir.Instruction{
			{Op: ir.OpObjectNew, Type: resultType, Result: object, Callee: resultShapeName, FieldCount: 2, Args: []string{done, value}, Span: span},
			{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: resultShapeName, Field: "done", FieldIndex: 0, Args: []string{object, done}, Span: span},
			{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: resultShapeName, Field: "value", FieldIndex: 1, Args: []string{object, value}, Span: span},
			{Op: ir.OpReturn, Type: resultType, Args: []string{object}, Span: span},
		}
	}

	items := nextTemp(counter)
	length := nextTemp(counter)
	hasMore := nextTemp(counter)
	body = append(body,
		ir.Instruction{Op: ir.OpFieldGet, Type: ir.Type(string(yieldType) + "[]"), Result: items, Callee: genClassName, Field: "__items", FieldIndex: itemsIndex, Args: []string{"this"}, Span: span},
		ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: length, Callee: "__array.length", Args: []string{items}, Span: span},
		ir.Instruction{Op: ir.OpCompare, Type: ir.TypeBool, Result: hasMore, Operator: "<", Args: []string{state, length}, Span: span},
	)

	value := nextTemp(counter)
	one := nextTemp(counter)
	advanced := nextTemp(counter)
	notDone := nextTemp(counter)
	then := []ir.Instruction{
		{Op: ir.OpIndex, Type: yieldType, Result: value, Args: []string{items, state}, Span: span},
		{Op: ir.OpConst, Type: ir.TypeNumber, Result: one, Value: "1", Span: span},
		{Op: ir.OpBinary, Type: ir.TypeNumber, Result: advanced, Operator: "+", Args: []string{state, one}, Span: span},
		{Op: ir.OpFieldSet, Type: ir.TypeVoid, Callee: genClassName, Field: "__state", FieldIndex: 0, Args: []string{"this", advanced}, Span: span},
		{Op: ir.OpConst, Type: ir.TypeBool, Result: notDone, Value: "false", Span: span},
	}
	_, ret := result(notDone, value)
	then = append(then, ret...)
	body = append(body, ir.Instruction{Op: ir.OpIf, Type: ir.TypeVoid, Args: []string{hasMore}, Then: then, Span: span})

	finished := nextTemp(counter)
	done := nextTemp(counter)
	finishedValue := "undefined" // a finished generator's value
	if yieldType == ir.TypeBool || yieldType == ir.TypeBigInt {
		finishedValue = "0"
	}
	body = append(body,
		ir.Instruction{Op: ir.OpConst, Type: yieldType, Result: finished, Value: finishedValue, Span: span},
		ir.Instruction{Op: ir.OpConst, Type: ir.TypeBool, Result: done, Value: "true", Span: span},
	)
	_, ret = result(done, finished)
	return append(body, ret...)
}

// ensureClosureGeneratorNext defines next() for a generator closure's class
// (its values are collected in __items, field 4) and its IteratorResult
// shape, once.
func ensureClosureGeneratorNext(span ir.SourceSpan, genClassName string, yieldType ir.Type, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) ir.Function {
	name := genClassName + "_next"
	if fn, ok := signatures[name]; ok {
		return fn
	}
	resultShapeName := "IteratorResult_" + string(yieldType)
	if _, exists := shapes[resultShapeName]; !exists {
		shapes[resultShapeName] = ir.ObjectShape{Name: resultShapeName, Span: span, Fields: []ir.Field{
			{Name: "done", Type: ir.TypeBool, Span: span},
			{Name: "value", Type: yieldType, Span: span},
		}}
	}
	counter := 0
	state := nextTemp(&counter)
	fn := ir.Function{
		Name:       name,
		Span:       span,
		ReturnType: ir.Type("object:" + resultShapeName),
		Parameters: []ir.Parameter{{Name: "this", Type: ir.Type("object:" + genClassName)}},
		Body:       []ir.Instruction{{Op: ir.OpFieldGet, Type: ir.TypeNumber, Result: state, Callee: genClassName, Field: "__state", FieldIndex: 0, Args: []string{"this"}, Span: span}},
	}
	fn.Body = appendItemsGeneratorNext(fn.Body, span, genClassName, 4, yieldType, resultShapeName, state, &counter)
	signatures[name] = fn
	extraFunctions = append(extraFunctions, fn)
	return fn
}

// generatorNextCall advances a generator known only by its Generator<T>
// type: every generator object stores its next() as the __next closure
// field, looked up by name.
func generatorNextCall(span ir.SourceSpan, generator, shapeName, result string, resultType ir.Type, counter *int) []ir.Instruction {
	next := nextTemp(counter)
	return []ir.Instruction{
		{Op: ir.OpFieldGet, Type: ir.TypeClosure, Result: next, Callee: shapeName, Field: "__next", FieldIndex: 0, DynamicField: true, Args: []string{generator}, Span: span},
		{Op: ir.OpClosureCall, Type: resultType, Result: result, Callee: next, Args: []string{generator}, Span: span},
	}
}

// lowerGeneratorNextMethod lowers gen.next() on a value typed only
// Generator<T> (no generator class of its own) through generatorNextCall.
func lowerGeneratorNextMethod(path string, expression *frontend.SyntaxExpression, receiver, className, result string, function *ir.Function, shapes map[string]ir.ObjectShape, counter *int) (string, ir.Type, error) {
	yieldType := ir.TypeUnknown
	if expression.Left != nil && expression.Left.Left != nil {
		if element, ok := iteratorElementType(expression.Left.Left); ok {
			yieldType = element
		}
	}
	span := toIRSpan(path, expression.Span)
	resultShapeName := "IteratorResult_" + string(yieldType)
	if _, exists := shapes[resultShapeName]; !exists {
		shapes[resultShapeName] = ir.ObjectShape{Name: resultShapeName, Span: span, Fields: []ir.Field{
			{Name: "done", Type: ir.TypeBool, Span: span},
			{Name: "value", Type: yieldType, Span: span},
		}}
	}
	if result == "" {
		result = nextTemp(counter)
	}
	resultType := ir.Type("object:" + resultShapeName)
	function.Body = append(function.Body, generatorNextCall(span, receiver, strings.TrimPrefix(className, "object:"), result, resultType, counter)...)
	return result, resultType, nil
}
