package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFFIManifestValidation(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Valid manifest
	validJSON := `{
		"ffi_format": 1,
		"name": "test_lib",
		"link": {
			"libraries": ["m"],
			"sources": ["./test.c"]
		},
		"symbols": {
			"foo": { "args": ["f64"], "returns": "f64" }
		}
	}`
	validPath := filepath.Join(tempDir, "valid.ffi.json")
	if err := os.WriteFile(validPath, []byte(validJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest, err := LoadFFIManifest(validPath)
	if err != nil {
		t.Fatalf("LoadFFIManifest failed on valid manifest: %v", err)
	}
	if manifest.Name != "test_lib" || manifest.FFIFormat != 1 {
		t.Fatalf("Unexpected manifest values: %+v", manifest)
	}
	if len(manifest.Link.Sources) != 1 || !filepath.IsAbs(manifest.Link.Sources[0]) {
		t.Fatalf("Expected absolute source path, got %v", manifest.Link.Sources)
	}

	// 2. Missing ffi_format
	missingFormatJSON := `{"name": "test_lib"}`
	missingPath := filepath.Join(tempDir, "missing_format.ffi.json")
	if err := os.WriteFile(missingPath, []byte(missingFormatJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFFIManifest(missingPath); err == nil || !strings.Contains(err.Error(), "missing or unsupported \"ffi_format\"") {
		t.Fatalf("Expected missing ffi_format error, got: %v", err)
	}

	// 3. Unsupported ffi_format (e.g. 99)
	unsupportedFormatJSON := `{"ffi_format": 99, "name": "test_lib"}`
	unsupportedPath := filepath.Join(tempDir, "unsupported_format.ffi.json")
	if err := os.WriteFile(unsupportedPath, []byte(unsupportedFormatJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFFIManifest(unsupportedPath); err == nil || !strings.Contains(err.Error(), "unsupported \"ffi_format\" 99") {
		t.Fatalf("Expected unsupported ffi_format error, got: %v", err)
	}

	// 4. Missing name
	missingNameJSON := `{"ffi_format": 1}`
	missingNamePath := filepath.Join(tempDir, "missing_name.ffi.json")
	if err := os.WriteFile(missingNamePath, []byte(missingNameJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFFIManifest(missingNamePath); err == nil || !strings.Contains(err.Error(), "missing mandatory field \"name\"") {
		t.Fatalf("Expected missing name error, got: %v", err)
	}
}
