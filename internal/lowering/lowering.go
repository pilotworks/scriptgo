// Package lowering converts the checked frontend model into backend-independent IR.
package lowering

import (
	"fmt"
	"sort"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// Lower lowers the currently supported synchronous TypeScript subset.
func Lower(program frontend.Program) (ir.Module, error) {
	var err error
	program, err = SpecializeGenerics(program)
	if err != nil {
		return ir.Module{}, err
	}
	if err := ValidateSubset(program); err != nil {
		return ir.Module{}, err
	}
	module := ir.Module{SourcePath: program.EntryPath, SourceFiles: make(map[string]string), StatementCount: program.StatementCount}
	for _, file := range program.Files {
		module.SourceFiles[file.FileName] = file.Source
	}
	hierarchy := buildClassHierarchy(program)
	shapes := map[string]ir.ObjectShape{}
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if statement.Kind == "class" && statement.Class != nil {
				shape := ir.ObjectShape{Name: statement.Class.Name, Span: toIRSpan(file.FileName, statement.Class.Span)}
				allFields := getInheritedFields(statement.Class.Name, hierarchy)
				for _, field := range allFields {
					val := ""
					if field.Initializer != nil {
						val = field.Initializer.Text
					} else if field.Type == "number" {
						val = "0"
					} else if field.Type == "bool" {
						val = "false"
					}
					shape.Fields = append(shape.Fields, ir.Field{Name: field.Name, Type: toIRType(field.Type), Value: val, Span: toIRSpan(file.FileName, field.Span)})
				}
				if len(shape.Fields) > 0 {
					shapes[shape.Name] = shape
					module.Shapes = append(module.Shapes, shape)
				}
			} else if statement.Kind == "enum" && statement.Enum != nil {
				shape := ir.ObjectShape{Name: statement.Enum.Name, Span: toIRSpan(file.FileName, statement.Enum.Span)}
				for _, member := range statement.Enum.Members {
					typ := ir.TypeNumber
					if member.Initializer != nil && member.Initializer.Kind == "string" {
						typ = ir.TypeString
					}
					shape.Fields = append(shape.Fields, ir.Field{
						Name:  member.Name,
						Type:  typ,
						Value: member.Value,
						Span:  toIRSpan(file.FileName, member.Span),
					})
				}
				shapes[shape.Name] = shape
				module.Shapes = append(module.Shapes, shape)
			}
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
			if statement.Kind == "module" || statement.Kind == "enum" || statement.Kind == "interface" || statement.Kind == "type_alias" {
				continue
			}
			if statement.Kind == "class" && statement.Class != nil {
				// Lower constructor if present
				if statement.Class.Constructor != nil {
					ctorMangled := statement.Class.Name + "_constructor"
					ctorStmt := typescriptgo.SyntaxStatement{
						Span: statement.Class.Constructor.Span,
						Kind: "function",
						Name: ctorMangled,
						Type: "void",
						Parameters: append([]typescriptgo.SyntaxParameter{
							{Name: "this", Type: "object:" + statement.Class.Name},
						}, statement.Class.Constructor.Parameters...),
						Body: statement.Class.Constructor.Body,
					}
					function, err := lowerFunction(file.FileName, ctorStmt, shapes, signatures)
					if err != nil {
						return ir.Module{}, fmt.Errorf("lower class constructor %q: %w", ctorMangled, sourceError(file.FileName, statement.Class.Constructor.Span, err))
					}
					module.Functions = append(module.Functions, function)
				}

				// Lower methods, static methods, getters, setters
				allMethods := getInheritedMethods(statement.Class.Name, hierarchy)
				for _, method := range allMethods {
					if method.IsAbstract || len(method.Body) == 0 {
						continue
					}
					var mangled string
					var params []typescriptgo.SyntaxParameter
					retType := method.Type
					if method.IsStatic {
						mangled = statement.Class.Name + "_" + method.Name
						params = method.Parameters
					} else if method.Kind == "get" {
						mangled = statement.Class.Name + "_get_" + method.Name
						params = []typescriptgo.SyntaxParameter{{Name: "this", Type: "object:" + statement.Class.Name}}
					} else if method.Kind == "set" {
						mangled = statement.Class.Name + "_set_" + method.Name
						params = append([]typescriptgo.SyntaxParameter{{Name: "this", Type: "object:" + statement.Class.Name}}, method.Parameters...)
						retType = "void"
					} else {
						mangled = statement.Class.Name + "_" + method.Name
						params = append([]typescriptgo.SyntaxParameter{{Name: "this", Type: "object:" + statement.Class.Name}}, method.Parameters...)
					}
					methodStmt := typescriptgo.SyntaxStatement{
						Span:       method.Span,
						Kind:       "function",
						Name:       mangled,
						Type:       retType,
						Parameters: params,
						Body:       method.Body,
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
	module.Shapes = nil
	for _, shape := range shapes {
		module.Shapes = append(module.Shapes, shape)
	}
	sort.Slice(module.Shapes, func(i, j int) bool {
		return module.Shapes[i].Name < module.Shapes[j].Name
	})
	main.Body = append(main.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid})
	module.Functions = append([]ir.Function{main}, module.Functions...)
	if err := module.Verify(); err != nil {
		return ir.Module{}, err
	}
	return module, nil
}
