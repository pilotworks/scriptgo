package runtime_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pilotworks/scriptgo/internal/runtime"
)

func TestNativeRuntimeBenchmarks(t *testing.T) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("clang is not installed")
	}

	dir := t.TempDir()
	runtimePath := filepath.Join(dir, "runtime.c")
	headerPath := filepath.Join(dir, "scriptgo_value.h")
	benchMainPath := filepath.Join(dir, "main.c")
	executable := filepath.Join(dir, "native_bench")

	if err := os.WriteFile(runtimePath, runtime.Source, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(headerPath, []byte(runtime.ValueHeader), 0o644); err != nil {
		t.Fatal(err)
	}

	mainSrc, err := os.ReadFile(filepath.Join("native", "bench", "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(benchMainPath, mainSrc, 0o644); err != nil {
		t.Fatal(err)
	}

	compileCmd := exec.Command(clang, "-O2", benchMainPath, runtimePath, "-I", dir, "-o", executable, "-lm", "-lresolv")
	if out, err := compileCmd.CombinedOutput(); err != nil {
		t.Fatalf("clang compilation failed: %v\n%s", err, out)
	}

	runCmd := exec.Command(executable)
	out, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native bench execution failed: %v\n%s", err, out)
	}

	outputStr := string(out)
	t.Logf("Benchmark Output:\n%s", outputStr)

	requiredBenchmarks := []string{
		"array.indexOf(number)",
		"string.split",
		"web.btoa",
		"web.atob",
		"json.parse(object)",
		"json.stringify(number[])",
	}

	for _, bench := range requiredBenchmarks {
		if !strings.Contains(outputStr, bench) {
			t.Errorf("benchmark output missing %q", bench)
		}
	}
}
