package lowering

import (
	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// typeOnlyShapes exposes fields needed by runtime casts without emitting the
// implementation of a type-only dependency.
func typeOnlyShapes(allProgram, runtime frontend.Program) []ir.ObjectShape {
	runtimeFiles := make(map[string]bool, len(runtime.Files))
	for _, file := range runtime.Files {
		runtimeFiles[file.FileName] = true
	}

	seen := map[string]bool{}
	var result []ir.ObjectShape
	var visit func(string, typescriptgo.SyntaxStatement)
	visit = func(fileName string, statement typescriptgo.SyntaxStatement) {
		if statement.Kind == "class" || statement.Kind == "interface" || statement.Kind == "type_alias" {
			if statement.Class == nil || seen[statement.Class.Name] {
				return
			}
			seen[statement.Class.Name] = true
			shape := ir.ObjectShape{
				Name: statement.Class.Name,
				Span: toIRSpan(fileName, statement.Class.Span),
			}
			for _, field := range statement.Class.Fields {
				fieldType := field.Type
				if fieldType == "" {
					fieldType = field.InferredType
				}
				value := ""
				if field.Initializer != nil {
					value = field.Initializer.Text
				}
				shape.Fields = append(shape.Fields, ir.Field{
					Name:     field.Name,
					Type:     toIRType(fieldType),
					Value:    value,
					Optional: field.Optional,
					Span:     toIRSpan(fileName, field.Span),
				})
			}
			if len(shape.Fields) > 0 {
				result = append(result, shape)
			}
			return
		}
		if statement.Kind == "block" || statement.Kind == "namespace" {
			for _, child := range statement.Body {
				visit(fileName, child)
			}
		}
	}

	for _, file := range allProgram.Files {
		if runtimeFiles[file.FileName] {
			continue
		}
		for _, statement := range file.Syntax.Statements {
			visit(file.FileName, statement)
		}
	}
	return result
}
