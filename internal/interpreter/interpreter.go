// Package interpreter provides the reference execution engine for typed IR.
// It is a test oracle and is not linked into native executables.
package interpreter

import (
	"bytes"
	"fmt"
	"math"
	"strconv"

	"github.com/pilotworks/scriptgo/internal/ir"
)

type Value struct {
	Type   ir.Type
	Number float64
	String string
	Bool   bool
	Array  []Value
	Object map[string]Value
}

type Result struct {
	Output string
	Return Value
}

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
	value, err := executeFunction(functions, main, nil, &output)
	if err != nil {
		return Result{}, err
	}
	return Result{Output: output.String(), Return: value}, nil
}

func executeFunction(functions map[string]ir.Function, function ir.Function, arguments []Value, output *bytes.Buffer) (Value, error) {
	env := make(map[string]Value, len(function.Parameters))
	if len(arguments) != len(function.Parameters) {
		return Value{}, fmt.Errorf("function %q received %d arguments, want %d", function.Name, len(arguments), len(function.Parameters))
	}
	for index, parameter := range function.Parameters {
		if arguments[index].Type != parameter.Type {
			return Value{}, fmt.Errorf("argument %d to %q has type %s, want %s", index, function.Name, arguments[index].Type, parameter.Type)
		}
		env[parameter.Name] = arguments[index]
	}

	for _, instruction := range function.Body {
		switch instruction.Op {
		case ir.OpConst:
			value, err := parseConstant(instruction.Type, instruction.Value)
			if err != nil {
				return Value{}, fmt.Errorf("%s: %w", instruction.Result, err)
			}
			env[instruction.Result] = value
		case ir.OpBinary:
			left, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, err
			}
			right, err := lookup(env, instruction.Args, 1)
			if err != nil {
				return Value{}, err
			}
			value, err := binary(instruction.Operator, left, right)
			if err != nil {
				return Value{}, err
			}
			env[instruction.Result] = value
		case ir.OpArray:
			array := make([]Value, 0, len(instruction.Args))
			for _, name := range instruction.Args {
				value, err := lookup(env, []string{name}, 0)
				if err != nil {
					return Value{}, err
				}
				array = append(array, value)
			}
			env[instruction.Result] = Value{Type: ir.TypeNumberArray, Array: array}
		case ir.OpIndex:
			array, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, err
			}
			index, err := lookup(env, instruction.Args, 1)
			if err != nil {
				return Value{}, err
			}
			if index.Type != ir.TypeNumber || math.Trunc(index.Number) != index.Number || index.Number < 0 {
				return Value{}, fmt.Errorf("array index must be a non-negative integer, got %v", index.Number)
			}
			position := int(index.Number)
			if position >= len(array.Array) {
				return Value{}, fmt.Errorf("array index %d out of bounds for length %d", position, len(array.Array))
			}
			env[instruction.Result] = array.Array[position]
		case ir.OpObjectNew:
			env[instruction.Result] = Value{Type: instruction.Type, Object: map[string]Value{}}
		case ir.OpFieldSet:
			object, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, err
			}
			value, err := lookup(env, instruction.Args, 1)
			if err != nil {
				return Value{}, err
			}
			if object.Object == nil {
				return Value{}, fmt.Errorf("field set on non-object value")
			}
			object.Object[instruction.Field] = value
		case ir.OpFieldGet:
			object, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, err
			}
			value, ok := object.Object[instruction.Field]
			if !ok {
				return Value{}, fmt.Errorf("unknown field %q", instruction.Field)
			}
			env[instruction.Result] = value
		case ir.OpPrint:
			value, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, err
			}
			if value.Type == ir.TypeNumberArray {
				return Value{}, fmt.Errorf("console.log does not support array values yet")
			}
			fmt.Fprintln(output, format(value))
		case ir.OpCall:
			callee, ok := functions[instruction.Callee]
			if !ok {
				return Value{}, fmt.Errorf("unknown function %q", instruction.Callee)
			}
			arguments := make([]Value, 0, len(instruction.Args))
			for _, name := range instruction.Args {
				value, ok := env[name]
				if !ok {
					return Value{}, fmt.Errorf("unknown call argument %q", name)
				}
				arguments = append(arguments, value)
			}
			value, err := executeFunction(functions, callee, arguments, output)
			if err != nil {
				return Value{}, err
			}
			env[instruction.Result] = value
		case ir.OpReturn:
			if len(instruction.Args) == 0 {
				return Value{Type: ir.TypeVoid}, nil
			}
			return lookup(env, instruction.Args, 0)
		default:
			return Value{}, fmt.Errorf("unsupported interpreter instruction %q", instruction.Op)
		}
	}
	return Value{Type: ir.TypeVoid}, nil
}

func parseConstant(typ ir.Type, value string) (Value, error) {
	switch typ {
	case ir.TypeNumber:
		number, err := strconv.ParseFloat(value, 64)
		return Value{Type: typ, Number: number}, err
	case ir.TypeString:
		return Value{Type: typ, String: value}, nil
	case ir.TypeBool:
		boolean, err := strconv.ParseBool(value)
		return Value{Type: typ, Bool: boolean}, err
	default:
		return Value{}, fmt.Errorf("unsupported constant type %s", typ)
	}
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

func binary(operator string, left, right Value) (Value, error) {
	if left.Type != right.Type {
		return Value{}, fmt.Errorf("binary operands have types %s and %s", left.Type, right.Type)
	}
	if left.Type == ir.TypeString && operator == "+" {
		return Value{Type: ir.TypeString, String: left.String + right.String}, nil
	}
	if left.Type != ir.TypeNumber {
		return Value{}, fmt.Errorf("operator %q is unsupported for %s", operator, left.Type)
	}
	switch operator {
	case "+":
		return Value{Type: ir.TypeNumber, Number: left.Number + right.Number}, nil
	case "-":
		return Value{Type: ir.TypeNumber, Number: left.Number - right.Number}, nil
	case "*":
		return Value{Type: ir.TypeNumber, Number: left.Number * right.Number}, nil
	case "/":
		return Value{Type: ir.TypeNumber, Number: left.Number / right.Number}, nil
	case "%":
		return Value{Type: ir.TypeNumber, Number: float64(int64(left.Number) % int64(right.Number))}, nil
	default:
		return Value{}, fmt.Errorf("operator %q is unsupported for number", operator)
	}
}

func format(value Value) string {
	switch value.Type {
	case ir.TypeNumber:
		return strconv.FormatFloat(value.Number, 'f', -1, 64)
	case ir.TypeString:
		return value.String
	case ir.TypeBool:
		return strconv.FormatBool(value.Bool)
	default:
		return ""
	}
}
