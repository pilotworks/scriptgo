package lowering

import (
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func isValidFieldName(s string) bool {
	if s == "" {
		return false
	}
	r := s[0]
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || r == '$'
}

func parseFlattenedShapeFields(s string) ([]ir.Field, bool) {
	s = strings.TrimPrefix(s, "object:")
	if strings.Contains(s, "<") || strings.Contains(s, "{") || strings.Contains(s, "(") || strings.Contains(s, "|") {
		return nil, false
	}
	s = strings.TrimPrefix(s, "__shape_")
	tokens := strings.Split(s, "_")
	if len(tokens) < 2 || len(tokens)%2 != 0 {
		return nil, false
	}
	var fields []ir.Field
	for i := 0; i < len(tokens); i += 2 {
		fName := tokens[i]
		fType := tokens[i+1]
		if !isValidFieldName(fName) {
			return nil, false
		}
		var typ ir.Type
		switch fType {
		case "number":
			typ = ir.TypeNumber
		case "string":
			typ = ir.TypeString
		case "bool", "boolean":
			typ = ir.TypeBool
		case "bigint":
			typ = ir.TypeBigInt
		default:
			if strings.HasSuffix(fType, "arr") {
				elem := strings.TrimSuffix(strings.TrimSuffix(fType, "_arr"), "arr")
				typ = ir.Type(string(toIRType(elem)) + "[]")
			} else {
				return nil, false
			}
		}
		fields = append(fields, ir.Field{Name: fName, Type: typ})
	}
	return fields, true
}

var anonymousShapes = make(map[string]ir.ObjectShape)

func registerAnonymousShape(name string, fields []ir.Field) {
	if _, exists := anonymousShapes[name]; !exists {
		anonymousShapes[name] = ir.ObjectShape{Name: name, Fields: fields}
	}
}

func tupleFields(typeStr string) ([]ir.Field, bool) {
	typeStr = strings.TrimSpace(typeStr)
	typeStr = strings.TrimPrefix(typeStr, "readonly ")
	if strings.HasPrefix(typeStr, "Readonly<") && strings.HasSuffix(typeStr, ">") {
		typeStr = strings.TrimSuffix(strings.TrimPrefix(typeStr, "Readonly<"), ">")
		typeStr = strings.TrimSpace(typeStr)
	}
	if typeAliasesIndex != nil && typeAliasesIndex[typeStr] != "" {
		return tupleFields(typeAliasesIndex[typeStr])
	}
	if strings.HasSuffix(typeStr, "[]") || !strings.HasPrefix(typeStr, "[") || !strings.HasSuffix(typeStr, "]") {
		return nil, false
	}
	inner := strings.TrimSpace(typeStr[1 : len(typeStr)-1])
	if inner == "" || strings.Contains(inner, "...") {
		return nil, false
	}
	parts := splitTopLevel(inner)
	var fields []ir.Field
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		if idx := strings.Index(trimmed, ":"); idx != -1 {
			trimmed = strings.TrimSpace(trimmed[idx+1:])
		}
		isRest := strings.HasPrefix(trimmed, "...")
		trimmed = strings.TrimPrefix(trimmed, "...")
		trimmed = strings.TrimSuffix(trimmed, "?")
		elemType := toIRType(trimmed)
		if isRest {
			if strings.HasSuffix(string(elemType), "[]") {
				elemType = arrayElementType(elemType)
			}
		}
		fields = append(fields, ir.Field{
			Name: strconv.Itoa(i),
			Type: elemType,
		})
	}
	return fields, true
}

func anonymousObjectFields(typeStr string, visited map[string]bool) ([]ir.Field, bool) {
	if strings.Contains(typeStr, "|") && len(splitTopLevelUnion(typeStr)) > 1 {
		return nil, false
	}
	if strings.Contains(typeStr, "&") {
		interParts := splitTopLevelIntersection(typeStr)
		if len(interParts) > 1 {
			var allFields []ir.Field
			seenFields := make(map[string]bool)
			for _, ip := range interParts {
				if fList, ok := anonymousObjectFields(strings.TrimSpace(ip), visited); ok {
					for _, f := range fList {
						if !seenFields[f.Name] {
							seenFields[f.Name] = true
							allFields = append(allFields, f)
						}
					}
				}
			}
			if len(allFields) > 0 {
				return allFields, true
			}
		}
	}
	if !strings.HasPrefix(typeStr, "{") || !strings.HasSuffix(typeStr, "}") {
		clean := strings.TrimSpace(strings.TrimPrefix(typeStr, "object:"))
		if s, ok := registeredShapes[clean]; ok {
			return s.Fields, true
		}
		if typeAliasesIndex != nil && typeAliasesIndex[clean] != "" {
			if visited == nil {
				visited = make(map[string]bool)
			}
			if !visited[clean] {
				visited[clean] = true
				return anonymousObjectFields(typeAliasesIndex[clean], visited)
			}
		}
		return nil, false
	}
	inner := strings.TrimSpace(typeStr[1 : len(typeStr)-1])
	if inner == "" {
		return nil, false
	}
	parts := splitTopLevel(inner)
	var fields []ir.Field
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		colonIdx := strings.Index(trimmed, ":")
		if colonIdx == -1 {
			continue
		}
		fName := strings.TrimSpace(trimmed[:colonIdx])
		if strings.HasPrefix(fName, "[") {
			continue
		}
		optional := strings.HasSuffix(fName, "?")
		fName = strings.TrimSuffix(fName, "?")
		isMethod := false
		if parenIdx := strings.Index(fName, "("); parenIdx != -1 {
			fName = strings.TrimSpace(fName[:parenIdx])
			isMethod = true
		}
		fName = strings.TrimSuffix(fName, "?")
		fTypeStr := strings.TrimSpace(trimmed[colonIdx+1:])
		fTypeStr = strings.TrimSpace(strings.TrimRight(fTypeStr, ";,}"))
		fieldType := toIRTypeInternal(fTypeStr, visited)
		if isMethod {
			fieldType = ir.TypeClosure
		}
		fields = append(fields, ir.Field{
			Name:     fName,
			Type:     fieldType,
			Optional: optional,
		})
	}
	if len(fields) == 0 {
		return nil, false
	}
	return fields, true
}

func fieldIndex(shape ir.ObjectShape, name string) int {
	for index, field := range shape.Fields {
		if field.Name == name {
			return index
		}
	}
	return -1
}

func dynamicFieldAccess(className string) bool {
	clean := strings.TrimPrefix(className, "object:")
	if strings.HasPrefix(clean, "__shape_") {
		return !strings.HasPrefix(clean, "__shape_0_")
	}
	meta, ok := classHierarchy[clean]
	return ok && (meta.IsInterface || meta.IsTypeAlias)
}

func isTupleShapeName(shapeName string) bool {
	if s, ok := registeredShapes[shapeName]; ok && isTupleShape(s) {
		return true
	}
	if s, ok := anonymousShapes[shapeName]; ok && isTupleShape(s) {
		return true
	}
	return false
}

func tupleCommonElementType(shapeName string) ir.Type {
	var s ir.ObjectShape
	if shape, ok := registeredShapes[shapeName]; ok {
		s = shape
	} else if shape, ok := anonymousShapes[shapeName]; ok {
		s = shape
	} else {
		return ir.TypeUnknownArray
	}
	if len(s.Fields) == 0 {
		return ir.TypeUnknownArray
	}
	firstType := s.Fields[0].Type
	for i := 1; i < len(s.Fields); i++ {
		if s.Fields[i].Type != firstType {
			return ir.TypeUnknownArray
		}
	}
	switch firstType {
	case ir.TypeNumber:
		return ir.TypeNumberArray
	case ir.TypeString:
		return ir.TypeStringArray
	case ir.TypeBool:
		return ir.TypeBoolArray
	case ir.TypeBigInt:
		return ir.TypeBigIntArray
	default:
		return ir.Type(string(firstType) + "[]")
	}
}
