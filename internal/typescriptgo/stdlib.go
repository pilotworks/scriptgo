package typescriptgo

import _ "embed"

// BuiltinModule is the TypeScript source and version contract for a
// scriptgo-provided module.
type BuiltinModule struct {
	Name    string
	Version string
	Source  string
}

//go:embed stdlib/path.ts
var pathSource string

var builtinModules = map[string]BuiltinModule{
	"path": {
		Name:    "path",
		Version: "1",
		Source:  pathSource,
	},
}

func BuiltinModuleManifest() []BuiltinModule {
	return []BuiltinModule{builtinModules["path"]}
}

func builtinModule(name string) (BuiltinModule, bool) {
	module, ok := builtinModules[name]
	return module, ok
}
