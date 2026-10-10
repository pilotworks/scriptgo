package compiler

import (
	"debug/elf"
	"os"
	"os/exec"
	"path/filepath"
	goRuntime "runtime"
	"slices"
	"strings"
	"testing"
)

func TestRuntimeLibraryFlags(t *testing.T) {
	libraries := []string{"-lz", "-lm"}
	tests := []struct {
		target string
		want   []string
	}{
		{target: "x86_64-linux-gnu", want: []string{"-Wl,--as-needed", "-lz", "-lm", "-Wl,--no-as-needed"}},
		{target: "aarch64-linux-musl", want: []string{"-Wl,--as-needed", "-lz", "-lm", "-Wl,--no-as-needed"}},
		{target: "arm64-apple-macos", want: []string{"-lz", "-lm", "-Wl,-dead_strip_dylibs"}},
		{target: "x86_64-windows-gnu", want: []string{"-lz", "-lm"}},
		{target: "wasm32-wasi", want: []string{"-lz", "-lm"}},
	}
	for _, tc := range tests {
		if got := runtimeLibraryFlags(tc.target, libraries); !slices.Equal(got, tc.want) {
			t.Errorf("runtimeLibraryFlags(%q) = %v, want %v", tc.target, got, tc.want)
		}
	}
	if got := runtimeLibraryFlags("x86_64-linux-gnu", nil); got != nil {
		t.Errorf("runtimeLibraryFlags with no libraries = %v, want nil", got)
	}
}

// TestBuildLinksOnlyReferencedLibraries checks that a program which uses no
// codec or math library does not depend on one at load time.
func TestBuildLinksOnlyReferencedLibraries(t *testing.T) {
	if goRuntime.GOOS != "linux" {
		t.Skip("inspects ELF dynamic dependencies")
	}
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang is not installed")
	}
	if !slices.Contains(linkerDCEFlags("native"), "-fuse-ld=lld") {
		// GNU ld records --as-needed libraries before --gc-sections runs, so
		// only lld drops the libraries of eliminated runtime code.
		t.Skip("lld is not installed")
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
	file, err := elf.Open(output)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	needed, err := file.ImportedLibraries()
	if err != nil {
		t.Fatal(err)
	}
	for _, library := range needed {
		for _, unused := range []string{"libz.", "libssl.", "libcrypto.", "libbrotli", "libzstd.", "libresolv.", "libm."} {
			if strings.HasPrefix(library, unused) {
				t.Errorf("hello world depends on %s; dependencies: %v", library, needed)
			}
		}
	}
}
