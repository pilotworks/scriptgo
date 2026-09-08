package lowering

import (
	"path/filepath"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// collectDynamicImports indexes named imports whose implementation is a local
// JavaScript module. The TypeScript frontend remains responsible for resolving
// the edge; lowering only records the deliberately small Dynamic boundary.
func collectDynamicImports(program frontend.Program) map[string]ir.DynamicModule {
	files := make(map[string]typescriptgo.SourceFile, len(program.Files))
	for _, file := range program.Files {
		files[filepath.Clean(file.FileName)] = file
	}
	result := make(map[string]ir.DynamicModule)
	for _, file := range program.Files {
		for _, reference := range file.Imports {
			if reference.TypeOnly || !isJavaScriptFile(reference.ResolvedFileName) {
				continue
			}
			dependency, ok := files[filepath.Clean(reference.ResolvedFileName)]
			if !ok {
				continue
			}
			for _, binding := range reference.Bindings {
				if binding.TypeOnly {
					continue
				}
				arity := -1
				for _, statement := range dependency.Syntax.Statements {
					if (statement.Kind == "function" || statement.Kind == "async_function") && statement.Name == binding.ImportedName {
						arity = len(statement.Parameters)
						break
					}
				}
				if arity >= 0 {
					result[binding.LocalName] = ir.DynamicModule{Path: filepath.Clean(dependency.FileName), Source: dependency.Source, Export: binding.ImportedName, Arity: arity}
				}
			}
		}
	}
	return result
}

func HasDynamicImports(program frontend.Program) bool {
	return len(collectDynamicImports(program)) != 0
}

func isJavaScriptFile(path string) bool {
	return strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".mjs") || strings.HasSuffix(path, ".cjs")
}

func nativeSourceFiles(files []typescriptgo.SourceFile) []typescriptgo.SourceFile {
	result := make([]typescriptgo.SourceFile, 0, len(files))
	for _, file := range files {
		if !isJavaScriptFile(file.FileName) {
			result = append(result, file)
		}
	}
	return result
}
