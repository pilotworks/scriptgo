package frontend

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/microsoft/typescript-go/scriptgo"
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

func NewProgram(entryPath, source string) (Program, error) {
	if filepath.Ext(entryPath) != ".ts" {
		return Program{}, fmt.Errorf("entry point must have a .ts extension: %s", entryPath)
	}
	if strings.TrimSpace(source) == "" {
		return Program{}, fmt.Errorf("entry point %q is empty", entryPath)
	}

	parsed, err := typescriptgo.Check(entryPath)
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
