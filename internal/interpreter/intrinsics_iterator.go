package interpreter

import (
	"bytes"
	"fmt"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func getIteratorElements(iterVal Value) []Value {
	if iterVal.Object != nil {
		if elemVal, ok := iterVal.Object["__elements"]; ok {
			arr := elemVal.GetArray()
			idx := 0
			if idxVal, ok := iterVal.Object["__index"]; ok {
				idx = int(idxVal.Number)
			}
			if idx >= len(arr) {
				return nil
			}
			return arr[idx:]
		}
		if itemsVal, ok := iterVal.Object["__items"]; ok {
			arr := itemsVal.GetArray()
			idx := 0
			if sVal, ok := iterVal.Object["__state"]; ok {
				idx = int(sVal.Number)
			}
			if idx >= len(arr) {
				return nil
			}
			return arr[idx:]
		}
	}
	arr := iterVal.GetArray()
	if len(arr) > 0 {
		return arr
	}
	if iterVal.Type == ir.TypeString {
		var elems []Value
		for _, r := range iterVal.String {
			elems = append(elems, Value{Type: ir.TypeString, String: string(r)})
		}
		return elems
	}
	return nil
}

func makeIteratorObject(elements []Value) Value {
	elemType := ir.TypeNumber
	if len(elements) > 0 {
		elemType = elements[0].Type
	}
	arrType := ir.Type(string(elemType) + "[]")
	return Value{
		Type: ir.TypeObject,
		Object: map[string]Value{
			"__elements": {Type: arrType, Array: elements},
			"__index":    {Type: ir.TypeNumber, Number: 0},
		},
	}
}

func isValTruthy(v Value) bool {
	switch v.Type {
	case ir.TypeBool:
		return v.Bool
	case ir.TypeNumber:
		return v.Number != 0
	case ir.TypeString:
		return v.String != "" && v.String != "0" && v.String != "false" && v.String != "undefined" && v.String != "null"
	case ir.TypeVoid:
		return false
	default:
		return true
	}
}

func executeIteratorIntrinsic(instruction ir.Instruction, env map[string]Value, functions map[string]ir.Function, output *bytes.Buffer) (Value, error) {
	switch instruction.Callee {
	case "__iterator.from":
		if len(instruction.Args) < 1 {
			return Value{}, fmt.Errorf("Iterator.from requires 1 argument")
		}
		sourceVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown argument %q", instruction.Args[0])
		}
		elements := getIteratorElements(sourceVal)
		return makeIteratorObject(elements), nil

	case "__iterator.map":
		if len(instruction.Args) < 2 {
			return Value{}, fmt.Errorf("iterator.map requires iterator and callback")
		}
		iterVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown iterator %q", instruction.Args[0])
		}
		closureVal, ok := env[instruction.Args[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("iterator.map callback must be a function")
		}
		elements := getIteratorElements(iterVal)
		var mapped []Value
		for i, el := range elements {
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{el, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			mapped = append(mapped, res)
		}
		return makeIteratorObject(mapped), nil

	case "__iterator.filter":
		if len(instruction.Args) < 2 {
			return Value{}, fmt.Errorf("iterator.filter requires iterator and predicate")
		}
		iterVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown iterator %q", instruction.Args[0])
		}
		closureVal, ok := env[instruction.Args[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("iterator.filter predicate must be a function")
		}
		elements := getIteratorElements(iterVal)
		var filtered []Value
		for i, el := range elements {
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{el, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if isValTruthy(res) {
				filtered = append(filtered, el)
			}
		}
		return makeIteratorObject(filtered), nil

	case "__iterator.take":
		if len(instruction.Args) < 2 {
			return Value{}, fmt.Errorf("iterator.take requires iterator and limit")
		}
		iterVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown iterator %q", instruction.Args[0])
		}
		limitVal, ok := env[instruction.Args[1]]
		if !ok {
			return Value{}, fmt.Errorf("unknown limit %q", instruction.Args[1])
		}
		limit := int(limitVal.Number)
		if limit < 0 {
			limit = 0
		}
		elements := getIteratorElements(iterVal)
		if limit < len(elements) {
			elements = elements[:limit]
		}
		return makeIteratorObject(elements), nil

	case "__iterator.drop":
		if len(instruction.Args) < 2 {
			return Value{}, fmt.Errorf("iterator.drop requires iterator and count")
		}
		iterVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown iterator %q", instruction.Args[0])
		}
		countVal, ok := env[instruction.Args[1]]
		if !ok {
			return Value{}, fmt.Errorf("unknown count %q", instruction.Args[1])
		}
		count := int(countVal.Number)
		if count < 0 {
			count = 0
		}
		elements := getIteratorElements(iterVal)
		if count >= len(elements) {
			elements = nil
		} else {
			elements = elements[count:]
		}
		return makeIteratorObject(elements), nil

	case "__iterator.flat_map":
		if len(instruction.Args) < 2 {
			return Value{}, fmt.Errorf("iterator.flatMap requires iterator and callback")
		}
		iterVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown iterator %q", instruction.Args[0])
		}
		closureVal, ok := env[instruction.Args[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("iterator.flatMap callback must be a function")
		}
		elements := getIteratorElements(iterVal)
		var flattened []Value
		for i, el := range elements {
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{el, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			subElems := getIteratorElements(res)
			if subElems != nil {
				flattened = append(flattened, subElems...)
			} else {
				flattened = append(flattened, res)
			}
		}
		return makeIteratorObject(flattened), nil

	case "__iterator.to_array":
		if len(instruction.Args) < 1 {
			return Value{}, fmt.Errorf("iterator.toArray requires iterator")
		}
		iterVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown iterator %q", instruction.Args[0])
		}
		elements := getIteratorElements(iterVal)
		elemType := ir.TypeNumber
		if len(elements) > 0 {
			elemType = elements[0].Type
		}
		arrType := ir.Type(string(elemType) + "[]")
		return Value{
			Type:  arrType,
			Array: elements,
		}, nil

	case "__iterator.for_each":
		if len(instruction.Args) < 2 {
			return Value{}, fmt.Errorf("iterator.forEach requires iterator and callback")
		}
		iterVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown iterator %q", instruction.Args[0])
		}
		closureVal, ok := env[instruction.Args[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("iterator.forEach callback must be a function")
		}
		elements := getIteratorElements(iterVal)
		for i, el := range elements {
			_, _, err := executeClosure(functions, closureVal.Closure, []Value{el, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
		}
		return Value{Type: ir.TypeVoid}, nil

	case "__iterator.reduce":
		if len(instruction.Args) < 2 {
			return Value{}, fmt.Errorf("iterator.reduce requires iterator and callback")
		}
		iterVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown iterator %q", instruction.Args[0])
		}
		closureVal, ok := env[instruction.Args[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("iterator.reduce callback must be a function")
		}
		elements := getIteratorElements(iterVal)
		if len(elements) == 0 && len(instruction.Args) < 3 {
			return Value{}, fmt.Errorf("reduce of empty iterator with no initial value")
		}
		startIdx := 0
		var accum Value
		if len(instruction.Args) >= 3 {
			if initVal, ok := env[instruction.Args[2]]; ok {
				accum = initVal
			}
		} else {
			accum = elements[0]
			startIdx = 1
		}
		for i := startIdx; i < len(elements); i++ {
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{accum, elements[i], {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			accum = res
		}
		return accum, nil

	case "__iterator.some":
		if len(instruction.Args) < 2 {
			return Value{}, fmt.Errorf("iterator.some requires iterator and predicate")
		}
		iterVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown iterator %q", instruction.Args[0])
		}
		closureVal, ok := env[instruction.Args[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("iterator.some predicate must be a function")
		}
		elements := getIteratorElements(iterVal)
		for i, el := range elements {
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{el, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if isValTruthy(res) {
				return Value{Type: ir.TypeBool, Bool: true}, nil
			}
		}
		return Value{Type: ir.TypeBool, Bool: false}, nil

	case "__iterator.every":
		if len(instruction.Args) < 2 {
			return Value{}, fmt.Errorf("iterator.every requires iterator and predicate")
		}
		iterVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown iterator %q", instruction.Args[0])
		}
		closureVal, ok := env[instruction.Args[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("iterator.every predicate must be a function")
		}
		elements := getIteratorElements(iterVal)
		for i, el := range elements {
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{el, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if !isValTruthy(res) {
				return Value{Type: ir.TypeBool, Bool: false}, nil
			}
		}
		return Value{Type: ir.TypeBool, Bool: true}, nil

	case "__iterator.find":
		if len(instruction.Args) < 2 {
			return Value{}, fmt.Errorf("iterator.find requires iterator and predicate")
		}
		iterVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown iterator %q", instruction.Args[0])
		}
		closureVal, ok := env[instruction.Args[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("iterator.find predicate must be a function")
		}
		elements := getIteratorElements(iterVal)
		for i, el := range elements {
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{el, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if isValTruthy(res) {
				return el, nil
			}
		}
		return Value{Type: ir.TypeVoid, String: "undefined"}, nil

	case "__iterator.next":
		if len(instruction.Args) < 1 {
			return Value{}, fmt.Errorf("iterator.next requires iterator")
		}
		iterVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown iterator %q", instruction.Args[0])
		}
		if iterVal.Object != nil {
			if elemVal, ok := iterVal.Object["__elements"]; ok && elemVal.Array != nil {
				idx := 0
				if idxVal, ok := iterVal.Object["__index"]; ok {
					idx = int(idxVal.Number)
				}
				if idx < len(elemVal.Array) {
					item := elemVal.Array[idx]
					iterVal.Object["__index"] = Value{Type: ir.TypeNumber, Number: float64(idx + 1)}
					env[instruction.Args[0]] = iterVal
					return Value{
						Type: ir.TypeObject,
						Object: map[string]Value{
							"value": item,
							"done":  {Type: ir.TypeBool, Bool: false},
						},
					}, nil
				}
				return Value{
					Type: ir.TypeObject,
					Object: map[string]Value{
						"value": {Type: ir.TypeVoid, String: "undefined"},
						"done":  {Type: ir.TypeBool, Bool: true},
					},
				}, nil
			}
		}
		return Value{
			Type: ir.TypeObject,
			Object: map[string]Value{
				"value": {Type: ir.TypeVoid, String: "undefined"},
				"done":  {Type: ir.TypeBool, Bool: true},
			},
		}, nil

	case "__iterator.return":
		return Value{
			Type: ir.TypeObject,
			Object: map[string]Value{
				"value": {Type: ir.TypeVoid, String: "undefined"},
				"done":  {Type: ir.TypeBool, Bool: true},
			},
		}, nil

	case "__iterator.throw":
		return Value{
			Type: ir.TypeObject,
			Object: map[string]Value{
				"value": {Type: ir.TypeVoid, String: "undefined"},
				"done":  {Type: ir.TypeBool, Bool: true},
			},
		}, nil
	}

	return Value{}, fmt.Errorf("unknown iterator intrinsic %q", instruction.Callee)
}
