package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitArrayCopyIntrinsic(out *strings.Builder, instruction ir.Instruction, arrayType ir.Type) error {
	switch instruction.Callee {
	case "__array.slice":
		if (len(instruction.Args) < 1 || len(instruction.Args) > 3) || (!strings.HasSuffix(string(instruction.Type), "[]") && instruction.Type != ir.TypeNumberArray && instruction.Type != ir.TypeStringArray && instruction.Type != ir.TypeBoolArray && instruction.Type != ir.TypeBigIntArray && instruction.Type != ir.TypeUnknownArray) {
			return fmt.Errorf("array.slice has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		startArg := "0.0"
		if len(instruction.Args) >= 2 {
			startArg = "%" + e.resolveArg(out, instruction.Args[1])
		}
		endArg := "0x7FF0000000000000" // +Infinity: end omitted
		if len(instruction.Args) == 3 {
			endArg = "%" + e.resolveArg(out, instruction.Args[2])
		}
		targetElemSize := 8
		if instruction.Type == ir.TypeBoolArray || instruction.Type == "bool[]" || instruction.Type == "boolean[]" {
			targetElemSize = 1
		} else if instruction.Type == ir.TypeUnknownArray || instruction.Type == "unknown[]" || instruction.Type == "any[]" {
			targetElemSize = 24
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_slice_with_size(ptr %%%s, double %s, double %s, i64 %d, ptr %%%s)\n", status, arrArg, startArg, endArg, targetElemSize, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.concat":
		if len(instruction.Args) != 2 || instruction.Type != arrayType {
			return fmt.Errorf("array.concat has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		otherArg := e.ensurePointerArg(out, instruction.Args[1])
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_concat(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, arrArg, otherArg, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.join":
		if (len(instruction.Args) != 1 && len(instruction.Args) != 2) || instruction.Type != ir.TypeString {
			return fmt.Errorf("array.join has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		sepArg := "null"
		if len(instruction.Args) == 2 {
			sepArg = "%" + e.resolveArg(out, instruction.Args[1])
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		switch arrayType {
		case ir.TypeNumberArray:
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_number(ptr %%%s, ptr %s, ptr %%%s)\n", status, arrArg, sepArg, resSlot)
		case ir.TypeStringArray:
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_string(ptr %%%s, ptr %s, ptr %%%s)\n", status, arrArg, sepArg, resSlot)
		case ir.TypeBigIntArray:
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_bigint(ptr %%%s, ptr %s, ptr %%%s)\n", status, arrArg, sepArg, resSlot)
		default:
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_unknown(ptr %%%s, ptr %s, ptr %%%s)\n", status, arrArg, sepArg, resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.toReversed":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_to_reversed(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.toSorted":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fnName := "scriptgo_array_to_sorted_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_to_sorted_string"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.with":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fnName := "scriptgo_array_with_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_with_string"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, double %%%s, %s %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], llvmType(arrayElementType(arrayType)), instruction.Args[2], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.toSpliced":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		delCount := "0.000000e+00"
		hasDel := 0
		if len(instruction.Args) > 2 {
			delCount = fmt.Sprintf("%%%s", instruction.Args[2])
			hasDel = 1
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_to_spliced(ptr %%%s, double %%%s, double %s, i32 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], delCount, hasDel, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.toString", "__array.toLocaleString":
		e.emitArrayToString(out, "%"+instruction.Args[0], arrayType, instruction.Result)
		return nil
	default:
		return fmt.Errorf("unknown array intrinsic %q", instruction.Callee)
	}
}

// emitArrayToString emits Array.prototype.toString (join with ",") for an
// array of any element type into result.
func (e *functionEmitter) emitArrayToString(out *strings.Builder, arrayArg string, arrayType ir.Type, result string) {
	resSlot := result + ".slot"
	fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	fn := "scriptgo_array_join_unknown"
	switch arrayType {
	case ir.TypeNumberArray:
		fn = "scriptgo_array_join_number"
	case ir.TypeStringArray:
		fn = "scriptgo_array_join_string"
	case ir.TypeBigIntArray:
		fn = "scriptgo_array_join_bigint"
	}
	fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %s, ptr null, ptr %%%s)\n", status, fn, arrayArg, resSlot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", result, resSlot)
}
