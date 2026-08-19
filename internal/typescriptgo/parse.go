package typescriptgo

import (
	"context"
	"path/filepath"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/bundled"
	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/parser"
	"github.com/microsoft/typescript-go/internal/tsoptions"
	"github.com/microsoft/typescript-go/internal/tspath"
	"github.com/microsoft/typescript-go/internal/vfs/osvfs"
	"github.com/microsoft/typescript-go/internal/vfs/wrapvfs"
)

// Parse parses one TypeScript source file with TypeScript-Go.
func Parse(fileName, source string) (ParseResult, error) {
	absoluteName, err := filepath.Abs(fileName)
	if err != nil {
		return ParseResult{}, err
	}
	absoluteName = filepath.Clean(absoluteName)

	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: absoluteName,
		Path:     tspath.ToPath(absoluteName, "", true),
	}, source, core.ScriptKindTS)

	result := ParseResult{}
	if file.Statements != nil {
		result.StatementCount = len(file.Statements.Nodes)
	}
	for _, diagnostic := range file.Diagnostics() {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Message: diagnostic.String(),
			Start:   diagnostic.Pos(),
			Length:  diagnostic.Len(),
			Code:    diagnostic.Code(),
		})
	}
	return result, nil
}

// Check creates a TypeScript-Go program. Program creation performs local module
// resolution and binds/checks the complete reachable source graph.
func Check(entryPath string) (ProgramResult, error) {
	absoluteEntry, err := filepath.Abs(entryPath)
	if err != nil {
		return ProgramResult{}, err
	}
	absoluteEntry = filepath.Clean(absoluteEntry)
	cwd := filepath.Dir(absoluteEntry)
	baseFS := bundled.WrapFS(osvfs.FS())
	fs := baseFS
	virtualFiles := map[string]string{}
	builtinPaths := map[string]string{}
	for name, module := range builtinModules {
		virtualPath := filepath.Join(cwd, "node_modules", name, "index.ts")
		virtualFiles[virtualPath] = module.Source
		builtinPaths[virtualPath] = name
	}
	fs = wrapvfs.Wrap(fs, wrapvfs.Replacements{
		FileExists: func(path string) bool {
			if _, ok := virtualFiles[filepath.Clean(path)]; ok {
				return true
			}
			return baseFS.FileExists(path)
		},
		ReadFile: func(path string) (string, bool) {
			if contents, ok := virtualFiles[filepath.Clean(path)]; ok {
				return contents, true
			}
			contents, ok := baseFS.ReadFile(path)
			if ok && strings.HasPrefix(filepath.Clean(path), filepath.Clean(bundled.LibPath())) && (strings.HasSuffix(path, "lib.es5.d.ts") || strings.HasSuffix(path, "lib.d.ts")) {
				return contents + "\n" + globalsSource, true
			}
			return contents, ok
		},
		DirectoryExists: func(path string) bool {
			clean := filepath.Clean(path)
			for virtualPath := range virtualFiles {
				if clean == filepath.Dir(virtualPath) || clean == filepath.Dir(filepath.Dir(virtualPath)) {
					return true
				}
			}
			if clean == cwd {
				return true
			}
			return baseFS.DirectoryExists(path)
		},
	})
	options := &core.CompilerOptions{
		Target:           core.ScriptTargetES2020,
		Module:           core.ModuleKindESNext,
		ModuleResolution: core.ModuleResolutionKindBundler,
		Strict:           core.TSTrue,
		NoEmit:           core.TSTrue,
	}
	comparePaths := tspath.ComparePathsOptions{
		CurrentDirectory:          cwd,
		UseCaseSensitiveFileNames: fs.UseCaseSensitiveFileNames(),
	}
	config := tsoptions.NewParsedCommandLine(options, []string{absoluteEntry}, comparePaths)
	host := compiler.NewCompilerHost(cwd, fs, bundled.LibPath(), nil, nil)
	program := compiler.NewProgram(compiler.ProgramOptions{
		Config:         config,
		Host:           host,
		SingleThreaded: core.TSTrue,
	})

	result := ProgramResult{}
	result.Options = CompilerOptions{
		Target:           "ES2020",
		Module:           "ESNext",
		ModuleResolution: "Bundler",
		Strict:           true,
	}
	ctx := context.Background()
	files := make(map[string]*ast.SourceFile)
	for _, file := range program.GetSourceFiles() {
		if program.IsSourceFileDefaultLibrary(file.Path()) || !isTypeScriptSource(file.FileName()) {
			continue
		}
		files[filepath.Clean(file.FileName())] = file
	}
	for _, file := range orderedSourceFiles(files, absoluteEntry, program) {
		if program.IsSourceFileDefaultLibrary(file.Path()) || !isTypeScriptSource(file.FileName()) {
			continue
		}
		result.Files = append(result.Files, SourceFile{
			FileName:       file.FileName(),
			Source:         file.Text(),
			StatementCount: statementCount(file),
			Span:           sourceSpan(&file.Node),
			Imports:        moduleReferences(program, file),
			Builtin:        builtinPaths[filepath.Clean(file.FileName())] != "",
			BuiltinName:    builtinPaths[filepath.Clean(file.FileName())],
			Symbols:        fileSymbols(program, ctx, file),
			Syntax:         syntaxFile(file),
		})
		result.Diagnostics = append(result.Diagnostics, convertDiagnostics("syntax", program.GetSyntacticDiagnostics(ctx, file))...)
		result.Diagnostics = append(result.Diagnostics, convertDiagnostics("type", program.GetSemanticDiagnostics(ctx, file))...)
	}
	result.Diagnostics = append(result.Diagnostics, convertDiagnostics("program", program.GetProgramDiagnostics())...)
	return result, nil
}

func fileSymbols(program *compiler.Program, ctx context.Context, file *ast.SourceFile) []Symbol {
	checkerInstance, done := program.GetTypeCheckerForFile(ctx, file)
	defer done()
	names := make([]string, 0, len(file.Locals))
	for name := range file.Locals {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]Symbol, 0, len(names))
	for _, name := range names {
		symbol := file.Locals[name]
		if symbol == nil || symbol.Name == "" {
			continue
		}
		declaration := symbol.ValueDeclaration
		if declaration == nil && len(symbol.Declarations) > 0 {
			declaration = symbol.Declarations[0]
		}
		if declaration == nil {
			continue
		}
		if declaration.Name() != nil {
			if namedSymbol := checkerInstance.GetSymbolAtLocation(declaration.Name()); namedSymbol != nil {
				symbol = namedSymbol
			}
		}
		typ := checkerInstance.GetTypeOfSymbol(symbol)
		if declaration.Name() != nil {
			typ = checkerInstance.GetTypeAtLocation(declaration.Name())
		}
		exported := false
		for parent := declaration; parent != nil; parent = parent.Parent {
			if ast.HasSyntacticModifier(parent, ast.ModifierFlagsExport) {
				exported = true
				break
			}
		}
		result = append(result, Symbol{
			Name:     symbol.Name,
			Kind:     symbolKind(symbol.Flags),
			Type:     checkerInstance.TypeToString(typ),
			Span:     sourceSpan(declaration),
			Exported: exported,
		})
	}
	return result
}

func symbolKind(flags ast.SymbolFlags) string {
	for _, candidate := range []struct {
		flag ast.SymbolFlags
		name string
	}{
		{ast.SymbolFlagsFunction, "function"},
		{ast.SymbolFlagsClass, "class"},
		{ast.SymbolFlagsInterface, "interface"},
		{ast.SymbolFlagsTypeAlias, "type"},
		{ast.SymbolFlagsVariable, "variable"},
	} {
		if flags&candidate.flag != 0 {
			return candidate.name
		}
	}
	return "symbol"
}

func moduleReferences(program *compiler.Program, file *ast.SourceFile) []ModuleReference {
	result := make([]ModuleReference, 0, len(file.Imports()))
	for _, specifier := range file.Imports() {
		localName := namespaceImportName(specifier.AsNode())
		if _, ok := builtinModule(specifier.Text()); ok {
			resolved := program.GetResolvedModuleFromModuleSpecifier(file, specifier)
			resolvedFileName := ""
			if resolved != nil {
				resolvedFileName = filepath.Clean(resolved.ResolvedFileName)
			}
			result = append(result, ModuleReference{Specifier: specifier.Text(), ResolvedFileName: resolvedFileName, LocalName: localName, Span: sourceSpan(specifier), Builtin: true})
			continue
		}
		resolved := program.GetResolvedModuleFromModuleSpecifier(file, specifier)
		if resolved == nil || resolved.ResolvedFileName == "" || resolved.IsExternalLibraryImport {
			continue
		}
		result = append(result, ModuleReference{
			Specifier:        specifier.Text(),
			ResolvedFileName: filepath.Clean(resolved.ResolvedFileName),
			LocalName:        localName,
			Span:             sourceSpan(specifier),
		})
	}
	return result
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
func orderedSourceFiles(files map[string]*ast.SourceFile, entry string, program *compiler.Program) []*ast.SourceFile {
	ordered := make([]*ast.SourceFile, 0, len(files))
	visited := make(map[string]bool, len(files))
	var visit func(string)
	visit = func(fileName string) {
		fileName = filepath.Clean(fileName)
		if visited[fileName] {
			return
		}
		file, ok := files[fileName]
		if !ok {
			return
		}
		visited[fileName] = true
		for _, reference := range moduleReferences(program, file) {
			visit(reference.ResolvedFileName)
		}
		ordered = append(ordered, file)
	}
	visit(entry)

	// Keep the adapter deterministic even if the compiler returns an unexpected
	// disconnected source file in a future TypeScript-Go revision.
	remaining := make([]string, 0, len(files)-len(ordered))
	for fileName := range files {
		if !visited[fileName] {
			remaining = append(remaining, fileName)
		}
	}
	sort.Strings(remaining)
	for _, fileName := range remaining {
		ordered = append(ordered, files[fileName])
	}
	return ordered
}
