package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_Init(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "scriptgo")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build scriptgo: %v\noutput: %s", err, string(out))
	}

	t.Run("init in new directory", func(t *testing.T) {
		projectDir := filepath.Join(tmpDir, "my-app")
		cmd := exec.Command(binPath, "init", projectDir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("scriptgo init failed: %v\noutput: %s", err, string(out))
		}

		outStr := string(out)
		if !strings.Contains(outStr, "Initialized ScriptGo project") {
			t.Errorf("unexpected init output: %s", outStr)
		}
		if !strings.Contains(outStr, "package.json") || !strings.Contains(outStr, "index.ts") {
			t.Errorf("expected file list in output: %s", outStr)
		}

		// Verify files exist
		for _, f := range []string{"package.json", "tsconfig.json", "index.ts", ".gitignore"} {
			if _, err := os.Stat(filepath.Join(projectDir, f)); err != nil {
				t.Errorf("expected %s to exist: %v", f, err)
			}
		}

		// Run the initialized project with scriptgo run start
		runCmd := exec.Command(binPath, "run", "start")
		runCmd.Dir = projectDir
		runCmd.Env = append(os.Environ(), "PATH="+filepath.Dir(binPath)+string(os.PathListSeparator)+os.Getenv("PATH"))
		runOut, runErr := runCmd.CombinedOutput()
		if runErr != nil {
			t.Fatalf("scriptgo run start failed: %v\noutput: %s", runErr, string(runOut))
		}
		if !strings.Contains(string(runOut), "Hello via ScriptGo!") {
			t.Errorf("unexpected run output: %s", string(runOut))
		}
	})

	t.Run("init conflict without force", func(t *testing.T) {
		conflictDir := filepath.Join(tmpDir, "conflict")
		if err := os.MkdirAll(conflictDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(conflictDir, "package.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(binPath, "init", conflictDir)
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected init to fail without force on existing package.json, output: %s", string(out))
		}
		if !strings.Contains(string(out), "already exists") {
			t.Errorf("unexpected error message: %s", string(out))
		}

		// With --force, it should succeed
		forceCmd := exec.Command(binPath, "init", "--force", conflictDir)
		forceOut, forceErr := forceCmd.CombinedOutput()
		if forceErr != nil {
			t.Fatalf("scriptgo init --force failed: %v\noutput: %s", forceErr, string(forceOut))
		}
	})

	t.Run("init with custom name", func(t *testing.T) {
		customDir := filepath.Join(tmpDir, "custom")
		cmd := exec.Command(binPath, "init", "--name", "custom-project", customDir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("scriptgo init --name failed: %v\noutput: %s", err, string(out))
		}

		content, err := os.ReadFile(filepath.Join(customDir, "package.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), `"name": "custom-project"`) {
			t.Errorf("expected package name 'custom-project', got: %s", string(content))
		}
	})

	t.Run("init with -y flag", func(t *testing.T) {
		yesDir := filepath.Join(tmpDir, "yes-app")
		cmd := exec.Command(binPath, "init", "-y", yesDir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("scriptgo init -y failed: %v\noutput: %s", err, string(out))
		}
		if !strings.Contains(string(out), "Initialized ScriptGo project") {
			t.Errorf("unexpected output: %s", string(out))
		}
		if _, err := os.Stat(filepath.Join(yesDir, "package.json")); err != nil {
			t.Fatalf("package.json not created: %v", err)
		}
	})
}
