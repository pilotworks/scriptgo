package typescriptgo

// Diagnostic is the stable diagnostic shape exposed to scriptgo.
type Diagnostic struct {
	FileName string
	Kind     string
	Message  string
	Start    int
	Length   int
	Code     int32
}

// SourceFile is the normalized source-file metadata exposed to scriptgo.
type SourceFile struct {
	FileName       string
	Source         string
	StatementCount int
	Span           SourceSpan
	Imports        []ModuleReference
	Symbols        []Symbol
	Syntax         SyntaxFile
	Builtin        bool
	BuiltinName    string
}

// Symbol is stable declaration metadata for frontend consumers. TypeScript-Go
// remains responsible for resolving and formatting the underlying type.
type Symbol struct {
	Name     string
	Kind     string
	Type     string
	Span     SourceSpan
	Exported bool
}

// ModuleReference records a TypeScript-Go-resolved local module edge.
type ModuleReference struct {
	Specifier        string
	ResolvedFileName string
	LocalName        string
	Span             SourceSpan
	Builtin          bool
}

type SourceSpan struct {
	Start  int
	Length int
}

// SyntaxFile is the small, backend-independent syntax model used by lowering.
// TypeScript-Go remains responsible for parsing and type checking; this model
// only prevents native stages from depending on TypeScript-Go internals.
type SyntaxFile struct {
	FileName   string
	Statements []SyntaxStatement
}

type SyntaxStatement struct {
	Span           SourceSpan
	Kind           string
	Name           string
	Label          string
	Type           string
	InferredType   string
	TypeParameters []string
	Expression     *SyntaxExpression
	Left           *SyntaxExpression
	Right          *SyntaxExpression
	Parameters     []SyntaxParameter
	Body           []SyntaxStatement
	Step           []SyntaxStatement
	Then           []SyntaxStatement
	Else           []SyntaxStatement
	Cases          []SyntaxSwitchCase
	CatchVar       string
	Catch          []SyntaxStatement
	Finally        []SyntaxStatement
	Class          *SyntaxClass
	Enum           *SyntaxEnum
	IsGenerator    bool
	IsAsync        bool
}

type SyntaxSwitchCase struct {
	Span       SourceSpan
	Expression *SyntaxExpression
	Statements []SyntaxStatement
}

type SyntaxEnum struct {
	Span    SourceSpan
	Name    string
	Members []SyntaxEnumMember
}

type SyntaxEnumMember struct {
	Span        SourceSpan
	Name        string
	Value       string
	Initializer *SyntaxExpression
}

type SyntaxClass struct {
	Span           SourceSpan
	Name           string
	TypeParameters []string
	Extends        string
	Implements     []string
	IsAbstract     bool
	Fields         []SyntaxField
	Constructor    *SyntaxConstructor
	Methods        []SyntaxMethod
	StaticBlocks   [][]SyntaxStatement
}

type SyntaxConstructor struct {
	Span       SourceSpan
	Parameters []SyntaxParameter
	Body       []SyntaxStatement
}

type SyntaxMethod struct {
	Span           SourceSpan
	Name           string
	Type           string
	InferredType   string
	TypeParameters []string
	Parameters     []SyntaxParameter
	Body           []SyntaxStatement
	IsStatic       bool
	IsAbstract     bool
	Kind           string
	IsGenerator    bool
	IsAsync        bool
}

type SyntaxField struct {
	Span         SourceSpan
	Name         string
	Type         string
	InferredType string
	Initializer  *SyntaxExpression
	IsStatic     bool
	IsPrivate    bool
	IsReadonly   bool
}

type SyntaxParameter struct {
	Span         SourceSpan
	Name         string
	Type         string
	InferredType string
	Rest         bool
	Initializer  *SyntaxExpression
}

type SyntaxExpression struct {
	Span          SourceSpan
	Kind          string
	Text          string
	InferredType  string
	Operator      string
	Left          *SyntaxExpression
	Right         *SyntaxExpression
	Arguments     []*SyntaxExpression
	TypeArguments []string
	WhenTrue      *SyntaxExpression
	WhenFalse     *SyntaxExpression
	Function      *SyntaxStatement
}

// ProgramResult contains the resolved files and all frontend diagnostics.
type ProgramResult struct {
	Options     CompilerOptions
	Files       []SourceFile
	Diagnostics []Diagnostic
}

// CompilerOptions records the normalized TypeScript-Go options used by the
// checked program, without exposing upstream option types across the boundary.
type CompilerOptions struct {
	Target           string
	Module           string
	ModuleResolution string
	Strict           bool
}

// ParseResult contains the parser output needed by the native frontend.
type ParseResult struct {
	StatementCount int
	Diagnostics    []Diagnostic
}
