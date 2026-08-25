package typescriptgo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindConfigFile(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "src", "nested")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	tsconfigPath := filepath.Join(dir, "tsconfig.json")
	if err := os.WriteFile(tsconfigPath, []byte(`{"compilerOptions":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	found := FindConfigFile(subDir)
	if filepath.Clean(found) != filepath.Clean(tsconfigPath) {
		t.Fatalf("FindConfigFile(%q) = %q, want %q", subDir, found, tsconfigPath)
	}

	notFound := FindConfigFile(t.TempDir())
	if notFound != "" {
		t.Fatalf("FindConfigFile in empty dir = %q, want empty", notFound)
	}
}

func TestCheckProjectValid(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}

	tsconfigPath := filepath.Join(dir, "tsconfig.json")
	tsconfigContent := `{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true
  },
  "include": ["src/**/*"]
}`
	if err := os.WriteFile(tsconfigPath, []byte(tsconfigContent), 0o644); err != nil {
		t.Fatal(err)
	}

	file1 := filepath.Join(srcDir, "a.ts")
	file2 := filepath.Join(srcDir, "b.ts")
	if err := os.WriteFile(file1, []byte("export const a: number = 10;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("import { a } from './a.js';\nexport const b: number = a + 20;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := CheckProject(tsconfigPath)
	if err != nil {
		t.Fatalf("CheckProject failed: %v", err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("CheckProject returned unexpected diagnostics: %+v", result.Diagnostics)
	}
	if len(result.Files) != 2 {
		t.Fatalf("CheckProject found %d files, want 2", len(result.Files))
	}
	if !result.Options.Strict {
		t.Fatalf("CheckProject strict = false, want true")
	}
}

func TestCheckProjectReportsDiagnostics(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}

	tsconfigPath := filepath.Join(dir, "tsconfig.json")
	tsconfigContent := `{
  "compilerOptions": {
    "target": "ES2022",
    "strict": true
  },
  "include": ["src/**/*"]
}`
	if err := os.WriteFile(tsconfigPath, []byte(tsconfigContent), 0o644); err != nil {
		t.Fatal(err)
	}

	file1 := filepath.Join(srcDir, "bad.ts")
	if err := os.WriteFile(file1, []byte("const x: number = 'not a number';\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := CheckProject(tsconfigPath)
	if err != nil {
		t.Fatalf("CheckProject failed: %v", err)
	}
	if len(result.Diagnostics) == 0 {
		t.Fatalf("CheckProject expected diagnostics for type error, got none")
	}
	found := false
	for _, diag := range result.Diagnostics {
		if strings.Contains(diag.Message, "Type 'string' is not assignable to type 'number'") || (strings.Contains(diag.Message, "number") && strings.Contains(diag.Message, "string")) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected assignability error in diagnostics: %+v", result.Diagnostics)
	}
}

func TestCheckDiscoversAncestorTSConfig(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}

	tsconfigPath := filepath.Join(dir, "tsconfig.json")
	// Strict is explicitly false
	tsconfigContent := `{
  "compilerOptions": {
    "strict": false,
    "noImplicitAny": false
  }
}`
	if err := os.WriteFile(tsconfigPath, []byte(tsconfigContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Implicit any parameter: in strict mode this fails, with strict: false this passes
	entryPath := filepath.Join(srcDir, "main.ts")
	entryContent := "function greet(name) { return 'hello ' + name; }\nconsole.log(greet('world'));\n"
	if err := os.WriteFile(entryPath, []byte(entryContent), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(entryPath)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Check with strict:false in tsconfig returned unexpected diagnostics: %+v", result.Diagnostics)
	}
	if result.Options.Strict {
		t.Fatalf("Check result options strict = true, want false from tsconfig")
	}
}
