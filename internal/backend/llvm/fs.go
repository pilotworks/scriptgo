package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitFsIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	if handled, err := e.emitFSWatcherIntrinsic(out, instruction); handled {
		return err
	}

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
