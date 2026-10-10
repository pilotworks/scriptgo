package lowering

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// The shape lowerings (linear awaits, one structured try, branch, or loop)
// each own a single suspending unit. lowerAsyncSequence composes them for a
// body with several units in sequence (`await a; try { await b } ...`): the
// body is split after its first unit, the remainder becomes a separate async
// function that settles the caller's promise, and every exit of the first
// unit that would fall off its end tail-calls that remainder instead.

// asyncTailMarker marks the synthetic return that ends the first unit; every
// async return rewriter expands it into the tail call (asyncTailCall).
const asyncTailMarker = "async.tail"

// Parameter names that carry the caller's promise and frame into a
// remainder function, which adopts them instead of creating its own.
const (
	asyncOuterPromise = "__async_outer_promise"
	asyncOuterFrame   = "__async_outer_frame"
)

func isAsyncTailReturn(instruction ir.Instruction) bool {
	return instruction.Op == ir.OpReturn && instruction.Value == asyncTailMarker
}

// asyncTailCall expands a tail marker on a path that returns returnType: call
// the remainder with this function's promise and frame, then return without
// settling (the remainder settles). An entry path returns its promise; a
// resume callback returns nothing.
func asyncTailCall(marker ir.Instruction, returnType ir.Type, promiseName, frameName string) []ir.Instruction {
	ret := ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid, Span: marker.Span}
	if returnType != "" && returnType != ir.TypeVoid {
		ret = ir.Instruction{Op: ir.OpReturn, Type: returnType, Args: []string{promiseName}, Span: marker.Span}
	}
	return []ir.Instruction{
		{Op: ir.OpCall, Type: ir.Type("object:Promise"), Result: marker.Result, Callee: marker.Callee, Args: append(append([]string{}, marker.Args...), promiseName, frameName), Span: marker.Span},
		ret,
	}
}

// lowerAsyncSequence lowers a body whose suspending units no single shape
// lowering accepts. It reports false when the body has fewer than two units.
func lowerAsyncSequence(path string, statement frontend.SyntaxStatement, lowered ir.Function, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) (ir.Function, bool, error) {
	split := asyncSequenceSplit(lowered.Body)
	if split <= 0 || split >= len(lowered.Body) {
		return ir.Function{}, false, nil
	}
	head, rest := lowered.Body[:split], lowered.Body[split:]

	asyncLowerCounter++
	restName := fmt.Sprintf("%s$rest%d", statement.Name, asyncLowerCounter)
	valueTypes := asyncValueTypes(lowered)
	// The caller's promise and frame are passed separately (asyncTailCall).
	seen := map[string]bool{asyncOuterPromise: true, asyncOuterFrame: true}
	var inputs []string
	for _, name := range asyncCaptures(rest, valueTypes, nil) {
		if !seen[name] {
			seen[name] = true
			inputs = append(inputs, name)
		}
	}
	for _, parameter := range lowered.Parameters {
		if !seen[parameter.Name] {
			seen[parameter.Name] = true
			inputs = append(inputs, parameter.Name)
		}
	}
	for _, local := range lowered.Locals {
		if !seen[local.Name] {
			seen[local.Name] = true
			inputs = append(inputs, local.Name)
		}
	}
	restParameters := make([]ir.Parameter, 0, len(inputs)+2)
	for _, name := range inputs {
		typ := valueTypes[name]
		if typ == "" {
			typ = ir.TypeUnknown
		}
		restParameters = append(restParameters, ir.Parameter{Name: name, Type: typ})
	}
	restParameters = append(restParameters,
		ir.Parameter{Name: asyncOuterPromise, Type: ir.Type("object:Promise")},
		ir.Parameter{Name: asyncOuterFrame, Type: ir.TypePointer},
	)

	restStatement := statement
	restStatement.Name = restName
	restFunction, err := lowerAsyncFunctionFromLowered(path, restStatement, ir.Function{
		Name:       restName,
		Span:       lowered.Span,
		Parameters: restParameters,
		Locals:     lowered.Locals,
		ReturnType: lowered.ReturnType,
		Body:       rest,
	}, shapes, signatures)
	if err != nil {
		return ir.Function{}, false, err
	}
	if !adoptOuterAsyncPromise(&restFunction) {
		return ir.Function{}, false, fmt.Errorf("async remainder %q does not create its promise", restName)
	}
	extraFunctions = append(extraFunctions, restFunction)

	span := lowered.Span
	if len(rest) > 0 {
		span = rest[0].Span
	}
	headLowered := lowered
	headLowered.Body = append(append([]ir.Instruction{}, head...), ir.Instruction{
		Op:     ir.OpReturn,
		Value:  asyncTailMarker,
		Result: restName + ".promise",
		Callee: restName,
		Args:   inputs,
		Span:   span,
	})
	headFunction, err := lowerAsyncFunctionFromLowered(path, statement, headLowered, shapes, signatures)
	if err != nil {
		return ir.Function{}, false, err
	}
	return headFunction, true, nil
}

// asyncSequenceSplit is the index after the body's first suspending unit: a
// run of top-level awaits ends at the first structured await, and a
// structured await ends at the next instruction that awaits. It is -1 when
// the body has a single unit.
func asyncSequenceSplit(body []ir.Instruction) int {
	first := -1
	for index, instruction := range body {
		if !hasAwait(instruction) {
			continue
		}
		if first < 0 {
			first = index
			continue
		}
		if isDirectAwait(body[first]) {
			if !isDirectAwait(instruction) {
				return index
			}
			continue
		}
		return index
	}
	return -1
}

func isDirectAwait(instruction ir.Instruction) bool {
	return instruction.Op == ir.OpCall && strings.HasPrefix(instruction.Callee, "__async.await")
}

// adoptOuterAsyncPromise makes a lowered remainder settle its caller's
// promise and frame: the remainder's own promise and frame creation become
// copies of the asyncOuterPromise and asyncOuterFrame parameters.
func adoptOuterAsyncPromise(function *ir.Function) bool {
	promise, frame := false, false
	for index, instruction := range function.Body {
		if instruction.Op != ir.OpCall {
			continue
		}
		switch {
		case !promise && instruction.Callee == "__async.promise_create":
			function.Body[index] = ir.Instruction{Op: ir.OpAssign, Type: ir.Type("object:Promise"), Result: instruction.Result, Args: []string{asyncOuterPromise}, Span: instruction.Span}
			promise = true
		case !frame && instruction.Callee == "__async.frame_new":
			function.Body[index] = ir.Instruction{Op: ir.OpAssign, Type: ir.TypePointer, Result: instruction.Result, Args: []string{asyncOuterFrame}, Span: instruction.Span}
			frame = true
		}
	}
	return promise && frame
}
