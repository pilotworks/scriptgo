package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitAtomicsIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	slot := fmt.Sprintf("atomics.slot.%d", e.loadCounter)
	e.loadCounter++

	switch instruction.Callee {
	case "__atomics.sharedArrayBufferNew":
		arg0 := e.resolveArg(out, instruction.Args[0])
		i64Val := fmt.Sprintf("atomics.len.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = fptosi double %%%s to i64\n", i64Val, arg0)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_shared_array_buffer_new(i64 %%%s, ptr %%%s)\n", status, i64Val, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)

	case "__atomics.isLockFree":
		arg0 := e.resolveArg(out, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_atomics_is_lock_free(double %%%s, ptr %%%s)\n", status, arg0, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		i32Val := fmt.Sprintf("atomics.i32.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", i32Val, slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Val)

	case "__atomics.add", "__atomics.sub", "__atomics.and", "__atomics.or", "__atomics.xor", "__atomics.store", "__atomics.exchange":
		cFunc := strings.Replace(instruction.Callee, "__atomics.", "@scriptgo_atomics_", 1)
		handle := e.resolveArg(out, instruction.Args[0])
		idx := e.resolveArg(out, instruction.Args[1])
		val := e.resolveArg(out, instruction.Args[2])
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 %s(ptr %%%s, double %%%s, double %%%s, ptr %%%s)\n", status, cFunc, handle, idx, val, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)

	case "__atomics.load":
		handle := e.resolveArg(out, instruction.Args[0])
		idx := e.resolveArg(out, instruction.Args[1])
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_atomics_load(ptr %%%s, double %%%s, ptr %%%s)\n", status, handle, idx, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)

	case "__atomics.compareExchange":
		handle := e.resolveArg(out, instruction.Args[0])
		idx := e.resolveArg(out, instruction.Args[1])
		exp := e.resolveArg(out, instruction.Args[2])
		des := e.resolveArg(out, instruction.Args[3])
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_atomics_compare_exchange(ptr %%%s, double %%%s, double %%%s, double %%%s, ptr %%%s)\n", status, handle, idx, exp, des, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)

	case "__atomics.wait":
		handle := e.resolveArg(out, instruction.Args[0])
		idx := e.resolveArg(out, instruction.Args[1])
		val := e.resolveArg(out, instruction.Args[2])
		timeout := "0.0"
		if len(instruction.Args) > 3 {
			timeout = "%" + e.resolveArg(out, instruction.Args[3])
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_atomics_wait(ptr %%%s, double %%%s, double %%%s, double %s, ptr %%%s)\n", status, handle, idx, val, timeout, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)

	case "__atomics.notify":
		handle := e.resolveArg(out, instruction.Args[0])
		idx := e.resolveArg(out, instruction.Args[1])
		count := "1.0"
		if len(instruction.Args) > 2 {
			count = "%" + e.resolveArg(out, instruction.Args[2])
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_atomics_notify(ptr %%%s, double %%%s, double %s, ptr %%%s)\n", status, handle, idx, count, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)

	default:
		return fmt.Errorf("unknown atomics intrinsic %q", instruction.Callee)
	}
	return nil
}
