package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerStructuredAsyncBranch splits a top-level if whose selected branch has
// one await. The condition and the non-suspending prefix still execute in the
// caller's turn; only the selected continuation is queued.
func lowerStructuredAsyncBranch(path string, statement typescriptgo.SyntaxStatement, lowered ir.Function, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (ir.Function, bool, error) {
	index := -1
	var branch ir.Instruction
	for i, instruction := range lowered.Body {
		if !hasAwait(instruction) {
			continue
		}
		if index != -1 || instruction.Op != ir.OpIf || hasNestedAwait(instruction.Cond) {
			return ir.Function{}, false, nil
		}
		index, branch = i, instruction
	}
	if index < 0 || len(branch.Args) != 1 || len(branch.Else) > 0 && hasNestedAwait(branch.Else) {
		if index >= 0 && len(branch.Args) == 1 && hasNestedAwait(branch.Else) {
			if structured, ok, err := lowerStructuredAsyncBranchChain(path, statement, lowered, shapes, signatures, index, branch); err != nil {
				return ir.Function{}, false, err
			} else if ok {
				return structured, true, nil
			}
		}
		return ir.Function{}, false, nil
	}

	thenSegments, thenAwaits := splitLinearAsyncBody(branch.Then)
	elseSegments, elseAwaits := splitLinearAsyncBody(branch.Else)
	if len(thenAwaits) == 1 && len(elseAwaits) != 0 || len(elseAwaits) == 1 && len(thenAwaits) != 0 {
		return ir.Function{}, false, nil
	}
	awaitInThen := len(thenAwaits) == 1
	if !awaitInThen && len(elseAwaits) != 1 {
		return ir.Function{}, false, nil
	}
	await := thenAwaits[0]
	if !awaitInThen {
		await = elseAwaits[0]
	}

	asyncLowerCounter++
	base := fmt.Sprintf("__async_branch_%d", asyncLowerCounter)
	promiseName, frameName := base+".promise", base+".frame"
	result := ir.Function{Name: statement.Name, Span: toIRSpan(path, statement.Span), Parameters: lowered.Parameters, Locals: lowered.Locals, ReturnType: ir.Type("object:Promise"), Async: true, EntryBlock: "async.entry"}
	result.Body = append(result.Body,
		ir.Instruction{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: promiseName, Callee: "__async.promise_create", Span: result.Span},
		ir.Instruction{Op: ir.OpCall, Type: ir.TypePointer, Result: frameName, Callee: "__async.frame_new", Value: fmt.Sprintf("%d", len(lowered.Parameters)+len(lowered.Captured)+len(lowered.Locals)+3), Span: result.Span},
	)
	result.AsyncFrame = buildAsyncFrame(result, lowered, []ir.Instruction{await})
	result.Blocks = buildAsyncBlocks([]ir.Instruction{await}, result.Span)
	result.Body = append(result.Body, rewriteAsyncReturns(lowered.Body[:index], promiseName, frameName, result.Span)...)

	valueTypes := asyncValueTypes(lowered)
	valueTypes[promiseName], valueTypes[frameName] = ir.Type("object:Promise"), ir.TypePointer
	scope := append([]ir.Instruction{}, branch.Cond...)
	scope = append(scope, branch.Then...)
	scope = append(scope, branch.Else...)
	scope = append(scope, lowered.Body[index+1:]...)
	captures := asyncCaptures(scope, valueTypes, lowered.Parameters)
	captures = removeCapture(captures, await.Result)
	seen := map[string]bool{}
	add := func(name string) {
		if name != "" && !seen[name] {
			captures = append(captures, name)
			seen[name] = true
		}
	}
	for _, name := range captures {
		seen[name] = true
	}
	add(promiseName)
	add(frameName)
	for _, parameter := range lowered.Parameters {
		add(parameter.Name)
	}
	for _, local := range lowered.Locals {
		add(local.Name)
	}

	tail := rewriteAsyncReturnsVoid(lowered.Body[index+1:], promiseName, frameName, result.Span)
	if len(tail) == 0 || !instructionListTerminates(tail) {
		tail = append(tail, settleAsyncVoid(promiseName, frameName, result.Span)...)
	}
	other := branch.Else
	if !awaitInThen {
		other = branch.Then
	}
	other = append(append([]ir.Instruction{}, other...), tail...)
	if len(other) == 0 || !instructionListTerminates(other) {
		other = append(other, settleAsyncVoid(promiseName, frameName, result.Span)...)
	}

	fulfilledName, rejectedName := base+".fulfilled", base+".rejected"
	fulfilledClosure, rejectedClosure := base+".fulfilled.closure", base+".rejected.closure"
	schedule := []ir.Instruction{
		{Op: ir.OpClosure, Type: ir.TypeClosure, Result: fulfilledClosure, Callee: fulfilledName, Args: captures, Span: await.Span},
		{Op: ir.OpClosure, Type: ir.TypeClosure, Result: rejectedClosure, Callee: rejectedName, Args: captures, Span: await.Span},
		{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_schedule_resume_pair", Args: []string{await.Args[0], fulfilledClosure, rejectedClosure}, Span: await.Span},
	}
	selected := thenSegments[0]
	if !awaitInThen {
		selected = elseSegments[0]
	}
	unselected := other
	branchBody := append(append([]ir.Instruction{}, selected...), schedule...)
	if awaitInThen {
		result.Body = append(result.Body, ir.Instruction{Op: ir.OpIf, Type: ir.TypeVoid, Args: branch.Args, Cond: branch.Cond, Then: branchBody, Else: stripTerminalReturn(unselected), Span: branch.Span})
	} else {
		result.Body = append(result.Body, ir.Instruction{Op: ir.OpIf, Type: ir.TypeVoid, Args: branch.Args, Cond: branch.Cond, Then: stripTerminalReturn(unselected), Else: branchBody, Span: branch.Span})
	}
	result.Body = append(result.Body, ir.Instruction{Op: ir.OpReturn, Type: result.ReturnType, Args: []string{promiseName}, Span: result.Span})

	fulfilled := ir.Function{Name: fulfilledName, Span: await.Span, Parameters: asyncResumeParameters(), Captured: asyncCapturedParameters(captures, valueTypes, promiseName), ReturnType: ir.TypeVoid}
	fulfilled.Body = append(fulfilled.Body, ir.Instruction{Op: ir.OpCheckedCast, Type: await.Type, Result: await.Result, Args: []string{"__resume_raw"}, Span: await.Span})
	if awaitInThen {
		fulfilled.Body = append(fulfilled.Body, thenSegments[1]...)
	} else {
		fulfilled.Body = append(fulfilled.Body, elseSegments[1]...)
	}
	fulfilled.Body = rewriteAsyncReturnsVoid(fulfilled.Body, promiseName, frameName, result.Span)
	if !instructionListTerminates(fulfilled.Body) {
		fulfilled.Body = append(fulfilled.Body, tail...)
	}
	rejected := ir.Function{Name: rejectedName, Span: await.Span, Parameters: asyncResumeParameters(), Captured: asyncCapturedParameters(captures, valueTypes, promiseName), ReturnType: ir.TypeVoid, Body: append([]ir.Instruction{}, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_reject_existing", Args: []string{promiseName, "__resume_raw"}, Span: await.Span}, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: await.Span}, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: await.Span})}
	extraFunctions = append(extraFunctions, fulfilled, rejected)
	return result, true, nil
}

func stripTerminalReturn(instructions []ir.Instruction) []ir.Instruction {
	result := append([]ir.Instruction{}, instructions...)
	if len(result) > 0 && result[len(result)-1].Op == ir.OpReturn {
		result = result[:len(result)-1]
	}
	return result
}
