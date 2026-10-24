package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

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
