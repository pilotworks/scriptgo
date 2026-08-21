package interpreter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

type SetValue struct {
	Entries []Value
	Index   map[string]int
}

func NewSetValue() *SetValue {
	return &SetValue{
		Entries: make([]Value, 0),
		Index:   make(map[string]int),
	}
}

func (s *SetValue) Add(val Value) {
	h := keyHash(val)
	if _, ok := s.Index[h]; ok {
		return
	}
	s.Index[h] = len(s.Entries)
	s.Entries = append(s.Entries, val)
}

func (s *SetValue) Has(val Value) bool {
	h := keyHash(val)
	_, ok := s.Index[h]
	return ok
}

func (s *SetValue) Delete(val Value) bool {
	h := keyHash(val)
	idx, ok := s.Index[h]
	if !ok {
		return false
	}
	delete(s.Index, h)
	s.Entries = append(s.Entries[:idx], s.Entries[idx+1:]...)
	for i := idx; i < len(s.Entries); i++ {
		s.Index[keyHash(s.Entries[i])] = i
	}
	return true
}

func (s *SetValue) Clear() {
	s.Entries = s.Entries[:0]
	s.Index = make(map[string]int)
}

func (s *SetValue) Size() int {
	return len(s.Entries)
}

func executeSetIntrinsic(inst ir.Instruction, env map[string]Value, functions map[string]ir.Function, output *bytes.Buffer) (Value, error) {
	callee := inst.Callee
	args := inst.Args

	switch callee {
	case "__set.new":
		return Value{
			Type:     ir.TypeSet,
			SetValue: NewSetValue(),
		}, nil

	case "__set.new_values":
		sv := NewSetValue()
		if len(args) > 0 {
			arrVal, ok := env[args[0]]
			if ok {
				entries := arrVal.GetArray()
				for _, val := range entries {
					sv.Add(val)
				}
			}
		}
		return Value{
			Type:     ir.TypeSet,
			SetValue: sv,
		}, nil

	case "__set.add":
		if len(args) != 2 {
			return Value{}, fmt.Errorf("Set.add requires 2 arguments (set, value)")
		}
		setVal, ok := env[args[0]]
		if !ok || setVal.SetValue == nil {
			return Value{}, fmt.Errorf("Set.add requires a Set instance")
		}
		vVal, ok := env[args[1]]
		if !ok {
			return Value{}, fmt.Errorf("Set.add missing value")
		}
		setVal.SetValue.Add(vVal)
		return setVal, nil

	case "__set.has":
		if len(args) != 2 {
			return Value{}, fmt.Errorf("Set.has requires 2 arguments (set, value)")
		}
		setVal, ok := env[args[0]]
		if !ok || setVal.SetValue == nil {
			return Value{}, fmt.Errorf("Set.has requires a Set instance")
		}
		vVal, ok := env[args[1]]
		if !ok {
			return Value{}, fmt.Errorf("Set.has missing value")
		}
		return Value{
			Type: ir.TypeBool,
			Bool: setVal.SetValue.Has(vVal),
		}, nil

	case "__set.delete":
		if len(args) != 2 {
			return Value{}, fmt.Errorf("Set.delete requires 2 arguments (set, value)")
		}
		setVal, ok := env[args[0]]
		if !ok || setVal.SetValue == nil {
			return Value{}, fmt.Errorf("Set.delete requires a Set instance")
		}
		vVal, ok := env[args[1]]
		if !ok {
			return Value{}, fmt.Errorf("Set.delete missing value")
		}
		return Value{
			Type: ir.TypeBool,
			Bool: setVal.SetValue.Delete(vVal),
		}, nil

	case "__set.clear":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Set.clear requires 1 argument")
		}
		setVal, ok := env[args[0]]
		if !ok || setVal.SetValue == nil {
			return Value{}, fmt.Errorf("Set.clear requires a Set instance")
		}
		setVal.SetValue.Clear()
		return Value{Type: ir.TypeVoid}, nil

	case "__set.size":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Set.size requires 1 argument")
		}
		setVal, ok := env[args[0]]
		if !ok || setVal.SetValue == nil {
			return Value{}, fmt.Errorf("Set.size requires a Set instance")
		}
		return Value{
			Type:   ir.TypeNumber,
			Number: float64(setVal.SetValue.Size()),
		}, nil

	case "__set.keys", "__set.values":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Set.values requires 1 argument")
		}
		setVal, ok := env[args[0]]
		if !ok || setVal.SetValue == nil {
			return Value{}, fmt.Errorf("Set.values requires a Set instance")
		}
		values := make([]Value, len(setVal.SetValue.Entries))
		copy(values, setVal.SetValue.Entries)
		return Value{
			Type:  inst.Type,
			Array: values,
		}, nil

	case "__set.entries":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Set.entries requires 1 argument")
		}
		setVal, ok := env[args[0]]
		if !ok || setVal.SetValue == nil {
			return Value{}, fmt.Errorf("Set.entries requires a Set instance")
		}
		pairs := make([]Value, len(setVal.SetValue.Entries))
		for i, e := range setVal.SetValue.Entries {
			pairs[i] = Value{
				Type:  ir.Type("object"),
				Array: []Value{e, e},
			}
		}
		return Value{
			Type:  inst.Type,
			Array: pairs,
		}, nil

	case "__set.forEach":
		if len(args) < 2 {
			return Value{}, fmt.Errorf("Set.forEach requires set and callback closure")
		}
		setVal, ok := env[args[0]]
		if !ok || setVal.SetValue == nil {
			return Value{}, fmt.Errorf("Set.forEach requires a Set instance")
		}
		closureVal, ok := env[args[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("Set.forEach callback must be a closure")
		}
		snapshot := make([]Value, len(setVal.SetValue.Entries))
		copy(snapshot, setVal.SetValue.Entries)
		for _, val := range snapshot {
			_, flow, err := executeClosure(functions, closureVal.Closure, []Value{val, val, setVal}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return Value{}, fmt.Errorf("uncaught exception in Set.forEach")
			}
		}
		return Value{Type: ir.TypeVoid}, nil

	case "__set.toString":
		if len(args) != 1 {
			return Value{}, fmt.Errorf("Set.toString requires 1 argument")
		}
		setVal, ok := env[args[0]]
		if !ok || setVal.SetValue == nil {
			return Value{Type: ir.TypeString, String: "Set(0) {}"}, nil
		}
		sv := setVal.SetValue
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Set(%d) {", len(sv.Entries)))
		if len(sv.Entries) > 0 {
			sb.WriteString(" ")
			for i, val := range sv.Entries {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(formatCollectionValue(val))
			}
			sb.WriteString(" ")
		}
		sb.WriteString("}")
		return Value{Type: ir.TypeString, String: sb.String()}, nil

	default:
		return Value{}, fmt.Errorf("unknown Set intrinsic %q", callee)
	}
}
