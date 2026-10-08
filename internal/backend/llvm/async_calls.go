package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

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
			fn := "scriptgo_promise_resolve_existing_boxed"
			boxed := fmt.Sprintf("promise.existing.box.%d", e.loadCounter)
			e.loadCounter++
			if err := e.emitBoxValue(out, argVal, argType, boxed); err != nil {
				return err
			}
			tag := fmt.Sprintf("promise.existing.tag.%d", e.loadCounter)
			payload := fmt.Sprintf("promise.existing.payload.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 0\n", tag, boxed))
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payload, boxed))
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
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 0\n", tagName, unknownVal)
			e.tempCounter++
			payloadName := fmt.Sprintf("await.unknown.payload.%d", e.tempCounter)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadName, unknownVal)
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
			fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 %%%s, 0\n", b0, resultTag)
			fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i32 0, 1\n", b1, b0)
			fmt.Fprintf(out, "  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i64 %%%s, 2\n", rawResult, b1, resultPayload)
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
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 %%%s, 0\n", b0, tagVal))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i32 0, 1\n", b1, b0))
			out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, b1, payloadVal))
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
				out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadName, instruction.Args[0]))
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
	boxedSlot := fmt.Sprintf("async.frame.value.slot.%d", e.loadCounter)
	e.loadCounter++
	out.WriteString(fmt.Sprintf("  %%%s = alloca { i32, i32, i64, i64 }\n", boxedSlot))
	out.WriteString(fmt.Sprintf("  store { i32, i32, i64, i64 } %%%s, ptr %%%s\n", boxed, boxedSlot))
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_async_frame_set(ptr %%%s, i64 %d, ptr %%%s)\n", status, frame, index, boxedSlot))
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
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 0\n", tag, boxed))
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payload, boxed))
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
