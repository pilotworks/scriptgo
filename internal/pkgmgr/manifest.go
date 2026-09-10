// Package pkgmgr owns package metadata and installation contracts. It does
// not perform TypeScript parsing or native compilation.
package pkgmgr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PackageManifest is the package.json subset needed for deterministic local
// resolution and lockfile generation.
type PackageManifest struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Type            string            `json:"type,omitempty"`
	Main            string            `json:"main,omitempty"`
	Module          string            `json:"module,omitempty"`
	Exports         json.RawMessage   `json:"exports,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
}

// LoadManifest reads and validates one package.json file.
func LoadManifest(path string) (PackageManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PackageManifest{}, fmt.Errorf("read package manifest %q: %w", path, err)
	}
	var manifest PackageManifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return PackageManifest{}, fmt.Errorf("parse package manifest %q: %w", path, err)
	}
	if err := ValidateManifest(manifest, path); err != nil {
		return PackageManifest{}, err
	}
	return manifest, nil
}

// ValidateManifest checks fields required by the local package contract.
func ValidateManifest(manifest PackageManifest, path string) error {
	if strings.TrimSpace(manifest.Name) == "" {
		return fmt.Errorf("package manifest %q: missing mandatory field \"name\"", path)
	}
	if strings.ContainsAny(manifest.Name, "\\/") {
		return fmt.Errorf("package manifest %q: invalid package name %q", path, manifest.Name)
	}
	if manifest.Type != "" && manifest.Type != "module" && manifest.Type != "commonjs" {
		return fmt.Errorf("package manifest %q: unsupported type %q", path, manifest.Type)
	}
	for name, spec := range manifest.Dependencies {
		if strings.TrimSpace(name) == "" || strings.TrimSpace(spec) == "" {
			return fmt.Errorf("package manifest %q: invalid dependency %q", path, name)
		}
	}
	return nil
}

func packageManifestPath(dir string) string {
	return filepath.Join(dir, "package.json")
}
