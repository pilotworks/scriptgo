package lowering

import (
	"fmt"
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerObjectLiteralExpression(path string, expression *typescriptgo.SyntaxExpression, result string, function *ir.Function, env map[string]ir.Type, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, ir.Type, error) {
	if len(expression.Arguments) == 0 {
		shapeName := "__shape_empty"
		if expression.InferredType != "" {
			cleanInf := strings.TrimPrefix(expression.InferredType, "object:")
			if _, ok := shapes[cleanInf]; ok {
				shapeName = cleanInf
			} else if s, ok := registeredShapes[cleanInf]; ok {
				shapes[cleanInf] = s
				shapeName = cleanInf
			} else if s, ok := anonymousShapes[cleanInf]; ok {
				shapes[cleanInf] = s
				shapeName = cleanInf
			}
		}
		if _, ok := shapes[shapeName]; !ok {
			shapes[shapeName] = ir.ObjectShape{
				Name:   shapeName,
				Span:   toIRSpan(path, expression.Span),
				Fields: nil,
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
		for i, field := range targetShape.Fields {
			defVal := "undefined"
			defType := field.Type
			switch field.Type {
			case ir.TypeNumber:
				defVal = "NaN"
			case ir.TypeBool:
				defVal = "false"
			}
			defConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   defType,
				Result: defConst,
				Value:  defVal,
				Span:   toIRSpan(path, expression.Span),
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:         ir.OpFieldSet,
				Type:       ir.TypeVoid,
				Callee:     shapeName,
				Field:      field.Name,
				FieldIndex: i,
				Args:       []string{result, defConst},
				Span:       toIRSpan(path, expression.Span),
			})
		}
		return result, objType, nil
	}
	var parentShape *ir.ObjectShape
	var cleanParent string
	if expression.InferredType != "" {
		cleanParent = strings.TrimPrefix(expression.InferredType, "object:")
		irT := toIRType(cleanParent)
		cleanT := strings.TrimPrefix(string(irT), "object:")
		if s, ok := shapes[cleanT]; ok {
			parentShape = &s
		} else if s, ok := anonymousShapes[cleanT]; ok {
			parentShape = &s
		} else if s, ok := registeredShapes[cleanT]; ok {
			parentShape = &s
		} else if s, ok := shapes[cleanParent]; ok {
			parentShape = &s
		} else if s, ok := anonymousShapes[cleanParent]; ok {
			parentShape = &s
		} else if s, ok := registeredShapes[cleanParent]; ok {
			parentShape = &s
		}
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
		if parentShape != nil && prop.Left != nil && prop.Left.Kind == "object_literal" && (prop.Left.InferredType == "" || strings.HasPrefix(prop.Left.InferredType, "{") || prop.Left.InferredType == "object") {
			for _, pf := range parentShape.Fields {
				if pf.Name == prop.Text {
					t := strings.TrimPrefix(string(pf.Type), "object:")
					if t == "" || t == "object" {
						t = cleanParent
					}
					prop.Left.InferredType = t
					break
				}
			}
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
	var targetS *ir.ObjectShape
	if expression.InferredType != "" {
		cleanInf := strings.TrimPrefix(expression.InferredType, "object:")
		for strings.HasPrefix(cleanInf, "(") && strings.HasSuffix(cleanInf, ")") && !strings.Contains(cleanInf, "=>") {
			cleanInf = strings.TrimSpace(cleanInf[1 : len(cleanInf)-1])
		}
		if s, ok := shapes[cleanInf]; ok {
			allFound := true
			for _, f := range fields {
				if fieldIndex(s, f.Name) < 0 {
					allFound = false
					break
				}
			}
			if allFound {
				targetS = &s
			}
		} else if s, ok := anonymousShapes[cleanInf]; ok {
			allFound := true
			for _, f := range fields {
				if fieldIndex(s, f.Name) < 0 {
					allFound = false
					break
				}
			}
			if allFound {
				targetS = &s
			}
		}
		if targetS == nil {
			irT := toIRType(cleanInf)
			cleanT := strings.TrimPrefix(string(irT), "object:")
			if s, ok := shapes[cleanT]; ok {
				allFound := true
				for _, f := range fields {
					if fieldIndex(s, f.Name) < 0 {
						allFound = false
						break
					}
				}
				if allFound {
					targetS = &s
				}
			} else if s, ok := anonymousShapes[cleanT]; ok {
				allFound := true
				for _, f := range fields {
					if fieldIndex(s, f.Name) < 0 {
						allFound = false
						break
					}
				}
				if allFound {
					targetS = &s
				}
			}
		}
		if targetS == nil && (strings.Contains(cleanInf, "|") || (typeAliasesIndex != nil && strings.Contains(typeAliasesIndex[cleanInf], "|"))) {
			unionStr := cleanInf
			if typeAliasesIndex != nil && strings.Contains(typeAliasesIndex[cleanInf], "|") {
				unionStr = typeAliasesIndex[cleanInf]
			}
			members := strings.Split(unionStr, "|")
			for _, m := range members {
				cleanM := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(m), "object:"))
				if s, ok := shapes[cleanM]; ok && !strings.HasPrefix(cleanM, "{") {
					allFound := true
					for _, f := range fields {
						if fieldIndex(s, f.Name) < 0 {
							allFound = false
							break
						}
					}
					if allFound && len(s.Fields) == len(fields) {
						targetS = &s
						break
					}
				}
			}
		}
		if targetS != nil {
			allFound := true
			for _, f := range fields {
				if fieldIndex(*targetS, f.Name) < 0 {
					allFound = false
					break
				}
			}
			if allFound {
				shapes[targetS.Name] = *targetS
				shapeName = targetS.Name
			}
		}
	}
	if shapeName == anonymousShapeName(fields) {
		for name, s := range registeredShapes {
			if !strings.HasPrefix(name, "__shape_") && len(s.Fields) == len(fields) {
				match := true
				for _, f := range fields {
					if fieldIndex(s, f.Name) < 0 {
						match = false
						break
					}
				}
				if match {
					shapes[name] = s
					shapeName = name
					break
				}
			}
		}
		if shapeName == anonymousShapeName(fields) {
			for name, s := range shapes {
				if !strings.HasPrefix(name, "__shape_") && len(s.Fields) == len(fields) {
					match := true
					for _, f := range fields {
						if fieldIndex(s, f.Name) < 0 {
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
	}
	var targetShape ir.ObjectShape
	if targetS != nil {
		targetShape = *targetS
		shapeName = targetS.Name
		shapes[shapeName] = *targetS
	} else if s, ok := shapes[shapeName]; ok {
		targetShape = s
	} else if s, ok := anonymousShapes[shapeName]; ok {
		targetShape = s
		shapes[shapeName] = s
	} else {
		targetShape = ir.ObjectShape{
			Name:   shapeName,
			Span:   toIRSpan(path, expression.Span),
			Fields: fields,
		}
		shapes[shapeName] = targetShape
	}
	if result == "" {
		result = nextTemp(counter)
	}
	var tagNames []string
	for _, f := range fields {
		tagNames = append(tagNames, f.Name)
	}
	typeTag := ":" + strings.Join(tagNames, ":") + ":"
	objType := ir.Type("object:" + shapeName)
	function.Body = append(function.Body, ir.Instruction{
		Op:         ir.OpObjectNew,
		Type:       objType,
		Result:     result,
		Callee:     shapeName,
		Value:      typeTag,
		FieldCount: len(targetShape.Fields),
		Span:       toIRSpan(path, expression.Span),
	})
	propMap := map[string]string{}
	propTypeMap := map[string]ir.Type{}
	for i, f := range fields {
		propMap[f.Name] = propValues[i]
		propTypeMap[f.Name] = f.Type
	}
	for i, field := range targetShape.Fields {
		if val, exists := propMap[field.Name]; exists {
			valType := propTypeMap[field.Name]
			if field.Type == ir.TypeUnknown && valType != ir.TypeUnknown {
				boxed := nextTemp(counter)
				env[boxed] = ir.TypeUnknown
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpBoxUnknown,
					Type:   ir.TypeUnknown,
					Result: boxed,
					Args:   []string{val},
					Span:   toIRSpan(path, expression.Span),
				})
				val = boxed
			} else if field.Type != ir.TypeUnknown && valType == ir.TypeUnknown {
				unboxed := nextTemp(counter)
				env[unboxed] = field.Type
				function.Body = append(function.Body, ir.Instruction{
					Op:     ir.OpCheckedCast,
					Type:   field.Type,
					Result: unboxed,
					Args:   []string{val},
					Span:   toIRSpan(path, expression.Span),
				})
				val = unboxed
			}
			function.Body = append(function.Body, ir.Instruction{
				Op:         ir.OpFieldSet,
				Type:       ir.TypeVoid,
				Callee:     shapeName,
				Field:      field.Name,
				FieldIndex: i,
				Args:       []string{result, val},
				Span:       toIRSpan(path, expression.Span),
			})
		} else {
			defVal := "undefined"
			defType := field.Type
			if field.Type == ir.TypeNumber {
				defVal = "NaN"
			} else if field.Type == ir.TypeBool {
				defVal = "false"
			}
			defConst := nextTemp(counter)
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpConst,
				Type:   defType,
				Result: defConst,
				Value:  defVal,
				Span:   toIRSpan(path, expression.Span),
			})
			function.Body = append(function.Body, ir.Instruction{
				Op:         ir.OpFieldSet,
				Type:       ir.TypeVoid,
				Callee:     shapeName,
				Field:      field.Name,
				FieldIndex: i,
				Args:       []string{result, defConst},
				Span:       toIRSpan(path, expression.Span),
			})
		}
	}
	return result, objType, nil
}

func anonymousShapeName(fields []ir.Field) string {
	var parts []string
	for i, f := range fields {
		fieldName := f.Name
		if fieldName == "" {
			fieldName = strconv.Itoa(i)
		}
		cleanType := strings.ReplaceAll(string(f.Type), ":", "_")
		cleanType = strings.ReplaceAll(cleanType, "[]", "_arr")
		parts = append(parts, fmt.Sprintf("%s_%s", fieldName, cleanType))
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
