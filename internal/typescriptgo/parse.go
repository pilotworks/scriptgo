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
				opts.AllowJs = core.TSTrue
				opts.AllowImportingTsExtensions = core.TSTrue
				parsedCfg.SetCompilerOptions(opts)
				validationDiags := ValidateCompilerOptions(opts, cfgAbs)
				compilerOpts = CompilerOptions{
					Target:           formatTarget(opts.Target),
					Module:           formatModule(opts.Module),
					ModuleResolution: formatResolution(opts.ModuleResolution),
					Strict:           opts.Strict == core.TSTrue,
				}
				if len(validationDiags) > 0 {
					return ProgramResult{
						Options:     compilerOpts,
						Diagnostics: validationDiags,
					}, nil
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
			AllowJs:                    core.TSTrue,
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

	program, config := newProgramWithResolvedJavaScript(config, host, cwd)

	result := ProgramResult{
		Options:     compilerOpts,
		Diagnostics: convertDiagnostics("config", configDiagnostics),
	}
	ctx := context.Background()
	files := make(map[string]*ast.SourceFile)
	for _, file := range program.GetSourceFiles() {
		if program.IsSourceFileDefaultLibrary(file.Path()) || !isSupportedSource(file.FileName()) {
			continue
		}
		files[filepath.Clean(file.FileName())] = file
	}
	for _, file := range orderedSourceFiles(files, absoluteEntry, program, builtinPaths) {
		if program.IsSourceFileDefaultLibrary(file.Path()) || !isSupportedSource(file.FileName()) {
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
		if builtinPaths[filepath.Clean(file.FileName())] == "" {
			result.Diagnostics = append(result.Diagnostics, convertDiagnostics("syntax", program.GetSyntacticDiagnostics(ctx, file))...)
			result.Diagnostics = append(result.Diagnostics, convertDiagnostics("type", program.GetSemanticDiagnostics(ctx, file))...)
		}
	}
	result.Diagnostics = append(result.Diagnostics, convertDiagnostics("program", program.GetProgramDiagnostics())...)
	return result, nil
}

// newProgramWithResolvedJavaScript adds external JavaScript package entries as
// roots. TypeScript-Go resolves these edges but excludes external-library JS
// from GetSourceFiles unless they are explicit program roots.
func newProgramWithResolvedJavaScript(config *tsoptions.ParsedCommandLine, host compiler.CompilerHost, cwd string) (*compiler.Program, *tsoptions.ParsedCommandLine) {
	for {
		program := compiler.NewProgram(compiler.ProgramOptions{
			Config:         config,
			Host:           host,
			SingleThreaded: core.TSTrue,
		})
		known := make(map[string]bool, len(config.FileNames()))
		for _, fileName := range config.FileNames() {
			known[filepath.Clean(fileName)] = true
		}
		var additions []string
		for _, file := range program.GetSourceFiles() {
			for _, reference := range moduleReferences(program, file, cwd) {
				if !isJavaScriptSource(reference.ResolvedFileName) {
					continue
				}
				resolved := filepath.Clean(reference.ResolvedFileName)
				if known[resolved] {
					continue
				}
				if _, err := os.Stat(resolved); err != nil {
					continue
				}
				known[resolved] = true
				additions = append(additions, resolved)
			}
		}
		if len(additions) == 0 {
			return program, config
		}
		sort.Strings(additions)
		config = config.WithFileNames(append(config.FileNames(), additions...))
	}
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

