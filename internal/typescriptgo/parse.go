package typescriptgo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/bundled"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
	"github.com/microsoft/TypeScript/tsc/internal/compiler"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
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

// ParseFileToSyntax parses one TypeScript source file into a SyntaxFile AST structure.
func ParseFileToSyntax(fileName, source string) (SyntaxFile, error) {
	absoluteName, err := filepath.Abs(fileName)
	if err != nil {
		return SyntaxFile{}, err
	}
	absoluteName = filepath.Clean(absoluteName)
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: absoluteName,
		Path:     tspath.ToPath(absoluteName, "", true),
	}, source, core.ScriptKindTS)
	return syntaxFile(file, nil), nil
}



// Check creates a TypeScript-Go program. Program creation performs local module
// resolution and binds/checks the complete reachable source graph.
func Check(entryPath string) (ProgramResult, error) {
	return CheckWithOptions(entryPath, CheckOptions{})
}

// CheckWithOptions creates a TypeScript-Go program with custom check options.
func CheckWithOptions(entryPath string, checkOpts CheckOptions) (ProgramResult, error) {
	if err := EnsureStdlib(""); err != nil {
		return ProgramResult{}, fmt.Errorf("load stdlib: %w", err)
	}

	if entryPath == "" && checkOpts.ConfigPath != "" {
		return CheckProject(checkOpts.ConfigPath)
	}

	absoluteEntry, err := filepath.Abs(entryPath)
	if err != nil {
		return ProgramResult{}, err
	}
	absoluteEntry = filepath.Clean(absoluteEntry)

	// If entryPath is a json file or a directory, check as a project
	if strings.HasSuffix(absoluteEntry, ".json") {
		return CheckProject(absoluteEntry)
	}
	if info, err := os.Stat(absoluteEntry); err == nil && info.IsDir() {
		return CheckProject(absoluteEntry)
	}

	cwd := filepath.Dir(absoluteEntry)
	fs, virtualFiles, builtinPaths := buildVirtualEnvironment(cwd)
	host := compiler.NewCompilerHost(cwd, fs, bundled.LibPath(), nil, nil, nil)

	var config *tsoptions.ParsedCommandLine
	var configDiagnostics []*ast.Diagnostic
	var compilerOpts CompilerOptions

	configPath := checkOpts.ConfigPath
	if configPath == "" {
		configPath = FindConfigFile(cwd)
	}

	if configPath != "" {
		cfgAbs, err := filepath.Abs(configPath)
		if err == nil {
			parsedCfg, parseErrs := tsoptions.GetParsedCommandLineOfConfigFile(cfgAbs, nil, nil, host, nil)
			if parsedCfg != nil {
				config = parsedCfg
				configDiagnostics = append(configDiagnostics, parseErrs...)
				configDiagnostics = append(configDiagnostics, parsedCfg.GetConfigFileParsingDiagnostics()...)
				configDiagnostics = append(configDiagnostics, parsedCfg.Errors...)
				opts := parsedCfg.CompilerOptions()
				if opts == nil {
					opts = &core.CompilerOptions{}
				}
				opts.NoEmit = core.TSTrue
				opts.AllowImportingTsExtensions = core.TSTrue
				parsedCfg.SetCompilerOptions(opts)
				compilerOpts = CompilerOptions{
					Target:           formatTarget(opts.Target),
					Module:           formatModule(opts.Module),
					ModuleResolution: formatResolution(opts.ModuleResolution),
					Strict:           opts.Strict == core.TSTrue,
				}
			}
		}
	}

	if config == nil {
		options := &core.CompilerOptions{
			Target:                     core.ScriptTargetESNext,
			Module:                     core.ModuleKindESNext,
			ModuleResolution:           core.ModuleResolutionKindBundler,
			AllowImportingTsExtensions: core.TSTrue,
			Lib:                        []string{"lib.esnext.d.ts"},
			Strict:                     core.TSTrue,
			NoEmit:                     core.TSTrue,
		}
		comparePaths := tspath.ComparePathsOptions{
			CurrentDirectory:          cwd,
			UseCaseSensitiveFileNames: fs.UseCaseSensitiveFileNames(),
		}
		rootFileNames := []string{absoluteEntry}
		for virtualPath := range virtualFiles {
			rootFileNames = append(rootFileNames, virtualPath)
		}
		sort.Strings(rootFileNames[1:])
		config = tsoptions.NewParsedCommandLine(options, rootFileNames, comparePaths)
		compilerOpts = CompilerOptions{
			Target:           "ES2023",
			Module:           "ESNext",
			ModuleResolution: "Bundler",
			Strict:           true,
		}
	} else {
		rootFileNames := append([]string{absoluteEntry}, config.FileNames()...)
		for virtualPath := range virtualFiles {
			rootFileNames = append(rootFileNames, virtualPath)
		}
		sort.Strings(rootFileNames)
		config = config.WithFileNames(rootFileNames)
	}

	program := compiler.NewProgram(compiler.ProgramOptions{
		Config:         config,
		Host:           host,
		SingleThreaded: core.TSTrue,
	})

	result := ProgramResult{
		Options:     compilerOpts,
		Diagnostics: convertDiagnostics("config", configDiagnostics),
	}
	ctx := context.Background()
	files := make(map[string]*ast.SourceFile)
	for _, file := range program.GetSourceFiles() {
		if program.IsSourceFileDefaultLibrary(file.Path()) || !isTypeScriptSource(file.FileName()) {
			continue
		}
		files[filepath.Clean(file.FileName())] = file
	}
	for _, file := range orderedSourceFiles(files, absoluteEntry, program, builtinPaths) {
		if program.IsSourceFileDefaultLibrary(file.Path()) || !isTypeScriptSource(file.FileName()) {
			continue
		}
		checkerInstance, done := program.GetTypeCheckerForFile(ctx, file)
		symbols := fileSymbolsWithChecker(checkerInstance, file)
		syntax := syntaxFile(file, checkerInstance)
		done()
		result.Files = append(result.Files, SourceFile{
			FileName:       file.FileName(),
			Source:         file.Text(),
			StatementCount: statementCount(file),
			Span:           sourceSpan(&file.Node),
			Imports:        moduleReferences(program, file, cwd),
			Builtin:        builtinPaths[filepath.Clean(file.FileName())] != "",
			BuiltinName:    builtinPaths[filepath.Clean(file.FileName())],
			Symbols:        symbols,
			Syntax:         syntax,
		})
		result.Diagnostics = append(result.Diagnostics, convertDiagnostics("syntax", program.GetSyntacticDiagnostics(ctx, file))...)
		result.Diagnostics = append(result.Diagnostics, convertDiagnostics("type", program.GetSemanticDiagnostics(ctx, file))...)
	}
	result.Diagnostics = append(result.Diagnostics, convertDiagnostics("program", program.GetProgramDiagnostics())...)
	return result, nil
}

func fileSymbolsWithChecker(checkerInstance *checker.Checker, file *ast.SourceFile) []Symbol {
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

func moduleReferences(program *compiler.Program, file *ast.SourceFile, cwd string) []ModuleReference {
	result := make([]ModuleReference, 0, len(file.Imports()))
	for _, specifier := range file.Imports() {
		localName := namespaceImportName(specifier.AsNode())
		if module, ok := builtinModule(specifier.Text()); ok {
			resolved := program.GetResolvedModuleFromModuleSpecifier(file, specifier)
			resolvedFileName := ""
			if resolved != nil && resolved.ResolvedFileName != "" && !strings.HasSuffix(resolved.ResolvedFileName, ".d.ts") {
				resolvedFileName = filepath.Clean(resolved.ResolvedFileName)
			} else {
				resolvedFileName = filepath.Clean(filepath.Join(cwd, "node_modules", module.Name, "index.ts"))
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
func orderedSourceFiles(files map[string]*ast.SourceFile, entry string, program *compiler.Program, builtinPaths map[string]string) []*ast.SourceFile {
	cwd := filepath.Dir(entry)
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
		for _, reference := range moduleReferences(program, file, cwd) {
			visit(reference.ResolvedFileName)
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
