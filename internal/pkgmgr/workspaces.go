package pkgmgr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type workspacePackage struct {
	Root     string
	Manifest PackageManifest
}

func discoverWorkspaces(root string, raw json.RawMessage) (map[string]workspacePackage, error) {
	patterns, err := workspacePatterns(raw)
	if err != nil {
		return nil, err
	}
	result := map[string]workspacePackage{}
	for _, pattern := range patterns {
		exclude := strings.HasPrefix(pattern, "!")
		if exclude {
			pattern = strings.TrimPrefix(pattern, "!")
		}
		matches, globErr := filepath.Glob(filepath.Join(root, filepath.FromSlash(pattern)))
		if globErr != nil {
			return nil, fmt.Errorf("invalid workspace pattern %q: %w", pattern, globErr)
		}
		for _, match := range matches {
			info, statErr := os.Stat(match)
			if statErr != nil || !info.IsDir() {
				continue
			}
			manifest, loadErr := LoadManifest(filepath.Join(match, "package.json"))
			if loadErr != nil {
				if os.IsNotExist(loadErr) {
					continue
				}
				return nil, loadErr
			}
			if exclude {
				delete(result, manifest.Name)
				continue
			}
			if _, exists := result[manifest.Name]; exists {
				return nil, fmt.Errorf("duplicate workspace package %q", manifest.Name)
			}
			result[manifest.Name] = workspacePackage{Root: match, Manifest: manifest}
		}
	}
	return result, nil
}

func workspacePatterns(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var patterns []string
	if err := json.Unmarshal(raw, &patterns); err == nil {
		return patterns, nil
	}
	var config struct {
		Packages []string `json:"packages"`
	}
	if err := json.Unmarshal(raw, &config); err != nil || len(config.Packages) == 0 {
		return nil, fmt.Errorf("workspaces must be an array or an object with a packages array")
	}
	return config.Packages, nil
}
