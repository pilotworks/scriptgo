package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitFsIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	resolvedArgs := make([]string, len(instruction.Args))
	for i, arg := range instruction.Args {
		resolvedArgs[i] = e.resolveArg(out, arg)
	}

	switch instruction.Callee {
	case "__fs.readFileSync":
		if len(instruction.Args) < 1 || len(instruction.Args) > 2 || (instruction.Type != ir.TypeString && instruction.Type != ir.TypeBuffer) {
			return fmt.Errorf("fs.readFileSync has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		if instruction.Type == ir.TypeBuffer {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_read_file_buffer_sync(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		} else {
			encoding := "null"
			if len(resolvedArgs) == 2 {
				encoding = fmt.Sprintf("%%%s", resolvedArgs[1])
			}
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_read_file_sync(ptr %%%s, ptr %s, ptr %%%s)\n", status, resolvedArgs[0], encoding, slot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__fs.writeFileSync":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeVoid {
			return fmt.Errorf("fs.writeFileSync has invalid signature")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_write_file_sync(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], resolvedArgs[1])
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
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_exists_sync(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
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
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_unlink_sync(ptr %%%s)\n", status, resolvedArgs[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.readdirSync":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_readdir_sync(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__fs.copyFileSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_copy_file_sync(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.renameSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_rename_sync(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.appendFileSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_append_file_sync(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.mkdirSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		recArg := "0.0"
		if len(instruction.Args) > 1 {
			recVal := resolvedArgs[1]
			recF64 := recVal + ".f64"
			fmt.Fprintf(out, "  %%%s = uitofp i1 %%%s to double\n", recF64, recVal)
			recArg = fmt.Sprintf("%%%s", recF64)
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_mkdir_sync(ptr %%%s, double %s)\n", status, resolvedArgs[0], recArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.rmSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		recArg := "0.0"
		forceArg := "0.0"
		if len(instruction.Args) > 1 {
			recVal := resolvedArgs[1]
			recF64 := recVal + ".f64"
			fmt.Fprintf(out, "  %%%s = uitofp i1 %%%s to double\n", recF64, recVal)
			recArg = fmt.Sprintf("%%%s", recF64)
		}
		if len(instruction.Args) > 2 {
			forceVal := resolvedArgs[2]
			forceF64 := forceVal + ".f64"
			fmt.Fprintf(out, "  %%%s = uitofp i1 %%%s to double\n", forceF64, forceVal)
			forceArg = fmt.Sprintf("%%%s", forceF64)
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_rm_sync(ptr %%%s, double %s, double %s)\n", status, resolvedArgs[0], recArg, forceArg)
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
			status, resolvedArgs[0], sizeSlot, mtimeSlot, birthtimeSlot, modeSlot)
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
	case "__fs.accessSync":
		if len(instruction.Args) < 1 || len(instruction.Args) > 2 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("fs.accessSync has invalid signature")
		}
		modeArg := "0.0"
		if len(instruction.Args) == 2 {
			modeArg = fmt.Sprintf("%%%s", resolvedArgs[1])
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_access_sync(ptr %%%s, double %s, ptr %%%s)\n", status, resolvedArgs[0], modeArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
		return nil
	case "__fs.chmodSync":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeVoid {
			return fmt.Errorf("fs.chmodSync has invalid signature")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_chmod_sync(ptr %%%s, double %%%s)\n", status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.realpathSync":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("fs.realpathSync has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_realpath_sync(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__fs.truncateSync":
		if len(instruction.Args) < 1 || len(instruction.Args) > 2 || instruction.Type != ir.TypeVoid {
			return fmt.Errorf("fs.truncateSync has invalid signature")
		}
		lenArg := "0.0"
		if len(instruction.Args) == 2 {
			lenArg = fmt.Sprintf("%%%s", resolvedArgs[1])
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_truncate_sync(ptr %%%s, double %s)\n", status, resolvedArgs[0], lenArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.mkdtempSync":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("fs.mkdtempSync has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_mkdtemp_sync(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__fs.openSync":
		if len(instruction.Args) < 1 || len(instruction.Args) > 3 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("fs.openSync has invalid signature")
		}
		flagArg := "null"
		modeArg := "0.0"
		if len(instruction.Args) >= 2 {
			flagArg = fmt.Sprintf("%%%s", resolvedArgs[1])
		}
		if len(instruction.Args) >= 3 {
			modeArg = fmt.Sprintf("%%%s", resolvedArgs[2])
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_open_sync(ptr %%%s, ptr %s, double %s, ptr %%%s)\n", status, resolvedArgs[0], flagArg, modeArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__fs.closeSync":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeVoid {
			return fmt.Errorf("fs.closeSync has invalid signature")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_close_sync(double %%%s)\n", status, resolvedArgs[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.readFdSync":
		if len(instruction.Args) < 5 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("fs.readFdSync has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_read_fd_sync(double %%%s, ptr %%%s, double %%%s, double %%%s, double %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2], resolvedArgs[3], resolvedArgs[4], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__fs.writeFdSync":
		if len(instruction.Args) < 4 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("fs.writeFdSync has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_write_fd_sync(double %%%s, ptr %%%s, double %%%s, double %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2], resolvedArgs[3], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__fs.opendirSync":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		dummySlot := instruction.Result + ".types.slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", dummySlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_opendir_sync(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot, dummySlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__fs.fstatSync":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("fs.fstatSync requires 1 arg")
		}
		sizeSlot := instruction.Result + ".size.slot"
		mtimeSlot := instruction.Result + ".mtime.slot"
		birthtimeSlot := instruction.Result + ".birthtime.slot"
		modeSlot := instruction.Result + ".mode.slot"
		uidSlot := instruction.Result + ".uid.slot"
		gidSlot := instruction.Result + ".gid.slot"
		inoSlot := instruction.Result + ".ino.slot"
		devSlot := instruction.Result + ".dev.slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", sizeSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", mtimeSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", birthtimeSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", modeSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", uidSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", gidSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", inoSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", devSlot)

		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_fstat_sync(double %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], sizeSlot, mtimeSlot, birthtimeSlot, modeSlot, uidSlot, gidSlot, inoSlot, devSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		sizeVal := instruction.Result + ".size"
		mtimeVal := instruction.Result + ".mtime"
		birthtimeVal := instruction.Result + ".birthtime"
		modeVal := instruction.Result + ".mode"
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", sizeVal, sizeSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", mtimeVal, mtimeSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", birthtimeVal, birthtimeSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", modeVal, modeSlot)

		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 4, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)

		f0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		f1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		f2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		f3 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 0, double %%%s)\n", f0, instruction.Result, sizeVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", f0)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 1, double %%%s)\n", f1, instruction.Result, mtimeVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", f1)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 2, double %%%s)\n", f2, instruction.Result, birthtimeVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", f2)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 3, double %%%s)\n", f3, instruction.Result, modeVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", f3)
		return nil
	case "__fs.statfsSync":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("fs.statfsSync requires 1 arg")
		}
		bsizeSlot := instruction.Result + ".bsize.slot"
		blocksSlot := instruction.Result + ".blocks.slot"
		bfreeSlot := instruction.Result + ".bfree.slot"
		bavailSlot := instruction.Result + ".bavail.slot"
		filesSlot := instruction.Result + ".files.slot"
		ffreeSlot := instruction.Result + ".ffree.slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", bsizeSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", blocksSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", bfreeSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", bavailSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", filesSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", ffreeSlot)

		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_statfs_sync(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, resolvedArgs[0], bsizeSlot, blocksSlot, bfreeSlot, bavailSlot, filesSlot, ffreeSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		bsizeVal := instruction.Result + ".bsize"
		blocksVal := instruction.Result + ".blocks"
		bfreeVal := instruction.Result + ".bfree"
		bavailVal := instruction.Result + ".bavail"
		filesVal := instruction.Result + ".files"
		ffreeVal := instruction.Result + ".ffree"
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", bsizeVal, bsizeSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", blocksVal, blocksSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", bfreeVal, bfreeSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", bavailVal, bavailSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", filesVal, filesSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", ffreeVal, ffreeSlot)

		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 6, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)

		for idx, v := range []string{bsizeVal, blocksVal, bfreeVal, bavailVal, filesVal, ffreeVal} {
			st := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 %d, double %%%s)\n", st, instruction.Result, idx, v)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st)
		}
		return nil
	case "__fs.chownSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_chown_sync(ptr %%%s, double %%%s, double %%%s)\n", status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.lchownSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_lchown_sync(ptr %%%s, double %%%s, double %%%s)\n", status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.fchownSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_fchown_sync(double %%%s, double %%%s, double %%%s)\n", status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.fchmodSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_fchmod_sync(double %%%s, double %%%s)\n", status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.linkSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_link_sync(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.symlinkSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_symlink_sync(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.readlinkSync":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_readlink_sync(ptr %%%s, ptr %%%s)\n", status, resolvedArgs[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__fs.utimesSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_utimes_sync(ptr %%%s, double %%%s, double %%%s)\n", status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.lutimesSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_lutimes_sync(ptr %%%s, double %%%s, double %%%s)\n", status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.futimesSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_futimes_sync(double %%%s, double %%%s, double %%%s)\n", status, resolvedArgs[0], resolvedArgs[1], resolvedArgs[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.fsyncSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_fsync_sync(double %%%s)\n", status, resolvedArgs[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.fdatasyncSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_fdatasync_sync(double %%%s)\n", status, resolvedArgs[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.ftruncateSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_ftruncate_sync(double %%%s, double %%%s)\n", status, resolvedArgs[0], resolvedArgs[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__fs.rmdirSync":
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_fs_rmdir_sync(ptr %%%s)\n", status, resolvedArgs[0])
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
	case "__crypto.hashDigestBuffer":
		if len(instruction.Args) < 2 || len(instruction.Args) > 3 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.hashDigestBuffer has invalid signature")
		}
		encodingArg := "null"
		if len(instruction.Args) == 3 {
			encodingArg = fmt.Sprintf("%%%s", instruction.Args[2])
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_hash_digest_buffer(ptr %%%s, ptr %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], encodingArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__crypto.randomBytes":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.randomBytes has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_random_bytes(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__crypto.randomInt":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("crypto.randomInt has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_random_int(double %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__crypto.randomFill":
		if len(instruction.Args) < 1 || len(instruction.Args) > 3 {
			return fmt.Errorf("crypto.randomFill has invalid signature")
		}
		offArg := "0.0"
		szArg := "0.0"
		if len(instruction.Args) >= 2 {
			offArg = fmt.Sprintf("%%%s", instruction.Args[1])
		}
		if len(instruction.Args) >= 3 {
			szArg = fmt.Sprintf("%%%s", instruction.Args[2])
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_random_fill(ptr %%%s, double %s, double %s)\n", status, instruction.Args[0], offArg, szArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		if instruction.Type == ir.TypeBuffer {
			fmt.Fprintf(out, "  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, instruction.Args[0])
		}
		return nil
	case "__crypto.timingSafeEqual":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("crypto.timingSafeEqual has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_timing_safe_equal(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
		return nil
	case "__crypto.hmacDigest":
		if len(instruction.Args) < 3 || len(instruction.Args) > 4 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.hmacDigest has invalid signature")
		}
		encodingArg := "null"
		if len(instruction.Args) == 4 {
			encodingArg = fmt.Sprintf("%%%s", instruction.Args[3])
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_hmac_digest(ptr %%%s, ptr %%%s, ptr %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], encodingArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__crypto.hmacDigestBuffer":
		if len(instruction.Args) < 3 || len(instruction.Args) > 4 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.hmacDigestBuffer has invalid signature")
		}
		encodingArg := "null"
		if len(instruction.Args) == 4 {
			encodingArg = fmt.Sprintf("%%%s", instruction.Args[3])
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_hmac_digest_buffer(ptr %%%s, ptr %%%s, ptr %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], encodingArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__crypto.pbkdf2Sync":
		if len(instruction.Args) < 4 || len(instruction.Args) > 5 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.pbkdf2Sync has invalid signature")
		}
		digestArg := "null"
		if len(instruction.Args) == 5 {
			digestArg = fmt.Sprintf("%%%s", instruction.Args[4])
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_pbkdf2_sync(ptr %%%s, ptr %%%s, double %%%s, double %%%s, ptr %s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3], digestArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__crypto.hkdfSync":
		if len(instruction.Args) != 5 || instruction.Type != ir.TypeArrayBuffer {
			return fmt.Errorf("crypto.hkdfSync has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_hkdf_sync(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3], instruction.Args[4], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__crypto.scryptSync":
		if len(instruction.Args) != 3 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.scryptSync has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_scrypt_sync(ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
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
	case "__os.platform", "__os.arch", "__os.homedir", "__os.type", "__os.release", "__os.tmpdir":
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
	argVal := instruction.Args[0]
	if len(instruction.Args) > 0 {
		argVal = e.resolveArg(out, instruction.Args[0])
	}
	argType := e.types[argVal]
	if argType == "" {
		argType = e.types[instruction.Args[0]]
	}
	switch instruction.Callee {
	case "__json.stringify_number":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify number has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_number(double %%%s, ptr %%%s)\n", status, argVal, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.stringify_bool":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify bool has invalid signature")
		}
		boolVal := fmt.Sprintf("%s.bool_i32.%d", instruction.Result, e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", boolVal, argVal)
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
		if argType == ir.TypeUnknown {
			payloadName := fmt.Sprintf("json.str.payload.%d", e.loadCounter)
			ptrName := fmt.Sprintf("json.str.ptr.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadName, argVal)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrName, payloadName)
			argVal = ptrName
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_string(ptr %%%s, ptr %%%s)\n", status, argVal, slot)
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
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_number_array(ptr %%%s, ptr %%%s)\n", status, argVal, slot)
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
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_string_array(ptr %%%s, ptr %%%s)\n", status, argVal, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.stringify_unknown":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify unknown has invalid signature")
		}
		if argType != ir.TypeUnknown {
			boxedVar := fmt.Sprintf("json.box.%d", e.loadCounter)
			e.loadCounter++
			if err := e.emitBoxValue(out, argVal, argType, boxedVar); err != nil {
				return err
			}
			argVal = boxedVar
		}
		tag := fmt.Sprintf("%s.tag", argVal)
		padding := fmt.Sprintf("%s.pad", argVal)
		val := fmt.Sprintf("%s.val", argVal)
		fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag, argVal)
		fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 1\n", padding, argVal)
		fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", val, argVal)
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_unknown(i32 %%%s, i32 %%%s, i64 %%%s, ptr %%%s)\n", status, tag, padding, val, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.parse_unknown":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeUnknown {
			return fmt.Errorf("JSON.parse has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca { i32, i32, i64 }\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_parse_unknown(ptr %%%s, ptr %%%s)\n", status, argVal, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load { i32, i32, i64 }, ptr %%%s\n", instruction.Result, slot)
		return nil
	default:
		return fmt.Errorf("unknown JSON intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitConsoleIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
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
