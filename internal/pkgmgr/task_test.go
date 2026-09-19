package pkgmgr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindPackageManifestUpwards(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, "package.json")
	if err := os.WriteFile(manifestPath, []byte(`{"name":"test-root"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	found, err := FindPackageManifest(nested)
	if err != nil {
		t.Fatalf("FindPackageManifest failed: %v", err)
	}
	if found != manifestPath {
		t.Fatalf("FindPackageManifest = %q, want %q", found, manifestPath)
	}
}

func TestResolveTaskAndListScripts(t *testing.T) {
	root := t.TempDir()
	manifestPath := filepath.Join(root, "package.json")
	manifestContent := `{
		"name": "my-app",
		"scripts": {
			"build": "echo building",
			"test": "echo testing"
		}
	}`
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0o644); err != nil {
		t.Fatal(err)
	}

	scripts, err := ListScripts(manifestPath)
	if err != nil {
		t.Fatalf("ListScripts failed: %v", err)
	}
	if len(scripts) != 2 || scripts["build"] != "echo building" || scripts["test"] != "echo testing" {
		t.Fatalf("unexpected scripts: %+v", scripts)
	}

	task, err := ResolveTask(root, "", "build")
	if err != nil {
		t.Fatalf("ResolveTask(build) failed: %v", err)
	}
	if task.Name != "build" || task.Command != "echo building" || task.ProjectRoot != root {
		t.Fatalf("unexpected task: %+v", task)
	}

	_, err = ResolveTask(root, "", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent script, got nil")
	}
	if !strings.Contains(err.Error(), "build") || !strings.Contains(err.Error(), "test") {
		t.Fatalf("error message should suggest available scripts, got: %v", err)
	}
}

func TestBuildTaskEnv(t *testing.T) {
	root := t.TempDir()
	binDir := filepath.Join(root, "node_modules", ".bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	baseEnv := []string{
		"PATH=/usr/bin:/bin",
		"USER=testuser",
	}

	env := BuildTaskEnv(root, "my-script", baseEnv)
	var foundPath, foundLifecycle bool
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			foundPath = true
			val := strings.TrimPrefix(e, "PATH=")
			if !strings.HasPrefix(val, binDir) {
				t.Fatalf("PATH should start with %q, got %q", binDir, val)
			}
		}
		if e == "npm_lifecycle_event=my-script" {
			foundLifecycle = true
		}
	}
	if !foundPath {
		t.Fatal("PATH not found in task environment")
	}
	if !foundLifecycle {
		t.Fatal("npm_lifecycle_event not found in task environment")
	}
}
