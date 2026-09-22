package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pilotworks/scriptgo/internal/pkgmgr"
)

func handleInstall(args []string) {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.Usage = printInstallUsage
	project := fs.String("project", ".", "project directory containing package.json")
	manifest := fs.String("manifest", "", "package manifest path (default: <project>/package.json)")
	lockfile := fs.String("lockfile", "", "lockfile path (default: <project>/scriptgo-lock.json)")
	store := fs.String("store", "", "content store path (default: <project>/.scriptgo/store)")
	registry := fs.String("registry", "", "npm-compatible registry URL")
	token := fs.String("registry-token", "", "registry bearer token (default: SCRIPTGO_NPM_TOKEN or NPM_TOKEN)")
	offline := fs.Bool("offline", false, "install only from the existing lockfile and content store")
	frozen := fs.Bool("frozen", false, "use exact versions and metadata from the existing lockfile")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}
	root, err := filepath.Abs(*project)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scriptgo: %v\n", err)
		os.Exit(1)
	}
	lock, err := pkgmgr.Install(pkgmgr.InstallOptions{
		ProjectRoot: root,
		Manifest:    *manifest,
		Lockfile:    *lockfile,
		StoreRoot:   *store,
		Registry:    pkgmgr.Registry{BaseURL: *registry, Token: *token},
		Offline:     *offline,
		Frozen:      *frozen,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "scriptgo install: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "installed %d packages\n", len(lock.Packages))
}

type packageInstallFlags struct {
	Install   bool
	Offline   bool
	Frozen    bool
	Registry  string
	Token     string
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
		Registry:    pkgmgr.Registry{BaseURL: flags.Registry, Token: flags.Token},
		Offline:     flags.Offline,
		Frozen:      flags.Frozen,
	}); err != nil {
		return fmt.Errorf("install dependencies: %w", err)
	}
	return nil
}
