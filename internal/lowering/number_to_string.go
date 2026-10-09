package lowering

import (
	"math"
	"strconv"
	"strings"
)

// formatJSNumber is Number::toString (ECMA-262 6.1.6.1.20) for radix 10: the
// shortest round-trip digits laid out as JavaScript prints them. It matches
// the runtime's scriptgo_number_format so compile-time keys equal run-time
// String(n).
func formatJSNumber(value float64) string {
	switch {
	case math.IsNaN(value):
		return "NaN"
	case value == 0:
		return "0"
	case math.IsInf(value, 1):
		return "Infinity"
	case math.IsInf(value, -1):
		return "-Infinity"
	}
	sign := ""
	if value < 0 {
		sign, value = "-", -value
	}
	// d.ddddde±x with the fewest digits that round-trip.
	sci := strconv.FormatFloat(value, 'e', -1, 64)
	mantissa, exponent, _ := strings.Cut(sci, "e")
	digits := strings.Replace(mantissa, ".", "", 1)
	exp, _ := strconv.Atoi(exponent)
	k, n := len(digits), exp+1
	switch {
	case k <= n && n <= 21:
		return sign + digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		return sign + digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		return sign + "0." + strings.Repeat("0", -n) + digits
	}
	e := n - 1
	expSign := "+"
	if e < 0 {
		expSign, e = "-", -e
	}
	if k == 1 {
		return sign + digits + "e" + expSign + strconv.Itoa(e)
	}
	return sign + digits[:1] + "." + digits[1:] + "e" + expSign + strconv.Itoa(e)
}
