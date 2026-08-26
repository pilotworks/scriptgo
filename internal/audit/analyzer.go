package audit

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pilotworks/scriptgo/internal/spec"
)

type APIStatus string

const (
	StatusVerified APIStatus = "VERIFIED" // In Official Spec & has passing Corpus Test
	StatusMissing  APIStatus = "MISSING"  // In Official Spec & not tested in Corpus
)

// APIAuditResult represents the audited status of a single API entry.
type APIAuditResult struct {
	CanonicalAPI spec.CanonicalAPI `json:"spec_api"`
	Status       APIStatus         `json:"status"`
	CorpusTests  []CorpusAPIItem   `json:"corpus_tests,omitempty"`
}

// ModuleAuditReport summarizes the audit result for a single module.
type ModuleAuditReport struct {
	ModuleName    string           `json:"module_name"`
	TotalOfficial int              `json:"total_official_apis"`
	VerifiedCount int              `json:"verified_apis"`
	MissingCount  int              `json:"missing_apis"`
	CoverageRate  float64          `json:"coverage_rate_percent"`
	Results       []APIAuditResult `json:"results"`
	CorpusExtra   []CorpusAPIItem  `json:"corpus_extra,omitempty"`
}

// OverallAuditReport represents the combined report across all audited modules.
type OverallAuditReport struct {
	TotalModules      int                          `json:"total_modules"`
	TotalOfficialAPIs int                          `json:"total_official_apis"`
	TotalVerifiedAPIs int                          `json:"total_verified_apis"`
	TotalMissingAPIs  int                          `json:"total_missing_apis"`
	OverallCoverage   float64                      `json:"overall_coverage_percent"`
	ModuleReports     map[string]ModuleAuditReport `json:"module_reports"`
}

// StandardNodeModules lists all 64 canonical API modules and documentation sections in Node.js 22 LTS.
var StandardNodeModules = []string{
	"addons",
	"assert",
	"async_context",
	"async_hooks",
	"buffer",
	"child_process",
	"cli",
	"cluster",
	"console",
	"crypto",
	"debugger",
	"deprecations",
	"dgram",
	"diagnostics_channel",
	"dns",
	"documentation",
	"domain",
	"embedding",
	"environment_variables",
	"errors",
	"esm",
	"events",
	"fs",
	"globals",
	"http",
	"http2",
	"https",
	"inspector",
	"intl",
	"module",
	"modules",
	"n-api",
	"net",
	"os",
	"packages",
	"path",
	"perf_hooks",
	"permissions",
	"process",
	"punycode",
	"querystring",
	"readline",
	"repl",
	"report",
	"single-executable-applications",
	"sqlite",
	"stream",
	"string_decoder",
	"synopsis",
	"test",
	"timers",
	"tls",
	"tracing",
	"tty",
	"typescript",
	"url",
	"util",
	"v8",
	"vm",
	"wasi",
	"webcrypto",
	"webstreams",
	"worker_threads",
	"zlib",
}

// AuditModule audits a single module against the corpus catalog.
func AuditModule(moduleName string, doc *spec.DocRoot, corpus *CorpusAPICatalog) ModuleAuditReport {
	canonicalList := spec.ExtractCanonicalAPIs(moduleName, doc)
	cleanMod := strings.ToLower(strings.TrimPrefix(moduleName, "node:"))

	report := ModuleAuditReport{
		ModuleName:    cleanMod,
		TotalOfficial: len(canonicalList),
		Results:       make([]APIAuditResult, 0, len(canonicalList)),
	}

	matchedCorpusItems := make(map[string]bool)

	for _, api := range canonicalList {
		normKey := strings.ToLower(api.NormalizedKey)
		// Check key matches: full key ("timers.settimeout") or class key ("immediate.hasref") or plain name ("settimeout")
		tests, found := corpus.ItemsByKey[normKey]
		if !found {
			// Try matching with dot suffix (e.g. timers.settimeout matching settimeout)
			if idx := strings.Index(normKey, "."); idx != -1 {
				shortKey := normKey[idx+1:]
				tests, found = corpus.ItemsByKey[shortKey]
			}
		}
		if !found {
			// Try matching with module prefix + name
			altKey := cleanMod + "." + strings.ToLower(api.Name)
			tests, found = corpus.ItemsByKey[altKey]
		}

		res := APIAuditResult{
			CanonicalAPI: api,
		}

		if found && len(tests) > 0 {
			res.Status = StatusVerified
			res.CorpusTests = tests
			report.VerifiedCount++
			for _, t := range tests {
				matchedCorpusItems[t.NormalizedKey] = true
			}
		} else {
			res.Status = StatusMissing
			report.MissingCount++
		}

		report.Results = append(report.Results, res)
	}

	// Check any corpus items for this module that were not directly in official canonical list
	if modItems, exists := corpus.ItemsByModule[cleanMod]; exists {
		for _, item := range modItems {
			if !matchedCorpusItems[item.NormalizedKey] {
				report.CorpusExtra = append(report.CorpusExtra, item)
			}
		}
	}

	if report.TotalOfficial > 0 {
		report.CoverageRate = float64(report.VerifiedCount) / float64(report.TotalOfficial) * 100.0
	}

	return report
}

// AuditAllModules runs audit across specified modules (or all StandardNodeModules if empty).
func AuditAllModules(cacheDir, corpusDir string, moduleNames []string) (*OverallAuditReport, error) {
	corpus, err := ScanCorpusAPIs(corpusDir)
	if err != nil {
		return nil, fmt.Errorf("failed to scan corpus APIs: %w", err)
	}

	targets := moduleNames
	if len(targets) == 0 {
		targets = StandardNodeModules
	}

	report := &OverallAuditReport{
		ModuleReports: make(map[string]ModuleAuditReport),
	}

	for _, mod := range targets {
		doc, err := spec.LoadModuleSpec(cacheDir, mod)
		if err != nil {
			// Skip or record warning if network/spec unavailable
			continue
		}
		modReport := AuditModule(mod, doc, corpus)
		report.ModuleReports[mod] = modReport
		report.TotalModules++
		report.TotalOfficialAPIs += modReport.TotalOfficial
		report.TotalVerifiedAPIs += modReport.VerifiedCount
		report.TotalMissingAPIs += modReport.MissingCount
	}

	if report.TotalOfficialAPIs > 0 {
		report.OverallCoverage = float64(report.TotalVerifiedAPIs) / float64(report.TotalOfficialAPIs) * 100.0
	}

	return report, nil
}

// FormatTerminalTable formats the audit results into a clean CLI table.
func FormatTerminalTable(report *OverallAuditReport, filterModule string, showMissing bool) string {
	var sb strings.Builder

	sb.WriteString("================================================================================\n")
	sb.WriteString("  ScriptGo Node.js API Coverage Audit (Corpus as Source of Truth)\n")
	sb.WriteString("================================================================================\n")
	sb.WriteString(fmt.Sprintf("Audited Modules     : %d\n", report.TotalModules))
	sb.WriteString(fmt.Sprintf("Official APIs Total : %d\n", report.TotalOfficialAPIs))
	sb.WriteString(fmt.Sprintf("Corpus Verified APIs: %d / %d (%.1f%%)\n", report.TotalVerifiedAPIs, report.TotalOfficialAPIs, report.OverallCoverage))
	sb.WriteString(fmt.Sprintf("Missing APIs (Gaps) : %d\n", report.TotalMissingAPIs))
	sb.WriteString("================================================================================\n\n")

	sb.WriteString("Module Parity & Coverage Matrix:\n")
	var modNames []string
	for name := range report.ModuleReports {
		if filterModule != "" && !strings.Contains(name, filterModule) {
			continue
		}
		modNames = append(modNames, name)
	}
	sort.Strings(modNames)

	for _, name := range modNames {
		mr := report.ModuleReports[name]
		bar := progressBar(mr.VerifiedCount, mr.TotalOfficial)
		sb.WriteString(fmt.Sprintf("  %-18s %s  %3d / %-3d (%5.1f%%)\n",
			"node:"+name, bar, mr.VerifiedCount, mr.TotalOfficial, mr.CoverageRate))

		if showMissing && mr.MissingCount > 0 {
			missingCountShown := 0
			for _, res := range mr.Results {
				if res.Status == StatusMissing {
					missingCountShown++
					if missingCountShown <= 8 {
						sb.WriteString(fmt.Sprintf("    \033[90m⏳ Missing:\033[0m %s\n", res.CanonicalAPI.FullName))
					}
				}
			}
			if missingCountShown > 8 {
				sb.WriteString(fmt.Sprintf("    \033[90m... and %d more missing APIs\033[0m\n", missingCountShown-8))
			}
		}
	}
	sb.WriteString("\n")

	return sb.String()
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
