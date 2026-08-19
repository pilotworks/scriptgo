package lowering

import (
	"path/filepath"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

var defaultParamsIndex = map[string]map[int]*typescriptgo.SyntaxExpression{}

// buildFunctionIndex collects function signatures and namespace import aliases
// from the checked module graph. Lowering does not need to know which module a
// namespace came from; the frontend has already resolved that edge.
func buildFunctionIndex(program frontend.Program) map[string]ir.Function {
	defaultParamsIndex = map[string]map[int]*typescriptgo.SyntaxExpression{}
	index := map[string]ir.Function{}
	functionsByFile := map[string][]ir.Function{}
	for _, file := range program.Files {
		fileName := filepath.Clean(file.FileName)
		for _, statement := range file.Syntax.Statements {
			if statement.Kind == "function" {
				function := ir.Function{Name: statement.Name, ReturnType: toIRType(statement.Type)}
				if function.ReturnType == "" {
					function.ReturnType = ir.TypeVoid
				}
				for pIdx, parameter := range statement.Parameters {
					typ := toIRType(parameter.Type)
					if parameter.Rest {
						if parameter.Type == "number[]" {
							typ = ir.TypeNumberArray
						} else {
							typ = ir.TypeStringArray
						}
					}
					if parameter.Initializer != nil {
						if defaultParamsIndex[function.Name] == nil {
							defaultParamsIndex[function.Name] = map[int]*typescriptgo.SyntaxExpression{}
						}
						defaultParamsIndex[function.Name][pIdx] = parameter.Initializer
					}
					function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: typ})
				}
				index[function.Name] = function
				if file.BuiltinName != "" {
					index[file.BuiltinName+"."+function.Name] = function
					if defaultParamsIndex[function.Name] != nil {
						defaultParamsIndex[file.BuiltinName+"."+function.Name] = defaultParamsIndex[function.Name]
					}
				}
				functionsByFile[fileName] = append(functionsByFile[fileName], function)
			} else if statement.Kind == "class" && statement.Class != nil {
				for _, method := range statement.Class.Methods {
					mangled := statement.Class.Name + "_" + method.Name
					function := ir.Function{Name: mangled, ReturnType: toIRType(method.Type)}
					if function.ReturnType == "" {
						function.ReturnType = ir.TypeVoid
					}
					function.Parameters = append(function.Parameters, ir.Parameter{Name: "this", Type: ir.Type("object:" + statement.Class.Name)})
					for pIdx, parameter := range method.Parameters {
						typ := toIRType(parameter.Type)
						if parameter.Rest {
							if parameter.Type == "number[]" {
								typ = ir.TypeNumberArray
							} else {
								typ = ir.TypeStringArray
							}
						}
						if parameter.Initializer != nil {
							if defaultParamsIndex[mangled] == nil {
								defaultParamsIndex[mangled] = map[int]*typescriptgo.SyntaxExpression{}
							}
							defaultParamsIndex[mangled][pIdx+1] = parameter.Initializer
						}
						function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: typ})
					}
					index[mangled] = function
					functionsByFile[fileName] = append(functionsByFile[fileName], function)
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
	return index
}
