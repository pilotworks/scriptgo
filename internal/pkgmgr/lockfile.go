package pkgmgr

import (
	"encoding/json"
	"fmt"
	"os"
)

// Lockfile is the deterministic, offline package graph contract.
type Lockfile struct {
	LockfileVersion int                      `json:"lockfileVersion"`
	Packages        map[string]LockedPackage `json:"packages"`
}

// LockedPackage records the resolved package metadata needed for later fetch
// and CAS installation stages.
type LockedPackage struct {
	Version      string            `json:"version,omitempty"`
	Resolved     string            `json:"resolved,omitempty"`
	Integrity    string            `json:"integrity,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

// NewLockfile creates a stable lockfile from package metadata keyed by name.
func NewLockfile(manifests map[string]PackageManifest) (Lockfile, error) {
	lock := Lockfile{LockfileVersion: 1, Packages: make(map[string]LockedPackage, len(manifests))}
	for name, manifest := range manifests {
		if err := ValidateManifest(manifest, name); err != nil {
			return Lockfile{}, err
		}
		dependencies := make(map[string]string, len(manifest.Dependencies))
		for dependency, spec := range manifest.Dependencies {
			dependencies[dependency] = spec
		}
		lock.Packages[name] = LockedPackage{Version: manifest.Version, Dependencies: dependencies}
	}
	return lock, nil
}

// LoadLockfile reads and validates a lockfile.
func LoadLockfile(path string) (Lockfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Lockfile{}, fmt.Errorf("read lockfile %q: %w", path, err)
	}
	var lock Lockfile
	if err := json.Unmarshal(data, &lock); err != nil {
		return Lockfile{}, fmt.Errorf("parse lockfile %q: %w", path, err)
	}
	if err := ValidateLockfile(lock); err != nil {
		return Lockfile{}, fmt.Errorf("validate lockfile %q: %w", path, err)
	}
	return lock, nil
}

// WriteLockfile writes canonical indented JSON. encoding/json sorts map keys,
// making output stable across runs and platforms.
func WriteLockfile(path string, lock Lockfile) error {
	if err := ValidateLockfile(lock); err != nil {
		return err
	}
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("encode lockfile: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write lockfile %q: %w", path, err)
	}
	return nil
}

// ValidateLockfile checks the schema contract without resolving network data.
func ValidateLockfile(lock Lockfile) error {
	if lock.LockfileVersion != 1 {
		return fmt.Errorf("unsupported lockfile version %d", lock.LockfileVersion)
	}
	if lock.Packages == nil {
		return fmt.Errorf("packages must be present")
	}
	for name, pkg := range lock.Packages {
		if name == "" {
			return fmt.Errorf("package name must not be empty")
		}
		for dependency, spec := range pkg.Dependencies {
			if dependency == "" || spec == "" {
				return fmt.Errorf("package %q has invalid dependency", name)
			}
		}
	}
	return nil
}
