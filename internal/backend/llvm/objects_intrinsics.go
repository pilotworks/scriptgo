package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitObjectIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__object.freeze", "__object.seal", "__object.preventExtensions":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("%s requires one argument", instruction.Callee)
		}
		obj := e.ensurePointerArg(out, instruction.Args[0])
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		callee := "@scriptgo_object_freeze"
		if instruction.Callee == "__object.seal" {
			callee = "@scriptgo_object_seal"
		}
		if instruction.Callee == "__object.preventExtensions" {
			callee = "@scriptgo_object_prevent_extensions"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 %s(ptr %%%s, ptr %%%s)\n", status, callee, obj, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = ir.TypeObject
		return nil
	case "__object.isFrozen", "__object.isSealed", "__object.isExtensible":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("%s requires one argument", instruction.Callee)
		}
		obj := e.ensurePointerArg(out, instruction.Args[0])
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
		callee := "@scriptgo_object_is_frozen"
		if instruction.Callee == "__object.isSealed" {
			callee = "@scriptgo_object_is_sealed"
		}
		if instruction.Callee == "__object.isExtensible" {
			callee = "@scriptgo_object_is_extensible"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 %s(ptr %%%s, ptr %%%s)\n", status, callee, obj, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		loaded := instruction.Result + ".i32"
		out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%%s\n", loaded, slot))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, loaded))
		e.types[instruction.Result] = ir.TypeBool
		return nil
	case "__object.is":
		e.types[instruction.Result] = ir.TypeBool
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		t1 := e.types[instruction.Args[0]]
		switch t1 {
		case ir.TypeNumber:
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_is_number(double %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		case ir.TypeString:
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_is_string(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		case ir.TypeUnknown:
			arg0 := instruction.Args[0]
			if slot0, ok := e.varSlots[arg0]; ok {
				loaded0 := fmt.Sprintf("%s.is.loaded.%d", arg0, e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded0, slot0))
				arg0 = loaded0
			}
			arg1 := instruction.Args[1]
			if slot1, ok := e.varSlots[arg1]; ok {
				loaded1 := fmt.Sprintf("%s.is.loaded.%d", arg1, e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded1, slot1))
				arg1 = loaded1
			}
			value0, err := e.emitCanonicalValuePointer(out, arg0, ir.TypeUnknown, fmt.Sprintf("object.is.0.%d", e.loadCounter))
			if err != nil {
				return err
			}
			value1, err := e.emitCanonicalValuePointer(out, arg1, ir.TypeUnknown, fmt.Sprintf("object.is.1.%d", e.loadCounter))
			if err != nil {
				return err
			}
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_is_unknown(ptr %s, ptr %s, ptr %%%s)\n", status, value0, value1, slot))
		default:
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_is_ptr(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		}
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		i32Val := instruction.Result + ".i32"
		out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%%s\n", i32Val, slot))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Val))
	case "__object.entries":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_new(i64 2, i64 8, ptr %%%s)\n", status, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	case "__object.keys":
		objVar := instruction.Args[0]
		if e.types[objVar] == ir.TypeUnknown {
			e.tempCounter++
			payloadName := fmt.Sprintf("keys.unbox.payload.%d", e.tempCounter)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadName, objVar)
			e.tempCounter++
			ptrName := fmt.Sprintf("keys.unbox.ptr.%d", e.tempCounter)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrName, payloadName)
			objVar = ptrName
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_keys(ptr %%%s, ptr %%%s)\n", status, objVar, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	case "__object.groupBy":
		itemsArg := e.ensurePointerArg(out, instruction.Args[0])
		cbArg := e.ensurePointerArg(out, instruction.Args[1])
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_group_by(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, itemsArg, cbArg, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	case "__object.get_prop":
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
			ptrObj := "%" + objArg
			if objType == ir.TypeUnknown {
				payloadName := fmt.Sprintf("dynamic.prop.payload.%d", e.loadCounter)
				ptrName := fmt.Sprintf("dynamic.prop.ptr.%d", e.loadCounter)
				e.loadCounter++
				fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadName, objArg)
				fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrName, payloadName)
				ptrObj = "%" + ptrName
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
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_unknown_get(ptr %s, ptr %s, ptr %%%s)\n", status, ptrObj, ptrProperty, valueSlot)
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
	case "__object.delete_prop":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("__object.delete_prop requires object and property")
		}
		objArg := e.resolveArg(out, instruction.Args[0])
		propertyArg := e.resolveArg(out, instruction.Args[1])
		objType := e.types[instruction.Args[0]]
		ptrObj := objArg
		if objType == ir.TypeUnknown {
			payloadName := fmt.Sprintf("dynamic.delete.payload.%d", e.loadCounter)
			ptrName := fmt.Sprintf("dynamic.delete.ptr.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadName, objArg)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrName, payloadName)
			ptrObj = ptrName
		}
		ptrProperty := propertyArg
		if e.types[instruction.Args[1]] == ir.TypeUnknown {
			propPayload := fmt.Sprintf("dynamic.delete.prop.payload.%d", e.loadCounter)
			propPtr := fmt.Sprintf("dynamic.delete.prop.ptr.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", propPayload, propertyArg)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", propPtr, propPayload)
			ptrProperty = propPtr
		}
		outSlot := fmt.Sprintf("delete.prop.out.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = alloca i32\n", outSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_delete_property(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, ptrObj, ptrProperty, outSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		i32Val := fmt.Sprintf("delete.prop.val.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", i32Val, outSlot)
		fmt.Fprintf(out, "  %%%s = trunc i32 %%%s to i1\n", instruction.Result, i32Val)
		e.types[instruction.Result] = ir.TypeBool
		return nil
	case "__object.set_prop":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("__object.set_prop requires object, property, and value")
		}
		objArg := e.resolveArg(out, instruction.Args[0])
		propertyArg := e.resolveArg(out, instruction.Args[1])
		valueArg := e.resolveArg(out, instruction.Args[2])
		objType := e.types[instruction.Args[0]]
		ptrObj := objArg
		if objType == ir.TypeUnknown {
			payloadName := fmt.Sprintf("dynamic.set.payload.%d", e.loadCounter)
			ptrName := fmt.Sprintf("dynamic.set.ptr.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadName, objArg)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrName, payloadName)
			ptrObj = ptrName
		}
		ptrProperty := propertyArg
		if e.types[instruction.Args[1]] == ir.TypeUnknown {
			propPayload := fmt.Sprintf("dynamic.set.prop.payload.%d", e.loadCounter)
			propPtr := fmt.Sprintf("dynamic.set.prop.ptr.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", propPayload, propertyArg)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", propPtr, propPayload)
			ptrProperty = propPtr
		}
		valueType := e.types[instruction.Args[2]]
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		switch valueType {
		case ir.TypeString:
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_string_set(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, ptrObj, ptrProperty, valueArg)
		case ir.TypeNumber:
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_number_set(ptr %%%s, ptr %%%s, double %%%s)\n", status, ptrObj, ptrProperty, valueArg)
		case ir.TypeBool:
			boolValue := fmt.Sprintf("dynamic.set.bool.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", boolValue, valueArg)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_bool_set(ptr %%%s, ptr %%%s, i32 %%%s)\n", status, ptrObj, ptrProperty, boolValue)
		case ir.TypeBigInt:
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_bigint_set(ptr %%%s, ptr %%%s, i64 %%%s)\n", status, ptrObj, ptrProperty, valueArg)
		default:
			boxed := valueArg
			if valueType != ir.TypeUnknown {
				boxed = fmt.Sprintf("dynamic.set.boxed.%d", e.loadCounter)
				e.loadCounter++
				if err := e.emitBoxValue(out, valueArg, valueType, boxed); err != nil {
					return err
				}
			}
			boxedSlot := fmt.Sprintf("dynamic.set.value.slot.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = alloca { i32, i32, i64, i64 }\n", boxedSlot)
			fmt.Fprintf(out, "  store { i32, i32, i64, i64 } %%%s, ptr %%%s\n", boxed, boxedSlot)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_unknown_set(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, ptrObj, ptrProperty, boxedSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	default:
		if strings.HasPrefix(instruction.Callee, "__object.") {
			if instruction.Type == ir.TypeBool {
				out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 1, 1\n", instruction.Result))
				return nil
			}
			if instruction.Type == ir.TypeNumber {
				out.WriteString(fmt.Sprintf("  %%%s = fadd double 0.0, 1.0\n", instruction.Result))
				return nil
			}
			if instruction.Type == ir.TypeUnknown {
				var argVal string
				argType := ir.TypeObject
				if len(instruction.Args) > 0 {
					argVal = instruction.Args[0]
					if t, ok := e.types[argVal]; ok && t != "" {
						argType = t
					}
				}
				if argVal != "" {
					return e.emitBoxValue(out, argVal, argType, instruction.Result)
				}
				out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 0, 0\n", instruction.Result))
				return nil
			}
			if len(instruction.Args) > 0 {
				ptrArg := e.ensurePointerArg(out, instruction.Args[0])
				out.WriteString(fmt.Sprintf("  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, ptrArg))
			} else {
				out.WriteString(fmt.Sprintf("  %%%s = alloca i8\n", instruction.Result))
			}
			return nil
		}
		return fmt.Errorf("unknown object intrinsic %q", instruction.Callee)
	}
	return nil
}
