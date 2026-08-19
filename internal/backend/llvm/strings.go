package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func emitStringIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__string.length":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("string.length has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call double @scriptgo_string_length(ptr %%%s)\n", instruction.Result, instruction.Args[0])
	case "__string.lastIndexOf":
		if (len(instruction.Args) != 2 && len(instruction.Args) != 3) || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("string.lastIndexOf has invalid signature")
		}
		position := "-1.0"
		if len(instruction.Args) == 3 {
			position = "%" + instruction.Args[2]
		}
		fmt.Fprintf(out, "  %%%s = call double @scriptgo_string_last_index(ptr %%%s, ptr %%%s, double %s)\n", instruction.Result, instruction.Args[0], instruction.Args[1], position)
	case "__string.slice":
		if len(instruction.Args) != 3 || instruction.Type != ir.TypeString {
			return fmt.Errorf("string.slice has invalid signature")
		}
		fmt.Fprintf(out, "  %%%s = call ptr @scriptgo_string_slice(ptr %%%s, double %%%s, double %%%s)\n", instruction.Result, instruction.Args[0], instruction.Args[1], instruction.Args[2])
	default:
		return fmt.Errorf("unknown string intrinsic %q", instruction.Callee)
	}
	return nil
}

func emitArrayIntrinsic(out *strings.Builder, instruction ir.Instruction, arrayType ir.Type) error {
	if instruction.Callee != "__array.length" || len(instruction.Args) != 1 || instruction.Type != ir.TypeNumber {
		return fmt.Errorf("array.length has invalid signature")
	}
	length := "scriptgo_array_number_length"
	if arrayType == ir.TypeStringArray {
		length = "scriptgo_array_string_length"
	}
	resultSlot := instruction.Result + ".slot"
	fmt.Fprintf(out, "  %%%s = alloca i64\n", resultSlot)
	fmt.Fprintf(out, "  call i32 @%s(ptr %%%s, ptr %%%s)\n", length, instruction.Args[0], resultSlot)
	fmt.Fprintf(out, "  %%%s.i64 = load i64, ptr %%%s\n", instruction.Result, resultSlot)
	fmt.Fprintf(out, "  %%%s = uitofp i64 %%%s.i64 to double\n", instruction.Result, instruction.Result)
	return nil
}
