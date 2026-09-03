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
	typeName := instruction.Value
	if typeName == "" {
		for _, s := range e.module.Shapes {
			if s.Name == instruction.Callee && len(s.Fields) > 0 {
				var names []string
				for _, f := range s.Fields {
					names = append(names, f.Name)
				}
				typeName = ":" + strings.Join(names, ":") + ":"
				break
			}
		}
	}
	if typeName == "" {
		typeName = instruction.Callee
	}
	if typeName != "" {
		if strGlobal, ok := e.stringsByValue[typeName]; ok {
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
	var valueType ir.Type
	if instruction.Callee != "" && len(e.module.Shapes) > 0 {
		for _, s := range e.module.Shapes {
			if s.Name == instruction.Callee {
				if instruction.FieldIndex >= 0 && instruction.FieldIndex < len(s.Fields) {
					valueType = s.Fields[instruction.FieldIndex].Type
				} else {
					for _, f := range s.Fields {
						if f.Name == instruction.Field {
							valueType = f.Type
							break
						}
					}
				}
				break
			}
		}
	}
	if valueType == "" || valueType == ir.TypeVoid {
		if instruction.Type != "" && instruction.Type != ir.TypeVoid {
			valueType = instruction.Type
		} else if typ, ok := e.types[instruction.Args[1]]; ok && typ != "" && typ != ir.TypeVoid {
			valueType = typ
		} else {
			valueType = ir.TypePointer
		}
	}
	valArg := e.resolveArg(out, instruction.Args[1])
	objArg := e.resolveArg(out, instruction.Args[0])
	objType := e.types[instruction.Args[0]]
	ptrObj := "%" + objArg
	if objType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, objArg))
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
		ptrObj = "%" + ptrVar
	}
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	actualType := e.types[instruction.Args[1]]
	if actualType == "" {
		actualType = e.types[valArg]
	}
	if actualType == "" {
		actualType = valueType
	}
	if valueType == ir.TypeUnknown && actualType != ir.TypeUnknown {
		boxedVar := fmt.Sprintf("box.fset.%d", e.loadCounter)
		if err := e.emitBoxValue(out, valArg, actualType, boxedVar); err != nil {
			return err
		}
		valArg = boxedVar
		actualType = ir.TypeUnknown
	}
	if instruction.DynamicField {
		return e.emitDynamicFieldSet(out, instruction, ptrObj, valArg, actualType, valueType)
	}
	switch {
	case actualType == ir.TypeNumber:
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_set(ptr %s, i64 %d, double %%%s)\n", status, ptrObj, instruction.FieldIndex, valArg))
	case actualType == ir.TypeBool:
		boolI32 := fmt.Sprintf("obj.bool.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i32\n", boolI32, valArg))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_bool_set(ptr %s, i64 %d, i32 %%%s)\n", status, ptrObj, instruction.FieldIndex, boolI32))
	case actualType == ir.TypeBigInt:
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_bigint_set(ptr %s, i64 %d, i64 %%%s)\n", status, ptrObj, instruction.FieldIndex, valArg))
	case actualType == ir.TypeUnknown:
		tagVar := fmt.Sprintf("tag.%d", e.loadCounter)
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, valArg))
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, valArg))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_unknown_set(ptr %s, i64 %d, i32 %%%s, i64 %%%s)\n", status, ptrObj, instruction.FieldIndex, tagVar, payloadVar))
	default:
		if (actualType == "ptr" || actualType == ir.TypePointer || actualType == ir.TypeVoid) && valueType == ir.TypeNumber {
			nanVar := fmt.Sprintf("nan.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = fdiv double 0.0, 0.0\n", nanVar))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_set(ptr %s, i64 %d, double %%%s)\n", status, ptrObj, instruction.FieldIndex, nanVar))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_ptr_set(ptr %s, i64 %d, ptr %%%s)\n", status, ptrObj, instruction.FieldIndex, valArg))
		}
	}
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	return nil
}

func (e *functionEmitter) emitFieldGet(out *strings.Builder, instruction ir.Instruction) error {
	if instruction.FieldIndex < 0 {
		return fmt.Errorf("object field %q has invalid index", instruction.Field)
	}
	e.types[instruction.Result] = instruction.Type
	objArg := e.resolveArg(out, instruction.Args[0])
	objType := e.types[instruction.Args[0]]
	ptrObj := "%" + objArg
	if objType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, objArg))
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
		ptrObj = "%" + ptrVar
	}
	e.types[instruction.Result] = instruction.Type
	actualFieldType := instruction.Type
	shapeName := instruction.Callee
	if shapeName == "" && strings.HasPrefix(string(objType), "object:") {
		shapeName = strings.TrimPrefix(string(objType), "object:")
	}
	if shapeName != "" {
		for _, s := range e.module.Shapes {
			if s.Name == shapeName || strings.HasPrefix(s.Name, shapeName+"<") {
				if instruction.FieldIndex >= 0 && instruction.FieldIndex < len(s.Fields) {
					actualFieldType = s.Fields[instruction.FieldIndex].Type
				} else {
					for _, f := range s.Fields {
						if f.Name == instruction.Field {
							actualFieldType = f.Type
							break
						}
					}
				}
				break
			}
		}
	}
	if instruction.Type == ir.TypeUnknown {
		actualFieldType = ir.TypeUnknown
	}
	if instruction.DynamicField {
		return e.emitDynamicFieldGet(out, instruction, ptrObj)
	}

	if actualFieldType == ir.TypeUnknown {
		slotTag := fmt.Sprintf("slot_tag.%d", e.loadCounter)
		slotPayload := fmt.Sprintf("slot_payload.%d", e.loadCounter)
		tagLoaded := fmt.Sprintf("tag.loaded.%d", e.loadCounter)
		payloadLoaded := fmt.Sprintf("payload.loaded.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slotTag))
		out.WriteString(fmt.Sprintf("  %%%s = alloca i64\n", slotPayload))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_unknown_get(ptr %s, i64 %d, ptr %%%s, ptr %%%s)\n", status, ptrObj, instruction.FieldIndex, slotTag, slotPayload))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%%s\n", tagLoaded, slotTag))
		out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s\n", payloadLoaded, slotPayload))
		if instruction.Type == ir.TypeUnknown {
			b0 := fmt.Sprintf("box.b0.%d", e.loadCounter)
			b1 := fmt.Sprintf("box.b1.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } undef, i32 %%%s, 0\n", b0, tagLoaded))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", b1, b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, b1, payloadLoaded))
		} else if instruction.Type == ir.TypeNumber {
			out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s to double\n", instruction.Result, payloadLoaded))
		} else if instruction.Type == ir.TypeBool {
			out.WriteString(fmt.Sprintf("  %%%s = trunc i64 %%%s to i1\n", instruction.Result, payloadLoaded))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", instruction.Result, payloadLoaded))
		}
		return nil
	}

	switch {
	case instruction.Type == ir.TypeNumber || actualFieldType == ir.TypeNumber:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_get(ptr %s, i64 %d, ptr %%__slot_double)\n", status, ptrObj, instruction.FieldIndex))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		if instruction.Type == ir.TypeUnknown {
			numLoaded := fmt.Sprintf("num.loaded.%d", e.loadCounter)
			payloadVal := fmt.Sprintf("payload.%d", e.loadCounter)
			b0 := fmt.Sprintf("box.b0.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%__slot_double\n", numLoaded))
			out.WriteString(fmt.Sprintf("  %%%s = bitcast double %%%s to i64\n", payloadVal, numLoaded))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } zeroinitializer, i32 3, 0\n", b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, b0, payloadVal))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%__slot_double\n", instruction.Result))
		}
	case instruction.Type == ir.TypeBool || actualFieldType == ir.TypeBool:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_bool_get(ptr %s, i64 %d, ptr %%__slot_i32)\n", status, ptrObj, instruction.FieldIndex))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		if instruction.Type == ir.TypeUnknown {
			boolLoaded := fmt.Sprintf("bool.loaded.%d", e.loadCounter)
			payloadVal := fmt.Sprintf("payload.%d", e.loadCounter)
			b0 := fmt.Sprintf("box.b0.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%__slot_i32\n", boolLoaded))
			out.WriteString(fmt.Sprintf("  %%%s = zext i32 %%%s to i64\n", payloadVal, boolLoaded))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } zeroinitializer, i32 2, 0\n", b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, b0, payloadVal))
		} else {
			boolI32 := instruction.Result + ".i32"
			out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%__slot_i32\n", boolI32))
			out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, boolI32))
		}
	case instruction.Type == ir.TypeBigInt || actualFieldType == ir.TypeBigInt:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		slotBigInt := fmt.Sprintf("slot.bigint.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = alloca i64\n", slotBigInt))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_bigint_get(ptr %s, i64 %d, ptr %%%s)\n", status, ptrObj, instruction.FieldIndex, slotBigInt))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		if instruction.Type == ir.TypeUnknown {
			biLoaded := fmt.Sprintf("bi.loaded.%d", e.loadCounter)
			b0 := fmt.Sprintf("box.b0.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s\n", biLoaded, slotBigInt))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } zeroinitializer, i32 8, 0\n", b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, b0, biLoaded))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s\n", instruction.Result, slotBigInt))
		}
	default:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_ptr_get(ptr %s, i64 %d, ptr %%__slot_ptr)\n", status, ptrObj, instruction.FieldIndex))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		if instruction.Type == ir.TypeUnknown {
			ptrLoaded := fmt.Sprintf("ptr.loaded.%d", e.loadCounter)
			payloadVal := fmt.Sprintf("payload.%d", e.loadCounter)
			b0 := fmt.Sprintf("box.b0.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%__slot_ptr\n", ptrLoaded))
			out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", payloadVal, ptrLoaded))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } zeroinitializer, i32 4, 0\n", b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, b0, payloadVal))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result))
		}
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
	arg := instruction.Args[0]
	argType := e.types[arg]
	argVal := arg
	if slot, ok := e.varSlots[arg]; ok {
		loaded := fmt.Sprintf("%s.instanceof_load.%d", arg, e.loadCounter)
		e.loadCounter++
		if argType == ir.TypeUnknown {
			out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64 }, ptr %%%s\n", loaded, slot))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = load volatile ptr, ptr %%%s\n", loaded, slot))
		}
		argVal = loaded
	}
	ptrArg := "%" + argVal
	if argType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, argVal))
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
		ptrArg = "%" + ptrVar
	}
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_instanceof(ptr %s, ptr %s, ptr %%%s)\n", status, ptrArg, strGlobal, slot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	i32Val := instruction.Result + ".i32"
	out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%%s\n", i32Val, slot))
	out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Val))
	return nil
}

func (e *functionEmitter) emitObjectIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
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
				out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64 }, ptr %%%s\n", loaded0, slot0))
				arg0 = loaded0
			}
			arg1 := instruction.Args[1]
			if slot1, ok := e.varSlots[arg1]; ok {
				loaded1 := fmt.Sprintf("%s.is.loaded.%d", arg1, e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64 }, ptr %%%s\n", loaded1, slot1))
				arg1 = loaded1
			}
			tag0 := fmt.Sprintf("tag0.%d", e.loadCounter)
			tag1 := fmt.Sprintf("tag1.%d", e.loadCounter)
			payload0 := fmt.Sprintf("payload0.%d", e.loadCounter)
			payload1 := fmt.Sprintf("payload1.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag0, arg0))
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag1, arg1))
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payload0, arg0))
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payload1, arg1))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_is_unknown(i32 %%%s, i64 %%%s, i32 %%%s, i64 %%%s, ptr %%%s)\n", status, tag0, payload0, tag1, payload1, slot))
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
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadName, objVar)
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
				tagName := fmt.Sprintf("unknown.prop.tag.%d", e.loadCounter)
				payloadName := fmt.Sprintf("unknown.prop.payload.%d", e.loadCounter)
				e.loadCounter++
				fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagName, objArg)
				fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadName, objArg)
				slot := instruction.Result + ".slot"
				status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
				e.runtimeStatus++
				out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_unknown_number_property(i32 %%%s, i64 %%%s, ptr %%%s, ptr %%%s)\n", status, tagName, payloadName, propArg, slot))
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
				fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadName, objArg)
				fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrName, payloadName)
				ptrObj = "%" + ptrName
			}
			propertyArg := e.resolveArg(out, instruction.Args[1])
			tagSlot := fmt.Sprintf("%s.dynamic.tag.slot", instruction.Result)
			payloadSlot := fmt.Sprintf("%s.dynamic.payload.slot", instruction.Result)
			tagValue := fmt.Sprintf("%s.dynamic.tag", instruction.Result)
			payloadValue := fmt.Sprintf("%s.dynamic.payload", instruction.Result)
			status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			e.types[instruction.Result] = ir.TypeUnknown
			fmt.Fprintf(out, "  %%%s = alloca i32\n", tagSlot)
			fmt.Fprintf(out, "  %%%s = alloca i64\n", payloadSlot)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_unknown_get(ptr %s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, ptrObj, propertyArg, tagSlot, payloadSlot)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
			fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", tagValue, tagSlot)
			fmt.Fprintf(out, "  %%%s = load i64, ptr %%%s\n", payloadValue, payloadSlot)
			box0 := fmt.Sprintf("%s.dynamic.box0", instruction.Result)
			box1 := fmt.Sprintf("%s.dynamic.box1", instruction.Result)
			fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64 } undef, i32 %%%s, 0\n", box0, tagValue)
			fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", box1, box0)
			fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, box1, payloadValue)
			return nil
		}
		objArg := e.resolveArg(out, instruction.Args[0])
		objType := e.types[instruction.Args[0]]
		ptrObj := objArg
		if objType == ir.TypeUnknown {
			payloadName := fmt.Sprintf("dynamic.prop.payload.%d", e.loadCounter)
			ptrName := fmt.Sprintf("dynamic.prop.ptr.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadName, objArg)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrName, payloadName)
			ptrObj = ptrName
		}
		propertyArg := e.resolveArg(out, instruction.Args[1])
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		e.types[instruction.Result] = instruction.Type
		switch instruction.Type {
		case ir.TypeNumber:
			slot := instruction.Result + ".slot"
			out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
			if objType == ir.TypeUnknown {
				tagName := fmt.Sprintf("dynamic.prop.tag.%d", e.loadCounter)
				payloadName := fmt.Sprintf("dynamic.prop.value.%d", e.loadCounter)
				e.loadCounter++
				fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagName, objArg)
				fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadName, objArg)
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_unknown_number_property(i32 %%%s, i64 %%%s, ptr %%%s, ptr %%%s)\n", status, tagName, payloadName, propertyArg, slot)
			} else {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_property_number_get(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, ptrObj, propertyArg, slot)
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
				var argType ir.Type = ir.TypeObject
				if len(instruction.Args) > 0 {
					argVal = instruction.Args[0]
					if t, ok := e.types[argVal]; ok && t != "" {
						argType = t
					}
				}
				if argVal != "" {
					return e.emitBoxValue(out, argVal, argType, instruction.Result)
				}
				out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } zeroinitializer, i32 0, 0\n", instruction.Result))
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
