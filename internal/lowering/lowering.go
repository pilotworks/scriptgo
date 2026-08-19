// Package lowering converts the checked frontend model into backend-independent IR.
package lowering

import (
	"fmt"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// Lower lowers the currently supported synchronous TypeScript subset.
func Lower(program frontend.Program) (ir.Module, error) {
	if err := ValidateSubset(program); err != nil {
		return ir.Module{}, err
	}
	module := ir.Module{SourcePath: program.EntryPath, SourceFiles: make(map[string]string), StatementCount: program.StatementCount}
	for _, file := range program.Files {
		module.SourceFiles[file.FileName] = file.Source
	}
	shapes := map[string]ir.ObjectShape{}
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if statement.Kind != "class" || statement.Class == nil {
				continue
			}
			shape := ir.ObjectShape{Name: statement.Class.Name, Span: toIRSpan(file.FileName, statement.Class.Span)}
			for _, field := range statement.Class.Fields {
				val := ""
				if field.Initializer != nil {
					val = field.Initializer.Text
				} else if field.Type == "number" {
					val = "0"
				}
				shape.Fields = append(shape.Fields, ir.Field{Name: field.Name, Type: toIRType(field.Type), Value: val, Span: toIRSpan(file.FileName, field.Span)})
			}
			shapes[shape.Name] = shape
			module.Shapes = append(module.Shapes, shape)
		}
	}

	main := ir.Function{Name: "main", ReturnType: ir.TypeVoid}
	env := map[string]ir.Type{}
	signatures := buildFunctionIndex(program)
	counter := 0
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if statement.Kind == "function" {
				function, err := lowerFunction(file.FileName, statement, shapes, signatures)
				if err != nil {
					return ir.Module{}, fmt.Errorf("lower function %q: %w", statement.Name, sourceError(file.FileName, statement.Span, err))
				}
				module.Functions = append(module.Functions, function)
				continue
			}
			if statement.Kind == "module" {
				continue
			}
			if statement.Kind == "class" && statement.Class != nil {
				for _, method := range statement.Class.Methods {
					mangled := statement.Class.Name + "_" + method.Name
					methodStmt := typescriptgo.SyntaxStatement{
						Span: method.Span,
						Kind: "function",
						Name: mangled,
						Type: method.Type,
						Parameters: append([]typescriptgo.SyntaxParameter{
							{Name: "this", Type: "object:" + statement.Class.Name},
						}, method.Parameters...),
						Body: method.Body,
					}
					function, err := lowerFunction(file.FileName, methodStmt, shapes, signatures)
					if err != nil {
						return ir.Module{}, fmt.Errorf("lower class method %q: %w", mangled, sourceError(file.FileName, method.Span, err))
					}
					module.Functions = append(module.Functions, function)
				}
				continue
			}
			if err := lowerStatement(file.FileName, statement, &main, env, &counter, shapes, signatures); err != nil {
				return ir.Module{}, fmt.Errorf("lower %s: %w", statement.Kind, sourceError(file.FileName, statement.Span, err))
			}
		}
	}
	main.Body = append(main.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid})
	module.Functions = append([]ir.Function{main}, module.Functions...)
	if err := module.Verify(); err != nil {
		return ir.Module{}, err
	}
	return module, nil
}
