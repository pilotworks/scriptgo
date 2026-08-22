package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type corpusDirectives struct {
	hasRunExpected    bool
	runExpected       string
	hasNativeExpected bool
	nativeExpected    string
	hasRunErr         bool
	runErr            string
	hasCheckErr       bool
	checkErr          string
	hasBuildErr       bool
	buildErr          string
}

func parseDirectiveLine(comment, prefix string) (string, bool) {
	if strings.HasPrefix(comment, prefix+":") {
		val := strings.TrimPrefix(comment, prefix+":")
		val = strings.TrimPrefix(val, " ")
		return val, true
	}
	if strings.HasPrefix(comment, prefix+" ") {
		val := strings.TrimPrefix(comment, prefix+" ")
		return val, true
	}
	if comment == prefix {
		return "", true
	}
	return "", false
}

func parseCorpusDirectives(content string) corpusDirectives {
	var d corpusDirectives
	var runLines []string
	var nativeLines []string
	var runErrLines []string
	var checkErrLines []string
	var buildErrLines []string

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmedLeading := strings.TrimLeft(line, " \t")
		if !strings.HasPrefix(trimmedLeading, "//") {
			continue
		}
		comment := strings.TrimLeft(strings.TrimPrefix(trimmedLeading, "//"), " \t")
		if val, ok := parseDirectiveLine(comment, "@expect"); ok {
			d.hasRunExpected = true
			runLines = append(runLines, val)
		} else if val, ok := parseDirectiveLine(comment, "@run.expected"); ok {
			d.hasRunExpected = true
			runLines = append(runLines, val)
		} else if val, ok := parseDirectiveLine(comment, "@native.expected"); ok {
			d.hasNativeExpected = true
			nativeLines = append(nativeLines, val)
		} else if val, ok := parseDirectiveLine(comment, "@run.err"); ok {
			d.hasRunErr = true
			runErrLines = append(runErrLines, val)
		} else if val, ok := parseDirectiveLine(comment, "@check.err"); ok {
			d.hasCheckErr = true
			checkErrLines = append(checkErrLines, val)
		} else if val, ok := parseDirectiveLine(comment, "@build.err"); ok {
			d.hasBuildErr = true
			buildErrLines = append(buildErrLines, val)
		}
	}

	if d.hasRunExpected {
		d.runExpected = strings.Join(runLines, "\n") + "\n"
	}
	if d.hasNativeExpected {
		d.nativeExpected = strings.Join(nativeLines, "\n") + "\n"
	}
	if d.hasRunErr {
		d.runErr = strings.Join(runErrLines, "\n")
	}
	if d.hasCheckErr {
		d.checkErr = strings.Join(checkErrLines, "\n")
	}
	if d.hasBuildErr {
		d.buildErr = strings.Join(buildErrLines, "\n")
	}

	return d
}

func TestCorpus(t *testing.T) {
	root := filepath.Join("testdata", "corpus")
	cases := corpusCases(t, root)
	if len(cases) == 0 {
		t.Fatal("corpus has no cases")
	}

	sanitizerEnv := os.Getenv("SCRIPTGO_SANITIZE")
	var buildOpts BuildOptions
	if sanitizerEnv != "" {
		buildOpts.Sanitizers = strings.Split(sanitizerEnv, ",")
	}

	for _, caseTarget := range cases {
		t.Run(filepath.ToSlash(caseTarget), func(t *testing.T) {
			t.Parallel()
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

			expectations := 0

			// 1. Run Expected
			runExp := directives.runExpected
			hasRunExp := directives.hasRunExpected
			if !hasRunExp && !isStandalone {
				if exp, ok := readCorpusFile(t, caseDir, "run.expected"); ok {
					runExp = exp
					hasRunExp = true
				}
			}

			if hasRunExp {
				expectations++
				got, err := Run(entry)
				if err != nil {
					t.Fatalf("Run failed: %v", err)
				}
				if got != runExp {
					t.Fatalf("Run output = %q, want %q", got, runExp)
				}
				if len(buildOpts.Sanitizers) > 0 {
					if _, err := exec.LookPath("clang"); err == nil {
						outputPath := filepath.Join(t.TempDir(), "main_sanitized")
						if err := BuildWithOptions(entry, outputPath, buildOpts); err == nil {
							nativeOut, err := exec.Command(outputPath).CombinedOutput()
							if err != nil {
								t.Fatalf("native sanitizer execution failed: %v\n%s", err, nativeOut)
							}
							if string(nativeOut) != runExp {
								t.Fatalf("native sanitizer output = %q, want %q", nativeOut, runExp)
							}
						}
					}
				}
			}

			// 2. Native Expected
			nativeExp := directives.nativeExpected
			hasNativeExp := directives.hasNativeExpected
			if !hasNativeExp && !isStandalone {
				if exp, ok := readCorpusFile(t, caseDir, "native.expected"); ok {
					nativeExp = exp
					hasNativeExp = true
				}
			}

			if hasNativeExp {
				expectations++
				if _, err := exec.LookPath("clang"); err != nil {
					t.Skip("clang is not installed")
				}
				outputPath := filepath.Join(t.TempDir(), "main")
				if err := Build(entry, outputPath); err != nil {
					t.Fatalf("Build failed: %v", err)
				}
				got, err := exec.Command(outputPath).CombinedOutput()
				if err != nil {
					t.Fatalf("native executable failed: %v\n%s", err, got)
				}
				if string(got) != nativeExp {
					t.Fatalf("native output = %q, want %q", got, nativeExp)
				}
			}

			// 3. Run Err
			runErrExp := directives.runErr
			hasRunErr := directives.hasRunErr
			if !hasRunErr && !isStandalone {
				if exp, ok := readCorpusFile(t, caseDir, "run.err"); ok {
					runErrExp = exp
					hasRunErr = true
				}
			}

			if hasRunErr {
				expectations++
				if _, err := Run(entry); err == nil || !strings.Contains(err.Error(), strings.TrimSpace(runErrExp)) {
					t.Fatalf("Run error = %v, want substring %q", err, strings.TrimSpace(runErrExp))
				}
			}

			// 4. Check Err & Build Err
			checkErrExp := directives.checkErr
			hasCheckErr := directives.hasCheckErr
			if !hasCheckErr && !isStandalone {
				if exp, ok := readCorpusFile(t, caseDir, "check.err"); ok {
					checkErrExp = exp
					hasCheckErr = true
				}
			}

			if hasCheckErr {
				expectations++
				if _, err := Compile(entry); err == nil || !strings.Contains(err.Error(), strings.TrimSpace(checkErrExp)) {
					t.Fatalf("Compile error = %v, want check.err substring %q", err, strings.TrimSpace(checkErrExp))
				}
			}

			buildErrExp := directives.buildErr
			hasBuildErr := directives.hasBuildErr
			if !hasBuildErr && !isStandalone {
				if exp, ok := readCorpusFile(t, caseDir, "build.err"); ok {
					buildErrExp = exp
					hasBuildErr = true
				}
			}

			if hasBuildErr {
				expectations++
				if _, err := Compile(entry); err == nil || !strings.Contains(err.Error(), strings.TrimSpace(buildErrExp)) {
					t.Fatalf("Compile error = %v, want build.err substring %q", err, strings.TrimSpace(buildErrExp))
				}
			}

			if expectations == 0 {
				t.Fatal("corpus case has no expectation (neither inline directives nor expectation files)")
			}
		})
	}
}

func corpusCases(t *testing.T, root string) []string {
	t.Helper()
	var cases []string
	dirsWithMain := make(map[string]bool)

	// First pass: find all directories that contain a main.ts
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && entry.Name() == "main.ts" {
			dirsWithMain[filepath.Dir(path)] = true
		}
		return nil
	})

	// Second pass: collect main.ts directories and standalone .ts files
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if dirsWithMain[path] {
				cases = append(cases, path)
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(entry.Name(), ".ts") && entry.Name() != "main.ts" {
			dir := filepath.Dir(path)
			if !dirsWithMain[dir] {
				cases = append(cases, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk corpus: %v", err)
	}
	sort.Strings(cases)
	return cases
}

func readCorpusFile(t *testing.T, casePath, name string) (string, bool) {
	t.Helper()
	path := filepath.Join(casePath, name)
	contents, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", false
	}
	if err != nil {
		t.Fatalf("read corpus expectation %q: %v", path, err)
	}
	return string(contents), true
}
