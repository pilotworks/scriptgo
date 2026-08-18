package typescriptgo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckResolvesLocalModules(t *testing.T) {
	dir := t.TempDir()
	dependency := filepath.Join(dir, "answer.ts")
	entry := filepath.Join(dir, "main.ts")
	if err := os.WriteFile(dependency, []byte("export const answer: number = 42;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entry, []byte("import { answer } from './answer';\nconsole.log(answer);\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 2 {
		t.Fatalf("Check resolved %d files, want 2: %+v", len(result.Files), result.Files)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Check returned unexpected diagnostics: %+v", result.Diagnostics)
	}
}

func TestCheckReportsSemanticDiagnostics(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, []byte("const answer: number = 'not a number';\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) == 0 {
		t.Fatal("Check returned no semantic diagnostics")
	}
	found := false
	for _, diagnostic := range result.Diagnostics {
		if strings.Contains(diagnostic.Message, "number") && strings.Contains(diagnostic.Message, "string") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Check diagnostics do not describe the type error: %+v", result.Diagnostics)
	}
}
