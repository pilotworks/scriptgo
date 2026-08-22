package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pilotworks/scriptgo/internal/ir"
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
	for _, expected := range []string{"ModuleID = 'scriptgo'", "define i32 @main(", "ret i32 0"} {
		if !strings.Contains(output, expected) {
			t.Errorf("output does not contain %q: %s", expected, output)
		}
	}
}

func TestCompileMetadataIsDeterministic(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, []byte("console.log(20 + 22);\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	first, err := CompileWithOptions(entry, BuildOptions{Target: "arm64-apple-darwin"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := CompileWithOptions(entry, BuildOptions{Target: "arm64-apple-darwin"})
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("LLVM output is not deterministic:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	for _, expected := range []string{
		`; scriptgo.compiler = "dev"`,
		`; scriptgo.runtime-abi = "scriptgo.runtime.v1"`,
		`; scriptgo.target = "arm64-apple-darwin"`,
		`; scriptgo.source-sha256 = "`,
	} {
		if !strings.Contains(first, expected) {
			t.Errorf("LLVM metadata does not contain %q:\n%s", expected, first)
		}
	}
}

func TestCompileDebugMetadataUsesStableSourceNames(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, []byte("console.log(20 + 22);\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err := CompileWithOptions(entry, BuildOptions{Debug: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"!llvm.dbg.cu = !{!0}",
		"!DIFile(filename: \"main.ts\", directory: \".\")",
		"!DISubprogram(name: \"main\"",
		"!dbg !",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("debug LLVM output does not contain %q:\n%s", expected, output)
		}
	}
	if strings.Contains(output, entry) || strings.Contains(output, filepath.Dir(entry)) {
		t.Fatalf("debug LLVM output contains an absolute source path:\n%s", output)
	}
}

func TestCompileRejectsUnsupportedSanitizer(t *testing.T) {
	if _, err := CompileWithOptions("main.ts", BuildOptions{Sanitizers: []string{"memory"}}); err == nil || !strings.Contains(err.Error(), "unsupported sanitizer") {
		t.Fatalf("CompileWithOptions error = %v, want unsupported sanitizer diagnostic", err)
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

func TestRunSupportsBuiltinPathFunctions(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "import * as path from 'path';\nconsole.log(path.join('a', 'b'));\nconsole.log(path.dirname('a/b.txt'));\nconsole.log(path.basename('a/b.txt'));\nconsole.log(path.extname('a/b.txt'));\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err := Run(entry)
	if err != nil {
		t.Fatal(err)
	}
	if output != "a/b\na\nb.txt\n.txt\n" {
		t.Fatalf("path output = %q, want basic path behavior", output)
	}
}

func TestBuildSupportsBuiltinPathFunctions(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang is not installed")
	}
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.ts")
	output := filepath.Join(dir, "main")
	source := "import * as path from 'path';\nconsole.log(path.join('a', 'b'));\nconsole.log(path.dirname('a/b.txt'));\nconsole.log(path.basename('a/b.txt'));\nconsole.log(path.extname('a/b.txt'));\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Build(entry, output); err != nil {
		t.Fatal(err)
	}
	result, err := exec.Command(output).CombinedOutput()
	if err != nil {
		t.Fatalf("path executable failed: %v\n%s", err, result)
	}
	if string(result) != "a/b\na\nb.txt\n.txt\n" {
		t.Fatalf("native path output = %q, want basic path behavior", result)
	}
}

func TestBuiltinPathMatchesNodeReference(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.ts")
	source := "import * as path from 'path';\n" +
		"console.log(path.join('a', 'b'));\n" +
		"console.log(path.join('a', 'b', 'c'));\n" +
		"console.log(path.join('a', 'b', 'c', 'd'));\n" +
		"console.log(path.join('a/', '/b'));\n" +
		"console.log(path.join('', 'a'));\n" +
		"console.log(path.join('a', ''));\n" +
		"console.log(path.dirname('a/b.txt'));\n" +
		"console.log(path.dirname('a/b/'));\n" +
		"console.log(path.dirname('/'));\n" +
		"console.log(path.dirname('/foo'));\n" +
		"console.log(path.basename('a/b.txt'));\n" +
		"console.log(path.basename('a/b/'));\n" +
		"console.log(path.basename('//'));\n" +
		"console.log(path.extname('a/b.txt'));\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	nodeScript := "const path = require('path'); console.log(path.join('a', 'b')); console.log(path.join('a', 'b', 'c')); console.log(path.join('a', 'b', 'c', 'd')); console.log(path.join('a/', '/b')); console.log(path.join('', 'a')); console.log(path.join('a', '')); console.log(path.dirname('a/b.txt')); console.log(path.dirname('a/b/')); console.log(path.dirname('/')); console.log(path.dirname('/foo')); console.log(path.basename('a/b.txt')); console.log(path.basename('a/b/')); console.log(path.basename('//')); console.log(path.extname('a/b.txt'));"
	reference, err := exec.Command("node", "-e", nodeScript).CombinedOutput()
	if err != nil {
		t.Fatalf("node reference failed: %v\n%s", err, reference)
	}
	interpreted, err := Run(entry)
	if err != nil {
		t.Fatal(err)
	}
	if interpreted != string(reference) {
		t.Fatalf("interpreter path output = %q, Node reference = %q", interpreted, reference)
	}

	output := filepath.Join(dir, "main")
	if err := Build(entry, output); err != nil {
		t.Fatal(err)
	}
	native, err := exec.Command(output).CombinedOutput()
	if err != nil {
		t.Fatalf("native path executable failed: %v\n%s", err, native)
	}
	if string(native) != string(reference) {
		t.Fatalf("native path output = %q, Node reference = %q", native, reference)
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

func TestMathBuiltinsMatchInterpreterAndNative(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang is not installed")
	}
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.ts")
	outputPath := filepath.Join(dir, "main")
	source := "console.log(Math.abs(3.5));\n" +
		"console.log(Math.ceil(2.1));\n" +
		"console.log(Math.floor(2.9));\n" +
		"console.log(Math.trunc(2.9));\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	interpreted, err := Run(entry)
	if err != nil {
		t.Fatal(err)
	}
	if interpreted != "3.5\n3\n2\n2\n" {
		t.Fatalf("interpreter output = %q", interpreted)
	}
	if err := Build(entry, outputPath); err != nil {
		t.Fatal(err)
	}
	native, err := exec.Command(outputPath).CombinedOutput()
	if err != nil {
		t.Fatalf("native builtin executable failed: %v\n%s", err, native)
	}
	if string(native) != interpreted {
		t.Fatalf("native output = %q, interpreter output = %q", native, interpreted)
	}
}

func TestBuiltinGlobalsLowerAsNumbers(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, []byte("const nan: number = NaN; const infinity: number = Infinity;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	module, err := CompileModule(entry)
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]string{}
	for _, instruction := range module.Functions[0].Body {
		if instruction.Op == ir.OpConst {
			values[instruction.Result] = instruction.Value
		}
	}
	if values["nan"] != "NaN" || values["infinity"] != "+Inf" {
		t.Fatalf("builtin global constants = %#v", values)
	}
}

func TestBuiltinGlobalsCompileToLLVM(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, []byte("console.log(NaN); console.log(Infinity);\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err := Compile(entry)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "0x7FF8000000000000") || !strings.Contains(output, "0x7FF0000000000000") {
		t.Fatalf("LLVM output does not contain builtin number constants:\n%s", output)
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

	if output, err := Compile(entry); err != nil || !strings.Contains(output, "scriptgo_array_get") {
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

	if _, err := Compile(entry); err == nil || !strings.Contains(err.Error(), "TS1109") {
		t.Fatalf("Compile did not report a TypeScript syntax error: %v", err)
	}
}

func TestResolveClangReportsMissingToolchain(t *testing.T) {
	mockEmptyLookup := func(file string) (string, error) {
		return "", fmt.Errorf("executable file not found in $PATH")
	}
	if _, err := resolveClangWithLookup(mockEmptyLookup); err == nil || !strings.Contains(err.Error(), "requires \"clang\" in PATH") {
		t.Fatalf("resolveClang error = %v, want missing-toolchain diagnostic", err)
	}
}

func TestBuildOptionsDefaultsAndEnvironment(t *testing.T) {
	// Default values when neither struct fields nor env vars are set
	t.Setenv("SCRIPTGO_CC", "")
	t.Setenv("SCRIPTGO_TARGET", "")
	t.Setenv("CC", "ignored-generic-cc")
	t.Setenv("TARGET", "ignored-generic-target")
	opts := BuildOptions{}.normalized()
	if opts.CC != "clang" {
		t.Errorf("default CC = %q, want %q", opts.CC, "clang")
	}
	if opts.Target != "native" {
		t.Errorf("default Target = %q, want %q", opts.Target, "native")
	}

	// SCRIPTGO_CC / SCRIPTGO_TARGET environment variables
	t.Setenv("SCRIPTGO_CC", "zig cc")
	t.Setenv("SCRIPTGO_TARGET", "x86_64-linux-gnu")
	envOpts := BuildOptions{}.normalized()
	if envOpts.CC != "zig cc" {
		t.Errorf("env CC = %q, want %q", envOpts.CC, "zig cc")
	}
	if envOpts.Target != "x86_64-linux-gnu" {
		t.Errorf("env Target = %q, want %q", envOpts.Target, "x86_64-linux-gnu")
	}

	// Explicit struct field overrides environment variables
	explicitOpts := BuildOptions{CC: "gcc", Target: "aarch64-macos"}.normalized()
	if explicitOpts.CC != "gcc" {
		t.Errorf("explicit CC = %q, want %q", explicitOpts.CC, "gcc")
	}
	if explicitOpts.Target != "aarch64-macos" {
		t.Errorf("explicit Target = %q, want %q", explicitOpts.Target, "aarch64-macos")
	}
}

func TestResolveCC(t *testing.T) {
	if _, err := exec.LookPath("clang"); err == nil {
		parts, err := resolveCC("clang")
		if err != nil {
			t.Fatalf("resolveCC(clang) error = %v", err)
		}
		if len(parts) != 1 || !strings.Contains(parts[0], "clang") {
			t.Errorf("resolveCC(clang) = %v", parts)
		}
	}

	if _, err := exec.LookPath("zig"); err == nil {
		parts, err := resolveCC("zigcc")
		if err != nil {
			t.Fatalf("resolveCC(zigcc) error = %v", err)
		}
		if len(parts) < 2 || !strings.Contains(parts[0], "zig") || parts[1] != "cc" {
			t.Errorf("resolveCC(zigcc) = %v, want [zig cc]", parts)
		}
	}

	mockEmptyLookup := func(file string) (string, error) {
		return "", fmt.Errorf("executable file not found in $PATH")
	}
	if _, err := resolveCCWithLookup("nonexistent_compiler_123", mockEmptyLookup); err == nil || !strings.Contains(err.Error(), "requires \"nonexistent_compiler_123\" in PATH") {
		t.Fatalf("resolveCC error = %v, want missing-toolchain diagnostic", err)
	}
	if _, err := resolveCCWithLookup("clang", mockEmptyLookup); err == nil || !strings.Contains(err.Error(), "requires \"clang\" or \"zig\" in PATH") {
		t.Fatalf("resolveCC(clang) with empty PATH error = %v, want missing-toolchain diagnostic", err)
	}
}
