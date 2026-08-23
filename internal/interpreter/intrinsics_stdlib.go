package interpreter

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/pilotworks/scriptgo/internal/ir"
)

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
