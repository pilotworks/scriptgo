package lowering

import "testing"

func TestBigIntLiteralValue(t *testing.T) {
	cases := map[string]string{
		"100":                  "100",
		"100n":                 "100",
		"0x10":                 "16",
		"0XFF":                 "255",
		"0o17":                 "15",
		"0b101":                "5",
		"1_000":                "1000",
		"9223372036854775807":  "9223372036854775807",
		"-9223372036854775808": "-9223372036854775808",
	}
	for text, want := range cases {
		got, ok := bigIntLiteralValue(text)
		if !ok || got != want {
			t.Errorf("bigIntLiteralValue(%q) = %q, %v; want %q", text, got, ok, want)
		}
	}
	for _, text := range []string{"0xfedcba9876543210", "9223372036854775808", "1e3", ""} {
		if got, ok := bigIntLiteralValue(text); ok {
			t.Errorf("bigIntLiteralValue(%q) = %q; want rejection", text, got)
		}
	}
}
