//go:build unix

package main

import (
	"errors"
	"os/exec"
	"testing"
)

func TestChildExitCodeReportsSignals(t *testing.T) {
	err := exec.Command("sh", "-c", "kill -SEGV $$").Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected an exit error, got %v", err)
	}
	if code := childExitCode("test", exitErr); code != 128+11 {
		t.Fatalf("signal exit code = %d, want 139", code)
	}
	err = exec.Command("sh", "-c", "exit 3").Run()
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected an exit error, got %v", err)
	}
	if code := childExitCode("test", exitErr); code != 3 {
		t.Fatalf("exit code = %d, want 3", code)
	}
}
