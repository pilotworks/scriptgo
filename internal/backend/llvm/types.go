package llvm

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func consoleRuntimeName(method string, typ ir.Type) (string, bool) {
	if method != "log" && method != "info" && method != "warn" && method != "error" {
		return "", false
	}
	suffix := map[ir.Type]string{
		ir.TypeNumber: "number",
		ir.TypeString: "string",
		ir.TypeBool:   "bool",
	}[typ]
	if suffix == "" {
		return "", false
	}
	return "scriptgo_console_" + method + "_" + suffix, true
}

func llvmType(typ ir.Type) string {
	switch typ {
	case ir.TypeNumber:
		return "double"
	case ir.TypeString:
		return "ptr"
	case ir.TypeNumberArray, ir.TypeStringArray:
		return "ptr"
	case ir.TypeClosure:
		return "ptr"
	case ir.TypeBool:
		return "i1"
	case ir.TypeVoid:
		return "void"
	case ir.TypeObject:
		return "ptr"
	default:
		if strings.HasPrefix(string(typ), string(ir.TypeObject)+":") {
			return "ptr"
		}
		return string(typ)
	}
}

func arrayElementType(arrayType ir.Type) ir.Type {
	str := string(arrayType)
	if strings.HasSuffix(str, "[]") {
		return ir.Type(strings.TrimSuffix(str, "[]"))
	}
	if arrayType == ir.TypeStringArray {
		return ir.TypeString
	}
	return ir.TypeNumber
}

func arrayElementSize(arrayType ir.Type) (int64, error) {
	switch arrayType {
	case ir.TypeNumberArray:
		return 8, nil // IEEE-754 binary64.
	case ir.TypeStringArray:
		return 8, nil // v1 targets use 64-bit opaque pointers.
	default:
		if strings.HasSuffix(string(arrayType), "[]") {
			return 8, nil
		}
		return 0, fmt.Errorf("unsupported array element layout %s", arrayType)
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
