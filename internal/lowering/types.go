package lowering

import (
	"slices"
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

var typeAliasesIndex = map[string]string{}

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

func toIRType(value string) ir.Type {
	value = strings.TrimSpace(value)
	if aliased, ok := typeAliasesIndex[value]; ok && aliased != value {
		return toIRType(aliased)
	}
	base := value
	if idx := strings.Index(base, "__"); idx != -1 {
		base = base[:idx]
	}
	if idx := strings.Index(base, "<"); idx != -1 {
		base = base[:idx]
	}
	if aliased, ok := typeAliasesIndex[base]; ok && aliased != base {
		if strings.Contains(aliased, "=>") {
			return ir.TypeClosure
		}
	}
	if strings.Contains(value, "=>") {
		return ir.TypeClosure
	}
	if strings.HasSuffix(value, "_arr") {
		elem := strings.TrimSuffix(value, "_arr")
		return toIRType(elem + "[]")
	}
	if strings.HasSuffix(value, "[]") {
		elem := strings.TrimSuffix(value, "[]")
		elemType := toIRType(elem)
		switch elemType {
		case ir.TypeNumber:
			return ir.TypeNumberArray
		case ir.TypeString:
			return ir.TypeStringArray
		case ir.TypeBool:
			return ir.TypeBoolArray
		case ir.TypeBigInt:
			return ir.TypeBigIntArray
		default:
			return ir.Type(string(elemType) + "[]")
		}
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
	if strings.Contains(value, "|") {
		parts := strings.Split(value, "|")
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
		return toIRType(nonNullish[0])
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
		typeArgs := splitTypeArguments(inner)
		return ir.Type("object:" + mangleGenericName(base, typeArgs))
	}
	if strings.HasPrefix(value, "object:") {
		trimmed := strings.TrimPrefix(value, "object:")
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
	case "RegExp":
		return ir.Type("object:RegExp")
	case "string", "null", "undefined":
		return ir.TypeString
	case "bool", "boolean":
		return ir.TypeBool
	case "bool[]", "boolean[]":
		return ir.TypeBoolArray
	case "number[]":
		return ir.TypeNumberArray
	case "string[]", "TemplateStringsArray":
		return ir.TypeStringArray
	case "closure", "function":
		return ir.TypeClosure
	case "unknown":
		return ir.TypeUnknown
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
	case "WeakMap":
		return ir.Type("object:WeakMap")
	case "WeakSet":
		return ir.Type("object:WeakSet")
	case "WeakRef":
		return ir.Type("object:WeakRef")
	case "TextEncoder":
		return ir.TypeTextEncoder
	case "TextDecoder":
		return ir.TypeTextDecoder
	case "Buffer":
		return ir.TypeBuffer
	case "void", "any", "":
		return ir.TypeVoid
	default:
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
			if fields, ok := anonymousObjectFields(value); ok {
				name := anonymousShapeName(fields)
				registerAnonymousShape(name, fields)
				return ir.Type("object:" + name)
			}
		}
		if strings.Contains(value, "=>") || strings.HasPrefix(value, "(") || strings.HasPrefix(value, "Function") {
			return ir.TypeClosure
		}
		return ir.Type("object:" + value)
	}
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

func tupleFields(typeStr string) ([]ir.Field, bool) {
	if !strings.HasPrefix(typeStr, "[") || !strings.HasSuffix(typeStr, "]") {
		return nil, false
	}
	inner := strings.TrimSpace(typeStr[1 : len(typeStr)-1])
	if inner == "" {
		return nil, false
	}
	parts := splitTopLevel(inner)
	var fields []ir.Field
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		if idx := strings.Index(trimmed, ":"); idx != -1 {
			trimmed = strings.TrimSpace(trimmed[idx+1:])
		}
		trimmed = strings.TrimPrefix(trimmed, "...")
		trimmed = strings.TrimSuffix(trimmed, "[]")
		trimmed = strings.TrimSuffix(trimmed, "?")
		elemType := toIRType(trimmed)
		fields = append(fields, ir.Field{
			Name: strconv.Itoa(i),
			Type: elemType,
		})
	}
	return fields, true
}

func anonymousObjectFields(typeStr string) ([]ir.Field, bool) {
	if !strings.HasPrefix(typeStr, "{") || !strings.HasSuffix(typeStr, "}") {
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
		fName = strings.TrimSuffix(fName, "?")
		fTypeStr := strings.TrimSpace(trimmed[colonIdx+1:])
		fields = append(fields, ir.Field{
			Name: fName,
			Type: toIRType(fTypeStr),
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

func isMapType(t ir.Type) bool {
	return t == ir.TypeMap
}

func isSetType(t ir.Type) bool {
	return t == ir.TypeSet
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
	default:
		return false
	}
}
