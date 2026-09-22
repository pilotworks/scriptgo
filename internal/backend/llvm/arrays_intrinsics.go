package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitArrayIntrinsic(out *strings.Builder, instruction ir.Instruction, arrayType ir.Type) error {
	switch instruction.Callee {
	case "__array.new_length":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		elementSize, err := arrayElementSizeForTarget(instruction.Type, e.pointerSize())
		if err != nil {
			return err
		}
		lenI64 := fmt.Sprintf("%s.i64", instruction.Args[0])
		out.WriteString(fmt.Sprintf("  %%%s = fptoui double %%%s to i64\n", lenI64, instruction.Args[0]))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_new(i64 %%%s, i64 %d, ptr %%%s)\n", status, lenI64, elementSize, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		if tag := arrayElementTag(instruction.Type); tag > 0 {
			status = fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_set_tag(ptr %%%s, i64 %d)\n", status, instruction.Result, tag))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		}
		return nil
	case "__array.isArray":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("array.isArray has invalid signature")
		}
		arg := instruction.Args[0]
		argType := e.types[arg]
		if argType == ir.TypeUnknown || e.isParamUnknown(arg) {
			if slot, ok := e.varSlots[arg]; ok {
				loaded := fmt.Sprintf("%s.isarr.loaded.%d", arg, e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded, slot))
				arg = loaded
			}
			e.tempCounter++
			tagVar := fmt.Sprintf("isarray.tag.%d", e.tempCounter)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 0\n", tagVar, arg)
			fmt.Fprintf(out, "  %%%s = icmp eq i32 %%%s, 6\n", instruction.Result, tagVar)
			return nil
		}
		if strings.HasSuffix(string(argType), "[]") {
			fmt.Fprintf(out, "  %%%s = or i1 false, true\n", instruction.Result)
			return nil
		}
		fmt.Fprintf(out, "  %%%s = or i1 false, false\n", instruction.Result)
		return nil
	case "__array.length":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("array.length has invalid signature")
		}
		ptrArg := e.ensurePointerArg(out, instruction.Args[0])
		resultSlot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i64\n", resultSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_length(ptr %%%s, ptr %%%s)\n", status, ptrArg, resultSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.i64 = load i64, ptr %%%s\n", instruction.Result, resultSlot)
		fmt.Fprintf(out, "  %%%s = uitofp i64 %%%s.i64 to double\n", instruction.Result, instruction.Result)
		if e.integerVars != nil {
			e.integerVars[instruction.Result] = true
		}
		return nil
	case "__array.set_length":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("array.set_length has invalid signature")
		}
		ptrArg := e.ensurePointerArg(out, instruction.Args[0])
		lenArg := e.resolveArg(out, instruction.Args[1])
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_set_length(ptr %%%s, double %%%s)\n", status, ptrArg, lenArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil
	case "__array.push":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("array.push has invalid signature")
		}
		arrArg := e.resolveArg(out, instruction.Args[0])
		elemType := arrayElementType(arrayType)
		arg1 := e.resolveArg(out, instruction.Args[1])
		arg1Type := e.types[arg1]
		if arg1Type == "" {
			arg1Type = e.types[instruction.Args[1]]
		}
		if arg1Type == "" {
			for _, g := range e.module.Globals {
				if g.Name == instruction.Args[1] {
					arg1Type = g.Type
					break
				}
			}
		}
		if arg1Type == ir.TypeUnknown && elemType != ir.TypeUnknown {
			e.tempCounter++
			payloadName := fmt.Sprintf("push.unbox.payload.%d", e.tempCounter)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadName, arg1)
			paramType := llvmType(elemType)
			switch paramType {
			case "double":
				e.tempCounter++
				valName := fmt.Sprintf("push.unbox.dbl.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = bitcast i64 %%%s to double\n", valName, payloadName)
				arg1 = valName
			case "i1":
				e.tempCounter++
				valName := fmt.Sprintf("push.unbox.bool.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = trunc i64 %%%s to i1\n", valName, payloadName)
				arg1 = valName
			case "i64":
				arg1 = payloadName
			default:
				e.tempCounter++
				valName := fmt.Sprintf("push.unbox.ptr.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to %s\n", valName, payloadName, paramType)
				arg1 = valName
			}
		} else if elemType == ir.TypeUnknown && arg1Type != ir.TypeUnknown {
			e.tempCounter++
			boxedName := fmt.Sprintf("push.box.%d", e.tempCounter)
			var tag int
			switch arg1Type {
			case ir.TypeNumber:
				tag = 3
				payload := fmt.Sprintf("box.payload.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = bitcast double %%%s to i64\n", payload, arg1)
				fmt.Fprintf(out, "  %%%s.0 = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 %d, 0\n", boxedName, tag)
				fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64, i64 } %%%s.0, i64 %%%s, 2\n", boxedName, boxedName, payload)
			case ir.TypeString:
				tag = 4
				payload := fmt.Sprintf("box.payload.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = ptrtoint ptr %%%s to i64\n", payload, arg1)
				fmt.Fprintf(out, "  %%%s.0 = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 %d, 0\n", boxedName, tag)
				fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64, i64 } %%%s.0, i64 %%%s, 2\n", boxedName, boxedName, payload)
			case ir.TypeBool:
				tag = 2
				payload := fmt.Sprintf("box.payload.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i64\n", payload, arg1)
				fmt.Fprintf(out, "  %%%s.0 = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 %d, 0\n", boxedName, tag)
				fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64, i64 } %%%s.0, i64 %%%s, 2\n", boxedName, boxedName, payload)
			default:
				tag = 5
				payload := fmt.Sprintf("box.payload.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = ptrtoint ptr %%%s to i64\n", payload, arg1)
				fmt.Fprintf(out, "  %%%s.0 = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 %d, 0\n", boxedName, tag)
				fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64, i64 } %%%s.0, i64 %%%s, 2\n", boxedName, boxedName, payload)
			}
			arg1 = boxedName
		}
		elemLLVMType := arrayElementLLVMType(arrayType)
		valSlot := fmt.Sprintf("%s.push.val.%d", instruction.Args[0], e.runtimeStatus)
		fmt.Fprintf(out, "  %%%s = alloca %s\n", valSlot, elemLLVMType)
		fmt.Fprintf(out, "  store %s %%%s, ptr %%%s\n", elemLLVMType, arg1, valSlot)
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_push(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, arrArg, valSlot, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.pop":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("array.pop has invalid signature")
		}
		arrArg := e.resolveArg(out, instruction.Args[0])
		elemLLVMType := arrayElementLLVMType(arrayType)
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca %s\n", resSlot, elemLLVMType)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_pop(ptr %%%s, ptr %%%s)\n", status, arrArg, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load %s, ptr %%%s\n", instruction.Result, elemLLVMType, resSlot)
		return nil
	case "__array.slice":
		if (len(instruction.Args) < 1 || len(instruction.Args) > 3) || (!strings.HasSuffix(string(instruction.Type), "[]") && instruction.Type != ir.TypeNumberArray && instruction.Type != ir.TypeStringArray && instruction.Type != ir.TypeBoolArray && instruction.Type != ir.TypeBigIntArray && instruction.Type != ir.TypeUnknownArray) {
			return fmt.Errorf("array.slice has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		startArg := "0.0"
		if len(instruction.Args) >= 2 {
			startArg = "%" + e.resolveArg(out, instruction.Args[1])
		}
		endArg := "-1.0"
		if len(instruction.Args) == 3 {
			endArg = "%" + e.resolveArg(out, instruction.Args[2])
		}
		targetElemSize := 8
		if instruction.Type == ir.TypeBoolArray || instruction.Type == "bool[]" || instruction.Type == "boolean[]" {
			targetElemSize = 1
		} else if instruction.Type == ir.TypeUnknownArray || instruction.Type == "unknown[]" || instruction.Type == "any[]" {
			targetElemSize = 24
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_slice_with_size(ptr %%%s, double %s, double %s, i64 %d, ptr %%%s)\n", status, arrArg, startArg, endArg, targetElemSize, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.indexOf":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("array.indexOf has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		targetArg := e.resolveArg(out, instruction.Args[1])
		fromArg := "0.0"
		if len(instruction.Args) == 3 {
			fromArg = "%" + e.resolveArg(out, instruction.Args[2])
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		elemLLVMType := arrayElementLLVMType(arrayType)
		if elemLLVMType == "ptr" {
			if arrayType == ir.TypeStringArray {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_index_of_string(ptr %%%s, ptr %%%s, double %s, ptr %%%s)\n", status, arrArg, targetArg, fromArg, resSlot)
			} else {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_index_of_ptr(ptr %%%s, ptr %%%s, double %s, ptr %%%s)\n", status, arrArg, targetArg, fromArg, resSlot)
			}
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_index_of_number(ptr %%%s, double %%%s, double %s, ptr %%%s)\n", status, arrArg, targetArg, fromArg, resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.includes":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("array.includes has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		targetArg := e.resolveArg(out, instruction.Args[1])
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		incElemLLVMType := arrayElementLLVMType(arrayType)
		if incElemLLVMType == "ptr" {
			if arrayType == ir.TypeStringArray {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_includes_string(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, arrArg, targetArg, resSlot)
			} else {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_includes_ptr(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, arrArg, targetArg, resSlot)
			}
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_includes_number(ptr %%%s, double %%%s, ptr %%%s)\n", status, arrArg, targetArg, resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, resSlot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
		return nil
	case "__array.at":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("array.at has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		idxArg := e.resolveArg(out, instruction.Args[1])
		elemLLVMType := arrayElementLLVMType(arrayType)
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca %s\n", resSlot, elemLLVMType)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_at(ptr %%%s, double %%%s, ptr %%%s)\n", status, arrArg, idxArg, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load %s, ptr %%%s\n", instruction.Result, elemLLVMType, resSlot)
		return nil
	case "__array.shift":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("array.shift has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		elemLLVMType := arrayElementLLVMType(arrayType)
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca %s\n", resSlot, elemLLVMType)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_shift(ptr %%%s, ptr %%%s)\n", status, arrArg, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load %s, ptr %%%s\n", instruction.Result, elemLLVMType, resSlot)
		return nil
	case "__array.unshift":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("array.unshift has invalid signature")
		}
		arrArg := e.resolveArg(out, instruction.Args[0])
		elemType := arrayElementType(arrayType)
		arg1 := e.resolveArg(out, instruction.Args[1])
		arg1Type := e.types[arg1]
		if arg1Type == "" {
			arg1Type = e.types[instruction.Args[1]]
		}
		if arg1Type == "" {
			for _, g := range e.module.Globals {
				if g.Name == instruction.Args[1] {
					arg1Type = g.Type
					break
				}
			}
		}
		if arg1Type == ir.TypeUnknown && elemType != ir.TypeUnknown {
			e.tempCounter++
			payloadName := fmt.Sprintf("unshift.unbox.payload.%d", e.tempCounter)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadName, arg1)
			paramType := llvmType(elemType)
			switch paramType {
			case "double":
				e.tempCounter++
				valName := fmt.Sprintf("unshift.unbox.dbl.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = bitcast i64 %%%s to double\n", valName, payloadName)
				arg1 = valName
			case "i1":
				e.tempCounter++
				valName := fmt.Sprintf("unshift.unbox.bool.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = trunc i64 %%%s to i1\n", valName, payloadName)
				arg1 = valName
			case "i64":
				arg1 = payloadName
			default:
				e.tempCounter++
				valName := fmt.Sprintf("unshift.unbox.ptr.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to %s\n", valName, payloadName, paramType)
				arg1 = valName
			}
		} else if elemType == ir.TypeUnknown && arg1Type != ir.TypeUnknown {
			e.tempCounter++
			boxedName := fmt.Sprintf("unshift.box.%d", e.tempCounter)
			var tag int
			switch arg1Type {
			case ir.TypeNumber:
				tag = 3
				payload := fmt.Sprintf("box.payload.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = bitcast double %%%s to i64\n", payload, arg1)
				fmt.Fprintf(out, "  %%%s.0 = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 %d, 0\n", boxedName, tag)
				fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64, i64 } %%%s.0, i64 %%%s, 2\n", boxedName, boxedName, payload)
			case ir.TypeString:
				tag = 4
				payload := fmt.Sprintf("box.payload.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = ptrtoint ptr %%%s to i64\n", payload, arg1)
				fmt.Fprintf(out, "  %%%s.0 = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 %d, 0\n", boxedName, tag)
				fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64, i64 } %%%s.0, i64 %%%s, 2\n", boxedName, boxedName, payload)
			case ir.TypeBool:
				tag = 2
				payload := fmt.Sprintf("box.payload.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i64\n", payload, arg1)
				fmt.Fprintf(out, "  %%%s.0 = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 %d, 0\n", boxedName, tag)
				fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64, i64 } %%%s.0, i64 %%%s, 2\n", boxedName, boxedName, payload)
			default:
				tag = 5
				payload := fmt.Sprintf("box.payload.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = ptrtoint ptr %%%s to i64\n", payload, arg1)
				fmt.Fprintf(out, "  %%%s.0 = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 %d, 0\n", boxedName, tag)
				fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64, i64 } %%%s.0, i64 %%%s, 2\n", boxedName, boxedName, payload)
			}
			arg1 = boxedName
		}
		elemLLVMType := arrayElementLLVMType(arrayType)
		valSlot := fmt.Sprintf("%s.unshift.val.%d", instruction.Args[0], e.runtimeStatus)
		fmt.Fprintf(out, "  %%%s = alloca %s\n", valSlot, elemLLVMType)
		fmt.Fprintf(out, "  store %s %%%s, ptr %%%s\n", elemLLVMType, arg1, valSlot)
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca double\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_unshift(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, arrArg, valSlot, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.reverse":
		if len(instruction.Args) != 1 || instruction.Type != arrayType {
			return fmt.Errorf("array.reverse has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_reverse(ptr %%%s, ptr %%%s)\n", status, arrArg, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.concat":
		if len(instruction.Args) != 2 || instruction.Type != arrayType {
			return fmt.Errorf("array.concat has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		otherArg := e.ensurePointerArg(out, instruction.Args[1])
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_concat(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, arrArg, otherArg, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.splice":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != arrayType {
			return fmt.Errorf("array.splice has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		startArg := "%" + e.resolveArg(out, instruction.Args[1])
		dcArg := "1000000000.0"
		if len(instruction.Args) == 3 {
			dcArg = "%" + e.resolveArg(out, instruction.Args[2])
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_splice(ptr %%%s, double %s, double %s, ptr %%%s)\n", status, arrArg, startArg, dcArg, resSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.join":
		if (len(instruction.Args) != 1 && len(instruction.Args) != 2) || instruction.Type != ir.TypeString {
			return fmt.Errorf("array.join has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		sepArg := "null"
		if len(instruction.Args) == 2 {
			sepArg = "%" + e.resolveArg(out, instruction.Args[1])
		}
		resSlot := instruction.Result + ".slot"
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", resSlot)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		switch arrayType {
		case ir.TypeNumberArray:
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_number(ptr %%%s, ptr %s, ptr %%%s)\n", status, arrArg, sepArg, resSlot)
		case ir.TypeStringArray:
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_string(ptr %%%s, ptr %s, ptr %%%s)\n", status, arrArg, sepArg, resSlot)
		case ir.TypeBigIntArray:
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_bigint(ptr %%%s, ptr %s, ptr %%%s)\n", status, arrArg, sepArg, resSlot)
		default:
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_unknown(ptr %%%s, ptr %s, ptr %%%s)\n", status, arrArg, sepArg, resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.flatMap", "__array.flatMap_scalar":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		fnName := "scriptgo_array_flat_map_number"
		if instruction.Callee == "__array.flatMap_scalar" {
			fnName = "scriptgo_array_flat_map_number_scalar"
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.map":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		var fnName string
		retElemType := arrayElementType(instruction.Type)
		inElemType := arrayElementType(arrayType)
		if retElemType == ir.TypeNumber {
			if inElemType == ir.TypeString {
				fnName = "scriptgo_array_map_number_from_string"
			} else if inElemType != ir.TypeNumber {
				fnName = "scriptgo_array_map_number_from_ptr"
			} else {
				fnName = "scriptgo_array_map_number"
			}
		} else if retElemType == ir.TypeString {
			if inElemType == ir.TypeNumber {
				fnName = "scriptgo_array_map_string_from_number"
			} else if inElemType != ir.TypeString {
				fnName = "scriptgo_array_map_string_from_ptr"
			} else {
				fnName = "scriptgo_array_map_string"
			}
		} else {
			if inElemType == ir.TypeNumber {
				fnName = "scriptgo_array_map_ptr_from_number"
			} else if inElemType == ir.TypeString {
				fnName = "scriptgo_array_map_ptr_from_string"
			} else {
				fnName = "scriptgo_array_map_ptr"
			}
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.filter":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		fnName := "scriptgo_array_filter_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_filter_string"
		} else if arrayType != ir.TypeNumberArray && arrayType != "number[]" {
			fnName = "scriptgo_array_filter_ptr"
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.forEach":
		fnName := "scriptgo_array_for_each_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_for_each_string"
		} else if arrayType != ir.TypeNumberArray && arrayType != "number[]" {
			fnName = "scriptgo_array_for_each_ptr"
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	case "__array.reduce":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_reduce_number(ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.find":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_find_number(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.some":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_some_number(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s.i32 = load i32, ptr %%%s\n", instruction.Result, slot))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s.i32, 0\n", instruction.Result, instruction.Result))
		return nil
	case "__array.every":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_every_number(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s.i32 = load i32, ptr %%%s\n", instruction.Result, slot))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s.i32, 0\n", instruction.Result, instruction.Result))
		return nil
	case "__array.findIndex":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fnName := "scriptgo_array_find_index_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_find_index_string"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.fill":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		startVal := "0.000000e+00"
		hasStart := 0
		if len(instruction.Args) > 2 {
			startVal = fmt.Sprintf("%%%s", instruction.Args[2])
			hasStart = 1
		}
		endVal := "0.000000e+00"
		hasEnd := 0
		if len(instruction.Args) > 3 {
			endVal = fmt.Sprintf("%%%s", instruction.Args[3])
			hasEnd = 1
		}
		if arrayType == ir.TypeStringArray {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_fill_string(ptr %%%s, ptr %%%s, double %s, double %s, i32 %d, i32 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], startVal, endVal, hasStart, hasEnd, slot))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_fill_number(ptr %%%s, double %%%s, double %s, double %s, i32 %d, i32 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], startVal, endVal, hasStart, hasEnd, slot))
		}
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.toReversed":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_to_reversed(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.toSorted":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fnName := "scriptgo_array_to_sorted_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_to_sorted_string"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.findLast":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_find_last_number(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.findLastIndex":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fnName := "scriptgo_array_find_last_index_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_find_last_index_string"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.lastIndexOf":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("array.lastIndexOf has invalid signature")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		targetArg := e.resolveArg(out, instruction.Args[1])
		resSlot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", resSlot))
		fromArg := "-1.000000e+00"
		if len(instruction.Args) > 2 {
			fromArg = fmt.Sprintf("%%%s", e.resolveArg(out, instruction.Args[2]))
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if arrayType == ir.TypeStringArray {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_last_index_of_string(ptr %%%s, ptr %%%s, double %s, ptr %%%s)\n", status, arrArg, targetArg, fromArg, resSlot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_last_index_of_number(ptr %%%s, double %%%s, double %s, ptr %%%s)\n", status, arrArg, targetArg, fromArg, resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.copyWithin":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		startVal := "0.000000e+00"
		hasStart := 0
		if len(instruction.Args) > 2 {
			startVal = fmt.Sprintf("%%%s", instruction.Args[2])
			hasStart = 1
		}
		endVal := "0.000000e+00"
		hasEnd := 0
		if len(instruction.Args) > 3 {
			endVal = fmt.Sprintf("%%%s", instruction.Args[3])
			hasEnd = 1
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_copy_within(ptr %%%s, double %%%s, double %s, double %s, i32 %d, i32 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], startVal, endVal, hasStart, hasEnd, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.with":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fnName := "scriptgo_array_with_number"
		if arrayType == ir.TypeStringArray {
			fnName = "scriptgo_array_with_string"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, double %%%s, %s %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], llvmType(arrayElementType(arrayType)), instruction.Args[2], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.toSpliced":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		delCount := "0.000000e+00"
		hasDel := 0
		if len(instruction.Args) > 2 {
			delCount = fmt.Sprintf("%%%s", instruction.Args[2])
			hasDel = 1
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_to_spliced(ptr %%%s, double %%%s, double %s, i32 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], delCount, hasDel, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.sort":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		elemLLVMType := arrayElementLLVMType(arrayType)
		if len(instruction.Args) > 1 {
			fnName := "scriptgo_array_sort_closure_number"
			if elemLLVMType == "ptr" {
				if arrayType == ir.TypeStringArray {
					fnName = "scriptgo_array_sort_closure_string"
				} else {
					fnName = "scriptgo_array_sort_closure_ptr"
				}
			}
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], instruction.Args[1], slot))
		} else {
			fnName := "scriptgo_array_sort_number"
			if arrayType == ir.TypeStringArray {
				fnName = "scriptgo_array_sort_string"
			}
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, ptr %%%s)\n", status, fnName, instruction.Args[0], slot))
		}
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.reduceRight":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_reduce_right_number(ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.toString", "__array.toLocaleString":
		resSlot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", resSlot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if arrayType == ir.TypeStringArray {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_string(ptr %%%s, ptr null, ptr %%%s)\n", status, instruction.Args[0], resSlot)
		} else if arrayType == ir.TypeUnknownArray {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_unknown(ptr %%%s, ptr null, ptr %%%s)\n", status, instruction.Args[0], resSlot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_join_number(ptr %%%s, ptr null, ptr %%%s)\n", status, instruction.Args[0], resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, resSlot)
		return nil
	case "__array.values", "__array.flat":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_slice(ptr %%%s, double 0.0, double -1.0, ptr %%%s)\n", status, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.keys":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_keys(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__array.entries":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_entries(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	default:
		return fmt.Errorf("unknown array intrinsic %q", instruction.Callee)
	}
}
