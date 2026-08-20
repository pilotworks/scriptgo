package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_MissingFileInput(t *testing.T) {
	// Build the scriptgo binary to test CLI invocations
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "scriptgo")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build scriptgo: %v\noutput: %s", err, string(out))
	}

	testCases := []struct {
		name string
		args []string
	}{
		{"run missing file", []string{"run", "nonexistent_file.ts"}},
		{"build missing file", []string{"build", "dadasdasdasapp.ts", "-o", filepath.Join(tmpDir, "out")}},
		{"check missing file", []string{"check", "missing_file.ts"}},
		{"emit missing file", []string{"emit", "missing_file.ts"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binPath, tc.args...)
			out, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("expected command to fail for missing file, but got success: %s", string(out))
			}
			outStr := string(out)
			if !strings.Contains(outStr, "file not found") {
				t.Errorf("expected output to contain 'file not found', got: %s", outStr)
			}
		})
	}
}

func TestCLI_InlineEvalFlag(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "scriptgo")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build scriptgo: %v\noutput: %s", err, string(out))
	}

	t.Run("run -e", func(t *testing.T) {
		cmd := exec.Command(binPath, "run", "-e", "console.log(42);")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("expected run -e to succeed, got: %v\noutput: %s", err, string(out))
		}
		if !strings.Contains(string(out), "42") {
			t.Errorf("expected output to contain '42', got: %s", string(out))
		}
	})

	t.Run("check -e", func(t *testing.T) {
		cmd := exec.Command(binPath, "check", "-e", "const x: number = 100; console.log(x);")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("expected check -e to succeed, got: %v\noutput: %s", err, string(out))
		}
	})
}
