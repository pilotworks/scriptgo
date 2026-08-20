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

type Value struct {
	Type    ir.Type
	Number  float64
	BigInt  int64
	String  string
	Bool    bool
	Array   []Value
	Object  map[string]Value
	Closure *Closure
	Boxed   *Value
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
	case ir.TypeString:
		return Value{Type: typ, String: value}, nil
	case ir.TypeBool:
		boolean, err := strconv.ParseBool(value)
		return Value{Type: typ, Bool: boolean}, err
	default:
		return Value{}, fmt.Errorf("unsupported constant type %s", typ)
	}
}

func binary(operator string, left, right Value) (Value, error) {
	if left.Type != right.Type {
		return Value{}, fmt.Errorf("binary operands have types %s and %s", left.Type, right.Type)
	}
	if left.Type == ir.TypeString && operator == "+" {
		return Value{Type: ir.TypeString, String: left.String + right.String}, nil
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
	default:
		return Value{}, fmt.Errorf("operator %q is unsupported for number", operator)
	}
}

func compare(operator string, left, right Value) (Value, error) {
	if left.Type != right.Type {
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
	default:
		return Value{}, fmt.Errorf("compare is unsupported for %s", left.Type)
	}
	return Value{Type: ir.TypeBool, Bool: result}, nil
}

func format(value Value) string {
	switch value.Type {
	case ir.TypeNumber:
		return strconv.FormatFloat(value.Number, 'f', -1, 64)
	case ir.TypeBigInt:
		return fmt.Sprintf("%dn", value.BigInt)
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
