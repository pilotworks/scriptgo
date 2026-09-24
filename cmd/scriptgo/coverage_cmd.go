package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pilotworks/scriptgo/internal/compiler"
)

func handleCoverage(args []string) {
	fs := flag.NewFlagSet("coverage", flag.ContinueOnError)
	fs.Usage = printCoverageUsage
	eval := fs.String("e", "", "evaluate inline script string")
	format := fs.String("format", "summary", "coverage output format: summary, json")
	output := fs.String("o", "", "write output to this path (default: stdout)")
	verbose := fs.Bool("v", false, "print analysis stages to stderr")
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
		printCoverageUsage()
		os.Exit(2)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "scriptgo: analyzing coverage for %s\n", entryPath)
	}
	options := compiler.BuildOptions{Dynamic: *dynamic}
	var result string
	var err error
	switch *format {
	case "summary":
		result, err = compiler.CoverageSummary(entryPath, options)
	case "json":
		result, err = compiler.CoverageReportJSON(entryPath, options)
	default:
		fmt.Fprintf(os.Stderr, "scriptgo: unsupported coverage format %q (supported: summary, json)\n", *format)
		os.Exit(2)
	}
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
