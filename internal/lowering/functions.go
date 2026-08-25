package lowering

import (
	"path/filepath"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

var defaultParamsIndex = map[string]map[int]*typescriptgo.SyntaxExpression{}
var restParamsIndex = map[string]bool{}

// buildFunctionIndex collects function signatures and namespace import aliases
// from the checked module graph. Lowering does not need to know which module a
// namespace came from; the frontend has already resolved that edge.
func buildFunctionIndex(program frontend.Program) map[string]ir.Function {
	typeAliasesIndex = map[string]string{}
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
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
	defaultParamsIndex = map[string]map[int]*typescriptgo.SyntaxExpression{}
	restParamsIndex = map[string]bool{}
	index := map[string]ir.Function{}
	functionsByFile := map[string][]ir.Function{}
	for _, file := range program.Files {
		fileName := filepath.Clean(file.FileName)
		for _, statement := range file.Syntax.Statements {
			if statement.Kind == "declare_function" {
				retType := statement.Type
				if retType == "" && statement.InferredType != "" {
					retType = statement.InferredType
				}
				function := ir.Function{Name: statement.Name, ReturnType: toIRType(retType)}
				if function.ReturnType == "" {
					function.ReturnType = ir.TypeVoid
				}
				for _, parameter := range statement.Parameters {
					pType := parameter.Type
					if pType == "" && parameter.InferredType != "" {
						pType = parameter.InferredType
					}
					function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: toIRType(pType)})
				}
				index[function.Name] = function
				continue
			}
			if statement.Kind == "function" || statement.Kind == "generator_function" || statement.Kind == "async_function" || statement.Kind == "async_generator_function" || statement.IsGenerator || statement.IsAsync {
				retType := statement.Type
				if retType == "" && statement.InferredType != "" {
					retType = statement.InferredType
				}
				if statement.IsGenerator || statement.Kind == "generator_function" || statement.Kind == "async_generator_function" {
					retType = "object:Generator_" + statement.Name
				}
				fnName := statement.Name
				if fnName == "main" {
					fnName = "main$user"
				}
				function := ir.Function{Name: fnName, ReturnType: toIRType(retType)}
				if function.ReturnType == "" {
					function.ReturnType = ir.TypeVoid
				}
				for pIdx, parameter := range statement.Parameters {
					pType := parameter.Type
					if pType == "" && parameter.InferredType != "" {
						pType = parameter.InferredType
					}
					typ := toIRType(pType)
					if parameter.Rest {
						restParamsIndex[function.Name] = true
						if typ == "" || typ == ir.TypeUnknown {
							if pType == "number[]" {
								typ = ir.TypeNumberArray
							} else {
								typ = ir.TypeStringArray
							}
						}
					}
					if parameter.Initializer != nil {
						if defaultParamsIndex[function.Name] == nil {
							defaultParamsIndex[function.Name] = map[int]*typescriptgo.SyntaxExpression{}
						}
						defaultParamsIndex[function.Name][pIdx] = parameter.Initializer
					} else if parameter.Optional {
						if defaultParamsIndex[function.Name] == nil {
							defaultParamsIndex[function.Name] = map[int]*typescriptgo.SyntaxExpression{}
						}
						defaultParamsIndex[function.Name][pIdx] = &typescriptgo.SyntaxExpression{Kind: "undefined"}
					}
					function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: typ})
				}
				index[statement.Name] = function
				index[function.Name] = function
				if defaultParamsIndex[function.Name] != nil {
					defaultParamsIndex[statement.Name] = defaultParamsIndex[function.Name]
				}
				if restParamsIndex[function.Name] {
					restParamsIndex[statement.Name] = true
				}
				if file.BuiltinName != "" {
					index[file.BuiltinName+"."+function.Name] = function
					if defaultParamsIndex[function.Name] != nil {
						defaultParamsIndex[file.BuiltinName+"."+function.Name] = defaultParamsIndex[function.Name]
					}
					if restParamsIndex[function.Name] {
						restParamsIndex[file.BuiltinName+"."+function.Name] = true
					}
				}
				functionsByFile[fileName] = append(functionsByFile[fileName], function)
			} else if statement.Kind == "namespace" {
				for _, subStmt := range statement.Body {
					if subStmt.Kind == "function" || subStmt.Kind == "async_function" {
						retType := subStmt.Type
						if retType == "" && subStmt.InferredType != "" {
							retType = subStmt.InferredType
						}
						fullName := statement.Name + "." + subStmt.Name
						function := ir.Function{Name: fullName, ReturnType: toIRType(retType)}
						if function.ReturnType == "" {
							function.ReturnType = ir.TypeVoid
						}
						for _, parameter := range subStmt.Parameters {
							pType := parameter.Type
							if pType == "" && parameter.InferredType != "" {
								pType = parameter.InferredType
							}
							function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: toIRType(pType)})
						}
						index[fullName] = function
						functionsByFile[fileName] = append(functionsByFile[fileName], function)
					}
				}
			} else if statement.Kind == "class" && statement.Class != nil {
				// 1. Index constructor
				if statement.Class.Constructor != nil {
					ctorMangled := statement.Class.Name + "_constructor"
					ctorFn := ir.Function{Name: ctorMangled, ReturnType: ir.TypeVoid}
					ctorFn.Parameters = append(ctorFn.Parameters, ir.Parameter{Name: "this", Type: ir.Type("object:" + statement.Class.Name)})
					for pIdx, parameter := range statement.Class.Constructor.Parameters {
						typ := toIRType(parameter.Type)
						if parameter.Initializer != nil {
							if defaultParamsIndex[ctorMangled] == nil {
								defaultParamsIndex[ctorMangled] = map[int]*typescriptgo.SyntaxExpression{}
							}
							defaultParamsIndex[ctorMangled][pIdx+1] = parameter.Initializer
						} else if parameter.Optional {
							if defaultParamsIndex[ctorMangled] == nil {
								defaultParamsIndex[ctorMangled] = map[int]*typescriptgo.SyntaxExpression{}
							}
							defaultParamsIndex[ctorMangled][pIdx+1] = &typescriptgo.SyntaxExpression{Kind: "undefined"}
						}
						ctorFn.Parameters = append(ctorFn.Parameters, ir.Parameter{Name: parameter.Name, Type: typ})
					}
					index[ctorMangled] = ctorFn
					functionsByFile[fileName] = append(functionsByFile[fileName], ctorFn)
				}

				// 2. Index all methods (including inherited)
				allMethods := getInheritedMethods(statement.Class.Name, hierarchy)
				for _, method := range allMethods {
					var mangled string
					var function ir.Function
					if method.IsStatic {
						mangled = statement.Class.Name + "_static_" + method.Name
						function = ir.Function{Name: mangled, ReturnType: toIRType(method.Type)}
						if function.ReturnType == "" {
							function.ReturnType = ir.TypeVoid
						}
						for pIdx, parameter := range method.Parameters {
							typ := toIRType(parameter.Type)
							if parameter.Initializer != nil {
								if defaultParamsIndex[mangled] == nil {
									defaultParamsIndex[mangled] = map[int]*typescriptgo.SyntaxExpression{}
								}
								defaultParamsIndex[mangled][pIdx] = parameter.Initializer
							} else if parameter.Optional {
								if defaultParamsIndex[mangled] == nil {
									defaultParamsIndex[mangled] = map[int]*typescriptgo.SyntaxExpression{}
								}
								defaultParamsIndex[mangled][pIdx] = &typescriptgo.SyntaxExpression{Kind: "undefined"}
							}
							function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: typ})
						}
						index[mangled] = function
						index[statement.Class.Name+"."+method.Name] = function
					} else if method.Kind == "get" {
						mangled = statement.Class.Name + "_get_" + method.Name
						function = ir.Function{Name: mangled, ReturnType: toIRType(method.Type)}
						function.Parameters = append(function.Parameters, ir.Parameter{Name: "this", Type: ir.Type("object:" + statement.Class.Name)})
					} else if method.Kind == "set" {
						mangled = statement.Class.Name + "_set_" + method.Name
						function = ir.Function{Name: mangled, ReturnType: ir.TypeVoid}
						function.Parameters = append(function.Parameters, ir.Parameter{Name: "this", Type: ir.Type("object:" + statement.Class.Name)})
						if len(method.Parameters) > 0 {
							function.Parameters = append(function.Parameters, ir.Parameter{Name: method.Parameters[0].Name, Type: toIRType(method.Parameters[0].Type)})
						}
					} else {
						mangled = statement.Class.Name + "_" + method.Name
						retType := toIRType(method.Type)
						if method.Type == "this" || retType == "this" || retType == "object:this" {
							retType = ir.Type("object:" + statement.Class.Name)
						}
						function = ir.Function{Name: mangled, ReturnType: retType}
						if function.ReturnType == "" {
							function.ReturnType = ir.TypeVoid
						}
						function.Parameters = append(function.Parameters, ir.Parameter{Name: "this", Type: ir.Type("object:" + statement.Class.Name)})
						for pIdx, parameter := range method.Parameters {
							typ := toIRType(parameter.Type)
							if parameter.Initializer != nil {
								if defaultParamsIndex[mangled] == nil {
									defaultParamsIndex[mangled] = map[int]*typescriptgo.SyntaxExpression{}
								}
								defaultParamsIndex[mangled][pIdx+1] = parameter.Initializer
							} else if parameter.Optional {
								if defaultParamsIndex[mangled] == nil {
									defaultParamsIndex[mangled] = map[int]*typescriptgo.SyntaxExpression{}
								}
								defaultParamsIndex[mangled][pIdx+1] = &typescriptgo.SyntaxExpression{Kind: "undefined"}
							}
							if parameter.Rest {
								restParamsIndex[mangled] = true
							}
							function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: typ})
						}
					}
					if !method.IsAbstract {
						index[mangled] = function
						functionsByFile[fileName] = append(functionsByFile[fileName], function)
					}
				}
			}
		}
	}
	for _, file := range program.Files {
		for _, reference := range file.Imports {
			if reference.LocalName == "" || reference.ResolvedFileName == "" {
				continue
			}
			for _, function := range functionsByFile[filepath.Clean(reference.ResolvedFileName)] {
				index[reference.LocalName+"."+function.Name] = function
				if defaultParamsIndex[function.Name] != nil {
					defaultParamsIndex[reference.LocalName+"."+function.Name] = defaultParamsIndex[function.Name]
				}
			}
		}
	}
	dispatchers := synthesizePolymorphicDispatchers(hierarchy, index)
	for _, df := range dispatchers {
		index[df.Name] = df
	}
	return index
}
