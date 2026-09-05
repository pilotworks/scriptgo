package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

type asyncTryState struct {
	index int
	inst  ir.Instruction
	body  [][]ir.Instruction
	await ir.Instruction
}

func lowerStructuredAsyncTryChain(path string, statement typescriptgo.SyntaxStatement, lowered ir.Function, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (ir.Function, bool, error) {
	var states []asyncTryState
	for index, instruction := range lowered.Body {
		if !hasAwait(instruction) {
			continue
		}
		if instruction.Op != ir.OpTry || hasStructuredAwait(instruction.Catch) || hasStructuredAwait(instruction.Finally) {
			return ir.Function{}, false, nil
		}
		segments, awaits := splitLinearAsyncBody(instruction.Body)
		if len(awaits) != 1 {
			return ir.Function{}, false, nil
		}
		states = append(states, asyncTryState{index: index, inst: instruction, body: segments, await: awaits[0]})
	}
	if len(states) < 2 {
		return ir.Function{}, false, nil
	}

	asyncLowerCounter++
	base := fmt.Sprintf("__async_try_chain_%d", asyncLowerCounter)
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
	allAwaits := make([]ir.Instruction, 0, len(states))
	for _, state := range states {
		allAwaits = append(allAwaits, state.await)
	}
	result.Body = append(result.Body,
		ir.Instruction{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: promiseName, Callee: "__async.promise_create", Span: result.Span},
		ir.Instruction{Op: ir.OpCall, Type: ir.TypePointer, Result: frameName, Callee: "__async.frame_new", Value: fmt.Sprintf("%d", len(lowered.Parameters)+len(lowered.Captured)+len(lowered.Locals)+len(allAwaits)+2), Span: result.Span},
	)
	result.AsyncFrame = buildAsyncFrame(result, lowered, allAwaits)
	result.Blocks = buildAsyncBlocks(allAwaits, result.Span)
	result.Body = append(result.Body, rewriteAsyncReturns(lowered.Body[:states[0].index], promiseName, frameName, result.Span)...)

	valueTypes := asyncValueTypes(lowered)
	valueTypes[promiseName] = ir.Type("object:Promise")
	valueTypes[frameName] = ir.TypePointer
	captures := asyncTryChainCaptures(lowered, states, promiseName, frameName, valueTypes)
	stateClosures := make([]string, len(states))
	for i := range states {
		stateClosures[i] = fmt.Sprintf("%s.state.%d.closure", base, i)
	}
	stateNames := make([]string, len(states))
	for i := range states {
		stateNames[i] = fmt.Sprintf("%s.state.%d", base, i)
	}

	result.Body = append(result.Body,
		ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: stateClosures[0], Callee: stateNames[0], Args: captures, Span: states[0].await.Span},
		ir.Instruction{Op: ir.OpClosureCall, Type: ir.TypeVoid, Callee: stateClosures[0], Span: states[0].await.Span},
		ir.Instruction{Op: ir.OpReturn, Type: result.ReturnType, Args: []string{promiseName}, Span: result.Span},
	)

	var functions []ir.Function
	tail := rewriteAsyncReturnsVoid(lowered.Body[states[len(states)-1].index+1:], promiseName, frameName, result.Span)
	if len(tail) == 0 || !instructionListTerminates(tail) {
		tail = append(tail, settleAsyncVoid(promiseName, frameName, result.Span)...)
	}
	for i, state := range states {
		stateFn := ir.Function{Name: stateNames[i], Span: state.inst.Span, Parameters: asyncRunnerParameters(), Captured: asyncCapturedParameters(captures, valueTypes, promiseName), ReturnType: ir.TypeVoid}
		stateFn.Body = append(stateFn.Body, state.body[0]...)
		fulfilledName := fmt.Sprintf("%s.fulfilled.%d", base, i)
		rejectedName := fmt.Sprintf("%s.rejected.%d", base, i)
		fulfilled := ir.Function{Name: fulfilledName, Span: state.await.Span, Parameters: asyncResumeParameters(), Captured: asyncCapturedParameters(captures, valueTypes, promiseName), ReturnType: ir.TypeVoid}
		fulfilled.Body = append(fulfilled.Body, ir.Instruction{Op: ir.OpCheckedCast, Type: state.await.Type, Result: state.await.Result, Args: []string{"__resume_raw"}, Span: state.await.Span})
		fulfilled.Body = append(fulfilled.Body, state.body[1]...)
		fulfilled.Body = append(fulfilled.Body, state.inst.Finally...)
		rejected := ir.Function{Name: rejectedName, Span: state.await.Span, Parameters: asyncResumeParameters(), Captured: asyncCapturedParameters(captures, valueTypes, promiseName), ReturnType: ir.TypeVoid}
		if state.inst.CatchVar != "" {
			rejected.Body = append(rejected.Body, ir.Instruction{Op: ir.OpCheckedCast, Type: ir.Type("object:Error"), Result: state.inst.CatchVar, Args: []string{"__resume_raw"}, Span: state.await.Span})
		}
		rejected.Body = append(rejected.Body, state.inst.Catch...)
		rejected.Body = append(rejected.Body, state.inst.Finally...)
		if i+1 < len(states) {
			next := ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: stateNames[i+1], Args: asyncRunnerCallArgs(), Span: state.await.Span}
			fulfilled.Body = append(fulfilled.Body, next)
			fulfilled.Body = append(fulfilled.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: state.await.Span})
			rejected.Body = append(rejected.Body, next)
			rejected.Body = append(rejected.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: state.await.Span})
		} else {
			fulfilled.Body = append(fulfilled.Body, tail...)
			rejected.Body = append(rejected.Body, tail...)
		}
		functions = append(functions, fulfilled, rejected)
		stateFn.Body = append(stateFn.Body,
			ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: fmt.Sprintf("%s.fulfilled.closure", stateNames[i]), Callee: fulfilledName, Args: captures, Span: state.await.Span},
			ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: fmt.Sprintf("%s.rejected.closure", stateNames[i]), Callee: rejectedName, Args: captures, Span: state.await.Span},
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_schedule_resume_pair", Args: []string{state.await.Args[0], fmt.Sprintf("%s.fulfilled.closure", stateNames[i]), fmt.Sprintf("%s.rejected.closure", stateNames[i])}, Span: state.await.Span},
			ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: state.await.Span},
		)
		functions = append(functions, stateFn)
	}
	extraFunctions = append(extraFunctions, functions...)
	return result, true, nil
}

func asyncTryChainCaptures(lowered ir.Function, states []asyncTryState, promiseName, frameName string, types map[string]ir.Type) []string {
	var scope []ir.Instruction
	for _, state := range states {
		scope = append(scope, state.body[0]...)
		scope = append(scope, state.body[1]...)
		scope = append(scope, state.inst.Catch...)
		scope = append(scope, state.inst.Finally...)
	}
	captures := asyncCaptures(scope, types, lowered.Parameters)
	for _, state := range states {
		captures = removeCapture(captures, state.await.Result)
		captures = removeCapture(captures, state.inst.CatchVar)
	}
	seen := map[string]bool{}
	result := make([]string, 0, len(captures)+2)
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}
	for _, name := range captures {
		add(name)
	}
	add(promiseName)
	add(frameName)
	for _, parameter := range lowered.Parameters {
		add(parameter.Name)
	}
	return result
}
