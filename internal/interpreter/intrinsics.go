package interpreter

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	goruntime "runtime"
	"sort"
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
	if val, handled, err := executeRegexStringIntrinsic(name, values); handled {
		return val, err
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
	default:
		return Value{}, fmt.Errorf("unknown number intrinsic %q", name)
	}
}

func executeArrayIntrinsic(name string, arguments []string, env map[string]Value, functions map[string]ir.Function, output *bytes.Buffer) (Value, error) {
	if name == "__array.isArray" {
		if len(arguments) == 0 {
			return Value{Type: ir.TypeBool, Bool: false}, nil
		}
		arg, ok := env[arguments[0]]
		if !ok {
			return Value{Type: ir.TypeBool, Bool: false}, nil
		}
		isArr := strings.HasSuffix(string(arg.Type), "[]") || arg.Type == ir.TypeNumberArray || arg.Type == ir.TypeStringArray || arg.Type == ir.TypeBoolArray || arg.Type == ir.TypeBigIntArray
		return Value{Type: ir.TypeBool, Bool: isArr}, nil
	}
	if name == "__array.of" {
		var arr []Value
		elemType := ir.TypeNumber
		for _, argName := range arguments {
			arg, ok := env[argName]
			if !ok {
				return Value{}, fmt.Errorf("unknown array.of argument %q", argName)
			}
			arr = append(arr, arg)
			if arg.Type != "" {
				elemType = arg.Type
			}
		}
		retType := ir.Type(string(elemType) + "[]")
		if elemType == ir.TypeNumber {
			retType = ir.TypeNumberArray
		} else if elemType == ir.TypeString {
			retType = ir.TypeStringArray
		}
		return Value{Type: retType, Array: arr}, nil
	}
	if name == "__array.from" {
		if len(arguments) == 0 {
			return Value{}, fmt.Errorf("array.from requires at least 1 argument")
		}
		arg, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown array.from argument %q", arguments[0])
		}
		if arg.Type == ir.TypeString {
			runes := []rune(arg.String)
			arr := make([]Value, len(runes))
			for i, r := range runes {
				arr[i] = Value{Type: ir.TypeString, String: string(r)}
			}
			return Value{Type: ir.TypeStringArray, Array: arr}, nil
		}
		newArr := append([]Value(nil), arg.GetArray()...)
		return Value{Type: arg.Type, Array: newArr}, nil
	}
	if len(arguments) == 0 {
		return Value{}, fmt.Errorf("array intrinsic requires at least one argument")
	}
	array, ok := env[arguments[0]]
	if !ok || (!strings.HasSuffix(string(array.Type), "[]") && array.Type != ir.TypeNumberArray && array.Type != ir.TypeStringArray) {
		return Value{}, fmt.Errorf("array intrinsic requires an array")
	}
	array.Array = array.GetArray()
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
		array.SetArray(array.Array)
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
		array.SetArray(array.Array)
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
			start = max(len(array.Array)+start, 0)
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
				end = max(len(array.Array)+end, 0)
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
		array.SetArray(array.Array)
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
		array.SetArray(array.Array)
		env[arguments[0]] = array
		return Value{Type: ir.TypeNumber, Number: float64(len(array.Array))}, nil
	case "__array.reverse":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("array.reverse requires 1 argument")
		}
		for i, j := 0, len(array.Array)-1; i < j; i, j = i+1, j-1 {
			array.Array[i], array.Array[j] = array.Array[j], array.Array[i]
		}
		array.SetArray(array.Array)
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
		return Value{Type: array.Type, Array: newItems, ArrayRef: &newItems}, nil
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
			start = max(len(array.Array)+start, 0)
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
		array.SetArray(array.Array)
		env[arguments[0]] = array
		return Value{Type: array.Type, Array: deleted, ArrayRef: &deleted}, nil
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
				parts[i] = strconv.FormatFloat(item.Number, 'f', -1, 64)
			} else if item.Type == ir.TypeBigInt {
				parts[i] = strconv.FormatInt(item.BigInt, 10)
			} else if item.Type == ir.TypeBool {
				parts[i] = strconv.FormatBool(item.Bool)
			} else {
				parts[i] = item.String
			}
		}
		return Value{Type: ir.TypeString, String: strings.Join(parts, sep)}, nil
	case "__array.map":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.map requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.map callback must be a closure")
		}
		mapped := make([]Value, 0, len(array.Array))
		for i, item := range array.Array {
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return res, fmt.Errorf("uncaught exception in array.map")
			}
			mapped = append(mapped, res)
		}
		retType := array.Type
		if len(mapped) > 0 {
			if mapped[0].Type == ir.TypeString {
				retType = ir.TypeStringArray
			} else {
				retType = ir.TypeNumberArray
			}
		}
		return Value{Type: retType, Array: mapped}, nil
	case "__array.filter":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.filter requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.filter callback must be a closure")
		}
		filtered := make([]Value, 0)
		for i, item := range array.Array {
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return res, fmt.Errorf("uncaught exception in array.filter")
			}
			if res.Bool {
				filtered = append(filtered, item)
			}
		}
		return Value{Type: array.Type, Array: filtered}, nil
	case "__array.forEach":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.forEach requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.forEach callback must be a closure")
		}
		for i, item := range array.Array {
			_, flow, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return Value{}, fmt.Errorf("uncaught exception in array.forEach")
			}
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__array.reduce":
		if len(arguments) < 3 {
			return Value{}, fmt.Errorf("array.reduce requires callback closure and initial value")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.reduce callback must be a closure")
		}
		acc, ok := env[arguments[2]]
		if !ok {
			return Value{}, fmt.Errorf("unknown reduce initial value")
		}
		for i, item := range array.Array {
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{acc, item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return res, fmt.Errorf("uncaught exception in array.reduce")
			}
			acc = res
		}
		return acc, nil
	case "__array.find":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.find requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.find callback must be a closure")
		}
		for i, item := range array.Array {
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return res, fmt.Errorf("uncaught exception in array.find")
			}
			if res.Bool {
				return item, nil
			}
		}
		if array.Type == ir.TypeNumberArray {
			return Value{Type: ir.TypeNumber, Number: 0}, nil
		} else if array.Type == ir.TypeStringArray {
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		elemType := ir.Type(strings.TrimSuffix(string(array.Type), "[]"))
		if !strings.HasPrefix(string(elemType), "object:") && elemType != ir.TypeNumber && elemType != ir.TypeString && elemType != ir.TypeBool && elemType != ir.TypeBigInt {
			elemType = ir.Type("object:" + string(elemType))
		}
		return Value{Type: elemType}, nil
	case "__array.some":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.some requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.some callback must be a closure")
		}
		for i, item := range array.Array {
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if res.Bool {
				return Value{Type: ir.TypeBool, Bool: true}, nil
			}
		}
		return Value{Type: ir.TypeBool, Bool: false}, nil
	case "__array.every":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.every requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.every callback must be a closure")
		}
		for i, item := range array.Array {
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if !res.Bool {
				return Value{Type: ir.TypeBool, Bool: false}, nil
			}
		}
		return Value{Type: ir.TypeBool, Bool: true}, nil
	case "__array.findIndex":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.findIndex requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.findIndex callback must be a closure")
		}
		for i, item := range array.Array {
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if res.Bool {
				return Value{Type: ir.TypeNumber, Number: float64(i)}, nil
			}
		}
		return Value{Type: ir.TypeNumber, Number: -1}, nil
	case "__array.fill":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.fill requires value argument")
		}
		fillVal, ok := env[arguments[1]]
		if !ok {
			return Value{}, fmt.Errorf("unknown fill value %q", arguments[1])
		}
		start := 0
		end := len(array.Array)
		if len(arguments) > 2 {
			if sVal, ok := env[arguments[2]]; ok && sVal.Type == ir.TypeNumber {
				start = int(sVal.Number)
				if start < 0 {
					start = max(len(array.Array)+start, 0)
				} else if start > len(array.Array) {
					start = len(array.Array)
				}
			}
		}
		if len(arguments) > 3 {
			if eVal, ok := env[arguments[3]]; ok && eVal.Type == ir.TypeNumber {
				end = int(eVal.Number)
				if end < 0 {
					end = max(len(array.Array)+end, 0)
				} else if end > len(array.Array) {
					end = len(array.Array)
				}
			}
		}
		for i := start; i < end; i++ {
			array.Array[i] = fillVal
		}
		array.SetArray(array.Array)
		env[arguments[0]] = array
		return array, nil
	case "__array.toReversed":
		reversed := make([]Value, len(array.Array))
		for i := range array.Array {
			reversed[i] = array.Array[len(array.Array)-1-i]
		}
		return Value{Type: array.Type, Array: reversed}, nil
	case "__array.toSorted":
		sorted := make([]Value, len(array.Array))
		copy(sorted, array.Array)
		if len(arguments) > 1 {
			if closureVal, ok := env[arguments[1]]; ok && closureVal.Closure != nil {
				var sortErr error
				sort.SliceStable(sorted, func(i, j int) bool {
					if sortErr != nil {
						return false
					}
					res, _, err := executeClosure(functions, closureVal.Closure, []Value{sorted[i], sorted[j]}, output)
					if err != nil {
						sortErr = err
						return false
					}
					return res.Number < 0
				})
				if sortErr != nil {
					return Value{}, sortErr
				}
				return Value{Type: array.Type, Array: sorted}, nil
			}
		}
		sort.SliceStable(sorted, func(i, j int) bool {
			if sorted[i].Type == ir.TypeNumber && sorted[j].Type == ir.TypeNumber {
				return sorted[i].Number < sorted[j].Number
			}
			if sorted[i].Type == ir.TypeString && sorted[j].Type == ir.TypeString {
				return sorted[i].String < sorted[j].String
			}
			return false
		})
		return Value{Type: array.Type, Array: sorted}, nil
	case "__array.findLast":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.findLast requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.findLast callback must be a closure")
		}
		for i := len(array.Array) - 1; i >= 0; i-- {
			item := array.Array[i]
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return res, fmt.Errorf("uncaught exception in array.findLast")
			}
			if res.Bool {
				return item, nil
			}
		}
		if array.Type == ir.TypeNumberArray {
			return Value{Type: ir.TypeNumber, Number: 0}, nil
		} else if array.Type == ir.TypeStringArray {
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		return Value{Type: ir.Type(strings.TrimSuffix(string(array.Type), "[]"))}, nil
	case "__array.findLastIndex":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.findLastIndex requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.findLastIndex callback must be a closure")
		}
		for i := len(array.Array) - 1; i >= 0; i-- {
			item := array.Array[i]
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if res.Bool {
				return Value{Type: ir.TypeNumber, Number: float64(i)}, nil
			}
		}
		return Value{Type: ir.TypeNumber, Number: -1}, nil
	case "__array.lastIndexOf":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.lastIndexOf requires target")
		}
		target, ok := env[arguments[1]]
		if !ok {
			return Value{}, fmt.Errorf("unknown lastIndexOf target")
		}
		fromIndex := len(array.Array) - 1
		if len(arguments) >= 3 {
			if fVal, ok := env[arguments[2]]; ok && fVal.Type == ir.TypeNumber {
				fromIndex = int(fVal.Number)
				if fromIndex < 0 {
					fromIndex = len(array.Array) + fromIndex
				}
				if fromIndex >= len(array.Array) {
					fromIndex = len(array.Array) - 1
				}
			}
		}
		for i := fromIndex; i >= 0; i-- {
			if array.Type == ir.TypeNumberArray && array.Array[i].Number == target.Number {
				return Value{Type: ir.TypeNumber, Number: float64(i)}, nil
			} else if array.Type == ir.TypeStringArray && array.Array[i].String == target.String {
				return Value{Type: ir.TypeNumber, Number: float64(i)}, nil
			} else if array.Array[i].Type == target.Type && array.Array[i].Number == target.Number && array.Array[i].String == target.String {
				return Value{Type: ir.TypeNumber, Number: float64(i)}, nil
			}
		}
		return Value{Type: ir.TypeNumber, Number: -1}, nil
	case "__array.toSpliced":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.toSpliced requires start")
		}
		startVal, ok := env[arguments[1]]
		if !ok || startVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("array.toSpliced start must be number")
		}
		start := int(startVal.Number)
		if start < 0 {
			start = max(len(array.Array)+start, 0)
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
		var insertItems []Value
		for i := 3; i < len(arguments); i++ {
			if item, ok := env[arguments[i]]; ok {
				insertItems = append(insertItems, item)
			}
		}
		newItems := make([]Value, 0, len(array.Array)-deleteCount+len(insertItems))
		newItems = append(newItems, array.Array[:start]...)
		newItems = append(newItems, insertItems...)
		newItems = append(newItems, array.Array[start+deleteCount:]...)
		return Value{Type: array.Type, Array: newItems}, nil
	case "__array.with":
		if len(arguments) < 3 {
			return Value{}, fmt.Errorf("array.with requires index and value")
		}
		idxVal, ok := env[arguments[1]]
		if !ok || idxVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("array.with index must be number")
		}
		newVal, ok := env[arguments[2]]
		if !ok {
			return Value{}, fmt.Errorf("array.with value missing")
		}
		idx := int(idxVal.Number)
		if idx < 0 {
			idx = len(array.Array) + idx
		}
		if idx < 0 || idx >= len(array.Array) {
			return Value{}, fmt.Errorf("array.with index out of bounds")
		}
		newArr := make([]Value, len(array.Array))
		copy(newArr, array.Array)
		newArr[idx] = newVal
		return Value{Type: array.Type, Array: newArr}, nil
	case "__array.sort":
		if len(arguments) > 1 {
			if closureVal, ok := env[arguments[1]]; ok && closureVal.Closure != nil {
				var sortErr error
				sort.SliceStable(array.Array, func(i, j int) bool {
					if sortErr != nil {
						return false
					}
					res, _, err := executeClosure(functions, closureVal.Closure, []Value{array.Array[i], array.Array[j]}, output)
					if err != nil {
						sortErr = err
						return false
					}
					return res.Number < 0
				})
				if sortErr != nil {
					return Value{}, sortErr
				}
				array.SetArray(array.Array)
				env[arguments[0]] = array
				return array, nil
			}
		}
		sort.SliceStable(array.Array, func(i, j int) bool {
			if array.Array[i].Type == ir.TypeNumber && array.Array[j].Type == ir.TypeNumber {
				return array.Array[i].Number < array.Array[j].Number
			}
			if array.Array[i].Type == ir.TypeString && array.Array[j].Type == ir.TypeString {
				return array.Array[i].String < array.Array[j].String
			}
			return false
		})
		array.SetArray(array.Array)
		env[arguments[0]] = array
		return array, nil
	case "__array.copyWithin":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.copyWithin requires target")
		}
		targetVal, ok := env[arguments[1]]
		if !ok || targetVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("array.copyWithin target must be number")
		}
		target := int(targetVal.Number)
		if target < 0 {
			target = max(len(array.Array)+target, 0)
		} else if target > len(array.Array) {
			target = len(array.Array)
		}
		start := 0
		if len(arguments) > 2 {
			if sVal, ok := env[arguments[2]]; ok && sVal.Type == ir.TypeNumber {
				start = int(sVal.Number)
				if start < 0 {
					start = max(len(array.Array)+start, 0)
				} else if start > len(array.Array) {
					start = len(array.Array)
				}
			}
		}
		end := len(array.Array)
		if len(arguments) > 3 {
			if eVal, ok := env[arguments[3]]; ok && eVal.Type == ir.TypeNumber {
				end = int(eVal.Number)
				if end < 0 {
					end = max(len(array.Array)+end, 0)
				} else if end > len(array.Array) {
					end = len(array.Array)
				}
			}
		}
		count := minInt(end-start, len(array.Array)-target)
		if count > 0 && start < len(array.Array) {
			temp := make([]Value, count)
			copy(temp, array.Array[start:start+count])
			copy(array.Array[target:target+count], temp)
		}
		array.SetArray(array.Array)
		env[arguments[0]] = array
		return array, nil
	case "__array.reduceRight":
		if len(arguments) < 3 {
			return Value{}, fmt.Errorf("array.reduceRight requires callback closure and initial value")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.reduceRight callback must be a closure")
		}
		acc, ok := env[arguments[2]]
		if !ok {
			return Value{}, fmt.Errorf("unknown reduceRight initial value")
		}
		for i := len(array.Array) - 1; i >= 0; i-- {
			item := array.Array[i]
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{acc, item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return res, fmt.Errorf("uncaught exception in array.reduceRight")
			}
			acc = res
		}
		return acc, nil
	case "__array.toString", "__array.toLocaleString":
		parts := make([]string, len(array.Array))
		for i, item := range array.Array {
			if item.Type == ir.TypeNumber {
				parts[i] = strconv.FormatFloat(item.Number, 'f', -1, 64)
			} else if item.Type == ir.TypeBigInt {
				parts[i] = strconv.FormatInt(item.BigInt, 10)
			} else if item.Type == ir.TypeBool {
				parts[i] = strconv.FormatBool(item.Bool)
			} else {
				parts[i] = item.String
			}
		}
		return Value{Type: ir.TypeString, String: strings.Join(parts, ",")}, nil
	case "__array.flat":
		var flatItems []Value
		for _, item := range array.Array {
			if len(item.GetArray()) > 0 {
				flatItems = append(flatItems, item.GetArray()...)
			} else {
				flatItems = append(flatItems, item)
			}
		}
		return Value{Type: array.Type, Array: flatItems}, nil
	case "__array.flatMap":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.flatMap requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.flatMap callback must be a closure")
		}
		var flatMapped []Value
		for i, item := range array.Array {
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return res, fmt.Errorf("uncaught exception in array.flatMap")
			}
			if len(res.GetArray()) > 0 {
				flatMapped = append(flatMapped, res.GetArray()...)
			} else {
				flatMapped = append(flatMapped, res)
			}
		}
		return Value{Type: array.Type, Array: flatMapped}, nil
	case "__array.entries":
		var pairs []Value
		for i, item := range array.Array {
			pairs = append(pairs, Value{Type: ir.TypeString, String: fmt.Sprintf("[%d, %s]", i, format(item))})
		}
		return Value{Type: ir.TypeStringArray, Array: pairs}, nil
	case "__array.keys":
		keys := make([]Value, len(array.Array))
		for i := range array.Array {
			keys[i] = Value{Type: ir.TypeNumber, Number: float64(i)}
		}
		return Value{Type: ir.TypeNumberArray, Array: keys}, nil
	case "__array.values":
		vals := append([]Value(nil), array.Array...)
		return Value{Type: array.Type, Array: vals}, nil
	default:
		return Value{}, fmt.Errorf("unknown array intrinsic %q", name)
	}
}

func executeFsIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__fs.readFileSync":
		if len(arguments) < 1 || len(arguments) > 2 {
			return Value{}, fmt.Errorf("fs.readFileSync requires 1 or 2 arguments")
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
	case "__fs.unlinkSync":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("fs.unlinkSync requires 1 argument")
		}
		pathVal, ok := env[arguments[0]]
		if !ok || pathVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("fs.unlinkSync requires a string path")
		}
		_ = os.Remove(pathVal.String)
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.readdirSync":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("fs.readdirSync requires 1 argument")
		}
		pathVal := env[arguments[0]].String
		entries, err := os.ReadDir(pathVal)
		if err != nil {
			return Value{}, err
		}
		var arr []Value
		for _, entry := range entries {
			arr = append(arr, Value{Type: ir.TypeString, String: entry.Name()})
		}
		return Value{Type: ir.TypeStringArray, Array: arr}, nil
	case "__fs.copyFileSync":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("fs.copyFileSync requires 2 arguments")
		}
		src := env[arguments[0]].String
		dest := env[arguments[1]].String
		data, err := os.ReadFile(src)
		if err != nil {
			return Value{}, err
		}
		err = os.WriteFile(dest, data, 0644)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.renameSync":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("fs.renameSync requires 2 arguments")
		}
		oldPath := env[arguments[0]].String
		newPath := env[arguments[1]].String
		err := os.Rename(oldPath, newPath)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.appendFileSync":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("fs.appendFileSync requires 2 arguments")
		}
		pathVal := env[arguments[0]].String
		content := env[arguments[1]].String
		f, err := os.OpenFile(pathVal, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return Value{}, err
		}
		defer f.Close()
		_, err = f.WriteString(content)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.mkdirSync":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("fs.mkdirSync requires 1 argument")
		}
		pathVal := env[arguments[0]].String
		isRec := false
		if len(arguments) > 1 && (env[arguments[1]].Bool || env[arguments[1]].Number > 0) {
			isRec = true
		}
		var err error
		if isRec {
			err = os.MkdirAll(pathVal, 0755)
		} else {
			err = os.Mkdir(pathVal, 0755)
		}
		if err != nil && !os.IsExist(err) {
			return Value{}, err
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.rmSync":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("fs.rmSync requires 1 argument")
		}
		pathVal := env[arguments[0]].String
		isRec := false
		if len(arguments) > 1 && (env[arguments[1]].Bool || env[arguments[1]].Number > 0) {
			isRec = true
		}
		var err error
		if isRec {
			err = os.RemoveAll(pathVal)
		} else {
			err = os.Remove(pathVal)
		}
		if err != nil {
			isForce := false
			if len(arguments) > 2 && (env[arguments[2]].Bool || env[arguments[2]].Number > 0) {
				isForce = true
			}
			if !isForce && !os.IsNotExist(err) {
				return Value{}, err
			}
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__fs.statSync":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("fs.statSync requires 1 argument")
		}
		pathVal := env[arguments[0]].String
		info, err := os.Stat(pathVal)
		if err != nil {
			return Value{}, err
		}
		size := float64(info.Size())
		mtimeMs := float64(info.ModTime().UnixMilli())
		birthtimeMs := mtimeMs
		var mode float64
		if info.IsDir() {
			mode = float64(0040755)
		} else {
			mode = float64(0100644)
		}
		return Value{
			Type: "object:Stats",
			Object: map[string]Value{
				"size":        {Type: ir.TypeNumber, Number: size},
				"mtimeMs":     {Type: ir.TypeNumber, Number: mtimeMs},
				"birthtimeMs": {Type: ir.TypeNumber, Number: birthtimeMs},
				"mode":        {Type: ir.TypeNumber, Number: mode},
			},
		}, nil
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
	case "__crypto.hashDigest":
		if len(arguments) < 2 || len(arguments) > 3 {
			return Value{}, fmt.Errorf("crypto.hashDigest requires 2 or 3 arguments")
		}
		algoVal, ok := env[arguments[0]]
		if !ok || algoVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("crypto.hashDigest requires a string algorithm")
		}
		dataVal, ok := env[arguments[1]]
		if !ok || dataVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("crypto.hashDigest requires a string data")
		}
		encoding := "hex"
		if len(arguments) == 3 {
			encVal, ok := env[arguments[2]]
			if ok && encVal.Type == ir.TypeString && encVal.String != "" {
				encoding = encVal.String
			}
		}
		var hashBytes []byte
		switch strings.ToLower(algoVal.String) {
		case "sha256":
			h := sha256.Sum256([]byte(dataVal.String))
			hashBytes = h[:]
		default:
			h := sha256.Sum256([]byte(dataVal.String))
			hashBytes = h[:]
		}
		var outStr string
		switch strings.ToLower(encoding) {
		case "hex":
			outStr = hex.EncodeToString(hashBytes)
		case "base64":
			outStr = base64.StdEncoding.EncodeToString(hashBytes)
		default:
			outStr = hex.EncodeToString(hashBytes)
		}
		return Value{Type: ir.TypeString, String: outStr}, nil
	default:
		return Value{}, fmt.Errorf("unknown crypto intrinsic %q", name)
	}
}

func executeDateIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__date.now":
		return Value{Type: ir.TypeNumber, Number: float64(time.Now().UnixMilli())}, nil
	case "__date.parse":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("date.parse requires 1 argument")
		}
		sVal, ok := env[arguments[0]]
		if !ok || sVal.Type != ir.TypeString {
			return Value{}, fmt.Errorf("date.parse requires a string date")
		}
		t, err := time.Parse(time.RFC3339, sVal.String)
		if err != nil {
			t, _ = time.Parse("2006-01-02", sVal.String)
		}
		return Value{Type: ir.TypeNumber, Number: float64(t.UnixMilli())}, nil
	case "__date.UTC":
		vals := make([]float64, 7)
		vals[2] = 1 // default date = 1
		for i := 0; i < len(arguments) && i < 7; i++ {
			if v, ok := env[arguments[i]]; ok && v.Type == ir.TypeNumber {
				vals[i] = v.Number
			}
		}
		y := int(vals[0])
		if y >= 0 && y <= 99 {
			y += 1900
		}
		t := time.Date(y, time.Month(int(vals[1])+1), int(vals[2]), int(vals[3]), int(vals[4]), int(vals[5]), int(vals[6])*1e6, time.UTC)
		return Value{Type: ir.TypeNumber, Number: float64(t.UnixMilli())}, nil
	case "__date.toISOString", "__date.toJSON":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("date formatter requires timestamp argument")
		}
		tVal, ok := env[arguments[0]]
		if !ok || tVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("date formatter requires a number timestamp")
		}
		t := time.UnixMilli(int64(tVal.Number)).UTC()
		return Value{Type: ir.TypeString, String: t.Format("2006-01-02T15:04:05.000Z")}, nil
	case "__date.toString", "__date.toLocaleString", "__date.toTemporalInstant":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("date.toString requires 1 argument")
		}
		tVal, ok := env[arguments[0]]
		if !ok || tVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("date.toString requires a number timestamp")
		}
		t := time.UnixMilli(int64(tVal.Number))
		return Value{Type: ir.TypeString, String: t.Format("Mon Jan 02 2006 15:04:05")}, nil
	case "__date.toDateString", "__date.toLocaleDateString":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("date.toDateString requires 1 argument")
		}
		tVal, ok := env[arguments[0]]
		if !ok || tVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("date.toDateString requires a number timestamp")
		}
		t := time.UnixMilli(int64(tVal.Number))
		return Value{Type: ir.TypeString, String: t.Format("Mon Jan 02 2006")}, nil
	case "__date.toTimeString", "__date.toLocaleTimeString":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("date.toTimeString requires 1 argument")
		}
		tVal, ok := env[arguments[0]]
		if !ok || tVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("date.toTimeString requires a number timestamp")
		}
		t := time.UnixMilli(int64(tVal.Number))
		return Value{Type: ir.TypeString, String: t.Format("15:04:05")}, nil
	case "__date.toUTCString":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("date.toUTCString requires 1 argument")
		}
		tVal, ok := env[arguments[0]]
		if !ok || tVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("date.toUTCString requires a number timestamp")
		}
		t := time.UnixMilli(int64(tVal.Number)).UTC()
		return Value{Type: ir.TypeString, String: t.Format("Mon, 02 Jan 2006 15:04:05 GMT")}, nil
	default:
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("date operation %q requires arguments", name)
		}
		tVal, ok := env[arguments[0]]
		if !ok || tVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("date operation %q requires numeric timestamp", name)
		}
		t := time.UnixMilli(int64(tVal.Number))

		switch name {
		case "__date.getDate":
			return Value{Type: ir.TypeNumber, Number: float64(t.Day())}, nil
		case "__date.getDay":
			return Value{Type: ir.TypeNumber, Number: float64(t.Weekday())}, nil
		case "__date.getFullYear":
			return Value{Type: ir.TypeNumber, Number: float64(t.Year())}, nil
		case "__date.getHours":
			return Value{Type: ir.TypeNumber, Number: float64(t.Hour())}, nil
		case "__date.getMilliseconds":
			return Value{Type: ir.TypeNumber, Number: float64(t.Nanosecond() / 1e6)}, nil
		case "__date.getMinutes":
			return Value{Type: ir.TypeNumber, Number: float64(t.Minute())}, nil
		case "__date.getMonth":
			return Value{Type: ir.TypeNumber, Number: float64(int(t.Month()) - 1)}, nil
		case "__date.getSeconds":
			return Value{Type: ir.TypeNumber, Number: float64(t.Second())}, nil
		case "__date.getTimezoneOffset":
			_, offsetSec := t.Zone()
			return Value{Type: ir.TypeNumber, Number: float64(-offsetSec / 60)}, nil
		case "__date.getUTCDate":
			return Value{Type: ir.TypeNumber, Number: float64(t.UTC().Day())}, nil
		case "__date.getUTCDay":
			return Value{Type: ir.TypeNumber, Number: float64(t.UTC().Weekday())}, nil
		case "__date.getUTCFullYear":
			return Value{Type: ir.TypeNumber, Number: float64(t.UTC().Year())}, nil
		case "__date.getUTCHours":
			return Value{Type: ir.TypeNumber, Number: float64(t.UTC().Hour())}, nil
		case "__date.getUTCMilliseconds":
			return Value{Type: ir.TypeNumber, Number: float64(t.UTC().Nanosecond() / 1e6)}, nil
		case "__date.getUTCMinutes":
			return Value{Type: ir.TypeNumber, Number: float64(t.UTC().Minute())}, nil
		case "__date.getUTCMonth":
			return Value{Type: ir.TypeNumber, Number: float64(int(t.UTC().Month()) - 1)}, nil
		case "__date.getUTCSeconds":
			return Value{Type: ir.TypeNumber, Number: float64(t.UTC().Second())}, nil
		}

		// Setters
		if len(arguments) >= 2 {
			argVal, ok := env[arguments[1]]
			if !ok || argVal.Type != ir.TypeNumber {
				return Value{}, fmt.Errorf("setter %q requires numeric value", name)
			}
			val := argVal.Number

			switch name {
			case "__date.setDate":
				tl := t.Local()
				updated := time.Date(tl.Year(), tl.Month(), int(val), tl.Hour(), tl.Minute(), tl.Second(), tl.Nanosecond(), tl.Location())
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			case "__date.setFullYear":
				tl := t.Local()
				updated := time.Date(int(val), tl.Month(), tl.Day(), tl.Hour(), tl.Minute(), tl.Second(), tl.Nanosecond(), tl.Location())
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			case "__date.setHours":
				tl := t.Local()
				updated := time.Date(tl.Year(), tl.Month(), tl.Day(), int(val), tl.Minute(), tl.Second(), tl.Nanosecond(), tl.Location())
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			case "__date.setMilliseconds":
				tl := t.Local()
				updated := time.Date(tl.Year(), tl.Month(), tl.Day(), tl.Hour(), tl.Minute(), tl.Second(), int(val)*1e6, tl.Location())
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			case "__date.setMinutes":
				tl := t.Local()
				updated := time.Date(tl.Year(), tl.Month(), tl.Day(), tl.Hour(), int(val), tl.Second(), tl.Nanosecond(), tl.Location())
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			case "__date.setMonth":
				tl := t.Local()
				updated := time.Date(tl.Year(), time.Month(int(val)+1), tl.Day(), tl.Hour(), tl.Minute(), tl.Second(), tl.Nanosecond(), tl.Location())
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			case "__date.setSeconds":
				tl := t.Local()
				updated := time.Date(tl.Year(), tl.Month(), tl.Day(), tl.Hour(), tl.Minute(), int(val), tl.Nanosecond(), tl.Location())
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			case "__date.setUTCDate":
				tu := t.UTC()
				updated := time.Date(tu.Year(), tu.Month(), int(val), tu.Hour(), tu.Minute(), tu.Second(), tu.Nanosecond(), time.UTC)
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			case "__date.setUTCFullYear":
				tu := t.UTC()
				updated := time.Date(int(val), tu.Month(), tu.Day(), tu.Hour(), tu.Minute(), tu.Second(), tu.Nanosecond(), time.UTC)
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			case "__date.setUTCHours":
				tu := t.UTC()
				updated := time.Date(tu.Year(), tu.Month(), tu.Day(), int(val), tu.Minute(), tu.Second(), tu.Nanosecond(), time.UTC)
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			case "__date.setUTCMilliseconds":
				tu := t.UTC()
				updated := time.Date(tu.Year(), tu.Month(), tu.Day(), tu.Hour(), tu.Minute(), tu.Second(), int(val)*1e6, time.UTC)
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			case "__date.setUTCMinutes":
				tu := t.UTC()
				updated := time.Date(tu.Year(), tu.Month(), tu.Day(), tu.Hour(), int(val), tu.Second(), tu.Nanosecond(), time.UTC)
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			case "__date.setUTCMonth":
				tu := t.UTC()
				updated := time.Date(tu.Year(), time.Month(int(val)+1), tu.Day(), tu.Hour(), tu.Minute(), tu.Second(), tu.Nanosecond(), time.UTC)
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			case "__date.setUTCSeconds":
				tu := t.UTC()
				updated := time.Date(tu.Year(), tu.Month(), tu.Day(), tu.Hour(), tu.Minute(), int(val), tu.Nanosecond(), time.UTC)
				return Value{Type: ir.TypeNumber, Number: float64(updated.UnixMilli())}, nil
			}
		}

		return Value{}, fmt.Errorf("unknown date intrinsic %q", name)
	}
}

func executeOsIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__os.platform":
		platform := goruntime.GOOS
		if platform == "windows" {
			platform = "win32"
		}
		return Value{Type: ir.TypeString, String: platform}, nil
	case "__os.arch":
		arch := goruntime.GOARCH
		if arch == "amd64" {
			arch = "x64"
		} else if arch == "386" {
			arch = "ia32"
		}
		return Value{Type: ir.TypeString, String: arch}, nil
	case "__os.homedir":
		dir, _ := os.UserHomeDir()
		return Value{Type: ir.TypeString, String: dir}, nil
	case "__os.type":
		typ := "Darwin"
		if goruntime.GOOS == "linux" {
			typ = "Linux"
		} else if goruntime.GOOS == "windows" {
			typ = "Windows_NT"
		}
		return Value{Type: ir.TypeString, String: typ}, nil
	case "__os.release":
		return Value{Type: ir.TypeString, String: "1.0.0"}, nil
	case "__os.uptime":
		return Value{Type: ir.TypeNumber, Number: 3600.0}, nil
	case "__os.totalmem":
		return Value{Type: ir.TypeNumber, Number: 16.0 * 1024 * 1024 * 1024}, nil
	case "__os.freemem":
		return Value{Type: ir.TypeNumber, Number: 8.0 * 1024 * 1024 * 1024}, nil
	case "__os.tmpdir":
		return Value{Type: ir.TypeString, String: os.TempDir()}, nil
	default:
		return Value{}, fmt.Errorf("unknown os intrinsic %q", name)
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

func executeJsonIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	if len(arguments) != 1 {
		return Value{}, fmt.Errorf("JSON intrinsic requires 1 argument")
	}
	arg, ok := env[arguments[0]]
	if !ok {
		return Value{}, fmt.Errorf("unknown argument %q", arguments[0])
	}
	switch name {
	case "__json.stringify_number":
		if arg.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("JSON.stringify expects a number")
		}
		if math.IsNaN(arg.Number) || math.IsInf(arg.Number, 0) {
			return Value{Type: ir.TypeString, String: "null"}, nil
		}
		return Value{Type: ir.TypeString, String: format(arg)}, nil
	case "__json.stringify_bool":
		if arg.Type != ir.TypeBool {
			return Value{}, fmt.Errorf("JSON.stringify expects a boolean")
		}
		return Value{Type: ir.TypeString, String: strconv.FormatBool(arg.Bool)}, nil
	case "__json.stringify_string":
		if arg.Type != ir.TypeString {
			return Value{}, fmt.Errorf("JSON.stringify expects a string")
		}
		b, err := json.Marshal(arg.String)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: ir.TypeString, String: string(b)}, nil
	case "__json.stringify_number_array", "__json.stringify_string_array":
		parts := make([]string, len(arg.Array))
		for i, item := range arg.Array {
			if item.Type == ir.TypeString {
				b, _ := json.Marshal(item.String)
				parts[i] = string(b)
			} else if item.Type == ir.TypeNumber && (math.IsNaN(item.Number) || math.IsInf(item.Number, 0)) {
				parts[i] = "null"
			} else {
				parts[i] = format(item)
			}
		}
		return Value{Type: ir.TypeString, String: "[" + strings.Join(parts, ",") + "]"}, nil
	case "__json.parse_string":
		if arg.Type != ir.TypeString {
			return Value{}, fmt.Errorf("JSON.parse expects a string")
		}
		str := strings.TrimSpace(arg.String)
		if strings.HasPrefix(str, "\"") && strings.HasSuffix(str, "\"") {
			var unmarshaled string
			if err := json.Unmarshal([]byte(str), &unmarshaled); err == nil {
				return Value{Type: ir.TypeString, String: unmarshaled}, nil
			}
		}
		return Value{Type: ir.TypeString, String: str}, nil
	default:
		return Value{}, fmt.Errorf("unknown JSON intrinsic %q", name)
	}
}

type microtaskItem struct {
	closure *Closure
	arg     Value
}

var microtasks []microtaskItem

func resetMicrotasks() {
	microtasks = nil
}

func executeAsyncIntrinsic(name string, arguments []string, env map[string]Value, functions map[string]ir.Function, output *bytes.Buffer) (Value, error) {
	switch name {
	case "__async.queueMicrotask":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("queueMicrotask requires 1 argument")
		}
		closureVal, ok := env[arguments[0]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("queueMicrotask argument must be a closure")
		}
		microtasks = append(microtasks, microtaskItem{closure: closureVal.Closure})
		return Value{Type: ir.TypeVoid}, nil
	case "__async.promise_resolve":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("Promise.resolve requires 1 argument")
		}
		val, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown argument %q", arguments[0])
		}
		return Value{Type: ir.TypeObject, Object: map[string]Value{"__state": {Type: ir.TypeNumber, Number: 1}, "__value": val}}, nil
	case "__async.promise_then":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("promise.then requires promise and callback")
		}
		promiseVal, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown promise %q", arguments[0])
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("promise.then callback must be a closure")
		}
		var val Value
		if promiseVal.Object != nil {
			val = promiseVal.Object["__value"]
		}
		microtasks = append(microtasks, microtaskItem{closure: closureVal.Closure, arg: val})
		return promiseVal, nil
	case "__async.promise_catch":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("promise.catch requires promise and callback")
		}
		promiseVal, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown promise %q", arguments[0])
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("promise.catch callback must be a closure")
		}
		return promiseVal, nil
	case "__async.await":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("await requires 1 argument")
		}
		promiseVal, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown await argument %q", arguments[0])
		}
		if promiseVal.Type == ir.TypeObject && promiseVal.Object != nil {
			return promiseVal.Object["__value"], nil
		}
		return promiseVal, nil
	default:
		return Value{}, fmt.Errorf("unknown async intrinsic %q", name)
	}
}

func drainMicrotasks(functions map[string]ir.Function, output *bytes.Buffer) error {
	for len(microtasks) > 0 {
		task := microtasks[0]
		microtasks = microtasks[1:]
		var args []Value
		if task.arg.Type != "" {
			args = []Value{task.arg}
		}
		_, _, err := executeClosure(functions, task.closure, args, output)
		if err != nil {
			return err
		}
	}
	return nil
}

func executeChildProcessIntrinsic(instruction ir.Instruction, env map[string]Value) (Value, error) {
	switch instruction.Callee {
	case "__child_process.execSync":
		if len(instruction.Args) < 1 {
			return Value{}, fmt.Errorf("child_process.execSync requires at least 1 argument")
		}
		command := env[instruction.Args[0]].String
		var cwd string
		var input string
		if len(instruction.Args) > 1 && instruction.Args[1] != "" {
			cwd = env[instruction.Args[1]].String
		}
		if len(instruction.Args) > 2 && instruction.Args[2] != "" {
			input = env[instruction.Args[2]].String
		}

		cmd := exec.Command("/bin/sh", "-c", command)
		if cwd != "" {
			cmd.Dir = cwd
		}
		if input != "" {
			cmd.Stdin = strings.NewReader(input)
		}
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		err := cmd.Run()
		var exitCode float64 = 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = float64(exitErr.ExitCode())
			} else {
				exitCode = 1
			}
		}
		if instruction.Type == ir.TypeString {
			return Value{Type: ir.TypeString, String: stdoutBuf.String()}, nil
		}
		return Value{
			Type: "object:SpawnSyncReturns",
			Object: map[string]Value{
				"stdout": {Type: ir.TypeString, String: stdoutBuf.String()},
				"stderr": {Type: ir.TypeString, String: stderrBuf.String()},
				"status": {Type: ir.TypeNumber, Number: exitCode},
			},
		}, nil
	case "__child_process.spawnSync":
		if len(instruction.Args) < 1 {
			return Value{}, fmt.Errorf("child_process.spawnSync requires at least 1 argument")
		}
		command := env[instruction.Args[0]].String
		var args []string
		if len(instruction.Args) > 1 && instruction.Args[1] != "" {
			arrVal := env[instruction.Args[1]]
			for _, elem := range arrVal.Array {
				args = append(args, elem.String)
			}
		}
		var cwd string
		var input string
		if len(instruction.Args) > 2 && instruction.Args[2] != "" {
			cwd = env[instruction.Args[2]].String
		}
		if len(instruction.Args) > 3 && instruction.Args[3] != "" {
			input = env[instruction.Args[3]].String
		}

		cmd := exec.Command(command, args...)
		if cwd != "" {
			cmd.Dir = cwd
		}
		if input != "" {
			cmd.Stdin = strings.NewReader(input)
		}
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		err := cmd.Run()
		var exitCode float64 = 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = float64(exitErr.ExitCode())
			} else {
				exitCode = 1
			}
		}
		return Value{
			Type: "object:SpawnSyncReturns",
			Object: map[string]Value{
				"stdout": {Type: ir.TypeString, String: stdoutBuf.String()},
				"stderr": {Type: ir.TypeString, String: stderrBuf.String()},
				"status": {Type: ir.TypeNumber, Number: exitCode},
			},
		}, nil
	default:
		return Value{}, fmt.Errorf("unknown child_process intrinsic %q", instruction.Callee)
	}
}

func executeHttpIntrinsic(instruction ir.Instruction, env map[string]Value) (Value, error) {
	switch instruction.Callee {
	case "__http.fetchSync":
		if len(instruction.Args) < 1 {
			return Value{}, fmt.Errorf("fetchSync requires at least 1 argument (url)")
		}
		url := env[instruction.Args[0]].String
		method := "GET"
		if len(instruction.Args) > 1 && instruction.Args[1] != "" {
			if mVal, ok := env[instruction.Args[1]]; ok && mVal.String != "" {
				method = strings.ToUpper(mVal.String)
			}
		}
		var headerPairs []string
		if len(instruction.Args) > 2 && instruction.Args[2] != "" {
			if hVal, ok := env[instruction.Args[2]]; ok {
				for _, elem := range hVal.Array {
					headerPairs = append(headerPairs, elem.String)
				}
			}
		}
		var body string
		if len(instruction.Args) > 3 && instruction.Args[3] != "" {
			if bVal, ok := env[instruction.Args[3]]; ok {
				body = bVal.String
			}
		}

		var reqBody io.Reader
		if body != "" {
			reqBody = strings.NewReader(body)
		}
		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return Value{}, fmt.Errorf("fetch error creating request: %w", err)
		}
		for i := 0; i+1 < len(headerPairs); i += 2 {
			req.Header.Add(headerPairs[i], headerPairs[i+1])
		}

		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		resp, err := client.Do(req)
		var statusCode float64 = 0
		statusText := ""
		var respHeaders []Value
		respBodyStr := ""

		if err != nil {
			statusText = err.Error()
			statusCode = 0
		} else {
			defer resp.Body.Close()
			statusCode = float64(resp.StatusCode)
			statusText = resp.Status
			if idx := strings.Index(statusText, " "); idx != -1 {
				statusText = strings.TrimSpace(statusText[idx:])
			}
			respBytes, _ := io.ReadAll(resp.Body)
			respBodyStr = string(respBytes)

			for k, vList := range resp.Header {
				for _, v := range vList {
					respHeaders = append(respHeaders, Value{Type: ir.TypeString, String: strings.ToLower(k)})
					respHeaders = append(respHeaders, Value{Type: ir.TypeString, String: v})
				}
			}
		}

		return Value{
			Type: "object:FetchResponseData",
			Object: map[string]Value{
				"status":     {Type: ir.TypeNumber, Number: statusCode},
				"statusText": {Type: ir.TypeString, String: statusText},
				"headers":    {Type: ir.TypeStringArray, Array: respHeaders},
				"body":       {Type: ir.TypeString, String: respBodyStr},
			},
		}, nil
	default:
		return Value{}, fmt.Errorf("unknown http intrinsic %q", instruction.Callee)
	}
}

func executeGeneratorIntrinsic(instruction ir.Instruction, env map[string]Value, functions map[string]ir.Function, output *bytes.Buffer) (Value, error) {
	switch instruction.Callee {
	case "__generator.next":
		if len(instruction.Args) < 1 {
			return Value{}, fmt.Errorf("__generator.next requires generator argument")
		}
		genVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown generator %q", instruction.Args[0])
		}
		if genVal.Closure != nil {
			val, _, err := executeClosure(functions, genVal.Closure, nil, output)
			return val, err
		}
		if genVal.Object != nil {
			if itemsVal, hasItems := genVal.Object["__items"]; hasItems && len(itemsVal.Array) > 0 {
				state := 0
				if sVal, hasState := genVal.Object["__state"]; hasState {
					state = int(sVal.Number)
				}
				if state < len(itemsVal.Array) {
					item := itemsVal.Array[state]
					genVal.Object["__state"] = Value{Type: ir.TypeNumber, Number: float64(state + 1)}
					env[instruction.Args[0]] = genVal
					return Value{
						Type: ir.TypeObject,
						Object: map[string]Value{
							"value": item,
							"done":  {Type: ir.TypeBool, Bool: false},
						},
					}, nil
				}
				return Value{
					Type: ir.TypeObject,
					Object: map[string]Value{
						"value": {},
						"done":  {Type: ir.TypeBool, Bool: true},
					},
				}, nil
			}
			clsName := genVal.String
			if clsName == "" {
				clsName = strings.TrimPrefix(string(genVal.Type), "object:")
			}
			nextFnName := clsName + "_next"
			if nextFn, ok := functions[nextFnName]; ok {
				res, _, err := executeFunction(functions, nextFn, []Value{genVal}, output)
				if err != nil {
					return Value{}, err
				}
				env[instruction.Args[0]] = genVal
				return res, nil
			}
		}
		return Value{
			Type: ir.TypeObject,
			Object: map[string]Value{
				"value": {},
				"done":  {Type: ir.TypeBool, Bool: true},
			},
		}, nil
	}
	return Value{}, fmt.Errorf("unsupported generator intrinsic %q", instruction.Callee)
}

