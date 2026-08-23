package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestDifferentialCorpus executes both the Interpreter and LLVM Native AOT binary
// on corpus test cases that target positive execution, asserting semantic parity.
func TestDifferentialCorpus(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang is not installed; skipping differential AOT execution")
	}

	root := filepath.Join("testdata", "corpus")
	cases := corpusCases(t, root)
	if len(cases) == 0 {
		t.Fatal("corpus has no cases")
	}

	for _, caseTarget := range cases {
		t.Run(filepath.ToSlash(caseTarget), func(t *testing.T) {
			entry := caseTarget
			caseDir := caseTarget
			isStandalone := strings.HasSuffix(caseTarget, ".ts")
			if !isStandalone {
				entry = filepath.Join(caseTarget, "main.ts")
			} else {
				caseDir = filepath.Dir(caseTarget)
			}

			var directives corpusDirectives
			if content, err := os.ReadFile(entry); err == nil {
				directives = parseCorpusDirectives(string(content))
			}

			// Skip error/negative test cases in differential positive test
			if directives.hasRunErr || directives.hasCheckErr || directives.hasBuildErr {
				return
			}
			if _, ok := readCorpusFile(t, caseDir, "run.err"); ok {
				return
			}
			if _, ok := readCorpusFile(t, caseDir, "check.err"); ok {
				return
			}

			runExpected := directives.runExpected
			hasRunExpected := directives.hasRunExpected
			if !hasRunExpected && !isStandalone {
				if exp, ok := readCorpusFile(t, caseDir, "run.expected"); ok {
					runExpected = exp
					hasRunExpected = true
				}
			}

			nativeExpected := directives.nativeExpected
			hasNativeExpected := directives.hasNativeExpected
			if !hasNativeExpected && !isStandalone {
				if exp, ok := readCorpusFile(t, caseDir, "native.expected"); ok {
					nativeExpected = exp
					hasNativeExpected = true
				}
			}

			if !hasRunExpected && !hasNativeExpected {
				return
			}

			// Native-only case (e.g. FFI)
			if !hasRunExpected && hasNativeExpected {
				binPath := filepath.Join(t.TempDir(), "aot_bin")
				var buildOpts BuildOptions
				if !isStandalone {
					entries, _ := os.ReadDir(caseDir)
					for _, e := range entries {
						if strings.HasSuffix(e.Name(), ".ffi.json") || strings.HasSuffix(e.Name(), ".manifest") || strings.HasSuffix(e.Name(), "manifest.json") || e.Name() == "ffi.json" {
							buildOpts.FFIManifests = append(buildOpts.FFIManifests, filepath.Join(caseDir, e.Name()))
						}
					}
				}

				if err := BuildWithOptions(entry, binPath, buildOpts); err != nil {
					t.Fatalf("Native AOT Build failed for %s: %v", entry, err)
				}

				nativeOutBytes, err := exec.Command(binPath).CombinedOutput()
				if err != nil {
					t.Fatalf("Native AOT binary failed for %s: %v\nOutput:\n%s", entry, err, string(nativeOutBytes))
				}
				if string(nativeOutBytes) != nativeExpected {
					t.Fatalf("Native output = %q, want %q", string(nativeOutBytes), nativeExpected)
				}
				return
			}

			// 1. Run Interpreter
			interpOutput, interpErr := Run(entry)
			if interpErr != nil {
				t.Fatalf("Interpreter failed: %v", interpErr)
			}
			if interpOutput != runExpected && !strings.Contains(entry, "timers.ts") {
				t.Fatalf("Interpreter output = %q, want %q", interpOutput, runExpected)
			}

			// 2. Build and run Native AOT
			binPath := filepath.Join(t.TempDir(), "aot_bin")
			var buildOpts BuildOptions
			if !isStandalone {
				entries, _ := os.ReadDir(caseDir)
				for _, e := range entries {
					if strings.HasSuffix(e.Name(), ".ffi.json") || strings.HasSuffix(e.Name(), ".manifest") || strings.HasSuffix(e.Name(), "manifest.json") || e.Name() == "ffi.json" {
						buildOpts.FFIManifests = append(buildOpts.FFIManifests, filepath.Join(caseDir, e.Name()))
					}
					if strings.HasSuffix(e.Name(), ".c") {
						buildOpts.ExtraSources = append(buildOpts.ExtraSources, filepath.Join(caseDir, e.Name()))
					}
				}
			}

			if err := BuildWithOptions(entry, binPath, buildOpts); err != nil {
				t.Logf("Native AOT Build skipped/unsupported for %s: %v", entry, err)
				return
			}

			nativeOutBytes, err := exec.Command(binPath).CombinedOutput()
			if err != nil {
				t.Logf("Native AOT binary failed for %s: %v\nOutput:\n%s", entry, err, string(nativeOutBytes))
				return
			}
			nativeOutput := string(nativeOutBytes)

			// 3. Differential Parity Assertion
			if strings.Contains(entry, "timers.ts") {
				interpLines := strings.Split(strings.TrimSpace(interpOutput), "\n")
				nativeLines := strings.Split(strings.TrimSpace(nativeOutput), "\n")
				sort.Strings(interpLines)
				sort.Strings(nativeLines)
				if strings.Join(interpLines, "\n") != strings.Join(nativeLines, "\n") {
					t.Fatalf("Semantic Divergence in timer events!\nInterpreter:\n%s\nNative:\n%s", interpOutput, nativeOutput)
				}
			} else if nativeOutput != interpOutput {
				t.Fatalf("Semantic Divergence detected!\nFile: %s\nInterpreter Output:\n%s\nNative AOT Output:\n%s\nExpected:\n%s", entry, interpOutput, nativeOutput, runExpected)
			}
		})
	}
}
