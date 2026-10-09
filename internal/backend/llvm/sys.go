package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

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
	case "__process.set_env":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("process.set_env has invalid signature")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_process_set_env(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__process.env_obj":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 0, ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__process.pid":
		if len(instruction.Args) != 0 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("process.pid has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_process_pid(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__process.ppid":
		if len(instruction.Args) != 0 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("process.ppid has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_process_ppid(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__process.version":
		if len(instruction.Args) != 0 || instruction.Type != ir.TypeString {
			return fmt.Errorf("process.version has invalid signature")
		}
		verStr := e.compilerVersion
		if verStr == "" || verStr == "dev" {
			verStr = "v0.1.0"
		}
		if !strings.HasPrefix(verStr, "v") {
			verStr = "v" + verStr
		}
		global := e.stringsByValue[verStr]
		length := len([]byte(verStr)) + 1
		fmt.Fprintf(out, "  %%%s = getelementptr inbounds [%d x i8], ptr %s, i64 0, i64 0\n", instruction.Result, length, global)
		return nil
	default:
		return fmt.Errorf("unknown process intrinsic %q", instruction.Callee)
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
	case "__date.UTC":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		// Needs 7 args (year, month, date, hours, min, sec, ms)
		args := make([]string, 7)
		for i := 0; i < 7; i++ {
			if i < len(instruction.Args) {
				args[i] = fmt.Sprintf("double %%%s", instruction.Args[i])
			} else if i == 2 { // date defaults to 1
				args[i] = "double 1.0"
			} else {
				args[i] = "double 0.0"
			}
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_date_utc(%s, ptr %%%s)\n", status, strings.Join(args, ", "), slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__date.toISOString", "__date.toJSON":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_date_to_iso_string(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__date.toString", "__date.toLocaleString", "__date.toLocaleDateString", "__date.toLocaleTimeString":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_date_to_string(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__date.toDateString":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_date_to_date_string(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__date.toTimeString":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_date_to_time_string(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__date.toUTCString":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_date_to_utc_string(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	default:
		// Map getters & setters
		getterMap := map[string]string{
			"__date.getDate":            "scriptgo_date_get_date",
			"__date.getDay":             "scriptgo_date_get_day",
			"__date.getFullYear":        "scriptgo_date_get_full_year",
			"__date.getHours":           "scriptgo_date_get_hours",
			"__date.getMilliseconds":    "scriptgo_date_get_milliseconds",
			"__date.getMinutes":         "scriptgo_date_get_minutes",
			"__date.getMonth":           "scriptgo_date_get_month",
			"__date.getSeconds":         "scriptgo_date_get_seconds",
			"__date.getTimezoneOffset":  "scriptgo_date_get_timezone_offset",
			"__date.getUTCDate":         "scriptgo_date_get_utc_date",
			"__date.getUTCDay":          "scriptgo_date_get_utc_day",
			"__date.getUTCFullYear":     "scriptgo_date_get_utc_full_year",
			"__date.getUTCHours":        "scriptgo_date_get_utc_hours",
			"__date.getUTCMilliseconds": "scriptgo_date_get_utc_milliseconds",
			"__date.getUTCMinutes":      "scriptgo_date_get_utc_minutes",
			"__date.getUTCMonth":        "scriptgo_date_get_utc_month",
			"__date.getUTCSeconds":      "scriptgo_date_get_utc_seconds",
		}
		if cFn, ok := getterMap[instruction.Callee]; ok {
			slot := instruction.Result + ".slot"
			status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
			fmt.Fprintf(out, "  %%%s = call i32 @%s(double %%%s, ptr %%%s)\n", status, cFn, instruction.Args[0], slot)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
			fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
			return nil
		}

		setterMap := map[string]string{
			"__date.setDate":            "scriptgo_date_set_date",
			"__date.setFullYear":        "scriptgo_date_set_full_year",
			"__date.setHours":           "scriptgo_date_set_hours",
			"__date.setMilliseconds":    "scriptgo_date_set_milliseconds",
			"__date.setMinutes":         "scriptgo_date_set_minutes",
			"__date.setMonth":           "scriptgo_date_set_month",
			"__date.setSeconds":         "scriptgo_date_set_seconds",
			"__date.setUTCDate":         "scriptgo_date_set_utc_date",
			"__date.setUTCFullYear":     "scriptgo_date_set_utc_full_year",
			"__date.setUTCHours":        "scriptgo_date_set_utc_hours",
			"__date.setUTCMilliseconds": "scriptgo_date_set_utc_milliseconds",
			"__date.setUTCMinutes":      "scriptgo_date_set_utc_minutes",
			"__date.setUTCMonth":        "scriptgo_date_set_utc_month",
			"__date.setUTCSeconds":      "scriptgo_date_set_utc_seconds",
		}
		if cFn, ok := setterMap[instruction.Callee]; ok {
			slot := instruction.Result + ".slot"
			status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
			fmt.Fprintf(out, "  %%%s = call i32 @%s(double %%%s, double %%%s, ptr %%%s)\n", status, cFn, instruction.Args[0], instruction.Args[1], slot)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
			fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
			return nil
		}

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
	case "__os.platform", "__os.arch", "__os.homedir", "__os.type", "__os.release", "__os.tmpdir",
		"__os.hostname", "__os.loadavg", "__os.cpus", "__os.network_interfaces", "__os.user_info",
		"__os.machine", "__os.version":
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
	case "__os.uptime", "__os.totalmem", "__os.freemem", "__os.available_parallelism":
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
	case "__os.get_priority":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("%s has invalid signature", instruction.Callee)
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_os_get_priority(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__os.set_priority":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeVoid {
			return fmt.Errorf("%s has invalid signature", instruction.Callee)
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_os_set_priority(double %%%s, double %%%s)\n", status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
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

func (e *functionEmitter) emitErrorIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	resolvedArgs := make([]string, len(instruction.Args))
	for i, arg := range instruction.Args {
		resolvedArgs[i] = e.resolveArg(out, arg)
	}
	switch instruction.Callee {
	case "__error.captureStack":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("__error.captureStack requires 2 arguments")
		}
		slot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_error_capture_stack(ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__error.stackTraceLimit":
		fmt.Fprintf(out, "  %%%s = call double @scriptgo_error_stack_trace_limit()\n", instruction.Result)
		return nil
	case "__error.setStackTraceLimit":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("__error.setStackTraceLimit requires 1 argument")
		}
		fmt.Fprintf(out, "  call void @scriptgo_error_set_stack_trace_limit(double %%%s)\n", resolvedArgs[0])
		return nil
	case "__error.throw":
		// Throws "<Name>: <message>" as an instance of the named built-in
		// error through the runtime unwinder, like any runtime failure.
		if len(instruction.Args) != 1 {
			return fmt.Errorf("__error.throw requires 1 argument")
		}
		fmt.Fprintf(out, "  call void @scriptgo_throw_error_message(ptr %%%s)\n", resolvedArgs[0])
		return nil
	default:
		return fmt.Errorf("unknown error intrinsic %q", instruction.Callee)
	}
}
