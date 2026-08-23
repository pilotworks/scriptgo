package interpreter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

type microtaskItem struct {
	closure *Closure
	arg     Value
}

var microtasks []microtaskItem

func resetMicrotasks() {
	microtasks = nil
}

func executeAsyncIntrinsic(name string, arguments []string, env map[string]Value, functions map[string]ir.Function, output *bytes.Buffer) (Value, error) {
	switch name {
	case "__async.queueMicrotask":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("queueMicrotask requires 1 argument")
		}
		closureVal, ok := env[arguments[0]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("queueMicrotask argument must be a closure")
		}
		microtasks = append(microtasks, microtaskItem{closure: closureVal.Closure})
		return Value{Type: ir.TypeVoid}, nil
	case "__async.promise_create":
		return Value{Type: ir.TypeObject, Object: map[string]Value{"__state": {Type: ir.TypeNumber, Number: 0}, "__value": {Type: ir.TypeVoid}}}, nil
	case "__async.promise_try":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("Promise.try requires a callback")
		}
		closureVal, ok := env[arguments[0]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("Promise.try callback must be a closure")
		}
		retVal, _, err := executeClosure(functions, closureVal.Closure, nil, output)
		if err != nil {
			return Value{Type: ir.TypeObject, Object: map[string]Value{"__state": {Type: ir.TypeNumber, Number: 2}, "__value": {Type: ir.TypeString, String: err.Error()}}}, nil
		}
		return Value{Type: ir.TypeObject, Object: map[string]Value{"__state": {Type: ir.TypeNumber, Number: 1}, "__value": retVal}}, nil
	case "__async.promise_resolve":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("Promise.resolve requires 1 argument")
		}
		val, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown argument %q", arguments[0])
		}
		return Value{Type: ir.TypeObject, Object: map[string]Value{"__state": {Type: ir.TypeNumber, Number: 1}, "__value": val}}, nil
	case "__async.promise_then":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("promise.then requires promise and callback")
		}
		promiseVal, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown promise %q", arguments[0])
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("promise.then callback must be a closure")
		}
		var val Value
		if promiseVal.Object != nil {
			val = promiseVal.Object["__value"]
		}
		microtasks = append(microtasks, microtaskItem{closure: closureVal.Closure, arg: val})
		return promiseVal, nil
	case "__async.promise_catch":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("promise.catch requires promise and callback")
		}
		promiseVal, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown promise %q", arguments[0])
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("promise.catch callback must be a closure")
		}
		return promiseVal, nil
	case "__async.await":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("await requires 1 argument")
		}
		promiseVal, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown await argument %q", arguments[0])
		}
		if promiseVal.Type == ir.TypeObject && promiseVal.Object != nil {
			return promiseVal.Object["__value"], nil
		}
		return promiseVal, nil
	case "__async.array_from_async":
		if len(arguments) < 1 {
			return Value{}, fmt.Errorf("Array.fromAsync requires at least 1 argument")
		}
		sourceVal, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown source %q", arguments[0])
		}
		var mapClosure *Closure
		if len(arguments) > 1 {
			if cv, exists := env[arguments[1]]; exists && cv.Closure != nil {
				mapClosure = cv.Closure
			}
		}

		var items []Value
		arr := sourceVal.GetArray()
		if len(arr) > 0 {
			items = arr
		} else if sourceVal.Type == ir.TypeString {
			for _, r := range sourceVal.String {
				items = append(items, Value{Type: ir.TypeString, String: string(r)})
			}
		} else if sourceVal.Object != nil {
			if it, hasIt := sourceVal.Object["__items"]; hasIt {
				items = it.GetArray()
			} else if it, hasIt := sourceVal.Object["__elements"]; hasIt {
				items = it.GetArray()
			}
		}

		var resultItems []Value
		for i, itm := range items {
			val := itm
			if val.Type == ir.TypeObject && val.Object != nil {
				if v, hasVal := val.Object["__value"]; hasVal {
					val = v
				}
			}
			if mapClosure != nil {
				ret, _, err := executeClosure(functions, mapClosure, []Value{val, {Type: ir.TypeNumber, Number: float64(i)}}, output)
				if err != nil {
					return Value{}, err
				}
				if ret.Type == ir.TypeObject && ret.Object != nil {
					if v, hasVal := ret.Object["__value"]; hasVal {
						ret = v
					}
				}
				val = ret
			}
			resultItems = append(resultItems, val)
		}

		elemType := ir.TypeNumber
		if len(resultItems) > 0 {
			elemType = resultItems[0].Type
		}
		arrVal := Value{
			Type:  ir.Type(string(elemType) + "[]"),
			Array: resultItems,
		}
		return Value{
			Type: ir.TypeObject,
			Object: map[string]Value{
				"__state": {Type: ir.TypeNumber, Number: 1},
				"__value": arrVal,
			},
		}, nil
	default:
		return Value{}, fmt.Errorf("unknown async intrinsic %q", name)
	}
}

func drainMicrotasks(functions map[string]ir.Function, output *bytes.Buffer) error {
	for len(microtasks) > 0 {
		task := microtasks[0]
		microtasks = microtasks[1:]
		var args []Value
		if task.arg.Type != "" {
			args = []Value{task.arg}
		}
		_, _, err := executeClosure(functions, task.closure, args, output)
		if err != nil {
			return err
		}
	}
	return nil
}

func executeGeneratorIntrinsic(instruction ir.Instruction, env map[string]Value, functions map[string]ir.Function, output *bytes.Buffer) (Value, error) {
	switch instruction.Callee {
	case "__generator.next":
		if len(instruction.Args) < 1 {
			return Value{}, fmt.Errorf("__generator.next requires generator argument")
		}
		genVal, ok := env[instruction.Args[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown generator %q", instruction.Args[0])
		}
		if genVal.Closure != nil {
			val, _, err := executeClosure(functions, genVal.Closure, nil, output)
			return val, err
		}
		if genVal.Object != nil {
			if itemsVal, hasItems := genVal.Object["__items"]; hasItems && len(itemsVal.Array) > 0 {
				state := 0
				if sVal, hasState := genVal.Object["__state"]; hasState {
					state = int(sVal.Number)
				}
				if state < len(itemsVal.Array) {
					item := itemsVal.Array[state]
					genVal.Object["__state"] = Value{Type: ir.TypeNumber, Number: float64(state + 1)}
					env[instruction.Args[0]] = genVal
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
						"value": {},
						"done":  {Type: ir.TypeBool, Bool: true},
					},
				}, nil
			}
			clsName := genVal.String
			if clsName == "" {
				clsName = strings.TrimPrefix(string(genVal.Type), "object:")
			}
			nextFnName := clsName + "_next"
			if nextFn, ok := functions[nextFnName]; ok {
				res, _, err := executeFunction(functions, nextFn, []Value{genVal}, output)
				if err != nil {
					return Value{}, err
				}
				env[instruction.Args[0]] = genVal
				return res, nil
			}
		}
		return Value{
			Type: ir.TypeObject,
			Object: map[string]Value{
				"value": {},
				"done":  {Type: ir.TypeBool, Bool: true},
			},
		}, nil
	}
	return Value{}, fmt.Errorf("unsupported generator intrinsic %q", instruction.Callee)
}
