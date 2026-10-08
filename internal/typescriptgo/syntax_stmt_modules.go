package typescriptgo

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

func syntaxImportDeclaration(node *ast.Node, span SourceSpan) (SyntaxStatement, bool) {
	var aliases []SyntaxStatement
	if imp := node.AsImportDeclaration(); imp != nil && imp.ImportClause != nil {
		if clause := imp.ImportClause.AsImportClause(); clause != nil && clause.NamedBindings != nil && clause.NamedBindings.Kind == ast.KindNamedImports {
			if namedImports := clause.NamedBindings.AsNamedImports(); namedImports != nil && namedImports.Elements != nil {
				for _, elem := range namedImports.Elements.Nodes {
					if spec := elem.AsImportSpecifier(); spec != nil {
						localName := spec.Name().Text()
						origName := localName
						if spec.PropertyName != nil {
							origName = spec.PropertyName.Text()
						}
						if origName != localName {
							aliases = append(aliases, SyntaxStatement{
								Span: span,
								Kind: "import_alias",
								Name: localName,
								Type: origName,
							})
						}
					}
				}
			}
		}
	}
	if len(aliases) > 0 {
		return SyntaxStatement{Span: span, Kind: "block", Body: aliases}, true
	}
	return SyntaxStatement{Span: span, Kind: "module", Type: node.Kind.String()}, true
}

func syntaxExportDeclaration(node *ast.Node, span SourceSpan) (SyntaxStatement, bool) {
	var aliases []SyntaxStatement
	if exp := node.AsExportDeclaration(); exp != nil && exp.ExportClause != nil && exp.ExportClause.Kind == ast.KindNamedExports {
		if namedExports := exp.ExportClause.AsNamedExports(); namedExports != nil && namedExports.Elements != nil {
			for _, elem := range namedExports.Elements.Nodes {
				if spec := elem.AsExportSpecifier(); spec != nil {
					exportName := spec.Name().Text()
					origName := exportName
					if spec.PropertyName != nil {
						origName = spec.PropertyName.Text()
					}
					if origName != exportName {
						aliases = append(aliases, SyntaxStatement{
							Span: span,
							Kind: "export_alias",
							Name: exportName,
							Type: origName,
						})
					}
				}
			}
		}
	}
	if len(aliases) > 0 {
		return SyntaxStatement{Span: span, Kind: "block", Body: aliases}, true
	}
	return SyntaxStatement{Span: span, Kind: "module", Type: node.Kind.String()}, true
}

func syntaxModuleDeclaration(node *ast.Node, chk *checker.Checker, span SourceSpan) (SyntaxStatement, bool) {
	modDecl := node.AsModuleDeclaration()
	name := ""
	if node.Name() != nil {
		name = strings.Trim(node.Name().Text(), "\"'")
	}
	var bodyStmts []SyntaxStatement
	if modDecl != nil && modDecl.Body != nil {
		if modBlock := modDecl.Body.AsModuleBlock(); modBlock != nil && modBlock.Statements != nil {
			for _, s := range modBlock.Statements.Nodes {
				if conv, ok := syntaxStatement(s, chk); ok {
					if conv.Kind == "class" && conv.Class != nil && !strings.Contains(conv.Class.Name, ".") {
						conv.Class.Name = name + "." + conv.Class.Name
						conv.Name = conv.Class.Name
					}
					if (conv.Kind == "function" || conv.Kind == "async_function") && conv.Type != "" && !strings.Contains(conv.Type, ".") {
						for _, other := range modBlock.Statements.Nodes {
							if other.Kind == ast.KindClassDeclaration {
								cDecl := other.AsClassDeclaration()
								if cDecl != nil && cDecl.Name() != nil && cDecl.Name().Text() == conv.Type {
									conv.Type = name + "." + conv.Type
									break
								}
							}
						}
					}
					bodyStmts = append(bodyStmts, conv)
				}
			}
		}
	}
	return SyntaxStatement{
		Span: span,
		Kind: "namespace",
		Name: name,
		Body: bodyStmts,
	}, true
}
