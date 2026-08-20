package lowering

import (
	"sort"
	"strings"

	typescriptgo "github.com/microsoft/typescript-go/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

type ClassMeta struct {
	Name       string
	Extends    string
	Implements []string
	IsAbstract bool
	Fields     []typescriptgo.SyntaxField
	Statics    map[string]typescriptgo.SyntaxField
	HasCtor    bool
}

var classHierarchy = map[string]ClassMeta{}
var classSyntax = map[string]typescriptgo.SyntaxClass{}

func buildClassHierarchy(program frontend.Program) map[string]ClassMeta {
	hierarchy := map[string]ClassMeta{}
	syntax := map[string]typescriptgo.SyntaxClass{}
	for _, file := range program.Files {
		for _, stmt := range file.Syntax.Statements {
			if (stmt.Kind == "class" || stmt.Kind == "interface" || stmt.Kind == "type_alias") && stmt.Class != nil {
				syntax[stmt.Class.Name] = *stmt.Class
				meta := ClassMeta{
					Name:       stmt.Class.Name,
					Extends:    stmt.Class.Extends,
					Implements: stmt.Class.Implements,
					IsAbstract: stmt.Class.IsAbstract,
					Statics:    map[string]typescriptgo.SyntaxField{},
					HasCtor:    stmt.Class.Constructor != nil,
				}
				for _, f := range stmt.Class.Fields {
					if f.IsStatic {
						meta.Statics[f.Name] = f
					} else {
						meta.Fields = append(meta.Fields, f)
					}
				}
				hierarchy[stmt.Class.Name] = meta
			}
		}
	}
	classHierarchy = hierarchy
	classSyntax = syntax
	return hierarchy
}

func getInheritedMethods(className string, hierarchy map[string]ClassMeta) []typescriptgo.SyntaxMethod {
	meta, ok := hierarchy[className]
	if !ok {
		return nil
	}
	stmtClass, hasStmt := classSyntax[className]
	if !hasStmt {
		return nil
	}
	var inherited []typescriptgo.SyntaxMethod
	if meta.Extends != "" {
		inherited = getInheritedMethods(meta.Extends, hierarchy)
	}
	methodMap := map[string]typescriptgo.SyntaxMethod{}
	for _, m := range inherited {
		if !m.IsStatic {
			key := m.Kind + ":" + m.Name
			methodMap[key] = m
		}
	}
	for _, m := range stmtClass.Methods {
		key := m.Kind + ":" + m.Name
		methodMap[key] = m
	}
	var result []typescriptgo.SyntaxMethod
	for _, m := range methodMap {
		result = append(result, m)
	}
	return result
}

func getInheritedFields(className string, hierarchy map[string]ClassMeta) []typescriptgo.SyntaxField {
	meta, ok := hierarchy[className]
	if !ok {
		return nil
	}
	var fields []typescriptgo.SyntaxField
	if meta.Extends != "" {
		if baseFields := getInheritedFields(meta.Extends, hierarchy); len(baseFields) > 0 {
			fields = append(fields, baseFields...)
		} else {
			baseName := meta.Extends
			var typeArgs []string
			if strings.Contains(meta.Extends, "<") && strings.HasSuffix(meta.Extends, ">") {
				idx := strings.Index(meta.Extends, "<")
				baseName = meta.Extends[:idx]
				inner := meta.Extends[idx+1 : len(meta.Extends)-1]
				typeArgs = splitTypeArguments(inner)
			} else if strings.Contains(meta.Extends, "__") {
				idx := strings.Index(meta.Extends, "__")
				baseName = meta.Extends[:idx]
				typeArgs = strings.Split(meta.Extends[idx+2:], "_")
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
	fields = append(fields, meta.Fields...)
	return fields
}

func findMethodInHierarchy(className, methodName string, signatures map[string]ir.Function, hierarchy map[string]ClassMeta) (ir.Function, string, bool) {
	curr := className
	for curr != "" {
		cleanCurr := curr
		if idx := strings.Index(curr, "<"); idx != -1 {
			cleanCurr = curr[:idx]
		}
		mangled := cleanCurr + "_" + methodName
		if fn, ok := signatures[mangled]; ok {
			return fn, mangled, true
		}
		if meta, ok := hierarchy[cleanCurr]; ok {
			curr = meta.Extends
		} else {
			break
		}
	}
	return ir.Function{}, "", false
}

func findGetterInHierarchy(className, propName string, signatures map[string]ir.Function, hierarchy map[string]ClassMeta) (ir.Function, string, bool) {
	curr := className
	for curr != "" {
		cleanCurr := curr
		if idx := strings.Index(curr, "<"); idx != -1 {
			cleanCurr = curr[:idx]
		}
		mangled := cleanCurr + "_get_" + propName
		if fn, ok := signatures[mangled]; ok {
			return fn, mangled, true
		}
		if meta, ok := hierarchy[curr]; ok {
			curr = meta.Extends
		} else {
			break
		}
	}
	return ir.Function{}, "", false
}

func findSetterInHierarchy(className, propName string, signatures map[string]ir.Function, hierarchy map[string]ClassMeta) (ir.Function, string, bool) {
	curr := className
	for curr != "" {
		mangled := curr + "_set_" + propName
		if fn, ok := signatures[mangled]; ok {
			return fn, mangled, true
		}
		if meta, ok := hierarchy[curr]; ok {
			curr = meta.Extends
		} else {
			break
		}
	}
	return ir.Function{}, "", false
}

func findConstructorInHierarchy(className string, signatures map[string]ir.Function, hierarchy map[string]ClassMeta) (ir.Function, string, bool) {
	curr := className
	for curr != "" {
		mangled := curr + "_constructor"
		if fn, ok := signatures[mangled]; ok {
			return fn, mangled, true
		}
		if meta, ok := hierarchy[curr]; ok {
			curr = meta.Extends
		} else {
			break
		}
	}
	return ir.Function{}, "", false
}

func getHierarchyTag(className string, hierarchy map[string]ClassMeta) string {
	if className == "" {
		return ""
	}
	chain := []string{className}
	curr := className
	for {
		meta, ok := hierarchy[curr]
		if !ok || meta.Extends == "" {
			break
		}
		chain = append(chain, meta.Extends)
		curr = meta.Extends
	}
	return ":" + strings.Join(chain, ":") + ":"
}

func isSubclassOf(derived, base string, hierarchy map[string]ClassMeta) bool {
	if derived == base {
		return true
	}
	curr := derived
	for {
		meta, ok := hierarchy[curr]
		if !ok || meta.Extends == "" {
			break
		}
		if meta.Extends == base {
			return true
		}
		curr = meta.Extends
	}
	return false
}

func classDepth(className string, hierarchy map[string]ClassMeta) int {
	depth := 0
	curr := className
	for {
		meta, ok := hierarchy[curr]
		if !ok || meta.Extends == "" {
			break
		}
		depth++
		curr = meta.Extends
	}
	return depth
}

func findOverridingSubclasses(baseClass, methodName string, hierarchy map[string]ClassMeta, signatures map[string]ir.Function) []string {
	var result []string
	for name := range hierarchy {
		if name != baseClass && isSubclassOf(name, baseClass, hierarchy) {
			mangled := name + "_" + methodName
			if _, ok := signatures[mangled]; ok {
				result = append(result, name)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		dI := classDepth(result[i], hierarchy)
		dJ := classDepth(result[j], hierarchy)
		if dI != dJ {
			return dI < dJ
		}
		return result[i] < result[j]
	})
	return result
}
