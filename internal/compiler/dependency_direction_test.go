package compiler

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const (
	modulePath  = "github.com/pilotworks/scriptgo/"
	adapterPath = "github.com/microsoft/TypeScript/tsc/scriptgo"
)

// allowedImports is the enforced form of the dependency direction documented
// in AGENTS.md and docs/application-structure.md. Keys and values are paths
// relative to the module root; "typescriptgo" names the pinned adapter. A new
// package must be added here (and to the documentation) before it can build.
var allowedImports = map[string][]string{
	"cmd/scriptgo":          {"internal/compiler", "internal/pkgmgr"},
	"cmd/scg":               {},
	"cmd/parity":            {"internal/audit", "internal/compiler"},
	"internal/compiler":     {"internal/backend/llvm", "internal/frontend", "internal/ir", "internal/lowering", "internal/opt", "internal/runtime"},
	"internal/frontend":     {"typescriptgo"},
	"internal/lowering":     {"internal/frontend", "internal/ir"},
	"internal/opt":          {"internal/ir"},
	"internal/backend/llvm": {"internal/ir"},
	"internal/ir":           {},
	"internal/runtime":      {},
	"internal/pkgmgr":       {},
	"internal/audit":        {"internal/frontend", "internal/spec"},
	"internal/spec":         {},
}

func TestDependencyDirection(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, top := range []string{"cmd", "internal"} {
		err := filepath.WalkDir(filepath.Join(root, top), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				name := entry.Name()
				if name == "testdata" || name == "native" || name == "typescriptgo" || name == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, _ := filepath.Rel(root, filepath.Dir(path))
			pkg := filepath.ToSlash(rel)
			found[pkg] = true
			allowed, known := allowedImports[pkg]
			if !known {
				t.Errorf("package %s is missing from the documented dependency direction", pkg)
				return nil
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, spec := range file.Imports {
				imported, _ := strconv.Unquote(spec.Path.Value)
				var dep string
				switch {
				case imported == adapterPath:
					dep = "typescriptgo"
				case strings.HasPrefix(imported, modulePath):
					dep = strings.TrimPrefix(imported, modulePath)
				default:
					continue
				}
				if !slices.Contains(allowed, dep) {
					t.Errorf("%s imports %s; allowed: %v", filepath.ToSlash(rel)+"/"+filepath.Base(path), dep, allowed)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	var stale []string
	for pkg := range allowedImports {
		if !found[pkg] {
			stale = append(stale, pkg)
		}
	}
	sort.Strings(stale)
	if len(stale) > 0 {
		t.Errorf("dependency rules name packages that no longer exist: %v", stale)
	}
}
