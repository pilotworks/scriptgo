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
		case ir.OpCompare:
			left, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, err
			}
			right, err := lookup(env, instruction.Args, 1)
			if err != nil {
				return Value{}, err
			}
			value, err := compare(instruction.Operator, left, right)
			if err != nil {
				return Value{}, err
			}
			env[instruction.Result] = value
		case ir.OpSelect:
			condition, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, err
			}
			selectedName := instruction.Args[2]
			if condition.Bool {
				selectedName = instruction.Args[1]
			}
			selected, err := lookup(env, []string{selectedName}, 0)
			if err != nil {
				return Value{}, err
			}
			env[instruction.Result] = selected
		case ir.OpArray:
			array := make([]Value, 0, len(instruction.Args))
			for _, name := range instruction.Args {
				value, err := lookup(env, []string{name}, 0)
				if err != nil {
					return Value{}, err
				}
				array = append(array, value)
			}
			env[instruction.Result] = Value{Type: instruction.Type, Array: array}
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
			if value.Type == ir.TypeNumberArray || value.Type == ir.TypeStringArray {
				return Value{}, fmt.Errorf("console.log does not support array values yet")
			}
			fmt.Fprintln(output, format(value))
		case ir.OpCall:
			if strings.HasPrefix(instruction.Callee, "__Math.") {
				value, err := executeMathIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, err
				}
				env[instruction.Result] = value
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__array.") {
				value, err := executeArrayIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, err
				}
				env[instruction.Result] = value
				continue
			}
			if strings.HasPrefix(instruction.Callee, "__string.") {
				value, err := executeStringIntrinsic(instruction.Callee, instruction.Args, env)
				if err != nil {
					return Value{}, err
				}
				env[instruction.Result] = value
				continue
			}
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
		case ir.OpIf:
			condition, err := lookup(env, instruction.Args, 0)
			if err != nil {
				return Value{}, err
			}
			branch := instruction.Else
			if condition.Bool {
				branch = instruction.Then
			}
			if len(branch) == 1 && branch[0].Op == ir.OpReturn {
				if len(branch[0].Args) == 0 {
					return Value{Type: ir.TypeVoid}, nil
				}
				return lookup(env, branch[0].Args, 0)
			}
			if len(branch) != 0 {
				return Value{}, fmt.Errorf("interpreter only supports if branches that return")
			}
		default:
			return Value{}, fmt.Errorf("unsupported interpreter instruction %q", instruction.Op)
		}
	}
	return Value{Type: ir.TypeVoid}, nil
}

func parseConstant(typ ir.Type, value string) (Value, error) {
	switch typ {
	case ir.TypeNumber:
		if value == "NaN" {
			return Value{Type: typ, Number: math.NaN()}, nil
		}
		if value == "+Inf" {
			return Value{Type: typ, Number: math.Inf(1)}, nil
		}
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

func executeMathIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	if len(arguments) != 1 {
		return Value{}, fmt.Errorf("%s requires one number", name)
	}
	argument, ok := env[arguments[0]]
	if !ok || argument.Type != ir.TypeNumber {
		return Value{}, fmt.Errorf("%s requires one number", name)
	}
	value := argument.Number
	switch name {
	case "__Math.abs":
		value = math.Abs(value)
	case "__Math.ceil":
		value = math.Ceil(value)
	case "__Math.floor":
		value = math.Floor(value)
	case "__Math.trunc":
		value = math.Trunc(value)
	default:
		return Value{}, fmt.Errorf("unknown math intrinsic %q", name)
	}
	return Value{Type: ir.TypeNumber, Number: value}, nil
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

func compare(operator string, left, right Value) (Value, error) {
	if left.Type != right.Type {
		return Value{}, fmt.Errorf("compare operands have types %s and %s", left.Type, right.Type)
	}
	var result bool
	switch left.Type {
	case ir.TypeNumber:
		switch operator {
		case "==":
			result = left.Number == right.Number
		case "!==":
			result = left.Number != right.Number
		case "<":
			result = left.Number < right.Number
		case "<=":
			result = left.Number <= right.Number
		case ">":
			result = left.Number > right.Number
		case ">=":
			result = left.Number >= right.Number
		}
	case ir.TypeString:
		switch operator {
		case "==":
			result = left.String == right.String
		case "!==":
			result = left.String != right.String
		case "<":
			result = left.String < right.String
		case "<=":
			result = left.String <= right.String
		case ">":
			result = left.String > right.String
		case ">=":
			result = left.String >= right.String
		}
	default:
		return Value{}, fmt.Errorf("compare is unsupported for %s", left.Type)
	}
	return Value{Type: ir.TypeBool, Bool: result}, nil
}

func executeStringIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	values := make([]Value, 0, len(arguments))
	for _, argument := range arguments {
		value, ok := env[argument]
		if !ok {
			return Value{}, fmt.Errorf("unknown string intrinsic argument %q", argument)
		}
		values = append(values, value)
	}
	switch name {
	case "__string.length":
		if len(values) != 1 || values[0].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.length requires one string")
		}
		return Value{Type: ir.TypeNumber, Number: float64(len(values[0].String))}, nil
	case "__string.lastIndexOf":
		if (len(values) != 2 && len(values) != 3) || values[0].Type != ir.TypeString || values[1].Type != ir.TypeString {
			return Value{}, fmt.Errorf("string.lastIndexOf requires two strings and an optional position")
		}
		value := values[0].String
		if len(values) == 3 {
			if values[2].Number >= 0 {
				position := minInt(int(values[2].Number)+1, len(value))
				value = value[:position]
			}
		}
		return Value{Type: ir.TypeNumber, Number: float64(strings.LastIndex(value, values[1].String))}, nil
	case "__string.slice":
		if len(values) != 3 || values[0].Type != ir.TypeString || values[1].Type != ir.TypeNumber || values[2].Type != ir.TypeNumber {
			return Value{}, fmt.Errorf("string.slice requires string and two numbers")
		}
		start := maxInt(0, minInt(int(values[1].Number), len(values[0].String)))
		end := maxInt(start, minInt(int(values[2].Number), len(values[0].String)))
		return Value{Type: ir.TypeString, String: values[0].String[start:end]}, nil
	default:
		return Value{}, fmt.Errorf("unknown string intrinsic %q", name)
	}
}

func executeArrayIntrinsic(name string, arguments []string, env map[string]Value) (Value, error) {
	if len(arguments) != 1 {
		return Value{}, fmt.Errorf("array intrinsic requires one array")
	}
	array, ok := env[arguments[0]]
	if !ok || (array.Type != ir.TypeNumberArray && array.Type != ir.TypeStringArray) {
		return Value{}, fmt.Errorf("array intrinsic requires an array")
	}
	switch name {
	case "__array.length":
		return Value{Type: ir.TypeNumber, Number: float64(len(array.Array))}, nil
	default:
		return Value{}, fmt.Errorf("unknown array intrinsic %q", name)
	}
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
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
