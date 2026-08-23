package interpreter

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func execIntlIntrinsic(callee string, args []Value) (Value, error) {
	switch callee {
	case "__intl.number_format_new":
		locale := "en-US"
		if len(args) > 0 && args[0].String != "" && args[0].String != "undefined" {
			locale = args[0].String
		}
		obj := make(map[string]Value)
		obj["locale"] = Value{Type: ir.TypeString, String: locale}
		if len(args) > 1 && args[1].Object != nil {
			for k, v := range args[1].Object {
				obj[k] = v
			}
		}
		return Value{Type: ir.Type("object:Intl.NumberFormat"), Object: obj}, nil

	case "__intl.number_format_format":
		if len(args) == 0 {
			return Value{Type: ir.TypeString, String: "0"}, nil
		}
		nf := args[0]
		var num float64
		if len(args) > 1 {
			num = args[1].Number
		}
		currency := ""
		style := "decimal"
		if nf.Object != nil {
			if s, ok := nf.Object["style"]; ok && s.String != "" {
				style = s.String
			}
			if c, ok := nf.Object["currency"]; ok && c.String != "" {
				currency = c.String
			}
		}
		formatted := formatNumberLocale(num, style, currency)
		return Value{Type: ir.TypeString, String: formatted}, nil

	case "__intl.date_time_format_new":
		locale := "en-US"
		if len(args) > 0 && args[0].String != "" && args[0].String != "undefined" {
			locale = args[0].String
		}
		obj := make(map[string]Value)
		obj["locale"] = Value{Type: ir.TypeString, String: locale}
		if len(args) > 1 && args[1].Object != nil {
			for k, v := range args[1].Object {
				obj[k] = v
			}
		}
		return Value{Type: ir.Type("object:Intl.DateTimeFormat"), Object: obj}, nil

	case "__intl.date_time_format_format":
		if len(args) == 0 {
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		var t time.Time
		if len(args) > 1 {
			arg := args[1]
			if arg.Object != nil {
				if ms, ok := arg.Object["_timestamp"]; ok {
					t = time.UnixMilli(int64(ms.Number)).UTC()
				}
			} else if arg.Number != 0 {
				t = time.UnixMilli(int64(arg.Number)).UTC()
			} else {
				t = time.Now().UTC()
			}
		} else {
			t = time.Now().UTC()
		}
		formatted := fmt.Sprintf("%d/%d/%d", t.Month(), t.Day(), t.Year())
		return Value{Type: ir.TypeString, String: formatted}, nil

	case "__intl.collator_new":
		locale := "en-US"
		if len(args) > 0 && args[0].String != "" && args[0].String != "undefined" {
			locale = args[0].String
		}
		obj := make(map[string]Value)
		obj["locale"] = Value{Type: ir.TypeString, String: locale}
		return Value{Type: ir.Type("object:Intl.Collator"), Object: obj}, nil

	case "__intl.collator_compare":
		if len(args) < 3 {
			return Value{Type: ir.TypeNumber, Number: 0}, nil
		}
		s1 := args[1].String
		s2 := args[2].String
		cmp := strings.Compare(s1, s2)
		return Value{Type: ir.TypeNumber, Number: float64(cmp)}, nil

	case "__intl.segmenter_new":
		locale := "en-US"
		if len(args) > 0 && args[0].String != "" && args[0].String != "undefined" {
			locale = args[0].String
		}
		obj := make(map[string]Value)
		obj["locale"] = Value{Type: ir.TypeString, String: locale}
		return Value{Type: ir.Type("object:Intl.Segmenter"), Object: obj}, nil

	case "__intl.segmenter_segment":
		if len(args) < 2 {
			return Value{Type: ir.TypeStringArray, Array: []Value{}}, nil
		}
		text := args[1].String
		words := strings.Fields(text)
		var arr []Value
		for _, w := range words {
			arr = append(arr, Value{Type: ir.TypeString, String: w})
		}
		return Value{Type: ir.TypeStringArray, Array: arr}, nil

	case "__intl.get_canonical_locales":
		var locales []Value
		for _, a := range args {
			if len(a.Array) > 0 {
				for _, elem := range a.Array {
					locales = append(locales, Value{Type: ir.TypeString, String: canonicalizeLocale(elem.String)})
				}
			} else if a.String != "" {
				locales = append(locales, Value{Type: ir.TypeString, String: canonicalizeLocale(a.String)})
			}
		}
		return Value{Type: ir.TypeStringArray, Array: locales}, nil

	case "__intl.display_names_new":
		locale := "en-US"
		if len(args) > 0 && args[0].String != "" {
			locale = args[0].String
		}
		obj := make(map[string]Value)
		obj["locale"] = Value{Type: ir.TypeString, String: locale}
		return Value{Type: ir.Type("object:Intl.DisplayNames"), Object: obj}, nil

	case "__intl.display_names_of":
		if len(args) < 2 {
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		code := args[1].String
		name := code
		switch code {
		case "en", "en-US":
			name = "English"
		case "vi", "vi-VN":
			name = "Vietnamese"
		case "fr", "fr-FR":
			name = "French"
		case "USD":
			name = "US Dollar"
		case "VND":
			name = "Vietnamese Dong"
		}
		return Value{Type: ir.TypeString, String: name}, nil

	case "__intl.list_format_new":
		locale := "en-US"
		if len(args) > 0 && args[0].String != "" {
			locale = args[0].String
		}
		obj := make(map[string]Value)
		obj["locale"] = Value{Type: ir.TypeString, String: locale}
		return Value{Type: ir.Type("object:Intl.ListFormat"), Object: obj}, nil

	case "__intl.list_format_format":
		if len(args) < 2 {
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		var items []string
		for _, v := range args[1].Array {
			items = append(items, v.String)
		}
		if len(items) == 0 {
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		if len(items) == 1 {
			return Value{Type: ir.TypeString, String: items[0]}, nil
		}
		if len(items) == 2 {
			return Value{Type: ir.TypeString, String: items[0] + " and " + items[1]}, nil
		}
		formatted := strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
		return Value{Type: ir.TypeString, String: formatted}, nil

	case "__intl.relative_time_format_new":
		locale := "en-US"
		if len(args) > 0 && args[0].String != "" {
			locale = args[0].String
		}
		obj := make(map[string]Value)
		obj["locale"] = Value{Type: ir.TypeString, String: locale}
		return Value{Type: ir.Type("object:Intl.RelativeTimeFormat"), Object: obj}, nil

	case "__intl.relative_time_format_format":
		if len(args) < 3 {
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		val := int(args[1].Number)
		unit := args[2].String
		var res string
		if val > 0 {
			res = fmt.Sprintf("in %d %s", val, unit)
			if val > 1 && !strings.HasSuffix(unit, "s") {
				res += "s"
			}
		} else if val < 0 {
			res = fmt.Sprintf("%d %s ago", -val, unit)
			if -val > 1 && !strings.HasSuffix(unit, "s") {
				res += "s"
			}
		} else {
			res = fmt.Sprintf("now")
		}
		return Value{Type: ir.TypeString, String: res}, nil

	case "__intl.plural_rules_new":
		locale := "en-US"
		if len(args) > 0 && args[0].String != "" {
			locale = args[0].String
		}
		obj := make(map[string]Value)
		obj["locale"] = Value{Type: ir.TypeString, String: locale}
		return Value{Type: ir.Type("object:Intl.PluralRules"), Object: obj}, nil

	case "__intl.plural_rules_select":
		if len(args) < 2 {
			return Value{Type: ir.TypeString, String: "other"}, nil
		}
		val := args[1].Number
		if val == 1 {
			return Value{Type: ir.TypeString, String: "one"}, nil
		}
		return Value{Type: ir.TypeString, String: "other"}, nil
	}

	return Value{}, fmt.Errorf("unknown intl intrinsic %q", callee)
}

func canonicalizeLocale(loc string) string {
	parts := strings.Split(loc, "-")
	for i, p := range parts {
		if i == 0 {
			parts[i] = strings.ToLower(p)
		} else if len(p) == 2 {
			parts[i] = strings.ToUpper(p)
		} else if len(p) == 4 {
			parts[i] = strings.Title(strings.ToLower(p))
		}
	}
	return strings.Join(parts, "-")
}

func formatNumberLocale(n float64, style, currency string) string {
	intPart := int64(n)
	frac := n - float64(intPart)
	
	s := strconv.FormatInt(intPart, 10)
	var formatted strings.Builder
	l := len(s)
	for i, ch := range s {
		if i > 0 && (l-i)%3 == 0 && s[0] != '-' {
			formatted.WriteRune(',')
		}
		formatted.WriteRune(ch)
	}

	if frac != 0 {
		fracStr := strconv.FormatFloat(frac, 'f', 2, 64)
		if strings.Contains(fracStr, ".") {
			parts := strings.Split(fracStr, ".")
			formatted.WriteString("." + parts[1])
		}
	}

	res := formatted.String()
	if style == "currency" && currency != "" {
		if currency == "USD" {
			res = "$" + res
		} else if currency == "EUR" {
			res = "€" + res
		} else {
			res = res + " " + currency
		}
	}
	return res
}
