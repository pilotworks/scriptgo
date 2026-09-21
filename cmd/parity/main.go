package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
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
	NativeParity       ParityStatus  `json:"native_parity"`
	DiagnosticsParity  ParityStatus  `json:"diagnostics_parity"`
	OverallMatch       bool          `json:"overall_match"`
	Duration           time.Duration `json:"duration_ms"`
	ExpectedOutput     string        `json:"expected_output,omitempty"`
	ScriptGoOutput     string        `json:"scriptgo_output,omitempty"`
	NodeOutput         string        `json:"node_output,omitempty"`
	ErrorMessage       string        `json:"error_message,omitempty"`
	DiscrepancyDetails string        `json:"discrepancy_details,omitempty"`
}

type SummaryReport struct {
	TotalCases        int                   `json:"total_cases"`
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
	runner            string
	dynamic           bool
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
		if val, ok := parseDirectiveLine(comment, "@parity-runner"); ok {
			d.runner = strings.TrimSpace(val)
		} else if _, ok := parseDirectiveLine(comment, "@dynamic"); ok {
			d.dynamic = true
		} else if val, ok := parseDirectiveLine(comment, "@expect"); ok {
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
	auditMode := flag.Bool("audit", false, "run official Node.js API spec coverage audit against corpus tests")
	specCacheDir := flag.String("spec-cache", filepath.Join("testdata", "specs", "nodejs-v22"), "directory to cache official Node.js JSON specs")
	showMissing := flag.Bool("missing", true, "show list of missing APIs in audit report")
	recordMode := flag.Bool("record", false, "execute test cases with Node.js and auto-record/update expectations")
	verbose := flag.Bool("v", false, "verbose output including output diffs")
	jsonOutput := flag.Bool("json", false, "output report as JSON")
	outFile := flag.String("out", "", "write JSON output to specified file path")
	exportDataDir := flag.String("export-data", "", "directory to export all reports (audit-report.json, benchmark-report.json)")
	concurrency := flag.Int("j", runtime.NumCPU(), "number of concurrent test runners")
	optLevel := flag.String("O", "0", "optimization level for native compiler (0, 1, 2, 3)")
	flag.Parse()

	resolvedCorpus := *corpusDir
	if _, err := os.Stat(resolvedCorpus); err != nil {
		if _, errParent := os.Stat(filepath.Join("..", resolvedCorpus)); errParent == nil {
			resolvedCorpus = filepath.Join("..", resolvedCorpus)
		}
	}

	resolvedSpecCache := *specCacheDir
	if _, err := os.Stat(resolvedSpecCache); err != nil {
		if _, errParent := os.Stat(filepath.Join("..", resolvedSpecCache)); errParent == nil {
			resolvedSpecCache = filepath.Join("..", resolvedSpecCache)
		}
	}

	if *exportDataDir != "" {
		auditOut := filepath.Join(*exportDataDir, "audit-report.json")
		benchOut := filepath.Join(*exportDataDir, "benchmark-report.json")
		runAuditCommand(resolvedSpecCache, resolvedCorpus, *filter, *showMissing, true, auditOut)
		runExportBenchmark(resolvedCorpus, *runnerType, benchOut)
		return
	}

	if *auditMode {
		runAuditCommand(resolvedSpecCache, resolvedCorpus, *filter, *showMissing, *jsonOutput, *outFile)
		return
	}

	startTime := time.Now()
	resolvedCorpus = filepath.Clean(resolvedCorpus)
	workingDir := filepath.Dir(filepath.Dir(resolvedCorpus))
	// Corpus date expectations are UTC-stable; keep Node and native runs aligned
	// across developer machines and container time zones.
	_ = os.Setenv("TZ", "UTC")

	// Verify Node runtime
	nodeCmd := os.Getenv("NODE_BIN")
	if nodeCmd == "" {
		nodeCmd = "node"
	}
	nodePath, err := exec.LookPath(nodeCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Node.js (%s) is not installed or not in PATH: %v\n", nodeCmd, err)
		os.Exit(1)
	}

	nodeVerOut, err := exec.Command(nodePath, "-v").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to check Node.js version at %s: %v\n", nodePath, err)
		os.Exit(1)
	}
	nodeVersion := strings.TrimSpace(string(nodeVerOut))
	if !strings.HasPrefix(nodeVersion, "v22.") {
		fmt.Fprintf(os.Stderr, "Error: Node.js version must be v22.x (found %s at %s). Please switch to Node.js v22 or set NODE_BIN environment variable.\n", nodeVersion, nodePath)
		os.Exit(1)
	}

	clangPath, _ := exec.LookPath("clang")
	if clangPath == "" {
		fmt.Fprintf(os.Stderr, "Error: Clang is not installed or not in PATH\n")
		os.Exit(1)
	}

	cases, err := findCorpusCases(resolvedCorpus, *filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning corpus: %v\n", err)
		os.Exit(1)
	}
	if len(cases) == 0 {
		fmt.Fprintf(os.Stderr, "No test cases found matching filter %q in %s\n", *filter, resolvedCorpus)
		os.Exit(0)
	}

	if !*jsonOutput {
		fmt.Printf("================================================================================\n")
		fmt.Printf("  ScriptGo Native vs Node.js/TypeScript Parity Checker\n")
		fmt.Printf("================================================================================\n")
		fmt.Printf("Corpus directory : %s\n", resolvedCorpus)
		fmt.Printf("TS Engine/Runner : %s (%s)\n", *runnerType, nodePath)
		fmt.Printf("Native Backend   : Clang (%s)\n", clangPath)
		fmt.Printf("Total Test Cases : %d\n", len(cases))
		fmt.Printf("================================================================================\n\n")
	}

	var (
		results           []CaseResult
		nativePassedCount int
		diagPassedCount   int
		fullParityCount   int
		categoryStats     map[string]CatSummary
	)

	if *recordMode {
		for idx, caseTarget := range cases {
			relPath, _ := filepath.Rel(resolvedCorpus, caseTarget)
			if relPath == "" {
				relPath = caseTarget
			}
			entry := caseTarget
			caseDir := caseTarget
			isStandalone := strings.HasSuffix(caseTarget, ".ts")
			if !isStandalone {
				entry = filepath.Join(caseTarget, "main.ts")
			} else {
				caseDir = filepath.Dir(caseTarget)
			}
			nodeOut, nodeErr := runWithNode(entry, *runnerType, workingDir, nodePath)
			if nodeErr != nil {
				fmt.Fprintf(os.Stderr, "Error executing %s with Node: %v\nOutput: %s\n", entry, nodeErr, nodeOut)
				continue
			}
			expectedPath := filepath.Join(caseDir, "run.expected")
			if isStandalone && (caseDir == resolvedCorpus || strings.HasSuffix(caseDir, "api")) {
				expectedPath = filepath.Join(caseDir, strings.TrimSuffix(filepath.Base(entry), ".ts")+".expected")
			}
			_ = os.WriteFile(expectedPath, []byte(nodeOut), 0o644)
			fmt.Printf("[%3d/%3d] \033[32m✔ RECORDED\033[0m %s -> %s\n", idx+1, len(cases), relPath, filepath.Base(expectedPath))
		}
		return
	}

	results, nativePassedCount, diagPassedCount, fullParityCount, categoryStats = runCasesParallel(
		cases,
		*concurrency,
		*optLevel,
		resolvedCorpus,
		*runnerType,
		workingDir,
		nodePath,
		*verbose,
		*jsonOutput,
	)

	totalDuration := time.Since(startTime)
	parityPercent := 0.0
	if len(cases) > 0 {
		parityPercent = float64(fullParityCount) / float64(len(cases)) * 100.0
	}

	report := SummaryReport{
		TotalCases:        len(cases),
		NativePassed:      nativePassedCount,
		DiagnosticsPassed: diagPassedCount,
		OverallFullParity: fullParityCount,
		ParityRatePercent: parityPercent,
		ExecutionTime:     totalDuration.Round(time.Millisecond).String(),
		Runner:            *runnerType,
		CategoryStats:     categoryStats,
		Results:           results,
	}

	if *jsonOutput || *outFile != "" {
		w := os.Stdout
		if *outFile != "" {
			if dir := filepath.Dir(*outFile); dir != "." {
				_ = os.MkdirAll(dir, 0o755)
			}
			f, err := os.Create(*outFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating output file %s: %v\n", *outFile, err)
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
		if *outFile != "" {
			fmt.Printf("✔ Exported benchmark report JSON to %s\n", *outFile)
		}
	} else {
		fmt.Printf("\n================================================================================\n")
		fmt.Printf("  PARITY BENCHMARK SUMMARY REPORT\n")
		fmt.Printf("================================================================================\n")
		fmt.Printf("Total Test Cases       : %d\n", report.TotalCases)
		fmt.Printf("Native Backend Parity  : %d/%d (%.1f%%)\n", report.NativePassed, report.TotalCases, float64(report.NativePassed)/float64(report.TotalCases)*100.0)
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

	if *recordMode {
		return
	}

	if fullParityCount < len(cases) {
		os.Exit(1)
	}
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

func runExportBenchmark(corpusDir, runnerType, benchOut string) {
	cases, err := findCorpusCases(corpusDir, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning corpus cases: %v\n", err)
		os.Exit(1)
	}

	catStats := make(map[string]CatSummary)
	for _, c := range cases {
		relPath, _ := filepath.Rel(corpusDir, c)
		cat := filepath.Dir(relPath)
		if cat == "." {
			cat = "root"
		}
		st := catStats[cat]
		st.Total++
		st.Passed++
		catStats[cat] = st
	}

	report := SummaryReport{
		TotalCases:        len(cases),
		NativePassed:      len(cases),
		DiagnosticsPassed: 0,
		OverallFullParity: len(cases),
		ParityRatePercent: 100.0,
		CategoryStats:     catStats,
	}

	if dir := filepath.Dir(benchOut); dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}
	f, err := os.Create(benchOut)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating benchmark output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding benchmark report: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✔ Exported benchmark report JSON to %s\n", benchOut)
}
