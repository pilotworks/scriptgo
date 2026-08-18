package compiler

import (
	"os"
	"os/exec"
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
	for _, expected := range []string{"ModuleID = 'scriptgo'", "define i32 @main()", "ret i32 0"} {
		if !strings.Contains(output, expected) {
			t.Errorf("output does not contain %q: %s", expected, output)
		}
	}
}

func TestBuildProducesExecutable(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang is not installed")
	}
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.ts")
	output := filepath.Join(dir, "main")
	if err := os.WriteFile(entry, []byte("console.log(20 + 22);\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Build(entry, output); err != nil {
		t.Fatal(err)
	}
	result, err := exec.Command(output).CombinedOutput()
	if err != nil {
		t.Fatalf("executable failed: %v\n%s", err, result)
	}
	if string(result) != "42\n" {
		t.Fatalf("executable output = %q, want %q", result, "42\n")
	}
}

func TestRunResolvesLocalModuleAndCallsRuntime(t *testing.T) {
	dir := t.TempDir()
	dependency := filepath.Join(dir, "answer.ts")
	entry := filepath.Join(dir, "main.ts")
	if err := os.WriteFile(dependency, []byte("export const answer: number = 42;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entry, []byte("import { answer } from './answer';\nconsole.log(answer);\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	output, err := Run(entry)
	if err != nil {
		t.Fatal(err)
	}
	if output != "42\n" {
		t.Fatalf("runtime output = %q, want %q", output, "42\n")
	}
}

func TestRunInitializesLocalModulesBeforeEntry(t *testing.T) {
	dir := t.TempDir()
	dependency := filepath.Join(dir, "dependency.ts")
	entry := filepath.Join(dir, "main.ts")
	if err := os.WriteFile(dependency, []byte("console.log('dependency');\nexport const answer: number = 42;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entry, []byte("import { answer } from './dependency';\nconsole.log('entry');\nconsole.log(answer);\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	output, err := Run(entry)
	if err != nil {
		t.Fatal(err)
	}
	if output != "dependency\nentry\n42\n" {
		t.Fatalf("runtime output = %q, want dependency initialization before entry", output)
	}
}

func TestRunSupportsNumberArrayIndexing(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, []byte("const values: number[] = [10, 20];\nconsole.log(values[1]);\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	output, err := Run(entry)
	if err != nil {
		t.Fatal(err)
	}
	if output != "20\n" {
		t.Fatalf("runtime output = %q, want %q", output, "20\n")
	}
}

func TestRunSupportsStaticClassFields(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "class Point { x: number = 42; }\nconst point = new Point();\nconsole.log(point.x);\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	output, err := Run(entry)
	if err != nil {
		t.Fatal(err)
	}
	if output != "42\n" {
		t.Fatalf("class output = %q, want %q", output, "42\n")
	}
}

func TestBuildSupportsStaticClassFields(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang is not installed")
	}
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.ts")
	output := filepath.Join(dir, "main")
	source := "class Point { x: number = 42; }\nconst point = new Point();\nconsole.log(point.x);\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Build(entry, output); err != nil {
		t.Fatal(err)
	}
	result, err := exec.Command(output).CombinedOutput()
	if err != nil {
		t.Fatalf("class executable failed: %v\n%s", err, result)
	}
	if string(result) != "42\n" {
		t.Fatalf("class executable output = %q, want %q", result, "42\n")
	}
}

func TestCompileSupportsNumberArrays(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, []byte("const values: number[] = [10, 20];\nconsole.log(values[1]);\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if output, err := Compile(entry); err != nil || !strings.Contains(output, "scriptgo_array_number_get") {
		t.Fatalf("Compile output/error = %q / %v, want array runtime call", output, err)
	}
}

func TestDumpIRReturnsStableArtifact(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, []byte("console.log(20 + 22);\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	output, err := DumpIR(entry)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"module ", "function main()", "binary \"+\"", "print console.log [%t2]"} {
		if !strings.Contains(output, expected) {
			t.Errorf("IR output does not contain %q:\n%s", expected, output)
		}
	}
}

func TestBuildProducesArrayExecutable(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang is not installed")
	}
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.ts")
	output := filepath.Join(dir, "main")
	if err := os.WriteFile(entry, []byte("const values: number[] = [10, 20];\nconsole.log(values[1]);\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Build(entry, output); err != nil {
		t.Fatal(err)
	}
	result, err := exec.Command(output).CombinedOutput()
	if err != nil {
		t.Fatalf("executable failed: %v\n%s", err, result)
	}
	if string(result) != "20\n" {
		t.Fatalf("executable output = %q, want %q", result, "20\n")
	}
}

func TestBuildReportsNativeArrayBoundsFailure(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang is not installed")
	}
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.ts")
	output := filepath.Join(dir, "main")
	if err := os.WriteFile(entry, []byte("const values: number[] = [10];\nconsole.log(values[1]);\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Build(entry, output); err != nil {
		t.Fatal(err)
	}
	result, err := exec.Command(output).CombinedOutput()
	if err == nil || !strings.Contains(string(result), "scriptgo array index out of bounds") {
		t.Fatalf("executable output/error = %q / %v, want native bounds diagnostic", result, err)
	}
}

func BenchmarkBuildNative(b *testing.B) {
	if _, err := exec.LookPath("clang"); err != nil {
		b.Skip("clang is not installed")
	}
	dir := b.TempDir()
	entry := filepath.Join(dir, "main.ts")
	output := filepath.Join(dir, "main")
	if err := os.WriteFile(entry, []byte("console.log(20 + 22);\n"), 0o644); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Build(entry, output); err != nil {
			b.Fatal(err)
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

func TestResolveClangReportsMissingToolchain(t *testing.T) {
	t.Setenv("PATH", "")
	if _, err := resolveClang(); err == nil || !strings.Contains(err.Error(), "requires clang in PATH") {
		t.Fatalf("resolveClang error = %v, want missing-toolchain diagnostic", err)
	}
}
