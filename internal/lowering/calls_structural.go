package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

// adaptStructuralObjectArgument materializes the destination shape required by
// a function signature when the source value has a different structural shape.
func adaptStructuralObjectArgument(
	path string,
	argumentSpan ir.SourceSpan,
	value string,
	valueType ir.Type,
	parameterType ir.Type,
	function *ir.Function,
	counter *int,
	shapes map[string]ir.ObjectShape,
) (string, ir.Type, bool) {
	if !strings.HasPrefix(string(valueType), "object:") || !strings.HasPrefix(string(parameterType), "object:") || valueType == parameterType {
		return value, valueType, false
	}
	sourceName := strings.TrimPrefix(string(valueType), "object:")
	destinationName := strings.TrimPrefix(string(parameterType), "object:")
	sourceShape, sourceOK := shapes[sourceName]
	if !sourceOK {
		if fields, ok := resolveShapeFields(sourceName, shapes); ok {
			sourceShape = ir.ObjectShape{Name: sourceName, Fields: fields}
			shapes[sourceName] = sourceShape
			sourceOK = true
		}
	}
	destinationShape, destinationOK := shapes[destinationName]
	if !destinationOK {
		if fields, ok := resolveShapeFields(destinationName, shapes); ok {
			destinationShape = ir.ObjectShape{Name: destinationName, Fields: fields}
			shapes[destinationName] = destinationShape
			destinationOK = true
		}
	}
	if !sourceOK || !destinationOK || len(destinationShape.Fields) == 0 {
		return value, valueType, false
	}

	adapted := nextTemp(counter)
	fieldNames := make([]string, 0, len(destinationShape.Fields))
	for _, field := range destinationShape.Fields {
		fieldNames = append(fieldNames, field.Name)
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:         ir.OpObjectNew,
		Type:       parameterType,
		Result:     adapted,
		Callee:     destinationName,
		Value:      ":" + strings.Join(fieldNames, ":") + ":",
		FieldCount: len(destinationShape.Fields),
		Span:       argumentSpan,
	})

	for destinationIndex, destinationField := range destinationShape.Fields {
		for sourceIndex, sourceField := range sourceShape.Fields {
			if sourceField.Name != destinationField.Name {
				continue
			}
			fieldValue := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:           ir.OpFieldGet,
				Type:         sourceField.Type,
				Result:       fieldValue,
				Callee:       sourceName,
				Field:        sourceField.Name,
				FieldIndex:   sourceIndex,
				DynamicField: dynamicFieldAccess(sourceName),
				Args:         []string{value},
				Span:         argumentSpan,
			})
			if destinationField.Type == ir.TypeUnknown && sourceField.Type != ir.TypeUnknown {
				boxed := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpBoxUnknown,
					Type:   ir.TypeUnknown,
					Result: boxed,
					Args:   []string{fieldValue},
					Span:   argumentSpan,
				})
				fieldValue = boxed
			} else if destinationField.Type != ir.TypeUnknown && sourceField.Type == ir.TypeUnknown {
				casted := nextTemp(counter)
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCheckedCast,
					Type:   destinationField.Type,
					Result: casted,
					Args:   []string{fieldValue},
					Span:   argumentSpan,
				})
				fieldValue = casted
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:           ir.OpFieldSet,
				Type:         ir.TypeVoid,
				Callee:       destinationName,
				Field:        destinationField.Name,
				FieldIndex:   destinationIndex,
				DynamicField: dynamicFieldAccess(destinationName),
				Args:         []string{adapted, fieldValue},
				Span:         argumentSpan,
			})
			break
		}
	}
	return adapted, parameterType, true
}
