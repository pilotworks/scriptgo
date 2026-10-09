package test262

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

// Summary counts outcomes for one group of tests.
type Summary struct {
	Group       string `json:"group"`
	Pass        int    `json:"pass"`
	Fail        int    `json:"fail"`
	Unsupported int    `json:"unsupported"`
	Skipped     int    `json:"skipped"`
}

// Total is the number of tests in the group.
func (s Summary) Total() int { return s.Pass + s.Fail + s.Unsupported + s.Skipped }

// Attempted is the number of tests the runner executed or compiled.
func (s Summary) Attempted() int { return s.Pass + s.Fail + s.Unsupported }

func (s *Summary) add(o Outcome) {
	switch o {
	case Pass:
		s.Pass++
	case Fail:
		s.Fail++
	case Unsupported:
		s.Unsupported++
	case Skipped:
		s.Skipped++
	}
}

// Summarize groups results by the first depth path segments and adds a
// final "total" row.
func Summarize(results []Result, depth int) []Summary {
	groups := map[string]*Summary{}
	total := &Summary{Group: "total"}
	for _, r := range results {
		parts := strings.Split(r.Path, "/")
		if len(parts) > depth {
			parts = parts[:depth]
		}
		key := strings.Join(parts, "/")
		if groups[key] == nil {
			groups[key] = &Summary{Group: key}
		}
		groups[key].add(r.Outcome)
		total.add(r.Outcome)
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]Summary, 0, len(keys)+1)
	for _, k := range keys {
		out = append(out, *groups[k])
	}
	return append(out, *total)
}

var addressPattern = regexp.MustCompile(`0x[0-9a-fA-F]+`)

// TopReasons returns the most frequent details for an outcome, normalized
// so that per-test names and addresses group together.
func TopReasons(results []Result, outcome Outcome, limit int) []string {
	counts := map[string]int{}
	for _, r := range results {
		if r.Outcome != outcome || r.Detail == "" {
			continue
		}
		detail := r.Detail
		if i := strings.Index(detail, " - error "); i >= 0 {
			detail = detail[i+3:]
		}
		detail = addressPattern.ReplaceAllString(detail, "0x…")
		counts[detail]++
	}
	type kv struct {
		reason string
		count  int
	}
	var list []kv
	for k, v := range counts {
		list = append(list, kv{k, v})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].count != list[j].count {
			return list[i].count > list[j].count
		}
		return list[i].reason < list[j].reason
	})
	var out []string
	for i := 0; i < len(list) && i < limit; i++ {
		out = append(out, fmt.Sprintf("%5d  %s", list[i].count, list[i].reason))
	}
	return out
}

// WriteText prints a Markdown summary table.
func WriteText(w io.Writer, summaries []Summary) {
	fmt.Fprintln(w, "| Group | Pass | Fail | Unsupported | Skipped | Pass rate (attempted) |")
	fmt.Fprintln(w, "| --- | ---: | ---: | ---: | ---: | ---: |")
	for _, s := range summaries {
		rate := "n/a"
		if s.Attempted() > 0 {
			rate = fmt.Sprintf("%.1f%%", 100*float64(s.Pass)/float64(s.Attempted()))
		}
		fmt.Fprintf(w, "| %s | %d | %d | %d | %d | %s |\n", s.Group, s.Pass, s.Fail, s.Unsupported, s.Skipped, rate)
	}
}

// Regressions returns the results that did not pass, in input order. A
// baseline run (-require-pass) lists tests that passed before, so any entry
// here is a conformance regression.
func Regressions(results []Result) []Result {
	var regressed []Result
	for _, r := range results {
		if r.Outcome != Pass {
			regressed = append(regressed, r)
		}
	}
	return regressed
}
