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
				if arity, ok := dynamicExportArity(dependency.Syntax.Statements, binding.ImportedName); ok {
					result[binding.LocalName] = ir.DynamicModule{Path: filepath.Clean(dependency.FileName), Source: dependency.Source, Export: binding.ImportedName, Arity: arity}
				}
			}
		}
	}
	return result
}

func dynamicExportArity(statements []typescriptgo.SyntaxStatement, exportName string) (int, bool) {
	for _, statement := range statements {
		if exportName == "default" && statement.DefaultExport {
			// Default function declarations are represented by their source name,
			// while the module boundary addresses them as "default".
		} else if statement.Name != exportName {
			continue
		}
		switch statement.Kind {
		case "function", "async_function":
			return len(statement.Parameters), true
		case "variable":
			if statement.Expression != nil && statement.Expression.Function != nil {
				return len(statement.Expression.Function.Parameters), true
			}
		}
	}
	return 0, false
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
