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
	varSlots      map[string]string
	loadCounter   int
	arrayTypes    []arrayReference
	objects       []string
	ownedStrings  []string
	labelCounter  int
	runtimeStatus int
	terminated    bool
}

func (e *functionEmitter) resolveArg(out *strings.Builder, arg string) string {
	slot, ok := e.varSlots[arg]
	if !ok {
		return arg
	}
	typ := e.types[arg]
	loadName := fmt.Sprintf("%s.load.%d", arg, e.loadCounter)
	e.loadCounter++
	e.types[loadName] = typ
	out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %%%s\n", loadName, llvmType(typ), slot))
	return loadName
}

func (e *functionEmitter) emitInstruction(out *strings.Builder, instruction ir.Instruction) error {
	resolvedArgs := make([]string, len(instruction.Args))
	for i, arg := range instruction.Args {
		resolvedArgs[i] = e.resolveArg(out, arg)
	}
	inst := instruction
	inst.Args = resolvedArgs

	targetResult := inst.Result
	if _, ok := e.varSlots[inst.Result]; ok {
		inst.Result = fmt.Sprintf("%s.val.%d", inst.Result, e.loadCounter)
		e.loadCounter++
	}

	switch inst.Op {
	case ir.OpConst:
		if err := e.emitConst(out, inst); err != nil {
			return err
		}
	case ir.OpAssign:
		typ := e.types[targetResult]
		out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", llvmType(typ), inst.Args[0], e.varSlots[targetResult]))
		return nil
	case ir.OpBinary:
		if err := e.emitBinary(out, inst); err != nil {
			return err
		}
	case ir.OpCompare:
		if err := e.emitCompare(out, inst); err != nil {
			return err
		}
	case ir.OpSelect:
		if err := e.emitSelect(out, inst); err != nil {
			return err
		}
	case ir.OpIf:
		return e.emitIf(out, inst)
	case ir.OpWhile:
		return e.emitWhile(out, inst)
	case ir.OpPrint:
		return e.emitPrint(out, inst)
	case ir.OpArray:
		if err := e.emitArray(out, inst); err != nil {
			return err
		}
	case ir.OpIndex:
		if err := e.emitIndex(out, inst); err != nil {
			return err
		}
	case ir.OpObjectNew:
		if err := e.emitObjectNew(out, inst); err != nil {
			return err
		}
	case ir.OpFieldSet:
		return e.emitFieldSet(out, inst)
	case ir.OpFieldGet:
		if err := e.emitFieldGet(out, inst); err != nil {
			return err
		}
	case ir.OpCall:
		if err := e.emitCall(out, inst); err != nil {
			return err
		}
	case ir.OpReturn:
		return e.emitReturn(out, inst)
	default:
		return fmt.Errorf("unsupported LLVM instruction %q", inst.Op)
	}

	if slot, ok := e.varSlots[targetResult]; ok {
		typ := e.types[targetResult]
		out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", llvmType(typ), inst.Result, slot))
	}
	return nil
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
	if leftType == ir.TypeBool {
		op, ok := map[string]string{"&&": "and", "||": "or"}[instruction.Operator]
		if !ok {
			return fmt.Errorf("unsupported LLVM binary bool operator %q", instruction.Operator)
		}
		e.types[instruction.Result] = ir.TypeBool
		out.WriteString(fmt.Sprintf("  %%%s = %s i1 %%%s, %%%s\n", instruction.Result, op, instruction.Args[0], instruction.Args[1]))
		return nil
	}
	if leftType != ir.TypeNumber {
		return fmt.Errorf("LLVM binary operator %q only supports number, bool, or string concatenation", instruction.Operator)
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
	if leftType == ir.TypeString {
		cmpResult := instruction.Result + ".cmp"
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @strcmp(ptr %%%s, ptr %%%s)\n", cmpResult, instruction.Args[0], instruction.Args[1]))
		predicate, ok := map[string]string{
			"==": "eq", "===": "eq",
			"!=": "ne", "!==": "ne",
			"<": "slt", "<=": "sle",
			">": "sgt", ">=": "sge",
		}[instruction.Operator]
		if !ok {
			return fmt.Errorf("unsupported LLVM string compare operator %q", instruction.Operator)
		}
		e.types[instruction.Result] = ir.TypeBool
		out.WriteString(fmt.Sprintf("  %%%s = icmp %s i32 %%%s, 0\n", instruction.Result, predicate, cmpResult))
		return nil
	}
	if leftType == ir.TypeBool {
		predicate, ok := map[string]string{
			"==": "eq", "===": "eq",
			"!=": "ne", "!==": "ne",
		}[instruction.Operator]
		if !ok {
			return fmt.Errorf("unsupported LLVM bool compare operator %q", instruction.Operator)
		}
		e.types[instruction.Result] = ir.TypeBool
		out.WriteString(fmt.Sprintf("  %%%s = icmp %s i1 %%%s, %%%s\n", instruction.Result, predicate, instruction.Args[0], instruction.Args[1]))
		return nil
	}
	if leftType != ir.TypeNumber {
		return fmt.Errorf("LLVM compare only supports number, string, or bool operands")
	}
	predicate, ok := map[string]string{
		"==": "oeq", "===": "oeq",
		"!=": "une", "!==": "une",
		"<": "olt", "<=": "ole",
		">": "ogt", ">=": "oge",
	}[instruction.Operator]
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
	if len(instruction.Args) != 1 || e.types[instruction.Args[0]] != ir.TypeBool {
		return fmt.Errorf("if requires a bool condition")
	}
	thenLabel := fmt.Sprintf("if.then.%d", e.labelCounter)
	elseLabel := fmt.Sprintf("if.else.%d", e.labelCounter)
	continueLabel := fmt.Sprintf("if.continue.%d", e.labelCounter)
	e.labelCounter++

	hasElse := len(instruction.Else) > 0
	if hasElse {
		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s\n", instruction.Args[0], thenLabel, elseLabel))
	} else {
		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s\n", instruction.Args[0], thenLabel, continueLabel))
	}

	out.WriteString(fmt.Sprintf("%s:\n", thenLabel))
	e.terminated = false
	for _, inst := range instruction.Then {
		if err := e.emitInstruction(out, inst); err != nil {
			return err
		}
	}
	thenTerminated := e.terminated
	if !thenTerminated {
		out.WriteString(fmt.Sprintf("  br label %%%s\n", continueLabel))
	}

	elseTerminated := false
	if hasElse {
		out.WriteString(fmt.Sprintf("%s:\n", elseLabel))
		e.terminated = false
		for _, inst := range instruction.Else {
			if err := e.emitInstruction(out, inst); err != nil {
				return err
			}
		}
		elseTerminated = e.terminated
		if !elseTerminated {
			out.WriteString(fmt.Sprintf("  br label %%%s\n", continueLabel))
		}
	}

	if !thenTerminated || (hasElse && !elseTerminated) || (!hasElse) {
		out.WriteString(fmt.Sprintf("%s:\n", continueLabel))
		e.terminated = false
	} else {
		e.terminated = true
	}
	return nil
}

func (e *functionEmitter) emitWhile(out *strings.Builder, instruction ir.Instruction) error {
	condLabel := fmt.Sprintf("while.cond.%d", e.labelCounter)
	bodyLabel := fmt.Sprintf("while.body.%d", e.labelCounter)
	endLabel := fmt.Sprintf("while.end.%d", e.labelCounter)
	e.labelCounter++

	out.WriteString(fmt.Sprintf("  br label %%%s\n", condLabel))
	out.WriteString(fmt.Sprintf("%s:\n", condLabel))

	e.terminated = false
	for _, inst := range instruction.Cond {
		if err := e.emitInstruction(out, inst); err != nil {
			return err
		}
	}
	condVal := e.resolveArg(out, instruction.Args[0])
	out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s\n", condVal, bodyLabel, endLabel))

	out.WriteString(fmt.Sprintf("%s:\n", bodyLabel))
	e.terminated = false
	for _, inst := range instruction.Body {
		if err := e.emitInstruction(out, inst); err != nil {
			return err
		}
	}
	if !e.terminated {
		out.WriteString(fmt.Sprintf("  br label %%%s\n", condLabel))
	}

	out.WriteString(fmt.Sprintf("%s:\n", endLabel))
	e.terminated = false
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
		if instruction.Type == ir.TypeString && (instruction.Callee == "__string.slice" || instruction.Callee == "__string.concat" || instruction.Callee == "__string.fromNumber" || instruction.Callee == "__string.fromBool") {
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
