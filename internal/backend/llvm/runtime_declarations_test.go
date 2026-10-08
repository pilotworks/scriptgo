package llvm

import (
	"regexp"
	"strings"
	"testing"

	"github.com/pilotworks/scriptgo/internal/ir"
)

var declaredSymbol = regexp.MustCompile(`@([A-Za-z_.$][A-Za-z0-9_.$]*)\(`)

func TestRuntimeDeclarationsAreWellFormedAndUnique(t *testing.T) {
	seen := map[string]bool{}
	for index, block := range runtimeDeclarations {
		if !strings.HasSuffix(block, "\n") {
			t.Errorf("block %d does not end with a newline", index)
		}
		for _, line := range strings.Split(block, "\n") {
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "declare ") {
				t.Errorf("block %d contains a non-declaration line %q", index, line)
				continue
			}
			match := declaredSymbol.FindStringSubmatch(line)
			if match == nil {
				t.Errorf("cannot find declared symbol in %q", line)
				continue
			}
			if seen[match[1]] {
				t.Errorf("runtime symbol %s is declared more than once", match[1])
			}
			seen[match[1]] = true
		}
	}
}

func TestEmitWritesRuntimeDeclarationsOnce(t *testing.T) {
	module := ir.Module{Functions: []ir.Function{{
		Name:       "main",
		ReturnType: ir.TypeVoid,
		Body:       []ir.Instruction{{Op: ir.OpReturn, Type: ir.TypeVoid}},
	}}}
	output, err := Emit(module)
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(runtimeDeclarations, "")
	if count := strings.Count(output, all); count != 1 {
		t.Fatalf("runtime declarations appear %d times in the module, want 1", count)
	}
}

// TestEmitCallDispatchesIntrinsicFamilies exercises each grouped dispatcher of
// emitCall with a representative intrinsic.
func TestEmitCallDispatchesIntrinsicFamilies(t *testing.T) {
	cases := []struct {
		name   string
		params []ir.Parameter
		call   ir.Instruction
		want   string
	}{
		{"core", []ir.Parameter{{Name: "x", Type: ir.TypeNumber}},
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: "r", Callee: "__Math.floor", Args: []string{"x"}}, "@llvm.floor.f64"},
		{"value", []ir.Parameter{{Name: "s", Type: ir.TypeString}},
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: "r", Callee: "__string.toUpperCase", Args: []string{"s"}}, "@scriptgo_string_to_upper("},
		{"system", []ir.Parameter{{Name: "xs", Type: ir.TypeNumberArray}},
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: "r", Callee: "__array.pop", Args: []string{"xs"}}, "@scriptgo_array_pop"},
		{"service", nil,
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: "r", Callee: "__date.now"}, "@scriptgo_date_now"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			module := ir.Module{Functions: []ir.Function{{
				Name:       "main",
				ReturnType: ir.TypeVoid,
				Body:       []ir.Instruction{{Op: ir.OpReturn, Type: ir.TypeVoid}},
			}, {
				Name:       "probe",
				ReturnType: ir.TypeVoid,
				Parameters: tc.params,
				Body:       []ir.Instruction{tc.call, {Op: ir.OpReturn, Type: ir.TypeVoid}},
			}}}
			output, err := Emit(module)
			if err != nil {
				t.Fatal(err)
			}
			body := output[strings.Index(output, "@probe("):]
			if !strings.Contains(body, "call") || !strings.Contains(body, tc.want) {
				t.Fatalf("probe body does not call %s:\n%s", tc.want, body)
			}
		})
	}
}

func TestEmitRejectsUnknownGroupedIntrinsic(t *testing.T) {
	module := ir.Module{Functions: []ir.Function{{
		Name:       "main",
		ReturnType: ir.TypeVoid,
		Parameters: []ir.Parameter{{Name: "xs", Type: ir.TypeNumberArray}},
		Body: []ir.Instruction{
			{Op: ir.OpCall, Type: ir.TypeNumber, Result: "r", Callee: "__array.notAMethod", Args: []string{"xs"}},
			{Op: ir.OpReturn, Type: ir.TypeVoid},
		},
	}}}
	if _, err := Emit(module); err == nil {
		t.Fatal("expected an error for an unknown array intrinsic")
	}
}
