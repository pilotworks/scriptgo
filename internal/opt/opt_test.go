package opt

import (
	"testing"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func TestConstFold_AlgebraicIdentities(t *testing.T) {
	fn := ir.Function{
		Name:       "testFn",
		ReturnType: ir.TypeNumber,
		Parameters: []ir.Parameter{{Name: "x", Type: ir.TypeNumber}},
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "zero", Value: "0"},
			{Op: ir.OpBinary, Type: ir.TypeNumber, Operator: "+", Result: "res1", Args: []string{"x", "zero"}},
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "one", Value: "1"},
			{Op: ir.OpBinary, Type: ir.TypeNumber, Operator: "*", Result: "res2", Args: []string{"res1", "one"}},
			{Op: ir.OpReturn, Type: ir.TypeNumber, Result: "res2", Args: []string{"res2"}},
		},
	}
	m := ir.Module{
		Functions: []ir.Function{fn},
	}

	optMod, err := Optimize(m, Options{Level: "2"})
	if err != nil {
		t.Fatalf("Optimize failed: %v", err)
	}

	body := optMod.Functions[0].Body
	// res2 should be folded/aliased to x, return should return x
	lastInst := body[len(body)-1]
	if lastInst.Op != ir.OpReturn {
		t.Fatalf("expected return instruction, got %s", lastInst.Op)
	}
	if len(lastInst.Args) > 0 && lastInst.Args[0] != "x" {
		t.Errorf("expected return arg 'x', got %q", lastInst.Args[0])
	}
}

func TestCSE_RedundantIndex(t *testing.T) {
	fn := ir.Function{
		Name:       "testCSE",
		ReturnType: ir.TypeNumber,
		Parameters: []ir.Parameter{
			{Name: "arr", Type: ir.TypeNumberArray},
			{Name: "idx", Type: ir.TypeNumber},
		},
		Body: []ir.Instruction{
			{Op: ir.OpIndex, Type: ir.TypeNumber, Result: "val1", Args: []string{"arr", "idx"}},
			{Op: ir.OpIndex, Type: ir.TypeNumber, Result: "val2", Args: []string{"arr", "idx"}},
			{Op: ir.OpBinary, Type: ir.TypeNumber, Operator: "+", Result: "sum", Args: []string{"val1", "val2"}},
			{Op: ir.OpReturn, Type: ir.TypeNumber, Result: "sum", Args: []string{"sum"}},
		},
	}
	m := ir.Module{
		Functions: []ir.Function{fn},
	}

	optMod, err := Optimize(m, Options{Level: "2"})
	if err != nil {
		t.Fatalf("Optimize failed: %v", err)
	}

	body := optMod.Functions[0].Body
	indexCount := 0
	for _, inst := range body {
		if inst.Op == ir.OpIndex {
			indexCount++
		}
	}
	if indexCount != 1 {
		t.Errorf("expected 1 index instruction after CSE, got %d", indexCount)
	}
}

func TestLICM_HoistInvariantIndex(t *testing.T) {
	fn := ir.Function{
		Name:       "testLICM",
		ReturnType: ir.TypeVoid,
		Parameters: []ir.Parameter{
			{Name: "a", Type: "number[][]"},
			{Name: "i", Type: ir.TypeNumber},
			{Name: "size", Type: ir.TypeNumber},
		},
		Locals: []ir.Parameter{
			{Name: "k", Type: ir.TypeNumber},
			{Name: "sum", Type: ir.TypeNumber},
		},
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "k", Value: "0"},
			{
				Op:   ir.OpWhile,
				Type: ir.TypeVoid,
				Args: []string{"cond"},
				Cond: []ir.Instruction{
					{Op: ir.OpCompare, Type: ir.TypeBool, Operator: "<", Result: "cond", Args: []string{"k", "size"}},
				},
				Body: []ir.Instruction{
					{Op: ir.OpIndex, Type: ir.TypeNumberArray, Result: "ai", Args: []string{"a", "i"}},
					{Op: ir.OpIndex, Type: ir.TypeNumber, Result: "elem", Args: []string{"ai", "k"}},
					{Op: ir.OpBinary, Type: ir.TypeNumber, Operator: "+", Result: "sumNext", Args: []string{"sum", "elem"}},
					{Op: ir.OpAssign, Result: "sum", Args: []string{"sumNext"}},
					{Op: ir.OpConst, Type: ir.TypeNumber, Result: "one", Value: "1"},
					{Op: ir.OpBinary, Type: ir.TypeNumber, Operator: "+", Result: "kNext", Args: []string{"k", "one"}},
					{Op: ir.OpAssign, Result: "k", Args: []string{"kNext"}},
				},
			},
			{Op: ir.OpReturn, Type: ir.TypeVoid},
		},
	}
	m := ir.Module{
		Functions: []ir.Function{fn},
	}

	optMod, err := Optimize(m, Options{Level: "2"})
	if err != nil {
		t.Fatalf("Optimize failed: %v", err)
	}

	body := optMod.Functions[0].Body
	// ai should have been hoisted before OpWhile
	foundHoisted := false
	for _, inst := range body {
		if inst.Op == ir.OpIndex && inst.Result == "ai" {
			foundHoisted = true
			break
		}
		if inst.Op == ir.OpWhile {
			break
		}
	}
	if !foundHoisted {
		t.Errorf("expected 'ai = index [a, i]' to be hoisted before while loop")
	}
}
