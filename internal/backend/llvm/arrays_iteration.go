package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitArrayIterationIntrinsic(out *strings.Builder, instruction ir.Instruction, arrayType ir.Type) error {
	switch instruction.Callee {
	case "__array.flatMap", "__array.flatMap_scalar":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		fnName := "scriptgo_array_flat_map_number"
		if instruction.Callee == "__array.flatMap_scalar" {
			fnName = "scriptgo_array_flat_map_number_scalar"
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.map":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		var fnName string
		retElemType := arrayElementType(instruction.Type)
		inElemType := arrayElementType(arrayType)
		if retElemType == ir.TypeNumber {
			if inElemType == ir.TypeString {
				fnName = "scriptgo_array_map_number_from_string"
			} else if inElemType != ir.TypeNumber {
				fnName = "scriptgo_array_map_number_from_ptr"
			} else {
				fnName = "scriptgo_array_map_number"
			}
		} else if retElemType == ir.TypeString {
			if inElemType == ir.TypeNumber {
				fnName = "scriptgo_array_map_string_from_number"
			} else if inElemType != ir.TypeString {
				fnName = "scriptgo_array_map_string_from_ptr"
			} else {
				fnName = "scriptgo_array_map_string"
			}
		} else {
			if inElemType == ir.TypeNumber {
				fnName = "scriptgo_array_map_ptr_from_number"
			} else if inElemType == ir.TypeString {
				fnName = "scriptgo_array_map_ptr_from_string"
			} else {
				fnName = "scriptgo_array_map_ptr"
			}
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.filter":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		fnName := "scriptgo_array_filter_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_filter_string"
		} else if arrayType != ir.TypeNumberArray && arrayType != "number[]" {
			fnName = "scriptgo_array_filter_ptr"
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.forEach":
		fnName := "scriptgo_array_for_each_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_for_each_string"
		} else if arrayType != ir.TypeNumberArray && arrayType != "number[]" {
			fnName = "scriptgo_array_for_each_ptr"
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	case "__array.reduce":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if len(instruction.Args) == 2 {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_reduce_number_seeded_left(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_reduce_number(ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot))
		}
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.reduceRight":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if len(instruction.Args) == 2 {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_reduce_number_seeded_right(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_reduce_right_number(ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot))
		}
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.values", "__array.flat":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_slice(ptr %%%s, double 0.0, double 0x7FF0000000000000, ptr %%%s)\n", status, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.keys":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_keys(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.entries":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_entries(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	default:
		return fmt.Errorf("unknown array intrinsic %q", instruction.Callee)
	}
}
