package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitMapIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__map.new":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_new(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__map.new_entries":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("map.new_entries requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_new_entries(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__map.set":
		if len(instruction.Args) < 3 {
			return fmt.Errorf("map.set requires 3 arguments")
		}
		mapArg := instruction.Args[0]
		keyArg := instruction.Args[1]
		valArg := instruction.Args[2]
		keyType := e.types[keyArg]
		valType := e.types[valArg]

		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)

		if keyType == ir.TypeNumber {
			if valType == ir.TypeNumber {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_set_number_number(ptr %%%s, double %%%s, double %%%s, ptr %%%s)\n", status, mapArg, keyArg, valArg, slot)
			} else if valType == ir.TypeString {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_set_number_string(ptr %%%s, double %%%s, ptr %%%s, ptr %%%s)\n", status, mapArg, keyArg, valArg, slot)
			} else {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_set_number_ptr(ptr %%%s, double %%%s, ptr %%%s, ptr %%%s)\n", status, mapArg, keyArg, valArg, slot)
			}
		} else {
			if valType == ir.TypeNumber {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_set_string_number(ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n", status, mapArg, keyArg, valArg, slot)
			} else if valType == ir.TypeString {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_set_string_string(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, mapArg, keyArg, valArg, slot)
			} else {
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_set_string_ptr(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, mapArg, keyArg, valArg, slot)
			}
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__map.get":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("map.get requires 2 arguments")
		}
		mapArg := instruction.Args[0]
		keyArg := instruction.Args[1]
		keyType := e.types[keyArg]
		retType := instruction.Type

		keyIsStr := "1"
		keyStrArg := fmt.Sprintf("%%%s", keyArg)
		keyNumArg := "0.0"
		if keyType == ir.TypeNumber {
			keyIsStr = "0"
			keyStrArg = "null"
			keyNumArg = fmt.Sprintf("%%%s", keyArg)
		}

		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++

		if retType == ir.TypeNumber {
			fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_get_number(ptr %%%s, ptr %s, double %s, i32 %s, ptr %%%s)\n", status, mapArg, keyStrArg, keyNumArg, keyIsStr, slot)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
			fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		} else if retType == ir.TypeString {
			fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_get_string(ptr %%%s, ptr %s, double %s, i32 %s, ptr %%%s)\n", status, mapArg, keyStrArg, keyNumArg, keyIsStr, slot)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		} else {
			fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_get_ptr(ptr %%%s, ptr %s, double %s, i32 %s, ptr %%%s)\n", status, mapArg, keyStrArg, keyNumArg, keyIsStr, slot)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		}
		return nil

	case "__map.has":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("map.has requires 2 arguments")
		}
		mapArg := instruction.Args[0]
		keyArg := instruction.Args[1]
		keyType := e.types[keyArg]

		keyIsStr := "1"
		keyStrArg := fmt.Sprintf("%%%s", keyArg)
		keyNumArg := "0.0"
		if keyType == ir.TypeNumber {
			keyIsStr = "0"
			keyStrArg = "null"
			keyNumArg = fmt.Sprintf("%%%s", keyArg)
		}

		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_has(ptr %%%s, ptr %s, double %s, i32 %s, ptr %%%s)\n", status, mapArg, keyStrArg, keyNumArg, keyIsStr, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", instruction.Result+".i32", slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, instruction.Result+".i32")
		return nil

	case "__map.delete":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("map.delete requires 2 arguments")
		}
		mapArg := instruction.Args[0]
		keyArg := instruction.Args[1]
		keyType := e.types[keyArg]

		keyIsStr := "1"
		keyStrArg := fmt.Sprintf("%%%s", keyArg)
		keyNumArg := "0.0"
		if keyType == ir.TypeNumber {
			keyIsStr = "0"
			keyStrArg = "null"
			keyNumArg = fmt.Sprintf("%%%s", keyArg)
		}

		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_delete(ptr %%%s, ptr %s, double %s, i32 %s, ptr %%%s)\n", status, mapArg, keyStrArg, keyNumArg, keyIsStr, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", instruction.Result+".i32", slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, instruction.Result+".i32")
		return nil

	case "__map.clear":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("map.clear requires 1 argument")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_clear(ptr %%%s)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__map.size":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("map.size requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_size(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__map.toString":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("map.toString requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_to_string(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__map.forEach":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("map.forEach requires 2 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_for_each(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__map.keys":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("map.keys requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_keys(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__map.values":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("map.values requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_values(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__map.entries":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("map.entries requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_map_entries(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	default:
		return fmt.Errorf("unsupported Map intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitSetIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__set.new":
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_new(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__set.new_values":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("set.new_values requires 1 argument")
		}
		arrArg := instruction.Args[0]
		arrType := e.types[arrArg]
		fnName := "scriptgo_set_new_values_ptr"
		if arrType == ir.TypeNumberArray {
			fnName = "scriptgo_set_new_values_number"
		} else if arrType == ir.TypeStringArray {
			fnName = "scriptgo_set_new_values_string"
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, ptr %%%s)\n", status, fnName, arrArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__set.add":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("set.add requires 2 arguments")
		}
		setArg := instruction.Args[0]
		valArg := instruction.Args[1]
		valType := e.types[valArg]

		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)

		if valType == ir.TypeNumber {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_add_number(ptr %%%s, double %%%s, ptr %%%s)\n", status, setArg, valArg, slot)
		} else if valType == ir.TypeString {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_add_string(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, setArg, valArg, slot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_add_ptr(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, setArg, valArg, slot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__set.has":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("set.has requires 2 arguments")
		}
		setArg := instruction.Args[0]
		valArg := instruction.Args[1]
		valType := e.types[valArg]

		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)

		if valType == ir.TypeNumber {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_has_number(ptr %%%s, double %%%s, ptr %%%s)\n", status, setArg, valArg, slot)
		} else if valType == ir.TypeString {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_has_string(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, setArg, valArg, slot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_has_ptr(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, setArg, valArg, slot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", instruction.Result+".i32", slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, instruction.Result+".i32")
		return nil

	case "__set.delete":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("set.delete requires 2 arguments")
		}
		setArg := instruction.Args[0]
		valArg := instruction.Args[1]
		valType := e.types[valArg]

		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca i32\n", slot)

		if valType == ir.TypeNumber {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_delete_number(ptr %%%s, double %%%s, ptr %%%s)\n", status, setArg, valArg, slot)
		} else if valType == ir.TypeString {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_delete_string(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, setArg, valArg, slot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_delete_ptr(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, setArg, valArg, slot)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load i32, ptr %%%s\n", instruction.Result+".i32", slot)
		fmt.Fprintf(out, "  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, instruction.Result+".i32")
		return nil

	case "__set.clear":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("set.clear requires 1 argument")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_clear(ptr %%%s)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__set.size":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("set.size requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_size(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__set.toString":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("set.toString requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_to_string(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__set.forEach":
		if len(instruction.Args) < 2 {
			return fmt.Errorf("set.forEach requires 2 arguments")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_set_for_each(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	default:
		return fmt.Errorf("unsupported Set intrinsic %q", instruction.Callee)
	}
}
