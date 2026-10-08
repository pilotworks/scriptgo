package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitHttpIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__http.fetchSync":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("fetchSync requires at least 1 argument")
		}
		methodArg := "null"
		headersArg := "null"
		bodyArg := "null"
		if len(instruction.Args) > 1 {
			methodArg = "%" + instruction.Args[1]
		}
		if len(instruction.Args) > 2 {
			headersArg = "%" + instruction.Args[2]
		}
		if len(instruction.Args) > 3 {
			bodyArg = "%" + instruction.Args[3]
		}
		statusSlot := instruction.Result + ".status.slot"
		statusTextSlot := instruction.Result + ".statusText.slot"
		headersSlot := instruction.Result + ".headers.slot"
		bodySlot := instruction.Result + ".body.slot"

		fmt.Fprintf(out, "  %%%s = alloca double\n", statusSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", statusTextSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", headersSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", bodySlot)

		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fetch_sync(ptr %%%s, ptr %s, ptr %s, ptr %s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], methodArg, headersArg, bodyArg, statusSlot, statusTextSlot, headersSlot, bodySlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		statusVal := instruction.Result + ".status"
		statusTextVal := instruction.Result + ".statusText"
		headersVal := instruction.Result + ".headers"
		bodyVal := instruction.Result + ".body"

		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", statusVal, statusSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", statusTextVal, statusTextSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", headersVal, headersSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", bodyVal, bodySlot)

		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 4, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)

		s1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		s2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		s3 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		s4 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++

		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 0, double %%%s)\n", s1, instruction.Result, statusVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s1)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 1, ptr %%%s)\n", s2, instruction.Result, statusTextVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s2)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 2, ptr %%%s)\n", s3, instruction.Result, headersVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s3)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 3, ptr %%%s)\n", s4, instruction.Result, bodyVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s4)
		return nil
	default:
		return fmt.Errorf("unknown http intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitStreamIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__stream.getDefaultHighWaterMark":
		arg := "0"
		if len(instruction.Args) > 0 {
			boolVal := fmt.Sprintf("%s.bool_i32.%d", instruction.Result, e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", boolVal, instruction.Args[0])
			arg = "%" + boolVal
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_stream_get_default_high_water_mark(i32 %s, ptr %%%s)\n", status, arg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__stream.setDefaultHighWaterMark":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("setDefaultHighWaterMark requires 2 arguments")
		}
		boolVal := fmt.Sprintf("%s.bool_i32.%d", instruction.Result, e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", boolVal, instruction.Args[0])
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_stream_set_default_high_water_mark(i32 %%%s, double %%%s)\n", status, boolVal, instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	default:
		return fmt.Errorf("unknown stream intrinsic %q", instruction.Callee)
	}
}
