package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitStringIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	status := instruction.Result + ".status"
	switch instruction.Callee {
	case "__string.length":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("string.length has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_length(ptr %%%s, ptr %%__slot_double)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%__slot_double\n", instruction.Result)
		if e.integerVars != nil {
			e.integerVars[instruction.Result] = true
		}
	case "__string.indexOf", "__string.lastIndexOf", "__string.startsWith", "__string.endsWith", "__string.includes", "__string.replace", "__string.replaceAll", "__string.split", "__string.match", "__string.search", "__string.replace_regex", "__string.matchAll", "__string.localeCompare":
		return e.emitStringSearchIntrinsic(out, instruction, status)

	case "__string.fromNumber", "__string.fromBool", "__string.fromUnknown", "__string.fromObject", "__string.inspectObject", "__string.inspectBuffer", "__string.inspectArray", "__string.fromBigInt", "__string.fromBigIntLocale", "__string.errorToString", "__string.fromCodePoint", "__string.fromCharCode", "__string.encodeURIComponent", "__string.decodeURIComponent", "__string.encodeURI", "__string.decodeURI", "__string.raw", "__string.new":
		return e.emitStringConversionIntrinsic(out, instruction, status)

	case "__string.slice", "__string.substring":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.slice has invalid signature")
		}
		endArg := "1000000000.0"
		if len(instruction.Args) == 3 {
			endArg = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_slice(ptr %%%s, double %%%s, double %s, ptr %%__slot_ptr)\n", status, instruction.Args[0], instruction.Args[1], endArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.trim":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.trim has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_trim(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.trimStart":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.trimStart has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_trim_start(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.trimEnd":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.trimEnd has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_trim_end(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.charAt":
		if len(instruction.Args) < 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.charAt has invalid signature")
		}
		pos := "0.0"
		if len(instruction.Args) >= 2 {
			pos = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_char_at(ptr %%%s, double %s, ptr %%__slot_ptr)\n", status, instruction.Args[0], pos)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.charCodeAt":
		if len(instruction.Args) < 1 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("string.charCodeAt has invalid signature")
		}
		pos := "0.0"
		if len(instruction.Args) >= 2 {
			pos = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_char_code_at(ptr %%%s, double %s, ptr %%__slot_double)\n", status, instruction.Args[0], pos)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%__slot_double\n", instruction.Result)

	case "__string.toLowerCase":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.toLowerCase has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_to_lower(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.toUpperCase":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.toUpperCase has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_to_upper(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.repeat":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.repeat has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_repeat(ptr %%%s, double %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)

	case "__string.padStart":
		if len(instruction.Args) < 2 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.padStart has invalid signature")
		}
		padArg := "null"
		if len(instruction.Args) >= 3 {
			padArg = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_pad_start(ptr %%%s, double %%%s, ptr %s, ptr %%__slot_ptr)\n", status, instruction.Args[0], instruction.Args[1], padArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.padEnd":
		if len(instruction.Args) < 2 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.padEnd has invalid signature")
		}
		padArg := "null"
		if len(instruction.Args) >= 3 {
			padArg = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_pad_end(ptr %%%s, double %%%s, ptr %s, ptr %%__slot_ptr)\n", status, instruction.Args[0], instruction.Args[1], padArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.concat":
		if len(instruction.Args) < 2 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.concat has invalid signature")
		}
		current := instruction.Args[0]
		for i := 1; i < len(instruction.Args); i++ {
			stepResult := instruction.Result
			if i < len(instruction.Args)-1 {
				stepResult = fmt.Sprintf("%s.step.%d", instruction.Result, i)
			}
			fmt.Fprintf(out, "  %%%s.status = call i32 @scriptgo_string_concat(ptr %%%s, ptr %%%s, ptr %%__slot_ptr)\n", stepResult, current, instruction.Args[i])
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s.status)\n", stepResult)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", stepResult)
			current = stepResult
		}

	case "__string.codePointAt":
		if (len(instruction.Args) != 1 && len(instruction.Args) != 2) || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("string.codePointAt has invalid signature")
		}
		pos := "0.0"
		if len(instruction.Args) == 2 {
			pos = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_code_point_at(ptr %%%s, double %s, ptr %%__slot_double)\n", status, instruction.Args[0], pos)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%__slot_double\n", instruction.Result)

	case "__string.isWellFormed":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("string.isWellFormed has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_is_well_formed(ptr %%%s, ptr %%__slot_double)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%__slot_double\n", instruction.Result)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
	case "__string.toWellFormed":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.toWellFormed has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_to_well_formed(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)

	case "__string.trimLeft", "__string.trimRight", "__string.substr", "__string.anchor", "__string.big", "__string.blink", "__string.bold", "__string.fixed", "__string.fontcolor", "__string.fontsize", "__string.italics", "__string.link", "__string.small", "__string.strike", "__string.sub", "__string.sup":
		return e.emitStringAnnexBIntrinsic(out, instruction, status)

	case "__string.toLocaleLowerCase":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_to_lower(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.toLocaleUpperCase":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_to_upper(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)

	case "__string.at":
		pos := "0.0"
		if len(instruction.Args) >= 2 {
			pos = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_at(ptr %%%s, double %s, ptr %%__slot_ptr)\n", status, instruction.Args[0], pos)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)

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
