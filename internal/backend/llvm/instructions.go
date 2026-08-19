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
	labelCounter       int
	loopBreakLabels    []string
	loopContinueLabels []string
	runtimeStatus      int
	terminated         bool
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
	case ir.OpDoWhile:
		return e.emitDoWhile(out, inst)
	case ir.OpBreak:
		if len(e.loopBreakLabels) == 0 {
			return fmt.Errorf("break outside of loop")
		}
		breakLabel := e.loopBreakLabels[len(e.loopBreakLabels)-1]
		out.WriteString(fmt.Sprintf("  br label %%%s\n", breakLabel))
		e.terminated = true
		return nil
	case ir.OpContinue:
		if len(e.loopContinueLabels) == 0 {
			return fmt.Errorf("continue outside of loop")
		}
		continueLabel := e.loopContinueLabels[len(e.loopContinueLabels)-1]
		out.WriteString(fmt.Sprintf("  br label %%%s\n", continueLabel))
		e.terminated = true
		return nil
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
	case ir.OpIndexSet:
		return e.emitIndexSet(out, inst)
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
	case ir.OpThrow:
		return e.emitThrow(out, inst)
	case ir.OpTry:
		return e.emitTry(out, inst)
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

	e.loopBreakLabels = append(e.loopBreakLabels, endLabel)
	e.loopContinueLabels = append(e.loopContinueLabels, condLabel)
	defer func() {
		e.loopBreakLabels = e.loopBreakLabels[:len(e.loopBreakLabels)-1]
		e.loopContinueLabels = e.loopContinueLabels[:len(e.loopContinueLabels)-1]
	}()

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

func (e *functionEmitter) emitDoWhile(out *strings.Builder, instruction ir.Instruction) error {
	bodyLabel := fmt.Sprintf("dowhile.body.%d", e.labelCounter)
	condLabel := fmt.Sprintf("dowhile.cond.%d", e.labelCounter)
	endLabel := fmt.Sprintf("dowhile.end.%d", e.labelCounter)
	e.labelCounter++

	e.loopBreakLabels = append(e.loopBreakLabels, endLabel)
	e.loopContinueLabels = append(e.loopContinueLabels, condLabel)
	defer func() {
		e.loopBreakLabels = e.loopBreakLabels[:len(e.loopBreakLabels)-1]
		e.loopContinueLabels = e.loopContinueLabels[:len(e.loopContinueLabels)-1]
	}()

	out.WriteString(fmt.Sprintf("  br label %%%s\n", bodyLabel))
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

	out.WriteString(fmt.Sprintf("%s:\n", condLabel))
	e.terminated = false
	for _, inst := range instruction.Cond {
		if err := e.emitInstruction(out, inst); err != nil {
			return err
		}
	}
	condVal := e.resolveArg(out, instruction.Args[0])
	out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s\n", condVal, bodyLabel, endLabel))

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
		if err := e.emitArrayIntrinsic(out, instruction, arrayType); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__fs.") {
		if err := e.emitFsIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__process.") {
		if err := e.emitProcessIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__crypto.") {
		if err := e.emitCryptoIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__web.") {
		if err := e.emitWebIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__performance.") {
		if err := e.emitPerformanceIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
		}
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

func (e *functionEmitter) emitThrow(out *strings.Builder, instruction ir.Instruction) error {
	argVal := instruction.Args[0]
	argType, ok := e.types[argVal]
	if !ok {
		return fmt.Errorf("unknown throw argument %q", argVal)
	}
	switch argType {
	case ir.TypeString:
		out.WriteString(fmt.Sprintf("  call void @scriptgo_throw_string(ptr %%%s)\n", argVal))
	case ir.TypeNumber:
		out.WriteString(fmt.Sprintf("  call void @scriptgo_throw_number(double %%%s)\n", argVal))
	case ir.TypeBool:
		boolVal := fmt.Sprintf("throw.bool.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i32\n", boolVal, argVal))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_throw_bool(i32 %%%s)\n", boolVal))
	default:
		out.WriteString(fmt.Sprintf("  call void @scriptgo_throw_string(ptr %%%s)\n", argVal))
	}
	out.WriteString("  unreachable\n")
	e.terminated = true
	return nil
}

func (e *functionEmitter) emitTry(out *strings.Builder, instruction ir.Instruction) error {
	labelId := e.labelCounter
	e.labelCounter++

	frameName := fmt.Sprintf("eh.frame.%d", labelId)
	bufName := fmt.Sprintf("eh.buf.%d", labelId)
	statusName := fmt.Sprintf("eh.status.%d", labelId)
	isThrowName := fmt.Sprintf("eh.is_throw.%d", labelId)

	tryBodyLabel := fmt.Sprintf("try.body.%d", labelId)
	catchLabel := fmt.Sprintf("try.catch.%d", labelId)
	finallyLabel := fmt.Sprintf("try.finally.%d", labelId)
	endLabel := fmt.Sprintf("try.end.%d", labelId)

	hasCatch := len(instruction.Catch) > 0 || instruction.CatchVar != ""
	hasFinally := len(instruction.Finally) > 0

	landingLabel := catchLabel
	if !hasCatch {
		landingLabel = finallyLabel
	}

	out.WriteString(fmt.Sprintf("  %%%s = alloca [1024 x i8], align 16\n", frameName))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_exception_push(ptr %%%s)\n", frameName))
	out.WriteString(fmt.Sprintf("  %%%s = call ptr @scriptgo_exception_buf(ptr %%%s)\n", bufName, frameName))
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @setjmp(ptr %%%s) returns_twice\n", statusName, bufName))
	out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", isThrowName, statusName))
	out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s\n", isThrowName, landingLabel, tryBodyLabel))

	out.WriteString(fmt.Sprintf("%s:\n", tryBodyLabel))
	e.terminated = false
	for _, inst := range instruction.Body {
		if err := e.emitInstruction(out, inst); err != nil {
			return err
		}
	}
	tryTerminated := e.terminated
	if !tryTerminated {
		out.WriteString(fmt.Sprintf("  call void @scriptgo_exception_pop(ptr %%%s)\n", frameName))
		if hasFinally {
			out.WriteString(fmt.Sprintf("  br label %%%s\n", finallyLabel))
		} else {
			out.WriteString(fmt.Sprintf("  br label %%%s\n", endLabel))
		}
	}

	catchTerminated := false
	if hasCatch {
		out.WriteString(fmt.Sprintf("%s:\n", catchLabel))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_exception_pop(ptr %%%s)\n", frameName))
		if instruction.CatchVar != "" {
			catchValName := fmt.Sprintf("caught.%d", labelId)
			out.WriteString(fmt.Sprintf("  %%%s = call ptr @scriptgo_exception_get_string(ptr %%%s)\n", catchValName, frameName))
			e.types[instruction.CatchVar] = ir.TypeString
			e.types[catchValName] = ir.TypeString
			if slot, ok := e.varSlots[instruction.CatchVar]; ok {
				out.WriteString(fmt.Sprintf("  store ptr %%%s, ptr %%%s\n", catchValName, slot))
			} else {
				slotName := fmt.Sprintf("slot.%s", instruction.CatchVar)
				out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slotName))
				out.WriteString(fmt.Sprintf("  store ptr %%%s, ptr %%%s\n", catchValName, slotName))
				e.varSlots[instruction.CatchVar] = slotName
			}
		}
		e.terminated = false
		for _, inst := range instruction.Catch {
			if err := e.emitInstruction(out, inst); err != nil {
				return err
			}
		}
		catchTerminated = e.terminated
		if !catchTerminated {
			if hasFinally {
				out.WriteString(fmt.Sprintf("  br label %%%s\n", finallyLabel))
			} else {
				out.WriteString(fmt.Sprintf("  br label %%%s\n", endLabel))
			}
		}
	}

	finallyTerminated := false
	if hasFinally {
		out.WriteString(fmt.Sprintf("%s:\n", finallyLabel))
		if !hasCatch {
			out.WriteString(fmt.Sprintf("  call void @scriptgo_exception_pop(ptr %%%s)\n", frameName))
		}
		e.terminated = false
		for _, inst := range instruction.Finally {
			if err := e.emitInstruction(out, inst); err != nil {
				return err
			}
		}
		finallyTerminated = e.terminated
		if !finallyTerminated {
			if !hasCatch {
				afterRethrow := fmt.Sprintf("finally.after_rethrow.%d", labelId)
				rethrowCond := fmt.Sprintf("rethrow.cond.%d", labelId)
				out.WriteString(fmt.Sprintf("  %%%s = icmp ne i32 %%%s, 0\n", rethrowCond, statusName))
				rethrowBlock := fmt.Sprintf("finally.rethrow.%d", labelId)
				out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s\n", rethrowCond, rethrowBlock, afterRethrow))
				out.WriteString(fmt.Sprintf("%s:\n", rethrowBlock))
				out.WriteString(fmt.Sprintf("  call void @scriptgo_exception_rethrow(ptr %%%s)\n", frameName))
				out.WriteString("  unreachable\n")
				out.WriteString(fmt.Sprintf("%s:\n", afterRethrow))
			}
			out.WriteString(fmt.Sprintf("  br label %%%s\n", endLabel))
		}
	}

	if (!tryTerminated) || (hasCatch && !catchTerminated) || (hasFinally && !finallyTerminated) {
		out.WriteString(fmt.Sprintf("%s:\n", endLabel))
		e.terminated = false
	} else {
		e.terminated = true
	}
	return nil
}
