package interpreter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func compileRegex(pattern, flags string) (*regexp.Regexp, error) {
	prefix := ""
	if strings.Contains(flags, "i") {
		prefix += "(?i)"
	}
	if strings.Contains(flags, "m") {
		prefix += "(?m)"
	}
	if strings.Contains(flags, "s") {
		prefix += "(?s)"
	}
	return regexp.Compile(prefix + pattern)
}

func executeRegexIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	values := make([]Value, 0, len(arguments))
	for _, argument := range arguments {
		value, ok := env[argument]
		if !ok {
			return Value{}, fmt.Errorf("unknown regex intrinsic argument %q", argument)
		}
		values = append(values, value)
	}

	switch name {
	case "__regex.test":
		if len(values) != 3 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString || values[2].Type != ir.TypeString {
			return Value{}, fmt.Errorf("__regex.test requires pattern, flags, and str")
		}
		re, err := compileRegex(values[0].String, values[1].String)
		if err != nil {
			return Value{}, fmt.Errorf("invalid regex %q: %w", values[0].String, err)
		}
		return Value{Type: ir.TypeBool, Bool: re.MatchString(values[2].String)}, nil

	case "__regex.exec":
		if len(values) != 3 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString || values[2].Type != ir.TypeString {
			return Value{}, fmt.Errorf("__regex.exec requires pattern, flags, and str")
		}
		re, err := compileRegex(values[0].String, values[1].String)
		if err != nil {
			return Value{}, fmt.Errorf("invalid regex %q: %w", values[0].String, err)
		}
		matches := re.FindStringSubmatch(values[2].String)
		arr := make([]Value, 0, len(matches))
		for _, m := range matches {
			arr = append(arr, Value{Type: ir.TypeString, String: m})
		}
		return Value{Type: ir.TypeStringArray, Array: arr}, nil

	default:
		return Value{}, fmt.Errorf("unknown regex intrinsic %q", name)
	}
}

func executeBigIntIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	values := make([]Value, 0, len(arguments))
	for _, argument := range arguments {
		value, ok := env[argument]
		if !ok {
			return Value{}, fmt.Errorf("unknown bigint intrinsic argument %q", argument)
		}
		values = append(values, value)
	}

	switch name {
	case "__bigint.fromNumber":
		if len(values) != 1 || values[0].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("__bigint.fromNumber requires one number")
		}
		return Value{Type: ir.TypeBigInt, BigInt: int64(values[0].Number)}, nil

	case "__bigint.fromString":
		if len(values) != 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("__bigint.fromString requires one string")
		}
		clean := strings.TrimSuffix(values[0].String, "n")
		bi, err := strconv.ParseInt(clean, 10, 64)
		if err != nil {
			return Value{}, fmt.Errorf("invalid bigint string %q: %w", values[0].String, err)
		}
		return Value{Type: ir.TypeBigInt, BigInt: bi}, nil

	default:
		return Value{}, fmt.Errorf("unknown bigint intrinsic %q", name)
	}
}

func executeRegexStringIntrinsic(name string, values []Value) (Value, bool, error) {
	switch name {
	case "__string.fromBigInt":
		if len(values) != 1 || values[0].Type != ir.TypeBigInt {
			return Value{}, true, fmt.Errorf("string.fromBigInt requires one bigint")
		}
		return Value{Type: ir.TypeString, String: strconv.FormatInt(values[0].BigInt, 10)}, true, nil

	case "__string.match":
		if len(values) != 3 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString || values[2].Type != ir.TypeString {
			return Value{}, true, fmt.Errorf("string.match requires str, pattern, and flags")
		}
		re, err := compileRegex(values[1].String, values[2].String)
		if err != nil {
			return Value{}, true, fmt.Errorf("invalid regex %q: %w", values[1].String, err)
		}
		var matches []string
		if strings.Contains(values[2].String, "g") {
			matches = re.FindAllString(values[0].String, -1)
		} else {
			matches = re.FindStringSubmatch(values[0].String)
		}
		arr := make([]Value, 0, len(matches))
		for _, m := range matches {
			arr = append(arr, Value{Type: ir.TypeString, String: m})
		}
		return Value{Type: ir.TypeStringArray, Array: arr}, true, nil

	case "__string.search":
		if len(values) != 3 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString || values[2].Type != ir.TypeString {
			return Value{}, true, fmt.Errorf("string.search requires str, pattern, and flags")
		}
		re, err := compileRegex(values[1].String, values[2].String)
		if err != nil {
			return Value{}, true, fmt.Errorf("invalid regex %q: %w", values[1].String, err)
		}
		loc := re.FindStringIndex(values[0].String)
		if loc == nil {
			return Value{Type: ir.TypeNumber, Number: -1}, true, nil
		}
		return Value{Type: ir.TypeNumber, Number: float64(loc[0])}, true, nil

	case "__string.replace_regex":
		if len(values) != 4 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString || values[2].Type != ir.TypeString || values[3].Type != ir.TypeString {
			return Value{}, true, fmt.Errorf("string.replace_regex requires str, pattern, flags, and replacement")
		}
		re, err := compileRegex(values[1].String, values[2].String)
		if err != nil {
			return Value{}, true, fmt.Errorf("invalid regex %q: %w", values[1].String, err)
		}
		src := values[0].String
		repl := values[3].String
		if strings.Contains(values[2].String, "g") {
			return Value{Type: ir.TypeString, String: re.ReplaceAllString(src, repl)}, true, nil
		}
		loc := re.FindStringIndex(src)
		if loc == nil {
			return Value{Type: ir.TypeString, String: src}, true, nil
		}
		res := src[:loc[0]] + repl + src[loc[1]:]
		return Value{Type: ir.TypeString, String: res}, true, nil

	default:
		return Value{}, false, nil
	}
}
