package opt

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

type csePass struct{}

// NewCSEPass creates a scoped common subexpression elimination pass.
func NewCSEPass() Pass {
	return &csePass{}
}

func (p *csePass) Name() string {
	return "cse"
}

func (p *csePass) Run(m *ir.Module) (bool, error) {
	changed := false
	for i := range m.Functions {
		fnChanged := p.runOnFunction(&m.Functions[i])
		if fnChanged {
			changed = true
		}
	}
	return changed, nil
}

func (p *csePass) runOnFunction(fn *ir.Function) bool {
	assigned := make(map[string]bool)
	collectAssigned(fn.Body, assigned)

	changed := false
	fn.Body = p.processBlock(fn, fn.Body, assigned, make(map[string]string), make(map[string]string), &changed)
	return changed
}

func (p *csePass) processBlock(
	fn *ir.Function,
	body []ir.Instruction,
	assigned map[string]bool,
	available map[string]string,
	aliases map[string]string,
	changed *bool,
) []ir.Instruction {
	localAvail := cloneMap(available)
	localAliases := cloneMap(aliases)
	var result []ir.Instruction

	for _, inst := range body {
		// Apply dominating aliases to current instruction operands inline
		for i, arg := range inst.Args {
			if target, ok := localAliases[arg]; ok {
				inst.Args[i] = target
				*changed = true
			}
		}
		if target, ok := localAliases[inst.Callee]; ok {
			inst.Callee = target
			*changed = true
		}
		if target, ok := localAliases[inst.This]; ok {
			inst.This = target
			*changed = true
		}

		// Handle control flow: child blocks inherit dominating aliases & available exprs
		if inst.Op == ir.OpIf {
			inst.Then = p.processBlock(fn, inst.Then, assigned, localAvail, localAliases, changed)
			inst.Else = p.processBlock(fn, inst.Else, assigned, localAvail, localAliases, changed)

			// Memory may be modified inside if/else branches
			p.invalidateMemory(localAvail)
			result = append(result, inst)
			continue
		}

		if inst.Op == ir.OpWhile || inst.Op == ir.OpDoWhile {
			inst.Cond = p.processBlock(fn, inst.Cond, assigned, localAvail, localAliases, changed)
			inst.Body = p.processBlock(fn, inst.Body, assigned, localAvail, localAliases, changed)
			if len(inst.Step) > 0 {
				inst.Step = p.processBlock(fn, inst.Step, assigned, localAvail, localAliases, changed)
			}

			// Loops may modify memory and loop-assigned variables
			p.invalidateMemory(localAvail)
			result = append(result, inst)
			continue
		}

		// Side effects and memory invalidations
		switch inst.Op {
		case ir.OpAssign:
			p.invalidateVariable(localAvail, inst.Result)
			delete(localAliases, inst.Result)
			result = append(result, inst)
			continue

		case ir.OpIndexSet:
			if len(inst.Args) > 0 {
				p.invalidateArray(localAvail, inst.Args[0])
			} else {
				p.invalidateMemory(localAvail)
			}
			result = append(result, inst)
			continue

		case ir.OpFieldSet:
			if len(inst.Args) > 0 {
				p.invalidateObject(localAvail, inst.Args[0])
			} else {
				p.invalidateMemory(localAvail)
			}
			result = append(result, inst)
			continue

		case ir.OpCall, ir.OpDynamicCall, ir.OpClosureCall, ir.OpExternCall:
			p.invalidateMemory(localAvail)
			result = append(result, inst)
			continue
		}

		// Check if pure expression is already available in dominating scope
		key := p.expressionKey(inst)
		if key != "" {
			if existing, ok := localAvail[key]; ok && existing != inst.Result {
				if !isMutableOrNamed(fn, assigned, inst.Result) {
					localAliases[inst.Result] = existing
					*changed = true
					continue // eliminate duplicate pure instruction
				}
			}
			localAvail[key] = inst.Result
		}

		result = append(result, inst)
	}

	return result
}

func (p *csePass) expressionKey(inst ir.Instruction) string {
	switch inst.Op {
	case ir.OpIndex:
		if len(inst.Args) == 2 {
			return fmt.Sprintf("index:%s:%s:%s", inst.Type, inst.Args[0], inst.Args[1])
		}
	case ir.OpFieldGet:
		if len(inst.Args) == 1 {
			return fmt.Sprintf("field.get:%s:%s:%d", inst.Type, inst.Args[0], inst.FieldIndex)
		}
	case ir.OpBinary, ir.OpCompare:
		if len(inst.Args) == 2 {
			return fmt.Sprintf("%s:%s:%s:%s:%s", inst.Op, inst.Type, inst.Operator, inst.Args[0], inst.Args[1])
		}
	case ir.OpSelect:
		if len(inst.Args) == 3 {
			return fmt.Sprintf("select:%s:%s:%s:%s", inst.Type, inst.Args[0], inst.Args[1], inst.Args[2])
		}
	}
	return ""
}

func (p *csePass) invalidateVariable(avail map[string]string, varName string) {
	for k := range avail {
		if strings.Contains(k, ":"+varName+":") || strings.HasSuffix(k, ":"+varName) {
			delete(avail, k)
		}
	}
}

func (p *csePass) invalidateArray(avail map[string]string, arrName string) {
	prefix := fmt.Sprintf("index:%s:", arrName)
	for k := range avail {
		if strings.Contains(k, prefix) {
			delete(avail, k)
		}
	}
}

func (p *csePass) invalidateObject(avail map[string]string, objName string) {
	prefix := fmt.Sprintf("field.get:%s:", objName)
	for k := range avail {
		if strings.Contains(k, prefix) {
			delete(avail, k)
		}
	}
}

func (p *csePass) invalidateMemory(avail map[string]string) {
	for k := range avail {
		if strings.HasPrefix(k, "index:") || strings.HasPrefix(k, "field.get:") {
			delete(avail, k)
		}
	}
}

func cloneMap(m map[string]string) map[string]string {
	res := make(map[string]string, len(m))
	for k, v := range m {
		res[k] = v
	}
	return res
}
