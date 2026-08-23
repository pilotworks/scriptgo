package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
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
					heritage := ""
					if classDecl.HeritageClauses != nil {
						for _, clause := range classDecl.HeritageClauses.Nodes {
							if clause != nil {
								hText := strings.TrimSpace(source[clause.Pos():clause.End()])
								if hText != "" {
									heritage += " " + hText
								}
							}
						}
					}
					for _, member := range classDecl.Members.Nodes {
						switch member.Kind {
						case ast.KindConstructor:
							sig := formatConstructorSignature(member, source)
							if sig != "" {
								memberLines = append(memberLines, "    "+sig)
							}
						case ast.KindMethodDeclaration:
							if !hasPrivateModifier(member.Modifiers()) {
								sig := formatFunctionSignature(member, source)
								if sig != "" {
									if hasStaticModifier(member.Modifiers()) && !strings.HasPrefix(sig, "static ") {
										sig = "static " + sig
									}
									memberLines = append(memberLines, "    "+sig)
								}
							}
						case ast.KindPropertyDeclaration:
							if !hasPrivateModifier(member.Modifiers()) {
								prop := member.AsPropertyDeclaration()
								name := prop.Name().Text()
								fType := "unknown"
								if prop.Type != nil {
									fType = strings.TrimSpace(source[prop.Type.Pos():prop.Type.End()])
								}
								staticPrefix := ""
								if hasStaticModifier(member.Modifiers()) {
									staticPrefix = "static "
								}
								memberLines = append(memberLines, fmt.Sprintf("    %s%s: %s;", staticPrefix, name, fType))
							}
						}
					}
					lines = append(lines, fmt.Sprintf("export class %s%s {\n%s\n}", className, heritage, strings.Join(memberLines, "\n")))
				}
				continue
			}

			// 6. Export default assignment (e.g. export default console;)
			if stmt.Kind == ast.KindExportAssignment {
				exp := stmt.AsExportAssignment().Expression
				if exp != nil {
					expText := strings.TrimSpace(source[exp.Pos():exp.End()])
					typeName := expText
					if expText == "console" {
						typeName = "Console"
					}
					lines = append(lines, fmt.Sprintf("declare const _default: %s;\nexport default _default;", typeName))
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

func hasStaticModifier(modifiers *ast.ModifierList) bool {
	if modifiers == nil {
		return false
	}
	for _, m := range modifiers.Nodes {
		if m.Kind == ast.KindStaticKeyword {
			return true
		}
	}
	return false
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
		return cleanSignature(sig + ";")
	}
	sig := strings.TrimSpace(src[startPos:stmt.End()])
	if !strings.HasSuffix(sig, ";") {
		sig += ";"
	}
	return cleanSignature(sig)
}

func formatFunctionSignature(stmt *ast.Node, src string) string {
	startPos := stmt.Pos()
	if body := stmt.Body(); body != nil {
		bodyPos := body.Pos()
		sig := strings.TrimSpace(src[startPos:bodyPos])
		return cleanSignature(sig + ";")
	}
	sig := strings.TrimSpace(src[startPos:stmt.End()])
	if !strings.HasSuffix(sig, ";") {
		sig += ";"
	}
	return cleanSignature(sig)
}

func cleanSignature(sig string) string {
	if !strings.Contains(sig, "=") {
		return sig
	}
	openParen := strings.Index(sig, "(")
	closeParen := strings.LastIndex(sig, ")")
	if openParen == -1 || closeParen == -1 || closeParen <= openParen {
		return sig
	}
	prefix := sig[:openParen+1]
	paramsStr := sig[openParen+1 : closeParen]
	suffix := sig[closeParen:]

	if strings.TrimSpace(paramsStr) == "" {
		return sig
	}

	params := splitParams(paramsStr)
	var newParams []string
	for _, p := range params {
		trimmed := strings.TrimSpace(p)
		if idx := strings.Index(trimmed, "="); idx != -1 {
			leftPart := strings.TrimSpace(trimmed[:idx])
			if colonIdx := strings.Index(leftPart, ":"); colonIdx != -1 {
				name := strings.TrimSpace(leftPart[:colonIdx])
				typ := strings.TrimSpace(leftPart[colonIdx+1:])
				if !strings.HasSuffix(name, "?") {
					name = name + "?"
				}
				newParams = append(newParams, name+": "+typ)
			} else {
				if !strings.HasSuffix(leftPart, "?") {
					leftPart = leftPart + "?"
				}
				newParams = append(newParams, leftPart)
			}
		} else {
			newParams = append(newParams, trimmed)
		}
	}
	return prefix + strings.Join(newParams, ", ") + suffix
}

func splitParams(s string) []string {
	var params []string
	var current strings.Builder
	depth := 0
	for _, r := range s {
		switch r {
		case '(', '[', '{', '<':
			depth++
			current.WriteRune(r)
		case ')', ']', '}', '>':
			depth--
			current.WriteRune(r)
		case ',':
			if depth == 0 {
				params = append(params, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		params = append(params, current.String())
	}
	return params
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
