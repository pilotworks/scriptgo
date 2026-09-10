package pkgmgr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolvedPackage identifies a package and its selected source entry point.
type ResolvedPackage struct {
	Root     string
	Manifest PackageManifest
	Entry    string
}

// ResolveLocalPackage resolves a package from node_modules without network
// access. Search follows Node's upward node_modules lookup convention.
func ResolveLocalPackage(projectRoot, specifier string) (ResolvedPackage, error) {
	if filepath.IsAbs(specifier) || strings.HasPrefix(specifier, ".") || specifier == "" {
		return ResolvedPackage{}, fmt.Errorf("package specifier %q is not a bare package name", specifier)
	}
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		return ResolvedPackage{}, fmt.Errorf("resolve project root: %w", err)
	}
	for dir := root; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "node_modules", filepath.FromSlash(specifier))
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return loadResolvedPackage(candidate)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return ResolvedPackage{}, fmt.Errorf("package %q was not found from %q", specifier, projectRoot)
}

func loadResolvedPackage(root string) (ResolvedPackage, error) {
	manifest, err := LoadManifest(packageManifestPath(root))
	if err != nil {
		return ResolvedPackage{}, err
	}
	entry, err := resolveEntry(root, manifest)
	if err != nil {
		return ResolvedPackage{}, err
	}
	return ResolvedPackage{Root: root, Manifest: manifest, Entry: entry}, nil
}

func resolveEntry(root string, manifest PackageManifest) (string, error) {
	if target := exportsEntry(manifest.Exports); target != "" {
		if entry, ok := existingPackageFile(root, target); ok {
			return entry, nil
		}
		return "", fmt.Errorf("package %q exports entry %q, but the file is missing", manifest.Name, target)
	}
	for _, target := range []string{manifest.Module, manifest.Main, "index.js", "index.mjs", "index.cjs", "index.ts"} {
		if entry, ok := existingPackageFile(root, target); ok {
			return entry, nil
		}
	}
	return "", fmt.Errorf("package %q has no resolvable entry point", manifest.Name)
}

func exportsEntry(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var direct string
	if json.Unmarshal(raw, &direct) == nil {
		return direct
	}
	var entries map[string]json.RawMessage
	if json.Unmarshal(raw, &entries) != nil {
		return ""
	}
	value, ok := entries["."]
	if !ok {
		return ""
	}
	if json.Unmarshal(value, &direct) == nil {
		return direct
	}
	var conditions map[string]string
	if json.Unmarshal(value, &conditions) != nil {
		return ""
	}
	for _, condition := range []string{"import", "require", "default"} {
		if target := conditions[condition]; target != "" {
			return target
		}
	}
	return ""
}

func existingPackageFile(root, target string) (string, bool) {
	if target == "" || filepath.IsAbs(target) {
		return "", false
	}
	cleanRoot := filepath.Clean(root)
	path := filepath.Clean(filepath.Join(cleanRoot, filepath.FromSlash(target)))
	if path != cleanRoot && !strings.HasPrefix(path, cleanRoot+string(filepath.Separator)) {
		return "", false
	}
	info, err := os.Stat(path)
	return path, err == nil && !info.IsDir()
}
