package pkgmgr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveLocalPackagePrefersExports(t *testing.T) {
	root := t.TempDir()
	packageRoot := filepath.Join(root, "node_modules", "demo")
	if err := os.MkdirAll(filepath.Join(packageRoot, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packageRoot, "package.json"), []byte(`{"name":"demo","exports":{".":{"import":"./src/entry.mjs","default":"./index.js"}},"main":"./wrong.js"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(packageRoot, "src", "entry.mjs")
	if err := os.WriteFile(entry, []byte("export {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	resolved, err := ResolveLocalPackage(root, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Entry != entry {
		t.Fatalf("entry = %q, want %q", resolved.Entry, entry)
	}
}

func TestResolveLocalPackageRejectsTraversal(t *testing.T) {
	if _, err := ResolveLocalPackage(t.TempDir(), "../outside"); err == nil {
		t.Fatal("expected bare-package validation error")
	}
}

func TestLockfileRoundTripIsStable(t *testing.T) {
	lock, err := NewLockfile(map[string]PackageManifest{
		"zeta":  {Name: "zeta", Dependencies: map[string]string{"alpha": "^1.0.0"}},
		"alpha": {Name: "alpha", Version: "1.2.3"},
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "scriptgo-lock.json")
	if err := WriteLockfile(path, lock); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "\"lockfileVersion\": 1") || strings.Index(text, "\"alpha\"") > strings.Index(text, "\"zeta\"") {
		t.Fatalf("lockfile is not canonical:\n%s", text)
	}
	loaded, err := LoadLockfile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Packages) != 2 || loaded.Packages["zeta"].Dependencies["alpha"] != "^1.0.0" {
		t.Fatalf("loaded lockfile = %+v", loaded)
	}
}
