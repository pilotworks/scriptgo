package interpreter

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func executeObjectIntrinsic(functions map[string]ir.Function, name string, arguments []string, env map[string]Value, output *bytes.Buffer) (Value, error) {
	switch name {
	case "__object.get_prop":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("__object.get_prop requires 2 arguments")
		}
		obj, ok1 := env[arguments[0]]
		prop, ok2 := env[arguments[1]]
		if !ok1 || !ok2 {
			return Value{}, fmt.Errorf("unknown argument for __object.get_prop")
		}
		if obj.Object != nil {
			if v, ok := obj.Object[prop.String]; ok {
				return v, nil
			}
		}
		return Value{Type: ir.TypeNumber, Number: 0}, nil
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
			var sortedKeys []string
			for k := range obj.Object {
				sortedKeys = append(sortedKeys, k)
			}
			sort.Strings(sortedKeys)
			for _, k := range sortedKeys {
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
			var sortedKeys []string
			for k := range obj.Object {
				sortedKeys = append(sortedKeys, k)
			}
			sort.Strings(sortedKeys)
			for _, k := range sortedKeys {
				vals = append(vals, obj.Object[k])
			}
		} else if obj.ArrayRef != nil || len(obj.Array) > 0 {
			vals = append(vals, obj.GetArray()...)
		}
		return Value{Type: ir.TypeNumberArray, Array: vals}, nil
	case "__object.entries":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("__object.entries requires 1 argument")
		}
		obj, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown argument for __object.entries")
		}
		var entries []Value
		if obj.Object != nil {
			var sortedKeys []string
			for k := range obj.Object {
				sortedKeys = append(sortedKeys, k)
			}
			sort.Strings(sortedKeys)
			for _, k := range sortedKeys {
				entries = append(entries, Value{Type: ir.TypeObject, Array: []Value{{Type: ir.TypeString, String: k}, obj.Object[k]}})
			}
		} else if obj.ArrayRef != nil || len(obj.Array) > 0 {
			arr := obj.GetArray()
			for i, v := range arr {
				entries = append(entries, Value{Type: ir.TypeObject, Array: []Value{{Type: ir.TypeString, String: strconv.Itoa(i)}, v}})
			}
		}
		return Value{Type: ir.Type("any[]"), Array: entries}, nil
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
	case "__object.fromEntries":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("__object.fromEntries requires 1 argument")
		}
		entries, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown argument for __object.fromEntries")
		}
		resultObj := make(map[string]Value)
		arr := entries.GetArray()
		for _, item := range arr {
			var k string
			var v Value
			if len(item.GetArray()) >= 2 {
				subArr := item.GetArray()
				k = subArr[0].String
				v = subArr[1]
			} else if item.Object != nil {
				k = item.Object["0"].String
				v = item.Object["1"]
			}
			if k != "" {
				resultObj[k] = v
			}
		}
		return Value{Type: ir.Type("object:Record"), Object: resultObj}, nil
	case "__object.groupBy":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("__object.groupBy requires 2 arguments")
		}
		items, ok1 := env[arguments[0]]
		closureVal, ok2 := env[arguments[1]]
		if !ok1 || !ok2 {
			return Value{}, fmt.Errorf("unknown argument for __object.groupBy")
		}
		arr := items.GetArray()
		grouped := make(map[string]Value)
		for _, item := range arr {
			k := item.String
			if closureVal.Closure != nil {
				res, _, err := executeClosure(functions, closureVal.Closure, []Value{item}, output)
				if err == nil {
					k = format(res)
				}
			} else if item.Type == ir.TypeNumber {
				k = format(item)
			}
			existing, ok := grouped[k]
			var list []Value
			if ok {
				list = existing.GetArray()
			}
			list = append(list, item)
			grouped[k] = Value{Type: ir.Type("any[]"), Array: list}
		}
		return Value{Type: ir.Type("object:Record"), Object: grouped}, nil
	case "__object.create":
		return Value{Type: ir.TypeObject, Object: make(map[string]Value)}, nil
	case "__object.freeze", "__object.seal", "__object.preventExtensions", "__object.setPrototypeOf", "__object.defineProperty", "__object.defineProperties":
		if len(arguments) > 0 {
			if obj, ok := env[arguments[0]]; ok {
				return obj, nil
			}
		}
		return Value{Type: ir.TypeObject, Object: make(map[string]Value)}, nil
	case "__object.isFrozen", "__object.isSealed", "__object.isExtensible":
		return Value{Type: ir.TypeBool, Bool: true}, nil
	case "__object.getOwnPropertyNames":
		return executeObjectIntrinsic(functions, "__object.keys", arguments, env, output)
	case "__object.getOwnPropertySymbols":
		return Value{Type: ir.Type("symbol[]"), Array: []Value{}}, nil
	case "__object.getOwnPropertyDescriptor":
		desc := make(map[string]Value)
		if len(arguments) >= 2 {
			if obj, ok := env[arguments[0]]; ok && obj.Object != nil {
				if prop, ok2 := env[arguments[1]]; ok2 {
					if val, ok3 := obj.Object[prop.String]; ok3 {
						desc["value"] = val
						desc["writable"] = Value{Type: ir.TypeBool, Bool: true}
						desc["enumerable"] = Value{Type: ir.TypeBool, Bool: true}
						desc["configurable"] = Value{Type: ir.TypeBool, Bool: true}
					}
				}
			}
		}
		return Value{Type: ir.TypeObject, Object: desc}, nil
	case "__object.getOwnPropertyDescriptors":
		descs := make(map[string]Value)
		if len(arguments) >= 1 {
			if obj, ok := env[arguments[0]]; ok && obj.Object != nil {
				for k, val := range obj.Object {
					desc := make(map[string]Value)
					desc["value"] = val
					desc["writable"] = Value{Type: ir.TypeBool, Bool: true}
					desc["enumerable"] = Value{Type: ir.TypeBool, Bool: true}
					desc["configurable"] = Value{Type: ir.TypeBool, Bool: true}
					descs[k] = Value{Type: ir.TypeObject, Object: desc}
				}
			}
		}
		return Value{Type: ir.TypeObject, Object: descs}, nil
	case "__object.getPrototypeOf":
		return Value{Type: ir.TypeObject, Object: make(map[string]Value)}, nil
	case "__object.isPrototypeOf":
		return Value{Type: ir.TypeBool, Bool: false}, nil
	case "__object.propertyIsEnumerable":
		return executeObjectIntrinsic(functions, "__object.hasOwn", arguments, env, output)
	case "__object.toString", "__object.toLocaleString":
		return Value{Type: ir.TypeString, String: "[object Object]"}, nil
	case "__object.valueOf":
		if len(arguments) > 0 {
			if obj, ok := env[arguments[0]]; ok {
				return obj, nil
			}
		}
		return Value{Type: ir.TypeObject, Object: make(map[string]Value)}, nil
	case "__object.new":
		return Value{Type: ir.TypeObject, Object: make(map[string]Value)}, nil
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
