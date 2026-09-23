package typescriptgo

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/compiler"
)

// webGlobalModules maps standard Web/Node global constructors to their stdlib builtin module names.
var webGlobalModules = map[string]string{
	"FormData":        "formdata",
	"URLSearchParams": "url",
	"URL":             "url",
	"Headers":         "http",
	"Response":        "http",
	"Request":         "http",
	"Blob":            "buffer",
	"File":            "buffer",
}

func moduleReferences(program *compiler.Program, file *ast.SourceFile, cwd string) []ModuleReference {
	result := make([]ModuleReference, 0, len(file.Imports()))
	for _, specifier := range file.Imports() {
		localName := namespaceImportName(specifier.AsNode())
		typeOnly := importIsTypeOnly(specifier.AsNode())
		if module, ok := builtinModule(specifier.Text()); ok {
			resolved := program.GetResolvedModuleFromModuleSpecifier(file, specifier)
			resolvedFileName := ""
			if module.Name == "stream_consumers" || module.Name == "stream/consumers" {
				resolvedFileName = filepath.Join(cwd, "node_modules", "stream", "consumers", "index.ts")
			} else if module.Name == "stream_promises" || module.Name == "stream/promises" {
				resolvedFileName = filepath.Join(cwd, "node_modules", "stream", "promises", "index.ts")
			} else if module.Name == "webstreams" || module.Name == "stream/web" {
				resolvedFileName = filepath.Join(cwd, "node_modules", "stream", "web", "index.ts")
			} else if module.Name == "readline_promises" || module.Name == "readline/promises" {
				resolvedFileName = filepath.Join(cwd, "node_modules", "readline", "promises", "index.ts")
			} else if resolved != nil && resolved.ResolvedFileName != "" && !strings.HasSuffix(resolved.ResolvedFileName, ".d.ts") {
				resolvedFileName = filepath.Clean(resolved.ResolvedFileName)
			} else {
				resolvedFileName = filepath.Clean(filepath.Join(cwd, "node_modules", filepath.FromSlash(module.Name), "index.ts"))
			}
			result = append(result, ModuleReference{Specifier: specifier.Text(), ResolvedFileName: resolvedFileName, LocalName: localName, Bindings: importBindings(specifier.AsNode()), Span: sourceSpan(specifier), Builtin: true, TypeOnly: typeOnly})
			continue
		}
		resolved := program.GetResolvedModuleFromModuleSpecifier(file, specifier)
		if resolved == nil || resolved.ResolvedFileName == "" {
			continue
		}
		// TypeScript-Go marks node_modules files as external library imports. Keep
		// resolved JavaScript entry points in the graph so Dynamic lowering can
		// bundle them; declaration-only package edges remain type-checking data.
		if resolved.IsExternalLibraryImport && !isJavaScriptSource(resolved.ResolvedFileName) {
			continue
		}
		result = append(result, ModuleReference{
			Specifier:        specifier.Text(),
			ResolvedFileName: filepath.Clean(resolved.ResolvedFileName),
			LocalName:        localName,
			Bindings:         importBindings(specifier.AsNode()),
			Span:             sourceSpan(specifier),
			TypeOnly:         typeOnly,
		})
	}
	return result
}

func importIsTypeOnly(specifier *ast.Node) bool {
	if specifier == nil || specifier.Parent == nil {
		return false
	}
	if specifier.Parent.Kind == ast.KindExportDeclaration {
		declaration := specifier.Parent.AsExportDeclaration()
		return declaration != nil && declaration.IsTypeOnly
	}
	if specifier.Parent.Kind != ast.KindImportDeclaration {
		return false
	}
	declaration := specifier.Parent.AsImportDeclaration()
	if declaration == nil || declaration.ImportClause == nil {
		return false
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause == nil || clause.IsTypeOnly() {
		return clause != nil && clause.IsTypeOnly()
	}
	if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
		return false
	}
	named := clause.NamedBindings.AsNamedImports()
	if named == nil || named.Elements == nil || len(named.Elements.Nodes) == 0 {
		return false
	}
	for _, element := range named.Elements.Nodes {
		if !element.IsTypeOnly() {
			return false
		}
	}
	return true
}

func importBindings(specifier *ast.Node) []ModuleBinding {
	if specifier == nil || specifier.Parent == nil {
		return nil
	}
	if specifier.Parent.Kind == ast.KindExportDeclaration {
		declaration := specifier.Parent.AsExportDeclaration()
		if declaration == nil || declaration.ExportClause == nil || declaration.ExportClause.Kind != ast.KindNamedExports {
			return nil
		}
		named := declaration.ExportClause.AsNamedExports()
		if named == nil || named.Elements == nil {
			return nil
		}
		bindings := make([]ModuleBinding, 0, len(named.Elements.Nodes))
		for _, element := range named.Elements.Nodes {
			exportSpecifier := element.AsExportSpecifier()
			if exportSpecifier == nil || exportSpecifier.Name() == nil {
				continue
			}
			exportedName := exportSpecifier.Name().Text()
			importedName := exportedName
			if exportSpecifier.PropertyName != nil {
				importedName = exportSpecifier.PropertyName.Text()
			}
			bindings = append(bindings, ModuleBinding{
				ImportedName: importedName,
				LocalName:    exportedName,
				TypeOnly:     declaration.IsTypeOnly || exportSpecifier.IsTypeOnly,
			})
		}
		return bindings
	}
	if specifier.Parent.Kind != ast.KindImportDeclaration {
		return nil
	}
	declaration := specifier.Parent.AsImportDeclaration()
	if declaration == nil || declaration.ImportClause == nil {
		return nil
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause == nil {
		return nil
	}
	var bindings []ModuleBinding
	if clause.Name() != nil {
		bindings = append(bindings, ModuleBinding{
			ImportedName: "default",
			LocalName:    clause.Name().Text(),
			TypeOnly:     clause.IsTypeOnly(),
		})
	}
	if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
		return bindings
	}
	named := clause.NamedBindings.AsNamedImports()
	if named == nil || named.Elements == nil {
		return bindings
	}
	for _, element := range named.Elements.Nodes {
		importSpecifier := element.AsImportSpecifier()
		if importSpecifier == nil || importSpecifier.Name() == nil {
			continue
		}
		importedName := importSpecifier.Name().Text()
		if importSpecifier.PropertyName != nil {
			importedName = importSpecifier.PropertyName.Text()
		}
		bindings = append(bindings, ModuleBinding{
			ImportedName: importedName,
			LocalName:    importSpecifier.Name().Text(),
			TypeOnly:     clause.IsTypeOnly() || importSpecifier.IsTypeOnly,
		})
	}
	return bindings
}

func namespaceImportName(specifier *ast.Node) string {
	if specifier == nil || specifier.Parent == nil || specifier.Parent.Kind != ast.KindImportDeclaration {
		return ""
	}
	clause := specifier.Parent.ImportClause()
	if clause == nil || clause.AsImportClause().NamedBindings == nil || !ast.IsNamespaceImport(clause.AsImportClause().NamedBindings) {
		return ""
	}
	return clause.AsImportClause().NamedBindings.AsNamespaceImport().Name().Text()
}

// orderedSourceFiles returns reachable local files in dependency-first order.
func orderedSourceFiles(files map[string]*ast.SourceFile, entry string, program *compiler.Program, builtinPaths map[string]string) []*ast.SourceFile {
	cwd := filepath.Dir(entry)
	ordered := make([]*ast.SourceFile, 0, len(files))
	visited := make(map[string]bool, len(files))

	modulePaths := make(map[string]string, len(builtinPaths))
	for path, name := range builtinPaths {
		if !strings.HasSuffix(path, ".d.ts") {
			modulePaths[name] = path
		}
	}

	var visit func(string)
	visit = func(fileName string) {
		fileName = filepath.Clean(fileName)
		if visited[fileName] {
			return
		}
		file, ok := files[fileName]
		if !ok {
			if canonical, err := filepath.EvalSymlinks(fileName); err == nil {
				canonical = filepath.Clean(canonical)
				file, ok = files[canonical]
			}
		}
		if !ok {
			return
		}
		visited[fileName] = true
		for _, reference := range moduleReferences(program, file, cwd) {
			visit(reference.ResolvedFileName)
		}

		// Auto-include standard Web globals referenced in source files when not locally declared
		fileBuiltin := builtinPaths[fileName]
		for globalName, modName := range webGlobalModules {
			if fileBuiltin != modName && file.HasIdentifier(globalName) {
				if file.Locals == nil || file.Locals[globalName] == nil {
					if modPath, exists := modulePaths[modName]; exists {
						visit(modPath)
					}
				}
			}
		}

		ordered = append(ordered, file)
	}
	visit(entry)

	// Keep the adapter deterministic even if the compiler returns an unexpected
	// disconnected source file in a future TypeScript-Go revision.
	remaining := make([]string, 0, len(files)-len(ordered))
	for fileName := range files {
		if !visited[fileName] && builtinPaths[fileName] == "" {
			remaining = append(remaining, fileName)
		}
	}
	sort.Strings(remaining)
	for _, fileName := range remaining {
		ordered = append(ordered, files[fileName])
	}
	return ordered
}
