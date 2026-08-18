// Package compiler owns the native compilation pipeline.
package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pilotworks/scriptgo/internal/backend/llvm"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/interpreter"
	"github.com/pilotworks/scriptgo/internal/ir"
	"github.com/pilotworks/scriptgo/internal/lowering"
	"github.com/pilotworks/scriptgo/internal/runtime"
)

// Compile reads one TypeScript entry point and returns LLVM IR.
func Compile(entryPath string) (string, error) {
	module, err := CompileModule(entryPath)
	if err != nil {
		return "", err
	}
	return llvm.Emit(module)
}

// CompileModule returns the verified typed IR for an entry point.
func CompileModule(entryPath string) (ir.Module, error) {
	source, err := os.ReadFile(entryPath)
	if err != nil {
		return ir.Module{}, fmt.Errorf("read entry point %q: %w", entryPath, err)
	}

	program, err := frontend.NewProgram(entryPath, string(source))
	if err != nil {
		return ir.Module{}, err
	}

	module, err := lowering.Lower(program)
	if err != nil {
		return ir.Module{}, fmt.Errorf("lower %q: %w", entryPath, err)
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
	clang, err := resolveClang()
	if err != nil {
		return err
	}
	output, err := Compile(entryPath)
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp("", "scriptgo-*.ll")
	if err != nil {
		return fmt.Errorf("create temporary LLVM file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.WriteString(output); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary LLVM file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary LLVM file: %w", err)
	}
	runtimeFile, err := os.CreateTemp("", "scriptgo-runtime-*.c")
	if err != nil {
		return fmt.Errorf("create temporary runtime file: %w", err)
	}
	runtimePath := runtimeFile.Name()
	defer os.Remove(runtimePath)
	if _, err := runtimeFile.Write(runtime.Source); err != nil {
		runtimeFile.Close()
		return fmt.Errorf("write temporary runtime file: %w", err)
	}
	if err := runtimeFile.Close(); err != nil {
		return fmt.Errorf("close temporary runtime file: %w", err)
	}
	command := exec.Command(clang, "-x", "ir", temporaryPath, "-x", "c", runtimePath, "-o", filepath.Clean(outputPath))
	if diagnostic, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("clang: %w: %s", err, diagnostic)
	}
	return nil
}

func resolveClang() (string, error) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		return "", fmt.Errorf("native backend llvm requires clang in PATH: %w", err)
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
