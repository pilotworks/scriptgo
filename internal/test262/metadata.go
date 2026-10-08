// Package test262 runs a subset of the TC39 test262 conformance suite through
// the scriptgo compiler and classifies each test as passing, failing,
// unsupported by the native subset, or skipped by metadata policy.
//
// The suite itself is not vendored; callers point the runner at a test262
// checkout. Tests run in the Static tier with a scriptgo-owned harness
// (harness.ts) that implements the parts of assert.js and sta.js expressible
// in the native subset.
package test262

import (
	"fmt"
	"strings"
)

// Negative describes an expected failure from test metadata.
type Negative struct {
	Phase string
	Type  string
}

// Metadata is the subset of test262 frontmatter the runner acts on.
type Metadata struct {
	Includes []string
	Flags    []string
	Features []string
	Negative *Negative
}

// HasFlag reports whether the test declares flag.
func (m Metadata) HasFlag(flag string) bool {
	for _, f := range m.Flags {
		if f == flag {
			return true
		}
	}
	return false
}

// ParseMetadata extracts the YAML frontmatter between "/*---" and "---*/".
// Only the keys the runner uses are interpreted; others are ignored.
func ParseMetadata(source string) (Metadata, error) {
	start := strings.Index(source, "/*---")
	if start < 0 {
		return Metadata{}, fmt.Errorf("missing test262 frontmatter")
	}
	end := strings.Index(source[start:], "---*/")
	if end < 0 {
		return Metadata{}, fmt.Errorf("unterminated test262 frontmatter")
	}
	lines := strings.Split(source[start+len("/*---"):start+end], "\n")
	var meta Metadata
	key := ""
	for _, raw := range lines {
		line := strings.TrimRight(raw, " \t\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		indented := line[0] == ' ' || line[0] == '\t'
		trimmed := strings.TrimSpace(line)
		if !indented {
			name, value, ok := strings.Cut(trimmed, ":")
			if !ok {
				key = ""
				continue
			}
			key = strings.TrimSpace(name)
			value = strings.TrimSpace(value)
			switch key {
			case "includes":
				meta.Includes = append(meta.Includes, inlineList(value)...)
			case "flags":
				meta.Flags = append(meta.Flags, inlineList(value)...)
			case "features":
				meta.Features = append(meta.Features, inlineList(value)...)
			case "negative":
				meta.Negative = &Negative{}
			}
			continue
		}
		switch {
		case strings.HasPrefix(trimmed, "- "):
			item := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			switch key {
			case "includes":
				meta.Includes = append(meta.Includes, item)
			case "flags":
				meta.Flags = append(meta.Flags, item)
			case "features":
				meta.Features = append(meta.Features, item)
			}
		case key == "negative" && meta.Negative != nil:
			name, value, ok := strings.Cut(trimmed, ":")
			if !ok {
				continue
			}
			switch strings.TrimSpace(name) {
			case "phase":
				meta.Negative.Phase = strings.TrimSpace(value)
			case "type":
				meta.Negative.Type = strings.TrimSpace(value)
			}
		}
	}
	return meta, nil
}

func inlineList(value string) []string {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return nil
	}
	var out []string
	for _, item := range strings.Split(value[1:len(value)-1], ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}
