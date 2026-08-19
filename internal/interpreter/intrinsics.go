package interpreter

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

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
	case "__Math.tan":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Tan(values[0])}, nil
	case "__Math.atan":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Atan(values[0])}, nil
	case "__Math.atan2":
		if len(values) != 2 {
			return Value{}, fmt.Errorf("%s requires 2 arguments", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Atan2(values[0], values[1])}, nil
	case "__Math.log":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Log(values[0])}, nil
	case "__Math.log2":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Log2(values[0])}, nil
	case "__Math.log10":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Log10(values[0])}, nil
	case "__Math.exp":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Exp(values[0])}, nil
	case "__Math.sign":
		if len(values) != 1 {
			return Value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		if math.IsNaN(values[0]) {
			return Value{Type: ir.TypeNumber, Number: math.NaN()}, nil
		}
		if values[0] > 0 {
			return Value{Type: ir.TypeNumber, Number: 1}, nil
		} else if values[0] < 0 {
			return Value{Type: ir.TypeNumber, Number: -1}, nil
		}
		return Value{Type: ir.TypeNumber, Number: values[0]}, nil
	case "__Math.hypot":
		if len(values) != 2 {
			return Value{}, fmt.Errorf("%s requires 2 arguments", name)
		}
		return Value{Type: ir.TypeNumber, Number: math.Hypot(values[0], values[1])}, nil
	case "__Math.random":
		return Value{Type: ir.TypeNumber, Number: 0.5}, nil
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
	case "__string.trimStart":
		if len(values) != 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.trimStart requires one string")
		}
		return Value{Type: ir.TypeString, String: strings.TrimLeft(values[0].String, " \t\n\r")}, nil
	case "__string.trimEnd":
		if len(values) != 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.trimEnd requires one string")
		}
		return Value{Type: ir.TypeString, String: strings.TrimRight(values[0].String, " \t\n\r")}, nil
	case "__string.charAt":
		if len(values) < 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.charAt requires a string")
		}
		pos := 0
		if len(values) >= 2 && values[1].Type == ir.TypeNumber {
			pos = int(values[1].Number)
		}
		if pos < 0 || pos >= len(values[0].String) {
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		return Value{Type: ir.TypeString, String: string(values[0].String[pos])}, nil
	case "__string.charCodeAt":
		if len(values) < 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.charCodeAt requires a string")
		}
		pos := 0
		if len(values) >= 2 && values[1].Type == ir.TypeNumber {
			pos = int(values[1].Number)
		}
		if pos < 0 || pos >= len(values[0].String) {
			return Value{Type: ir.TypeNumber, Number: math.NaN()}, nil
		}
		return Value{Type: ir.TypeNumber, Number: float64(values[0].String[pos])}, nil
	case "__string.includes":
		if (len(values) != 2 && len(values) != 3) || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.includes requires two strings and optional start position")
		}
		start := 0
		if len(values) == 3 && values[2].Type == ir.TypeNumber && values[2].Number > 0 {
			start = minInt(int(values[2].Number), len(values[0].String))
		}
		return Value{Type: ir.TypeBool, Bool: strings.Contains(values[0].String[start:], values[1].String)}, nil
	case "__string.toLowerCase":
		if len(values) != 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.toLowerCase requires one string")
		}
		return Value{Type: ir.TypeString, String: strings.ToLower(values[0].String)}, nil
	case "__string.toUpperCase":
		if len(values) != 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.toUpperCase requires one string")
		}
		return Value{Type: ir.TypeString, String: strings.ToUpper(values[0].String)}, nil
	case "__string.repeat":
		if len(values) != 2 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("string.repeat requires a string and count")
		}
		return Value{Type: ir.TypeString, String: strings.Repeat(values[0].String, maxInt(0, int(values[1].Number)))}, nil
	case "__string.replace":
		if len(values) != 3 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString || values[2].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.replace requires three strings")
		}
		return Value{Type: ir.TypeString, String: strings.Replace(values[0].String, values[1].String, values[2].String, 1)}, nil
	case "__string.replaceAll":
		if len(values) != 3 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString || values[2].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.replaceAll requires three strings")
		}
		return Value{Type: ir.TypeString, String: strings.ReplaceAll(values[0].String, values[1].String, values[2].String)}, nil
	case "__string.padStart":
		if len(values) < 2 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("string.padStart requires string and targetLength")
		}
		targetLen := int(values[1].Number)
		padStr := " "
		if len(values) >= 3 && values[2].Type == ir.TypeString && values[2].String != "" {
			padStr = values[2].String
		}
		s := values[0].String
		if len(s) >= targetLen {
			return Value{Type: ir.TypeString, String: s}, nil
		}
		diff := targetLen - len(s)
		repeats := (diff / len(padStr)) + 1
		pad := strings.Repeat(padStr, repeats)[:diff]
		return Value{Type: ir.TypeString, String: pad + s}, nil
	case "__string.padEnd":
		if len(values) < 2 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("string.padEnd requires string and targetLength")
		}
		targetLen := int(values[1].Number)
		padStr := " "
		if len(values) >= 3 && values[2].Type == ir.TypeString && values[2].String != "" {
			padStr = values[2].String
		}
		s := values[0].String
		if len(s) >= targetLen {
			return Value{Type: ir.TypeString, String: s}, nil
		}
		diff := targetLen - len(s)
		repeats := (diff / len(padStr)) + 1
		pad := strings.Repeat(padStr, repeats)[:diff]
		return Value{Type: ir.TypeString, String: s + pad}, nil
	case "__string.concat":
		var sb strings.Builder
		for _, v := range values {
			if v.Type == ir.TypeString {
				sb.WriteString(v.String)
			}
		}
		return Value{Type: ir.TypeString, String: sb.String()}, nil
	case "__string.split":
		if len(values) != 2 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.split requires two strings")
		}
		var parts []string
		if values[1].String == "" {
			for _, r := range values[0].String {
				parts = append(parts, string(r))
			}
		} else {
			parts = strings.Split(values[0].String, values[1].String)
		}
		elems := make([]Value, len(parts))
		for i, p := range parts {
			elems[i] = Value{Type: ir.TypeString, String: p}
		}
		return Value{Type: ir.TypeStringArray, Array: elems}, nil
	default:
		return Value{}, fmt.Errorf("unknown string intrinsic %q", name)
	}
}

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
	case "__number.toFixed":
		if len(values) < 1 || values[0].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("toFixed requires a number")
		}
		digits := 0
		if len(values) >= 2 && values[1].Type == ir.TypeNumber {
			digits = int(values[1].Number)
		}
		return Value{Type: ir.TypeString, String: fmt.Sprintf("%.*f", digits, values[0].Number)}, nil
	default:
		return Value{}, fmt.Errorf("unknown number intrinsic %q", name)
	}
}

func executeArrayIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	if len(arguments) == 0 {
		return Value{}, fmt.Errorf("array intrinsic requires at least one argument")
	}
	array, ok := env[arguments[0]]
	if !ok || (array.Type != ir.TypeNumberArray && array.Type != ir.TypeStringArray) {
		return Value{}, fmt.Errorf("array intrinsic requires an array")
	}
	switch name {
	case "__array.length":
		return Value{Type: ir.TypeNumber, Number: float64(len(array.Array))}, nil
	case "__array.push":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("array.push requires array and element")
		}
		elem, ok := env[arguments[1]]
		if !ok {
			return Value{}, fmt.Errorf("unknown push argument %q", arguments[1])
		}
		array.Array = append(array.Array, elem)
		env[arguments[0]] = array
		return Value{Type: ir.TypeNumber, Number: float64(len(array.Array))}, nil
	case "__array.pop":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("array.pop requires 1 argument")
		}
		if len(array.Array) == 0 {
			if array.Type == ir.TypeNumberArray {
				return Value{Type: ir.TypeNumber, Number: 0}, nil
			}
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		last := array.Array[len(array.Array)-1]
		array.Array = array.Array[:len(array.Array)-1]
		env[arguments[0]] = array
		return last, nil
	case "__array.slice":
		if len(arguments) < 2 || len(arguments) > 3 {
			return Value{}, fmt.Errorf("array.slice requires start and optional end")
		}
		startVal, ok := env[arguments[1]]
		if !ok || startVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("array.slice start must be a number")
		}
		start := int(startVal.Number)
		if start < 0 {
			start = len(array.Array) + start
			if start < 0 {
				start = 0
			}
		} else if start > len(array.Array) {
			start = len(array.Array)
		}
		end := len(array.Array)
		if len(arguments) == 3 {
			endVal, ok := env[arguments[2]]
			if !ok || endVal.Type != ir.TypeNumber {
				return Value{}, fmt.Errorf("array.slice end must be a number")
			}
			end = int(endVal.Number)
			if end < 0 {
				end = len(array.Array) + end
				if end < 0 {
					end = 0
				}
			} else if end > len(array.Array) {
				end = len(array.Array)
			}
		}
		if end < start {
			end = start
		}
		sub := append([]Value(nil), array.Array[start:end]...)
		return Value{Type: array.Type, Array: sub}, nil
	case "__array.indexOf":
		if len(arguments) < 2 || len(arguments) > 3 {
			return Value{}, fmt.Errorf("array.indexOf requires target and optional fromIndex")
		}
		target, ok := env[arguments[1]]
		if !ok {
			return Value{}, fmt.Errorf("unknown indexOf target")
		}
		start := 0
		if len(arguments) == 3 {
			fromVal, ok := env[arguments[2]]
			if ok && fromVal.Type == ir.TypeNumber && fromVal.Number > 0 {
				start = int(fromVal.Number)
			}
		}
		idx := -1
		for i := start; i < len(array.Array); i++ {
			if array.Type == ir.TypeNumberArray && array.Array[i].Number == target.Number {
				idx = i
				break
			} else if array.Type == ir.TypeStringArray && array.Array[i].String == target.String {
				idx = i
				break
			}
		}
		return Value{Type: ir.TypeNumber, Number: float64(idx)}, nil
	case "__array.includes":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("array.includes requires target")
		}
		target, ok := env[arguments[1]]
		if !ok {
			return Value{}, fmt.Errorf("unknown includes target")
		}
		found := false
		for i := 0; i < len(array.Array); i++ {
			if array.Type == ir.TypeNumberArray && array.Array[i].Number == target.Number {
				found = true
				break
			} else if array.Type == ir.TypeStringArray && array.Array[i].String == target.String {
				found = true
				break
			}
		}
		return Value{Type: ir.TypeBool, Bool: found}, nil
	case "__array.at":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("array.at requires index")
		}
		idxVal, ok := env[arguments[1]]
		if !ok || idxVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("array.at index must be a number")
		}
		idx := int(idxVal.Number)
		if idx < 0 {
			idx = len(array.Array) + idx
		}
		if idx < 0 || idx >= len(array.Array) {
			if array.Type == ir.TypeNumberArray {
				return Value{Type: ir.TypeNumber, Number: 0}, nil
			}
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		return array.Array[idx], nil
	case "__array.shift":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("array.shift requires 1 argument")
		}
		if len(array.Array) == 0 {
			if array.Type == ir.TypeNumberArray {
				return Value{Type: ir.TypeNumber, Number: 0}, nil
			}
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		first := array.Array[0]
		array.Array = array.Array[1:]
		env[arguments[0]] = array
		return first, nil
	case "__array.unshift":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("array.unshift requires array and element")
		}
		elem, ok := env[arguments[1]]
		if !ok {
			return Value{}, fmt.Errorf("unknown unshift argument %q", arguments[1])
		}
		array.Array = append([]Value{elem}, array.Array...)
		env[arguments[0]] = array
		return Value{Type: ir.TypeNumber, Number: float64(len(array.Array))}, nil
	case "__array.reverse":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("array.reverse requires 1 argument")
		}
		for i, j := 0, len(array.Array)-1; i < j; i, j = i+1, j-1 {
			array.Array[i], array.Array[j] = array.Array[j], array.Array[i]
		}
		env[arguments[0]] = array
		return array, nil
	case "__array.concat":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("array.concat requires other array")
		}
		other, ok := env[arguments[1]]
		if !ok || other.Type != array.Type {
			return Value{}, fmt.Errorf("array.concat requires matching array type")
		}
		newItems := make([]Value, 0, len(array.Array)+len(other.Array))
		newItems = append(newItems, array.Array...)
		newItems = append(newItems, other.Array...)
		return Value{Type: array.Type, Array: newItems}, nil
	case "__array.splice":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.splice requires start")
		}
		startVal, ok := env[arguments[1]]
		if !ok || startVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("array.splice start must be number")
		}
		start := int(startVal.Number)
		if start < 0 {
			start = len(array.Array) + start
			if start < 0 {
				start = 0
			}
		} else if start > len(array.Array) {
			start = len(array.Array)
		}
		deleteCount := len(array.Array) - start
		if len(arguments) >= 3 {
			dcVal, ok := env[arguments[2]]
			if ok && dcVal.Type == ir.TypeNumber {
				if dcVal.Number < 0 {
					deleteCount = 0
				} else {
					deleteCount = int(dcVal.Number)
					if start+deleteCount > len(array.Array) {
						deleteCount = len(array.Array) - start
					}
				}
			}
		}
		deleted := append([]Value(nil), array.Array[start:start+deleteCount]...)
		array.Array = append(array.Array[:start], array.Array[start+deleteCount:]...)
		env[arguments[0]] = array
		return Value{Type: array.Type, Array: deleted}, nil
	case "__array.join":
		sep := ","
		if len(arguments) >= 2 {
			sepVal, ok := env[arguments[1]]
			if ok && sepVal.Type == ir.TypeString {
				sep = sepVal.String
			}
		}
		parts := make([]string, len(array.Array))
		for i, item := range array.Array {
			if item.Type == ir.TypeNumber {
				parts[i] = format(item)
			} else {
				parts[i] = item.String
			}
		}
		return Value{Type: ir.TypeString, String: strings.Join(parts, sep)}, nil
	default:
		return Value{}, fmt.Errorf("unknown array intrinsic %q", name)
	}
}

func executeFsIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__fs.readFileSync":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("fs.readFileSync requires 1 argument")
		}
		pathVal, ok := env[arguments[0]]
		if !ok || pathVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("fs.readFileSync requires a string path")
		}
		content, err := os.ReadFile(pathVal.String)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeString, String: string(content)}, nil
	case "__fs.writeFileSync":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("fs.writeFileSync requires 2 arguments")
		}
		pathVal, ok := env[arguments[0]]
		if !ok || pathVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("fs.writeFileSync requires a string path")
		}
		contentVal, ok := env[arguments[1]]
		if !ok || contentVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("fs.writeFileSync requires a string content")
		}
		err := os.WriteFile(pathVal.String, []byte(contentVal.String), 0644)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.existsSync":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("fs.existsSync requires 1 argument")
		}
		pathVal, ok := env[arguments[0]]
		if !ok || pathVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("fs.existsSync requires a string path")
		}
		_, err := os.Stat(pathVal.String)
		return Value{Type: ir.TypeBool, Bool: err == nil}, nil
	default:
		return Value{}, fmt.Errorf("unknown fs intrinsic %q", name)
	}
}

func executeProcessIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__process.exit":
		return Value{Type: ir.TypeVoid}, nil
	case "__process.cwd":
		cwd, err := os.Getwd()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeString, String: cwd}, nil
	case "__process.argv":
		args := os.Args
		arr := make([]Value, len(args))
		for i, a := range args {
			arr[i] = Value{Type: ir.TypeString, String: a}
		}
		return Value{Type: ir.TypeStringArray, Array: arr}, nil
	case "__process.env":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("process.env requires 1 argument")
		}
		keyVal, ok := env[arguments[0]]
		if !ok || keyVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("process.env requires a string key")
		}
		return Value{Type: ir.TypeString, String: os.Getenv(keyVal.String)}, nil
	default:
		return Value{}, fmt.Errorf("unknown process intrinsic %q", name)
	}
}

func executeCryptoIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__crypto.randomUUID":
		var b [16]byte
		if _, err := rand.Read(b[:]); err != nil {
			return Value{}, err
		}
		b[6] = (b[6] & 0x0f) | 0x40
		b[8] = (b[8] & 0x3f) | 0x80
		uuid := fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
			b[0], b[1], b[2], b[3],
			b[4], b[5],
			b[6], b[7],
			b[8], b[9],
			b[10], b[11], b[12], b[13], b[14], b[15],
		)
		return Value{Type: ir.TypeString, String: uuid}, nil
	default:
		return Value{}, fmt.Errorf("unknown crypto intrinsic %q", name)
	}
}

func executeWebIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__web.btoa":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("btoa requires 1 argument")
		}
		val, ok := env[arguments[0]]
		if !ok || val.Type != ir.TypeString {
			return Value{}, fmt.Errorf("btoa requires string argument")
		}
		encoded := base64.StdEncoding.EncodeToString([]byte(val.String))
		return Value{Type: ir.TypeString, String: encoded}, nil
	case "__web.atob":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("atob requires 1 argument")
		}
		val, ok := env[arguments[0]]
		if !ok || val.Type != ir.TypeString {
			return Value{}, fmt.Errorf("atob requires string argument")
		}
		decoded, err := base64.StdEncoding.DecodeString(val.String)
		if err != nil {
			return Value{}, fmt.Errorf("InvalidCharacterError: %w", err)
		}
		return Value{Type: ir.TypeString, String: string(decoded)}, nil
	default:
		return Value{}, fmt.Errorf("unknown web intrinsic %q", name)
	}
}

func executePerformanceIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__performance.now":
		ms := float64(time.Now().UnixNano()) / 1e6
		return Value{Type: ir.TypeNumber, Number: ms}, nil
	default:
		return Value{}, fmt.Errorf("unknown performance intrinsic %q", name)
	}
}


