package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitIndex(out *strings.Builder, instruction ir.Instruction) error {
	if len(instruction.Args) != 2 {
		return fmt.Errorf("index instruction requires array and index operands")
	}
	arrayType := e.types[instruction.Args[0]]
	if arrayType == ir.TypeBigInt64Array || arrayType == ir.TypeBigUint64Array {
		e.types[instruction.Result] = ir.TypeBigInt
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = alloca i64\n", slot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_typedarray_get_bigint(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s\n", instruction.Result, slot))
		return nil
	}
	arrArg := e.ensurePointerArg(out, instruction.Args[0])
	idxArg := e.resolveArg(out, instruction.Args[1])
	if isTypedArrayType(arrayType) {
		if arrayType == ir.TypeFloat64Array {
			e.types[instruction.Result] = ir.TypeNumber
			slot, hasSlot := e.varSlots[instruction.Result]
			slowSlot := "%__slot_double"
			if hasSlot {
				slowSlot = "%" + slot
			} else {
				e.localSSAs[instruction.Result] = true
			}

			id := e.labelCounter
			e.labelCounter++

			checkLabel := fmt.Sprintf("taget.check.%d", id)
			fastLabel := fmt.Sprintf("taget.fast.%d", id)
			slowLabel := fmt.Sprintf("taget.slow.%d", id)
			doneLabel := fmt.Sprintf("taget.done.%d", id)

			isNotNull := fmt.Sprintf("taget.is_not_null.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = icmp ne ptr %%%s, null\n", isNotNull, arrArg))
			out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\"branch_weights\", i32 10000, i32 1}\n", isNotNull, checkLabel, slowLabel))

			out.WriteString(fmt.Sprintf("\n%s:\n", checkLabel))
			idxI64 := fmt.Sprintf("taget.i64.%d", id)
			idxRoundtrip := fmt.Sprintf("taget.roundtrip.%d", id)
			isInt := fmt.Sprintf("taget.is_int.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i64\n", idxI64, idxArg))
			out.WriteString(fmt.Sprintf("  %%%s = sitofp i64 %%%s to double\n", idxRoundtrip, idxI64))
			out.WriteString(fmt.Sprintf("  %%%s = fcmp oeq double %%%s, %%%s\n", isInt, idxArg, idxRoundtrip))

			arrLenPtr := fmt.Sprintf("taget.len.ptr.%d", id)
			arrLen := fmt.Sprintf("taget.len.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %%%s, i64 8\n", arrLenPtr, arrArg))
			out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s, !invariant.load !{}\n", arrLen, arrLenPtr))

			inBounds := fmt.Sprintf("taget.in_bounds.%d", id)
			condFast := fmt.Sprintf("taget.cond_fast.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = icmp ult i64 %%%s, %%%s\n", inBounds, idxI64, arrLen))
			out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", condFast, isInt, inBounds))
			out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\"branch_weights\", i32 10000, i32 1}\n", condFast, fastLabel, slowLabel))

			out.WriteString(fmt.Sprintf("\n%s:\n", fastLabel))
			dataPtrPtr := fmt.Sprintf("taget.data.ptr.%d", id)
			dataPtr := fmt.Sprintf("taget.data.%d", id)
			elemPtr := fmt.Sprintf("taget.elem.ptr.%d", id)
			fastVal := fmt.Sprintf("taget.fast.val.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %%%s, i64 40\n", dataPtrPtr, arrArg))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s, !invariant.load !{}\n", dataPtr, dataPtrPtr))
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds double, ptr %%%s, i64 %%%s\n", elemPtr, dataPtr, idxI64))
			out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", fastVal, elemPtr))
			out.WriteString(fmt.Sprintf("  br label %%%s\n", doneLabel))

			out.WriteString(fmt.Sprintf("\n%s:\n", slowLabel))
			status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_typedarray_get(ptr %%%s, double %%%s, ptr %s)\n", status, arrArg, idxArg, slowSlot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			slowVal := fmt.Sprintf("taget.slow.val.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %s\n", slowVal, slowSlot))
			out.WriteString(fmt.Sprintf("  br label %%%s\n", doneLabel))

			out.WriteString(fmt.Sprintf("\n%s:\n", doneLabel))
			out.WriteString(fmt.Sprintf("  %%%s = phi double [ %%%s, %%%s ], [ %%%s, %%%s ]\n", instruction.Result, fastVal, fastLabel, slowVal, slowLabel))
			if hasSlot {
				out.WriteString(fmt.Sprintf("  store%s double %%%s, ptr %%%s\n", e.vol(), instruction.Result, slot))
			}
			return nil
		}
		e.types[instruction.Result] = ir.TypeNumber
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_typedarray_get(ptr %%%s, double %%%s, ptr %%%s)\n", status, arrArg, idxArg, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	}
	e.types[instruction.Result] = instruction.Type
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	if arrayType == ir.TypeUnknownArray && instruction.Type != ir.TypeUnknown {
		unknownResult := instruction.Result + ".unknown"
		slot := unknownResult + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca { i32, i32, i64, i64 }\n", slot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_get_unknown(ptr %%%s, double %%%s, ptr %%%s)\n", status, arrArg, idxArg, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", unknownResult, slot))
		e.types[unknownResult] = ir.TypeUnknown
		cast := instruction
		cast.Args = []string{unknownResult}
		return e.emitCheckedCast(out, cast)
	}
	if instruction.Type == ir.TypeUnknown {
		elemType := arrayElementType(arrayType)
		if elemType != "" && elemType != ir.TypeUnknown {
			rawSlot := fmt.Sprintf("%s.raw.slot", instruction.Result)
			rawVal := fmt.Sprintf("%s.raw", instruction.Result)
			llvmT := llvmType(elemType)
			out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", rawSlot, llvmT))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_get(ptr %%%s, double %%%s, ptr %%%s)\n", status, arrArg, idxArg, rawSlot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %%%s\n", rawVal, llvmT, rawSlot))
			return e.emitBoxValue(out, rawVal, elemType, instruction.Result)
		}
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca { i32, i32, i64, i64 }\n", slot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_get_unknown(ptr %%%s, double %%%s, ptr %%%s)\n", status, arrArg, idxArg, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", instruction.Result, slot))
		return nil
	}
	llvmT := llvmType(instruction.Type)
	if llvmT == "void" || llvmT == "" {
		llvmT = "{ i32, i32, i64, i64 }"
	}
	slot, hasSlot := e.varSlots[instruction.Result]
	slowSlot := slot
	if !hasSlot {
		switch llvmT {
		case "double":
			slowSlot = "%__slot_double"
		case "ptr":
			slowSlot = "%__slot_ptr"
		case "i64":
			slowSlot = "%__slot_i64"
		case "i32":
			slowSlot = "%__slot_i32"
		case "i1":
			slowSlot = "%__slot_i1"
		default:
			slowSlot = "%__slot_ptr"
		}
		e.localSSAs[instruction.Result] = true
	} else {
		slowSlot = "%" + slot
	}

	canFastPath := (instruction.Type == ir.TypeNumber || isPointerType(instruction.Type) || instruction.Type == ir.TypeBool) &&
		arrayType != ir.TypeUnknownArray && arrayType != ""

	if canFastPath {
		expectedSize := int64(8)
		if instruction.Type == ir.TypeBool {
			expectedSize = 1
		} else if isPointerType(instruction.Type) {
			expectedSize = e.pointerSize()
		}

		id := e.labelCounter
		e.labelCounter++

		checkLabel := fmt.Sprintf("idx.check.%d", id)
		fastLabel := fmt.Sprintf("idx.fast.%d", id)
		slowLabel := fmt.Sprintf("idx.slow.%d", id)
		doneLabel := fmt.Sprintf("idx.done.%d", id)

		isNotNull := fmt.Sprintf("idx.is_not_null.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne ptr %%%s, null\n", isNotNull, arrArg))
		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\"branch_weights\", i32 10000, i32 1}\n", isNotNull, checkLabel, slowLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", checkLabel))
		idxI64 := fmt.Sprintf("idx.i64.%d", id)
		idxRoundtrip := fmt.Sprintf("idx.roundtrip.%d", id)
		isInt := fmt.Sprintf("idx.is_int.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i64\n", idxI64, idxArg))
		out.WriteString(fmt.Sprintf("  %%%s = sitofp i64 %%%s to double\n", idxRoundtrip, idxI64))
		out.WriteString(fmt.Sprintf("  %%%s = fcmp oeq double %%%s, %%%s\n", isInt, idxArg, idxRoundtrip))

		arrLen := fmt.Sprintf("idx.len.%d", id)
		invLen := ""
		if !e.hasArrayResize {
			invLen = ", !invariant.load !{}"
		}
		out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s%s\n", arrLen, arrArg, invLen))

		elemSizePtr := fmt.Sprintf("idx.esize.ptr.%d", id)
		elemSize := fmt.Sprintf("idx.esize.%d", id)
		isExpectedSize := fmt.Sprintf("idx.is_esize.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %%%s, i64 16\n", elemSizePtr, arrArg))
		out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s, !invariant.load !{}\n", elemSize, elemSizePtr))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i64 %%%s, %d\n", isExpectedSize, elemSize, expectedSize))

		inBounds := fmt.Sprintf("idx.in_bounds.%d", id)
		cond1 := fmt.Sprintf("idx.cond1.%d", id)
		condFast := fmt.Sprintf("idx.cond_fast.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp ult i64 %%%s, %%%s\n", inBounds, idxI64, arrLen))
		out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", cond1, isInt, inBounds))
		out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", condFast, cond1, isExpectedSize))
		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\"branch_weights\", i32 10000, i32 1}\n", condFast, fastLabel, slowLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", fastLabel))
		dataPtrPtr := fmt.Sprintf("idx.data.ptr.%d", id)
		dataPtr := fmt.Sprintf("idx.data.%d", id)
		elemPtr := fmt.Sprintf("idx.elem.ptr.%d", id)
		fastVal := fmt.Sprintf("idx.fast.val.%d", id)
		invData := ""
		if !e.hasArrayResize {
			invData = ", !invariant.load !{}"
		}
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %%%s, i64 24\n", dataPtrPtr, arrArg))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s%s\n", dataPtr, dataPtrPtr, invData))

		elemPtrType := llvmT
		if instruction.Type == ir.TypeBool {
			elemPtrType = "i8"
		}
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds %s, ptr %%%s, i64 %%%s\n", elemPtr, elemPtrType, dataPtr, idxI64))
		if instruction.Type == ir.TypeBool {
			rawByte := fmt.Sprintf("idx.raw.byte.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = load i8, ptr %%%s\n", rawByte, elemPtr))
			out.WriteString(fmt.Sprintf("  %%%s = icmp ne i8 %%%s, 0\n", fastVal, rawByte))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %%%s\n", fastVal, elemPtrType, elemPtr))
		}
		out.WriteString(fmt.Sprintf("  br label %%%s\n", doneLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", slowLabel))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_get(ptr %%%s, double %%%s, ptr %s)\n", status, arrArg, idxArg, slowSlot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		slowVal := fmt.Sprintf("idx.slow.val.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %s\n", slowVal, llvmT, slowSlot))
		out.WriteString(fmt.Sprintf("  br label %%%s\n", doneLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", doneLabel))
		out.WriteString(fmt.Sprintf("  %%%s = phi %s [ %%%s, %%%s ], [ %%%s, %%%s ]\n", instruction.Result, llvmT, fastVal, fastLabel, slowVal, slowLabel))
		if hasSlot {
			out.WriteString(fmt.Sprintf("  store%s %s %%%s, ptr %%%s\n", e.vol(), llvmT, instruction.Result, slot))
		}
		return nil
	}

	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_get(ptr %%%s, double %%%s, ptr %s)\n", status, arrArg, idxArg, slowSlot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %s\n", instruction.Result, llvmT, slowSlot))
	if hasSlot {
		out.WriteString(fmt.Sprintf("  store%s %s %%%s, ptr %%%s\n", e.vol(), llvmT, instruction.Result, slot))
	}
	return nil
}
