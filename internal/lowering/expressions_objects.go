package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerObjectLiteralExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if len(expression.Arguments) == 0 {
		return "", "", fmt.Errorf("empty object literal needs explicit type or shape")
	}
	var fields []ir.Field
	var propValues []string
	for _, prop := range expression.Arguments {
		if prop.Kind == "spread" {
			val, valType, err := lowerExpression(path, prop.Left, "", function, env, counter, shapes, signatures)
			if err != nil {
				return "", "", err
			}
			srcShapeName := strings.TrimPrefix(string(valType), "object:")
			if s, ok := shapes[srcShapeName]; ok {
				for sIdx, f := range s.Fields {
					fieldVal := nextTemp(counter)
					function.Body = append(function.Body, ir.Instruction{
						Op:         ir.OpFieldGet,
						Type:       f.Type,
						Result:     fieldVal,
						Callee:     srcShapeName,
						Field:      f.Name,
						FieldIndex: sIdx,
						Args:       []string{val},
						Span:       toIRSpan(path, prop.Span),
					})
					fields = append(fields, ir.Field{
						Name: f.Name,
						Type: f.Type,
						Span: toIRSpan(path, prop.Span),
					})
					propValues = append(propValues, fieldVal)
				}
			}
			continue
		}
		val, valType, err := lowerExpression(path, prop.Left, "", function, env, counter, shapes, signatures)
		if err != nil {
			return "", "", err
		}
		fields = append(fields, ir.Field{
			Name: prop.Text,
			Type: valType,
			Span: toIRSpan(path, prop.Span),
		})
		propValues = append(propValues, val)
	}
	shapeName := anonymousShapeName(fields)
	if expression.InferredType != "" {
		cleanInf := strings.TrimPrefix(expression.InferredType, "object:")
		if _, ok := shapes[cleanInf]; ok {
			shapeName = cleanInf
		}
	}
	if shapeName == anonymousShapeName(fields) {
		for name, s := range shapes {
			if !strings.HasPrefix(name, "__shape_") && len(s.Fields) == len(fields) {
				match := true
				for i, f := range fields {
					if s.Fields[i].Name != f.Name {
						match = false
						break
					}
				}
				if match {
					shapeName = name
					break
				}
			}
		}
	}
	if _, ok := shapes[shapeName]; !ok {
		shapes[shapeName] = ir.ObjectShape{
			Name:   shapeName,
			Span:   toIRSpan(path, expression.Span),
			Fields: fields,
		}
	}
	targetShape := shapes[shapeName]
	if result == "" {
		result = nextTemp(counter)
	}
	objType := ir.Type("object:" + shapeName)
	function.Body = append(function.Body, ir.Instruction{
		Op:         ir.OpObjectNew,
		Type:       objType,
		Result:     result,
		Callee:     shapeName,
		FieldCount: len(targetShape.Fields),
		Span:       toIRSpan(path, expression.Span),
	})
	propMap := map[string]string{}
	for i, f := range fields {
		propMap[f.Name] = propValues[i]
	}
	for i, field := range targetShape.Fields {
		if val, exists := propMap[field.Name]; exists {
			function.Body = append(function.Body, ir.Instruction{
				Op:         ir.OpFieldSet,
				Type:       ir.TypeVoid,
				Callee:     shapeName,
				Field:      field.Name,
				FieldIndex: i,
				Args:       []string{result, val},
				Span:       toIRSpan(path, expression.Span),
			})
		}
	}
	return result, objType, nil
}

func anonymousShapeName(fields []ir.Field) string {
	var parts []string
	for _, f := range fields {
		cleanType := strings.ReplaceAll(string(f.Type), ":", "_")
		cleanType = strings.ReplaceAll(cleanType, "[]", "_arr")
		parts = append(parts, fmt.Sprintf("%s_%s", f.Name, cleanType))
	}
	return "__shape_" + strings.Join(parts, "_")
}

func ensureDateShape(shapes map[string]ir.ObjectShape) {
	if _, ok := shapes["Date"]; !ok {
		shapes["Date"] = ir.ObjectShape{
			Name: "Date",
			Fields: []ir.Field{
				{Name: "time", Type: ir.TypeNumber},
			},
		}
	}
}

func ensureTextEncoderEncodeIntoResultShape(shapes map[string]ir.ObjectShape) string {
	shapeName := "TextEncoderEncodeIntoResult"
	if _, ok := shapes[shapeName]; !ok {
		shapes[shapeName] = ir.ObjectShape{
			Name: shapeName,
			Fields: []ir.Field{
				{Name: "read", Type: ir.TypeNumber},
				{Name: "written", Type: ir.TypeNumber},
			},
		}
	}
	return shapeName
}
