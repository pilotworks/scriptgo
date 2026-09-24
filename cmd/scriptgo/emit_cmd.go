package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pilotworks/scriptgo/internal/compiler"
)

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
	optLevel := fs.String("O", "", "optimization level (0, 1, 2, 3, s, z, fast)")
	release := fs.Bool("release", false, "build with release optimizations")
	dynamic := registerDynamicFlag(fs)
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
		OptLevel:         *optLevel,
		Release:          *release,
		WarnRuntimeCasts: *warnRuntimeCasts,
		StrictCasts:      *strictCasts,
		Dynamic:          *dynamic,
	}
	var result string
	var err error

	switch *mode {
	case "typed-ir":
		if *verbose {
			fmt.Fprintf(os.Stderr, "scriptgo: emitting typed IR for %s\n", entryPath)
		}
		result, err = compiler.DumpIRWithOptions(entryPath, options)
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
