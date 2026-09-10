package ir

import (
	"strings"
	"testing"
)

func TestVerifyRejectsDynamicDependencyOutsideClosedGraph(t *testing.T) {
	module := Module{
		DynamicModules: []DynamicModule{{
			Path:    "/pkg/index.js",
			Source:  `import "./missing.js";`,
			Kind:    "esm",
			Imports: []DynamicImport{{Specifier: "./missing.js", Path: "/pkg/missing.js"}},
		}},
		Functions: []Function{{Name: "main", ReturnType: TypeVoid, Body: []Instruction{{Op: OpReturn, Type: TypeVoid}}}},
	}

	err := module.Verify()
	if err == nil || !strings.Contains(err.Error(), "invalid dependency") {
		t.Fatalf("Verify error = %v, want closed graph dependency error", err)
	}
}
