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
	case "__string.trimStart":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.trimStart has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_trim_start(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.trimEnd":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.trimEnd has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_trim_end(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.charAt":
		if len(instruction.Args) < 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.charAt has invalid signature")
		}
		pos := "0.0"
		if len(instruction.Args) >= 2 {
			pos = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_char_at(ptr %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], pos, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.charCodeAt":
		if len(instruction.Args) < 1 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("string.charCodeAt has invalid signature")
		}
		pos := "0.0"
		if len(instruction.Args) >= 2 {
			pos = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_char_code_at(ptr %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], pos, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case "__string.includes":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeBool {
			return fmt.Errorf("string.includes has invalid signature")
		}
		pos := "0.0"
		if len(instruction.Args) == 3 {
			pos = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_includes(ptr %%%s, ptr %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], pos, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
	case "__string.toLowerCase":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.toLowerCase has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_to_lower(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.toUpperCase":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.toUpperCase has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_to_upper(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.repeat":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.repeat has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_repeat(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
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
	case "__string.replaceAll":
		if len(instruction.Args) != 3 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.replaceAll has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_replace_all(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.padStart":
		if len(instruction.Args) < 2 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.padStart has invalid signature")
		}
		padArg := "null"
		if len(instruction.Args) >= 3 {
			padArg = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_pad_start(ptr %%%s, double %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], padArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.padEnd":
		if len(instruction.Args) < 2 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.padEnd has invalid signature")
		}
		padArg := "null"
		if len(instruction.Args) >= 3 {
			padArg = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_pad_end(ptr %%%s, double %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], padArg, slot)
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
	case "__string.split":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeStringArray {
			return fmt.Errorf("string.split has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_split(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.fromBigInt":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.fromBigInt has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_bigint(i64 %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.match":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("string.match has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_match(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.search":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("string.search has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_search(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case "__string.replace_regex":
		if len(instruction.Args) != 4 {
			return fmt.Errorf("string.replace_regex has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_replace_regex(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.codePointAt":
		if (len(instruction.Args) != 1 && len(instruction.Args) != 2) || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("string.codePointAt has invalid signature")
		}
		pos := "0.0"
		if len(instruction.Args) == 2 {
			pos = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_code_point_at(ptr %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], pos, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case "__string.fromCodePoint":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.fromCodePoint has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_code_point(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.isWellFormed":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("string.isWellFormed has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_is_well_formed(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
	case "__string.toWellFormed":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.toWellFormed has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_to_well_formed(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.matchAll":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("string.matchAll has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_match(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.trimLeft":
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_trim_start(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.trimRight":
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_trim_end(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.toLocaleLowerCase":
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_to_lower(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.toLocaleUpperCase":
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_to_upper(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.fromCharCode":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("string.fromCharCode requires 1 arg")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_code_point(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.at", "__string.substr":
		pos := "0.0"
		if len(instruction.Args) >= 2 {
			pos = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_char_at(ptr %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], pos, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__string.localeCompare":
		other := "\"\""
		if len(instruction.Args) >= 2 {
			other = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_index_of(ptr %%%s, ptr %s, double 0.0, ptr %%%s)\n", status, instruction.Args[0], other, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	default:
		if strings.HasPrefix(instruction.Callee, "__string.") {
			if instruction.Type == ir.TypeString {
				if len(instruction.Args) > 0 {
					fmt.Fprintf(out, "  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, instruction.Args[0])
				} else {
					fmt.Fprintf(out, "  %%%s = alloca i8\n", instruction.Result)
				}
				return nil
			}
			if instruction.Type == ir.TypeNumber {
				fmt.Fprintf(out, "  %%%s = fadd double 0.0, 0.0\n", instruction.Result)
				return nil
			}
			if instruction.Type == ir.TypeBool {
				fmt.Fprintf(out, "  %%%s = icmp eq i32 1, 1\n", instruction.Result)
				return nil
			}
		}
		return fmt.Errorf("unknown string intrinsic %q", instruction.Callee)
	}
	return nil
}
