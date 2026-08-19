// Package ir defines the stable contract between the frontend and native backends.
package ir

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
	TypeStringArray Type = "string[]"
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
	Then       []Instruction
	Else       []Instruction
	Cond       []Instruction
	Body       []Instruction
}

const (
	OpConst     = "const"
	OpBinary    = "binary"
	OpCompare   = "compare"
	OpSelect    = "select"
	OpCall      = "call"
	OpPrint     = "print"
	OpParam     = "param"
	OpReturn    = "return"
	OpArray     = "array"
	OpIndex     = "index"
	OpIndexSet  = "index.set"
	OpObjectNew = "object.new"
	OpFieldGet  = "field.get"
	OpFieldSet  = "field.set"
	OpIf        = "if"
	OpWhile     = "while"
	OpDoWhile   = "dowhile"
	OpBreak     = "break"
	OpContinue  = "continue"
	OpAssign    = "assign"
)
