package lowering

import (
	"fmt"
	"strconv"
	"strings"

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
			hasVoid := false
			for _, p := range parts {
				trimmed := strings.TrimSpace(p)
				if trimmed == "void" {
					hasVoid = true
				} else if trimmed != "null" && trimmed != "undefined" && trimmed != "" {
					nonNullish = append(nonNullish, trimmed)
				}
			}
			if len(nonNullish) == 0 {
				return ir.TypeVoid
			}
			if hasVoid {
				return ir.TypeUnknown
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
		if base == "IteratorObject" {
			// Iterator helpers run over a materialized copy of the sequence,
			// so an IteratorObject<T> is stored as a T[] (see calls_iterator.go).
			elemType := "unknown"
			if len(typeArgs) > 0 && typeArgs[0] != "" && typeArgs[0] != "any" {
				elemType = typeArgs[0]
			}
			return toIRType(elemType + "[]")
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
