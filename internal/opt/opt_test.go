package opt

import (
	"testing"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func TestTemporaryObjectRegionPassWrapsNonEscapingBuilderAndVisitor(t *testing.T) {
	builder := ir.Function{
		Name: "buildTree", ReturnType: "object:Node",
		Body: []ir.Instruction{
			{Op: ir.OpObjectNew, Type: "object:Node", Result: "node", Callee: "Node", FieldCount: 1},
			{Op: ir.OpReturn, Type: "object:Node", Args: []string{"node"}},
		},
	}
	visitor := ir.Function{
		Name: "checkTree", ReturnType: ir.TypeNumber,
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "zero", Value: "0"},
			{Op: ir.OpReturn, Type: ir.TypeNumber, Args: []string{"zero"}},
		},
	}
	caller := ir.Function{
		Name: "run", ReturnType: ir.TypeVoid,
		Body: []ir.Instruction{{
			Op: ir.OpWhile,
			Body: []ir.Instruction{
				{Op: ir.OpCall, Type: "object:Node", Result: "tree", Callee: "buildTree"},
				{Op: ir.OpCall, Type: ir.TypeNumber, Result: "sum", Callee: "checkTree", Args: []string{"tree"}},
			},
		}},
	}
	m := ir.Module{Functions: []ir.Function{builder, visitor, caller}}
	changed, err := NewTemporaryObjectRegionPass().Run(&m)
	if err != nil || !changed {
		t.Fatalf("region pass = changed %v, err %v", changed, err)
	}
	body := m.Functions[2].Body[0].Body
	if len(body) != 4 || body[0].Op != ir.OpRegionBegin || body[3].Op != ir.OpRegionEnd {
		t.Fatalf("region boundaries = %#v", body)
	}
}

func TestTemporaryObjectRegionPassRejectsEscapingBuilderResult(t *testing.T) {
	builder := ir.Function{
		Name: "buildTree", ReturnType: "object:Node",
		Body: []ir.Instruction{
			{Op: ir.OpObjectNew, Type: "object:Node", Result: "node", Callee: "Node", FieldCount: 1},
			{Op: ir.OpReturn, Type: "object:Node", Args: []string{"node"}},
		},
	}
	caller := ir.Function{
		Name: "run", ReturnType: ir.TypeVoid,
		Body: []ir.Instruction{{
			Op: ir.OpWhile,
			Body: []ir.Instruction{
				{Op: ir.OpCall, Type: "object:Node", Result: "tree", Callee: "buildTree"},
				{Op: ir.OpPrint, Type: ir.TypeVoid, Args: []string{"tree"}},
			},
		}},
	}
	m := ir.Module{Functions: []ir.Function{builder, caller}}
	changed, err := NewTemporaryObjectRegionPass().Run(&m)
	if err != nil || changed {
		t.Fatalf("escaping result must not form a region: changed %v, err %v", changed, err)
	}
}

func TestTemporaryObjectRegionPassOnlyWrapsLoopBodies(t *testing.T) {
	builder := ir.Function{
		Name: "buildTree", ReturnType: "object:Node",
		Body: []ir.Instruction{
			{Op: ir.OpObjectNew, Type: "object:Node", Result: "node", Callee: "Node", FieldCount: 1},
			{Op: ir.OpReturn, Type: "object:Node", Args: []string{"node"}},
		},
	}
	visitor := ir.Function{Name: "checkTree", ReturnType: ir.TypeNumber, Body: []ir.Instruction{{Op: ir.OpConst, Type: ir.TypeNumber, Result: "zero", Value: "0"}, {Op: ir.OpReturn, Type: ir.TypeNumber, Args: []string{"zero"}}}}
	caller := ir.Function{Name: "run", ReturnType: ir.TypeVoid, Body: []ir.Instruction{{Op: ir.OpCall, Type: "object:Node", Result: "tree", Callee: "buildTree"}, {Op: ir.OpCall, Type: ir.TypeNumber, Result: "sum", Callee: "checkTree", Args: []string{"tree"}}}}
	m := ir.Module{Functions: []ir.Function{builder, visitor, caller}}
	changed, err := NewTemporaryObjectRegionPass().Run(&m)
	if err != nil || changed {
		t.Fatalf("top-level calls must not form a region: changed %v, err %v", changed, err)
	}
}

func TestTemporaryObjectRegionPassRejectsUnsafeConstructor(t *testing.T) {
	unsafeConstructor := ir.Function{
		Name: "Node_constructor", ReturnType: ir.TypeVoid,
		Parameters: []ir.Parameter{{Name: "this", Type: "object:Node"}},
		Body:       []ir.Instruction{{Op: ir.OpPrint, Type: ir.TypeVoid, Args: []string{"this"}}},
	}
	builder := ir.Function{
		Name: "buildTree", ReturnType: "object:Node",
		Body: []ir.Instruction{
			{Op: ir.OpObjectNew, Type: "object:Node", Result: "node", Callee: "Node", FieldCount: 1},
			{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "Node_constructor", Args: []string{"node"}},
			{Op: ir.OpReturn, Type: "object:Node", Args: []string{"node"}},
		},
	}
	visitor := ir.Function{Name: "checkTree", ReturnType: ir.TypeNumber, Body: []ir.Instruction{{Op: ir.OpConst, Type: ir.TypeNumber, Result: "zero", Value: "0"}, {Op: ir.OpReturn, Type: ir.TypeNumber, Args: []string{"zero"}}}}
	caller := ir.Function{Name: "run", ReturnType: ir.TypeVoid, Body: []ir.Instruction{{Op: ir.OpWhile, Body: []ir.Instruction{{Op: ir.OpCall, Type: "object:Node", Result: "tree", Callee: "buildTree"}, {Op: ir.OpCall, Type: ir.TypeNumber, Result: "sum", Callee: "checkTree", Args: []string{"tree"}}}}}}
	m := ir.Module{Functions: []ir.Function{unsafeConstructor, builder, visitor, caller}}
	changed, err := NewTemporaryObjectRegionPass().Run(&m)
	if err != nil || changed {
		t.Fatalf("unsafe constructor must not form a region: changed %v, err %v", changed, err)
	}
}

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
