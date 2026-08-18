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
		if err := compiler.Build(flag.Arg(0), *output); err != nil {
			fmt.Fprintln(os.Stderr, "scriptgo:", err)
			os.Exit(1)
		}
		return
	}
	if *emit == "run" {
		result, err := compiler.Run(flag.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, "scriptgo:", err)
			os.Exit(1)
		}
		fmt.Print(result)
		return
	}
	if *emit == "typed-ir" {
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
