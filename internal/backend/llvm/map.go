package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitMapIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__map.new":
		e.emitMapCall(out, instruction.Result, "ptr", "scriptgo_map_new", "")
		return nil

	case "__map.new_entries":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("map.new_entries requires 1 argument")
		}
		arrArg := e.ensurePointerArg(out, instruction.Args[0])
		e.emitMapCall(out, instruction.Result, "ptr", "scriptgo_map_new_entries", "ptr %"+arrArg)
		return nil

	case "__map.set":
		if len(instruction.Args) < 3 {
			return fmt.Errorf("map.set requires 3 arguments")
		}
		mapArg := e.ensurePointerArg(out, instruction.Args[0])
		keyTag, keyPayload, err := e.mapKeyArgs(out, instruction.Args[1])
		if err != nil {
			return err
		}
		origVal := instruction.Args[2]
		var callee, valArg string
		switch e.types[origVal] {
		case ir.TypeNumber:
			callee, valArg = "scriptgo_map_set_number", "double %"+e.resolveArg(out, origVal)
		case ir.TypeString:
			callee, valArg = "scriptgo_map_set_string", "ptr %"+e.resolveArg(out, origVal)
		case ir.TypeBigInt:
			callee, valArg = "scriptgo_map_set_bigint", "i64 %"+e.resolveArg(out, origVal)
		case ir.TypeBool:
			flag := fmt.Sprintf("map.value.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", flag, e.resolveArg(out, origVal))
			callee, valArg = "scriptgo_map_set_bool", "i32 %"+flag
		case ir.TypeUnknown:
			boxed := e.resolveArg(out, origVal)
			value := fmt.Sprintf("map.value.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s.tag = extractvalue { i32, i32, i64, i64 } %%%s, 0\n", value, boxed)
			fmt.Fprintf(out, "  %%%s.payload = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", value, boxed)
			callee, valArg = "scriptgo_map_set_unknown", fmt.Sprintf("i32 %%%s.tag, i64 %%%s.payload", value, value)
		default:
			callee, valArg = "scriptgo_map_set_ptr", "ptr %"+e.ensurePointerArg(out, origVal)
		}
		e.emitMapCall(out, instruction.Result, "ptr", callee, fmt.Sprintf("ptr %%%s, i32 %s, i64 %s, %s", mapArg, keyTag, keyPayload, valArg))
		return nil

	case "__map.get":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("map.get requires 2 arguments")
		}
		mapArg := e.ensurePointerArg(out, instruction.Args[0])
		keyTag, keyPayload, err := e.mapKeyArgs(out, instruction.Args[1])
		if err != nil {
			return err
		}
		args := fmt.Sprintf("ptr %%%s, i32 %s, i64 %s", mapArg, keyTag, keyPayload)
		switch instruction.Type {
		case ir.TypeNumber:
			e.emitMapCall(out, instruction.Result, "double", "scriptgo_map_get_number", args)
		case ir.TypeString:
			e.emitMapCall(out, instruction.Result, "ptr", "scriptgo_map_get_string", args)
		case ir.TypeBigInt:
			e.emitMapCall(out, instruction.Result, "i64", "scriptgo_map_get_bigint", args)
		case ir.TypeBool:
			e.emitMapCall(out, instruction.Result+".i32", "i32", "scriptgo_map_get_bool", args)
			fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, instruction.Result+".i32")
		case ir.TypeUnknown:
			e.emitMapCall(out, instruction.Result, "{ i32, i32, i64, i64 }", "scriptgo_map_get_unknown", args)
		default:
			e.emitMapCall(out, instruction.Result, "ptr", "scriptgo_map_get_ptr", args)
		}
		return nil

	case "__map.has", "__map.delete":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("%s requires 2 arguments", instruction.Callee)
		}
		mapArg := e.ensurePointerArg(out, instruction.Args[0])
		keyTag, keyPayload, err := e.mapKeyArgs(out, instruction.Args[1])
		if err != nil {
			return err
		}
		callee := "scriptgo_map_has"
		if instruction.Callee == "__map.delete" {
			callee = "scriptgo_map_delete"
		}
		e.emitMapCall(out, instruction.Result+".i32", "i32", callee, fmt.Sprintf("ptr %%%s, i32 %s, i64 %s", mapArg, keyTag, keyPayload))
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, instruction.Result+".i32")
		return nil

	case "__map.clear", "__map.forEach":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("%s requires a receiver", instruction.Callee)
		}
		mapArg := e.ensurePointerArg(out, instruction.Args[0])
		callee, args := "scriptgo_map_clear", "ptr %"+mapArg
		if instruction.Callee == "__map.forEach" {
			if len(instruction.Args) < 2 {
				return fmt.Errorf("map.forEach requires 2 arguments")
			}
			callee, args = "scriptgo_map_for_each", fmt.Sprintf("ptr %%%s, ptr %%%s", mapArg, instruction.Args[1])
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @%s(%s)\n", status, callee, args)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__map.size", "__map.toString", "__map.values", "__map.entries", "__map.keys":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("%s requires a receiver", instruction.Callee)
		}
		mapArg := e.ensurePointerArg(out, instruction.Args[0])
		resultType, callee, args := "ptr", "scriptgo_map_"+strings.TrimPrefix(instruction.Callee, "__map."), "ptr %"+mapArg
		switch instruction.Callee {
		case "__map.size":
			resultType = "double"
		case "__map.toString":
			callee = "scriptgo_map_to_string"
		case "__map.keys":
			// The runtime lays keys out for the element type the program reads.
			elementSize, err := arrayElementSizeForTarget(instruction.Type, e.pointerSize())
			if err != nil {
				return err
			}
			args += fmt.Sprintf(", i64 %d", elementSize)
		}
		e.emitMapCall(out, instruction.Result, resultType, callee, args)
		return nil

	default:
		return fmt.Errorf("unsupported Map intrinsic %q", instruction.Callee)
	}
}

// emitMapCall calls a Map runtime function whose last parameter is the out
// slot for result, aborting on a failed status.
func (e *functionEmitter) emitMapCall(out *strings.Builder, result, resultType, callee, args string) {
	slot := result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	if args != "" {
		args += ", "
	}
	fmt.Fprintf(out, "  %%%s = alloca %s\n", slot, resultType)
	fmt.Fprintf(out, "  %%%s = call i32 @%s(%sptr %%%s)\n", status, callee, args, slot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load %s, ptr %%%s\n", result, resultType, slot)
}

// mapKeyArgs passes a Map key as the value tag and payload of its boxed
// form, so the runtime compares keys with SameValueZero by their own type:
// the number 1 and the string "1" are distinct keys.
func (e *functionEmitter) mapKeyArgs(out *strings.Builder, key string) (tag, payload string, err error) {
	name := fmt.Sprintf("map.key.%d", e.loadCounter)
	e.loadCounter++
	boxed := name
	if keyType := e.types[key]; keyType == ir.TypeUnknown {
		boxed = e.resolveArg(out, key)
	} else if err := e.emitBoxValue(out, key, keyType, boxed); err != nil {
		return "", "", err
	}
	fmt.Fprintf(out, "  %%%s.tag = extractvalue { i32, i32, i64, i64 } %%%s, 0\n", name, boxed)
	fmt.Fprintf(out, "  %%%s.payload = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", name, boxed)
	return "%" + name + ".tag", "%" + name + ".payload", nil
}
