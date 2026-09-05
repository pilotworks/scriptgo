package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerStructuredAsyncLoop lowers a loop whose body contains one await. The
// loop itself becomes a resumable runner: each iteration executes synchronously
// until the await, then the continuation re-enters the runner on the next turn.
// This keeps loop state in closure cells and never blocks the event loop.
func lowerStructuredAsyncLoop(path string, statement typescriptgo.SyntaxStatement, lowered ir.Function, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (ir.Function, bool, error) {
	loopIndex, loop, ok := findAsyncLoop(lowered.Body)
	if !ok {
		return ir.Function{}, false, nil
	}
	if len(loop.Cond) == 0 || loop.Args == nil || len(loop.Args) != 1 || len(loop.Step) > 0 && hasNestedAwait(loop.Step) {
		return ir.Function{}, false, nil
	}
	bodySegments, awaits := splitLinearAsyncBody(loop.Body)
	if len(bodySegments) == 1 && len(loop.Body) == 1 && loop.Body[0].Op == ir.OpTry {
		return lowerStructuredAsyncLoopTry(path, statement, lowered, shapes, signatures, loopIndex, loop, loop.Body[0])
	}
	if len(awaits) != 1 {
		if len(awaits) > 1 {
			return lowerStructuredAsyncLoopLinearMulti(path, statement, lowered, shapes, signatures, loopIndex, loop, bodySegments, awaits)
		}
		return ir.Function{}, false, nil
	}

	asyncLowerCounter++
	base := fmt.Sprintf("__async_structured_%d", asyncLowerCounter)
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
		ir.Instruction{Op: ir.OpCall, Type: ir.TypePointer, Result: frameName, Callee: "__async.frame_new", Value: fmt.Sprintf("%d", len(lowered.Parameters)+len(lowered.Captured)+len(lowered.Locals)+len(awaits)+2), Span: result.Span},
	)

	awaits = append(awaits, collectNestedAwaits(loop.Cond)...)
	result.AsyncFrame = buildAsyncFrame(result, lowered, awaits)
	result.Blocks = buildAsyncBlocks(awaits, result.Span)
	result.Body = append(result.Body, rewriteAsyncReturns(lowered.Body[:loopIndex], promiseName, frameName, result.Span)...)

	valueTypes := asyncValueTypes(lowered)
	valueTypes[promiseName] = ir.Type("object:Promise")
	valueTypes[frameName] = ir.TypePointer
	runnerCaptures, continuationCaptures := asyncLoopCaptures(lowered, loop, bodySegments, promiseName, frameName, valueTypes)
	runnerName := base + ".runner"
	runnerClosure := base + ".runner.closure"

	tail := rewriteAsyncReturnsVoid(lowered.Body[loopIndex+1:], promiseName, frameName, result.Span)
	if len(tail) == 0 || !instructionListTerminates(tail) {
		tail = append(tail, settleAsyncVoid(promiseName, frameName, result.Span)...)
	}
	loopPrefix := rewriteAsyncReturnsVoid(bodySegments[0], promiseName, frameName, result.Span)
	loopPrefix = rewriteLoopControl(loopPrefix, tail, runnerName, result.Span)

	runner := ir.Function{
		Name:       runnerName,
		Span:       loop.Span,
		Parameters: asyncRunnerParameters(),
		Captured:   asyncCapturedParameters(runnerCaptures, valueTypes, promiseName),
		ReturnType: ir.TypeVoid,
	}
	runner.Body = append(runner.Body, loop.Cond...)
	await := awaits[0]
	successName := base + ".fulfilled"
	rejectName := base + ".rejected"
	resumeFulfilled := base + ".fulfilled.closure"
	resumeRejected := base + ".rejected.closure"
	resumePath := append(loopPrefix,
		ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: resumeFulfilled, Callee: successName, Args: continuationCaptures, Span: await.Span},
		ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: resumeRejected, Callee: rejectName, Args: continuationCaptures, Span: await.Span},
		ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_schedule_resume_pair", Args: []string{await.Args[0], resumeFulfilled, resumeRejected}, Span: await.Span},
		ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: await.Span},
	)
	runner.Body = append(runner.Body, ir.Instruction{
		Op: ir.OpIf, Type: ir.TypeVoid, Args: []string{loop.Args[0]},
		Then: resumePath, Else: tail, Span: loop.Span,
	})

	success := ir.Function{
		Name:       successName,
		Span:       await.Span,
		Parameters: asyncResumeParameters(),
		Captured:   asyncCapturedParameters(continuationCaptures, valueTypes, promiseName),
		ReturnType: ir.TypeVoid,
	}
	success.Body = append(success.Body, ir.Instruction{Op: ir.OpCheckedCast, Type: await.Type, Result: await.Result, Args: []string{"__resume_raw"}, Span: await.Span})
	// The resumed segment is no longer emitted inside a loop. Rewrite loop
	// control before emitting it so break/continue retain their source target.
	resumed := rewriteAsyncReturnsVoid(bodySegments[1], promiseName, frameName, result.Span)
	resumed = rewriteLoopControl(resumed, tail, runnerName, await.Span)
	success.Body = append(success.Body, resumed...)
	if !instructionListTerminates(resumed) {
		success.Body = append(success.Body, loop.Step...)
		success.Body = append(success.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: runnerName, Args: asyncRunnerCallArgs(), Span: await.Span})
		success.Body = append(success.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: await.Span})
	}

	reject := ir.Function{
		Name:       rejectName,
		Span:       await.Span,
		Parameters: asyncResumeParameters(),
		Captured:   asyncCapturedParameters(continuationCaptures, valueTypes, promiseName),
		ReturnType: ir.TypeVoid,
		Body: append([]ir.Instruction{},
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_reject_existing", Args: []string{promiseName, "__resume_raw"}, Span: await.Span},
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: await.Span},
			ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: await.Span},
		),
	}
	result.Body = append(result.Body,
		ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: runnerClosure, Callee: runnerName, Args: runnerCaptures, Span: loop.Span},
		ir.Instruction{Op: ir.OpClosureCall, Type: ir.TypeVoid, Callee: runnerClosure, Span: loop.Span},
		ir.Instruction{Op: ir.OpReturn, Type: result.ReturnType, Args: []string{promiseName}, Span: result.Span},
	)

	extraFunctions = append(extraFunctions, runner, success, reject)
	return result, true, nil
}

// lowerStructuredAsyncLoopTry handles the common retry pattern where the loop
// body is a try/catch containing one await. The catch path resumes the loop;
// the fulfilled path may either continue or settle the async function.
func lowerStructuredAsyncLoopTry(path string, statement typescriptgo.SyntaxStatement, lowered ir.Function, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, loopIndex int, loop ir.Instruction, try ir.Instruction) (ir.Function, bool, error) {
	segments, awaits := splitLinearAsyncBody(try.Body)
	if len(awaits) != 1 || len(try.Finally) > 0 || hasNestedAwait(try.Catch) {
		return ir.Function{}, false, nil
	}

	asyncLowerCounter++
	base := fmt.Sprintf("__async_structured_try_loop_%d", asyncLowerCounter)
	promiseName := base + ".promise"
	frameName := base + ".frame"
	result := ir.Function{
		Name: statement.Name, Span: toIRSpan(path, statement.Span),
		Parameters: lowered.Parameters, Locals: lowered.Locals,
		ReturnType: ir.Type("object:Promise"), Async: true, EntryBlock: "async.entry",
	}
	result.Body = append(result.Body,
		ir.Instruction{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: promiseName, Callee: "__async.promise_create", Span: result.Span},
		ir.Instruction{Op: ir.OpCall, Type: ir.TypePointer, Result: frameName, Callee: "__async.frame_new", Value: fmt.Sprintf("%d", len(lowered.Parameters)+len(lowered.Captured)+len(lowered.Locals)+len(awaits)+2), Span: result.Span},
	)
	result.AsyncFrame = buildAsyncFrame(result, lowered, awaits)
	result.Blocks = buildAsyncBlocks(awaits, result.Span)
	result.Body = append(result.Body, rewriteAsyncReturns(lowered.Body[:loopIndex], promiseName, frameName, result.Span)...)

	valueTypes := asyncValueTypes(lowered)
	valueTypes[promiseName] = ir.Type("object:Promise")
	valueTypes[frameName] = ir.TypePointer
	runnerScope := append(append([]ir.Instruction{}, loop.Cond...), segments[0]...)
	continuationScope := append(append([]ir.Instruction{}, segments[1]...), try.Catch...)
	runnerCaptures := asyncCaptures(runnerScope, valueTypes, lowered.Parameters)
	continuationCaptures := asyncCaptures(continuationScope, valueTypes, lowered.Parameters)
	continuationCaptures = removeCapture(continuationCaptures, awaits[0].Result)
	continuationCaptures = removeCapture(continuationCaptures, try.CatchVar)
	seen := map[string]bool{}
	for _, name := range runnerCaptures {
		seen[name] = true
	}
	addCapture := func(list *[]string, name string) {
		if name != "" && !seen[name] {
			*list = append(*list, name)
			seen[name] = true
		}
	}
	for _, name := range runnerCaptures {
		_ = name
	}
	for _, name := range continuationCaptures {
		if !seen[name] {
			runnerCaptures = append(runnerCaptures, name)
			seen[name] = true
		}
	}
	// Continuations use the runner environment as a stable prefix.
	ordered := append([]string{}, runnerCaptures...)
	for _, name := range continuationCaptures {
		addCapture(&ordered, name)
	}
	continuationCaptures = ordered
	addCapture(&runnerCaptures, promiseName)
	addCapture(&runnerCaptures, frameName)
	for _, p := range lowered.Parameters {
		addCapture(&runnerCaptures, p.Name)
	}
	for _, l := range lowered.Locals {
		addCapture(&runnerCaptures, l.Name)
	}
	// Rebuild continuation ordering after adding the common async captures.
	for _, name := range runnerCaptures {
		if !containsString(continuationCaptures, name) {
			continuationCaptures = append(continuationCaptures, name)
		}
	}
	addCapture(&continuationCaptures, promiseName)
	addCapture(&continuationCaptures, frameName)

	runnerName := base + ".runner"
	runnerClosure := base + ".runner.closure"
	successName := base + ".fulfilled"
	rejectName := base + ".rejected"
	fulfilledClosure := base + ".fulfilled.closure"
	rejectedClosure := base + ".rejected.closure"

	tail := rewriteAsyncReturnsVoid(lowered.Body[loopIndex+1:], promiseName, frameName, result.Span)
	if len(tail) == 0 || !instructionListTerminates(tail) {
		tail = append(tail, settleAsyncVoid(promiseName, frameName, result.Span)...)
	}
	continuePath := append([]ir.Instruction{}, loop.Step...)
	continuePath = append(continuePath, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: runnerName, Args: asyncRunnerCallArgs(), Span: loop.Span}, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: loop.Span})

	runner := ir.Function{Name: runnerName, Span: loop.Span, Parameters: asyncRunnerParameters(), Captured: asyncCapturedParameters(runnerCaptures, valueTypes, promiseName), ReturnType: ir.TypeVoid}
	runner.Body = append(runner.Body, loop.Cond...)
	await := awaits[0]
	schedule := []ir.Instruction{
		{Op: ir.OpClosure, Type: ir.TypeClosure, Result: fulfilledClosure, Callee: successName, Args: continuationCaptures, Span: await.Span},
		{Op: ir.OpClosure, Type: ir.TypeClosure, Result: rejectedClosure, Callee: rejectName, Args: continuationCaptures, Span: await.Span},
		{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_schedule_resume_pair", Args: []string{await.Args[0], fulfilledClosure, rejectedClosure}, Span: await.Span},
		{Op: ir.OpReturn, Type: ir.TypeVoid, Span: await.Span},
	}
	runner.Body = append(runner.Body, ir.Instruction{Op: ir.OpIf, Type: ir.TypeVoid, Args: []string{loop.Args[0]}, Then: append(append([]ir.Instruction{}, segments[0]...), schedule...), Else: tail, Span: loop.Span})

	success := ir.Function{Name: successName, Span: await.Span, Parameters: asyncResumeParameters(), Captured: asyncCapturedParameters(continuationCaptures, valueTypes, promiseName), ReturnType: ir.TypeVoid}
	success.Body = append(success.Body, ir.Instruction{Op: ir.OpCheckedCast, Type: await.Type, Result: await.Result, Args: []string{"__resume_raw"}, Span: await.Span})
	success.Body = append(success.Body, rewriteAsyncReturnsVoid(segments[1], promiseName, frameName, result.Span)...)
	if !instructionListTerminates(success.Body) {
		success.Body = append(success.Body, continuePath...)
	}

	reject := ir.Function{Name: rejectName, Span: await.Span, Parameters: asyncResumeParameters(), Captured: asyncCapturedParameters(continuationCaptures, valueTypes, promiseName), ReturnType: ir.TypeVoid}
	if try.CatchVar != "" {
		reject.Body = append(reject.Body, ir.Instruction{Op: ir.OpCheckedCast, Type: ir.Type("object:Error"), Result: try.CatchVar, Args: []string{"__resume_raw"}, Span: await.Span})
	}
	reject.Body = append(reject.Body, try.Catch...)
	reject.Body = append(reject.Body, rewriteAsyncReturnsVoid(nil, promiseName, frameName, result.Span)...)
	if !instructionListTerminates(reject.Body) {
		reject.Body = append(reject.Body, continuePath...)
	}

	result.Body = append(result.Body,
		ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: runnerClosure, Callee: runnerName, Args: runnerCaptures, Span: loop.Span},
		ir.Instruction{Op: ir.OpClosureCall, Type: ir.TypeVoid, Callee: runnerClosure, Span: loop.Span},
		ir.Instruction{Op: ir.OpReturn, Type: result.ReturnType, Args: []string{promiseName}, Span: result.Span},
	)
	extraFunctions = append(extraFunctions, runner, success, reject)
	return result, true, nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func findAsyncLoop(body []ir.Instruction) (int, ir.Instruction, bool) {
	found := -1
	var candidate ir.Instruction
	for i, instruction := range body {
		if !hasAwait(instruction) {
			continue
		}
		if found >= 0 {
			if instruction.Op == ir.OpCall && strings.HasPrefix(instruction.Callee, "__async.await") {
				continue
			}
			return -1, ir.Instruction{}, false
		}
		if instruction.Op != ir.OpWhile && instruction.Op != ir.OpDoWhile {
			return -1, ir.Instruction{}, false
		}
		if hasNestedAwait(instruction.Cond) || hasNestedAwait(instruction.Step) || !supportedAsyncLoopBody(instruction.Body) {
			return -1, ir.Instruction{}, false
		}
		found, candidate = i, instruction
		continue
	}
	return found, candidate, found >= 0
}

func supportedAsyncLoopBody(body []ir.Instruction) bool {
	if countDirectAwaits(body) >= 1 {
		return true
	}
	if len(body) != 1 || body[0].Op != ir.OpTry || len(body[0].Finally) > 0 || hasNestedAwait(body[0].Catch) {
		return false
	}
	_, awaits := splitLinearAsyncBody(body[0].Body)
	return len(awaits) == 1
}

func hasAwait(instructions ir.Instruction) bool {
	if instructions.Op == ir.OpCall && strings.HasPrefix(instructions.Callee, "__async.await") {
		return true
	}
	return hasNestedAwait([]ir.Instruction{instructions})
}

func countDirectAwaits(instructions []ir.Instruction) int {
	count := 0
	for _, instruction := range instructions {
		if instruction.Op == ir.OpCall && strings.HasPrefix(instruction.Callee, "__async.await") {
			count++
		}
	}
	return count
}

func collectNestedAwaits(instructions []ir.Instruction) []ir.Instruction {
	var result []ir.Instruction
	for _, instruction := range instructions {
		if instruction.Op == ir.OpCall && strings.HasPrefix(instruction.Callee, "__async.await") {
			result = append(result, instruction)
		}
		for _, nested := range [][]ir.Instruction{instruction.Then, instruction.Else, instruction.Cond, instruction.Body, instruction.Step, instruction.Catch, instruction.Finally} {
			result = append(result, collectNestedAwaits(nested)...)
		}
	}
	return result
}

func asyncRunnerParameters() []ir.Parameter {
	parameters := []ir.Parameter{{Name: "__env_ctx", Type: ir.TypePointer}}
	for i := 0; i < 4; i++ {
		parameters = append(parameters, ir.Parameter{Name: fmt.Sprintf("__runner_arg_%d$raw", i), Type: ir.TypeUnknown})
	}
	return parameters
}

func asyncResumeParameters() []ir.Parameter {
	parameters := []ir.Parameter{{Name: "__env_ctx", Type: ir.TypePointer}, {Name: "__resume_raw", Type: ir.TypeUnknown}}
	for i := 1; i < 4; i++ {
		parameters = append(parameters, ir.Parameter{Name: fmt.Sprintf("__resume_arg_%d$raw", i), Type: ir.TypeUnknown})
	}
	return parameters
}

func asyncRunnerCallArgs() []string {
	return []string{"__env_ctx", "__resume_raw", "__resume_arg_1$raw", "__resume_arg_2$raw", "__resume_arg_3$raw"}
}

func asyncLoopCaptures(lowered ir.Function, loop ir.Instruction, segments [][]ir.Instruction, promiseName, frameName string, types map[string]ir.Type) ([]string, []string) {
	runnerScope := append(append([]ir.Instruction{}, loop.Cond...), segments[0]...)
	runner := asyncCaptures(runnerScope, types, lowered.Parameters)
	continuationScope := append(append([]ir.Instruction{}, segments[1]...), loop.Step...)
	continuation := asyncCaptures(continuationScope, types, lowered.Parameters)
	continuation = removeCapture(continuation, loopAwaitResult(loop))
	seenRunner := make(map[string]bool, len(runner))
	for _, name := range runner {
		seenRunner[name] = true
	}
	seenContinuation := make(map[string]bool, len(continuation))
	for _, name := range continuation {
		seenContinuation[name] = true
	}
	add := func(list *[]string, seen map[string]bool, name string) {
		if name != "" && !seen[name] {
			*list = append(*list, name)
			seen[name] = true
		}
	}
	add(&runner, seenRunner, promiseName)
	add(&runner, seenRunner, frameName)
	add(&continuation, seenContinuation, promiseName)
	add(&continuation, seenContinuation, frameName)
	for _, parameter := range lowered.Parameters {
		add(&runner, seenRunner, parameter.Name)
		add(&continuation, seenContinuation, parameter.Name)
	}
	for _, local := range lowered.Locals {
		// Compiler-generated temporaries created in the loop are local to the
		// runner; the continuation captures them after they have been defined.
		if !strings.HasPrefix(local.Name, "__") {
			add(&runner, seenRunner, local.Name)
		}
		add(&continuation, seenContinuation, local.Name)
	}
	// Recursive runner calls pass the continuation environment directly. Keep
	// the runner's fields as the stable prefix of that environment.
	orderedContinuation := append([]string{}, runner...)
	for _, name := range continuation {
		add(&orderedContinuation, seenContinuation, name)
	}
	continuation = orderedContinuation
	var addDefined func([]ir.Instruction)
	addDefined = func(instructions []ir.Instruction) {
		for _, instruction := range instructions {
			if instruction.Result != "" && strings.HasPrefix(instruction.Result, "__") {
				add(&continuation, seenContinuation, instruction.Result)
			}
			addDefined(instruction.Then)
			addDefined(instruction.Else)
			addDefined(instruction.Cond)
			addDefined(instruction.Body)
			addDefined(instruction.Step)
			addDefined(instruction.Catch)
			addDefined(instruction.Finally)
		}
	}
	addDefined(segments[0])
	return runner, continuation
}

func loopAwaitResult(loop ir.Instruction) string {
	for _, instruction := range loop.Body {
		if instruction.Op == ir.OpCall && strings.HasPrefix(instruction.Callee, "__async.await") {
			return instruction.Result
		}
	}
	return ""
}

func removeCapture(captures []string, name string) []string {
	if name == "" {
		return captures
	}
	result := captures[:0]
	for _, capture := range captures {
		if capture != name {
			result = append(result, capture)
		}
	}
	return result
}

func rewriteLoopControl(instructions, exit []ir.Instruction, runnerName string, span ir.SourceSpan) []ir.Instruction {
	result := make([]ir.Instruction, 0, len(instructions))
	for _, instruction := range instructions {
		if instruction.Op == ir.OpBreak {
			result = append(result, exit...)
			continue
		}
		if instruction.Op == ir.OpContinue {
			result = append(result, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: runnerName, Args: []string{}, Span: span}, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: span})
			continue
		}
		instruction.Then = rewriteLoopControl(instruction.Then, exit, runnerName, span)
		instruction.Else = rewriteLoopControl(instruction.Else, exit, runnerName, span)
		instruction.Cond = rewriteLoopControl(instruction.Cond, exit, runnerName, span)
		instruction.Body = rewriteLoopControl(instruction.Body, exit, runnerName, span)
		instruction.Step = rewriteLoopControl(instruction.Step, exit, runnerName, span)
		instruction.Catch = rewriteLoopControl(instruction.Catch, exit, runnerName, span)
		instruction.Finally = rewriteLoopControl(instruction.Finally, exit, runnerName, span)
		result = append(result, instruction)
	}
	return result
}

func rewriteAsyncReturnsVoid(instructions []ir.Instruction, promiseName, frameName string, span ir.SourceSpan) []ir.Instruction {
	result := make([]ir.Instruction, 0, len(instructions)+3)
	for _, instruction := range instructions {
		instruction.Then = rewriteAsyncReturnsVoid(instruction.Then, promiseName, frameName, span)
		instruction.Else = rewriteAsyncReturnsVoid(instruction.Else, promiseName, frameName, span)
		instruction.Cond = rewriteAsyncReturnsVoid(instruction.Cond, promiseName, frameName, span)
		instruction.Body = rewriteAsyncReturnsVoid(instruction.Body, promiseName, frameName, span)
		instruction.Step = rewriteAsyncReturnsVoid(instruction.Step, promiseName, frameName, span)
		instruction.Catch = rewriteAsyncReturnsVoid(instruction.Catch, promiseName, frameName, span)
		instruction.Finally = rewriteAsyncReturnsVoid(instruction.Finally, promiseName, frameName, span)
		if instruction.Op == ir.OpThrow {
			if len(instruction.Args) > 0 {
				result = append(result, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_reject_existing", Args: []string{promiseName, instruction.Args[0]}, Span: instruction.Span})
			}
			result = append(result,
				ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: instruction.Span},
				ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: instruction.Span},
			)
			continue
		}
		if instruction.Op != ir.OpReturn {
			result = append(result, instruction)
			continue
		}
		value := "__async.undefined"
		if len(instruction.Args) > 0 {
			value = instruction.Args[0]
		}
		if len(instruction.Args) == 0 {
			result = append(result, ir.Instruction{Op: ir.OpConst, Type: ir.TypeUnknown, Result: value, Value: "undefined", Span: instruction.Span})
		}
		result = append(result, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_resolve_existing", Args: []string{promiseName, value}, Span: instruction.Span}, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: instruction.Span}, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: instruction.Span})
	}
	return result
}

// rewriteAsyncEntryReturns rewrites returns on a path that still executes in
// the async entry function. Unlike a resume callback, this path must return
// the outer Promise after settling it.
func rewriteAsyncEntryReturns(instructions []ir.Instruction, promiseName, frameName string, returnType ir.Type, span ir.SourceSpan) []ir.Instruction {
	result := make([]ir.Instruction, 0, len(instructions)+3)
	for _, instruction := range instructions {
		instruction.Then = rewriteAsyncEntryReturns(instruction.Then, promiseName, frameName, returnType, span)
		instruction.Else = rewriteAsyncEntryReturns(instruction.Else, promiseName, frameName, returnType, span)
		instruction.Cond = rewriteAsyncEntryReturns(instruction.Cond, promiseName, frameName, returnType, span)
		instruction.Body = rewriteAsyncEntryReturns(instruction.Body, promiseName, frameName, returnType, span)
		instruction.Step = rewriteAsyncEntryReturns(instruction.Step, promiseName, frameName, returnType, span)
		instruction.Catch = rewriteAsyncEntryReturns(instruction.Catch, promiseName, frameName, returnType, span)
		instruction.Finally = rewriteAsyncEntryReturns(instruction.Finally, promiseName, frameName, returnType, span)
		if instruction.Op != ir.OpReturn && instruction.Op != ir.OpThrow {
			result = append(result, instruction)
			continue
		}
		value := "__async.undefined"
		if len(instruction.Args) > 0 {
			value = instruction.Args[0]
		} else {
			result = append(result, ir.Instruction{Op: ir.OpConst, Type: ir.TypeUnknown, Result: value, Value: "undefined", Span: instruction.Span})
		}
		callee := "__async.promise_resolve_existing"
		if instruction.Op == ir.OpThrow {
			callee = "__async.promise_reject_existing"
		}
		result = append(result,
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: callee, Args: []string{promiseName, value}, Span: instruction.Span},
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: instruction.Span},
			ir.Instruction{Op: ir.OpReturn, Type: returnType, Args: []string{promiseName}, Span: instruction.Span},
		)
	}
	return result
}

func settleAsyncEntry(promiseName, frameName string, returnType ir.Type, span ir.SourceSpan) []ir.Instruction {
	return []ir.Instruction{
		{Op: ir.OpConst, Type: ir.TypeUnknown, Result: "__async.undefined", Value: "undefined", Span: span},
		{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_resolve_existing", Args: []string{promiseName, "__async.undefined"}, Span: span},
		{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: span},
		{Op: ir.OpReturn, Type: returnType, Args: []string{promiseName}, Span: span},
	}
}

func settleAsyncVoid(promiseName, frameName string, span ir.SourceSpan) []ir.Instruction {
	return []ir.Instruction{
		{Op: ir.OpConst, Type: ir.TypeUnknown, Result: "__async.undefined", Value: "undefined", Span: span},
		{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_resolve_existing", Args: []string{promiseName, "__async.undefined"}, Span: span},
		{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: span},
		{Op: ir.OpReturn, Type: ir.TypeVoid, Span: span},
	}
}

func instructionListTerminates(instructions []ir.Instruction) bool {
	if len(instructions) == 0 {
		return false
	}
	last := instructions[len(instructions)-1]
	switch last.Op {
	case ir.OpReturn, ir.OpThrow:
		return true
	case ir.OpIf:
		return len(last.Else) > 0 && instructionListTerminates(last.Then) && instructionListTerminates(last.Else)
	case ir.OpTry:
		if len(last.Finally) > 0 && instructionListTerminates(last.Finally) {
			return true
		}
		return instructionListTerminates(last.Body) && len(last.Catch) > 0 && instructionListTerminates(last.Catch)
	default:
		return false
	}
}
