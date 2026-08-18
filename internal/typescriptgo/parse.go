package typescriptgo

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/bundled"
	"github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/parser"
	"github.com/microsoft/typescript-go/internal/tsoptions"
	"github.com/microsoft/typescript-go/internal/tspath"
	"github.com/microsoft/typescript-go/internal/vfs/osvfs"
)

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
	StatementCount int
	Syntax         SyntaxFile
}

// SyntaxFile is the small, backend-independent syntax model used by lowering.
// TypeScript-Go remains responsible for parsing and type checking; this model
// only prevents native stages from depending on TypeScript-Go internals.
type SyntaxFile struct {
	FileName   string
	Statements []SyntaxStatement
}

type SyntaxStatement struct {
	Kind       string
	Name       string
	Type       string
	Expression *SyntaxExpression
	Parameters []SyntaxParameter
	Body       []SyntaxStatement
}

type SyntaxParameter struct {
	Name string
	Type string
}

type SyntaxExpression struct {
	Kind      string
	Text      string
	Operator  string
	Left      *SyntaxExpression
	Right     *SyntaxExpression
	Arguments []*SyntaxExpression
}

// ProgramResult contains the resolved files and all frontend diagnostics.
type ProgramResult struct {
	Files       []SourceFile
	Diagnostics []Diagnostic
}

// ParseResult contains the parser output needed by the native frontend.
type ParseResult struct {
	StatementCount int
	Diagnostics    []Diagnostic
}

// Parse parses one TypeScript source file with TypeScript-Go.
func Parse(fileName, source string) (ParseResult, error) {
	absoluteName, err := filepath.Abs(fileName)
	if err != nil {
		return ParseResult{}, err
	}
	absoluteName = filepath.Clean(absoluteName)

	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: absoluteName,
		Path:     tspath.ToPath(absoluteName, "", true),
	}, source, core.ScriptKindTS)

	result := ParseResult{}
	if file.Statements != nil {
		result.StatementCount = len(file.Statements.Nodes)
	}
	for _, diagnostic := range file.Diagnostics() {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Message: diagnostic.String(),
			Start:   diagnostic.Pos(),
			Length:  diagnostic.Len(),
			Code:    diagnostic.Code(),
		})
	}
	return result, nil
}

// Check creates a TypeScript-Go program. Program creation performs local module
// resolution and binds/checks the complete reachable source graph.
func Check(entryPath string) (ProgramResult, error) {
	absoluteEntry, err := filepath.Abs(entryPath)
	if err != nil {
		return ProgramResult{}, err
	}
	absoluteEntry = filepath.Clean(absoluteEntry)
	cwd := filepath.Dir(absoluteEntry)
	fs := bundled.WrapFS(osvfs.FS())
	options := &core.CompilerOptions{
		Target:           core.ScriptTargetES2020,
		Module:           core.ModuleKindESNext,
		ModuleResolution: core.ModuleResolutionKindBundler,
		Strict:           core.TSTrue,
		NoEmit:           core.TSTrue,
	}
	comparePaths := tspath.ComparePathsOptions{
		CurrentDirectory:          cwd,
		UseCaseSensitiveFileNames: fs.UseCaseSensitiveFileNames(),
	}
	config := tsoptions.NewParsedCommandLine(options, []string{absoluteEntry}, comparePaths)
	host := compiler.NewCompilerHost(cwd, fs, bundled.LibPath(), nil, nil)
	program := compiler.NewProgram(compiler.ProgramOptions{
		Config:         config,
		Host:           host,
		SingleThreaded: core.TSTrue,
	})

	result := ProgramResult{}
	ctx := context.Background()
	for _, file := range program.GetSourceFiles() {
		if program.IsSourceFileDefaultLibrary(file.Path()) || !isTypeScriptSource(file.FileName()) {
			continue
		}
		result.Files = append(result.Files, SourceFile{
			FileName:       file.FileName(),
			StatementCount: statementCount(file),
			Syntax:         syntaxFile(file),
		})
		result.Diagnostics = append(result.Diagnostics, convertDiagnostics("syntax", program.GetSyntacticDiagnostics(ctx, file))...)
		result.Diagnostics = append(result.Diagnostics, convertDiagnostics("type", program.GetSemanticDiagnostics(ctx, file))...)
	}
	result.Diagnostics = append(result.Diagnostics, convertDiagnostics("program", program.GetProgramDiagnostics())...)
	return result, nil
}

func syntaxFile(file *ast.SourceFile) SyntaxFile {
	result := SyntaxFile{FileName: file.FileName()}
	if file.Statements == nil {
		return result
	}
	for _, statement := range file.Statements.Nodes {
		if converted, ok := syntaxStatement(statement); ok {
			result.Statements = append(result.Statements, converted)
		}
	}
	return result
}

func syntaxStatement(node *ast.Node) (SyntaxStatement, bool) {
	switch node.Kind {
	case ast.KindVariableStatement:
		declarations := node.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes
		if len(declarations) != 1 {
			return SyntaxStatement{Kind: "unsupported", Type: "multiple variable declarations"}, true
		}
		declaration := declarations[0]
		result := SyntaxStatement{Kind: "variable", Name: declaration.Name().Text()}
		result.Type = syntaxType(declaration.Type())
		result.Expression = syntaxExpression(declaration.Initializer())
		return result, true
	case ast.KindFunctionDeclaration:
		result := SyntaxStatement{Kind: "function", Name: node.Name().Text(), Type: syntaxType(node.Type())}
		for _, parameter := range node.Parameters() {
			result.Parameters = append(result.Parameters, SyntaxParameter{
				Name: parameter.Name().Text(),
				Type: syntaxType(parameter.Type()),
			})
		}
		if body := node.Body(); body != nil {
			for _, statement := range body.Statements() {
				if converted, ok := syntaxStatement(statement); ok {
					result.Body = append(result.Body, converted)
				}
			}
		}
		return result, true
	case ast.KindReturnStatement:
		return SyntaxStatement{Kind: "return", Expression: syntaxExpression(node.Expression())}, true
	case ast.KindExpressionStatement:
		return SyntaxStatement{Kind: "expression", Expression: syntaxExpression(node.Expression())}, true
	case ast.KindImportDeclaration, ast.KindExportDeclaration:
		return SyntaxStatement{Kind: "module", Type: node.Kind.String()}, true
	default:
		return SyntaxStatement{Kind: "unsupported", Type: node.Kind.String()}, true
	}
}

func syntaxExpression(node *ast.Node) *SyntaxExpression {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindNumericLiteral:
		return &SyntaxExpression{Kind: "number", Text: node.Text()}
	case ast.KindStringLiteral:
		return &SyntaxExpression{Kind: "string", Text: node.Text()}
	case ast.KindTrueKeyword:
		return &SyntaxExpression{Kind: "bool", Text: "true"}
	case ast.KindFalseKeyword:
		return &SyntaxExpression{Kind: "bool", Text: "false"}
	case ast.KindIdentifier:
		return &SyntaxExpression{Kind: "identifier", Text: node.Text()}
	case ast.KindParenthesizedExpression:
		return syntaxExpression(node.Expression())
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		return &SyntaxExpression{
			Kind:     "binary",
			Operator: binaryOperator(binary.OperatorToken.Kind.String()),
			Left:     syntaxExpression(binary.Left),
			Right:    syntaxExpression(binary.Right),
		}
	case ast.KindCallExpression:
		result := &SyntaxExpression{Kind: "call", Left: syntaxExpression(node.Expression())}
		for _, argument := range node.Arguments() {
			result.Arguments = append(result.Arguments, syntaxExpression(argument))
		}
		return result
	case ast.KindPropertyAccessExpression:
		return &SyntaxExpression{
			Kind: "property",
			Text: node.Name().Text(),
			Left: syntaxExpression(node.Expression()),
		}
	default:
		return &SyntaxExpression{Kind: "unsupported", Text: node.Kind.String()}
	}
}

func syntaxType(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindNumberKeyword:
		return "number"
	case ast.KindStringKeyword:
		return "string"
	case ast.KindBooleanKeyword:
		return "bool"
	case ast.KindVoidKeyword:
		return "void"
	default:
		return node.Kind.String()
	}
}

func binaryOperator(kind string) string {
	kind = strings.TrimPrefix(kind, "Kind")
	operators := map[string]string{
		"PlusToken":               "+",
		"MinusToken":              "-",
		"AsteriskToken":           "*",
		"SlashToken":              "/",
		"PercentToken":            "%",
		"EqualsEqualsToken":       "==",
		"EqualsEqualsEqualsToken": "===",
		"LessThanToken":           "<",
		"LessThanEqualsToken":     "<=",
		"GreaterThanToken":        ">",
		"GreaterThanEqualsToken":  ">=",
	}
	if operator, ok := operators[kind]; ok {
		return operator
	}
	return kind
}

func convertDiagnostics(kind string, diagnostics []*ast.Diagnostic) []Diagnostic {
	result := make([]Diagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		fileName := ""
		if file := diagnostic.File(); file != nil {
			fileName = file.FileName()
		}
		result = append(result, Diagnostic{
			FileName: fileName,
			Kind:     kind,
			Message:  diagnostic.String(),
			Start:    diagnostic.Pos(),
			Length:   diagnostic.Len(),
			Code:     diagnostic.Code(),
		})
	}
	return result
}

func statementCount(file *ast.SourceFile) int {
	if file.Statements == nil {
		return 0
	}
	return len(file.Statements.Nodes)
}

func isTypeScriptSource(fileName string) bool {
	for _, extension := range []string{".ts", ".tsx", ".mts", ".cts"} {
		if strings.HasSuffix(fileName, extension) {
			return true
		}
	}
	return false
}
