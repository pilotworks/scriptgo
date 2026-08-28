package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitNumberIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	status := instruction.Result + ".status"
	slot := instruction.Result + ".slot"
	switch instruction.Callee {
	case "__number.parseInt":
		if len(instruction.Args) < 1 || len(instruction.Args) > 2 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("parseInt has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		if len(instruction.Args) == 2 {
			radixArg := e.resolveArg(out, instruction.Args[1])
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_parse_int_radix(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], radixArg, slot)
		} else {
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_parse_int(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		}
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
	case "__number.isSafeInteger":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("isSafeInteger has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_is_safe_integer(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
	case "__number.toFixed":
		if len(instruction.Args) < 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("toFixed has invalid signature")
		}
		numArg := instruction.Args[0]
		if e.types[numArg] == ir.TypeUnknown {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			numVar := fmt.Sprintf("num.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, numArg)
			fmt.Fprintf(out, "  %%%s = bitcast i64 %%%s to double\n", numVar, payloadVar)
			numArg = numVar
		}
		digits := "0.0"
		if len(instruction.Args) >= 2 {
			digits = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_to_fixed(double %%%s, double %s, ptr %%%s)\n", status, numArg, digits, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__number.toString":
		if len(instruction.Args) < 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("toString has invalid signature")
		}
		numArg := instruction.Args[0]
		if e.types[numArg] == ir.TypeUnknown {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			numVar := fmt.Sprintf("num.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, numArg)
			fmt.Fprintf(out, "  %%%s = bitcast i64 %%%s to double\n", numVar, payloadVar)
			numArg = numVar
		}
		radix := "10.0"
		if len(instruction.Args) >= 2 {
			radix = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_to_string(double %%%s, double %s, ptr %%%s)\n", status, numArg, radix, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__number.toExponential":
		if len(instruction.Args) < 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("toExponential has invalid signature")
		}
		numArg := instruction.Args[0]
		if e.types[numArg] == ir.TypeUnknown {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			numVar := fmt.Sprintf("num.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, numArg)
			fmt.Fprintf(out, "  %%%s = bitcast i64 %%%s to double\n", numVar, payloadVar)
			numArg = numVar
		}
		fractionDigits := "0.0 / 0.0"
		if len(instruction.Args) >= 2 {
			fractionDigits = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_to_exponential(double %%%s, double %s, ptr %%%s)\n", status, numArg, fractionDigits, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__number.toPrecision":
		if len(instruction.Args) < 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("toPrecision has invalid signature")
		}
		numArg := instruction.Args[0]
		if e.types[numArg] == ir.TypeUnknown {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			numVar := fmt.Sprintf("num.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, numArg)
			fmt.Fprintf(out, "  %%%s = bitcast i64 %%%s to double\n", numVar, payloadVar)
			numArg = numVar
		}
		precision := "0.0 / 0.0"
		if len(instruction.Args) >= 2 {
			precision = "%" + instruction.Args[1]
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_to_precision(double %%%s, double %s, ptr %%%s)\n", status, numArg, precision, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__number.toLocaleString":
		if len(instruction.Args) < 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("toLocaleString has invalid signature")
		}
		numArg := instruction.Args[0]
		if e.types[numArg] == ir.TypeUnknown {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			numVar := fmt.Sprintf("num.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, numArg)
			fmt.Fprintf(out, "  %%%s = bitcast i64 %%%s to double\n", numVar, payloadVar)
			numArg = numVar
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_to_locale_string(double %%%s, ptr %%%s)\n", status, numArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	case "__number.new":
		if len(instruction.Args) == 0 {
			fmt.Fprintf(out, "  %%%s = fadd double 0.0, 0.0\n", instruction.Result)
		} else {
			argType := e.types[instruction.Args[0]]
			if argType == ir.TypeString {
				fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
				fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_number_parse_float(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
				fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
				fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
			} else if argType == ir.TypeBool {
				boolF64 := instruction.Result + ".f64"
				fmt.Fprintf(out, "  %%%s = uitofp i1 %%%s to double\n", boolF64, instruction.Args[0])
				fmt.Fprintf(out, "  %%%s = fadd double %%%s, 0.0\n", instruction.Result, boolF64)
			} else {
				fmt.Fprintf(out, "  %%%s = fadd double %%%s, 0.0\n", instruction.Result, instruction.Args[0])
			}
		}
	default:
		return fmt.Errorf("unknown number intrinsic %q", instruction.Callee)
	}
	return nil
}
