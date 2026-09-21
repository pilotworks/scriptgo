package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

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
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadVar, objArg))
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
		slot := fmt.Sprintf("slot_value.%d", e.loadCounter)
		loaded := fmt.Sprintf("value.loaded.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = alloca { i32, i32, i64, i64 }\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_unknown_get(ptr %s, i64 %d, ptr %%%s)\n", status, ptrObj, instruction.FieldIndex, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		if instruction.Type == ir.TypeUnknown {
			out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", instruction.Result, slot))
		} else if instruction.Type == ir.TypeNumber {
			payloadLoaded := fmt.Sprintf("payload.loaded.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded, slot))
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadLoaded, loaded))
			out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s to double\n", instruction.Result, payloadLoaded))
		} else if instruction.Type == ir.TypeBool {
			payloadLoaded := fmt.Sprintf("payload.loaded.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded, slot))
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadLoaded, loaded))
			out.WriteString(fmt.Sprintf("  %%%s = trunc i64 %%%s to i1\n", instruction.Result, payloadLoaded))
		} else {
			payloadLoaded := fmt.Sprintf("payload.loaded.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded, slot))
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadLoaded, loaded))
			out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", instruction.Result, payloadLoaded))
		}
		return nil
	}

	if !instruction.DynamicField && instruction.FieldIndex >= 0 && actualFieldType != ir.TypeUnknown && instruction.Type != ir.TypeUnknown && (instruction.Type == ir.TypeNumber || actualFieldType == ir.TypeNumber || isPointerType(instruction.Type) || isPointerType(actualFieldType)) {
		isNum := instruction.Type == ir.TypeNumber || actualFieldType == ir.TypeNumber
		id := e.labelCounter
		e.labelCounter++

		if objArg == "this" {
			fieldPtr := fmt.Sprintf("fget.field_ptr.%d", id)
			byteOffset := 40 + instruction.FieldIndex*8
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %s, i64 %d\n", fieldPtr, ptrObj, byteOffset))
			typStr := "ptr"
			if isNum {
				typStr = "double"
				out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, fieldPtr))
			} else {
				rawPtr := fmt.Sprintf("fget.raw_ptr.%d", id)
				rawI64 := fmt.Sprintf("fget.raw_i64.%d", id)
				isNan := fmt.Sprintf("fget.is_nan.%d", id)
				out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", rawPtr, fieldPtr))
				out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", rawI64, rawPtr))
				out.WriteString(fmt.Sprintf("  %%%s = icmp eq i64 %%%s, 9221120237041090560\n", isNan, rawI64))
				out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, ptr @scriptgo_undefined_sentinel, ptr %%%s\n", instruction.Result, isNan, rawPtr))
			}
			if slot, hasSlot := e.varSlots[instruction.Result]; hasSlot {
				out.WriteString(fmt.Sprintf("  store%s %s %%%s, ptr %%%s\n", e.vol(), typStr, instruction.Result, slot))
			} else if e.localSSAs != nil {
				e.localSSAs[instruction.Result] = true
			}
			return nil
		}

		checkLabel := fmt.Sprintf("fget.check.%d", id)
		fastLabel := fmt.Sprintf("fget.fast.%d", id)
		slowLabel := fmt.Sprintf("fget.slow.%d", id)
		doneLabel := fmt.Sprintf("fget.done.%d", id)

		notNull := fmt.Sprintf("fget.not_null.%d", id)
		notUndef := fmt.Sprintf("fget.not_undef.%d", id)
		validPtr := fmt.Sprintf("fget.valid_ptr.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne ptr %s, null\n", notNull, ptrObj))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne ptr %s, @scriptgo_undefined_sentinel\n", notUndef, ptrObj))
		out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", validPtr, notNull, notUndef))
		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\x22branch_weights\x22, i32 10000, i32 1}\n", validPtr, checkLabel, slowLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", checkLabel))
		magicVal := fmt.Sprintf("fget.magic.%d", id)
		isMagic := fmt.Sprintf("fget.is_magic.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %s\n", magicVal, ptrObj))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i64 %%%s, 6001091566378403156\n", isMagic, magicVal))

		fcPtr := fmt.Sprintf("fget.fc_ptr.%d", id)
		fcVal := fmt.Sprintf("fget.fc.%d", id)
		inBounds := fmt.Sprintf("fget.in_bounds.%d", id)
		condFast := fmt.Sprintf("fget.cond_fast.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %s, i64 8\n", fcPtr, ptrObj))
		out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s\n", fcVal, fcPtr))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ugt i64 %%%s, %d\n", inBounds, fcVal, instruction.FieldIndex))
		out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", condFast, isMagic, inBounds))
		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\x22branch_weights\x22, i32 10000, i32 1}\n", condFast, fastLabel, slowLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", fastLabel))
		fieldPtr := fmt.Sprintf("fget.field_ptr.%d", id)
		byteOffset := 40 + instruction.FieldIndex*8
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %s, i64 %d\n", fieldPtr, ptrObj, byteOffset))
		fastVal := fmt.Sprintf("fget.fast.val.%d", id)
		if isNum {
			out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", fastVal, fieldPtr))
		} else {
			rawPtr := fmt.Sprintf("fget.raw_ptr.%d", id)
			rawI64 := fmt.Sprintf("fget.raw_i64.%d", id)
			isNan := fmt.Sprintf("fget.is_nan.%d", id)
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", rawPtr, fieldPtr))
			out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", rawI64, rawPtr))
			out.WriteString(fmt.Sprintf("  %%%s = icmp eq i64 %%%s, 9221120237041090560\n", isNan, rawI64))
			out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, ptr @scriptgo_undefined_sentinel, ptr %%%s\n", fastVal, isNan, rawPtr))
		}
		out.WriteString(fmt.Sprintf("  br label %%%s\n", doneLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", slowLabel))
		slowStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		slowVal := fmt.Sprintf("fget.slow.val.%d", id)
		if isNum {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_get(ptr %s, i64 %d, ptr %%__slot_double)\n", slowStatus, ptrObj, instruction.FieldIndex))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", slowStatus))
			out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%__slot_double\n", slowVal))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_ptr_get(ptr %s, i64 %d, ptr %%__slot_ptr)\n", slowStatus, ptrObj, instruction.FieldIndex))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", slowStatus))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%__slot_ptr\n", slowVal))
		}
		out.WriteString(fmt.Sprintf("  br label %%%s\n", doneLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", doneLabel))
		typStr := "ptr"
		if isNum {
			typStr = "double"
		}
		out.WriteString(fmt.Sprintf("  %%%s = phi %s [ %%%s, %%%s ], [ %%%s, %%%s ]\n", instruction.Result, typStr, fastVal, fastLabel, slowVal, slowLabel))
		if slot, hasSlot := e.varSlots[instruction.Result]; hasSlot {
			out.WriteString(fmt.Sprintf("  store%s %s %%%s, ptr %%%s\n", e.vol(), typStr, instruction.Result, slot))
		} else if e.localSSAs != nil {
			e.localSSAs[instruction.Result] = true
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
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 3, 0\n", b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, b0, payloadVal))
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
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 2, 0\n", b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, b0, payloadVal))
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
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 8, 0\n", b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, b0, biLoaded))
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
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 4, 0\n", b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, b0, payloadVal))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result))
		}
	}
	return nil
}
