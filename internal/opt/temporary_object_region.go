package opt

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

// temporaryObjectRegionPass brackets a tightly constrained builder/visitor
// sequence in a loop. The runtime can then reclaim the complete object graph
// at once, without registering every temporary node with the tracing GC.
type temporaryObjectRegionPass struct{}

func NewTemporaryObjectRegionPass() Pass          { return &temporaryObjectRegionPass{} }
func (p *temporaryObjectRegionPass) Name() string { return "temporary-object-regions" }

func (p *temporaryObjectRegionPass) Run(m *ir.Module) (bool, error) {
	builders := make(map[string]bool)
	inspectors := make(map[string]bool)
	constructors := make(map[string]bool)
	for _, fn := range m.Functions {
		if isRegionConstructor(fn) {
			constructors[fn.Name] = true
		}
	}
	for _, fn := range m.Functions {
		if isRegionBuilder(fn, constructors) {
			builders[fn.Name] = true
		}
		if isRegionInspector(fn) {
			inspectors[fn.Name] = true
		}
	}
	changed := false
	for i := range m.Functions {
		m.Functions[i].Body = p.processBlock(m.Functions[i].Body, builders, inspectors, false, &changed)
	}
	return changed, nil
}

func (p *temporaryObjectRegionPass) processBlock(body []ir.Instruction, builders, inspectors map[string]bool, regionEligible bool, changed *bool) []ir.Instruction {
	for i := range body {
		inst := &body[i]
		inst.Cond = p.processBlock(inst.Cond, builders, inspectors, false, changed)
		inst.Body = p.processBlock(inst.Body, builders, inspectors, inst.Op == ir.OpWhile || inst.Op == ir.OpDoWhile, changed)
		inst.Step = p.processBlock(inst.Step, builders, inspectors, false, changed)
		inst.Then = p.processBlock(inst.Then, builders, inspectors, false, changed)
		inst.Else = p.processBlock(inst.Else, builders, inspectors, false, changed)
		inst.Catch = p.processBlock(inst.Catch, builders, inspectors, false, changed)
		inst.Finally = p.processBlock(inst.Finally, builders, inspectors, false, changed)
	}

	if !regionEligible {
		return body
	}
	start, end, ok := regionInterval(body, builders, inspectors)
	if !ok {
		return body
	}
	result := make([]ir.Instruction, 0, len(body)+2)
	result = append(result, body[:start]...)
	result = append(result, ir.Instruction{Op: ir.OpRegionBegin})
	result = append(result, body[start:end+1]...)
	result = append(result, ir.Instruction{Op: ir.OpRegionEnd})
	result = append(result, body[end+1:]...)
	*changed = true
	return result
}

func regionInterval(body []ir.Instruction, builders, inspectors map[string]bool) (int, int, bool) {
	start, end := -1, -1
	for i, inst := range body {
		if inst.Op == ir.OpRegionBegin || inst.Op == ir.OpRegionEnd {
			return 0, 0, false
		}
		if inst.Op != ir.OpCall || !builders[inst.Callee] || inst.Result == "" {
			continue
		}
		consumer := -1
		for j := i + 1; j < len(body); j++ {
			if contains(body[j].Args, inst.Result) {
				if body[j].Op != ir.OpCall || !inspectors[body[j].Callee] {
					return 0, 0, false
				}
				consumer = j
			}
		}
		if consumer < 0 {
			continue
		}
		if start < 0 {
			start = i
		}
		if consumer > end {
			end = consumer
		}
	}
	if start < 0 {
		return 0, 0, false
	}
	for _, inst := range body[start : end+1] {
		switch inst.Op {
		case ir.OpConst, ir.OpBinary, ir.OpCompare, ir.OpAssign:
		case ir.OpCall:
			if !builders[inst.Callee] && !inspectors[inst.Callee] {
				return 0, 0, false
			}
		default:
			return 0, 0, false
		}
	}
	return start, end, true
}

func isRegionBuilder(fn ir.Function, constructors map[string]bool) bool {
	if !strings.HasPrefix(string(fn.ReturnType), "object:") {
		return false
	}
	hasObjectNew, valid := false, true
	var scan func([]ir.Instruction)
	scan = func(body []ir.Instruction) {
		for _, inst := range body {
			switch inst.Op {
			case ir.OpConst, ir.OpBinary, ir.OpCompare, ir.OpIf, ir.OpReturn, ir.OpObjectNew, ir.OpFieldSet:
			case ir.OpCall:
				if inst.Callee != fn.Name && !constructors[inst.Callee] {
					valid = false
				}
			default:
				valid = false
			}
			if inst.Op == ir.OpObjectNew {
				hasObjectNew = true
			}
			scan(inst.Cond)
			scan(inst.Body)
			scan(inst.Step)
			scan(inst.Then)
			scan(inst.Else)
		}
	}
	scan(fn.Body)
	return valid && hasObjectNew
}

func isRegionConstructor(fn ir.Function) bool {
	if !strings.HasSuffix(fn.Name, "_constructor") || fn.ReturnType != ir.TypeVoid || len(fn.Parameters) == 0 || fn.Parameters[0].Name != "this" {
		return false
	}
	valid := true
	var scan func([]ir.Instruction)
	scan = func(body []ir.Instruction) {
		for _, inst := range body {
			switch inst.Op {
			case ir.OpConst, ir.OpBinary, ir.OpCompare, ir.OpIf, ir.OpReturn:
			case ir.OpFieldSet:
				if len(inst.Args) != 2 || inst.Args[0] != "this" {
					valid = false
				}
			default:
				valid = false
			}
			scan(inst.Cond)
			scan(inst.Body)
			scan(inst.Step)
			scan(inst.Then)
			scan(inst.Else)
		}
	}
	scan(fn.Body)
	return valid
}

func isRegionInspector(fn ir.Function) bool {
	if fn.ReturnType == "" || strings.HasPrefix(string(fn.ReturnType), "object:") {
		return false
	}
	valid := true
	var scan func([]ir.Instruction)
	scan = func(body []ir.Instruction) {
		for _, inst := range body {
			switch inst.Op {
			case ir.OpConst, ir.OpBinary, ir.OpCompare, ir.OpIf, ir.OpReturn, ir.OpFieldGet, ir.OpAssign:
			case ir.OpCall:
				if inst.Callee != fn.Name {
					valid = false
				}
			default:
				valid = false
			}
			scan(inst.Cond)
			scan(inst.Body)
			scan(inst.Step)
			scan(inst.Then)
			scan(inst.Else)
		}
	}
	scan(fn.Body)
	return valid
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
