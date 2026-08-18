package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pilotworks/scriptgo/internal/compiler"
)

func main() {
	output := flag.String("o", "", "write generated output to this path")
	emit := flag.String("emit", "llvm-ir", "output mode: typed-ir, llvm-ir, exe, or run")
	verbose := flag.Bool("v", false, "print compilation stages to stderr")
	target := flag.String("target", "native", "native target triple, or native for the host")
	debug := flag.Bool("debug", false, "include native debug metadata")
	sanitize := flag.String("sanitize", "", "enable clang sanitizers (comma-separated: address,undefined,leak)")
	flag.Parse()
	options := compiler.BuildOptions{Target: *target, Debug: *debug, Sanitizers: splitList(*sanitize)}

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: scriptgo [flags] entry.ts")
		flag.PrintDefaults()
		os.Exit(2)
	}

	if *emit == "exe" {
		if *output == "" {
			fmt.Fprintln(os.Stderr, "scriptgo: -o is required with -emit exe")
			os.Exit(2)
		}
		if *verbose {
			fmt.Fprintf(os.Stderr, "scriptgo: build %s -> %s\n", flag.Arg(0), *output)
		}
		if err := compiler.BuildWithOptions(flag.Arg(0), *output, options); err != nil {
			fmt.Fprintln(os.Stderr, "scriptgo:", err)
			os.Exit(1)
		}
		return
	}
	if *emit == "run" {
		if *verbose {
			fmt.Fprintf(os.Stderr, "scriptgo: interpret %s\n", flag.Arg(0))
		}
		result, err := compiler.Run(flag.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, "scriptgo:", err)
			os.Exit(1)
		}
		fmt.Print(result)
		return
	}
	if *emit == "typed-ir" {
		if *verbose {
			fmt.Fprintf(os.Stderr, "scriptgo: lower and dump typed IR for %s\n", flag.Arg(0))
		}
		result, err := compiler.DumpIR(flag.Arg(0))
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
		return
	}
	if *emit != "llvm-ir" {
		fmt.Fprintln(os.Stderr, "scriptgo: unsupported -emit mode:", *emit)
		os.Exit(2)
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "scriptgo: lower and emit LLVM IR for %s\n", flag.Arg(0))
	}
	result, err := compiler.CompileWithOptions(flag.Arg(0), options)
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
