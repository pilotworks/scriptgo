package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pilotworks/scriptgo/internal/compiler"
)

func main() {
	output := flag.String("o", "", "write generated output to this path")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: scriptgo [flags] entry.ts")
		flag.PrintDefaults()
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
