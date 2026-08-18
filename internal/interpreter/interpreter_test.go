package interpreter

import (
	"strings"
	"testing"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func TestExecutePrintsArithmeticResult(t *testing.T) {
	module := ir.Module{Functions: []ir.Function{{
		Name:       "main",
		ReturnType: ir.TypeVoid,
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "left", Value: "20"},
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "right", Value: "22"},
			{Op: ir.OpBinary, Type: ir.TypeNumber, Result: "sum", Operator: "+", Args: []string{"left", "right"}},
			{Op: ir.OpPrint, Type: ir.TypeVoid, Args: []string{"sum"}},
			{Op: ir.OpReturn, Type: ir.TypeVoid},
		},
	}}}

	result, err := Execute(module)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(result.Output, "42\n") {
		t.Fatalf("interpreter output = %q, want %q", result.Output, "42\n")
	}
}

func TestExecuteIndexesNumberArray(t *testing.T) {
	module := ir.Module{Functions: []ir.Function{{
		Name:       "main",
		ReturnType: ir.TypeVoid,
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "first", Value: "10"},
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "second", Value: "20"},
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "index", Value: "1"},
			{Op: ir.OpArray, Type: ir.TypeNumberArray, Result: "values", Args: []string{"first", "second"}},
			{Op: ir.OpIndex, Type: ir.TypeNumber, Result: "value", Args: []string{"values", "index"}},
			{Op: ir.OpPrint, Type: ir.TypeVoid, Args: []string{"value"}},
			{Op: ir.OpReturn, Type: ir.TypeVoid},
		},
	}}}

	result, err := Execute(module)
	if err != nil {
		t.Fatal(err)
	}
	if result.Output != "20\n" {
		t.Fatalf("array output = %q, want %q", result.Output, "20\n")
	}
}

func TestExecuteRejectsArrayIndexOutOfBounds(t *testing.T) {
	module := ir.Module{Functions: []ir.Function{{
		Name:       "main",
		ReturnType: ir.TypeVoid,
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "value", Value: "10"},
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "index", Value: "1"},
			{Op: ir.OpArray, Type: ir.TypeNumberArray, Result: "values", Args: []string{"value"}},
			{Op: ir.OpIndex, Type: ir.TypeNumber, Result: "result", Args: []string{"values", "index"}},
			{Op: ir.OpReturn, Type: ir.TypeVoid},
		},
	}}}

	if _, err := Execute(module); err == nil || !strings.Contains(err.Error(), "out of bounds") {
		t.Fatalf("Execute error = %v, want bounds diagnostic", err)
	}
}
