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
			if (arg == "-o" || arg == "-target" || arg == "--target" || arg == "-sanitize" || arg == "--sanitize" || arg == "-mode" || arg == "--mode" || arg == "-e" || arg == "--eval") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
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
	if _, err := os.Stat(arg); err == nil {
		return arg, func() {}, nil
	}
	return createInlineSourceFile(arg)
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
	target := fs.String("target", "native", "native target triple, or native for the host")
	debug := fs.Bool("debug", false, "include native debug metadata")
	sanitize := fs.String("sanitize", "", "enable clang sanitizers (comma-separated: address,undefined,leak)")
	native := fs.Bool("native", false, "compile to native executable and execute directly")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	var entryPath string
	var cleanup func()
	var extraArgs []string

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
		extraArgs = fs.Args()[1:]
	} else {
		printRunUsage()
		os.Exit(2)
	}

	options := compiler.BuildOptions{Target: *target, Debug: *debug, Sanitizers: splitList(*sanitize)}

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
			fmt.Fprintf(os.Stderr, "scriptgo: %v\n", err)
			os.Exit(1)
		}
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
	if err != nil {
		fmt.Fprintln(os.Stderr, "scriptgo:", err)
		os.Exit(1)
	}
	fmt.Print(result)
}

func handleBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.Usage = printBuildUsage
	output := fs.String("o", "", "write generated binary to this path (default: ./<entry_name>)")
	verbose := fs.Bool("v", false, "print compilation stages to stderr")
	target := fs.String("target", "native", "native target triple, or native for the host")
	debug := fs.Bool("debug", false, "include native debug metadata")
	sanitize := fs.String("sanitize", "", "enable clang sanitizers (comma-separated: address,undefined,leak)")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	if fs.NArg() != 1 {
		printBuildUsage()
		os.Exit(2)
	}

	input := fs.Arg(0)
	entryPath, cleanup, err := resolveInput(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scriptgo:", err)
		os.Exit(1)
	}
	defer cleanup()

	outputPath := *output
	if outputPath == "" {
		if _, err := os.Stat(input); err == nil {
			base := filepath.Base(input)
			ext := filepath.Ext(base)
			outputPath = strings.TrimSuffix(base, ext)
		}
		if outputPath == "" {
			outputPath = "app"
		}
	}

	options := compiler.BuildOptions{Target: *target, Debug: *debug, Sanitizers: splitList(*sanitize)}
	if *verbose {
		fmt.Fprintf(os.Stderr, "scriptgo: build %s -> %s\n", entryPath, outputPath)
	}
	if err := compiler.BuildWithOptions(entryPath, outputPath, options); err != nil {
		fmt.Fprintln(os.Stderr, "scriptgo:", err)
		os.Exit(1)
	}
}

func handleCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.Usage = printCheckUsage
	verbose := fs.Bool("v", false, "print compilation stages to stderr")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	if fs.NArg() != 1 {
		printCheckUsage()
		os.Exit(2)
	}

	entryPath, cleanup, err := resolveInput(fs.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "scriptgo:", err)
		os.Exit(1)
	}
	defer cleanup()

	if *verbose {
		fmt.Fprintf(os.Stderr, "scriptgo: checking %s\n", entryPath)
	}
	if err := compiler.Check(entryPath); err != nil {
		fmt.Fprintln(os.Stderr, "scriptgo:", err)
		os.Exit(1)
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "scriptgo: %s checked successfully\n", entryPath)
	}
}

func handleEmit(args []string) {
	fs := flag.NewFlagSet("emit", flag.ContinueOnError)
	fs.Usage = printEmitUsage
	mode := fs.String("mode", "llvm-ir", "output mode: llvm-ir, typed-ir")
	output := fs.String("o", "", "write output to this path (default: stdout)")
	verbose := fs.Bool("v", false, "print compilation stages to stderr")
	target := fs.String("target", "native", "native target triple, or native for the host")
	debug := fs.Bool("debug", false, "include native debug metadata")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	if fs.NArg() != 1 {
		printEmitUsage()
		os.Exit(2)
	}

	entryPath, cleanup, err := resolveInput(fs.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "scriptgo:", err)
		os.Exit(1)
	}
	defer cleanup()

	var result string

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
		options := compiler.BuildOptions{Target: *target, Debug: *debug}
		result, err = compiler.CompileWithOptions(entryPath, options)
	default:
		fmt.Fprintf(os.Stderr, "scriptgo: unsupported emit mode %q (supported: llvm-ir, typed-ir)\n", *mode)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "scriptgo:", err)
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
  -v                 Verbose output
  --target <triple>  Target architecture triple (default: native)
  --debug            Emit native DWARF debug symbols
  --sanitize <list>  Enable Clang sanitizers (address, undefined, leak)
  -h, --help         Show help message

Use 'scriptgo help <command>' or 'scriptgo <command> --help' for detailed command usage.`)
}

func printRunUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo run [flags] <entry.ts | "code string"> [-- <args...>]

Description:
  Executes a TypeScript file or inline code string using either the reference semantic
  interpreter (default) or by compiling directly to a temporary native binary.

Flags:
  -e <string>        Evaluate inline script string
  --native           Compile to native executable and execute directly on host
  -v                 Verbose output (print compilation stages)
  --target <triple>  Target architecture triple (default: native)
  --debug            Include DWARF debug symbols
  --sanitize <list>  Enable Clang sanitizers (address, undefined, leak)
  -h, --help         Show this help message

Examples:
  scriptgo run app.ts
  scriptgo run "console.log(1231)"
  scriptgo run -e "console.log('hello ' + 42)"
  scriptgo run --native "console.log(100 * 20)"
  scriptgo run app.ts -- arg1 arg2`)
}

func printBuildUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo build [flags] <entry.ts | "code string"> [-o <output>]

Description:
  Compiles a TypeScript program into a standalone, optimized native executable
  linked with the host C runtime.

Flags:
  -o <path>          Output binary path (default: ./<entry_name>)
  -v                 Verbose output (print compilation stages)
  --target <triple>  Target architecture triple (default: native)
  --debug            Include DWARF debug symbols (O0 with debug metadata)
  --sanitize <list>  Enable Clang sanitizers (address, undefined, leak)
  -h, --help         Show this help message

Examples:
  scriptgo build server.ts
  scriptgo build "console.log('built natively')" -o hello
  scriptgo build server.ts -o /usr/local/bin/server
  scriptgo build cli.ts --debug --sanitize address -o cli_debug`)
}

func printCheckUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo check [flags] <entry.ts | "code string">

Description:
  Type-checks and validates the reachable source graph and native subset
  eligibility without invoking code generation or Clang.

Flags:
  -v                 Verbose output (print check stages and confirmation)
  -h, --help         Show this help message

Examples:
  scriptgo check app.ts
  scriptgo check "const x: number = 42; console.log(x);"
  scriptgo check -v src/main.ts`)
}

func printEmitUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo emit [flags] <entry.ts | "code string"> [--mode llvm-ir|typed-ir] [-o <output>]

Description:
  Emits intermediate representations (Typed IR or LLVM IR) for debugging
  and compiler inspection.

Flags:
  --mode <mode>      Output mode: llvm-ir (default), typed-ir
  -o <path>          Write emitted IR to file instead of stdout
  -v                 Verbose output (print compilation stages)
  --target <triple>  Target architecture triple (default: native)
  --debug            Include DWARF debug symbols in LLVM IR
  -h, --help         Show this help message

Examples:
  scriptgo emit app.ts
  scriptgo emit "console.log(123)" --mode typed-ir
  scriptgo emit app.ts --mode llvm-ir -o app.ll`)
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
