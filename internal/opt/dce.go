package opt

import "github.com/pilotworks/scriptgo/internal/ir"

type dcePass struct{}

// NewDCEPass creates a Dead Code Elimination pass.
func NewDCEPass() Pass {
	return &dcePass{}
}

func (p *dcePass) Name() string {
	return "dce"
}

func (p *dcePass) Run(m *ir.Module) (bool, error) {
	changed := false
	for i := range m.Functions {
		fnChanged := p.runOnFunction(&m.Functions[i])
		if fnChanged {
			changed = true
		}
	}
	return changed, nil
}

func (p *dcePass) runOnFunction(fn *ir.Function) bool {
	changed := false
	for {
		used := make(map[string]int)
		p.collectUses(fn.Body, used)

		iterChanged := false
		fn.Body = p.pruneBlock(fn, fn.Body, used, &iterChanged)
		if iterChanged {
			changed = true
		} else {
			break
		}
	}
	return changed
}

func (p *dcePass) pruneBlock(fn *ir.Function, body []ir.Instruction, used map[string]int, changed *bool) []ir.Instruction {
	var result []ir.Instruction

	for _, inst := range body {
		if len(inst.Cond) > 0 {
			inst.Cond = p.pruneBlock(fn, inst.Cond, used, changed)
		}
		if len(inst.Body) > 0 {
			inst.Body = p.pruneBlock(fn, inst.Body, used, changed)
		}
		if len(inst.Step) > 0 {
			inst.Step = p.pruneBlock(fn, inst.Step, used, changed)
		}
		if len(inst.Then) > 0 {
			inst.Then = p.pruneBlock(fn, inst.Then, used, changed)
		}
		if len(inst.Else) > 0 {
			inst.Else = p.pruneBlock(fn, inst.Else, used, changed)
		}
		if len(inst.Catch) > 0 {
			inst.Catch = p.pruneBlock(fn, inst.Catch, used, changed)
		}
		if len(inst.Finally) > 0 {
			inst.Finally = p.pruneBlock(fn, inst.Finally, used, changed)
		}

		if p.isDead(fn, inst, used) {
			*changed = true
			continue
		}

		result = append(result, inst)
	}

	return result
}

func (p *dcePass) isDead(fn *ir.Function, inst ir.Instruction, used map[string]int) bool {
	if inst.Result == "" {
		return false
	}
	if used[inst.Result] > 0 {
		return false
	}

	// Never eliminate named variables, parameters, or captured variables
	for _, l := range fn.Locals {
		if l.Name == inst.Result {
			return false
		}
	}
	for _, p := range fn.Parameters {
		if p.Name == inst.Result {
			return false
		}
	}
	for _, c := range fn.Captured {
		if c.Name == inst.Result {
			return false
		}
	}

	// Instructions with side effects cannot be eliminated
	switch inst.Op {
	case ir.OpConst, ir.OpBinary, ir.OpCompare, ir.OpSelect, ir.OpTypeOf:
		return true
	case ir.OpIndex, ir.OpFieldGet:
		// Pure reads with unused results can be removed
		return true
	default:
		return false
	}
}

func (p *dcePass) collectUses(body []ir.Instruction, used map[string]int) {
	for _, inst := range body {
		for _, arg := range inst.Args {
			used[arg]++
		}
		if inst.Callee != "" {
			used[inst.Callee]++
		}
		if inst.This != "" {
			used[inst.This]++
		}
		if inst.Op == ir.OpReturn && inst.Result != "" {
			used[inst.Result]++
		}
		// OpAssign requires inst.Result to be defined in known
		if inst.Op == ir.OpAssign && inst.Result != "" {
			used[inst.Result]++
		}

		p.collectUses(inst.Cond, used)
		p.collectUses(inst.Body, used)
		p.collectUses(inst.Step, used)
		p.collectUses(inst.Then, used)
		p.collectUses(inst.Else, used)
		p.collectUses(inst.Catch, used)
		p.collectUses(inst.Finally, used)
	}
}
