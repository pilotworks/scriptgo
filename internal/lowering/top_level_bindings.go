package lowering

import (
	"fmt"
	"path/filepath"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

// Module top-level variables all become locals of the merged main function,
// so two modules may declare the same name (fs, os and crypto each export
// `constants`). Every module-level declaration gets a storage name before
// lowering: the first module (in program order) keeps the plain name and a
// later one is spelled name$m<file index>. Reads resolve a name through the
// file that declares or imports it, so functions see their own module's
// binding regardless of lowering order.

type topLevelBinding struct {
	Storage string
	File    string
	Decl    frontend.SyntaxStatement
}

type importedBinding struct {
	File string
	Name string
}

var (
	topLevelBindingsByFile = map[string]map[string]topLevelBinding{}
	topLevelBindingStorage = map[string]topLevelBinding{}
	importedBindingsByFile = map[string]map[string]importedBinding{}
)

func initializeTopLevelBindings(program frontend.Program) {
	topLevelBindingsByFile = map[string]map[string]topLevelBinding{}
	topLevelBindingStorage = map[string]topLevelBinding{}
	importedBindingsByFile = map[string]map[string]importedBinding{}
	claimed := map[string]bool{}
	for index, file := range program.Files {
		fileName := filepath.Clean(file.FileName)
		for _, statement := range file.Syntax.Statements {
			if (statement.Kind != "variable" && statement.Kind != "using" && statement.Kind != "await_using") || statement.Name == "" {
				continue
			}
			if _, declared := topLevelBindingsByFile[fileName][statement.Name]; declared {
				continue
			}
			storage := statement.Name
			if claimed[storage] {
				storage = fmt.Sprintf("%s$m%d", statement.Name, index)
			}
			claimed[statement.Name] = true
			binding := topLevelBinding{Storage: storage, File: fileName, Decl: statement}
			if topLevelBindingsByFile[fileName] == nil {
				topLevelBindingsByFile[fileName] = map[string]topLevelBinding{}
			}
			topLevelBindingsByFile[fileName][statement.Name] = binding
			topLevelBindingStorage[storage] = binding
		}
	}
	for _, file := range program.Files {
		fileName := filepath.Clean(file.FileName)
		for _, reference := range file.Imports {
			if reference.ResolvedFileName == "" || reference.TypeOnly {
				continue
			}
			for _, binding := range reference.Bindings {
				if binding.LocalName == "" || binding.TypeOnly || binding.ImportedName == "default" {
					continue
				}
				if importedBindingsByFile[fileName] == nil {
					importedBindingsByFile[fileName] = map[string]importedBinding{}
				}
				importedBindingsByFile[fileName][binding.LocalName] = importedBinding{File: filepath.Clean(reference.ResolvedFileName), Name: binding.ImportedName}
			}
		}
	}
}

// resolveTopLevelBinding finds the module-level variable a name refers to in
// the file at path: its own declaration, an imported binding, or a storage
// name produced by an earlier resolution.
func resolveTopLevelBinding(path string, name string) (topLevelBinding, bool) {
	fileName := filepath.Clean(path)
	if binding, ok := topLevelBindingsByFile[fileName][name]; ok {
		return binding, true
	}
	if imported, ok := importedBindingsByFile[fileName][name]; ok {
		if binding, ok := topLevelBindingsByFile[imported.File][imported.Name]; ok {
			return binding, true
		}
	}
	binding, ok := topLevelBindingStorage[name]
	return binding, ok
}

// moduleExportBinding finds the module-level variable `name` exported by the
// module that a namespace qualifier in the file at path refers to.
func moduleExportBinding(path string, qualifier string, name string) (topLevelBinding, bool) {
	cleanPath := filepath.Clean(path)
	target, ok := functionNamespacesByFile[cleanPath][qualifier]
	if !ok {
		target, ok = defaultNamespacesByFile[cleanPath][qualifier]
	}
	if !ok {
		return topLevelBinding{}, false
	}
	binding, ok := topLevelBindingsByFile[target][name]
	return binding, ok
}

// topLevelDeclarationBinding reports the module storage for a declaration
// statement lowered at module level (in main).
func topLevelDeclarationBinding(path string, statement frontend.SyntaxStatement) (topLevelBinding, bool) {
	binding, ok := topLevelBindingsByFile[filepath.Clean(path)][statement.Name]
	if !ok || binding.Decl.Span != statement.Span {
		return topLevelBinding{}, false
	}
	return binding, true
}

// bindingType is the IR type of a module-level variable, from its annotation
// or the checker's inferred type.
func bindingType(binding topLevelBinding) ir.Type {
	typ := toIRTypeForPath(binding.File, binding.Decl.Type)
	if typ == "" {
		typ = toIRTypeForPath(binding.File, binding.Decl.InferredType)
	}
	if typ == "" {
		typ = ir.TypeNumber
	}
	return typ
}
