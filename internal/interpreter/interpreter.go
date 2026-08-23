// Package interpreter provides the reference execution engine for typed IR.
// It is a test oracle and is not linked into native executables.
package interpreter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/pilotworks/scriptgo/internal/ir"
)

var execMu sync.Mutex

// Execute interprets a verified module using the reference IR semantics.
func Execute(module ir.Module) (Result, error) {
	execMu.Lock()
	defer execMu.Unlock()

	if err := module.Verify(); err != nil {
		return Result{}, err
	}
	functions := make(map[string]ir.Function, len(module.Functions))
	for _, function := range module.Functions {
		functions[function.Name] = function
	}
	main, ok := functions["main"]
	if !ok {
		return Result{}, fmt.Errorf("module has no main function")
	}
	resetMicrotasks()
	resetTimers()
	resetStreamDefaults()
	var output bytes.Buffer
	value, flow, err := executeFunction(functions, main, nil, &output)
	if err != nil {
		return Result{}, err
	}
	if flow.kind == kindThrow {
		return Result{}, fmt.Errorf("uncaught exception: %s", format(value))
	}
	if err := drainTimers(functions, &output); err != nil {
		return Result{}, err
	}
	return Result{Output: output.String(), Return: value}, nil
}

type flowKind int

const (
	kindNormal flowKind = iota
	kindReturn
	kindBreak
	kindContinue
	kindThrow
	kindExit
)

type controlFlow struct {
	kind   flowKind
	target string
}

var (
	flowNormal = controlFlow{kind: kindNormal}
	flowReturn = controlFlow{kind: kindReturn}
	flowThrow  = controlFlow{kind: kindThrow}
	flowExit   = controlFlow{kind: kindExit}
)

func executeFunction(functions map[string]ir.Function, function ir.Function, arguments []Value, output *bytes.Buffer) (Value, controlFlow, error) {
	env := make(map[string]Value, len(function.Parameters))
	if len(arguments) < len(function.Parameters) {
		for i := len(arguments); i < len(function.Parameters); i++ {
			switch function.Parameters[i].Type {
			case ir.TypeUnknown:
				arguments = append(arguments, Value{Type: ir.TypeUnknown, Boxed: &Value{Type: ir.TypeString, String: "undefined"}})
			case ir.TypeNumber:
				arguments = append(arguments, Value{Type: ir.TypeNumber, Number: 0})
			case ir.TypeBool:
				arguments = append(arguments, Value{Type: ir.TypeBool, Bool: false})
			default:
				arguments = append(arguments, Value{Type: ir.TypeString, String: "undefined"})
			}
		}
	} else if len(arguments) > len(function.Parameters) {
		return Value{}, flowNormal, fmt.Errorf("function %q received %d arguments, want %d", function.Name, len(arguments), len(function.Parameters))
	}
	for index, parameter := range function.Parameters {
		argType := arguments[index].Type
		paramType := parameter.Type
		if argType != paramType {
			if !(strings.HasPrefix(string(argType), "object:") && strings.HasPrefix(string(paramType), "object:")) &&
				!(argType == ir.TypeBuffer && paramType == ir.TypeUint8Array) &&
				!(isValNullish(arguments[index]) && (strings.HasPrefix(string(paramType), "object:") || paramType == "ptr" || paramType == ir.TypeClosure)) {
				return Value{}, flowNormal, fmt.Errorf("argument %d to %q has type %s, want %s", index, function.Name, argType, paramType)
			}
		}
		env[parameter.Name] = arguments[index]
	}

	val, _, flow, err := executeBlock(functions, function.Body, env, output)
	return val, flow, err
}

func executeClosure(functions map[string]ir.Function, closure *Closure, arguments []Value, output *bytes.Buffer) (Value, controlFlow, error) {
	if closure == nil {
		return Value{}, flowNormal, fmt.Errorf("cannot execute nil closure")
	}
	closureEnv := make(map[string]Value, len(closure.Env)+len(arguments))
	for k, v := range closure.Env {
		closureEnv[k] = v
	}
	for k, ref := range closure.RefEnv {
		if ref != nil {
			closureEnv[k] = *ref
		}
	}
	userParams := closure.Function.Parameters
	if len(userParams) > 0 && userParams[0].Name == "__env_ctx" {
		userParams = userParams[1:]
	}
	for index, parameter := range userParams {
		if index < len(arguments) {
			arg := arguments[index]
			if arg.Type == ir.TypeUnknown && arg.Boxed != nil && parameter.Type != ir.TypeUnknown {
				arg = *arg.Boxed
			}
			closureEnv[parameter.Name] = arg
		} else {
			closureEnv[parameter.Name] = Value{Type: ir.TypeString, String: "undefined"}
		}
	}
	val, _, flow, err := executeBlock(functions, closure.Function.Body, closureEnv, output)
	for k, v := range closureEnv {
		if ref, ok := closure.RefEnv[k]; ok && ref != nil {
			*ref = v
		}
		if _, ok := closure.Env[k]; ok {
			closure.Env[k] = v
		}
	}
	return val, flow, err
}

func executeBlock(functions map[string]ir.Function, body []ir.Instruction, env map[string]Value, output *bytes.Buffer) (Value, bool, controlFlow, error) {
	cellMap := make(map[string]*Value)
	for _, instruction := range body {
		switch instruction.Op {
		case ir.OpClosure:
			targetFn, ok := functions[instruction.Callee]
			if !ok {
				return Value{}, false, flowNormal, fmt.Errorf("unknown function for closure %q", instruction.Callee)
			}
			captured := make(map[string]Value, len(instruction.Args))
			capturedRefs := make(map[string]*Value, len(instruction.Args))
			for _, argName := range instruction.Args {
				val, err := lookup(env, []string{argName}, 0)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				captured[argName] = val
				if cell, ok := cellMap[argName]; ok && cell != nil {
					capturedRefs[argName] = cell
				} else {
					cell := new(Value)
					*cell = val
					cellMap[argName] = cell
					capturedRefs[argName] = cell
				}
			}
			env[instruction.Result] = Value{
				Type: ir.TypeClosure,
				Closure: &Closure{
					Function: targetFn,
					Env:      captured,
					RefEnv:   capturedRefs,
				},
			}
		case ir.OpClosureCall:
			closureVal, err := lookup(env, []string{instruction.Callee}, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if closureVal.Type == ir.TypeUnknown && closureVal.Boxed != nil {
				closureVal = *closureVal.Boxed
			}
			if closureVal.Closure == nil {
				return Value{}, false, flowNormal, fmt.Errorf("%q is not a callable closure at span %+v (type=%s, val=%+v)", instruction.Callee, instruction.Span, closureVal.Type, closureVal)
			}
			args := make([]Value, 0, len(instruction.Args))
			for _, argName := range instruction.Args {
				val, err := lookup(env, []string{argName}, 0)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				args = append(args, val)
			}
			val, flow, err := executeClosure(functions, closureVal.Closure, args, output)
			if err != nil {
				return Value{}, false, flow, err
			}
			if flow != flowNormal && flow != flowReturn {
				return val, true, flow, nil
			}
			if instruction.Result != "" {
				env[instruction.Result] = val
			}
		case ir.OpConst:
			value, err := parseConstant(instruction.Type, instruction.Value)
			if err != nil {
				return Value{}, false, flowNormal, fmt.Errorf("%s: %w", instruction.Result, err)
			}
			env[instruction.Result] = value
		case ir.OpAssign:
			val, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			env[instruction.Result] = val
			if cell, ok := cellMap[instruction.Result]; ok && cell != nil {
				*cell = val
			}
		case ir.OpBinary:
			left, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			right, err := lookup(env, instruction.Args, 1)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			value, err := binary(instruction.Operator, left, right)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			env[instruction.Result] = value
		case ir.OpCompare:
			left, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			right, err := lookup(env, instruction.Args, 1)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			value, err := compare(instruction.Operator, left, right)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			env[instruction.Result] = value
		case ir.OpSelect:
			condition, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			selectedName := instruction.Args[2]
			if condition.Bool {
				selectedName = instruction.Args[1]
			}
			selected, err := lookup(env, []string{selectedName}, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			env[instruction.Result] = selected
		case ir.OpArray:
			array := make([]Value, 0)
			for _, arg := range instruction.Args {
				val, err := lookup(env, []string{arg}, 0)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				array = append(array, val)
			}
			ref := new([]Value)
			*ref = array
			env[instruction.Result] = Value{Type: instruction.Type, Array: array, ArrayRef: ref}
		case ir.OpIndex:
			array, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if array.Type == ir.TypeUnknown && array.Boxed != nil {
				array = *array.Boxed
			}
			index, err := lookup(env, instruction.Args, 1)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if index.Type != ir.TypeNumber || math.Trunc(index.Number) != index.Number || index.Number < 0 {
				return Value{}, false, flowNormal, fmt.Errorf("array index must be a non-negative integer, got %v", index.Number)
			}
			position := int(index.Number)
			if array.TypedArray != nil {
				if array.TypedArray.Kind == ir.TypeBigInt64Array || array.TypedArray.Kind == ir.TypeBigUint64Array {
					env[instruction.Result] = Value{Type: ir.TypeBigInt, BigInt: array.TypedArray.GetBigInt(position)}
				} else {
					env[instruction.Result] = Value{Type: ir.TypeNumber, Number: array.TypedArray.Get(position)}
				}
				continue
			}
			arr := array.GetArray()
			if len(arr) == 0 && array.Object != nil {
				posKey := strconv.Itoa(position)
				if val, ok := array.Object[posKey]; ok {
					env[instruction.Result] = val
					continue
				}
			}
			if position >= len(arr) {
				return Value{}, false, flowNormal, fmt.Errorf("array index %d out of bounds for length %d", position, len(arr))
			}
			env[instruction.Result] = arr[position]
		case ir.OpIndexSet:
			array, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if array.Type == ir.TypeUnknown && array.Boxed != nil {
				array = *array.Boxed
			}
			index, err := lookup(env, instruction.Args, 1)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			val, err := lookup(env, instruction.Args, 2)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if index.Type != ir.TypeNumber || math.Trunc(index.Number) != index.Number || index.Number < 0 {
				return Value{}, false, flowNormal, fmt.Errorf("array index must be a non-negative integer, got %v", index.Number)
			}
			position := int(index.Number)
			if array.TypedArray != nil {
				if array.TypedArray.Kind == ir.TypeBigInt64Array || array.TypedArray.Kind == ir.TypeBigUint64Array {
					array.TypedArray.SetBigInt(position, val.BigInt)
				} else {
					array.TypedArray.Set(position, val.Number)
				}
				continue
			}
			arr := array.GetArray()
			if position >= len(arr) {
				return Value{}, false, flowNormal, fmt.Errorf("array index %d out of bounds for length %d", position, len(arr))
			}
			arr[position] = val
			array.SetArray(arr)
			env[instruction.Args[0]] = array
		case ir.OpObjectNew:
			env[instruction.Result] = Value{Type: instruction.Type, Object: map[string]Value{}, String: instruction.Value}
		case ir.OpInstanceOf:
			object, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if object.Type == ir.TypeUnknown && object.Boxed != nil {
				object = *object.Boxed
			}
			isInstance := false
			if object.Type != "" {
				if object.String != "" && strings.Contains(object.String, ":"+instruction.Value+":") {
					isInstance = true
				} else if strings.HasPrefix(string(object.Type), "object:"+instruction.Value) {
					isInstance = true
				}
			}
			env[instruction.Result] = Value{Type: ir.TypeBool, Bool: isInstance}
		case ir.OpFieldSet:
			object, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			value, err := lookup(env, instruction.Args, 1)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if len(object.Array) > 0 {
				idx, err := strconv.Atoi(instruction.Field)
				if err == nil && idx >= 0 && idx < len(object.Array) {
					object.Array[idx] = value
					env[instruction.Args[0]] = object
					continue
				}
			}
			if object.Object == nil {
				object.Object = make(map[string]Value)
			}
			object.Object[instruction.Field] = value
			env[instruction.Args[0]] = object
		case ir.OpFieldGet:
			object, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if instruction.Field == "length" && (object.ArrayRef != nil || len(object.Array) > 0 || strings.HasSuffix(string(object.Type), "[]")) {
				env[instruction.Result] = Value{Type: ir.TypeNumber, Number: float64(len(object.GetArray()))}
				continue
			}
			if len(object.Array) > 0 {
				idx, err := strconv.Atoi(instruction.Field)
				if err == nil && idx >= 0 && idx < len(object.Array) {
					env[instruction.Result] = object.Array[idx]
					continue
				}
			}
			value, ok := object.Object[instruction.Field]
			if !ok {
				value = Value{Type: ir.TypeString, String: "undefined"}
			} else if value.ArrayRef != nil {
				value.Array = *value.ArrayRef
			}
			env[instruction.Result] = value
		case ir.OpBoxUnknown:
			val, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if val.Type == ir.TypeUnknown {
				env[instruction.Result] = val
			} else {
				env[instruction.Result] = Value{
					Type:  ir.TypeUnknown,
					Boxed: &val,
				}
			}
		case ir.OpCheckedCast:
			val, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			casted, err := castValue(val, instruction.Type)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			env[instruction.Result] = casted
		case ir.OpTypeOf:
			val, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if val.Type == ir.TypeUnknown && val.Boxed != nil {
				val = *val.Boxed
			}
			var typeStr string
			if isValNullish(val) {
				if val.String == "undefined" || val.Type == ir.TypeVoid {
					typeStr = "undefined"
				} else {
					typeStr = "object"
				}
			} else {
				switch val.Type {
				case ir.TypeNumber:
					typeStr = "number"
				case ir.TypeBigInt:
					typeStr = "bigint"
				case ir.TypeSymbol:
					typeStr = "symbol"
				case ir.TypeString:
					typeStr = "string"
				case ir.TypeBool:
					typeStr = "boolean"
				case ir.TypeVoid:
					typeStr = "undefined"
				case ir.TypeClosure:
					typeStr = "function"
				default:
					typeStr = "object"
				}
			}
			env[instruction.Result] = Value{Type: ir.TypeString, String: typeStr}
		case ir.OpPrint:
			value, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			indent := getConsoleIndent()
			fmt.Fprintf(output, "%s%s\n", indent, format(value))
		case ir.OpCall:
			if strings.HasPrefix(instruction.Callee, "__console.") {
				value, err := executeConsoleIntrinsic(instruction.Callee, instruction.Args, env, output)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__Math.") {
				value, err := executeMathIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				env[instruction.Result] = value
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__number.") {
				value, err := executeNumberIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				env[instruction.Result] = value
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__array.") {
				value, err := executeArrayIntrinsic(instruction.Callee, instruction.Args, env, functions, output)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				env[instruction.Result] = value
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__stream.") {
				value, err := executeStreamIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__object.") {
				value, err := executeObjectIntrinsic(functions, instruction.Callee, instruction.Args, env, output)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				env[instruction.Result] = value
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__clone.") {
				if len(instruction.Args) > 0 {
					val := env[instruction.Args[0]]
					env[instruction.Result] = deepCloneValue(val)
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__map.") {
				value, err := executeMapIntrinsic(instruction, env, functions, output)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__set.") {
				value, err := executeSetIntrinsic(instruction, env, functions, output)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__typedarray.") || strings.HasPrefix(instruction.Callee, "__arraybuffer.") || strings.HasPrefix(instruction.Callee, "__dataview.") {
				value, err := executeTypedArrayIntrinsic(instruction, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__buffer.") {
				value, err := executeBufferIntrinsic(instruction, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__text_encoder.") || strings.HasPrefix(instruction.Callee, "__text_decoder.") {
				value, err := executeTextEncodingIntrinsic(instruction, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__async.") {
				value, err := executeAsyncIntrinsic(instruction.Callee, instruction.Args, env, functions, output)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__generator.") {
				value, err := executeGeneratorIntrinsic(instruction, env, functions, output)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__iterator.") {
				value, err := executeIteratorIntrinsic(instruction, env, functions, output)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__timers.") {
				value, err := executeTimerIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__fs.") {
				value, err := executeFsIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__child_process.") {
				value, err := executeChildProcessIntrinsic(instruction, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__http.") {
				value, err := executeHttpIntrinsic(instruction, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__process.") {
				if instruction.Callee == "__process.exit" {
					return Value{Type: ir.TypeVoid}, true, flowExit, nil
				}
				value, err := executeProcessIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__crypto.") {
				value, err := executeCryptoIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__date.") {
				value, err := executeDateIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__os.") {
				value, err := executeOsIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__web.") {
				value, err := executeWebIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__performance.") {
				value, err := executePerformanceIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__symbol.") {
				value, err := executeSymbolIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__json.") {
				value, err := executeJsonIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if instruction.Result != "" {
					env[instruction.Result] = value
				}
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__string.") {
				value, err := executeStringIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				env[instruction.Result] = value
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__regex.") || strings.HasPrefix(instruction.Callee, "__regexp.") {
				value, err := executeRegexIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				env[instruction.Result] = value
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__bigint.") {
				value, err := executeBigIntIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				env[instruction.Result] = value
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__weak") || strings.HasPrefix(instruction.Callee, "__gc.") {
				value, err := executeWeakIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				env[instruction.Result] = value
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__intl.") {
				args := make([]Value, 0, len(instruction.Args))
				for _, name := range instruction.Args {
					v, ok := env[name]
					if ok {
						args = append(args, v)
					}
				}
				value, err := execIntlIntrinsic(instruction.Callee, args)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				env[instruction.Result] = value
				continue
			}
			callee, ok := functions[instruction.Callee]
			if !ok {
				return Value{}, false, flowNormal, fmt.Errorf("native FFI call %q requires native compilation (--native)", instruction.Callee)
			}
			arguments := make([]Value, 0, len(instruction.Args))
			for _, name := range instruction.Args {
				value, ok := env[name]
				if !ok {
					return Value{}, false, flowNormal, fmt.Errorf("unknown call argument %q", name)
				}
				arguments = append(arguments, value)
			}
			value, flow, err := executeFunction(functions, callee, arguments, output)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if flow.kind == kindThrow {
				return value, false, flow, nil
			}
			env[instruction.Result] = value
		case ir.OpThrow:
			val, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			return val, false, controlFlow{kind: kindThrow}, nil
		case ir.OpTry:
			ret, returned, flow, err := executeBlock(functions, instruction.Body, env, output)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if flow.kind == kindThrow {
				if len(instruction.Catch) > 0 {
					if instruction.CatchVar != "" {
						env[instruction.CatchVar] = ret
					}
					ret, returned, flow, err = executeBlock(functions, instruction.Catch, env, output)
					if err != nil {
						return Value{}, false, flowNormal, err
					}
				}
			}
			if len(instruction.Finally) > 0 {
				fRet, fReturned, fFlow, fErr := executeBlock(functions, instruction.Finally, env, output)
				if fErr != nil {
					return Value{}, false, flowNormal, fErr
				}
				if fReturned || fFlow.kind != kindNormal {
					ret = fRet
					returned = fReturned
					flow = fFlow
				}
			}
			if returned || flow.kind != kindNormal {
				return ret, returned, flow, nil
			}
		case ir.OpDebugger:
			// No-op in headless reference interpreter
		case ir.OpBreak:
			return Value{}, false, controlFlow{kind: kindBreak, target: instruction.Value}, nil
		case ir.OpContinue:
			return Value{}, false, controlFlow{kind: kindContinue, target: instruction.Value}, nil
		case ir.OpReturn:
			if len(instruction.Args) == 0 {
				return Value{Type: ir.TypeVoid}, true, controlFlow{kind: kindReturn}, nil
			}
			retVal, err := lookup(env, instruction.Args, 0)
			return retVal, true, controlFlow{kind: kindReturn}, err
		case ir.OpIf:
			condition, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			branch := instruction.Else
			if condition.Bool {
				branch = instruction.Then
			}
			if len(branch) > 0 {
				ret, returned, flow, err := executeBlock(functions, branch, env, output)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if returned || flow.kind != kindNormal {
					return ret, returned, flow, nil
				}
			}
		case ir.OpWhile:
			for {
				if len(instruction.Cond) > 0 {
					ret, returned, flow, err := executeBlock(functions, instruction.Cond, env, output)
					if err != nil {
						return Value{}, false, flowNormal, err
					}
					if returned || flow.kind != kindNormal {
						return ret, returned, flow, nil
					}
				}
				condition, err := lookup(env, instruction.Args, 0)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if !condition.Bool {
					break
				}
				ret, returned, flow, err := executeBlock(functions, instruction.Body, env, output)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if returned || flow.kind == kindReturn || flow.kind == kindThrow || flow.kind == kindExit {
					return ret, returned, flow, nil
				}
				if flow.kind == kindBreak {
					if flow.target == "" || flow.target == instruction.Value {
						break
					}
					return ret, returned, flow, nil
				}
				if flow.kind == kindContinue {
					if flow.target != "" && flow.target != instruction.Value {
						return ret, returned, flow, nil
					}
				}
				if len(instruction.Step) > 0 {
					ret, returned, flow, err := executeBlock(functions, instruction.Step, env, output)
					if err != nil {
						return Value{}, false, flowNormal, err
					}
					if returned || flow.kind == kindReturn || flow.kind == kindThrow || flow.kind == kindExit {
						return ret, returned, flow, nil
					}
				}
				if flow.kind == kindContinue {
					if flow.target == "" || flow.target == instruction.Value {
						continue
					}
					return ret, returned, flow, nil
				}
			}
		case ir.OpDoWhile:
			for {
				ret, returned, flow, err := executeBlock(functions, instruction.Body, env, output)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if returned || flow.kind == kindReturn || flow.kind == kindThrow || flow.kind == kindExit {
					return ret, returned, flow, nil
				}
				if flow.kind == kindBreak {
					if flow.target == "" || flow.target == instruction.Value {
						break
					}
					return ret, returned, flow, nil
				}
				if flow.kind == kindContinue {
					if flow.target != "" && flow.target != instruction.Value {
						return ret, returned, flow, nil
					}
				}
				if len(instruction.Step) > 0 {
					ret, returned, flow, err := executeBlock(functions, instruction.Step, env, output)
					if err != nil {
						return Value{}, false, flowNormal, err
					}
					if returned || flow.kind == kindReturn || flow.kind == kindThrow || flow.kind == kindExit {
						return ret, returned, flow, nil
					}
				}
				if flow.kind == kindContinue {
					if flow.target == "" || flow.target == instruction.Value {
						continue
					}
					return ret, returned, flow, nil
				}
				if len(instruction.Cond) > 0 {
					ret, returned, flow, err := executeBlock(functions, instruction.Cond, env, output)
					if err != nil {
						return Value{}, false, flowNormal, err
					}
					if returned || flow.kind != kindNormal {
						return ret, returned, flow, nil
					}
				}
				condition, err := lookup(env, instruction.Args, 0)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if !condition.Bool {
					break
				}
			}
		default:
			return Value{}, false, flowNormal, fmt.Errorf("unsupported interpreter instruction %q", instruction.Op)
		}
	}
	return Value{Type: ir.TypeVoid}, false, flowNormal, nil
}

func lookup(env map[string]Value, arguments []string, index int) (Value, error) {
	if index >= len(arguments) {
		return Value{}, fmt.Errorf("missing value at argument %d", index)
	}
	value, ok := env[arguments[index]]
	if !ok {
		return Value{}, fmt.Errorf("unknown value %q", arguments[index])
	}
	return value, nil
}

func castValue(val Value, targetType ir.Type) (Value, error) {
	if val.Type == ir.TypeUnknown {
		if val.Boxed == nil {
			return Value{}, fmt.Errorf("TypeError: SG4002: cannot cast uninitialized unknown to %s", targetType)
		}
		val = *val.Boxed
	}
	if val.Type == targetType {
		return val, nil
	}
	if (strings.Contains(string(val.Type), "[]") || strings.Contains(string(val.Type), "__shape_0_") || val.ArrayRef != nil || len(val.Array) > 0) && (strings.Contains(string(targetType), "[]") || targetType == ir.TypeObject) {
		res := val
		res.Type = targetType
		return res, nil
	}
	if (strings.HasPrefix(string(val.Type), "object:") || val.Object != nil) && (strings.HasPrefix(string(targetType), "object:") || targetType == ir.TypeObject) {
		res := val
		res.Type = targetType
		return res, nil
	}
	if val.Type == ir.TypeString && strings.HasPrefix(string(targetType), "object:") {
		var objMap map[string]interface{}
		if err := json.Unmarshal([]byte(val.String), &objMap); err == nil {
			objVal := Value{
				Type:   targetType,
				Object: make(map[string]Value),
			}
			for k, v := range objMap {
				switch vv := v.(type) {
				case float64:
					objVal.Object[k] = Value{Type: ir.TypeNumber, Number: vv}
				case string:
					objVal.Object[k] = Value{Type: ir.TypeString, String: vv}
				case bool:
					objVal.Object[k] = Value{Type: ir.TypeBool, Bool: vv}
				}
			}
			return objVal, nil
		}
	}
	if val.Type == ir.TypeString && targetType == ir.TypeNumber {
		if num, err := strconv.ParseFloat(val.String, 64); err == nil {
			return Value{Type: ir.TypeNumber, Number: num}, nil
		}
	}
	if val.Type == ir.TypeString && targetType == ir.TypeBool {
		return Value{Type: ir.TypeBool, Bool: val.String == "true"}, nil
	}
	if val.Type == ir.TypeBuffer && targetType == ir.TypeUint8Array {
		return val, nil
	}
	return Value{}, fmt.Errorf("TypeError: SG4002: cannot cast %s to %s", val.Type, targetType)
}
