package main

import (
	"fmt"
	"os"
	"os/exec"
)

// childExitCode maps a child process failure to this process's exit code.
// A child killed by a signal has no exit status (ExitCode is -1): report the
// signal on stderr and exit with 128+signal, the shell convention, instead
// of an unexplained 255.
func childExitCode(prefix string, exitErr *exec.ExitError) int {
	if code := exitErr.ExitCode(); code >= 0 {
		return code
	}
	fmt.Fprintf(os.Stderr, "%s: program terminated by %s\n", prefix, exitErr.ProcessState.String())
	return signalExitCode(exitErr.ProcessState)
}
