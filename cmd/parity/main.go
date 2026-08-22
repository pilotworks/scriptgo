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

type corpusDirectives struct {
	hasRunExpected    bool
	runExpected       string
	hasNativeExpected bool
	nativeExpected    string
	hasRunErr         bool
	runErr            string
	hasCheckErr       bool
	checkErr          string
	hasBuildErr       bool
	buildErr          string
}

func parseDirectiveLine(comment, prefix string) (string, bool) {
	if strings.HasPrefix(comment, prefix+":") {
		val := strings.TrimPrefix(comment, prefix+":")
		val = strings.TrimPrefix(val, " ")
		return val, true
	}
	if strings.HasPrefix(comment, prefix+" ") {
		val := strings.TrimPrefix(comment, prefix+" ")
		return val, true
	}
	if comment == prefix {
		return "", true
	}
	return "", false
}

func parseCorpusDirectives(content string) corpusDirectives {
	var d corpusDirectives
	var runLines []string
	var nativeLines []string
	var runErrLines []string
	var checkErrLines []string
	var buildErrLines []string

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmedLeading := strings.TrimLeft(line, " \t")
		if !strings.HasPrefix(trimmedLeading, "//") {
			continue
		}
		comment := strings.TrimLeft(strings.TrimPrefix(trimmedLeading, "//"), " \t")
		if val, ok := parseDirectiveLine(comment, "@expect"); ok {
			d.hasRunExpected = true
			runLines = append(runLines, val)
		} else if val, ok := parseDirectiveLine(comment, "@run.expected"); ok {
			d.hasRunExpected = true
			runLines = append(runLines, val)
		} else if val, ok := parseDirectiveLine(comment, "@native.expected"); ok {
			d.hasNativeExpected = true
			nativeLines = append(nativeLines, val)
		} else if val, ok := parseDirectiveLine(comment, "@run.err"); ok {
			d.hasRunErr = true
			runErrLines = append(runErrLines, val)
		} else if val, ok := parseDirectiveLine(comment, "@check.err"); ok {
			d.hasCheckErr = true
			checkErrLines = append(checkErrLines, val)
		} else if val, ok := parseDirectiveLine(comment, "@build.err"); ok {
			d.hasBuildErr = true
			buildErrLines = append(buildErrLines, val)
		}
	}

	if d.hasRunExpected {
		d.runExpected = strings.Join(runLines, "\n") + "\n"
	}
	if d.hasNativeExpected {
		d.nativeExpected = strings.Join(nativeLines, "\n") + "\n"
	}
	if d.hasRunErr {
		d.runErr = strings.Join(runErrLines, "\n")
	}
	if d.hasCheckErr {
		d.checkErr = strings.Join(checkErrLines, "\n")
	}
	if d.hasBuildErr {
		d.buildErr = strings.Join(buildErrLines, "\n")
	}

	return d
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

	for idx, caseTarget := range cases {
		caseStart := time.Now()
		relPath, _ := filepath.Rel(*corpusDir, caseTarget)
		if relPath == "" {
			relPath = caseTarget
		}
		category := filepath.Dir(relPath)
		if category == "." {
			category = "root"
		}

		entry := caseTarget
		caseDir := caseTarget
		isStandalone := strings.HasSuffix(caseTarget, ".ts")
		if !isStandalone {
			entry = filepath.Join(caseTarget, "main.ts")
		} else {
			caseDir = filepath.Dir(caseTarget)
		}

		var directives corpusDirectives
		if content, err := os.ReadFile(entry); err == nil {
			directives = parseCorpusDirectives(string(content))
		}

		res := CaseResult{
			Path:              relPath,
			Category:          category,
			InterpreterParity: StatusSkip,
			NativeParity:      StatusSkip,
			DiagnosticsParity: StatusSkip,
		}

		runExpected := directives.runExpected
		hasRunExpected := directives.hasRunExpected
		if !hasRunExpected && !isStandalone {
			runExpected, hasRunExpected = readCorpusFile(caseDir, "run.expected")
		}

		nativeExpected := directives.nativeExpected
		hasNativeExpected := directives.hasNativeExpected
		if !hasNativeExpected && !isStandalone {
			nativeExpected, hasNativeExpected = readCorpusFile(caseDir, "native.expected")
		}

		runErr := directives.runErr
		hasRunErr := directives.hasRunErr
		if !hasRunErr && !isStandalone {
			runErr, hasRunErr = readCorpusFile(caseDir, "run.err")
		}

		checkErr := directives.checkErr
		hasCheckErr := directives.hasCheckErr
		if !hasCheckErr && !isStandalone {
			checkErr, hasCheckErr = readCorpusFile(caseDir, "check.err")
		}

		buildErr := directives.buildErr
		hasBuildErr := directives.hasBuildErr
		if !hasBuildErr && !isStandalone {
			buildErr, hasBuildErr = readCorpusFile(caseDir, "build.err")
		}

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
			cleanNodeOut := cleanTraceOutput(nodeOut)
			cleanSgOut := cleanTraceOutput(sgOut)
			cleanExpected := cleanTraceOutput(expectedTarget)

			interpMatchesTarget := (sgErr == nil && (sgOut == runExpected || strings.TrimSpace(sgOut) == strings.TrimSpace(runExpected) || strings.TrimSpace(cleanSgOut) == strings.TrimSpace(cleanExpected)))
			nodeMatchesTarget := (nodeErr == nil && (nodeOut == expectedTarget || strings.TrimSpace(nodeOut) == strings.TrimSpace(expectedTarget) || strings.TrimSpace(cleanNodeOut) == strings.TrimSpace(cleanExpected)))
			interpMatchesNode := (sgErr == nil && nodeErr == nil && (sgOut == nodeOut || strings.TrimSpace(sgOut) == strings.TrimSpace(nodeOut) || strings.TrimSpace(cleanSgOut) == strings.TrimSpace(cleanNodeOut)))

			if interpMatchesTarget {
				res.InterpreterParity = StatusPass
				interpPassedCount++
				if !nodeMatchesTarget && !interpMatchesNode {
					res.DiscrepancyDetails = fmt.Sprintf("ScriptGo matches run.expected, but Node.js produced different output (%q vs %q)", sgOut, nodeOut)
				}
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
			nativeOK := (res.NativeParity == StatusPass || res.NativeParity == StatusSkip || (!hasNativeExpected && !*strictNative))
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
					fmt.Printf("          \033[31mDetails:\033[0m %s\n", res.DiscrepancyDetails)
				}
				if res.ErrorMessage != "" {
					fmt.Printf("          \033[31mError:\033[0m %s\n", res.ErrorMessage)
				}
			}
		}
	}

	totalDuration := time.Since(startTime)
	parityPercent := 0.0
	if len(cases) > 0 {
		parityPercent = float64(fullParityCount) / float64(len(cases)) * 100.0
	}

	report := SummaryReport{
		TotalCases:        len(cases),
		InterpreterPassed: interpPassedCount,
		NativePassed:      nativePassedCount,
		DiagnosticsPassed: diagPassedCount,
		OverallFullParity: fullParityCount,
		ParityRatePercent: parityPercent,
		ExecutionTime:     totalDuration.Round(time.Millisecond).String(),
		Runner:            *runnerType,
		CategoryStats:     categoryStats,
		Results:           results,
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON output: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("\n================================================================================\n")
		fmt.Printf("  PARITY BENCHMARK SUMMARY REPORT\n")
		fmt.Printf("================================================================================\n")
		fmt.Printf("Total Test Cases       : %d\n", report.TotalCases)
		fmt.Printf("Interpreter Parity     : %d/%d (%.1f%%)\n", report.InterpreterPassed, report.TotalCases, float64(report.InterpreterPassed)/float64(report.TotalCases)*100.0)
		if *checkNative && hasClang {
			fmt.Printf("Native Backend Parity  : %d/%d\n", report.NativePassed, report.TotalCases)
		}
		if report.DiagnosticsPassed > 0 {
			fmt.Printf("Diagnostic Parity      : %d/%d\n", report.DiagnosticsPassed, report.TotalCases)
		}
		fmt.Printf("Overall Full Parity    : %d/%d (%.1f%%)\n", report.OverallFullParity, report.TotalCases, report.ParityRatePercent)
		fmt.Printf("Total Time Elapsed     : %s\n", report.ExecutionTime)
		fmt.Printf("================================================================================\n\n")

		fmt.Printf("Category Breakdown:\n")
		var catNames []string
		for name := range categoryStats {
			catNames = append(catNames, name)
		}
		sort.Strings(catNames)

		for _, name := range catNames {
			cat := categoryStats[name]
			pct := 0.0
			if cat.Total > 0 {
				pct = float64(cat.Passed) / float64(cat.Total) * 100.0
			}
			bar := progressBar(cat.Passed, cat.Total)
			fmt.Printf("  %-32s %s %3d/%-3d (%5.1f%%)\n", name, bar, cat.Passed, cat.Total, pct)
		}
		fmt.Println()
	}

	if fullParityCount < len(cases) {
		os.Exit(1)
	}
}

func runWithNode(entry, runner string) (string, error) {
	var cmd *exec.Cmd

	switch runner {
	case "tsx":
		cmd = exec.Command("tsx", entry)
	case "tsc":
		cmd = exec.Command("tsc", "--noEmit", entry)
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
	dirsWithMain := make(map[string]bool)

	// First pass: identify directories containing main.ts
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && entry.Name() == "main.ts" {
			dirsWithMain[filepath.Dir(path)] = true
		}
		return nil
	})

	// Second pass: collect main.ts directories and standalone .ts files
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if dirsWithMain[path] {
				if filter == "" || strings.Contains(filepath.ToSlash(path), filter) {
					cases = append(cases, path)
				}
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(entry.Name(), ".ts") && entry.Name() != "main.ts" {
			dir := filepath.Dir(path)
			if !dirsWithMain[dir] {
				if filter == "" || strings.Contains(filepath.ToSlash(path), filter) {
					cases = append(cases, path)
				}
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

func cleanTraceOutput(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "debugger") || strings.HasPrefix(trimmed, "Trace:") {
			continue
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}
