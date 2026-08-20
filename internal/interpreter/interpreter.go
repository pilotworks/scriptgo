// Package interpreter provides the reference execution engine for typed IR.
// It is a test oracle and is not linked into native executables.
package interpreter

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

// Execute interprets a verified module using the reference IR semantics.
func Execute(module ir.Module) (Result, error) {
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
	var output bytes.Buffer
	value, flow, err := executeFunction(functions, main, nil, &output)
	if err != nil {
		return Result{}, err
	}
	if flow.kind == kindThrow {
		return Result{}, fmt.Errorf("uncaught exception: %s", format(value))
	}
	if err := drainMicrotasks(functions, &output); err != nil {
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
	flowNormal   = controlFlow{kind: kindNormal}
	flowReturn   = controlFlow{kind: kindReturn}
	flowBreak    = controlFlow{kind: kindBreak}
	flowContinue = controlFlow{kind: kindContinue}
	flowThrow    = controlFlow{kind: kindThrow}
	flowExit     = controlFlow{kind: kindExit}
)

func executeFunction(functions map[string]ir.Function, function ir.Function, arguments []Value, output *bytes.Buffer) (Value, controlFlow, error) {
	env := make(map[string]Value, len(function.Parameters))
	if len(arguments) != len(function.Parameters) {
		return Value{}, flowNormal, fmt.Errorf("function %q received %d arguments, want %d", function.Name, len(arguments), len(function.Parameters))
	}
	for index, parameter := range function.Parameters {
		argType := arguments[index].Type
		paramType := parameter.Type
		if argType != paramType {
			if !(strings.HasPrefix(string(argType), "object:") && strings.HasPrefix(string(paramType), "object:")) {
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
	env := make(map[string]Value)
	for k, v := range closure.Env {
		env[k] = v
	}
	userParams := closure.Function.Parameters
	if len(userParams) > 0 && userParams[0].Name == "__env_ctx" {
		userParams = userParams[1:]
	}
	for index, parameter := range userParams {
		if index < len(arguments) {
			env[parameter.Name] = arguments[index]
		}
	}
	val, _, flow, err := executeBlock(functions, closure.Function.Body, env, output)
	return val, flow, err
}

func executeBlock(functions map[string]ir.Function, body []ir.Instruction, env map[string]Value, output *bytes.Buffer) (Value, bool, controlFlow, error) {
	for _, instruction := range body {
		switch instruction.Op {
		case ir.OpClosure:
			targetFn, ok := functions[instruction.Callee]
			if !ok {
				return Value{}, false, flowNormal, fmt.Errorf("unknown function for closure %q", instruction.Callee)
			}
			captured := make(map[string]Value, len(instruction.Args))
			for _, argName := range instruction.Args {
				val, err := lookup(env, []string{argName}, 0)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				captured[argName] = val
			}
			env[instruction.Result] = Value{
				Type: ir.TypeClosure,
				Closure: &Closure{
					Function: targetFn,
					Env:      captured,
				},
			}
		case ir.OpClosureCall:
			closureVal, err := lookup(env, []string{instruction.Callee}, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if closureVal.Closure == nil {
				return Value{}, false, flowNormal, fmt.Errorf("%q is not a callable closure", instruction.Callee)
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
			index, err := lookup(env, instruction.Args, 1)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if index.Type != ir.TypeNumber || math.Trunc(index.Number) != index.Number || index.Number < 0 {
				return Value{}, false, flowNormal, fmt.Errorf("array index must be a non-negative integer, got %v", index.Number)
			}
			position := int(index.Number)
			arr := array.GetArray()
			if position >= len(arr) {
				return Value{}, false, flowNormal, fmt.Errorf("array index %d out of bounds for length %d", position, len(arr))
			}
			env[instruction.Result] = arr[position]
		case ir.OpIndexSet:
			array, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
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
					continue
				}
			}
			if object.Object == nil {
				object.Object = make(map[string]Value)
			}
			object.Object[instruction.Field] = value
		case ir.OpFieldGet:
			object, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
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
				return Value{}, false, flowNormal, fmt.Errorf("unknown field %q", instruction.Field)
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
			if val.Type == ir.TypeUnknown {
				if val.Boxed == nil {
					return Value{}, false, flowNormal, fmt.Errorf("TypeError: SG4002: cannot cast uninitialized unknown to %s", instruction.Type)
				}
				inner := *val.Boxed
				if inner.Type != instruction.Type && !(strings.HasPrefix(string(inner.Type), "object:") && strings.HasPrefix(string(instruction.Type), "object:")) {
					return Value{}, false, flowNormal, fmt.Errorf("TypeError: SG4002: cannot cast %s to %s", inner.Type, instruction.Type)
				}
				env[instruction.Result] = inner
			} else {
				env[instruction.Result] = val
			}
		case ir.OpTypeOf:
			val, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			typeStr := "object"
			actualType := val.Type
			if val.Type == ir.TypeUnknown && val.Boxed != nil {
				actualType = val.Boxed.Type
			}
			switch actualType {
			case ir.TypeNumber:
				typeStr = "number"
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
			env[instruction.Result] = Value{Type: ir.TypeString, String: typeStr}
		case ir.OpPrint:
			value, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if value.Type == ir.TypeNumberArray || value.Type == ir.TypeStringArray {
				return Value{}, false, flowNormal, fmt.Errorf("console.log does not support array values yet")
			}
			fmt.Fprintln(output, format(value))
		case ir.OpCall:
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
			if strings.HasPrefix(instruction.Callee, "__regex.") {
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
			callee, ok := functions[instruction.Callee]
			if !ok {
				return Value{}, false, flowNormal, fmt.Errorf("unknown function %q", instruction.Callee)
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
