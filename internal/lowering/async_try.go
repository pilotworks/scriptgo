package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerStructuredAsyncTry handles a try statement with one direct await. The
// fulfilled and rejected callbacks are the two asynchronous exits from the
// try body; both execute finally before continuing the outer sequence.
func lowerStructuredAsyncTry(path string, statement typescriptgo.SyntaxStatement, lowered ir.Function, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (ir.Function, bool, error) {
	tryIndex, tryInstruction, ok := findAsyncTry(lowered.Body)
	if !ok {
		return ir.Function{}, false, nil
	}
	bodySegments, awaits := splitLinearAsyncBody(tryInstruction.Body)
	if len(awaits) != 1 || hasStructuredAwait(tryInstruction.Catch) || hasStructuredAwait(tryInstruction.Finally) {
		return ir.Function{}, false, nil
	}
	await := awaits[0]

	asyncLowerCounter++
	base := fmt.Sprintf("__async_try_%d", asyncLowerCounter)
	promiseName := base + ".promise"
	frameName := base + ".frame"
	result := ir.Function{
		Name:       statement.Name,
		Span:       toIRSpan(path, statement.Span),
		Parameters: lowered.Parameters,
		Locals:     lowered.Locals,
		ReturnType: ir.Type("object:Promise"),
		Async:      true,
		EntryBlock: "async.entry",
	}
	result.Body = append(result.Body,
		ir.Instruction{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: promiseName, Callee: "__async.promise_create", Span: result.Span},
		ir.Instruction{Op: ir.OpCall, Type: ir.TypePointer, Result: frameName, Callee: "__async.frame_new", Value: fmt.Sprintf("%d", len(lowered.Parameters)+len(lowered.Captured)+len(lowered.Locals)+3), Span: result.Span},
	)
	result.AsyncFrame = buildAsyncFrame(result, lowered, []ir.Instruction{await})
	result.Blocks = buildAsyncBlocks([]ir.Instruction{await}, result.Span)
	result.Body = append(result.Body, rewriteAsyncReturns(lowered.Body[:tryIndex], promiseName, frameName, result.Span)...)
	result.Body = append(result.Body, bodySegments[0]...)

	valueTypes := asyncValueTypes(lowered)
	valueTypes[promiseName] = ir.Type("object:Promise")
	valueTypes[frameName] = ir.TypePointer
	captures := asyncTryCaptures(lowered, tryInstruction, bodySegments, promiseName, frameName, valueTypes, await.Result)
	fulfilledName := base + ".fulfilled"
	rejectedName := base + ".rejected"
	fulfilledClosure := base + ".fulfilled.closure"
	rejectedClosure := base + ".rejected.closure"

	tail := rewriteAsyncReturnsVoid(lowered.Body[tryIndex+1:], promiseName, frameName, result.Span)
	if len(tail) == 0 || !instructionListTerminates(tail) {
		tail = append(tail, settleAsyncVoid(promiseName, frameName, result.Span)...)
	}
	finally := tryInstruction.Finally
	catchFinally := finally
	if instructionListTerminates(tryInstruction.Catch) {
		// lowerTry has already emitted finally for a source-level return in the
		// catch branch; do not emit that cleanup a second time.
		catchFinally = nil
	}

	fulfilled := ir.Function{
		Name:       fulfilledName,
		Span:       await.Span,
		Parameters: asyncResumeParameters(),
		Captured:   asyncCapturedParameters(captures, valueTypes, promiseName),
		ReturnType: ir.TypeVoid,
	}
	fulfilled.Body = append(fulfilled.Body,
		ir.Instruction{Op: ir.OpCheckedCast, Type: await.Type, Result: await.Result, Args: []string{"__resume_raw"}, Span: await.Span},
	)
	fulfilled.Body = append(fulfilled.Body, rewriteAsyncTryPath(bodySegments[1], tryInstruction.CatchVar, tryInstruction.Catch, catchFinally, tail, promiseName, frameName, await.Span)...)
	if !instructionListTerminates(fulfilled.Body) {
		fulfilled.Body = append(fulfilled.Body, settleAsyncVoid(promiseName, frameName, await.Span)...)
	}

	rejected := ir.Function{
		Name:       rejectedName,
		Span:       await.Span,
		Parameters: asyncResumeParameters(),
		Captured:   asyncCapturedParameters(captures, valueTypes, promiseName),
		ReturnType: ir.TypeVoid,
	}
	if tryInstruction.CatchVar != "" {
		rejected.Body = append(rejected.Body, ir.Instruction{Op: ir.OpCheckedCast, Type: ir.Type("object:Error"), Result: tryInstruction.CatchVar, Args: []string{"__resume_raw"}, Span: await.Span})
	}
	rejected.Body = append(rejected.Body, rewriteAsyncTryPath(tryInstruction.Catch, "", nil, catchFinally, tail, promiseName, frameName, await.Span)...)
	if !instructionListTerminates(rejected.Body) {
		rejected.Body = append(rejected.Body, settleAsyncVoid(promiseName, frameName, await.Span)...)
	}

	result.Body = append(result.Body,
		ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: fulfilledClosure, Callee: fulfilledName, Args: captures, Span: await.Span},
		ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: rejectedClosure, Callee: rejectedName, Args: captures, Span: await.Span},
		ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_schedule_resume_pair", Args: []string{await.Args[0], fulfilledClosure, rejectedClosure}, Span: await.Span},
		ir.Instruction{Op: ir.OpReturn, Type: result.ReturnType, Args: []string{promiseName}, Span: result.Span},
	)
	extraFunctions = append(extraFunctions, fulfilled, rejected)
	return result, true, nil
}

// rewriteAsyncTryPath preserves try/finally ordering when a continuation path
// contains a terminal return or throw. A terminal instruction must not be
// emitted before finally; doing so changes JavaScript completion semantics.
func rewriteAsyncTryPath(body []ir.Instruction, catchVar string, catchBody, finally, tail []ir.Instruction, promiseName, frameName string, span ir.SourceSpan) []ir.Instruction {
	result := make([]ir.Instruction, 0, len(body)+len(catchBody)+len(finally)+len(tail)+4)
	for _, instruction := range body {
		switch instruction.Op {
		case ir.OpThrow:
			if len(catchBody) > 0 {
				if catchVar != "" && len(instruction.Args) > 0 {
					result = append(result, ir.Instruction{Op: ir.OpCheckedCast, Type: ir.Type("object:Error"), Result: catchVar, Args: []string{instruction.Args[0]}, Span: instruction.Span})
				}
				result = append(result, rewriteAsyncTryPath(catchBody, "", nil, finally, tail, promiseName, frameName, span)...)
			} else {
				result = append(result, finally...)
				if len(instruction.Args) > 0 {
					result = append(result, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_reject_existing", Args: []string{promiseName, instruction.Args[0]}, Span: instruction.Span})
				}
				result = append(result, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: instruction.Span}, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: instruction.Span})
			}
			return result
		case ir.OpReturn:
			result = append(result, finally...)
			value := "__async.undefined"
			if len(instruction.Args) > 0 {
				value = instruction.Args[0]
			} else {
				result = append(result, ir.Instruction{Op: ir.OpConst, Type: ir.TypeUnknown, Result: value, Value: "undefined", Span: instruction.Span})
			}
			result = append(result,
				ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_resolve_existing", Args: []string{promiseName, value}, Span: instruction.Span},
				ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: instruction.Span},
				ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: instruction.Span},
			)
			return result
		default:
			result = append(result, instruction)
		}
	}
	result = append(result, finally...)
	result = append(result, tail...)
	return result
}

func findAsyncTry(body []ir.Instruction) (int, ir.Instruction, bool) {
	found := -1
	var candidate ir.Instruction
	for i, instruction := range body {
		if !hasAwait(instruction) {
			continue
		}
		if instruction.Op != ir.OpTry || found != -1 {
			return -1, ir.Instruction{}, false
		}
		found, candidate = i, instruction
	}
	return found, candidate, found >= 0
}

func asyncTryCaptures(lowered ir.Function, tryInstruction ir.Instruction, segments [][]ir.Instruction, promiseName, frameName string, types map[string]ir.Type, awaitResult string) []string {
	used := asyncCaptures(append(append([]ir.Instruction{}, segments[0]...), tryInstruction.Finally...), types, lowered.Parameters)
	continuationScope := append(append([]ir.Instruction{}, segments[1]...), tryInstruction.Catch...)
	continuationScope = append(continuationScope, tryInstruction.Finally...)
	used = append(used, asyncCaptures(continuationScope, types, lowered.Parameters)...)
	used = removeCapture(used, awaitResult)
	seen := map[string]bool{}
	result := make([]string, 0, len(used)+len(lowered.Parameters)+len(lowered.Locals)+2)
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}
	for _, name := range used {
		add(name)
	}
	add(promiseName)
	add(frameName)
	for _, parameter := range lowered.Parameters {
		add(parameter.Name)
	}
	for _, local := range lowered.Locals {
		add(local.Name)
	}
	return result
}
