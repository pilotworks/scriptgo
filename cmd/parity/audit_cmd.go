package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pilotworks/scriptgo/internal/audit"
)

func runAuditCommand(specCacheDir, corpusDir, filter string, showMissing, jsonOutput bool, outFile string) {
	var targetModules []string
	if filter != "" {
		// Clean up filter in case user passed "api/timers" or "node:timers" or "timers.ts"
		clean := strings.TrimPrefix(filter, "api/")
		clean = strings.TrimPrefix(clean, "node:")
		clean = strings.TrimSuffix(clean, ".ts")
		targetModules = []string{clean}
	} else {
		targetModules = audit.StandardNodeModules
	}

	report, err := audit.AuditAllModules(specCacheDir, corpusDir, targetModules)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error performing API spec audit: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput || outFile != "" {
		var w *os.File = os.Stdout
		if outFile != "" {
			if dir := filepath.Dir(outFile); dir != "." {
				_ = os.MkdirAll(dir, 0o755)
			}
			f, err := os.Create(outFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating output file %s: %v\n", outFile, err)
				os.Exit(1)
			}
			defer f.Close()
			w = f
		}

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON output: %v\n", err)
			os.Exit(1)
		}
		if outFile != "" {
			fmt.Printf("✔ Exported audit report JSON to %s\n", outFile)
		}
		return
	}

	output := audit.FormatTerminalTable(report, filter, showMissing)
	fmt.Print(output)
}
