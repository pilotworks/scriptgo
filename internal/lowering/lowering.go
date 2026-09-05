// Package lowering converts the checked frontend model into backend-independent IR.
package lowering

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
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
	checkedProgram := program
	program = runtimeProgram(checkedProgram)
	extraFunctions = nil
	closureCounter = 0
	topLevelVars = map[string]typescriptgo.SyntaxStatement{}
	inProgressVars = map[string]bool{}
	anonymousShapes = make(map[string]ir.ObjectShape)
	registeredShapes = nil
	generatorASTIndex = map[string]typescriptgo.SyntaxStatement{}
	defaultParamsIndex = map[string]map[int]*typescriptgo.SyntaxExpression{}
	restParamsIndex = map[string]bool{}
	activeReturnFinallyStack = nil
	activeThrowFinallyStack = nil
	loopFinallyScopeStack = nil
	usingScopeStack = nil
	clearMetadataRegistry()
	ClearDiagnostics()
	var err error
	program, err = SpecializeGenerics(program)
	if err != nil {
		return ir.Module{}, err
	}
	if err := validateSubsetLocked(program); err != nil {
		return ir.Module{}, err
	}
	initializeClassIdentities(program)
	initializeFunctionIdentities(program)
	module := ir.Module{SourcePath: program.EntryPath, SourceFiles: make(map[string]string), StatementCount: program.StatementCount}
	typeAliasesIndex = map[string]string{}
	for _, file := range program.Files {
		module.SourceFiles[file.FileName] = file.Source
		var collectTopLevel func(s typescriptgo.SyntaxStatement)
		collectTopLevel = func(s typescriptgo.SyntaxStatement) {
			if (s.Kind == "variable" || s.Kind == "using" || s.Kind == "await_using") && s.Name != "" {
				if s.Expression != nil || topLevelVars[s.Name].Expression == nil {
					topLevelVars[s.Name] = s
				}
			} else if s.Kind == "block" || s.Kind == "namespace" {
				for _, sub := range s.Body {
					collectTopLevel(sub)
				}
			}
		}
		for _, statement := range file.Syntax.Statements {
			collectTopLevel(statement)
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
		{Name: "Error", Fields: []ir.Field{{Name: "message", Type: ir.TypeString}, {Name: "name", Type: ir.TypeString}, {Name: "stack", Type: ir.TypeString}, {Name: "cause", Type: ir.TypeString}}},
		{Name: "TypeError", Fields: []ir.Field{{Name: "message", Type: ir.TypeString}, {Name: "name", Type: ir.TypeString}, {Name: "stack", Type: ir.TypeString}, {Name: "cause", Type: ir.TypeString}}},
		{Name: "RangeError", Fields: []ir.Field{{Name: "message", Type: ir.TypeString}, {Name: "name", Type: ir.TypeString}, {Name: "stack", Type: ir.TypeString}, {Name: "cause", Type: ir.TypeString}}},
		{Name: "ReferenceError", Fields: []ir.Field{{Name: "message", Type: ir.TypeString}, {Name: "name", Type: ir.TypeString}, {Name: "stack", Type: ir.TypeString}, {Name: "cause", Type: ir.TypeString}}},
		{Name: "SyntaxError", Fields: []ir.Field{{Name: "message", Type: ir.TypeString}, {Name: "name", Type: ir.TypeString}, {Name: "stack", Type: ir.TypeString}, {Name: "cause", Type: ir.TypeString}}},
		{Name: "URIError", Fields: []ir.Field{{Name: "message", Type: ir.TypeString}, {Name: "name", Type: ir.TypeString}, {Name: "stack", Type: ir.TypeString}, {Name: "cause", Type: ir.TypeString}}},
		{Name: "EvalError", Fields: []ir.Field{{Name: "message", Type: ir.TypeString}, {Name: "name", Type: ir.TypeString}, {Name: "stack", Type: ir.TypeString}, {Name: "cause", Type: ir.TypeString}}},
		{Name: "Date", Fields: []ir.Field{{Name: "time", Type: ir.TypeNumber}}},
		{Name: "RegExp", Fields: []ir.Field{{Name: "source", Type: ir.TypeString}, {Name: "flags", Type: ir.TypeString}, {Name: "lastIndex", Type: ir.TypeNumber}}},
		{Name: "ResponseInit", Fields: []ir.Field{{Name: "status", Type: ir.TypeNumber}, {Name: "statusText", Type: ir.TypeString}, {Name: "headers", Type: ir.TypeObject}}},
		{Name: "RequestInit", Fields: []ir.Field{{Name: "method", Type: ir.TypeString}, {Name: "headers", Type: ir.TypeObject}, {Name: "body", Type: ir.TypeUnknown}, {Name: "mode", Type: ir.TypeString}, {Name: "credentials", Type: ir.TypeString}, {Name: "cache", Type: ir.TypeString}, {Name: "redirect", Type: ir.TypeString}, {Name: "referrer", Type: ir.TypeString}}},
		{Name: "TextDecoderOptions", Fields: []ir.Field{{Name: "fatal", Type: ir.TypeBool}, {Name: "ignoreBOM", Type: ir.TypeBool}}},
		{Name: "TextDecodeOptions", Fields: []ir.Field{{Name: "stream", Type: ir.TypeBool}}},
		{Name: "TextEncoderEncodeIntoResult", Fields: []ir.Field{{Name: "read", Type: ir.TypeNumber}, {Name: "written", Type: ir.TypeNumber}}},
		{Name: "IteratorResult", Fields: []ir.Field{{Name: "done", Type: ir.TypeBool}, {Name: "value", Type: ir.TypeUnknown}}},
		{Name: "TypedPropertyDescriptor", Fields: []ir.Field{{Name: "enumerable", Type: ir.TypeBool}, {Name: "configurable", Type: ir.TypeBool}, {Name: "writable", Type: ir.TypeBool}, {Name: "value", Type: ir.TypeUnknown}, {Name: "get", Type: ir.TypeClosure}, {Name: "set", Type: ir.TypeClosure}}},
		{Name: "PropertyDescriptor", Fields: []ir.Field{{Name: "enumerable", Type: ir.TypeBool}, {Name: "configurable", Type: ir.TypeBool}, {Name: "writable", Type: ir.TypeBool}, {Name: "value", Type: ir.TypeUnknown}, {Name: "get", Type: ir.TypeClosure}, {Name: "set", Type: ir.TypeClosure}}},
	}
	for _, s := range builtinShapes {
		shapes[s.Name] = s
		module.Shapes = append(module.Shapes, s)
	}
	for _, file := range program.Files {
		var collectShapes func(fileName string, statement typescriptgo.SyntaxStatement)
		collectShapes = func(fileName string, statement typescriptgo.SyntaxStatement) {
			if (statement.Kind == "class" || statement.Kind == "interface" || statement.Kind == "type_alias") && statement.Class != nil {
				if statement.Kind == "type_alias" && len(statement.Class.Fields) == 0 {
					return
				}
				className := classIdentityForPath(fileName, statement.Class.Name)
				shape := ir.ObjectShape{Name: className, Span: toIRSpan(fileName, statement.Class.Span)}
				allFields := getInheritedFields(className, hierarchy)
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
					fTypeStr := field.Type
					if fTypeStr == "" {
						fTypeStr = field.InferredType
					}
					shape.Fields = append(shape.Fields, ir.Field{Name: field.Name, Type: toIRTypeForPath(fileName, fTypeStr), Value: val, Optional: field.Optional, Span: toIRSpan(fileName, field.Span)})
				}
				if statement.Kind == "interface" {
					if _, exists := shapes[shape.Name]; exists {
						shapes[shape.Name] = shape
						for idx, s := range module.Shapes {
							if s.Name == shape.Name {
								module.Shapes[idx] = shape
								break
							}
						}
						return
					}
				}
				shapes[shape.Name] = shape
				baseName := className
				if idx := strings.Index(baseName, "<"); idx != -1 {
					baseName = baseName[:idx]
					shapes[baseName] = shape
				}
				module.Shapes = append(module.Shapes, shape)
			} else if statement.Kind == "enum" && statement.Enum != nil {
				shape := ir.ObjectShape{Name: statement.Enum.Name, Span: toIRSpan(fileName, statement.Enum.Span)}
				for _, member := range statement.Enum.Members {
					typ := ir.TypeNumber
					if member.Initializer != nil && member.Initializer.Kind == "string" {
						typ = ir.TypeString
					}
					shape.Fields = append(shape.Fields, ir.Field{
						Name:  member.Name,
						Type:  typ,
						Value: member.Value,
						Span:  toIRSpan(fileName, member.Span),
					})
				}
				shapes[shape.Name] = shape
				module.Shapes = append(module.Shapes, shape)
			} else if statement.Kind == "namespace" || statement.Kind == "block" {
				for _, sub := range statement.Body {
					collectShapes(fileName, sub)
				}
			}
		}
		for _, statement := range file.Syntax.Statements {
			collectShapes(file.FileName, statement)
		}
	}
	for _, shape := range typeOnlyShapes(checkedProgram, program) {
		if _, exists := shapes[shape.Name]; exists {
			continue
		}
		shapes[shape.Name] = shape
		module.Shapes = append(module.Shapes, shape)
	}
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
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
			if fields, ok := anonymousObjectFields(statement.Type, nil); ok {
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
	var allUnionTypes []string
	addUnionType := func(t string) {
		t = strings.TrimSpace(t)
		if strings.Contains(t, "|") {
			allUnionTypes = append(allUnionTypes, t)
		}
	}
	for _, file := range program.Files {
		for _, stmt := range file.Syntax.Statements {
			addUnionType(stmt.Type)
			addUnionType(stmt.InferredType)
			for _, param := range stmt.Parameters {
				addUnionType(param.Type)
				addUnionType(param.InferredType)
			}
			for _, sub := range stmt.Body {
				addUnionType(sub.Type)
				addUnionType(sub.InferredType)
				for _, param := range sub.Parameters {
					addUnionType(param.Type)
					addUnionType(param.InferredType)
				}
			}
		}
	}
	for _, unionTypeStr := range allUnionTypes {
		unionIR := toIRType(unionTypeStr)
		unionShapeName := strings.TrimPrefix(string(unionIR), "object:")
		if unionShape, ok := anonymousShapes[unionShapeName]; ok && len(unionShape.Fields) > 0 {
			for _, member := range splitTopLevelUnion(unionTypeStr) {
				cleanM := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(member), "object:"))
				if _, isVariant := shapes[cleanM]; isVariant {
					shapes[cleanM] = unionShape
				}
				var memberTypeStr string
				if aliased, ok := typeAliasesIndex[cleanM]; ok {
					memberTypeStr = aliased
				} else {
					memberTypeStr = cleanM
				}
				if f, ok := anonymousObjectFields(memberTypeStr, nil); ok {
					anonName := anonymousShapeName(f)
					shapes[anonName] = unionShape
					anonymousShapes[anonName] = unionShape
				}
			}
		}
	}
	SetRegisteredShapes(shapes)

	for _, shape := range anonymousShapes {
		if _, exists := shapes[shape.Name]; !exists {
			shapes[shape.Name] = shape
			module.Shapes = append(module.Shapes, shape)
		}
	}

	for clsName, meta := range hierarchy {
		for fName, field := range meta.Statics {
			globalName := clsName + "_" + fName
			globalType := toIRTypeForPath(meta.FileName, field.Type)
			if globalType == "" {
				globalType = ir.TypeNumber
			}
			globalVal := ""
			if field.Initializer != nil {
				globalVal = field.Initializer.Text
			}
			module.Globals = append(module.Globals, ir.Global{
				Name:  globalName,
				Type:  globalType,
				Value: globalVal,
			})
		}
	}

	seenGlobals := map[string]bool{}
	for _, file := range program.Files {
		var collectGlobals func(statement typescriptgo.SyntaxStatement)
		collectGlobals = func(statement typescriptgo.SyntaxStatement) {
			if statement.Kind == "variable" && statement.Name != "" {
				if seenGlobals[statement.Name] {
					return
				}
				seenGlobals[statement.Name] = true
				vType := statement.Type
				if vType == "" && statement.InferredType != "" {
					vType = statement.InferredType
				}
				typ := toIRType(vType)
				if statement.Expression != nil {
					if (statement.Type == "" && statement.InferredType == "") || typ == "" {
						if inferred := inferExprType(statement.Expression, nil, nil); inferred != "" {
							if ityp := toIRType(inferred); ityp != "" {
								typ = ityp
							}
						}
					}
					if typ == "" || typ == ir.TypeNumber {
						switch statement.Expression.Kind {
						case "arrow_function", "function":
							typ = ir.TypeClosure
						case "new":
							if statement.Expression.Left != nil {
								typ = ir.Type("object:" + statement.Expression.Left.Text)
							}
						case "call":
							if statement.Expression.Left != nil {
								calleeName := statement.Expression.Left.Text
								if calleeVar, ok := topLevelVars[calleeName]; ok {
									if isReturningClosure(calleeVar) {
										typ = ir.TypeClosure
									} else if calleeVar.Type != "" {
										t := toIRType(calleeVar.Type)
										if t == ir.TypeClosure || strings.HasPrefix(string(t), "object:") {
											typ = t
										}
									}
								}
							}
						}
					}
					if typ == "" || (typ == ir.TypePointer && strings.Contains(vType, "|")) {
						switch statement.Expression.Kind {
						case "number", "literal":
							if _, err := strconv.ParseFloat(statement.Expression.Text, 64); err == nil {
								typ = ir.TypeNumber
							}
						case "string":
							typ = ir.TypeString
						case "bool":
							typ = ir.TypeBool
						case "bigint":
							typ = ir.TypeBigInt
						}
					}
				}
				if typ == "" {
					typ = ir.TypeNumber
				}
				module.Globals = append(module.Globals, ir.Global{
					Name: statement.Name,
					Type: typ,
				})
			} else if statement.Kind == "namespace" {
				for _, sub := range statement.Body {
					collectGlobals(sub)
				}
			}
		}
		for _, statement := range file.Syntax.Statements {
			collectGlobals(statement)
		}
	}

	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if statement.IsGenerator || statement.Kind == "generator_function" || statement.Kind == "async_generator_function" {
				RegisterGeneratorStatement(functionIdentityForPath(file.FileName, statement.Name), statement)
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
				for _, fn := range extraFns {
					signatures[fn.Name] = fn
				}
				continue
			}
		}
	}

	implementedFunctions := map[string]bool{}
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if (statement.Kind == "function" || statement.Kind == "async_function") && len(statement.Body) > 0 {
				implementedFunctions[functionIdentityForPath(file.FileName, statement.Name)] = true
			}
		}
	}

	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if statement.Kind == "declare_function" {
				if implementedFunctions[functionIdentityForPath(file.FileName, statement.Name)] {
					continue
				}
				retType := toIRType(statement.Type)
				if retType == "" && statement.InferredType != "" {
					retType = toIRType(statement.InferredType)
				}
				if retType == "" {
					retType = ir.TypeVoid
				}
				var params []ir.Parameter
				for _, p := range statement.Parameters {
					pType := p.Type
					if pType == "" && p.InferredType != "" {
						pType = p.InferredType
					}
					params = append(params, ir.Parameter{Name: p.Name, Type: toIRType(pType)})
				}
				module.Externs = append(module.Externs, ir.ExternFunction{
					Name:       functionIdentityForPath(file.FileName, statement.Name),
					Span:       toIRSpan(file.FileName, statement.Span),
					Parameters: params,
					ReturnType: retType,
				})
				continue
			}
			if statement.Kind == "function" || statement.Kind == "async_function" {
				if len(statement.TypeParameters) > 0 {
					continue
				}
				if len(statement.Body) == 0 && implementedFunctions[functionIdentityForPath(file.FileName, statement.Name)] {
					continue
				}
				fnCopy := statement
				fnCopy.Name = functionIdentityForPath(file.FileName, statement.Name)
				function, err := lowerFunction(file.FileName, fnCopy, shapes, signatures)
				if err != nil {
					return ir.Module{}, fmt.Errorf("lower function %q: %w", statement.Name, sourceError(file.FileName, statement.Span, err))
				}
				module.Functions = append(module.Functions, function)
				signatures[statement.Name] = function
				signatures[function.Name] = function
				continue
			}
			lowerClassStatement := func(fileName string, statement typescriptgo.SyntaxStatement) error {
				if statement.Class == nil || len(statement.Class.TypeParameters) > 0 {
					return nil
				}
				className := classIdentityForPath(fileName, statement.Class.Name)
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
					ctorMangled := className + "_constructor"
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
							{Name: "this", Type: "object:" + className},
						}, statement.Class.Constructor.Parameters...),
						Body: ctorBody,
					}
					function, err := lowerFunction(fileName, ctorStmt, shapes, signatures)
					if err != nil {
						return fmt.Errorf("lower class constructor %q: %w", ctorMangled, sourceError(fileName, statement.Class.Constructor.Span, err))
					}
					module.Functions = append(module.Functions, function)
					signatures[ctorMangled] = function
				}

				// Lower methods, static methods, getters, setters
				allMethods := getInheritedMethods(className, hierarchy)
				for _, method := range allMethods {
					if method.IsAbstract || method.Body == nil {
						continue
					}
					var mangled string
					var params []typescriptgo.SyntaxParameter
					retType := method.Type
					// TypeScript's polymorphic `this` return type is the concrete
					// class type at each method boundary. Keeping it as the literal
					// `this` type makes the next chained call lose its receiver class.
					if retType == "this" || retType == "object:this" {
						retType = "object:" + className
					}
					var cleanParams []typescriptgo.SyntaxParameter
					for _, p := range method.Parameters {
						if p.Name != "this" {
							cleanParams = append(cleanParams, p)
						}
					}
					if method.IsStatic {
						mangled = className + "_static_" + method.Name
						params = cleanParams
					} else if method.Kind == "get" {
						mangled = className + "_get_" + method.Name
						params = []typescriptgo.SyntaxParameter{{Name: "this", Type: "object:" + className}}
					} else if method.Kind == "set" {
						mangled = className + "_set_" + method.Name
						params = append([]typescriptgo.SyntaxParameter{{Name: "this", Type: "object:" + className}}, cleanParams...)
						retType = "void"
					} else {
						mangled = methodImplementationName(className, method.Name)
						params = append([]typescriptgo.SyntaxParameter{{Name: "this", Type: "object:" + className}}, cleanParams...)
					}
					methodStmt := typescriptgo.SyntaxStatement{
						Span:       method.Span,
						Kind:       "function",
						IsAsync:    method.IsAsync,
						Name:       mangled,
						Type:       retType,
						Parameters: params,
						Body:       method.Body,
					}
					function, err := lowerFunction(fileName, methodStmt, shapes, signatures)
					if err != nil {
						fmt.Printf("DEBUG method error: %s: %v\n", mangled, err)
						return fmt.Errorf("lower class method %q: %w", mangled, sourceError(fileName, method.Span, err))
					}
					module.Functions = append(module.Functions, function)
					signatures[mangled] = function
					if method.IsStatic {
						signatures[className+"."+method.Name] = function
					}
				}

				// Lower static field initializers and static blocks in class definition order
				if len(statement.Class.StaticElements) > 0 {
					for _, elem := range statement.Class.StaticElements {
						switch elem.Kind {
						case typescriptgo.StaticElementField:
							f := elem.Field
							if f != nil && f.IsStatic && f.Initializer != nil {
								staticVar := className + "_" + f.Name
								_, valType, err := lowerExpression(fileName, f.Initializer, staticVar, &main, env, &counter, shapes, signatures)
								if err == nil {
									env[staticVar] = valType
								}
							}
						case typescriptgo.StaticElementBlock:
							for _, stmt := range elem.Statements {
								if err := lowerStatement(fileName, stmt, &main, env, &counter, shapes, signatures); err != nil {
									return fmt.Errorf("lower class %s static block: %w", statement.Class.Name, sourceError(fileName, stmt.Span, err))
								}
							}
						}
					}
				} else {
					for _, f := range statement.Class.Fields {
						if f.IsStatic && f.Initializer != nil {
							staticVar := className + "_" + f.Name
							_, valType, err := lowerExpression(fileName, f.Initializer, staticVar, &main, env, &counter, shapes, signatures)
							if err == nil {
								env[staticVar] = valType
							}
						}
					}
					for _, block := range statement.Class.StaticBlocks {
						for _, stmt := range block {
							if err := lowerStatement(fileName, stmt, &main, env, &counter, shapes, signatures); err != nil {
								return fmt.Errorf("lower class %s static block: %w", statement.Class.Name, sourceError(fileName, stmt.Span, err))
							}
						}
					}
				}
				if err := lowerClassDecorators(fileName, statement.Class, &main, env, &counter, shapes, signatures); err != nil {
					return fmt.Errorf("lower class %s decorators: %w", statement.Class.Name, err)
				}
				return nil
			}

			if statement.Kind == "namespace" {
				for _, subStmt := range statement.Body {
					if subStmt.Kind == "declare_function" || subStmt.Kind == "interface" || subStmt.Kind == "type_alias" || subStmt.Kind == "module" || subStmt.Kind == "enum" {
						continue
					}
					if subStmt.Kind == "function" || subStmt.Kind == "async_function" {
						if len(subStmt.TypeParameters) > 0 {
							continue
						}
						fnCopy := subStmt
						fnCopy.Name = statement.Name + "." + subStmt.Name
						function, err := lowerFunction(file.FileName, fnCopy, shapes, signatures)
						if err != nil {
							return ir.Module{}, fmt.Errorf("lower namespace function %q: %w", fnCopy.Name, sourceError(file.FileName, fnCopy.Span, err))
						}
						module.Functions = append(module.Functions, function)
					} else if subStmt.Kind == "class" && subStmt.Class != nil {
						if err := lowerClassStatement(file.FileName, subStmt); err != nil {
							return ir.Module{}, err
						}
					} else {
						if err := lowerStatement(file.FileName, subStmt, &main, env, &counter, shapes, signatures); err != nil {
							return ir.Module{}, fmt.Errorf("lower namespace statement: %w", sourceError(file.FileName, subStmt.Span, err))
						}
					}
				}
				continue
			}
			if statement.Kind == "module" || statement.Kind == "enum" || statement.Kind == "interface" || statement.Kind == "type_alias" || statement.Kind == "generator_function" || statement.Kind == "async_generator_function" || statement.IsGenerator {
				continue
			}
			if statement.Kind == "class" && statement.Class != nil {
				if err := lowerClassStatement(file.FileName, statement); err != nil {
					return ir.Module{}, err
				}
				continue
			}
			if err := lowerStatement(file.FileName, statement, &main, env, &counter, shapes, signatures); err != nil {
				return ir.Module{}, fmt.Errorf("lower %s: %w", statement.Kind, sourceError(file.FileName, statement.Span, err))
			}
		}
	}
	for name, s := range anonymousShapes {
		if _, exists := shapes[name]; !exists {
			shapes[name] = s
		}
	}
	module.Shapes = nil
	for _, shape := range shapes {
		module.Shapes = append(module.Shapes, shape)
	}
	sort.Slice(module.Shapes, func(i, j int) bool {
		return module.Shapes[i].Name < module.Shapes[j].Name
	})
	if len(collectNestedAwaits(main.Body)) > 0 {
		functions, ok, err := lowerTopLevelAsyncSequence(program.EntryPath, main, shapes, signatures)
		if err != nil {
			return ir.Module{}, fmt.Errorf("lower top-level await: %w", sourceError(program.EntryPath, typescriptgo.SourceSpan{}, err))
		}
		if ok {
			mainFunction := functions[len(functions)-1]
			stageFunctions := functions[:len(functions)-1]
			module.Functions = append(append([]ir.Function{mainFunction}, stageFunctions...), module.Functions...)
		} else {
			main.Body = append(main.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid})
			module.Functions = append([]ir.Function{main}, module.Functions...)
		}
	} else {
		main.Body = append(main.Body, ir.Instruction{Op: ir.OpReturn, Type: ir.TypeVoid})
		module.Functions = append([]ir.Function{main}, module.Functions...)
	}
	module.Functions = append(module.Functions, extraFunctions...)

	dispatchers := synthesizePolymorphicDispatchers(hierarchy, signatures)
	for _, df := range dispatchers {
		replaced := false
		for idx, f := range module.Functions {
			if f.Name == df.Name {
				module.Functions[idx] = df
				replaced = true
				break
			}
		}
		if !replaced {
			module.Functions = append(module.Functions, df)
		}
	}

	if err := module.Verify(); err != nil {
		return ir.Module{}, err
	}
	return module, nil
}
