package main

import (
	"os"
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

func TestCLI_DynamicCompatibilitySurface(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "scriptgo")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build scriptgo: %v\noutput: %s", err, string(out))
	}
	for _, command := range []string{"run", "build", "check", "emit", "coverage"} {
		cmd := exec.Command(binPath, command, "--help")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s --help failed: %v\n%s", command, err, out)
		}
		if !strings.Contains(string(out), "--dynamic") {
			t.Errorf("%s --help does not list --dynamic: %s", command, out)
		}
	}

	entry := filepath.Join(tmpDir, "main.ts")
	if err := os.WriteFile(entry, []byte("const value: any = 42;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command(binPath, "coverage", "--dynamic", entry)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("coverage failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "ScriptGo coverage summary") || !strings.Contains(string(out), "Result: dynamic-runtime-required") {
		t.Fatalf("unexpected compatibility output: %s", out)
	}
	if strings.HasPrefix(strings.TrimSpace(string(out)), "{") {
		t.Fatalf("default compatibility output must not be JSON: %s", out)
	}

	cmd = exec.Command(binPath, "coverage", "--format", "json", "--dynamic", entry)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("JSON coverage failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), `"mode": "dynamic-enabled"`) || !strings.Contains(string(out), `"tier": "dynamic"`) {
		t.Fatalf("unexpected JSON compatibility output: %s", out)
	}

	cmd = exec.Command(binPath, "coverage", "--format", "yaml", entry)
	out, err = cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "unsupported coverage format") {
		t.Fatalf("invalid coverage format output/error = %s / %v", out, err)
	}

	cmd = exec.Command(binPath, "check", "--dynamic", entry)
	out, err = cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "SG5001") {
		t.Fatalf("dynamic check output/error = %s / %v, want SG5001", out, err)
	}
}

func TestCLI_CheckTSConfig(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "scriptgo")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build scriptgo: %v\noutput: %s", err, string(out))
	}

	projDir := filepath.Join(tmpDir, "project")
	srcDir := filepath.Join(projDir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}

	tsconfigPath := filepath.Join(projDir, "tsconfig.json")
	tsconfigContent := `{
  "compilerOptions": {
    "target": "ES2022",
    "strict": true
  },
  "include": ["src/**/*"]
}`
	if err := os.WriteFile(tsconfigPath, []byte(tsconfigContent), 0o644); err != nil {
		t.Fatal(err)
	}

	validTS := filepath.Join(srcDir, "valid.ts")
	if err := os.WriteFile(validTS, []byte("export const msg: string = 'hello';\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("check tsconfig.json explicitly", func(t *testing.T) {
		cmd := exec.Command(binPath, "check", tsconfigPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("expected check tsconfig.json to succeed, got %v: %s", err, string(out))
		}
	})

	t.Run("check -p project dir", func(t *testing.T) {
		cmd := exec.Command(binPath, "check", "-p", projDir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("expected check -p to succeed, got %v: %s", err, string(out))
		}
	})

	t.Run("check in cwd", func(t *testing.T) {
		cmd := exec.Command(binPath, "check")
		cmd.Dir = projDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("expected check in cwd to succeed, got %v: %s", err, string(out))
		}
	})

	t.Run("check type error diagnostics", func(t *testing.T) {
		badTS := filepath.Join(srcDir, "bad.ts")
		if err := os.WriteFile(badTS, []byte("const val: number = 'text';\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binPath, "check", "-p", projDir)
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected check on project with type error to fail, got success")
		}
		outStr := string(out)
		if !strings.Contains(outStr, "TS2322") && !strings.Contains(outStr, "number") {
			t.Errorf("expected TS diagnostic in output, got: %s", outStr)
		}
	})
}
