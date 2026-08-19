package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// BuiltinCategory specifies the standard architectural group of a built-in symbol.
//
// Dual-Surface APIs (e.g. process, Buffer, console, crypto, timers, URL):
// Symbols that exist both as auto-globals (Category 2/3) and as explicit module
// imports (Category 4 via node:process, node:buffer, node:timers, etc.) resolve
// to identical underlying native implementations.
type BuiltinCategory string

const (
	// CategoryECMAScript: Core JavaScript/ECMAScript standard library (Array, Math, Promise, etc.).
	CategoryECMAScript BuiltinCategory = "ECMAScript"
	// CategoryWebCompat: WinterCG/Node.js server-compatible Web APIs (fetch, URL, AbortController, etc.).
	CategoryWebCompat BuiltinCategory = "WebCompat"
	// CategoryNodeGlobal: Node.js runtime host environment globals (process, Buffer, console, etc.).
	CategoryNodeGlobal BuiltinCategory = "NodeGlobal"
	// CategoryNodeModule: Node.js standard modules loaded explicitly (node:fs, node:path, node:crypto, etc.).
	CategoryNodeModule BuiltinCategory = "NodeModule"
)

// BuiltinGlobal describes a globally available value admitted by the native
// subset. TypeScript-Go remains responsible for resolving the declaration;
// this table defines the native representation and constant value.
type BuiltinGlobal struct {
	Category BuiltinCategory
	Name     string
	Type     ir.Type
	Value    string
}

// BuiltinIntrinsic describes a small, explicitly promoted intrinsic. The
// intrinsic name is backend-independent; each backend owns its implementation.
type BuiltinIntrinsic struct {
	Category      BuiltinCategory
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
	// Category 1: ECMAScript built-ins
	"NaN":      {Category: CategoryECMAScript, Name: "NaN", Type: ir.TypeNumber, Value: "NaN"},
	"Infinity": {Category: CategoryECMAScript, Name: "Infinity", Type: ir.TypeNumber, Value: "+Inf"},
}

var builtinIntrinsics = map[string]BuiltinIntrinsic{
	// Category 1: ECMAScript built-ins
	"Math.abs":   {Category: CategoryECMAScript, Name: "Math.abs", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerMathIntrinsic},
	"Math.ceil":  {Category: CategoryECMAScript, Name: "Math.ceil", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerMathIntrinsic},
	"Math.floor": {Category: CategoryECMAScript, Name: "Math.floor", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerMathIntrinsic},
	"Math.trunc": {Category: CategoryECMAScript, Name: "Math.trunc", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerMathIntrinsic},
	"Math.min":   {Category: CategoryECMAScript, Name: "Math.min", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 2, MaxArgs: 2, Lower: lowerMathIntrinsic},
	"Math.max":   {Category: CategoryECMAScript, Name: "Math.max", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 2, MaxArgs: 2, Lower: lowerMathIntrinsic},
	"Math.pow":   {Category: CategoryECMAScript, Name: "Math.pow", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 2, MaxArgs: 2, Lower: lowerMathIntrinsic},
	"Math.sqrt":  {Category: CategoryECMAScript, Name: "Math.sqrt", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerMathIntrinsic},
	"Math.round": {Category: CategoryECMAScript, Name: "Math.round", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerMathIntrinsic},
	"Math.sin":   {Category: CategoryECMAScript, Name: "Math.sin", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerMathIntrinsic},
	"Math.cos":   {Category: CategoryECMAScript, Name: "Math.cos", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerMathIntrinsic},

	// Category 3: Node-specific globals
	"console.log":     {Category: CategoryNodeGlobal, Name: "console.log", ArgumentTypes: []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeBool}, MinArgs: 1, MaxArgs: 1, Lower: lowerConsoleIntrinsic},
	"console.info":    {Category: CategoryNodeGlobal, Name: "console.info", ArgumentTypes: []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeBool}, MinArgs: 1, MaxArgs: 1, Lower: lowerConsoleIntrinsic},
	"console.warn":    {Category: CategoryNodeGlobal, Name: "console.warn", ArgumentTypes: []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeBool}, MinArgs: 1, MaxArgs: 1, Lower: lowerConsoleIntrinsic},
	"console.error":   {Category: CategoryNodeGlobal, Name: "console.error", ArgumentTypes: []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeBool}, MinArgs: 1, MaxArgs: 1, Lower: lowerConsoleIntrinsic},
	"process.exit":    {Category: CategoryNodeGlobal, Name: "process.exit", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerProcessExit},
	"process.cwd":     {Category: CategoryNodeGlobal, Name: "process.cwd", ArgumentTypes: []ir.Type{}, MinArgs: 0, MaxArgs: 0, Lower: lowerProcessCwd},
	"__scriptgo.exit": {Category: CategoryNodeGlobal, Name: "process.exit", ArgumentTypes: []ir.Type{ir.TypeNumber}, MinArgs: 1, MaxArgs: 1, Lower: lowerProcessExit},
	"__scriptgo.cwd":  {Category: CategoryNodeGlobal, Name: "process.cwd", ArgumentTypes: []ir.Type{}, MinArgs: 0, MaxArgs: 0, Lower: lowerProcessCwd},

	// Category 4: Node built-in modules
	"fs.readFileSync":          {Category: CategoryNodeModule, Name: "fs.readFileSync", ArgumentTypes: []ir.Type{ir.TypeString}, MinArgs: 1, MaxArgs: 1, Lower: lowerFsReadFileSync},
	"fs.writeFileSync":         {Category: CategoryNodeModule, Name: "fs.writeFileSync", ArgumentTypes: []ir.Type{ir.TypeString}, MinArgs: 2, MaxArgs: 2, Lower: lowerFsWriteFileSync},
	"fs.existsSync":            {Category: CategoryNodeModule, Name: "fs.existsSync", ArgumentTypes: []ir.Type{ir.TypeString}, MinArgs: 1, MaxArgs: 1, Lower: lowerFsExistsSync},
	"__scriptgo.readFileSync":  {Category: CategoryNodeModule, Name: "fs.readFileSync", ArgumentTypes: []ir.Type{ir.TypeString}, MinArgs: 1, MaxArgs: 1, Lower: lowerFsReadFileSync},
	"__scriptgo.writeFileSync": {Category: CategoryNodeModule, Name: "fs.writeFileSync", ArgumentTypes: []ir.Type{ir.TypeString}, MinArgs: 2, MaxArgs: 2, Lower: lowerFsWriteFileSync},
	"__scriptgo.existsSync":    {Category: CategoryNodeModule, Name: "fs.existsSync", ArgumentTypes: []ir.Type{ir.TypeString}, MinArgs: 1, MaxArgs: 1, Lower: lowerFsExistsSync},
	"readFileSync":             {Category: CategoryNodeModule, Name: "fs.readFileSync", ArgumentTypes: []ir.Type{ir.TypeString}, MinArgs: 1, MaxArgs: 1, Lower: lowerFsReadFileSync},
	"writeFileSync":            {Category: CategoryNodeModule, Name: "fs.writeFileSync", ArgumentTypes: []ir.Type{ir.TypeString}, MinArgs: 2, MaxArgs: 2, Lower: lowerFsWriteFileSync},
	"existsSync":               {Category: CategoryNodeModule, Name: "fs.existsSync", ArgumentTypes: []ir.Type{ir.TypeString}, MinArgs: 1, MaxArgs: 1, Lower: lowerFsExistsSync},
}

func builtinGlobal(name string) (BuiltinGlobal, bool) {
	global, ok := builtinGlobals[name]
	return global, ok
}

func builtinIntrinsic(name string) (BuiltinIntrinsic, bool) {
	intrinsic, ok := builtinIntrinsics[name]
	return intrinsic, ok
}

// BuiltinsByCategory returns all registered globals and intrinsics for a category.
func BuiltinsByCategory(cat BuiltinCategory) ([]BuiltinGlobal, []BuiltinIntrinsic) {
	var globals []BuiltinGlobal
	for _, g := range builtinGlobals {
		if g.Category == cat {
			globals = append(globals, g)
		}
	}
	var intrinsics []BuiltinIntrinsic
	for _, i := range builtinIntrinsics {
		if i.Category == cat {
			intrinsics = append(intrinsics, i)
		}
	}
	return globals, intrinsics
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

func lowerFsReadFileSync(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	args, _, err := call.arguments(intrinsic)
	if err != nil {
		return "", "", err
	}
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__fs.readFileSync", Args: args, Span: toIRSpan(call.Path, call.Expression.Span)})
	return result, ir.TypeString, nil
}

func lowerFsWriteFileSync(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	args, _, err := call.arguments(intrinsic)
	if err != nil {
		return "", "", err
	}
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Result: result, Callee: "__fs.writeFileSync", Args: args, Span: toIRSpan(call.Path, call.Expression.Span)})
	return result, ir.TypeVoid, nil
}

func lowerFsExistsSync(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	args, _, err := call.arguments(intrinsic)
	if err != nil {
		return "", "", err
	}
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeBool, Result: result, Callee: "__fs.existsSync", Args: args, Span: toIRSpan(call.Path, call.Expression.Span)})
	return result, ir.TypeBool, nil
}

func lowerProcessExit(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	args, _, err := call.arguments(intrinsic)
	if err != nil {
		return "", "", err
	}
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeVoid, Result: result, Callee: "__process.exit", Args: args, Span: toIRSpan(call.Path, call.Expression.Span)})
	return result, ir.TypeVoid, nil
}

func lowerProcessCwd(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
	result := call.Result
	if result == "" {
		result = nextTemp(call.Counter)
	}
	call.Function.Body = append(call.Function.Body, ir.Instruction{Op: ir.OpCall, Type: ir.TypeString, Result: result, Callee: "__process.cwd", Args: nil, Span: toIRSpan(call.Path, call.Expression.Span)})
	return result, ir.TypeString, nil
}
