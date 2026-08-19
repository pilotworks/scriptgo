package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func emitStringIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	status := instruction.Result + ".status"
	slot := instruction.Result + ".slot"
	switch instruction.Callee {
	case "__string.length":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("string.length has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_length(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case "__string.indexOf":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("string.indexOf has invalid signature")
		}
		position := "0.0"
		if len(instruction.Args) == 3 {
			position = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_index_of(ptr %%%s, ptr %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], position, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case "__string.lastIndexOf":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("string.lastIndexOf has invalid signature")
		}
		position := "-1.0"
		if len(instruction.Args) == 3 {
			position = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_last_index(ptr %%%s, ptr %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], position, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case "__string.startsWith":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("string.startsWith has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_starts_with(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
	case "__string.endsWith":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("string.endsWith has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_ends_with(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
	case "__string.fromNumber":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.fromNumber has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_number(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.fromBool":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.fromBool has invalid signature")
		}
		boolI32 := instruction.Result + ".i32"
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", boolI32, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_bool(i32 %%%s, ptr %%%s)\n", status, boolI32, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.slice", "__string.substring":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.slice has invalid signature")
		}
		endArg := "-1.0"
		if len(instruction.Args) == 3 {
			endArg = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_slice(ptr %%%s, double %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], endArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.trim":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.trim has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_trim(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.replace":
		if len(instruction.Args) != 3 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.replace has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_replace(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.concat":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.concat has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_concat(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	default:
		return fmt.Errorf("unknown string intrinsic %q", instruction.Callee)
	}
	return nil
}

func emitArrayIntrinsic(out *strings.Builder, instruction ir.Instruction, arrayType ir.Type) error {
	if instruction.Callee != "__array.length" || len(instruction.Args) != 1 || instruction.Type != ir.TypeNumber {
		return fmt.Errorf("array.length has invalid signature")
	}
	_ = arrayType
	length := "scriptgo_array_length"
	resultSlot := instruction.Result + ".slot"
	status := instruction.Result + ".status"
	fmt.Fprintf(out, "  %%%s = alloca i64\n", resultSlot)
	fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, ptr %%%s)\n", status, length, instruction.Args[0], resultSlot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s.i64 = load i64, ptr %%%s\n", instruction.Result, resultSlot)
	fmt.Fprintf(out, "  %%%s = uitofp i64 %%%s.i64 to double\n", instruction.Result, instruction.Result)
	return nil
}
