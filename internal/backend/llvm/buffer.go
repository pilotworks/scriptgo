package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitBufferIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__buffer.alloc":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		hasFillI32 := instruction.Args[2] + ".i32"
		isStrFillI32 := instruction.Args[3] + ".i32"
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", hasFillI32, instruction.Args[2])
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", isStrFillI32, instruction.Args[3])
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)

		fillPtr := "null"
		fillNum := "0.0"
		if e.types[instruction.Args[1]] == ir.TypeString {
			fillPtr = "%" + instruction.Args[1]
		} else {
			fillNum = "%" + instruction.Args[1]
		}

		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_alloc(double %%%s, ptr %s, double %s, i32 %%%s, i32 %%%s, ptr %%%s)\n",
			status, instruction.Args[0], fillPtr, fillNum, hasFillI32, isStrFillI32, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__buffer.from_string":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_from_string(ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__buffer.from_array":
		arg := instruction.Args[0]
		if e.types[arg] == ir.TypeUnknown || e.isParamUnknown(arg) {
			unboxed := fmt.Sprintf("unboxed.arr.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", unboxed, arg)
			ptrVal := fmt.Sprintf("ptr.arr.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrVal, unboxed)
			arg = ptrVal
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_from_array(ptr %%%s, ptr %%%s)\n",
			status, arg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__buffer.concat":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_concat(ptr %%%s, double %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__buffer.isBuffer":
		arg := instruction.Args[0]
		if e.types[arg] == ir.TypeUnknown || e.isParamUnknown(arg) {
			unboxed := fmt.Sprintf("unboxed.isbuf.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", unboxed, arg)
			ptrVal := fmt.Sprintf("ptr.isbuf.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrVal, unboxed)
			arg = ptrVal
		}
		slot := instruction.Result + ".i32.slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_is_buffer(ptr %%%s, ptr %%%s)\n",
			status, arg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", instruction.Result+".i32", slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, instruction.Result+".i32")
		return nil

	case "__buffer.byteLength":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_byte_length(ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__buffer.toString":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		hasStartI32 := instruction.Args[4] + ".i32"
		hasEndI32 := instruction.Args[5] + ".i32"
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", hasStartI32, instruction.Args[4])
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", hasEndI32, instruction.Args[5])
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_to_string(ptr %%%s, ptr %%%s, double %%%s, double %%%s, i32 %%%s, i32 %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3], hasStartI32, hasEndI32, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__buffer.copy":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		hasTSI32 := instruction.Args[5] + ".i32"
		hasSSI32 := instruction.Args[6] + ".i32"
		hasSEI32 := instruction.Args[7] + ".i32"
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", hasTSI32, instruction.Args[5])
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", hasSSI32, instruction.Args[6])
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", hasSEI32, instruction.Args[7])
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_copy(ptr %%%s, ptr %%%s, double %%%s, double %%%s, double %%%s, i32 %%%s, i32 %%%s, i32 %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3], instruction.Args[4], hasTSI32, hasSSI32, hasSEI32, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__buffer.fill":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		isStrI32 := instruction.Args[2] + ".i32"
		hasSI32 := instruction.Args[5] + ".i32"
		hasEI32 := instruction.Args[6] + ".i32"
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", isStrI32, instruction.Args[2])
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", hasSI32, instruction.Args[5])
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", hasEI32, instruction.Args[6])
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)

		fillPtr := "null"
		fillNum := "0.0"
		if e.types[instruction.Args[1]] == ir.TypeString {
			fillPtr = "%" + instruction.Args[1]
		} else {
			fillNum = "%" + instruction.Args[1]
		}

		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_fill(ptr %%%s, ptr %s, double %s, i32 %%%s, double %%%s, double %%%s, i32 %%%s, i32 %%%s, ptr %%%s)\n",
			status, instruction.Args[0], fillPtr, fillNum, isStrI32, instruction.Args[3], instruction.Args[4], hasSI32, hasEI32, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__buffer.equals":
		slot := instruction.Result + ".i32.slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_equals(ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", instruction.Result+".i32", slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, instruction.Result+".i32")
		return nil

	case "__buffer.compare":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_compare(ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__buffer.indexOf":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		isStrI32 := instruction.Args[2] + ".i32"
		hasOffI32 := instruction.Args[4] + ".i32"
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", isStrI32, instruction.Args[2])
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", hasOffI32, instruction.Args[4])
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)

		valPtr := "null"
		valNum := "0.0"
		if e.types[instruction.Args[1]] == ir.TypeString {
			valPtr = "%" + instruction.Args[1]
		} else {
			valNum = "%" + instruction.Args[1]
		}

		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_buffer_index_of(ptr %%%s, ptr %s, double %s, i32 %%%s, double %%%s, i32 %%%s, ptr %%%s)\n",
			status, instruction.Args[0], valPtr, valNum, isStrI32, instruction.Args[3], hasOffI32, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	// Binary Read operations
	case "__buffer.readUInt8":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_u8", -1)
	case "__buffer.readInt8":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_i8", -1)
	case "__buffer.readUInt16LE":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_u16", 1)
	case "__buffer.readUInt16BE":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_u16", 0)
	case "__buffer.readInt16LE":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_i16", 1)
	case "__buffer.readInt16BE":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_i16", 0)
	case "__buffer.readUInt32LE":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_u32", 1)
	case "__buffer.readUInt32BE":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_u32", 0)
	case "__buffer.readInt32LE":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_i32", 1)
	case "__buffer.readInt32BE":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_i32", 0)
	case "__buffer.readFloatLE":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_float", 1)
	case "__buffer.readFloatBE":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_float", 0)
	case "__buffer.readDoubleLE":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_double", 1)
	case "__buffer.readDoubleBE":
		return e.emitBufferRead(out, instruction, "scriptgo_buffer_read_double", 0)

	// Binary Write operations
	case "__buffer.writeUInt8":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_u8", -1)
	case "__buffer.writeInt8":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_i8", -1)
	case "__buffer.writeUInt16LE":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_u16", 1)
	case "__buffer.writeUInt16BE":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_u16", 0)
	case "__buffer.writeInt16LE":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_i16", 1)
	case "__buffer.writeInt16BE":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_i16", 0)
	case "__buffer.writeUInt32LE":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_u32", 1)
	case "__buffer.writeUInt32BE":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_u32", 0)
	case "__buffer.writeInt32LE":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_i32", 1)
	case "__buffer.writeInt32BE":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_i32", 0)
	case "__buffer.writeFloatLE":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_float", 1)
	case "__buffer.writeFloatBE":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_float", 0)
	case "__buffer.writeDoubleLE":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_double", 1)
	case "__buffer.writeDoubleBE":
		return e.emitBufferWrite(out, instruction, "scriptgo_buffer_write_double", 0)

	// BigInt Read operations
	case "__buffer.readBigInt64LE", "__buffer.readBigUInt64LE":
		return e.emitBufferReadBigInt(out, instruction, 1)
	case "__buffer.readBigInt64BE", "__buffer.readBigUInt64BE":
		return e.emitBufferReadBigInt(out, instruction, 0)

	// BigInt Write operations
	case "__buffer.writeBigInt64LE", "__buffer.writeBigUInt64LE":
		return e.emitBufferWriteBigInt(out, instruction, 1)
	case "__buffer.writeBigInt64BE", "__buffer.writeBigUInt64BE":
		return e.emitBufferWriteBigInt(out, instruction, 0)

	// IntN Read operations
	case "__buffer.readIntLE":
		return e.emitBufferReadIntN(out, instruction, 1, 1)
	case "__buffer.readIntBE":
		return e.emitBufferReadIntN(out, instruction, 0, 1)
	case "__buffer.readUIntLE":
		return e.emitBufferReadIntN(out, instruction, 1, 0)
	case "__buffer.readUIntBE":
		return e.emitBufferReadIntN(out, instruction, 0, 0)

	// IntN Write operations
	case "__buffer.writeIntLE":
		return e.emitBufferWriteIntN(out, instruction, 1, 1)
	case "__buffer.writeIntBE":
		return e.emitBufferWriteIntN(out, instruction, 0, 1)
	case "__buffer.writeUIntLE":
		return e.emitBufferWriteIntN(out, instruction, 1, 0)
	case "__buffer.writeUIntBE":
		return e.emitBufferWriteIntN(out, instruction, 0, 0)

	// Swap operations
	case "__buffer.swap16":
		return e.emitBufferSwap(out, instruction, 2)
	case "__buffer.swap32":
		return e.emitBufferSwap(out, instruction, 4)
	case "__buffer.swap64":
		return e.emitBufferSwap(out, instruction, 8)

	case "__buffer.lastIndexOf":
		return e.emitBufferLastIndexOf(out, instruction)
	case "__buffer.includes":
		return e.emitBufferIncludes(out, instruction)
	case "__buffer.write":
		return e.emitBufferWriteStr(out, instruction)

	default:
		return fmt.Errorf("unknown buffer intrinsic %q", instruction.Callee)
	}
}

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

