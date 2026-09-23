package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitIndexSet(out *strings.Builder, instruction ir.Instruction) error {
	if len(instruction.Args) != 3 {
		return fmt.Errorf("index.set instruction requires array, index, and value operands")
	}
	arrayType, ok := e.types[instruction.Args[0]]
	if !ok {
		return fmt.Errorf("unknown index.set array %q", instruction.Args[0])
	}
	arrArg := e.ensurePointerArg(out, instruction.Args[0])
	idxArg := e.resolveArg(out, instruction.Args[1])
	valArg := e.resolveArg(out, instruction.Args[2])
	if arrayType == ir.TypeBigInt64Array || arrayType == ir.TypeBigUint64Array {
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_typedarray_set_bigint(ptr %%%s, double %%%s, i64 %%%s)\n", status, arrArg, idxArg, valArg))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	}
	if isTypedArrayType(arrayType) {
		if arrayType == ir.TypeFloat64Array {
			id := e.labelCounter
			e.labelCounter++
			checkLabel := fmt.Sprintf("taset.check.%d", id)
			fastLabel := fmt.Sprintf("taset.fast.%d", id)
			slowLabel := fmt.Sprintf("taset.slow.%d", id)
			doneLabel := fmt.Sprintf("taset.done.%d", id)

			isNotNull := fmt.Sprintf("taset.is_not_null.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = icmp ne ptr %%%s, null\n", isNotNull, arrArg))
			out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\"branch_weights\", i32 10000, i32 1}\n", isNotNull, checkLabel, slowLabel))

			out.WriteString(fmt.Sprintf("\n%s:\n", checkLabel))
			idxI64 := fmt.Sprintf("taset.i64.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i64\n", idxI64, idxArg))

			arrLenPtr := fmt.Sprintf("taset.len.ptr.%d", id)
			arrLen := fmt.Sprintf("taset.len.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %%%s, i64 8\n", arrLenPtr, arrArg))
			out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s, !invariant.load !{}\n", arrLen, arrLenPtr))

			inBounds := fmt.Sprintf("taset.in_bounds.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = icmp ult i64 %%%s, %%%s\n", inBounds, idxI64, arrLen))

			condFast := inBounds
			if !e.isInteger(instruction.Args[1]) && !e.isInteger(idxArg) {
				idxRoundtrip := fmt.Sprintf("taset.roundtrip.%d", id)
				isInt := fmt.Sprintf("taset.is_int.%d", id)
				out.WriteString(fmt.Sprintf("  %%%s = sitofp i64 %%%s to double\n", idxRoundtrip, idxI64))
				out.WriteString(fmt.Sprintf("  %%%s = fcmp oeq double %%%s, %%%s\n", isInt, idxArg, idxRoundtrip))

				condFastInt := fmt.Sprintf("taset.cond_fast.%d", id)
				out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", condFastInt, isInt, inBounds))
				condFast = condFastInt
			}
			out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\"branch_weights\", i32 10000, i32 1}\n", condFast, fastLabel, slowLabel))

			out.WriteString(fmt.Sprintf("\n%s:\n", fastLabel))
			dataPtrPtr := fmt.Sprintf("taset.data.ptr.%d", id)
			dataPtr := fmt.Sprintf("taset.data.%d", id)
			elemPtr := fmt.Sprintf("taset.elem.ptr.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %%%s, i64 40\n", dataPtrPtr, arrArg))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s, !invariant.load !{}\n", dataPtr, dataPtrPtr))
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds double, ptr %%%s, i64 %%%s\n", elemPtr, dataPtr, idxI64))
			out.WriteString(fmt.Sprintf("  store double %%%s, ptr %%%s\n", valArg, elemPtr))
			out.WriteString(fmt.Sprintf("  br label %%%s\n", doneLabel))

			out.WriteString(fmt.Sprintf("\n%s:\n", slowLabel))
			status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_typedarray_set(ptr %%%s, double %%%s, double %%%s)\n", status, arrArg, idxArg, valArg))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			out.WriteString(fmt.Sprintf("  br label %%%s\n", doneLabel))

			out.WriteString(fmt.Sprintf("\n%s:\n", doneLabel))
			return nil
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_typedarray_set(ptr %%%s, double %%%s, double %%%s)\n", status, arrArg, idxArg, valArg))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	}
	elemLLVMType := arrayElementLLVMType(arrayType)
	elemType := arrayElementType(arrayType)
	canFastPath := (elemType == ir.TypeNumber || isPointerType(elemType) || elemType == ir.TypeBool) &&
		arrayType != ir.TypeUnknownArray && arrayType != ""

	if canFastPath {
		expectedSize := int64(8)
		if elemType == ir.TypeBool {
			expectedSize = 1
		} else if isPointerType(elemType) {
			expectedSize = e.pointerSize()
		}

		id := e.labelCounter
		e.labelCounter++

		checkLabel := fmt.Sprintf("idxset.check.%d", id)
		fastLabel := fmt.Sprintf("idxset.fast.%d", id)
		slowLabel := fmt.Sprintf("idxset.slow.%d", id)
		doneLabel := fmt.Sprintf("idxset.done.%d", id)

		isNotNull := fmt.Sprintf("idxset.is_not_null.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne ptr %%%s, null\n", isNotNull, arrArg))
		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\"branch_weights\", i32 10000, i32 1}\n", isNotNull, checkLabel, slowLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", checkLabel))
		idxI64 := fmt.Sprintf("idxset.i64.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i64\n", idxI64, idxArg))

		arrLen := fmt.Sprintf("idxset.len.%d", id)
		invLen := ""
		if !e.hasArrayResize {
			invLen = ", !invariant.load !{}"
		}
		out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s%s\n", arrLen, arrArg, invLen))

		inBounds := fmt.Sprintf("idxset.in_bounds.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp ult i64 %%%s, %%%s\n", inBounds, idxI64, arrLen))

		elemSizePtr := fmt.Sprintf("idxset.esize.ptr.%d", id)
		elemSize := fmt.Sprintf("idxset.esize.%d", id)
		isExpectedSize := fmt.Sprintf("idxset.is_esize.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %%%s, i64 16\n", elemSizePtr, arrArg))
		out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s, !invariant.load !{}\n", elemSize, elemSizePtr))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i64 %%%s, %d\n", isExpectedSize, elemSize, expectedSize))

		condFast := fmt.Sprintf("idxset.cond_fast.%d", id)
		if !e.isInteger(instruction.Args[1]) && !e.isInteger(idxArg) {
			idxRoundtrip := fmt.Sprintf("idxset.roundtrip.%d", id)
			isInt := fmt.Sprintf("idxset.is_int.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = sitofp i64 %%%s to double\n", idxRoundtrip, idxI64))
			out.WriteString(fmt.Sprintf("  %%%s = fcmp oeq double %%%s, %%%s\n", isInt, idxArg, idxRoundtrip))

			cond1 := fmt.Sprintf("idxset.cond1.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", cond1, isInt, inBounds))
			out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", condFast, cond1, isExpectedSize))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", condFast, inBounds, isExpectedSize))
		}

		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\"branch_weights\", i32 10000, i32 1}\n", condFast, fastLabel, slowLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", fastLabel))
		dataPtrPtr := fmt.Sprintf("idxset.data.ptr.%d", id)
		dataPtr := fmt.Sprintf("idxset.data.%d", id)
		elemPtr := fmt.Sprintf("idxset.elem.ptr.%d", id)
		invData := ""
		if !e.hasArrayResize {
			invData = ", !invariant.load !{}"
		}
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %%%s, i64 24\n", dataPtrPtr, arrArg))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s%s\n", dataPtr, dataPtrPtr, invData))

		elemPtrType := elemLLVMType
		if elemType == ir.TypeBool {
			elemPtrType = "i8"
		}
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds %s, ptr %%%s, i64 %%%s\n", elemPtr, elemPtrType, dataPtr, idxI64))
		if elemType == ir.TypeBool {
			boolByte := fmt.Sprintf("idxset.bool.byte.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i8\n", boolByte, valArg))
			out.WriteString(fmt.Sprintf("  store i8 %%%s, ptr %%%s\n", boolByte, elemPtr))
		} else {
			out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", elemPtrType, valArg, elemPtr))
		}
		out.WriteString(fmt.Sprintf("  br label %%%s\n", doneLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", slowLabel))
		valSlot := fmt.Sprintf("%s.set.slot.%d", instruction.Args[0], e.runtimeStatus)
		out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", valSlot, elemLLVMType))
		out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", elemLLVMType, valArg, valSlot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		valueSize, err := arrayElementSizeForTarget(arrayType, e.pointerSize())
		if err != nil {
			return err
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_set_typed(ptr %%%s, double %%%s, ptr %%%s, i64 %d, i64 %d)\n", status, arrArg, idxArg, valSlot, valueSize, arrayElementTag(arrayType)))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  br label %%%s\n", doneLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", doneLabel))
		return nil
	}

	valSlot := fmt.Sprintf("%s.set.slot.%d", instruction.Args[0], e.runtimeStatus)
	out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", valSlot, elemLLVMType))
	out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", elemLLVMType, valArg, valSlot))
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	valueSize, err := arrayElementSizeForTarget(arrayType, e.pointerSize())
	if err != nil {
		return err
	}
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_set_typed(ptr %%%s, double %%%s, ptr %%%s, i64 %d, i64 %d)\n", status, arrArg, idxArg, valSlot, valueSize, arrayElementTag(arrayType)))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	return nil
}
