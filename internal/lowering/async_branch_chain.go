package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

type asyncBranchLeaf struct {
	body   []ir.Instruction
	await  ir.Instruction
	suffix []ir.Instruction
}

// lowerStructuredAsyncBranchChain lowers an else-if tree with one direct
// await per leaf. It keeps the original branch tree in the entry function and
// gives every suspension point its own fulfilled/rejected continuation.
func lowerStructuredAsyncBranchChain(path string, statement typescriptgo.SyntaxStatement, lowered ir.Function, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function, index int, root ir.Instruction) (ir.Function, bool, error) {
	leaves, ok := collectAsyncBranchLeaves(root)
	if !ok || len(leaves) < 2 {
		return ir.Function{}, false, nil
	}
	awaits := make([]ir.Instruction, 0, len(leaves))
	for _, leaf := range leaves {
		if len(leaf.await.Args) == 0 {
			continue
		}
		awaits = append(awaits, leaf.await)
	}
	if len(awaits) == 0 {
		return ir.Function{}, false, nil
	}

	asyncLowerCounter++
	base := fmt.Sprintf("__async_branch_chain_%d", asyncLowerCounter)
	promiseName, frameName := base+".promise", base+".frame"
	result := ir.Function{Name: statement.Name, Span: toIRSpan(path, statement.Span), Parameters: lowered.Parameters, Locals: lowered.Locals, ReturnType: ir.Type("object:Promise"), Async: true, EntryBlock: "async.entry"}
	result.Body = append(result.Body,
		ir.Instruction{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: promiseName, Callee: "__async.promise_create", Span: result.Span},
		ir.Instruction{Op: ir.OpCall, Type: ir.TypePointer, Result: frameName, Callee: "__async.frame_new", Value: fmt.Sprintf("%d", len(lowered.Parameters)+len(lowered.Captured)+len(lowered.Locals)+len(awaits)+2), Span: result.Span},
	)
	result.AsyncFrame = buildAsyncFrame(result, lowered, awaits)
	result.Blocks = buildAsyncBlocks(awaits, result.Span)
	result.Body = append(result.Body, rewriteAsyncReturns(lowered.Body[:index], promiseName, frameName, result.Span)...)

	valueTypes := asyncValueTypes(lowered)
	valueTypes[promiseName], valueTypes[frameName] = ir.Type("object:Promise"), ir.TypePointer
	scope := append([]ir.Instruction{}, root.Cond...)
	scope = append(scope, root.Then...)
	scope = append(scope, root.Else...)
	scope = append(scope, lowered.Body[index+1:]...)
	captures := asyncCaptures(scope, valueTypes, lowered.Parameters)
	seen := make(map[string]bool, len(captures))
	for _, name := range captures {
		seen[name] = true
	}
	addCapture := func(name string) {
		if name != "" && !seen[name] {
			captures = append(captures, name)
			seen[name] = true
		}
	}
	for _, await := range awaits {
		captures = removeCapture(captures, await.Result)
	}
	addCapture(promiseName)
	addCapture(frameName)
	for _, parameter := range lowered.Parameters {
		addCapture(parameter.Name)
	}
	for _, local := range lowered.Locals {
		addCapture(local.Name)
	}

	tail := rewriteAsyncReturnsVoid(lowered.Body[index+1:], promiseName, frameName, result.Span)
	if len(tail) == 0 || !instructionListTerminates(tail) {
		tail = append(tail, settleAsyncVoid(promiseName, frameName, result.Span)...)
	}

	continuationByAwait := make(map[string]int, len(awaits))
	for i, await := range awaits {
		continuationByAwait[await.Result] = i
	}
	continuationBodies := make([][]ir.Instruction, len(awaits))
	for _, leaf := range leaves {
		if len(leaf.await.Args) == 0 {
			continue
		}
		continuationBodies[continuationByAwait[leaf.await.Result]] = append([]ir.Instruction{}, leaf.suffix...)
	}

	fulfilledNames := make([]string, len(awaits))
	rejectedNames := make([]string, len(awaits))
	for i, await := range awaits {
		fulfilledNames[i], rejectedNames[i] = fmt.Sprintf("%s.fulfilled.%d", base, i), fmt.Sprintf("%s.rejected.%d", base, i)
		fulfilled := ir.Function{Name: fulfilledNames[i], Span: await.Span, Parameters: asyncResumeParameters(), Captured: asyncCapturedParameters(captures, valueTypes, promiseName), ReturnType: ir.TypeVoid}
		fulfilled.Body = append(fulfilled.Body, ir.Instruction{Op: ir.OpCheckedCast, Type: await.Type, Result: await.Result, Args: []string{"__resume_raw"}, Span: await.Span})
		fulfilled.Body = append(fulfilled.Body, rewriteAsyncReturnsVoid(continuationBodies[i], promiseName, frameName, result.Span)...)
		if !instructionListTerminates(fulfilled.Body) {
			fulfilled.Body = append(fulfilled.Body, tail...)
		}
		rejected := ir.Function{Name: rejectedNames[i], Span: await.Span, Parameters: asyncResumeParameters(), Captured: asyncCapturedParameters(captures, valueTypes, promiseName), ReturnType: ir.TypeVoid, Body: append([]ir.Instruction{}, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_reject_existing", Args: []string{promiseName, "__resume_raw"}, Span: await.Span}, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: await.Span}, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: await.Span})}
		extraFunctions = append(extraFunctions, fulfilled, rejected)
	}

	leafIndex := 0
	var rewrite func(ir.Instruction) ir.Instruction
	rewrite = func(node ir.Instruction) ir.Instruction {
		if node.Op != ir.OpIf {
			return node
		}
		node.Cond = rewriteList(node.Cond, rewrite)
		if segments, nodeAwaits := splitLinearAsyncBody(node.Then); len(nodeAwaits) == 1 {
			node.Then = append(segments[0], branchSchedule(base, continuationByAwait[nodeAwaits[0].Result], nodeAwaits[0], fulfilledNames, rejectedNames, captures)...)
		} else if len(nodeAwaits) == 0 && !hasNestedAwait(node.Then) {
			node.Then = settleBranchLeaf(node.Then, promiseName, frameName, result.Span)
		} else {
			node.Then = rewriteList(node.Then, rewrite)
		}
		nestedIndex := -1
		for i, nested := range node.Else {
			if hasAwait(nested) {
				nestedIndex = i
				break
			}
		}
		if nestedIndex == len(node.Else)-1 && nestedIndex >= 0 && node.Else[nestedIndex].Op == ir.OpIf {
			node.Else = rewriteList(node.Else, rewrite)
		} else if segments, nodeAwaits := splitLinearAsyncBody(node.Else); len(nodeAwaits) == 1 {
			node.Else = append(segments[0], branchSchedule(base, continuationByAwait[nodeAwaits[0].Result], nodeAwaits[0], fulfilledNames, rejectedNames, captures)...)
		} else if len(nodeAwaits) == 0 && !hasNestedAwait(node.Else) {
			node.Else = settleBranchLeaf(node.Else, promiseName, frameName, result.Span)
		} else {
			node.Else = rewriteList(node.Else, rewrite)
		}
		leafIndex++
		return node
	}
	result.Body = append(result.Body, rewrite(root))
	result.Body = append(result.Body, ir.Instruction{Op: ir.OpReturn, Type: result.ReturnType, Args: []string{promiseName}, Span: result.Span})
	return result, true, nil
}

func settleBranchLeaf(body []ir.Instruction, promiseName, frameName string, span ir.SourceSpan) []ir.Instruction {
	// A branch leaf can execute in the async entry function when it contains no
	// suspension. Keep the entry's Promise return type distinct from the void
	// return used by resume callbacks.
	result := rewriteAsyncEntryReturns(body, promiseName, frameName, ir.Type("object:Promise"), span)
	if len(result) == 0 || !instructionListTerminates(result) {
		result = append(result, settleAsyncEntry(promiseName, frameName, ir.Type("object:Promise"), span)...)
	}
	return result
}

func collectAsyncBranchLeaves(node ir.Instruction) ([]asyncBranchLeaf, bool) {
	if node.Op != ir.OpIf || len(node.Then) == 0 {
		return nil, false
	}
	var leaves []asyncBranchLeaf
	collect := func(body []ir.Instruction) bool {
		segments, awaits := splitLinearAsyncBody(body)
		if len(awaits) == 1 {
			leaves = append(leaves, asyncBranchLeaf{body: body, await: awaits[0], suffix: segments[1]})
			return true
		}
		return len(awaits) == 0 && !hasNestedAwait(body)
	}
	if !collect(node.Then) {
		return nil, false
	}
	nestedIndex := -1
	for i, nested := range node.Else {
		if hasAwait(nested) {
			nestedIndex = i
			break
		}
	}
	if nestedIndex == len(node.Else)-1 && nestedIndex >= 0 && node.Else[nestedIndex].Op == ir.OpIf && hasAwait(node.Else[nestedIndex]) {
		nested, ok := collectAsyncBranchLeaves(node.Else[nestedIndex])
		if !ok {
			return nil, false
		}
		leaves = append(leaves, nested...)
		return leaves, true
	}
	if !collect(node.Else) {
		return nil, false
	}
	return leaves, true
}

func rewriteList(list []ir.Instruction, rewrite func(ir.Instruction) ir.Instruction) []ir.Instruction {
	result := make([]ir.Instruction, len(list))
	for i, node := range list {
		result[i] = rewrite(node)
	}
	return result
}

func branchSchedule(base string, index int, await ir.Instruction, fulfilled, rejected, captures []string) []ir.Instruction {
	return []ir.Instruction{
		{Op: ir.OpClosure, Type: ir.TypeClosure, Result: fmt.Sprintf("%s.fulfilled.closure.%d", base, index), Callee: fulfilled[index], Args: captures, Span: await.Span},
		{Op: ir.OpClosure, Type: ir.TypeClosure, Result: fmt.Sprintf("%s.rejected.closure.%d", base, index), Callee: rejected[index], Args: captures, Span: await.Span},
		{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_schedule_resume_pair", Args: []string{await.Args[0], fmt.Sprintf("%s.fulfilled.closure.%d", base, index), fmt.Sprintf("%s.rejected.closure.%d", base, index)}, Span: await.Span},
	}
}

func settleWithoutReturn(promiseName, frameName string, span ir.SourceSpan) []ir.Instruction {
	return []ir.Instruction{{Op: ir.OpConst, Type: ir.TypeUnknown, Result: "__async.undefined", Value: "undefined", Span: span}, {Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.promise_resolve_existing", Args: []string{promiseName, "__async.undefined"}, Span: span}, {Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__async.frame_release", Args: []string{frameName}, Span: span}}
}
