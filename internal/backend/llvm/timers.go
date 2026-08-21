package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitTimerIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__timers.setTimeout":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("setTimeout requires callback")
		}
		delay := "0.0"
		if len(instruction.Args) > 1 {
			delay = "%" + instruction.Args[1]
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_timer_set_timeout(ptr %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], delay, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__timers.setInterval":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("setInterval requires callback")
		}
		delay := "0.0"
		if len(instruction.Args) > 1 {
			delay = "%" + instruction.Args[1]
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_timer_set_interval(ptr %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], delay, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__timers.setImmediate":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("setImmediate requires callback")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_timer_set_immediate(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__timers.clearTimeout":
		if len(instruction.Args) < 1 {
			return nil
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_timer_clear_timeout(double %%%s)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__timers.clearInterval":
		if len(instruction.Args) < 1 {
			return nil
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_timer_clear_interval(double %%%s)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__timers.clearImmediate":
		if len(instruction.Args) < 1 {
			return nil
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_timer_clear_immediate(double %%%s)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	default:
		return fmt.Errorf("unknown timer intrinsic %q", instruction.Callee)
	}
}
