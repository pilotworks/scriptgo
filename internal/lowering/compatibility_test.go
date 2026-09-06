package lowering

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/pilotworks/scriptgo/internal/frontend"
)

func compatibilityProgram(t *testing.T, source string) frontend.Program {
	t.Helper()
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}
	return program
}

func TestAnalyzeCompatibilityClassifiesAnyByMode(t *testing.T) {
	program := compatibilityProgram(t, "const value: any = 42;\nconsole.log(value);\n")
	staticReport, err := AnalyzeCompatibility(program, CompatibilityPolicy{Mode: ModeStatic})
	if err != nil {
		t.Fatal(err)
	}
	dynamicReport, err := AnalyzeCompatibility(program, CompatibilityPolicy{Mode: ModeDynamicEnabled})
	if err != nil {
		t.Fatal(err)
	}
	if staticReport.Summary.Unsupported == 0 || staticReport.Summary.Dynamic != 0 {
		t.Fatalf("static summary = %+v, want unsupported any site", staticReport.Summary)
	}
	if dynamicReport.Summary.Dynamic == 0 || dynamicReport.Summary.Unsupported != 0 {
		t.Fatalf("dynamic summary = %+v, want Dynamic any site", dynamicReport.Summary)
	}
	if dynamicReport.Decisions[0].FileName != "main.ts" {
		t.Fatalf("report path = %q, want relative main.ts", dynamicReport.Decisions[0].FileName)
	}
}

func TestAnalyzeCompatibilityCollectsAndSortsSites(t *testing.T) {
	program := compatibilityProgram(t, "let first: any = 1;\nlet second: any = 2;\n")
	report, err := AnalyzeCompatibility(program, CompatibilityPolicy{})
	if err != nil {
		t.Fatal(err)
	}
	unsupported := 0
	lastStart := -1
	seen := map[string]bool{}
	for _, decision := range report.Decisions {
		key := decision.FileName + ":" + decision.Kind + ":" + strconv.Itoa(decision.Start)
		if seen[key] {
			t.Fatalf("duplicate compatibility site: %+v", decision)
		}
		seen[key] = true
		if decision.Start < lastStart {
			t.Fatalf("report is not sorted: %+v", report.Decisions)
		}
		lastStart = decision.Start
		if decision.Code == CodeAnyBoundary {
			unsupported++
		}
	}
	if unsupported != 2 {
		t.Fatalf("any decisions = %d, want 2; report=%+v", unsupported, report)
	}
}

func TestEnforceCompatibilityRejectsUnavailableDynamicRuntime(t *testing.T) {
	program := compatibilityProgram(t, "const value: any = 42;\n")
	report, err := AnalyzeCompatibility(program, CompatibilityPolicy{Mode: ModeDynamicEnabled})
	if err != nil {
		t.Fatal(err)
	}
	err = EnforceCompatibility(report, CompatibilityCapabilities{})
	if err == nil || !strings.Contains(err.Error(), string(CodeDynamicRuntimeUnavailable)) {
		t.Fatalf("enforcement error = %v, want SG5001", err)
	}
}

func TestAnalyzeCompatibilityDoesNotMutateGenericProgram(t *testing.T) {
	program := compatibilityProgram(t, "function identity<T>(value: T): T { return value; }\nconsole.log(identity<number>(42));\n")
	if _, err := AnalyzeCompatibility(program, CompatibilityPolicy{}); err != nil {
		t.Fatal(err)
	}
	if _, err := Lower(program); err != nil {
		t.Fatalf("lowering after compatibility analysis failed: %v", err)
	}
}
