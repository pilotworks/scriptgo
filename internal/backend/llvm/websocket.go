package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitWebSocketIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__websocket.connect":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("websocket.connect requires at least 1 argument")
		}
		urlArg := "%" + instruction.Args[0]
		protoArg := "null"
		if len(instruction.Args) > 1 {
			protoArg = "%" + instruction.Args[1]
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_websocket_connect(ptr %s, ptr %s, ptr %%%s)\n",
			status, urlArg, protoArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__websocket.sendText":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("websocket.sendText requires 2 arguments")
		}
		handleArg := "%" + instruction.Args[0]
		dataArg := "%" + instruction.Args[1]
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_websocket_send_text(double %s, ptr %s, ptr %%%s)\n",
			status, handleArg, dataArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__websocket.close":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("websocket.close requires at least 1 argument")
		}
		handleArg := "%" + instruction.Args[0]
		codeArg := "1000.0"
		reasonArg := "null"
		if len(instruction.Args) > 1 {
			codeArg = "%" + instruction.Args[1]
		}
		if len(instruction.Args) > 2 {
			reasonArg = "%" + instruction.Args[2]
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_websocket_close(double %s, double %s, ptr %s)\n",
			status, handleArg, codeArg, reasonArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__websocket.poll":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("websocket.poll requires 1 argument")
		}
		handleArg := "%" + instruction.Args[0]
		evSlot := instruction.Result + ".ev_slot"
		dataSlot := instruction.Result + ".data_slot"
		codeSlot := instruction.Result + ".code_slot"
		reasonSlot := instruction.Result + ".reason_slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", evSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", dataSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", codeSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", reasonSlot)

		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_websocket_poll(double %s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, handleArg, evSlot, dataSlot, codeSlot, reasonSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		evVal := instruction.Result + ".ev"
		dataVal := instruction.Result + ".data"
		codeVal := instruction.Result + ".code"
		reasonVal := instruction.Result + ".reason"
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", evVal, evSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", dataVal, dataSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", codeVal, codeSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", reasonVal, reasonSlot)

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

		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 0, double %%%s)\n", s1, instruction.Result, evVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s1)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 1, ptr %%%s)\n", s2, instruction.Result, dataVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s2)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 2, double %%%s)\n", s3, instruction.Result, codeVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s3)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 3, ptr %%%s)\n", s4, instruction.Result, reasonVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s4)
		return nil

	case "__websocket.readyState":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("websocket.readyState requires 1 argument")
		}
		handleArg := "%" + instruction.Args[0]
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_websocket_ready_state(double %s, ptr %%%s)\n",
			status, handleArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	default:
		return fmt.Errorf("unknown websocket intrinsic %q", instruction.Callee)
	}
}
