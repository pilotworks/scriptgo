package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitTextEncodingIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__text_encoder.new":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_text_encoder_new(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__text_encoder.encoding":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("text_encoder.encoding requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_text_encoder_encoding(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__text_encoder.encode":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)

		inputArg := "null"
		if len(instruction.Args) > 0 {
			inputArg = "%" + instruction.Args[0]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_text_encoder_encode(ptr null, ptr %s, ptr %%%s)\n", status, inputArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__text_encoder.encode_into":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("text_encoder.encode_into requires 2 arguments")
		}
		srcArg := instruction.Args[0]
		destArg := instruction.Args[1]

		readSlot := instruction.Result + ".read_slot"
		writtenSlot := instruction.Result + ".written_slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++

		fmt.Fprintf(out, "  %%%s = alloca double\n", readSlot)
		fmt.Fprintf(out, "  %%%s = alloca double\n", writtenSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_text_encoder_encode_into(ptr null, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, srcArg, destArg, readSlot, writtenSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)

		readVal := instruction.Result + ".read_val"
		writtenVal := instruction.Result + ".written_val"
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", readVal, readSlot)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", writtenVal, writtenSlot)

		// Create standard ScriptGo object with 2 fields
		objSlot := instruction.Result + ".obj_slot"
		objStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", objSlot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_new(i64 2, ptr %%%s)\n", objStatus, objSlot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", objStatus)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, objSlot)

		setReadStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 0, double %%%s)\n", setReadStatus, instruction.Result, readVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", setReadStatus)

		setWrittenStatus := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_object_number_set(ptr %%%s, i64 1, double %%%s)\n", setWrittenStatus, instruction.Result, writtenVal)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", setWrittenStatus)
		return nil

	case "__text_decoder.new":
		labelArg := "null"
		fatalVal := 0
		ignoreBOMVal := 0

		if len(instruction.Args) > 0 {
			labelArg = "%" + instruction.Args[0]
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_text_decoder_new(ptr %s, i32 %d, i32 %d, ptr %%%s)\n", status, labelArg, fatalVal, ignoreBOMVal, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__text_decoder.encoding":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("text_decoder.encoding requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_text_decoder_encoding(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__text_decoder.fatal":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("text_decoder.fatal requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_text_decoder_fatal(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		i32Val := instruction.Result + ".i32"
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", i32Val, slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Val)
		return nil

	case "__text_decoder.ignore_bom":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("text_decoder.ignore_bom requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_text_decoder_ignore_bom(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		i32Val := instruction.Result + ".i32"
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", i32Val, slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Val)
		return nil

	case "__text_decoder.decode":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("text_decoder.decode requires receiver argument")
		}
		decArg := instruction.Args[0]
		inputArg := "null"
		if len(instruction.Args) > 1 {
			inputArg = "%" + instruction.Args[1]
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_text_decoder_decode(ptr %%%s, ptr %s, ptr %%%s)\n", status, decArg, inputArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	default:
		return fmt.Errorf("unknown text encoding intrinsic %s", instruction.Callee)
	}
}
