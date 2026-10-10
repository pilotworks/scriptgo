package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitConsoleIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__console.layout":
		// The console's line layout of a value's single-line inspection.
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("console.layout has invalid signature")
		}
		text := e.ensurePointerArg(out, instruction.Args[0])
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_console_layout(ptr %%%s, ptr %%__slot_ptr)\n", status, text)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
		e.types[instruction.Result] = ir.TypeString
		return nil
	case "__console.inspectUnknown":
		// A dynamically typed console argument: strings print as they are,
		// other values as their laid-out inspection.
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("console.inspectUnknown has invalid signature")
		}
		value := e.resolveArg(out, instruction.Args[0])
		argType := e.types[value]
		if argType == "" {
			argType = e.types[instruction.Args[0]]
		}
		if argType != ir.TypeUnknown && argType != "" {
			boxed := fmt.Sprintf("console.box.%d", e.loadCounter)
			e.loadCounter++
			if err := e.emitBoxValue(out, value, argType, boxed); err != nil {
				return err
			}
			value = boxed
		}
		valuePtr, err := e.emitCanonicalValuePointer(out, value, ir.TypeUnknown, "console.unknown")
		if err != nil {
			return err
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_console_inspect_unknown(ptr %s, ptr %%__slot_ptr)\n", status, valuePtr)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
		e.types[instruction.Result] = ir.TypeString
		return nil
	case "__console.new":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 0, ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__console.clear":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_console_clear()\n", status)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__console.group":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_console_group()\n", status)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__console.groupEnd":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_console_group_end()\n", status)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__console.count":
		arg := "null"
		if len(instruction.Args) > 0 {
			arg = "%" + instruction.Args[0]
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_console_count(ptr %s)\n", status, arg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__console.countReset":
		arg := "null"
		if len(instruction.Args) > 0 {
			arg = "%" + instruction.Args[0]
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_console_count_reset(ptr %s)\n", status, arg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__console.time":
		arg := "null"
		if len(instruction.Args) > 0 {
			arg = "%" + instruction.Args[0]
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_console_time(ptr %s)\n", status, arg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__console.timeLog":
		lblArg := "null"
		dataArg := "null"
		if len(instruction.Args) > 0 {
			lblArg = "%" + instruction.Args[0]
		}
		if len(instruction.Args) > 1 {
			dataArg = "%" + instruction.Args[1]
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_console_time_log(ptr %s, ptr %s)\n", status, lblArg, dataArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__console.timeEnd":
		arg := "null"
		if len(instruction.Args) > 0 {
			arg = "%" + instruction.Args[0]
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_console_time_end(ptr %s)\n", status, arg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__console.trace":
		arg := "null"
		if len(instruction.Args) > 0 {
			arg = "%" + instruction.Args[0]
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_console_trace(ptr %s)\n", status, arg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	default:
		return fmt.Errorf("unknown console intrinsic %q", instruction.Callee)
	}
}
