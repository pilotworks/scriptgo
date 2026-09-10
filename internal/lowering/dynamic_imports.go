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
		if canonical, err := filepath.EvalSymlinks(file.FileName); err == nil {
			files[filepath.Clean(canonical)] = file
		}
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
					exportName := binding.ImportedName
					if dynamicCommonJSDefaultExport(dependency.Syntax.Statements) {
						exportName = "default"
					}
					result[binding.LocalName] = ir.DynamicModule{Path: filepath.Clean(dependency.FileName), Source: dependency.Source, Export: exportName, Arity: arity}
				}
			}
		}
	}
	for _, file := range program.Files {
		collectDynamicImportAliases(file.Syntax.Statements, result)
	}
	return result
}

func collectDynamicImportAliases(statements []typescriptgo.SyntaxStatement, modules map[string]ir.DynamicModule) {
	for _, statement := range statements {
		if statement.Kind == "variable" && statement.Name != "" && statement.Expression != nil && statement.Expression.Kind == "identifier" {
			if module, ok := modules[statement.Expression.Text]; ok {
				modules[statement.Name] = module
			}
		}
		collectDynamicImportAliases(statement.Body, modules)
		collectDynamicImportAliases(statement.Then, modules)
		collectDynamicImportAliases(statement.Else, modules)
		collectDynamicImportAliases(statement.Catch, modules)
		collectDynamicImportAliases(statement.Finally, modules)
	}
}

func dynamicExportArity(statements []typescriptgo.SyntaxStatement, exportName string) (int, bool) {
	for _, statement := range statements {
		if arity, ok := dynamicCommonJSExportArity(statement, exportName); ok {
			return arity, true
		}
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

// dynamicCommonJSExportArity recognizes the small CommonJS function-export
// surface supported by the Dynamic runtime. TypeScript-Go has already
// normalized assignment targets into field_set statements for us.
func dynamicCommonJSExportArity(statement typescriptgo.SyntaxStatement, exportName string) (int, bool) {
	if statement.Kind != "field_set" || statement.Expression == nil || statement.Expression.Function == nil {
		return 0, false
	}
	if statement.Left == nil {
		return 0, false
	}
	if statement.Left.Kind == "identifier" && statement.Left.Text == "exports" && statement.Name == exportName {
		return len(statement.Expression.Function.Parameters), true
	}
	if statement.Left.Kind == "identifier" && statement.Left.Text == "module" && statement.Name == "exports" {
		return len(statement.Expression.Function.Parameters), true
	}
	if statement.Left.Kind == "property" && statement.Left.Text == "exports" && statement.Left.Left != nil && statement.Left.Left.Kind == "identifier" && statement.Left.Left.Text == "module" && statement.Name == exportName {
		return len(statement.Expression.Function.Parameters), true
	}
	return 0, false
}

func dynamicCommonJSDefaultExport(statements []typescriptgo.SyntaxStatement) bool {
	for _, statement := range statements {
		if statement.Kind == "field_set" && statement.Left != nil && statement.Left.Kind == "identifier" && statement.Left.Text == "module" && statement.Name == "exports" && statement.Expression != nil && statement.Expression.Function != nil {
			return true
		}
	}
	return false
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
