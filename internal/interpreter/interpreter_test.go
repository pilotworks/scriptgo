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
