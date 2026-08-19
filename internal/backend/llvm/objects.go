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
	switch valueType {
	case ir.TypeNumber:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 %d, double %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case ir.TypeString:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_string_set(ptr %%%s, i64 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, instruction.Args[1]))
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
	switch instruction.Type {
	case ir.TypeNumber:
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_get(ptr %%%s, i64 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
	case ir.TypeString:
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_string_get(ptr %%%s, i64 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.FieldIndex, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	default:
		return fmt.Errorf("unsupported object field type %s", instruction.Type)
	}
	return nil
}
