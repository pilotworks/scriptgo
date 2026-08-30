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
	"github.com/microsoft/TypeScript/tsc/internal/compiler"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/vfs"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/osvfs"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/wrapvfs"
)

// CheckOptions configures type checking behavior.
type CheckOptions struct {
	ConfigPath string
}

// FindConfigFile searches for a tsconfig.json file starting in searchDir and walking up ancestor directories.
func FindConfigFile(searchDir string) string {
	abs, err := filepath.Abs(searchDir)
	if err != nil {
		return ""
	}
	for dir := abs; ; {
		candidate := filepath.Join(dir, "tsconfig.json")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir || parent == "." || parent == "" {
			break
		}
		dir = parent
	}
	return ""
}

// buildVirtualEnvironment prepares the virtual filesystem for builtin modules and types.
func buildVirtualEnvironment(cwd string) (vfs.FS, map[string]string, map[string]string) {
	baseFS := bundled.WrapFS(osvfs.FS())
	virtualFiles := map[string]string{}
	builtinPaths := map[string]string{}
	for name, module := range builtinModules {
		if name == "stream_consumers" {
			virtualPath := filepath.Join(cwd, "node_modules", "stream", "consumers", "index.ts")
			virtualFiles[virtualPath] = module.Source
			builtinPaths[virtualPath] = "stream/consumers"
			continue
		}
		virtualPath := filepath.Join(cwd, "node_modules", name, "index.ts")
		virtualFiles[virtualPath] = module.Source
		builtinPaths[virtualPath] = name
		if name == "webstreams" {
			vStreamWeb := filepath.Join(cwd, "node_modules", "stream", "web", "index.ts")
			virtualFiles[vStreamWeb] = module.Source
			builtinPaths[vStreamWeb] = "stream/web"
		}
	}

	var nodeTypesDts strings.Builder
	for name := range builtinModules {
		if name == "stream_consumers" {
			nodeTypesDts.WriteString("declare module \"node:stream/consumers\" {\n    export * from \"stream/consumers\";\n    import d from \"stream/consumers\";\n    export default d;\n}\n")
			continue
		}
		nodeTypesDts.WriteString(fmt.Sprintf("declare module \"node:%s\" {\n    export * from \"%s\";\n    import d from \"%s\";\n    export default d;\n}\n", name, name, name))
		if name == "webstreams" {
			nodeTypesDts.WriteString("declare module \"node:stream/web\" {\n    export * from \"webstreams\";\n    import d from \"webstreams\";\n    export default d;\n}\ndeclare module \"stream/web\" {\n    export * from \"webstreams\";\n    import d from \"webstreams\";\n    export default d;\n}\n")
		}
	}
	nodeTypesPath := filepath.Join(cwd, "node_modules", "@types", "node", "index.d.ts")
	virtualFiles[nodeTypesPath] = nodeTypesDts.String()

	fs := wrapvfs.Wrap(baseFS, wrapvfs.Replacements{
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
			if clean == cwd {
				return true
			}
			for virtualPath := range virtualFiles {
				for dir := filepath.Dir(virtualPath); dir != "." && dir != "/" && len(dir) >= len(cwd); dir = filepath.Dir(dir) {
					if clean == dir {
						return true
					}
				}
			}
			return baseFS.DirectoryExists(path)
		},
	})

	return fs, virtualFiles, builtinPaths
}

func formatTarget(target core.ScriptTarget) string {
	if target == 0 {
		return "ES2023"
	}
	return target.String()
}

func formatModule(module core.ModuleKind) string {
	if module == 0 {
		return "ESNext"
	}
	return module.String()
}

func formatResolution(res core.ModuleResolutionKind) string {
	if res == 0 {
		return "Bundler"
	}
	return res.String()
}

// CheckProject typechecks an entire TypeScript project rooted at a tsconfig.json file or directory.
func CheckProject(configPath string) (ProgramResult, error) {
	if err := EnsureStdlib(""); err != nil {
		return ProgramResult{}, fmt.Errorf("load stdlib: %w", err)
	}

	if configPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return ProgramResult{}, err
		}
		configPath = filepath.Join(cwd, "tsconfig.json")
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return ProgramResult{}, err
	}
	absPath = filepath.Clean(absPath)
	if info, err := os.Stat(absPath); err == nil && info.IsDir() {
		absPath = filepath.Join(absPath, "tsconfig.json")
	}

	cwd := filepath.Dir(absPath)
	fs, virtualFiles, builtinPaths := buildVirtualEnvironment(cwd)
	host := compiler.NewCompilerHost(cwd, fs, bundled.LibPath(), nil, nil, nil)

	parsedConfig, parseErrors := tsoptions.GetParsedCommandLineOfConfigFile(absPath, nil, nil, host, nil)
	if len(parseErrors) > 0 && parsedConfig == nil {
		return ProgramResult{
			Diagnostics: convertDiagnostics("config", parseErrors),
		}, nil
	}
	if parsedConfig == nil {
		return ProgramResult{
			Diagnostics: []Diagnostic{{
				FileName: absPath,
				Kind:     "config",
				Message:  fmt.Sprintf("could not load tsconfig at %s", absPath),
			}},
		}, nil
	}

	projectFileNames := append([]string(nil), parsedConfig.FileNames()...)
	rootFiles := append([]string(nil), projectFileNames...)
	for virtualPath := range virtualFiles {
		rootFiles = append(rootFiles, virtualPath)
	}
	sort.Strings(rootFiles)
	parsedConfig = parsedConfig.WithFileNames(rootFiles)

	opts := parsedConfig.CompilerOptions()
	if opts == nil {
		opts = &core.CompilerOptions{}
	}
	opts.NoEmit = core.TSTrue
	opts.AllowImportingTsExtensions = core.TSTrue
	parsedConfig.SetCompilerOptions(opts)

	program := compiler.NewProgram(compiler.ProgramOptions{
		Config:         parsedConfig,
		Host:           host,
		SingleThreaded: core.TSTrue,
	})

	result := ProgramResult{
		Options: CompilerOptions{
			Target:           formatTarget(opts.Target),
			Module:           formatModule(opts.Module),
			ModuleResolution: formatResolution(opts.ModuleResolution),
			Strict:           opts.Strict == core.TSTrue,
		},
	}

	result.Diagnostics = append(result.Diagnostics, convertDiagnostics("config", parseErrors)...)
	result.Diagnostics = append(result.Diagnostics, convertDiagnostics("config", parsedConfig.GetConfigFileParsingDiagnostics())...)
	result.Diagnostics = append(result.Diagnostics, convertDiagnostics("config", parsedConfig.Errors)...)

	ctx := context.Background()
	files := make(map[string]*ast.SourceFile)
	for _, file := range program.GetSourceFiles() {
		if program.IsSourceFileDefaultLibrary(file.Path()) || !isTypeScriptSource(file.FileName()) {
			continue
		}
		files[filepath.Clean(file.FileName())] = file
	}

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

	for _, root := range projectFileNames {
		visit(root)
	}

	for _, file := range ordered {
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
