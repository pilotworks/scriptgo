package lowering

import (
	"math"
	"testing"
)

func TestFormatJSNumber(t *testing.T) {
	for value, want := range map[float64]string{
		0: "0", 16: "16", 1.5: "1.5", -2.25: "-2.25", 0.1: "0.1", 1e21: "1e+21", 1e20: "100000000000000000000",
		123e-7: "0.0000123", 1e-7: "1e-7", 1.5e-10: "1.5e-10", math.Inf(1): "Infinity", 2.999232: "2.999232",
	} {
		if got := formatJSNumber(value); got != want {
			t.Errorf("formatJSNumber(%v) = %q, want %q", value, got, want)
		}
	}
	if got := formatJSNumber(math.NaN()); got != "NaN" {
		t.Errorf("NaN = %q", got)
	}
}
