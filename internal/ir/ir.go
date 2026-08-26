// Package ir defines the stable contract between the frontend and native backends.
package ir

// Module is the first typed IR unit. Instructions will be added as lowering grows.
type Module struct {
	SourcePath     string
	SourceFiles    map[string]string
	StatementCount int
	Shapes         []ObjectShape
	Globals        []Global
	Functions      []Function
	Externs        []ExternFunction
}

type Global struct {
	Name  string
	Type  Type
	Value string
	Span  SourceSpan
}

type ExternFunction struct {
	Name       string
	Span       SourceSpan
	Parameters []Parameter
	ReturnType Type
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
	Captured   []Parameter
	ReturnType Type
	Body       []Instruction
}

type Parameter struct {
	Name string
	Type Type
}

type Type string

const (
	TypeVoid              Type = "void"
	TypeBool              Type = "bool"
	TypeNumber            Type = "number"
	TypeBigInt            Type = "bigint"
	TypeSymbol            Type = "symbol"
	TypeString            Type = "string"
	TypeNumberArray       Type = "number[]"
	TypeStringArray       Type = "string[]"
	TypeBoolArray         Type = "bool[]"
	TypeBigIntArray       Type = "bigint[]"
	TypeSymbolArray       Type = "symbol[]"
	TypeUnknownArray      Type = "unknown[]"
	TypeObject            Type = "object"
	TypeClosure           Type = "closure"
	TypeUnknown           Type = "unknown"
	TypeUint8Array        Type = "Uint8Array"
	TypeInt8Array         Type = "Int8Array"
	TypeUint8ClampedArray Type = "Uint8ClampedArray"
	TypeInt16Array        Type = "Int16Array"
	TypeUint16Array       Type = "Uint16Array"
	TypeInt32Array        Type = "Int32Array"
	TypeUint32Array       Type = "Uint32Array"
	TypeFloat32Array      Type = "Float32Array"
	TypeFloat64Array      Type = "Float64Array"
	TypeBigInt64Array     Type = "BigInt64Array"
	TypeBigUint64Array    Type = "BigUint64Array"
	TypeDataView          Type = "DataView"
	TypeArrayBuffer       Type = "ArrayBuffer"
	TypeMap               Type = "Map"
	TypeSet               Type = "Set"
	TypeTextEncoder       Type = "TextEncoder"
	TypeTextDecoder       Type = "TextDecoder"
	TypeBuffer            Type = "Buffer"
	TypePointer           Type = "ptr"
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
	Step       []Instruction
	CatchVar   string
	Catch      []Instruction
	Finally    []Instruction
}

const (
	OpConst       = "const"
	OpBinary      = "binary"
	OpCompare     = "compare"
	OpSelect      = "select"
	OpCall        = "call"
	OpExternCall  = "extern.call"
	OpPrint       = "print"
	OpParam       = "param"
	OpReturn      = "return"
	OpArray       = "array"
	OpIndex       = "index"
	OpIndexSet    = "index.set"
	OpObjectNew   = "object.new"
	OpFieldGet    = "field.get"
	OpFieldSet    = "field.set"
	OpIf          = "if"
	OpWhile       = "while"
	OpDoWhile     = "dowhile"
	OpBreak       = "break"
	OpContinue    = "continue"
	OpAssign      = "assign"
	OpThrow       = "throw"
	OpTry         = "try"
	OpClosure     = "closure"
	OpClosureCall = "closure.call"
	OpInstanceOf  = "instanceof"
	OpBoxUnknown  = "box.unknown"
	OpCheckedCast = "checked_cast"
	OpTypeOf      = "typeof"
	OpDebugger    = "debugger"
)
