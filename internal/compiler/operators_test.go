package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOperatorsEndToEnd(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{
			name: "strict equality and inequality numbers",
			source: `
const a: number = 42;
const b: number = 42;
const c: number = 10;
console.log(a === b);
console.log(a !== c);
console.log(a == b);
console.log(a != c);
`,
			expected: "true\ntrue\ntrue\ntrue\n",
		},
		{
			name: "string equality and inequality",
			source: `
const s1: string = "hello";
const s2: string = "hello";
const s3: string = "world";
console.log(s1 === s2);
console.log(s1 !== s3);
console.log(s1 == s2);
console.log(s1 != s3);
`,
			expected: "true\ntrue\ntrue\ntrue\n",
		},
		{
			name: "boolean logic and unary not",
			source: `
const t: boolean = true;
const f: boolean = false;
console.log(t && !f);
console.log(f || t);
console.log(!t || f);
`,
			expected: "true\ntrue\nfalse\n",
		},
		{
			name: "unary minus and ternary operator",
			source: `
const x: number = -10;
const y: number = -x;
const chosen: number = y > 0 ? y + 2 : 0;
console.log(chosen);
`,
			expected: "12\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			entry := filepath.Join(dir, "main.ts")
			if err := os.WriteFile(entry, []byte(strings.TrimSpace(tt.source)+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}

			// 1. Test via Interpreter (--emit run)
			runOutput, err := Run(entry)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
			if runOutput != tt.expected {
				t.Fatalf("Run output = %q, want %q", runOutput, tt.expected)
			}

			// 2. Test via Native Clang Build (--emit exe)
			if _, err := exec.LookPath("clang"); err == nil {
				bin := filepath.Join(dir, "main_bin")
				if err := Build(entry, bin); err != nil {
					t.Fatalf("Build failed: %v", err)
				}
				binOutput, err := exec.Command(bin).CombinedOutput()
				if err != nil {
					t.Fatalf("Native execution failed: %v\nOutput: %s", err, binOutput)
				}
				if string(binOutput) != tt.expected {
					t.Fatalf("Native output = %q, want %q", string(binOutput), tt.expected)
				}
			}
		})
	}
}
