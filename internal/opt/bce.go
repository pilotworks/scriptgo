package opt

import (
	"strconv"

	"github.com/pilotworks/scriptgo/internal/ir"
)

type bcePass struct{}

// NewBCEPass creates a Bounds Check Elimination and loop vectorization candidate pass.
func NewBCEPass() Pass {
	return &bcePass{}
}

func (p *bcePass) Name() string {
	return "bce"
}

func (p *bcePass) Run(m *ir.Module) (bool, error) {
	changed := false
	for i := range m.Functions {
		if p.runOnFunction(&m.Functions[i]) {
			changed = true
		}
	}
	return changed, nil
}

type loopContext struct {
	indVar        string
	boundVar      string
	inclusive     bool
	lowerZero     bool
	stepOne       bool
	secondaryVars map[string]bool // monotonic secondary counters bounded by indVar
	guardedArrays map[string]bool // arrays with bounds check hoisted to pre-header
}

type bceState struct {
	constants     map[string]float64
	arrayLengths  map[string]string          // lengthVar -> arrayVar
	arraysByBound map[string]map[string]bool // boundVar -> set of arrayVars proven <= boundVar
	priorIndexed  map[string]map[string]bool // arrayVar -> set of indexVars safely accessed
	arrayParents  map[string]string          // childArrayVar -> parentArrayVar (e.g. row -> matrix)
	loopStack     []*loopContext
}

func (p *bcePass) runOnFunction(fn *ir.Function) bool {
	state := &bceState{
		constants:     make(map[string]float64),
		arrayLengths:  make(map[string]string),
		arraysByBound: make(map[string]map[string]bool),
		priorIndexed:  make(map[string]map[string]bool),
		arrayParents:  make(map[string]string),
	}

	for _, param := range fn.Parameters {
		if param.Type == ir.TypeNumber {
			if state.arraysByBound[param.Name] == nil {
				state.arraysByBound[param.Name] = make(map[string]bool)
			}
		}
	}

	changed := false
	fn.Body = p.processBlock(fn.Body, state, &changed, fn)
	return changed
}

func (p *bcePass) processBlock(block []ir.Instruction, state *bceState, changed *bool, fn *ir.Function) []ir.Instruction {
	var result []ir.Instruction

	for i := range block {
		inst := block[i]

		switch inst.Op {
		case ir.OpConst:
			if inst.Type == ir.TypeNumber && inst.Value != "" {
				if val, err := strconv.ParseFloat(inst.Value, 64); err == nil {
					state.constants[inst.Result] = val
				}
			}
			result = append(result, inst)

		case ir.OpFieldGet:
			if inst.Field == "length" && len(inst.Args) == 1 {
				arr := inst.Args[0]
				state.arrayLengths[inst.Result] = arr
				if state.arraysByBound[inst.Result] == nil {
					state.arraysByBound[inst.Result] = make(map[string]bool)
				}
				state.arraysByBound[inst.Result][arr] = true
			}
			result = append(result, inst)

		case ir.OpCall:
			if (inst.Callee == "__array.length" || inst.Callee == "__typedarray.length") && len(inst.Args) == 1 {
				arr := inst.Args[0]
				state.arrayLengths[inst.Result] = arr
				if state.arraysByBound[inst.Result] == nil {
					state.arraysByBound[inst.Result] = make(map[string]bool)
				}
				state.arraysByBound[inst.Result][arr] = true
			}
			result = append(result, inst)

		case ir.OpIndex:
			if len(inst.Args) >= 2 {
				arr := inst.Args[0]
				idx := inst.Args[1]
				if p.canEliminateBoundsCheck(arr, idx, state) {
					if !inst.NoBoundsCheck {
						inst.NoBoundsCheck = true
						*changed = true
					}
				}
				if state.priorIndexed[arr] == nil {
					state.priorIndexed[arr] = make(map[string]bool)
				}
				state.priorIndexed[arr][idx] = true
				if inst.Result != "" {
					state.arrayParents[inst.Result] = arr
				}
			}
			result = append(result, inst)

		case ir.OpIndexSet:
			if len(inst.Args) >= 2 {
				arr := inst.Args[0]
				idx := inst.Args[1]
				if p.canEliminateBoundsCheck(arr, idx, state) {
					if !inst.NoBoundsCheck {
						inst.NoBoundsCheck = true
						*changed = true
					}
				}
				if state.priorIndexed[arr] == nil {
					state.priorIndexed[arr] = make(map[string]bool)
				}
				state.priorIndexed[arr][idx] = true
			}
			result = append(result, inst)

		case ir.OpIf:
			inst.Then = p.processBlock(inst.Then, state, changed, fn)
			inst.Else = p.processBlock(inst.Else, state, changed, fn)
			result = append(result, inst)

		case ir.OpWhile, ir.OpDoWhile:
			loopCtx := p.analyzeLoopHeader(&inst, result, state)
			if loopCtx != nil {
				// Pre-header loop bounds check hoisting
				guards := p.generatePreheaderGuards(&inst, loopCtx, state, result, fn)
				if len(guards) > 0 {
					result = append(result, guards...)
					*changed = true
				}
				state.loopStack = append(state.loopStack, loopCtx)
			}

			inst.Cond = p.processBlock(inst.Cond, state, changed, fn)
			inst.Body = p.processBlock(inst.Body, state, changed, fn)
			if len(inst.Step) > 0 {
				inst.Step = p.processBlock(inst.Step, state, changed, fn)
			}

			if loopCtx != nil {
				state.loopStack = state.loopStack[:len(state.loopStack)-1]
				// Check loop vectorization eligibility
				if p.isVectorizableLoop(&inst, loopCtx) {
					if !inst.Vectorize {
						inst.Vectorize = true
						*changed = true
					}
				}
			}
			result = append(result, inst)

		default:
			result = append(result, inst)
		}
	}
	return result
}

func (p *bcePass) analyzeLoopHeader(loop *ir.Instruction, preceding []ir.Instruction, state *bceState) *loopContext {
	if len(loop.Cond) == 0 {
		return nil
	}

	// Pre-scan loop condition for direct array length queries
	for j := range loop.Cond {
		cInst := &loop.Cond[j]
		if cInst.Op == ir.OpFieldGet && cInst.Field == "length" && len(cInst.Args) == 1 {
			arr := cInst.Args[0]
			state.arrayLengths[cInst.Result] = arr
			if state.arraysByBound[cInst.Result] == nil {
				state.arraysByBound[cInst.Result] = make(map[string]bool)
			}
			state.arraysByBound[cInst.Result][arr] = true
		}
		if cInst.Op == ir.OpCall && (cInst.Callee == "__array.length" || cInst.Callee == "__typedarray.length") && len(cInst.Args) == 1 {
			arr := cInst.Args[0]
			state.arrayLengths[cInst.Result] = arr
			if state.arraysByBound[cInst.Result] == nil {
				state.arraysByBound[cInst.Result] = make(map[string]bool)
			}
			state.arraysByBound[cInst.Result][arr] = true
		}
	}

	var cmpInst *ir.Instruction
	for j := range loop.Cond {
		if loop.Cond[j].Op == ir.OpCompare {
			cmpInst = &loop.Cond[j]
		}
	}
	if cmpInst == nil || len(cmpInst.Args) != 2 {
		return nil
	}

	var indVar, boundVar string
	inclusive := false

	switch cmpInst.Operator {
	case "<":
		indVar = cmpInst.Args[0]
		boundVar = cmpInst.Args[1]
	case "<=":
		indVar = cmpInst.Args[0]
		boundVar = cmpInst.Args[1]
		inclusive = true
	case ">":
		indVar = cmpInst.Args[1]
		boundVar = cmpInst.Args[0]
	case ">=":
		indVar = cmpInst.Args[1]
		boundVar = cmpInst.Args[0]
		inclusive = true
	default:
		return nil
	}

	lowerZero := p.isVarNonNegative(indVar, preceding, state)

	stepOne := false
	checkStepList := loop.Step
	if len(checkStepList) == 0 && len(loop.Body) > 0 {
		checkStepList = loop.Body[len(loop.Body)-1:]
	}
	for _, inst := range checkStepList {
		if inst.Op == ir.OpBinary && inst.Operator == "+" && len(inst.Args) == 2 {
			if inst.Args[0] == indVar || inst.Args[1] == indVar {
				stepOne = true
			}
		}
		if inst.Op == ir.OpAssign && inst.Result == indVar {
			stepOne = true
		}
	}

	ctx := &loopContext{
		indVar:        indVar,
		boundVar:      boundVar,
		inclusive:     inclusive,
		lowerZero:     lowerZero,
		stepOne:       stepOne,
		secondaryVars: make(map[string]bool),
		guardedArrays: make(map[string]bool),
	}

	// Identify secondary monotonic counters bounded by indVar (e.g. partition index in Quicksort)
	p.analyzeSecondaryMonotonicVars(loop, ctx)

	return ctx
}

func (p *bcePass) isVarNonNegative(v string, preceding []ir.Instruction, state *bceState) bool {
	if c, ok := state.constants[v]; ok && c >= 0 {
		return true
	}
	for k := len(preceding) - 1; k >= 0; k-- {
		inst := preceding[k]
		if inst.Result == v {
			if inst.Op == ir.OpConst {
				if c, ok := state.constants[inst.Result]; ok && c >= 0 {
					return true
				}
			} else if inst.Op == ir.OpAssign && len(inst.Args) == 1 {
				return p.isVarNonNegative(inst.Args[0], preceding[:k], state)
			} else if inst.Op == ir.OpBinary && (inst.Operator == "+" || inst.Operator == "-") {
				return true
			}
			break
		}
	}
	return true
}

func (p *bcePass) analyzeSecondaryMonotonicVars(loop *ir.Instruction, ctx *loopContext) {
	incVars := make(map[string]bool)
	otherAssigned := make(map[string]bool)

	p.collectMonotonicUpdates(loop.Body, incVars, otherAssigned)

	for w := range incVars {
		if otherAssigned[w] || w == ctx.indVar {
			continue
		}
		ctx.secondaryVars[w] = true
	}
}

func (p *bcePass) collectMonotonicUpdates(block []ir.Instruction, incVars map[string]bool, otherAssigned map[string]bool) {
	for i := range block {
		inst := block[i]
		if inst.Op == ir.OpAssign && inst.Result != "" {
			if len(inst.Args) == 1 {
				arg := inst.Args[0]
				isInc := false
				for j := i - 1; j >= 0; j-- {
					prev := block[j]
					if prev.Result == arg && prev.Op == ir.OpBinary && prev.Operator == "+" {
						if len(prev.Args) == 2 && (prev.Args[0] == inst.Result || prev.Args[1] == inst.Result) {
							isInc = true
							break
						}
					}
				}
				if isInc {
					incVars[inst.Result] = true
				} else {
					otherAssigned[inst.Result] = true
				}
			}
		}
		if len(inst.Then) > 0 {
			p.collectMonotonicUpdates(inst.Then, incVars, otherAssigned)
		}
		if len(inst.Else) > 0 {
			p.collectMonotonicUpdates(inst.Else, incVars, otherAssigned)
		}
	}
}

func (p *bcePass) generatePreheaderGuards(loop *ir.Instruction, ctx *loopContext, state *bceState, preceding []ir.Instruction, fn *ir.Function) []ir.Instruction {
	if ctx.boundVar == "" {
		return nil
	}

	// Dominance invariant: boundVar must be defined before entering the loop!
	if !p.isVarDefinedBefore(ctx.boundVar, preceding, fn) {
		return nil
	}

	var guards []ir.Instruction
	accessedArrays := make(map[string]bool)

	p.collectLoopIndexedArrays(loop.Body, ctx, accessedArrays)

	for arr := range accessedArrays {
		if p.isBoundValidForArray(arr, ctx.boundVar, state, ctx) {
			continue
		}

		// Dominance invariant: arr must be defined before entering the loop!
		if !p.isVarDefinedBefore(arr, preceding, fn) {
			continue
		}

		// Emit pre-header guard: __array.bounds_guard(arr, boundVar)
		guards = append(guards, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeVoid,
			Callee: "__array.bounds_guard",
			Args:   []string{arr, ctx.boundVar},
		})
		ctx.guardedArrays[arr] = true
		if state.arraysByBound[ctx.boundVar] == nil {
			state.arraysByBound[ctx.boundVar] = make(map[string]bool)
		}
		state.arraysByBound[ctx.boundVar][arr] = true
	}

	return guards
}

func (p *bcePass) isVarDefinedBefore(v string, preceding []ir.Instruction, fn *ir.Function) bool {
	if fn != nil {
		for _, param := range fn.Parameters {
			if param.Name == v {
				return true
			}
		}
	}
	for _, inst := range preceding {
		if inst.Result == v {
			return true
		}
	}
	return false
}

func (p *bcePass) collectLoopIndexedArrays(block []ir.Instruction, ctx *loopContext, arrays map[string]bool) {
	for _, inst := range block {
		if (inst.Op == ir.OpIndex || inst.Op == ir.OpIndexSet) && len(inst.Args) >= 2 {
			arr := inst.Args[0]
			idx := inst.Args[1]
			if idx == ctx.indVar || ctx.secondaryVars[idx] {
				arrays[arr] = true
			}
		}
		if len(inst.Then) > 0 {
			p.collectLoopIndexedArrays(inst.Then, ctx, arrays)
		}
		if len(inst.Else) > 0 {
			p.collectLoopIndexedArrays(inst.Else, ctx, arrays)
		}
	}
}

func (p *bcePass) canEliminateBoundsCheck(arr string, idx string, state *bceState) bool {
	if len(state.loopStack) == 0 {
		return false
	}

	for i := len(state.loopStack) - 1; i >= 0; i-- {
		loop := state.loopStack[i]

		isMatchingIndex := (idx == loop.indVar) || (loop.secondaryVars != nil && loop.secondaryVars[idx])
		if !isMatchingIndex {
			continue
		}

		if p.isBoundValidForArray(arr, loop.boundVar, state, loop) {
			return true
		}

		if parent, hasParent := state.arrayParents[arr]; hasParent {
			if p.isBoundValidForArray(parent, loop.boundVar, state, loop) {
				return true
			}
		}
	}

	return false
}

func (p *bcePass) isBoundValidForArray(arr string, boundVar string, state *bceState, ctx *loopContext) bool {
	if boundVar == "" {
		return false
	}
	// 1. Direct SSA length derivation: boundVar was defined as arr.length
	if targetArr, ok := state.arrayLengths[boundVar]; ok && targetArr == arr {
		return true
	}
	if arrs, ok := state.arraysByBound[boundVar]; ok && arrs[arr] {
		return true
	}
	// 2. Dominating prior access: arr[boundVar] was safely accessed before loop
	if prior, ok := state.priorIndexed[arr]; ok && prior[boundVar] {
		return true
	}
	// 3. Pre-header guarded for this loop context
	if ctx != nil && ctx.guardedArrays != nil && ctx.guardedArrays[arr] {
		return true
	}
	return false
}

func (p *bcePass) isVectorizableLoop(loop *ir.Instruction, ctx *loopContext) bool {
	if len(loop.Cond) == 0 || len(loop.Body) == 0 || !ctx.stepOne {
		return false
	}
	return p.isBranchVectorizable(loop.Body)
}

func (p *bcePass) isBranchVectorizable(branch []ir.Instruction) bool {
	for _, inst := range branch {
		switch inst.Op {
		case ir.OpWhile, ir.OpDoWhile, ir.OpCall, ir.OpClosureCall, ir.OpDynamicCall,
			ir.OpExternCall, ir.OpArray, ir.OpObjectNew, ir.OpTry, ir.OpThrow, ir.OpBreak:
			return false
		case ir.OpIf:
			if !p.isBranchVectorizable(inst.Then) || !p.isBranchVectorizable(inst.Else) {
				return false
			}
		}
	}
	return true
}
