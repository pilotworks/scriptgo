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
}

func NewProgram(entryPath, source string) (Program, error) {
	if filepath.Ext(entryPath) != ".ts" {
		return Program{}, fmt.Errorf("entry point must have a .ts extension: %s", entryPath)
	}
	if strings.TrimSpace(source) == "" {
		return Program{}, fmt.Errorf("entry point %q is empty", entryPath)
	}

	parsed, err := typescriptgo.Parse(entryPath, source)
	if err != nil {
		return Program{}, fmt.Errorf("parse entry point %q: %w", entryPath, err)
	}
	if len(parsed.Diagnostics) > 0 {
		diagnostic := parsed.Diagnostics[0]
		return Program{}, fmt.Errorf(
			"TypeScript syntax error in %s at offset %d (TS%d): %s",
			entryPath,
			diagnostic.Start,
			diagnostic.Code,
			diagnostic.Message,
		)
	}

	return Program{
		EntryPath:      entryPath,
		Source:         source,
		StatementCount: parsed.StatementCount,
	}, nil
}
