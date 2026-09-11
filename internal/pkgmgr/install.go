package pkgmgr

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// InstallOptions controls a deterministic package installation.
type InstallOptions struct {
	ProjectRoot string
	Manifest    string
	Lockfile    string
	StoreRoot   string
	Registry    Registry
	Offline     bool
	Frozen      bool
}

type packageNode struct {
	Name     string
	Version  string
	Manifest PackageManifest
	Deps     map[string]string
}

// Install resolves production dependencies, verifies and caches their
// tarballs, then creates an isolated node_modules tree using hardlinks.
// Lifecycle scripts are intentionally not executed.
func Install(options InstallOptions) (Lockfile, error) {
	root, err := filepath.Abs(options.ProjectRoot)
	if err != nil {
		return Lockfile{}, fmt.Errorf("resolve project root: %w", err)
	}
	manifestPath := options.Manifest
	if manifestPath == "" {
		manifestPath = filepath.Join(root, "package.json")
	}
	manifest, err := LoadProjectManifest(manifestPath)
	if err != nil {
		return Lockfile{}, err
	}
	lockPath := options.Lockfile
	if lockPath == "" {
		lockPath = filepath.Join(root, "scriptgo-lock.json")
	}
	store := Store{Root: options.StoreRoot}
	if store.Root == "" {
		store.Root = filepath.Join(root, ".scriptgo", "store")
	}
	var locked Lockfile
	if options.Offline || options.Frozen {
		locked, err = LoadLockfile(lockPath)
		if err != nil {
			return Lockfile{}, fmt.Errorf("lockfile required for frozen/offline install: %w", err)
		}
	}
	if options.Frozen {
		if err := validateProjectLock(manifest, locked); err != nil {
			return Lockfile{}, err
		}
	}
	nodes := map[string]*packageNode{}
	metadata := map[string]map[string]PackageManifest{}
	edges := map[string]map[string]string{}
	parents := map[string]string{}
	var resolve func(string, string, string) (string, error)
	resolve = func(name, spec, parent string) (string, error) {
		if parent != "" {
			if edge := edges[parent][name]; edge != "" {
				return edge, nil
			}
		}
		versions := metadata[name]
		if versions == nil {
			if options.Offline || options.Frozen {
				versions = lockedVersions(name, locked)
			} else {
				versions, err = options.Registry.FetchPackage(name)
				if err != nil {
					return "", err
				}
			}
			metadata[name] = versions
		}
		candidates := make([]Version, 0, len(versions))
		byVersion := map[Version]PackageManifest{}
		for versionText, candidate := range versions {
			version, parseErr := ParseVersion(versionText)
			if parseErr == nil {
				candidates = append(candidates, version)
				byVersion[version] = candidate
			}
		}
		selected, err := SelectVersion(spec, candidates)
		if err != nil {
			return "", fmt.Errorf("resolve %s@%s: %w", name, spec, err)
		}
		identity := name + "@" + selected.String()
		if parent != "" && edges[parent] == nil {
			edges[parent] = map[string]string{}
		}
		if parent != "" {
			edges[parent][name] = identity
			parents[identity] = parent
		}
		if _, exists := nodes[identity]; exists {
			return identity, nil
		}
		candidate := byVersion[selected]
		candidate.Name, candidate.Version = name, selected.String()
		deps := mergedDependencies(candidate)
		nodes[identity] = &packageNode{Name: name, Version: selected.String(), Manifest: candidate, Deps: deps}
		edges[identity] = map[string]string{}
		for dependency, dependencySpec := range deps {
			if _, err := resolve(dependency, dependencySpec, identity); err != nil {
				if isOptional(candidate, dependency) {
					continue
				}
				return "", err
			}
		}
		return identity, nil
	}
	rootEdges := map[string]string{}
	for name, spec := range manifest.Dependencies {
		identity, resolveErr := resolve(name, spec, "")
		if resolveErr != nil {
			return Lockfile{}, resolveErr
		}
		rootEdges[name] = identity
	}
	for name, spec := range manifest.OptionalDependencies {
		identity, resolveErr := resolve(name, spec, "")
		if resolveErr == nil {
			rootEdges[name] = identity
		}
	}
	if err := validatePeerDependencies(nodes, edges, parents, rootEdges); err != nil {
		return Lockfile{}, err
	}
	lock := graphLockfile(nodes, manifest)
	for identity, node := range nodes {
		if err := materializePackage(store, options.Registry, options.Offline, node.Manifest); err != nil {
			if removeOptionalNode(identity, node, nodes, edges, rootEdges, manifest) {
				continue
			}
			return Lockfile{}, fmt.Errorf("install %s: %w", identity, err)
		}
	}
	pruneGraph(nodes, edges, rootEdges)
	lock = graphLockfile(nodes, manifest)
	if err := commitInstall(root, lockPath, lock, store, nodes, edges, rootEdges); err != nil {
		return Lockfile{}, err
	}
	return lock, nil
}

func mergedDependencies(manifest PackageManifest) map[string]string {
	deps := map[string]string{}
	for name, spec := range manifest.Dependencies {
		deps[name] = spec
	}
	for name, spec := range manifest.OptionalDependencies {
		deps[name] = spec
	}
	return deps
}

func isOptional(manifest PackageManifest, name string) bool {
	_, ok := manifest.OptionalDependencies[name]
	return ok
}

func lockedVersions(name string, lock Lockfile) map[string]PackageManifest {
	versions := map[string]PackageManifest{}
	for identity, pkg := range lock.Packages {
		if (identity == name || strings.HasPrefix(identity, name+"@")) && pkg.Version != "" {
			versions[pkg.Version] = PackageManifest{
				Name: name, Version: pkg.Version, Dist: PackageDist{Tarball: pkg.Resolved, Integrity: pkg.Integrity},
				Dependencies: pkg.Dependencies, OptionalDependencies: pkg.OptionalDependencies,
				PeerDependencies: pkg.PeerDependencies, Bin: pkg.Bin,
			}
		}
	}
	return versions
}

func validatePeerDependencies(nodes map[string]*packageNode, edges map[string]map[string]string, parents, rootEdges map[string]string) error {
	for identity, node := range nodes {
		for name, spec := range node.Manifest.PeerDependencies {
			candidate := ""
			for parent := parents[identity]; parent != ""; parent = parents[parent] {
				candidate = edges[parent][name]
				if candidate != "" {
					break
				}
			}
			if candidate == "" {
				candidate = rootEdges[name]
			}
			peer := nodes[candidate]
			if peer == nil {
				return fmt.Errorf("package %s requires peer %s@%s", identity, name, spec)
			}
			version, err := ParseVersion(peer.Version)
			if err != nil || !Satisfies(version, spec) {
				return fmt.Errorf("package %s requires peer %s@%s, found %s@%s", identity, name, spec, name, peer.Version)
			}
		}
	}
	return nil
}

func graphLockfile(nodes map[string]*packageNode, project PackageManifest) Lockfile {
	lock := Lockfile{LockfileVersion: 1, Project: LockfileProject{
		Dependencies:         cloneStrings(project.Dependencies),
		OptionalDependencies: cloneStrings(project.OptionalDependencies),
		PeerDependencies:     cloneStrings(project.PeerDependencies),
	}, Packages: map[string]LockedPackage{}}
	for identity, node := range nodes {
		lock.Packages[identity] = LockedPackage{
			Version: node.Version, Resolved: node.Manifest.Dist.Tarball, Integrity: node.Manifest.Dist.Integrity,
			Dependencies: cloneStrings(node.Manifest.Dependencies), OptionalDependencies: cloneStrings(node.Manifest.OptionalDependencies),
			PeerDependencies: cloneStrings(node.Manifest.PeerDependencies), Bin: node.Manifest.Bin,
		}
	}
	return lock
}

func cloneStrings(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func validateProjectLock(manifest PackageManifest, lock Lockfile) error {
	if !equalStrings(manifest.Dependencies, lock.Project.Dependencies) ||
		!equalStrings(manifest.OptionalDependencies, lock.Project.OptionalDependencies) ||
		!equalStrings(manifest.PeerDependencies, lock.Project.PeerDependencies) {
		return fmt.Errorf("frozen install rejected: project manifest dependencies differ from lockfile")
	}
	return nil
}

func equalStrings(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func materializePackage(store Store, registry Registry, offline bool, manifest PackageManifest) error {
	var archive []byte
	var err error
	if offline {
		archive, err = cachedArchive(store, manifest.Dist.Integrity)
	} else {
		if manifest.Dist.Tarball == "" {
			return fmt.Errorf("package has no tarball URL")
		}
		archive, err = registry.download(manifest.Dist.Tarball)
	}
	if err != nil {
		return err
	}
	if manifest.Dist.Integrity != "" {
		if err := VerifyIntegrity(archive, manifest.Dist.Integrity); err != nil {
			return err
		}
	}
	if _, err := store.Put(archive); err != nil {
		return err
	}
	destination := filepath.Join(store.Root, "packages", manifest.Name, manifest.Version)
	if info, err := os.Stat(destination); err == nil && info.IsDir() {
		marker, markerErr := os.ReadFile(filepath.Join(destination, ".scriptgo-integrity"))
		if markerErr == nil && strings.TrimSpace(string(marker)) != manifest.Dist.Integrity {
			return fmt.Errorf("package %s@%s already exists with different integrity", manifest.Name, manifest.Version)
		}
		return nil
	}
	packagesRoot := filepath.Dir(destination)
	if err := os.MkdirAll(packagesRoot, 0o755); err != nil {
		return err
	}
	staging, err := os.MkdirTemp(packagesRoot, ".package-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	if err := extractTarGz(archive, staging); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(staging, ".scriptgo-integrity"), []byte(manifest.Dist.Integrity+"\n"), 0o644); err != nil {
		return err
	}
	if err := os.Rename(staging, destination); err != nil {
		if info, statErr := os.Stat(destination); statErr == nil && info.IsDir() {
			return nil
		}
		return fmt.Errorf("publish extracted package: %w", err)
	}
	return nil
}

func cachedArchive(store Store, integrity string) ([]byte, error) {
	const prefix = "sha512-"
	if !strings.HasPrefix(integrity, prefix) {
		return nil, fmt.Errorf("offline package has no usable integrity")
	}
	digest, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(integrity, prefix))
	if err != nil {
		return nil, err
	}
	if len(digest) != sha512.Size {
		return nil, fmt.Errorf("offline package has invalid integrity")
	}
	return store.Read(hex.EncodeToString(digest))
}

func extractTarGz(data []byte, destination string) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gz.Close()
	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := filepath.Clean(filepath.FromSlash(strings.TrimPrefix(header.Name, "package/")))
		if name == "." || filepath.IsAbs(name) || strings.HasPrefix(name, ".."+string(filepath.Separator)) || header.Size < 0 || header.Size > 256<<20 {
			return fmt.Errorf("archive path traversal: %q", header.Name)
		}
		path := filepath.Join(destination, name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			mode := os.FileMode(header.Mode) & 0o777
			if mode == 0 {
				mode = 0o644
			}
			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				return err
			}
			written, copyErr := io.CopyN(file, reader, header.Size)
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			if written != header.Size {
				return fmt.Errorf("archive entry %q is truncated", header.Name)
			}
		default:
			return fmt.Errorf("unsupported archive entry %q", header.Name)
		}
	}
}

func removeOptionalNode(identity string, node *packageNode, nodes map[string]*packageNode, edges map[string]map[string]string, rootEdges map[string]string, project PackageManifest) bool {
	rootReferences := make([]string, 0)
	for name, candidate := range rootEdges {
		if candidate == identity {
			rootReferences = append(rootReferences, name)
			if _, required := project.Dependencies[name]; required {
				return false
			}
		}
	}
	if len(rootReferences) > 0 {
		for _, name := range rootReferences {
			delete(rootEdges, name)
		}
		delete(nodes, identity)
		return true
	}
	removed := false
	for parent, dependencies := range edges {
		for name, candidate := range dependencies {
			if candidate == identity {
				parentNode := nodes[parent]
				if parentNode == nil || !isOptionalName(name, node.Name, parentNode.Manifest.OptionalDependencies) {
					return false
				}
				delete(dependencies, name)
				removed = true
			}
		}
	}
	if removed {
		delete(nodes, identity)
	}
	return removed
}

func isOptionalName(name, target string, optional map[string]string) bool {
	if optional != nil {
		_, ok := optional[name]
		return ok
	}
	return name == target
}

func pruneGraph(nodes map[string]*packageNode, edges map[string]map[string]string, rootEdges map[string]string) {
	reachable := map[string]bool{}
	var visit func(string)
	visit = func(identity string) {
		if reachable[identity] {
			return
		}
		reachable[identity] = true
		for _, dependency := range edges[identity] {
			visit(dependency)
		}
	}
	for _, identity := range rootEdges {
		visit(identity)
	}
	for identity := range nodes {
		if !reachable[identity] {
			delete(nodes, identity)
			delete(edges, identity)
		}
	}
}

func commitInstall(root, lockPath string, lock Lockfile, store Store, nodes map[string]*packageNode, edges map[string]map[string]string, rootEdges map[string]string) error {
	staging, err := os.MkdirTemp(root, ".scriptgo-install-")
	if err != nil {
		return fmt.Errorf("create install staging directory: %w", err)
	}
	defer os.RemoveAll(staging)
	stagedModules := filepath.Join(staging, "node_modules")
	if err := linkGraphAt(stagedModules, store, nodes, edges, rootEdges); err != nil {
		return err
	}
	stagedLock := filepath.Join(staging, "scriptgo-lock.json")
	if err := WriteLockfile(stagedLock, lock); err != nil {
		return err
	}
	modules := filepath.Join(root, "node_modules")
	backup := filepath.Join(staging, "node_modules.old")
	if _, err := os.Stat(modules); err == nil {
		if err := os.Rename(modules, backup); err != nil {
			return fmt.Errorf("stage existing node_modules: %w", err)
		}
	}
	if err := os.Rename(stagedModules, modules); err != nil {
		_ = os.Rename(backup, modules)
		return fmt.Errorf("commit node_modules: %w", err)
	}
	lockBackup := lockPath + ".scriptgo-old"
	if _, err := os.Stat(lockPath); err == nil {
		if err := os.Rename(lockPath, lockBackup); err != nil {
			_ = os.RemoveAll(modules)
			_ = os.Rename(backup, modules)
			return fmt.Errorf("stage existing lockfile: %w", err)
		}
	}
	if err := os.Rename(stagedLock, lockPath); err != nil {
		_ = os.Remove(lockPath)
		_ = os.Rename(lockBackup, lockPath)
		_ = os.RemoveAll(modules)
		_ = os.Rename(backup, modules)
		return fmt.Errorf("commit lockfile: %w", err)
	}
	_ = os.Remove(lockBackup)
	_ = os.RemoveAll(backup)
	return nil
}

func linkGraph(root string, store Store, nodes map[string]*packageNode, edges map[string]map[string]string, rootEdges map[string]string) error {
	return linkGraphAt(filepath.Join(root, "node_modules"), store, nodes, edges, rootEdges)
}

func linkGraphAt(nodeModules string, store Store, nodes map[string]*packageNode, edges map[string]map[string]string, rootEdges map[string]string) error {
	if err := os.MkdirAll(nodeModules, 0o755); err != nil {
		return err
	}
	for name, identity := range rootEdges {
		if err := linkNode(store, nodes, edges, identity, filepath.Join(nodeModules, filepath.FromSlash(name)), map[string]bool{}); err != nil {
			return fmt.Errorf("link %s: %w", name, err)
		}
	}
	if err := linkBins(nodeModules, nodes, rootEdges); err != nil {
		return err
	}
	return nil
}

func linkNode(store Store, nodes map[string]*packageNode, edges map[string]map[string]string, identity, destination string, stack map[string]bool) error {
	node := nodes[identity]
	if node == nil {
		return fmt.Errorf("missing graph node %q", identity)
	}
	if stack[identity] {
		// A cyclic back edge resolves through an ancestor node_modules directory.
		return nil
	}
	stack[identity] = true
	defer delete(stack, identity)
	if err := os.RemoveAll(destination); err != nil {
		return err
	}
	source := filepath.Join(store.Root, "packages", node.Name, node.Version)
	if err := linkTree(source, destination); err != nil {
		return err
	}
	for name, dependency := range edges[identity] {
		if err := linkNode(store, nodes, edges, dependency, filepath.Join(destination, "node_modules", filepath.FromSlash(name)), stack); err != nil {
			return err
		}
	}
	return nil
}

func linkTree(source, destination string) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() == ".scriptgo-integrity" {
			continue
		}
		src, dst := filepath.Join(source, entry.Name()), filepath.Join(destination, entry.Name())
		if entry.IsDir() {
			if err := linkTree(src, dst); err != nil {
				return err
			}
			continue
		}
		if err := os.Link(src, dst); err != nil {
			return fmt.Errorf("hardlink %q: %w", dst, err)
		}
	}
	return nil
}

func linkBins(nodeModules string, nodes map[string]*packageNode, rootEdges map[string]string) error {
	binDir := filepath.Join(nodeModules, ".bin")
	for _, identity := range rootEdges {
		node := nodes[identity]
		if node == nil {
			continue
		}
		for name, target := range node.Manifest.Bin {
			if name == "" {
				name = node.Name
			}
			if filepath.Base(name) != name || filepath.IsAbs(target) || strings.HasPrefix(filepath.Clean(filepath.FromSlash(target)), ".."+string(filepath.Separator)) {
				return fmt.Errorf("invalid bin declaration for %s", node.Name)
			}
			if err := os.MkdirAll(binDir, 0o755); err != nil {
				return err
			}
			if err := linkBin(binDir, name, node.Name, target); err != nil {
				return err
			}
		}
	}
	return nil
}

func linkBin(binDir, name, packageName, target string) error {
	linkTarget := filepath.Join("..", filepath.FromSlash(packageName), filepath.FromSlash(target))
	if runtime.GOOS != "windows" {
		if err := os.Symlink(linkTarget, filepath.Join(binDir, name)); err != nil {
			return fmt.Errorf("link bin %s: %w", name, err)
		}
		return nil
	}
	cmdTarget := strings.ReplaceAll(linkTarget, "/", "\\")
	relativeTarget := strings.TrimPrefix(cmdTarget, `..\`)
	cmd := fmt.Sprintf("@echo off\r\n\"%%~dp0\\%s\" %%*\r\n", relativeTarget)
	if err := os.WriteFile(filepath.Join(binDir, name+".cmd"), []byte(cmd), 0o755); err != nil {
		return fmt.Errorf("write bin shim %s: %w", name, err)
	}
	powershellTarget := strings.ReplaceAll(relativeTarget, `\`, "/")
	powershell := fmt.Sprintf("& (Join-Path $PSScriptRoot '%s') $args\r\n", powershellTarget)
	if err := os.WriteFile(filepath.Join(binDir, name+".ps1"), []byte(powershell), 0o755); err != nil {
		return fmt.Errorf("write PowerShell bin shim %s: %w", name, err)
	}
	return nil
}
