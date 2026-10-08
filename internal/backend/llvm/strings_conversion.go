package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitStringConversionIntrinsic(out *strings.Builder, instruction ir.Instruction, status string) error {
	switch instruction.Callee {
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
				fmt.Fprintf(out, "  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded, slot)
				arg = loaded
			} else if llvmType(argType) != "void" && llvmType(argType) != "" {
				fmt.Fprintf(out, "  %%%s = load%s %s, ptr %%%s\n", loaded, e.vol(), llvmType(argType), slot)
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
		valuePtr, err := e.emitCanonicalValuePointer(out, arg, ir.TypeUnknown, "string.unknown")
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_unknown(ptr %s, ptr %%__slot_ptr)\n", status, valuePtr)
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
				fmt.Fprintf(out, "  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded, slot)
				arg = loaded
			} else if llvmType(argType) != "void" && llvmType(argType) != "" {
				fmt.Fprintf(out, "  %%%s = load%s %s, ptr %%%s\n", loaded, e.vol(), llvmType(argType), slot)
				arg = loaded
			}
		}
		if argType == ir.TypeUnknown {
			valuePtr, err := e.emitCanonicalValuePointer(out, arg, ir.TypeUnknown, "object.string.unknown")
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_unknown(ptr %s, ptr %%__slot_ptr)\n", status, valuePtr)
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
				fmt.Fprintf(out, "  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded, slot)
				arg = loaded
			} else {
				fmt.Fprintf(out, "  %%%s = load%s %s, ptr %%%s\n", loaded, e.vol(), llvmType(argType), slot)
				arg = loaded
			}
		}
		if argType == ir.TypeUnknown {
			tagVar := fmt.Sprintf("inspect.tag.%d", e.loadCounter)
			valVar := fmt.Sprintf("inspect.val.%d", e.loadCounter)
			ptrVar := fmt.Sprintf("inspect.ptr.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 0\n", tagVar, arg)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", valVar, arg)
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
	case "__string.fromCodePoint":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.fromCodePoint has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_code_point(double %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
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
				fmt.Fprintf(out, "  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded, slot)
				arg = loaded
			} else if llvmType(argType) != "void" && llvmType(argType) != "" {
				fmt.Fprintf(out, "  %%%s = load%s %s, ptr %%%s\n", loaded, e.vol(), llvmType(argType), slot)
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
			valuePtr, err := e.emitCanonicalValuePointer(out, arg, ir.TypeUnknown, "stringify.unknown")
			if err != nil {
				return err
			}
			statusVar := fmt.Sprintf("status.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_from_unknown(ptr %s, ptr %%__slot_ptr)\n", statusVar, valuePtr)
			fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", statusVar)
			fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
			return nil
		}
		fmt.Fprintf(out, "  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, arg)
		return nil
	}
	return nil
}
