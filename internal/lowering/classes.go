package lowering

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pilotworks/scriptgo/internal/frontend"
)

type ClassMeta struct {
	Name        string
	FileName    string
	Extends     string
	Implements  []string
	IsAbstract  bool
	IsInterface bool
	IsTypeAlias bool
	Fields      []frontend.SyntaxField
	Statics     map[string]frontend.SyntaxField
	HasCtor     bool
}

var classHierarchy = map[string]ClassMeta{}
var classSyntax = map[string]frontend.SyntaxClass{}

func buildClassHierarchy(program frontend.Program) map[string]ClassMeta {
	hierarchy := map[string]ClassMeta{}
	syntax := map[string]frontend.SyntaxClass{}
	for _, file := range program.Files {
		fileName := filepath.Clean(file.FileName)
		var visitStmt func(stmt frontend.SyntaxStatement)
		visitStmt = func(stmt frontend.SyntaxStatement) {
			if (stmt.Kind == "class" || stmt.Kind == "interface" || stmt.Kind == "type_alias") && stmt.Class != nil {
				if stmt.Class.Name == "" {
					return
				}
				className := classIdentityForPath(fileName, stmt.Class.Name)
				if className == "" {
					return
				}
				classDef := *stmt.Class
				if existingSyntax, exists := syntax[className]; exists {
					if stmt.Kind == "interface" || stmt.Kind == "type_alias" {
						return
					}
					mergedMethods := make([]frontend.SyntaxMethod, 0, len(existingSyntax.Methods)+len(stmt.Class.Methods))
					methodSeen := map[string]int{}
					for _, m := range existingSyntax.Methods {
						key := fmt.Sprintf("%v:%s:%s", m.IsStatic, m.Kind, m.Name)
						methodSeen[key] = len(mergedMethods)
						mergedMethods = append(mergedMethods, m)
					}
					for _, m := range classDef.Methods {
						key := fmt.Sprintf("%v:%s:%s", m.IsStatic, m.Kind, m.Name)
						if idx, found := methodSeen[key]; found {
							if mergedMethods[idx].Body == nil && m.Body != nil {
								mergedMethods[idx] = m
							}
						} else {
							methodSeen[key] = len(mergedMethods)
							mergedMethods = append(mergedMethods, m)
						}
					}
					classDef.Methods = mergedMethods
				}
				syntax[className] = classDef
				meta := ClassMeta{
					Name:        className,
					FileName:    fileName,
					Extends:     qualifyClassType(fileName, classDef.Extends),
					Implements:  make([]string, 0, len(classDef.Implements)),
					IsAbstract:  classDef.IsAbstract,
					IsInterface: stmt.Kind == "interface",
					IsTypeAlias: stmt.Kind == "type_alias",
					Statics:     map[string]frontend.SyntaxField{},
					HasCtor:     classDef.Constructor != nil,
				}
				for _, implemented := range classDef.Implements {
					meta.Implements = append(meta.Implements, qualifyClassType(fileName, implemented))
				}
				for _, f := range classDef.Fields {
					if f.IsStatic {
						meta.Statics[f.Name] = f
					} else {
						meta.Fields = append(meta.Fields, f)
					}
				}
				hierarchy[className] = meta
			} else if stmt.Kind == "namespace" || stmt.Kind == "block" {
				for _, sub := range stmt.Body {
					visitStmt(sub)
				}
			}
		}
		for _, stmt := range file.Syntax.Statements {
			visitStmt(stmt)
		}
	}
	classHierarchy = hierarchy
	classSyntax = syntax
	return hierarchy
}

func isSubtype(subType, superType string) bool {
	sub := strings.TrimPrefix(subType, "object:")
	super := strings.TrimPrefix(superType, "object:")
	if sub == super {
		return true
	}
	subNorm := strings.ReplaceAll(strings.ReplaceAll(sub, "<", "_"), ">", "")
	superNorm := strings.ReplaceAll(strings.ReplaceAll(super, "<", "_"), ">", "")
	if subNorm == superNorm {
		return true
	}
	if (strings.HasPrefix(subNorm, "Promise_") || subNorm == "Promise") && (strings.HasPrefix(superNorm, "Promise_") || superNorm == "Promise") {
		return true
	}
	if sub == "__shape_empty" && (strings.HasPrefix(super, "Record") || super == "Object" || super == "object" || strings.HasPrefix(super, "__shape_")) {
		return true
	}
	meta, ok := classHierarchy[sub]
	if !ok {
		return false
	}
	if meta.Extends != "" {
		for _, rawBase := range strings.Split(meta.Extends, ",") {
			base := strings.TrimSpace(rawBase)
			if base != "" && (base == super || isSubtype(base, super)) {
				return true
			}
		}
	}
	for _, imp := range meta.Implements {
		imp = strings.TrimSpace(imp)
		if imp != "" && (imp == super || isSubtype(imp, super)) {
			return true
		}
	}
	return false
}

func getInheritedMethods(className string, hierarchy map[string]ClassMeta) []frontend.SyntaxMethod {
	meta, ok := hierarchy[className]
	if !ok {
		return nil
	}
	stmtClass, hasStmt := classSyntax[className]
	if !hasStmt {
		return nil
	}
	var inherited []frontend.SyntaxMethod
	if meta.Extends != "" {
		for _, rawBase := range strings.Split(meta.Extends, ",") {
			base := strings.TrimSpace(rawBase)
			if base == "" {
				continue
			}
			inherited = append(inherited, getInheritedMethods(base, hierarchy)...)
		}
	}
	// Keep declaration order (base members first, overrides in place) so the
	// lowered module is deterministic.
	methodMap := map[string]frontend.SyntaxMethod{}
	var order []string
	remember := func(key string, m frontend.SyntaxMethod) {
		if _, seen := methodMap[key]; !seen {
			order = append(order, key)
		}
		methodMap[key] = m
	}
	for _, m := range inherited {
		remember(fmt.Sprintf("%v:%s:%s", m.IsStatic, m.Kind, m.Name), m)
	}
	for _, m := range stmtClass.Methods {
		key := fmt.Sprintf("%v:%s:%s", m.IsStatic, m.Kind, m.Name)
		if existing, ok := methodMap[key]; ok && existing.Body != nil && m.Body == nil {
			continue
		}
		remember(key, m)
	}
	result := make([]frontend.SyntaxMethod, 0, len(order))
	for _, key := range order {
		result = append(result, methodMap[key])
	}
	return result
}

func cleanGenericBase(className string) string {
	if className == "" {
		return ""
	}
	if idx := strings.Index(className, "<"); idx != -1 {
		return className[:idx]
	}
	hasPrefix := strings.HasPrefix(className, "__")
	s := className
	if hasPrefix {
		s = strings.TrimPrefix(className, "__")
	}
	parts := strings.Split(s, "__")
	if hasPrefix {
		return "__" + parts[0]
	}
	return parts[0]
}

func getInheritedFields(className string, hierarchy map[string]ClassMeta) []frontend.SyntaxField {
	meta, ok := hierarchy[className]
	if !ok {
		if candidates, hasCand := classCandidates[className]; hasCand && len(candidates) > 0 {
			meta, ok = hierarchy[candidates[0].Internal]
		}
	}
	if !ok {
		return nil
	}
	var fields []frontend.SyntaxField
	if meta.Extends != "" {
		for _, rawBase := range strings.Split(meta.Extends, ",") {
			base := strings.TrimSpace(rawBase)
			if base == "" {
				continue
			}
			if baseFields := getInheritedFields(base, hierarchy); len(baseFields) > 0 {
				fields = append(fields, baseFields...)
			} else if isBuiltinErrorClass(base) {
				fields = append(fields, builtinErrorFields()...)
			} else {
				baseName := base
				var typeArgs []string
				if strings.Contains(base, "<") && strings.HasSuffix(base, ">") {
					idx := strings.Index(base, "<")
					baseName = base[:idx]
					inner := base[idx+1 : len(base)-1]
					typeArgs = splitTypeArguments(inner)
				} else {
					for i := len(base) - 2; i > 0; i-- {
						if base[i] == '_' && base[i+1] == '_' && base[i-1] != '_' {
							baseName = base[:i]
							typeArgs = strings.Split(base[i+2:], "_")
							break
						}
					}
				}
				baseFields = getInheritedFields(baseName, hierarchy)
				if len(typeArgs) > 0 {
					baseClass := classSyntax[baseName]
					subst := map[string]string{}
					for i, tp := range baseClass.TypeParameters {
						if i < len(typeArgs) {
							subst[tp] = typeArgs[i]
						}
					}
					for _, bf := range baseFields {
						f := bf
						f.Type = substituteType(f.Type, subst)
						f.InferredType = substituteType(f.InferredType, subst)
						fields = append(fields, f)
					}
				} else {
					fields = append(fields, baseFields...)
				}
			}
		}
	}
	seenFields := map[string]int{}
	for idx, f := range fields {
		seenFields[f.Name] = idx
	}
	for _, f := range meta.Fields {
		if existingIdx, exists := seenFields[f.Name]; exists {
			if f.Initializer != nil {
				fields[existingIdx].Initializer = f.Initializer
			}
			if f.Type != "" {
				fields[existingIdx].Type = f.Type
			}
			if f.InferredType != "" {
				fields[existingIdx].InferredType = f.InferredType
			}
		} else {
			seenFields[f.Name] = len(fields)
			fields = append(fields, f)
		}
	}
	return fields
}

func normalizeGenericName(s string) string {
	s = strings.ReplaceAll(s, "<", "__")
	s = strings.ReplaceAll(s, ">", "")
	s = strings.ReplaceAll(s, ", ", "_")
	s = strings.ReplaceAll(s, ",", "_")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

func methodImplementationName(className, methodName string) string {
	return className + "_" + methodName + "_impl"
}

// isBuiltinErrorClass reports Error and the native error constructors, whose
// instances carry the built-in Error shape.
func isBuiltinErrorClass(name string) bool {
	return name == "Error" || builtinErrorBase[name] != ""
}

// builtinErrorFields is the built-in Error layout (see the Error builtin
// shape), inherited first by user classes that extend an error class.
func builtinErrorFields() []frontend.SyntaxField {
	return []frontend.SyntaxField{
		{Name: "message", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "stack", Type: "string"},
		{Name: "cause", Type: "string"},
	}
}
