package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitFsIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__fs.readFileSync":
		if len(instruction.Args) < 1 || len(instruction.Args) > 2 || instruction.Type != ir.TypeString {
			return fmt.Errorf("fs.readFileSync has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_read_file_sync(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__fs.writeFileSync":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeVoid {
			return fmt.Errorf("fs.writeFileSync has invalid signature")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_write_file_sync(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.existsSync":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("fs.existsSync has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_exists_sync(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
		return nil
	case "__fs.unlinkSync":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeVoid {
			return fmt.Errorf("fs.unlinkSync has invalid signature")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_unlink_sync(ptr %%%s)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	default:
		return fmt.Errorf("unknown fs intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitProcessIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__process.exit":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeVoid {
			return fmt.Errorf("process.exit has invalid signature")
		}
		fmt.Fprintf(out, "  call i32 @scriptgo_process_exit(double %%%s)\n", instruction.Args[0])
		return nil
	case "__process.cwd":
		if len(instruction.Args) != 0 || instruction.Type != ir.TypeString {
			return fmt.Errorf("process.cwd has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_process_cwd(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__process.argv":
		if len(instruction.Args) != 0 || instruction.Type != ir.TypeStringArray {
			return fmt.Errorf("process.argv has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_process_argv(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__process.env":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("process.env has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_process_env(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	default:
		return fmt.Errorf("unknown process intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitCryptoIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__crypto.randomUUID":
		if len(instruction.Args) != 0 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.randomUUID has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_random_uuid(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__crypto.hashDigest":
		if len(instruction.Args) < 2 || len(instruction.Args) > 3 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.hashDigest has invalid signature")
		}
		encodingArg := "null"
		if len(instruction.Args) == 3 {
			encodingArg = fmt.Sprintf("%%%s", instruction.Args[2])
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_hash_digest(ptr %%%s, ptr %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], encodingArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	default:
		return fmt.Errorf("unknown crypto intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitDateIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__date.now":
		if len(instruction.Args) != 0 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("date.now has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_date_now(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__date.toISOString":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("date.toISOString has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_date_to_iso_string(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__date.toString", "__date.toUTCString":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("date.toString has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_date_to_string(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__date.parse":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("date.parse has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_date_parse(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil
	default:
		return fmt.Errorf("unknown date intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitWebIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__web.btoa":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("btoa has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_web_btoa(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__web.atob":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("atob has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_web_atob(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	default:
		return fmt.Errorf("unknown web intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitOsIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__os.platform", "__os.arch", "__os.homedir", "__os.type", "__os.release":
		if len(instruction.Args) != 0 || instruction.Type != ir.TypeString {
			return fmt.Errorf("%s has invalid signature", instruction.Callee)
		}
		cFn := "scriptgo_os_" + strings.TrimPrefix(instruction.Callee, "__os.")
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s)\n", status, cFn, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__os.uptime", "__os.totalmem", "__os.freemem":
		if len(instruction.Args) != 0 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("%s has invalid signature", instruction.Callee)
		}
		cFn := "scriptgo_os_" + strings.TrimPrefix(instruction.Callee, "__os.")
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s)\n", status, cFn, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil
	default:
		return fmt.Errorf("unknown os intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitPerformanceIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__performance.now":
		if len(instruction.Args) != 0 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("performance.now has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_performance_now(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil
	default:
		return fmt.Errorf("unknown performance intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitJsonIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__json.stringify_number":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify number has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_number(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.stringify_bool":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify bool has invalid signature")
		}
		boolVal := fmt.Sprintf("%s.i32", instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", boolVal, instruction.Args[0])
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_bool(i32 %%%s, ptr %%%s)\n", status, boolVal, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.stringify_string":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify string has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_string(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.stringify_number_array":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify number array has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_number_array(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.stringify_string_array":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify string array has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_string_array(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.parse_string":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.parse has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_parse_string(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	default:
		return fmt.Errorf("unknown JSON intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitConsoleIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
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



