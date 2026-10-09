package test262

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseMetadataReadsListsAndNegative(t *testing.T) {
	source := `// Copyright
/*---
esid: sec-addition
description: >
  multi-line description with: a colon
includes: [compareArray.js, sta.js]
flags:
  - onlyStrict
  - raw
features: [BigInt]
negative:
  phase: parse
  type: SyntaxError
---*/
x = 1;
`
	meta, err := ParseMetadata(source)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(meta.Includes, ",") != "compareArray.js,sta.js" {
		t.Errorf("includes = %v", meta.Includes)
	}
	if !meta.HasFlag("onlyStrict") || !meta.HasFlag("raw") || meta.HasFlag("module") {
		t.Errorf("flags = %v", meta.Flags)
	}
	if strings.Join(meta.Features, ",") != "BigInt" {
		t.Errorf("features = %v", meta.Features)
	}
	if meta.Negative == nil || meta.Negative.Phase != "parse" || meta.Negative.Type != "SyntaxError" {
		t.Errorf("negative = %+v", meta.Negative)
	}
}

func TestParseMetadataRequiresFrontmatter(t *testing.T) {
	if _, err := ParseMetadata("x = 1;"); err == nil {
		t.Fatal("expected an error for a file without frontmatter")
	}
}

func TestPrepareSkipsUnmodelledModesAndPrependsHarness(t *testing.T) {
	for _, meta := range []Metadata{
		{Flags: []string{"module"}},
		{Flags: []string{"async"}},
		{Flags: []string{"noStrict"}},
		{Includes: []string{"propertyHelper.js"}},
		{Negative: &Negative{Phase: "resolution", Type: "SyntaxError"}},
	} {
		if _, skip := Prepare("x;", meta); skip == "" {
			t.Errorf("expected %+v to be skipped", meta)
		}
	}
	entry, skip := Prepare("assert(true);", Metadata{Includes: []string{"sta.js"}})
	if skip != "" {
		t.Fatalf("unexpected skip: %s", skip)
	}
	if !strings.HasPrefix(entry, "// @ts-nocheck\n") || !strings.Contains(entry, "class Test262Error") || !strings.HasSuffix(entry, "assert(true);") {
		t.Fatalf("unexpected prepared entry:\n%s", entry)
	}
	early, _ := Prepare("x;", Metadata{Negative: &Negative{Phase: "parse", Type: "SyntaxError"}})
	if strings.Contains(early, "@ts-nocheck") {
		t.Fatal("early-error tests must keep TypeScript grammar checking on")
	}
	throws, _ := Prepare("assert.throws(TypeError, f);\nassert.throws(Foo, f);", Metadata{Includes: []string{"compareArray.js"}})
	if !strings.Contains(throws, `assert.throwsNamed("TypeError", f);`) || !strings.Contains(throws, "assert.throws(Foo, f);") {
		t.Fatalf("assert.throws rewrite: %s", throws[strings.LastIndex(throws, "}")+1:])
	}
	raw, _ := Prepare("1;", Metadata{Flags: []string{"raw"}})
	if strings.Contains(raw, "Test262Error") {
		t.Fatal("raw tests must not receive the harness")
	}
}

// fakeSuite writes a minimal test262 layout and returns its root.
func fakeSuite(t *testing.T, tests map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range tests {
		path := filepath.Join(root, "test", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestRunClassifiesOutcomes(t *testing.T) {
	const front = "/*---\ndescription: d\n---*/\n"
	const negParse = "/*---\nnegative:\n  phase: parse\n  type: SyntaxError\n---*/\n"
	const negRuntime = "/*---\nnegative:\n  phase: runtime\n  type: Test262Error\n---*/\n"
	root := fakeSuite(t, map[string]string{
		"a/pass.js":            front + "// exit 0",
		"a/fail.js":            front + "// exit 1",
		"a/unsupported.js":     front + "// reject",
		"a/clang.js":           front + "// clang",
		"a/panic.js":           front + "// panic",
		"a/module.js":          "/*---\nflags: [module]\n---*/\n",
		"a/syntax.js":          negParse + "// syntax",
		"a/parse-ok.js":        negParse + "// exit 0",
		"a/throws.js":          negRuntime + "// throws Test262Error",
		"a/throws-other.js":    negRuntime + "// throws TypeError",
		"a/throws-crash.js":    negRuntime + "// exit 1",
		"a/x_FIXTURE.js":       "fixture",
		"b/timeout.js":         front + "// sleep",
		"b/not-javascript.txt": "ignored",
	})
	build := func(entry, output string) error {
		data, err := os.ReadFile(entry)
		if err != nil {
			return err
		}
		source := string(data)
		switch {
		case strings.Contains(source, "// reject"):
			return errors.New("main.ts:1:1 - error SG1001: The any type is not supported in native subset.")
		case strings.Contains(source, "// panic"):
			panic("nil pointer")
		case strings.Contains(source, "// clang"):
			return errors.New("clang: exit status 1: invalid IR")
		case strings.Contains(source, "// syntax"):
			return errors.New("main.ts:1:1 - error TS1005: ';' expected.")
		}
		script := "#!/bin/sh\nexit 0\n"
		if strings.Contains(source, "// exit 1") {
			script = "#!/bin/sh\necho 'Uncaught exception object: 0x1234' >&2\nexit 1\n"
		}
		if strings.Contains(source, "// throws ") {
			name := strings.TrimSpace(source[strings.Index(source, "// throws ")+len("// throws "):])
			script = "#!/bin/sh\necho 'Uncaught exception: " + name + ": boom' >&2\nexit 1\n"
		}
		if strings.Contains(source, "// sleep") {
			script = "#!/bin/sh\nsleep 5\n"
		}
		return os.WriteFile(output, []byte(script), 0o755)
	}
	results, err := Run(context.Background(), Config{Root: root, Paths: []string{"a", "b"}, Parallel: 3, Timeout: 300 * time.Millisecond, Build: build})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Outcome{}
	for _, r := range results {
		got[r.Path] = r.Outcome
	}
	want := map[string]Outcome{
		"a/pass.js":         Pass,
		"a/fail.js":         Fail,
		"a/unsupported.js":  Unsupported,
		"a/clang.js":        Fail,
		"a/panic.js":        Fail,
		"a/module.js":       Skipped,
		"a/syntax.js":       Pass,
		"a/parse-ok.js":     Fail,
		"a/throws.js":       Pass,
		"a/throws-other.js": Fail,
		"a/throws-crash.js": Fail,
		"b/timeout.js":      Fail,
	}
	if len(got) != len(want) {
		t.Fatalf("discovered %d tests, want %d: %v", len(got), len(want), got)
	}
	for path, outcome := range want {
		if got[path] != outcome {
			t.Errorf("%s: got %s, want %s", path, got[path], outcome)
		}
	}
	summaries := Summarize(results, 1)
	total := summaries[len(summaries)-1]
	if total.Group != "total" || total.Pass != 3 || total.Fail != 7 || total.Unsupported != 1 || total.Skipped != 1 {
		t.Errorf("unexpected total summary %+v", total)
	}
	if reasons := TopReasons(results, Fail, 10); len(reasons) == 0 || !strings.Contains(strings.Join(reasons, "\n"), "0x…") {
		t.Errorf("fail reasons should normalize addresses: %v", reasons)
	}
}

func TestIsSyntaxDiagnosticAcceptsOnlyEarlyErrorCodes(t *testing.T) {
	for _, code := range []string{"TS1005", "TS1156", "TS2364", "TS2462", "TS2337", "TS18016"} {
		if !isSyntaxDiagnostic(errors.New("main.ts:1:1 - error " + code + ": message")) {
			t.Errorf("%s should count as an early error", code)
		}
	}
	for _, code := range []string{"TS2322", "TS2695", "TS7031", "TS2678", "SG1001"} {
		if isSyntaxDiagnostic(errors.New("main.ts:1:1 - error " + code + ": message")) {
			t.Errorf("%s must not count as an early error", code)
		}
	}
}

func TestRegressionsListsNonPassingResults(t *testing.T) {
	results := []Result{
		{Path: "a.js", Outcome: Pass},
		{Path: "b.js", Outcome: Fail, Detail: "boom"},
		{Path: "c.js", Outcome: Unsupported},
	}
	regressed := Regressions(results)
	if len(regressed) != 2 || regressed[0].Path != "b.js" || regressed[1].Path != "c.js" {
		t.Fatalf("Regressions = %+v", regressed)
	}
	if Regressions(results[:1]) != nil {
		t.Fatal("an all-pass run has no regressions")
	}
}
