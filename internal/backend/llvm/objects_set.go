package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

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
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadVar, objArg))
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

	if instruction.FieldIndex >= 0 && (actualType == ir.TypeNumber || isPointerType(actualType)) && !((actualType == "ptr" || actualType == ir.TypePointer || actualType == ir.TypeVoid) && valueType == ir.TypeNumber) {
		id := e.labelCounter
		e.labelCounter++

		checkLabel := fmt.Sprintf("fset.check.%d", id)
		fastLabel := fmt.Sprintf("fset.fast.%d", id)
		slowLabel := fmt.Sprintf("fset.slow.%d", id)
		doneLabel := fmt.Sprintf("fset.done.%d", id)

		notNull := fmt.Sprintf("fset.not_null.%d", id)
		notUndef := fmt.Sprintf("fset.not_undef.%d", id)
		validPtr := fmt.Sprintf("fset.valid_ptr.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne ptr %s, null\n", notNull, ptrObj))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne ptr %s, @scriptgo_undefined_sentinel\n", notUndef, ptrObj))
		out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", validPtr, notNull, notUndef))
		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\x22branch_weights\x22, i32 10000, i32 1}\n", validPtr, checkLabel, slowLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", checkLabel))
		magicVal := fmt.Sprintf("fset.magic.%d", id)
		isMagic := fmt.Sprintf("fset.is_magic.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %s\n", magicVal, ptrObj))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq i64 %%%s, 6001091566378403156\n", isMagic, magicVal))

		fcPtr := fmt.Sprintf("fset.fc_ptr.%d", id)
		fcVal := fmt.Sprintf("fset.fc.%d", id)
		inBounds := fmt.Sprintf("fset.in_bounds.%d", id)
		condFast := fmt.Sprintf("fset.cond_fast.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %s, i64 8\n", fcPtr, ptrObj))
		out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s\n", fcVal, fcPtr))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ugt i64 %%%s, %d\n", inBounds, fcVal, instruction.FieldIndex))
		out.WriteString(fmt.Sprintf("  %%%s = and i1 %%%s, %%%s\n", condFast, isMagic, inBounds))
		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\x22branch_weights\x22, i32 10000, i32 1}\n", condFast, fastLabel, slowLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", fastLabel))
		fieldPtr := fmt.Sprintf("fset.field_ptr.%d", id)
		byteOffset := 40 + instruction.FieldIndex*8
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %s, i64 %d\n", fieldPtr, ptrObj, byteOffset))
		if actualType == ir.TypeNumber {
			out.WriteString(fmt.Sprintf("  store double %%%s, ptr %%%s\n", valArg, fieldPtr))
		} else {
			out.WriteString(fmt.Sprintf("  store ptr %%%s, ptr %%%s\n", valArg, fieldPtr))
		}
		out.WriteString(fmt.Sprintf("  br label %%%s\n", doneLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", slowLabel))
		slowStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if actualType == ir.TypeNumber {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_set(ptr %s, i64 %d, double %%%s)\n", slowStatus, ptrObj, instruction.FieldIndex, valArg))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_ptr_set(ptr %s, i64 %d, ptr %%%s)\n", slowStatus, ptrObj, instruction.FieldIndex, valArg))
		}
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", slowStatus))
		out.WriteString(fmt.Sprintf("  br label %%%s\n", doneLabel))

		out.WriteString(fmt.Sprintf("\n%s:\n", doneLabel))
		return nil
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
		valuePtr, err := e.emitCanonicalValuePointer(out, valArg, ir.TypeUnknown, fmt.Sprintf("object.field.%d", e.loadCounter))
		if err != nil {
			return err
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_unknown_set(ptr %s, i64 %d, ptr %s)\n", status, ptrObj, instruction.FieldIndex, valuePtr))
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
