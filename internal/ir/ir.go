// Package ir defines the stable contract between the frontend and native backends.
package ir

import (
	"fmt"
	"strings"
)

// Module is the first typed IR unit. Instructions will be added as lowering grows.
type Module struct {
	SourcePath     string
	SourceFiles    map[string]string
	StatementCount int
	Shapes         []ObjectShape
	Functions      []Function
}

type ObjectShape struct {
	Name   string
	Span   SourceSpan
	Fields []Field
}

type Field struct {
	Name  string
	Type  Type
	Value string
	Span  SourceSpan
}

type SourceSpan struct {
	Path   string
	Offset int
	Length int
}

type Function struct {
	Name       string
	Span       SourceSpan
	Parameters []Parameter
	ReturnType Type
	Body       []Instruction
}

type Parameter struct {
	Name string
	Type Type
}

type Type string

const (
	TypeVoid        Type = "void"
	TypeBool        Type = "bool"
	TypeNumber      Type = "number"
	TypeString      Type = "string"
	TypeNumberArray Type = "number[]"
	TypeObject      Type = "object"
)

type Instruction struct {
	Op         string
	Type       Type
	Result     string
	Value      string
	Operator   string
	Callee     string
	Field      string
	FieldIndex int
	FieldCount int
	Args       []string
	Span       SourceSpan
}

const (
	OpConst     = "const"
	OpBinary    = "binary"
	OpCall      = "call"
	OpPrint     = "print"
	OpParam     = "param"
	OpReturn    = "return"
	OpArray     = "array"
	OpIndex     = "index"
	OpObjectNew = "object.new"
	OpFieldGet  = "field.get"
	OpFieldSet  = "field.set"
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
		case OpConst, OpBinary, OpCall, OpParam, OpArray, OpIndex, OpObjectNew, OpFieldGet:
			if instruction.Result == "" || instruction.Type == "" {
				return fmt.Errorf("%s instruction must define result and type", instruction.Op)
			}
			if instruction.Op == OpArray {
				if instruction.Type != TypeNumberArray {
					return fmt.Errorf("array instruction has unsupported type %q", instruction.Type)
				}
				for _, argument := range instruction.Args {
					if known[argument] != TypeNumber {
						return fmt.Errorf("array element %q is not a number", argument)
					}
				}
			}
			if instruction.Op == OpIndex {
				if len(instruction.Args) != 2 {
					return fmt.Errorf("index instruction requires array and index operands")
				}
				if known[instruction.Args[0]] != TypeNumberArray || known[instruction.Args[1]] != TypeNumber {
					return fmt.Errorf("index instruction requires number[] and number operands")
				}
				if instruction.Type != TypeNumber {
					return fmt.Errorf("index instruction must produce a number")
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
