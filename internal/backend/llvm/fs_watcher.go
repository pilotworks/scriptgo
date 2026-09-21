package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitFSWatcherIntrinsic(out *strings.Builder, instruction ir.Instruction) (bool, error) {
	resolvedArgs := make([]string, len(instruction.Args))
	for i, arg := range instruction.Args {
		resolvedArgs[i] = e.resolveArg(out, arg)
	}

	switch instruction.Callee {
	case "__fs.watchCreate":
		pathArg := "%" + resolvedArgs[0]
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_watch_create(ptr %s, ptr %%%s)\n", status, pathArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return true, nil

	case "__fs.watchPoll":
		idArg := "%" + resolvedArgs[0]
		typeSlot := instruction.Result + ".type.slot"
		fileSlot := instruction.Result + ".file.slot"
		hasEventSlot := instruction.Result + ".has_event.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", typeSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", fileSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", hasEventSlot)

		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_watch_poll(double %s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, idArg, typeSlot, fileSlot, hasEventSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		typeVal := instruction.Result + ".type"
		fileVal := instruction.Result + ".file"
		hasEventVal := instruction.Result + ".has_event"
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", typeVal, typeSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", fileVal, fileSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", hasEventVal, hasEventSlot)

		cmpHas := fmt.Sprintf("cmp.has.%d", e.loadCounter)
		e.loadCounter++
		zextHas := fmt.Sprintf("zext.has.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = fcmp ogt double %%%s, 0.0\n", cmpHas, hasEventVal)
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", zextHas, cmpHas)

		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 3, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)

		s1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		s2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		s3 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 0, ptr %%%s)\n", s1, instruction.Result, typeVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s1)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 1, ptr %%%s)\n", s2, instruction.Result, fileVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s2)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_bool_set(ptr %%%s, i64 2, i32 %%%s)\n", s3, instruction.Result, zextHas)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s3)

		descriptor := intrinsicObjectDescriptor(instruction.Callee)
		if descriptor != "" {
			if global, ok := e.stringsByValue[descriptor]; ok {
				tStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
				e.runtimeStatus++
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_type_set(ptr %%%s, ptr %s)\n", tStatus, instruction.Result, global)
				fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", tStatus)
			}
		}
		return true, nil

	case "__fs.watchClose":
		idArg := "%" + resolvedArgs[0]
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_watch_close(double %s)\n", status, idArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return true, nil
	}
	return false, nil
}
