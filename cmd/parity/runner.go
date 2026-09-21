package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pilotworks/scriptgo/internal/compiler"
)

type workerTask struct {
	idx        int
	caseTarget string
}

type workerResult struct {
	idx int
	res CaseResult
}

func runCasesParallel(
	cases []string,
	concurrency int,
	optLevel string,
	resolvedCorpus string,
	runnerType string,
	workingDir string,
	nodePath string,
	verbose bool,
	jsonOutput bool,
) ([]CaseResult, int, int, int, map[string]CatSummary) {
	if concurrency <= 0 {
		concurrency = runtime.NumCPU()
	}
	if concurrency > len(cases) {
		concurrency = len(cases)
	}

	tasksChan := make(chan workerTask, len(cases))
	resultsChan := make(chan workerResult, len(cases))

	var wg sync.WaitGroup
	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasksChan {
				res := runCase(task.caseTarget, resolvedCorpus, runnerType, workingDir, nodePath, optLevel)
				resultsChan <- workerResult{idx: task.idx, res: res}
			}
		}()
	}

	for idx, caseTarget := range cases {
		tasksChan <- workerTask{idx: idx, caseTarget: caseTarget}
	}
	close(tasksChan)

	// Close resultsChan when all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	results := make([]CaseResult, len(cases))
	completed := make(map[int]CaseResult)
	nextToPrint := 0

	nativePassedCount := 0
	diagPassedCount := 0
	fullParityCount := 0
	categoryStats := make(map[string]CatSummary)

	for wr := range resultsChan {
		completed[wr.idx] = wr.res

		// Print in deterministic order as soon as contiguous sequence is available
		for {
			res, ok := completed[nextToPrint]
			if !ok {
				break
			}
			delete(completed, nextToPrint)
			results[nextToPrint] = res

			if res.NativeParity == StatusPass {
				nativePassedCount++
			}
			if res.DiagnosticsParity == StatusPass {
				diagPassedCount++
			}
			if res.OverallMatch {
				fullParityCount++
			}

			cat := categoryStats[res.Category]
			cat.Total++
			if res.OverallMatch {
				cat.Passed++
			} else {
				cat.Failed++
			}
			categoryStats[res.Category] = cat

			if !jsonOutput {
				statusSymbol := "\033[32m✔ PASS\033[0m"
				if !res.OverallMatch {
					statusSymbol = "\033[31m✖ FAIL\033[0m"
				}
				fmt.Printf("[%3d/%3d] %s %-54s [Native: %s | Diag: %s] (%dms)\n",
					nextToPrint+1, len(cases), statusSymbol, res.Path,
					formatStatus(res.NativeParity),
					formatStatus(res.DiagnosticsParity),
					res.Duration.Milliseconds())

				if verbose && !res.OverallMatch {
					if res.DiscrepancyDetails != "" {
						fmt.Printf("          \033[31mDetails:\033[0m %s\n", res.DiscrepancyDetails)
					}
					if res.ErrorMessage != "" {
						fmt.Printf("          \033[31mError:\033[0m %s\n", res.ErrorMessage)
					}
				}
			}

			nextToPrint++
		}
	}

	return results, nativePassedCount, diagPassedCount, fullParityCount, categoryStats
}

func runCase(caseTarget, resolvedCorpus, runnerType, workingDir, nodePath, optLevel string) CaseResult {
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

	var directives corpusDirectives
	if content, err := os.ReadFile(entry); err == nil {
		directives = parseCorpusDirectives(string(content))
	}
	runner := runnerType
	if directives.runner != "" {
		runner = directives.runner
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
	caseOptions := compiler.BuildOptions{
		WorkingDir: workingDir,
		Dynamic:    directives.dynamic,
		OptLevel:   optLevel,
	}

	// 1. Runtime Cases (run.expected / native.expected)
	if hasRunExpected || hasNativeExpected {
		var nodeOut string
		var nodeErr error
		if hasRunExpected {
			nodeOut, nodeErr = runWithNode(entry, runner, workingDir, nodePath)
		}
		res.NodeOutput = nodeOut

		cleanNodeOut := cleanTraceOutput(nodeOut)
		cleanExpected := cleanTraceOutput(expectedTarget)
		nodeMatchesTarget := !hasRunExpected || (nodeErr == nil && (nodeOut == expectedTarget || strings.TrimSpace(nodeOut) == strings.TrimSpace(expectedTarget) || strings.TrimSpace(cleanNodeOut) == strings.TrimSpace(cleanExpected)))

		// Run ScriptGo Native
		sgOut, sgErr := compiler.RunWithOptions(entry, caseOptions)
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
			if nodeMatchesTarget {
				res.OverallMatch = true
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
			_, sgErr = compiler.RunWithOptions(entry, caseOptions)
		} else {
			_, sgErr = compiler.CompileWithOptions(entry, caseOptions)
		}

		if sgErr != nil && strings.Contains(sgErr.Error(), strings.TrimSpace(errExp)) {
			res.DiagnosticsParity = StatusPass
			res.OverallMatch = true
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
	return res
}

func runWithNode(entry, runner, workingDir, nodePath string) (string, error) {
	entry = filepath.Clean(entry)
	absoluteEntry, err := filepath.Abs(entry)
	if err != nil {
		return "", fmt.Errorf("resolve Node entry point %q: %w", entry, err)
	}

	var cmd *exec.Cmd
	var emittedEntry string

	switch runner {
	case "tsx":
		cmd = exec.Command("tsx", absoluteEntry)
	case "tsc":
		cmd = exec.Command("tsc", "--noEmit", absoluteEntry)
	case "tsc-node22":
		tscPath := os.Getenv("TSC_BIN")
		hasTSC := tscPath != ""
		if tscPath == "" {
			tscPath = filepath.Join(workingDir, "node_modules", ".bin", "tsc")
			if absoluteTSC, err := filepath.Abs(tscPath); err == nil {
				tscPath = absoluteTSC
			}
			hasTSC = fileExists(tscPath)
			if !hasTSC {
				tscPath = "tsc"
			}
		}
		if !hasTSC {
			_, err := exec.LookPath(tscPath)
			hasTSC = err == nil
		}
		if !hasTSC {
			cmd = exec.Command(nodePath, "--expose-gc", "--no-warnings", "--experimental-transform-types", absoluteEntry)
			break
		}
		tempDir, err := os.MkdirTemp("", "scriptgo-parity-tsc-")
		if err != nil {
			return "", fmt.Errorf("create TypeScript output directory: %w", err)
		}
		defer os.RemoveAll(tempDir)
		emittedEntry = filepath.Join(tempDir, strings.TrimSuffix(filepath.Base(absoluteEntry), filepath.Ext(absoluteEntry))+".js")
		compile := exec.Command(tscPath, "--target", "ES2022", "--module", "ES2022", "--moduleResolution", "Bundler", "--experimentalDecorators", "--emitDecoratorMetadata", "--skipLibCheck", "--noEmitOnError", "false", "--outDir", tempDir, "--rootDir", filepath.Dir(absoluteEntry), absoluteEntry)
		compile.Dir = workingDir
		var compileOutput bytes.Buffer
		compile.Stdout = &compileOutput
		compile.Stderr = &compileOutput
		if err := compile.Run(); err != nil && !fileExists(emittedEntry) {
			return compileOutput.String(), fmt.Errorf("tsc: %w\n%s", err, compileOutput.String())
		}
		if emittedEntry == "" || !fileExists(emittedEntry) {
			return compileOutput.String(), fmt.Errorf("tsc did not emit %s", emittedEntry)
		}
		packageJSON := filepath.Join(tempDir, "package.json")
		if err := os.WriteFile(packageJSON, []byte(`{"type":"module"}`), 0o644); err != nil {
			return "", fmt.Errorf("write temporary package metadata: %w", err)
		}
		cmd = exec.Command(nodePath, "--expose-gc", "--no-warnings", emittedEntry)
		if workingDir != "" {
			nodeModules := filepath.Join(workingDir, "node_modules")
			if absoluteNodeModules, err := filepath.Abs(nodeModules); err == nil {
				nodeModules = absoluteNodeModules
			}
			if info, err := os.Stat(nodeModules); err == nil && info.IsDir() {
				if err := os.Symlink(nodeModules, filepath.Join(tempDir, "node_modules")); err != nil {
					return "", fmt.Errorf("link temporary Node dependencies: %w", err)
				}
			}
			cmd.Env = append(os.Environ(), "NODE_PATH="+nodeModules)
		}
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func findCorpusCases(root, filter string) ([]string, error) {
	var cases []string
	dirsWithMain := make(map[string]bool)

	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && entry.Name() == "main.ts" {
			dirsWithMain[filepath.Dir(path)] = true
		}
		return nil
	})

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
