package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitObjectIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__object.reflect_set", "__object.reflect_delete":
		// Reflect.set(target, key, value) / Reflect.deleteProperty(target, key):
		// the runtime reports false where the integrity level forbids the change.
		want := 3
		if instruction.Callee == "__object.reflect_delete" {
			want = 2
		}
		if len(instruction.Args) != want {
			return fmt.Errorf("%s requires %d arguments", instruction.Callee, want)
		}
		obj := e.ensurePointerArg(out, instruction.Args[0])
		property := e.resolveArg(out, instruction.Args[1])
		slot := instruction.Result + ".reflect.slot"
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if want == 3 {
			valuePtr, err := e.emitCanonicalValuePointer(out, e.resolveArg(out, instruction.Args[2]), ir.TypeUnknown, instruction.Result+".reflect.value")
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_reflect_set(ptr %%%s, ptr %%%s, ptr %s, ptr %%%s)\n", status, obj, property, valuePtr, slot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_reflect_delete(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, obj, property, slot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		loaded := instruction.Result + ".reflect.i32"
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", loaded, slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, loaded)
		e.types[instruction.Result] = ir.TypeBool
		return nil
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
		e.types[instruction.Result] = instruction.Type
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
		t0, t1 := e.types[instruction.Args[0]], e.types[instruction.Args[1]]
		if t1 == "" {
			t1 = t0
		}
		switch {
		case t0 == ir.TypeNumber && t1 == ir.TypeNumber:
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_is_number(double %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		case t0 == ir.TypeString && t1 == ir.TypeString:
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_is_string(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		case t0 == ir.TypeUnknown || t1 == ir.TypeUnknown || t0 != t1:
			// Mixed or boxed operands compare as canonical runtime values so
			// that SameValue semantics (NaN, signed zero, tags) hold.
			types := [2]ir.Type{t0, t1}
			operands := [2]string{instruction.Args[0], instruction.Args[1]}
			for index, arg := range operands {
				if types[index] != ir.TypeUnknown {
					operands[index] = e.resolveArg(out, arg)
					continue
				}
				if argSlot, ok := e.varSlots[arg]; ok {
					loaded := fmt.Sprintf("%s.is.loaded.%d", arg, e.loadCounter)
					e.loadCounter++
					out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded, argSlot))
					operands[index] = loaded
				}
			}
			var values [2]string
			for index, arg := range operands {
				value, err := e.emitCanonicalValuePointer(out, arg, types[index], fmt.Sprintf("object.is.%d.%d", index, e.loadCounter))
				if err != nil {
					return err
				}
				values[index] = value
			}
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_is_unknown(ptr %s, ptr %s, ptr %%%s)\n", status, values[0], values[1], slot))
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
	case "__object.isPrototypeOf":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("__object.isPrototypeOf requires a receiver and a value")
		}
		arg := instruction.Args[1]
		argType := e.types[arg]
		boxed := e.resolveArg(out, arg)
		if !e.isParamUnknown(arg) && argType != ir.TypeUnknown {
			boxed = fmt.Sprintf("%s.isproto.boxed.%d", instruction.Result, e.loadCounter)
			e.loadCounter++
			if err := e.emitBoxValue(out, arg, argType, boxed); err != nil {
				return err
			}
		}
		slot := instruction.Result + ".isproto.slot"
		fmt.Fprintf(out, "  %%%s = alloca { i32, i32, i64, i64 }\n", slot)
		fmt.Fprintf(out, "  store { i32, i32, i64, i64 } %%%s, ptr %%%s\n", boxed, slot)
		fmt.Fprintf(out, "  %%%s.i32 = call i32 @scriptgo_object_is_prototype_of(ptr %%%s, ptr %%%s)\n", instruction.Result, e.resolveArg(out, instruction.Args[0]), slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s.i32, 0\n", instruction.Result, instruction.Result)
		return nil
	case "__object.getPrototypeOf":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("__object.getPrototypeOf requires one argument")
		}
		arg := instruction.Args[0]
		argType := e.types[arg]
		if e.isParamUnknown(arg) {
			argType = ir.TypeUnknown
		}
		if llvmType(argType) == "ptr" {
			fmt.Fprintf(out, "  %%%s = call ptr @scriptgo_object_get_prototype(ptr %%%s)\n", instruction.Result, e.resolveArg(out, arg))
			return nil
		}
		boxed := arg
		if argType == ir.TypeUnknown {
			boxed = e.resolveArg(out, arg)
		} else {
			boxed = fmt.Sprintf("%s.proto.boxed.%d", instruction.Result, e.loadCounter)
			e.loadCounter++
			if err := e.emitBoxValue(out, arg, argType, boxed); err != nil {
				return err
			}
		}
		slot := instruction.Result + ".proto.slot"
		fmt.Fprintf(out, "  %%%s = alloca { i32, i32, i64, i64 }\n", slot)
		fmt.Fprintf(out, "  store { i32, i32, i64, i64 } %%%s, ptr %%%s\n", boxed, slot)
		fmt.Fprintf(out, "  %%%s = call ptr @scriptgo_object_get_prototype_value(ptr %%%s)\n", instruction.Result, slot)
		return nil
	case "__object.new":
		// Object(value) for a value that is already an object (lowering
		// handles every other case): the result is the value itself.
		if instruction.Type == ir.TypeUnknown {
			if len(instruction.Args) == 0 {
				return fmt.Errorf("__object.new requires an argument")
			}
			argType := e.types[instruction.Args[0]]
			if argType == "" {
				argType = ir.TypeObject
			}
			return e.emitBoxValue(out, instruction.Args[0], argType, instruction.Result)
		}
		if len(instruction.Args) == 0 {
			return fmt.Errorf("__object.new requires an argument")
		}
		ptrArg := e.ensurePointerArg(out, instruction.Args[0])
		out.WriteString(fmt.Sprintf("  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, ptrArg))
		return nil
	default:
		return fmt.Errorf("unknown object intrinsic %q", instruction.Callee)
	}
	return nil
}
