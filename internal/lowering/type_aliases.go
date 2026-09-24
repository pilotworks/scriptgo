package lowering

import (
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
)

// initializeTypeAliases builds the type alias and enum index across all files
// in the program (including declaration files) so that monomorphization and
// lowering have access to spec-defined and user-defined type aliases.
func initializeTypeAliases(program frontend.Program) {
	if typeAliasesIndex == nil {
		typeAliasesIndex = make(map[string]string)
	}
	for _, file := range program.Files {
		for _, statement := range file.Syntax.Statements {
			if statement.Kind == "type_alias" && statement.Name != "" && statement.Type != "" {
				typeAliasesIndex[statement.Name] = statement.Type
			}
			if statement.Kind == "enum" && statement.Enum != nil {
				hasString := false
				for _, m := range statement.Enum.Members {
					if m.Value != "" && (strings.HasPrefix(m.Value, "\"") || strings.HasPrefix(m.Value, "'")) {
						hasString = true
						break
					}
				}
				if hasString {
					typeAliasesIndex[statement.Enum.Name] = "string"
				} else {
					typeAliasesIndex[statement.Enum.Name] = "number"
				}
			}
		}
	}
}
