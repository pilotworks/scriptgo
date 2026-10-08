package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitJsonIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	argVal := instruction.Args[0]
	if len(instruction.Args) > 0 {
		argVal = e.resolveArg(out, instruction.Args[0])
	}
	argType := e.types[argVal]
	if argType == "" {
		argType = e.types[instruction.Args[0]]
	}
	switch instruction.Callee {
	case "__json.stringify_number":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify number has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_number(double %%%s, ptr %%%s)\n", status, argVal, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.stringify_bool":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify bool has invalid signature")
		}
		boolVal := fmt.Sprintf("%s.bool_i32.%d", instruction.Result, e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", boolVal, argVal)
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_bool(i32 %%%s, ptr %%%s)\n", status, boolVal, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.stringify_string":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify string has invalid signature")
		}
		if argType == ir.TypeUnknown {
			payloadName := fmt.Sprintf("json.str.payload.%d", e.loadCounter)
			ptrName := fmt.Sprintf("json.str.ptr.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadName, argVal)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrName, payloadName)
			argVal = ptrName
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_string(ptr %%%s, ptr %%%s)\n", status, argVal, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.stringify_number_array":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify number array has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_number_array(ptr %%%s, ptr %%%s)\n", status, argVal, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.stringify_string_array":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify string array has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_string_array(ptr %%%s, ptr %%%s)\n", status, argVal, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.stringify_object_array":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify object array has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_object_array(ptr %%%s, ptr %%%s)\n", status, argVal, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.stringify_unknown":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("JSON.stringify unknown has invalid signature")
		}
		if argType != ir.TypeUnknown {
			boxedVar := fmt.Sprintf("json.box.%d", e.loadCounter)
			e.loadCounter++
			if err := e.emitBoxValue(out, argVal, argType, boxedVar); err != nil {
				return err
			}
			argVal = boxedVar
		}
		valuePtr, err := e.emitCanonicalValuePointer(out, argVal, ir.TypeUnknown, "json.value")
		if err != nil {
			return err
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_stringify_unknown(ptr %s, ptr %%%s)\n", status, valuePtr, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil
	case "__json.parse_unknown":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeUnknown {
			return fmt.Errorf("JSON.parse has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca { i32, i32, i64, i64 }\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_parse_unknown(ptr %%%s, ptr %%%s)\n", status, argVal, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", instruction.Result, slot)
		return nil
	default:
		return fmt.Errorf("unknown JSON intrinsic %q", instruction.Callee)
	}
}
