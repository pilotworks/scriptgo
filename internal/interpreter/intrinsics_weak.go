package interpreter

import (
	"fmt"
	goruntime "runtime"

	"github.com/pilotworks/scriptgo/internal/ir"
)

var (
	interpreterWeakMaps = make(map[string]map[string]Value)
	interpreterWeakSets = make(map[string]map[string]bool)
	weakMapCounter      = 0
	weakSetCounter      = 0
)

func executeWeakIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__gc.collect":
		goruntime.GC()
		return Value{Type: ir.TypeNumber, Number: 0}, nil

	case "__weakref.new":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("WeakRef requires target argument")
		}
		targetVal := env[arguments[0]]
		return Value{
			Type: ir.Type("object:WeakRef"),
			Object: map[string]Value{
				"__target": targetVal,
			},
		}, nil

	case "__weakref.deref":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("WeakRef.deref requires 1 argument")
		}
		refVal := env[arguments[0]]
		if refVal.Object != nil {
			if target, ok := refVal.Object["__target"]; ok {
				return target, nil
			}
		}
		return Value{Type: ir.TypeUnknown}, nil

	case "__weakmap.new":
		weakMapCounter++
		id := fmt.Sprintf("wm_%d", weakMapCounter)
		interpreterWeakMaps[id] = make(map[string]Value)
		return Value{
			Type: ir.Type("object:WeakMap"),
			Object: map[string]Value{
				"__id": {Type: ir.TypeString, String: id},
			},
		}, nil

	case "__weakmap.set":
		if len(arguments) < 3 {
			return Value{}, fmt.Errorf("WeakMap.set requires 3 arguments")
		}
		mapVal := env[arguments[0]]
		id := mapVal.Object["__id"].String
		keyVal := arguments[1]
		valVal := env[arguments[2]]
		if store, ok := interpreterWeakMaps[id]; ok {
			store[keyVal] = valVal
		}
		return mapVal, nil

	case "__weakmap.get":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("WeakMap.get requires 2 arguments")
		}
		mapVal := env[arguments[0]]
		id := mapVal.Object["__id"].String
		keyVal := arguments[1]
		if store, ok := interpreterWeakMaps[id]; ok {
			if val, exists := store[keyVal]; exists {
				return val, nil
			}
		}
		return Value{Type: ir.TypeUnknown}, nil

	case "__weakmap.has":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("WeakMap.has requires 2 arguments")
		}
		mapVal := env[arguments[0]]
		id := mapVal.Object["__id"].String
		keyVal := arguments[1]
		has := false
		if store, ok := interpreterWeakMaps[id]; ok {
			_, has = store[keyVal]
		}
		return Value{Type: ir.TypeBool, Bool: has}, nil

	case "__weakmap.delete":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("WeakMap.delete requires 2 arguments")
		}
		mapVal := env[arguments[0]]
		id := mapVal.Object["__id"].String
		keyVal := arguments[1]
		deleted := false
		if store, ok := interpreterWeakMaps[id]; ok {
			if _, exists := store[keyVal]; exists {
				delete(store, keyVal)
				deleted = true
			}
		}
		return Value{Type: ir.TypeBool, Bool: deleted}, nil

	case "__weakset.new":
		weakSetCounter++
		id := fmt.Sprintf("ws_%d", weakSetCounter)
		interpreterWeakSets[id] = make(map[string]bool)
		return Value{
			Type: ir.Type("object:WeakSet"),
			Object: map[string]Value{
				"__id": {Type: ir.TypeString, String: id},
			},
		}, nil

	case "__weakset.add":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("WeakSet.add requires 2 arguments")
		}
		setVal := env[arguments[0]]
		id := setVal.Object["__id"].String
		valVal := arguments[1]
		if store, ok := interpreterWeakSets[id]; ok {
			store[valVal] = true
		}
		return setVal, nil

	case "__weakset.has":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("WeakSet.has requires 2 arguments")
		}
		setVal := env[arguments[0]]
		id := setVal.Object["__id"].String
		valVal := arguments[1]
		has := false
		if store, ok := interpreterWeakSets[id]; ok {
			has = store[valVal]
		}
		return Value{Type: ir.TypeBool, Bool: has}, nil

	case "__weakset.delete":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("WeakSet.delete requires 2 arguments")
		}
		setVal := env[arguments[0]]
		id := setVal.Object["__id"].String
		valVal := arguments[1]
		deleted := false
		if store, ok := interpreterWeakSets[id]; ok {
			if store[valVal] {
				delete(store, valVal)
				deleted = true
			}
		}
		return Value{Type: ir.TypeBool, Bool: deleted}, nil

	default:
		return Value{}, fmt.Errorf("unknown weak intrinsic %q", name)
	}
}
