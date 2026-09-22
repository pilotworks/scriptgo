package pkgmgr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizePackageName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"my-app", "my-app"},
		{"My App!", "my-app"},
		{"..", "scriptgo-project"},
		{".", "scriptgo-project"},
		{"", "scriptgo-project"},
		{"@scope/package", "scope-package"},
		{"___test___", "test"},
		{"Foo_Bar.Baz", "foo_bar.baz"},
	}

	for _, tc := range tests {
		got := SanitizePackageName(tc.input)
		if got != tc.want {
			t.Errorf("SanitizePackageName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestInit_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "sample-app")

	res, err := Init(InitOptions{
		TargetDir: target,
		Name:      "sample-app",
	})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	expectedFiles := []string{"package.json", "tsconfig.json", "index.ts", ".gitignore"}
	if len(res.CreatedFiles) != len(expectedFiles) {
		t.Fatalf("CreatedFiles = %v, want %v", res.CreatedFiles, expectedFiles)
	}

	for _, f := range expectedFiles {
		p := filepath.Join(target, f)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected file %s to exist: %v", f, err)
		}
	}

	manifest, err := LoadProjectManifest(filepath.Join(target, "package.json"))
	if err != nil {
		t.Fatalf("load package.json failed: %v", err)
	}

	if manifest.Name != "sample-app" {
		t.Errorf("manifest.Name = %q, want %q", manifest.Name, "sample-app")
	}
	if manifest.Main != "index.ts" {
		t.Errorf("manifest.Main = %q, want %q", manifest.Main, "index.ts")
	}
	if manifest.Scripts["start"] != "scriptgo run index.ts" {
		t.Errorf("manifest.Scripts[start] = %q, want %q", manifest.Scripts["start"], "scriptgo run index.ts")
	}
}

func TestInit_ExistingConflict(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "conflict-app")

	if _, err := Init(InitOptions{TargetDir: target}); err != nil {
		t.Fatalf("first Init failed: %v", err)
	}

	// Running Init again without Force should error
	_, err := Init(InitOptions{TargetDir: target})
	if err == nil {
		t.Fatal("expected error on re-init without force, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error message: %v", err)
	}

	// Running Init with Force should succeed
	res, err := Init(InitOptions{TargetDir: target, Force: true})
	if err != nil {
		t.Fatalf("Init with force failed: %v", err)
	}
	if len(res.CreatedFiles) != 4 {
		t.Errorf("expected 4 files created with force, got %d", len(res.CreatedFiles))
	}
}

func TestInit_PreserveExistingOtherFiles(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "preserve-app")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}

	customTS := `console.log("custom code");`
	if err := os.WriteFile(filepath.Join(target, "index.ts"), []byte(customTS), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Init(InitOptions{TargetDir: target})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// index.ts should be in SkippedFiles, package.json in CreatedFiles
	foundSkipped := false
	for _, f := range res.SkippedFiles {
		if f == "index.ts" {
			foundSkipped = true
			break
		}
	}
	if !foundSkipped {
		t.Errorf("expected index.ts to be skipped, got skipped: %v", res.SkippedFiles)
	}

	content, err := os.ReadFile(filepath.Join(target, "index.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != customTS {
		t.Errorf("index.ts was modified, want %q, got %q", customTS, string(content))
	}
}
