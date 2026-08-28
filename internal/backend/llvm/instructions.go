package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

type arrayReference struct {
	name string
	typ  ir.Type
}

type functionEmitter struct {
	function        ir.Function
	functions       map[string]ir.Function
	stringsByValue  map[string]string
	debug           *debugInfo
	module          ir.Module
	compilerVersion string

	types              map[string]ir.Type
	varSlots           map[string]string
	loadCounter        int
	arrayTypes         []arrayReference
	objects            []string
	ownedStrings       []string
	labelCounter       int
	loopBreakLabels    []string
	loopContinueLabels []string
	labeledBreak       map[string][]string
	labeledContinue    map[string][]string
	sharedEnvCells     map[string]string
	runtimeStatus      int
	tempCounter        int
	terminated         bool
	localSSAs          map[string]bool
}

func (e *functionEmitter) dbg(span ir.SourceSpan) string {
	if e.debug == nil {
		return ""
	}
	loc := e.debug.location(span, e.function.Name, e.module)
	if loc == "" {
		return ""
	}
	return ", " + loc
}

func (e *functionEmitter) isParamUnknown(arg string) bool {
	for _, p := range e.function.Parameters {
		if p.Name == arg && p.Type == ir.TypeUnknown {
			return true
		}
	}
	return false
}

func (e *functionEmitter) ensurePointerArg(out *strings.Builder, arg string) string {
	if e.types[arg] == ir.TypeUnknown || e.isParamUnknown(arg) {
		if slot, ok := e.varSlots[arg]; ok {
			loaded := fmt.Sprintf("%s.ptr_load.%d", arg, e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = load volatile { i32, i32, i64 }, ptr %%%s\n", loaded, slot))
			arg = loaded
		}
		payloadVar := fmt.Sprintf("%s.payload.%d", arg, e.loadCounter)
		ptrVar := fmt.Sprintf("%s.ptr.%d", arg, e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg))
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
		return ptrVar
	}
	return e.resolveArg(out, arg)
}

func (e *functionEmitter) resolveArg(out *strings.Builder, arg string) string {
	slot, ok := e.varSlots[arg]
	if cellSlot, isCell := e.sharedEnvCells[arg]; isCell {
		slot = cellSlot
		ok = true
	}
	if ok {
		typ := e.types[arg]
		if e.isParamUnknown(arg) {
			typ = ir.TypeUnknown
		}
		if typ == ir.TypeVoid || typ == "" {
			return arg
		}
		lt := llvmType(typ)
		if lt == "void" || lt == "" {
			return arg
		}
		loadName := fmt.Sprintf("%s.load.%d", arg, e.loadCounter)
		e.loadCounter++
		e.types[loadName] = typ
		out.WriteString(fmt.Sprintf("  %%%s = load volatile %s, ptr %%%s\n", loadName, lt, slot))
		return loadName
	}
	if e.localSSAs == nil || !e.localSSAs[arg] {
		for _, g := range e.module.Globals {
			if g.Name == arg {
				loadName := fmt.Sprintf("%s.gload.%d", arg, e.loadCounter)
				e.loadCounter++
				typ := e.types[arg]
				if typ == "" || typ == ir.TypeVoid {
					typ = g.Type
					if typ == "" || typ == ir.TypeVoid {
						typ = ir.TypePointer
					}
				}
				e.types[loadName] = typ
				if strings.HasSuffix(string(typ), "[]") || typ == ir.TypeStringArray || typ == ir.TypeNumberArray {
					e.arrayTypes = append(e.arrayTypes, arrayReference{name: loadName, typ: typ})
				}
				out.WriteString(fmt.Sprintf("  %%%s = load volatile %s, ptr @%s\n", loadName, llvmType(typ), g.Name))
				return loadName
			}
		}
	}
	return arg
}

func (e *functionEmitter) emitInstruction(out *strings.Builder, instruction ir.Instruction) error {
	if instruction.Result != "" && instruction.Type != "" {
		e.types[instruction.Result] = instruction.Type
	}
	switch instruction.Op {
	case ir.OpWhile:
		return e.emitWhile(out, instruction)
	case ir.OpDoWhile:
		return e.emitDoWhile(out, instruction)
	case ir.OpClosure:
		return e.emitClosure(out, instruction)
	}

	resolvedArgs := make([]string, len(instruction.Args))
	for i, arg := range instruction.Args {
		resolvedArgs[i] = e.resolveArg(out, arg)
	}
	inst := instruction
	inst.Args = resolvedArgs

	targetResult := inst.Result
	isGlobalResult := false
	var globalResultType ir.Type
	if _, ok := e.varSlots[inst.Result]; ok {
		inst.Result = fmt.Sprintf("%s.val.%d", inst.Result, e.loadCounter)
		e.loadCounter++
	} else {
		for _, g := range e.module.Globals {
			if g.Name == inst.Result {
				isGlobalResult = true
				globalResultType = g.Type
				break
			}
		}
		if isGlobalResult {
			inst.Result = fmt.Sprintf("%s.gres.%d", inst.Result, e.loadCounter)
			e.loadCounter++
		}
	}

	switch inst.Op {
	case ir.OpConst:
		if err := e.emitConst(out, inst); err != nil {
			return err
		}
	case ir.OpAssign:
		typ := inst.Type
		if typ == "" {
			typ = e.types[targetResult]
		}
		isGlobal := false
		if _, ok := e.varSlots[targetResult]; !ok {
			for _, g := range e.module.Globals {
				if g.Name == targetResult {
					if typ == "" {
						typ = g.Type
					}
					isGlobal = true
					break
				}
			}
		}
		arg := inst.Args[0]
		argType := e.types[arg]
		argVal := "%" + arg
		if typ != ir.TypeUnknown && argType == ir.TypeUnknown {
			payloadVar := fmt.Sprintf("payload.%d", e.loadCounter)
			e.loadCounter++
			out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVar, arg))
			switch typ {
			case ir.TypeNumber:
				numVar := fmt.Sprintf("num.%d", e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = bitcast i64 %%%s to double\n", numVar, payloadVar))
				argVal = "%" + numVar
			case ir.TypeBool:
				boolVar := fmt.Sprintf("bool.%d", e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = trunc i64 %%%s to i1\n", boolVar, payloadVar))
				argVal = "%" + boolVar
			default:
				ptrVar := fmt.Sprintf("ptr.%d", e.loadCounter)
				e.loadCounter++
				out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVar, payloadVar))
				argVal = "%" + ptrVar
			}
		}
		if isGlobal {
			out.WriteString(fmt.Sprintf("  store %s %s, ptr @%s\n", llvmType(typ), argVal, targetResult))
			if cellSlot, isCell := e.sharedEnvCells[targetResult]; isCell {
				out.WriteString(fmt.Sprintf("  store volatile %s %s, ptr %%%s\n", llvmType(typ), argVal, cellSlot))
			}
		} else {
			slot := e.varSlots[targetResult]
			if cellSlot, isCell := e.sharedEnvCells[targetResult]; isCell {
				slot = cellSlot
			}
			out.WriteString(fmt.Sprintf("  store volatile %s %s, ptr %%%s\n", llvmType(typ), argVal, slot))
		}
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
		breakLabel := ""
		if inst.Value != "" {
			if stack, ok := e.labeledBreak[inst.Value]; ok && len(stack) > 0 {
				breakLabel = stack[len(stack)-1]
			} else {
				return fmt.Errorf("break to unknown label %q", inst.Value)
			}
		} else {
			if len(e.loopBreakLabels) == 0 {
				return fmt.Errorf("break outside of loop")
			}
			breakLabel = e.loopBreakLabels[len(e.loopBreakLabels)-1]
		}
		out.WriteString(fmt.Sprintf("  br label %%%s\n", breakLabel))
		e.terminated = true
		return nil
	case ir.OpContinue:
		continueLabel := ""
		if inst.Value != "" {
			if stack, ok := e.labeledContinue[inst.Value]; ok && len(stack) > 0 {
				continueLabel = stack[len(stack)-1]
			} else {
				return fmt.Errorf("continue to unknown label %q", inst.Value)
			}
		} else {
			if len(e.loopContinueLabels) == 0 {
				return fmt.Errorf("continue outside of loop")
			}
			continueLabel = e.loopContinueLabels[len(e.loopContinueLabels)-1]
		}
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
	case ir.OpDebugger:
		pathStr := inst.Span.Path
		if pathStr == "" {
			pathStr = e.module.SourcePath
		}
		global, ok := e.stringsByValue[pathStr]
		if !ok {
			global = fmt.Sprintf("@.str.%d", len(e.stringsByValue))
			e.stringsByValue[pathStr] = global
		}
		pathPtr := fmt.Sprintf("debugger.path.%d", e.loadCounter)
		e.loadCounter++
		length := len([]byte(pathStr)) + 1
		line := sourceLine(e.module.SourceFiles[pathStr], inst.Span.Offset)
		out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds [%d x i8], ptr %s, i64 0, i64 0\n", pathPtr, length, global))
		dbgLoc := e.dbg(inst.Span)
		out.WriteString(fmt.Sprintf("  call void @scriptgo_debugger_break(ptr %%%s, i32 %d)%s\n", pathPtr, line, dbgLoc))
		return nil
	default:
		return fmt.Errorf("unsupported LLVM instruction %q", inst.Op)
	}

	slot, hasSlot := e.varSlots[targetResult]
	if cellSlot, isCell := e.sharedEnvCells[targetResult]; isCell {
		slot = cellSlot
		hasSlot = true
	}
	if hasSlot {
		typ := e.types[targetResult]
		if typ == "" {
			typ = inst.Type
			if typ == "" {
				typ = e.types[inst.Result]
			}
			e.types[targetResult] = typ
		}
		if typ != ir.TypeVoid {
			lt := llvmType(typ)
			if lt == "" {
				lt = "ptr"
			}
			if lt != "void" {
				out.WriteString(fmt.Sprintf("  store volatile %s %%%s, ptr %%%s\n", lt, inst.Result, slot))
			}
		}
	}
	if isGlobalResult && inst.Type != ir.TypeVoid {
		typ := e.types[inst.Result]
		if typ == "" {
			typ = inst.Type
			if typ == "" {
				typ = globalResultType
			}
		}
		if typ != ir.TypeVoid {
			e.types[targetResult] = typ
			if strings.HasSuffix(string(typ), "[]") || typ == ir.TypeStringArray || typ == ir.TypeNumberArray {
				e.arrayTypes = append(e.arrayTypes, arrayReference{name: targetResult, typ: typ})
			}
			lt := llvmType(typ)
			if lt == "" {
				lt = "ptr"
			}
			if lt != "void" {
				out.WriteString(fmt.Sprintf("  store volatile %s %%%s, ptr @%s\n", lt, inst.Result, targetResult))
			}
		}
	}
	return nil
}
