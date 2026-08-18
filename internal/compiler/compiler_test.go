package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileReturnsPipelineStub(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, []byte("const answer: number = 42;"), 0o644); err != nil {
		t.Fatal(err)
	}

	output, err := Compile(entry)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"scriptgo scaffold", entry, "statements: 1", "typed-ir"} {
		if !strings.Contains(output, expected) {
			t.Errorf("output does not contain %q: %s", expected, output)
		}
	}
}

func TestCompileRejectsUnsupportedEntry(t *testing.T) {
	if _, err := Compile("main.js"); err == nil {
		t.Fatal("Compile accepted a non-TypeScript entry point")
	}
}

func TestCompileRejectsTypeScriptSyntaxErrors(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "broken.ts")
	if err := os.WriteFile(entry, []byte("const answer: number = ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Compile(entry); err == nil || !strings.Contains(err.Error(), "TypeScript syntax error") {
		t.Fatalf("Compile did not report a TypeScript syntax error: %v", err)
	}
}
