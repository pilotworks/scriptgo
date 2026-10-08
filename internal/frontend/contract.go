package frontend

import "github.com/microsoft/TypeScript/tsc/scriptgo"

// The types below are the frontend-to-lowering contract. They name the stable,
// normalized data produced by the pinned TypeScript-Go adapter so that later
// stages depend on internal/frontend only and never import the adapter
// directly. They are aliases, not copies: the adapter remains the single
// definition of each shape, and a change to the adapter model is a change to
// this contract.

// Checked program and source-file data.
type (
	CompilerOptions = typescriptgo.CompilerOptions
	Diagnostic      = typescriptgo.Diagnostic
	ModuleBinding   = typescriptgo.ModuleBinding
	ModuleReference = typescriptgo.ModuleReference
	ProgramResult   = typescriptgo.ProgramResult
	SourceFile      = typescriptgo.SourceFile
	SourceSpan      = typescriptgo.SourceSpan
	Symbol          = typescriptgo.Symbol
)

// Normalized syntax consumed by lowering.
type (
	StaticElementKind   = typescriptgo.StaticElementKind
	SyntaxClass         = typescriptgo.SyntaxClass
	SyntaxDecorator     = typescriptgo.SyntaxDecorator
	SyntaxEnumMember    = typescriptgo.SyntaxEnumMember
	SyntaxExpression    = typescriptgo.SyntaxExpression
	SyntaxField         = typescriptgo.SyntaxField
	SyntaxFile          = typescriptgo.SyntaxFile
	SyntaxMethod        = typescriptgo.SyntaxMethod
	SyntaxParameter     = typescriptgo.SyntaxParameter
	SyntaxStatement     = typescriptgo.SyntaxStatement
	SyntaxStaticElement = typescriptgo.SyntaxStaticElement
	SyntaxSwitchCase    = typescriptgo.SyntaxSwitchCase
)

// Static class element kinds.
const (
	StaticElementBlock = typescriptgo.StaticElementBlock
	StaticElementField = typescriptgo.StaticElementField
)
