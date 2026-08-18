// Package ir defines the stable contract between the frontend and native backends.
package ir

import "fmt"

// Module is the first typed IR unit. Instructions will be added as lowering grows.
type Module struct {
	SourcePath     string
	StatementCount int
	Functions      []Function
}

type Function struct {
	Name       string
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
	TypeVoid   Type = "void"
	TypeBool   Type = "bool"
	TypeNumber Type = "number"
	TypeString Type = "string"
)

type Instruction struct {
	Op       string
	Type     Type
	Result   string
	Value    string
	Operator string
	Callee   string
	Args     []string
	Offset   int
}

const (
	OpConst  = "const"
	OpBinary = "binary"
	OpCall   = "call"
	OpPrint  = "print"
	OpParam  = "param"
	OpReturn = "return"
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
		case OpConst, OpBinary, OpCall, OpParam:
			if instruction.Result == "" || instruction.Type == "" {
				return fmt.Errorf("%s instruction must define result and type", instruction.Op)
			}
			known[instruction.Result] = instruction.Type
		case OpPrint:
			if len(instruction.Args) != 1 {
				return fmt.Errorf("print instruction requires one argument")
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
