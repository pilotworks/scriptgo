package interpreter

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func executeArrayTransformIntrinsic(name string, arguments []string, array Value, env map[string]Value, functions map[string]ir.Function, output *bytes.Buffer) (Value, bool, error) {
	switch name {
	case "__array.splice":
		if len(arguments) < 2 {
			return Value{}, true, fmt.Errorf("array.splice requires start")
		}
		startVal, ok := env[arguments[1]]
		if !ok || startVal.Type != ir.TypeNumber {
			return Value{}, true, fmt.Errorf("array.splice start must be a number")
		}
		start := int(startVal.Number)
		if start < 0 {
			start = max(len(array.Array)+start, 0)
		} else if start > len(array.Array) {
			start = len(array.Array)
		}
		deleteCount := len(array.Array) - start
		if len(arguments) >= 3 {
			dcVal, ok := env[arguments[2]]
			if ok && dcVal.Type == ir.TypeNumber {
				if dcVal.Number < 0 {
					deleteCount = 0
				} else {
					deleteCount = int(dcVal.Number)
					if start+deleteCount > len(array.Array) {
						deleteCount = len(array.Array) - start
					}
				}
			}
		}
		var insertItems []Value
		for i := 3; i < len(arguments); i++ {
			if item, ok := env[arguments[i]]; ok {
				insertItems = append(insertItems, item)
			}
		}
		deleted := append([]Value(nil), array.Array[start:start+deleteCount]...)
		newArray := make([]Value, 0, len(array.Array)-deleteCount+len(insertItems))
		newArray = append(newArray, array.Array[:start]...)
		newArray = append(newArray, insertItems...)
		newArray = append(newArray, array.Array[start+deleteCount:]...)
		array.Array = newArray
		array.SetArray(newArray)
		env[arguments[0]] = array
		return Value{Type: array.Type, Array: deleted}, true, nil

	case "__array.toSpliced":
		if len(arguments) < 2 {
			return Value{}, true, fmt.Errorf("array.toSpliced requires start")
		}
		startVal, ok := env[arguments[1]]
		if !ok || startVal.Type != ir.TypeNumber {
			return Value{}, true, fmt.Errorf("array.toSpliced start must be number")
		}
		start := int(startVal.Number)
		if start < 0 {
			start = max(len(array.Array)+start, 0)
		} else if start > len(array.Array) {
			start = len(array.Array)
		}
		deleteCount := len(array.Array) - start
		if len(arguments) >= 3 {
			dcVal, ok := env[arguments[2]]
			if ok && dcVal.Type == ir.TypeNumber {
				if dcVal.Number < 0 {
					deleteCount = 0
				} else {
					deleteCount = int(dcVal.Number)
					if start+deleteCount > len(array.Array) {
						deleteCount = len(array.Array) - start
					}
				}
			}
		}
		var insertItems []Value
		for i := 3; i < len(arguments); i++ {
			if item, ok := env[arguments[i]]; ok {
				insertItems = append(insertItems, item)
			}
		}
		newItems := make([]Value, 0, len(array.Array)-deleteCount+len(insertItems))
		newItems = append(newItems, array.Array[:start]...)
		newItems = append(newItems, insertItems...)
		newItems = append(newItems, array.Array[start+deleteCount:]...)
		return Value{Type: array.Type, Array: newItems}, true, nil

	case "__array.toSorted":
		sorted := make([]Value, len(array.Array))
		copy(sorted, array.Array)
		if len(arguments) > 1 {
			if closureVal, ok := env[arguments[1]]; ok && closureVal.Closure != nil {
				var sortErr error
				sort.SliceStable(sorted, func(i, j int) bool {
					if sortErr != nil {
						return false
					}
					res, _, err := executeClosure(functions, closureVal.Closure, []Value{sorted[i], sorted[j]}, output)
					if err != nil {
						sortErr = err
						return false
					}
					return res.Number < 0
				})
				if sortErr != nil {
					return Value{}, true, sortErr
				}
				return Value{Type: array.Type, Array: sorted}, true, nil
			}
		}
		sort.SliceStable(sorted, func(i, j int) bool {
			if sorted[i].Type == ir.TypeNumber && sorted[j].Type == ir.TypeNumber {
				return sorted[i].Number < sorted[j].Number
			}
			if sorted[i].Type == ir.TypeString && sorted[j].Type == ir.TypeString {
				return sorted[i].String < sorted[j].String
			}
			return false
		})
		return Value{Type: array.Type, Array: sorted}, true, nil

	case "__array.sort":
		if len(arguments) > 1 {
			if closureVal, ok := env[arguments[1]]; ok && closureVal.Closure != nil {
				var sortErr error
				sort.SliceStable(array.Array, func(i, j int) bool {
					if sortErr != nil {
						return false
					}
					res, _, err := executeClosure(functions, closureVal.Closure, []Value{array.Array[i], array.Array[j]}, output)
					if err != nil {
						sortErr = err
						return false
					}
					return res.Number < 0
				})
				if sortErr != nil {
					return Value{}, true, sortErr
				}
				array.SetArray(array.Array)
				env[arguments[0]] = array
				return array, true, nil
			}
		}
		sort.SliceStable(array.Array, func(i, j int) bool {
			if array.Array[i].Type == ir.TypeNumber && array.Array[j].Type == ir.TypeNumber {
				return array.Array[i].Number < array.Array[j].Number
			}
			if array.Array[i].Type == ir.TypeString && array.Array[j].Type == ir.TypeString {
				return array.Array[i].String < array.Array[j].String
			}
			return false
		})
		array.SetArray(array.Array)
		env[arguments[0]] = array
		return array, true, nil

	case "__array.with":
		if len(arguments) < 3 {
			return Value{}, true, fmt.Errorf("array.with requires index and value")
		}
		idxVal, ok := env[arguments[1]]
		if !ok || idxVal.Type != ir.TypeNumber {
			return Value{}, true, fmt.Errorf("array.with index must be number")
		}
		newVal, ok := env[arguments[2]]
		if !ok {
			return Value{}, true, fmt.Errorf("array.with value missing")
		}
		idx := int(idxVal.Number)
		if idx < 0 {
			idx = len(array.Array) + idx
		}
		if idx < 0 || idx >= len(array.Array) {
			return Value{}, true, fmt.Errorf("array.with index out of bounds")
		}
		newArr := make([]Value, len(array.Array))
		copy(newArr, array.Array)
		newArr[idx] = newVal
		return Value{Type: array.Type, Array: newArr}, true, nil

	case "__array.findLast":
		if len(arguments) < 2 {
			return Value{}, true, fmt.Errorf("array.findLast requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, true, fmt.Errorf("array.findLast callback must be a closure")
		}
		for i := len(array.Array) - 1; i >= 0; i-- {
			item := array.Array[i]
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, true, err
			}
			if flow == flowThrow {
				return res, true, fmt.Errorf("uncaught exception in array.findLast")
			}
			if res.Bool {
				return item, true, nil
			}
		}
		if array.Type == ir.TypeNumberArray {
			return Value{Type: ir.TypeNumber, Number: 0}, true, nil
		} else if array.Type == ir.TypeStringArray {
			return Value{Type: ir.TypeString, String: ""}, true, nil
		}
		return Value{Type: ir.Type(strings.TrimSuffix(string(array.Type), "[]"))}, true, nil

	case "__array.findLastIndex":
		if len(arguments) < 2 {
			return Value{}, true, fmt.Errorf("array.findLastIndex requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, true, fmt.Errorf("array.findLastIndex callback must be a closure")
		}
		for i := len(array.Array) - 1; i >= 0; i-- {
			item := array.Array[i]
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, true, err
			}
			if res.Bool {
				return Value{Type: ir.TypeNumber, Number: float64(i)}, true, nil
			}
		}
		return Value{Type: ir.TypeNumber, Number: -1}, true, nil

	case "__array.lastIndexOf":
		if len(arguments) < 2 {
			return Value{}, true, fmt.Errorf("array.lastIndexOf requires target")
		}
		target, ok := env[arguments[1]]
		if !ok {
			return Value{}, true, fmt.Errorf("unknown lastIndexOf target")
		}
		fromIndex := len(array.Array) - 1
		if len(arguments) >= 3 {
			if fVal, ok := env[arguments[2]]; ok && fVal.Type == ir.TypeNumber {
				fromIndex = int(fVal.Number)
				if fromIndex < 0 {
					fromIndex = len(array.Array) + fromIndex
				}
				if fromIndex >= len(array.Array) {
					fromIndex = len(array.Array) - 1
				}
			}
		}
		for i := fromIndex; i >= 0; i-- {
			if array.Type == ir.TypeNumberArray && array.Array[i].Number == target.Number {
				return Value{Type: ir.TypeNumber, Number: float64(i)}, true, nil
			} else if array.Type == ir.TypeStringArray && array.Array[i].String == target.String {
				return Value{Type: ir.TypeNumber, Number: float64(i)}, true, nil
			} else if array.Array[i].Type == target.Type && array.Array[i].Number == target.Number && array.Array[i].String == target.String {
				return Value{Type: ir.TypeNumber, Number: float64(i)}, true, nil
			}
		}
		return Value{Type: ir.TypeNumber, Number: -1}, true, nil

	case "__array.copyWithin":
		if len(arguments) < 2 {
			return Value{}, true, fmt.Errorf("array.copyWithin requires target")
		}
		targetVal, ok := env[arguments[1]]
		if !ok || targetVal.Type != ir.TypeNumber {
			return Value{}, true, fmt.Errorf("array.copyWithin target must be number")
		}
		target := int(targetVal.Number)
		if target < 0 {
			target = max(len(array.Array)+target, 0)
		} else if target > len(array.Array) {
			target = len(array.Array)
		}
		start := 0
		if len(arguments) > 2 {
			if sVal, ok := env[arguments[2]]; ok && sVal.Type == ir.TypeNumber {
				start = int(sVal.Number)
				if start < 0 {
					start = max(len(array.Array)+start, 0)
				} else if start > len(array.Array) {
					start = len(array.Array)
				}
			}
		}
		end := len(array.Array)
		if len(arguments) > 3 {
			if eVal, ok := env[arguments[3]]; ok && eVal.Type == ir.TypeNumber {
				end = int(eVal.Number)
				if end < 0 {
					end = max(len(array.Array)+end, 0)
				} else if end > len(array.Array) {
					end = len(array.Array)
				}
			}
		}
		count := minInt(end-start, len(array.Array)-target)
		if count > 0 && start < len(array.Array) {
			temp := make([]Value, count)
			copy(temp, array.Array[start:start+count])
			copy(array.Array[target:target+count], temp)
		}
		array.SetArray(array.Array)
		env[arguments[0]] = array
		return array, true, nil

	case "__array.reduceRight":
		if len(arguments) < 3 {
			return Value{}, true, fmt.Errorf("array.reduceRight requires callback closure and initial value")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, true, fmt.Errorf("array.reduceRight callback must be a closure")
		}
		acc, ok := env[arguments[2]]
		if !ok {
			return Value{}, true, fmt.Errorf("unknown reduceRight initial value")
		}
		for i := len(array.Array) - 1; i >= 0; i-- {
			item := array.Array[i]
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{acc, item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, true, err
			}
			if flow == flowThrow {
				return res, true, fmt.Errorf("uncaught exception in array.reduceRight")
			}
			acc = res
		}
		return acc, true, nil

	case "__array.flatMap":
		if len(arguments) < 2 {
			return Value{}, true, fmt.Errorf("array.flatMap requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, true, fmt.Errorf("array.flatMap callback must be a closure")
		}
		var flatMapped []Value
		for i, item := range array.Array {
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, true, err
			}
			if flow == flowThrow {
				return res, true, fmt.Errorf("uncaught exception in array.flatMap")
			}
			if len(res.GetArray()) > 0 {
				flatMapped = append(flatMapped, res.GetArray()...)
			} else {
				flatMapped = append(flatMapped, res)
			}
		}
		return Value{Type: array.Type, Array: flatMapped}, true, nil
	}
	return Value{}, false, nil
}
