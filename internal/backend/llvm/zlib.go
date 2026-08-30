package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitZlibIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	if len(instruction.Args) != 2 || instruction.Result == "" || instruction.Type != ir.TypeUint8Array {
		return fmt.Errorf("zlib intrinsic %q has invalid signature", instruction.Callee)
	}

	input := e.resolveArg(out, instruction.Args[0])
	mode := e.resolveArg(out, instruction.Args[1])
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
	switch instruction.Callee {
	case "__zlib.transform_string":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_zlib_transform_string(ptr %%%s, double %%%s, ptr %%%s)\n", status, input, mode, slot)
	case "__zlib.transform_buffer":
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_zlib_transform_buffer(ptr %%%s, double %%%s, ptr %%%s)\n", status, input, mode, slot)
	default:
		return fmt.Errorf("unknown zlib intrinsic %q", instruction.Callee)
	}
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	return nil
}
