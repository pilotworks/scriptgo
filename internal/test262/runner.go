package test262

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// Outcome classifies one test.
type Outcome string

const (
	// Pass: the test behaved as test262 requires.
	Pass Outcome = "pass"
	// Fail: the program compiled but produced the wrong behavior.
	Fail Outcome = "fail"
	// Unsupported: the compiler rejected the test (native subset gap).
	Unsupported Outcome = "unsupported"
	// Skipped: metadata requires a mode the runner does not model.
	Skipped Outcome = "skipped"
)

// Result is the outcome of one test file.
type Result struct {
	Path    string  `json:"path"`
	Outcome Outcome `json:"outcome"`
	Detail  string  `json:"detail,omitempty"`
}

// Builder compiles a TypeScript entry into a native executable.
type Builder func(entryPath, outputPath string) error

// Config controls a test262 run.
type Config struct {
	// Root is the test262 checkout (containing test/ and harness/).
	Root string
	// Paths are directories or files relative to Root/test.
	Paths []string
	// Parallel is the number of concurrent tests (default 1).
	Parallel int
	// Timeout bounds each test's execution (default 10s).
	Timeout time.Duration
	// Build compiles one test; required.
	Build Builder
	// WorkDir holds per-test build directories (default os.TempDir()).
	WorkDir string
}

// Discover lists the test files selected by cfg, sorted for stable reports.
// Fixture files (_FIXTURE) are not tests and are excluded.
func Discover(cfg Config) ([]string, error) {
	var files []string
	for _, p := range cfg.Paths {
		base := filepath.Join(cfg.Root, "test", filepath.FromSlash(p))
		err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".js") || strings.Contains(d.Name(), "_FIXTURE") {
				return nil
			}
			rel, err := filepath.Rel(filepath.Join(cfg.Root, "test"), path)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(rel))
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	return files, nil
}

// Run executes the selected tests and returns results in discovery order.
func Run(ctx context.Context, cfg Config) ([]Result, error) {
	if cfg.Build == nil {
		return nil, errors.New("test262: Config.Build is required")
	}
	if cfg.Parallel < 1 {
		cfg.Parallel = 1
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	files, err := Discover(cfg)
	if err != nil {
		return nil, err
	}
	results := make([]Result, len(files))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < cfg.Parallel; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i] = runOne(ctx, cfg, files[i])
			}
		}()
	}
	for i := range files {
		select {
		case jobs <- i:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return nil, ctx.Err()
		}
	}
	close(jobs)
	wg.Wait()
	return results, nil
}

func runOne(ctx context.Context, cfg Config, rel string) (result Result) {
	result.Path = rel
	defer func() {
		if r := recover(); r != nil {
			// A crash is a compiler bug, never a subset limitation.
			result.Outcome, result.Detail = Fail, fmt.Sprintf("compiler panic: %v", r)
		}
	}()
	source, err := os.ReadFile(filepath.Join(cfg.Root, "test", filepath.FromSlash(rel)))
	if err != nil {
		return Result{Path: rel, Outcome: Fail, Detail: err.Error()}
	}
	meta, err := ParseMetadata(string(source))
	if err != nil {
		return Result{Path: rel, Outcome: Skipped, Detail: err.Error()}
	}
	entry, skip := Prepare(string(source), meta)
	if skip != "" {
		return Result{Path: rel, Outcome: Skipped, Detail: skip}
	}
	dir, err := os.MkdirTemp(cfg.WorkDir, "test262-")
	if err != nil {
		return Result{Path: rel, Outcome: Fail, Detail: err.Error()}
	}
	defer os.RemoveAll(dir)
	entryPath := filepath.Join(dir, "main.ts")
	if err := os.WriteFile(entryPath, []byte(entry), 0o644); err != nil {
		return Result{Path: rel, Outcome: Fail, Detail: err.Error()}
	}
	binary := filepath.Join(dir, "main")
	buildErr := cfg.Build(entryPath, binary)
	negative := meta.Negative
	if negative != nil && (negative.Phase == "parse" || negative.Phase == "early") {
		if buildErr != nil && isSyntaxDiagnostic(buildErr) {
			return Result{Path: rel, Outcome: Pass}
		}
		if buildErr != nil && isToolchainFailure(buildErr) {
			return Result{Path: rel, Outcome: Fail, Detail: "invalid native code: " + toolchainError(buildErr)}
		}
		if buildErr != nil {
			return Result{Path: rel, Outcome: Unsupported, Detail: firstLine(buildErr.Error())}
		}
		return Result{Path: rel, Outcome: Fail, Detail: "expected " + negative.Type + " at " + negative.Phase + " phase, compiled successfully"}
	}
	if buildErr != nil {
		if isToolchainFailure(buildErr) {
			// The compiler accepted the program but emitted code the native
			// toolchain rejects: a compiler bug, not a subset limitation.
			return Result{Path: rel, Outcome: Fail, Detail: "invalid native code: " + toolchainError(buildErr)}
		}
		return Result{Path: rel, Outcome: Unsupported, Detail: firstLine(buildErr.Error())}
	}
	runCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	cmd := exec.CommandContext(runCtx, binary)
	cmd.Dir = dir
	// Do not wait on pipes held open by processes the test spawned.
	cmd.WaitDelay = time.Second
	var stderr bytes.Buffer
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if runCtx.Err() == context.DeadlineExceeded {
		return Result{Path: rel, Outcome: Fail, Detail: "timeout"}
	}
	if negative != nil && negative.Phase == "runtime" {
		// The runtime reports an uncaught Error instance as
		// "Uncaught exception: <name>: <message>", so the thrown type must
		// match; a crash or a different error is a failure.
		if runErr == nil {
			return Result{Path: rel, Outcome: Fail, Detail: "expected " + negative.Type + " at runtime, completed normally"}
		}
		if uncaughtErrorName(stderr.String()) == negative.Type {
			return Result{Path: rel, Outcome: Pass}
		}
		detail := firstLine(stderr.String())
		if detail == "" {
			detail = runErr.Error()
		}
		return Result{Path: rel, Outcome: Fail, Detail: "expected " + negative.Type + " at runtime, got: " + detail}
	}
	if runErr != nil {
		detail := firstLine(stderr.String())
		if detail == "" {
			detail = runErr.Error()
		}
		return Result{Path: rel, Outcome: Fail, Detail: detail}
	}
	return Result{Path: rel, Outcome: Pass}
}

// syntaxDiagnostic matches TypeScript-Go diagnostics that implement
// ECMAScript early errors: the syntax/grammar range TS1xxx, plus checker codes
// that are spec early errors rather than type errors — invalid assignment
// target (2364), rest element not last (2462), misplaced super() (2337), and
// a private name outside a class body (18016).
var syntaxDiagnostic = regexp.MustCompile(`error TS(1\d{3}|2364|2462|2337|18016):`)

// isSyntaxDiagnostic reports a TypeScript-Go diagnostic for an early error,
// the compile-time equivalent of an early or parse-phase SyntaxError.
func isSyntaxDiagnostic(err error) bool {
	return syntaxDiagnostic.MatchString(err.Error())
}

// uncaughtError matches the runtime's report of an uncaught Error instance.
var uncaughtError = regexp.MustCompile(`(?m)^Uncaught exception: ([A-Za-z_$][\w$]*)(?::|$)`)

// uncaughtErrorName returns the error name of an uncaught exception report,
// or "" when the program did not die from an uncaught Error.
func uncaughtErrorName(stderr string) string {
	if match := uncaughtError.FindStringSubmatch(stderr); match != nil {
		return match[1]
	}
	return ""
}

var toolchainLocation = regexp.MustCompile(`^\S*module\.ll:\d+:\d+: `)

// toolchainError returns the first diagnostic Clang reported, without the
// temporary file location, so identical backend defects group together.
func toolchainError(err error) string {
	for _, line := range strings.Split(err.Error(), "\n") {
		if strings.Contains(line, "error:") {
			return firstLine(toolchainLocation.ReplaceAllString(strings.TrimSpace(line), ""))
		}
	}
	return firstLine(err.Error())
}

// isToolchainFailure reports a build error raised by Clang after scriptgo
// accepted and lowered the program.
func isToolchainFailure(err error) bool {
	return strings.Contains(err.Error(), "clang: exit status")
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 200 {
		s = s[:200] + "…"
	}
	return s
}
