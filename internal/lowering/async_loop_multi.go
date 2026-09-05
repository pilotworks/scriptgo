package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerStructuredAsyncLoopLinearMulti lowers a loop whose body contains a
// linear sequence of direct awaits. Each fulfilled continuation runs the next
// segment synchronously and registers only the next await.
func lowerStructuredAsyncLoopLinearMulti(path string, statement typescriptgo.SyntaxStatement, lowered ir.Function, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, loopIndex int, loop ir.Instruction, segments [][]ir.Instruction, awaits []ir.Instruction) (ir.Function, bool, error) {
	if len(awaits) < 2 || len(segments) != len(awaits)+1 {
		return ir.Function{}, false, nil
	}
	tailSegments, tailAwaits := splitLinearAsyncBody(lowered.Body[loopIndex+1:])
	if len(tailAwaits) > 1 {
		return ir.Function{}, false, nil
	}
	asyncLowerCounter++
	base := fmt.Sprintf("__async_loop_multi_%d", asyncLowerCounter)
	promiseName, frameName := base+".promise", base+".frame"
	allAwaits := append(append([]ir.Instruction{}, awaits...), tailAwaits...)
	result := ir.Function{Name: statement.Name, Span: toIRSpan(path, statement.Span), Parameters: lowered.Parameters, Locals: lowered.Locals, ReturnType: ir.Type("object:Promise"), Async: true, EntryBlock: "async.entry"}
	result.Body = append(result.Body,
		ir.Instruction{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: promiseName, Callee: "__async.promise_create", Span: result.Span},
		ir.Instruction{Op: ir.OpCall, Type: ir.TypePointer, Result: frameName, Callee: "__async.frame_new", Value: fmt.Sprintf("%d", len(lowered.Parameters)+len(lowered.Captured)+len(lowered.Locals)+len(allAwaits)+2), Span: result.Span},
	)
	result.AsyncFrame = buildAsyncFrame(result, lowered, allAwaits)
	result.Blocks = buildAsyncBlocks(allAwaits, result.Span)
	result.Body = append(result.Body, rewriteAsyncReturns(lowered.Body[:loopIndex], promiseName, frameName, result.Span)...)

	valueTypes := asyncValueTypes(lowered)
	valueTypes[promiseName], valueTypes[frameName] = ir.Type("object:Promise"), ir.TypePointer
	var scope []ir.Instruction
	scope = append(scope, loop.Cond...)
	for _, segment := range segments {
		scope = append(scope, segment...)
	}
	scope = append(scope, awaits...)
	for _, segment := range tailSegments {
		scope = append(scope, segment...)
	}
	scope = append(scope, tailAwaits...)
	captures := asyncCaptures(scope, valueTypes, lowered.Parameters)
	seen := map[string]bool{}
	for _, name := range captures {
		seen[name] = true
	}
	add := func(name string) {
		if name != "" && !seen[name] {
			captures = append(captures, name)
			seen[name] = true
		}
	}
	add(promiseName)
	add(frameName)
	for _, parameter := range lowered.Parameters {
		add(parameter.Name)
	}
	for _, local := range lowered.Locals {
		add(local.Name)
	}

	tailBody := tailSegments[0]
	if len(tailAwaits) == 1 {
		tailBody = tailSegments[1]
	}
	tail := rewriteAsyncReturnsVoid(tailBody, promiseName, frameName, result.Span)
	if len(tail) == 0 || !instructionListTerminates(tail) {
		tail = append(tail, settleAsyncVoid(promiseName, frameName, result.Span)...)
	}
	runnerName := base + ".runner"
	runnerClosure := base + ".runner.closure"
	runner := ir.Function{Name: runnerName, Span: loop.Span, Parameters: asyncRunnerParameters(), Captured: asyncCapturedParameters(captures, valueTypes, promiseName), ReturnType: ir.TypeVoid}
	runner.Body = append(runner.Body, loop.Cond...)
	exitPath := append([]ir.Instruction{}, tailSegments[0]...)
	if len(tailAwaits) == 1 {
		exitPath = append(exitPath, multiAwaitSchedule(base, len(awaits), tailAwaits[0], captures)...)
	} else {
		exitPath = append(exitPath, tail...)
	}
	firstSegment := rewriteAsyncReturnsVoid(segments[0], promiseName, frameName, result.Span)
	firstSegment = rewriteLoopControl(firstSegment, exitPath, runnerName, loop.Span)
	runner.Body = append(runner.Body, ir.Instruction{Op: ir.OpIf, Type: ir.TypeVoid, Args: loop.Args, Then: append(firstSegment, multiAwaitSchedule(base, 0, awaits[0], captures)...), Else: exitPath, Span: loop.Span})

	continuationFunctions := []ir.Function{runner}
	for i, await := range awaits {
		successName, rejectName := fmt.Sprintf("%s.fulfilled.%d", base, i), fmt.Sprintf("%s.rejected.%d", base, i)
		success := ir.Function{Name: successName, Span: await.Span, Parameters: asyncResumeParameters(), Captured: asyncCapturedParameters(captures, valueTypes, promiseName), ReturnType: ir.TypeVoid}
		success.Body = append(success.Body, ir.Instruction{Op: ir.OpCheckedCast, Type: await.Type, Result: await.Result, Args: []string{"__resume_raw"}, Span: await.Span})
		nextSegment := rewriteAsyncReturnsVoid(segments[i+1], promiseName, frameName, result.Span)
		// A break exits the loop and must execute the post-loop path (including
		// any await after the loop), not the already-settled final tail.
		nextSegment = rewriteLoopControl(nextSegment, exitPath, runnerName, await.Span)
		success.Body = append(success.Body, nextSegment...)
		if instructionListTerminates(nextSegment) {
			// A break/return path already settled the outer promise.
		} else if i+1 < len(awaits) {
			success.Body = append(success.Body, multiAwaitSchedule(base, i+1, awaits[i+1], captures)...)
		} else {
			success.Body = append(success.Body, loop.Step...)
			success.Body = append(success.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: runnerName, Args: asyncRunnerCallArgs(), Span: await.Span}, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: await.Span})
		}
		reject := ir.Function{Name: rejectName, Span: await.Span, Parameters: asyncResumeParameters(), Captured: asyncCapturedParameters(captures, valueTypes, promiseName), ReturnType: ir.TypeVoid, Body: append([]ir.Instruction{}, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_reject_existing", Args: []string{promiseName, "__resume_raw"}, Span: await.Span}, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: await.Span}, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: await.Span})}
		continuationFunctions = append(continuationFunctions, success, reject)
	}
	if len(tailAwaits) == 1 {
		await := tailAwaits[0]
		index := len(awaits)
		success := ir.Function{Name: fmt.Sprintf("%s.fulfilled.%d", base, index), Span: await.Span, Parameters: asyncResumeParameters(), Captured: asyncCapturedParameters(captures, valueTypes, promiseName), ReturnType: ir.TypeVoid}
		success.Body = append(success.Body, ir.Instruction{Op: ir.OpCheckedCast, Type: await.Type, Result: await.Result, Args: []string{"__resume_raw"}, Span: await.Span})
		success.Body = append(success.Body, tail...)
		reject := ir.Function{Name: fmt.Sprintf("%s.rejected.%d", base, index), Span: await.Span, Parameters: asyncResumeParameters(), Captured: asyncCapturedParameters(captures, valueTypes, promiseName), ReturnType: ir.TypeVoid, Body: append([]ir.Instruction{}, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_reject_existing", Args: []string{promiseName, "__resume_raw"}, Span: await.Span}, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: await.Span}, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: await.Span})}
		continuationFunctions = append(continuationFunctions, success, reject)
	}
	result.Body = append(result.Body, ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: runnerClosure, Callee: runnerName, Args: captures, Span: loop.Span}, ir.Instruction{Op: ir.OpClosureCall, Type: ir.TypeVoid, Callee: runnerClosure, Span: loop.Span}, ir.Instruction{Op: ir.OpReturn, Type: result.ReturnType, Args: []string{promiseName}, Span: result.Span})
	extraFunctions = append(extraFunctions, continuationFunctions...)
	return result, true, nil
}

func multiAwaitSchedule(base string, index int, await ir.Instruction, captures []string) []ir.Instruction {
	fulfilled := fmt.Sprintf("%s.fulfilled.%d.closure", base, index)
	rejected := fmt.Sprintf("%s.rejected.%d.closure", base, index)
	return []ir.Instruction{{Op: ir.OpClosure, Type: ir.TypeClosure, Result: fulfilled, Callee: fmt.Sprintf("%s.fulfilled.%d", base, index), Args: captures, Span: await.Span}, {Op: ir.OpClosure, Type: ir.TypeClosure, Result: rejected, Callee: fmt.Sprintf("%s.rejected.%d", base, index), Args: captures, Span: await.Span}, {Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_schedule_resume_pair", Args: []string{await.Args[0], fulfilled, rejected}, Span: await.Span}, {Op: ir.OpReturn, Type: ir.TypeVoid, Span: await.Span}}
}
