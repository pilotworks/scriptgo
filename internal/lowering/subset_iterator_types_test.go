package lowering

import "testing"

func TestWithoutIteratorDefaultAny(t *testing.T) {
	cases := map[string]string{
		"Generator<string, any, any>":                      "Generator<string>",
		"AsyncGenerator<number, any, any>":                 "AsyncGenerator<number>",
		"Generator<any, any, any>":                         "Generator<any>",
		"Map<string, Generator<number, any, any>>":         "Map<string, Generator<number>>",
		"Generator<(x: number) => any, void, any>":         "Generator<(x: number) => any, void>",
		"MyGenerator<string, any>":                         "MyGenerator<string, any>",
		"IteratorResult<string, any>":                      "IteratorResult<string, any>",
		"Generator<Generator<string, any, any>, any, any>": "Generator<Generator<string>>",
	}
	for input, want := range cases {
		if got := withoutIteratorDefaultAny(input); got != want {
			t.Errorf("withoutIteratorDefaultAny(%q) = %q, want %q", input, got, want)
		}
	}
	if !isOrContainsAny("Generator<any, any, any>") || isOrContainsAny("Generator<string, any, any>") {
		t.Fatal("isOrContainsAny must ignore only the iterator default type arguments")
	}
}
