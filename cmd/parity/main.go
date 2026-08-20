package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pilotworks/scriptgo/internal/compiler"
)

type ParityStatus string

const (
	StatusPass ParityStatus = "PASS"
	StatusDiff ParityStatus = "DIFF"
	StatusFail ParityStatus = "FAIL"
	StatusSkip ParityStatus = "SKIP"
)

type CaseResult struct {
	Path               string        `json:"path"`
	Category           string        `json:"category"`
	ExpectationType    string        `json:"expectation_type"`
	InterpreterParity  ParityStatus  `json:"interpreter_parity"`
	NativeParity       ParityStatus  `json:"native_parity"`
	DiagnosticsParity  ParityStatus  `json:"diagnostics_parity"`
	OverallMatch       bool          `json:"overall_match"`
	Duration           time.Duration `json:"duration_ms"`
	ExpectedOutput     string        `json:"expected_output,omitempty"`
	ScriptGoOutput     string        `json:"scriptgo_output,omitempty"`
	NativeOutput       string        `json:"native_output,omitempty"`
	NodeOutput         string        `json:"node_output,omitempty"`
	ErrorMessage       string        `json:"error_message,omitempty"`
	DiscrepancyDetails string        `json:"discrepancy_details,omitempty"`
}

type SummaryReport struct {
	TotalCases        int                   `json:"total_cases"`
	InterpreterPassed int                   `json:"interpreter_passed"`
	NativePassed      int                   `json:"native_passed"`
	DiagnosticsPassed int                   `json:"diagnostics_passed"`
	OverallFullParity int                   `json:"overall_full_parity"`
	ParityRatePercent float64               `json:"parity_rate_percent"`
	ExecutionTime     string                `json:"execution_time"`
	Runner            string                `json:"runner"`
	CategoryStats     map[string]CatSummary `json:"category_stats"`
	Results           []CaseResult          `json:"results"`
}

type CatSummary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

func main() {
	corpusDir := flag.String("corpus", filepath.Join("internal", "compiler", "testdata", "corpus"), "path to corpus test directory")
	filter := flag.String("filter", "", "filter test cases by substring in path")
	runnerType := flag.String("runner", "node", "typescript runner: node, tsx, tsc")
	checkNative := flag.Bool("native", true, "verify ScriptGo native binary compilation and execution")
	strictNative := flag.Bool("strict-native", false, "require native to succeed even if case only specifies run.expected")
	verbose := flag.Bool("v", false, "verbose output including output diffs")
	jsonOutput := flag.Bool("json", false, "output report as JSON")
	flag.Parse()

	startTime := time.Now()

	// Verify Node runtime
	nodePath, err := exec.LookPath("node")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Node.js is not installed or not in PATH: %v\n", err)
		os.Exit(1)
	}

	clangPath, _ := exec.LookPath("clang")
	hasClang := clangPath != ""
	if *checkNative && !hasClang {
		if !*jsonOutput {
			fmt.Println("[WARN] Clang is not installed; skipping native binary execution.")
		}
	}

	cases, err := findCorpusCases(*corpusDir, *filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning corpus: %v\n", err)
		os.Exit(1)
	}
	if len(cases) == 0 {
		fmt.Fprintf(os.Stderr, "No test cases found matching filter %q in %s\n", *filter, *corpusDir)
		os.Exit(0)
	}

	if !*jsonOutput {
		fmt.Printf("================================================================================\n")
		fmt.Printf("  ScriptGo vs Node.js/TypeScript Parity Checker\n")
		fmt.Printf("================================================================================\n")
		fmt.Printf("Corpus directory : %s\n", *corpusDir)
		fmt.Printf("TS Engine/Runner : %s (%s)\n", *runnerType, nodePath)
		fmt.Printf("Native Backend   : %v (Clang: %t)\n", *checkNative, hasClang)
		fmt.Printf("Total Test Cases : %d\n", len(cases))
		fmt.Printf("================================================================================\n\n")
	}

	var results []CaseResult
	interpPassedCount := 0
	nativePassedCount := 0
	diagPassedCount := 0
	fullParityCount := 0
	categoryStats := make(map[string]CatSummary)

	for idx, casePath := range cases {
		caseStart := time.Now()
		relPath, _ := filepath.Rel(*corpusDir, casePath)
		if relPath == "" {
			relPath = casePath
		}
		category := filepath.Dir(relPath)
		if category == "." {
			category = "root"
		}

		entry := filepath.Join(casePath, "main.ts")
		res := CaseResult{
			Path:              relPath,
			Category:          category,
			InterpreterParity: StatusSkip,
			NativeParity:      StatusSkip,
			DiagnosticsParity: StatusSkip,
		}

		runExpected, hasRunExpected := readCorpusFile(casePath, "run.expected")
		nativeExpected, hasNativeExpected := readCorpusFile(casePath, "native.expected")
		runErr, hasRunErr := readCorpusFile(casePath, "run.err")
		checkErr, hasCheckErr := readCorpusFile(casePath, "check.err")
		buildErr, hasBuildErr := readCorpusFile(casePath, "build.err")

		expectedTarget := ""
		if hasRunExpected {
			res.ExpectationType = "run.expected"
			expectedTarget = runExpected
		} else if hasNativeExpected {
			res.ExpectationType = "native.expected"
			expectedTarget = nativeExpected
		} else if hasRunErr {
			res.ExpectationType = "run.err"
			expectedTarget = runErr
		} else if hasCheckErr {
			res.ExpectationType = "check.err"
			expectedTarget = checkErr
		} else if hasBuildErr {
			res.ExpectationType = "build.err"
			expectedTarget = buildErr
		}
		res.ExpectedOutput = expectedTarget

		// 1. Runtime Cases (run.expected / native.expected)
		if hasRunExpected || hasNativeExpected {
			// Run Node.js / TypeScript
			nodeOut, nodeErr := runWithNode(entry, *runnerType)
			res.NodeOutput = nodeOut

			// Run ScriptGo Interpreter
			sgOut, sgErr := compiler.Run(entry)
			res.ScriptGoOutput = sgOut
			if sgErr != nil {
				res.ErrorMessage = sgErr.Error()
			}

			// Check Interpreter Parity
			interpMatchesTarget := (sgErr == nil && (sgOut == runExpected || strings.TrimSpace(sgOut) == strings.TrimSpace(runExpected)))
			nodeMatchesTarget := (nodeErr == nil && (nodeOut == expectedTarget || strings.TrimSpace(nodeOut) == strings.TrimSpace(expectedTarget)))
			interpMatchesNode := (sgErr == nil && nodeErr == nil && (sgOut == nodeOut || strings.TrimSpace(sgOut) == strings.TrimSpace(nodeOut)))

			if interpMatchesTarget && (nodeMatchesTarget || interpMatchesNode) {
				res.InterpreterParity = StatusPass
				interpPassedCount++
			} else if !nodeMatchesTarget && interpMatchesTarget {
				res.InterpreterParity = StatusDiff
				res.DiscrepancyDetails = fmt.Sprintf("ScriptGo matches run.expected, but Node.js produced different output (%q vs %q)", sgOut, nodeOut)
			} else {
				res.InterpreterParity = StatusFail
				if sgErr != nil {
					res.DiscrepancyDetails = fmt.Sprintf("ScriptGo interpreter error: %v", sgErr)
				} else {
					res.DiscrepancyDetails = fmt.Sprintf("Output mismatch: want %q, got ScriptGo %q, Node %q", expectedTarget, sgOut, nodeOut)
				}
			}

			// Check Native Parity
			if *checkNative && hasClang {
				tmpDir, err := os.MkdirTemp("", "scriptgo-parity-")
				if err == nil {
					binPath := filepath.Join(tmpDir, "main")
					if err := compiler.Build(entry, binPath); err == nil {
						cmd := exec.Command(binPath)
						var stdout bytes.Buffer
						cmd.Stdout = &stdout
						cmd.Stderr = &stdout
						if err := cmd.Run(); err == nil {
							res.NativeOutput = stdout.String()
							target := runExpected
							if hasNativeExpected {
								target = nativeExpected
							}
							if res.NativeOutput == target || strings.TrimSpace(res.NativeOutput) == strings.TrimSpace(target) {
								res.NativeParity = StatusPass
								nativePassedCount++
							} else {
								res.NativeParity = StatusDiff
							}
						} else {
							res.NativeParity = StatusFail
						}
					} else {
						if hasNativeExpected || *strictNative {
							res.NativeParity = StatusFail
						} else {
							res.NativeParity = StatusSkip
						}
					}
					os.RemoveAll(tmpDir)
				}
			} else {
				res.NativeParity = StatusSkip
			}

			// Overall Match
			nativeOK := (res.NativeParity == StatusPass || res.NativeParity == StatusSkip)
			if res.InterpreterParity == StatusPass && nativeOK {
				res.OverallMatch = true
				fullParityCount++
			}
		} else if hasCheckErr || hasBuildErr || hasRunErr {
			// 2. Diagnostic & Error Cases
			errExp := checkErr
			if hasBuildErr {
				errExp = buildErr
			} else if hasRunErr {
				errExp = runErr
			}

			var sgErr error
			if hasRunErr {
				_, sgErr = compiler.Run(entry)
			} else {
				_, sgErr = compiler.Compile(entry)
			}

			if sgErr != nil && strings.Contains(sgErr.Error(), strings.TrimSpace(errExp)) {
				res.DiagnosticsParity = StatusPass
				res.OverallMatch = true
				diagPassedCount++
				fullParityCount++
			} else {
				res.DiagnosticsParity = StatusFail
				if sgErr != nil {
					res.DiscrepancyDetails = fmt.Sprintf("Diagnostic mismatch: want substring %q, got %q", strings.TrimSpace(errExp), sgErr.Error())
				} else {
					res.DiscrepancyDetails = fmt.Sprintf("Expected error %q, but compilation succeeded", strings.TrimSpace(errExp))
				}
			}
		}

		res.Duration = time.Since(caseStart)
		results = append(results, res)

		cat := categoryStats[category]
		cat.Total++
		if res.OverallMatch {
			cat.Passed++
		} else {
			cat.Failed++
		}
		categoryStats[category] = cat

		if !*jsonOutput {
			statusSymbol := "\033[32m✔ PASS\033[0m"
			if !res.OverallMatch {
				statusSymbol = "\033[31m✖ FAIL\033[0m"
			}
			fmt.Printf("[%3d/%3d] %s %-52s [Interp: %s | Native: %s | Diag: %s] (%dms)\n",
				idx+1, len(cases), statusSymbol, relPath,
				formatStatus(res.InterpreterParity),
				formatStatus(res.NativeParity),
				formatStatus(res.DiagnosticsParity),
				res.Duration.Milliseconds())

			if *verbose && !res.OverallMatch {
				if res.DiscrepancyDetails != "" {
					fmt.Printf("        \033[33mNote:\033[0m %s\n", res.DiscrepancyDetails)
				}
				if res.ExpectedOutput != "" {
					fmt.Printf("        Expected: %q\n", truncateString(res.ExpectedOutput, 60))
				}
				if res.ScriptGoOutput != "" {
					fmt.Printf("        ScriptGo: %q\n", truncateString(res.ScriptGoOutput, 60))
				}
				if res.NodeOutput != "" {
					fmt.Printf("        Node.js : %q\n", truncateString(res.NodeOutput, 60))
				}
				if res.NativeOutput != "" {
					fmt.Printf("        Native  : %q\n", truncateString(res.NativeOutput, 60))
				}
			}
		}
	}

	totalDuration := time.Since(startTime)
	parityRate := 0.0
	if len(cases) > 0 {
		parityRate = float64(fullParityCount) / float64(len(cases)) * 100.0
	}

	if *jsonOutput {
		report := SummaryReport{
			TotalCases:        len(cases),
			InterpreterPassed: interpPassedCount,
			NativePassed:      nativePassedCount,
			DiagnosticsPassed: diagPassedCount,
			OverallFullParity: fullParityCount,
			ParityRatePercent: parityRate,
			ExecutionTime:     totalDuration.String(),
			Runner:            *runnerType,
			CategoryStats:     categoryStats,
			Results:           results,
		}
		data, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Printf("\n================================================================================\n")
	fmt.Printf("  Category Breakdown\n")
	fmt.Printf("================================================================================\n")
	var sortedCategories []string
	for cat := range categoryStats {
		sortedCategories = append(sortedCategories, cat)
	}
	sort.Strings(sortedCategories)
	for _, cat := range sortedCategories {
		stats := categoryStats[cat]
		pct := float64(stats.Passed) / float64(stats.Total) * 100.0
		bar := progressBar(stats.Passed, stats.Total)
		fmt.Printf("  %-25s %s %3d/%-3d (%5.1f%%)\n", cat, bar, stats.Passed, stats.Total, pct)
	}

	fmt.Printf("\n================================================================================\n")
	fmt.Printf("  Overall Parity Summary\n")
	fmt.Printf("================================================================================\n")
	fmt.Printf("Total Test Cases          : %d\n", len(cases))
	fmt.Printf("Full Parity Compatibility : \033[1;32m%d / %d (%.1f%%)\033[0m\n", fullParityCount, len(cases), parityRate)
	fmt.Printf("ScriptGo Interpreter OK   : %d\n", interpPassedCount)
	fmt.Printf("ScriptGo Native Binary OK : %d\n", nativePassedCount)
	fmt.Printf("Diagnostics Parity OK     : %d\n", diagPassedCount)
	fmt.Printf("Total Elapsed Time        : %s\n", totalDuration.Round(time.Millisecond))
	fmt.Printf("================================================================================\n")
}

func runWithNode(entry string, runnerType string) (string, error) {
	var cmd *exec.Cmd
	switch runnerType {
	case "tsx":
		cmd = exec.Command("npx", "-y", "tsx", entry)
	case "tsc":
		tmpDir, err := os.MkdirTemp("", "tsc-parity-")
		if err != nil {
			return "", err
		}
		defer os.RemoveAll(tmpDir)
		tscCmd := exec.Command("npx", "-y", "--package", "typescript", "tsc",
			"--target", "ES2022",
			"--module", "ESNext",
			"--moduleResolution", "bundler",
			"--strict", "true",
			"--outDir", tmpDir,
			entry)
		if out, err := tscCmd.CombinedOutput(); err != nil {
			return string(out), err
		}
		jsFile := filepath.Join(tmpDir, strings.TrimSuffix(filepath.Base(entry), ".ts")+".js")
		cmd = exec.Command("node", jsFile)
	case "node":
		fallthrough
	default:
		loader := "data:text/javascript,export async function resolve(specifier, context, nextResolve) { try { return await nextResolve(specifier, context); } catch (e) { if (specifier.startsWith(\"./\") || specifier.startsWith(\"../\")) { for (const ext of [\".ts\", \".js\", \"/index.ts\", \"/index.js\"]) { try { return await nextResolve(specifier + ext, context); } catch {} } } throw e; } }"
		cmd = exec.Command("node", "--no-warnings", "--loader", loader, "--experimental-transform-types", entry)
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	err := cmd.Run()
	return stdout.String(), err
}

func findCorpusCases(root, filter string) ([]string, error) {
	var cases []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Name() == "main.ts" {
			dir := filepath.Dir(path)
			if filter == "" || strings.Contains(filepath.ToSlash(dir), filter) {
				cases = append(cases, dir)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(cases)
	return cases, nil
}

func readCorpusFile(casePath, name string) (string, bool) {
	path := filepath.Join(casePath, name)
	contents, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return string(contents), true
}

func formatStatus(s ParityStatus) string {
	switch s {
	case StatusPass:
		return "\033[32mPASS\033[0m"
	case StatusDiff:
		return "\033[33mDIFF\033[0m"
	case StatusFail:
		return "\033[31mFAIL\033[0m"
	case StatusSkip:
		return "\033[90mSKIP\033[0m"
	default:
		return string(s)
	}
}

func progressBar(passed, total int) string {
	width := 20
	if total == 0 {
		return "[" + strings.Repeat(" ", width) + "]"
	}
	filled := (passed * width) / total
	empty := width - filled
	if filled > width {
		filled = width
		empty = 0
	}
	return "[" + strings.Repeat("=", filled) + strings.Repeat(" ", empty) + "]"
}

func truncateString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
