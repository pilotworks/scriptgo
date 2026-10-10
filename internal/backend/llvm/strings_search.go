package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitStringSearchIntrinsic(out *strings.Builder, instruction ir.Instruction, status string) error {
	switch instruction.Callee {
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
		position := "0x7FF0000000000000" // an omitted position searches from the end
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
	case "__string.replaceAll_regex", "__string.replace_regex_fn", "__string.replaceAll_regex_fn":
		if len(instruction.Args) != 4 {
			return fmt.Errorf("%s has invalid signature", instruction.Callee)
		}
		fn := map[string]string{
			"__string.replaceAll_regex":    "scriptgo_string_replace_all_regex",
			"__string.replace_regex_fn":    "scriptgo_string_replace_regex_fn",
			"__string.replaceAll_regex_fn": "scriptgo_string_replace_all_regex_fn",
		}[instruction.Callee]
		replacement := e.resolveArg(out, instruction.Args[3])
		if strings.HasSuffix(instruction.Callee, "_fn") {
			replacement = e.ensurePointerArg(out, instruction.Args[3])
		}
		fmt.Fprintf(out, "  %%%s = call i32 @%s(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%__slot_ptr)\n", status, fn, instruction.Args[0], instruction.Args[1], instruction.Args[2], replacement)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.split_regex":
		if len(instruction.Args) < 3 || len(instruction.Args) > 4 || instruction.Type != ir.TypeStringArray {
			return fmt.Errorf("string.split_regex has invalid signature")
		}
		limitArg := "-1.000000e+00" // omitted: no limit
		if len(instruction.Args) == 4 {
			limitArg = "%" + instruction.Args[3]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_split_regex(ptr %%%s, ptr %%%s, ptr %%%s, double %s, ptr %%__slot_ptr)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], limitArg)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.replace_regex":
		if len(instruction.Args) != 4 {
			return fmt.Errorf("string.replace_regex has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_replace_regex(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.matchAll":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("string.matchAll has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_match_all(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.localeCompare":
		other := "@scriptgo_undefined_sentinel" // compares with "undefined"
		if len(instruction.Args) >= 2 {
			other = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_locale_compare(ptr %%%s, ptr %s, ptr %%__slot_double)\n", status, instruction.Args[0], other)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%__slot_double\n", instruction.Result)
	}
	return nil
}
