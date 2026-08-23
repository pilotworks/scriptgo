package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pilotworks/scriptgo/internal/compiler"
)

func main() {
	if len(os.Args) < 2 {
		printMainUsage()
		os.Exit(2)
	}

	firstArg := os.Args[1]
	switch firstArg {
	case "run":
		handleRun(normalizeFlagsFirst(os.Args[2:]))
	case "build":
		handleBuild(normalizeFlagsFirst(os.Args[2:]))
	case "check":
		handleCheck(normalizeFlagsFirst(os.Args[2:]))
	case "emit":
		handleEmit(normalizeFlagsFirst(os.Args[2:]))
	case "version", "--version", "-V":
		fmt.Printf("scriptgo version %s (runtime %s)\n", compiler.Version, compiler.RuntimeABIVersion)
	case "help", "--help", "-h":
		if len(os.Args) > 2 {
			handleHelpCommand(os.Args[2])
			return
		}
		printMainUsage()
	default:
		fmt.Fprintf(os.Stderr, "scriptgo: unknown command %q\n\n", firstArg)
		printMainUsage()
		os.Exit(2)
	}
}

func handleHelpCommand(cmd string) {
	switch cmd {
	case "run":
		printRunUsage()
	case "build":
		printBuildUsage()
	case "check":
		printCheckUsage()
	case "emit":
		printEmitUsage()
	case "version":
		fmt.Println("Usage: scriptgo version\n\nPrints the current compiler version and runtime ABI version.")
	default:
		fmt.Fprintf(os.Stderr, "scriptgo: unknown command %q for help\n\n", cmd)
		printMainUsage()
		os.Exit(2)
	}
}

func normalizeFlagsFirst(args []string) []string {
	var flags []string
	var positionals []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if (arg == "-o" || arg == "-target" || arg == "--target" || arg == "-cc" || arg == "--cc" || arg == "-sanitize" || arg == "--sanitize" || arg == "-mode" || arg == "--mode" || arg == "-e" || arg == "--eval" || arg == "-m" || arg == "-ffi-manifest" || arg == "--ffi-manifest") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
		} else {
			positionals = append(positionals, arg)
		}
	}
	return append(flags, positionals...)
}

func resolveInput(arg string) (string, func(), error) {
	if _, err := os.Stat(arg); err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("file not found: %s", arg)
		}
		return "", nil, err
	}
	return arg, func() {}, nil
}

func createInlineSourceFile(code string) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "scriptgo-inline-")
	if err != nil {
		return "", nil, fmt.Errorf("create temporary directory: %w", err)
	}
	filePath := filepath.Join(tempDir, "main.ts")
	if err := os.WriteFile(filePath, []byte(code), 0o644); err != nil {
		os.RemoveAll(tempDir)
		return "", nil, fmt.Errorf("write temporary script: %w", err)
	}
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	return filePath, cleanup, nil
}

func handleRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.Usage = printRunUsage
	eval := fs.String("e", "", "evaluate inline script string")
	verbose := fs.Bool("v", false, "print compilation stages to stderr")
	target := fs.String("target", "", "native target triple (default: $SCRIPTGO_TARGET or native)")
	cc := fs.String("cc", "", "C compiler / toolchain driver (default: $SCRIPTGO_CC or clang)")
	debug := fs.Bool("debug", false, "include native debug metadata")
	sanitize := fs.String("sanitize", "", "enable clang sanitizers (comma-separated: address,undefined,leak)")
	native := fs.Bool("native", false, "compile to native executable and execute directly")
	warnRuntimeCasts := fs.Bool("warn-runtime-casts", false, "warn on runtime checked casts")
	strictCasts := fs.Bool("strict-casts", false, "treat cast warnings as errors")
	ffiManifest := fs.String("ffi-manifest", "", "path to FFI JSON metadata manifest (*.ffi.json)")
	fs.StringVar(ffiManifest, "m", "", "path to FFI JSON metadata manifest (shorthand)")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	var entryPath string
	var cleanup func()
	var extraArgs []string
	var extraSources []string

	if *eval != "" {
		var err error
		entryPath, cleanup, err = createInlineSourceFile(*eval)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scriptgo: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()
		extraArgs = fs.Args()
	} else if fs.NArg() >= 1 {
		var err error
		entryPath, cleanup, err = resolveInput(fs.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "scriptgo: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()
		for _, arg := range fs.Args()[1:] {
			ext := filepath.Ext(arg)
			if ext == ".c" || ext == ".o" || ext == ".a" {
				extraSources = append(extraSources, arg)
			} else {
				extraArgs = append(extraArgs, arg)
			}
		}
	} else {
		printRunUsage()
		os.Exit(2)
	}

	var manifests []string
	if *ffiManifest != "" {
		manifests = append(manifests, *ffiManifest)
	}

	options := compiler.BuildOptions{
		CC:               *cc,
		Target:           *target,
		Debug:            *debug,
		Sanitizers:       splitList(*sanitize),
		WarnRuntimeCasts: *warnRuntimeCasts,
		StrictCasts:      *strictCasts,
		FFIManifests:     manifests,
		ExtraSources:     extraSources,
	}

	if *native {
		if *verbose {
			fmt.Fprintf(os.Stderr, "scriptgo: compiling %s to native binary for execution\n", entryPath)
		}
		tempDir, err := os.MkdirTemp("", "scriptgo-run-")
		if err != nil {
			fmt.Fprintf(os.Stderr, "scriptgo: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(tempDir)
		binPath := filepath.Join(tempDir, "app")
		if err := compiler.BuildWithOptions(entryPath, binPath, options); err != nil {
			printError(err)
			os.Exit(1)
		}
		printCompilerWarnings()
		cmd := exec.Command(binPath, extraArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "scriptgo: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "scriptgo: interpreting %s\n", entryPath)
	}
	result, err := compiler.Run(entryPath)
	printCompilerWarnings()
	if err != nil {
		printError(err)
		os.Exit(1)
	}
	fmt.Print(result)
}

func handleBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.Usage = printBuildUsage
	eval := fs.String("e", "", "evaluate inline script string")
	output := fs.String("o", "", "write generated binary to this path (default: ./<entry_name>)")
	verbose := fs.Bool("v", false, "print compilation stages to stderr")
	target := fs.String("target", "", "native target triple (default: $SCRIPTGO_TARGET or native)")
	cc := fs.String("cc", "", "C compiler / toolchain driver (default: $SCRIPTGO_CC or clang)")
	debug := fs.Bool("debug", false, "include native debug metadata")
	sanitize := fs.String("sanitize", "", "enable clang sanitizers (comma-separated: address,undefined,leak)")
	warnRuntimeCasts := fs.Bool("warn-runtime-casts", false, "warn on runtime checked casts")
	strictCasts := fs.Bool("strict-casts", false, "treat cast warnings as errors")
	ffiManifest := fs.String("ffi-manifest", "", "path to FFI JSON metadata manifest (*.ffi.json)")
	fs.StringVar(ffiManifest, "m", "", "path to FFI JSON metadata manifest (shorthand)")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	var entryPath string
	var cleanup func()
	var input string
	var extraSources []string

	if *eval != "" {
		var err error
		entryPath, cleanup, err = createInlineSourceFile(*eval)
		if err != nil {
			fmt.Fprintln(os.Stderr, "scriptgo:", err)
			os.Exit(1)
		}
		defer cleanup()
		for _, arg := range fs.Args() {
			ext := filepath.Ext(arg)
			if ext == ".c" || ext == ".o" || ext == ".a" {
				extraSources = append(extraSources, arg)
			}
		}
	} else if fs.NArg() >= 1 {
		input = fs.Arg(0)
		var err error
		entryPath, cleanup, err = resolveInput(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, "scriptgo:", err)
			os.Exit(1)
		}
		defer cleanup()
		for _, arg := range fs.Args()[1:] {
			ext := filepath.Ext(arg)
			if ext == ".c" || ext == ".o" || ext == ".a" {
				extraSources = append(extraSources, arg)
			}
		}
	} else {
		printBuildUsage()
		os.Exit(2)
	}

	outputPath := *output
	if outputPath == "" {
		if input != "" {
			base := filepath.Base(input)
			ext := filepath.Ext(base)
			outputPath = strings.TrimSuffix(base, ext)
		}
		if outputPath == "" {
			outputPath = "app"
		}
	}

	var manifests []string
	if *ffiManifest != "" {
		manifests = append(manifests, *ffiManifest)
	}

	options := compiler.BuildOptions{
		CC:               *cc,
		Target:           *target,
		Debug:            *debug,
		Sanitizers:       splitList(*sanitize),
		WarnRuntimeCasts: *warnRuntimeCasts,
		StrictCasts:      *strictCasts,
		FFIManifests:     manifests,
		ExtraSources:     extraSources,
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "scriptgo: build %s -> %s\n", entryPath, outputPath)
	}
	if err := compiler.BuildWithOptions(entryPath, outputPath, options); err != nil {
		printCompilerWarnings()
		printError(err)
		os.Exit(1)
	}
	printCompilerWarnings()
}

func handleCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.Usage = printCheckUsage
	eval := fs.String("e", "", "evaluate inline script string")
	verbose := fs.Bool("v", false, "print compilation stages to stderr")
	warnRuntimeCasts := fs.Bool("warn-runtime-casts", false, "warn on runtime checked casts")
	strictCasts := fs.Bool("strict-casts", false, "treat cast warnings as errors")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	var entryPath string
	var cleanup func()

	if *eval != "" {
		var err error
		entryPath, cleanup, err = createInlineSourceFile(*eval)
		if err != nil {
			fmt.Fprintln(os.Stderr, "scriptgo:", err)
			os.Exit(1)
		}
		defer cleanup()
	} else if fs.NArg() == 1 {
		var err error
		entryPath, cleanup, err = resolveInput(fs.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, "scriptgo:", err)
			os.Exit(1)
		}
		defer cleanup()
	} else {
		printCheckUsage()
		os.Exit(2)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "scriptgo: checking %s\n", entryPath)
	}
	options := compiler.BuildOptions{
		WarnRuntimeCasts: *warnRuntimeCasts,
		StrictCasts:      *strictCasts,
	}
	if _, err := compiler.CompileWithOptions(entryPath, options); err != nil {
		printCompilerWarnings()
		printError(err)
		os.Exit(1)
	}
	printCompilerWarnings()
	if *verbose {
		fmt.Fprintf(os.Stderr, "scriptgo: %s checked successfully\n", entryPath)
	}
}

func handleEmit(args []string) {
	fs := flag.NewFlagSet("emit", flag.ContinueOnError)
	fs.Usage = printEmitUsage
	eval := fs.String("e", "", "evaluate inline script string")
	mode := fs.String("mode", "llvm-ir", "output mode: llvm-ir, typed-ir")
	output := fs.String("o", "", "write output to this path (default: stdout)")
	verbose := fs.Bool("v", false, "print compilation stages to stderr")
	target := fs.String("target", "", "native target triple (default: $SCRIPTGO_TARGET or native)")
	debug := fs.Bool("debug", false, "include native debug metadata")
	warnRuntimeCasts := fs.Bool("warn-runtime-casts", false, "warn on runtime checked casts")
	strictCasts := fs.Bool("strict-casts", false, "treat cast warnings as errors")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	var entryPath string
	var cleanup func()

	if *eval != "" {
		var err error
		entryPath, cleanup, err = createInlineSourceFile(*eval)
		if err != nil {
			fmt.Fprintln(os.Stderr, "scriptgo:", err)
			os.Exit(1)
		}
		defer cleanup()
	} else if fs.NArg() == 1 {
		var err error
		entryPath, cleanup, err = resolveInput(fs.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, "scriptgo:", err)
			os.Exit(1)
		}
		defer cleanup()
	} else {
		printEmitUsage()
		os.Exit(2)
	}

	options := compiler.BuildOptions{
		Target:           *target,
		Debug:            *debug,
		WarnRuntimeCasts: *warnRuntimeCasts,
		StrictCasts:      *strictCasts,
	}
	var result string
	var err error

	switch *mode {
	case "typed-ir":
		if *verbose {
			fmt.Fprintf(os.Stderr, "scriptgo: emitting typed IR for %s\n", entryPath)
		}
		result, err = compiler.DumpIR(entryPath)
	case "llvm-ir":
		if *verbose {
			fmt.Fprintf(os.Stderr, "scriptgo: emitting LLVM IR for %s\n", entryPath)
		}
		result, err = compiler.CompileWithOptions(entryPath, options)
	default:
		fmt.Fprintf(os.Stderr, "scriptgo: unsupported emit mode %q (supported: llvm-ir, typed-ir)\n", *mode)
		os.Exit(2)
	}

	printCompilerWarnings()
	if err != nil {
		printError(err)
		os.Exit(1)
	}

	if *output == "" {
		fmt.Print(result)
		return
	}
	if err := os.WriteFile(*output, []byte(result), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "scriptgo:", err)
		os.Exit(1)
	}
}

func printMainUsage() {
	fmt.Fprintln(os.Stderr, `ScriptGo - TypeScript Native Compiler

Usage:
  scriptgo <command> [flags] <arguments>

Commands:
  run       Execute a TypeScript program or code string via interpreter or native binary
  build     Compile TypeScript into a standalone native executable
  check     Verify TypeScript syntax, types, and native subset rules
  emit      Emit intermediate representation (LLVM IR or Typed IR)
  version   Print compiler and runtime ABI version
  help      Show help for ScriptGo or a specific command

Global Flags:
  -v                     Verbose output
  --target <triple>      Target architecture triple (default: $SCRIPTGO_TARGET or native)
  --cc <driver>          C compiler / toolchain driver (default: $SCRIPTGO_CC or clang)
  --debug                Emit native DWARF debug symbols
  --sanitize <list>      Enable Clang sanitizers (address, undefined, leak)
  --warn-runtime-casts   Warn on runtime checked casts (SG4005)
  --strict-casts         Treat cast warnings as errors
  -h, --help             Show help message

Use 'scriptgo help <command>' or 'scriptgo <command> --help' for detailed command usage.`)
}

func printRunUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo run [flags] <entry.ts> [-- <args...>]
  scriptgo run [flags] -e "<code string>" [-- <args...>]

Description:
  Executes a TypeScript file or inline code string using either the reference semantic
  interpreter (default) or by compiling directly to a temporary native binary.

Flags:
  -e <string>            Evaluate inline script string
  --native               Compile to native executable and execute directly on host
  -m, --ffi-manifest     Path to FFI JSON metadata manifest (*.ffi.json)
  -v                     Verbose output (print compilation stages)
  --target <triple>      Target architecture triple (default: $SCRIPTGO_TARGET or native)
  --cc <driver>          C compiler / toolchain driver (default: $SCRIPTGO_CC or clang)
  --debug                Include DWARF debug symbols
  --sanitize <list>      Enable Clang sanitizers (address, undefined, leak)
  --warn-runtime-casts   Warn on runtime checked casts (SG4005)
  --strict-casts         Treat cast warnings as errors
  -h, --help             Show this help message

Examples:
  scriptgo run app.ts
  scriptgo run -e "console.log('hello ' + 42)"
  scriptgo run --native -e "console.log(100 * 20)"
  scriptgo run --native app.ts --ffi-manifest mylib.ffi.json
  scriptgo run --native app.ts helper.c
  scriptgo run --native --cc "zig cc" app.ts
  scriptgo run app.ts -- arg1 arg2`)
}

func printBuildUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo build [flags] <entry.ts> [sources.c...] [-o <output>]
  scriptgo build [flags] -e "<code string>" [-o <output>]

Description:
  Compiles a TypeScript program into a standalone, optimized native executable
  linked with the host C runtime.

Flags:
  -e <string>            Evaluate inline script string
  -o <path>              Output binary path (default: ./<entry_name>)
  -m, --ffi-manifest     Path to FFI JSON metadata manifest (*.ffi.json)
  -v                     Verbose output (print compilation stages)
  --target <triple>      Target architecture triple (default: $SCRIPTGO_TARGET or native)
  --cc <driver>          C compiler / toolchain driver (default: $SCRIPTGO_CC or clang)
  --debug                Include DWARF debug symbols (O0 with debug metadata)
  --sanitize <list>      Enable Clang sanitizers (address, undefined, leak)
  --warn-runtime-casts   Warn on runtime checked casts (SG4005)
  --strict-casts         Treat cast warnings as errors
  -h, --help             Show this help message

Examples:
  scriptgo build server.ts
  scriptgo build server.ts -o /usr/local/bin/server
  scriptgo build app.ts --ffi-manifest sqlite3.ffi.json -o myapp
  scriptgo build app.ts helper.c -o myapp
  scriptgo build cli.ts --cc "zig cc" --target x86_64-linux-gnu -o cli_linux
  scriptgo build cli.ts --debug --sanitize address -o cli_debug`)
}

func printCheckUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo check [flags] <entry.ts>
  scriptgo check [flags] -e "<code string>"

Description:
  Type-checks and validates the reachable source graph and native subset
  eligibility without invoking code generation or Clang.

Flags:
  -e <string>            Evaluate inline script string
  -v                     Verbose output (print check stages and confirmation)
  --warn-runtime-casts   Warn on runtime checked casts (SG4005)
  --strict-casts         Treat cast warnings as errors
  -h, --help             Show this help message

Examples:
  scriptgo check app.ts
  scriptgo check -e "const x: number = 42; console.log(x);"
  scriptgo check -v src/main.ts`)
}

func printEmitUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo emit [flags] <entry.ts> [--mode llvm-ir|typed-ir] [-o <output>]
  scriptgo emit [flags] -e "<code string>" [--mode llvm-ir|typed-ir] [-o <output>]

Description:
  Emits intermediate representations (Typed IR or LLVM IR) for debugging
  and compiler inspection.

Flags:
  -e <string>            Evaluate inline script string
  --mode <mode>          Output mode: llvm-ir (default), typed-ir
  -o <path>              Write emitted IR to file instead of stdout
  -v                     Verbose output (print compilation stages)
  --target <triple>      Target architecture triple (default: $SCRIPTGO_TARGET, $TARGET, or native)
  --debug                Include DWARF debug symbols in LLVM IR
  --warn-runtime-casts   Warn on runtime checked casts (SG4005)
  --strict-casts         Treat cast warnings as errors
  -h, --help             Show this help message

Examples:
  scriptgo emit app.ts
  scriptgo emit -e "console.log(123)" --mode typed-ir
  scriptgo emit app.ts --mode llvm-ir -o app.ll`)
}

func printCompilerWarnings() {
	warns := compiler.GetWarnings()
	for _, w := range warns {
		fmt.Fprintln(os.Stderr, w.Format())
	}
}

func printError(err error) {
	if err == nil {
		return
	}
	errStr := err.Error()
	if strings.Contains(errStr, " - error ") || strings.Contains(errStr, " - warning ") {
		fmt.Fprintln(os.Stderr, errStr)
	} else {
		fmt.Fprintln(os.Stderr, "scriptgo:", errStr)
	}
}

func splitList(value string) []string {
	if value == "" {
		return nil
	}
	values := strings.Split(value, ",")
	result := make([]string, 0, len(values))
	for _, item := range values {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
