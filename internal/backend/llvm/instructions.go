package llvm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

type arrayReference struct {
	name string
	typ  ir.Type
}

type functionEmitter struct {
	function       ir.Function
	functions      map[string]ir.Function
	stringsByValue map[string]string
	debug          *debugInfo

	types         map[string]ir.Type
	arrayTypes    []arrayReference
	objects       []string
	ownedStrings  []string
	labelCounter  int
	runtimeStatus int
	terminated    bool
}

func (e *functionEmitter) emitInstruction(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Op {
	case ir.OpConst:
		return e.emitConst(out, instruction)
	case ir.OpBinary:
		return e.emitBinary(out, instruction)
	case ir.OpCompare:
		return e.emitCompare(out, instruction)
	case ir.OpSelect:
		return e.emitSelect(out, instruction)
	case ir.OpIf:
		return e.emitIf(out, instruction)
	case ir.OpPrint:
		return e.emitPrint(out, instruction)
	case ir.OpArray:
		return e.emitArray(out, instruction)
	case ir.OpIndex:
		return e.emitIndex(out, instruction)
	case ir.OpObjectNew:
		return e.emitObjectNew(out, instruction)
	case ir.OpFieldSet:
		return e.emitFieldSet(out, instruction)
	case ir.OpFieldGet:
		return e.emitFieldGet(out, instruction)
	case ir.OpCall:
		return e.emitCall(out, instruction)
	case ir.OpReturn:
		return e.emitReturn(out, instruction)
	default:
		return fmt.Errorf("unsupported LLVM instruction %q", instruction.Op)
	}
}

func (e *functionEmitter) emitConst(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = instruction.Type
	switch instruction.Type {
	case ir.TypeNumber:
		number, err := strconv.ParseFloat(instruction.Value, 64)
		if err != nil {
			return fmt.Errorf("invalid number %q: %w", instruction.Value, err)
		}
		out.WriteString(fmt.Sprintf("  %%%s = fadd double 0.0, %s\n", instruction.Result, llvmNumber(number)))
	case ir.TypeString:
		global := e.stringsByValue[instruction.Value]
		length := len([]byte(instruction.Value)) + 1
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [%d x i8], ptr %s, i64 0, i64 0\n", instruction.Result, length, global))
	case ir.TypeBool:
		out.WriteString(fmt.Sprintf("  %%%s = or i1 false, %s\n", instruction.Result, instruction.Value))
	default:
		return fmt.Errorf("unsupported constant type %s", instruction.Type)
	}
	return nil
}

func (e *functionEmitter) emitBinary(out *strings.Builder, instruction ir.Instruction) error {
	leftType, ok := e.types[instruction.Args[0]]
	if !ok {
		return fmt.Errorf("unknown binary value %q", instruction.Args[0])
	}
	if _, ok := e.types[instruction.Args[1]]; !ok {
		return fmt.Errorf("unknown binary value %q", instruction.Args[1])
	}
	if leftType == ir.TypeString && instruction.Operator == "+" {
		e.types[instruction.Result] = ir.TypeString
		slot := instruction.Result + ".slot"
		status := instruction.Result + ".status"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_string_concat(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		e.ownedStrings = append(e.ownedStrings, instruction.Result)
		return nil
	}
	if leftType != ir.TypeNumber {
		return fmt.Errorf("LLVM binary operator %q only supports number or string concatenation", instruction.Operator)
	}
	op, ok := map[string]string{"+": "fadd", "-": "fsub", "*": "fmul", "/": "fdiv", "%": "frem"}[instruction.Operator]
	if !ok {
		return fmt.Errorf("unsupported LLVM binary operator %q", instruction.Operator)
	}
	e.types[instruction.Result] = instruction.Type
	out.WriteString(fmt.Sprintf("  %%%s = %s double %%%s, %%%s\n", instruction.Result, op, instruction.Args[0], instruction.Args[1]))
	return nil
}

func (e *functionEmitter) emitCompare(out *strings.Builder, instruction ir.Instruction) error {
	leftType, ok := e.types[instruction.Args[0]]
	if !ok || e.types[instruction.Args[1]] != leftType {
		return fmt.Errorf("unknown or mismatched compare operands")
	}
	if leftType != ir.TypeNumber {
		return fmt.Errorf("LLVM compare only supports number operands")
	}
	predicate, ok := map[string]string{"==": "oeq", "!==": "une", "<": "olt", "<=": "ole", ">": "ogt", ">=": "oge"}[instruction.Operator]
	if !ok {
		return fmt.Errorf("unsupported LLVM compare operator %q", instruction.Operator)
	}
	e.types[instruction.Result] = ir.TypeBool
	out.WriteString(fmt.Sprintf("  %%%s = fcmp %s double %%%s, %%%s\n", instruction.Result, predicate, instruction.Args[0], instruction.Args[1]))
	return nil
}

func (e *functionEmitter) emitSelect(out *strings.Builder, instruction ir.Instruction) error {
	if e.types[instruction.Args[0]] != ir.TypeBool || e.types[instruction.Args[1]] != e.types[instruction.Args[2]] {
		return fmt.Errorf("select operands have incompatible types")
	}
	e.types[instruction.Result] = instruction.Type
	out.WriteString(fmt.Sprintf("  %%%s = select i1 %%%s, %s %%%s, %s %%%s\n", instruction.Result, instruction.Args[0], llvmType(instruction.Type), instruction.Args[1], llvmType(instruction.Type), instruction.Args[2]))
	return nil
}

func (e *functionEmitter) emitIf(out *strings.Builder, instruction ir.Instruction) error {
	if len(instruction.Args) != 1 || e.types[instruction.Args[0]] != ir.TypeBool || len(instruction.Then) != 1 || instruction.Then[0].Op != ir.OpReturn || len(instruction.Else) != 0 {
		return fmt.Errorf("LLVM if currently requires a returning then branch and empty else branch")
	}
	thenLabel := fmt.Sprintf("if.then.%d", e.labelCounter)
	continueLabel := fmt.Sprintf("if.continue.%d", e.labelCounter)
	e.labelCounter++
	out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s\n", instruction.Args[0], thenLabel, continueLabel))
	out.WriteString(fmt.Sprintf("%s:\n", thenLabel))
	branchReturn := instruction.Then[0]
	if len(branchReturn.Args) == 0 {
		out.WriteString("  ret void\n")
	} else {
		out.WriteString(fmt.Sprintf("  ret %s %%%s\n", llvmType(branchReturn.Type), branchReturn.Args[0]))
	}
	out.WriteString(fmt.Sprintf("%s:\n", continueLabel))
	return nil
}

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
	default:
		return fmt.Errorf("unsupported print type %s", valueType)
	}
	return nil
}

func (e *functionEmitter) emitCall(out *strings.Builder, instruction ir.Instruction) error {
	if strings.HasPrefix(instruction.Callee, "__Math.") {
		if err := emitMathIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__array.") {
		arrayType, ok := e.types[instruction.Args[0]]
		if !ok {
			return fmt.Errorf("unknown array intrinsic argument %q", instruction.Args[0])
		}
		if err := emitArrayIntrinsic(out, instruction, arrayType); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__string.") {
		if err := emitStringIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString && (instruction.Callee == "__string.slice" || instruction.Callee == "__string.concat") {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return nil
	}
	callee, ok := e.functions[instruction.Callee]
	if !ok {
		return fmt.Errorf("unknown function %q", instruction.Callee)
	}
	if len(callee.Parameters) != len(instruction.Args) {
		return fmt.Errorf("call to %q has wrong arity", instruction.Callee)
	}
	returnType := llvmType(callee.ReturnType)
	if returnType == "void" {
		out.WriteString(fmt.Sprintf("  call void @%s(", instruction.Callee))
	} else {
		e.types[instruction.Result] = callee.ReturnType
		out.WriteString(fmt.Sprintf("  %%%s = call %s @%s(", instruction.Result, returnType, instruction.Callee))
	}
	for index, argument := range instruction.Args {
		if index > 0 {
			out.WriteString(", ")
		}
		out.WriteString(fmt.Sprintf("%s %%%s", llvmType(callee.Parameters[index].Type), argument))
	}
	out.WriteString(")\n")
	return nil
}

func (e *functionEmitter) emitReturn(out *strings.Builder, instruction ir.Instruction) error {
	for _, arrayRef := range e.arrayTypes {
		out.WriteString(fmt.Sprintf("  call i32 @scriptgo_array_release(ptr %%%s)\n", arrayRef.name))
	}
	for _, object := range e.objects {
		out.WriteString(fmt.Sprintf("  call i32 @scriptgo_object_release(ptr %%%s)\n", object))
	}
	returnValue := ""
	if len(instruction.Args) != 0 {
		returnValue = instruction.Args[0]
	}
	if e.function.Name == "main" {
		for _, value := range e.ownedStrings {
			if value != returnValue {
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_string_release(ptr %%%s)\n", value))
			}
		}
	}
	if e.function.Name == "main" {
		out.WriteString("  ret i32 0\n")
	} else if len(instruction.Args) == 0 {
		out.WriteString("  ret void\n")
	} else {
		out.WriteString(fmt.Sprintf("  ret %s %%%s\n", llvmType(instruction.Type), instruction.Args[0]))
	}
	e.terminated = true
	return nil
}
