package interpreter

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func executeArrayIntrinsic(name string, arguments []string, env map[string]Value, functions map[string]ir.Function, output *bytes.Buffer) (Value, error) {
	if name == "__array.isArray" {
		if len(arguments) == 0 {
			return Value{Type: ir.TypeBool, Bool: false}, nil
		}
		arg, ok := env[arguments[0]]
		if !ok {
			return Value{Type: ir.TypeBool, Bool: false}, nil
		}
		for arg.Type == ir.TypeUnknown && arg.Boxed != nil {
			arg = *arg.Boxed
		}
		isArr := strings.HasSuffix(string(arg.Type), "[]") || strings.Contains(string(arg.Type), "[]") || strings.Contains(string(arg.Type), "__shape_0_") || arg.Type == ir.TypeNumberArray || arg.Type == ir.TypeStringArray || arg.Type == ir.TypeBoolArray || arg.Type == ir.TypeBigIntArray || arg.Type == ir.TypeSymbolArray || arg.ArrayRef != nil || len(arg.Array) > 0 || len(arg.GetArray()) > 0 || arg.TypedArray != nil || strings.Contains(string(arg.Type), "Array")
		return Value{Type: ir.TypeBool, Bool: isArr}, nil
	}
	if name == "__array.of" {
		var arr []Value
		elemType := ir.TypeNumber
		for _, argName := range arguments {
			elem, ok := env[argName]
			if !ok {
				return Value{}, fmt.Errorf("unknown array element %q", argName)
			}
			arr = append(arr, elem)
			elemType = elem.Type
		}
		ref := new([]Value)
		*ref = arr
		return Value{Type: ir.Type(string(elemType) + "[]"), Array: arr, ArrayRef: ref}, nil
	}
	if name == "__array.from" {
		if len(arguments) == 0 {
			return Value{}, fmt.Errorf("array.from requires at least 1 argument")
		}
		arg, ok := env[arguments[0]]
		if !ok {
			return Value{}, fmt.Errorf("unknown array.from argument %q", arguments[0])
		}
		for arg.Type == ir.TypeUnknown && arg.Boxed != nil {
			arg = *arg.Boxed
		}
		if arg.Type == ir.TypeString {
			var arr []Value
			for _, r := range arg.String {
				arr = append(arr, Value{Type: ir.TypeString, String: string(r)})
			}
			ref := new([]Value)
			*ref = arr
			return Value{Type: ir.TypeStringArray, Array: arr, ArrayRef: ref}, nil
		}
		newArr := append([]Value(nil), arg.GetArray()...)
		ref := new([]Value)
		*ref = newArr
		return Value{Type: arg.Type, Array: newArr, ArrayRef: ref}, nil
	}
	if len(arguments) == 0 {
		return Value{}, fmt.Errorf("array intrinsic requires at least one argument")
	}
	array, ok := env[arguments[0]]
	if !ok {
		return Value{}, fmt.Errorf("array intrinsic requires an array")
	}
	for array.Type == ir.TypeUnknown && array.Boxed != nil {
		array = *array.Boxed
	}
	if !strings.HasSuffix(string(array.Type), "[]") && !strings.Contains(string(array.Type), "__shape_0_") && array.Type != ir.TypeNumberArray && array.Type != ir.TypeStringArray && array.Type != ir.TypeBoolArray && array.Type != ir.TypeBigIntArray && array.ArrayRef == nil && len(array.Array) == 0 {
		return Value{}, fmt.Errorf("array intrinsic requires an array")
	}
	array.Array = array.GetArray()
	if res, handled, err := executeArrayTransformIntrinsic(name, arguments, array, env, functions, output); handled {
		return res, err
	}
	switch name {
	case "__array.length":
		if array.TypedArray != nil {
			return Value{Type: ir.TypeNumber, Number: float64(array.TypedArray.Length)}, nil
		}
		return Value{Type: ir.TypeNumber, Number: float64(len(array.GetArray()))}, nil
	case "__array.push":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("array.push requires array and element")
		}
		elem, ok := env[arguments[1]]
		if !ok {
			return Value{}, fmt.Errorf("unknown push argument %q", arguments[1])
		}
		arr := append(array.GetArray(), elem)
		array.SetArray(arr)
		env[arguments[0]] = array
		return Value{Type: ir.TypeNumber, Number: float64(len(arr))}, nil
	case "__array.pop":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("array.pop requires 1 argument")
		}
		arr := array.GetArray()
		if len(arr) == 0 {
			if array.Type == ir.TypeNumberArray {
				return Value{Type: ir.TypeNumber, Number: 0}, nil
			}
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		last := arr[len(arr)-1]
		arr = arr[:len(arr)-1]
		array.SetArray(arr)
		env[arguments[0]] = array
		return last, nil
	case "__array.slice":
		if len(arguments) < 2 || len(arguments) > 3 {
			return Value{}, fmt.Errorf("array.slice requires start and optional end")
		}
		startVal, ok := env[arguments[1]]
		if !ok || startVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("array.slice start must be a number")
		}
		arr := array.GetArray()
		start := int(startVal.Number)
		if start < 0 {
			start = max(len(arr)+start, 0)
		} else if start > len(arr) {
			start = len(arr)
		}
		end := len(arr)
		if len(arguments) == 3 {
			endVal, ok := env[arguments[2]]
			if !ok || endVal.Type != ir.TypeNumber {
				return Value{}, fmt.Errorf("array.slice end must be a number")
			}
			end = int(endVal.Number)
			if end < 0 {
				end = max(len(arr)+end, 0)
			} else if end > len(arr) {
				end = len(arr)
			}
		}
		if end < start {
			end = start
		}
		sub := append([]Value(nil), arr[start:end]...)
		return Value{Type: array.Type, Array: sub}, nil
	case "__array.indexOf":
		if len(arguments) < 2 || len(arguments) > 3 {
			return Value{}, fmt.Errorf("array.indexOf requires target and optional fromIndex")
		}
		target, ok := env[arguments[1]]
		if !ok {
			return Value{}, fmt.Errorf("unknown indexOf target")
		}
		start := 0
		if len(arguments) == 3 {
			fromVal, ok := env[arguments[2]]
			if ok && fromVal.Type == ir.TypeNumber && fromVal.Number > 0 {
				start = int(fromVal.Number)
			}
		}
		idx := -1
		arr := array.GetArray()
		for i := start; i < len(arr); i++ {
			if array.Type == ir.TypeNumberArray && arr[i].Number == target.Number {
				idx = i
				break
			} else if array.Type == ir.TypeStringArray && arr[i].String == target.String {
				idx = i
				break
			}
		}
		return Value{Type: ir.TypeNumber, Number: float64(idx)}, nil
	case "__array.includes":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("array.includes requires target")
		}
		target, ok := env[arguments[1]]
		if !ok {
			return Value{}, fmt.Errorf("unknown includes target")
		}
		found := false
		arr := array.GetArray()
		for i := 0; i < len(arr); i++ {
			if array.Type == ir.TypeNumberArray && arr[i].Number == target.Number {
				found = true
				break
			} else if array.Type == ir.TypeStringArray && arr[i].String == target.String {
				found = true
				break
			}
		}
		return Value{Type: ir.TypeBool, Bool: found}, nil
	case "__array.at":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("array.at requires index")
		}
		idxVal, ok := env[arguments[1]]
		if !ok || idxVal.Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("array.at index must be a number")
		}
		arr := array.GetArray()
		idx := int(idxVal.Number)
		if idx < 0 {
			idx = len(arr) + idx
		}
		if idx < 0 || idx >= len(arr) {
			if array.Type == ir.TypeNumberArray {
				return Value{Type: ir.TypeNumber, Number: 0}, nil
			}
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		return arr[idx], nil
	case "__array.shift":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("array.shift requires 1 argument")
		}
		arr := array.GetArray()
		if len(arr) == 0 {
			if array.Type == ir.TypeNumberArray {
				return Value{Type: ir.TypeNumber, Number: 0}, nil
			}
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		first := arr[0]
		arr = arr[1:]
		array.SetArray(arr)
		env[arguments[0]] = array
		return first, nil
	case "__array.unshift":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("array.unshift requires array and element")
		}
		elem, ok := env[arguments[1]]
		if !ok {
			return Value{}, fmt.Errorf("unknown unshift argument %q", arguments[1])
		}
		arr := append([]Value{elem}, array.GetArray()...)
		array.SetArray(arr)
		env[arguments[0]] = array
		return Value{Type: ir.TypeNumber, Number: float64(len(arr))}, nil
	case "__array.reverse":
		if len(arguments) != 1 {
			return Value{}, fmt.Errorf("array.reverse requires 1 argument")
		}
		arr := array.GetArray()
		for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
			arr[i], arr[j] = arr[j], arr[i]
		}
		array.SetArray(arr)
		env[arguments[0]] = array
		return array, nil
	case "__array.concat":
		if len(arguments) != 2 {
			return Value{}, fmt.Errorf("array.concat requires other array")
		}
		other, ok := env[arguments[1]]
		if !ok || other.Type != array.Type {
			return Value{}, fmt.Errorf("array.concat requires matching array type")
		}
		newItems := make([]Value, 0, len(array.Array)+len(other.Array))
		newItems = append(newItems, array.Array...)
		newItems = append(newItems, other.Array...)
		return Value{Type: array.Type, Array: newItems, ArrayRef: &newItems}, nil
	case "__array.join":
		sep := ","
		if len(arguments) >= 2 {
			sepVal, ok := env[arguments[1]]
			if ok && sepVal.Type == ir.TypeString {
				sep = sepVal.String
			}
		}
		parts := make([]string, len(array.Array))
		for i, item := range array.Array {
			if item.Type == ir.TypeNumber {
				parts[i] = strconv.FormatFloat(item.Number, 'f', -1, 64)
			} else if item.Type == ir.TypeBigInt {
				parts[i] = strconv.FormatInt(item.BigInt, 10)
			} else if item.Type == ir.TypeBool {
				parts[i] = strconv.FormatBool(item.Bool)
			} else {
				parts[i] = item.String
			}
		}
		return Value{Type: ir.TypeString, String: strings.Join(parts, sep)}, nil
	case "__array.map":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.map requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.map callback must be a closure")
		}
		mapped := make([]Value, 0, len(array.Array))
		for i, item := range array.Array {
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return res, fmt.Errorf("uncaught exception in array.map")
			}
			mapped = append(mapped, res)
		}
		retType := array.Type
		if len(mapped) > 0 {
			if mapped[0].Type == ir.TypeString {
				retType = ir.TypeStringArray
			} else {
				retType = ir.TypeNumberArray
			}
		}
		return Value{Type: retType, Array: mapped}, nil
	case "__array.filter":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.filter requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.filter callback must be a closure")
		}
		filtered := make([]Value, 0)
		for i, item := range array.Array {
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return res, fmt.Errorf("uncaught exception in array.filter")
			}
			if res.Bool {
				filtered = append(filtered, item)
			}
		}
		return Value{Type: array.Type, Array: filtered}, nil
	case "__array.forEach":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.forEach requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.forEach callback must be a closure")
		}
		for i, item := range array.Array {
			_, flow, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return Value{}, fmt.Errorf("uncaught exception in array.forEach")
			}
		}
		return Value{Type: ir.TypeVoid}, nil
	case "__array.reduce":
		if len(arguments) < 3 {
			return Value{}, fmt.Errorf("array.reduce requires callback closure and initial value")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.reduce callback must be a closure")
		}
		acc, ok := env[arguments[2]]
		if !ok {
			return Value{}, fmt.Errorf("unknown reduce initial value")
		}
		for i, item := range array.Array {
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{acc, item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return res, fmt.Errorf("uncaught exception in array.reduce")
			}
			acc = res
		}
		return acc, nil
	case "__array.find":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.find requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.find callback must be a closure")
		}
		for i, item := range array.Array {
			res, flow, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if flow == flowThrow {
				return res, fmt.Errorf("uncaught exception in array.find")
			}
			if res.Bool {
				return item, nil
			}
		}
		if array.Type == ir.TypeNumberArray {
			return Value{Type: ir.TypeNumber, Number: 0}, nil
		} else if array.Type == ir.TypeStringArray {
			return Value{Type: ir.TypeString, String: ""}, nil
		}
		elemType := ir.Type(strings.TrimSuffix(string(array.Type), "[]"))
		if !strings.HasPrefix(string(elemType), "object:") && elemType != ir.TypeNumber && elemType != ir.TypeString && elemType != ir.TypeBool && elemType != ir.TypeBigInt {
			elemType = ir.Type("object:" + string(elemType))
		}
		return Value{Type: elemType}, nil
	case "__array.some":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.some requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.some callback must be a closure")
		}
		for i, item := range array.Array {
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if res.Bool {
				return Value{Type: ir.TypeBool, Bool: true}, nil
			}
		}
		return Value{Type: ir.TypeBool, Bool: false}, nil
	case "__array.every":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.every requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.every callback must be a closure")
		}
		for i, item := range array.Array {
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if !res.Bool {
				return Value{Type: ir.TypeBool, Bool: false}, nil
			}
		}
		return Value{Type: ir.TypeBool, Bool: true}, nil
	case "__array.findIndex":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.findIndex requires callback closure")
		}
		closureVal, ok := env[arguments[1]]
		if !ok || closureVal.Closure == nil {
			return Value{}, fmt.Errorf("array.findIndex callback must be a closure")
		}
		for i, item := range array.Array {
			res, _, err := executeClosure(functions, closureVal.Closure, []Value{item, {Type: ir.TypeNumber, Number: float64(i)}}, output)
			if err != nil {
				return Value{}, err
			}
			if res.Bool {
				return Value{Type: ir.TypeNumber, Number: float64(i)}, nil
			}
		}
		return Value{Type: ir.TypeNumber, Number: -1}, nil
	case "__array.fill":
		if len(arguments) < 2 {
			return Value{}, fmt.Errorf("array.fill requires value argument")
		}
		fillVal, ok := env[arguments[1]]
		if !ok {
			return Value{}, fmt.Errorf("unknown fill value %q", arguments[1])
		}
		start := 0
		end := len(array.Array)
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
		for i := start; i < end; i++ {
			array.Array[i] = fillVal
		}
		array.SetArray(array.Array)
		env[arguments[0]] = array
		return array, nil
	case "__array.toReversed":
		reversed := make([]Value, len(array.Array))
		for i := range array.Array {
			reversed[i] = array.Array[len(array.Array)-1-i]
		}
		return Value{Type: array.Type, Array: reversed}, nil
	case "__array.toString", "__array.toLocaleString":
		parts := make([]string, len(array.Array))
		for i, item := range array.Array {
			if item.Type == ir.TypeNumber {
				parts[i] = strconv.FormatFloat(item.Number, 'f', -1, 64)
			} else if item.Type == ir.TypeBigInt {
				parts[i] = strconv.FormatInt(item.BigInt, 10)
			} else if item.Type == ir.TypeBool {
				parts[i] = strconv.FormatBool(item.Bool)
			} else {
				parts[i] = item.String
			}
		}
		return Value{Type: ir.TypeString, String: strings.Join(parts, ",")}, nil
	case "__array.flat":
		var flatItems []Value
		for _, item := range array.Array {
			if len(item.GetArray()) > 0 {
				flatItems = append(flatItems, item.GetArray()...)
			} else {
				flatItems = append(flatItems, item)
			}
		}
		return Value{Type: array.Type, Array: flatItems}, nil
	case "__array.entries":
		var pairs []Value
		for i, item := range array.Array {
			pairs = append(pairs, Value{Type: ir.TypeString, String: fmt.Sprintf("[%d, %s]", i, format(item))})
		}
		return Value{Type: ir.TypeStringArray, Array: pairs}, nil
	case "__array.keys":
		keys := make([]Value, len(array.Array))
		for i := range array.Array {
			keys[i] = Value{Type: ir.TypeNumber, Number: float64(i)}
		}
		return Value{Type: ir.TypeNumberArray, Array: keys}, nil
	case "__array.values":
		vals := append([]Value(nil), array.Array...)
		return Value{Type: array.Type, Array: vals}, nil
	default:
		return Value{}, fmt.Errorf("unknown array intrinsic %q", name)
	}
}
