// Package interpreter provides the reference execution engine for typed IR.
// It is a test oracle and is not linked into native executables.
package interpreter

import (
	"bytes"
	"fmt"
	"math"
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
	if flow == flowThrow {
		return Result{}, fmt.Errorf("uncaught exception: %s", format(value))
	}
	return Result{Output: output.String(), Return: value}, nil
}

type controlFlow int

const (
	flowNormal controlFlow = iota
	flowReturn
	flowBreak
	flowContinue
	flowThrow
	flowExit
)

func executeFunction(functions map[string]ir.Function, function ir.Function, arguments []Value, output *bytes.Buffer) (Value, controlFlow, error) {
	env := make(map[string]Value, len(function.Parameters))
	if len(arguments) != len(function.Parameters) {
		return Value{}, flowNormal, fmt.Errorf("function %q received %d arguments, want %d", function.Name, len(arguments), len(function.Parameters))
	}
	for index, parameter := range function.Parameters {
		if arguments[index].Type != parameter.Type {
			return Value{}, flowNormal, fmt.Errorf("argument %d to %q has type %s, want %s", index, function.Name, arguments[index].Type, parameter.Type)
		}
		env[parameter.Name] = arguments[index]
	}

	val, _, flow, err := executeBlock(functions, function.Body, env, output)
	return val, flow, err
}

func executeBlock(functions map[string]ir.Function, body []ir.Instruction, env map[string]Value, output *bytes.Buffer) (Value, bool, controlFlow, error) {
	for _, instruction := range body {
		switch instruction.Op {
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
			array := make([]Value, 0, len(instruction.Args))
			for _, name := range instruction.Args {
				value, err := lookup(env, []string{name}, 0)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				array = append(array, value)
			}
			env[instruction.Result] = Value{Type: instruction.Type, Array: array}
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
			if position >= len(array.Array) {
				return Value{}, false, flowNormal, fmt.Errorf("array index %d out of bounds for length %d", position, len(array.Array))
			}
			env[instruction.Result] = array.Array[position]
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
			if position >= len(array.Array) {
				return Value{}, false, flowNormal, fmt.Errorf("array index %d out of bounds for length %d", position, len(array.Array))
			}
			array.Array[position] = val
		case ir.OpObjectNew:
			env[instruction.Result] = Value{Type: instruction.Type, Object: map[string]Value{}}
		case ir.OpFieldSet:
			object, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			value, err := lookup(env, instruction.Args, 1)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if object.Object == nil {
				return Value{}, false, flowNormal, fmt.Errorf("field set on non-object value")
			}
			object.Object[instruction.Field] = value
		case ir.OpFieldGet:
			object, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			value, ok := object.Object[instruction.Field]
			if !ok {
				return Value{}, false, flowNormal, fmt.Errorf("unknown field %q", instruction.Field)
			}
			env[instruction.Result] = value
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
				value, err := executeArrayIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				env[instruction.Result] = value
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
			if strings.HasPrefix(instruction.Callee, "__string.") {
				value, err := executeStringIntrinsic(instruction.Callee, instruction.Args, env)
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
			if flow == flowThrow {
				return value, false, flowThrow, nil
			}
			env[instruction.Result] = value
		case ir.OpThrow:
			val, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			return val, false, flowThrow, nil
		case ir.OpTry:
			ret, returned, flow, err := executeBlock(functions, instruction.Body, env, output)
			if err != nil {
				return Value{}, false, flowNormal, err
			}
			if flow == flowThrow {
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
				if fReturned || fFlow != flowNormal {
					ret = fRet
					returned = fReturned
					flow = fFlow
				}
			}
			if returned || flow != flowNormal {
				return ret, returned, flow, nil
			}
		case ir.OpBreak:
			return Value{}, false, flowBreak, nil
		case ir.OpContinue:
			return Value{}, false, flowContinue, nil
		case ir.OpReturn:
			if len(instruction.Args) == 0 {
				return Value{Type: ir.TypeVoid}, true, flowReturn, nil
			}
			retVal, err := lookup(env, instruction.Args, 0)
			return retVal, true, flowReturn, err
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
				if returned || flow != flowNormal {
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
					if returned || flow != flowNormal {
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
				if returned || flow == flowReturn || flow == flowThrow {
					return ret, returned, flow, nil
				}
				if flow == flowBreak {
					break
				}
				if flow == flowContinue {
					continue
				}
			}
		case ir.OpDoWhile:
			for {
				ret, returned, flow, err := executeBlock(functions, instruction.Body, env, output)
				if err != nil {
					return Value{}, false, flowNormal, err
				}
				if returned || flow == flowReturn || flow == flowThrow {
					return ret, returned, flow, nil
				}
				if flow == flowBreak {
					break
				}
				if len(instruction.Cond) > 0 {
					ret, returned, flow, err := executeBlock(functions, instruction.Cond, env, output)
					if err != nil {
						return Value{}, false, flowNormal, err
					}
					if returned || flow != flowNormal {
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
