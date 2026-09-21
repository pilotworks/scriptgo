package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func arrayElementLLVMType(arrayType ir.Type) string {
	elem := arrayElementType(arrayType)
	t := llvmType(elem)
	if t == "void" || t == "" {
		return "{ i32, i32, i64, i64 }"
	}
	return t
}

func arrayElementTag(arrayType ir.Type) int {
	elemType := arrayElementType(arrayType)
	switch elemType {
	case ir.TypeNumber:
		return 3
	case ir.TypeString:
		return 4
	case ir.TypeBool:
		return 2
	case ir.TypeBigInt:
		return 8
	case ir.TypeSymbol:
		return 9
	case ir.TypeClosure:
		return 7
	}
	if elemType == "closure" || elemType == "Function" || elemType == "function" || strings.Contains(string(elemType), "=>") {
		return 7
	}
	if strings.HasPrefix(string(elemType), "object:") || elemType == ir.TypeObject {
		return 5
	}
	return 0
}

func (e *functionEmitter) emitArray(out *strings.Builder, instruction ir.Instruction) error {
	if !strings.HasSuffix(string(instruction.Type), "[]") && instruction.Type != ir.TypeNumberArray && instruction.Type != ir.TypeStringArray {
		return fmt.Errorf("unsupported LLVM array type %s", instruction.Type)
	}
	e.types[instruction.Result] = instruction.Type
	e.arrayTypes = append(e.arrayTypes, arrayReference{name: instruction.Result, typ: instruction.Type})
	slot := instruction.Result + ".slot"
	if existingSlot, ok := e.varSlots[instruction.Result]; ok {
		slot = existingSlot
	} else {
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		e.varSlots[instruction.Result] = slot
	}
	elementSize, err := arrayElementSizeForTarget(instruction.Type, e.pointerSize())
	if err != nil {
		return err
	}
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_new(i64 %d, i64 %d, ptr %%%s)\n", status, len(instruction.Args), elementSize, slot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	if tag := arrayElementTag(instruction.Type); tag > 0 {
		status = fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_set_tag(ptr %%%s, i64 %d)\n", status, instruction.Result, tag))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	}
	for index, argument := range instruction.Args {
		argVal := e.resolveArg(out, argument)
		valueSlot := fmt.Sprintf("%s.element.%d", instruction.Result, index)
		elementLLVMType := arrayElementLLVMType(instruction.Type)
		out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", valueSlot, elementLLVMType))
		out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", elementLLVMType, argVal, valueSlot))
		status = fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_set(ptr %%%s, double %s, ptr %%%s)\n", status, instruction.Result, llvmNumber(float64(index)), valueSlot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	}
	return nil
}

func isTypedArrayType(t ir.Type) bool {
	switch t {
	case ir.TypeInt8Array, ir.TypeUint8Array, ir.TypeUint8ClampedArray,
		ir.TypeInt16Array, ir.TypeUint16Array, ir.TypeInt32Array, ir.TypeUint32Array,
		ir.TypeFloat32Array, ir.TypeFloat64Array, ir.TypeBigInt64Array, ir.TypeBigUint64Array,
		ir.TypeBuffer:
		return true
	default:
		return false
	}
}
