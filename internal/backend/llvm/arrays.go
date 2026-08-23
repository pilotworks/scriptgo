package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitArray(out *strings.Builder, instruction ir.Instruction) error {
	if !strings.HasSuffix(string(instruction.Type), "[]") && instruction.Type != ir.TypeNumberArray && instruction.Type != ir.TypeStringArray {
		return fmt.Errorf("unsupported LLVM array type %s", instruction.Type)
	}
	e.types[instruction.Result] = instruction.Type
	e.arrayTypes = append(e.arrayTypes, arrayReference{name: instruction.Result, typ: instruction.Type})
	slot := instruction.Result + ".slot"
	out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
	elementSize, err := arrayElementSize(instruction.Type)
	if err != nil {
		return err
	}
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_new(i64 %d, i64 %d, ptr %%%s)\n", status, len(instruction.Args), elementSize, slot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	for index, argument := range instruction.Args {
		valueSlot := fmt.Sprintf("%s.element.%d", instruction.Result, index)
		elementLLVMType := llvmType(arrayElementType(instruction.Type))
		out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", valueSlot, elementLLVMType))
		out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", elementLLVMType, argument, valueSlot))
		status = fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_set(ptr %%%s, double %s, ptr %%%s)\n", status, instruction.Result, llvmNumber(float64(index)), valueSlot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	}
	return nil
}

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
	if isTypedArrayType(arrayType) {
		e.types[instruction.Result] = ir.TypeNumber
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_typedarray_get(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	}
	e.types[instruction.Result] = instruction.Type
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	llvmT := llvmType(instruction.Type)
	out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", slot, llvmT))
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_get(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %%%s\n", instruction.Result, llvmT, slot))
	return nil
}

func (e *functionEmitter) emitIndexSet(out *strings.Builder, instruction ir.Instruction) error {
	if len(instruction.Args) != 3 {
		return fmt.Errorf("index.set instruction requires array, index, and value operands")
	}
	arrayType, ok := e.types[instruction.Args[0]]
	if !ok {
		return fmt.Errorf("unknown index.set array %q", instruction.Args[0])
	}
	if arrayType == ir.TypeBigInt64Array || arrayType == ir.TypeBigUint64Array {
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_typedarray_set_bigint(ptr %%%s, double %%%s, i64 %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	}
	if isTypedArrayType(arrayType) {
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_typedarray_set(ptr %%%s, double %%%s, double %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	}
	elemLLVMType := llvmType(arrayElementType(arrayType))
	valSlot := fmt.Sprintf("%s.set.slot.%d", instruction.Args[0], e.runtimeStatus)
	out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", valSlot, elemLLVMType))
	out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", elemLLVMType, instruction.Args[2], valSlot))
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_set(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], valSlot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	return nil
}

func (e *functionEmitter) emitArrayIntrinsic(out *strings.Builder, instruction ir.Instruction, arrayType ir.Type) error {
	switch instruction.Callee {
	case "__array.length":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("array.length has invalid signature")
		}
		resultSlot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i64\n", resultSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_length(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], resultSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.i64 = load i64, ptr %%%s\n", instruction.Result, resultSlot)
		fmt.Fprintf(out, "  %%%s = uitofp i64 %%%s.i64 to double\n", instruction.Result, instruction.Result)
		return nil
	case "__array.push":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("array.push has invalid signature")
		}
		elemType := arrayElementType(arrayType)
		valSlot := fmt.Sprintf("%s.push.val.%d", instruction.Args[0], e.runtimeStatus)
		fmt.Fprintf(out, "  %%%s = alloca %s\n", valSlot, llvmType(elemType))
		fmt.Fprintf(out, "  store %s %%%s, ptr %%%s\n", llvmType(elemType), instruction.Args[1], valSlot)
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_push(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], valSlot, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.pop":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("array.pop has invalid signature")
		}
		elemType := arrayElementType(arrayType)
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca %s\n", resSlot, llvmType(elemType))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_pop(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load %s, ptr %%%s\n", instruction.Result, llvmType(elemType), resSlot)
		return nil
	case "__array.slice":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != arrayType {
			return fmt.Errorf("array.slice has invalid signature")
		}
		startArg := "%" + instruction.Args[1]
		endArg := "-1.0"
		if len(instruction.Args) == 3 {
			endArg = "%" + instruction.Args[2]
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_slice(ptr %%%s, double %s, double %s, ptr %%%s)\n", status, instruction.Args[0], startArg, endArg, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.indexOf":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("array.indexOf has invalid signature")
		}
		fromArg := "0.0"
		if len(instruction.Args) == 3 {
			fromArg = "%" + instruction.Args[2]
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if arrayType == ir.TypeStringArray {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_index_of_string(ptr %%%s, ptr %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], fromArg, resSlot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_index_of_number(ptr %%%s, double %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], fromArg, resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.includes":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("array.includes has invalid signature")
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if arrayType == ir.TypeStringArray {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_includes_string(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], resSlot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_includes_number(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, resSlot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
		return nil
	case "__array.at":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("array.at has invalid signature")
		}
		elemType := arrayElementType(arrayType)
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca %s\n", resSlot, llvmType(elemType))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_at(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load %s, ptr %%%s\n", instruction.Result, llvmType(elemType), resSlot)
		return nil
	case "__array.shift":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("array.shift has invalid signature")
		}
		elemType := arrayElementType(arrayType)
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca %s\n", resSlot, llvmType(elemType))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_shift(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load %s, ptr %%%s\n", instruction.Result, llvmType(elemType), resSlot)
		return nil
	case "__array.unshift":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("array.unshift has invalid signature")
		}
		elemType := arrayElementType(arrayType)
		valSlot := fmt.Sprintf("%s.unshift.val.%d", instruction.Args[0], e.runtimeStatus)
		fmt.Fprintf(out, "  %%%s = alloca %s\n", valSlot, llvmType(elemType))
		fmt.Fprintf(out, "  store %s %%%s, ptr %%%s\n", llvmType(elemType), instruction.Args[1], valSlot)
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_unshift(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], valSlot, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.reverse":
		if len(instruction.Args) != 1 || instruction.Type != arrayType {
			return fmt.Errorf("array.reverse has invalid signature")
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_reverse(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.concat":
		if len(instruction.Args) != 2 || instruction.Type != arrayType {
			return fmt.Errorf("array.concat has invalid signature")
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_concat(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.splice":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != arrayType {
			return fmt.Errorf("array.splice has invalid signature")
		}
		startArg := "%" + instruction.Args[1]
		dcArg := "1000000000.0"
		if len(instruction.Args) == 3 {
			dcArg = "%" + instruction.Args[2]
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_splice(ptr %%%s, double %s, double %s, ptr %%%s)\n", status, instruction.Args[0], startArg, dcArg, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.join":
		if (len(instruction.Args) != 1 && len(instruction.Args) != 2) || instruction.Type != ir.TypeString {
			return fmt.Errorf("array.join has invalid signature")
		}
		sepArg := "null"
		if len(instruction.Args) == 2 {
			sepArg = "%" + instruction.Args[1]
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if arrayType == ir.TypeStringArray {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_string(ptr %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], sepArg, resSlot)
		} else if arrayType == ir.TypeBigIntArray || arrayType == "bigint[]" {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_bigint(ptr %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], sepArg, resSlot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_number(ptr %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], sepArg, resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.map", "__array.flatMap":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		fnName := "scriptgo_array_map_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_map_string"
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.filter":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		fnName := "scriptgo_array_filter_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_filter_string"
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.forEach":
		fnName := "scriptgo_array_for_each_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_for_each_string"
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	case "__array.reduce":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_reduce_number(ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.find":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_find_number(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.some":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_some_number(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s.i32 = load i32, ptr %%%s\n", instruction.Result, slot))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s.i32, 0\n", instruction.Result, instruction.Result))
		return nil
	case "__array.every":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_every_number(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s.i32 = load i32, ptr %%%s\n", instruction.Result, slot))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s.i32, 0\n", instruction.Result, instruction.Result))
		return nil
	case "__array.findIndex":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fnName := "scriptgo_array_find_index_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_find_index_string"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.fill":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		startVal := "0.000000e+00"
		hasStart := 0
		if len(instruction.Args) > 2 {
			startVal = fmt.Sprintf("%%%s", instruction.Args[2])
			hasStart = 1
		}
		endVal := "0.000000e+00"
		hasEnd := 0
		if len(instruction.Args) > 3 {
			endVal = fmt.Sprintf("%%%s", instruction.Args[3])
			hasEnd = 1
		}
		if arrayType == ir.TypeStringArray {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_fill_string(ptr %%%s, ptr %%%s, double %s, double %s, i32 %d, i32 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], startVal, endVal, hasStart, hasEnd, slot))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_fill_number(ptr %%%s, double %%%s, double %s, double %s, i32 %d, i32 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], startVal, endVal, hasStart, hasEnd, slot))
		}
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.toReversed":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_to_reversed(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.toSorted":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fnName := "scriptgo_array_to_sorted_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_to_sorted_string"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.findLast":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_find_last_number(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.findLastIndex":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fnName := "scriptgo_array_find_last_index_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_find_last_index_string"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.lastIndexOf":
		resSlot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", resSlot))
		fromArg := "-1.000000e+00"
		if len(instruction.Args) > 2 {
			fromArg = fmt.Sprintf("%%%s", instruction.Args[2])
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if arrayType == ir.TypeStringArray {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_last_index_of_string(ptr %%%s, ptr %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], fromArg, resSlot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_last_index_of_number(ptr %%%s, double %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], fromArg, resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.copyWithin":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		startVal := "0.000000e+00"
		hasStart := 0
		if len(instruction.Args) > 2 {
			startVal = fmt.Sprintf("%%%s", instruction.Args[2])
			hasStart = 1
		}
		endVal := "0.000000e+00"
		hasEnd := 0
		if len(instruction.Args) > 3 {
			endVal = fmt.Sprintf("%%%s", instruction.Args[3])
			hasEnd = 1
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_copy_within(ptr %%%s, double %%%s, double %s, double %s, i32 %d, i32 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], startVal, endVal, hasStart, hasEnd, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.with":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fnName := "scriptgo_array_with_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_with_string"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, double %%%s, %s %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], llvmType(arrayElementType(arrayType)), instruction.Args[2], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.toSpliced":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		delCount := "0.000000e+00"
		hasDel := 0
		if len(instruction.Args) > 2 {
			delCount = fmt.Sprintf("%%%s", instruction.Args[2])
			hasDel = 1
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_to_spliced(ptr %%%s, double %%%s, double %s, i32 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], delCount, hasDel, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.sort":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fnName := "scriptgo_array_sort_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_sort_string"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.reduceRight":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_reduce_right_number(ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.toString", "__array.toLocaleString":
		resSlot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", resSlot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if arrayType == ir.TypeStringArray {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_string(ptr %%%s, ptr null, ptr %%%s)\n", status, instruction.Args[0], resSlot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_number(ptr %%%s, ptr null, ptr %%%s)\n", status, instruction.Args[0], resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.values", "__array.flat":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_slice(ptr %%%s, double 0.0, double -1.0, ptr %%%s)\n", status, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	default:
		return fmt.Errorf("unknown array intrinsic %q", instruction.Callee)
	}
}

func isTypedArrayType(t ir.Type) bool {
	switch t {
	case ir.TypeInt8Array, ir.TypeUint8Array, ir.TypeUint8ClampedArray,
		ir.TypeInt16Array, ir.TypeUint16Array, ir.TypeInt32Array, ir.TypeUint32Array,
		ir.TypeFloat32Array, ir.TypeFloat64Array, ir.TypeBigInt64Array, ir.TypeBigUint64Array:
		return true
	default:
		return false
	}
}
