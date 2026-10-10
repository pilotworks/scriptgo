package lowering

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func ensureFunctionClosureTrampoline(path string, sig ir.Function, shapes map[string]ir.ObjectShape, signatures map[string]ir.Function) string {
	if strings.HasPrefix(sig.Name, "__closure_") {
		return sig.Name
	}
	trampolineName := "__closure_trampoline_" + sig.Name
	if _, exists := signatures[trampolineName]; exists {
		return trampolineName
	}

	trampolineFn := ir.Function{
		Name:       trampolineName,
		ReturnType: sig.ReturnType,
		Span:       sig.Span,
	}

	// Closure ABI parameters: (__env_ctx: ptr, param0$raw: unknown, param1$raw: unknown, param2$raw: unknown, param3$raw: unknown)
	trampolineFn.Parameters = append(trampolineFn.Parameters, ir.Parameter{
		Name: "__env_ctx",
		Type: ir.TypePointer,
	})

	var callArgs []string
	counter := 0
	// Call sites substitute defaults for direct calls; a closure call cannot,
	// so the trampoline applies them. Parameters keep their source names
	// then, because an initializer may refer to earlier parameters.
	defaults := defaultParamsIndex[sig.Name]
	trampolineEnv := map[string]ir.Type{}
	trampolineArgName := func(i int, param ir.Parameter) string {
		if len(defaults) > 0 && param.Name != "" {
			return param.Name
		}
		return fmt.Sprintf("arg_%d", i)
	}
	// Arguments past the fourth, and a rest parameter, are read first,
	// before any default initializer can make another closure call. A
	// function's rest parameter collects the closure call's arguments.
	restIndex := restParameterIndex(sig, sig.Name)
	for i := closureRawParameterCount; i < len(sig.Parameters) && i != restIndex; i++ {
		trampolineFn.Body = append(trampolineFn.Body, closureMoreArgument(trampolineArgName(i, sig.Parameters[i])+"$raw", i-closureRawParameterCount, sig.Span)...)
	}
	if restIndex >= 0 {
		trampolineFn.Body = append(trampolineFn.Body, closureRestArgument(trampolineArgName(restIndex, sig.Parameters[restIndex]), sig.Parameters[restIndex].Type, restIndex, sig.Span)...)
	} else if len(sig.Parameters) > closureRawParameterCount {
		trampolineFn.Body = append(trampolineFn.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Callee: "__closure.more_done", Span: sig.Span})
	}
	for i, param := range sig.Parameters {
		unboxedName := trampolineArgName(i, param)
		rawName := unboxedName + "$raw"
		if i < closureRawParameterCount {
			trampolineFn.Parameters = append(trampolineFn.Parameters, ir.Parameter{
				Name: rawName,
				Type: ir.TypeUnknown,
			})
		}
		if i == restIndex {
			trampolineEnv[unboxedName] = param.Type
			callArgs = append(callArgs, unboxedName)
			counter++
			continue
		}
		trampolineFn.Body = append(trampolineFn.Body, ir.Instruction{
			Op:     ir.OpCheckedCast,
			Type:   param.Type,
			Result: unboxedName,
			Args:   []string{rawName},
			Span:   sig.Span,
		})
		trampolineEnv[rawName] = ir.TypeUnknown
		trampolineEnv[unboxedName] = param.Type
		if initializer := defaults[i]; initializer != nil && initializer.Kind != "undefined" {
			statement := parameterDefaultStatement(frontend.SyntaxParameter{Span: initializer.Span, Name: unboxedName, Initializer: initializer})
			bodyLen := len(trampolineFn.Body)
			if err := lowerStatement(path, statement, &trampolineFn, trampolineEnv, &counter, shapes, signatures); err != nil {
				// An initializer that only lowers in its declaring scope
				// (it reads a local of an enclosing function) stays a
				// call-site default; closure calls then pass undefined.
				trampolineFn.Body = trampolineFn.Body[:bodyLen]
			}
		}
		callArgs = append(callArgs, unboxedName)
		counter++
	}

	// Fill remaining up to 4 raw args if sig has fewer than 4 parameters
	for i := len(sig.Parameters); i < 4; i++ {
		rawName := fmt.Sprintf("__unused_arg_%d$raw", i)
		trampolineFn.Parameters = append(trampolineFn.Parameters, ir.Parameter{
			Name: rawName,
			Type: ir.TypeUnknown,
		})
	}

	if sig.ReturnType != ir.TypeVoid {
		retVal := fmt.Sprintf("ret_%d", counter)
		trampolineFn.Body = append(trampolineFn.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   sig.ReturnType,
			Result: retVal,
			Callee: sig.Name,
			Args:   callArgs,
			Span:   sig.Span,
		})
		trampolineFn.Body = append(trampolineFn.Body, ir.Instruction{
			Op:   ir.OpReturn,
			Type: sig.ReturnType,
			Args: []string{retVal},
			Span: sig.Span,
		})
	} else {
		trampolineFn.Body = append(trampolineFn.Body, ir.Instruction{
			Op:     ir.OpCall,
			Type:   ir.TypeVoid,
			Callee: sig.Name,
			Args:   callArgs,
			Span:   sig.Span,
		})
		trampolineFn.Body = append(trampolineFn.Body, ir.Instruction{
			Op:   ir.OpReturn,
			Type: ir.TypeVoid,
			Span: sig.Span,
		})
	}

	extraFunctions = append(extraFunctions, trampolineFn)
	signatures[trampolineName] = trampolineFn
	return trampolineName
}

// closureRawParameterCount is the number of arguments the closure ABI passes
// as raw parameters (see the backend's closure invoke adapters).
const closureRawParameterCount = 4

// closureMoreArgument reads closure argument closureRawParameterCount+index
// into result (an unknown value, undefined when the call passed fewer).
func closureMoreArgument(result string, index int, span ir.SourceSpan) []ir.Instruction {
	indexValue := result + ".more_index"
	return []ir.Instruction{
		{Op: ir.OpConst, Type: ir.TypeNumber, Result: indexValue, Value: strconv.Itoa(index), Span: span},
		{Op: ir.OpCall, Type: ir.TypeUnknown, Result: result, Callee: "__closure.more_arg", Args: []string{indexValue}, Span: span},
	}
}

// closureRestArgument builds rest parameter result, of array type typ and at
// parameter index, from the closure call's arguments.
func closureRestArgument(result string, typ ir.Type, index int, span ir.SourceSpan) []ir.Instruction {
	indexValue := result + ".rest_index"
	return []ir.Instruction{
		{Op: ir.OpConst, Type: ir.TypeNumber, Result: indexValue, Value: strconv.Itoa(index), Span: span},
		{Op: ir.OpCall, Type: typ, Result: result, Callee: "__closure.rest", Args: []string{indexValue}, Span: span},
	}
}
