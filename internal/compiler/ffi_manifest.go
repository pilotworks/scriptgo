// Package compiler owns the native compilation pipeline.
package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CurrentFFIFormat is the locked schema format version for FFI metadata files.
const CurrentFFIFormat = 1

// FFIManifest defines the C library metadata configuration.
type FFIManifest struct {
	FFIFormat int                  `json:"ffi_format"`
	Name      string               `json:"name"`
	Link      FFILinkOptions       `json:"link"`
	Symbols   map[string]FFISymbol `json:"symbols,omitempty"`
}

// FFILinkOptions defines linker flags and source assets for native linking.
type FFILinkOptions struct {
	Libraries   []string `json:"libraries,omitempty"`
	LibDirs     []string `json:"libDirs,omitempty"`
	IncludeDirs []string `json:"includeDirs,omitempty"`
	Sources     []string `json:"sources,omitempty"`
	CFlags      []string `json:"cflags,omitempty"`
	Frameworks  []string `json:"frameworks,omitempty"`
}

// FFISymbol defines the signature of a C library exported symbol.
type FFISymbol struct {
	Symbol  string   `json:"symbol,omitempty"`
	Args    []string `json:"args,omitempty"`
	Returns string   `json:"returns,omitempty"`
}

// LoadFFIManifest reads and validates a JSON metadata file.
func LoadFFIManifest(manifestPath string) (*FFIManifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read FFI metadata file %q: %w", manifestPath, err)
	}

	var manifest FFIManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse FFI metadata file %q: %w", manifestPath, err)
	}

	if err := ValidateFFIManifest(&manifest, manifestPath); err != nil {
		return nil, err
	}

	// Resolve relative source paths against the directory of the manifest file
	manifestDir := filepath.Dir(manifestPath)
	for i, src := range manifest.Link.Sources {
		if !filepath.IsAbs(src) {
			manifest.Link.Sources[i] = filepath.Clean(filepath.Join(manifestDir, src))
		}
	}
	for i, dir := range manifest.Link.LibDirs {
		if !filepath.IsAbs(dir) {
			manifest.Link.LibDirs[i] = filepath.Clean(filepath.Join(manifestDir, dir))
		}
	}
	for i, dir := range manifest.Link.IncludeDirs {
		if !filepath.IsAbs(dir) {
			manifest.Link.IncludeDirs[i] = filepath.Clean(filepath.Join(manifestDir, dir))
		}
	}

	return &manifest, nil
}

// ValidateFFIManifest checks the schema version and mandatory fields of a manifest.
func ValidateFFIManifest(manifest *FFIManifest, manifestPath string) error {
	if manifest.FFIFormat == 0 {
		return fmt.Errorf("FFI metadata file %q: missing or unsupported \"ffi_format\" (expected %d)", manifestPath, CurrentFFIFormat)
	}
	if manifest.FFIFormat != CurrentFFIFormat {
		return fmt.Errorf("FFI metadata file %q: unsupported \"ffi_format\" %d (expected %d)", manifestPath, manifest.FFIFormat, CurrentFFIFormat)
	}
	if manifest.Name == "" {
		return fmt.Errorf("FFI metadata file %q: missing mandatory field \"name\"", manifestPath)
	}
	return nil
}
