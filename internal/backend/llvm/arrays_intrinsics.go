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
	case "__array.bounds_guard":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("array.bounds_guard requires array and bound arguments")
		}
		ptrArg := e.ensurePointerArg(out, instruction.Args[0])
		boundArg := e.resolveArg(out, instruction.Args[1])
		id := e.labelCounter
		e.labelCounter++
		boundI64 := fmt.Sprintf("guard.bound.i64.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i64\n", boundI64, boundArg))
		isNotNull := fmt.Sprintf("guard.not_null.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne ptr %%%s, null\n", isNotNull, ptrArg))
		checkLabel := fmt.Sprintf("guard.check.%d", id)
		failLabel := fmt.Sprintf("guard.fail.%d", id)
		passLabel := fmt.Sprintf("guard.pass.%d", id)
		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\"branch_weights\", i32 10000, i32 1}\n", isNotNull, checkLabel, failLabel))
		out.WriteString(fmt.Sprintf("\n%s:\n", checkLabel))
		lenPtr := fmt.Sprintf("guard.len.ptr.%d", id)
		lenVal := fmt.Sprintf("guard.len.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds i8, ptr %%%s, i64 8\n", lenPtr, ptrArg))
		out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s, !invariant.load !{}\n", lenVal, lenPtr))
		cmpZero := fmt.Sprintf("guard.zero.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp sle i64 %%%s, 0\n", cmpZero, boundI64))
		cmpLen := fmt.Sprintf("guard.cmp.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = icmp ule i64 %%%s, %%%s\n", cmpLen, boundI64, lenVal))
		cmpOk := fmt.Sprintf("guard.ok.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = or i1 %%%s, %%%s\n", cmpOk, cmpZero, cmpLen))
		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\"branch_weights\", i32 10000, i32 1}\n", cmpOk, passLabel, failLabel))
		out.WriteString(fmt.Sprintf("\n%s:\n", failLabel))
		out.WriteString("  call void @scriptgo_runtime_abort_if_failed(i32 -1)\n")
		out.WriteString("  unreachable\n")
		out.WriteString(fmt.Sprintf("\n%s:\n", passLabel))
		return nil
	case "__array.set_length", "__array.push", "__array.pop", "__array.shift", "__array.unshift", "__array.reverse", "__array.splice", "__array.fill", "__array.copyWithin", "__array.sort":
		return e.emitArrayMutationIntrinsic(out, instruction, arrayType)

	case "__array.slice", "__array.concat", "__array.join", "__array.toReversed", "__array.toSorted", "__array.with", "__array.toSpliced", "__array.toString", "__array.toLocaleString":
		return e.emitArrayCopyIntrinsic(out, instruction, arrayType)
	case "__array.indexOf", "__array.includes", "__array.at", "__array.find", "__array.some", "__array.every", "__array.findIndex", "__array.findLast", "__array.findLastIndex", "__array.lastIndexOf":
		return e.emitArraySearchIntrinsic(out, instruction, arrayType)

	case "__array.flatMap", "__array.flatMap_scalar", "__array.map", "__array.filter", "__array.forEach", "__array.reduce", "__array.reduceRight", "__array.values", "__array.flat", "__array.keys", "__array.entries":
		return e.emitArrayIterationIntrinsic(out, instruction, arrayType)

	default:
		return fmt.Errorf("unknown array intrinsic %q", instruction.Callee)
	}
}
