package ir

import (
	"fmt"
	"strings"
)

// Verify checks the invariants required by every native backend.
func (m Module) Verify() error {
	if len(m.Functions) == 0 {
		return fmt.Errorf("module has no functions")
	}
	for _, function := range m.Functions {
		if err := function.Verify(); err != nil {
			return fmt.Errorf("function %q: %w", function.Name, err)
		}
	}
	for _, shape := range m.Shapes {
		if shape.Name == "" || len(shape.Fields) == 0 {
			return fmt.Errorf("invalid object shape %q", shape.Name)
		}
		for _, field := range shape.Fields {
			if field.Name == "" || field.Type == "" {
				return fmt.Errorf("invalid field in object shape %q", shape.Name)
			}
		}
	}
	return nil
}

func (f Function) Verify() error {
	if f.Name == "" {
		return fmt.Errorf("function has no name")
	}
	known := map[string]Type{}
	for _, parameter := range f.Parameters {
		if parameter.Name == "" || parameter.Type == "" {
			return fmt.Errorf("invalid parameter")
		}
		known[parameter.Name] = parameter.Type
	}
	terminated := false
	for _, instruction := range f.Body {
		if terminated {
			return fmt.Errorf("instruction follows return")
		}
		for _, arg := range instruction.Args {
			if _, ok := known[arg]; !ok && instruction.Op != OpCall && instruction.Op != OpPrint {
				return fmt.Errorf("unknown value %q", arg)
			}
		}
		switch instruction.Op {
		case OpConst, OpBinary, OpCompare, OpSelect, OpCall, OpParam, OpArray, OpIndex, OpObjectNew, OpFieldGet:
			if instruction.Result == "" || instruction.Type == "" {
				return fmt.Errorf("%s instruction must define result and type", instruction.Op)
			}
			if instruction.Op == OpArray {
				if instruction.Type != TypeNumberArray && instruction.Type != TypeStringArray {
					return fmt.Errorf("array instruction has unsupported type %q", instruction.Type)
				}
				for _, argument := range instruction.Args {
					if known[argument] != elementType(instruction.Type) {
						return fmt.Errorf("array element %q has type %s, want %s", argument, known[argument], elementType(instruction.Type))
					}
				}
			}
			if instruction.Op == OpBinary {
				if len(instruction.Args) != 2 {
					return fmt.Errorf("binary instruction requires two operands")
				}
				leftType := known[instruction.Args[0]]
				rightType := known[instruction.Args[1]]
				if leftType != rightType {
					return fmt.Errorf("binary operands must have the same type")
				}
				if leftType == TypeBool {
					if instruction.Operator != "&&" && instruction.Operator != "||" {
						return fmt.Errorf("binary bool instruction has unsupported operator %q", instruction.Operator)
					}
					if instruction.Type != TypeBool {
						return fmt.Errorf("binary bool instruction must have bool result type")
					}
				}
			}
			if instruction.Op == OpCompare {
				if len(instruction.Args) != 2 || instruction.Type != TypeBool {
					return fmt.Errorf("compare instruction requires two operands and bool result")
				}
				if instruction.Operator != "==" && instruction.Operator != "===" && instruction.Operator != "!=" && instruction.Operator != "!==" && instruction.Operator != "<" && instruction.Operator != "<=" && instruction.Operator != ">" && instruction.Operator != ">=" {
					return fmt.Errorf("compare instruction has unsupported operator %q", instruction.Operator)
				}
				if known[instruction.Args[0]] != known[instruction.Args[1]] {
					return fmt.Errorf("compare operands must have the same type")
				}
			}
			if instruction.Op == OpSelect && (len(instruction.Args) != 3 || known[instruction.Args[0]] != TypeBool || known[instruction.Args[1]] != instruction.Type || known[instruction.Args[2]] != instruction.Type) {
				return fmt.Errorf("select requires bool condition and matching values")
			}
			if instruction.Op == OpIndex {
				if len(instruction.Args) != 2 {
					return fmt.Errorf("index instruction requires array and index operands")
				}
				if (known[instruction.Args[0]] != TypeNumberArray && known[instruction.Args[0]] != TypeStringArray) || known[instruction.Args[1]] != TypeNumber {
					return fmt.Errorf("index instruction requires an array and number operands")
				}
				if instruction.Type != elementType(known[instruction.Args[0]]) {
					return fmt.Errorf("index instruction has incompatible result type %s", instruction.Type)
				}
			}
			if instruction.Op == OpObjectNew && !strings.HasPrefix(string(instruction.Type), string(TypeObject)+":") {
				return fmt.Errorf("object.new must produce a shaped object")
			}
			if instruction.Op == OpObjectNew && instruction.FieldCount < 0 {
				return fmt.Errorf("object.new has invalid field count")
			}
			if instruction.Op == OpFieldGet && len(instruction.Args) != 1 {
				return fmt.Errorf("field.get requires one object operand")
			}
			if instruction.Op == OpFieldGet && instruction.FieldIndex < 0 {
				return fmt.Errorf("field.get has invalid field index")
			}
			known[instruction.Result] = instruction.Type
		case OpPrint:
			if len(instruction.Args) != 1 {
				return fmt.Errorf("print instruction requires one argument")
			}
		case OpFieldSet:
			if instruction.Type != TypeVoid || len(instruction.Args) != 2 || instruction.Field == "" || instruction.FieldIndex < 0 {
				return fmt.Errorf("field.set requires object, value, and field")
			}
		case OpIf:
			if instruction.Type != TypeVoid || instruction.Result != "" || len(instruction.Args) != 1 || known[instruction.Args[0]] != TypeBool {
				return fmt.Errorf("if requires one bool condition")
			}
			thenKnown := cloneTypes(known)
			elseKnown := cloneTypes(known)
			if err := verifyBlock(f, instruction.Then, thenKnown); err != nil {
				return fmt.Errorf("if then block: %w", err)
			}
			if err := verifyBlock(f, instruction.Else, elseKnown); err != nil {
				return fmt.Errorf("if else block: %w", err)
			}
		case OpReturn:
			if instruction.Type != f.ReturnType {
				return fmt.Errorf("return type %q does not match function return type %q", instruction.Type, f.ReturnType)
			}
			terminated = true
		default:
			return fmt.Errorf("unknown instruction %q", instruction.Op)
		}
	}
	return nil
}

func elementType(arrayType Type) Type {
	if arrayType == TypeStringArray {
		return TypeString
	}
	return TypeNumber
}

func cloneTypes(types map[string]Type) map[string]Type {
	clone := make(map[string]Type, len(types))
	for name, typ := range types {
		clone[name] = typ
	}
	return clone
}

func verifyBlock(f Function, body []Instruction, known map[string]Type) error {
	for _, instruction := range body {
		if instruction.Op == OpIf {
			if instruction.Type != TypeVoid || instruction.Result != "" || len(instruction.Args) != 1 || known[instruction.Args[0]] != TypeBool {
				return fmt.Errorf("nested if requires one bool condition")
			}
			if err := verifyBlock(f, instruction.Then, cloneTypes(known)); err != nil {
				return err
			}
			if err := verifyBlock(f, instruction.Else, cloneTypes(known)); err != nil {
				return err
			}
			continue
		}
		if instruction.Op == OpReturn {
			if instruction.Type != f.ReturnType {
				return fmt.Errorf("return type %q does not match function return type %q", instruction.Type, f.ReturnType)
			}
			continue
		}
		if instruction.Result != "" && instruction.Type != "" {
			known[instruction.Result] = instruction.Type
		}
	}
	return nil
}
