package frontend

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/scriptgo"
)

// Program is the normalized input passed from the TypeScript frontend to lowering.
type Program struct {
	EntryPath      string
	Source         string
	StatementCount int
	Options        typescriptgo.CompilerOptions
	Diagnostics    []typescriptgo.Diagnostic
	// Files are ordered with local dependencies before their importers.
	Files []typescriptgo.SourceFile
}

// ProgramOptions controls frontend program construction.
type ProgramOptions struct {
	ConfigPath string
}

func NewProgram(entryPath, source string) (Program, error) {
	return NewProgramWithOptions(entryPath, source, ProgramOptions{})
}

// NewProgramWithOptions creates a Program using custom options (e.g. tsconfig path).
func NewProgramWithOptions(entryPath, source string, opts ProgramOptions) (Program, error) {
	if filepath.Ext(entryPath) != ".ts" {
		return Program{}, fmt.Errorf("entry point must have a .ts extension: %s", entryPath)
	}
	if strings.TrimSpace(source) == "" {
		return Program{}, fmt.Errorf("entry point %q is empty", entryPath)
	}

	parsed, err := typescriptgo.CheckWithOptions(entryPath, typescriptgo.CheckOptions{
		ConfigPath: opts.ConfigPath,
	})
	if err != nil {
		return Program{}, fmt.Errorf("check entry point %q: %w", entryPath, err)
	}
	if len(parsed.Diagnostics) > 0 {
		diagnostic := parsed.Diagnostics[0]
		return Program{}, fmt.Errorf("%s", typescriptgo.FormatDiagnostic(diagnostic, source))
	}

	absoluteEntry, err := filepath.Abs(entryPath)
	if err != nil {
		return Program{}, err
	}
	absoluteEntry = filepath.Clean(absoluteEntry)
	statementCount := 0
	for _, file := range parsed.Files {
		if filepath.Clean(file.FileName) == absoluteEntry {
			statementCount = file.StatementCount
			break
		}
	}

	return Program{
		EntryPath:      absoluteEntry,
		Source:         source,
		StatementCount: statementCount,
		Options:        parsed.Options,
		Diagnostics:    parsed.Diagnostics,
		Files:          parsed.Files,
	}, nil
}

// CheckProject typechecks a tsconfig.json project through frontend.
func CheckProject(configPath string) (typescriptgo.ProgramResult, error) {
	return typescriptgo.CheckProject(configPath)
}
