// Package pkgmgr owns package metadata and installation contracts. It does
// not perform TypeScript parsing or native compilation.
package pkgmgr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// InitOptions configures project scaffolding.
type InitOptions struct {
	TargetDir string
	Name      string
	Force     bool
}

// InitResult contains the outcome of project scaffolding.
type InitResult struct {
	TargetDir    string
	CreatedFiles []string
	SkippedFiles []string
}

// SanitizePackageName converts an arbitrary string into a valid npm/package.json package name.
func SanitizePackageName(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "." || raw == ".." {
		return "scriptgo-project"
	}
	var b strings.Builder
	for _, r := range raw {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else if r == ' ' || r == '/' || r == '\\' {
			b.WriteRune('-')
		}
	}
	res := strings.Trim(b.String(), "-._")
	if res == "" {
		return "scriptgo-project"
	}
	return res
}

// Init scaffolds a new ScriptGo TypeScript project.
func Init(opts InitOptions) (*InitResult, error) {
	targetDir := opts.TargetDir
	if targetDir == "" {
		targetDir = "."
	}
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("resolve target directory: %w", err)
	}

	if err := os.MkdirAll(absTargetDir, 0o755); err != nil {
		return nil, fmt.Errorf("create directory %s: %w", absTargetDir, err)
	}

	pkgJSONPath := filepath.Join(absTargetDir, "package.json")
	if !opts.Force {
		if _, err := os.Stat(pkgJSONPath); err == nil {
			return nil, fmt.Errorf("package.json already exists in %s (use --force to overwrite)", absTargetDir)
		}
	}

	name := opts.Name
	if name == "" {
		base := filepath.Base(absTargetDir)
		name = SanitizePackageName(base)
	} else {
		name = SanitizePackageName(name)
	}

	manifest := PackageManifest{
		Name:    name,
		Version: "1.0.0",
		Type:    "module",
		Main:    "index.ts",
		Scripts: map[string]string{
			"start":    "scriptgo run index.ts",
			"build":    "scriptgo build index.ts",
			"check":    "scriptgo check index.ts",
			"coverage": "scriptgo coverage index.ts",
		},
		Dependencies:    map[string]string{},
		DevDependencies: map[string]string{},
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("serialize package.json: %w", err)
	}
	manifestData = append(manifestData, '\n')

	tsconfigContent := `{
  "compilerOptions": {
    "target": "ESNext",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  }
}
`

	indexContent := `console.log("Hello via ScriptGo!");
`

	gitignoreContent := `node_modules/
.scriptgo/
bin/
dist/
*.ll
*.s
*.o
`

	files := []struct {
		relPath string
		content []byte
	}{
		{"package.json", manifestData},
		{"tsconfig.json", []byte(tsconfigContent)},
		{"index.ts", []byte(indexContent)},
		{".gitignore", []byte(gitignoreContent)},
	}

	result := &InitResult{
		TargetDir: absTargetDir,
	}

	for _, f := range files {
		fullPath := filepath.Join(absTargetDir, f.relPath)
		if !opts.Force && f.relPath != "package.json" {
			if _, err := os.Stat(fullPath); err == nil {
				result.SkippedFiles = append(result.SkippedFiles, f.relPath)
				continue
			}
		}
		if err := os.WriteFile(fullPath, f.content, 0o644); err != nil {
			return nil, fmt.Errorf("write %s: %w", f.relPath, err)
		}
		result.CreatedFiles = append(result.CreatedFiles, f.relPath)
	}

	return result, nil
}
