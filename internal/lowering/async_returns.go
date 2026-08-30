package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

const asyncResolvedReturnTypeEnvKey = "__async.resolved_return_type"

// asyncResolvedReturnType preserves the declared Promise payload before the
// generic name is mangled into an IR function return type.
func asyncResolvedReturnType(typeName string) (ir.Type, bool) {
	typeName = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(typeName), "object:"))
	if !strings.HasPrefix(typeName, "Promise<") || !strings.HasSuffix(typeName, ">") {
		return "", false
	}
	args := splitTypeArguments(typeName[len("Promise<") : len(typeName)-1])
	if len(args) != 1 {
		return "", false
	}
	resolved := toIRType(args[0])
	return resolved, resolved != ""
}

// appendResolvedPromiseConstant creates a fulfilled Promise for an async
// return path. The value keeps its semantic type so undefined and null do not
// get mistaken for the number zero at the native ABI boundary.
func appendResolvedPromiseConstant(function *ir.Function, counter *int, valueType ir.Type, value string, span ir.SourceSpan) string {
	valueName := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpConst,
		Type:   valueType,
		Result: valueName,
		Value:  value,
		Span:   span,
	})
	promiseName := nextTemp(counter)
	function.Body = append(function.Body, ir.Instruction{
		Op:     ir.OpCall,
		Type:   ir.Type("object:Promise"),
		Result: promiseName,
		Callee: "__async.promise_resolve",
		Args:   []string{valueName},
		Span:   span,
	})
	return promiseName
}

func appendResolvedPromiseUndefined(function *ir.Function, counter *int, span ir.SourceSpan) string {
	return appendResolvedPromiseConstant(function, counter, ir.TypeVoid, "undefined", span)
}

func appendResolvedPromiseNull(function *ir.Function, counter *int, span ir.SourceSpan) string {
	return appendResolvedPromiseConstant(function, counter, ir.TypePointer, "null", span)
}
