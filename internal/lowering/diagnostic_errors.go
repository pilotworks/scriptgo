package lowering

import (
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func toIRSpan(path string, span frontend.SourceSpan) ir.SourceSpan {
	return ir.SourceSpan{Path: path, Offset: span.Start, Length: span.Length}
}

func sourceError(path string, span frontend.SourceSpan, err error) error {
	if err == nil {
		return nil
	}
	return &DiagnosticError{Path: path, Offset: span.Start, Length: span.Length, Err: err}
}

type DiagnosticError struct {
	Path   string
	Offset int
	Length int
	Err    error
}

func (e *DiagnosticError) Error() string {
	return e.Err.Error()
}
