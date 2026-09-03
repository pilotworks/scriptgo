package lowering

import (
	"path/filepath"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
)

// runtimeProgram keeps type-only dependencies out of lowering. Their checked
// declarations remain available to the frontend, but their runtime bodies do
// not belong in the emitted module.
func runtimeProgram(program frontend.Program) frontend.Program {
	if program.EntryPath == "" || len(program.Files) == 0 {
		return program
	}

	files := make(map[string]typescriptgo.SourceFile, len(program.Files))
	for _, file := range program.Files {
		files[filepath.Clean(file.FileName)] = file
	}
	entry := filepath.Clean(program.EntryPath)
	if _, ok := files[entry]; !ok {
		return program
	}

	reachable := map[string]bool{}
	var visit func(string)
	visit = func(fileName string) {
		fileName = filepath.Clean(fileName)
		if reachable[fileName] {
			return
		}
		file, ok := files[fileName]
		if !ok {
			return
		}
		reachable[fileName] = true
		for _, reference := range file.Imports {
			if !reference.TypeOnly {
				visit(reference.ResolvedFileName)
			}
		}
	}
	visit(entry)

	filtered := make([]typescriptgo.SourceFile, 0, len(reachable))
	for _, file := range program.Files {
		if reachable[filepath.Clean(file.FileName)] {
			filtered = append(filtered, file)
		}
	}
	program.Files = filtered
	return program
}
