package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func emitRegexIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	status := instruction.Result + ".status"
	slot := instruction.Result + ".slot"

	switch instruction.Callee {
	case "__regex.test":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("regex.test has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_regex_test(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		dblVal := instruction.Result + ".dbl"
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", dblVal, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s, 0.0\n", instruction.Result, dblVal)

	case "__regex.exec":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("regex.exec has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_regex_exec(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)

	case "__regex.global", "__regexp.global", "__regex.ignoreCase", "__regexp.ignoreCase", "__regex.multiline", "__regexp.multiline":
		fmt.Fprintf(out, "  %%%s = icmp eq i32 1, 1\n", instruction.Result)

	case "__regex.dotAll", "__regexp.dotAll", "__regex.unicode", "__regexp.unicode", "__regex.sticky", "__regexp.sticky", "__regex.hasIndices", "__regexp.hasIndices", "__regex.unicodeSets", "__regexp.unicodeSets":
		fmt.Fprintf(out, "  %%%s = icmp eq i32 1, 0\n", instruction.Result)

	case "__regex.escape", "__regexp.escape":
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_regex_escape(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)

	case "__regex.compile", "__regexp.compile":
		if len(instruction.Args) > 0 {
			fmt.Fprintf(out, "  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, instruction.Args[0])
		} else {
			fmt.Fprintf(out, "  %%%s = bitcast ptr null to ptr\n", instruction.Result)
		}

	default:
		return fmt.Errorf("unknown regex intrinsic %q", instruction.Callee)
	}
	return nil
}

func emitBigIntIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	status := instruction.Result + ".status"
	slot := instruction.Result + ".slot"

	switch instruction.Callee {
	case "__bigint.fromNumber":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("bigint.fromNumber has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca i64\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_bigint_from_number(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load i64, ptr %%%s\n", instruction.Result, slot)

	case "__bigint.fromString":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("bigint.fromString has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = alloca i64\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_bigint_from_string(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load i64, ptr %%%s\n", instruction.Result, slot)

	case "__bigint.asIntN":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("bigint.asIntN has invalid signature")
		}
		bitsI64 := instruction.Result + ".bits_i64"
		fmt.Fprintf(out, "  %%%s = fptosi double %%%s to i64\n", bitsI64, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = alloca i64\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_bigint_as_int_n(i64 %%%s, i64 %%%s, ptr %%%s)\n", status, bitsI64, instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load i64, ptr %%%s\n", instruction.Result, slot)

	case "__bigint.asUintN":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("bigint.asUintN has invalid signature")
		}
		bitsI64 := instruction.Result + ".bits_i64"
		fmt.Fprintf(out, "  %%%s = fptosi double %%%s to i64\n", bitsI64, instruction.Args[0])
		fmt.Fprintf(out, "  %%%s = alloca i64\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_bigint_as_uint_n(i64 %%%s, i64 %%%s, ptr %%%s)\n", status, bitsI64, instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load i64, ptr %%%s\n", instruction.Result, slot)

	default:
		return fmt.Errorf("unknown bigint intrinsic %q", instruction.Callee)
	}
	return nil
}
