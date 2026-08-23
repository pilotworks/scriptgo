package interpreter

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func executeNumberIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	values := make([]Value, 0, len(arguments))
	for _, arg := range arguments {
		val, ok := env[arg]
		if !ok {
			return Value{}, fmt.Errorf("unknown number intrinsic argument %q", arg)
		}
		values = append(values, val)
	}
	switch name {
	case "__number.parseInt":
		if len(values) < 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("parseInt requires a string")
		}
		s := strings.TrimSpace(values[0].String)
		sign := 1
		if strings.HasPrefix(s, "-") {
			sign = -1
			s = s[1:]
		} else if strings.HasPrefix(s, "+") {
			s = s[1:]
		}
		var n int64
		parsed := false
		for _, ch := range s {
			if ch >= '0' && ch <= '9' {
				n = n*10 + int64(ch-'0')
				parsed = true
			} else {
				break
			}
		}
		if !parsed {
			return Value{Type: ir.TypeNumber, Number: math.NaN()}, nil
		}
		return Value{Type: ir.TypeNumber, Number: float64(int64(sign) * n)}, nil
	case "__number.parseFloat":
		if len(values) < 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("parseFloat requires a string")
		}
		f, err := strconv.ParseFloat(strings.TrimSpace(values[0].String), 64)
		if err != nil {
			return Value{Type: ir.TypeNumber, Number: math.NaN()}, nil
		}
		return Value{Type: ir.TypeNumber, Number: f}, nil
	case "__number.isNaN":
		if len(values) < 1 || values[0].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("isNaN requires a number")
		}
		return Value{Type: ir.TypeBool, Bool: math.IsNaN(values[0].Number)}, nil
	case "__number.isFinite":
		if len(values) < 1 || values[0].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("isFinite requires a number")
		}
		return Value{Type: ir.TypeBool, Bool: !math.IsNaN(values[0].Number) && !math.IsInf(values[0].Number, 0)}, nil
	case "__number.isInteger":
		if len(values) < 1 || values[0].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("isInteger requires a number")
		}
		return Value{Type: ir.TypeBool, Bool: !math.IsNaN(values[0].Number) && !math.IsInf(values[0].Number, 0) && math.Trunc(values[0].Number) == values[0].Number}, nil
	case "__number.isSafeInteger":
		if len(values) < 1 || values[0].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("isSafeInteger requires a number")
		}
		v := values[0].Number
		isSafe := !math.IsNaN(v) && !math.IsInf(v, 0) && math.Trunc(v) == v && math.Abs(v) <= 9007199254740991
		return Value{Type: ir.TypeBool, Bool: isSafe}, nil
	case "__number.toFixed":
		if len(values) < 1 || values[0].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("toFixed requires a number")
		}
		digits := 0
		if len(values) >= 2 && values[1].Type == ir.TypeNumber {
			digits = int(values[1].Number)
		}
		return Value{Type: ir.TypeString, String: fmt.Sprintf("%.*f", digits, values[0].Number)}, nil
	case "__number.toString":
		if len(values) < 1 || values[0].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("toString requires a number")
		}
		radix := 10
		if len(values) >= 2 && values[1].Type == ir.TypeNumber {
			r := int(values[1].Number)
			if r >= 2 && r <= 36 {
				radix = r
			}
		}
		v := values[0].Number
		if radix == 16 {
			return Value{Type: ir.TypeString, String: fmt.Sprintf("%x", int64(v))}, nil
		}
		if radix == 8 {
			return Value{Type: ir.TypeString, String: fmt.Sprintf("%o", int64(v))}, nil
		}
		if radix == 2 {
			return Value{Type: ir.TypeString, String: fmt.Sprintf("%b", int64(v))}, nil
		}
		if v == math.Trunc(v) {
			return Value{Type: ir.TypeString, String: fmt.Sprintf("%d", int64(v))}, nil
		}
		return Value{Type: ir.TypeString, String: fmt.Sprintf("%g", v)}, nil
	case "__number.toExponential":
		if len(values) < 1 || values[0].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("toExponential requires a number")
		}
		digits := -1
		if len(values) >= 2 && values[1].Type == ir.TypeNumber {
			digits = int(values[1].Number)
		}
		if digits >= 0 {
			return Value{Type: ir.TypeString, String: strconv.FormatFloat(values[0].Number, 'e', digits, 64)}, nil
		}
		return Value{Type: ir.TypeString, String: strconv.FormatFloat(values[0].Number, 'e', -1, 64)}, nil
	case "__number.toPrecision":
		if len(values) < 1 || values[0].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("toPrecision requires a number")
		}
		prec := -1
		if len(values) >= 2 && values[1].Type == ir.TypeNumber {
			prec = int(values[1].Number)
		}
		if prec > 0 {
			return Value{Type: ir.TypeString, String: strconv.FormatFloat(values[0].Number, 'g', prec, 64)}, nil
		}
		return Value{Type: ir.TypeString, String: format(values[0])}, nil
	case "__number.toLocaleString":
		if len(values) < 1 || values[0].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("toLocaleString requires a number")
		}
		return Value{Type: ir.TypeString, String: format(values[0])}, nil
	case "__number.new":
		if len(values) == 0 {
			return Value{Type: ir.TypeNumber, Number: 0}, nil
		}
		if values[0].Type == ir.TypeNumber {
			return values[0], nil
		}
		if values[0].Type == ir.TypeBool {
			if values[0].Bool {
				return Value{Type: ir.TypeNumber, Number: 1}, nil
			}
			return Value{Type: ir.TypeNumber, Number: 0}, nil
		}
		if num, err := strconv.ParseFloat(strings.TrimSpace(values[0].String), 64); err == nil {
			return Value{Type: ir.TypeNumber, Number: num}, nil
		}
		return Value{Type: ir.TypeNumber, Number: 0}, nil
	default:
		return Value{}, fmt.Errorf("unknown number intrinsic %q", name)
	}
}
