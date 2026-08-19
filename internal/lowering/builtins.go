package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// BuiltinGlobal describes a globally available value admitted by the native
// subset. TypeScript-Go remains responsible for resolving the declaration;
// this table defines the native representation and constant value.
type BuiltinGlobal struct {
	Name  string
	Type  ir.Type
	Value string
}

// BuiltinIntrinsic describes a small, explicitly promoted intrinsic. The
// intrinsic name is backend-independent; each backend owns its implementation.
type BuiltinIntrinsic struct {
	Name          string
	ArgumentTypes []ir.Type
	MinArgs       int
	MaxArgs       int
	Lower         func(IntrinsicCall, BuiltinIntrinsic) (string, ir.Type, error)
}

type IntrinsicCall struct {
	Path            string
	Expression      *typescriptgo.SyntaxExpression
	Result          string
	Function        *ir.Function
	Env             map[string]ir.Type
	Counter         *int
	Shapes          map[string]ir.ObjectShape
	Signatures      map[string]ir.Function
	LowerExpression lowerExpressionFunc
}

type lowerExpressionFunc func(string, *typescriptgo.SyntaxExpression, string, *ir.Function, map[string]ir.Type, *int, map[string]ir.ObjectShape, map[string]ir.Function) (string, ir.Type, error)

var builtinGlobals = map[string]BuiltinGlobal{
	"NaN":      {Name: "NaN", Type: ir.TypeNumber, Value: "NaN"},
	"Infinity": {Name: "Infinity", Type: ir.TypeNumber, Value: "+Inf"},
}

var builtinIntrinsics = map[string]BuiltinIntrinsic{
	"Math.abs":      {Name: "Math.abs", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerMathIntrinsic},
	"Math.ceil":     {Name: "Math.ceil", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerMathIntrinsic},
	"Math.floor":    {Name: "Math.floor", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerMathIntrinsic},
	"Math.trunc":    {Name: "Math.trunc", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerMathIntrinsic},
	"console.log":   {Name: "console.log", ArgumentTypes: []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeBool}, MinArgs: 1, MaxArgs: 1, Lower: lowerConsoleIntrinsic},
	"console.info":  {Name: "console.info", ArgumentTypes: []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeBool}, MinArgs: 1, MaxArgs: 1, Lower: lowerConsoleIntrinsic},
	"console.warn":  {Name: "console.warn", ArgumentTypes: []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeBool}, MinArgs: 1, MaxArgs: 1, Lower: lowerConsoleIntrinsic},
	"console.error": {Name: "console.error", ArgumentTypes: []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeBool}, MinArgs: 1, MaxArgs: 1, Lower: lowerConsoleIntrinsic},
}

func builtinGlobal(name string) (BuiltinGlobal, bool) {
	global, ok := builtinGlobals[name]
	return global, ok
}

func builtinIntrinsic(name string) (BuiltinIntrinsic, bool) {
	intrinsic, ok := builtinIntrinsics[name]
	return intrinsic, ok
}

func (call IntrinsicCall) arguments(intrinsic BuiltinIntrinsic) ([]string, []ir.Type, error) {
	if len(call.Expression.Arguments) < intrinsic.MinArgs || len(call.Expression.Arguments) > intrinsic.MaxArgs {
		return nil, nil, fmt.Errorf("builtin %s expects between %d and %d argument(s)", intrinsic.Name, intrinsic.MinArgs, intrinsic.MaxArgs)
	}
	args := make([]string, 0, len(call.Expression.Arguments))
	types := make([]ir.Type, 0, len(call.Expression.Arguments))
	for _, argument := range call.Expression.Arguments {
		value, typ, err := call.LowerExpression(call.Path, argument, "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
		if err != nil {
			return nil, nil, err
		}
		accepted := false
		for _, allowed := range intrinsic.ArgumentTypes {
			if typ == allowed {
				accepted = true
				break
			}
		}
		if !accepted {
			return nil, nil, fmt.Errorf("builtin %s does not support %s", intrinsic.Name, typ)
		}
		args = append(args, value)
		types = append(types, typ)
	}
	return args, types, nil
}

func lowerMathIntrinsic(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	args, _, err := call.arguments(intrinsic)
	if err != nil {
		return "", "", err
	}
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeNumber, Result: result, Callee: "__" + intrinsic.Name, Args: args, Span: toIRSpan(call.Path, call.Expression.Span)})
	return result, ir.TypeNumber, nil
}

func lowerConsoleIntrinsic(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	args, _, err := call.arguments(intrinsic)
	if err != nil {
		return "", "", err
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpPrint, Type: ir.TypeVoid, Callee: intrinsic.Name, Args: args, Span: toIRSpan(call.Path, call.Expression.Span)})
	return "", ir.TypeVoid, nil
}
