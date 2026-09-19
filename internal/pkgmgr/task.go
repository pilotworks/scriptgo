package pkgmgr

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TaskDefinition describes a runnable script from a package manifest.
type TaskDefinition struct {
	Name         string
	Command      string
	ProjectRoot  string
	ManifestPath string
}

// FindPackageManifest searches for a package.json file starting in startDir and
// traversing upward through ancestor directories.
func FindPackageManifest(startDir string) (string, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolve start directory %q: %w", startDir, err)
	}
	for dir := abs; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "package.json")
		if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("no package.json found in %q or any ancestor directory", startDir)
}

// ListScripts returns all scripts defined in the specified manifest.
func ListScripts(manifestPath string) (map[string]string, error) {
	manifest, err := LoadProjectManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	if len(manifest.Scripts) == 0 {
		return map[string]string{}, nil
	}
	scripts := make(map[string]string, len(manifest.Scripts))
	for k, v := range manifest.Scripts {
		scripts[k] = v
	}
	return scripts, nil
}

// ResolveTask finds a script by name from startDir (or explicit manifestPath).
func ResolveTask(startDir, manifestPath, scriptName string) (*TaskDefinition, error) {
	path := manifestPath
	var err error
	if path == "" {
		path, err = FindPackageManifest(startDir)
		if err != nil {
			return nil, err
		}
	} else {
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, err
		}
	}

	manifest, err := LoadProjectManifest(path)
	if err != nil {
		return nil, err
	}

	projectRoot := filepath.Dir(path)
	if scriptName == "" {
		return nil, fmt.Errorf("missing script name")
	}

	cmd, ok := manifest.Scripts[scriptName]
	if !ok {
		available := make([]string, 0, len(manifest.Scripts))
		for name := range manifest.Scripts {
			available = append(available, name)
		}
		sort.Strings(available)
		if len(available) == 0 {
			return nil, fmt.Errorf("package.json at %q has no scripts defined", path)
		}
		return nil, fmt.Errorf("script %q not found in %q (available: %s)", scriptName, path, strings.Join(available, ", "))
	}

	return &TaskDefinition{
		Name:         scriptName,
		Command:      cmd,
		ProjectRoot:  projectRoot,
		ManifestPath: path,
	}, nil
}

// AncestorBinPaths collects node_modules/.bin directories starting from projectRoot
// up to the file system root. The projectRoot's node_modules/.bin is always included first.
func AncestorBinPaths(projectRoot string) []string {
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		abs = filepath.Clean(projectRoot)
	}
	seen := map[string]bool{}
	var binDirs []string

	rootBin := filepath.Join(abs, "node_modules", ".bin")
	binDirs = append(binDirs, rootBin)
	seen[rootBin] = true

	for dir := abs; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "node_modules", ".bin")
		if !seen[candidate] {
			if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
				binDirs = append(binDirs, candidate)
				seen[candidate] = true
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return binDirs
}

// BuildTaskEnv returns a copy of baseEnv with PATH prepended with ancestor
// node_modules/.bin directories, plus npm-compatible lifecycle variables.
func BuildTaskEnv(projectRoot, scriptName string, baseEnv []string) []string {
	binDirs := AncestorBinPaths(projectRoot)
	pathKey := "PATH"
	var existingPath string
	var otherEnv []string

	for _, e := range baseEnv {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "PATH") {
			pathKey = parts[0]
			existingPath = parts[1]
		} else {
			otherEnv = append(otherEnv, e)
		}
	}

	allBins := append(binDirs, existingPath)
	var validParts []string
	for _, p := range allBins {
		if p != "" {
			validParts = append(validParts, p)
		}
	}
	newPath := strings.Join(validParts, string(os.PathListSeparator))

	env := append(otherEnv,
		pathKey+"="+newPath,
		"npm_lifecycle_event="+scriptName,
	)
	if cwd, err := os.Getwd(); err == nil {
		env = append(env, "INIT_CWD="+cwd)
	}
	return env
}
