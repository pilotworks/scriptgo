package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitPromiseSettlement(out *strings.Builder, instruction ir.Instruction, promise string, status string, rejected bool) error {
	if len(instruction.Args) == 0 {
		if rejected {
			return fmt.Errorf("promise.reject requires 1 argument")
		}
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_boxed(ptr %%%s, i32 0, i64 0)\n", status, promise))
		return nil
	}
	if len(instruction.Args) != 1 {
		return fmt.Errorf("promise settlement requires 1 argument")
	}

	arg := instruction.Args[0]
	argType := e.types[arg]
	argVal := e.resolveArg(out, arg)
	if !rejected && argType == ir.TypeNumber {
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_number(ptr %%%s, double %%%s)\n", status, promise, argVal))
		return nil
	}
	if !rejected && argType == ir.TypeBool {
		boolVal := fmt.Sprintf("promise.bool.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i32\n", boolVal, argVal))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_bool(ptr %%%s, i32 %%%s)\n", status, promise, boolVal))
		return nil
	}
	if !rejected && argType == ir.TypeBigInt {
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_bigint(ptr %%%s, i64 %%%s)\n", status, promise, argVal))
		return nil
	}

	boxed := fmt.Sprintf("promise.box.%d", e.loadCounter)
	e.loadCounter++
	if err := e.emitBoxValue(out, argVal, argType, boxed); err != nil {
		return err
	}
	tag := fmt.Sprintf("promise.tag.%d", e.loadCounter)
	payload := fmt.Sprintf("promise.payload.%d", e.loadCounter)
	e.loadCounter++
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 0\n", tag, boxed))
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payload, boxed))
	settleFn := "scriptgo_promise_resolve_boxed"
	if rejected {
		settleFn = "scriptgo_promise_reject_boxed"
	}
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s, i32 %%%s, i64 %%%s)\n", status, settleFn, promise, tag, payload))
	return nil
}

func promiseReactionTag(typeName string) int {
	switch typeName {
	case string(ir.TypeNumber):
		return 3
	case string(ir.TypeBool):
		return 2
	case string(ir.TypeString):
		return 4
	case string(ir.TypeVoid), "":
		return 0
	default:
		if strings.HasSuffix(typeName, "[]") {
			return 6 // array payload
		}
		return 5
	}
}
