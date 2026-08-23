package interpreter

import (
	"fmt"
	"math"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

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
	case "__string.codePointAt":
		if len(values) < 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.codePointAt requires a string")
		}
		pos := 0
		if len(values) >= 2 && values[1].Type == ir.TypeNumber {
			pos = int(values[1].Number)
		}
		str := values[0].String
		if pos < 0 || pos >= len(str) {
			return Value{Type: ir.TypeNumber, Number: math.NaN()}, nil
		}
		r := []rune(str[pos:])
		if len(r) == 0 {
			return Value{Type: ir.TypeNumber, Number: math.NaN()}, nil
		}
		return Value{Type: ir.TypeNumber, Number: float64(r[0])}, nil
	case "__string.fromCodePoint":
		if len(values) < 1 || values[0].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("string.fromCodePoint requires a number")
		}
		cp := rune(int(values[0].Number))
		return Value{Type: ir.TypeString, String: string(cp)}, nil
	case "__string.isWellFormed":
		if len(values) < 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.isWellFormed requires a string")
		}
		return Value{Type: ir.TypeBool, Bool: true}, nil
	case "__string.toWellFormed":
		if len(values) < 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.toWellFormed requires a string")
		}
		return Value{Type: ir.TypeString, String: values[0].String}, nil
	case "__string.matchAll":
		val, _, err := executeRegexStringIntrinsic("__string.match", values)
		return val, err
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
	case "__string.trimLeft":
		if len(values) < 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("trimLeft requires a string")
		}
		return Value{Type: ir.TypeString, String: strings.TrimLeft(values[0].String, " \t\n\r")}, nil
	case "__string.trimRight":
		if len(values) < 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("trimRight requires a string")
		}
		return Value{Type: ir.TypeString, String: strings.TrimRight(values[0].String, " \t\n\r")}, nil
	case "__string.toLocaleLowerCase":
		if len(values) < 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("toLocaleLowerCase requires a string")
		}
		return Value{Type: ir.TypeString, String: strings.ToLower(values[0].String)}, nil
	case "__string.toLocaleUpperCase":
		if len(values) < 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("toLocaleUpperCase requires a string")
		}
		return Value{Type: ir.TypeString, String: strings.ToUpper(values[0].String)}, nil
	case "__string.fromCharCode":
		var b strings.Builder
		for _, v := range values {
			if v.Type == ir.TypeNumber {
				b.WriteRune(rune(uint16(v.Number)))
			}
		}
		return Value{Type: ir.TypeString, String: b.String()}, nil
	case "__string.at":
		if len(values) < 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("at requires a string")
		}
		runes := []rune(values[0].String)
		idx := 0
		if len(values) >= 2 && values[1].Type == ir.TypeNumber {
			idx = int(values[1].Number)
		}
		if idx < 0 {
			idx += len(runes)
		}
		if idx >= 0 && idx < len(runes) {
			return Value{Type: ir.TypeString, String: string(runes[idx])}, nil
		}
		return Value{Type: ir.TypeString, String: ""}, nil
	case "__string.substr":
		if len(values) < 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("substr requires a string")
		}
		runes := []rune(values[0].String)
		start := 0
		if len(values) >= 2 && values[1].Type == ir.TypeNumber {
			start = int(values[1].Number)
		}
		if start < 0 {
			start += len(runes)
			if start < 0 {
				start = 0
			}
		}
		length := len(runes)
		if len(values) >= 3 && values[2].Type == ir.TypeNumber {
			length = int(values[2].Number)
		}
		if start >= len(runes) || length <= 0 {
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		end := start + length
		if end > len(runes) {
			end = len(runes)
		}
		return Value{Type: ir.TypeString, String: string(runes[start:end])}, nil
	case "__string.localeCompare":
		if len(values) < 2 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString {
			return Value{Type: ir.TypeNumber, Number: 0}, nil
		}
		s1 := values[0].String
		s2 := values[1].String
		if s1 < s2 {
			return Value{Type: ir.TypeNumber, Number: -1}, nil
		} else if s1 > s2 {
			return Value{Type: ir.TypeNumber, Number: 1}, nil
		}
		return Value{Type: ir.TypeNumber, Number: 0}, nil
	case "__string.normalize", "__string.valueOf", "__string.toString":
		if len(values) >= 1 && values[0].Type == ir.TypeString {
			return values[0], nil
		}
		return Value{Type: ir.TypeString, String: ""}, nil
	case "__string.raw":
		if len(values) >= 1 {
			if values[0].Type == ir.TypeStringArray && len(values[0].Array) > 0 {
				return values[0].Array[0], nil
			}
			return values[0], nil
		}
		return Value{Type: ir.TypeString, String: ""}, nil
	case "__string.new":
		if len(values) >= 1 {
			return Value{Type: ir.TypeString, String: format(values[0])}, nil
		}
		return Value{Type: ir.TypeString, String: ""}, nil
	case "__string.anchor":
		nameStr := ""
		if len(values) >= 2 {
			nameStr = values[1].String
		}
		return Value{Type: ir.TypeString, String: fmt.Sprintf("<a name=\"%s\">%s</a>", nameStr, values[0].String)}, nil
	case "__string.big":
		return Value{Type: ir.TypeString, String: fmt.Sprintf("<big>%s</big>", values[0].String)}, nil
	case "__string.blink":
		return Value{Type: ir.TypeString, String: fmt.Sprintf("<blink>%s</blink>", values[0].String)}, nil
	case "__string.bold":
		return Value{Type: ir.TypeString, String: fmt.Sprintf("<b>%s</b>", values[0].String)}, nil
	case "__string.fixed":
		return Value{Type: ir.TypeString, String: fmt.Sprintf("<tt>%s</tt>", values[0].String)}, nil
	case "__string.fontcolor":
		c := ""
		if len(values) >= 2 {
			c = values[1].String
		}
		return Value{Type: ir.TypeString, String: fmt.Sprintf("<font color=\"%s\">%s</font>", c, values[0].String)}, nil
	case "__string.fontsize":
		s := ""
		if len(values) >= 2 {
			s = format(values[1])
		}
		return Value{Type: ir.TypeString, String: fmt.Sprintf("<font size=\"%s\">%s</font>", s, values[0].String)}, nil
	case "__string.italics":
		return Value{Type: ir.TypeString, String: fmt.Sprintf("<i>%s</i>", values[0].String)}, nil
	case "__string.link":
		u := ""
		if len(values) >= 2 {
			u = values[1].String
		}
		return Value{Type: ir.TypeString, String: fmt.Sprintf("<a href=\"%s\">%s</a>", u, values[0].String)}, nil
	case "__string.small":
		return Value{Type: ir.TypeString, String: fmt.Sprintf("<small>%s</small>", values[0].String)}, nil
	case "__string.strike":
		return Value{Type: ir.TypeString, String: fmt.Sprintf("<strike>%s</strike>", values[0].String)}, nil
	case "__string.sub":
		return Value{Type: ir.TypeString, String: fmt.Sprintf("<sub>%s</sub>", values[0].String)}, nil
	case "__string.sup":
		return Value{Type: ir.TypeString, String: fmt.Sprintf("<sup>%s</sup>", values[0].String)}, nil
	default:
		return Value{}, fmt.Errorf("unknown string intrinsic %q", name)
	}
}
