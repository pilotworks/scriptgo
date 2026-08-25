package typescriptgo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddedStdlibLoaded(t *testing.T) {
	if err := EnsureStdlib(""); err != nil {
		t.Fatalf("EnsureStdlib failed: %v", err)
	}

	manifest := BuiltinModuleManifest()
	if len(manifest) == 0 {
		t.Fatal("expected embedded builtin modules, got 0")
	}

	expectedModules := []string{"console", "crypto", "fs", "os", "path", "process"}
	for _, modName := range expectedModules {
		mod, ok := builtinModule(modName)
		if !ok {
			t.Fatalf("expected builtin module %q to be found", modName)
		}
		if mod.Source == "" {
			t.Fatalf("expected builtin module %q to have non-empty source", modName)
		}
		if mod.Version != DefaultVersion {
			t.Fatalf("expected builtin module %q version %q, got %q", modName, DefaultVersion, mod.Version)
		}
	}

	if globalsSource == "" {
		t.Fatal("expected globalsSource to be populated from embedded globals.d.ts")
	}
}

func TestSeedVersionDeclarations(t *testing.T) {
	tempDir := t.TempDir()
	versionDir := filepath.Join(tempDir, "v0.1.0")

	if err := SeedVersionDeclarations("v0.1.0", versionDir); err != nil {
		t.Fatalf("SeedVersionDeclarations failed: %v", err)
	}

	entries, err := os.ReadDir(versionDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected files in version directory, got 0")
	}

	// Verify required stdlib files exist
	expectedFiles := []string{"globals.d.ts", "console.ts", "path.ts", "fs.ts", "os.ts", "crypto.ts", "process.ts", "net.ts", "http.ts", "stream.ts"}
	for _, expected := range expectedFiles {
		filePath := filepath.Join(versionDir, expected)
		if _, err := os.Stat(filePath); err != nil {
			t.Fatalf("expected file %s to exist in version directory: %v", expected, err)
		}
	}
}
