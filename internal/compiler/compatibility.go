package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/lowering"
)

func compatibilityMode(options BuildOptions) lowering.CompatibilityMode {
	if options.Dynamic {
		return lowering.ModeDynamicEnabled
	}
	return lowering.ModeStatic
}

func loadProgram(entryPath string, options BuildOptions) (frontend.Program, error) {
	source, err := os.ReadFile(entryPath)
	if err != nil {
		return frontend.Program{}, fmt.Errorf("read entry point %q: %w", entryPath, err)
	}
	return frontend.NewProgramWithOptions(entryPath, string(source), frontend.ProgramOptions{
		ConfigPath:                 options.TSConfig,
		AllowJavaScriptImplicitAny: options.Dynamic,
	})
}

func analyzeProgram(program frontend.Program, options BuildOptions) (lowering.CompatibilityReport, error) {
	return lowering.AnalyzeCompatibility(program, lowering.CompatibilityPolicy{Mode: compatibilityMode(options)})
}

func enforceProgram(program frontend.Program, report lowering.CompatibilityReport, options BuildOptions) error {
	if !options.Dynamic {
		// Preserve the established Static first-diagnostic contract exactly.
		if lowering.HasDynamicImports(program) {
			return lowering.EnforceCompatibility(report, lowering.CompatibilityCapabilities{})
		}
		return lowering.ValidateSubsetWithOptions(program, lowering.Options{WarnRuntimeCasts: options.WarnRuntimeCasts})
	}
	return lowering.EnforceCompatibility(report, lowering.CompatibilityCapabilities{DynamicRuntime: true})
}

func loadCoverageReport(entryPath string, options BuildOptions) (lowering.CompatibilityReport, error) {
	program, err := loadProgram(entryPath, options)
	if err != nil {
		return lowering.CompatibilityReport{}, err
	}
	return analyzeProgram(program, options)
}

// CoverageSummary returns a human-readable whole-program overview.
func CoverageSummary(entryPath string, options BuildOptions) (string, error) {
	report, err := loadCoverageReport(entryPath, options)
	if err != nil {
		return "", err
	}
	return formatCoverageSummary(report), nil
}

func formatCoverageSummary(report lowering.CompatibilityReport) string {
	total := len(report.Decisions)
	result := "static-compatible"
	if report.Summary.Unsupported > 0 {
		result = "unsupported"
	} else if report.Summary.Dynamic > 0 {
		result = "dynamic-runtime-required"
	}
	percentage := func(count int) float64 {
		if total == 0 {
			return 0
		}
		return float64(count) * 100 / float64(total)
	}

	var out strings.Builder
	out.WriteString("ScriptGo coverage summary\n")
	fmt.Fprintf(&out, "Mode: %s\n", report.Mode)
	fmt.Fprintf(&out, "Result: %s\n", result)
	fmt.Fprintf(&out, "Sites: %d total\n", total)
	fmt.Fprintf(&out, "  Static:      %d (%.1f%%)\n", report.Summary.Static, percentage(report.Summary.Static))
	fmt.Fprintf(&out, "  Dynamic:     %d (%.1f%%)\n", report.Summary.Dynamic, percentage(report.Summary.Dynamic))
	fmt.Fprintf(&out, "  Unsupported: %d (%.1f%%)\n", report.Summary.Unsupported, percentage(report.Summary.Unsupported))

	counts := make(map[string]int)
	for _, decision := range report.Decisions {
		if decision.Code != "" {
			counts[string(decision.Code)]++
		}
	}
	if len(counts) > 0 {
		codes := make([]string, 0, len(counts))
		for code := range counts {
			codes = append(codes, code)
		}
		sort.Strings(codes)
		out.WriteString("Findings:\n")
		for _, code := range codes {
			fmt.Fprintf(&out, "  %s: %d site(s)\n", code, counts[code])
		}
	}
	out.WriteString("Details: rerun with --format json\n")
	return out.String()
}

// CoverageReportJSON returns deterministic detailed JSON without invoking lowering or LLVM.
func CoverageReportJSON(entryPath string, options BuildOptions) (string, error) {
	report, err := loadCoverageReport(entryPath, options)
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode compatibility report: %w", err)
	}
	return string(data) + "\n", nil
}
