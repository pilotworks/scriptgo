package frontend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSource(t *testing.T, dir, name, source string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestNewProgramRejectsNonTypeScriptEntry(t *testing.T) {
	entry := writeSource(t, t.TempDir(), "main.js", "console.log(1);\n")
	_, err := NewProgram(entry, "console.log(1);\n")
	if err == nil || !strings.Contains(err.Error(), ".ts extension") {
		t.Fatalf("expected extension error, got %v", err)
	}
}

func TestNewProgramRejectsEmptySource(t *testing.T) {
	entry := writeSource(t, t.TempDir(), "main.ts", "  \n")
	_, err := NewProgram(entry, "  \n")
	if err == nil || !strings.Contains(err.Error(), "is empty") {
		t.Fatalf("expected empty-source error, got %v", err)
	}
}

func TestNewProgramOrdersDependenciesBeforeImporters(t *testing.T) {
	dir := t.TempDir()
	writeSource(t, dir, "lib.ts", "export function double(value: number): number { return value * 2; }\n")
	source := "import { double } from \"./lib\";\nconsole.log(double(21));\n"
	entry := writeSource(t, dir, "main.ts", source)

	program, err := NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}
	if program.EntryPath != filepath.Clean(entry) || program.StatementCount != 2 {
		t.Fatalf("unexpected entry metadata: path=%q statements=%d", program.EntryPath, program.StatementCount)
	}
	var local []string
	for _, file := range program.Files {
		if !file.Builtin {
			local = append(local, filepath.Base(file.FileName))
		}
	}
	if strings.Join(local, ",") != "lib.ts,main.ts" {
		t.Fatalf("expected dependency before importer, got %v", local)
	}
}

func TestNewProgramReportsTypeErrorsWithSourceLocation(t *testing.T) {
	source := "const answer: number = \"forty-two\";\n"
	entry := writeSource(t, t.TempDir(), "main.ts", source)
	_, err := NewProgram(entry, source)
	if err == nil {
		t.Fatal("expected a type error")
	}
	for _, want := range []string{"main.ts", "2322"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("diagnostic %q does not contain %q", err.Error(), want)
		}
	}
}

func TestFormatSpanAnchorsLaterStageDiagnostics(t *testing.T) {
	source := "const value = compute();\n"
	got := FormatSpan("/work/main.ts", 14, 9, "error", "SG1001", "unsupported call", source)
	for _, want := range []string{"main.ts", "SG1001", "unsupported call", "compute()"} {
		if !strings.Contains(got, want) {
			t.Errorf("formatted diagnostic %q does not contain %q", got, want)
		}
	}
}
