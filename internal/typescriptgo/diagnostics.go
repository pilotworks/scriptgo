package typescriptgo

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/internal/ast"
)

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

// FormatDiagnostic formats a Diagnostic into standard TypeScript compiler (tsc) diagnostic style.
func FormatDiagnostic(diag Diagnostic, source string) string {
	codeStr := ""
	if diag.Code > 0 {
		codeStr = fmt.Sprintf("TS%d", diag.Code)
	}
	category := "error"
	if diag.Kind == "warning" {
		category = "warning"
	}
	return Format(diag.FileName, diag.Start, diag.Length, category, codeStr, diag.Message, source)
}

// Format formats a diagnostic into standard TypeScript compiler (tsc) diagnostic style:
// <file>:<line>:<col> - <category> <codeStr>: <message>
//
// <line> <lineText>
//        <squiggles>
func Format(fileName string, start, length int, category, codeStr, message, source string) string {
	if fileName == "" {
		fileName = "input.ts"
	}

	displayPath := fileName
	if cwd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(cwd, fileName); err == nil && !strings.HasPrefix(rel, "..") {
			displayPath = rel
		}
	}

	if source == "" {
		if content, err := os.ReadFile(fileName); err == nil {
			source = string(content)
		}
	}

	if source == "" {
		if codeStr != "" {
			return fmt.Sprintf("%s:1:1 - %s %s: %s", displayPath, category, codeStr, message)
		}
		return fmt.Sprintf("%s:1:1 - %s: %s", displayPath, category, message)
	}

	if start < 0 {
		start = 0
	}
	if start > len(source) {
		start = len(source)
	}

	// Compute line (1-indexed) and line start/end offsets
	line := 1
	lineStart := 0
	for i := 0; i < start; i++ {
		if source[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}

	lineEnd := len(source)
	for i := lineStart; i < len(source); i++ {
		if source[i] == '\n' || source[i] == '\r' {
			lineEnd = i
			break
		}
	}

	lineText := source[lineStart:lineEnd]

	// Compute 1-indexed column count using runes
	col := 1
	if start > lineStart {
		col = utf8.RuneCountInString(source[lineStart:start]) + 1
	}

	// Compute squiggles
	colIdx := col - 1
	if colIdx > len(lineText) {
		colIdx = len(lineText)
	}

	var prefixBuilder strings.Builder
	for i, r := range lineText {
		if i >= colIdx {
			break
		}
		if r == '\t' {
			prefixBuilder.WriteRune('\t')
		} else {
			prefixBuilder.WriteByte(' ')
		}
	}
	prefixWhitespace := prefixBuilder.String()

	spanLen := length
	if spanLen <= 0 {
		spanLen = 1
	}
	remaining := len(lineText) - colIdx
	if remaining <= 0 {
		spanLen = 1
	} else if spanLen > remaining {
		spanLen = remaining
	}

	squiggles := strings.Repeat("~", spanLen)
	lineNumStr := strconv.Itoa(line)
	gutterSpace := strings.Repeat(" ", len(lineNumStr))

	var header string
	if codeStr != "" {
		header = fmt.Sprintf("%s:%d:%d - %s %s: %s", displayPath, line, col, category, codeStr, message)
	} else {
		header = fmt.Sprintf("%s:%d:%d - %s: %s", displayPath, line, col, category, message)
	}

	return fmt.Sprintf("%s\n\n%s %s\n%s %s%s", header, lineNumStr, lineText, gutterSpace, prefixWhitespace, squiggles)
}
