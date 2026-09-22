package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_TaskRunner(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "scriptgo")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build scriptgo: %v\noutput: %s", err, string(out))
	}

	projectDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}

	manifestContent := `{
		"name": "task-test-project",
		"scripts": {
			"hello": "echo Hello from ScriptGo Task Runner",
			"echo-args": "echo Args:",
			"exit-code": "exit 42"
		}
	}`
	if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(manifestContent), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("list available scripts", func(t *testing.T) {
		cmd := exec.Command(binPath, "task", "--project", projectDir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("task list failed: %v\noutput: %s", err, string(out))
		}
		outStr := string(out)
		if !strings.Contains(outStr, "hello") || !strings.Contains(outStr, "echo-args") {
			t.Fatalf("expected script names in list, got: %s", outStr)
		}
	})

	t.Run("run task directly", func(t *testing.T) {
		cmd := exec.Command(binPath, "task", "--project", projectDir, "hello")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("task hello failed: %v\noutput: %s", err, string(out))
		}
		if !strings.Contains(string(out), "Hello from ScriptGo Task Runner") {
			t.Fatalf("unexpected output: %s", string(out))
		}
	})

	t.Run("run task with forwarded arguments", func(t *testing.T) {
		cmd := exec.Command(binPath, "task", "--project", projectDir, "echo-args", "--", "foo", "bar")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("task with args failed: %v\noutput: %s", err, string(out))
		}
		if !strings.Contains(string(out), "Args: foo bar") {
			t.Fatalf("expected args to be forwarded, got: %s", string(out))
		}
	})

	t.Run("run task exit code propagation", func(t *testing.T) {
		cmd := exec.Command(binPath, "task", "--project", projectDir, "exit-code")
		err := cmd.Run()
		if err == nil {
			t.Fatal("expected non-zero exit code")
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 42 {
				t.Fatalf("exit code = %d, want 42", exitErr.ExitCode())
			}
		} else {
			t.Fatalf("unexpected error type: %v", err)
		}
	})

	t.Run("fallback from scriptgo run", func(t *testing.T) {
		cmd := exec.Command(binPath, "run", "hello")
		cmd.Dir = projectDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("scriptgo run fallback failed: %v\noutput: %s", err, string(out))
		}
		if !strings.Contains(string(out), "Hello from ScriptGo Task Runner") {
			t.Fatalf("expected fallback to execute hello script, got: %s", string(out))
		}
	})

	t.Run("prioritize scriptgo run cmd over file with same name", func(t *testing.T) {
		helloFile := filepath.Join(projectDir, "hello")
		if err := os.WriteFile(helloFile, []byte("console.log('from-file-output');"), 0o644); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(helloFile)

		cmd := exec.Command(binPath, "run", "hello")
		cmd.Dir = projectDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("scriptgo run hello failed: %v\noutput: %s", err, string(out))
		}
		if !strings.Contains(string(out), "Hello from ScriptGo Task Runner") {
			t.Fatalf("expected package.json script to take priority over file, got: %s", string(out))
		}
	})

	t.Run("run task with binary from node_modules/.bin in PATH", func(t *testing.T) {
		binDir := filepath.Join(projectDir, "node_modules", ".bin")
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			t.Fatal(err)
		}
		dummyBin := filepath.Join(binDir, "my-tool")
		if err := os.WriteFile(dummyBin, []byte("#!/bin/sh\necho tool-output-ok\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		customManifest := `{
			"name": "task-test-project",
			"scripts": {
				"use-tool": "my-tool"
			}
		}`
		if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(customManifest), 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(binPath, "task", "--project", projectDir, "use-tool")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("task use-tool failed: %v\noutput: %s", err, string(out))
		}
		if !strings.Contains(string(out), "tool-output-ok") {
			t.Fatalf("expected binary from node_modules/.bin to be invoked, got: %s", string(out))
		}
	})
}
