package lowering

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

type classIdentity struct {
	PublicName string
	Internal   string
	FileName   string
}

var (
	classIdentitiesByFile = map[string]map[string]classIdentity{}
	classIdentitiesByName = map[string]classIdentity{}
	classCandidates       = map[string][]classIdentity{}
	classImportsByFile    = map[string]map[string]classIdentity{}
	classNamespacesByFile = map[string]map[string]string{}
)

// initializeClassIdentities gives declarations with the same public name a
// stable IR identity. TypeScript modules may legally export the same name,
// while IR shapes and functions share one module-wide namespace.
func initializeClassIdentities(program frontend.Program) {
	classIdentitiesByFile = map[string]map[string]classIdentity{}
	classIdentitiesByName = map[string]classIdentity{}
	classCandidates = map[string][]classIdentity{}
	classImportsByFile = map[string]map[string]classIdentity{}
	classNamespacesByFile = map[string]map[string]string{}

	fileBuiltins := make(map[string]string, len(program.Files))
	filesByPrefix := make(map[string][]string)
	for _, file := range program.Files {
		fileName := filepath.Clean(file.FileName)
		fileBuiltins[fileName] = file.BuiltinName
		prefix := classModulePrefix(fileName, file.BuiltinName)
		filesByPrefix[prefix] = append(filesByPrefix[prefix], fileName)
	}
	modulePrefixes := make(map[string]string, len(program.Files))
	for prefix, files := range filesByPrefix {
		sort.Strings(files)
		for index, fileName := range files {
			if index == 0 {
				modulePrefixes[fileName] = prefix
			} else {
				modulePrefixes[fileName] = prefix + "_" + strconv.Itoa(index+1)
			}
		}
	}

	publicFiles := map[string]map[string]bool{}
	for _, file := range program.Files {
		fileName := filepath.Clean(file.FileName)
		var visit func(typescriptgo.SyntaxStatement)
		visit = func(statement typescriptgo.SyntaxStatement) {
			if (statement.Kind == "class" || statement.Kind == "interface" || statement.Kind == "type_alias") && statement.Class != nil {
				if publicFiles[statement.Class.Name] == nil {
					publicFiles[statement.Class.Name] = map[string]bool{}
				}
				publicFiles[statement.Class.Name][fileName] = true
				return
			}
			if statement.Kind == "block" || statement.Kind == "namespace" {
				for _, child := range statement.Body {
					visit(child)
				}
			}
		}
		for _, statement := range file.Syntax.Statements {
			visit(statement)
		}
	}
	for _, file := range program.Files {
		fileName := filepath.Clean(file.FileName)
		var visit func(typescriptgo.SyntaxStatement)
		visit = func(statement typescriptgo.SyntaxStatement) {
			if (statement.Kind == "class" || statement.Kind == "interface" || statement.Kind == "type_alias") && statement.Class != nil {
				publicName := statement.Class.Name
				identity := classIdentity{PublicName: publicName, Internal: publicName, FileName: fileName}
				if len(publicFiles[publicName]) > 1 {
					identity.Internal = modulePrefixes[fileName] + "_" + publicName
				}
				if classIdentitiesByFile[fileName] == nil {
					classIdentitiesByFile[fileName] = map[string]classIdentity{}
				}
				if existing, ok := classIdentitiesByFile[fileName][publicName]; ok {
					identity = existing
				}
				classIdentitiesByFile[fileName][publicName] = identity
				classCandidates[publicName] = appendUniqueClassIdentity(classCandidates[publicName], identity)
				return
			}
			if statement.Kind == "block" || statement.Kind == "namespace" {
				for _, child := range statement.Body {
					visit(child)
				}
			}
		}
		for _, statement := range file.Syntax.Statements {
			visit(statement)
		}
	}
	for publicName, candidates := range classCandidates {
		if len(candidates) == 1 {
			classIdentitiesByName[publicName] = candidates[0]
		} else {
			delete(classIdentitiesByName, publicName)
		}
	}

	for _, file := range program.Files {
		fileName := filepath.Clean(file.FileName)
		for _, reference := range file.Imports {
			if reference.ResolvedFileName == "" {
				continue
			}
			targetFile := filepath.Clean(reference.ResolvedFileName)
			if reference.LocalName != "" {
				if classNamespacesByFile[fileName] == nil {
					classNamespacesByFile[fileName] = map[string]string{}
				}
				classNamespacesByFile[fileName][reference.LocalName] = targetFile
			}
			for _, binding := range reference.Bindings {
				if binding.LocalName == "" || binding.TypeOnly {
					continue
				}
				if target, ok := classIdentitiesByFile[targetFile][binding.ImportedName]; ok {
					if classImportsByFile[fileName] == nil {
						classImportsByFile[fileName] = map[string]classIdentity{}
					}
					classImportsByFile[fileName][binding.LocalName] = target
				}
			}
		}
	}
}

func appendUniqueClassIdentity(values []classIdentity, value classIdentity) []classIdentity {
	for _, existing := range values {
		if existing.FileName == value.FileName && existing.PublicName == value.PublicName {
			return values
		}
	}
	return append(values, value)
}

func classModulePrefix(fileName, builtinName string) string {
	if builtinName != "" {
		return "__builtin_" + sanitizeClassIdentityPart(builtinName)
	}
	base := filepath.Base(fileName)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return "__module_" + sanitizeClassIdentityPart(base)
}

func sanitizeClassIdentityPart(value string) string {
	var builder strings.Builder
	for _, ch := range value {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			builder.WriteRune(ch)
		} else {
			builder.WriteByte('_')
		}
	}
	return builder.String()
}

func classIdentityForPath(path, name string) string {
	name = strings.TrimSpace(strings.TrimPrefix(name, "object:"))
	if name == "" {
		return name
	}
	if _, ok := classIdentityByInternal(name); ok {
		return name
	}
	if idx := strings.Index(name, "<"); idx > 0 && strings.HasSuffix(name, ">") {
		base := classIdentityForPath(path, name[:idx])
		return base + name[idx:]
	}
	fileName := filepath.Clean(path)
	if imported, ok := classImportsByFile[fileName][name]; ok {
		return imported.Internal
	}
	if idx := strings.Index(name, "."); idx > 0 {
		if targetFile, ok := classNamespacesByFile[fileName][name[:idx]]; ok {
			if imported, ok := classIdentitiesByFile[targetFile][name[idx+1:]]; ok {
				return imported.Internal
			}
		}
	}
	if local, ok := classIdentitiesByFile[fileName][name]; ok {
		return local.Internal
	}
	if unique, ok := classIdentitiesByName[name]; ok {
		return unique.Internal
	}
	return name
}

func classIdentityByInternal(name string) (classIdentity, bool) {
	for _, candidates := range classCandidates {
		for _, candidate := range candidates {
			if candidate.Internal == name {
				return candidate, true
			}
		}
	}
	return classIdentity{}, false
}

func classPublicName(name string) string {
	if identity, ok := classIdentityByInternal(strings.TrimPrefix(name, "object:")); ok {
		return identity.PublicName
	}
	return strings.TrimPrefix(name, "object:")
}

func qualifyClassType(path, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if strings.HasSuffix(value, "[]") {
		return qualifyClassType(path, strings.TrimSuffix(value, "[]")) + "[]"
	}
	if strings.HasPrefix(value, "object:") {
		return "object:" + classIdentityForPath(path, strings.TrimPrefix(value, "object:"))
	}
	return classIdentityForPath(path, value)
}

func toIRTypeForPath(path, value string) ir.Type {
	typ := toIRType(value)
	if strings.HasPrefix(string(typ), "object:") {
		return ir.Type(qualifyClassType(path, string(typ)))
	}
	return typ
}
