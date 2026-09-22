package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pilotworks/scriptgo/internal/pkgmgr"
)

func handleInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.Usage = printInitUsage
	yes := fs.Bool("y", false, "initialize with default settings")
	fs.BoolVar(yes, "yes", false, "initialize with default settings (long)")
	force := fs.Bool("f", false, "overwrite existing files")
	fs.BoolVar(force, "force", false, "overwrite existing files (long)")
	name := fs.String("name", "", "package name (default: directory name)")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	targetDir := "."
	if fs.NArg() > 0 {
		targetDir = fs.Arg(0)
	}

	result, err := pkgmgr.Init(pkgmgr.InitOptions{
		TargetDir: targetDir,
		Name:      *name,
		Force:     *force,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "scriptgo init: %v\n", err)
		os.Exit(1)
	}

	displayDir := targetDir
	if abs, err := filepath.Abs(targetDir); err == nil {
		displayDir = abs
	}

	fmt.Printf("Initialized ScriptGo project in %s:\n", displayDir)
	for _, f := range result.CreatedFiles {
		fmt.Printf("  + %s\n", f)
	}
	for _, f := range result.SkippedFiles {
		fmt.Printf("  ~ %s (already exists, skipped)\n", f)
	}

	fmt.Println("\nRun your project with:")
	if targetDir != "." {
		fmt.Printf("  cd %s\n", targetDir)
	}
	fmt.Println("  scriptgo run start")
	fmt.Println("  # or")
	fmt.Println("  scriptgo run index.ts")
}
