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
	case ir.OpClosure:
		if err := e.emitClosure(out, inst); err != nil {
			return err
		}
	case ir.OpClosureCall:
		if err := e.emitClosureCall(out, inst); err != nil {
			return err
		}
	case ir.OpReturn:
		return e.emitReturn(out, inst)
	case ir.OpThrow:
		return e.emitThrow(out, inst)
	case ir.OpTry:
		return e.emitTry(out, inst)
	case ir.OpInstanceOf:
		return e.emitInstanceOf(out, inst)
	case ir.OpBoxUnknown:
		if err := e.emitBoxUnknown(out, inst); err != nil {
			return err
		}
	case ir.OpCheckedCast:
		if err := e.emitCheckedCast(out, inst); err != nil {
			return err
		}
	case ir.OpTypeOf:
		if err := e.emitTypeOf(out, inst); err != nil {
			return err
		}
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
	if instruction.Operator == "**" {
		e.types[instruction.Result] = instruction.Type
		out.WriteString(fmt.Sprintf("  %%%s = call double @llvm.pow.f64(double %%%s, double %%%s)\n", instruction.Result, instruction.Args[0], instruction.Args[1]))
		return nil
	}
	if op, ok := map[string]string{"+": "fadd", "-": "fsub", "*": "fmul", "/": "fdiv", "%": "frem"}[instruction.Operator]; ok {
		e.types[instruction.Result] = instruction.Type
		out.WriteString(fmt.Sprintf("  %%%s = %s double %%%s, %%%s\n", instruction.Result, op, instruction.Args[0], instruction.Args[1]))
		return nil
	}
	if bitOp, ok := map[string]string{"&": "and", "|": "or", "^": "xor"}[instruction.Operator]; ok {
		e.types[instruction.Result] = instruction.Type
		lI32 := instruction.Result + ".l_i32"
		rI32 := instruction.Result + ".r_i32"
		resI32 := instruction.Result + ".res_i32"
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", lI32, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", rI32, instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  %%%s = %s i32 %%%s, %%%s\n", resI32, bitOp, lI32, rI32))
		out.WriteString(fmt.Sprintf("  %%%s = sitofp i32 %%%s to double\n", instruction.Result, resI32))
		return nil
	}
	if shiftOp, ok := map[string]string{"<<": "shl", ">>": "ashr"}[instruction.Operator]; ok {
		e.types[instruction.Result] = instruction.Type
		lI32 := instruction.Result + ".l_i32"
		rI32 := instruction.Result + ".r_i32"
		shift := instruction.Result + ".shift"
		resI32 := instruction.Result + ".res_i32"
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", lI32, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", rI32, instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  %%%s = and i32 %%%s, 31\n", shift, rI32))
		out.WriteString(fmt.Sprintf("  %%%s = %s i32 %%%s, %%%s\n", resI32, shiftOp, lI32, shift))
		out.WriteString(fmt.Sprintf("  %%%s = sitofp i32 %%%s to double\n", instruction.Result, resI32))
		return nil
	}
	if instruction.Operator == ">>>" {
		e.types[instruction.Result] = instruction.Type
		lU32 := instruction.Result + ".l_u32"
		rU32 := instruction.Result + ".r_u32"
		shift := instruction.Result + ".shift"
		resU32 := instruction.Result + ".res_u32"
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", lU32, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  %%%s = fptosi double %%%s to i32\n", rU32, instruction.Args[1]))
		out.WriteString(fmt.Sprintf("  %%%s = and i32 %%%s, 31\n", shift, rU32))
		out.WriteString(fmt.Sprintf("  %%%s = lshr i32 %%%s, %%%s\n", resU32, lU32, shift))
		out.WriteString(fmt.Sprintf("  %%%s = uitofp i32 %%%s to double\n", instruction.Result, resU32))
		return nil
	}
	return fmt.Errorf("unsupported LLVM binary operator %q", instruction.Operator)
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
	stepLabel := fmt.Sprintf("while.step.%d", e.labelCounter)
	endLabel := fmt.Sprintf("while.end.%d", e.labelCounter)
	e.labelCounter++

	targetCont := condLabel
	if len(instruction.Step) > 0 {
		targetCont = stepLabel
	}

	e.loopBreakLabels = append(e.loopBreakLabels, endLabel)
	e.loopContinueLabels = append(e.loopContinueLabels, targetCont)
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
		out.WriteString(fmt.Sprintf("  br label %%%s\n", targetCont))
	}

	if len(instruction.Step) > 0 {
		out.WriteString(fmt.Sprintf("%s:\n", stepLabel))
		e.terminated = false
		for _, inst := range instruction.Step {
			if err := e.emitInstruction(out, inst); err != nil {
				return err
			}
		}
		if !e.terminated {
			out.WriteString(fmt.Sprintf("  br label %%%s\n", condLabel))
		}
	}

	out.WriteString(fmt.Sprintf("%s:\n", endLabel))
	e.terminated = false
	return nil
}

func (e *functionEmitter) emitDoWhile(out *strings.Builder, instruction ir.Instruction) error {
	bodyLabel := fmt.Sprintf("dowhile.body.%d", e.labelCounter)
	condLabel := fmt.Sprintf("dowhile.cond.%d", e.labelCounter)
	stepLabel := fmt.Sprintf("dowhile.step.%d", e.labelCounter)
	endLabel := fmt.Sprintf("dowhile.end.%d", e.labelCounter)
	e.labelCounter++

	targetCont := condLabel
	if len(instruction.Step) > 0 {
		targetCont = stepLabel
	}

	e.loopBreakLabels = append(e.loopBreakLabels, endLabel)
	e.loopContinueLabels = append(e.loopContinueLabels, targetCont)
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
		out.WriteString(fmt.Sprintf("  br label %%%s\n", targetCont))
	}

	if len(instruction.Step) > 0 {
		out.WriteString(fmt.Sprintf("%s:\n", stepLabel))
		e.terminated = false
		for _, inst := range instruction.Step {
			if err := e.emitInstruction(out, inst); err != nil {
				return err
			}
		}
		if !e.terminated {
			out.WriteString(fmt.Sprintf("  br label %%%s\n", condLabel))
		}
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
	if strings.HasPrefix(instruction.Callee, "__async.") {
		if err := e.emitAsyncIntrinsic(out, instruction); err != nil {
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
	if strings.HasPrefix(instruction.Callee, "__os.") {
		if err := e.emitOsIntrinsic(out, instruction); err != nil {
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
	if strings.HasPrefix(instruction.Callee, "__json.") {
		if err := e.emitJsonIntrinsic(out, instruction); err != nil {
			return err
		}
		if instruction.Result != "" {
			e.types[instruction.Result] = instruction.Type
			if instruction.Type == ir.TypeString {
				e.ownedStrings = append(e.ownedStrings, instruction.Result)
			}
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__number.") {
		if err := emitNumberIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
			e.ownedStrings = append(e.ownedStrings, instruction.Result)
		}
		return nil
	}
	if strings.HasPrefix(instruction.Callee, "__string.") {
		if err := emitStringIntrinsic(out, instruction); err != nil {
			return err
		}
		e.types[instruction.Result] = instruction.Type
		if instruction.Type == ir.TypeString {
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
	returnValue := ""
	if len(instruction.Args) != 0 {
		returnValue = instruction.Args[0]
	}
	for _, arrayRef := range e.arrayTypes {
		if arrayRef.name != returnValue {
			out.WriteString(fmt.Sprintf("  call i32 @scriptgo_array_release(ptr %%%s)\n", arrayRef.name))
		}
	}
	for _, object := range e.objects {
		if object != returnValue {
			out.WriteString(fmt.Sprintf("  call i32 @scriptgo_object_release(ptr %%%s)\n", object))
		}
	}
	if e.function.Name == "main" {
		for _, value := range e.ownedStrings {
			if value != returnValue {
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_string_release(ptr %%%s)\n", value))
			}
		}
	}
	if e.function.Name == "main" {
		out.WriteString("  call i32 @scriptgo_event_loop_run()\n")
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

func (e *functionEmitter) emitClosure(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = ir.TypeClosure
	slot := instruction.Result + ".slot"
	out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))

	var envPtr string
	if len(instruction.Args) == 0 {
		envPtr = "null"
	} else {
		typesList := make([]string, len(instruction.Args))
		for i, arg := range instruction.Args {
			typ, ok := e.types[arg]
			if !ok {
				typ = ir.TypeNumber
			}
			typesList[i] = llvmType(typ)
		}
		structType := fmt.Sprintf("{ %s }", strings.Join(typesList, ", "))
		envAlloc := fmt.Sprintf("%s.env.%d", instruction.Result, e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", envAlloc, structType))
		for i, arg := range instruction.Args {
			typ, ok := e.types[arg]
			if !ok {
				typ = ir.TypeNumber
			}
			fieldPtr := fmt.Sprintf("%s.field.%d", envAlloc, i)
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds %s, ptr %%%s, i32 0, i32 %d\n", fieldPtr, structType, envAlloc, i))
			out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", llvmType(typ), arg, fieldPtr))
		}
		envPtr = "%" + envAlloc
	}

	status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
	e.runtimeStatus++
	out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_closure_create(ptr @%s, ptr %s, ptr %%%s)\n", status, instruction.Callee, envPtr, slot))
	out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
	return nil
}

func (e *functionEmitter) emitClosureCall(out *strings.Builder, instruction ir.Instruction) error {
	closureVar := instruction.Callee
	fnPtrSlot := fmt.Sprintf("%s.fn_ptr_slot.%d", instruction.Result, e.loadCounter)
	fnPtr := fmt.Sprintf("%s.fn_ptr.%d", instruction.Result, e.loadCounter)
	envSlot := fmt.Sprintf("%s.env_slot.%d", instruction.Result, e.loadCounter)
	envCtx := fmt.Sprintf("%s.env_ctx.%d", instruction.Result, e.loadCounter)
	e.loadCounter++

	out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds { ptr, ptr }, ptr %%%s, i32 0, i32 0\n", fnPtrSlot, closureVar))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", fnPtr, fnPtrSlot))
	out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds { ptr, ptr }, ptr %%%s, i32 0, i32 1\n", envSlot, closureVar))
	out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", envCtx, envSlot))

	var callArgs []string
	callArgs = append(callArgs, fmt.Sprintf("ptr %%%s", envCtx))

	for _, arg := range instruction.Args {
		typ, ok := e.types[arg]
		if !ok {
			typ = ir.TypeNumber
		}
		callArgs = append(callArgs, fmt.Sprintf("%s %%%s", llvmType(typ), arg))
	}

	retType := llvmType(instruction.Type)
	if instruction.Type != ir.TypeVoid && instruction.Result != "" {
		e.types[instruction.Result] = instruction.Type
		out.WriteString(fmt.Sprintf("  %%%s = call %s %%%s(%s)\n", instruction.Result, retType, fnPtr, strings.Join(callArgs, ", ")))
	} else {
		out.WriteString(fmt.Sprintf("  call void %%%s(%s)\n", fnPtr, strings.Join(callArgs, ", ")))
	}
	return nil
}

func (e *functionEmitter) emitAsyncIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__async.queueMicrotask":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("queueMicrotask requires 1 argument")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_queue_microtask(ptr %%%s, ptr null)\n", status, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		return nil
	case "__async.promise_resolve":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("Promise.resolve requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_create(ptr %%%s)\n", status, slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		pVal := fmt.Sprintf("%s.p", instruction.Result)
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", pVal, slot))
		status2 := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_resolve_number(ptr %%%s, double %%%s)\n", status2, pVal, instruction.Args[0]))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status2))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__async.promise_then":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("promise.then requires promise and callback")
		}
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_then(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot))
		return nil
	case "__async.await":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("await requires 1 argument")
		}
		slot := instruction.Result + ".slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca double\n", slot))
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		out.WriteString(fmt.Sprintf("  %%%s = call i32 @scriptgo_promise_await_number(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status))
		out.WriteString(fmt.Sprintf("  %%%s = load double, ptr %%%s\n", instruction.Result, slot))
		return nil
	default:
		return fmt.Errorf("unknown async intrinsic %q", instruction.Callee)
	}
}

func (e *functionEmitter) emitBoxUnknown(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = ir.TypeUnknown
	arg := instruction.Args[0]
	argType := e.types[arg]
	id := e.loadCounter
	e.loadCounter++

	var tag int
	var payloadVal string

	switch argType {
	case ir.TypeNumber:
		tag = 3 // SCRIPTGO_TAG_NUMBER
		payloadVal = fmt.Sprintf("payload.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = bitcast double %%%s to i64\n", payloadVal, arg))
	case ir.TypeBool:
		tag = 2 // SCRIPTGO_TAG_BOOLEAN
		payloadVal = fmt.Sprintf("payload.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = zext i1 %%%s to i64\n", payloadVal, arg))
	case ir.TypeString:
		tag = 4 // SCRIPTGO_TAG_STRING
		payloadVal = fmt.Sprintf("payload.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", payloadVal, arg))
	case ir.TypeVoid:
		tag = 0 // SCRIPTGO_TAG_UNDEFINED
		payloadVal = "0"
	case ir.TypeUnknown:
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", instruction.Result, arg))
		return nil
	default:
		if strings.HasSuffix(string(argType), "[]") || argType == ir.TypeNumberArray || argType == ir.TypeStringArray {
			tag = 6 // SCRIPTGO_TAG_ARRAY
		} else if argType == ir.TypeClosure {
			tag = 7 // SCRIPTGO_TAG_FUNCTION
		} else {
			tag = 5 // SCRIPTGO_TAG_OBJECT
		}
		payloadVal = fmt.Sprintf("payload.%d", id)
		out.WriteString(fmt.Sprintf("  %%%s = ptrtoint ptr %%%s to i64\n", payloadVal, arg))
	}

	b0 := fmt.Sprintf("box.b0.%d", id)
	b1 := fmt.Sprintf("box.b1.%d", id)
	out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } undef, i32 %d, 0\n", b0, tag))
	out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i32 0, 1\n", b1, b0))
	if payloadVal == "0" {
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 0, 2\n", instruction.Result, b1))
	} else {
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64 } %%%s, i64 %%%s, 2\n", instruction.Result, b1, payloadVal))
	}
	return nil
}

func (e *functionEmitter) emitCheckedCast(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = instruction.Type
	arg := instruction.Args[0]
	id := e.labelCounter
	e.labelCounter++

	var expectedTag int
	switch instruction.Type {
	case ir.TypeNumber:
		expectedTag = 3
	case ir.TypeBool:
		expectedTag = 2
	case ir.TypeString:
		expectedTag = 4
	case ir.TypeVoid:
		expectedTag = 0
	case ir.TypeClosure:
		expectedTag = 7
	case ir.TypeNumberArray, ir.TypeStringArray:
		expectedTag = 6
	default:
		if strings.HasSuffix(string(instruction.Type), "[]") {
			expectedTag = 6
		} else {
			expectedTag = 5
		}
	}

	tagVar := fmt.Sprintf("cast.tag.%d", id)
	cmpVar := fmt.Sprintf("cast.cmp.%d", id)
	castOk := fmt.Sprintf("cast_ok.%d", id)
	castFail := fmt.Sprintf("cast_fail.%d", id)

	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, arg))
	out.WriteString(fmt.Sprintf("  %%%s = icmp eq i32 %%%s, %d\n", cmpVar, tagVar, expectedTag))
	out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s\n", cmpVar, castOk, castFail))

	out.WriteString(fmt.Sprintf("\n%s:\n", castFail))
	out.WriteString(fmt.Sprintf("  call void @__scriptgo_fail_checked_cast(i32 %%%s, i32 %d, ptr null)\n", tagVar, expectedTag))
	out.WriteString("  unreachable\n")

	out.WriteString(fmt.Sprintf("\n%s:\n", castOk))
	rawPayload := fmt.Sprintf("cast.raw.%d", id)
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", rawPayload, arg))

	switch instruction.Type {
	case ir.TypeNumber:
		out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s to double\n", instruction.Result, rawPayload))
	case ir.TypeBool:
		out.WriteString(fmt.Sprintf("  %%%s = trunc i64 %%%s to i1\n", instruction.Result, rawPayload))
	case ir.TypeString:
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", instruction.Result, rawPayload))
	default:
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", instruction.Result, rawPayload))
	}
	return nil
}

func (e *functionEmitter) emitTypeOf(out *strings.Builder, instruction ir.Instruction) error {
	e.types[instruction.Result] = ir.TypeString
	arg := instruction.Args[0]
	id := e.loadCounter
	e.loadCounter++

	tagVar := fmt.Sprintf("typeof.tag.%d", id)
	out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 0\n", tagVar, arg))
	out.WriteString(fmt.Sprintf("  %%%s = call ptr @__scriptgo_typeof_unknown(i32 %%%s)\n", instruction.Result, tagVar))
	return nil
}

