package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/core"
	"github.com/microsoft/typescript-go/internal/parser"
	"github.com/microsoft/typescript-go/internal/tspath"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting wd: %v\n", err)
		os.Exit(1)
	}

	var stdlibDir string

	candidates := []string{
		filepath.Join(wd, "internal", "typescriptgo", "stdlib"),
		filepath.Join(wd, "stdlib"),
		filepath.Join(wd, "..", "stdlib"),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			stdlibDir = c
			break
		}
	}
	if stdlibDir == "" {
		fmt.Fprintf(os.Stderr, "Could not locate stdlib directory from %s\n", wd)
		os.Exit(1)
	}

	entries, err := os.ReadDir(stdlibDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdlib dir %s: %v\n", stdlibDir, err)
		os.Exit(1)
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == "globals.d.ts" {
			continue
		}

		if !strings.HasSuffix(name, ".ts") || strings.HasSuffix(name, ".d.ts") {
			continue
		}

		srcPath := filepath.Join(stdlibDir, name)
		contentBytes, err := os.ReadFile(srcPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", srcPath, err)
			continue
		}
		source := string(contentBytes)

		file := parser.ParseSourceFile(ast.SourceFileParseOptions{
			FileName: srcPath,
			Path:     tspath.ToPath(srcPath, "", true),
		}, source, core.ScriptKindTS)

		var lines []string

		for _, stmt := range file.Statements.Nodes {
			// 1. Skip internal module/namespace __scriptgo
			if stmt.Kind == ast.KindModuleDeclaration {
				continue
			}

			// 2. Exported functions
			if stmt.Kind == ast.KindFunctionDeclaration {
				if hasExportModifier(stmt.Modifiers()) {
					sig := formatFunctionSignature(stmt, source)
					if sig != "" {
						lines = append(lines, sig)
					}
				}
				continue
			}

			// 3. Exported variable statements (e.g. export const argv: string[] = ...)
			if stmt.Kind == ast.KindVariableStatement {
				if hasExportModifier(stmt.Modifiers()) {
					varStmt := stmt.AsVariableStatement()
					decl := formatVariableStatement(varStmt, source)
					if decl != "" {
						lines = append(lines, decl)
					}
				}
				continue
			}

			// 4. Exported interfaces / types
			if stmt.Kind == ast.KindInterfaceDeclaration {
				if hasExportModifier(stmt.Modifiers()) {
					lines = append(lines, strings.TrimSpace(source[stmt.Pos():stmt.End()]))
				}
				continue
			}
			if stmt.Kind == ast.KindTypeAliasDeclaration {
				if hasExportModifier(stmt.Modifiers()) {
					lines = append(lines, strings.TrimSpace(source[stmt.Pos():stmt.End()]))
				}
				continue
			}

			// 5. Exported classes
			if stmt.Kind == ast.KindClassDeclaration {
				if hasExportModifier(stmt.Modifiers()) {
					classDecl := stmt.AsClassDeclaration()
					className := classDecl.Name().Text()
					var memberLines []string
					for _, member := range classDecl.Members.Nodes {
						if member.Kind == ast.KindConstructor {
							sig := formatConstructorSignature(member, source)
							if sig != "" {
								memberLines = append(memberLines, "    "+sig)
							}
						} else if member.Kind == ast.KindMethodDeclaration {
							if !hasPrivateModifier(member.Modifiers()) {
								sig := formatFunctionSignature(member, source)
								if sig != "" {
									memberLines = append(memberLines, "    "+sig)
								}
							}
						}
					}
					lines = append(lines, fmt.Sprintf("export class %s {\n%s\n}", className, strings.Join(memberLines, "\n")))
				}
				continue
			}
		}

		dtsContent := strings.Join(lines, "\n") + "\n"
		dtsName := strings.TrimSuffix(name, ".ts") + ".d.ts"
		stdlibDtsPath := filepath.Join(stdlibDir, dtsName)

		if err := os.WriteFile(stdlibDtsPath, []byte(dtsContent), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", stdlibDtsPath, err)
			continue
		}
		fmt.Printf("Generated %s -> %s\n", dtsName, stdlibDtsPath)
	}
}

func hasExportModifier(modifiers *ast.ModifierList) bool {
	if modifiers == nil {
		return false
	}
	for _, m := range modifiers.Nodes {
		if m.Kind == ast.KindExportKeyword {
			return true
		}
	}
	return false
}

func hasPrivateModifier(modifiers *ast.ModifierList) bool {
	if modifiers == nil {
		return false
	}
	for _, m := range modifiers.Nodes {
		if m.Kind == ast.KindPrivateKeyword {
			return true
		}
	}
	return false
}

func formatConstructorSignature(stmt *ast.Node, src string) string {
	startPos := stmt.Pos()
	if body := stmt.Body(); body != nil {
		bodyPos := body.Pos()
		sig := strings.TrimSpace(src[startPos:bodyPos])
		return sig + ";"
	}
	sig := strings.TrimSpace(src[startPos:stmt.End()])
	if !strings.HasSuffix(sig, ";") {
		sig += ";"
	}
	return sig
}

func formatFunctionSignature(stmt *ast.Node, src string) string {
	startPos := stmt.Pos()
	if body := stmt.Body(); body != nil {
		bodyPos := body.Pos()
		sig := strings.TrimSpace(src[startPos:bodyPos])
		return sig + ";"
	}
	sig := strings.TrimSpace(src[startPos:stmt.End()])
	if !strings.HasSuffix(sig, ";") {
		sig += ";"
	}
	return sig
}

func formatVariableStatement(varStmt *ast.VariableStatement, src string) string {
	var decls []string
	declList := varStmt.DeclarationList.AsVariableDeclarationList()
	isConst := declList.Flags&ast.NodeFlagsConst != 0
	keyword := "export let"
	if isConst {
		keyword = "export const"
	}

	for _, decl := range declList.Declarations.Nodes {
		name := ""
		if decl.Name() != nil {
			name = strings.TrimSpace(src[decl.Name().Pos():decl.Name().End()])
		}
		typeAnnotation := ""
		if decl.Type() != nil {
			typeAnnotation = ": " + strings.TrimSpace(src[decl.Type().Pos():decl.Type().End()])
		}
		decls = append(decls, fmt.Sprintf("%s %s%s;", keyword, name, typeAnnotation))
	}

	return strings.Join(decls, "\n")
}
