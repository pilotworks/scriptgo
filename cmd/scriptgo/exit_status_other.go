//go:build !unix

package main

import "os"

func signalExitCode(*os.ProcessState) int { return 1 }
