package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitArraySearchIntrinsic(out *strings.Builder, instruction ir.Instruction, arrayType ir.Type) error {
	switch instruction.Callee {
	case "__array.indexOf":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("array.indexOf has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		targetArg := e.resolveArg(out, instruction.Args[1])
		fromArg := "0.0"
		if len(instruction.Args) == 3 {
			fromArg = "%" + e.resolveArg(out, instruction.Args[2])
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		elemLLVMType := arrayElementLLVMType(arrayType)
		if elemLLVMType == "ptr" {
			if arrayType == ir.TypeStringArray {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_index_of_string(ptr %%%s, ptr %%%s, double %s, ptr %%%s)\n", status, arrArg, targetArg, fromArg, resSlot)
			} else {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_index_of_ptr(ptr %%%s, ptr %%%s, double %s, ptr %%%s)\n", status, arrArg, targetArg, fromArg, resSlot)
			}
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_index_of_number(ptr %%%s, double %%%s, double %s, ptr %%%s)\n", status, arrArg, targetArg, fromArg, resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.includes":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("array.includes has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		targetArg := e.resolveArg(out, instruction.Args[1])
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		incElemLLVMType := arrayElementLLVMType(arrayType)
		if incElemLLVMType == "ptr" {
			if arrayType == ir.TypeStringArray {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_includes_string(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, arrArg, targetArg, resSlot)
			} else {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_includes_ptr(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, arrArg, targetArg, resSlot)
			}
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_includes_number(ptr %%%s, double %%%s, ptr %%%s)\n", status, arrArg, targetArg, resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, resSlot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
		return nil
	case "__array.at":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("array.at has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		idxArg := e.resolveArg(out, instruction.Args[1])
		elemLLVMType := arrayElementLLVMType(arrayType)
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca %s\n", resSlot, elemLLVMType)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_at(ptr %%%s, double %%%s, ptr %%%s)\n", status, arrArg, idxArg, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load %s, ptr %%%s\n", instruction.Result, elemLLVMType, resSlot)
		return nil
	case "__array.find":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_find_number(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.some":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_some_number(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s.i32 = load i32, ptr %%%s\n", instruction.Result, slot))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s.i32, 0\n", instruction.Result, instruction.Result))
		return nil
	case "__array.every":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_every_number(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s.i32 = load i32, ptr %%%s\n", instruction.Result, slot))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s.i32, 0\n", instruction.Result, instruction.Result))
		return nil
	case "__array.findIndex":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fnName := "scriptgo_array_find_index_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_find_index_string"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.findLast":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_find_last_number(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.findLastIndex":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fnName := "scriptgo_array_find_last_index_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_find_last_index_string"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.lastIndexOf":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("array.lastIndexOf has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		targetArg := e.resolveArg(out, instruction.Args[1])
		resSlot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", resSlot))
		fromArg := "-1.000000e+00"
		if len(instruction.Args) > 2 {
			fromArg = fmt.Sprintf("%%%s", e.resolveArg(out, instruction.Args[2]))
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if arrayType == ir.TypeStringArray {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_last_index_of_string(ptr %%%s, ptr %%%s, double %s, ptr %%%s)\n", status, arrArg, targetArg, fromArg, resSlot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_last_index_of_number(ptr %%%s, double %%%s, double %s, ptr %%%s)\n", status, arrArg, targetArg, fromArg, resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	default:
		return fmt.Errorf("unknown array intrinsic %q", instruction.Callee)
	}
}
