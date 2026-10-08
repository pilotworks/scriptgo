package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitStringAnnexBIntrinsic(out *strings.Builder, instruction ir.Instruction, status string) error {
	switch instruction.Callee {
	case "__string.trimLeft":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_trim_start(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%__slot_ptr\n", instruction.Result)
	case "__string.trimRight":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_string_trim_end(ptr %%%s, ptr %%__slot_ptr)\n", status, instruction.Args[0])
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
	}
	return nil
}
