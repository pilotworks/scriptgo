package opt

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

type licmPass struct{}

// NewLICMPass creates a Loop-Invariant Code Motion pass.
func NewLICMPass() Pass {
	return &licmPass{}
}

func (p *licmPass) Name() string {
	return "licm"
}

func (p *licmPass) Run(m *ir.Module) (bool, error) {
	changed := false
	for i := range m.Functions {
		fnChanged := p.runOnFunction(&m.Functions[i])
		if fnChanged {
			changed = true
		}
	}
	return changed, nil
}

func (p *licmPass) runOnFunction(fn *ir.Function) bool {
	changed := false
	fn.Body = p.processBlock(fn.Body, &changed)
	return changed
}

func (p *licmPass) processBlock(body []ir.Instruction, changed *bool) []ir.Instruction {
	var result []ir.Instruction

	for i := range body {
		inst := body[i]

		// Recursively optimize inside nested control flow
		if inst.Op == ir.OpIf {
			inst.Then = p.processBlock(inst.Then, changed)
			inst.Else = p.processBlock(inst.Else, changed)
			result = append(result, inst)
			continue
		}

		if inst.Op == ir.OpWhile || inst.Op == ir.OpDoWhile {
			// First, process inner blocks bottom-up
			inst.Cond = p.processBlock(inst.Cond, changed)
			inst.Body = p.processBlock(inst.Body, changed)
			if len(inst.Step) > 0 {
				inst.Step = p.processBlock(inst.Step, changed)
			}

			// Now analyze this loop for invariant instructions to hoist
			hoisted := p.hoistFromLoop(&inst, changed)
			if len(hoisted) > 0 {
				result = append(result, hoisted...)
			}
			result = append(result, inst)
			continue
		}

		result = append(result, inst)
	}

	return result
}

func (p *licmPass) hoistFromLoop(loop *ir.Instruction, changed *bool) []ir.Instruction {
	var hoisted []ir.Instruction

	// Collect loop-assigned variables and mutated arrays
	assigned := make(map[string]bool)
	mutatedArrays := make(map[string]bool)
	p.collectLoopMutations(loop, assigned, mutatedArrays)

	// Collect all SSA definitions created inside the loop
	localDefs := make(map[string]bool)
	p.collectLocalDefs(loop.Cond, localDefs)
	p.collectLocalDefs(loop.Body, localDefs)
	p.collectLocalDefs(loop.Step, localDefs)

	// Repeatedly scan loop Body for hoistable instructions
	for {
		found := false
		var newBody []ir.Instruction

		for _, inst := range loop.Body {
			if p.canHoist(inst, assigned, mutatedArrays, localDefs) {
				hoisted = append(hoisted, inst)
				delete(localDefs, inst.Result) // now defined outside the loop
				found = true
				*changed = true
			} else {
				newBody = append(newBody, inst)
			}
		}

		loop.Body = newBody
		if !found {
			break
		}
	}

	return hoisted
}

func (p *licmPass) canHoist(
	inst ir.Instruction,
	assigned map[string]bool,
	mutatedArrays map[string]bool,
	localDefs map[string]bool,
) bool {
	// Cannot hoist instructions without a result or control flow
	if inst.Result == "" {
		return false
	}
	if assigned[inst.Result] {
		return false
	}

	switch inst.Op {
	case ir.OpConst:
		return true

	case ir.OpBinary, ir.OpCompare:
		if len(inst.Args) != 2 {
			return false
		}
		for _, arg := range inst.Args {
			if assigned[arg] || localDefs[arg] {
				return false
			}
		}
		return true

	case ir.OpSelect:
		if len(inst.Args) != 3 {
			return false
		}
		for _, arg := range inst.Args {
			if assigned[arg] || localDefs[arg] {
				return false
			}
		}
		return true

	case ir.OpIndex:
		if len(inst.Args) != 2 {
			return false
		}
		arr, idx := inst.Args[0], inst.Args[1]
		if assigned[arr] || localDefs[arr] || mutatedArrays[arr] {
			return false
		}
		if assigned[idx] || localDefs[idx] {
			return false
		}
		return true

	case ir.OpFieldGet:
		if len(inst.Args) != 1 {
			return false
		}
		obj := inst.Args[0]
		if assigned[obj] || localDefs[obj] {
			return false
		}
		return true
	}

	return false
}

func (p *licmPass) collectLoopMutations(
	loop *ir.Instruction,
	assigned map[string]bool,
	mutatedArrays map[string]bool,
) {
	var scan func(list []ir.Instruction)
	scan = func(list []ir.Instruction) {
		for _, inst := range list {
			if inst.Op == ir.OpAssign && inst.Result != "" {
				assigned[inst.Result] = true
			}
			if inst.Op == ir.OpIndexSet && len(inst.Args) > 0 {
				mutatedArrays[inst.Args[0]] = true
			}
			if inst.Op == ir.OpCall {
				// Builtin mutating array methods
				if strings.HasPrefix(inst.Callee, "__array.push") ||
					strings.HasPrefix(inst.Callee, "__array.pop") ||
					strings.HasPrefix(inst.Callee, "__array.shift") ||
					strings.HasPrefix(inst.Callee, "__array.unshift") ||
					strings.HasPrefix(inst.Callee, "__array.splice") {
					if len(inst.Args) > 0 {
						mutatedArrays[inst.Args[0]] = true
					}
				} else {
					// Conservative: any array passed to a call could be mutated
					for _, arg := range inst.Args {
						mutatedArrays[arg] = true
					}
				}
			}
			if inst.Op == ir.OpClosureCall || inst.Op == ir.OpDynamicCall {
				for _, arg := range inst.Args {
					mutatedArrays[arg] = true
				}
			}

			scan(inst.Cond)
			scan(inst.Body)
			scan(inst.Step)
			scan(inst.Then)
			scan(inst.Else)
			scan(inst.Catch)
			scan(inst.Finally)
		}
	}

	scan(loop.Cond)
	scan(loop.Body)
	scan(loop.Step)
}

func (p *licmPass) collectLocalDefs(list []ir.Instruction, defs map[string]bool) {
	for _, inst := range list {
		if inst.Result != "" {
			defs[inst.Result] = true
		}
		p.collectLocalDefs(inst.Cond, defs)
		p.collectLocalDefs(inst.Body, defs)
		p.collectLocalDefs(inst.Step, defs)
		p.collectLocalDefs(inst.Then, defs)
		p.collectLocalDefs(inst.Else, defs)
		p.collectLocalDefs(inst.Catch, defs)
		p.collectLocalDefs(inst.Finally, defs)
	}
}
