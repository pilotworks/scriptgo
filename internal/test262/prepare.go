package test262

import (
	_ "embed"
	"strings"
)

//go:embed harness.ts
var harnessSource string

// Harness returns the scriptgo-owned harness prepended to every test.
func Harness() string { return harnessSource }

// supportedIncludes are harness files whose behavior harness.ts provides.
var supportedIncludes = map[string]bool{"assert.js": true, "sta.js": true}

// unsupportedFlags are execution modes the runner does not model yet.
var unsupportedFlags = map[string]string{
	"module":         "ES module tests need module-goal loading",
	"async":          "asynchronous tests need the $DONE protocol",
	"noStrict":       "scriptgo compiles strict-mode modules only",
	"CanBlockIsTrue": "agent blocking is not modelled",
}

// Prepare returns the TypeScript entry source for a test, or a non-empty
// skip reason when metadata places the test outside what the runner models.
func Prepare(source string, meta Metadata) (entry string, skip string) {
	for _, flag := range meta.Flags {
		if reason, ok := unsupportedFlags[flag]; ok {
			return "", "flag " + flag + ": " + reason
		}
	}
	if meta.Negative != nil && meta.Negative.Phase == "resolution" {
		return "", "negative resolution tests need module loading"
	}
	for _, include := range meta.Includes {
		if !supportedIncludes[include] {
			return "", "include " + include + " is not provided by harness.ts"
		}
	}
	var b strings.Builder
	// Type errors are expected in conformance code (implicit any, coercions);
	// the native subset gate still validates every construct.
	b.WriteString("// @ts-nocheck\n")
	if !meta.HasFlag("raw") {
		b.WriteString(harnessSource)
		b.WriteString("\n")
	}
	b.WriteString(source)
	return b.String(), ""
}
