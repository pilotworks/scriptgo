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
	loopConstructors := make(map[string]bool)

	// Builtin JSON builders and inspectors
	builders["__json.parse_unknown"] = true
	builders["__json.parse"] = true
	inspectors["__json.stringify_unknown"] = true
	inspectors["__json.stringify_object_array"] = true
	inspectors["__json.stringify_number_array"] = true
	inspectors["__json.stringify_string_array"] = true
	inspectors["__json.stringify_bool_array"] = true
	inspectors["__json.inspect_object"] = true

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
	for _, fn := range m.Functions {
		if isLoopRegionConstructor(fn, inspectors) {
			loopConstructors[fn.Name] = true
		}
	}
	changed := false
	for i := range m.Functions {
		m.Functions[i].Body = p.processBlock(m.Functions[i].Body, builders, inspectors, loopConstructors, false, &changed)
	}
	return changed, nil
}

func (p *temporaryObjectRegionPass) processBlock(body []ir.Instruction, builders, inspectors, loopConstructors map[string]bool, regionEligible bool, changed *bool) []ir.Instruction {
	for i := range body {
		inst := &body[i]
		inst.Cond = p.processBlock(inst.Cond, builders, inspectors, loopConstructors, false, changed)
		inst.Body = p.processBlock(inst.Body, builders, inspectors, loopConstructors, inst.Op == ir.OpWhile || inst.Op == ir.OpDoWhile, changed)
		inst.Step = p.processBlock(inst.Step, builders, inspectors, loopConstructors, false, changed)
		inst.Then = p.processBlock(inst.Then, builders, inspectors, loopConstructors, false, changed)
		inst.Else = p.processBlock(inst.Else, builders, inspectors, loopConstructors, false, changed)
		inst.Catch = p.processBlock(inst.Catch, builders, inspectors, loopConstructors, false, changed)
		inst.Finally = p.processBlock(inst.Finally, builders, inspectors, loopConstructors, false, changed)
	}

	if !regionEligible {
		return body
	}
	start, end, ok := regionInterval(body, builders, inspectors)
	if !ok {
		start, end, ok = loopTemporaryObjectInterval(body, loopConstructors)
	}
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

// loopTemporaryObjectInterval recognizes a complete loop body containing only
// temporary class instances, their constructors, field accesses, and calls to
// closures stored on those instances. No value may escape this interval.
func loopTemporaryObjectInterval(body []ir.Instruction, constructors map[string]bool) (int, int, bool) {
	if len(body) == 0 {
		return 0, 0, false
	}
	objects := make(map[string]bool)
	closures := make(map[string]bool)
	for _, inst := range body {
		switch inst.Op {
		case ir.OpConst:
		case ir.OpBinary, ir.OpCompare, ir.OpAssign:
			if containsTemporary(inst.Args, objects, closures) {
				return 0, 0, false
			}
		case ir.OpObjectNew:
			if inst.Result == "" {
				return 0, 0, false
			}
			objects[inst.Result] = true
		case ir.OpCall:
			if !constructors[inst.Callee] || len(inst.Args) == 0 || !objects[inst.Args[0]] {
				return 0, 0, false
			}
		case ir.OpFieldSet:
			if len(inst.Args) != 2 || !objects[inst.Args[0]] {
				return 0, 0, false
			}
		case ir.OpFieldGet:
			if len(inst.Args) != 1 || !objects[inst.Args[0]] || inst.Result == "" {
				return 0, 0, false
			}
			if inst.Type == ir.TypeClosure {
				closures[inst.Result] = true
			}
		case ir.OpClosureCall:
			if !closures[inst.Callee] {
				return 0, 0, false
			}
		default:
			return 0, 0, false
		}
	}
	if len(objects) == 0 {
		return 0, 0, false
	}
	return 0, len(body) - 1, true
}

func containsTemporary(values []string, objects, closures map[string]bool) bool {
	for _, value := range values {
		if objects[value] || closures[value] {
			return true
		}
	}
	return false
}

func containsAny(values []string, targets map[string]bool) bool {
	for _, v := range values {
		if targets[v] {
			return true
		}
	}
	return false
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
		aliases := map[string]bool{inst.Result: true}
		consumer := -1
		valid := true
		for j := i + 1; j < len(body); j++ {
			if containsAny(body[j].Args, aliases) {
				if body[j].Op == ir.OpAssign && body[j].Result != "" {
					aliases[body[j].Result] = true
				} else if body[j].Op == ir.OpCall && inspectors[body[j].Callee] {
					consumer = j
				} else {
					valid = false
					break
				}
			}
		}
		if !valid || consumer < 0 {
			continue
		}
		escapesAfter := false
		for j := consumer + 1; j < len(body); j++ {
			if containsAny(body[j].Args, aliases) {
				escapesAfter = true
				break
			}
		}
		if escapesAfter {
			continue
		}
		if start < 0 || i < start {
			start = i
		}
		if consumer > end {
			end = consumer
		}
	}
	if start < 0 || end < start {
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

// isLoopRegionConstructor permits a scalar inspector closure stored on the
// freshly allocated receiver. The caller proves the receiver and closure die
// at the end of the loop region.
func isLoopRegionConstructor(fn ir.Function, inspectors map[string]bool) bool {
	if !strings.HasSuffix(fn.Name, "_constructor") || fn.ReturnType != ir.TypeVoid || len(fn.Parameters) == 0 || fn.Parameters[0].Name != "this" {
		return false
	}
	valid := true
	var scan func([]ir.Instruction)
	scan = func(body []ir.Instruction) {
		for _, inst := range body {
			switch inst.Op {
			case ir.OpConst, ir.OpBinary, ir.OpCompare, ir.OpIf, ir.OpReturn:
			case ir.OpClosure:
				if !inspectors[inst.Callee] || inst.Result == "" {
					valid = false
				}
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
