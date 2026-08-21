package interpreter

import (
	"fmt"
	"math"
	"strconv"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func executeObjectIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	switch name {
	case "__object.is":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("__object.is requires 2 arguments")
		}
		v1, ok1 := env[arguments[0]]
		v2, ok2 := env[arguments[1]]
		if !ok1 || !ok2 {
			return Value{}, fmt.Errorf("unknown argument for __object.is")
		}
		isSame := sameValue(v1, v2)
		return Value{Type: ir.TypeBool, Bool: isSame}, nil
	case "__object.hasOwn":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("__object.hasOwn requires 2 arguments")
		}
		obj, ok1 := env[arguments[0]]
		prop, ok2 := env[arguments[1]]
		if !ok1 || !ok2 {
			return Value{}, fmt.Errorf("unknown argument for __object.hasOwn")
		}
		if obj.Object != nil {
			_, exists := obj.Object[prop.String]
			return Value{Type: ir.TypeBool, Bool: exists}, nil
		}
		if obj.ArrayRef != nil || len(obj.Array) > 0 {
			arr := obj.GetArray()
			idx, err := strconv.Atoi(prop.String)
			if err == nil && idx >= 0 && idx < len(arr) {
				return Value{Type: ir.TypeBool, Bool: true}, nil
			}
			return Value{Type: ir.TypeBool, Bool: false}, nil
		}
		return Value{Type: ir.TypeBool, Bool: false}, nil
	case "__object.keys":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("__object.keys requires 1 argument")
		}
		obj, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown argument for __object.keys")
		}
		var keys []Value
		if obj.Object != nil {
			for k := range obj.Object {
				keys = append(keys, Value{Type: ir.TypeString, String: k})
			}
		} else if obj.ArrayRef != nil || len(obj.Array) > 0 {
			arr := obj.GetArray()
			for i := range arr {
				keys = append(keys, Value{Type: ir.TypeString, String: strconv.Itoa(i)})
			}
		}
		return Value{Type: ir.TypeStringArray, Array: keys}, nil
	case "__object.values":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("__object.values requires 1 argument")
		}
		obj, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown argument for __object.values")
		}
		var vals []Value
		if obj.Object != nil {
			for _, v := range obj.Object {
				vals = append(vals, v)
			}
		} else if obj.ArrayRef != nil || len(obj.Array) > 0 {
			vals = append(vals, obj.GetArray()...)
		}
		return Value{Type: ir.TypeNumberArray, Array: vals}, nil
	case "__object.assign":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("__object.assign requires at least 1 argument")
		}
		target, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown target for __object.assign")
		}
		if target.Object == nil {
			target.Object = make(map[string]Value)
		}
		for i := 1; i < len(arguments); i++ {
			src, ok := env[arguments[i]]
			if ok && src.Object != nil {
				for k, v := range src.Object {
					target.Object[k] = v
				}
			}
		}
		env[arguments[0]] = target
		return target, nil
	default:
		return Value{}, fmt.Errorf("unknown object intrinsic %q", name)
	}
}

func sameValue(v1, v2 Value) bool {
	if v1.Type != v2.Type {
		return false
	}
	switch v1.Type {
	case ir.TypeNumber:
		if math.IsNaN(v1.Number) && math.IsNaN(v2.Number) {
			return true
		}
		if v1.Number == 0 && v2.Number == 0 {
			return math.Copysign(1, v1.Number) == math.Copysign(1, v2.Number)
		}
		return v1.Number == v2.Number
	case ir.TypeString:
		return v1.String == v2.String
	case ir.TypeBool:
		return v1.Bool == v2.Bool
	case ir.TypeBigInt:
		return v1.BigInt == v2.BigInt
	case ir.TypeSymbol:
		return v1.SymbolID == v2.SymbolID
	default:
		return v1.Number == v2.Number && v1.String == v2.String && v1.Bool == v2.Bool
	}
}
