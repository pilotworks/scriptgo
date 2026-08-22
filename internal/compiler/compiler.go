// Package compiler owns the native compilation pipeline.
package compiler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pilotworks/scriptgo/internal/backend/llvm"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/interpreter"
	"github.com/pilotworks/scriptgo/internal/ir"
	"github.com/pilotworks/scriptgo/internal/lowering"
	"github.com/pilotworks/scriptgo/internal/runtime"
)

// Compile reads one TypeScript entry point and returns LLVM IR.
func Compile(entryPath string) (string, error) {
	return compileWithOptions(entryPath, BuildOptions{})
}

// CompileWithOptions emits LLVM IR with the requested target metadata.
func CompileWithOptions(entryPath string, options BuildOptions) (string, error) {
	return compileWithOptions(entryPath, options)
}

func compileWithOptions(entryPath string, options BuildOptions) (string, error) {
	options = options.normalized()
	if err := validateOptions(options); err != nil {
		return "", err
	}
	module, err := CompileModuleWithOptions(entryPath, options)
	if err != nil {
		return "", err
	}
	if options.StrictCasts {
		warns := lowering.GetWarnings()
		if len(warns) > 0 {
			return "", fmt.Errorf("strict casts: %s: %s at offset %d: %s", warns[0].Code, warns[0].FileName, warns[0].Span.Start, warns[0].Message)
		}
	}
	hash := hashSources(module.SourceFiles)
	return llvm.EmitWithOptions(module, llvm.Options{
		CompilerVersion: Version,
		RuntimeABI:      RuntimeABIVersion,
		Target:          options.Target,
		SourceHash:      hash,
		Debug:           options.Debug,
	})
}

// GetWarnings returns any compiler warnings from the most recent lowering run.
func GetWarnings() []lowering.Warning {
	return lowering.GetWarnings()
}

func validateOptions(options BuildOptions) error {
	for _, sanitizer := range options.Sanitizers {
		switch sanitizer {
		case "address", "undefined", "leak":
		default:
			return fmt.Errorf("unsupported sanitizer %q; supported sanitizers: address, undefined, leak", sanitizer)
		}
	}
	return nil
}

func hashSources(sources map[string]string) string {
	paths := make([]string, 0, len(sources))
	for path := range sources {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	hasher := sha256.New()
	for _, path := range paths {
		hasher.Write([]byte{0})
		hasher.Write([]byte(sources[path]))
		hasher.Write([]byte{0})
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

// CompileModule returns the verified typed IR for an entry point.
func CompileModule(entryPath string) (ir.Module, error) {
	return CompileModuleWithOptions(entryPath, BuildOptions{})
}

// CompileModuleWithOptions returns the verified typed IR for an entry point using custom build options.
func CompileModuleWithOptions(entryPath string, options BuildOptions) (ir.Module, error) {
	source, err := os.ReadFile(entryPath)
	if err != nil {
		return ir.Module{}, fmt.Errorf("read entry point %q: %w", entryPath, err)
	}

	program, err := frontend.NewProgram(entryPath, string(source))
	if err != nil {
		return ir.Module{}, err
	}

	module, err := lowering.LowerWithOptions(program, lowering.Options{
		WarnRuntimeCasts: options.WarnRuntimeCasts,
	})
	if err != nil {
		return ir.Module{}, err
	}
	return module, nil
}

// DumpIR returns the stable backend-independent IR artifact for an entry point.
func DumpIR(entryPath string) (string, error) {
	module, err := CompileModule(entryPath)
	if err != nil {
		return "", err
	}
	return module.Dump()
}

// Build compiles the generated LLVM IR into a native executable with clang.
func Build(entryPath, outputPath string) error {
	return BuildWithOptions(entryPath, outputPath, BuildOptions{})
}

// BuildWithOptions compiles a native executable with deterministic inputs.
func BuildWithOptions(entryPath, outputPath string, options BuildOptions) error {
	options = options.normalized()
	ccParts, err := resolveCC(options.CC)
	if err != nil {
		return err
	}
	output, err := compileWithOptions(entryPath, options)
	if err != nil {
		return err
	}
	temporaryDir, err := os.MkdirTemp("", "scriptgo-build-")
	if err != nil {
		return fmt.Errorf("create temporary build directory: %w", err)
	}
	defer os.RemoveAll(temporaryDir)
	temporaryPath := filepath.Join(temporaryDir, "module.ll")
	if err := os.WriteFile(temporaryPath, []byte(output), 0o644); err != nil {
		return fmt.Errorf("write temporary LLVM file: %w", err)
	}
	runtimePath := filepath.Join(temporaryDir, "runtime.c")
	runtimeSource := append([]byte("#line 1 \"scriptgo-runtime.c\"\n"), runtime.Source...)
	if err := os.WriteFile(runtimePath, runtimeSource, 0o644); err != nil {
		return fmt.Errorf("write temporary runtime file: %w", err)
	}
	args := []string{"-x", "ir", temporaryPath, "-x", "c", runtimePath, "-x", "none"}
	if options.Debug {
		args = append(args, "-O0")
	} else {
		args = append(args, "-O2")
	}
	if options.Target != "native" {
		args = append(args, "--target="+options.Target)
	}
	if options.Debug {
		args = append(args, "-g", "-fdebug-compilation-dir=.")
	}
	if len(options.Sanitizers) > 0 {
		args = append(args, "-fsanitize="+strings.Join(options.Sanitizers, ","))
	}
	args = append(args, "-o", filepath.Clean(outputPath), "-lm")
	cmdArgs := append(ccParts[1:], args...)
	command := exec.Command(ccParts[0], cmdArgs...)
	if diagnostic, err := command.CombinedOutput(); err != nil {
		driverName := filepath.Base(ccParts[0])
		return fmt.Errorf("%s: %w: %s", driverName, err, diagnostic)
	}
	return nil
}

func resolveCC(cc string) ([]string, error) {
	return resolveCCWithLookup(cc, exec.LookPath)
}

func resolveCCWithLookup(cc string, lookPath func(string) (string, error)) ([]string, error) {
	cc = strings.TrimSpace(cc)
	if cc == "" || cc == "clang" {
		if bin, err := lookPath("clang"); err == nil {
			return []string{bin}, nil
		}
		if bin, err := lookPath("zig"); err == nil {
			return []string{bin, "cc"}, nil
		}
		return nil, fmt.Errorf("native backend requires \"clang\" or \"zig\" in PATH")
	}

	parts := strings.Fields(cc)
	if len(parts) == 0 {
		return resolveCCWithLookup("", lookPath)
	}
	if parts[0] == "zigcc" {
		if bin, err := lookPath("zigcc"); err == nil {
			parts[0] = bin
			return parts, nil
		}
		if bin, err := lookPath("zig"); err == nil {
			return append([]string{bin, "cc"}, parts[1:]...), nil
		}
	}
	bin, err := lookPath(parts[0])
	if err != nil {
		return nil, fmt.Errorf("native backend requires %q in PATH: %w", parts[0], err)
	}
	parts[0] = bin
	return parts, nil
}

func resolveClang() (string, error) {
	return resolveClangWithLookup(exec.LookPath)
}

func resolveClangWithLookup(lookPath func(string) (string, error)) (string, error) {
	clang, err := lookPath("clang")
	if err != nil {
		return "", fmt.Errorf("native backend llvm requires \"clang\" in PATH: %w", err)
	}
	return clang, nil
}

// Run executes the verified IR with the reference interpreter.
func Run(entryPath string) (string, error) {
	module, err := CompileModule(entryPath)
	if err != nil {
		return "", err
	}
	result, err := interpreter.Execute(module)
	if err != nil {
		return "", fmt.Errorf("interpreter: %w", err)
	}
	return result.Output, nil
}

// Check parses, type checks, and validates the native subset for an entry point.
func Check(entryPath string) error {
	source, err := os.ReadFile(entryPath)
	if err != nil {
		return fmt.Errorf("read entry point %q: %w", entryPath, err)
	}
	program, err := frontend.NewProgram(entryPath, string(source))
	if err != nil {
		return err
	}
	return lowering.ValidateSubset(program)
}
