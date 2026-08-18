package typescriptgo

import (
	"context"
	"path/filepath"
	"sort"
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
	Span           SourceSpan
	Imports        []ModuleReference
	Syntax         SyntaxFile
}

// ModuleReference records a TypeScript-Go-resolved local module edge.
type ModuleReference struct {
	Specifier        string
	ResolvedFileName string
	Span             SourceSpan
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
	Parameters []SyntaxParameter
	Body       []SyntaxStatement
}

type SyntaxParameter struct {
	Span SourceSpan
	Name string
	Type string
}

type SyntaxExpression struct {
	Span      SourceSpan
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
	files := make(map[string]*ast.SourceFile)
	for _, file := range program.GetSourceFiles() {
		if program.IsSourceFileDefaultLibrary(file.Path()) || !isTypeScriptSource(file.FileName()) {
			continue
		}
		files[filepath.Clean(file.FileName())] = file
	}
	for _, file := range orderedSourceFiles(files, absoluteEntry, program) {
		if program.IsSourceFileDefaultLibrary(file.Path()) || !isTypeScriptSource(file.FileName()) {
			continue
		}
		result.Files = append(result.Files, SourceFile{
			FileName:       file.FileName(),
			StatementCount: statementCount(file),
			Span:           sourceSpan(&file.Node),
			Imports:        moduleReferences(program, file),
			Syntax:         syntaxFile(file),
		})
		result.Diagnostics = append(result.Diagnostics, convertDiagnostics("syntax", program.GetSyntacticDiagnostics(ctx, file))...)
		result.Diagnostics = append(result.Diagnostics, convertDiagnostics("type", program.GetSemanticDiagnostics(ctx, file))...)
	}
	result.Diagnostics = append(result.Diagnostics, convertDiagnostics("program", program.GetProgramDiagnostics())...)
	return result, nil
}

func moduleReferences(program *compiler.Program, file *ast.SourceFile) []ModuleReference {
	result := make([]ModuleReference, 0, len(file.Imports()))
	for _, specifier := range file.Imports() {
		resolved := program.GetResolvedModuleFromModuleSpecifier(file, specifier)
		if resolved == nil || resolved.ResolvedFileName == "" || resolved.IsExternalLibraryImport {
			continue
		}
		result = append(result, ModuleReference{
			Specifier:        specifier.Text(),
			ResolvedFileName: filepath.Clean(resolved.ResolvedFileName),
			Span:             sourceSpan(specifier),
		})
	}
	return result
}

// orderedSourceFiles returns reachable local files in dependency-first order.
func orderedSourceFiles(files map[string]*ast.SourceFile, entry string, program *compiler.Program) []*ast.SourceFile {
	ordered := make([]*ast.SourceFile, 0, len(files))
	visited := make(map[string]bool, len(files))
	var visit func(string)
	visit = func(fileName string) {
		fileName = filepath.Clean(fileName)
		if visited[fileName] {
			return
		}
		file, ok := files[fileName]
		if !ok {
			return
		}
		visited[fileName] = true
		for _, reference := range moduleReferences(program, file) {
			visit(reference.ResolvedFileName)
		}
		ordered = append(ordered, file)
	}
	visit(entry)

	// Keep the adapter deterministic even if the compiler returns an unexpected
	// disconnected source file in a future TypeScript-Go revision.
	remaining := make([]string, 0, len(files)-len(ordered))
	for fileName := range files {
		if !visited[fileName] {
			remaining = append(remaining, fileName)
		}
	}
	sort.Strings(remaining)
	for _, fileName := range remaining {
		ordered = append(ordered, files[fileName])
	}
	return ordered
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
	span := sourceSpan(node)
	switch node.Kind {
	case ast.KindVariableStatement:
		declarations := node.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes
		if len(declarations) != 1 {
			return SyntaxStatement{Span: span, Kind: "unsupported", Type: "multiple variable declarations"}, true
		}
		declaration := declarations[0]
		result := SyntaxStatement{Span: span, Kind: "variable", Name: declaration.Name().Text()}
		result.Type = syntaxType(declaration.Type())
		result.Expression = syntaxExpression(declaration.Initializer())
		return result, true
	case ast.KindFunctionDeclaration:
		result := SyntaxStatement{Span: span, Kind: "function", Name: node.Name().Text(), Type: syntaxType(node.Type())}
		for _, parameter := range node.Parameters() {
			result.Parameters = append(result.Parameters, SyntaxParameter{
				Span: parameterSpan(parameter),
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
		return SyntaxStatement{Span: span, Kind: "return", Expression: syntaxExpression(node.Expression())}, true
	case ast.KindExpressionStatement:
		return SyntaxStatement{Span: span, Kind: "expression", Expression: syntaxExpression(node.Expression())}, true
	case ast.KindImportDeclaration, ast.KindExportDeclaration:
		return SyntaxStatement{Span: span, Kind: "module", Type: node.Kind.String()}, true
	default:
		return SyntaxStatement{Span: span, Kind: "unsupported", Type: node.Kind.String()}, true
	}
}

func syntaxExpression(node *ast.Node) *SyntaxExpression {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindNumericLiteral:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "number", Text: node.Text()}
	case ast.KindStringLiteral:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "string", Text: node.Text()}
	case ast.KindTrueKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "bool", Text: "true"}
	case ast.KindFalseKeyword:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "bool", Text: "false"}
	case ast.KindIdentifier:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "identifier", Text: node.Text()}
	case ast.KindParenthesizedExpression:
		return syntaxExpression(node.Expression())
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		return &SyntaxExpression{
			Span:     sourceSpan(node),
			Kind:     "binary",
			Operator: binaryOperator(binary.OperatorToken.Kind.String()),
			Left:     syntaxExpression(binary.Left),
			Right:    syntaxExpression(binary.Right),
		}
	case ast.KindCallExpression:
		result := &SyntaxExpression{Span: sourceSpan(node), Kind: "call", Left: syntaxExpression(node.Expression())}
		for _, argument := range node.Arguments() {
			result.Arguments = append(result.Arguments, syntaxExpression(argument))
		}
		return result
	case ast.KindPropertyAccessExpression:
		return &SyntaxExpression{
			Span: sourceSpan(node),
			Kind: "property",
			Text: node.Name().Text(),
			Left: syntaxExpression(node.Expression()),
		}
	default:
		return &SyntaxExpression{Span: sourceSpan(node), Kind: "unsupported", Text: node.Kind.String()}
	}
}

func sourceSpan(node *ast.Node) SourceSpan {
	if node == nil {
		return SourceSpan{}
	}
	return SourceSpan{Start: node.Pos(), Length: node.End() - node.Pos()}
}

func parameterSpan(node *ast.Node) SourceSpan {
	return sourceSpan(node)
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
