package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// lowerTopLevelAsyncSequence composes top-level suspension-bearing statements
// into stages. Each stage executes its synchronous prefix immediately when
// entered, and the next stage is entered by a Promise reaction.
func lowerTopLevelAsyncSequence(path string, main ir.Function, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) ([]ir.Function, bool, error) {
	var boundaries []int
	for index, instruction := range main.Body {
		if hasAwait(instruction) {
			boundaries = append(boundaries, index)
		}
	}
	if len(boundaries) == 0 {
		// Nested suspension must not fall back to a synchronous blocking await.
		// Feed the complete module body through the structured async lowering;
		// unsupported shapes are reported to the caller instead of being hidden.
		stageName := "__top_level_async_main"
		stageStatement := typescriptgo.SyntaxStatement{Name: stageName, Kind: "async_function", IsAsync: true, Type: "Promise<void>", InferredType: "Promise<void>"}
		stage := main
		stage.Name = stageName
		stage.Body = append([]ir.Instruction{}, main.Body...)
		lowered, err := lowerAsyncFunctionFromLowered(path, stageStatement, stage, shapes, signatures)
		if err != nil {
			return nil, false, err
		}
		signatures[stageName] = lowered
		main.Body = []ir.Instruction{{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: "__top_level_async_result", Callee: stageName, Span: main.Span}}
		return []ir.Function{lowered, main}, true, nil
	}
	starts := make([]int, 0, len(boundaries))
	start := 0
	for _, boundary := range boundaries {
		starts = append(starts, start)
		start = boundary + 1
	}
	starts = append(starts, start)
	stages := make([]ir.Function, 0, len(boundaries))
	for index, boundary := range boundaries {
		end := boundary + 1
		if index+1 == len(boundaries) {
			end = len(main.Body)
		}
		stageName := fmt.Sprintf("__top_level_async_stage_%d", index)
		stageStatement := typescriptgo.SyntaxStatement{Name: stageName, Kind: "async_function", IsAsync: true, Type: "Promise<void>", InferredType: "Promise<void>"}
		stage := main
		stage.Name = stageName
		// Module bindings are emitted as module globals; keeping the wrapper's
		// local declarations here would create a fresh stack cell per stage.
		stage.Locals = nil
		stage.Body = append([]ir.Instruction{}, main.Body[starts[index]:end]...)
		lowered, err := lowerAsyncFunctionFromLowered(path, stageStatement, stage, shapes, signatures)
		if err != nil {
			return nil, false, err
		}
		lowered.Locals = nil
		stages = append(stages, lowered)
		signatures[stageName] = lowered
	}
	continuations := make([]ir.Function, len(stages)-1)
	for index := range continuations {
		continuations[index] = ir.Function{Name: fmt.Sprintf("__top_level_async_continue_%d", index), Parameters: asyncResumeParameters(), ReturnType: ir.Type("object:Promise")}
	}
	for index := len(continuations) - 1; index >= 0; index-- {
		nextPromise := fmt.Sprintf("__top_level_stage_promise_%d", index+1)
		body := []ir.Instruction{{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: nextPromise, Callee: stages[index+1].Name}}
		if index+1 < len(continuations) {
			closure := fmt.Sprintf("__top_level_continue_closure_%d", index+1)
			chained := fmt.Sprintf("__top_level_chained_promise_%d", index+1)
			body = append(body, ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: closure, Callee: continuations[index+1].Name}, ir.Instruction{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: chained, Callee: "__async.promise_then", Args: []string{nextPromise, closure}, Value: "object:Promise"}, ir.Instruction{Op: ir.OpReturn, Type: ir.Type("object:Promise"), Args: []string{chained}})
		} else {
			body = append(body, ir.Instruction{Op: ir.OpReturn, Type: ir.Type("object:Promise"), Args: []string{nextPromise}})
		}
		continuations[index].Body = body
	}
	entryPromise := "__top_level_stage_promise_0"
	mainBody := []ir.Instruction{{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: entryPromise, Callee: stages[0].Name}}
	if len(continuations) > 0 {
		closure := "__top_level_continue_closure_0"
		mainBody = append(mainBody, ir.Instruction{Op: ir.OpClosure, Type: ir.TypeClosure, Result: closure, Callee: continuations[0].Name}, ir.Instruction{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: "__top_level_chain", Callee: "__async.promise_then", Args: []string{entryPromise, closure}, Value: "object:Promise"})
	}
	mainBody = append(mainBody, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid})
	main.Body = mainBody
	functions := append(stages, main)
	extraFunctions = append(extraFunctions, continuations...)
	return functions, true, nil
}
