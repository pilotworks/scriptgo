// Package lowering converts the checked frontend model into backend-independent IR.
package lowering

import (
	"fmt"
	"sort"
	"sync"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

var (
	lowerMu        sync.Mutex
	topLevelVars   = map[string]typescriptgo.SyntaxStatement{}
	inProgressVars = map[string]bool{}
)

// Options specifies optional flags for the lowering phase.
type Options struct {
	WarnRuntimeCasts bool
}

// Lower lowers the currently supported synchronous TypeScript subset.
func Lower(program frontend.Program) (ir.Module, error) {
	return LowerWithOptions(program, Options{WarnRuntimeCasts: WarnRuntimeCasts})
}

// LowerWithOptions lowers the program using custom options.
func LowerWithOptions(program frontend.Program, options Options) (ir.Module, error) {
	lowerMu.Lock()
	defer lowerMu.Unlock()

	prevWarn := WarnRuntimeCasts
	WarnRuntimeCasts = options.WarnRuntimeCasts || prevWarn
	defer func() {
		WarnRuntimeCasts = prevWarn
	}()
	extraFunctions = nil
	closureCounter = 0
	topLevelVars = map[string]typescriptgo.SyntaxStatement{}
	inProgressVars = map[string]bool{}
	ClearDiagnostics()
	var err error
	program, err = SpecializeGenerics(program)
	if err != nil {
		return ir.Module{}, err
	}
	if err := validateSubsetLocked(program); err != nil {
		return ir.Module{}, err
	}
	module := ir.Module{SourcePath: program.EntryPath, SourceFiles: make(map[string]string), StatementCount: program.StatementCount}
	typeAliasesIndex = map[string]string{}
	for _, file := range program.Files {
		module.SourceFiles[file.FileName] = file.Source
		for _, statement := range file.Syntax.Statements {
			if statement.Kind == "variable" && statement.Expression != nil {
				topLevelVars[statement.Name] = statement
			}
			if statement.Kind == "type_alias" && statement.Name != "" && statement.Type != "" {
				typeAliasesIndex[statement.Name] = statement.Type
			}
			if statement.Kind == "enum" && statement.Enum != nil {
				isStringEnum := false
				for _, m := range statement.Enum.Members {
					if m.Initializer != nil && m.Initializer.Kind == "string" {
						isStringEnum = true
						break
					}
				}
				if isStringEnum {
					typeAliasesIndex[statement.Enum.Name] = "string"
				} else {
					typeAliasesIndex[statement.Enum.Name] = "number"
				}
			}
		}
	}
	hierarchy := buildClassHierarchy(program)
	shapes := map[string]ir.ObjectShape{}
	builtinShapes := []ir.ObjectShape{
		{Name: "Error", Fields: []ir.Field{{Name: "message", Type: ir.TypeString}, {Name: "name", Type: ir.TypeString}}},
		{Name: "TypeError", Fields: []ir.Field{{Name: "message", Type: ir.TypeString}, {Name: "name", Type: ir.TypeString}}},
		{Name: "RangeError", Fields: []ir.Field{{Name: "message", Type: ir.TypeString}, {Name: "name", Type: ir.TypeString}}},
		{Name: "SyntaxError", Fields: []ir.Field{{Name: "message", Type: ir.TypeString}, {Name: "name", Type: ir.TypeString}}},
		{Name: "Date", Fields: []ir.Field{{Name: "time", Type: ir.TypeNumber}}},
		{Name: "RegExp", Fields: []ir.Field{{Name: "source", Type: ir.TypeString}, {Name: "flags", Type: ir.TypeString}}},
	}
	for _, s := range builtinShapes {
		shapes[s.Name] = s
		module.Shapes = append(module.Shapes, s)
	}
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if (statement.Kind == "class" || statement.Kind == "interface" || statement.Kind == "type_alias") && statement.Class != nil {
				shape := ir.ObjectShape{Name: statement.Class.Name, Span: toIRSpan(file.FileName, statement.Class.Span)}
				allFields := getInheritedFields(statement.Class.Name, hierarchy)
				if len(allFields) == 0 {
					allFields = statement.Class.Fields
				}
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
			if fields, ok := tupleFields(statement.Type); ok {
				shapeName := anonymousShapeName(fields)
				if _, exists := shapes[shapeName]; !exists {
					shape := ir.ObjectShape{Name: shapeName, Fields: fields}
					shapes[shapeName] = shape
					module.Shapes = append(module.Shapes, shape)
				}
				if statement.Name != "" {
					shapes[statement.Name] = ir.ObjectShape{Name: statement.Name, Fields: fields}
				}
			}
		}
	}

	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if statement.IsGenerator || statement.Kind == "generator_function" || statement.Kind == "async_generator_function" {
				RegisterGeneratorStatement(statement.Name, statement)
			}
		}
	}

	main := ir.Function{Name: "main", ReturnType: ir.TypeVoid}
	env := map[string]ir.Type{}
	signatures := buildFunctionIndex(program)
	counter := 0
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if statement.IsGenerator || statement.Kind == "generator_function" || statement.Kind == "async_generator_function" {
				factoryFn, extraFns, newShapes, err := lowerGeneratorFunction(file.FileName, statement, shapes, signatures)
				if err != nil {
					return ir.Module{}, fmt.Errorf("lower generator %q: %w", statement.Name, sourceError(file.FileName, statement.Span, err))
				}
				module.Functions = append(module.Functions, factoryFn)
				module.Functions = append(module.Functions, extraFns...)
				module.Shapes = append(module.Shapes, newShapes...)
				continue
			}
			if statement.Kind == "function" || statement.Kind == "async_function" {
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
				var fieldInits []typescriptgo.SyntaxStatement
				for _, f := range statement.Class.Fields {
					if !f.IsStatic && f.Initializer != nil {
						fieldInits = append(fieldInits, typescriptgo.SyntaxStatement{
							Span: f.Span,
							Kind: "field_set",
							Name: f.Name,
							Left: &typescriptgo.SyntaxExpression{
								Span: f.Span,
								Kind: "identifier",
								Text: "this",
							},
							Expression: f.Initializer,
						})
					}
				}

				// Lower constructor if present
				if statement.Class.Constructor != nil {
					ctorMangled := statement.Class.Name + "_constructor"
					var ctorBody []typescriptgo.SyntaxStatement
					if len(statement.Class.Constructor.Body) > 0 && statement.Class.Constructor.Body[0].Expression != nil && statement.Class.Constructor.Body[0].Expression.Kind == "call" && statement.Class.Constructor.Body[0].Expression.Left != nil && statement.Class.Constructor.Body[0].Expression.Left.Text == "super" {
						ctorBody = append(ctorBody, statement.Class.Constructor.Body[0])
						ctorBody = append(ctorBody, fieldInits...)
						ctorBody = append(ctorBody, statement.Class.Constructor.Body[1:]...)
					} else {
						ctorBody = append(ctorBody, fieldInits...)
						ctorBody = append(ctorBody, statement.Class.Constructor.Body...)
					}
					ctorStmt := typescriptgo.SyntaxStatement{
						Span: statement.Class.Constructor.Span,
						Kind: "function",
						Name: ctorMangled,
						Type: "void",
						Parameters: append([]typescriptgo.SyntaxParameter{
							{Name: "this", Type: "object:" + statement.Class.Name},
						}, statement.Class.Constructor.Parameters...),
						Body: ctorBody,
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
						mangled = statement.Class.Name + "_static_" + method.Name
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

				// Lower static field initializers and static blocks in class definition order
				for _, f := range statement.Class.Fields {
					if f.IsStatic && f.Initializer != nil {
						staticVar := statement.Class.Name + "_" + f.Name
						_, valType, err := lowerExpression(file.FileName, f.Initializer, staticVar, &main, env, &counter, shapes, signatures)
						if err == nil {
							env[staticVar] = valType
						}
					}
				}
				for _, block := range statement.Class.StaticBlocks {
					for _, stmt := range block {
						if err := lowerStatement(file.FileName, stmt, &main, env, &counter, shapes, signatures); err != nil {
							return ir.Module{}, fmt.Errorf("lower class %s static block: %w", statement.Class.Name, sourceError(file.FileName, stmt.Span, err))
						}
					}
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
	module.Functions = append(module.Functions, extraFunctions...)
	if err := module.Verify(); err != nil {
		return ir.Module{}, err
	}
	return module, nil
}
