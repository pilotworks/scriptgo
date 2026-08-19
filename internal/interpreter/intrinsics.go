package interpreter

import (
	"fmt"
	"math"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func executeMathIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	if len(arguments) != 1 {
		return Value{}, fmt.Errorf("%s requires one number", name)
	}
	argument, ok := env[arguments[0]]
	if !ok || argument.Type != ir.TypeNumber {
		return Value{}, fmt.Errorf("%s requires one number", name)
	}
	value := argument.Number
	switch name {
	case "__Math.abs":
		value = math.Abs(value)
	case "__Math.ceil":
		value = math.Ceil(value)
	case "__Math.floor":
		value = math.Floor(value)
	case "__Math.trunc":
		value = math.Trunc(value)
	default:
		return Value{}, fmt.Errorf("unknown math intrinsic %q", name)
	}
	return Value{Type: ir.TypeNumber, Number: value}, nil
}

func executeStringIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	values := make([]Value, 0, len(arguments))
	for _, argument := range arguments {
		value, ok := env[argument]
		if !ok {
			return Value{}, fmt.Errorf("unknown string intrinsic argument %q", argument)
		}
		values = append(values, value)
	}
	switch name {
	case "__string.length":
		if len(values) != 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.length requires one string")
		}
		return Value{Type: ir.TypeNumber, Number: float64(len(values[0].String))}, nil
	case "__string.lastIndexOf":
		if (len(values) != 2 && len(values) != 3) || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.lastIndexOf requires two strings and an optional position")
		}
		value := values[0].String
		if len(values) == 3 {
			if values[2].Number >= 0 {
				position := minInt(int(values[2].Number)+1, len(value))
				value = value[:position]
			}
		}
		return Value{Type: ir.TypeNumber, Number: float64(strings.LastIndex(value, values[1].String))}, nil
	case "__string.slice":
		if len(values) != 3 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeNumber || values[2].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("string.slice requires string and two numbers")
		}
		start := maxInt(0, minInt(int(values[1].Number), len(values[0].String)))
		end := maxInt(start, minInt(int(values[2].Number), len(values[0].String)))
		return Value{Type: ir.TypeString, String: values[0].String[start:end]}, nil
	default:
		return Value{}, fmt.Errorf("unknown string intrinsic %q", name)
	}
}

func executeArrayIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	if len(arguments) != 1 {
		return Value{}, fmt.Errorf("array intrinsic requires one array")
	}
	array, ok := env[arguments[0]]
	if !ok || (array.Type != ir.TypeNumberArray && array.Type != ir.TypeStringArray) {
		return Value{}, fmt.Errorf("array intrinsic requires an array")
	}
	switch name {
	case "__array.length":
		return Value{Type: ir.TypeNumber, Number: float64(len(array.Array))}, nil
	default:
		return Value{}, fmt.Errorf("unknown array intrinsic %q", name)
	}
}
