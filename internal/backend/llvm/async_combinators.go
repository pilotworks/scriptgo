package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

var promiseCombinatorKinds = map[string]int{
	"__async.promise_all":         0,
	"__async.promise_all_settled": 1,
	"__async.promise_any":         2,
	"__async.promise_race":        3,
}

// emitPromiseCombinator calls scriptgo_promise_combine. The input array's
// element layout comes from its IR type (-1 reads tagged unknown elements).
// The result is either the tuple object in Args[1] or a runtime-allocated
// array laid out as the IR array type named by Value.
func (e *functionEmitter) emitPromiseCombinator(out *strings.Builder, instruction ir.Instruction) error {
	kind := promiseCombinatorKinds[instruction.Callee]
	if len(instruction.Args) < 1 || len(instruction.Args) > 2 {
		return fmt.Errorf("%s requires an input array and an optional result object", instruction.Callee)
	}
	input := instruction.Args[0]
	inputType := e.types[input]
	inputTag := -1
	if arrayElementType(inputType) != ir.TypeUnknown {
		inputTag = arrayElementTag(inputType)
	}
	container := "null"
	if len(instruction.Args) == 2 {
		container = "%" + e.resolveArg(out, instruction.Args[1])
	}
	outputSize, outputTag := int64(0), 0
	if instruction.Value != "" {
		settled := ir.Type(instruction.Value)
		size, err := arrayElementSizeForTarget(settled, e.pointerSize())
		if err != nil {
			return err
		}
		outputSize = size
		if arrayElementType(settled) != ir.TypeUnknown {
			outputTag = arrayElementTag(settled)
		}
	} else if (kind == 0 || kind == 1) && len(instruction.Args) == 1 {
		return fmt.Errorf("%s needs a result array type or object", instruction.Callee)
	}
	slot := instruction.Result + ".slot"
	fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_promise_combine(i32 %d, ptr %%%s, i64 %d, ptr %s, i64 %d, i64 %d, ptr %%%s)\n", status, kind, e.resolveArg(out, input), inputTag, container, outputSize, outputTag, slot)
	fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
	fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
	e.types[instruction.Result] = instruction.Type
	return nil
}
