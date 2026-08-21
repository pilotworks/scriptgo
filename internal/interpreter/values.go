package interpreter

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

type Closure struct {
	Function ir.Function
	Env      map[string]Value
}

type ArrayBuffer struct {
	Data []byte
}

type TypedArray struct {
	Kind       ir.Type
	Buffer     *ArrayBuffer
	ByteOffset int
	Length     int
}

func (ta *TypedArray) ElementSize() int {
	switch ta.Kind {
	case ir.TypeInt32Array:
		return 4
	case ir.TypeFloat64Array:
		return 8
	default:
		return 1
	}
}

func (ta *TypedArray) ByteLength() int {
	return ta.Length * ta.ElementSize()
}

func (ta *TypedArray) Get(index int) float64 {
	if index < 0 || index >= ta.Length || ta.Buffer == nil {
		return 0
	}
	offset := ta.ByteOffset + index*ta.ElementSize()
	if offset+ta.ElementSize() > len(ta.Buffer.Data) {
		return 0
	}
	switch ta.Kind {
	case ir.TypeUint8Array:
		return float64(ta.Buffer.Data[offset])
	case ir.TypeInt32Array:
		u := uint32(ta.Buffer.Data[offset]) | uint32(ta.Buffer.Data[offset+1])<<8 | uint32(ta.Buffer.Data[offset+2])<<16 | uint32(ta.Buffer.Data[offset+3])<<24
		return float64(int32(u))
	case ir.TypeFloat64Array:
		u := uint64(ta.Buffer.Data[offset]) | uint64(ta.Buffer.Data[offset+1])<<8 | uint64(ta.Buffer.Data[offset+2])<<16 | uint64(ta.Buffer.Data[offset+3])<<24 |
			uint64(ta.Buffer.Data[offset+4])<<32 | uint64(ta.Buffer.Data[offset+5])<<40 | uint64(ta.Buffer.Data[offset+6])<<48 | uint64(ta.Buffer.Data[offset+7])<<56
		return math.Float64frombits(u)
	default:
		return float64(ta.Buffer.Data[offset])
	}
}

func (ta *TypedArray) Set(index int, val float64) {
	if index < 0 || index >= ta.Length || ta.Buffer == nil {
		return
	}
	offset := ta.ByteOffset + index*ta.ElementSize()
	if offset+ta.ElementSize() > len(ta.Buffer.Data) {
		return
	}
	switch ta.Kind {
	case ir.TypeUint8Array:
		ta.Buffer.Data[offset] = byte(uint32(val) & 0xff)
	case ir.TypeInt32Array:
		i32 := uint32(int32(val))
		ta.Buffer.Data[offset] = byte(i32)
		ta.Buffer.Data[offset+1] = byte(i32 >> 8)
		ta.Buffer.Data[offset+2] = byte(i32 >> 16)
		ta.Buffer.Data[offset+3] = byte(i32 >> 24)
	case ir.TypeFloat64Array:
		bits := math.Float64bits(val)
		ta.Buffer.Data[offset] = byte(bits)
		ta.Buffer.Data[offset+1] = byte(bits >> 8)
		ta.Buffer.Data[offset+2] = byte(bits >> 16)
		ta.Buffer.Data[offset+3] = byte(bits >> 24)
		ta.Buffer.Data[offset+4] = byte(bits >> 32)
		ta.Buffer.Data[offset+5] = byte(bits >> 40)
		ta.Buffer.Data[offset+6] = byte(bits >> 48)
		ta.Buffer.Data[offset+7] = byte(bits >> 56)
	}
}

type Value struct {
	Type       ir.Type
	Number     float64
	BigInt     int64
	SymbolID   uint64
	SymbolDesc string
	String     string
	Bool       bool
	Array      []Value
	ArrayRef   *[]Value
	Object     map[string]Value
	Closure    *Closure
	Boxed      *Value
	Buffer     *ArrayBuffer
	TypedArray *TypedArray
}

func (v Value) GetArray() []Value {
	if v.ArrayRef != nil {
		return *v.ArrayRef
	}
	return v.Array
}

func (v *Value) SetArray(arr []Value) {
	if v.ArrayRef != nil {
		*v.ArrayRef = arr
	} else {
		ref := new([]Value)
		*ref = arr
		v.ArrayRef = ref
	}
	v.Array = arr
}

type Result struct {
	Output string
	Return Value
}

func parseConstant(typ ir.Type, value string) (Value, error) {
	switch typ {
	case ir.TypeNumber:
		if value == "NaN" {
			return Value{Type: typ, Number: math.NaN()}, nil
		}
		if value == "+Inf" {
			return Value{Type: typ, Number: math.Inf(1)}, nil
		}
		number, err := strconv.ParseFloat(value, 64)
		return Value{Type: typ, Number: number}, err
	case ir.TypeBigInt:
		val := strings.TrimSuffix(value, "n")
		bi, err := strconv.ParseInt(val, 10, 64)
		return Value{Type: typ, BigInt: bi}, err
	case ir.TypeSymbol:
		return getOrCreateWellKnownSymbol(value), nil
	case ir.TypeString:
		return Value{Type: typ, String: value}, nil
	case ir.TypeBool:
		boolean, err := strconv.ParseBool(value)
		return Value{Type: typ, Bool: boolean}, err
	default:
		if strings.HasPrefix(string(typ), "object:") || typ == "ptr" || typ == ir.TypeVoid {
			return Value{Type: typ}, nil
		}
		return Value{}, fmt.Errorf("unsupported constant type %s", typ)
	}
}

func binary(operator string, left, right Value) (Value, error) {
	if left.Type != right.Type {
		return Value{}, fmt.Errorf("binary operands have types %s and %s", left.Type, right.Type)
	}
	if left.Type == ir.TypeString {
		switch operator {
		case "+":
			return Value{Type: ir.TypeString, String: left.String + right.String}, nil
		case "||":
			if left.String != "" {
				return left, nil
			}
			return right, nil
		case "&&":
			if left.String == "" {
				return left, nil
			}
			return right, nil
		default:
			return Value{}, fmt.Errorf("operator %q is unsupported for string", operator)
		}
	}
	if left.Type == ir.TypeBool {
		switch operator {
		case "&&":
			return Value{Type: ir.TypeBool, Bool: left.Bool && right.Bool}, nil
		case "||":
			return Value{Type: ir.TypeBool, Bool: left.Bool || right.Bool}, nil
		default:
			return Value{}, fmt.Errorf("operator %q is unsupported for bool", operator)
		}
	}
	if left.Type == ir.TypeBigInt {
		switch operator {
		case "+":
			return Value{Type: ir.TypeBigInt, BigInt: left.BigInt + right.BigInt}, nil
		case "-":
			return Value{Type: ir.TypeBigInt, BigInt: left.BigInt - right.BigInt}, nil
		case "*":
			return Value{Type: ir.TypeBigInt, BigInt: left.BigInt * right.BigInt}, nil
		case "/":
			if right.BigInt == 0 {
				return Value{}, fmt.Errorf("division by zero")
			}
			return Value{Type: ir.TypeBigInt, BigInt: left.BigInt / right.BigInt}, nil
		case "%":
			if right.BigInt == 0 {
				return Value{}, fmt.Errorf("division by zero")
			}
			return Value{Type: ir.TypeBigInt, BigInt: left.BigInt % right.BigInt}, nil
		case "&":
			return Value{Type: ir.TypeBigInt, BigInt: left.BigInt & right.BigInt}, nil
		case "|":
			return Value{Type: ir.TypeBigInt, BigInt: left.BigInt | right.BigInt}, nil
		case "^":
			return Value{Type: ir.TypeBigInt, BigInt: left.BigInt ^ right.BigInt}, nil
		case "<<":
			return Value{Type: ir.TypeBigInt, BigInt: left.BigInt << uint(right.BigInt)}, nil
		case ">>":
			return Value{Type: ir.TypeBigInt, BigInt: left.BigInt >> uint(right.BigInt)}, nil
		default:
			return Value{}, fmt.Errorf("operator %q is unsupported for bigint", operator)
		}
	}
	if left.Type != ir.TypeNumber {
		return Value{}, fmt.Errorf("operator %q is unsupported for %s", operator, left.Type)
	}
	switch operator {
	case "+":
		return Value{Type: ir.TypeNumber, Number: left.Number + right.Number}, nil
	case "-":
		return Value{Type: ir.TypeNumber, Number: left.Number - right.Number}, nil
	case "*":
		return Value{Type: ir.TypeNumber, Number: left.Number * right.Number}, nil
	case "/":
		return Value{Type: ir.TypeNumber, Number: left.Number / right.Number}, nil
	case "%":
		return Value{Type: ir.TypeNumber, Number: float64(int64(left.Number) % int64(right.Number))}, nil
	case "**":
		return Value{Type: ir.TypeNumber, Number: math.Pow(left.Number, right.Number)}, nil
	case "&":
		return Value{Type: ir.TypeNumber, Number: float64(int32(int64(left.Number)) & int32(int64(right.Number)))}, nil
	case "|":
		return Value{Type: ir.TypeNumber, Number: float64(int32(int64(left.Number)) | int32(int64(right.Number)))}, nil
	case "^":
		return Value{Type: ir.TypeNumber, Number: float64(int32(int64(left.Number)) ^ int32(int64(right.Number)))}, nil
	case "<<":
		shift := uint32(uint64(int64(right.Number))) & 0x1F
		return Value{Type: ir.TypeNumber, Number: float64(int32(int64(left.Number)) << shift)}, nil
	case ">>":
		shift := uint32(uint64(int64(right.Number))) & 0x1F
		return Value{Type: ir.TypeNumber, Number: float64(int32(int64(left.Number)) >> shift)}, nil
	case ">>>":
		shift := uint32(uint64(int64(right.Number))) & 0x1F
		return Value{Type: ir.TypeNumber, Number: float64(uint32(uint64(int64(left.Number))) >> shift)}, nil
	case "||":
		if left.Number != 0 {
			return left, nil
		}
		return right, nil
	case "&&":
		if left.Number == 0 {
			return left, nil
		}
		return right, nil
	default:
		return Value{}, fmt.Errorf("operator %q is unsupported for number", operator)
	}
}

func compare(operator string, left, right Value) (Value, error) {
	if left.Type != right.Type {
		if (strings.HasPrefix(string(left.Type), "object:") || left.Type == "ptr") && (strings.HasPrefix(string(right.Type), "object:") || right.Type == "ptr") {
			leftNull := len(left.Object) == 0 && left.Boxed == nil && len(left.GetArray()) == 0
			rightNull := len(right.Object) == 0 && right.Boxed == nil && len(right.GetArray()) == 0
			switch operator {
			case "==", "===":
				return Value{Type: ir.TypeBool, Bool: leftNull == rightNull}, nil
			case "!=", "!==":
				return Value{Type: ir.TypeBool, Bool: leftNull != rightNull}, nil
			}
		}
		return Value{}, fmt.Errorf("compare operands have types %s and %s", left.Type, right.Type)
	}
	var result bool
	switch left.Type {
	case ir.TypeNumber:
		switch operator {
		case "==", "===":
			result = left.Number == right.Number
		case "!=", "!==":
			result = left.Number != right.Number
		case "<":
			result = left.Number < right.Number
		case "<=":
			result = left.Number <= right.Number
		case ">":
			result = left.Number > right.Number
		case ">=":
			result = left.Number >= right.Number
		default:
			return Value{}, fmt.Errorf("operator %q is unsupported for number comparison", operator)
		}
	case ir.TypeString:
		switch operator {
		case "==", "===":
			result = left.String == right.String
		case "!=", "!==":
			result = left.String != right.String
		case "<":
			result = left.String < right.String
		case "<=":
			result = left.String <= right.String
		case ">":
			result = left.String > right.String
		case ">=":
			result = left.String >= right.String
		default:
			return Value{}, fmt.Errorf("operator %q is unsupported for string comparison", operator)
		}
	case ir.TypeBool:
		switch operator {
		case "==", "===":
			result = left.Bool == right.Bool
		case "!=", "!==":
			result = left.Bool != right.Bool
		default:
			return Value{}, fmt.Errorf("operator %q is unsupported for bool comparison", operator)
		}
	case ir.TypeBigInt:
		switch operator {
		case "==", "===":
			result = left.BigInt == right.BigInt
		case "!=", "!==":
			result = left.BigInt != right.BigInt
		case "<":
			result = left.BigInt < right.BigInt
		case "<=":
			result = left.BigInt <= right.BigInt
		case ">":
			result = left.BigInt > right.BigInt
		case ">=":
			result = left.BigInt >= right.BigInt
		default:
			return Value{}, fmt.Errorf("operator %q is unsupported for bigint comparison", operator)
		}
	case ir.TypeSymbol:
		switch operator {
		case "==", "===":
			result = left.SymbolID == right.SymbolID
		case "!=", "!==":
			result = left.SymbolID != right.SymbolID
		default:
			return Value{}, fmt.Errorf("operator %q is unsupported for symbol comparison", operator)
		}
	default:
		if strings.HasPrefix(string(left.Type), "object:") || left.Type == "ptr" {
			leftNull := len(left.Object) == 0 && left.Boxed == nil && len(left.GetArray()) == 0
			rightNull := len(right.Object) == 0 && right.Boxed == nil && len(right.GetArray()) == 0
			switch operator {
			case "==", "===":
				result = leftNull == rightNull
			case "!=", "!==":
				result = leftNull != rightNull
			default:
				return Value{}, fmt.Errorf("operator %q is unsupported for object comparison", operator)
			}
			return Value{Type: ir.TypeBool, Bool: result}, nil
		}
		return Value{}, fmt.Errorf("compare is unsupported for %s", left.Type)
	}
	return Value{Type: ir.TypeBool, Bool: result}, nil
}

func format(value Value) string {
	if value.TypedArray != nil {
		ta := value.TypedArray
		parts := make([]string, ta.Length)
		for i := 0; i < ta.Length; i++ {
			parts[i] = strconv.FormatFloat(ta.Get(i), 'f', -1, 64)
		}
		return fmt.Sprintf("%s(%d) [ %s ]", string(ta.Kind), ta.Length, strings.Join(parts, ", "))
	}
	if value.Buffer != nil {
		return fmt.Sprintf("ArrayBuffer { byteLength: %d }", len(value.Buffer.Data))
	}
	if len(value.Array) > 0 || strings.HasSuffix(string(value.Type), "[]") || value.Type == ir.TypeNumberArray || value.Type == ir.TypeStringArray {
		parts := make([]string, len(value.Array))
		for i, item := range value.Array {
			parts[i] = format(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}
	switch value.Type {
	case ir.TypeNumber:
		return strconv.FormatFloat(value.Number, 'f', -1, 64)
	case ir.TypeBigInt:
		return fmt.Sprintf("%dn", value.BigInt)
	case ir.TypeSymbol:
		if value.SymbolDesc != "" {
			return fmt.Sprintf("Symbol(%s)", value.SymbolDesc)
		}
		return "Symbol()"
	case ir.TypeString:
		return value.String
	case ir.TypeBool:
		return strconv.FormatBool(value.Bool)
	case ir.TypeUnknown:
		if value.Boxed != nil {
			return format(*value.Boxed)
		}
		return "undefined"
	default:
		return ""
	}
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
