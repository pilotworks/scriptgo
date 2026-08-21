package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitTypedArrayIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__arraybuffer.new":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("arraybuffer.new requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = fptosi double %%%s to i64\n", instruction.Args[0]+".i64", instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_arraybuffer_new(i64 %%%s, ptr %%%s)\n", status, instruction.Args[0]+".i64", slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__arraybuffer.byteLength":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("arraybuffer.byteLength requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_arraybuffer_byte_length(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__arraybuffer.slice":
		if len(instruction.Args) < 3 {
			return fmt.Errorf("arraybuffer.slice requires 3 arguments")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_arraybuffer_slice(ptr %%%s, double %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__arraybuffer.isView":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("arraybuffer.isView requires 1 argument")
		}
		argType, ok := e.types[instruction.Args[0]]
		isView := ok && (argType == ir.TypeUint8Array || argType == ir.TypeInt32Array || argType == ir.TypeFloat64Array)
		val := "0"
		if isView {
			val = "1"
		}
		fmt.Fprintf(out, "  %%%s = icmp eq i1 %s, 1\n", instruction.Result, val)
		return nil

	case "__typedarray.new_length":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("typedarray.new_length requires 1 argument")
		}
		kind := 1
		if instruction.Type == ir.TypeInt32Array {
			kind = 2
		} else if instruction.Type == ir.TypeFloat64Array {
			kind = 3
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = fptosi double %%%s to i64\n", instruction.Args[0]+".i64", instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_new(i64 %d, i64 %%%s, ptr null, i64 0, ptr %%%s)\n", status, kind, instruction.Args[0]+".i64", slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.new_buffer":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("typedarray.new_buffer requires 3 arguments")
		}
		kind := 1
		if instruction.Type == ir.TypeInt32Array {
			kind = 2
		} else if instruction.Type == ir.TypeFloat64Array {
			kind = 3
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = fptosi double %%%s to i64\n", instruction.Args[1]+".offset.i64", instruction.Args[1])
		fmt.Fprintf(out, "  %%%s = fptosi double %%%s to i64\n", instruction.Args[2]+".len.i64", instruction.Args[2])
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_new(i64 %d, i64 %%%s, ptr %%%s, i64 %%%s, ptr %%%s)\n", status, kind, instruction.Args[2]+".len.i64", instruction.Args[0], instruction.Args[1]+".offset.i64", slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.length":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("typedarray.length requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_length(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.byteLength":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("typedarray.byteLength requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_byte_length(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.byteOffset":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("typedarray.byteOffset requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_byte_offset(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.buffer":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("typedarray.buffer requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_buffer(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.subarray":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("typedarray.subarray requires 3 arguments")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_subarray(ptr %%%s, double %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.slice":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("typedarray.slice requires 3 arguments")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_slice(ptr %%%s, double %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.set":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("typedarray.set requires 3 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_set_array(ptr %%%s, ptr %%%s, double %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__typedarray.new_array":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("typedarray.new_array requires 2 arguments")
		}
		kind := 1
		if instruction.Type == ir.TypeInt32Array {
			kind = 2
		} else if instruction.Type == ir.TypeFloat64Array {
			kind = 3
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_from_array(i64 %d, ptr %%%s, ptr %%%s)\n", status, kind, instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.toString":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("typedarray.toString requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_to_string(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__arraybuffer.toString":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("arraybuffer.toString requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_arraybuffer_to_string(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__typedarray.fill":
		if len(instruction.Args) != 4 {
			return fmt.Errorf("typedarray.fill requires 4 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_typedarray_fill(ptr %%%s, double %%%s, double %%%s, double %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	default:
		return fmt.Errorf("unknown typedarray intrinsic %q", instruction.Callee)
	}
}
