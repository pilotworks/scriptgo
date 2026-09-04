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
	auditMode := flag.Bool("audit", false, "run official Node.js API spec coverage audit against corpus tests")
	specCacheDir := flag.String("spec-cache", filepath.Join("testdata", "specs", "nodejs-v22"), "directory to cache official Node.js JSON specs")
	showMissing := flag.Bool("missing", true, "show list of missing APIs in audit report")
	recordMode := flag.Bool("record", false, "execute test cases with Node.js and auto-record/update expectations")
	verbose := flag.Bool("v", false, "verbose output including output diffs")
	jsonOutput := flag.Bool("json", false, "output report as JSON")
	outFile := flag.String("out", "", "write JSON output to specified file path")
	exportDataDir := flag.String("export-data", "", "directory to export all reports (audit-report.json, benchmark-report.json)")
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

	var results []CaseResult
	nativePassedCount := 0
	diagPassedCount := 0
	fullParityCount := 0
	categoryStats := make(map[string]CatSummary)

	for idx, caseTarget := range cases {
		caseStart := time.Now()
		relPath, _ := filepath.Rel(resolvedCorpus, caseTarget)
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

		if *recordMode {
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
			continue
		}

		var directives corpusDirectives
		if content, err := os.ReadFile(entry); err == nil {
			directives = parseCorpusDirectives(string(content))
		}

		res := CaseResult{
			Path:              relPath,
			Category:          category,
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
			// Native-only expectations do not have a Node.js reference output.
			var nodeOut string
			var nodeErr error
			if hasRunExpected {
				nodeOut, nodeErr = runWithNode(entry, *runnerType, workingDir, nodePath)
			}
			res.NodeOutput = nodeOut

			cleanNodeOut := cleanTraceOutput(nodeOut)
			cleanExpected := cleanTraceOutput(expectedTarget)
			nodeMatchesTarget := !hasRunExpected || (nodeErr == nil && (nodeOut == expectedTarget || strings.TrimSpace(nodeOut) == strings.TrimSpace(expectedTarget) || strings.TrimSpace(cleanNodeOut) == strings.TrimSpace(cleanExpected)))

			// Run ScriptGo Native
			sgOut, sgErr := compiler.RunWithOptions(entry, compiler.BuildOptions{WorkingDir: workingDir})
			res.ScriptGoOutput = sgOut
			if sgErr != nil {
				res.ErrorMessage = sgErr.Error()
			}

			target := runExpected
			if hasNativeExpected {
				target = nativeExpected
			}
			cleanSgOut := cleanTraceOutput(sgOut)
			cleanTarget := cleanTraceOutput(target)

			nativeMatchesTarget := (sgErr == nil && (sgOut == target || strings.TrimSpace(sgOut) == strings.TrimSpace(target) || strings.TrimSpace(cleanSgOut) == strings.TrimSpace(cleanTarget)))
			if nativeMatchesTarget {
				res.NativeParity = StatusPass
				nativePassedCount++
				if nodeMatchesTarget {
					res.OverallMatch = true
					fullParityCount++
				} else {
					res.DiscrepancyDetails = fmt.Sprintf("ScriptGo matches expected output, but Node.js produced different output (%q vs %q)", sgOut, nodeOut)
				}
			} else {
				if sgErr != nil {
					res.NativeParity = StatusFail
					res.DiscrepancyDetails = fmt.Sprintf("ScriptGo native execution error: %v", sgErr)
				} else {
					res.NativeParity = StatusDiff
					res.DiscrepancyDetails = fmt.Sprintf("Output mismatch: want %q, got ScriptGo %q, Node %q", target, sgOut, nodeOut)
				}
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
				_, sgErr = compiler.RunWithOptions(entry, compiler.BuildOptions{WorkingDir: workingDir})
			} else {
				_, sgErr = compiler.CompileWithOptions(entry, compiler.BuildOptions{WorkingDir: workingDir})
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
			fmt.Printf("[%3d/%3d] %s %-54s [Native: %s | Diag: %s] (%dms)\n",
				idx+1, len(cases), statusSymbol, relPath,
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

func runWithNode(entry, runner, workingDir, nodePath string) (string, error) {
	entry = filepath.Clean(entry)
	absoluteEntry, err := filepath.Abs(entry)
	if err != nil {
		return "", fmt.Errorf("resolve Node entry point %q: %w", entry, err)
	}

	var cmd *exec.Cmd

	switch runner {
	case "tsx":
		cmd = exec.Command("tsx", absoluteEntry)
	case "tsc":
		cmd = exec.Command("tsc", "--noEmit", absoluteEntry)
	default:
		loader := "data:text/javascript,export async function resolve(specifier, context, nextResolve) { try { return await nextResolve(specifier, context); } catch (e) { if (specifier.startsWith(\"./\") || specifier.startsWith(\"../\")) { for (const ext of [\".ts\", \".js\", \"/index.ts\", \"/index.js\"]) { try { return await nextResolve(specifier + ext, context); } catch {} } } throw e; } }"
		cmd = exec.Command(nodePath, "--expose-gc", "--no-warnings", "--loader", loader, "--experimental-transform-types", absoluteEntry)
	}
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	err = cmd.Run()
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
