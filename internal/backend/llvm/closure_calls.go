package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func closureReturnTag(valueType ir.Type) int {
	switch valueType {
	case ir.TypeUnknown:
		return -1
	case ir.TypeVoid:
		return 0
	case ir.TypeBool:
		return 2
	case ir.TypeNumber:
		return 3
	case ir.TypeString:
		return 4
	case ir.TypeClosure:
		return 7
	case ir.TypeBigInt:
		return 8
	case ir.TypeSymbol:
		return 9
	default:
		if strings.HasSuffix(string(valueType), "[]") {
			return 6
		}
		return 5
	}
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
	} else if len(instruction.Args) == 1 && instruction.Args[0] == "this" {
		argVal := e.resolveArg(out, "this")
		if !strings.HasPrefix(argVal, "%") && !strings.HasPrefix(argVal, "@") {
			envPtr = "%" + argVal
		} else {
			envPtr = argVal
		}
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
			if arg == "this" {
				argVal := e.resolveArg(out, "this")
				if !strings.HasPrefix(argVal, "%") && !strings.HasPrefix(argVal, "@") {
					argVal = "%" + argVal
				}
				out.WriteString(fmt.Sprintf("  store ptr %s, ptr %%%s\n", argVal, fieldPtr))
			} else if cellSlot, ok := e.sharedEnvCells[arg]; ok && len(e.loopBreakLabels) == 0 {
				out.WriteString(fmt.Sprintf("  store ptr %%%s, ptr %%%s\n", cellSlot, fieldPtr))
			} else {
				cellAlloc := fmt.Sprintf("closure.cell.%s.%d", arg, e.loadCounter)
				e.loadCounter++
				allocSize := 8
				if typ == ir.TypeUnknown {
					allocSize = 24
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
	returnTag := 0
	if callee, ok := e.functions[instruction.Callee]; ok {
		returnTag = closureReturnTag(callee.ReturnType)
	}
	calleeName := mangleFunctionName(instruction.Callee)
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_closure_create(ptr %s, ptr %s, ptr %s, i32 %d, ptr %%%s)\n", status, functionSymbol(calleeName), envPtr, functionSymbol(calleeName+"$invoke"), returnTag, slot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	for _, g := range e.module.Globals {
		if g.Name == instruction.Result {
			out.WriteString(fmt.Sprintf("  store volatile ptr %%%s, ptr %s\n", instruction.Result, functionSymbol(g.Name)))
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
				out.WriteString(fmt.Sprintf("  %%%s = load volatile %s, ptr %s\n", loadName, llvmType(g.Type), functionSymbol(g.Name)))
				closureVar = loadName
				break
			}
		}
	}
	if e.types[closureVar] == ir.TypeUnknown || e.types[instruction.Callee] == ir.TypeUnknown {
		e.tempCounter++
		payloadName := fmt.Sprintf("closure.unbox.payload.%d", e.tempCounter)
		fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadName, closureVar)
		e.tempCounter++
		ptrName := fmt.Sprintf("closure.unbox.ptr.%d", e.tempCounter)
		fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrName, payloadName)
		closureVar = ptrName
	}

	callID := e.loadCounter
	e.loadCounter++
	hasEnvCtx := false
	for _, p := range e.function.Parameters {
		if p.Name == "__env_ctx" {
			hasEnvCtx = true
			break
		}
	}
	_, isLocalVar := e.varSlots[instruction.Callee]
	recursive := hasEnvCtx && !isLocalVar && (instruction.Callee == e.function.Name || (!strings.Contains(e.function.Name, "__") && strings.HasSuffix(e.function.Name, "_"+instruction.Callee)))
	useValueDispatcher := !recursive && (instruction.Type == ir.TypeUnknown || instruction.Type == ir.TypeVoid)
	fnPtrSlot := fmt.Sprintf("%s.fn_ptr_slot.%d", instruction.Result, callID)
	fnPtr := fmt.Sprintf("%s.fn_ptr.%d", instruction.Result, callID)
	envSlot := fmt.Sprintf("%s.env_slot.%d", instruction.Result, callID)
	envCtx := fmt.Sprintf("%s.env_ctx.%d", instruction.Result, callID)
	closureIsNull := fmt.Sprintf("closure.is_null.%d", callID)
	closureIsUndef := fmt.Sprintf("closure.is_undef.%d", callID)
	closureInvalid := fmt.Sprintf("closure.invalid.%d", callID)
	nullBlock := fmt.Sprintf("closure.null.%d", callID)
	callBlock := fmt.Sprintf("closure.call.%d", callID)
	contBlock := fmt.Sprintf("closure.cont.%d", callID)
	callRes := fmt.Sprintf("%s.call_res.%d", instruction.Result, callID)

	var callArgs []string
	if recursive {
		// Direct recursive call to current closure
		fnPtr = functionSymbol(mangleFunctionName(e.function.Name))
		callArgs = append(callArgs, "ptr %__env_ctx")
	} else {
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, null\n", closureIsNull, closureVar))
		out.WriteString(fmt.Sprintf("  %%%s = icmp eq ptr %%%s, @scriptgo_undefined_sentinel\n", closureIsUndef, closureVar))
		out.WriteString(fmt.Sprintf("  %%%s = or i1 %%%s, %%%s\n", closureInvalid, closureIsNull, closureIsUndef))
		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s, !prof !{!\x22branch_weights\x22, i32 1, i32 10000}\n", closureInvalid, nullBlock, callBlock))
		out.WriteString(fmt.Sprintf("%s:\n", nullBlock))
		out.WriteString(fmt.Sprintf("  br label %%%s\n", contBlock))
		out.WriteString(fmt.Sprintf("%s:\n", callBlock))
	}

	if useValueDispatcher {
		argPointers := []string{"ptr null", "ptr null", "ptr null", "ptr null"}
		for index, arg := range instruction.Args {
			if index >= len(argPointers) {
				return fmt.Errorf("closure call supports at most 4 arguments")
			}
			argType := e.types[arg]
			if argType == "" {
				argType = ir.TypeNumber
			}
			argPointer, err := e.emitCanonicalValuePointer(out, arg, argType, fmt.Sprintf("closure.arg.%d", index))
			if err != nil {
				return err
			}
			argPointers[index] = "ptr " + argPointer
		}
		resultSlot := fmt.Sprintf("closure.result.slot.%d", callID)
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = alloca { i32, i32, i64, i64 }\n", resultSlot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_closure_invoke_value(ptr %%%s, i32 %d, %s, ptr %%%s)\n", status, closureVar, len(instruction.Args), strings.Join(argPointers, ", "), resultSlot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		if instruction.Type == ir.TypeUnknown && instruction.Result != "" {
			e.types[instruction.Result] = ir.TypeUnknown
			out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", callRes, resultSlot))
		}
		out.WriteString(fmt.Sprintf("  br label %%%s\n", contBlock))
		out.WriteString(fmt.Sprintf("%s:\n", contBlock))
		if instruction.Type == ir.TypeUnknown && instruction.Result != "" {
			out.WriteString(fmt.Sprintf("  %%%s = phi { i32, i32, i64, i64 } [ zeroinitializer, %%%s ], [ %%%s, %%%s ]\n", instruction.Result, nullBlock, callRes, callBlock))
		}
		return nil
	}

	if !recursive {
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds { ptr, ptr, ptr, i32 }, ptr %%%s, i32 0, i32 0\n", fnPtrSlot, closureVar))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", fnPtr, fnPtrSlot))
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds { ptr, ptr, ptr, i32 }, ptr %%%s, i32 0, i32 1\n", envSlot, closureVar))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", envCtx, envSlot))
		callArgs = append(callArgs, fmt.Sprintf("ptr %%%s", envCtx))
	}

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
		// Static closure callbacks use the C-compatible flattened ABI. The
		// canonical value remains the source representation inside LLVM, but
		// each callback argument is expanded at the call boundary.
		for field, fieldType := range []string{"i32", "i32", "i64"} {
			fieldName := fmt.Sprintf("closure.call.%d.%d", field, e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, %d\n", fieldName, argVal, field))
			callArgs = append(callArgs, fmt.Sprintf("%s %%%s", fieldType, fieldName))
		}
	}

	for len(callArgs) < 13 { // env + 4 flattened value triples
		callArgs = append(callArgs, "i32 0")
		callArgs = append(callArgs, "i32 0")
		callArgs = append(callArgs, "i64 0")
	}

	retType := llvmType(instruction.Type)
	fnSig := "(ptr, i32, i32, i64, i32, i32, i64, i32, i32, i64, i32, i32, i64) "
	fnTarget := "%" + fnPtr
	if strings.HasPrefix(fnPtr, "@") {
		fnTarget = fnPtr
	}
	if instruction.Result != "" && retType != "void" {
		e.types[instruction.Result] = instruction.Type
		if recursive {
			out.WriteString(fmt.Sprintf("  %%%s = call %s %s%s(%s)\n", instruction.Result, retType, fnSig, fnTarget, strings.Join(callArgs, ", ")))
			return nil
		}
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
		out.WriteString(fmt.Sprintf("  call void %s%s(%s)\n", fnSig, fnTarget, strings.Join(callArgs, ", ")))
		if recursive {
			return nil
		}
		out.WriteString(fmt.Sprintf("  br label %%%s\n", contBlock))
		out.WriteString(fmt.Sprintf("%s:\n", contBlock))
	}
	return nil
}
