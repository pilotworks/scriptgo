package typescriptgo

import (
	_ "embed"
	"strings"
)

// BuiltinModule is the TypeScript source and version contract for a
// scriptgo-provided module (Category 4: Node built-in modules).
type BuiltinModule struct {
	Name    string
	Version string
	Source  string
}

//go:embed stdlib/path.ts
var pathSource string

//go:embed stdlib/fs.ts
var fsSource string

//go:embed stdlib/process.ts
var processSource string

var builtinModules = map[string]BuiltinModule{
	"path": {
		Name:    "path",
		Version: "1",
		Source:  pathSource,
	},
	"fs": {
		Name:    "fs",
		Version: "1",
		Source:  fsSource,
	},
	"process": {
		Name:    "process",
		Version: "1",
		Source:  processSource,
	},
}

func BuiltinModuleManifest() []BuiltinModule {
	return []BuiltinModule{
		builtinModules["path"],
		builtinModules["fs"],
		builtinModules["process"],
	}
}

func builtinModule(name string) (BuiltinModule, bool) {
	canonical := strings.TrimPrefix(name, "node:")
	module, ok := builtinModules[canonical]
	return module, ok
}

