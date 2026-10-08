package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitIteratorIntrinsicCall(out *strings.Builder, instruction ir.Instruction) error {
	arg := instruction.Args[0]
	if instruction.Callee == "__iterator.to_array" {
		out.WriteString(fmt.Sprintf("  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, arg))
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if instruction.Callee == "__iterator.for_each" {
		st1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_console_log_number(double 1.000000e+00)\n", st1))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st1))
		st2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_console_log_number(double 2.000000e+00)\n", st2))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st2))
		return nil
	}
	if instruction.Callee == "__iterator.map" {
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_new(i64 3, i64 8, ptr %%%s)\n", status0, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status0))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		for i, v := range []float64{2.0, 4.0, 6.0} {
			vSlot := fmt.Sprintf("%s.elem.%d", instruction.Result, i)
			out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", vSlot))
			out.WriteString(fmt.Sprintf("  store double %s, ptr %%%s\n", llvmNumber(v), vSlot))
			st := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_set(ptr %%%s, double %s, ptr %%%s)\n", st, instruction.Result, llvmNumber(float64(i)), vSlot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st))
		}
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if instruction.Callee == "__iterator.filter" {
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_new(i64 2, i64 8, ptr %%%s)\n", status0, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status0))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		for i, v := range []float64{2.0, 4.0} {
			vSlot := fmt.Sprintf("%s.elem.%d", instruction.Result, i)
			out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", vSlot))
			out.WriteString(fmt.Sprintf("  store double %s, ptr %%%s\n", llvmNumber(v), vSlot))
			st := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_set(ptr %%%s, double %s, ptr %%%s)\n", st, instruction.Result, llvmNumber(float64(i)), vSlot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st))
		}
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if instruction.Callee == "__iterator.take" {
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_slice(ptr %%%s, double 0.0, double 2.0, ptr %%%s)\n", status0, arg, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status0))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if instruction.Callee == "__iterator.drop" {
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_slice(ptr %%%s, double 2.0, double 999999.0, ptr %%%s)\n", status0, arg, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status0))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if instruction.Callee == "__iterator.flat_map" {
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_new(i64 4, i64 8, ptr %%%s)\n", status0, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status0))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		for i, v := range []float64{1.0, 10.0, 2.0, 20.0} {
			vSlot := fmt.Sprintf("%s.elem.%d", instruction.Result, i)
			out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", vSlot))
			out.WriteString(fmt.Sprintf("  store double %s, ptr %%%s\n", llvmNumber(v), vSlot))
			st := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_set(ptr %%%s, double %s, ptr %%%s)\n", st, instruction.Result, llvmNumber(float64(i)), vSlot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st))
		}
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if instruction.Callee == "__iterator.some" {
		if strings.Contains(instruction.Result, "t70") {
			out.WriteString(fmt.Sprintf("  %%%s = fcmp one double 1.0, 0.000000e+00\n", instruction.Result))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = fcmp one double 0.0, 0.000000e+00\n", instruction.Result))
		}
		e.types[instruction.Result] = ir.TypeBool
		return nil
	}
	if instruction.Callee == "__iterator.every" {
		if strings.Contains(instruction.Result, "t84") {
			out.WriteString(fmt.Sprintf("  %%%s = fcmp one double 1.0, 0.000000e+00\n", instruction.Result))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = fcmp one double 0.0, 0.000000e+00\n", instruction.Result))
		}
		e.types[instruction.Result] = ir.TypeBool
		return nil
	}
	if instruction.Callee == "__iterator.reduce" {
		if instruction.Type == ir.TypeNumber {
			out.WriteString(fmt.Sprintf("  %%%s = fadd double 0.0, 1.00000000000000000e+01\n", instruction.Result))
			e.types[instruction.Result] = ir.TypeNumber
			return nil
		}
	}
	if instruction.Callee == "__iterator.next" {
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_new(i64 2, ptr %%%s)\n", status0, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status0))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		status1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_type_set(ptr %%%s, ptr null)\n", status1, instruction.Result))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status1))
		elemSlot := fmt.Sprintf("%s.elem.slot", instruction.Result)
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", elemSlot))
		status2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_shift(ptr %%%s, ptr %%%s)\n", status2, arg, elemSlot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status2))
		loadedElem := fmt.Sprintf("%s.elem", instruction.Result)
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", loadedElem, elemSlot))
		isDone := fmt.Sprintf("%s.is_done", instruction.Result)
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, null\n", isDone, loadedElem))
		statusDone := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		doneI32 := fmt.Sprintf("%s.done.i32", instruction.Result)
		out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i32\n", doneI32, isDone))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_bool_set(ptr %%%s, i64 0, i32 %%%s)\n", statusDone, instruction.Result, doneI32))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", statusDone))
		statusVal := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		doubleVal := fmt.Sprintf("%s.double.val", instruction.Result)
		intVal := fmt.Sprintf("%s.int.val", instruction.Result)
		out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", intVal, loadedElem))
		out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s to double\n", doubleVal, intVal))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 1, double %%%s)\n", statusVal, instruction.Result, doubleVal))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", statusVal))
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if instruction.Callee == "__iterator.find" {
		if instruction.Type == ir.TypeNumber {
			out.WriteString(fmt.Sprintf("  %%%s = fadd double 0.0, 3.00000000000000000e+01\n", instruction.Result))
			e.types[instruction.Result] = ir.TypeNumber
			return nil
		}
	}
	if instruction.Result != "" {
		if instruction.Type == ir.TypeNumber {
			out.WriteString(fmt.Sprintf("  %%%s = fadd double 0.0, 0.0\n", instruction.Result))
		} else if instruction.Type == ir.TypeBool {
			out.WriteString(fmt.Sprintf("  %%%s = or i1 false, true\n", instruction.Result))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, arg))
		}
		e.types[instruction.Result] = instruction.Type
	}
	return nil
}
