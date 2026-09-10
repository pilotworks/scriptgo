package lowering

import (
	"path/filepath"
	"sort"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

type dynamicImportBinding struct {
	Path   string
	Export string
	Arity  int
}

// collectDynamicImports builds the closed JavaScript module graph and indexes
// bindings used by native TypeScript callers. TypeScript-Go remains the sole
// owner of resolving every graph edge.
func collectDynamicImports(program frontend.Program) (map[string]dynamicImportBinding, []ir.DynamicModule) {
	files := make(map[string]typescriptgo.SourceFile, len(program.Files))
	for _, file := range program.Files {
		files[filepath.Clean(file.FileName)] = file
		if canonical, err := filepath.EvalSymlinks(file.FileName); err == nil {
			files[filepath.Clean(canonical)] = file
		}
	}
	bindings := make(map[string]dynamicImportBinding)
	exportsByPath := make(map[string]map[string]bool)
	for _, file := range program.Files {
		for _, reference := range file.Imports {
			if reference.TypeOnly || !isJavaScriptFile(reference.ResolvedFileName) {
				continue
			}
			dependency, ok := files[filepath.Clean(reference.ResolvedFileName)]
			if !ok {
				continue
			}
			dependencyPath := filepath.Clean(dependency.FileName)
			if exportsByPath[dependencyPath] == nil {
				exportsByPath[dependencyPath] = make(map[string]bool)
			}
			for _, binding := range reference.Bindings {
				if binding.TypeOnly {
					continue
				}
				exportsByPath[dependencyPath][binding.ImportedName] = true
				if isJavaScriptFile(file.FileName) {
					continue
				}
				if arity, ok := dynamicResolvedExportArity(files, dependency, binding.ImportedName, make(map[string]bool)); ok {
					exportName := binding.ImportedName
					if dynamicCommonJSDefaultExport(dependency.Syntax.Statements) {
						exportName = "default"
					}
					bindings[binding.LocalName] = dynamicImportBinding{Path: dependencyPath, Export: exportName, Arity: arity}
				}
			}
		}
	}
	for _, file := range program.Files {
		if !isJavaScriptFile(file.FileName) {
			collectDynamicImportAliases(file.Syntax.Statements, bindings)
		}
	}

	modules := make([]ir.DynamicModule, 0)
	for _, file := range program.Files {
		if !isJavaScriptFile(file.FileName) {
			continue
		}
		path := filepath.Clean(file.FileName)
		module := ir.DynamicModule{Path: path, Source: file.Source, Kind: dynamicModuleKind(file)}
		for name := range exportsByPath[path] {
			module.Exports = append(module.Exports, name)
		}
		sort.Strings(module.Exports)
		for _, reference := range file.Imports {
			if reference.TypeOnly || !isJavaScriptFile(reference.ResolvedFileName) {
				continue
			}
			module.Imports = append(module.Imports, ir.DynamicImport{
				Specifier: reference.Specifier,
				Path:      filepath.Clean(reference.ResolvedFileName),
			})
		}
		sort.Slice(module.Imports, func(i, j int) bool {
			if module.Imports[i].Specifier == module.Imports[j].Specifier {
				return module.Imports[i].Path < module.Imports[j].Path
			}
			return module.Imports[i].Specifier < module.Imports[j].Specifier
		})
		modules = append(modules, module)
	}
	sort.Slice(modules, func(i, j int) bool { return modules[i].Path < modules[j].Path })
	return bindings, modules
}

func dynamicResolvedExportArity(files map[string]typescriptgo.SourceFile, file typescriptgo.SourceFile, exportName string, seen map[string]bool) (int, bool) {
	path := filepath.Clean(file.FileName)
	key := path + "#" + exportName
	if seen[key] {
		return 0, false
	}
	seen[key] = true
	if arity, ok := dynamicExportArity(file.Syntax.Statements, exportName); ok {
		return arity, true
	}
	for _, reference := range file.Imports {
		if reference.TypeOnly || !isJavaScriptFile(reference.ResolvedFileName) {
			continue
		}
		for _, binding := range reference.Bindings {
			if binding.TypeOnly || binding.LocalName != exportName {
				continue
			}
			dependency, ok := files[filepath.Clean(reference.ResolvedFileName)]
			if !ok {
				continue
			}
			if arity, ok := dynamicResolvedExportArity(files, dependency, binding.ImportedName, seen); ok {
				return arity, true
			}
		}
	}
	return 0, false
}

func collectDynamicImportAliases(statements []typescriptgo.SyntaxStatement, modules map[string]dynamicImportBinding) {
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

func dynamicModuleKind(file typescriptgo.SourceFile) string {
	switch strings.ToLower(filepath.Ext(file.FileName)) {
	case ".cjs":
		return "commonjs"
	case ".mjs":
		return "esm"
	}
	for _, statement := range file.Syntax.Statements {
		if _, ok := dynamicCommonJSExportArity(statement, ""); ok || isCommonJSExport(statement) {
			return "commonjs"
		}
	}
	return "esm"
}

func isCommonJSExport(statement typescriptgo.SyntaxStatement) bool {
	if statement.Kind != "field_set" || statement.Left == nil {
		return false
	}
	if statement.Left.Kind == "identifier" {
		return statement.Left.Text == "exports" || (statement.Left.Text == "module" && statement.Name == "exports")
	}
	return statement.Left.Kind == "property" && statement.Left.Text == "exports" && statement.Left.Left != nil && statement.Left.Left.Kind == "identifier" && statement.Left.Left.Text == "module"
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
	bindings, _ := collectDynamicImports(program)
	return len(bindings) != 0
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
