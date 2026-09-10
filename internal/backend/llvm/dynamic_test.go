package llvm

import (
	"strings"
	"testing"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func TestEmitDynamicCallRegistersClosedModuleGraph(t *testing.T) {
	module := ir.Module{
		DynamicModules: []ir.DynamicModule{
			{
				Path:    "/pkg/index.js",
				Source:  `import { add } from "./math.js"; export function run() { return add(20, 22); }`,
				Kind:    "esm",
				Exports: []string{"run"},
				Imports: []ir.DynamicImport{{Specifier: "./math.js", Path: "/pkg/math.js"}},
			},
			{Path: "/pkg/math.js", Source: `export function add(a, b) { return a + b; }`, Kind: "esm", Exports: []string{"add"}},
		},
		Functions: []ir.Function{{
			Name:       "main",
			ReturnType: ir.TypeVoid,
			Body: []ir.Instruction{
				{Op: ir.OpDynamicCall, Type: ir.TypeNumber, Result: "answer", Callee: "/pkg/index.js#run", Field: "run"},
				{Op: ir.OpReturn, Type: ir.TypeVoid},
			},
		}},
	}

	output, err := Emit(module)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"declare i32 @scriptgo_dynamic_register_module",
		"declare i32 @scriptgo_dynamic_register_dependency",
		"call i32 @scriptgo_dynamic_register_module",
		"call i32 @scriptgo_dynamic_register_dependency",
		"call i32 @scriptgo_dynamic_call_module",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("Dynamic LLVM output does not contain %q:\n%s", expected, output)
		}
	}
}
