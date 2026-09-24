package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pilotworks/scriptgo/internal/pkgmgr"
)

func handleAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.Usage = printAddUsage
	project := fs.String("project", ".", "project directory containing package.json")
	manifest := fs.String("manifest", "", "package manifest path (default: <project>/package.json)")
	lockfile := fs.String("lockfile", "", "lockfile path (default: <project>/scriptgo-lock.json)")
	store := fs.String("store", "", "content store path (default: <project>/.scriptgo/store)")
	registry := fs.String("registry", "", "npm-compatible registry URL")
	token := fs.String("registry-token", "", "registry bearer token (default: SCRIPTGO_NPM_TOKEN or NPM_TOKEN)")
	dev := fs.Bool("dev", false, "save package to devDependencies")
	fs.BoolVar(dev, "D", false, "save package to devDependencies (shorthand)")
	optional := fs.Bool("optional", false, "save package to optionalDependencies")
	fs.BoolVar(optional, "O", false, "save package to optionalDependencies (shorthand)")
	peer := fs.Bool("peer", false, "save package to peerDependencies")
	exact := fs.Bool("exact", false, "save exact version instead of ^x.y.z")
	fs.BoolVar(exact, "E", false, "save exact version instead of ^x.y.z (shorthand)")
	offline := fs.Bool("offline", false, "install only from the existing lockfile and content store")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	depCount := 0
	if *dev {
		depCount++
	}
	if *optional {
		depCount++
	}
	if *peer {
		depCount++
	}
	if depCount > 1 {
		fmt.Fprintf(os.Stderr, "scriptgo add: --dev, --optional, and --peer are mutually exclusive\n")
		os.Exit(2)
	}

	packages := fs.Args()
	if len(packages) == 0 {
		fmt.Fprintf(os.Stderr, "scriptgo add: at least one package must be specified\n\n")
		printAddUsage()
		os.Exit(2)
	}

	depType := pkgmgr.DepTypeRegular
	if *dev {
		depType = pkgmgr.DepTypeDev
	} else if *optional {
		depType = pkgmgr.DepTypeOptional
	} else if *peer {
		depType = pkgmgr.DepTypePeer
	}

	root, err := filepath.Abs(*project)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scriptgo: %v\n", err)
		os.Exit(1)
	}

	result, err := pkgmgr.Add(pkgmgr.AddOptions{
		ProjectRoot:  root,
		Manifest:     *manifest,
		Lockfile:     *lockfile,
		StoreRoot:    *store,
		Registry:     pkgmgr.Registry{BaseURL: *registry, Token: *token},
		Dependencies: packages,
		DepType:      depType,
		Exact:        *exact,
		Offline:      *offline,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "scriptgo add: %v\n", err)
		os.Exit(1)
	}

	if len(result.Added) == 1 {
		for name, spec := range result.Added {
			fmt.Fprintf(os.Stdout, "added %s@%s, installed %d packages\n", name, spec, len(result.Lockfile.Packages))
		}
	} else {
		fmt.Fprintf(os.Stdout, "added %d packages, installed %d packages\n", len(result.Added), len(result.Lockfile.Packages))
	}
}

// normalizeFlagsFirstForAdd moves flags to the front so that positional packages
// can appear before, after, or between flags.
func normalizeFlagsFirstForAdd(args []string) []string {
	var flags []string
	var positionals []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if (arg == "-project" || arg == "--project" || arg == "-manifest" || arg == "--manifest" ||
				arg == "-lockfile" || arg == "--lockfile" || arg == "-store" || arg == "--store" ||
				arg == "-registry" || arg == "--registry" || arg == "-registry-token" || arg == "--registry-token") &&
				i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
		} else {
			positionals = append(positionals, arg)
		}
	}
	return append(flags, positionals...)
}
