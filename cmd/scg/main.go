package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// 1. Look for sibling "scriptgo" in the same directory as this binary
	var scriptgoPath string
	if selfExe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(selfExe), "scriptgo")
		if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
			scriptgoPath = candidate
		}
	}
	// 2. Look in PATH
	if scriptgoPath == "" {
		if path, lookErr := exec.LookPath("scriptgo"); lookErr == nil {
			scriptgoPath = path
		}
	}
	if scriptgoPath == "" {
		fmt.Fprintf(os.Stderr, "scg: could not locate 'scriptgo' binary in PATH or executable directory\n")
		os.Exit(1)
	}

	cmd := exec.Command(scriptgoPath, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}
