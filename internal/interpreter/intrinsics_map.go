package interpreter

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

type MapEntry struct {
	Key   Value
	Value Value
}

type MapValue struct {
	Entries []MapEntry
	Index   map[string]int
}

func NewMapValue() *MapValue {
	return &MapValue{
		Entries: make([]MapEntry, 0),
		Index:   make(map[string]int),
	}
}

func keyHash(k Value) string {
	switch k.Type {
	case ir.TypeString:
		return "s:" + k.String
	case ir.TypeNumber:
		return "n:" + strconv.FormatFloat(k.Number, 'f', -1, 64)
	case ir.TypeBigInt:
		return "b:" + strconv.FormatInt(k.BigInt, 10)
	case ir.TypeBool:
		return "bool:" + strconv.FormatBool(k.Bool)
	case ir.TypeSymbol:
		return "sym:" + strconv.FormatUint(k.SymbolID, 10)
	default:
		if k.String != "" {
			return "s:" + k.String
		}
		if k.Number != 0 {
			return "n:" + strconv.FormatFloat(k.Number, 'f', -1, 64)
		}
		return fmt.Sprintf("o:%p", k.Object)
	}
}

func (m *MapValue) Set(key, val Value) {
	h := keyHash(key)
	if idx, ok := m.Index[h]; ok {
		m.Entries[idx].Value = val
		return
	}
	m.Index[h] = len(m.Entries)
	m.Entries = append(m.Entries, MapEntry{Key: key, Value: val})
}

func (m *MapValue) Get(key Value) (Value, bool) {
	h := keyHash(key)
	if idx, ok := m.Index[h]; ok {
		return m.Entries[idx].Value, true
	}
	return Value{Type: ir.TypeUnknown}, false
}

func (m *MapValue) Has(key Value) bool {
	h := keyHash(key)
	_, ok := m.Index[h]
	return ok
}

func (m *MapValue) Delete(key Value) bool {
	h := keyHash(key)
	idx, ok := m.Index[h]
	if !ok {
		return false
	}
	delete(m.Index, h)
	m.Entries = append(m.Entries[:idx], m.Entries[idx+1:]...)
	for i := idx; i < len(m.Entries); i++ {
		m.Index[keyHash(m.Entries[i].Key)] = i
	}
	return true
}

func (m *MapValue) Clear() {
	m.Entries = m.Entries[:0]
	m.Index = make(map[string]int)
}

func (m *MapValue) Size() int {
	return len(m.Entries)
}

func executeMapIntrinsic(inst ir.Instruction, env map[string]Value, functions map[string]ir.Function, output *bytes.Buffer) (Value, error) {
	callee := inst.Callee
	args := inst.Args

	switch callee {
	case "__map.new":
		return Value{
			Type:     ir.TypeMap,
			MapValue: NewMapValue(),
		}, nil

	case "__map.new_entries":
		mv := NewMapValue()
		if len(args) > 0 {
			arrVal, ok := env[args[0]]
			if ok {
				entries := arrVal.GetArray()
				for _, entryVal := range entries {
					pair := entryVal.GetArray()
					if len(pair) >= 2 {
						mv.Set(pair[0], pair[1])
					} else if len(entryVal.Object) >= 2 {
						k, ok1 := entryVal.Object["0"]
						v, ok2 := entryVal.Object["1"]
						if ok1 && ok2 {
							mv.Set(k, v)
						}
					}
				}
			}
		}
		return Value{
			Type:     ir.TypeMap,
			MapValue: mv,
		}, nil

	case "__map.set":
		if len(args) != 3 {
			return Value{}, fmt.Errorf("Map.set requires 3 arguments (map, key, value)")
		}
		mapVal, ok := env[args[0]]
		if !ok || mapVal.MapValue == nil {
			return Value{}, fmt.Errorf("Map.set requires a Map instance")
		}
		kVal, ok := env[args[1]]
		if !ok {
			return Value{}, fmt.Errorf("Map.set missing key")
		}
		vVal, ok := env[args[2]]
		if !ok {
			return Value{}, fmt.Errorf("Map.set missing value")
		}
		mapVal.MapValue.Set(kVal, vVal)
		return mapVal, nil

	case "__map.get":
		if len(args) != 2 {
			return Value{}, fmt.Errorf("Map.get requires 2 arguments (map, key)")
		}
		mapVal, ok := env[args[0]]
		if !ok || mapVal.MapValue == nil {
			return Value{}, fmt.Errorf("Map.get requires a Map instance")
		}
		kVal, ok := env[args[1]]
		if !ok {
			return Value{}, fmt.Errorf("Map.get missing key")
		}
		val, found := mapVal.MapValue.Get(kVal)
		if !found {
			switch inst.Type {
			case ir.TypeNumber:
				return Value{Type: ir.TypeNumber, Number: 0}, nil
			case ir.TypeString:
				return Value{Type: ir.TypeString, String: ""}, nil
			case ir.TypeBool:
				return Value{Type: ir.TypeBool, Bool: false}, nil
			case ir.TypeBigInt:
				return Value{Type: ir.TypeBigInt, BigInt: 0}, nil
			default:
				return Value{Type: inst.Type}, nil
			}
		}
		return val, nil

	case "__map.has":
		if len(args) != 2 {
			return Value{}, fmt.Errorf("Map.has requires 2 arguments (map, key)")
		}
		mapVal, ok := env[args[0]]
		if !ok || mapVal.MapValue == nil {
			return Value{}, fmt.Errorf("Map.has requires a Map instance")
		}
		kVal, ok := env[args[1]]
		if !ok {
			return Value{}, fmt.Errorf("Map.has missing key")
		}
		return Value{
			Type: ir.TypeBool,
			Bool: mapVal.MapValue.Has(kVal),
		}, nil

	case "__map.delete":
		if len(args) != 2 {
			return Value{}, fmt.Errorf("Map.delete requires 2 arguments (map, key)")
		}
		mapVal, ok := env[args[0]]
		if !ok || mapVal.MapValue == nil {
			return Value{}, fmt.Errorf("Map.delete requires a Map instance")
		}
		kVal, ok := env[args[1]]
		if !ok {
			return Value{}, fmt.Errorf("Map.delete missing key")
		}
		return Value{
			Type: ir.TypeBool,
			Bool: mapVal.MapValue.Delete(kVal),
		}, nil

	case "__map.clear":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Map.clear requires 1 argument")
		}
		mapVal, ok := env[args[0]]
		if !ok || mapVal.MapValue == nil {
			return Value{}, fmt.Errorf("Map.clear requires a Map instance")
		}
		mapVal.MapValue.Clear()
		return Value{Type: ir.TypeVoid}, nil

	case "__map.size":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Map.size requires 1 argument")
		}
		mapVal, ok := env[args[0]]
		if !ok || mapVal.MapValue == nil {
			return Value{}, fmt.Errorf("Map.size requires a Map instance")
		}
		return Value{
			Type:   ir.TypeNumber,
			Number: float64(mapVal.MapValue.Size()),
		}, nil

	case "__map.keys":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Map.keys requires 1 argument")
		}
		mapVal, ok := env[args[0]]
		if !ok || mapVal.MapValue == nil {
			return Value{}, fmt.Errorf("Map.keys requires a Map instance")
		}
		keys := make([]Value, len(mapVal.MapValue.Entries))
		for i, e := range mapVal.MapValue.Entries {
			keys[i] = e.Key
		}
		return Value{
			Type:  inst.Type,
			Array: keys,
		}, nil

	case "__map.values":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Map.values requires 1 argument")
		}
		mapVal, ok := env[args[0]]
		if !ok || mapVal.MapValue == nil {
			return Value{}, fmt.Errorf("Map.values requires a Map instance")
		}
		values := make([]Value, len(mapVal.MapValue.Entries))
		for i, e := range mapVal.MapValue.Entries {
			values[i] = e.Value
		}
		return Value{
			Type:  inst.Type,
			Array: values,
		}, nil

	case "__map.entries":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Map.entries requires 1 argument")
		}
		mapVal, ok := env[args[0]]
		if !ok || mapVal.MapValue == nil {
			return Value{}, fmt.Errorf("Map.entries requires a Map instance")
		}
		pairs := make([]Value, len(mapVal.MapValue.Entries))
		for i, e := range mapVal.MapValue.Entries {
			pairs[i] = Value{
				Type:  ir.Type("object"),
				Array: []Value{e.Key, e.Value},
			}
		}
		return Value{
			Type:  inst.Type,
			Array: pairs,
		}, nil

	case "__map.forEach":
		if len(args) < 2 {
			return Value{}, fmt.Errorf("Map.forEach requires map and callback closure")
		}
		mapVal, ok := env[args[0]]
		if !ok || mapVal.MapValue == nil {
			return Value{}, fmt.Errorf("Map.forEach requires a Map instance")
		}
		closureVal, ok := env[args[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("Map.forEach callback must be a closure")
		}
		snapshot := make([]MapEntry, len(mapVal.MapValue.Entries))
		copy(snapshot, mapVal.MapValue.Entries)
		for _, entry := range snapshot {
			_, flow, err := executeClosure(functions, closureVal.Closure, []Value{entry.Value, entry.Key, mapVal}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return Value{}, fmt.Errorf("uncaught exception in Map.forEach")
			}
		}
		return Value{Type: ir.TypeVoid}, nil

	case "__map.toString":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Map.toString requires 1 argument")
		}
		mapVal, ok := env[args[0]]
		if !ok || mapVal.MapValue == nil {
			return Value{Type: ir.TypeString, String: "Map(0) {}"}, nil
		}
		mv := mapVal.MapValue
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Map(%d) {", len(mv.Entries)))
		if len(mv.Entries) > 0 {
			sb.WriteString(" ")
			for i, entry := range mv.Entries {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%s => %s", formatCollectionValue(entry.Key), formatCollectionValue(entry.Value)))
			}
			sb.WriteString(" ")
		}
		sb.WriteString("}")
		return Value{Type: ir.TypeString, String: sb.String()}, nil

	default:
		return Value{}, fmt.Errorf("unknown Map intrinsic %q", callee)
	}
}
