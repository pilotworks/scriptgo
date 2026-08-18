// Package compiler owns the native compilation pipeline.
package compiler

import (
	"fmt"
	"os"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// Compile reads one TypeScript entry point and returns the current backend output.
// The frontend is backed by TypeScript-Go while the typed IR and backend remain small.
func Compile(entryPath string) (string, error) {
	source, err := os.ReadFile(entryPath)
	if err != nil {
		return "", fmt.Errorf("read entry point %q: %w", entryPath, err)
	}

	program, err := frontend.NewProgram(entryPath, string(source))
	if err != nil {
		return "", err
	}

	module := ir.Module{
		SourcePath:     program.EntryPath,
		StatementCount: program.StatementCount,
	}
	return GenerateStub(module), nil
}

// GenerateStub documents the pipeline boundary while native lowering is being built.
func GenerateStub(module ir.Module) string {
	return fmt.Sprintf("; scriptgo scaffold\n; input: %s\n; statements: %d\n; functions: %d\n; pipeline: typescript -> typed-ir -> llvm\n", module.SourcePath, module.StatementCount, len(module.Functions))
}
