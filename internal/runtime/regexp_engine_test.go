package runtime

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRegExpEngineSymbolsArePrefixed compiles the engine and checks that
// every symbol it defines carries the scriptgo_re_ prefix, so it never
// collides with the QuickJS-ng object linked by the dynamic runtime.
func TestRegExpEngineSymbolsArePrefixed(t *testing.T) {
	source, err := RegExpEngineSource()
	if err != nil {
		t.Fatal(err)
	}
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("clang is not installed")
	}
	nm, err := exec.LookPath("nm")
	if err != nil {
		t.Skip("nm is not installed")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "regexp_engine.c")
	obj := filepath.Join(dir, "regexp_engine.o")
	if err := os.WriteFile(src, source, 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(clang, "-O1", "-c", src, "-o", obj).CombinedOutput(); err != nil {
		t.Fatalf("compile regexp engine: %v\n%s", err, out)
	}
	out, err := exec.Command(nm, "-g", "--defined-only", obj).Output()
	if err != nil {
		t.Fatalf("nm: %v", err)
	}
	defined := 0
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		name := strings.TrimPrefix(fields[2], "_") // Mach-O prefixes C symbols with _
		defined++
		if !strings.HasPrefix(name, "scriptgo_re_") {
			t.Errorf("regexp engine defines unprefixed symbol %q; add it to regexpEngineSymbols", name)
		}
	}
	if defined == 0 {
		t.Fatal("regexp engine defines no symbols")
	}
}

func TestBuildRegExpEngineSourceRejectsMissingMarkers(t *testing.T) {
	if _, err := buildRegExpEngineSource("int x;"); err == nil {
		t.Fatal("expected an error for an amalgam without engine markers")
	}
}
