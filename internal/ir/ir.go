// Package ir defines the stable contract between the frontend and native backends.
package ir

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
	Op   string
	Type Type
}
