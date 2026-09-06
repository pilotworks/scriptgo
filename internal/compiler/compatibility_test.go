package compiler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pilotworks/scriptgo/internal/lowering"
)

func writeCompatibilityEntry(t *testing.T, source string) string {
	t.Helper()
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	return entry
}

func TestCoverageReportJSONIsDeterministicAndInspectable(t *testing.T) {
	entry := writeCompatibilityEntry(t, "const value: any = 42;\n")
	first, err := CoverageReportJSON(entry, BuildOptions{Dynamic: true})
	if err != nil {
		t.Fatal(err)
	}
	second, err := CoverageReportJSON(entry, BuildOptions{Dynamic: true})
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("compatibility report is not deterministic")
	}
	if strings.Contains(first, filepath.Dir(entry)) {
		t.Fatalf("compatibility report contains absolute path: %s", first)
	}
	var report lowering.CompatibilityReport
	if err := json.Unmarshal([]byte(first), &report); err != nil {
		t.Fatal(err)
	}
	if report.Format != 1 || report.Mode != lowering.ModeDynamicEnabled || report.Summary.Dynamic == 0 {
		t.Fatalf("unexpected compatibility report: %+v", report)
	}
	if report.Summary.Static+report.Summary.Dynamic+report.Summary.Unsupported != len(report.Decisions) {
		t.Fatalf("summary does not match site count: %+v", report)
	}
	otherEntry := writeCompatibilityEntry(t, "const value: any = 42;\n")
	other, err := CoverageReportJSON(otherEntry, BuildOptions{Dynamic: true})
	if err != nil {
		t.Fatal(err)
	}
	if first != other {
		t.Fatal("compatibility report depends on the temporary source root")
	}
}

func TestCoverageSummaryShowsWholeProgramStatus(t *testing.T) {
	entry := writeCompatibilityEntry(t, "const value: any = 42;\n")
	summary, err := CoverageSummary(entry, BuildOptions{Dynamic: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"ScriptGo coverage summary",
		"Mode: dynamic-enabled",
		"Result: dynamic-runtime-required",
		"Dynamic:",
		"SG1001: 1 site(s)",
		"Details: rerun with --format json",
	} {
		if !strings.Contains(summary, expected) {
			t.Errorf("summary does not contain %q:\n%s", expected, summary)
		}
	}
}

func TestDynamicModeDoesNotReclassifyStaticLoweringFence(t *testing.T) {
	entry := writeCompatibilityEntry(t, "const value = {};\n")
	reportJSON, err := CoverageReportJSON(entry, BuildOptions{Dynamic: true})
	if err != nil {
		t.Fatal(err)
	}
	var report lowering.CompatibilityReport
	if err := json.Unmarshal([]byte(reportJSON), &report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.Unsupported == 0 || report.Summary.Dynamic != 0 {
		t.Fatalf("empty-object fence was reclassified as Dynamic: %+v", report)
	}
	if _, err := CompileWithOptions(entry, BuildOptions{Dynamic: true}); err == nil || !strings.Contains(err.Error(), "SG2005") {
		t.Fatalf("CompileWithOptions error = %v, want SG2005", err)
	}
}

func TestDynamicImportIsReportedBeforeIR(t *testing.T) {
	entry := filepath.Join("testdata", "corpus", "language", "compatibility_tiers", "dynamic_import", "main.ts")
	reportJSON, err := CoverageReportJSON(entry, BuildOptions{Dynamic: true})
	if err != nil {
		t.Fatal(err)
	}
	var report lowering.CompatibilityReport
	if err := json.Unmarshal([]byte(reportJSON), &report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.Dynamic == 0 {
		t.Fatalf("dynamic import was not classified Dynamic: %+v", report)
	}
	if _, err := CompileWithOptions(entry, BuildOptions{Dynamic: true}); err == nil || !strings.Contains(err.Error(), "SG5001") {
		t.Fatalf("dynamic import compilation error = %v, want SG5001", err)
	}
}

func TestDynamicCompilationStopsBeforeIR(t *testing.T) {
	entry := writeCompatibilityEntry(t, "const value: any = 42;\n")
	if _, err := CompileWithOptions(entry, BuildOptions{Dynamic: true}); err == nil || !strings.Contains(err.Error(), "SG5001") {
		t.Fatalf("CompileWithOptions error = %v, want SG5001", err)
	}
	if _, err := DumpIRWithOptions(entry, BuildOptions{Dynamic: true}); err == nil || !strings.Contains(err.Error(), "SG5001") {
		t.Fatalf("DumpIRWithOptions error = %v, want SG5001", err)
	}
}

func TestCompatibilityMetadataRecordsModeAndCounts(t *testing.T) {
	entry := writeCompatibilityEntry(t, "console.log(42);\n")
	output, err := CompileWithOptions(entry, BuildOptions{Dynamic: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		`; scriptgo.compatibility-mode = "dynamic-enabled"`,
		`; scriptgo.compatibility-report-format = "1"`,
		`; scriptgo.dynamic-sites = "0"`,
		`; scriptgo.unsupported-sites = "0"`,
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("LLVM output does not contain %q", expected)
		}
	}
}
