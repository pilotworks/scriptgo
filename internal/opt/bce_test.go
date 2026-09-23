package opt

import (
	"testing"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func findLoop(body []ir.Instruction) *ir.Instruction {
	for i := range body {
		if body[i].Op == ir.OpWhile || body[i].Op == ir.OpDoWhile {
			return &body[i]
		}
	}
	return nil
}

func findInstruction(body []ir.Instruction, op string, callee string) *ir.Instruction {
	for i := range body {
		if body[i].Op == op {
			if callee == "" || body[i].Callee == callee {
				return &body[i]
			}
		}
	}
	return nil
}

func TestBCEPassEliminatesBoundsCheckForArrayLengthLoop(t *testing.T) {
	fn := ir.Function{
		Name:       "sumArray",
		ReturnType: ir.TypeNumber,
		Parameters: []ir.Parameter{{Name: "a", Type: ir.TypeNumberArray}},
		Body: []ir.Instruction{
			{Op: ir.OpFieldGet, Type: ir.TypeNumber, Result: "len", Args: []string{"a"}, Field: "length"},
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "i", Value: "0"},
			{
				Op:   ir.OpWhile,
				Type: ir.TypeVoid,
				Args: []string{"cmp"},
				Cond: []ir.Instruction{
					{Op: ir.OpCompare, Type: ir.TypeBool, Result: "cmp", Operator: "<", Args: []string{"i", "len"}},
				},
				Body: []ir.Instruction{
					{Op: ir.OpIndex, Type: ir.TypeNumber, Result: "elem", Args: []string{"a", "i"}},
					{Op: ir.OpBinary, Type: ir.TypeNumber, Result: "add", Operator: "+", Args: []string{"sum", "elem"}},
					{Op: ir.OpAssign, Type: ir.TypeNumber, Result: "sum", Args: []string{"add"}},
				},
				Step: []ir.Instruction{
					{Op: ir.OpConst, Type: ir.TypeNumber, Result: "one", Value: "1"},
					{Op: ir.OpBinary, Type: ir.TypeNumber, Result: "next_i", Operator: "+", Args: []string{"i", "one"}},
					{Op: ir.OpAssign, Type: ir.TypeNumber, Result: "i", Args: []string{"next_i"}},
				},
			},
			{Op: ir.OpReturn, Type: ir.TypeNumber, Args: []string{"sum"}},
		},
	}

	m := ir.Module{Functions: []ir.Function{fn}}
	changed, err := NewBCEPass().Run(&m)
	if err != nil {
		t.Fatalf("bce pass error: %v", err)
	}
	if !changed {
		t.Fatalf("bce pass expected changes but returned changed=false")
	}

	loopInst := findLoop(m.Functions[0].Body)
	if loopInst == nil {
		t.Fatalf("expected loop instruction")
	}
	indexInst := loopInst.Body[0]
	if !indexInst.NoBoundsCheck {
		t.Errorf("expected NoBoundsCheck = true on index instruction, got false")
	}
	if !loopInst.Vectorize {
		t.Errorf("expected Vectorize = true on clean countable loop, got false")
	}
}

func TestBCEPassEliminatesBoundsCheckNumericLoop(t *testing.T) {
	fn := ir.Function{
		Name:       "scaleArray",
		ReturnType: ir.TypeVoid,
		Parameters: []ir.Parameter{
			{Name: "arr", Type: ir.TypeNumberArray},
			{Name: "n", Type: ir.TypeNumber},
		},
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "i", Value: "0"},
			{
				Op:   ir.OpWhile,
				Type: ir.TypeVoid,
				Args: []string{"cmp"},
				Cond: []ir.Instruction{
					{Op: ir.OpCompare, Type: ir.TypeBool, Result: "cmp", Operator: "<", Args: []string{"i", "n"}},
				},
				Body: []ir.Instruction{
					{Op: ir.OpIndex, Type: ir.TypeNumber, Result: "val", Args: []string{"arr", "i"}},
					{Op: ir.OpConst, Type: ir.TypeNumber, Result: "two", Value: "2"},
					{Op: ir.OpBinary, Type: ir.TypeNumber, Result: "doubled", Operator: "*", Args: []string{"val", "two"}},
					{Op: ir.OpIndexSet, Type: ir.TypeVoid, Args: []string{"arr", "i", "doubled"}},
				},
				Step: []ir.Instruction{
					{Op: ir.OpConst, Type: ir.TypeNumber, Result: "one", Value: "1"},
					{Op: ir.OpBinary, Type: ir.TypeNumber, Result: "next_i", Operator: "+", Args: []string{"i", "one"}},
					{Op: ir.OpAssign, Type: ir.TypeNumber, Result: "i", Args: []string{"next_i"}},
				},
			},
			{Op: ir.OpReturn, Type: ir.TypeVoid},
		},
	}

	m := ir.Module{Functions: []ir.Function{fn}}
	changed, err := NewBCEPass().Run(&m)
	if err != nil {
		t.Fatalf("bce pass error: %v", err)
	}
	if !changed {
		t.Fatalf("bce pass expected changes but returned changed=false")
	}

	guardInst := findInstruction(m.Functions[0].Body, ir.OpCall, "__array.bounds_guard")
	if guardInst == nil {
		t.Errorf("expected pre-header guard __array.bounds_guard to be hoisted")
	}

	loopInst := findLoop(m.Functions[0].Body)
	if loopInst == nil {
		t.Fatalf("expected loop instruction")
	}
	indexInst := loopInst.Body[0]
	indexSetInst := loopInst.Body[3]

	if !indexInst.NoBoundsCheck {
		t.Errorf("expected NoBoundsCheck = true on OpIndex, got false")
	}
	if !indexSetInst.NoBoundsCheck {
		t.Errorf("expected NoBoundsCheck = true on OpIndexSet, got false")
	}
	if !loopInst.Vectorize {
		t.Errorf("expected Vectorize = true on numeric loop, got false")
	}
}

func TestBCEPassVectorizeDisabledOnEscapingCall(t *testing.T) {
	fn := ir.Function{
		Name:       "loopWithCall",
		ReturnType: ir.TypeVoid,
		Parameters: []ir.Parameter{
			{Name: "arr", Type: ir.TypeNumberArray},
			{Name: "n", Type: ir.TypeNumber},
		},
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "i", Value: "0"},
			{
				Op:   ir.OpWhile,
				Type: ir.TypeVoid,
				Args: []string{"cmp"},
				Cond: []ir.Instruction{
					{Op: ir.OpCompare, Type: ir.TypeBool, Result: "cmp", Operator: "<", Args: []string{"i", "n"}},
				},
				Body: []ir.Instruction{
					{Op: ir.OpIndex, Type: ir.TypeNumber, Result: "val", Args: []string{"arr", "i"}},
					{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "printValue", Args: []string{"val"}},
				},
				Step: []ir.Instruction{
					{Op: ir.OpConst, Type: ir.TypeNumber, Result: "one", Value: "1"},
					{Op: ir.OpBinary, Type: ir.TypeNumber, Result: "next_i", Operator: "+", Args: []string{"i", "one"}},
					{Op: ir.OpAssign, Type: ir.TypeNumber, Result: "i", Args: []string{"next_i"}},
				},
			},
			{Op: ir.OpReturn, Type: ir.TypeVoid},
		},
	}

	m := ir.Module{Functions: []ir.Function{fn}}
	_, err := NewBCEPass().Run(&m)
	if err != nil {
		t.Fatalf("bce pass error: %v", err)
	}

	loopInst := findLoop(m.Functions[0].Body)
	if loopInst == nil {
		t.Fatalf("expected loop instruction")
	}
	if loopInst.Vectorize {
		t.Errorf("expected Vectorize = false because of function call inside loop body, got true")
	}
}

func TestBCEPassWithArbitraryVariableNames(t *testing.T) {
	// Demonstrates that BCE works with arbitrary, non-conventional variable names
	fn := ir.Function{
		Name:       "customTransform",
		ReturnType: ir.TypeVoid,
		Parameters: []ir.Parameter{
			{Name: "dataBuffer", Type: ir.TypeNumberArray},
			{Name: "customLimit", Type: ir.TypeNumber},
		},
		Body: []ir.Instruction{
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "customIndex", Value: "0"},
			{
				Op:   ir.OpWhile,
				Type: ir.TypeVoid,
				Args: []string{"customCmp"},
				Cond: []ir.Instruction{
					{Op: ir.OpCompare, Type: ir.TypeBool, Result: "customCmp", Operator: "<", Args: []string{"customIndex", "customLimit"}},
				},
				Body: []ir.Instruction{
					{Op: ir.OpIndex, Type: ir.TypeNumber, Result: "itemVal", Args: []string{"dataBuffer", "customIndex"}},
					{Op: ir.OpConst, Type: ir.TypeNumber, Result: "tenVal", Value: "10"},
					{Op: ir.OpBinary, Type: ir.TypeNumber, Result: "multVal", Operator: "*", Args: []string{"itemVal", "tenVal"}},
					{Op: ir.OpIndexSet, Type: ir.TypeVoid, Args: []string{"dataBuffer", "customIndex", "multVal"}},
				},
				Step: []ir.Instruction{
					{Op: ir.OpConst, Type: ir.TypeNumber, Result: "stepDelta", Value: "1"},
					{Op: ir.OpBinary, Type: ir.TypeNumber, Result: "nextIndex", Operator: "+", Args: []string{"customIndex", "stepDelta"}},
					{Op: ir.OpAssign, Type: ir.TypeNumber, Result: "customIndex", Args: []string{"nextIndex"}},
				},
			},
			{Op: ir.OpReturn, Type: ir.TypeVoid},
		},
	}

	m := ir.Module{Functions: []ir.Function{fn}}
	changed, err := NewBCEPass().Run(&m)
	if err != nil {
		t.Fatalf("bce pass error: %v", err)
	}
	if !changed {
		t.Fatalf("expected bce to optimize arbitrary named loop")
	}

	loopInst := findLoop(m.Functions[0].Body)
	if loopInst == nil {
		t.Fatalf("expected loop instruction")
	}
	if !loopInst.Body[0].NoBoundsCheck {
		t.Errorf("expected NoBoundsCheck = true on dataBuffer[customIndex]")
	}
	if !loopInst.Body[3].NoBoundsCheck {
		t.Errorf("expected NoBoundsCheck = true on dataBuffer[customIndex] = multVal")
	}
	if !loopInst.Vectorize {
		t.Errorf("expected Vectorize = true on custom named countable loop")
	}
}

func TestBCEPassSecondaryMonotonicCounter(t *testing.T) {
	// Demonstrates partition pattern in Quicksort:
	// dominating access arr[highBound] proves highBound < len(arr)
	// wall counter incremented at most once per cursor iteration
	fn := ir.Function{
		Name:       "customPartition",
		ReturnType: ir.TypeNumber,
		Parameters: []ir.Parameter{
			{Name: "dataArr", Type: ir.TypeNumberArray},
			{Name: "lowBound", Type: ir.TypeNumber},
			{Name: "highBound", Type: ir.TypeNumber},
		},
		Body: []ir.Instruction{
			// Dominating prior access
			{Op: ir.OpIndex, Type: ir.TypeNumber, Result: "pivotVal", Args: []string{"dataArr", "highBound"}},
			// Secondary counter initialized before loop
			{Op: ir.OpConst, Type: ir.TypeNumber, Result: "oneConst", Value: "1"},
			{Op: ir.OpBinary, Type: ir.TypeNumber, Result: "wallInit", Operator: "-", Args: []string{"lowBound", "oneConst"}},
			{Op: ir.OpAssign, Type: ir.TypeNumber, Result: "wall", Args: []string{"wallInit"}},
			// Loop initialization
			{Op: ir.OpAssign, Type: ir.TypeNumber, Result: "cursor", Args: []string{"lowBound"}},
			{
				Op:   ir.OpWhile,
				Type: ir.TypeVoid,
				Args: []string{"loopCmp"},
				Cond: []ir.Instruction{
					{Op: ir.OpCompare, Type: ir.TypeBool, Result: "loopCmp", Operator: "<", Args: []string{"cursor", "highBound"}},
				},
				Body: []ir.Instruction{
					// Primary induction access
					{Op: ir.OpIndex, Type: ir.TypeNumber, Result: "currVal", Args: []string{"dataArr", "cursor"}},
					{Op: ir.OpCompare, Type: ir.TypeBool, Result: "branchCmp", Operator: "<=", Args: []string{"currVal", "pivotVal"}},
					{
						Op:   ir.OpIf,
						Type: ir.TypeVoid,
						Args: []string{"branchCmp"},
						Then: []ir.Instruction{
							{Op: ir.OpBinary, Type: ir.TypeNumber, Result: "nextWall", Operator: "+", Args: []string{"wall", "oneConst"}},
							{Op: ir.OpAssign, Type: ir.TypeNumber, Result: "wall", Args: []string{"nextWall"}},
							// Secondary counter access
							{Op: ir.OpIndex, Type: ir.TypeNumber, Result: "wallVal", Args: []string{"dataArr", "wall"}},
							{Op: ir.OpIndexSet, Type: ir.TypeVoid, Args: []string{"dataArr", "wall", "currVal"}},
							{Op: ir.OpIndexSet, Type: ir.TypeVoid, Args: []string{"dataArr", "cursor", "wallVal"}},
						},
					},
				},
				Step: []ir.Instruction{
					{Op: ir.OpBinary, Type: ir.TypeNumber, Result: "nextCursor", Operator: "+", Args: []string{"cursor", "oneConst"}},
					{Op: ir.OpAssign, Type: ir.TypeNumber, Result: "cursor", Args: []string{"nextCursor"}},
				},
			},
			{Op: ir.OpReturn, Type: ir.TypeNumber, Args: []string{"wall"}},
		},
	}

	m := ir.Module{Functions: []ir.Function{fn}}
	changed, err := NewBCEPass().Run(&m)
	if err != nil {
		t.Fatalf("bce pass error: %v", err)
	}
	if !changed {
		t.Fatalf("expected bce to optimize secondary monotonic counter loop")
	}

	loopInst := findLoop(m.Functions[0].Body)
	if loopInst == nil {
		t.Fatalf("expected loop instruction")
	}
	// Check primary induction access arr[cursor]
	if !loopInst.Body[0].NoBoundsCheck {
		t.Errorf("expected NoBoundsCheck = true on dataArr[cursor]")
	}
	// Check secondary counter access arr[wall] inside If
	ifBranch := loopInst.Body[2]
	if !ifBranch.Then[2].NoBoundsCheck {
		t.Errorf("expected NoBoundsCheck = true on dataArr[wall] (OpIndex)")
	}
	if !ifBranch.Then[3].NoBoundsCheck {
		t.Errorf("expected NoBoundsCheck = true on dataArr[wall] (OpIndexSet)")
	}
	if !ifBranch.Then[4].NoBoundsCheck {
		t.Errorf("expected NoBoundsCheck = true on dataArr[cursor] (OpIndexSet)")
	}
}
