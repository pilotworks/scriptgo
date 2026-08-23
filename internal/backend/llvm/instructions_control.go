package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

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
	if instruction.Value != "" {
		if e.labeledBreak == nil {
			e.labeledBreak = make(map[string][]string)
			e.labeledContinue = make(map[string][]string)
		}
		lbl := instruction.Value
		e.labeledBreak[lbl] = append(e.labeledBreak[lbl], endLabel)
		e.labeledContinue[lbl] = append(e.labeledContinue[lbl], targetCont)
		defer func() {
			e.labeledBreak[lbl] = e.labeledBreak[lbl][:len(e.labeledBreak[lbl])-1]
			e.labeledContinue[lbl] = e.labeledContinue[lbl][:len(e.labeledContinue[lbl])-1]
		}()
	}

	out.WriteString(fmt.Sprintf("  br label %%%s\n", condLabel))
	out.WriteString(fmt.Sprintf("%s:\n", condLabel))

	e.terminated = false
	for _, inst := range instruction.Cond {
		if err := e.emitInstruction(out, inst); err != nil {
			return err
		}
	}
	if len(instruction.Args) > 0 {
		condVal := e.resolveArg(out, instruction.Args[0])
		out.WriteString(fmt.Sprintf("  br i1 %%%s, label %%%s, label %%%s\n", condVal, bodyLabel, endLabel))
	} else {
		out.WriteString(fmt.Sprintf("  br label %%%s\n", bodyLabel))
	}

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
	if instruction.Value != "" {
		if e.labeledBreak == nil {
			e.labeledBreak = make(map[string][]string)
			e.labeledContinue = make(map[string][]string)
		}
		lbl := instruction.Value
		e.labeledBreak[lbl] = append(e.labeledBreak[lbl], endLabel)
		e.labeledContinue[lbl] = append(e.labeledContinue[lbl], targetCont)
		defer func() {
			e.labeledBreak[lbl] = e.labeledBreak[lbl][:len(e.labeledBreak[lbl])-1]
			e.labeledContinue[lbl] = e.labeledContinue[lbl][:len(e.labeledContinue[lbl])-1]
		}()
	}

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

func (e *functionEmitter) emitReturn(out *strings.Builder, instruction ir.Instruction) error {
	if e.function.Name == "main" {
		out.WriteString("  call i32 @scriptgo_timers_drain()\n")
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
	case ir.TypeUnknown:
		payloadVal := fmt.Sprintf("throw.payload.%d", e.loadCounter)
		ptrVal := fmt.Sprintf("throw.ptr.%d", e.loadCounter)
		e.loadCounter++
		out.WriteString(fmt.Sprintf("  %%%s = extractvalue { i32, i32, i64 } %%%s, 2\n", payloadVal, argVal))
		out.WriteString(fmt.Sprintf("  %%%s = inttoptr i64 %%%s to ptr\n", ptrVal, payloadVal))
		out.WriteString(fmt.Sprintf("  call void @scriptgo_throw_string(ptr %%%s)\n", ptrVal))
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
