package main

import (
	"fmt"
	"path/filepath"

	"github.com/pilotworks/scriptgo/internal/pkgmgr"
)

type packageInstallFlags struct {
	Install   bool
	Offline   bool
	Frozen    bool
	Registry  string
	StoreRoot string
}

func installForEntry(entryPath string, flags packageInstallFlags) error {
	if !flags.Install {
		if flags.Offline || flags.Frozen {
			return fmt.Errorf("--offline and --frozen require --install")
		}
		return nil
	}
	if entryPath == "" {
		return fmt.Errorf("--install requires a file entry point")
	}
	root := filepath.Dir(entryPath)
	if _, err := pkgmgr.Install(pkgmgr.InstallOptions{
		ProjectRoot: root,
		StoreRoot:   flags.StoreRoot,
		Registry:    pkgmgr.Registry{BaseURL: flags.Registry},
		Offline:     flags.Offline,
		Frozen:      flags.Frozen,
	}); err != nil {
		return fmt.Errorf("install dependencies: %w", err)
	}
	return nil
}
