package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) dynamicFieldName(out *strings.Builder, field string) string {
	global, ok := e.stringsByValue[field]
	if !ok {
		return ""
	}
	name := fmt.Sprintf("dynamic.field.name.%d", e.loadCounter)
	e.loadCounter++
	fmt.Fprintf(out, "  %%%s = getelementptr inbounds [%d x i8], ptr %s, i64 0, i64 0\n", name, len([]byte(field))+1, global)
	return name
}

func (e *functionEmitter) emitDynamicFieldGet(out *strings.Builder, instruction ir.Instruction, object string) error {
	property := e.dynamicFieldName(out, instruction.Field)
	if property == "" {
		return fmt.Errorf("unknown dynamic object field name %q", instruction.Field)
	}
	e.types[instruction.Result] = instruction.Type

	switch instruction.Type {
	case ir.TypeNumber:
		slot := instruction.Result + ".dynamic.number.slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_number_get(ptr %s, ptr %%%s, ptr %%%s)\n", status, object, property, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case ir.TypeBool:
		slot := instruction.Result + ".dynamic.bool.slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_bool_get(ptr %s, ptr %%%s, ptr %%%s)\n", status, object, property, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		loaded := instruction.Result + ".dynamic.bool.i32"
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", loaded, slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, loaded)
	case ir.TypeBigInt:
		slot := instruction.Result + ".dynamic.bigint.slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i64\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_bigint_get(ptr %s, ptr %%%s, ptr %%%s)\n", status, object, property, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load i64, ptr %%%s\n", instruction.Result, slot)
	case ir.TypeUnknown:
		tagSlot := instruction.Result + ".dynamic.tag.slot"
		payloadSlot := instruction.Result + ".dynamic.payload.slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i32\n", tagSlot)
		fmt.Fprintf(out, "  %%%s = alloca i64\n", payloadSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_unknown_get(ptr %s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, object, property, tagSlot, payloadSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		tag := instruction.Result + ".dynamic.tag"
		payload := instruction.Result + ".dynamic.payload"
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", tag, tagSlot)
		fmt.Fprintf(out, "  %%%s = load i64, ptr %%%s\n", payload, payloadSlot)
		box0 := instruction.Result + ".dynamic.box0"
		box1 := instruction.Result + ".dynamic.box1"
		fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64 } undef, i32 %%%s, 0\n", box0, tag)
		fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", box1, box0)
		fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, box1, payload)
	default:
		slot := instruction.Result + ".dynamic.ptr.slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_ptr_get(ptr %s, ptr %%%s, ptr %%%s)\n", status, object, property, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	}
	return nil
}

func (e *functionEmitter) emitDynamicFieldSet(out *strings.Builder, instruction ir.Instruction, object, value string, actualType, expectedType ir.Type) error {
	property := e.dynamicFieldName(out, instruction.Field)
	if property == "" {
		return fmt.Errorf("unknown dynamic object field name %q", instruction.Field)
	}
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++

	if expectedType == ir.TypeNumber && (actualType == ir.TypePointer || actualType == ir.TypeVoid) {
		nan := fmt.Sprintf("dynamic.field.nan.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = fdiv double 0.0, 0.0\n", nan)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_number_set(ptr %s, ptr %%%s, double %%%s)\n", status, object, property, nan)
	} else {
		switch actualType {
		case ir.TypeNumber:
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_number_set(ptr %s, ptr %%%s, double %%%s)\n", status, object, property, value)
		case ir.TypeBool:
			boolValue := fmt.Sprintf("dynamic.field.bool.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", boolValue, value)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_bool_set(ptr %s, ptr %%%s, i32 %%%s)\n", status, object, property, boolValue)
		case ir.TypeBigInt:
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_bigint_set(ptr %s, ptr %%%s, i64 %%%s)\n", status, object, property, value)
		case ir.TypeUnknown:
			tag := fmt.Sprintf("dynamic.field.tag.%d", e.loadCounter)
			payload := fmt.Sprintf("dynamic.field.payload.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag, value)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payload, value)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_unknown_set(ptr %s, ptr %%%s, i32 %%%s, i64 %%%s)\n", status, object, property, tag, payload)
		default:
			boxed := fmt.Sprintf("dynamic.field.boxed.%d", e.loadCounter)
			e.loadCounter++
			if err := e.emitBoxValue(out, value, actualType, boxed); err != nil {
				return err
			}
			tag := fmt.Sprintf("dynamic.field.tag.%d", e.loadCounter)
			payload := fmt.Sprintf("dynamic.field.payload.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag, boxed)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payload, boxed)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_unknown_set(ptr %s, ptr %%%s, i32 %%%s, i64 %%%s)\n", status, object, property, tag, payload)
		}
	}
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	return nil
}
