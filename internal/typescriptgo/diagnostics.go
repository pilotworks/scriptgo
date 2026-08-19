package typescriptgo

import "github.com/microsoft/typescript-go/internal/ast"

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
