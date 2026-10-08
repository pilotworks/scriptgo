package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitArrayMutationIntrinsic(out *strings.Builder, instruction ir.Instruction, arrayType ir.Type) error {
	switch instruction.Callee {
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
	default:
		return fmt.Errorf("unknown array intrinsic %q", instruction.Callee)
	}
}
