package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pilotworks/scriptgo/internal/compiler"
)

func main() {
	output := flag.String("o", "", "write generated output to this path")
	emit := flag.String("emit", "llvm-ir", "output mode: typed-ir, llvm-ir, exe, or run")
	verbose := flag.Bool("v", false, "print compilation stages to stderr")
	flag.Parse()

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
		if err := compiler.Build(flag.Arg(0), *output); err != nil {
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
	result, err := compiler.Compile(flag.Arg(0))
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
