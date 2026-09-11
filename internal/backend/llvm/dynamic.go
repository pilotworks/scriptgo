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
		if instruction.Op == ir.OpDynamicCall || instruction.Op == ir.OpDynamicFunctionCall || hasDynamicInstruction(instruction.Then) || hasDynamicInstruction(instruction.Else) || hasDynamicInstruction(instruction.Body) || hasDynamicInstruction(instruction.Catch) || hasDynamicInstruction(instruction.Finally) {
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
	pathGlobal, ok := e.stringsByValue[path]
	if !ok {
		return fmt.Errorf("dynamic module path was not interned for %q", path)
	}
	exportGlobal, ok := e.stringsByValue[export]
	if !ok {
		return fmt.Errorf("dynamic export was not interned for %q", export)
	}
	if err := e.emitDynamicModuleRegistry(out); err != nil {
		return err
	}
	pathPtr := e.dynamicStringPointer(out, pathGlobal, "dynamic.path")
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
	fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dynamic_call_module(ptr %%%s, ptr %%%s, ptr %%%s, i32 %d, i32 %d, i32 %d, ptr %%%s)\n", status, pathPtr, exportPtr, argsSlot, count, instruction.FieldIndex, dynamicResultTag(instruction.Type), outSlot)
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

func (e *functionEmitter) emitDynamicFunctionCall(out *strings.Builder, instruction ir.Instruction) error {
	if len(instruction.Args) == 0 || len(instruction.Args) > 5 {
		return fmt.Errorf("dynamic function call supports one callable and up to four arguments")
	}
	callable := e.resolveArg(out, instruction.Args[0])
	if e.types[instruction.Args[0]] == ir.TypeUnknown {
		payload := fmt.Sprintf("dynamic.function.handle.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue %s %%%s, 2\n", payload, boxedLLVMType, callable))
		callable = fmt.Sprintf("dynamic.function.handle.ptr.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", callable, payload))
	}
	pointers := []string{"ptr null", "ptr null", "ptr null", "ptr null"}
	for i, arg := range instruction.Args[1:] {
		argType := e.types[arg]
		pointer, err := e.emitCanonicalValuePointer(out, arg, argType, fmt.Sprintf("dynamic.function.arg.%d", i))
		if err != nil {
			return err
		}
		pointers[i] = "ptr " + pointer
	}
	thisPointer := e.dynamicUndefinedValuePointer(out)
	if instruction.This != "" {
		var err error
		thisPointer, err = e.emitCanonicalValuePointer(out, instruction.This, e.types[instruction.This], "dynamic.function.this")
		if err != nil {
			return err
		}
	}
	outSlot := fmt.Sprintf("dynamic.function.out.%d", e.loadCounter)
	e.loadCounter++
	out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", outSlot, boxedLLVMType))
	status := fmt.Sprintf("dynamic.function.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_dynamic_invoke_function(ptr %%%s, ptr %s, i32 %d, %s, i32 %d, ptr %%%s)\n", status, callable, thisPointer, len(instruction.Args)-1, strings.Join(pointers, ", "), dynamicResultTag(instruction.Type), outSlot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_dynamic_abort_if_failed(i32 %%%s)\n", status))
	boxed := fmt.Sprintf("dynamic.function.result.%d", e.loadCounter)
	e.loadCounter++
	out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %%%s\n", boxed, boxedLLVMType, outSlot))
	return e.emitDynamicResult(out, instruction.Type, boxed, instruction.Result)
}

func (e *functionEmitter) dynamicUndefinedValuePointer(out *strings.Builder) string {
	slot := fmt.Sprintf("dynamic.function.undefined.this.%d", e.loadCounter)
	e.loadCounter++
	out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", slot, boxedLLVMType))
	out.WriteString(fmt.Sprintf("  store %s { i32 0, i32 0, i64 0, i64 0 }, ptr %%%s\n", boxedLLVMType, slot))
	return "%" + slot
}

func (e *functionEmitter) emitDynamicModuleRegistry(out *strings.Builder) error {
	for _, module := range e.module.DynamicModules {
		pathGlobal, pathOK := e.stringsByValue[module.Path]
		sourceGlobal, sourceOK := e.stringsByValue[module.Source]
		kindGlobal, kindOK := e.stringsByValue[module.Kind]
		exportsGlobal, exportsOK := e.stringsByValue[strings.Join(module.Exports, "\x1f")]
		if !pathOK || !sourceOK || !kindOK || !exportsOK {
			return fmt.Errorf("Dynamic module registry strings were not interned for %q", module.Path)
		}
		path := e.dynamicStringPointer(out, pathGlobal, "dynamic.module.path")
		source := e.dynamicStringPointer(out, sourceGlobal, "dynamic.module.source")
		kind := e.dynamicStringPointer(out, kindGlobal, "dynamic.module.kind")
		exports := e.dynamicStringPointer(out, exportsGlobal, "dynamic.module.exports")
		status := fmt.Sprintf("dynamic.module.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dynamic_register_module(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n", status, path, source, kind, exports)
		fmt.Fprintf(out, "  call void @scriptgo_dynamic_abort_if_failed(i32 %%%s)\n", status)
		for _, dependency := range module.Imports {
			specifierGlobal, specifierOK := e.stringsByValue[dependency.Specifier]
			resolvedGlobal, resolvedOK := e.stringsByValue[dependency.Path]
			if !specifierOK || !resolvedOK {
				return fmt.Errorf("Dynamic dependency strings were not interned for %q", module.Path)
			}
			specifier := e.dynamicStringPointer(out, specifierGlobal, "dynamic.dependency.specifier")
			resolved := e.dynamicStringPointer(out, resolvedGlobal, "dynamic.dependency.path")
			dependencyStatus := fmt.Sprintf("dynamic.dependency.status.%d", e.runtimeStatus)
			e.runtimeStatus++
			fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_dynamic_register_dependency(ptr %%%s, ptr %%%s, ptr %%%s)\n", dependencyStatus, path, specifier, resolved)
			fmt.Fprintf(out, "  call void @scriptgo_dynamic_abort_if_failed(i32 %%%s)\n", dependencyStatus)
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
	case ir.TypeDynamicFunction:
		return 7
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
	case ir.TypeDynamicFunction:
		payload := fmt.Sprintf("dynamic.function.payload.%d", e.loadCounter)
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
