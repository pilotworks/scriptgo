package lowering

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// restParameterIndex is the index of target's array-typed rest parameter, or
// -1 when it has none. key is the mangled callee name in restParamsIndex.
func restParameterIndex(target ir.Function, key string) int {
	if len(target.Parameters) == 0 || !(restParamsIndex[key] || restParamsIndex[strings.Split(key, "__")[0]]) {
		return -1
	}
	last := len(target.Parameters) - 1
	if !strings.HasSuffix(string(target.Parameters[last].Type), "[]") {
		return -1
	}
	return last
}

// spreadRestArguments gathers the arguments that fill a rest parameter into
// one array literal when any of them is a spread (`f(a, ...xs, b)`), so the
// array literal's spread lowering builds the rest array; with always set it
// gathers them regardless. restIndex is the rest parameter's index and offset
// the number of leading parameters (the receiver) not written as arguments;
// span locates the call. packed reports whether the rest arguments were
// gathered. A spread of a fixed-length tuple variable expands into its
// elements first; any other spread outside the rest arguments is rejected.
func spreadRestArguments(arguments []*frontend.SyntaxExpression, span frontend.SourceSpan, target ir.Function, restIndex, offset int, always bool) (rewritten []*frontend.SyntaxExpression, packed bool, err error) {
	arguments = expandTupleSpreads(arguments)
	restStart := len(arguments)
	if restIndex >= 0 {
		restStart = restIndex - offset
	}
	for index, argument := range arguments {
		if argument.Kind != "spread" {
			continue
		}
		if index < restStart {
			return nil, false, fmt.Errorf("spread argument for a non-rest parameter of %q is not supported in the native subset", target.Name)
		}
		packed = true
	}
	if restIndex < 0 || restStart > len(arguments) || !(packed || always) {
		return arguments, false, nil
	}
	rest := arguments[restStart:]
	if len(rest) > 0 {
		last := rest[len(rest)-1].Span
		span = frontend.SourceSpan{Start: rest[0].Span.Start, Length: last.Start + last.Length - rest[0].Span.Start}
	}
	rewritten = append(append([]*frontend.SyntaxExpression(nil), arguments[:restStart]...), &frontend.SyntaxExpression{
		Span:         span,
		Kind:         "array",
		Arguments:    rest,
		InferredType: string(target.Parameters[restIndex].Type),
	})
	return rewritten, true, nil
}

// emptyRestArray is the rest argument of a call that passes no rest
// arguments and leaves an earlier optional parameter out.
func emptyRestArray(path string, span frontend.SourceSpan, restType ir.Type, function *ir.Function, counter *int) string {
	array := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{Op: ir.OpArray, Type: restType, Result: array, Span: toIRSpan(path, span)})
	return array
}

// expandTupleSpreads replaces each spread of a fixed-length tuple variable by
// its elements (`f(...pair)` is `f(pair[0], pair[1])`), so they can fill
// ordinary parameters. Reading a variable has no effects, so indexing it per
// element matches evaluating the spread once. Other spreads are kept.
func expandTupleSpreads(arguments []*frontend.SyntaxExpression) []*frontend.SyntaxExpression {
	var expanded []*frontend.SyntaxExpression
	for index, argument := range arguments {
		elements, ok := fixedTupleElementTypes(argument)
		if !ok {
			if expanded != nil {
				expanded = append(expanded, argument)
			}
			continue
		}
		if expanded == nil {
			expanded = append([]*frontend.SyntaxExpression(nil), arguments[:index]...)
		}
		for position, elementType := range elements {
			expanded = append(expanded, &frontend.SyntaxExpression{
				Span:         argument.Span,
				Kind:         "index",
				Left:         argument.Left,
				Right:        &frontend.SyntaxExpression{Span: argument.Span, Kind: "number", Text: strconv.Itoa(position), InferredType: "number"},
				InferredType: elementType,
			})
		}
	}
	if expanded == nil {
		return arguments
	}
	return expanded
}

// fixedTupleElementTypes is the element types of a spread of a tuple-typed
// variable whose length is fixed (no optional or rest elements).
func fixedTupleElementTypes(argument *frontend.SyntaxExpression) ([]string, bool) {
	if argument.Kind != "spread" || argument.Left == nil || argument.Left.Kind != "identifier" {
		return nil, false
	}
	typ := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(argument.Left.InferredType), "readonly "))
	if alias := typeAliasesIndex[typ]; alias != "" {
		typ = strings.TrimSpace(strings.TrimPrefix(alias, "readonly "))
	}
	if !strings.HasPrefix(typ, "[") || !strings.HasSuffix(typ, "]") || strings.HasSuffix(typ, "[]") {
		return nil, false
	}
	inner := strings.TrimSpace(typ[1 : len(typ)-1])
	if inner == "" {
		return nil, false
	}
	var elements []string
	for _, part := range splitTopLevel(inner) {
		element := strings.TrimSpace(part)
		if label, rest, labeled := strings.Cut(element, ":"); labeled && isIdentifierName(strings.TrimSuffix(strings.TrimSpace(label), "?")) {
			if strings.HasSuffix(strings.TrimSpace(label), "?") {
				return nil, false
			}
			element = strings.TrimSpace(rest)
		}
		if strings.HasPrefix(element, "...") || strings.HasSuffix(element, "?") {
			return nil, false
		}
		elements = append(elements, element)
	}
	return elements, true
}

// isIdentifierName reports whether text is an ASCII identifier, the shape of
// a tuple element label.
func isIdentifierName(text string) bool {
	if text == "" {
		return false
	}
	for index, r := range text {
		if r == '_' || r == '$' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (index > 0 && r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}
