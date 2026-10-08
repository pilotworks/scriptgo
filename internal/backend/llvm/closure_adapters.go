package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func collectClosureCallees(instructions []ir.Instruction, callees map[string]bool) {
	for _, instruction := range instructions {
		if instruction.Op == ir.OpClosure {
			callees[instruction.Callee] = true
		}
		collectClosureCallees(instruction.Then, callees)
		collectClosureCallees(instruction.Else, callees)
		collectClosureCallees(instruction.Cond, callees)
		collectClosureCallees(instruction.Body, callees)
		collectClosureCallees(instruction.Step, callees)
		collectClosureCallees(instruction.Catch, callees)
		collectClosureCallees(instruction.Finally, callees)
	}
}

func emitClosureInvokeAdapter(function ir.Function) string {
	const valueType = "{ i32, i32, i64, i64 }"
	name := mangleFunctionName(function.Name)
	params := "ptr %env, i32 %t0, i32 %f0, i64 %p0, i32 %t1, i32 %f1, i64 %p1, i32 %t2, i32 %f2, i64 %p2, i32 %t3, i32 %f3, i64 %p3"
	var out strings.Builder
	fmt.Fprintf(&out, "define internal void @%s$invoke(%s, ptr %%out) nounwind {\n", name, params)

	var callArgs []string
	valIdx := 0
	for _, param := range function.Parameters {
		if param.Name == "__env_ctx" {
			callArgs = append(callArgs, "ptr %env")
			continue
		}
		if isRawCallbackParameter(param) {
			callArgs = append(callArgs, fmt.Sprintf("i32 %%t%d, i32 %%f%d, i64 %%p%d", valIdx, valIdx, valIdx))
			valIdx++
			continue
		}
		pType := llvmType(param.Type)
		switch param.Type {
		case ir.TypeNumber:
			casted := fmt.Sprintf("arg.num.%d", valIdx)
			fmt.Fprintf(&out, "  %%%s = bitcast i64 %%p%d to double\n", casted, valIdx)
			callArgs = append(callArgs, fmt.Sprintf("double %%%s", casted))
		case ir.TypeBool:
			casted := fmt.Sprintf("arg.bool.%d", valIdx)
			fmt.Fprintf(&out, "  %%%s = trunc i64 %%p%d to i1\n", casted, valIdx)
			if pType == "zeroext i1" {
				callArgs = append(callArgs, fmt.Sprintf("zeroext i1 %%%s", casted))
			} else {
				callArgs = append(callArgs, fmt.Sprintf("i1 %%%s", casted))
			}
		case ir.TypeUnknown:
			b0 := fmt.Sprintf("arg.box0.%d", valIdx)
			b1 := fmt.Sprintf("arg.box1.%d", valIdx)
			b2 := fmt.Sprintf("arg.box2.%d", valIdx)
			fmt.Fprintf(&out, "  %%%s = insertvalue %s zeroinitializer, i32 %%t%d, 0\n", b0, valueType, valIdx)
			fmt.Fprintf(&out, "  %%%s = insertvalue %s %%%s, i32 %%f%d, 1\n", b1, valueType, b0, valIdx)
			fmt.Fprintf(&out, "  %%%s = insertvalue %s %%%s, i64 %%p%d, 2\n", b2, valueType, b1, valIdx)
			callArgs = append(callArgs, fmt.Sprintf("%s %%%s", valueType, b2))
		default:
			if pType == "ptr" {
				casted := fmt.Sprintf("arg.ptr.%d", valIdx)
				fmt.Fprintf(&out, "  %%%s = inttoptr i64 %%p%d to ptr\n", casted, valIdx)
				callArgs = append(callArgs, fmt.Sprintf("ptr %%%s", casted))
			} else {
				callArgs = append(callArgs, fmt.Sprintf("%s zeroinitializer", pType))
			}
		}
		valIdx++
	}
	args := strings.Join(callArgs, ", ")

	if function.ReturnType == ir.TypeUnknown {
		fmt.Fprintf(&out, "  %%result = call %s @%s(%s)\n", valueType, name, args)
		fmt.Fprintf(&out, "  store %s %%result, ptr %%out\n", valueType)
		out.WriteString("  ret void\n}\n\n")
		return out.String()
	}
	if function.ReturnType == ir.TypeVoid {
		fmt.Fprintf(&out, "  call void @%s(%s)\n", name, args)
		fmt.Fprintf(&out, "  store %s zeroinitializer, ptr %%out\n", valueType)
		out.WriteString("  ret void\n}\n\n")
		return out.String()
	}

	returnType := llvmType(function.ReturnType)
	fmt.Fprintf(&out, "  %%typed = call %s @%s(%s)\n", returnType, name, args)
	tag := closureReturnTag(function.ReturnType)
	payload := "%typed"
	switch function.ReturnType {
	case ir.TypeNumber:
		out.WriteString("  %payload = bitcast double %typed to i64\n")
		payload = "%payload"
	case ir.TypeBool:
		out.WriteString("  %payload = zext i1 %typed to i64\n")
		payload = "%payload"
	case ir.TypeBigInt:
	default:
		out.WriteString("  %payload = ptrtoint ptr %typed to i64\n")
		payload = "%payload"
	}
	fmt.Fprintf(&out, "  %%boxed.0 = insertvalue %s zeroinitializer, i32 %d, 0\n", valueType, tag)
	fmt.Fprintf(&out, "  %%boxed.1 = insertvalue %s %%boxed.0, i64 %s, 2\n", valueType, payload)
	fmt.Fprintf(&out, "  store %s %%boxed.1, ptr %%out\n", valueType)
	out.WriteString("  ret void\n}\n\n")
	return out.String()
}

func mangleFunctionName(name string) string {
	switch name {
	case "close", "open", "read", "write", "exit", "abort", "link", "unlink", "remove", "rename", "stat", "pipe", "fork", "kill", "signal", "listen", "connect", "bind", "accept", "send", "recv", "select", "poll", "system", "pause", "sleep", "alarm":
		return name + "$user"
	default:
		return name
	}
}
