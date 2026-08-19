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
	Span       SourceSpan
	Kind       string
	Name       string
	Type       string
	Expression *SyntaxExpression
	Left       *SyntaxExpression
	Right      *SyntaxExpression
	Parameters []SyntaxParameter
	Body       []SyntaxStatement
	Then       []SyntaxStatement
	Else       []SyntaxStatement
	CatchVar   string
	Catch      []SyntaxStatement
	Finally    []SyntaxStatement
	Class      *SyntaxClass
}

type SyntaxClass struct {
	Span        SourceSpan
	Name        string
	Fields      []SyntaxField
	Constructor *SyntaxConstructor
	Methods     []SyntaxMethod
}

type SyntaxConstructor struct {
	Span       SourceSpan
	Parameters []SyntaxParameter
	Body       []SyntaxStatement
}

type SyntaxMethod struct {
	Span       SourceSpan
	Name       string
	Type       string
	Parameters []SyntaxParameter
	Body       []SyntaxStatement
}

type SyntaxField struct {
	Span        SourceSpan
	Name        string
	Type        string
	Initializer *SyntaxExpression
}

type SyntaxParameter struct {
	Span        SourceSpan
	Name        string
	Type        string
	Rest        bool
	Initializer *SyntaxExpression
}

type SyntaxExpression struct {
	Span      SourceSpan
	Kind      string
	Text      string
	Operator  string
	Left      *SyntaxExpression
	Right     *SyntaxExpression
	Arguments []*SyntaxExpression
	WhenTrue  *SyntaxExpression
	WhenFalse *SyntaxExpression
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
