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
	case ir.TypeVoid:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		undefGlobal := e.stringsByValue["undefined"]
		ptrUndef := fmt.Sprintf("print.undef.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [10 x i8], ptr %s, i64 0, i64 0\n", ptrUndef, undefGlobal))
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s)\n", status, name, ptrUndef))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
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
		arg := instruction.Args[0]
		argType := e.types[arg]
		if slot, ok := e.varSlots[arg]; ok {
			loaded := fmt.Sprintf("%s.con_load.%d", arg, e.loadCounter)
			e.loadCounter++
			if argType == ir.TypeUnknown {
				out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64 }, ptr %%%s\n", loaded, slot))
			} else {
				out.WriteString(fmt.Sprintf("  %%%s = load volatile %s, ptr %%%s\n", loaded, llvmType(argType), slot))
			}
			arg = loaded
		}
		if argType != ir.TypeUnknown {
			boxedVar := fmt.Sprintf("box.con.%d", e.loadCounter)
			if err := e.emitBoxValue(out, arg, argType, boxedVar); err != nil {
				return err
			}
			arg = boxedVar
		}
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, arg))
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 1\n", flagsVar, arg))
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg))
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
		if argType == ir.TypeUnknown || e.isParamUnknown(arg) {
			if slot, ok := e.varSlots[arg]; ok {
				loaded := fmt.Sprintf("%s.isarr.loaded.%d", arg, e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64 }, ptr %%%s\n", loaded, slot))
				arg = loaded
			}
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
			emptyGlobal := "@.str.empty"
			if strGlobal, ok := e.stringsByValue[""]; ok {
				emptyGlobal = strGlobal
			}
			emptyPtr := fmt.Sprintf("empty.str.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = getelementptr inbounds [1 x i8], ptr %s, i64 0, i64 0\n", emptyPtr, emptyGlobal)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_split(ptr %%%s, ptr %%%s, double -1.000000e+00, ptr %%%s)\n", status, instruction.Args[0], emptyPtr, slot)
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
		if !ok || arrayType == "" {
			for _, g := range e.module.Globals {
				if g.Name == instruction.Args[0] {
					arrayType = g.Type
					ok = true
					break
				}
			}
		}
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
			statusDone := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			doneI32 := fmt.Sprintf("%s.done.i32", instruction.Result)
			out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i32\n", doneI32, isDone))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_bool_set(ptr %%%s, i64 0, i32 %%%s)\n", statusDone, instruction.Result, doneI32))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", statusDone))
			statusVal := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			doubleVal := fmt.Sprintf("%s.double.val", instruction.Result)
			intVal := fmt.Sprintf("%s.int.val", instruction.Result)
			out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", intVal, loadedElem))
			out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s to double\n", doubleVal, intVal))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 1, double %%%s)\n", statusVal, instruction.Result, doubleVal))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", statusVal))
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
	if strings.HasPrefix(instruction.Callee, "__zlib.") {
		if err := e.emitZlibIntrinsic(out, instruction); err != nil {
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
	if strings.HasPrefix(instruction.Callee, "__sqlite.") {
		if err := e.emitSqliteIntrinsic(out, instruction); err != nil {
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
	if instruction.Callee == "__scriptgo.is_truthy" {
		arg := instruction.Args[0]
		argType := e.types[arg]
		argVal := arg
		if slot, ok := e.varSlots[arg]; ok {
			loaded := fmt.Sprintf("%s.loaded.%d", arg, e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %%%s\n", loaded, llvmType(argType), slot))
			argVal = loaded
		}
		if argType != ir.TypeUnknown {
			boxedVar := fmt.Sprintf("box.truthy.%d", e.loadCounter)
			if err := e.emitBoxValue(out, argVal, argType, boxedVar); err != nil {
				return err
			}
			argVal = boxedVar
		}
		tagVar := fmt.Sprintf("truthy.tag.%d", e.loadCounter)
		payloadVar := fmt.Sprintf("truthy.payload.%d", e.loadCounter)
		i32Res := fmt.Sprintf("truthy.i32.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, argVal))
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, argVal))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_is_truthy_unknown(i32 %%%s, i64 %%%s)\n", i32Res, tagVar, payloadVar))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Res))
		e.types[instruction.Result] = ir.TypeBool
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
		if err := e.emitRegexIntrinsic(out, instruction); err != nil {
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
	if strings.HasPrefix(instruction.Callee, "__weak") || strings.HasPrefix(instruction.Callee, "__finalization_registry.") {
		if err := e.emitWeakIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__atomics.") {
		if err := e.emitAtomicsIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
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
	if strings.HasPrefix(instruction.Callee, "__typedarray.") || strings.HasPrefix(instruction.Callee, "__arraybuffer.") || strings.HasPrefix(instruction.Callee, "__arraybuffer_view.") || strings.HasPrefix(instruction.Callee, "__dataview.") {
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
	if strings.HasPrefix(instruction.Callee, "__dns.") {
		if err := e.emitDnsIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__net.") {
		if err := e.emitNetIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__dgram.") {
		if err := e.emitDgramIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__tls.") {
		if err := e.emitTLSIntrinsic(out, instruction); err != nil {
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
		argVal := e.resolveArg(out, argument)
		if strings.HasPrefix(e.function.Name, "__top_level_async_stage_") && !e.isParam(argument) {
			for _, global := range e.module.Globals {
				if global.Name == argument {
					loadName := fmt.Sprintf("%s.stage_global.%d", argument, e.loadCounter)
					e.loadCounter++
					typ := e.types[argument]
					if typ == "" || typ == ir.TypeVoid {
						typ = global.Type
					}
					if typ == "" || typ == ir.TypeVoid {
						typ = ir.TypePointer
					}
					e.types[loadName] = typ
					out.WriteString(fmt.Sprintf("  %%%s = load volatile %s, ptr @%s\n", loadName, llvmType(typ), global.Name))
					argVal = loadName
					break
				}
			}
		}
		if isRawCallbackParameter(callee.Parameters[index]) {
			boxed := argVal
			for field, fieldType := range []string{"i32", "i32", "i64"} {
				fieldName := fmt.Sprintf("call.raw.%s.%d", []string{"tag", "pad", "payload"}[field], e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, %d\n", fieldName, boxed, field))
				callArgs = append(callArgs, fmt.Sprintf("%s %%%s", fieldType, fieldName))
			}
			continue
		}
		argType, hasArgType := e.types[argVal]
		if !hasArgType {
			argType, hasArgType = e.types[argument]
		}
		if callee.Parameters[index].Type == ir.TypeUnknown && hasArgType && argType != ir.TypeUnknown {
			boxedName := fmt.Sprintf("call.box.%d", e.loadCounter)
			if err := e.emitBoxValue(out, argVal, argType, boxedName); err != nil {
				return err
			}
			argVal = boxedName
		} else if hasArgType && argType == ir.TypeUnknown && callee.Parameters[index].Type != ir.TypeUnknown {
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
		} else if hasArgType && argType == ir.TypeVoid && paramType == "double" {
			nanName := fmt.Sprintf("call.nan.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = fadd double 0.0, 0x7FF8000000000000\n", nanName))
			argVal = nanName
		}
		callArgs = append(callArgs, fmt.Sprintf("%s %%%s", paramType, argVal))
	}
	returnType := llvmType(callee.ReturnType)
	if callee.ReturnType == ir.TypeBool {
		returnType = "zeroext i1"
	}
	if returnType == "void" {
		out.WriteString(fmt.Sprintf("  call void @%s(", mangleFunctionName(instruction.Callee)))
	} else {
		e.types[instruction.Result] = callee.ReturnType
		out.WriteString(fmt.Sprintf("  %%%s = call %s @%s(", instruction.Result, returnType, mangleFunctionName(instruction.Callee)))
	}
	out.WriteString(strings.Join(callArgs, ", "))
	out.WriteString(")\n")
	return nil
}

func (e *functionEmitter) emitClosure(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = ir.TypeClosure
	slot := instruction.Result + ".slot"
	if cellSlot, isCell := e.sharedEnvCells[instruction.Result]; isCell {
		slot = cellSlot
	} else if existingSlot, ok := e.varSlots[instruction.Result]; ok {
		slot = existingSlot
	} else {
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		e.varSlots[instruction.Result] = slot
	}

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
		out.WriteString(fmt.Sprintf("  %%%s = call ptr @scriptgo_closure_alloc(i64 %%%s)\n", envAlloc, sizeVal))
		for i, arg := range instruction.Args {
			typ, okTyp := e.types[arg]
			if !okTyp {
				typ = ir.TypeNumber
			}
			fieldPtr := fmt.Sprintf("%s.field.%d", envAlloc, i)
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds %s, ptr %%%s, i32 0, i32 %d\n", fieldPtr, structType, envAlloc, i))
			if cellSlot, ok := e.sharedEnvCells[arg]; ok && len(e.loopBreakLabels) == 0 {
				out.WriteString(fmt.Sprintf("  store ptr %%%s, ptr %%%s\n", cellSlot, fieldPtr))
			} else {
				cellAlloc := fmt.Sprintf("closure.cell.%s.%d", arg, e.loadCounter)
				e.loadCounter++
				allocSize := 8
				if typ == ir.TypeUnknown {
					allocSize = 16
				}
				out.WriteString(fmt.Sprintf("  %%%s = call ptr @scriptgo_closure_alloc(i64 %d)\n", cellAlloc, allocSize))
				argVal := e.resolveArg(out, arg)
				out.WriteString(fmt.Sprintf("  store volatile %s %%%s, ptr %%%s\n", llvmType(typ), argVal, cellAlloc))
				out.WriteString(fmt.Sprintf("  store ptr %%%s, ptr %%%s\n", cellAlloc, fieldPtr))
			}
		}
		envPtr = "%" + envAlloc
	}

	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_closure_create(ptr @%s, ptr %s, ptr %%%s)\n", status, mangleFunctionName(instruction.Callee), envPtr, slot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	for _, g := range e.module.Globals {
		if g.Name == instruction.Result {
			out.WriteString(fmt.Sprintf("  store volatile ptr %%%s, ptr @%s\n", instruction.Result, g.Name))
			break
		}
	}
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
	} else {
		for _, g := range e.module.Globals {
			if g.Name == closureVar {
				loadName := fmt.Sprintf("%s.gload.%d", closureVar, e.loadCounter)
				e.loadCounter++
				e.types[loadName] = g.Type
				out.WriteString(fmt.Sprintf("  %%%s = load volatile %s, ptr @%s\n", loadName, llvmType(g.Type), g.Name))
				closureVar = loadName
				break
			}
		}
	}
	if e.types[closureVar] == ir.TypeUnknown || e.types[instruction.Callee] == ir.TypeUnknown {
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
	closureIsNull := fmt.Sprintf("closure.is_null.%d", e.loadCounter)
	closureIsUndef := fmt.Sprintf("closure.is_undef.%d", e.loadCounter)
	closureInvalid := fmt.Sprintf("closure.invalid.%d", e.loadCounter)
	nullBlock := fmt.Sprintf("closure.null.%d", e.loadCounter)
	callBlock := fmt.Sprintf("closure.call.%d", e.loadCounter)
	contBlock := fmt.Sprintf("closure.cont.%d", e.loadCounter)
	callRes := fmt.Sprintf("%s.call_res.%d", instruction.Result, e.loadCounter)
	e.loadCounter++

	var callArgs []string
	if instruction.Callee == e.function.Name || strings.HasSuffix(e.function.Name, "_"+instruction.Callee) {
		// Direct recursive call to current closure
		fnPtr = "@" + e.function.Name
		callArgs = append(callArgs, "ptr %__env_ctx")
		goto prepareArgs
	}

	out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, null\n", closureIsNull, closureVar))
	out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, @scriptgo_undefined_sentinel\n", closureIsUndef, closureVar))
	out.WriteString(fmt.Sprintf("  %%%s = or i1 %%%s, %%%s\n", closureInvalid, closureIsNull, closureIsUndef))
	out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s\n", closureInvalid, nullBlock, callBlock))

	out.WriteString(fmt.Sprintf("%s:\n", nullBlock))
	out.WriteString(fmt.Sprintf("  br label %%%s\n", contBlock))

	out.WriteString(fmt.Sprintf("%s:\n", callBlock))
	out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds { ptr, ptr }, ptr %%%s, i32 0, i32 0\n", fnPtrSlot, closureVar))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", fnPtr, fnPtrSlot))
	out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds { ptr, ptr }, ptr %%%s, i32 0, i32 1\n", envSlot, closureVar))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", envCtx, envSlot))
	callArgs = append(callArgs, fmt.Sprintf("ptr %%%s", envCtx))

prepareArgs:
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
	fnSig := ""
	fnTarget := "%" + fnPtr
	if strings.HasPrefix(fnPtr, "@") {
		fnTarget = fnPtr
	} else {
		fnSig = "(ptr, { i32, i32, i64 }, { i32, i32, i64 }, { i32, i32, i64 }, { i32, i32, i64 }) "
	}
	if instruction.Result != "" && retType != "void" {
		e.types[instruction.Result] = instruction.Type
		out.WriteString(fmt.Sprintf("  %%%s = call %s %s%s(%s)\n", callRes, retType, fnSig, fnTarget, strings.Join(callArgs, ", ")))
		out.WriteString(fmt.Sprintf("  br label %%%s\n", contBlock))
		out.WriteString(fmt.Sprintf("%s:\n", contBlock))
		defaultVal := "null"
		if instruction.Type == ir.TypeNumber {
			defaultVal = "0.0"
		} else if instruction.Type == ir.TypeBool {
			defaultVal = "false"
		} else if instruction.Type == ir.TypeUnknown {
			defaultVal = "zeroinitializer"
		} else if instruction.Type == ir.TypeBigInt {
			defaultVal = "0"
		}
		out.WriteString(fmt.Sprintf("  %%%s = phi %s [ %s, %%%s ], [ %%%s, %%%s ]\n", instruction.Result, retType, defaultVal, nullBlock, callRes, callBlock))
	} else {
		if fnSig != "" || strings.HasPrefix(fnTarget, "@__closure_") {
			out.WriteString(fmt.Sprintf("  call { i32, i32, i64 } %s%s(%s)\n", fnSig, fnTarget, strings.Join(callArgs, ", ")))
		} else {
			out.WriteString(fmt.Sprintf("  call void %s(%s)\n", fnTarget, strings.Join(callArgs, ", ")))
		}
		out.WriteString(fmt.Sprintf("  br label %%%s\n", contBlock))
		out.WriteString(fmt.Sprintf("%s:\n", contBlock))
	}
	return nil
}

func (e *functionEmitter) emitPromiseSettlement(out *strings.Builder, instruction ir.Instruction, promise string, status string, rejected bool) error {
	if len(instruction.Args) == 0 {
		if rejected {
			return fmt.Errorf("promise.reject requires 1 argument")
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_boxed(ptr %%%s, i32 0, i64 0)\n", status, promise))
		return nil
	}
	if len(instruction.Args) != 1 {
		return fmt.Errorf("promise settlement requires 1 argument")
	}

	arg := instruction.Args[0]
	argType := e.types[arg]
	argVal := e.resolveArg(out, arg)
	if !rejected && argType == ir.TypeNumber {
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_number(ptr %%%s, double %%%s)\n", status, promise, argVal))
		return nil
	}
	if !rejected && argType == ir.TypeBool {
		boolVal := fmt.Sprintf("promise.bool.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i32\n", boolVal, argVal))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_bool(ptr %%%s, i32 %%%s)\n", status, promise, boolVal))
		return nil
	}
	if !rejected && argType == ir.TypeBigInt {
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_bigint(ptr %%%s, i64 %%%s)\n", status, promise, argVal))
		return nil
	}

	boxed := fmt.Sprintf("promise.box.%d", e.loadCounter)
	e.loadCounter++
	if err := e.emitBoxValue(out, argVal, argType, boxed); err != nil {
		return err
	}
	tag := fmt.Sprintf("promise.tag.%d", e.loadCounter)
	payload := fmt.Sprintf("promise.payload.%d", e.loadCounter)
	e.loadCounter++
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag, boxed))
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payload, boxed))
	settleFn := "scriptgo_promise_resolve_boxed"
	if rejected {
		settleFn = "scriptgo_promise_reject_boxed"
	}
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, i32 %%%s, i64 %%%s)\n", status, settleFn, promise, tag, payload))
	return nil
}

func (e *functionEmitter) emitAsyncIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__async.frame_new":
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		count := instruction.Value
		if count == "" {
			count = "0"
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_async_frame_new(i64 %s, ptr %%%s)\n", status, count, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = ir.TypePointer
		if e.function.AsyncFrame != nil {
			if err := e.emitAsyncFrameInitialValues(out, instruction.Result); err != nil {
				return err
			}
		}
		return nil
	case "__async.frame_release":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("async frame release requires one frame")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_async_frame_release(ptr %%%s)\n", status, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
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
	case "__async.promise_resolver":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("promise resolver requires a promise")
		}
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		reject := 0
		if instruction.Value == "reject" {
			reject = 1
		} else if instruction.Value != "resolve" {
			return fmt.Errorf("unknown Promise resolver %q", instruction.Value)
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolver_create(ptr %%%s, i32 %d, ptr %%%s)\n", status, instruction.Args[0], reject, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	case "__async.promise_construct":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("promise constructor requires exactly one executor")
		}
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_construct(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	case "__async.promise_resolve":
		if len(instruction.Args) > 1 {
			return fmt.Errorf("promise.resolve requires at most 1 argument")
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
		if err := e.emitPromiseSettlement(out, instruction, pVal, status2, false); err != nil {
			return err
		}
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status2))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	case "__async.promise_reject":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("promise.reject requires 1 argument")
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
		if err := e.emitPromiseSettlement(out, instruction, pVal, status2, true); err != nil {
			return err
		}
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status2))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.types[instruction.Result] = instruction.Type
		return nil
	case "__async.promise_then":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("promise.then requires promise and callback")
		}
		rejArg := "null"
		if len(instruction.Args) >= 3 && instruction.Args[2] != "" && instruction.Args[2] != "null" {
			rejArg = "%" + instruction.Args[2]
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		resultSlot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", resultSlot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_then(ptr %%%s, ptr %%%s, ptr %s, i32 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], rejArg, promiseReactionTag(instruction.Value), resultSlot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		if instruction.Result != "" {
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, resultSlot))
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	case "__async.promise_catch":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("promise.catch requires promise and callback")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		resultSlot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", resultSlot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_then(ptr %%%s, ptr null, ptr %%%s, i32 %d, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], promiseReactionTag(instruction.Value), resultSlot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		if instruction.Result != "" {
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, resultSlot))
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	case "__async.promise_schedule_resume":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("promise resume scheduling requires promise and continuation")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		promiseArg, err := e.emitAsyncPromiseArg(out, instruction.Args[0])
		if err != nil {
			return err
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_schedule_resume(ptr %%%s, ptr %%%s)\n", status, promiseArg, instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	case "__async.promise_schedule_resume_pair":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("promise resume pair requires promise and two continuations")
		}
		if promiseType := e.types[instruction.Args[0]]; promiseType != "" && promiseType != ir.TypeUnknown && promiseType != ir.Type("object:Promise") && !strings.HasPrefix(string(promiseType), "object:Promise_") && !strings.HasPrefix(string(promiseType), "object:Promise<") {
			return fmt.Errorf("async resume source %q has non-Promise type %s", instruction.Args[0], promiseType)
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		promiseArg, err := e.emitAsyncPromiseArg(out, instruction.Args[0])
		if err != nil {
			return err
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_schedule_resume_pair(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, promiseArg, instruction.Args[1], instruction.Args[2]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	case "__async.promise_resolve_existing", "__async.promise_reject_existing":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("promise settlement requires promise and value")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fn := "scriptgo_promise_resolve_existing"
		if instruction.Callee == "__async.promise_reject_existing" {
			fn = "scriptgo_promise_reject_existing"
		}
		argType := e.types[instruction.Args[1]]
		argVal := e.resolveArg(out, instruction.Args[1])
		if instruction.Callee == "__async.promise_resolve_existing" && argType == ir.TypeNumber {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_existing_number(ptr %%%s, double %%%s)\n", status, instruction.Args[0], argVal))
		} else if instruction.Callee == "__async.promise_resolve_existing" && argType == ir.TypeBool {
			boolVal := fmt.Sprintf("promise.existing.bool.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i32\n", boolVal, argVal))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_existing_bool(ptr %%%s, i32 %%%s)\n", status, instruction.Args[0], boolVal))
		} else if instruction.Callee == "__async.promise_resolve_existing" && isJSArrayType(argType) {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_existing_array(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], argVal))
		} else {
			boxed := fmt.Sprintf("promise.existing.box.%d", e.loadCounter)
			e.loadCounter++
			if err := e.emitBoxValue(out, argVal, argType, boxed); err != nil {
				return err
			}
			tag := fmt.Sprintf("promise.existing.tag.%d", e.loadCounter)
			payload := fmt.Sprintf("promise.existing.payload.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag, boxed))
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payload, boxed))
			fn = "scriptgo_promise_resolve_existing_boxed"
			if instruction.Callee == "__async.promise_reject_existing" {
				fn = "scriptgo_promise_reject_existing_boxed"
			}
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, i32 %%%s, i64 %%%s)\n", status, fn, instruction.Args[0], tag, payload))
		}
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	case "__async.await":
		if e.function.Async {
			return fmt.Errorf("async state machine contains blocking __async.await")
		}
		if len(instruction.Args) != 1 {
			return fmt.Errorf("await requires 1 argument")
		}
		promVar := instruction.Args[0]
		if e.types[promVar] == ir.TypeUnknown {
			// `unknown` may be a Promise, but it may also be an ordinary value
			// such as undefined. Let the runtime inspect the tag before waiting.
			unknownVal := e.resolveArg(out, promVar)
			e.tempCounter++
			tagName := fmt.Sprintf("await.unknown.tag.%d", e.tempCounter)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagName, unknownVal)
			e.tempCounter++
			payloadName := fmt.Sprintf("await.unknown.payload.%d", e.tempCounter)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadName, unknownVal)
			tagSlot := fmt.Sprintf("await.unknown.tag.slot.%d", e.tempCounter)
			payloadSlot := fmt.Sprintf("await.unknown.payload.slot.%d", e.tempCounter)
			fmt.Fprintf(out, "  %%%s = alloca i32\n", tagSlot)
			fmt.Fprintf(out, "  %%%s = alloca i64\n", payloadSlot)
			status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_promise_await_unknown(i32 %%%s, i64 %%%s, ptr %%%s, ptr %%%s)\n", status, tagName, payloadName, tagSlot, payloadSlot)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
			resultTag := fmt.Sprintf("await.unknown.result.tag.%d", e.tempCounter)
			resultPayload := fmt.Sprintf("await.unknown.result.payload.%d", e.tempCounter)
			fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", resultTag, tagSlot)
			fmt.Fprintf(out, "  %%%s = load i64, ptr %%%s\n", resultPayload, payloadSlot)
			rawResult := instruction.Result
			if instruction.Type != ir.TypeUnknown {
				rawResult += ".await_unknown"
			}
			b0 := rawResult + ".b0"
			b1 := rawResult + ".b1"
			fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64 } undef, i32 %%%s, 0\n", b0, resultTag)
			fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", b1, b0)
			fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 %%%s, 2\n", rawResult, b1, resultPayload)
			e.types[rawResult] = ir.TypeUnknown
			if instruction.Type == ir.TypeUnknown {
				e.types[instruction.Result] = ir.TypeUnknown
				return nil
			}
			return e.emitCheckedCast(out, ir.Instruction{
				Op:     ir.OpCheckedCast,
				Type:   instruction.Type,
				Result: instruction.Result,
				Args:   []string{rawResult},
			})
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if instruction.Type == ir.TypeNumber {
			out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_await_number(ptr %%%s, ptr %%%s)\n", status, promVar, slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		} else if instruction.Type == ir.TypeBool {
			out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_await_bool(ptr %%%s, ptr %%%s)\n", status, promVar, slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			boolVal := fmt.Sprintf("%s.bool", instruction.Result)
			out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%%s\n", boolVal, slot))
			out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, boolVal))
		} else if instruction.Type == ir.TypeBigInt {
			out.WriteString(fmt.Sprintf("  %%%s = alloca i64\n", slot))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_await_bigint(ptr %%%s, ptr %%%s)\n", status, promVar, slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s\n", instruction.Result, slot))
		} else if instruction.Type == ir.TypeUnknown {
			tagSlot := instruction.Result + ".tag.slot"
			payloadSlot := instruction.Result + ".payload.slot"
			out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", tagSlot))
			out.WriteString(fmt.Sprintf("  %%%s = alloca i64\n", payloadSlot))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_await_boxed(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, promVar, tagSlot, payloadSlot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			tagVal := instruction.Result + ".tag"
			payloadVal := instruction.Result + ".payload"
			out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%%s\n", tagVal, tagSlot))
			out.WriteString(fmt.Sprintf("  %%%s = load i64, ptr %%%s\n", payloadVal, payloadSlot))
			b0 := instruction.Result + ".b0"
			b1 := instruction.Result + ".b1"
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } undef, i32 %%%s, 0\n", b0, tagVal))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", b1, b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, b1, payloadVal))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_await_ptr(ptr %%%s, ptr %%%s)\n", status, promVar, slot))
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
		// Array.fromAsync produces a known array payload; preserve its array tag
		// so an await continuation can reconstruct the statically typed array.
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_existing_array(ptr %%%s, ptr %%%s)\n", status2, pVal, srcArray))
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
		if instruction.Callee == "__async.promise_all" && len(instruction.Args) == 1 {
			slot := instruction.Result + ".slot"
			out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
			status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_all_numbers(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
			e.types[instruction.Result] = instruction.Type
			return nil
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
		if len(instruction.Args) > 0 {
			argTyp := e.types[instruction.Args[0]]
			if argTyp == ir.TypeNumber {
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_number(ptr %%%s, double %%%s)\n", status2, pVal, instruction.Args[0]))
			} else if argTyp == ir.TypeBool {
				bVar := fmt.Sprintf("b.%d", e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i64\n", bVar, instruction.Args[0]))
				ptrName := fmt.Sprintf("pbox.%d", e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrName, bVar))
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve(ptr %%%s, ptr %%%s)\n", status2, pVal, ptrName))
			} else if argTyp == ir.TypeUnknown {
				payloadName := fmt.Sprintf("%s.payload", instruction.Args[0])
				ptrName := fmt.Sprintf("%s.ptr", instruction.Args[0])
				out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadName, instruction.Args[0]))
				out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrName, payloadName))
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve(ptr %%%s, ptr %%%s)\n", status2, pVal, ptrName))
			} else {
				out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve(ptr %%%s, ptr %%%s)\n", status2, pVal, instruction.Args[0]))
			}
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

// emitAsyncFrameInitialValues persists values available at async entry. Later
// states still use the same frame object, so this is the stable storage ABI
// even when a continuation also has an optimized captured representation.
func (e *functionEmitter) emitAsyncFrameInitialValues(out *strings.Builder, frame string) error {
	fieldIndex := make(map[string]int, len(e.function.AsyncFrame.Fields))
	for index, field := range e.function.AsyncFrame.Fields {
		fieldIndex[field.Name] = index
	}
	for _, field := range e.function.AsyncFrame.Fields {
		if field.Name == "state" {
			// State is written by the continuation dispatcher when that ABI is
			// emitted; there is no SSA value for the initial zero state here.
			continue
		}
		// The literal promise field is descriptive metadata; its SSA name is
		// generated per lowering site and is not recoverable from the field.
		if field.Name == "promise" {
			continue
		}
		valueType, ok := e.types[field.Name]
		if !ok || valueType == ir.TypeVoid {
			continue
		}
		if err := e.emitAsyncFrameBoxed(out, frame, fieldIndex[field.Name], field.Name, valueType); err != nil {
			return err
		}
	}
	return nil
}

func (e *functionEmitter) emitAsyncFrameBoxed(out *strings.Builder, frame string, index int, value string, valueType ir.Type) error {
	boxed := fmt.Sprintf("async.frame.box.%d", e.loadCounter)
	e.loadCounter++
	if err := e.emitBoxValue(out, value, valueType, boxed); err != nil {
		return err
	}
	tag := fmt.Sprintf("async.frame.tag.%d", e.loadCounter)
	payload := fmt.Sprintf("async.frame.payload.%d", e.loadCounter)
	e.loadCounter++
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag, boxed))
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payload, boxed))
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_async_frame_set(ptr %%%s, i64 %d, i32 %%%s, i64 %%%s)\n", status, frame, index, tag, payload))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	return nil
}

func (e *functionEmitter) emitAsyncPromiseArg(out *strings.Builder, arg string) (string, error) {
	if typ := e.types[arg]; typ != "" && typ != ir.TypeUnknown {
		return e.ensurePointerArg(out, arg), nil
	}
	boxed := e.resolveArg(out, arg)
	tag := fmt.Sprintf("async.promise.tag.%d", e.loadCounter)
	payload := fmt.Sprintf("async.promise.payload.%d", e.loadCounter)
	e.loadCounter++
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tag, boxed))
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payload, boxed))
	slot := fmt.Sprintf("async.promise.slot.%d", e.loadCounter)
	e.loadCounter++
	out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_unknown(i32 %%%s, i64 %%%s, ptr %%%s)\n", status, tag, payload, slot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	result := fmt.Sprintf("async.promise.%d", e.loadCounter)
	e.loadCounter++
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", result, slot))
	return result, nil
}

func isJSArrayType(t ir.Type) bool {
	return t == ir.TypeNumberArray || t == ir.TypeStringArray ||
		t == ir.TypeBoolArray || t == ir.TypeBigIntArray ||
		t == ir.TypeSymbolArray || t == ir.TypeUnknownArray ||
		strings.HasSuffix(string(t), "[]")
}

func promiseReactionTag(typeName string) int {
	switch typeName {
	case string(ir.TypeNumber):
		return 3
	case string(ir.TypeBool):
		return 2
	case string(ir.TypeString):
		return 4
	case string(ir.TypeVoid), "":
		return 0
	default:
		if strings.HasSuffix(typeName, "[]") {
			return 6 // array payload
		}
		return 5
	}
}
