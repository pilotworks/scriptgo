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
		ir.TypeBuffer:  "buffer",
	}[typ]
	if suffix == "" {
		if typ == ir.TypeVoid {
			return "scriptgo_console_" + method + "_string", true
		}
		if strings.HasPrefix(string(typ), "object:") || typ == ir.TypeObject || typ == ir.TypeClosure || strings.HasSuffix(string(typ), "[]") || typ == ir.TypePointer || typ == "ptr" {
			return "scriptgo_console_" + method + "_object", true
		}
		return "", false
	}
	return "scriptgo_console_" + method + "_" + suffix, true
}

func llvmType(typ ir.Type) string {
	switch typ {
	case ir.TypeNumber, "double":
		return "double"
	case ir.TypeBigInt, "i64":
		return "i64"
	case ir.TypeBool, "i1", "boolean":
		return "i1"
	case ir.TypeVoid:
		return "void"
	case ir.TypeUnknown, "{ i32, i32, i64 }":
		return "{ i32, i32, i64 }"
	default:
		return "ptr"
	}
}

func arrayElementType(arrayType ir.Type) ir.Type {
	str := string(arrayType)
	if before, ok := strings.CutSuffix(str, "[]"); ok {
		elem := before
		if elem == "boolean" {
			return ir.TypeBool
		}
		if elem == "void" || elem == "undefined" || elem == "unknown" {
			return ir.TypeUnknown
		}
		if elem == "never" || elem == "object:never" || strings.HasPrefix(elem, "never") || strings.HasPrefix(elem, "object:never") {
			return ir.TypeObject
		}
		return ir.Type(elem)
	}
	if arrayType == ir.TypeStringArray {
		return ir.TypeString
	}
	if arrayType == ir.TypeBoolArray {
		return ir.TypeBool
	}
	if arrayType == ir.TypeUnknown {
		return ir.TypeUnknown
	}
	return ir.TypeNumber
}

func arrayElementSizeForTarget(arrayType ir.Type, ptrSize int64) (int64, error) {
	if ptrSize <= 0 {
		ptrSize = 8
	}
	elem := arrayElementType(arrayType)
	switch elem {
	case ir.TypeBool:
		return 1, nil
	case ir.TypeNumber:
		return 8, nil
	case ir.TypeString, ir.TypeObject, ir.TypeClosure:
		return ptrSize, nil
	case ir.TypeUnknown:
		return 16, nil
	default:
		if strings.HasPrefix(string(elem), "object:") || strings.HasSuffix(string(elem), "[]") {
			return ptrSize, nil
		}
		return ptrSize, nil
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
