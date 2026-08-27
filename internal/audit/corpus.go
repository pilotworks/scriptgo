package audit

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pilotworks/scriptgo/internal/spec"
)

// CorpusAPIItem represents a single @api annotation found in a corpus test.
type CorpusAPIItem struct {
	Tag           string `json:"tag"`            // Raw tag after @api: e.g. "timers.setTimeout" or "assert.ok(value[, message])"
	NormalizedKey string `json:"normalized_key"` // Lowercase normalized key: "timers.settimeout", "assert.ok", "fs.readfilesync"
	Module        string `json:"module"`         // Inferred module name, e.g. "timers", "assert", "fs"
	FilePath      string `json:"file_path"`      // File path relative to repo root
	LineNumber    int    `json:"line_number"`
	CodeSnippet   string `json:"code_snippet,omitempty"` // Extracted corpus test code snippet
}

// CorpusAPICatalog stores all indexed corpus API annotations.
type CorpusAPICatalog struct {
	ItemsByKey    map[string][]CorpusAPIItem `json:"items_by_key"`
	ItemsByModule map[string][]CorpusAPIItem `json:"items_by_module"`
	AllItems      []CorpusAPIItem            `json:"all_items"`
}

// ScanCorpusAPIs scans the given corpus directory and indexes all @api annotations and code snippets.
func ScanCorpusAPIs(corpusRoot string) (*CorpusAPICatalog, error) {
	catalog := &CorpusAPICatalog{
		ItemsByKey:    make(map[string][]CorpusAPIItem),
		ItemsByModule: make(map[string][]CorpusAPIItem),
		AllItems:      make([]CorpusAPIItem, 0),
	}

	err := filepath.WalkDir(corpusRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".ts") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		baseName := strings.TrimSuffix(d.Name(), ".ts")
		relDir, _ := filepath.Rel(corpusRoot, filepath.Dir(path))
		defaultModule := baseName
		if d.Name() == "main.ts" {
			defaultModule = filepath.Base(filepath.Dir(path))
		}
		if relDir == "api" {
			defaultModule = baseName
		}

		lines := strings.Split(string(content), "\n")
		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if !strings.HasPrefix(line, "//") {
				continue
			}
			comment := strings.TrimSpace(strings.TrimPrefix(line, "//"))
			if !strings.HasPrefix(comment, "@api:") && !strings.HasPrefix(comment, "@api ") {
				continue
			}

			rawTag := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(comment, "@api:"), "@api "))
			if rawTag == "" {
				continue
			}

			norm := spec.NormalizeAPIName(rawTag)
			normKey := strings.ToLower(norm)

			// Infer module name
			module := defaultModule
			if dotIdx := strings.Index(norm, "."); dotIdx != -1 {
				prefix := norm[:dotIdx]
				if prefix != "" {
					module = prefix
				}
			}

			// If normKey doesn't contain a dot, prefix with module
			if !strings.Contains(normKey, ".") {
				normKey = strings.ToLower(module + "." + normKey)
			}

			// Extract code snippet until next separated @api tag or block boundary
			var snippetLines []string
			seenCode := false
			for j := i; j < len(lines); j++ {
				trimmed := strings.TrimSpace(lines[j])
				if j > i && strings.HasPrefix(trimmed, "//") {
					c := strings.TrimSpace(strings.TrimPrefix(trimmed, "//"))
					if (strings.HasPrefix(c, "@api:") || strings.HasPrefix(c, "@api ")) && seenCode {
						break
					}
				}
				if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
					seenCode = true
				}
				snippetLines = append(snippetLines, lines[j])
				if len(snippetLines) >= 40 {
					break
				}
			}

			// Clean snippet
			snippetStr := strings.TrimSpace(strings.Join(snippetLines, "\n"))

			cleanFilePath := filepath.Clean(path)
			for strings.HasPrefix(cleanFilePath, "../") || strings.HasPrefix(cleanFilePath, "..\\") {
				cleanFilePath = strings.TrimPrefix(strings.TrimPrefix(cleanFilePath, "../"), "..\\")
			}
			cleanFilePath = strings.TrimPrefix(cleanFilePath, "./")

			item := CorpusAPIItem{
				Tag:           rawTag,
				NormalizedKey: normKey,
				Module:        strings.ToLower(module),
				FilePath:      cleanFilePath,
				LineNumber:    i + 1,
				CodeSnippet:   snippetStr,
			}

			catalog.AllItems = append(catalog.AllItems, item)
			catalog.ItemsByKey[normKey] = append(catalog.ItemsByKey[normKey], item)
			catalog.ItemsByModule[strings.ToLower(module)] = append(catalog.ItemsByModule[strings.ToLower(module)], item)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return catalog, nil
}
