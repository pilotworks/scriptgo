package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

// emitObjectGetProp lowers __object.get_prop, a property read by name whose
// receiver or result type is only known at run time.
func (e *functionEmitter) emitObjectGetProp(out *strings.Builder, instruction ir.Instruction) error {
	// Union values stay tagged until a narrowing operation selects a
	// concrete representation. Numeric properties on those values need the
	// runtime type tag, rather than the old placeholder number.
	if len(instruction.Args) == 2 && instruction.Type == ir.TypeNumber {
		objArg := e.resolveArg(out, instruction.Args[0])
		if e.types[instruction.Args[0]] == ir.TypeUnknown || e.types[objArg] == ir.TypeUnknown {
			propArg := e.resolveArg(out, instruction.Args[1])
			valuePtr, err := e.emitCanonicalValuePointer(out, objArg, ir.TypeUnknown, fmt.Sprintf("unknown.prop.%d", e.loadCounter))
			if err != nil {
				return err
			}
			slot := instruction.Result + ".slot"
			status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_unknown_number_property(ptr %s, ptr %%%s, ptr %%%s)\n", status, valuePtr, propArg, slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
			return nil
		}
	}
	if instruction.Type == ir.TypeUnknown {
		objArg := e.resolveArg(out, instruction.Args[0])
		objType := e.types[instruction.Args[0]]
		getter := "scriptgo_object_property_unknown_get"
		ptrObj := "%" + objArg
		if objType == ir.TypeUnknown {
			// A dynamic receiver may be a string or another primitive, so
			// the runtime reads the property from the whole tagged value.
			valuePtr, err := e.emitCanonicalValuePointer(out, objArg, ir.TypeUnknown, fmt.Sprintf("dynamic.prop.%d", e.loadCounter))
			if err != nil {
				return err
			}
			e.loadCounter++
			getter = "scriptgo_unknown_property_get"
			ptrObj = valuePtr
		}
		propertyArg := e.resolveArg(out, instruction.Args[1])
		ptrProperty := "%" + propertyArg
		if e.types[instruction.Args[1]] == ir.TypeUnknown {
			propPayload := fmt.Sprintf("dynamic.propname.payload.%d", e.loadCounter)
			propPtr := fmt.Sprintf("dynamic.propname.ptr.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", propPayload, propertyArg)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", propPtr, propPayload)
			ptrProperty = "%" + propPtr
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		e.types[instruction.Result] = ir.TypeUnknown
		valueSlot := fmt.Sprintf("%s.dynamic.value.slot", instruction.Result)
		fmt.Fprintf(out, "  %%%s = alloca { i32, i32, i64, i64 }\n", valueSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %s, ptr %s, ptr %%%s)\n", status, getter, ptrObj, ptrProperty, valueSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", instruction.Result, valueSlot)
		return nil
	}
	objArg := e.resolveArg(out, instruction.Args[0])
	objType := e.types[instruction.Args[0]]
	ptrObj := objArg
	if objType == ir.TypeUnknown {
		payloadName := fmt.Sprintf("dynamic.prop.payload.%d", e.loadCounter)
		ptrName := fmt.Sprintf("dynamic.prop.ptr.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadName, objArg)
		fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrName, payloadName)
		ptrObj = ptrName
	}
	propertyArg := e.resolveArg(out, instruction.Args[1])
	ptrProperty := "%" + propertyArg
	if e.types[instruction.Args[1]] == ir.TypeUnknown {
		propPayload := fmt.Sprintf("dynamic.propname.payload.%d", e.loadCounter)
		propPtr := fmt.Sprintf("dynamic.propname.ptr.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", propPayload, propertyArg)
		fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", propPtr, propPayload)
		ptrProperty = "%" + propPtr
	}
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	e.types[instruction.Result] = instruction.Type
	switch instruction.Type {
	case ir.TypeNumber:
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		if objType == ir.TypeUnknown {
			valuePtr, err := e.emitCanonicalValuePointer(out, objArg, ir.TypeUnknown, fmt.Sprintf("dynamic.prop.%d", e.loadCounter))
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_unknown_number_property(ptr %s, ptr %s, ptr %%%s)\n", status, valuePtr, ptrProperty, slot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_number_get(ptr %%%s, ptr %s, ptr %%%s)\n", status, ptrObj, ptrProperty, slot)
		}
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
	case ir.TypeString:
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_string_get(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, ptrObj, propertyArg, slot)
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	case ir.TypeBool:
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_bool_get(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, ptrObj, propertyArg, slot)
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		loaded := instruction.Result + ".i32"
		out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%%s\n", loaded, slot))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, loaded))
	case ir.TypeBigInt:
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca i64\n", slot))
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_bigint_get(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, ptrObj, propertyArg, slot)
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s\n", instruction.Result, slot))
	default:
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_ptr_get(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, ptrObj, propertyArg, slot)
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	}
	return nil
}
