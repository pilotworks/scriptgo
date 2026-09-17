package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitTtyIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	resolvedArgs := make([]string, len(instruction.Args))
	for i, arg := range instruction.Args {
		resolvedArgs[i] = e.resolveArg(out, arg)
	}

	switch instruction.Callee {
	case "__tty.isatty":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("tty.isatty has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tty_isatty(double %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
		return nil

	case "__tty.getWindowSize":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("tty.getWindowSize has invalid signature")
		}
		colsSlot := instruction.Result + ".cols.slot"
		rowsSlot := instruction.Result + ".rows.slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", colsSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", rowsSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tty_get_window_size(double %%%s, ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], colsSlot, rowsSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		arrSlot := instruction.Result + ".arr.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", arrSlot)
		statusArr := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_new(i64 2, i64 8, ptr %%%s)\n", statusArr, arrSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", statusArr)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, arrSlot)

		statusSet0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_set(ptr %%%s, double 0.0, ptr %%%s)\n", statusSet0, instruction.Result, colsSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", statusSet0)

		statusSet1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_set(ptr %%%s, double 1.0, ptr %%%s)\n", statusSet1, instruction.Result, rowsSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", statusSet1)
		return nil

	case "__tty.setRawMode":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("tty.setRawMode has invalid signature")
		}
		modeF64 := fmt.Sprintf("%s.mode.f64", instruction.Result)
		fmt.Fprintf(out, "  %%%s = select i1 %%%s, double 1.0, double 0.0\n", modeF64, resolvedArgs[1])
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tty_set_raw_mode(double %%%s, double %%%s, ptr %%%s)\n", status, resolvedArgs[0], modeF64, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
		return nil

	case "__tty.read":
		if len(instruction.Args) < 1 || len(instruction.Args) > 2 || instruction.Type != ir.TypeString {
			return fmt.Errorf("tty.read has invalid signature")
		}
		maxLenArg := "4096.0"
		if len(instruction.Args) == 2 {
			maxLenArg = fmt.Sprintf("%%%s", resolvedArgs[1])
		}
		dataSlot := instruction.Result + ".data.slot"
		readSlot := instruction.Result + ".read.slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", dataSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", readSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tty_read(double %%%s, double %s, ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], maxLenArg, dataSlot, readSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, dataSlot)
		return nil

	case "__tty.readLine":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("tty.readLine has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tty_read_line(double %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__tty.write":
		if len(instruction.Args) < 2 || len(instruction.Args) > 3 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("tty.write has invalid signature")
		}
		lenArg := "0.0"
		if len(instruction.Args) == 3 {
			lenArg = fmt.Sprintf("%%%s", resolvedArgs[2])
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_tty_write(double %%%s, ptr %%%s, double %s, ptr %%%s)\n", status, resolvedArgs[0], resolvedArgs[1], lenArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	default:
		return fmt.Errorf("unknown tty intrinsic %q", instruction.Callee)
	}
}
