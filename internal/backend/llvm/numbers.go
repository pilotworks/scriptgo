package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func emitNumberIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	status := instruction.Result + ".status"
	slot := instruction.Result + ".slot"
	switch instruction.Callee {
	case "__number.parseInt":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("parseInt has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_parse_int(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case "__number.parseFloat":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("parseFloat has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_parse_float(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
	case "__number.isNaN":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("isNaN has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_is_nan(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
	case "__number.isFinite":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("isFinite has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_is_finite(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
	case "__number.isInteger":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("isInteger has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_is_integer(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
	case "__number.toFixed":
		if len(instruction.Args) < 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("toFixed has invalid signature")
		}
		digits := "0.0"
		if len(instruction.Args) >= 2 {
			digits = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_to_fixed(double %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], digits, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__number.toString":
		if len(instruction.Args) < 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("toString has invalid signature")
		}
		radix := "10.0"
		if len(instruction.Args) >= 2 {
			radix = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_to_string(double %%%s, double %s, ptr %%%s)\n", status, instruction.Args[0], radix, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	default:
		return fmt.Errorf("unknown number intrinsic %q", instruction.Callee)
	}
	return nil
}
