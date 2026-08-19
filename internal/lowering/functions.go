package lowering

import (
	"path/filepath"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// buildFunctionIndex collects function signatures and namespace import aliases
// from the checked module graph. Lowering does not need to know which module a
// namespace came from; the frontend has already resolved that edge.
func buildFunctionIndex(program frontend.Program) map[string]ir.Function {
	index := map[string]ir.Function{}
	functionsByFile := map[string][]ir.Function{}
	for _, file := range program.Files {
		fileName := filepath.Clean(file.FileName)
		for _, statement := range file.Syntax.Statements {
			if statement.Kind != "function" {
				continue
			}
			function := ir.Function{Name: statement.Name, ReturnType: toIRType(statement.Type)}
			if function.ReturnType == "" {
				function.ReturnType = ir.TypeVoid
			}
			for _, parameter := range statement.Parameters {
				typ := toIRType(parameter.Type)
				if parameter.Rest {
					typ = ir.TypeStringArray
				}
				function.Parameters = append(function.Parameters, ir.Parameter{Name: parameter.Name, Type: typ})
			}
			index[function.Name] = function
			if file.BuiltinName != "" {
				index[file.BuiltinName+"."+function.Name] = function
			}
			functionsByFile[fileName] = append(functionsByFile[fileName], function)
		}
	}
	for _, file := range program.Files {
		for _, reference := range file.Imports {
			if reference.LocalName == "" || reference.ResolvedFileName == "" {
				continue
			}
			for _, function := range functionsByFile[filepath.Clean(reference.ResolvedFileName)] {
				index[reference.LocalName+"."+function.Name] = function
			}
		}
	}
	return index
}
