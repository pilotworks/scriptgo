package llvm

import (
	"strings"
	"testing"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func TestEmitArithmeticAndPrint(t *testing.T) {
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

	output, err := Emit(module)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"define i32 @main()", "fadd double", "@printf", "ret i32 0"} {
		if !strings.Contains(output, expected) {
			t.Errorf("LLVM output does not contain %q:\n%s", expected, output)
		}
	}
}
