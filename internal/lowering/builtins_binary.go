package lowering

import (
	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerWeakRefIntrinsics(m map[string]BuiltinIntrinsic) {
	// WeakRef & FinalizationRegistry (Category 1: ECMAScript)
	registerCallIntrinsic(m, []string{"__scriptgo.weakrefNew", "WeakRef"}, CategoryECMAScript, "__weakref.new", []ir.Type{ir.TypeObject}, ir.TypeObject, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.weakrefDeref"}, CategoryECMAScript, "__weakref.deref", []ir.Type{ir.TypeObject}, ir.TypeObject, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.finalizationRegistryNew", "FinalizationRegistry"}, CategoryECMAScript, "__finalization_registry.new", []ir.Type{ir.TypeClosure}, ir.TypeObject, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.finalizationRegistryRegister"}, CategoryECMAScript, "__finalization_registry.register", []ir.Type{ir.TypeObject, ir.TypeObject, ir.TypeUnknown, ir.TypeObject}, ir.TypeVoid, 3, 4)
	registerCallIntrinsic(m, []string{"__scriptgo.finalizationRegistryUnregister"}, CategoryECMAScript, "__finalization_registry.unregister", []ir.Type{ir.TypeObject, ir.TypeObject}, ir.TypeBool, 2, 2)
}

func registerAtomicsIntrinsics(m map[string]BuiltinIntrinsic) {
	// Atomics & SharedArrayBuffer globals (Category 1: ECMAScript)
	registerCallIntrinsic(m, []string{"__scriptgo.sharedArrayBufferNew", "SharedArrayBuffer"}, CategoryECMAScript, "__atomics.sharedArrayBufferNew", []ir.Type{ir.TypeNumber}, ir.TypeArrayBuffer, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.atomicsIsLockFree", "Atomics.isLockFree"}, CategoryECMAScript, "__atomics.isLockFree", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.atomicsAdd", "Atomics.add"}, CategoryECMAScript, "__atomics.add", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.atomicsSub", "Atomics.sub"}, CategoryECMAScript, "__atomics.sub", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.atomicsAnd", "Atomics.and"}, CategoryECMAScript, "__atomics.and", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.atomicsOr", "Atomics.or"}, CategoryECMAScript, "__atomics.or", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.atomicsXor", "Atomics.xor"}, CategoryECMAScript, "__atomics.xor", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.atomicsLoad", "Atomics.load"}, CategoryECMAScript, "__atomics.load", []ir.Type{ir.TypeInt32Array, ir.TypeNumber}, ir.TypeNumber, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.atomicsStore", "Atomics.store"}, CategoryECMAScript, "__atomics.store", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.atomicsExchange", "Atomics.exchange"}, CategoryECMAScript, "__atomics.exchange", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.atomicsCompareExchange", "Atomics.compareExchange"}, CategoryECMAScript, "__atomics.compareExchange", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 4, 4)
	registerCallIntrinsic(m, []string{"__scriptgo.atomicsWait", "Atomics.wait"}, CategoryECMAScript, "__atomics.wait", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber, ir.TypeNumber}, ir.TypeString, 3, 4)
	registerCallIntrinsic(m, []string{"__scriptgo.atomicsNotify", "Atomics.notify"}, CategoryECMAScript, "__atomics.notify", []ir.Type{ir.TypeInt32Array, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 2, 3)
}

func registerTypedArrayIntrinsics(m map[string]BuiltinIntrinsic) {
	// TypedArray & ArrayBuffer globals (Category 1: ECMAScript)
	m["ArrayBuffer.isView"] = BuiltinIntrinsic{
		Category: CategoryECMAScript,
		Name:     "ArrayBuffer.isView",
		MinArgs:  1,
		MaxArgs:  1,
		Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
			argVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
			if err != nil {
				return "", "", err
			}
			result := call.Result
			if result == "" {
				result = nextTemp(call.Counter)
			}
			call.Function.Body = append(call.Function.Body, ir.Instruction{
				Op:     ir.OpCall,
				Type:   ir.TypeBool,
				Result: result,
				Callee: "__arraybuffer.isView",
				Args:   []string{argVal},
				Span:   toIRSpan(call.Path, call.Expression.Span),
			})
			return result, ir.TypeBool, nil
		},
	}
	for _, name := range []string{
		"Uint8Array", "Int8Array", "Uint8ClampedArray",
		"Int16Array", "Uint16Array", "Int32Array", "Uint32Array",
		"Float32Array", "Float64Array", "BigInt64Array", "BigUint64Array",
	} {
		targetKind := ir.Type(name)
		className := name
		m[name+".from"] = BuiltinIntrinsic{
			Category: CategoryECMAScript,
			Name:     name + ".from",
			MinArgs:  1,
			MaxArgs:  1,
			Lower: func(call IntrinsicCall, intrinsic BuiltinIntrinsic) (string, ir.Type, error) {
				argVal, _, err := call.LowerExpression(call.Path, call.Expression.Arguments[0], "", call.Function, call.Env, call.Counter, call.Shapes, call.Signatures)
				if err != nil {
					return "", "", err
				}
				result := call.Result
				if result == "" {
					result = nextTemp(call.Counter)
				}
				call.Function.Body = append(call.Function.Body, ir.Instruction{
					Op:     ir.OpCall,
					Type:   targetKind,
					Result: result,
					Callee: "__typedarray.new_array",
					Value:  className,
					Args:   []string{argVal},
					Span:   toIRSpan(call.Path, call.Expression.Span),
				})
				return result, targetKind, nil
			},
		}
	}
}
