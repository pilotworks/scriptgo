package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/pilotworks/scriptgo/internal/pkgmgr"
)

func handleTask(args []string) {
	var cmdArgs, passThroughArgs []string
	dashDashIdx := -1
	for i, a := range args {
		if a == "--" {
			dashDashIdx = i
			break
		}
	}
	if dashDashIdx >= 0 {
		cmdArgs = args[:dashDashIdx]
		passThroughArgs = args[dashDashIdx+1:]
	} else {
		cmdArgs = args
	}

	fs := flag.NewFlagSet("task", flag.ContinueOnError)
	fs.Usage = printTaskUsage
	project := fs.String("project", ".", "project directory containing package.json")
	manifest := fs.String("manifest", "", "package manifest path (default: <project>/package.json)")

	if err := fs.Parse(cmdArgs); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	positionals := fs.Args()
	if len(positionals) == 0 {
		exitCode := listAvailableScripts(*project, *manifest)
		os.Exit(exitCode)
	}

	scriptName := positionals[0]
	extraArgs := append(positionals[1:], passThroughArgs...)

	manifestPath := *manifest
	if manifestPath == "" && *project != "." {
		manifestPath = filepath.Join(*project, "package.json")
	}

	task, err := pkgmgr.ResolveTask(*project, manifestPath, scriptName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scriptgo task: %v\n", err)
		os.Exit(1)
	}

	exitCode := executeTask(task, extraArgs)
	os.Exit(exitCode)
}

func listAvailableScripts(projectRoot, manifestPath string) int {
	path := manifestPath
	var err error
	if path == "" {
		path, err = pkgmgr.FindPackageManifest(projectRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scriptgo task: %v\n", err)
			return 1
		}
	}
	scripts, err := pkgmgr.ListScripts(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scriptgo task: %v\n", err)
		return 1
	}
	if len(scripts) == 0 {
		fmt.Printf("No scripts found in %s\n", path)
		return 0
	}
	names := make([]string, 0, len(scripts))
	for name := range scripts {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Printf("Scripts available in %s:\n", path)
	for _, name := range names {
		fmt.Printf("  %s\n    %s\n", name, scripts[name])
	}
	return 0
}

func executeTask(task *pkgmgr.TaskDefinition, extraArgs []string) int {
	fullCmd := formatShellCommand(task.Command, extraArgs)
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		comspec := os.Getenv("COMSPEC")
		if comspec == "" {
			comspec = "cmd.exe"
		}
		cmd = exec.Command(comspec, "/d", "/s", "/c", fullCmd)
	} else {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		cmd = exec.Command(shell, "-c", fullCmd)
	}

	cmd.Dir = task.ProjectRoot
	env := pkgmgr.BuildTaskEnv(task.ProjectRoot, task.Name, os.Environ())
	if selfExe, err := os.Executable(); err == nil {
		selfDir := filepath.Dir(selfExe)
		var prependDirs []string
		prependDirs = append(prependDirs, selfDir)
		if realExe, err := filepath.EvalSymlinks(selfExe); err == nil {
			realDir := filepath.Dir(realExe)
			if realDir != selfDir {
				prependDirs = append(prependDirs, realDir)
			}
		}
		pathPrefix := strings.Join(prependDirs, string(os.PathListSeparator)) + string(os.PathListSeparator)
		for i, e := range env {
			if strings.HasPrefix(e, "PATH=") || strings.HasPrefix(e, "Path=") {
				env[i] = e[:5] + pathPrefix + e[5:]
				break
			}
		}
	}
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "scriptgo task: %v\n", err)
		return 1
	}
	return 0
}

func formatShellCommand(command string, args []string) string {
	if len(args) == 0 {
		return command
	}
	var b strings.Builder
	b.WriteString(command)
	for _, a := range args {
		b.WriteString(" ")
		b.WriteString(shellQuote(a))
	}
	return b.String()
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	needsQuote := false
	for _, c := range s {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') &&
			c != '_' && c != '-' && c != '.' && c != '/' && c != ':' && c != '@' {
			needsQuote = true
			break
		}
	}
	if !needsQuote {
		return s
	}
	if runtime.GOOS == "windows" {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
