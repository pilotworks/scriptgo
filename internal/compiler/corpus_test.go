package compiler

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestCorpus(t *testing.T) {
	root := filepath.Join("testdata", "corpus")
	cases := corpusCases(t, root)
	if len(cases) == 0 {
		t.Fatal("corpus has no cases")
	}

	for _, casePath := range cases {
		casePath := casePath
		t.Run(filepath.ToSlash(casePath), func(t *testing.T) {
			entry := filepath.Join(casePath, "main.ts")
			expectations := 0
			if expected, ok := readCorpusFile(t, casePath, "run.out"); ok {
				expectations++
				got, err := Run(entry)
				if err != nil {
					t.Fatalf("Run failed: %v", err)
				}
				if got != expected {
					t.Fatalf("Run output = %q, want %q", got, expected)
				}
			}
			if expected, ok := readCorpusFile(t, casePath, "run.err"); ok {
				expectations++
				if _, err := Run(entry); err == nil || !strings.Contains(err.Error(), strings.TrimSpace(expected)) {
					t.Fatalf("Run error = %v, want substring %q", err, strings.TrimSpace(expected))
				}
			}
			for _, fileName := range []string{"check.err", "build.err"} {
				if expected, ok := readCorpusFile(t, casePath, fileName); ok {
					expectations++
					if _, err := Compile(entry); err == nil || !strings.Contains(err.Error(), strings.TrimSpace(expected)) {
						t.Fatalf("Compile error = %v, want %s substring %q", err, fileName, strings.TrimSpace(expected))
					}
				}
			}
			if expectations == 0 {
				t.Fatal("corpus case has no expectation file")
			}
		})
	}
}

func corpusCases(t *testing.T, root string) []string {
	t.Helper()
	var cases []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Name() == "main.ts" {
			cases = append(cases, filepath.Dir(path))
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
