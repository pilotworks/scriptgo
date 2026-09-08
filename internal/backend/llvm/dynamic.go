package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

const boxedLLVMType = "{ i32, i32, i64, i64 }"

func moduleHasDynamic(module ir.Module) bool {
	for _, function := range module.Functions {
		if hasDynamicInstruction(function.Body) {
			return true
		}
	}
	return false
}

func hasDynamicInstruction(list []ir.Instruction) bool {
	for _, instruction := range list {
		if instruction.Op == ir.OpDynamicCall || hasDynamicInstruction(instruction.Then) || hasDynamicInstruction(instruction.Else) || hasDynamicInstruction(instruction.Body) || hasDynamicInstruction(instruction.Catch) || hasDynamicInstruction(instruction.Finally) {
			return true
		}
	}
	return false
}

func (e *functionEmitter) emitDynamicCall(out *strings.Builder, instruction ir.Instruction, targetResult string, originalArgs []string) error {
	if len(originalArgs) != len(instruction.Args) {
		return fmt.Errorf("dynamic call argument resolution mismatch")
	}
	path, export, ok := strings.Cut(instruction.Callee, "#")
	if !ok || path == "" || export == "" {
		return fmt.Errorf("invalid dynamic callee %q", instruction.Callee)
	}
	sourceGlobal, ok := e.stringsByValue[instruction.Value]
	if !ok {
		return fmt.Errorf("dynamic module source was not interned for %q", path)
	}
	pathGlobal, ok := e.stringsByValue[path]
	if !ok {
		return fmt.Errorf("dynamic module path was not interned for %q", path)
	}
	exportGlobal, ok := e.stringsByValue[export]
	if !ok {
		return fmt.Errorf("dynamic export was not interned for %q", export)
	}
	pathPtr := e.dynamicStringPointer(out, pathGlobal, "dynamic.path")
	sourcePtr := e.dynamicStringPointer(out, sourceGlobal, "dynamic.source")
	exportPtr := e.dynamicStringPointer(out, exportGlobal, "dynamic.export")

	count := len(originalArgs)
	arrayCount := count
	if arrayCount == 0 {
		arrayCount = 1
	}
	argsSlot := fmt.Sprintf("dynamic.args.%d", e.loadCounter)
	e.loadCounter++
	fmt.Fprintf(out, "  %%%s = alloca [%d x %s]\n", argsSlot, arrayCount, boxedLLVMType)
	for i, arg := range instruction.Args {
		argType := e.types[originalArgs[i]]
		if argType == "" {
			argType = e.types[arg]
		}
		boxed := fmt.Sprintf("dynamic.arg.boxed.%d", e.loadCounter)
		e.loadCounter++
		if err := e.emitBoxValue(out, arg, argType, boxed); err != nil {
			return err
		}
		index := fmt.Sprintf("dynamic.arg.index.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = getelementptr inbounds [%d x %s], ptr %%%s, i64 0, i64 %d\n", index, arrayCount, boxedLLVMType, argsSlot, i)
		fmt.Fprintf(out, "  store %s %%%s, ptr %%%s\n", boxedLLVMType, boxed, index)
	}
	outSlot := fmt.Sprintf("dynamic.out.%d", e.loadCounter)
	e.loadCounter++
	fmt.Fprintf(out, "  %%%s = alloca %s\n", outSlot, boxedLLVMType)
	status := fmt.Sprintf("dynamic.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dynamic_call_module(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, i32 %d, i32 %d, i32 %d, ptr %%%s)\n", status, pathPtr, sourcePtr, exportPtr, argsSlot, count, instruction.FieldIndex, dynamicResultTag(instruction.Type), outSlot)
	fmt.Fprintf(out, "  call void @scriptgo_dynamic_abort_if_failed(i32 %%%s)\n", status)
	boxedResult := fmt.Sprintf("dynamic.result.boxed.%d", e.loadCounter)
	e.loadCounter++
	fmt.Fprintf(out, "  %%%s = load %s, ptr %%%s\n", boxedResult, boxedLLVMType, outSlot)
	if err := e.emitDynamicResult(out, instruction.Type, boxedResult, instruction.Result); err != nil {
		return err
	}
	if targetResult != "" && targetResult != instruction.Result {
		if slot, ok := e.varSlots[targetResult]; ok {
			fmt.Fprintf(out, "  store %s %%%s, ptr %%%s\n", llvmType(instruction.Type), instruction.Result, slot)
		} else {
			for _, global := range e.module.Globals {
				if global.Name == targetResult {
					fmt.Fprintf(out, "  store %s %%%s, ptr @%s\n", llvmType(instruction.Type), instruction.Result, targetResult)
					break
				}
			}
		}
	}
	return nil
}

func dynamicResultTag(typ ir.Type) int {
	switch typ {
	case ir.TypeBool:
		return 2
	case ir.TypeNumber:
		return 3
	case ir.TypeString:
		return 4
	case ir.TypePointer:
		return 1
	case ir.TypeObject:
		return 5
	default:
		if strings.HasSuffix(string(typ), "[]") {
			return 6
		}
		if strings.HasPrefix(string(typ), "object:") {
			return 5
		}
		return -1
	}
}

func (e *functionEmitter) dynamicStringPointer(out *strings.Builder, global, prefix string) string {
	name := fmt.Sprintf("%s.%d", prefix, e.loadCounter)
	e.loadCounter++
	fmt.Fprintf(out, "  %%%s = getelementptr inbounds i8, ptr %s, i64 0\n", name, global)
	return name
}

func (e *functionEmitter) emitDynamicResult(out *strings.Builder, typ ir.Type, boxed, result string) error {
	tag := fmt.Sprintf("dynamic.tag.%d", e.loadCounter)
	e.loadCounter++
	fmt.Fprintf(out, "  %%%s = extractvalue %s %%%s, 0\n", tag, boxedLLVMType, boxed)
	switch typ {
	case ir.TypeUnknown:
		fmt.Fprintf(out, "  %%%s = load %s, ptr %%%s\n", result, boxedLLVMType, e.dynamicResultSlot(out, boxed))
	case ir.TypeNumber:
		payload := fmt.Sprintf("dynamic.number.payload.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = extractvalue %s %%%s, 2\n", payload, boxedLLVMType, boxed)
		fmt.Fprintf(out, "  %%%s = bitcast i64 %%%s to double\n", result, payload)
	case ir.TypeBool:
		payload := fmt.Sprintf("dynamic.bool.payload.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = extractvalue %s %%%s, 2\n", payload, boxedLLVMType, boxed)
		fmt.Fprintf(out, "  %%%s = trunc i64 %%%s to i1\n", result, payload)
	case ir.TypeString, ir.TypePointer:
		payload := fmt.Sprintf("dynamic.ptr.payload.%d", e.loadCounter)
		e.loadCounter++
		fmt.Fprintf(out, "  %%%s = extractvalue %s %%%s, 2\n", payload, boxedLLVMType, boxed)
		fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", result, payload)
	default:
		if strings.HasSuffix(string(typ), "[]") || strings.HasPrefix(string(typ), "object:") {
			payload := fmt.Sprintf("dynamic.ptr.payload.%d", e.loadCounter)
			e.loadCounter++
			fmt.Fprintf(out, "  %%%s = extractvalue %s %%%s, 2\n", payload, boxedLLVMType, boxed)
			fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to ptr\n", result, payload)
			break
		}
		return fmt.Errorf("dynamic result type %q is outside the primitive boundary", typ)
	}
	return nil
}

func (e *functionEmitter) dynamicResultSlot(out *strings.Builder, boxed string) string {
	name := fmt.Sprintf("dynamic.result.slot.%d", e.loadCounter)
	e.loadCounter++
	fmt.Fprintf(out, "  %%%s = alloca %s\n", name, boxedLLVMType)
	fmt.Fprintf(out, "  store %s %%%s, ptr %%%s\n", boxedLLVMType, boxed, name)
	return name
}
