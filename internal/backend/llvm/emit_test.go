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
	for _, expected := range []string{"define i32 @main(", "fadd double", "@scriptgo_console_log_number", "ret i32 0"} {
		if !strings.Contains(output, expected) {
			t.Errorf("LLVM output does not contain %q:\n%s", expected, output)
		}
	}
}

func TestEmitNumberArrayAndIndex(t *testing.T) {
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

	output, err := Emit(module)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"declare i32 @scriptgo_array_new(i64, i64, ptr)",
		"call i32 @scriptgo_array_set",
		"call i32 @scriptgo_array_get",
		"call i32 @scriptgo_array_release",
		"load double",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("LLVM output does not contain %q:\n%s", expected, output)
		}
	}
}
