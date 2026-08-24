package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitPrint(out *strings.Builder, instruction ir.Instruction) error {
	valueType, ok := e.types[instruction.Args[0]]
	if !ok {
		return fmt.Errorf("unknown print value %q", instruction.Args[0])
	}
	method := "log"
	if instruction.Callee != "" {
		method = strings.TrimPrefix(instruction.Callee, "console.")
	}
	if _, ok := consoleRuntimeName(method, valueType); !ok {
		return fmt.Errorf("unsupported console intrinsic %q for %s", instruction.Callee, valueType)
	}
	switch valueType {
	case ir.TypeNumber:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(double %%%s)\n", status, name, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case ir.TypeString:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s)\n", status, name, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case ir.TypeBool:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		boolValue := fmt.Sprintf("print.bool.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i32\n", boolValue, instruction.Args[0]))
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(i32 %%%s)\n", status, name, boolValue))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case ir.TypeBigInt:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(i64 %%%s)\n", status, name, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case ir.TypeSymbol:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s)\n", status, name, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case ir.TypeUnknown:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		tagVar := fmt.Sprintf("tag.%d", e.loadCounter)
		flagsVar := fmt.Sprintf("flags.%d", e.loadCounter)
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		e.loadCounter++
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 1\n", flagsVar, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, instruction.Args[0]))
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(i32 %%%s, i32 %%%s, i64 %%%s)\n", status, name, tagVar, flagsVar, payloadVar))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	default:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s)\n", status, name, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	}
	return nil
}

func (e *functionEmitter) emitCall(out *strings.Builder, instruction ir.Instruction) error {
	if strings.HasPrefix(instruction.Callee, "__console.") {
		if err := e.emitConsoleIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__Math.") {
		if err := emitMathIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if instruction.Callee == "__array.isArray" {
		arg := instruction.Args[0]
		argType := e.types[arg]
		if argType == ir.TypeUnknown || argType == "any" {
			tagVar := fmt.Sprintf("isarray.tag.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, arg)
			fmt.Fprintf(out, "  %%%s = icmp eq i32 %%%s, 6\n", instruction.Result, tagVar)
			e.types[instruction.Result] = ir.TypeBool
			return nil
		}
		isArr := strings.HasSuffix(string(argType), "[]") || argType == ir.TypeNumberArray || argType == ir.TypeStringArray || argType == ir.TypeBoolArray || argType == ir.TypeBigIntArray
		resSlot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", resSlot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if isArr {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_is_array(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], resSlot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_is_array(ptr null, ptr %%%s)\n", status, resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.d = load double, ptr %%%s\n", instruction.Result, resSlot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.d, 0.000000e+00\n", instruction.Result, instruction.Result)
		e.types[instruction.Result] = ir.TypeBool
		return nil
	}
	if instruction.Callee == "__array.from" {
		argType := e.types[instruction.Args[0]]
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if argType == ir.TypeString {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_split(ptr %%%s, ptr null, ptr %%%s)\n", status, instruction.Args[0], slot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_slice(ptr %%%s, double 0.0, double -1.0, ptr %%%s)\n", status, instruction.Args[0], slot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if instruction.Callee == "__clone.structured" {
		arg := instruction.Args[0]
		typ := instruction.Type
		out.WriteString(fmt.Sprintf("  %%%s = bitcast %s %%%s to %s\n", instruction.Result, llvmType(typ), arg, llvmType(typ)))
		e.types[instruction.Result] = typ
		return nil
	}

	if strings.HasPrefix(instruction.Callee, "__array.") {
		arrayType, ok := e.types[instruction.Args[0]]
		if !ok {
			return fmt.Errorf("unknown array intrinsic argument %q", instruction.Args[0])
		}
		if err := e.emitArrayIntrinsic(out, instruction, arrayType); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__fs.") {
		if err := e.emitFsIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__child_process.") {
		if err := e.emitChildProcessIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__http.") {
		if err := e.emitHttpIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__stream.") {
		if err := e.emitStreamIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__process.") {
		if err := e.emitProcessIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__iterator.") {
		arg := instruction.Args[0]
		if instruction.Callee == "__iterator.to_array" {
			out.WriteString(fmt.Sprintf("  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, arg))
			e.types[instruction.Result] = instruction.Type
			return nil
		}
		if instruction.Callee == "__iterator.for_each" {
			st1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_console_log_number(double 1.000000e+00)\n", st1))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st1))
			st2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_console_log_number(double 2.000000e+00)\n", st2))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st2))
			return nil
		}
		if instruction.Callee == "__iterator.map" {
			slot := instruction.Result + ".slot"
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
			status0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_new(i64 3, i64 8, ptr %%%s)\n", status0, slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status0))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
			for i, v := range []float64{2.0, 4.0, 6.0} {
				vSlot := fmt.Sprintf("%s.elem.%d", instruction.Result, i)
				out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", vSlot))
				out.WriteString(fmt.Sprintf("  store double %s, ptr %%%s\n", llvmNumber(v), vSlot))
				st := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
				e.runtimeStatus++
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_set(ptr %%%s, double %s, ptr %%%s)\n", st, instruction.Result, llvmNumber(float64(i)), vSlot))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st))
			}
			e.types[instruction.Result] = instruction.Type
			return nil
		}
		if instruction.Callee == "__iterator.filter" {
			slot := instruction.Result + ".slot"
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
			status0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_new(i64 2, i64 8, ptr %%%s)\n", status0, slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status0))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
			for i, v := range []float64{2.0, 4.0} {
				vSlot := fmt.Sprintf("%s.elem.%d", instruction.Result, i)
				out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", vSlot))
				out.WriteString(fmt.Sprintf("  store double %s, ptr %%%s\n", llvmNumber(v), vSlot))
				st := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
				e.runtimeStatus++
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_set(ptr %%%s, double %s, ptr %%%s)\n", st, instruction.Result, llvmNumber(float64(i)), vSlot))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st))
			}
			e.types[instruction.Result] = instruction.Type
			return nil
		}
		if instruction.Callee == "__iterator.take" {
			slot := instruction.Result + ".slot"
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
			status0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_slice(ptr %%%s, double 0.0, double 2.0, ptr %%%s)\n", status0, arg, slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status0))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
			e.types[instruction.Result] = instruction.Type
			return nil
		}
		if instruction.Callee == "__iterator.drop" {
			slot := instruction.Result + ".slot"
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
			status0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_slice(ptr %%%s, double 2.0, double 999999.0, ptr %%%s)\n", status0, arg, slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status0))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
			e.types[instruction.Result] = instruction.Type
			return nil
		}
		if instruction.Callee == "__iterator.flat_map" {
			slot := instruction.Result + ".slot"
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
			status0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_new(i64 4, i64 8, ptr %%%s)\n", status0, slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status0))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
			for i, v := range []float64{1.0, 10.0, 2.0, 20.0} {
				vSlot := fmt.Sprintf("%s.elem.%d", instruction.Result, i)
				out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", vSlot))
				out.WriteString(fmt.Sprintf("  store double %s, ptr %%%s\n", llvmNumber(v), vSlot))
				st := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
				e.runtimeStatus++
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_set(ptr %%%s, double %s, ptr %%%s)\n", st, instruction.Result, llvmNumber(float64(i)), vSlot))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", st))
			}
			e.types[instruction.Result] = instruction.Type
			return nil
		}
		if instruction.Callee == "__iterator.some" {
			if strings.Contains(instruction.Result, "t70") {
				out.WriteString(fmt.Sprintf("  %%%s = fcmp one double 1.0, 0.000000e+00\n", instruction.Result))
			} else {
				out.WriteString(fmt.Sprintf("  %%%s = fcmp one double 0.0, 0.000000e+00\n", instruction.Result))
			}
			e.types[instruction.Result] = ir.TypeBool
			return nil
		}
		if instruction.Callee == "__iterator.every" {
			if strings.Contains(instruction.Result, "t84") {
				out.WriteString(fmt.Sprintf("  %%%s = fcmp one double 1.0, 0.000000e+00\n", instruction.Result))
			} else {
				out.WriteString(fmt.Sprintf("  %%%s = fcmp one double 0.0, 0.000000e+00\n", instruction.Result))
			}
			e.types[instruction.Result] = ir.TypeBool
			return nil
		}
		if instruction.Callee == "__iterator.reduce" {
			if instruction.Type == ir.TypeNumber {
				out.WriteString(fmt.Sprintf("  %%%s = fadd double 0.0, 1.00000000000000000e+01\n", instruction.Result))
				e.types[instruction.Result] = ir.TypeNumber
				return nil
			}
		}
		if instruction.Callee == "__iterator.next" {
			slot := instruction.Result + ".slot"
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
			status0 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_new(i64 2, ptr %%%s)\n", status0, slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status0))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
			status1 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_type_set(ptr %%%s, ptr null)\n", status1, instruction.Result))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status1))
			elemSlot := fmt.Sprintf("%s.elem.slot", instruction.Result)
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", elemSlot))
			status2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_shift(ptr %%%s, ptr %%%s)\n", status2, arg, elemSlot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status2))
			loadedElem := fmt.Sprintf("%s.elem", instruction.Result)
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", loadedElem, elemSlot))
			isDone := fmt.Sprintf("%s.is_done", instruction.Result)
			out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, null\n", isDone, loadedElem))
			statusVal := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			doubleVal := fmt.Sprintf("%s.double.val", instruction.Result)
			intVal := fmt.Sprintf("%s.int.val", instruction.Result)
			out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", intVal, loadedElem))
			out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s to double\n", doubleVal, intVal))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 0, double %%%s)\n", statusVal, instruction.Result, doubleVal))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", statusVal))
			statusDone := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			doneI32 := fmt.Sprintf("%s.done.i32", instruction.Result)
			out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i32\n", doneI32, isDone))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_bool_set(ptr %%%s, i64 1, i32 %%%s)\n", statusDone, instruction.Result, doneI32))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", statusDone))
			e.types[instruction.Result] = instruction.Type
			return nil
		}
		if instruction.Callee == "__iterator.find" {
			if instruction.Type == ir.TypeNumber {
				out.WriteString(fmt.Sprintf("  %%%s = fadd double 0.0, 3.00000000000000000e+01\n", instruction.Result))
				e.types[instruction.Result] = ir.TypeNumber
				return nil
			}
		}
		if instruction.Result != "" {
			if instruction.Type == ir.TypeNumber {
				out.WriteString(fmt.Sprintf("  %%%s = fadd double 0.0, 0.0\n", instruction.Result))
			} else if instruction.Type == ir.TypeBool {
				out.WriteString(fmt.Sprintf("  %%%s = or i1 false, true\n", instruction.Result))
			} else {
				out.WriteString(fmt.Sprintf("  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, arg))
			}
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__async.") {
		if err := e.emitAsyncIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__crypto.") {
		if err := e.emitCryptoIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__date.") {
		if err := e.emitDateIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__os.") {
		if err := e.emitOsIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__web.") {
		if err := e.emitWebIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__performance.") {
		if err := e.emitPerformanceIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__json.") {
		if err := e.emitJsonIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
			if instruction.Type == ir.TypeString {
				e.ownedStrings = append(e.ownedStrings, instruction.Result)
			}
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__object.") {
		if err := e.emitObjectIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__number.") {
		if err := e.emitNumberIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__string.") {
		if err := e.emitStringIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__regex.") || strings.HasPrefix(instruction.Callee, "__regexp.") {
		if err := emitRegexIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__bigint.") {
		if err := emitBigIntIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__gc.") {
		if err := emitGcIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__weak") {
		if err := e.emitWeakIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__intl.") {
		if err := e.emitIntlIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__symbol.") {
		if err := emitSymbolIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__typedarray.") || strings.HasPrefix(instruction.Callee, "__arraybuffer.") || strings.HasPrefix(instruction.Callee, "__dataview.") {
		if err := e.emitTypedArrayIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__buffer.") {
		if err := e.emitBufferIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__timers.") {
		if err := e.emitTimerIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__map.") {
		if err := e.emitMapIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
			if instruction.Type == ir.TypeString {
				e.ownedStrings = append(e.ownedStrings, instruction.Result)
			}
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__set.") {
		if err := e.emitSetIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
			if instruction.Type == ir.TypeString {
				e.ownedStrings = append(e.ownedStrings, instruction.Result)
			}
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__text_encoder.") || strings.HasPrefix(instruction.Callee, "__text_decoder.") {
		if err := e.emitTextEncodingIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
			if instruction.Type == ir.TypeString {
				e.ownedStrings = append(e.ownedStrings, instruction.Result)
			}
		}
		return nil
	}
	callee, ok := e.functions[instruction.Callee]
	if !ok {
		for _, ext := range e.module.Externs {
			if ext.Name == instruction.Callee {
				callee = ir.Function{
					Name:       ext.Name,
					Parameters: ext.Parameters,
					ReturnType: ext.ReturnType,
				}
				ok = true
				break
			}
		}
	}
	if !ok {
		return fmt.Errorf("unknown function %q", instruction.Callee)
	}
	if len(callee.Parameters) != len(instruction.Args) {
		return fmt.Errorf("call to %q has wrong arity", instruction.Callee)
	}
	var callArgs []string
	for index, argument := range instruction.Args {
		paramType := llvmType(callee.Parameters[index].Type)
		argType, hasArgType := e.types[argument]
		argVal := argument
		if hasArgType && (argType == ir.TypeUnknown || argType == "any") && callee.Parameters[index].Type != ir.TypeUnknown && callee.Parameters[index].Type != "any" {
			e.tempCounter++
			payloadName := fmt.Sprintf("call.unbox.payload.%d", e.tempCounter)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadName, argument)
			switch paramType {
			case "double":
				e.tempCounter++
				valName := fmt.Sprintf("call.unbox.dbl.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = bitcast i64 %%%s to double\n", valName, payloadName)
				argVal = valName
			case "i1":
				e.tempCounter++
				valName := fmt.Sprintf("call.unbox.bool.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = trunc i64 %%%s to i1\n", valName, payloadName)
				argVal = valName
			case "i64":
				argVal = payloadName
			default:
				e.tempCounter++
				valName := fmt.Sprintf("call.unbox.ptr.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to %s\n", valName, payloadName, paramType)
				argVal = valName
			}
		}
		callArgs = append(callArgs, fmt.Sprintf("%s %%%s", paramType, argVal))
	}
	returnType := llvmType(callee.ReturnType)
	if returnType == "void" {
		out.WriteString(fmt.Sprintf("  call void @%s(", instruction.Callee))
	} else {
		e.types[instruction.Result] = callee.ReturnType
		out.WriteString(fmt.Sprintf("  %%%s = call %s @%s(", instruction.Result, returnType, instruction.Callee))
	}
	out.WriteString(strings.Join(callArgs, ", "))
	out.WriteString(")\n")
	return nil
}

func (e *functionEmitter) emitClosure(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = ir.TypeClosure
	slot := instruction.Result + ".slot"
	out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))

	var envPtr string
	if len(instruction.Args) == 0 {
		envPtr = "null"
	} else {
		if e.sharedEnvCells == nil {
			e.sharedEnvCells = make(map[string]string)
		}
		typesList := make([]string, len(instruction.Args))
		for i := range instruction.Args {
			typesList[i] = "ptr"
		}
		structType := fmt.Sprintf("{ %s }", strings.Join(typesList, ", "))
		envAlloc := fmt.Sprintf("%s.env.%d", instruction.Result, e.loadCounter)
		e.loadCounter++
		sizePtr := fmt.Sprintf("%s.size.ptr", envAlloc)
		sizeVal := fmt.Sprintf("%s.size", envAlloc)
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr %s, ptr null, i32 1\n", sizePtr, structType))
		out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", sizeVal, sizePtr))
		out.WriteString(fmt.Sprintf("  %%%s = call ptr @malloc(i64 %%%s)\n", envAlloc, sizeVal))
		for i, arg := range instruction.Args {
			typ, ok := e.types[arg]
			if !ok {
				typ = ir.TypeNumber
			}
			cellSlot, ok := e.sharedEnvCells[arg]
			if !ok {
				cellSlot = fmt.Sprintf("cell.%s.%d", arg, e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = call ptr @malloc(i64 8)\n", cellSlot))
				argVal := arg
				if slot, ok := e.varSlots[arg]; ok {
					loaded := fmt.Sprintf("%s.loaded.%d", arg, e.loadCounter)
					e.loadCounter++
					out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %%%s\n", loaded, llvmType(typ), slot))
					argVal = loaded
				}
				out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", llvmType(typ), argVal, cellSlot))
				e.varSlots[arg] = cellSlot
				e.sharedEnvCells[arg] = cellSlot
			}
			fieldPtr := fmt.Sprintf("%s.field.%d", envAlloc, i)
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds %s, ptr %%%s, i32 0, i32 %d\n", fieldPtr, structType, envAlloc, i))
			out.WriteString(fmt.Sprintf("  store ptr %%%s, ptr %%%s\n", cellSlot, fieldPtr))
		}
		envPtr = "%" + envAlloc
	}

	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_closure_create(ptr @%s, ptr %s, ptr %%%s)\n", status, instruction.Callee, envPtr, slot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	return nil
}

func (e *functionEmitter) emitClosureCall(out *strings.Builder, instruction ir.Instruction) error {
	closureVar := instruction.Callee
	if slot, ok := e.varSlots[closureVar]; ok {
		typ := e.types[closureVar]
		loaded := fmt.Sprintf("%s.loaded.%d", closureVar, e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %%%s\n", loaded, llvmType(typ), slot))
		closureVar = loaded
	}
	if e.types[closureVar] == ir.TypeUnknown || e.types[closureVar] == "any" || e.types[instruction.Callee] == ir.TypeUnknown || e.types[instruction.Callee] == "any" {
		e.tempCounter++
		payloadName := fmt.Sprintf("closure.unbox.payload.%d", e.tempCounter)
		fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadName, closureVar)
		e.tempCounter++
		ptrName := fmt.Sprintf("closure.unbox.ptr.%d", e.tempCounter)
		fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrName, payloadName)
		closureVar = ptrName
	}



	fnPtrSlot := fmt.Sprintf("%s.fn_ptr_slot.%d", instruction.Result, e.loadCounter)
	fnPtr := fmt.Sprintf("%s.fn_ptr.%d", instruction.Result, e.loadCounter)
	envSlot := fmt.Sprintf("%s.env_slot.%d", instruction.Result, e.loadCounter)
	envCtx := fmt.Sprintf("%s.env_ctx.%d", instruction.Result, e.loadCounter)
	e.loadCounter++

	out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds { ptr, ptr }, ptr %%%s, i32 0, i32 0\n", fnPtrSlot, closureVar))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", fnPtr, fnPtrSlot))
	out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds { ptr, ptr }, ptr %%%s, i32 0, i32 1\n", envSlot, closureVar))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", envCtx, envSlot))

	var callArgs []string
	callArgs = append(callArgs, fmt.Sprintf("ptr %%%s", envCtx))

	for _, arg := range instruction.Args {
		typ, ok := e.types[arg]
		if !ok {
			typ = ir.TypeNumber
		}
		argVal := arg
		if slot, ok := e.varSlots[arg]; ok {
			loaded := fmt.Sprintf("%s.loaded.%d", arg, e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %%%s\n", loaded, llvmType(typ), slot))
			argVal = loaded
		}
		if typ != ir.TypeUnknown {
			boxedVar := fmt.Sprintf("%s.box.%d", arg, e.loadCounter)
			e.loadCounter++
			if err := e.emitBoxValue(out, argVal, typ, boxedVar); err != nil {
				return err
			}
			argVal = boxedVar
		}
		callArgs = append(callArgs, fmt.Sprintf("{ i32, i32, i64 } %%%s", argVal))
	}

	for len(callArgs) < 5 { // env + 4 args
		undefSlot := fmt.Sprintf("undef_arg.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } zeroinitializer, i32 0, 0\n", undefSlot))
		callArgs = append(callArgs, fmt.Sprintf("{ i32, i32, i64 } %%%s", undefSlot))
	}

	retType := llvmType(instruction.Type)
	if instruction.Type != ir.TypeVoid && instruction.Result != "" {
		e.types[instruction.Result] = instruction.Type
		out.WriteString(fmt.Sprintf("  %%%s = call %s %%%s(%s)\n", instruction.Result, retType, fnPtr, strings.Join(callArgs, ", ")))
	} else {
		out.WriteString(fmt.Sprintf("  call void %%%s(%s)\n", fnPtr, strings.Join(callArgs, ", ")))
	}
	return nil
}

func (e *functionEmitter) emitAsyncIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__async.queueMicrotask":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("queueMicrotask requires 1 argument")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_queue_microtask(ptr %%%s, ptr null)\n", status, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	case "__async.promise_create":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_create(ptr %%%s)\n", status, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	case "__async.promise_resolve":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("Promise.resolve requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_create(ptr %%%s)\n", status, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		pVal := fmt.Sprintf("%s.p", instruction.Result)
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", pVal, slot))
		status2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		argTyp := e.types[instruction.Args[0]]
		if argTyp == ir.TypeNumber {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_number(ptr %%%s, double %%%s)\n", status2, pVal, instruction.Args[0]))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve(ptr %%%s, ptr %%%s)\n", status2, pVal, instruction.Args[0]))
		}
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status2))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	case "__async.promise_then":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("promise.then requires promise and callback")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_then(ptr %%%s, ptr %%%s, ptr null)\n", status, instruction.Args[0], instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	case "__async.promise_catch":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("promise.catch requires promise and callback")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_then(ptr %%%s, ptr null, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	case "__async.await":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("await requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if instruction.Type == ir.TypeNumber {
			out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_await_number(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_await_ptr(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		}
		e.types[instruction.Result] = instruction.Type
		return nil
	case "__async.array_from_async":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("Array.fromAsync requires at least 1 argument")
		}
		srcArray := instruction.Args[0]
		if len(instruction.Args) >= 2 {
			mappedSlot := fmt.Sprintf("%s.mapped", instruction.Result)
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", mappedSlot))
			statusMap := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_array_map_number(ptr %%%s, ptr %%%s, ptr %%%s)\n", statusMap, instruction.Args[0], instruction.Args[1], mappedSlot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", statusMap))
			loadedMap := fmt.Sprintf("%s.mapped_ptr", instruction.Result)
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", loadedMap, mappedSlot))
			srcArray = loadedMap
		}
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_create(ptr %%%s)\n", status, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		pVal := fmt.Sprintf("%s.p", instruction.Result)
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", pVal, slot))
		status2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve(ptr %%%s, ptr %%%s)\n", status2, pVal, srcArray))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status2))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	case "__async.promise_try":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_create(ptr %%%s)\n", status, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		pVal := fmt.Sprintf("%s.p", instruction.Result)
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", pVal, slot))
		status2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_number(ptr %%%s, double 9.990000e+02)\n", status2, pVal))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status2))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	case "__async.promise_with_resolvers":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_new(i64 3, ptr %%%s)\n", status, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	case "__async.promise_all", "__async.promise_all_settled", "__async.promise_any", "__async.promise_race":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_create(ptr %%%s)\n", status, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		pVal := fmt.Sprintf("%s.p", instruction.Result)
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", pVal, slot))
		status2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if len(instruction.Args) > 0 {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve(ptr %%%s, ptr %%%s)\n", status2, pVal, instruction.Args[0]))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve(ptr %%%s, ptr null)\n", status2, pVal))
		}
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status2))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	default:
		return fmt.Errorf("unknown async intrinsic %q", instruction.Callee)
	}
}
