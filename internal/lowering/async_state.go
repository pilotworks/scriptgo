package lowering

import (
	"fmt"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

var asyncLowerCounter int

// lowerAsyncFunction lowers the synchronous prefix of an async function and
// turns each linear await into a pair of closure continuations. This is the
// first stage of async CPS lowering; nested control-flow is handled by the
// structured lowering pass before this representation is marked as a CFG.
func lowerAsyncFunction(path string, statement typescriptgo.SyntaxStatement, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (ir.Function, error) {
	payloadType, ok := asyncResolvedReturnType(statement.Type)
	if !ok {
		payloadType, ok = asyncResolvedReturnType(statement.InferredType)
	}
	if !ok {
		payloadType = ir.TypeUnknown
	}

	plain := statement
	plain.Kind = "function"
	plain.IsAsync = false
	plain.Type = string(payloadType)
	plain.InferredType = string(payloadType)
	lowered, err := lowerSyncFunction(path, plain, shapes, signatures)
	if err != nil {
		return ir.Function{}, err
	}
	return lowerAsyncFunctionFromLowered(path, statement, lowered, shapes, signatures)
}

func lowerAsyncFunctionFromLowered(path string, statement typescriptgo.SyntaxStatement, lowered ir.Function, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (ir.Function, error) {
	finish := func(function ir.Function) (ir.Function, error) {
		function.Captured = lowered.Captured
		return function, nil
	}
	if hasStructuredAwait(lowered.Body) {
		if structured, ok, err := lowerStructuredAsyncTryChain(path, statement, lowered, shapes, signatures); err != nil {
			return ir.Function{}, err
		} else if ok {
			return finish(structured)
		}
		if structured, ok, err := lowerStructuredAsyncTry(path, statement, lowered, shapes, signatures); err != nil {
			return ir.Function{}, err
		} else if ok {
			return finish(structured)
		}
		if structured, ok, err := lowerStructuredAsyncBranchLoop(path, statement, lowered, shapes, signatures); err != nil {
			return ir.Function{}, err
		} else if ok {
			return finish(structured)
		}
		if structured, ok, err := lowerStructuredAsyncBranch(path, statement, lowered, shapes, signatures); err != nil {
			return ir.Function{}, err
		} else if ok {
			return finish(structured)
		}
		if structured, ok, err := lowerStructuredAsyncLoop(path, statement, lowered, shapes, signatures); err != nil {
			return ir.Function{}, err
		} else if ok {
			return finish(structured)
		}
		return ir.Function{}, fmt.Errorf("async function %q contains an unsupported suspension shape", statement.Name)
	}

	asyncLowerCounter++
	promiseName := fmt.Sprintf("__async_promise_%d", asyncLowerCounter)
	result := ir.Function{
		Name:       statement.Name,
		Span:       toIRSpan(path, statement.Span),
		Parameters: lowered.Parameters,
		Locals:     lowered.Locals,
		ReturnType: ir.Type("object:Promise"),
		Captured:   lowered.Captured,
	}
	segments, awaits := splitLinearAsyncBody(lowered.Body)
	if len(awaits) == 0 {
		result, err := lowerAsyncImmediateFunction(path, statement, lowered)
		if err != nil {
			return ir.Function{}, err
		}
		return finish(result)
	}
	result.Body = append(result.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.Type("object:Promise"),
		Result: promiseName,
		Callee: "__async.promise_create",
		Span:   result.Span,
	})
	frameName := promiseName + ".frame"
	result.Body = append(result.Body, ir.Instruction{
		Op: ir.OpCall, Type: ir.TypePointer, Result: frameName,
		Callee: "__async.frame_new", Value: fmt.Sprintf("%d", len(lowered.Parameters)+len(lowered.Captured)+len(lowered.Locals)+len(awaits)+2),
		Span: result.Span,
	})

	result.Async = true
	result.EntryBlock = "async.entry"
	result.AsyncFrame = buildAsyncFrame(result, lowered, awaits)
	result.Blocks = buildAsyncBlocks(awaits, result.Span)
	segments[0] = rewriteAsyncReturns(segments[0], promiseName, frameName, result.Span)

	result.Body = append(result.Body, segments[0]...)
	appendAsyncSuspension(&result, promiseName, frameName, awaits, segments, 0, lowered, path, shapes, signatures)
	result.Body = append(result.Body, ir.Instruction{Op: ir.OpReturn, Type: result.ReturnType, Args: []string{promiseName}, Span: result.Span})
	return result, nil
}

func buildAsyncFrame(function ir.Function, lowered ir.Function, awaits []ir.Instruction) *ir.AsyncFrame {
	frame := &ir.AsyncFrame{Name: "__async_frame_" + function.Name}
	seen := map[string]bool{}
	add := func(name string, typ ir.Type, span ir.SourceSpan) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		if typ == "" {
			typ = ir.TypeUnknown
		}
		frame.Fields = append(frame.Fields, ir.AsyncField{Name: name, Type: typ, Span: span})
	}
	add("state", ir.TypeNumber, function.Span)
	add("promise", ir.Type("object:Promise"), function.Span)
	for _, parameter := range lowered.Parameters {
		add(parameter.Name, parameter.Type, function.Span)
	}
	for _, captured := range lowered.Captured {
		add(captured.Name, captured.Type, function.Span)
	}
	for _, local := range lowered.Locals {
		add(local.Name, local.Type, function.Span)
	}
	for _, await := range awaits {
		add(await.Result, await.Type, await.Span)
	}
	return frame
}

func buildAsyncBlocks(awaits []ir.Instruction, span ir.SourceSpan) []ir.BasicBlock {
	blocks := make([]ir.BasicBlock, 0, len(awaits)*2+1)
	for index, await := range awaits {
		name := fmt.Sprintf("async.state.%d", index)
		next := fmt.Sprintf("async.state.%d", index+1)
		if index+1 == len(awaits) {
			next = "async.done"
		}
		blocks = append(blocks, ir.BasicBlock{
			Name: name,
			Span: await.Span,
			Terminator: ir.Terminator{
				Kind:       ir.TermAwait,
				AwaitValue: await.Args[0],
				Fulfilled:  next,
				Rejected:   fmt.Sprintf("async.reject.%d", index),
				State:      index,
				Span:       await.Span,
			},
		})
		blocks = append(blocks, ir.BasicBlock{
			Name:       nameForAsyncReject(index),
			Span:       await.Span,
			Terminator: ir.Terminator{Kind: ir.TermThrow, Value: "__resume_raw", Span: await.Span},
		})
	}
	entryTarget := "async.state.0"
	if len(awaits) == 0 {
		entryTarget = "async.done"
	}
	blocks = append([]ir.BasicBlock{{Name: "async.entry", Span: span, Terminator: ir.Terminator{Kind: ir.TermJump, Target: entryTarget, Span: span}}}, blocks...)
	blocks = append(blocks, ir.BasicBlock{Name: "async.done", Span: span, Terminator: ir.Terminator{Kind: ir.TermReturn, Span: span}})
	return blocks
}

func nameForAsyncReject(index int) string {
	return fmt.Sprintf("async.reject.%d", index)
}

func splitLinearAsyncBody(body []ir.Instruction) ([][]ir.Instruction, []ir.Instruction) {
	segments := [][]ir.Instruction{{}}
	var awaits []ir.Instruction
	for _, instruction := range body {
		if instruction.Op == ir.OpCall && strings.HasPrefix(instruction.Callee, "__async.await") {
			awaits = append(awaits, instruction)
			segments = append(segments, []ir.Instruction{})
			continue
		}
		segments[len(segments)-1] = append(segments[len(segments)-1], instruction)
	}
	return segments, awaits
}

func hasNestedAwait(instructions []ir.Instruction) bool {
	for _, instruction := range instructions {
		if instruction.Op == ir.OpCall && strings.HasPrefix(instruction.Callee, "__async.await") {
			return true
		}
		if hasNestedAwait(instruction.Then) || hasNestedAwait(instruction.Else) ||
			hasNestedAwait(instruction.Cond) || hasNestedAwait(instruction.Body) ||
			hasNestedAwait(instruction.Step) || hasNestedAwait(instruction.Catch) ||
			hasNestedAwait(instruction.Finally) {
			return true
		}
	}
	return false
}

func hasStructuredAwait(instructions []ir.Instruction) bool {
	for _, instruction := range instructions {
		nested := make([]ir.Instruction, 0, len(instruction.Then)+len(instruction.Else)+len(instruction.Cond)+len(instruction.Body)+len(instruction.Step)+len(instruction.Catch)+len(instruction.Finally))
		nested = append(nested, instruction.Then...)
		nested = append(nested, instruction.Else...)
		nested = append(nested, instruction.Cond...)
		nested = append(nested, instruction.Body...)
		nested = append(nested, instruction.Step...)
		nested = append(nested, instruction.Catch...)
		nested = append(nested, instruction.Finally...)
		if hasNestedAwait(nested) {
			return true
		}
	}
	return false
}

func stripReturn(instructions []ir.Instruction) []ir.Instruction {
	result := make([]ir.Instruction, 0, len(instructions))
	for _, instruction := range instructions {
		if instruction.Op != ir.OpReturn {
			result = append(result, instruction)
		}
	}
	return result
}

// rewriteAsyncReturns converts early returns in structured prefix instructions
// into settlement of the outer promise before returning from the async entry.
func rewriteAsyncReturns(instructions []ir.Instruction, promiseName, frameName string, span ir.SourceSpan) []ir.Instruction {
	result := make([]ir.Instruction, 0, len(instructions))
	for _, instruction := range instructions {
		instruction.Then = rewriteAsyncReturns(instruction.Then, promiseName, frameName, span)
		instruction.Else = rewriteAsyncReturns(instruction.Else, promiseName, frameName, span)
		instruction.Cond = rewriteAsyncReturns(instruction.Cond, promiseName, frameName, span)
		instruction.Body = rewriteAsyncReturns(instruction.Body, promiseName, frameName, span)
		instruction.Step = rewriteAsyncReturns(instruction.Step, promiseName, frameName, span)
		instruction.Catch = rewriteAsyncReturns(instruction.Catch, promiseName, frameName, span)
		instruction.Finally = rewriteAsyncReturns(instruction.Finally, promiseName, frameName, span)
		if instruction.Op == ir.OpThrow {
			value := "__async.undefined"
			if len(instruction.Args) > 0 {
				value = instruction.Args[0]
			} else {
				result = append(result, ir.Instruction{Op: ir.OpConst, Type: ir.TypeUnknown, Result: value, Value: "undefined", Span: instruction.Span})
			}
			result = append(result,
				ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_reject_existing", Args: []string{promiseName, value}, Span: instruction.Span},
				ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: instruction.Span},
				ir.Instruction{Op: ir.OpReturn, Type: ir.Type("object:Promise"), Args: []string{promiseName}, Span: instruction.Span},
			)
			continue
		}
		if instruction.Op != ir.OpReturn {
			result = append(result, instruction)
			continue
		}
		if len(instruction.Args) == 0 {
			result = append(result,
				ir.Instruction{Op: ir.OpConst, Type: ir.TypeUnknown, Result: "__async.undefined", Value: "undefined", Span: instruction.Span},
				ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_resolve_existing", Args: []string{promiseName, "__async.undefined"}, Span: instruction.Span},
				ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: instruction.Span},
				ir.Instruction{Op: ir.OpReturn, Type: ir.Type("object:Promise"), Args: []string{promiseName}, Span: instruction.Span})
			continue
		}
		result = append(result,
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_resolve_existing", Args: []string{promiseName, instruction.Args[0]}, Span: instruction.Span},
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: instruction.Span},
			ir.Instruction{Op: ir.OpReturn, Type: ir.Type("object:Promise"), Args: []string{promiseName}, Span: instruction.Span})
	}
	return result
}

func lowerAsyncImmediateFunction(path string, statement typescriptgo.SyntaxStatement, lowered ir.Function) (ir.Function, error) {
	asyncLowerCounter++
	base := fmt.Sprintf("__async_immediate_%d", asyncLowerCounter)
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
	result.AsyncFrame = &ir.AsyncFrame{Name: "__async_frame_" + statement.Name, Fields: []ir.AsyncField{
		{Name: "state", Type: ir.TypeNumber, Span: result.Span},
		{Name: "promise", Type: ir.Type("object:Promise"), Span: result.Span},
	}}
	result.Blocks = buildAsyncBlocks(nil, result.Span)
	result.Body = append(result.Body,
		ir.Instruction{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: promiseName, Callee: "__async.promise_create", Span: result.Span},
		ir.Instruction{Op: ir.OpCall, Type: ir.TypePointer, Result: frameName, Callee: "__async.frame_new", Value: "2", Span: result.Span},
	)
	result.Body = append(result.Body, rewriteAsyncReturns(lowered.Body, promiseName, frameName, result.Span)...)
	if !instructionListTerminates(result.Body) {
		result.Body = append(result.Body, settleAsyncVoid(promiseName, frameName, result.Span)...)
		result.Body = append(result.Body, ir.Instruction{Op: ir.OpReturn, Type: result.ReturnType, Args: []string{promiseName}, Span: result.Span})
	}
	return result, nil
}

func findReturnValue(instructions []ir.Instruction) string {
	for i := len(instructions) - 1; i >= 0; i-- {
		if instructions[i].Op == ir.OpReturn && len(instructions[i].Args) > 0 {
			return instructions[i].Args[0]
		}
	}
	return ""
}

func appendAsyncResolve(function *ir.Function, promiseName, value string, span ir.SourceSpan) (ir.Function, error) {
	if value == "" {
		value = appendResolvedPromiseUndefined(function, new(int), span)
		function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: function.ReturnType, Args: []string{value}, Span: span})
		return *function, nil
	}
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeVoid,
		Callee: "__async.promise_resolve_existing",
		Args:   []string{promiseName, value},
		Span:   span,
	})
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpReturn, Type: function.ReturnType, Args: []string{promiseName}, Span: span})
	return *function, nil
}

func appendAsyncSuspension(function *ir.Function, promiseName, frameName string, awaits []ir.Instruction, segments [][]ir.Instruction, index int, source ir.Function, path string, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) {
	if index >= len(awaits) {
		return
	}
	successName, rejectName := makeAsyncContinuation(function, promiseName, frameName, awaits, segments, index, source, path, shapes, signatures)
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.TypeVoid,
		Callee: "__async.promise_schedule_resume_pair",
		Args:   []string{awaits[index].Args[0], successName, rejectName},
		Span:   awaits[index].Span,
	})
}

func makeAsyncContinuation(owner *ir.Function, promiseName, frameName string, awaits []ir.Instruction, segments [][]ir.Instruction, index int, source ir.Function, path string, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (string, string) {
	asyncLowerCounter++
	base := fmt.Sprintf("__closure_async_%d", asyncLowerCounter)
	valueTypes := asyncValueTypes(source)
	valueTypes[frameName] = ir.TypePointer
	segment := segments[index+1]
	captures := asyncCaptures(segment, valueTypes, source.Parameters)
	// The awaited result is materialized by the callback's cast, not a value
	// available in the function that creates the callback.
	if current := awaits[index].Result; current != "" {
		filtered := captures[:0]
		for _, capture := range captures {
			if capture != current {
				filtered = append(filtered, capture)
			}
		}
		captures = filtered
	}
	addCapture := func(name string) {
		for _, existing := range captures {
			if existing == name {
				return
			}
		}
		captures = append(captures, name)
	}
	addCapture(promiseName)
	addCapture(frameName)
	// The continuation must retain the awaited operand when later lowering
	// emits additional Promise combinator states in the same callback.
	addCapture(awaits[index].Args[0])
	for _, parameter := range source.Parameters {
		addCapture(parameter.Name)
	}
	for _, local := range source.Locals {
		addCapture(local.Name)
	}

	success := ir.Function{
		Name:       base + "_fulfilled",
		Span:       awaits[index].Span,
		Parameters: []ir.Parameter{{Name: "__env_ctx", Type: ir.TypePointer}, {Name: "__resume_raw", Type: ir.TypeUnknown}},
		Captured:   asyncCapturedParameters(captures, valueTypes, promiseName),
		Locals:     source.Locals,
		ReturnType: ir.TypeVoid,
	}
	cast := ir.Instruction{Op: ir.OpCheckedCast, Type: awaits[index].Type, Result: awaits[index].Result, Args: []string{"__resume_raw"}, Span: awaits[index].Span}
	success.Body = append(success.Body, cast)
	success.Body = append(success.Body, rewriteAsyncReturnsVoid(segment, promiseName, frameName, awaits[index].Span)...)
	if !instructionListTerminates(success.Body) {
		if index+1 < len(awaits) {
			appendAsyncSuspension(&success, promiseName, frameName, awaits, segments, index+1, source, path, shapes, signatures)
		} else {
			success.Body = append(success.Body, settleAsyncVoid(promiseName, frameName, awaits[index].Span)...)
		}
	}
	if !instructionListTerminates(success.Body) {
		success.Body = append(success.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: awaits[index].Span})
	}

	reject := ir.Function{
		Name:       base + "_rejected",
		Span:       awaits[index].Span,
		Parameters: []ir.Parameter{{Name: "__env_ctx", Type: ir.TypePointer}, {Name: "__resume_raw", Type: ir.TypeUnknown}},
		Captured:   asyncCapturedParameters(captures, valueTypes, promiseName),
		Locals:     source.Locals,
		ReturnType: ir.TypeVoid,
		Body:       []ir.Instruction{{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_reject_existing", Args: []string{promiseName, "__resume_raw"}, Span: awaits[index].Span}, {Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: awaits[index].Span}, {Op: ir.OpReturn, Type: ir.TypeVoid, Span: awaits[index].Span}},
	}

	extraFunctions = append(extraFunctions, success, reject)
	fulfilledClosure := base + "_closure"
	rejectedClosure := base + "_reject_closure"
	owner.Body = append(owner.Body,
		ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: fulfilledClosure, Callee: success.Name, Args: captures, Span: awaits[index].Span},
		ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: rejectedClosure, Callee: reject.Name, Args: captures, Span: awaits[index].Span},
	)
	return fulfilledClosure, rejectedClosure
}

func segmentWithoutReturn(segment []ir.Instruction) []ir.Instruction {
	result := make([]ir.Instruction, 0, len(segment))
	for _, instruction := range segment {
		if instruction.Op != ir.OpReturn {
			result = append(result, instruction)
		}
	}
	return result
}

func rewriteAsyncContinuationThrows(instructions []ir.Instruction, promiseName, frameName string) []ir.Instruction {
	result := make([]ir.Instruction, 0, len(instructions))
	for _, instruction := range instructions {
		instruction.Then = rewriteAsyncContinuationThrows(instruction.Then, promiseName, frameName)
		instruction.Else = rewriteAsyncContinuationThrows(instruction.Else, promiseName, frameName)
		instruction.Cond = rewriteAsyncContinuationThrows(instruction.Cond, promiseName, frameName)
		instruction.Body = rewriteAsyncContinuationThrows(instruction.Body, promiseName, frameName)
		instruction.Step = rewriteAsyncContinuationThrows(instruction.Step, promiseName, frameName)
		instruction.Catch = rewriteAsyncContinuationThrows(instruction.Catch, promiseName, frameName)
		instruction.Finally = rewriteAsyncContinuationThrows(instruction.Finally, promiseName, frameName)
		if instruction.Op != ir.OpThrow {
			result = append(result, instruction)
			continue
		}
		value := "__async.undefined"
		if len(instruction.Args) > 0 {
			value = instruction.Args[0]
		} else {
			result = append(result, ir.Instruction{Op: ir.OpConst, Type: ir.TypeUnknown, Result: value, Value: "undefined", Span: instruction.Span})
		}
		result = append(result,
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_reject_existing", Args: []string{promiseName, value}, Span: instruction.Span},
			ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: instruction.Span},
			ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: instruction.Span},
		)
	}
	return result
}

func findFirstAwait(instructions []ir.Instruction) ir.Instruction {
	for _, instruction := range instructions {
		if instruction.Op == ir.OpCall && strings.HasPrefix(instruction.Callee, "__async.await") {
			return instruction
		}
	}
	return ir.Instruction{}
}

func asyncValueTypes(function ir.Function) map[string]ir.Type {
	result := make(map[string]ir.Type, len(function.Parameters)+len(function.Locals))
	for _, parameter := range function.Parameters {
		result[parameter.Name] = parameter.Type
	}
	for _, local := range function.Locals {
		result[local.Name] = local.Type
	}
	var collect func([]ir.Instruction)
	collect = func(instructions []ir.Instruction) {
		for _, instruction := range instructions {
			if instruction.Result != "" && instruction.Type != "" {
				result[instruction.Result] = instruction.Type
			}
			collect(instruction.Then)
			collect(instruction.Else)
			collect(instruction.Cond)
			collect(instruction.Body)
			collect(instruction.Step)
			collect(instruction.Catch)
			collect(instruction.Finally)
		}
	}
	collect(function.Body)
	return result
}

func asyncCaptures(segment []ir.Instruction, valueTypes map[string]ir.Type, parameters []ir.Parameter) []string {
	params := make(map[string]bool, len(parameters)+1)
	for _, parameter := range parameters {
		params[parameter.Name] = true
	}
	result := []string{}
	defined := map[string]bool{}
	var collectDefinitions func([]ir.Instruction)
	collectDefinitions = func(instructions []ir.Instruction) {
		for _, instruction := range instructions {
			if instruction.Result != "" {
				defined[instruction.Result] = true
			}
			collectDefinitions(instruction.Then)
			collectDefinitions(instruction.Else)
			collectDefinitions(instruction.Cond)
			collectDefinitions(instruction.Body)
			collectDefinitions(instruction.Step)
			collectDefinitions(instruction.Catch)
			collectDefinitions(instruction.Finally)
		}
	}
	collectDefinitions(segment)
	var collect func([]ir.Instruction)
	collect = func(instructions []ir.Instruction) {
		for _, instruction := range instructions {
			for _, arg := range instruction.Args {
				if arg != "" && valueTypes[arg] != "" && !params[arg] && !defined[arg] {
					seen := false
					for _, current := range result {
						if current == arg {
							seen = true
							break
						}
					}
					if !seen {
						result = append(result, arg)
					}
				}
			}
			collect(instruction.Then)
			collect(instruction.Else)
			collect(instruction.Cond)
			collect(instruction.Body)
			collect(instruction.Step)
			collect(instruction.Catch)
			collect(instruction.Finally)
		}
	}
	collect(segment)
	return result
}

func asyncCapturedParameters(names []string, types map[string]ir.Type, promiseName string) []ir.Parameter {
	result := make([]ir.Parameter, 0, len(names))
	for _, name := range names {
		typ := types[name]
		if name == promiseName {
			typ = ir.Type("object:Promise")
		}
		if typ == "" {
			typ = ir.TypeUnknown
		}
		result = append(result, ir.Parameter{Name: name, Type: typ})
	}
	return result
}
