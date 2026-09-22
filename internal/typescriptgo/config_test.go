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
	tsconfigContent := `{
  "compilerOptions": {
    "target": "ES2022",
    "strict": true
  }
}`
	if err := os.WriteFile(tsconfigPath, []byte(tsconfigContent), 0o644); err != nil {
		t.Fatal(err)
	}

	entryPath := filepath.Join(srcDir, "main.ts")
	entryContent := "export function greet(name: string): string { return 'hello ' + name; }\nconsole.log(greet('world'));\n"
	if err := os.WriteFile(entryPath, []byte(entryContent), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(entryPath)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Check with valid tsconfig returned unexpected diagnostics: %+v", result.Diagnostics)
	}
	if !result.Options.Strict {
		t.Fatalf("Check result options strict = false, want true from tsconfig")
	}
	if result.Options.Target != "ES2022" {
		t.Fatalf("Check result options target = %q, want ES2022 from tsconfig", result.Options.Target)
	}
}

func hasDiagnosticCode(diags []Diagnostic, code int32) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func TestValidateCompilerOptions_RejectsInvalidSettings(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.ts"), []byte("export const x = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// 1. Rejects strict: false
	strictFalsePath := filepath.Join(dir, "strict_false.json")
	if err := os.WriteFile(strictFalsePath, []byte(`{"compilerOptions":{"strict":false}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := CheckProject(strictFalsePath)
	if err != nil {
		t.Fatalf("CheckProject failed: %v", err)
	}
	if !hasDiagnosticCode(res.Diagnostics, 6001) {
		t.Fatalf("expected SG6001 diagnostic for strict: false, got: %+v", res.Diagnostics)
	}

	// 2. Rejects target < ES2020
	es5Path := filepath.Join(dir, "target_es5.json")
	if err := os.WriteFile(es5Path, []byte(`{"compilerOptions":{"target":"ES5","strict":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err = CheckProject(es5Path)
	if err != nil {
		t.Fatalf("CheckProject failed: %v", err)
	}
	if !hasDiagnosticCode(res.Diagnostics, 6002) {
		t.Fatalf("expected SG6002 diagnostic for target: ES5, got: %+v", res.Diagnostics)
	}

	// 3. Rejects moduleResolution: classic
	classicPath := filepath.Join(dir, "classic_res.json")
	if err := os.WriteFile(classicPath, []byte(`{"compilerOptions":{"moduleResolution":"classic","strict":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err = CheckProject(classicPath)
	if err != nil {
		t.Fatalf("CheckProject failed: %v", err)
	}
	if !hasDiagnosticCode(res.Diagnostics, 6003) {
		t.Fatalf("expected SG6003 diagnostic for moduleResolution: classic, got: %+v", res.Diagnostics)
	}

	// 4. Rejects module: AMD
	amdPath := filepath.Join(dir, "module_amd.json")
	if err := os.WriteFile(amdPath, []byte(`{"compilerOptions":{"module":"AMD","strict":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err = CheckProject(amdPath)
	if err != nil {
		t.Fatalf("CheckProject failed: %v", err)
	}
	if !hasDiagnosticCode(res.Diagnostics, 6004) {
		t.Fatalf("expected SG6004 diagnostic for module: AMD, got: %+v", res.Diagnostics)
	}
}
