package typescriptgo

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// DefaultVersion is the default ScriptGo SDK / stdlib version.
const DefaultVersion = "v0.1.0"

// BuiltinModule is the TypeScript source and version contract for a
// scriptgo-provided module (Category 4: Node built-in modules).
type BuiltinModule struct {
	Name    string
	Version string
	Source  string
}

var (
	stdlibMu       sync.RWMutex
	loadedDir      string
	builtinModules = map[string]BuiltinModule{}
	globalsSource  string
)

// ResolveStdlibDir finds the directory containing the standard library for a version.
func ResolveStdlibDir(version string) (string, error) {
	if version == "" {
		if v := os.Getenv("SCRIPTGO_VERSION"); v != "" {
			version = v
		} else {
			version = DefaultVersion
		}
	}

	// 1. Direct override via SCRIPTGO_STDLIB_PATH
	if customPath := os.Getenv("SCRIPTGO_STDLIB_PATH"); customPath != "" {
		if info, err := os.Stat(customPath); err == nil && info.IsDir() {
			return filepath.Clean(customPath), nil
		}
	}

	// 2. User versioned directory: ~/.scriptgo/versions/<version>/stdlib
	homeDir, err := os.UserHomeDir()
	if homeDir != "" {
		versionDir := filepath.Join(homeDir, ".scriptgo", "versions", version, "stdlib")
		if info, err := os.Stat(versionDir); err == nil && info.IsDir() {
			return versionDir, nil
		}
	}

	// 3. Fallback to local workspace repository stdlib/ if present
	for _, candidate := range []string{
		"stdlib",
		"../stdlib",
		"../../stdlib",
		"../../../stdlib",
	} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			abs, err := filepath.Abs(candidate)
			if err == nil {
				// Seed into ~/.scriptgo/versions/<version>/stdlib if user home is available
				if homeDir != "" {
					versionDir := filepath.Join(homeDir, ".scriptgo", "versions", version, "stdlib")
					_ = copyDir(abs, versionDir)
					if info, err := os.Stat(versionDir); err == nil && info.IsDir() {
						return versionDir, nil
					}
				}
				return abs, nil
			}
		}
	}

	if homeDir != "" {
		versionDir := filepath.Join(homeDir, ".scriptgo", "versions", version, "stdlib")
		return "", fmt.Errorf("stdlib not found in %s or workspace repository", versionDir)
	}
	return "", fmt.Errorf("stdlib for version %s could not be resolved: %w", version, err)
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

// LoadStdlibFromDir reads all .ts / .d.ts files from the given directory into memory.
func LoadStdlibFromDir(dir string, version string) error {
	stdlibMu.Lock()
	defer stdlibMu.Unlock()

	cleanDir := filepath.Clean(dir)
	if loadedDir == cleanDir && len(builtinModules) > 0 {
		return nil
	}

	entries, err := os.ReadDir(cleanDir)
	if err != nil {
		return fmt.Errorf("read stdlib directory %s: %w", cleanDir, err)
	}

	newModules := map[string]BuiltinModule{}
	var newGlobals string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		fullPath := filepath.Join(cleanDir, name)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("read stdlib file %s: %w", fullPath, err)
		}

		if name == "globals.d.ts" {
			newGlobals = string(content)
			continue
		}

		if strings.HasSuffix(name, ".ts") && !strings.HasSuffix(name, ".d.ts") {
			modName := strings.TrimSuffix(name, ".ts")
			newModules[modName] = BuiltinModule{
				Name:    modName,
				Version: version,
				Source:  string(content),
			}
		}
	}

	builtinModules = newModules
	globalsSource = newGlobals
	loadedDir = cleanDir
	return nil
}

// EnsureStdlib ensures that standard library modules are loaded for the requested version.
func EnsureStdlib(version string) error {
	stdlibMu.RLock()
	if len(builtinModules) > 0 && globalsSource != "" && (version == "" || version == DefaultVersion) {
		stdlibMu.RUnlock()
		return nil
	}
	stdlibMu.RUnlock()

	dir, err := ResolveStdlibDir(version)
	if err != nil {
		return err
	}
	if version == "" {
		version = DefaultVersion
	}
	return LoadStdlibFromDir(dir, version)
}

// BuiltinModuleManifest returns the loaded standard library modules.
func BuiltinModuleManifest() []BuiltinModule {
	_ = EnsureStdlib("")
	stdlibMu.RLock()
	defer stdlibMu.RUnlock()

	names := make([]string, 0, len(builtinModules))
	for name := range builtinModules {
		names = append(names, name)
	}
	sort.Strings(names)

	manifest := make([]BuiltinModule, 0, len(names))
	for _, name := range names {
		manifest = append(manifest, builtinModules[name])
	}
	return manifest
}

func builtinModule(name string) (BuiltinModule, bool) {
	_ = EnsureStdlib("")
	stdlibMu.RLock()
	defer stdlibMu.RUnlock()

	canonical := strings.TrimPrefix(name, "node:")
	module, ok := builtinModules[canonical]
	return module, ok
}

