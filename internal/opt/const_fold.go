package opt

import (
	"math"
	"strconv"

	"github.com/pilotworks/scriptgo/internal/ir"
)

type constFoldPass struct{}

// NewConstFoldPass creates a constant folding and algebraic simplification pass.
func NewConstFoldPass() Pass {
	return &constFoldPass{}
}

func (p *constFoldPass) Name() string {
	return "const-fold"
}

func (p *constFoldPass) Run(m *ir.Module) (bool, error) {
	changed := false
	for i := range m.Functions {
		fnChanged := p.runOnFunction(&m.Functions[i])
		if fnChanged {
			changed = true
		}
	}
	return changed, nil
}

func (p *constFoldPass) runOnFunction(fn *ir.Function) bool {
	assigned := make(map[string]bool)
	collectAssigned(fn.Body, assigned)

	numConsts := make(map[string]float64)
	boolConsts := make(map[string]bool)
	strConsts := make(map[string]string)
	aliases := make(map[string]string)

	changed := false
	fn.Body = p.foldBlock(fn, fn.Body, assigned, numConsts, boolConsts, strConsts, aliases, &changed)
	return changed
}

func (p *constFoldPass) foldBlock(
	fn *ir.Function,
	body []ir.Instruction,
	assigned map[string]bool,
	numConsts map[string]float64,
	boolConsts map[string]bool,
	strConsts map[string]string,
	aliases map[string]string,
	changed *bool,
) []ir.Instruction {
	localNums := cloneMapFloat(numConsts)
	localBools := cloneMapBool(boolConsts)
	localStrs := cloneMapStr(strConsts)
	localAliases := cloneMap(aliases)
	var result []ir.Instruction

	for _, inst := range body {
		// Resolve alias operands first
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

		// Recursively fold nested blocks with downward-propagating environments
		if len(inst.Cond) > 0 {
			inst.Cond = p.foldBlock(fn, inst.Cond, assigned, localNums, localBools, localStrs, localAliases, changed)
		}
		if len(inst.Body) > 0 {
			inst.Body = p.foldBlock(fn, inst.Body, assigned, localNums, localBools, localStrs, localAliases, changed)
		}
		if len(inst.Step) > 0 {
			inst.Step = p.foldBlock(fn, inst.Step, assigned, localNums, localBools, localStrs, localAliases, changed)
		}
		if len(inst.Then) > 0 {
			inst.Then = p.foldBlock(fn, inst.Then, assigned, localNums, localBools, localStrs, localAliases, changed)
		}
		if len(inst.Else) > 0 {
			inst.Else = p.foldBlock(fn, inst.Else, assigned, localNums, localBools, localStrs, localAliases, changed)
		}
		if len(inst.Catch) > 0 {
			inst.Catch = p.foldBlock(fn, inst.Catch, assigned, localNums, localBools, localStrs, localAliases, changed)
		}
		if len(inst.Finally) > 0 {
			inst.Finally = p.foldBlock(fn, inst.Finally, assigned, localNums, localBools, localStrs, localAliases, changed)
		}

		switch inst.Op {
		case ir.OpConst:
			if !isMutableOrNamed(fn, assigned, inst.Result) {
				if inst.Type == ir.TypeNumber {
					if v, err := strconv.ParseFloat(inst.Value, 64); err == nil {
						localNums[inst.Result] = v
					}
				} else if inst.Type == ir.TypeBool {
					localBools[inst.Result] = (inst.Value == "true")
				} else if inst.Type == ir.TypeString && inst.StringLiteral {
					localStrs[inst.Result] = inst.Value
				}
			}

		case ir.OpAssign:
			delete(localAliases, inst.Result)
			delete(localNums, inst.Result)
			delete(localBools, inst.Result)
			delete(localStrs, inst.Result)

		case ir.OpBinary:
			if len(inst.Args) == 2 && inst.Type == ir.TypeNumber {
				left, right := inst.Args[0], inst.Args[1]
				leftVal, leftIsConst := localNums[left]
				rightVal, rightIsConst := localNums[right]

				// Constant folding: both are known numbers
				if leftIsConst && rightIsConst {
					if res, ok := evalBinaryNumber(inst.Operator, leftVal, rightVal); ok {
						inst.Op = ir.OpConst
						inst.Value = formatNumber(res)
						inst.Args = nil
						inst.Operator = ""
						if !isMutableOrNamed(fn, assigned, inst.Result) {
							localNums[inst.Result] = res
						}
						*changed = true
						result = append(result, inst)
						continue
					}
				}

				// Algebraic identities
				// x + 0 -> x
				if inst.Operator == "+" && rightIsConst && rightVal == 0 {
					if p.simplifyTo(fn, assigned, localAliases, &inst, left, changed) {
						continue
					}
					result = append(result, inst)
					continue
				}
				// 0 + x -> x
				if inst.Operator == "+" && leftIsConst && leftVal == 0 {
					if p.simplifyTo(fn, assigned, localAliases, &inst, right, changed) {
						continue
					}
					result = append(result, inst)
					continue
				}
				// x - 0 -> x
				if inst.Operator == "-" && rightIsConst && rightVal == 0 {
					if p.simplifyTo(fn, assigned, localAliases, &inst, left, changed) {
						continue
					}
					result = append(result, inst)
					continue
				}
				// x * 1 -> x
				if inst.Operator == "*" && rightIsConst && rightVal == 1 {
					if p.simplifyTo(fn, assigned, localAliases, &inst, left, changed) {
						continue
					}
					result = append(result, inst)
					continue
				}
				// 1 * x -> x
				if inst.Operator == "*" && leftIsConst && leftVal == 1 {
					if p.simplifyTo(fn, assigned, localAliases, &inst, right, changed) {
						continue
					}
					result = append(result, inst)
					continue
				}
				// x / 1 -> x
				if inst.Operator == "/" && rightIsConst && rightVal == 1 {
					if p.simplifyTo(fn, assigned, localAliases, &inst, left, changed) {
						continue
					}
					result = append(result, inst)
					continue
				}
			}

		case ir.OpCompare:
			if len(inst.Args) == 2 {
				left, right := inst.Args[0], inst.Args[1]
				leftNum, leftIsNum := localNums[left]
				rightNum, rightIsNum := localNums[right]
				if leftIsNum && rightIsNum {
					if res, ok := evalCompareNumber(inst.Operator, leftNum, rightNum); ok {
						inst.Op = ir.OpConst
						inst.Type = ir.TypeBool
						inst.Value = strconv.FormatBool(res)
						inst.Args = nil
						inst.Operator = ""
						if !isMutableOrNamed(fn, assigned, inst.Result) {
							localBools[inst.Result] = res
						}
						*changed = true
						result = append(result, inst)
						continue
					}
				}
			}
		}

		result = append(result, inst)
	}
	return result
}

func (p *constFoldPass) simplifyTo(
	fn *ir.Function,
	assigned map[string]bool,
	aliases map[string]string,
	inst *ir.Instruction,
	target string,
	changed *bool,
) bool {
	if isMutableOrNamed(fn, assigned, inst.Result) {
		return false
	}
	aliases[inst.Result] = target
	*changed = true
	return true
}

func isMutableOrNamed(fn *ir.Function, assigned map[string]bool, name string) bool {
	if assigned[name] {
		return true
	}
	for _, p := range fn.Parameters {
		if p.Name == name {
			return true
		}
	}
	for _, l := range fn.Locals {
		if l.Name == name {
			return true
		}
	}
	for _, c := range fn.Captured {
		if c.Name == name {
			return true
		}
	}
	return false
}

func evalBinaryNumber(op string, l, r float64) (float64, bool) {
	switch op {
	case "+":
		return l + r, true
	case "-":
		return l - r, true
	case "*":
		return l * r, true
	case "/":
		if r == 0 {
			return 0, false
		}
		return l / r, true
	case "%":
		if r == 0 {
			return 0, false
		}
		return math.Mod(l, r), true
	case "&":
		return float64(int64(l) & int64(r)), true
	case "|":
		return float64(int64(l) | int64(r)), true
	case "^":
		return float64(int64(l) ^ int64(r)), true
	case "<<":
		shift := uint64(int64(r) & 31)
		return float64(int32(int64(l) << shift)), true
	case ">>":
		shift := uint64(int64(r) & 31)
		return float64(int32(int64(l) >> shift)), true
	default:
		return 0, false
	}
}

func evalCompareNumber(op string, l, r float64) (bool, bool) {
	switch op {
	case "==", "===":
		return l == r, true
	case "!=", "!==":
		return l != r, true
	case "<":
		return l < r, true
	case "<=":
		return l <= r, true
	case ">":
		return l > r, true
	case ">=":
		return l >= r, true
	default:
		return false, false
	}
}

func formatNumber(val float64) string {
	if val == math.Trunc(val) && !math.IsNaN(val) && !math.IsInf(val, 0) {
		return strconv.FormatInt(int64(val), 10)
	}
	return strconv.FormatFloat(val, 'f', -1, 64)
}

func collectAssigned(instructions []ir.Instruction, assigned map[string]bool) {
	for _, inst := range instructions {
		if inst.Op == ir.OpAssign && inst.Result != "" {
			assigned[inst.Result] = true
		}
		collectAssigned(inst.Cond, assigned)
		collectAssigned(inst.Body, assigned)
		collectAssigned(inst.Step, assigned)
		collectAssigned(inst.Then, assigned)
		collectAssigned(inst.Else, assigned)
		collectAssigned(inst.Catch, assigned)
		collectAssigned(inst.Finally, assigned)
	}
}

func cloneMapFloat(m map[string]float64) map[string]float64 {
	res := make(map[string]float64, len(m))
	for k, v := range m {
		res[k] = v
	}
	return res
}

func cloneMapBool(m map[string]bool) map[string]bool {
	res := make(map[string]bool, len(m))
	for k, v := range m {
		res[k] = v
	}
	return res
}

func cloneMapStr(m map[string]string) map[string]string {
	res := make(map[string]string, len(m))
	for k, v := range m {
		res[k] = v
	}
	return res
}
