package lowering

import (
	"slices"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

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

func statementAlwaysReturns(stmt frontend.SyntaxStatement) bool {
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

func isReturningClosure(stmt frontend.SyntaxStatement) bool {
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
