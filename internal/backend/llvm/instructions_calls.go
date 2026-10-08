package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitPrint(out *strings.Builder, instruction ir.Instruction) error {
	valueType, ok := e.types[instruction.Args[0]]
	if !ok {
		return fmt.Errorf("unknown print value %q", instruction.Args[0])
	}
	method := "log"
	if instruction.Callee != "" {
		method = strings.TrimPrefix(instruction.Callee, "console.")
	}
	if _, ok := consoleRuntimeName(method, valueType); !ok {
		return fmt.Errorf("unsupported console intrinsic %q for %s", instruction.Callee, valueType)
	}
	switch valueType {
	case ir.TypeVoid:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		undefGlobal := e.stringsByValue["undefined"]
		ptrUndef := fmt.Sprintf("print.undef.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [10 x i8], ptr %s, i64 0, i64 0\n", ptrUndef, undefGlobal))
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s)\n", status, name, ptrUndef))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case ir.TypeNumber:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(double %%%s)\n", status, name, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case ir.TypeString:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s)\n", status, name, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case ir.TypeBool:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		boolValue := fmt.Sprintf("print.bool.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i32\n", boolValue, instruction.Args[0]))
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(i32 %%%s)\n", status, name, boolValue))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case ir.TypeBigInt:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(i64 %%%s)\n", status, name, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case ir.TypeSymbol:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s)\n", status, name, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	case ir.TypeUnknown:
		valuePtr, err := e.emitCanonicalValuePointer(out, instruction.Args[0], ir.TypeUnknown, "console.value")
		if err != nil {
			return err
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %s)\n", status, name, valuePtr))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	default:
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		name, _ := consoleRuntimeName(method, valueType)
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @%s(ptr %%%s)\n", status, name, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	}
	return nil
}

func (e *functionEmitter) emitCall(out *strings.Builder, instruction ir.Instruction) error {
	if handled, err := e.tryEmitCoreIntrinsicCall(out, instruction); handled {
		return err
	}

	if handled, err := e.tryEmitSystemIntrinsicCall(out, instruction); handled {
		return err
	}

	if strings.HasPrefix(instruction.Callee, "__iterator.") {
		return e.emitIteratorIntrinsicCall(out, instruction)
	}

	if handled, err := e.tryEmitServiceIntrinsicCall(out, instruction); handled {
		return err
	}

	if handled, err := e.tryEmitValueIntrinsicCall(out, instruction); handled {
		return err
	}

	callee, ok := e.functions[instruction.Callee]
	if !ok {
		for _, ext := range e.module.Externs {
			if ext.Name == instruction.Callee {
				callee = ir.Function{
					Name:       ext.Name,
					Parameters: ext.Parameters,
					ReturnType: ext.ReturnType,
				}
				ok = true
				break
			}
		}
	}
	if !ok {
		return fmt.Errorf("unknown function %q", instruction.Callee)
	}
	if len(callee.Parameters) != len(instruction.Args) {
		return fmt.Errorf("call to %q has wrong arity", instruction.Callee)
	}
	var callArgs []string
	for index, argument := range instruction.Args {
		paramType := llvmType(callee.Parameters[index].Type)
		argVal := e.resolveArg(out, argument)
		if strings.HasPrefix(e.function.Name, "__top_level_async_stage_") && !e.isParam(argument) {
			for _, global := range e.module.Globals {
				if global.Name == argument {
					loadName := fmt.Sprintf("%s.stage_global.%d", argument, e.loadCounter)
					e.loadCounter++
					typ := e.types[argument]
					if typ == "" || typ == ir.TypeVoid {
						typ = global.Type
					}
					if typ == "" || typ == ir.TypeVoid {
						typ = ir.TypePointer
					}
					e.types[loadName] = typ
					out.WriteString(fmt.Sprintf("  %%%s = load volatile %s, ptr @%s\n", loadName, llvmType(typ), global.Name))
					argVal = loadName
					break
				}
			}
		}
		if isRawCallbackParameter(callee.Parameters[index]) {
			boxed := argVal
			for field, fieldType := range []string{"i32", "i32", "i64"} {
				fieldName := fmt.Sprintf("call.raw.%s.%d", []string{"tag", "pad", "payload"}[field], e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, %d\n", fieldName, boxed, field))
				callArgs = append(callArgs, fmt.Sprintf("%s %%%s", fieldType, fieldName))
			}
			continue
		}
		argType, hasArgType := e.types[argVal]
		if !hasArgType {
			argType, hasArgType = e.types[argument]
		}
		if callee.Parameters[index].Type == ir.TypeUnknown && hasArgType && argType != ir.TypeUnknown {
			boxedName := fmt.Sprintf("call.box.%d", e.loadCounter)
			if err := e.emitBoxValue(out, argVal, argType, boxedName); err != nil {
				return err
			}
			argVal = boxedName
		} else if hasArgType && argType == ir.TypeUnknown && callee.Parameters[index].Type != ir.TypeUnknown {
			e.tempCounter++
			payloadName := fmt.Sprintf("call.unbox.payload.%d", e.tempCounter)
			fmt.Fprintf(out, "  %%%s = extractvalue { i32, i32, i64, i64 } %%%s, 2\n", payloadName, argument)
			switch paramType {
			case "double":
				e.tempCounter++
				valName := fmt.Sprintf("call.unbox.dbl.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = bitcast i64 %%%s to double\n", valName, payloadName)
				argVal = valName
			case "i1":
				e.tempCounter++
				valName := fmt.Sprintf("call.unbox.bool.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = trunc i64 %%%s to i1\n", valName, payloadName)
				argVal = valName
			case "i64":
				argVal = payloadName
			default:
				e.tempCounter++
				valName := fmt.Sprintf("call.unbox.ptr.%d", e.tempCounter)
				fmt.Fprintf(out, "  %%%s = inttoptr i64 %%%s to %s\n", valName, payloadName, paramType)
				argVal = valName
			}
		} else if hasArgType && argType == ir.TypeVoid && paramType == "double" {
			nanName := fmt.Sprintf("call.nan.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = fadd double 0.0, 0x7FF8000000000000\n", nanName))
			argVal = nanName
		}
		callArgs = append(callArgs, fmt.Sprintf("%s %%%s", paramType, argVal))
	}
	returnType := llvmType(callee.ReturnType)
	if callee.ReturnType == ir.TypeBool {
		returnType = "zeroext i1"
	}
	if returnType == "void" {
		out.WriteString(fmt.Sprintf("  call void @%s(", mangleFunctionName(instruction.Callee)))
	} else {
		e.types[instruction.Result] = callee.ReturnType
		out.WriteString(fmt.Sprintf("  %%%s = call %s @%s(", instruction.Result, returnType, mangleFunctionName(instruction.Callee)))
	}
	out.WriteString(strings.Join(callArgs, ", "))
	out.WriteString(")\n")
	return nil
}

func isJSArrayType(t ir.Type) bool {
	return t == ir.TypeNumberArray || t == ir.TypeStringArray ||
		t == ir.TypeBoolArray || t == ir.TypeBigIntArray ||
		t == ir.TypeSymbolArray || t == ir.TypeUnknownArray ||
		strings.HasSuffix(string(t), "[]")
}
