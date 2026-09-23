package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitBufferRead(out *strings.Builder, instruction ir.Instruction, cFn string, isLE int) error {
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
	if isLE >= 0 {
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, double %%%s, i32 %d, ptr %%%s)\n",
			status, cFn, instruction.Args[0], instruction.Args[1], isLE, slot)
	} else {
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, double %%%s, ptr %%%s)\n",
			status, cFn, instruction.Args[0], instruction.Args[1], slot)
	}
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	return nil
}

func (e *functionEmitter) emitBufferWrite(out *strings.Builder, instruction ir.Instruction, cFn string, isLE int) error {
	// Fast path for writeUInt32LE on native little-endian architecture (ARM64 / x86_64)
	if cFn == "scriptgo_buffer_write_u32" && isLE == 1 {
		buf := instruction.Args[0]
		val := instruction.Args[1]
		off := instruction.Args[2]
		resultUsed := e.usedResults[instruction.Result]
		id := e.labelCounter
		e.labelCounter++

		checkLbl := fmt.Sprintf("bw.check.%d", id)
		fastLbl := fmt.Sprintf("bw.fast.%d", id)
		slowLbl := fmt.Sprintf("bw.slow.%d", id)
		doneLbl := fmt.Sprintf("bw.done.%d", id)

		// 1. Pointer null/undefined check
		fmt.Fprintf(out, "  %%bw.not_null.%d = icmp ne ptr %%%s, null\n", id, buf)
		fmt.Fprintf(out, "  %%bw.not_undef.%d = icmp ne ptr %%%s, @scriptgo_undefined_sentinel\n", id, buf)
		fmt.Fprintf(out, "  %%bw.valid_ptr.%d = and i1 %%bw.not_null.%d, %%bw.not_undef.%d\n", id, id, id)
		fmt.Fprintf(out, "  br i1 %%bw.valid_ptr.%d, label %%%s, label %%%s, !prof !{!\"branch_weights\", i32 10000, i32 1}\n\n", id, checkLbl, slowLbl)

		// 2. Check buffer magic, data pointer, and bounds
		fmt.Fprintf(out, "%s:\n", checkLbl)
		fmt.Fprintf(out, "  %%bw.magic.%d = load i32, ptr %%%s, align 4, !invariant.load !{}\n", id, buf)
		fmt.Fprintf(out, "  %%bw.is_buf.%d = icmp eq i32 %%bw.magic.%d, 1112884806\n", id, id) // SCRIPTGO_MAGIC_BUFFER 0x42554646
		fmt.Fprintf(out, "  %%bw.data_ptr.%d = getelementptr inbounds i8, ptr %%%s, i64 40\n", id, buf)
		fmt.Fprintf(out, "  %%bw.data.%d = load ptr, ptr %%bw.data_ptr.%d, align 8, !invariant.load !{}\n", id, id)
		fmt.Fprintf(out, "  %%bw.data_not_null.%d = icmp ne ptr %%bw.data.%d, null\n", id, id)
		fmt.Fprintf(out, "  %%bw.off_i64.%d = fptosi double %%%s to i64\n", id, off)
		fmt.Fprintf(out, "  %%bw.len_ptr.%d = getelementptr inbounds i8, ptr %%%s, i64 8\n", id, buf)
		fmt.Fprintf(out, "  %%bw.len.%d = load i64, ptr %%bw.len_ptr.%d, align 8, !invariant.load !{}\n", id, id)
		fmt.Fprintf(out, "  %%bw.max_off.%d = sub nsw i64 %%bw.len.%d, 4\n", id, id)
		fmt.Fprintf(out, "  %%bw.bounds_ok.%d = icmp ule i64 %%bw.off_i64.%d, %%bw.max_off.%d\n", id, id, id)
		fmt.Fprintf(out, "  %%bw.fast_cond0.%d = and i1 %%bw.is_buf.%d, %%bw.data_not_null.%d\n", id, id, id)
		fmt.Fprintf(out, "  %%bw.fast_cond.%d = and i1 %%bw.fast_cond0.%d, %%bw.bounds_ok.%d\n", id, id, id)
		fmt.Fprintf(out, "  br i1 %%bw.fast_cond.%d, label %%%s, label %%%s, !prof !{!\"branch_weights\", i32 10000, i32 1}\n\n", id, fastLbl, slowLbl)

		// 3. Fast path: store directly to buffer data pointer
		fmt.Fprintf(out, "%s:\n", fastLbl)
		fmt.Fprintf(out, "  %%bw.dest.%d = getelementptr inbounds i8, ptr %%bw.data.%d, i64 %%bw.off_i64.%d\n", id, id, id)
		fmt.Fprintf(out, "  %%bw.val_i32.%d = fptoui double %%%s to i32\n", id, val)
		fmt.Fprintf(out, "  store i32 %%bw.val_i32.%d, ptr %%bw.dest.%d, align 1\n", id, id)
		if resultUsed {
			fmt.Fprintf(out, "  %%bw.end_off.%d = add nsw i64 %%bw.off_i64.%d, 4\n", id, id)
			fmt.Fprintf(out, "  %%bw.fast_res.%d = sitofp i64 %%bw.end_off.%d to double\n", id, id)
		}
		fmt.Fprintf(out, "  br label %%%s\n\n", doneLbl)

		// 4. Slow path: fallback to runtime C helper
		fmt.Fprintf(out, "%s:\n", slowLbl)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		slot := "null"
		if resultUsed {
			slot = "%" + instruction.Result + ".slot"
			fmt.Fprintf(out, "  %%%s = alloca double\n", instruction.Result+".slot")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, double %%%s, double %%%s, i32 %d, ptr %s)\n",
			status, cFn, instruction.Args[0], instruction.Args[1], instruction.Args[2], isLE, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		if resultUsed {
			fmt.Fprintf(out, "  %%bw.slow_res.%d = load double, ptr %s\n", id, slot)
		}
		fmt.Fprintf(out, "  br label %%%s\n\n", doneLbl)

		// 5. Merge the paths; only materialize the JS return value when used.
		fmt.Fprintf(out, "%s:\n", doneLbl)
		if resultUsed {
			fmt.Fprintf(out, "  %%%s = phi double [ %%bw.fast_res.%d, %%%s ], [ %%bw.slow_res.%d, %%%s ]\n",
				instruction.Result, id, fastLbl, id, slowLbl)
		}
		return nil
	}

	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
	if isLE >= 0 {
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, double %%%s, double %%%s, i32 %d, ptr %%%s)\n",
			status, cFn, instruction.Args[0], instruction.Args[1], instruction.Args[2], isLE, slot)
	} else {
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, double %%%s, double %%%s, ptr %%%s)\n",
			status, cFn, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
	}
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	return nil
}

func (e *functionEmitter) emitBufferReadBigInt(out *strings.Builder, instruction ir.Instruction, isLE int) error {
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	fmt.Fprintf(out, "  %%%s = alloca i64\n", slot)
	fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_read_bigint64(ptr %%%s, double %%%s, i32 %d, ptr %%%s)\n",
		status, instruction.Args[0], instruction.Args[1], isLE, slot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load i64, ptr %%%s\n", instruction.Result, slot)
	return nil
}

func (e *functionEmitter) emitBufferWriteBigInt(out *strings.Builder, instruction ir.Instruction, isLE int) error {
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
	fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_write_bigint64(ptr %%%s, i64 %%%s, double %%%s, i32 %d, ptr %%%s)\n",
		status, instruction.Args[0], instruction.Args[1], instruction.Args[2], isLE, slot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	return nil
}

func (e *functionEmitter) emitBufferReadIntN(out *strings.Builder, instruction ir.Instruction, isLE, isSigned int) error {
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
	fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_read_int(ptr %%%s, double %%%s, double %%%s, i32 %d, i32 %d, ptr %%%s)\n",
		status, instruction.Args[0], instruction.Args[1], instruction.Args[2], isLE, isSigned, slot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	return nil
}

func (e *functionEmitter) emitBufferWriteIntN(out *strings.Builder, instruction ir.Instruction, isLE, isSigned int) error {
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
	fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_write_int(ptr %%%s, double %%%s, double %%%s, double %%%s, i32 %d, i32 %d, ptr %%%s)\n",
		status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3], isLE, isSigned, slot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	return nil
}

func (e *functionEmitter) emitBufferSwap(out *strings.Builder, instruction ir.Instruction, width int) error {
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
	fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_swap(ptr %%%s, i32 %d, ptr %%%s)\n",
		status, instruction.Args[0], width, slot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	return nil
}

func (e *functionEmitter) emitBufferLastIndexOf(out *strings.Builder, instruction ir.Instruction) error {
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	valPtr := "null"
	valNum := "0.000000e+00"
	if e.types[instruction.Args[1]] == ir.TypeString {
		valPtr = fmt.Sprintf("%%%s", instruction.Args[1])
	} else {
		valNum = fmt.Sprintf("%%%s", instruction.Args[1])
	}
	isStrI32 := fmt.Sprintf("cast.%s", instruction.Args[2])
	fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", isStrI32, instruction.Args[2])
	hasOffI32 := fmt.Sprintf("cast.%s", instruction.Args[4])
	fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", hasOffI32, instruction.Args[4])
	fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
	fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_last_index_of(ptr %%%s, ptr %s, double %s, i32 %%%s, double %%%s, i32 %%%s, ptr %%%s)\n",
		status, instruction.Args[0], valPtr, valNum, isStrI32, instruction.Args[3], hasOffI32, slot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	return nil
}

func (e *functionEmitter) emitBufferIncludes(out *strings.Builder, instruction ir.Instruction) error {
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	valPtr := "null"
	valNum := "0.000000e+00"
	if e.types[instruction.Args[1]] == ir.TypeString {
		valPtr = fmt.Sprintf("%%%s", instruction.Args[1])
	} else {
		valNum = fmt.Sprintf("%%%s", instruction.Args[1])
	}
	isStrI32 := fmt.Sprintf("cast.%s", instruction.Args[2])
	fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", isStrI32, instruction.Args[2])
	hasOffI32 := fmt.Sprintf("cast.%s", instruction.Args[4])
	fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", hasOffI32, instruction.Args[4])
	fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
	fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_includes(ptr %%%s, ptr %s, double %s, i32 %%%s, double %%%s, i32 %%%s, ptr %%%s)\n",
		status, instruction.Args[0], valPtr, valNum, isStrI32, instruction.Args[3], hasOffI32, slot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	raw := instruction.Result + ".raw"
	fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", raw, slot)
	fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, raw)
	return nil
}

func (e *functionEmitter) emitBufferWriteStr(out *strings.Builder, instruction ir.Instruction) error {
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
	fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_write(ptr %%%s, ptr %%%s, double %%%s, double %%%s, ptr %%%s, ptr %%%s)\n",
		status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3], instruction.Args[4], slot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	return nil
}
