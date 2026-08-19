package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitArray(out *strings.Builder, instruction ir.Instruction) error {
	if instruction.Type != ir.TypeNumberArray && instruction.Type != ir.TypeStringArray {
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
	arrayType, ok := e.types[instruction.Args[0]]
	if !ok {
		return fmt.Errorf("unknown index array %q", instruction.Args[0])
	}
	e.types[instruction.Result] = instruction.Type
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	if arrayType == ir.TypeStringArray {
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_get(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	} else {
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_get(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
	}
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
	default:
		return fmt.Errorf("unknown array intrinsic %q", instruction.Callee)
	}
}


