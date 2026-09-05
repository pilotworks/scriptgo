package compiler

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinkerDCEFlags(t *testing.T) {
	tests := []struct {
		target   string
		expected string
	}{
		{target: "x86_64-linux-gnu", expected: "-Wl,--gc-sections"},
		{target: "aarch64-linux-gnu", expected: "-Wl,--gc-sections"},
		{target: "aarch64-apple-darwin", expected: "-Wl,-dead_strip"},
		{target: "arm64-apple-macos", expected: "-Wl,-dead_strip"},
		{target: "x86_64-pc-windows-gnu", expected: "-Wl,--gc-sections"},
	}

	for _, tc := range tests {
		flags := linkerDCEFlags(tc.target)
		found := false
		for _, f := range flags {
			if f == tc.expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("linkerDCEFlags(%q) = %v, want to contain %q", tc.target, flags, tc.expected)
		}
	}
}

func TestBuildDeadCodeElimination(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang is not installed")
	}
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.ts")
	output := filepath.Join(dir, "main")
	if err := os.WriteFile(entry, []byte("console.log('hello');\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Build(entry, output); err != nil {
		t.Fatal(err)
	}

	// Verify execution output
	result, err := exec.Command(output).CombinedOutput()
	if err != nil {
		t.Fatalf("executable failed: %v\n%s", err, result)
	}
	if string(result) != "hello\n" {
		t.Fatalf("executable output = %q, want %q", result, "hello\n")
	}

	// If nm is available, inspect symbols to ensure unused runtime modules are eliminated
	if nmPath, err := exec.LookPath("nm"); err == nil {
		nmOut, err := exec.Command(nmPath, output).Output()
		if err == nil {
			unusedSymbols := []string{
				"_scriptgo_regex_exec",
				"_scriptgo_crypto_random_uuid",
				"_scriptgo_child_process_exec",
				"_scriptgo_array_filter_number",
			}
			for _, sym := range unusedSymbols {
				if bytes.Contains(nmOut, []byte(" T "+sym)) || bytes.Contains(nmOut, []byte(" T "+strings.TrimPrefix(sym, "_"))) {
					t.Errorf("binary unexpectedly contains unused runtime symbol %s", sym)
				}
			}
		}
	}
}

func TestBuildWithLTO(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang is not installed")
	}
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.ts")
	code := `
const pt = { x: 10, y: 20 };
console.log(pt.x + pt.y);
`
	if err := os.WriteFile(entry, []byte(code), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, ltoMode := range []string{"thin", "full"} {
		output := filepath.Join(dir, "main_"+ltoMode)
		options := BuildOptions{
			LTO:      ltoMode,
			OptLevel: "2",
		}
		if err := BuildWithOptions(entry, output, options); err != nil {
			t.Fatalf("BuildWithOptions with LTO=%s failed: %v", ltoMode, err)
		}
		result, err := exec.Command(output).CombinedOutput()
		if err != nil {
			t.Fatalf("executable (LTO=%s) failed: %v\n%s", ltoMode, err, result)
		}
		if string(result) != "30\n" {
			t.Fatalf("executable output = %q, want %q", result, "30\n")
		}
	}
}
