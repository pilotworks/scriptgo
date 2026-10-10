package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerObjectEnumerationIntrinsics(m map[string]BuiltinIntrinsic) {
	m["Object.keys"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.keys",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			objVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}

			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeStringArray,
				Result: result,
				Callee: "__object.keys",
				Args:   []string{objVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeStringArray, nil
		},
	}

	m["Object.values"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.values",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			return lowerObjectEnumeration(call, "__object.values")
		},
	}

	m["Object.entries"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "Object.entries",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			return lowerObjectEnumeration(call, "__object.entries")
		},
	}
}

// lowerObjectEnumeration lowers Object.values/entries to the runtime walk of
// the fields the object actually has (absent optional fields are skipped).
// The result is the array type TypeScript gives the call, so values land in
// that element layout and entries are [key, value] tuples.
func lowerObjectEnumeration(call IntrinsicCall, callee string) (string, ir.Type, error) {
	object, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
	if err != nil {
		return "", "", err
	}
	resultType := toIRType(call.Expression.InferredType)
	if declared := call.Env["__storage_type."+call.Result]; call.Result != "" && strings.HasSuffix(string(declared), "[]") &&
		(resultType == ir.TypeUnknownArray || !strings.HasSuffix(string(resultType), "[]")) {
		// TypeScript types Object.values(instance) as any[]; the declared
		// variable type gives the element layout.
		resultType = declared
	}
	if !strings.HasSuffix(string(resultType), "[]") {
		resultType = ir.TypeUnknownArray
	}
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: resultType, Result: result, Callee: callee, Args: []string{object}, Span: toIRSpan(call.Path, call.Expression.Span)})
	return result, resultType, nil
}
