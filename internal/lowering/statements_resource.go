package lowering

import (
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

type usingResourceInfo struct {
	Name    string
	Type    ir.Type
	IsAwait bool
	Span    typescriptgo.SourceSpan
}

var usingScopeStack [][]usingResourceInfo

func pushUsingScope() {
	usingScopeStack = append(usingScopeStack, []usingResourceInfo{})
}

func recordUsingResource(name string, resType ir.Type, isAwait bool, span typescriptgo.SourceSpan) {
	if len(usingScopeStack) == 0 {
		pushUsingScope()
	}
	topIdx := len(usingScopeStack) - 1
	usingScopeStack[topIdx] = append(usingScopeStack[topIdx], usingResourceInfo{
		Name:    name,
		Type:    resType,
		IsAwait: isAwait,
		Span:    span,
	})
}

func emitDisposeResource(path string, res usingResourceInfo, function *ir.Function, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) {
	clsName := strings.TrimPrefix(string(res.Type), "object:")
	candidates := []string{
		clsName + "_Symbol.dispose",
		clsName + ".Symbol.dispose",
		clsName + "_Symbol_dispose",
		clsName + "_dispose",
		clsName + "_close",
	}
	if res.IsAwait {
		candidates = append([]string{
			clsName + "_Symbol.asyncDispose",
			clsName + ".Symbol.asyncDispose",
			clsName + "_Symbol_asyncDispose",
			clsName + "_asyncDispose",
		}, candidates...)
	}

	for _, targetCallee := range candidates {
		if _, ok := signatures[targetCallee]; ok {
			function.Body = append(function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeVoid,
				Callee: targetCallee,
				Args:   []string{res.Name},
				Span:   toIRSpan(path, res.Span),
			})
			return
		}
	}
}

func popAndEmitUsingScope(path string, function *ir.Function, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) {
	if len(usingScopeStack) == 0 {
		return
	}
	topIdx := len(usingScopeStack) - 1
	scope := usingScopeStack[topIdx]
	usingScopeStack = usingScopeStack[:topIdx]

	for i := len(scope) - 1; i >= 0; i-- {
		emitDisposeResource(path, scope[i], function, counter, shapes, signatures)
	}
}

func emitAllActiveUsingScopes(path string, function *ir.Function, counter *int, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) {
	for sIdx := len(usingScopeStack) - 1; sIdx >= 0; sIdx-- {
		scope := usingScopeStack[sIdx]
		for i := len(scope) - 1; i >= 0; i-- {
			emitDisposeResource(path, scope[i], function, counter, shapes, signatures)
		}
	}
}
