package lowering

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

var typeAliasesIndex = map[string]string{}

func extractTopLevelReturnType(sig string) string {
	depth := 0
	for i := 0; i < len(sig)-1; i++ {
		switch sig[i] {
		case '(', '<', '{', '[':
			depth++
		case ')', '>', '}', ']':
			if depth > 0 {
				depth--
			}
		case '=':
			if depth == 0 && sig[i+1] == '>' {
				return strings.TrimSpace(sig[i+2:])
			}
		}
	}
	if idx := strings.LastIndex(sig, "=>"); idx != -1 {
		return strings.TrimSpace(sig[idx+2:])
	}
	return sig
}

func mangleGenericTypeString(t string) string {
	t = strings.TrimSpace(t)
	if strings.HasSuffix(t, "[]") {
		elem := strings.TrimSuffix(t, "[]")
		return mangleGenericTypeString(elem) + "_arr"
	}
	if strings.Contains(t, "<") && strings.HasSuffix(t, ">") {
		idx := strings.Index(t, "<")
		base := t[:idx]
		inner := t[idx+1 : len(t)-1]
		typeArgs := splitTypeArguments(inner)
		return mangleGenericName(base, typeArgs)
	}
	return t
}

var registeredShapes map[string]ir.ObjectShape

func SetRegisteredShapes(s map[string]ir.ObjectShape) {
	registeredShapes = s
}

func toIRType(value string) ir.Type {
	return toIRTypeInternal(value, nil)
}

// nonNullishIRType models TypeScript's non-null assertion: it removes only
// null/undefined/void from a union and preserves the remaining type. This is
// needed at dynamic API boundaries (for example Map.get(...)) where the
// nullable return ABI must not be used after `!` has narrowed the value.
func nonNullishIRType(value string) ir.Type {
	parts := splitTopLevelUnion(strings.TrimSpace(value))
	if len(parts) <= 1 {
		return toIRType(value)
	}
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && part != "null" && part != "undefined" && part != "void" {
			filtered = append(filtered, part)
		}
	}
	if len(filtered) == 0 {
		return ir.TypeVoid
	}
	return toIRType(strings.Join(filtered, " | "))
}

func typeContainsNullish(typeStr string) bool {
	for _, part := range splitTopLevelUnion(strings.TrimSpace(typeStr)) {
		switch strings.TrimSpace(part) {
		case "null", "undefined", "void":
			return true
		}
	}
	return false
}

func toIRTypeInternal(value string, visited map[string]bool) ir.Type {
	value = strings.TrimSpace(value)
	for strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")") && !strings.Contains(value, "=>") {
		value = strings.TrimSpace(value[1 : len(value)-1])
	}
	// Generic instantiation rewrites concrete typed arrays to names such as
	// `Uint8Array_ArrayBufferLike`. Preserve the runtime typed-array kind after
	// that mangling so methods still use the typed-array ABI.
	if typedArray, ok := mangledTypedArrayType(value); ok {
		return typedArray
	}
	if visited != nil && visited[value] {
		return ir.TypeObject
	}
	cleanVal := strings.TrimPrefix(value, "object:")
	// Shape names encode array-valued fields with a trailing `_arr`. Resolve
	// the shape marker before interpreting that suffix as an array type.
	if strings.HasPrefix(cleanVal, "__shape_") {
		if strings.HasSuffix(cleanVal, "[]") {
			element := toIRTypeInternal(strings.TrimSuffix(value, "[]"), visited)
			return ir.Type(string(element) + "[]")
		}
		return ir.Type("object:" + cleanVal)
	}
	if aliased, ok := typeAliasesIndex[cleanVal]; ok && aliased != cleanVal {
		newVisited := make(map[string]bool, len(visited)+1)
		for k, v := range visited {
			newVisited[k] = v
		}
		newVisited[value] = true
		newVisited[cleanVal] = true
		return toIRTypeInternal(aliased, newVisited)
	}
	base := value
	if idx := strings.Index(base, "__"); idx != -1 {
		base = base[:idx]
	}
	if idx := strings.Index(base, "<"); idx != -1 {
		base = base[:idx]
	}
	if aliased, ok := typeAliasesIndex[base]; ok && aliased != base {
		if strings.Contains(aliased, "=>") && !strings.HasPrefix(aliased, "{") {
			return ir.TypeClosure
		}
		if aliased == "number" {
			return ir.TypeNumber
		}
		if aliased == "string" {
			return ir.TypeString
		}
		if aliased == "boolean" || aliased == "bool" {
			return ir.TypeBool
		}
	}
	if strings.Contains(value, "|") {
		parts := splitTopLevelUnion(value)
		if len(parts) > 1 {
			var nonNullish []string
			for _, p := range parts {
				trimmed := strings.TrimSpace(p)
				if trimmed != "null" && trimmed != "undefined" && trimmed != "void" && trimmed != "" {
					nonNullish = append(nonNullish, trimmed)
				}
			}
			if len(nonNullish) == 0 {
				return ir.TypeVoid
			}
			hasNullish := len(nonNullish) < len(parts)
			hasNull := false
			for _, p := range parts {
				if strings.TrimSpace(p) == "null" {
					hasNull = true
					break
				}
			}
			allStrings := true
			for _, p := range nonNullish {
				if p != "string" && !((strings.HasPrefix(p, "\"") && strings.HasSuffix(p, "\"")) || (strings.HasPrefix(p, "'") && strings.HasSuffix(p, "'"))) {
					allStrings = false
					break
				}
			}
			if allStrings {
				if hasNull {
					return ir.TypeUnknown
				}
				return ir.TypeString
			}
			allNumbers := true
			for _, p := range nonNullish {
				if p != "number" {
					if _, err := strconv.ParseFloat(p, 64); err != nil {
						allNumbers = false
						break
					}
				}
			}
			if allNumbers && !hasNullish {
				return ir.TypeNumber
			}
			allByteBuffers := true
			for _, p := range nonNullish {
				irT := toIRTypeInternal(p, visited)
				if irT != ir.TypeUint8Array && irT != ir.TypeBuffer && irT != ir.TypeUint8ClampedArray {
					allByteBuffers = false
					break
				}
			}
			if allByteBuffers {
				return ir.TypeUint8Array
			}
			allSameIR := true
			var firstIR ir.Type
			for i, p := range nonNullish {
				irT := toIRTypeInternal(p, visited)
				if i == 0 {
					firstIR = irT
				} else if irT != firstIR {
					allSameIR = false
					break
				}
			}
			if allSameIR && firstIR != "" && firstIR != ir.TypeUnknown && firstIR != ir.TypeObject {
				if !hasNullish || isPointerLikeType(firstIR) {
					return firstIR
				}
				return ir.TypeUnknown
			}
			if len(nonNullish) > 1 {
				allObjects := true
				var unionFields []ir.Field
				seenFields := map[string]int{}
				for _, branch := range nonNullish {
					var fields []ir.Field
					cleanBranch := strings.TrimPrefix(branch, "object:")
					if f, ok := anonymousObjectFields(branch, visited); ok {
						fields = f
					} else if shape, ok := registeredShapes[cleanBranch]; ok && len(shape.Fields) > 0 {
						fields = shape.Fields
					} else if shape, ok := anonymousShapes[cleanBranch]; ok && len(shape.Fields) > 0 {
						fields = shape.Fields
					} else if aliased, ok := typeAliasesIndex[cleanBranch]; ok && aliased != cleanBranch {
						if f, ok := anonymousObjectFields(aliased, visited); ok {
							fields = f
						} else if shape, ok := registeredShapes[aliased]; ok && len(shape.Fields) > 0 {
							fields = shape.Fields
						} else if shape, ok := anonymousShapes[aliased]; ok && len(shape.Fields) > 0 {
							fields = shape.Fields
						}
					}
					if len(fields) == 0 {
						allObjects = false
						break
					}
					for _, f := range fields {
						if existingIdx, exists := seenFields[f.Name]; exists {
							if unionFields[existingIdx].Type != f.Type {
								unionFields[existingIdx].Type = ir.TypeUnknown
							}
						} else {
							seenFields[f.Name] = len(unionFields)
							unionFields = append(unionFields, f)
						}
					}
				}
				if allObjects && len(unionFields) > 0 {
					name := anonymousShapeName(unionFields)
					registerAnonymousShape(name, unionFields)
					return ir.Type("object:" + name)
				}
			}
			if len(nonNullish) == 1 {
				singleIR := toIRTypeInternal(nonNullish[0], visited)
				if hasNullish && !isPointerLikeType(singleIR) {
					return ir.TypeUnknown
				}
				return singleIR
			}
			return ir.TypeUnknown
		}
	}
	if strings.HasSuffix(value, "_arr") {
		elem := strings.TrimSuffix(value, "_arr")
		return toIRTypeInternal(elem+"[]", visited)
	}
	if strings.HasSuffix(value, "[]") {
		elem := strings.TrimSuffix(value, "[]")
		elem = strings.TrimPrefix(elem, "(")
		elem = strings.TrimSuffix(elem, ")")
		elemType := toIRTypeInternal(elem, visited)
		switch elemType {
		case ir.TypeNumber:
			return ir.TypeNumberArray
		case ir.TypeString:
			return ir.TypeStringArray
		case ir.TypeBool:
			return ir.TypeBoolArray
		case ir.TypeBigInt:
			return ir.TypeBigIntArray
		case ir.TypeUnknown, "never", "object:never":
			return ir.TypeUnknownArray
		default:
			return ir.Type(string(elemType) + "[]")
		}
	}
	if strings.Contains(value, "=>") && !strings.HasPrefix(value, "{") && !strings.Contains(value, "|") && !(strings.Contains(value, "<") && strings.HasSuffix(value, ">")) {
		return ir.TypeClosure
	}
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) || (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return ir.TypeString
	}
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return ir.TypeNumber
	}
	if value == "true" || value == "false" {
		return ir.TypeBool
	}
	if strings.Contains(value, "&") && !strings.Contains(value, "=>") {
		if fields, ok := anonymousObjectFields(value, visited); ok && len(fields) > 0 {
			shapeName := anonymousShapeName(fields)
			if _, ok := anonymousShapes[shapeName]; !ok {
				anonymousShapes[shapeName] = ir.ObjectShape{
					Name:   shapeName,
					Fields: fields,
				}
			}
			return ir.Type("object:" + shapeName)
		}
	}
	if strings.HasSuffix(value, "]") && strings.Contains(value, "[") {
		idx := strings.Index(value, "[")
		objName := strings.TrimSpace(value[:idx])
		propName := strings.Trim(strings.TrimSpace(value[idx+1:len(value)-1]), "\"'")
		if aliased, ok := typeAliasesIndex[objName]; ok {
			if fields, ok := anonymousObjectFields(aliased, visited); ok {
				for _, f := range fields {
					if f.Name == propName {
						return f.Type
					}
				}
			}
		}
		objType := toIRTypeInternal(objName, visited)
		shapeName := strings.TrimPrefix(string(objType), "object:")
		if s, ok := registeredShapes[shapeName]; ok {
			for _, f := range s.Fields {
				if f.Name == propName {
					return f.Type
				}
			}
		} else if s, ok := anonymousShapes[shapeName]; ok {
			for _, f := range s.Fields {
				if f.Name == propName {
					return f.Type
				}
			}
		}
	}
	if strings.TrimSpace(value) == "{}" {
		return ir.TypeUnknown
	}
	if strings.HasPrefix(value, "{") {
		if !strings.HasSuffix(value, "}") {
			value = value + "}"
		}
		if fields, ok := anonymousObjectFields(value, visited); ok {
			name := anonymousShapeName(fields)
			registerAnonymousShape(name, fields)
			return ir.Type("object:" + name)
		}
		return ir.TypeObject
	}
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		if fields, ok := tupleFields(value); ok {
			name := anonymousShapeName(fields)
			registerAnonymousShape(name, fields)
			return ir.Type("object:" + name)
		}
		if strings.Contains(value, "...") {
			return ir.TypeUnknownArray
		}
	}
	if strings.Contains(value, "<") && strings.HasSuffix(value, ">") {
		clean := strings.TrimPrefix(value, "object:")
		idx := strings.Index(clean, "<")
		base := clean[:idx]
		inner := clean[idx+1 : len(clean)-1]
		if base == "Promise" {
			return ir.Type("object:Promise_" + mangleGenericTypeString(inner))
		}
		if base == "Array" {
			return ir.Type(mangleGenericTypeString(inner) + "[]")
		}
		if base == "Map" {
			return ir.TypeMap
		}
		if base == "Set" {
			return ir.TypeSet
		}
		if base == "WeakMap" {
			return ir.Type("object:WeakMap")
		}
		if base == "WeakSet" {
			return ir.Type("object:WeakSet")
		}
		if base == "WeakRef" {
			return ir.Type("object:WeakRef")
		}
		// These utility types are compile-time transformations, not runtime
		// objects. Reduce the common forms before mapping to the native IR.
		typeArgs := splitTypeArguments(inner)
		switch base {
		case "NonNullable":
			if len(typeArgs) == 1 {
				return nonNullishIRType(typeArgs[0])
			}
		case "Exclude":
			if len(typeArgs) == 2 {
				parts := splitTopLevelUnion(typeArgs[0])
				removed := strings.TrimSpace(typeArgs[1])
				kept := make([]string, 0, len(parts))
				for _, part := range parts {
					if strings.TrimSpace(part) != removed {
						kept = append(kept, strings.TrimSpace(part))
					}
				}
				if len(kept) == 0 {
					return ir.TypeVoid
				}
				return toIRType(strings.Join(kept, " | "))
			}
		case "Extract":
			if len(typeArgs) == 2 {
				parts := splitTopLevelUnion(typeArgs[0])
				wanted := strings.TrimSpace(typeArgs[1])
				kept := make([]string, 0, len(parts))
				for _, part := range parts {
					if strings.TrimSpace(part) == wanted {
						kept = append(kept, strings.TrimSpace(part))
					}
				}
				if len(kept) == 0 {
					return ir.TypeVoid
				}
				return toIRType(strings.Join(kept, " | "))
			}
		}
		if base == "IteratorResult" {
			elemType := "number"
			if len(typeArgs) > 0 && typeArgs[0] != "" && typeArgs[0] != "any" && typeArgs[0] != "unknown" {
				elemType = typeArgs[0]
			}
			return toIRType(fmt.Sprintf("{ done: boolean; value: %s; }", elemType))
		}
		switch base {
		case "Uint8Array":
			return ir.TypeUint8Array
		case "Int8Array":
			return ir.TypeInt8Array
		case "Uint8ClampedArray":
			return ir.TypeUint8ClampedArray
		case "Int16Array":
			return ir.TypeInt16Array
		case "Uint16Array":
			return ir.TypeUint16Array
		case "Int32Array":
			return ir.TypeInt32Array
		case "Uint32Array":
			return ir.TypeUint32Array
		case "Float32Array":
			return ir.TypeFloat32Array
		case "Float64Array":
			return ir.TypeFloat64Array
		case "BigInt64Array":
			return ir.TypeBigInt64Array
		case "BigUint64Array":
			return ir.TypeBigUint64Array
		case "ArrayBuffer":
			return ir.TypeArrayBuffer
		case "DataView":
			return ir.TypeDataView
		}
		return ir.Type("object:" + mangleGenericName(base, typeArgs))
	}
	if strings.HasPrefix(value, "object:") {
		trimmed := strings.TrimPrefix(value, "object:")
		if aliased, ok := typeAliasesIndex[trimmed]; ok && aliased != "" && aliased != trimmed {
			return toIRType(aliased)
		}
		switch trimmed {
		case "Buffer":
			return ir.TypeBuffer
		case "Uint8Array":
			return ir.TypeUint8Array
		case "Int8Array":
			return ir.TypeInt8Array
		case "Uint8ClampedArray":
			return ir.TypeUint8ClampedArray
		case "Int16Array":
			return ir.TypeInt16Array
		case "Uint16Array":
			return ir.TypeUint16Array
		case "Int32Array":
			return ir.TypeInt32Array
		case "Uint32Array":
			return ir.TypeUint32Array
		case "Float32Array":
			return ir.TypeFloat32Array
		case "Float64Array":
			return ir.TypeFloat64Array
		case "BigInt64Array":
			return ir.TypeBigInt64Array
		case "BigUint64Array":
			return ir.TypeBigUint64Array
		case "DataView":
			return ir.TypeDataView
		case "ArrayBuffer":
			return ir.TypeArrayBuffer
		case "Map":
			return ir.TypeMap
		case "Set":
			return ir.TypeSet
		case "TextEncoder":
			return ir.TypeTextEncoder
		case "TextDecoder":
			return ir.TypeTextDecoder
		case "RegExpMatchArray", "RegExpExecArray":
			return ir.TypeStringArray
		case "RegExp":
			return ir.Type("object:RegExp")
		}
		return ir.Type(value)
	}
	switch value {
	case "number":
		return ir.TypeNumber
	case "bigint":
		return ir.TypeBigInt
	case "bigint[]":
		return ir.TypeBigIntArray
	case "symbol":
		return ir.TypeSymbol
	case "symbol[]":
		return ir.TypeSymbolArray
	case "RegExpMatchArray", "RegExpExecArray":
		return ir.TypeStringArray
	case "RegExp":
		return ir.Type("object:RegExp")
	case "string":
		return ir.TypeString
	case "null":
		return ir.TypePointer
	case "undefined":
		return ir.TypeVoid
	case "bool", "boolean":
		return ir.TypeBool
	case "bool[]", "boolean[]":
		return ir.TypeBoolArray
	case "number[]":
		return ir.TypeNumberArray
	case "string[]", "TemplateStringsArray":
		return ir.TypeStringArray
	case "closure", "function", "Function":
		return ir.TypeClosure
	case "unknown", "any":
		return ir.TypeUnknown
	case "unknown[]", "any[]":
		return ir.TypeUnknownArray
	case "Uint8Array":
		return ir.TypeUint8Array
	case "Int8Array":
		return ir.TypeInt8Array
	case "Uint8ClampedArray":
		return ir.TypeUint8ClampedArray
	case "Int16Array":
		return ir.TypeInt16Array
	case "Uint16Array":
		return ir.TypeUint16Array
	case "Int32Array":
		return ir.TypeInt32Array
	case "Uint32Array":
		return ir.TypeUint32Array
	case "Float32Array":
		return ir.TypeFloat32Array
	case "Float64Array":
		return ir.TypeFloat64Array
	case "BigInt64Array":
		return ir.TypeBigInt64Array
	case "BigUint64Array":
		return ir.TypeBigUint64Array
	case "DataView":
		return ir.TypeDataView
	case "ArrayBuffer", "SharedArrayBuffer":
		return ir.TypeArrayBuffer
	case "Map":
		return ir.TypeMap
	case "Set":
		return ir.TypeSet
	case "WeakMap":
		return ir.Type("object:WeakMap")
	case "WeakSet":
		return ir.Type("object:WeakSet")
	case "WeakRef":
		return ir.Type("object:WeakRef")
	case "FinalizationRegistry":
		return ir.Type("object:FinalizationRegistry")
	case "TextEncoder":
		return ir.TypeTextEncoder
	case "TextDecoder":
		return ir.TypeTextDecoder
	case "Buffer":
		return ir.TypeBuffer
	case "void", "":
		return ir.TypeVoid
	default:
		if strings.HasPrefix(value, "Intl.") {
			return ir.Type("object:" + value)
		}
		if strings.HasSuffix(value, "n") && len(value) > 1 {
			if _, err := strconv.ParseInt(value[:len(value)-1], 10, 64); err == nil {
				return ir.TypeBigInt
			}
		}
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			return ir.TypeNumber
		}
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) || (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			return ir.TypeString
		}
		if value == "true" || value == "false" {
			return ir.TypeBool
		}
		if value == "unique symbol" {
			return ir.TypeSymbol
		}
		if strings.HasPrefix(value, "Map<") && strings.HasSuffix(value, ">") {
			return ir.TypeMap
		}
		if strings.HasPrefix(value, "Set<") && strings.HasSuffix(value, ">") {
			return ir.TypeSet
		}
		if before, ok := strings.CutSuffix(value, "[]"); ok {
			elem := before
			return ir.Type(string(toIRType(elem)) + "[]")
		}
		if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			if fields, ok := tupleFields(value); ok {
				name := anonymousShapeName(fields)
				registerAnonymousShape(name, fields)
				return ir.Type("object:" + name)
			}
		}
		if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
			if fields, ok := anonymousObjectFields(value, nil); ok {
				name := anonymousShapeName(fields)
				registerAnonymousShape(name, fields)
				return ir.Type("object:" + name)
			}
			return ir.TypeObject
		}
		if fields, ok := parseFlattenedShapeFields(value); ok {
			name := anonymousShapeName(fields)
			registerAnonymousShape(name, fields)
			return ir.Type("object:" + name)
		}
		if strings.Contains(value, "=>") || strings.HasPrefix(value, "(") || strings.HasPrefix(value, "Function") {
			return ir.TypeClosure
		}
		if strings.HasPrefix(value, "object:") {
			return ir.Type(value)
		}
		return ir.Type("object:" + value)
	}
}

func mangledTypedArrayType(value string) (ir.Type, bool) {
	clean := strings.TrimPrefix(value, "object:")
	for name, typ := range map[string]ir.Type{
		"Int8Array":         ir.TypeInt8Array,
		"Uint8Array":        ir.TypeUint8Array,
		"Uint8ClampedArray": ir.TypeUint8ClampedArray,
		"Int16Array":        ir.TypeInt16Array,
		"Uint16Array":       ir.TypeUint16Array,
		"Int32Array":        ir.TypeInt32Array,
		"Uint32Array":       ir.TypeUint32Array,
		"Float32Array":      ir.TypeFloat32Array,
		"Float64Array":      ir.TypeFloat64Array,
		"BigInt64Array":     ir.TypeBigInt64Array,
		"BigUint64Array":    ir.TypeBigUint64Array,
	} {
		if clean == name || strings.HasPrefix(clean, name+"_") {
			return typ, true
		}
	}
	return "", false
}

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

func splitTopLevel(s string) []string {
	var parts []string
	var cur strings.Builder
	depthBrace := 0
	depthBracket := 0
	depthAngle := 0
	depthParen := 0

	for _, r := range s {
		switch r {
		case '{':
			depthBrace++
			cur.WriteRune(r)
		case '}':
			if depthBrace > 0 {
				depthBrace--
			}
			cur.WriteRune(r)
		case '[':
			depthBracket++
			cur.WriteRune(r)
		case ']':
			if depthBracket > 0 {
				depthBracket--
			}
			cur.WriteRune(r)
		case '<':
			depthAngle++
			cur.WriteRune(r)
		case '>':
			if depthAngle > 0 {
				depthAngle--
			}
			cur.WriteRune(r)
		case '(':
			depthParen++
			cur.WriteRune(r)
		case ')':
			if depthParen > 0 {
				depthParen--
			}
			cur.WriteRune(r)
		case ';', ',':
			if depthBrace == 0 && depthBracket == 0 && depthAngle == 0 && depthParen == 0 {
				if cur.Len() > 0 {
					parts = append(parts, strings.TrimSpace(cur.String()))
					cur.Reset()
				}
			} else {
				cur.WriteRune(r)
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		trimmed := strings.TrimSpace(cur.String())
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitTopLevelUnion(s string) []string {
	var parts []string
	var cur strings.Builder
	depthBrace := 0
	depthBracket := 0
	depthAngle := 0
	depthParen := 0

	for _, r := range s {
		switch r {
		case '{':
			depthBrace++
			cur.WriteRune(r)
		case '}':
			depthBrace--
			cur.WriteRune(r)
		case '[':
			depthBracket++
			cur.WriteRune(r)
		case ']':
			depthBracket--
			cur.WriteRune(r)
		case '<':
			depthAngle++
			cur.WriteRune(r)
		case '>':
			depthAngle--
			cur.WriteRune(r)
		case '(':
			depthParen++
			cur.WriteRune(r)
		case ')':
			depthParen--
			cur.WriteRune(r)
		case '|':
			if depthBrace == 0 && depthBracket == 0 && depthAngle == 0 && depthParen == 0 {
				if cur.Len() > 0 {
					trimmed := strings.TrimSpace(cur.String())
					if trimmed != "" {
						parts = append(parts, trimmed)
					}
					cur.Reset()
				}
			} else {
				cur.WriteRune(r)
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		trimmed := strings.TrimSpace(cur.String())
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitTopLevelIntersection(s string) []string {
	var parts []string
	var cur strings.Builder
	depthBrace := 0
	depthBracket := 0
	depthAngle := 0
	depthParen := 0

	for _, r := range s {
		switch r {
		case '{':
			depthBrace++
			cur.WriteRune(r)
		case '}':
			depthBrace--
			cur.WriteRune(r)
		case '[':
			depthBracket++
			cur.WriteRune(r)
		case ']':
			depthBracket--
			cur.WriteRune(r)
		case '<':
			depthAngle++
			cur.WriteRune(r)
		case '>':
			depthAngle--
			cur.WriteRune(r)
		case '(':
			depthParen++
			cur.WriteRune(r)
		case ')':
			depthParen--
			cur.WriteRune(r)
		case '&':
			if depthBrace == 0 && depthBracket == 0 && depthAngle == 0 && depthParen == 0 {
				if cur.Len() > 0 {
					trimmed := strings.TrimSpace(cur.String())
					if trimmed != "" {
						parts = append(parts, trimmed)
					}
					cur.Reset()
				}
			} else {
				cur.WriteRune(r)
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		trimmed := strings.TrimSpace(cur.String())
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
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

func toIRSpan(path string, span typescriptgo.SourceSpan) ir.SourceSpan {
	return ir.SourceSpan{Path: path, Offset: span.Start, Length: span.Length}
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
		if strings.HasPrefix(clean, "__shape_0_") {
			return false
		}
		return true
	}
	meta, ok := classHierarchy[clean]
	return ok && (meta.IsInterface || meta.IsTypeAlias)
}

func sourceError(path string, span typescriptgo.SourceSpan, err error) error {
	if err == nil {
		return nil
	}
	return &DiagnosticError{Path: path, Offset: span.Start, Length: span.Length, Err: err}
}

type DiagnosticError struct {
	Path   string
	Offset int
	Length int
	Err    error
}

func (e *DiagnosticError) Error() string {
	return e.Err.Error()
}

func isNumberTypedArray(t ir.Type) bool {
	switch t {
	case ir.TypeBuffer, ir.TypeInt8Array, ir.TypeUint8Array, ir.TypeUint8ClampedArray,
		ir.TypeInt16Array, ir.TypeUint16Array, ir.TypeInt32Array, ir.TypeUint32Array,
		ir.TypeFloat32Array, ir.TypeFloat64Array:
		return true
	default:
		return false
	}
}

func isTypedArrayType(t ir.Type) bool {
	switch t {
	case ir.TypeBuffer, ir.TypeInt8Array, ir.TypeUint8Array, ir.TypeUint8ClampedArray,
		ir.TypeInt16Array, ir.TypeUint16Array, ir.TypeInt32Array, ir.TypeUint32Array,
		ir.TypeFloat32Array, ir.TypeFloat64Array, ir.TypeBigInt64Array, ir.TypeBigUint64Array:
		return true
	default:
		return false
	}
}

// ArrayBufferView is a TypeScript union of typed arrays and DataView. The
// frontend keeps that alias name when it cannot collapse the union; property
// lowering routes its shared fields through a tag-aware ABI.
func isArrayBufferViewType(t ir.Type) bool {
	if t == "ArrayBufferView" || t == "object:ArrayBufferView" {
		return true
	}
	return isTypedArrayType(t) || t == ir.TypeDataView
}

func isMapType(t ir.Type) bool {
	return t == ir.TypeMap || t == "object:Map" || strings.HasPrefix(string(t), "object:Map<") || strings.HasPrefix(string(t), "object:Map__")
}

func isSetType(t ir.Type) bool {
	return t == ir.TypeSet || t == "object:Set" || strings.HasPrefix(string(t), "object:Set<") || strings.HasPrefix(string(t), "object:Set__")
}

func statementAlwaysReturns(stmt typescriptgo.SyntaxStatement) bool {
	switch stmt.Kind {
	case "return", "throw":
		return true
	case "block":
		return slices.ContainsFunc(stmt.Body, statementAlwaysReturns)
	case "if":
		if len(stmt.Then) == 0 || len(stmt.Else) == 0 {
			return false
		}
		thenReturns := slices.ContainsFunc(stmt.Then, statementAlwaysReturns)
		elseReturns := slices.ContainsFunc(stmt.Else, statementAlwaysReturns)
		return thenReturns && elseReturns
	case "switch":
		hasDefault := false
		fallthroughReturns := false
		for i := len(stmt.Cases) - 1; i >= 0; i-- {
			c := stmt.Cases[i]
			if c.Expression == nil {
				hasDefault = true
			}
			caseReturns := slices.ContainsFunc(c.Statements, statementAlwaysReturns)
			if len(c.Statements) == 0 {
				caseReturns = fallthroughReturns
			}
			if !caseReturns {
				return false
			}
			fallthroughReturns = caseReturns
		}
		return hasDefault
	case "try":
		bodyReturns := slices.ContainsFunc(stmt.Body, statementAlwaysReturns)
		catchReturns := len(stmt.Catch) > 0 && slices.ContainsFunc(stmt.Catch, statementAlwaysReturns)
		finallyReturns := len(stmt.Finally) > 0 && slices.ContainsFunc(stmt.Finally, statementAlwaysReturns)
		return finallyReturns || (bodyReturns && catchReturns)
	default:
		return false
	}
}

func isReturningClosure(stmt typescriptgo.SyntaxStatement) bool {
	if stmt.Expression != nil && (stmt.Expression.Kind == "arrow_function" || stmt.Expression.Kind == "function") {
		return true
	}
	for _, s := range stmt.Body {
		if s.Kind == "return" && s.Expression != nil {
			if s.Expression.Kind == "arrow_function" || s.Expression.Kind == "function" || s.Expression.Function != nil {
				return true
			}
		}
	}
	if stmt.Expression != nil && stmt.Expression.Function != nil {
		return isReturningClosure(*stmt.Expression.Function)
	}
	return false
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
