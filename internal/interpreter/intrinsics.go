package interpreter

import (
	"fmt"
	"math"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func executeMathIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	values := make([]float64, 0, len(arguments))
	for _, arg := range arguments {
		val, ok := env[arg]
		if !ok || val.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("%s requires number operands", name)
		}
		values = append(values, val.Number)
	}
	switch name {
	case "__Math.abs":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Abs(values[0])}, nil
	case "__Math.ceil":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Ceil(values[0])}, nil
	case "__Math.floor":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Floor(values[0])}, nil
	case "__Math.trunc":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Trunc(values[0])}, nil
	case "__Math.sqrt":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Sqrt(values[0])}, nil
	case "__Math.round":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Round(values[0])}, nil
	case "__Math.sin":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Sin(values[0])}, nil
	case "__Math.cos":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Cos(values[0])}, nil
	case "__Math.min":
		if len(values) != 2 {
			return Value{}, fmt.Errorf("%s requires 2 arguments", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Min(values[0], values[1])}, nil
	case "__Math.max":
		if len(values) != 2 {
			return Value{}, fmt.Errorf("%s requires 2 arguments", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Max(values[0], values[1])}, nil
	case "__Math.pow":
		if len(values) != 2 {
			return Value{}, fmt.Errorf("%s requires 2 arguments", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Pow(values[0], values[1])}, nil
	default:
		return Value{}, fmt.Errorf("unknown math intrinsic %q", name)
	}
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
	case "__string.indexOf":
		if (len(values) != 2 && len(values) != 3) || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.indexOf requires two strings and an optional position")
		}
		value := values[0].String
		start := 0
		if len(values) == 3 && values[2].Number > 0 {
			start = minInt(int(values[2].Number), len(value))
		}
		idx := strings.Index(value[start:], values[1].String)
		if idx != -1 {
			idx += start
		}
		return Value{Type: ir.TypeNumber, Number: float64(idx)}, nil
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
	case "__string.startsWith":
		if len(values) != 2 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.startsWith requires two strings")
		}
		return Value{Type: ir.TypeBool, Bool: strings.HasPrefix(values[0].String, values[1].String)}, nil
	case "__string.endsWith":
		if len(values) != 2 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.endsWith requires two strings")
		}
		return Value{Type: ir.TypeBool, Bool: strings.HasSuffix(values[0].String, values[1].String)}, nil
	case "__string.fromNumber":
		if len(values) != 1 || values[0].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("string.fromNumber requires one number")
		}
		return Value{Type: ir.TypeString, String: format(values[0])}, nil
	case "__string.fromBool":
		if len(values) != 1 || values[0].Type != ir.TypeBool {
			return Value{}, fmt.Errorf("string.fromBool requires one bool")
		}
		return Value{Type: ir.TypeString, String: format(values[0])}, nil
	case "__string.slice", "__string.substring":
		if (len(values) != 2 && len(values) != 3) || values[0].Type != ir.TypeString || values[1].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("string.slice requires string and number arguments")
		}
		start := maxInt(0, minInt(int(values[1].Number), len(values[0].String)))
		end := len(values[0].String)
		if len(values) == 3 && values[2].Type == ir.TypeNumber && values[2].Number >= 0 {
			end = maxInt(start, minInt(int(values[2].Number), len(values[0].String)))
		}
		return Value{Type: ir.TypeString, String: values[0].String[start:end]}, nil
	case "__string.trim":
		if len(values) != 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.trim requires one string")
		}
		return Value{Type: ir.TypeString, String: strings.TrimSpace(values[0].String)}, nil
	case "__string.replace":
		if len(values) != 3 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString || values[2].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.replace requires three strings")
		}
		return Value{Type: ir.TypeString, String: strings.Replace(values[0].String, values[1].String, values[2].String, 1)}, nil
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
