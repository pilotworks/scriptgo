package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitObjectNew(out *strings.Builder, instruction ir.Instruction) error {
	if instruction.FieldCount < 0 {
		return fmt.Errorf("object shape %q has invalid field count", instruction.Callee)
	}
	e.types[instruction.Result] = instruction.Type
	e.objects = append(e.objects, instruction.Result)
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_new(i64 %d, ptr %%%s)\n", status, instruction.FieldCount, slot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	if instruction.Value != "" {
		if strGlobal, ok := e.stringsByValue[instruction.Value]; ok {
			typeStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_type_set(ptr %%%s, ptr %s)\n", typeStatus, instruction.Result, strGlobal))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", typeStatus))
		}
	}
	return nil
}

func (e *functionEmitter) emitFieldSet(out *strings.Builder, instruction ir.Instruction) error {
	if instruction.FieldIndex < 0 {
		return fmt.Errorf("object field %q has invalid index", instruction.Field)
	}
	valueType, ok := e.types[instruction.Args[1]]
	if !ok {
		return fmt.Errorf("unknown object field value %q", instruction.Args[1])
	}
	switch {
	case valueType == ir.TypeNumber:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 %d, double %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case valueType == ir.TypeBool:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		boolI32 := fmt.Sprintf("obj.bool.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i32\n", boolI32, instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_bool_set(ptr %%%s, i64 %d, i32 %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, boolI32))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case valueType == ir.TypeString || valueType == ir.TypeNumberArray || valueType == ir.TypeStringArray || strings.HasPrefix(string(valueType), "object:"):
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_ptr_set(ptr %%%s, i64 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	default:
		return fmt.Errorf("unsupported object field type %s", valueType)
	}
	return nil
}

func (e *functionEmitter) emitFieldGet(out *strings.Builder, instruction ir.Instruction) error {
	if instruction.FieldIndex < 0 {
		return fmt.Errorf("object field %q has invalid index", instruction.Field)
	}
	e.types[instruction.Result] = instruction.Type
	slot := instruction.Result + ".slot"
	switch {
	case instruction.Type == ir.TypeNumber:
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_get(ptr %%%s, i64 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
	case instruction.Type == ir.TypeBool:
		out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_bool_get(ptr %%%s, i64 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		boolI32 := instruction.Result + ".i32"
		out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%%s\n", boolI32, slot))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, boolI32))
	case instruction.Type == ir.TypeString || instruction.Type == ir.TypeNumberArray || instruction.Type == ir.TypeStringArray || strings.HasPrefix(string(instruction.Type), "object:"):
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_ptr_get(ptr %%%s, i64 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	default:
		return fmt.Errorf("unsupported object field type %s", instruction.Type)
	}
	return nil
}

func (e *functionEmitter) emitInstanceOf(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = ir.TypeBool
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
	strGlobal, ok := e.stringsByValue[instruction.Value]
	if !ok {
		return fmt.Errorf("unknown string literal %q for instanceof", instruction.Value)
	}
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_instanceof(ptr %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], strGlobal, slot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	i32Val := instruction.Result + ".i32"
	out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%%s\n", i32Val, slot))
	out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Val))
	return nil
}
