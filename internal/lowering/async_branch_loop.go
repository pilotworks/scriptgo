package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerStructuredAsyncBranchLoop composes an async loop state machine with a
// surrounding branch. The branch condition and its side effects run before
// selecting the loop runner; the other branch settles the same outer promise.
func lowerStructuredAsyncBranchLoop(path string, statement typescriptgo.SyntaxStatement, lowered ir.Function, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (ir.Function, bool, error) {
	branchIndex := -1
	var branch ir.Instruction
	for i, instruction := range lowered.Body {
		if !hasAwait(instruction) {
			continue
		}
		if branchIndex != -1 || instruction.Op != ir.OpIf || hasNestedAwait(instruction.Cond) || hasNestedAwait(instruction.Else) {
			return ir.Function{}, false, nil
		}
		branchIndex, branch = i, instruction
	}
	if branchIndex < 0 || len(branch.Args) != 1 {
		return ir.Function{}, false, nil
	}
	loopIndex, _, ok := findAsyncLoop(branch.Then)
	if !ok || loopIndex < 0 || hasNestedAwait(branch.Then[:loopIndex]) {
		return ir.Function{}, false, nil
	}
	for _, instruction := range branch.Then[loopIndex+1:] {
		if hasAwait(instruction) {
			return ir.Function{}, false, nil
		}
	}

	// Reuse the loop lowering implementation by presenting the selected branch
	// as a temporary top-level async body. Its generated state functions and
	// captures remain valid when the entry runner is guarded below.
	synthetic := lowered
	synthetic.Body = append([]ir.Instruction{}, branch.Then...)
	loopFunction, ok, err := lowerStructuredAsyncLoop(path, statement, synthetic, shapes, signatures)
	if err != nil || !ok {
		return ir.Function{}, false, err
	}
	if len(loopFunction.Body) < 3 {
		return ir.Function{}, false, nil
	}
	promiseName := loopFunction.Body[0].Result
	frameName := loopFunction.Body[1].Result
	if promiseName == "" || frameName == "" {
		return ir.Function{}, false, fmt.Errorf("async branch loop lost promise/frame allocation")
	}

	result := loopFunction
	result.Name = statement.Name
	result.Span = toIRSpan(path, statement.Span)
	entry := append([]ir.Instruction{}, loopFunction.Body[2:]...)
	if len(entry) > 0 && entry[len(entry)-1].Op == ir.OpReturn {
		entry = entry[:len(entry)-1]
	}
	entry = rewriteAsyncReturnsVoid(entry, promiseName, frameName, result.Span)

	fallback := append([]ir.Instruction{}, branch.Else...)
	fallback = append(fallback, lowered.Body[branchIndex+1:]...)
	fallback = rewriteAsyncReturnsVoid(fallback, promiseName, frameName, result.Span)
	if len(fallback) == 0 || !instructionListTerminates(fallback) {
		fallback = append(fallback, settleAsyncVoid(promiseName, frameName, result.Span)...)
	}
	if len(fallback) > 0 && fallback[len(fallback)-1].Op == ir.OpReturn {
		fallback = fallback[:len(fallback)-1]
	}

	prefix := rewriteAsyncReturns(lowered.Body[:branchIndex], promiseName, frameName, result.Span)
	result.Body = append([]ir.Instruction{}, loopFunction.Body[:2]...)
	result.Body = append(result.Body, prefix...)
	result.Body = append(result.Body, branch.Cond...)
	result.Body = append(result.Body, ir.Instruction{Op: ir.OpIf, Type: ir.TypeVoid, Args: branch.Args, Then: entry, Else: fallback, Span: branch.Span})
	result.Body = append(result.Body, ir.Instruction{Op: ir.OpReturn, Type: result.ReturnType, Args: []string{promiseName}, Span: result.Span})
	return result, true, nil
}
