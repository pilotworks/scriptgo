package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) tryEmitCoreIntrinsicCall(out *strings.Builder, instruction ir.Instruction) (bool, error) {
	if strings.HasPrefix(instruction.Callee, "__dynamic.") {
		if err := e.emitDynamicIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__console.") {
		if err := e.emitConsoleIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__Math.") {
		if err := emitMathIntrinsic(out, instruction); err != nil {
			return true, err
		}
		e.types[instruction.Result] = instruction.Type
		if e.integerVars != nil && (instruction.Callee == "__Math.floor" || instruction.Callee == "__Math.ceil" || instruction.Callee == "__Math.trunc" || instruction.Callee == "__Math.round") {
			e.integerVars[instruction.Result] = true
		}
		return true, nil
	}
	if instruction.Callee == "__array.isArray" {
		arg := instruction.Args[0]
		argType := e.types[arg]
		if argType == ir.TypeUnknown || e.isParamUnknown(arg) {
			if slot, ok := e.varSlots[arg]; ok {
				loaded := fmt.Sprintf("%s.isarr.loaded.%d", arg, e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded, slot))
				arg = loaded
			}
			tagVar := fmt.Sprintf("isarray.tag.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 0\n", tagVar, arg)
			fmt.Fprintf(out, "  %%%s = icmp eq i32 %%%s, 6\n", instruction.Result, tagVar)
			e.types[instruction.Result] = ir.TypeBool
			return true, nil
		}
		isNotArr := argType == ir.TypeNumber || argType == ir.TypeBool || argType == ir.TypeBigInt
		resSlot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", resSlot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		if !isNotArr {
			argVal := e.resolveArg(out, instruction.Args[0])
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_is_array(ptr %%%s, ptr %%%s)\n", status, argVal, resSlot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_is_array(ptr null, ptr %%%s)\n", status, resSlot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.d = load double, ptr %%%s\n", instruction.Result, resSlot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.d, 0.000000e+00\n", instruction.Result, instruction.Result)
		e.types[instruction.Result] = ir.TypeBool
		return true, nil
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
		return true, nil
	}
	if instruction.Callee == "__clone.structured" {
		arg := instruction.Args[0]
		typ := instruction.Type
		out.WriteString(fmt.Sprintf("  %%%s = bitcast %s %%%s to %s\n", instruction.Result, llvmType(typ), arg, llvmType(typ)))
		e.types[instruction.Result] = typ
		return true, nil
	}
	return false, nil
}

func (e *functionEmitter) tryEmitSystemIntrinsicCall(out *strings.Builder, instruction ir.Instruction) (bool, error) {
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
			return true, fmt.Errorf("unknown array intrinsic argument %q", instruction.Args[0])
		}
		if err := e.emitArrayIntrinsic(out, instruction, arrayType); err != nil {
			return true, err
		}
		e.types[instruction.Result] = instruction.Type
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__fs.") {
		if err := e.emitFsIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__child_process.") {
		if err := e.emitChildProcessIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__http.") {
		if err := e.emitHttpIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__stream.") {
		if err := e.emitStreamIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__websocket.") {
		if err := e.emitWebSocketIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__process.") {
		if err := e.emitProcessIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	return false, nil
}

func (e *functionEmitter) tryEmitServiceIntrinsicCall(out *strings.Builder, instruction ir.Instruction) (bool, error) {
	if strings.HasPrefix(instruction.Callee, "__async.") {
		if err := e.emitAsyncIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__crypto.") {
		if err := e.emitCryptoIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__zlib.") {
		if err := e.emitZlibIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__date.") {
		if err := e.emitDateIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__os.") {
		if err := e.emitOsIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__error.") {
		if err := e.emitErrorIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
			if instruction.Type == ir.TypeString {
				e.ownedStrings = append(e.ownedStrings, instruction.Result)
			}
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__tty.") {
		if err := e.emitTtyIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
			if instruction.Type == ir.TypeString {
				e.ownedStrings = append(e.ownedStrings, instruction.Result)
			}
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__sqlite.") {
		if err := e.emitSqliteIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__web.") {
		if err := e.emitWebIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__performance.") {
		if err := e.emitPerformanceIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__json.") {
		if err := e.emitJsonIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
			if instruction.Type == ir.TypeString {
				e.ownedStrings = append(e.ownedStrings, instruction.Result)
			}
		}
		return true, nil
	}
	if instruction.Callee == "__scriptgo.is_truthy" {
		arg := instruction.Args[0]
		argType := e.types[arg]
		valuePtr, err := e.emitCanonicalValuePointer(out, arg, argType, "truthy.value")
		if err != nil {
			return true, err
		}
		i32Res := fmt.Sprintf("truthy.i32.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_is_truthy_unknown(ptr %s)\n", i32Res, valuePtr))
		out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Res))
		e.types[instruction.Result] = ir.TypeBool
		return true, nil
	}
	return false, nil
}

func (e *functionEmitter) tryEmitValueIntrinsicCall(out *strings.Builder, instruction ir.Instruction) (bool, error) {
	if strings.HasPrefix(instruction.Callee, "__object.") {
		if err := e.emitObjectIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__number.") {
		if err := e.emitNumberIntrinsic(out, instruction); err != nil {
			return true, err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__string.") {
		if err := e.emitStringIntrinsic(out, instruction); err != nil {
			return true, err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__regex.") || strings.HasPrefix(instruction.Callee, "__regexp.") {
		if err := e.emitRegexIntrinsic(out, instruction); err != nil {
			return true, err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__bigint.") {
		if err := emitBigIntIntrinsic(out, instruction); err != nil {
			return true, err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__gc.") {
		if err := emitGcIntrinsic(out, instruction); err != nil {
			return true, err
		}
		e.types[instruction.Result] = instruction.Type
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__weak") || strings.HasPrefix(instruction.Callee, "__finalization_registry.") {
		if err := e.emitWeakIntrinsic(out, instruction); err != nil {
			return true, err
		}
		e.types[instruction.Result] = instruction.Type
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__atomics.") {
		if err := e.emitAtomicsIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__intl.") {
		if err := e.emitIntlIntrinsic(out, instruction); err != nil {
			return true, err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__symbol.") {
		if err := emitSymbolIntrinsic(out, instruction); err != nil {
			return true, err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__typedarray.") || strings.HasPrefix(instruction.Callee, "__arraybuffer.") || strings.HasPrefix(instruction.Callee, "__arraybuffer_view.") || strings.HasPrefix(instruction.Callee, "__dataview.") {
		if err := e.emitTypedArrayIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__buffer.") {
		if err := e.emitBufferIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__timers.") {
		if err := e.emitTimerIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__map.") {
		if err := e.emitMapIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
			if instruction.Type == ir.TypeString {
				e.ownedStrings = append(e.ownedStrings, instruction.Result)
			}
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__set.") {
		if err := e.emitSetIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
			if instruction.Type == ir.TypeString {
				e.ownedStrings = append(e.ownedStrings, instruction.Result)
			}
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__text_encoder.") || strings.HasPrefix(instruction.Callee, "__text_decoder.") {
		if err := e.emitTextEncodingIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
			if instruction.Type == ir.TypeString {
				e.ownedStrings = append(e.ownedStrings, instruction.Result)
			}
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__dns.") {
		if err := e.emitDnsIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__net.") {
		if err := e.emitNetIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__dgram.") {
		if err := e.emitDgramIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return true, nil
	}
	if strings.HasPrefix(instruction.Callee, "__tls.") {
		if err := e.emitTLSIntrinsic(out, instruction); err != nil {
			return true, err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
			if instruction.Type == ir.TypeString {
				e.ownedStrings = append(e.ownedStrings, instruction.Result)
			}
		}
		return true, nil
	}
	return false, nil
}
