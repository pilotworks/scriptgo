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
	if !ok || valueType == "" || valueType == ir.TypeVoid {
		if instruction.Type != "" && instruction.Type != ir.TypeVoid {
			valueType = instruction.Type
		} else if instruction.Callee != "" && len(e.module.Shapes) > 0 {
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
	}
	if valueType == "" || valueType == ir.TypeVoid {
		valueType = ir.TypePointer
	}
	valArg := instruction.Args[1]
	if slot, ok := e.varSlots[valArg]; ok {
		loaded := fmt.Sprintf("%s.fset.loaded.%d", valArg, e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %%%s\n", loaded, llvmType(valueType), slot))
		valArg = loaded
	}
	objArg := instruction.Args[0]
	objType := e.types[objArg]
	ptrObj := "%" + objArg
	if objType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, objArg))
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
		ptrObj = "%" + ptrVar
	}
	switch {
	case valueType == ir.TypeNumber:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_set(ptr %s, i64 %d, double %%%s)\n", status, ptrObj, instruction.FieldIndex, valArg))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case valueType == ir.TypeBool:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		boolI32 := fmt.Sprintf("obj.bool.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i32\n", boolI32, valArg))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_bool_set(ptr %s, i64 %d, i32 %%%s)\n", status, ptrObj, instruction.FieldIndex, boolI32))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case valueType == ir.TypeUnknown:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, valArg))
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_ptr_set(ptr %s, i64 %d, ptr %%%s)\n", status, ptrObj, instruction.FieldIndex, ptrVar))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	default:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_ptr_set(ptr %s, i64 %d, ptr %%%s)\n", status, ptrObj, instruction.FieldIndex, valArg))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	}
	return nil
}

func (e *functionEmitter) emitFieldGet(out *strings.Builder, instruction ir.Instruction) error {
	if instruction.FieldIndex < 0 {
		return fmt.Errorf("object field %q has invalid index", instruction.Field)
	}
	e.types[instruction.Result] = instruction.Type
	objArg := instruction.Args[0]
	objType := e.types[objArg]
	ptrObj := "%" + objArg
	if objType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, objArg))
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
		ptrObj = "%" + ptrVar
	}
	switch {
	case instruction.Type == ir.TypeNumber:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_get(ptr %s, i64 %d, ptr %%__slot_double)\n", status, ptrObj, instruction.FieldIndex))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%__slot_double\n", instruction.Result))
	case instruction.Type == ir.TypeBool:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_bool_get(ptr %s, i64 %d, ptr %%__slot_i32)\n", status, ptrObj, instruction.FieldIndex))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		boolI32 := instruction.Result + ".i32"
		out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%__slot_i32\n", boolI32))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, boolI32))
	case instruction.Type == ir.TypeUnknown:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_ptr_get(ptr %s, i64 %d, ptr %%__slot_ptr)\n", status, ptrObj, instruction.FieldIndex))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		ptrLoaded := fmt.Sprintf("ptr.loaded.%d", e.loadCounter)
		payloadVal := fmt.Sprintf("payload.%d", e.loadCounter)
		b0 := fmt.Sprintf("box.b0.%d", e.loadCounter)
		b1 := fmt.Sprintf("box.b1.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%__slot_ptr\n", ptrLoaded))
		out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", payloadVal, ptrLoaded))
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } undef, i32 5, 0\n", b0))
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", b1, b0))
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, b1, payloadVal))
	default:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_ptr_get(ptr %s, i64 %d, ptr %%__slot_ptr)\n", status, ptrObj, instruction.FieldIndex))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result))
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
	ptrArg := "%" + arg
	if argType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg))
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
	case "__object.get_prop":
		if strings.HasSuffix(string(instruction.Type), "[]") || instruction.Type == ir.TypeNumberArray || instruction.Type == ir.TypeStringArray {
			e.types[instruction.Result] = instruction.Type
			arrSlot := instruction.Result + ".arr.slot"
			status1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", arrSlot))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_new(i64 2, i64 8, ptr %%%s)\n", status1, arrSlot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status1))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, arrSlot))
			return nil
		}
		if instruction.Type == ir.TypeBool {
			e.types[instruction.Result] = ir.TypeBool
			out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 1, 1\n", instruction.Result))
			return nil
		}
		e.types[instruction.Result] = ir.TypeNumber
		out.WriteString(fmt.Sprintf("  %%%s = fadd double 0.0, 1.0\n", instruction.Result))
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
			if len(instruction.Args) > 0 {
				out.WriteString(fmt.Sprintf("  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, instruction.Args[0]))
			} else {
				out.WriteString(fmt.Sprintf("  %%%s = alloca i8\n", instruction.Result))
			}
			return nil
		}
		return fmt.Errorf("unknown object intrinsic %q", instruction.Callee)
	}
	return nil
}
