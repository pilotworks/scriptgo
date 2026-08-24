package llvm

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func consoleRuntimeName(method string, typ ir.Type) (string, bool) {
	if method == "dir" || method == "dirxml" || method == "table" || method == "trace" {
		method = "log"
	}
	if method != "log" && method != "info" && method != "debug" && method != "warn" && method != "error" {
		return "", false
	}
	suffix := map[ir.Type]string{
		ir.TypeNumber:  "number",
		ir.TypeBigInt:  "bigint",
		ir.TypeSymbol:  "symbol",
		ir.TypeString:  "string",
		ir.TypeBool:    "bool",
		ir.TypeUnknown: "unknown",
	}[typ]
	if suffix == "" {
		if strings.HasPrefix(string(typ), "object:") || typ == ir.TypeObject || typ == ir.TypeClosure || strings.HasSuffix(string(typ), "[]") {
			return "scriptgo_console_" + method + "_object", true
		}
		return "", false
	}
	return "scriptgo_console_" + method + "_" + suffix, true
}

func llvmType(typ ir.Type) string {
	switch typ {
	case ir.TypeNumber:
		return "double"
	case ir.TypeBigInt:
		return "i64"
	case ir.TypeSymbol:
		return "ptr"
	case ir.TypeString:
		return "ptr"
	case ir.TypeNumberArray, ir.TypeStringArray, ir.TypeBigIntArray, ir.TypeSymbolArray, ir.TypeBoolArray,
		ir.TypeBuffer, ir.TypeUint8Array, ir.TypeInt8Array, ir.TypeUint8ClampedArray,
		ir.TypeInt16Array, ir.TypeUint16Array, ir.TypeInt32Array, ir.TypeUint32Array,
		ir.TypeFloat32Array, ir.TypeFloat64Array, ir.TypeBigInt64Array, ir.TypeBigUint64Array,
		ir.TypeDataView, ir.TypeArrayBuffer, ir.TypePointer:
		return "ptr"
	case ir.TypeClosure:
		return "ptr"
	case ir.TypeBool:
		return "i1"
	case ir.TypeVoid:
		return "void"
	case ir.TypeObject:
		return "ptr"
	case ir.TypeUnknown, "any":
		return "{ i32, i32, i64 }"
	default:
		s := string(typ)
		if strings.HasPrefix(s, string(ir.TypeObject)+":") || strings.HasPrefix(s, "object:") ||
			strings.HasSuffix(s, "[]") || strings.HasPrefix(s, "tuple:") ||
			s == "Map" || s == "Set" || s == "Promise" || s == "Date" || s == "RegExp" ||
			s == "Error" || s == "TypeError" || s == "RangeError" || s == "SyntaxError" ||
			strings.HasPrefix(s, "Map<") || strings.HasPrefix(s, "Set<") ||
			strings.HasPrefix(s, "Promise<") || strings.HasPrefix(s, "Promise_") ||
			strings.HasPrefix(s, "Iterator") || strings.HasPrefix(s, "AsyncIterator") ||
			strings.HasPrefix(s, "Generator") || strings.HasPrefix(s, "AsyncGenerator") {
			return "ptr"
		}
		return s
	}
}

func arrayElementType(arrayType ir.Type) ir.Type {
	str := string(arrayType)
	if before, ok := strings.CutSuffix(str, "[]"); ok {
		elem := before
		if elem == "boolean" {
			return ir.TypeBool
		}
		return ir.Type(elem)
	}
	if arrayType == ir.TypeStringArray {
		return ir.TypeString
	}
	if arrayType == ir.TypeBoolArray {
		return ir.TypeBool
	}
	return ir.TypeNumber
}

func arrayElementSize(arrayType ir.Type) (int64, error) {
	elem := arrayElementType(arrayType)
	switch elem {
	case ir.TypeBool:
		return 1, nil
	case ir.TypeNumber:
		return 8, nil
	case ir.TypeString, ir.TypeObject:
		return 8, nil
	case ir.TypeUnknown, "any":
		return 16, nil
	default:
		if strings.HasPrefix(string(elem), "object:") || strings.HasSuffix(string(elem), "[]") {
			return 8, nil
		}
		return 8, nil
	}
}

func llvmNumber(value float64) string {
	if math.IsNaN(value) {
		return "0x7FF8000000000000"
	}
	if math.IsInf(value, 1) {
		return "0x7FF0000000000000"
	}
	if math.IsInf(value, -1) {
		return "0xFFF0000000000000"
	}
	return strconv.FormatFloat(value, 'e', 17, 64)
}

func escapeString(value string) string {
	var out strings.Builder
	for _, b := range []byte(value) {
		if b >= 32 && b <= 126 && b != '"' && b != '\\' {
			out.WriteByte(b)
			continue
		}
		out.WriteString(fmt.Sprintf("\\%02X", b))
	}
	return out.String()
}
