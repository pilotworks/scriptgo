package typescriptgo

import (
	"embed"
	"fmt"
	"io/fs"
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
	Name          string
	Version       string
	Source        string
	IsDeclaration bool
}

//go:embed stdlib/*
var embeddedStdlibFS embed.FS

var (
	stdlibMu       sync.RWMutex
	builtinModules = map[string]BuiltinModule{}
	globalsSource  string
	loaded         bool
)

func init() {
	_ = loadEmbeddedStdlib(DefaultVersion)
}

func loadEmbeddedStdlib(version string) error {
	stdlibMu.Lock()
	defer stdlibMu.Unlock()

	newModules := map[string]BuiltinModule{}
	var newGlobals string

	entries, err := embeddedStdlibFS.ReadDir("stdlib")
	if err != nil {
		return fmt.Errorf("read embedded stdlib: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirName := entry.Name()
			isDecl := false
			indexPath := filepath.Join("stdlib", dirName, "index.ts")
			content, err := embeddedStdlibFS.ReadFile(indexPath)
			if err != nil {
				indexPath = filepath.Join("stdlib", dirName, "index.d.ts")
				content, err = embeddedStdlibFS.ReadFile(indexPath)
				if err == nil {
					isDecl = true
				}
			}
			if err != nil {
				indexPath = filepath.Join("stdlib", dirName, dirName+".ts")
				content, err = embeddedStdlibFS.ReadFile(indexPath)
			}
			if err == nil {
				newModules[dirName] = BuiltinModule{
					Name:          dirName,
					Version:       version,
					Source:        string(content),
					IsDeclaration: isDecl,
				}
			}
			continue
		}
		name := entry.Name()
		content, err := embeddedStdlibFS.ReadFile(filepath.Join("stdlib", name))
		if err != nil {
			return fmt.Errorf("read embedded stdlib file %s: %w", name, err)
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
	loaded = true
	return nil
}

// EnsureStdlib ensures that standard library modules are loaded in memory.
func EnsureStdlib(version string) error {
	stdlibMu.RLock()
	if loaded && len(builtinModules) > 0 && globalsSource != "" {
		stdlibMu.RUnlock()
		return nil
	}
	stdlibMu.RUnlock()

	// 1. Direct override via SCRIPTGO_STDLIB_PATH if set
	if customPath := os.Getenv("SCRIPTGO_STDLIB_PATH"); customPath != "" {
		if info, err := os.Stat(customPath); err == nil && info.IsDir() {
			return LoadStdlibFromDir(customPath, version)
		}
	}

	if version == "" {
		version = DefaultVersion
	}
	return loadEmbeddedStdlib(version)
}

// LoadStdlibFromDir reads .ts / .d.ts files from the given directory into memory (for local overrides).
func LoadStdlibFromDir(dir string, version string) error {
	stdlibMu.Lock()
	defer stdlibMu.Unlock()

	cleanDir := filepath.Clean(dir)
	entries, err := os.ReadDir(cleanDir)
	if err != nil {
		return fmt.Errorf("read stdlib directory %s: %w", cleanDir, err)
	}

	newModules := map[string]BuiltinModule{}
	var newGlobals string

	for _, entry := range entries {
		if entry.IsDir() {
			dirName := entry.Name()
			isDecl := false
			indexPath := filepath.Join(cleanDir, dirName, "index.ts")
			content, err := os.ReadFile(indexPath)
			if err != nil {
				indexPath = filepath.Join(cleanDir, dirName, "index.d.ts")
				content, err = os.ReadFile(indexPath)
				if err == nil {
					isDecl = true
				}
			}
			if err != nil {
				indexPath = filepath.Join(cleanDir, dirName, dirName+".ts")
				content, err = os.ReadFile(indexPath)
			}
			if err == nil {
				newModules[dirName] = BuiltinModule{
					Name:          dirName,
					Version:       version,
					Source:        string(content),
					IsDeclaration: isDecl,
				}
			}
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
	loaded = true
	return nil
}

// SeedVersionDeclarations writes the standard library definition files for a version
// to the target directory (e.g. ~/.scriptgo/versions/<version>/ or custom target)
// for IDE IntelliSense support.
func SeedVersionDeclarations(version string, targetDir string) error {
	if version == "" {
		version = DefaultVersion
	}
	if targetDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil || homeDir == "" {
			return fmt.Errorf("resolve user home directory: %w", err)
		}
		targetDir = filepath.Join(homeDir, ".scriptgo", "versions", version)
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create version directory %s: %w", targetDir, err)
	}

	return fs.WalkDir(embeddedStdlibFS, "stdlib", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel("stdlib", path)
		if err != nil || rel == "." {
			return nil
		}
		dstPath := filepath.Join(targetDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}
		content, err := embeddedStdlibFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded file %s: %w", path, err)
		}
		return os.WriteFile(dstPath, content, 0o644)
	})
}

// ResolveVersionDir returns the version directory containing .d.ts declarations,
// automatically seeding it with .d.ts files if needed.
func ResolveVersionDir(version string) (string, error) {
	if version == "" {
		if v := os.Getenv("SCRIPTGO_VERSION"); v != "" {
			version = v
		} else {
			version = DefaultVersion
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	versionDir := filepath.Join(homeDir, ".scriptgo", "versions", version)
	if err := SeedVersionDeclarations(version, versionDir); err != nil {
		return "", err
	}
	return versionDir, nil
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
	if canonical == "stream/web" {
		canonical = "webstreams"
	} else if canonical == "stream/consumers" {
		canonical = "stream_consumers"
	}
	module, ok := builtinModules[canonical]
	return module, ok
}

// EmbeddedFS returns the embedded filesystem containing stdlib sources and declarations.
func EmbeddedFS() fs.FS {
	return embeddedStdlibFS
}
