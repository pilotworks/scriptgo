package pkgmgr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DependencyType indicates which dependency field in package.json to populate.
type DependencyType int

const (
	DepTypeRegular DependencyType = iota
	DepTypeDev
	DepTypeOptional
	DepTypePeer
)

// AddOptions specifies packages and target configuration for scriptgo add.
type AddOptions struct {
	ProjectRoot  string
	Manifest     string
	Lockfile     string
	StoreRoot    string
	Registry     Registry
	Dependencies []string
	DepType      DependencyType
	Exact        bool
	Offline      bool
}

// AddResult contains metadata about the added packages and resulting lockfile.
type AddResult struct {
	Added    map[string]string // packageName -> recordedSpec
	Lockfile Lockfile
}

// ParsePackageSpec splits a package specification string into its package name
// and version/range constraint.
// Supports:
//   - "lodash" -> ("lodash", "")
//   - "lodash@^4.17.21" -> ("lodash", "^4.17.21")
//   - "lodash@4.17.21" -> ("lodash", "4.17.21")
//   - "@types/node" -> ("@types/node", "")
//   - "@types/node@20.0.0" -> ("@types/node", "20.0.0")
//   - "shared@workspace:*" -> ("shared", "workspace:*")
func ParsePackageSpec(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("empty package specification")
	}
	var name, spec string
	if strings.HasPrefix(raw, "@") {
		idx := strings.Index(raw[1:], "@")
		if idx != -1 {
			name = raw[:idx+1]
			spec = strings.TrimSpace(raw[idx+2:])
		} else {
			name = raw
		}
	} else {
		idx := strings.Index(raw, "@")
		if idx != -1 {
			name = raw[:idx]
			spec = strings.TrimSpace(raw[idx+1:])
		} else {
			name = raw
		}
	}
	if !validPackageName(name) {
		return "", "", fmt.Errorf("invalid package name %q", name)
	}
	return name, spec, nil
}

// Add resolves new dependency versions, updates package.json, and executes
// a deterministic Install to project node_modules and scriptgo-lock.json.
// If installation fails, package.json is rolled back.
func Add(options AddOptions) (AddResult, error) {
	if len(options.Dependencies) == 0 {
		return AddResult{}, fmt.Errorf("no package specified to add")
	}
	root, err := filepath.Abs(options.ProjectRoot)
	if err != nil {
		return AddResult{}, fmt.Errorf("resolve project root: %w", err)
	}
	manifestPath := options.Manifest
	if manifestPath == "" {
		manifestPath = filepath.Join(root, "package.json")
	}
	originalContent, err := os.ReadFile(manifestPath)
	if err != nil {
		return AddResult{}, fmt.Errorf("read package manifest %q: %w", manifestPath, err)
	}

	manifest, err := LoadProjectManifest(manifestPath)
	if err != nil {
		return AddResult{}, err
	}
	if manifest.Scriptgo.RegistryToken != "" {
		return AddResult{}, fmt.Errorf("package manifest must not contain a registry token; use registryTokenEnv or --registry-token")
	}

	var rawMap map[string]any
	if err := json.Unmarshal(originalContent, &rawMap); err != nil {
		return AddResult{}, fmt.Errorf("parse package manifest %q: %w", manifestPath, err)
	}

	workspaces, err := discoverWorkspaces(root, manifest.Workspaces)
	if err != nil {
		return AddResult{}, err
	}

	registry := options.Registry
	if registry.BaseURL == "" {
		registry.BaseURL = manifest.Scriptgo.Registry
	}
	if registry.Token == "" && manifest.Scriptgo.RegistryTokenEnv != "" {
		registry.Token = os.Getenv(manifest.Scriptgo.RegistryTokenEnv)
	}

	lockPath := options.Lockfile
	if lockPath == "" {
		lockPath = filepath.Join(root, "scriptgo-lock.json")
	}

	var locked Lockfile
	if options.Offline {
		locked, err = LoadLockfile(lockPath)
		if err != nil {
			return AddResult{}, fmt.Errorf("lockfile required for offline add: %w", err)
		}
	}

	resolvedSpecs := make(map[string]string, len(options.Dependencies))
	for _, rawSpec := range options.Dependencies {
		name, spec, err := ParsePackageSpec(rawSpec)
		if err != nil {
			return AddResult{}, err
		}

		recordedSpec, err := resolveAddSpec(name, spec, options, registry, workspaces, locked)
		if err != nil {
			return AddResult{}, err
		}
		resolvedSpecs[name] = recordedSpec
	}

	sectionName := "dependencies"
	switch options.DepType {
	case DepTypeDev:
		sectionName = "devDependencies"
	case DepTypeOptional:
		sectionName = "optionalDependencies"
	case DepTypePeer:
		sectionName = "peerDependencies"
	}

	allSections := []string{"dependencies", "devDependencies", "optionalDependencies", "peerDependencies"}
	for _, sec := range allSections {
		if sec != sectionName {
			if secMap, ok := rawMap[sec].(map[string]any); ok {
				for name := range resolvedSpecs {
					delete(secMap, name)
				}
			}
		}
	}

	var targetSection map[string]any
	if existing, ok := rawMap[sectionName].(map[string]any); ok {
		targetSection = existing
	} else {
		targetSection = map[string]any{}
		rawMap[sectionName] = targetSection
	}
	for name, spec := range resolvedSpecs {
		targetSection[name] = spec
	}

	topLevelKeys := getTopLevelKeyOrder(originalContent)
	topLevelKeys = insertDependencyKey(topLevelKeys, sectionName)

	newContent, err := formatPackageJSON(rawMap, topLevelKeys)
	if err != nil {
		return AddResult{}, fmt.Errorf("format updated package.json: %w", err)
	}

	if err := os.WriteFile(manifestPath, newContent, 0o644); err != nil {
		return AddResult{}, fmt.Errorf("write updated package.json: %w", err)
	}

	lock, err := Install(InstallOptions{
		ProjectRoot: root,
		Manifest:    manifestPath,
		Lockfile:    lockPath,
		StoreRoot:   options.StoreRoot,
		Registry:    registry,
		Offline:     options.Offline,
		Frozen:      false,
	})
	if err != nil {
		_ = os.WriteFile(manifestPath, originalContent, 0o644)
		return AddResult{}, fmt.Errorf("install after add: %w", err)
	}

	return AddResult{
		Added:    resolvedSpecs,
		Lockfile: lock,
	}, nil
}

func resolveAddSpec(name, spec string, options AddOptions, registry Registry, workspaces map[string]workspacePackage, locked Lockfile) (string, error) {
	// Workspace package resolution
	if workspace, ok := workspaces[name]; ok {
		if strings.HasPrefix(spec, "workspace:") {
			return spec, nil
		}
		if spec == "" {
			if options.Exact {
				return "workspace:" + workspace.Manifest.Version, nil
			}
			return "workspace:*", nil
		}
	}
	if strings.HasPrefix(spec, "workspace:") {
		return "", fmt.Errorf("workspace package %q not found", name)
	}

	// Registry package resolution
	if options.Offline {
		if lockedPkg, ok := locked.Packages[name]; ok && lockedPkg.Version != "" {
			if spec == "" {
				if options.Exact {
					return lockedPkg.Version, nil
				}
				return "^" + lockedPkg.Version, nil
			}
			return spec, nil
		}
		return "", fmt.Errorf("offline add: package %q not found in lockfile", name)
	}

	meta, err := registry.FetchPackageMetadata(name)
	if err != nil {
		return "", err
	}

	candidates := make([]Version, 0, len(meta.Versions))
	for vText := range meta.Versions {
		if v, parseErr := ParseVersion(vText); parseErr == nil {
			candidates = append(candidates, v)
		}
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("package %q has no valid semver versions in registry", name)
	}

	if spec != "" {
		hasRangePrefix := strings.HasPrefix(spec, "^") || strings.HasPrefix(spec, "~") ||
			strings.HasPrefix(spec, ">") || strings.HasPrefix(spec, "<") ||
			strings.HasPrefix(spec, "=") || strings.Contains(spec, "||")
		if hasRangePrefix {
			if _, err := SelectVersion(spec, candidates); err != nil {
				return "", fmt.Errorf("no version in registry satisfies %q for package %q: %w", spec, name, err)
			}
			return spec, nil
		}
		// Exact version like "4.17.21"
		v, parseErr := ParseVersion(spec)
		if parseErr != nil {
			return "", fmt.Errorf("invalid version %q for package %q: %w", spec, name, parseErr)
		}
		found := false
		for _, c := range candidates {
			if CompareVersions(c, v) == 0 {
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("version %q for package %q not found in registry", spec, name)
		}
		if options.Exact {
			return v.String(), nil
		}
		return "^" + v.String(), nil
	}

	// spec is empty: choose latest
	var targetVer Version
	if tag, ok := meta.DistTags["latest"]; ok && tag != "" {
		if v, parseErr := ParseVersion(tag); parseErr == nil {
			targetVer = v
		}
	}
	if targetVer == (Version{}) {
		highest, selErr := SelectVersion("*", candidates)
		if selErr != nil {
			return "", fmt.Errorf("select highest version for %q: %w", name, selErr)
		}
		targetVer = highest
	}
	if options.Exact {
		return targetVer.String(), nil
	}
	return "^" + targetVer.String(), nil
}

func getTopLevelKeyOrder(data []byte) []string {
	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil || t != json.Delim('{') {
		return nil
	}
	var keys []string
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			break
		}
		if key, ok := t.(string); ok {
			keys = append(keys, key)
			var v json.RawMessage
			if err := dec.Decode(&v); err != nil {
				break
			}
		}
	}
	return keys
}

func insertDependencyKey(keys []string, keyToInsert string) []string {
	for _, k := range keys {
		if k == keyToInsert {
			return keys
		}
	}
	order := []string{
		"name", "version", "private", "description", "type", "main", "module",
		"types", "exports", "bin", "scripts", "workspaces",
		"dependencies", "devDependencies", "peerDependencies", "optionalDependencies",
	}
	targetIdx := -1
	for i, o := range order {
		if o == keyToInsert {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		return append(keys, keyToInsert)
	}

	insertAfter := -1
	for i := targetIdx - 1; i >= 0; i-- {
		prev := order[i]
		for j, k := range keys {
			if k == prev {
				insertAfter = j
				break
			}
		}
		if insertAfter != -1 {
			break
		}
	}
	if insertAfter != -1 {
		res := make([]string, 0, len(keys)+1)
		res = append(res, keys[:insertAfter+1]...)
		res = append(res, keyToInsert)
		res = append(res, keys[insertAfter+1:]...)
		return res
	}
	return append(keys, keyToInsert)
}

func formatPackageJSON(rawMap map[string]any, keys []string) ([]byte, error) {
	var activeKeys []string
	for _, k := range keys {
		if rawMap[k] != nil {
			activeKeys = append(activeKeys, k)
		}
	}

	var b bytes.Buffer
	b.WriteString("{\n")
	for i, key := range activeKeys {
		val := rawMap[key]
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		valBytes, err := json.MarshalIndent(val, "  ", "  ")
		if err != nil {
			return nil, err
		}
		b.WriteString("  ")
		b.Write(keyBytes)
		b.WriteString(": ")
		b.Write(valBytes)
		if i < len(activeKeys)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("}\n")
	return b.Bytes(), nil
}
