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

type functionIdentity struct {
	PublicName string
	Internal   string
	FileName   string
}

type indexedFunction struct {
	Function   ir.Function
	PublicName string
}

var (
	functionIdentitiesByFile = map[string]map[string]functionIdentity{}
	functionIdentitiesByName = map[string]functionIdentity{}
	functionCandidates       = map[string][]functionIdentity{}
	functionImportsByFile    = map[string]map[string]functionIdentity{}
	functionNamespacesByFile = map[string]map[string]string{}
)

func initializeFunctionIdentities(program frontend.Program) {
	functionIdentitiesByFile = map[string]map[string]functionIdentity{}
	functionIdentitiesByName = map[string]functionIdentity{}
	functionCandidates = map[string][]functionIdentity{}
	functionImportsByFile = map[string]map[string]functionIdentity{}
	functionNamespacesByFile = map[string]map[string]string{}

	publicFiles := map[string]map[string]bool{}
	filesByPrefix := map[string][]string{}
	for _, file := range program.Files {
		fileName := filepath.Clean(file.FileName)
		prefix := functionModulePrefix(fileName, file.BuiltinName)
		filesByPrefix[prefix] = append(filesByPrefix[prefix], fileName)
		for _, statement := range file.Syntax.Statements {
			if !isTopLevelFunctionDeclaration(statement) || statement.Name == "" {
				continue
			}
			if publicFiles[statement.Name] == nil {
				publicFiles[statement.Name] = map[string]bool{}
			}
			publicFiles[statement.Name][fileName] = true
		}
	}

	modulePrefixes := map[string]string{}
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

	for _, file := range program.Files {
		fileName := filepath.Clean(file.FileName)
		for _, statement := range file.Syntax.Statements {
			if !isTopLevelFunctionDeclaration(statement) || statement.Name == "" {
				continue
			}
			identity := functionIdentity{
				PublicName: statement.Name,
				Internal:   statement.Name,
				FileName:   fileName,
			}
			if statement.Name == "main" && len(publicFiles[statement.Name]) == 1 {
				identity.Internal = "main$user"
			} else if len(publicFiles[statement.Name]) > 1 {
				identity.Internal = modulePrefixes[fileName] + "_" + statement.Name
			}
			if functionIdentitiesByFile[fileName] == nil {
				functionIdentitiesByFile[fileName] = map[string]functionIdentity{}
			}
			if existing, ok := functionIdentitiesByFile[fileName][statement.Name]; ok {
				identity = existing
			}
			functionIdentitiesByFile[fileName][statement.Name] = identity
			functionCandidates[statement.Name] = appendUniqueFunctionIdentity(functionCandidates[statement.Name], identity)
		}
	}
	for publicName, candidates := range functionCandidates {
		if len(candidates) == 1 {
			functionIdentitiesByName[publicName] = candidates[0]
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
				if functionNamespacesByFile[fileName] == nil {
					functionNamespacesByFile[fileName] = map[string]string{}
				}
				functionNamespacesByFile[fileName][reference.LocalName] = targetFile
			}
			for _, binding := range reference.Bindings {
				if binding.LocalName == "" || binding.TypeOnly {
					continue
				}
				if target, ok := functionIdentitiesByFile[targetFile][binding.ImportedName]; ok {
					if functionImportsByFile[fileName] == nil {
						functionImportsByFile[fileName] = map[string]functionIdentity{}
					}
					functionImportsByFile[fileName][binding.LocalName] = target
				}
			}
		}
	}
}

func isTopLevelFunctionDeclaration(statement typescriptgo.SyntaxStatement) bool {
	switch statement.Kind {
	case "declare_function", "function", "generator_function", "async_function", "async_generator_function":
		return true
	default:
		return statement.IsGenerator || statement.IsAsync
	}
}

func appendUniqueFunctionIdentity(values []functionIdentity, value functionIdentity) []functionIdentity {
	for _, existing := range values {
		if existing.FileName == value.FileName && existing.PublicName == value.PublicName {
			return values
		}
	}
	return append(values, value)
}

func functionModulePrefix(fileName, builtinName string) string {
	if builtinName != "" {
		return "__builtin_" + sanitizeFunctionIdentityPart(builtinName)
	}
	base := filepath.Base(fileName)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return "__module_" + sanitizeFunctionIdentityPart(base)
}

func sanitizeFunctionIdentityPart(value string) string {
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

func functionIdentityForPath(path, name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}
	fileName := filepath.Clean(path)
	if imported, ok := functionImportsByFile[fileName][name]; ok {
		return imported.Internal
	}
	if dot := strings.Index(name, "."); dot > 0 {
		if targetFile, ok := functionNamespacesByFile[fileName][name[:dot]]; ok {
			if imported, ok := functionIdentitiesByFile[targetFile][name[dot+1:]]; ok {
				return imported.Internal
			}
		}
	}
	if local, ok := functionIdentitiesByFile[fileName][name]; ok {
		return local.Internal
	}
	if identity, ok := functionIdentityByInternal(name); ok {
		return identity.Internal
	}
	if unique, ok := functionIdentitiesByName[name]; ok {
		return unique.Internal
	}
	return name
}

func functionIdentityByInternal(name string) (functionIdentity, bool) {
	for _, candidates := range functionCandidates {
		for _, candidate := range candidates {
			if candidate.Internal == name {
				return candidate, true
			}
		}
	}
	return functionIdentity{}, false
}

func resolveFunctionSignature(path, name string, signatures map[string]ir.Function) (ir.Function, bool) {
	if internal := functionIdentityForPath(path, name); internal != name {
		if function, ok := signatures[internal]; ok {
			return function, true
		}
	}
	function, ok := signatures[name]
	return function, ok
}

func functionPublicName(name string) string {
	if identity, ok := functionIdentityByInternal(name); ok {
		return identity.PublicName
	}
	return name
}
