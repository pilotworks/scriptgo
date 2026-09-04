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
	case "__string.indexOf":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("string.indexOf has invalid signature")
		}
		position := "0.0"
		if len(instruction.Args) == 3 {
			position = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_index_of(ptr %%%s, ptr %%%s, double %s, ptr %%__slot_double)\n", status, instruction.Args[0], instruction.Args[1], position)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%__slot_double\n", instruction.Result)
	case "__string.lastIndexOf":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("string.lastIndexOf has invalid signature")
		}
		position := "-1.0"
		if len(instruction.Args) == 3 {
			position = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_last_index(ptr %%%s, ptr %%%s, double %s, ptr %%__slot_double)\n", status, instruction.Args[0], instruction.Args[1], position)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%__slot_double\n", instruction.Result)
	case "__string.startsWith":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("string.startsWith has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_starts_with(ptr %%%s, ptr %%%s, ptr %%__slot_double)\n", status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%__slot_double\n", instruction.Result)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
	case "__string.endsWith":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("string.endsWith has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_ends_with(ptr %%%s, ptr %%%s, ptr %%__slot_double)\n", status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%__slot_double\n", instruction.Result)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
	case "__string.fromNumber":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.fromNumber has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_number(double %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.fromBool":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.fromBool has invalid signature")
		}
		boolI32 := instruction.Result + ".i32"
		fmt.Fprintf(out, "  %%%s = zext i1 %%%s to i32\n", boolI32, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_bool(i32 %%%s, ptr %%__slot_ptr)\n", status, boolI32)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.fromUnknown":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.fromUnknown has invalid signature")
		}
		arg := instruction.Args[0]
		argType := e.types[arg]
		if slot, ok := e.varSlots[arg]; ok {
			loaded := fmt.Sprintf("%s.str_load.%d", arg, e.loadCounter)
			e.loadCounter++
			if argType == ir.TypeUnknown {
				fmt.Fprintf(out, "  %%%s = load { i32, i32, i64 }, ptr %%%s\n", loaded, slot)
				arg = loaded
			} else if llvmType(argType) != "void" && llvmType(argType) != "" {
				fmt.Fprintf(out, "  %%%s = load volatile %s, ptr %%%s\n", loaded, llvmType(argType), slot)
				arg = loaded
			}
		}
		if argType != ir.TypeUnknown {
			boxedVar := fmt.Sprintf("box.stru.%d", e.loadCounter)
			if err := e.emitBoxValue(out, arg, argType, boxedVar); err != nil {
				return err
			}
			arg = boxedVar
		}
		tagVar := fmt.Sprintf("tag.%d", e.loadCounter)
		padVar := fmt.Sprintf("pad.%d", e.loadCounter)
		valVar := fmt.Sprintf("val.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, arg)
		fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 1\n", padVar, arg)
		fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", valVar, arg)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_unknown(i32 %%%s, i32 %%%s, i64 %%%s, ptr %%__slot_ptr)\n", status, tagVar, padVar, valVar)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.fromObject":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.fromObject has invalid signature")
		}
		arg := instruction.Args[0]
		argType := e.types[arg]
		if slot, ok := e.varSlots[arg]; ok {
			loaded := fmt.Sprintf("%s.str_load.%d", arg, e.loadCounter)
			e.loadCounter++
			if argType == ir.TypeUnknown {
				fmt.Fprintf(out, "  %%%s = load { i32, i32, i64 }, ptr %%%s\n", loaded, slot)
				arg = loaded
			} else if llvmType(argType) != "void" && llvmType(argType) != "" {
				fmt.Fprintf(out, "  %%%s = load volatile %s, ptr %%%s\n", loaded, llvmType(argType), slot)
				arg = loaded
			}
		}
		if argType == ir.TypeUnknown {
			tagVar := fmt.Sprintf("tag.%d", e.loadCounter)
			padVar := fmt.Sprintf("pad.%d", e.loadCounter)
			valVar := fmt.Sprintf("val.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, arg)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 1\n", padVar, arg)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", valVar, arg)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_unknown(i32 %%%s, i32 %%%s, i64 %%%s, ptr %%__slot_ptr)\n", status, tagVar, padVar, valVar)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
			return nil
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_object(ptr %%%s, ptr %%__slot_ptr)\n", status, arg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.inspectObject":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.inspectObject has invalid signature")
		}
		arg := instruction.Args[0]
		argType := e.types[arg]
		if slot, ok := e.varSlots[arg]; ok {
			loaded := fmt.Sprintf("%s.inspect_load.%d", arg, e.loadCounter)
			e.loadCounter++
			if argType == ir.TypeUnknown {
				fmt.Fprintf(out, "  %%%s = load { i32, i32, i64 }, ptr %%%s\n", loaded, slot)
				arg = loaded
			} else {
				fmt.Fprintf(out, "  %%%s = load volatile %s, ptr %%%s\n", loaded, llvmType(argType), slot)
				arg = loaded
			}
		}
		if argType == ir.TypeUnknown {
			tagVar := fmt.Sprintf("inspect.tag.%d", e.loadCounter)
			valVar := fmt.Sprintf("inspect.val.%d", e.loadCounter)
			ptrVar := fmt.Sprintf("inspect.ptr.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, arg)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", valVar, arg)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, valVar)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_inspect_object(ptr %%%s, ptr %%__slot_ptr)\n", status, ptrVar)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_json_inspect_object(ptr %%%s, ptr %%__slot_ptr)\n", status, arg)
		}
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.inspectBuffer":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.inspectBuffer has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_console_inspect_buffer(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.inspectArray":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.inspectArray has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_console_inspect_array(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
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
	case "__string.includes":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeBool {
			return fmt.Errorf("string.includes has invalid signature")
		}
		pos := "0.0"
		if len(instruction.Args) == 3 {
			pos = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_includes(ptr %%%s, ptr %%%s, double %s, ptr %%__slot_double)\n", status, instruction.Args[0], instruction.Args[1], pos)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%__slot_double\n", instruction.Result)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
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
	case "__string.replace":
		if len(instruction.Args) != 3 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.replace has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_replace(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.replaceAll":
		if len(instruction.Args) != 3 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.replaceAll has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_replace_all(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2])
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
	case "__string.split":
		if len(instruction.Args) < 1 || len(instruction.Args) > 3 || instruction.Type != ir.TypeStringArray {
			return fmt.Errorf("string.split has invalid signature")
		}
		sepArg := "null"
		if len(instruction.Args) >= 2 {
			sepArg = "%" + instruction.Args[1]
		}
		limitArg := "-1.000000e+00"
		if len(instruction.Args) >= 3 {
			limitArg = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_split(ptr %%%s, ptr %s, double %s, ptr %%__slot_ptr)\n", status, instruction.Args[0], sepArg, limitArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.fromBigInt":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.fromBigInt has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_bigint(i64 %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.fromBigIntLocale":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.fromBigIntLocale has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_bigint_locale(i64 %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.errorToString":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("error.toString has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_error_to_string(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.match":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("string.match has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_match(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.search":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("string.search has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_search(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%__slot_double)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%__slot_double\n", instruction.Result)
	case "__string.replace_regex":
		if len(instruction.Args) != 4 {
			return fmt.Errorf("string.replace_regex has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_replace_regex(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
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
	case "__string.fromCodePoint":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.fromCodePoint has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_code_point(double %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
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
	case "__string.matchAll":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("string.matchAll has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_match(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.trimLeft":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_trim_start(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.trimRight":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_trim_end(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.toLocaleLowerCase":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_to_lower(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.toLocaleUpperCase":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_to_upper(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.fromCharCode":
		if len(instruction.Args) < 1 {
			return fmt.Errorf("string.fromCharCode requires 1 arg")
		}
		if len(instruction.Args) == 1 {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_code_point(double %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
		} else {
			codesArr := fmt.Sprintf("%s.codes", instruction.Result)
			fmt.Fprintf(out, "  %%%s = alloca [%d x double]\n", codesArr, len(instruction.Args))
			for i, arg := range instruction.Args {
				elemPtr := fmt.Sprintf("%s.elem.%d", instruction.Result, i)
				fmt.Fprintf(out, "  %%%s = getelementptr inbounds [%d x double], ptr %%%s, i64 0, i64 %d\n", elemPtr, len(instruction.Args), codesArr, i)
				fmt.Fprintf(out, "  store double %%%s, ptr %%%s\n", arg, elemPtr)
			}
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_char_codes(ptr %%%s, i64 %d, ptr %%__slot_ptr)\n", status, codesArr, len(instruction.Args))
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
		}
	case "__string.at":
		pos := "0.0"
		if len(instruction.Args) >= 2 {
			pos = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_at(ptr %%%s, double %s, ptr %%__slot_ptr)\n", status, instruction.Args[0], pos)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.substr":
		startArg := "0.0"
		if len(instruction.Args) >= 2 {
			startArg = "%" + instruction.Args[1]
		}
		lenArg := "1000000000.0"
		if len(instruction.Args) >= 3 {
			lenArg = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_substr(ptr %%%s, double %s, double %s, ptr %%__slot_ptr)\n", status, instruction.Args[0], startArg, lenArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.anchor":
		nameArg := "null"
		if len(instruction.Args) >= 2 {
			nameArg = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_anchor(ptr %%%s, ptr %s, ptr %%__slot_ptr)\n", status, instruction.Args[0], nameArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.big":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_big(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.blink":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_blink(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.bold":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_bold(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.fixed":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_fixed(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.fontcolor":
		cArg := "null"
		if len(instruction.Args) >= 2 {
			cArg = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_fontcolor(ptr %%%s, ptr %s, ptr %%__slot_ptr)\n", status, instruction.Args[0], cArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.fontsize":
		sArg := "null"
		if len(instruction.Args) >= 2 {
			argTyp := e.types[instruction.Args[1]]
			if argTyp == ir.TypeNumber {
				numStrVal := fmt.Sprintf("%s.numstr", instruction.Result)
				fmt.Fprintf(out, "  %%%s.status = call i32 @scriptgo_string_from_number(double %%%s, ptr %%__slot_ptr)\n", numStrVal, instruction.Args[1])
				fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s.status)\n", numStrVal)
				fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", numStrVal)
				sArg = "%" + numStrVal
			} else {
				sArg = "%" + instruction.Args[1]
			}
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_fontsize(ptr %%%s, ptr %s, ptr %%__slot_ptr)\n", status, instruction.Args[0], sArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.italics":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_italics(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.link":
		uArg := "null"
		if len(instruction.Args) >= 2 {
			uArg = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_link(ptr %%%s, ptr %s, ptr %%__slot_ptr)\n", status, instruction.Args[0], uArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.small":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_small(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.strike":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_strike(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.sub":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_sub(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.sup":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_sup(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.encodeURIComponent":
		arg0 := e.resolveArg(out, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_encode_uri_component(ptr %%%s, ptr %%__slot_ptr)\n", status, arg0)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.decodeURIComponent":
		arg0 := e.resolveArg(out, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_decode_uri_component(ptr %%%s, ptr %%__slot_ptr)\n", status, arg0)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.encodeURI":
		arg0 := e.resolveArg(out, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_encode_uri(ptr %%%s, ptr %%__slot_ptr)\n", status, arg0)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.decodeURI":
		arg0 := e.resolveArg(out, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_decode_uri(ptr %%%s, ptr %%__slot_ptr)\n", status, arg0)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.raw":
		if len(instruction.Args) < 1 {
			fmt.Fprintf(out, "  %%%s = inttoptr i64 0 to ptr\n", instruction.Result)
			return nil
		}
		argTyp := e.types[instruction.Args[0]]
		if argTyp == ir.TypeStringArray || strings.HasSuffix(string(argTyp), "[]") {
			statusRaw := fmt.Sprintf("%s.raw.status", instruction.Result)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_get(ptr %%%s, double 0.0, ptr %%__slot_ptr)\n", statusRaw, instruction.Args[0])
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", statusRaw)
			currentStr := fmt.Sprintf("%s.str.0", instruction.Result)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", currentStr)

			for i := 1; i < len(instruction.Args); i++ {
				subArg := instruction.Args[i]
				subTyp := e.types[subArg]
				subStrVar := fmt.Sprintf("%s.sub.%d", instruction.Result, i)
				if subTyp == ir.TypeNumber {
					subStatus := fmt.Sprintf("%s.sub_status.%d", instruction.Result, i)
					fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_number(double %%%s, ptr %%__slot_ptr)\n", subStatus, subArg)
					fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", subStatus)
					fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", subStrVar)
				} else if subTyp == ir.TypeBigInt {
					subStatus := fmt.Sprintf("%s.sub_status.%d", instruction.Result, i)
					fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_bigint(i64 %%%s, ptr %%__slot_ptr)\n", subStatus, subArg)
					fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", subStatus)
					fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", subStrVar)
				} else {
					subStrVar = subArg
				}

				concat1 := fmt.Sprintf("%s.concat1.%d", instruction.Result, i)
				concat1Status := fmt.Sprintf("%s.concat1_status.%d", instruction.Result, i)
				concat1Slot := fmt.Sprintf("%s.concat1_slot.%d", instruction.Result, i)
				fmt.Fprintf(out, "  %%%s = alloca ptr\n", concat1Slot)
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_concat(ptr %%%s, ptr %%%s, ptr %%%s)\n", concat1Status, currentStr, subStrVar, concat1Slot)
				fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", concat1Status)
				fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", concat1, concat1Slot)
				currentStr = concat1

				litGetStatus := fmt.Sprintf("%s.lit_get_status.%d", instruction.Result, i)
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_array_get(ptr %%%s, double %f, ptr %%__slot_ptr)\n", litGetStatus, instruction.Args[0], float64(i))
				fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", litGetStatus)
				litStr := fmt.Sprintf("%s.lit_str.%d", instruction.Result, i)
				fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", litStr)

				concat2 := fmt.Sprintf("%s.concat2.%d", instruction.Result, i)
				concat2Status := fmt.Sprintf("%s.concat2_status.%d", instruction.Result, i)
				concat2Slot := fmt.Sprintf("%s.concat2_slot.%d", instruction.Result, i)
				fmt.Fprintf(out, "  %%%s = alloca ptr\n", concat2Slot)
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_concat(ptr %%%s, ptr %%%s, ptr %%%s)\n", concat2Status, currentStr, litStr, concat2Slot)
				fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", concat2Status)
				fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", concat2, concat2Slot)
				currentStr = concat2
			}
			fmt.Fprintf(out, "  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, currentStr)
		} else {
			fmt.Fprintf(out, "  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, instruction.Args[0])
		}
		return nil
	case "__string.localeCompare":
		other := "\"\""
		if len(instruction.Args) >= 2 {
			other = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_index_of(ptr %%%s, ptr %s, double 0.0, ptr %%__slot_double)\n", status, instruction.Args[0], other)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%__slot_double\n", instruction.Result)
	case "__string.new":
		if len(instruction.Args) == 0 {
			if strGlobal, ok := e.stringsByValue[""]; ok {
				fmt.Fprintf(out, "  %%%s = getelementptr inbounds [1 x i8], ptr %s, i64 0, i64 0\n", instruction.Result, strGlobal)
				return nil
			}
		}
		arg := instruction.Args[0]
		argType := e.types[arg]
		if slot, ok := e.varSlots[arg]; ok {
			loaded := fmt.Sprintf("%s.str_load.%d", arg, e.loadCounter)
			e.loadCounter++
			if argType == ir.TypeUnknown {
				fmt.Fprintf(out, "  %%%s = load { i32, i32, i64 }, ptr %%%s\n", loaded, slot)
				arg = loaded
			} else if llvmType(argType) != "void" && llvmType(argType) != "" {
				fmt.Fprintf(out, "  %%%s = load volatile %s, ptr %%%s\n", loaded, llvmType(argType), slot)
				arg = loaded
			}
		}
		if argType == ir.TypeString {
			fmt.Fprintf(out, "  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, arg)
			return nil
		}
		if argType == ir.TypeNumber {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_to_string(double %%%s, double 10.0, ptr %%__slot_ptr)\n", status, arg)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
			return nil
		}
		if argType == ir.TypeBool {
			trueGlobal := e.stringsByValue["true"]
			falseGlobal := e.stringsByValue["false"]
			truePtr := fmt.Sprintf("str.true.%d", e.loadCounter)
			e.loadCounter++
			falsePtr := fmt.Sprintf("str.false.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = getelementptr inbounds [5 x i8], ptr %s, i64 0, i64 0\n", truePtr, trueGlobal)
			fmt.Fprintf(out, "  %%%s = getelementptr inbounds [6 x i8], ptr %s, i64 0, i64 0\n", falsePtr, falseGlobal)
			fmt.Fprintf(out, "  %%%s = select i1 %%%s, ptr %%%s, ptr %%%s\n", instruction.Result, arg, truePtr, falsePtr)
			return nil
		}
		if argType == ir.TypeBigInt {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_bigint_to_string(i64 %%%s, double 10.0, ptr %%__slot_ptr)\n", status, arg)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
			return nil
		}
		if argType == ir.TypeVoid {
			undefGlobal := e.stringsByValue["undefined"]
			fmt.Fprintf(out, "  %%%s = getelementptr inbounds [10 x i8], ptr %s, i64 0, i64 0\n", instruction.Result, undefGlobal)
			return nil
		}
		if argType == ir.TypeUnknown {
			tagVar := fmt.Sprintf("tag.%d", e.loadCounter)
			padVar := fmt.Sprintf("pad.%d", e.loadCounter)
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			statusVar := fmt.Sprintf("status.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, arg)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 1\n", padVar, arg)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg)
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_unknown(i32 %%%s, i32 %%%s, i64 %%%s, ptr %%__slot_ptr)\n", statusVar, tagVar, padVar, payloadVar)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", statusVar)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
			return nil
		}
		fmt.Fprintf(out, "  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, arg)
		return nil
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
