package lowering

import "github.com/pilotworks/scriptgo/internal/ir"

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
	Name       string
	ResultType ir.Type
	Argument   ir.Type
	MinArgs    int
	MaxArgs    int
}

var builtinGlobals = map[string]BuiltinGlobal{
	"NaN":      {Name: "NaN", Type: ir.TypeNumber, Value: "NaN"},
	"Infinity": {Name: "Infinity", Type: ir.TypeNumber, Value: "+Inf"},
}

var builtinIntrinsics = map[string]BuiltinIntrinsic{
	"Math.abs":   {Name: "Math.abs", ResultType: ir.TypeNumber, Argument: ir.TypeNumber, MinArgs: 1, MaxArgs: 1},
	"Math.ceil":  {Name: "Math.ceil", ResultType: ir.TypeNumber, Argument: ir.TypeNumber, MinArgs: 1, MaxArgs: 1},
	"Math.floor": {Name: "Math.floor", ResultType: ir.TypeNumber, Argument: ir.TypeNumber, MinArgs: 1, MaxArgs: 1},
	"Math.trunc": {Name: "Math.trunc", ResultType: ir.TypeNumber, Argument: ir.TypeNumber, MinArgs: 1, MaxArgs: 1},
}

func builtinGlobal(name string) (BuiltinGlobal, bool) {
	global, ok := builtinGlobals[name]
	return global, ok
}

func builtinIntrinsic(name string) (BuiltinIntrinsic, bool) {
	intrinsic, ok := builtinIntrinsics[name]
	return intrinsic, ok
}
