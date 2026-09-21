package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitObjectNew(out *strings.Builder, instruction ir.Instruction) error {
	if instruction.FieldCount < 0 {
		return fmt.Errorf("object shape %q has invalid field count", instruction.Callee)
	}
	e.types[instruction.Result] = instruction.Type
	e.objects = append(e.objects, instruction.Result)
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))

	typeName := instruction.Value
	if typeName == "" {
		for _, s := range e.module.Shapes {
			if s.Name == instruction.Callee && len(s.Fields) > 0 {
				var names []string
				for _, f := range s.Fields {
					names = append(names, f.Name)
				}
				typeName = ":" + strings.Join(names, ":") + ":"
				break
			}
		}
	}
	if typeName == "" {
		typeName = instruction.Callee
	}
	if typeName != "" {
		if strGlobal, ok := e.stringsByValue[typeName]; ok {
			out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_new_typed(i64 %d, ptr %s, ptr %%%s)\n", status, instruction.FieldCount, strGlobal, slot))
			out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
			return nil
		}
	}

	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_new(i64 %d, ptr %%%s)\n", status, instruction.FieldCount, slot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	return nil
}

func (e *functionEmitter) emitInstanceOf(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = ir.TypeBool
	slot := instruction.Result + ".slot"
	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = alloca i32\n", slot))
	var classArg string
	if instruction.Value != "" {
		strGlobal, ok := e.stringsByValue[instruction.Value]
		if !ok {
			return fmt.Errorf("unknown string literal %q for instanceof", instruction.Value)
		}
		classArg = strGlobal
	} else if len(instruction.Args) > 1 {
		cls := instruction.Args[1]
		clsVal := cls
		if slot, ok := e.varSlots[cls]; ok {
			loaded := fmt.Sprintf("%s.instanceof_cls.%d", cls, e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", loaded, slot))
			clsVal = loaded
		}
		classArg = "%" + clsVal
	} else {
		return fmt.Errorf("instanceof requires target class or property")
	}

	arg := instruction.Args[0]
	argType := e.types[arg]
	argVal := arg
	if slot, ok := e.varSlots[arg]; ok {
		loaded := fmt.Sprintf("%s.instanceof_load.%d", arg, e.loadCounter)
		e.loadCounter++
		if argType == ir.TypeUnknown {
			out.WriteString(fmt.Sprintf("  %%%s = load { i32, i32, i64, i64 }, ptr %%%s\n", loaded, slot))
		} else {
			out.WriteString(fmt.Sprintf("  %%%s = load%s ptr, ptr %%%s\n", loaded, e.vol(), slot))
		}
		argVal = loaded
	}
	ptrArg := "%" + argVal
	if argType == ir.TypeUnknown {
		payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
		ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadVar, argVal))
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
		ptrArg = "%" + ptrVar
	}
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_object_instanceof(ptr %s, ptr %s, ptr %%%s)\n", status, ptrArg, classArg, slot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	i32Val := instruction.Result + ".i32"
	out.WriteString(fmt.Sprintf("  %%%s = load i32, ptr %%%s\n", i32Val, slot))
	out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", instruction.Result, i32Val))
	return nil
}
