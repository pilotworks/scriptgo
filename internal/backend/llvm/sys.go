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
	case "__fs.readdirSync":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_readdir_sync(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__fs.copyFileSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_copy_file_sync(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.renameSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_rename_sync(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.appendFileSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_append_file_sync(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.mkdirSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		recArg := "0.0"
		if len(instruction.Args) > 1 {
			recVal := instruction.Args[1]
			recF64 := recVal + ".f64"
			fmt.Fprintf(out, "  %%%s = uitofp i1 %%%s to double\n", recF64, recVal)
			recArg = fmt.Sprintf("%%%s", recF64)
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_mkdir_sync(ptr %%%s, double %s)\n", status, instruction.Args[0], recArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.rmSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		recArg := "0.0"
		forceArg := "0.0"
		if len(instruction.Args) > 1 {
			recVal := instruction.Args[1]
			recF64 := recVal + ".f64"
			fmt.Fprintf(out, "  %%%s = uitofp i1 %%%s to double\n", recF64, recVal)
			recArg = fmt.Sprintf("%%%s", recF64)
		}
		if len(instruction.Args) > 2 {
			forceVal := instruction.Args[2]
			forceF64 := forceVal + ".f64"
			fmt.Fprintf(out, "  %%%s = uitofp i1 %%%s to double\n", forceF64, forceVal)
			forceArg = fmt.Sprintf("%%%s", forceF64)
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_rm_sync(ptr %%%s, double %s, double %s)\n", status, instruction.Args[0], recArg, forceArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.statSync":
		// Allocate Stats object struct with fields size, mtimeMs, birthtimeMs, mode
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		sizeSlot := instruction.Result + ".size.slot"
		mtimeSlot := instruction.Result + ".mtime.slot"
		birthtimeSlot := instruction.Result + ".birthtime.slot"
		modeSlot := instruction.Result + ".mode.slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", sizeSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", mtimeSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", birthtimeSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", modeSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_stat_sync(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], sizeSlot, mtimeSlot, birthtimeSlot, modeSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		sizeVal := instruction.Result + ".size"
		mtimeVal := instruction.Result + ".mtime"
		birthtimeVal := instruction.Result + ".birthtime"
		modeVal := instruction.Result + ".mode"
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", sizeVal, sizeSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", mtimeVal, mtimeSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", birthtimeVal, birthtimeSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", modeVal, modeSlot)

		// Create object:Stats instance with 4 fields
		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 4, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)

		field0Status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		field1Status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		field2Status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		field3Status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 0, double %%%s)\n", field0Status, instruction.Result, sizeVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", field0Status)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 1, double %%%s)\n", field1Status, instruction.Result, mtimeVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", field1Status)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 2, double %%%s)\n", field2Status, instruction.Result, birthtimeVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", field2Status)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 3, double %%%s)\n", field3Status, instruction.Result, modeVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", field3Status)
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
	case "__date.toString", "__date.toTemporalInstant", "__date.toLocaleString", "__date.toLocaleDateString", "__date.toLocaleTimeString":
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

func (e *functionEmitter) emitChildProcessIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__child_process.execSync":
		cwdArg := "null"
		inputArg := "null"
		if len(instruction.Args) > 1 {
			cwdArg = "%" + instruction.Args[1]
		}
		if len(instruction.Args) > 2 {
			inputArg = "%" + instruction.Args[2]
		}
		stdoutSlot := instruction.Result + ".stdout.slot"
		stderrSlot := instruction.Result + ".stderr.slot"
		statusSlot := instruction.Result + ".status.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", stdoutSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", stderrSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", statusSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_child_process_exec_sync(ptr %%%s, ptr %s, ptr %s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], cwdArg, inputArg, stdoutSlot, stderrSlot, statusSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		if instruction.Type == ir.TypeString {
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, stdoutSlot)
		} else {
			stdoutVal := instruction.Result + ".stdout"
			stderrVal := instruction.Result + ".stderr"
			statusVal := instruction.Result + ".status"
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", stdoutVal, stdoutSlot)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", stderrVal, stderrSlot)
			fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", statusVal, statusSlot)

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
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 0, ptr %%%s)\n", s1, instruction.Result, stdoutVal)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s1)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 1, ptr %%%s)\n", s2, instruction.Result, stderrVal)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s2)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 2, double %%%s)\n", s3, instruction.Result, statusVal)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s3)
		}
		return nil
	case "__child_process.spawnSync":
		argsArg := "null"
		cwdArg := "null"
		inputArg := "null"
		if len(instruction.Args) > 1 {
			argsArg = "%" + instruction.Args[1]
		}
		if len(instruction.Args) > 2 {
			cwdArg = "%" + instruction.Args[2]
		}
		if len(instruction.Args) > 3 {
			inputArg = "%" + instruction.Args[3]
		}
		stdoutSlot := instruction.Result + ".stdout.slot"
		stderrSlot := instruction.Result + ".stderr.slot"
		statusSlot := instruction.Result + ".status.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", stdoutSlot)
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", stderrSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", statusSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_child_process_spawn_sync(ptr %%%s, ptr %s, ptr %s, ptr %s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], argsArg, cwdArg, inputArg, stdoutSlot, stderrSlot, statusSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		stdoutVal := instruction.Result + ".stdout"
		stderrVal := instruction.Result + ".stderr"
		statusVal := instruction.Result + ".status"
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", stdoutVal, stdoutSlot)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", stderrVal, stderrSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", statusVal, statusSlot)

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
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 0, ptr %%%s)\n", s1, instruction.Result, stdoutVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s1)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 1, ptr %%%s)\n", s2, instruction.Result, stderrVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s2)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 2, double %%%s)\n", s3, instruction.Result, statusVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", s3)
		return nil
	default:
		return fmt.Errorf("unknown child_process intrinsic %q", instruction.Callee)
	}
}

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



