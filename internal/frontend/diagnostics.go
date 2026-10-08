package frontend

import "github.com/microsoft/TypeScript/tsc/scriptgo"

// FormatDiagnostic renders a frontend diagnostic with its source code frame.
func FormatDiagnostic(diagnostic Diagnostic, source string) string {
	return typescriptgo.FormatDiagnostic(diagnostic, source)
}

// FormatSpan renders a diagnostic for a source span reported by a later stage
// (for example a native subset error from lowering) in the same format as
// TypeScript-Go diagnostics, so every stage reports source-anchored errors.
func FormatSpan(fileName string, start, length int, category, code, message, source string) string {
	return typescriptgo.Format(fileName, start, length, category, code, message, source)
}
