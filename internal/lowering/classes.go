package lowering

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

type ClassMeta struct {
	Name        string
	FileName    string
	Extends     string
	Implements  []string
	IsAbstract  bool
	IsInterface bool
	IsTypeAlias bool
	Fields      []typescriptgo.SyntaxField
	Statics     map[string]typescriptgo.SyntaxField
	HasCtor     bool
}

var classHierarchy = map[string]ClassMeta{}
var classSyntax = map[string]typescriptgo.SyntaxClass{}

func buildClassHierarchy(program frontend.Program) map[string]ClassMeta {
	hierarchy := map[string]ClassMeta{}
	syntax := map[string]typescriptgo.SyntaxClass{}
	for _, file := range program.Files {
		fileName := filepath.Clean(file.FileName)
		var visitStmt func(stmt typescriptgo.SyntaxStatement)
		visitStmt = func(stmt typescriptgo.SyntaxStatement) {
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
					mergedMethods := make([]typescriptgo.SyntaxMethod, 0, len(existingSyntax.Methods)+len(stmt.Class.Methods))
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
					Statics:     map[string]typescriptgo.SyntaxField{},
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
		for _, rawBase := range strings.Split(meta.Extends, ",") {
			base := strings.TrimSpace(rawBase)
			if base == "" {
				continue
			}
			inherited = append(inherited, getInheritedMethods(base, hierarchy)...)
		}
	}
	methodMap := map[string]typescriptgo.SyntaxMethod{}
	for _, m := range inherited {
		key := fmt.Sprintf("%v:%s:%s", m.IsStatic, m.Kind, m.Name)
		methodMap[key] = m
	}
	for _, m := range stmtClass.Methods {
		key := fmt.Sprintf("%v:%s:%s", m.IsStatic, m.Kind, m.Name)
		if existing, ok := methodMap[key]; ok && existing.Body != nil && m.Body == nil {
			continue
		}
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
		for _, rawBase := range strings.Split(meta.Extends, ",") {
			base := strings.TrimSpace(rawBase)
			if base == "" {
				continue
			}
			if baseFields := getInheritedFields(base, hierarchy); len(baseFields) > 0 {
				fields = append(fields, baseFields...)
			} else {
				baseName := base
				var typeArgs []string
				if strings.Contains(base, "<") && strings.HasSuffix(base, ">") {
					idx := strings.Index(base, "<")
					baseName = base[:idx]
					inner := base[idx+1 : len(base)-1]
					typeArgs = splitTypeArguments(inner)
				} else if strings.Contains(base, "__") {
					idx := strings.Index(base, "__")
					baseName = base[:idx]
					typeArgs = strings.Split(base[idx+2:], "_")
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

func findStaticMethodInHierarchy(className, methodName string, signatures map[string]ir.Function, hierarchy map[string]ClassMeta) (ir.Function, string, bool) {
	curr := className
	for curr != "" {
		if strings.Contains(curr, "<") && strings.HasSuffix(curr, ">") {
			mangledCls := mangleGenericTypeString(curr)
			mangled := mangledCls + "_static_" + methodName
			if fn, ok := signatures[mangled]; ok {
				return fn, mangled, true
			}
			mangledOld := mangledCls + "_" + methodName
			if fn, ok := signatures[mangledOld]; ok && (len(fn.Parameters) == 0 || fn.Parameters[0].Name != "this") {
				return fn, mangledOld, true
			}
		}
		cleanCurr := curr
		if idx := strings.Index(curr, "<"); idx != -1 {
			cleanCurr = curr[:idx]
		}
		mangled := cleanCurr + "_static_" + methodName
		if fn, ok := signatures[mangled]; ok {
			return fn, mangled, true
		}
		mangledOld := cleanCurr + "_" + methodName
		if fn, ok := signatures[mangledOld]; ok && (len(fn.Parameters) == 0 || fn.Parameters[0].Name != "this") {
			return fn, mangledOld, true
		}
		if meta, ok := hierarchy[cleanCurr]; ok {
			curr = meta.Extends
		} else {
			break
		}
	}
	return ir.Function{}, "", false
}

func normalizeGenericName(s string) string {
	s = strings.ReplaceAll(s, "<", "__")
	s = strings.ReplaceAll(s, ">", "")
	s = strings.ReplaceAll(s, ", ", "_")
	s = strings.ReplaceAll(s, ",", "_")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

func findImplementationInHierarchy(className, methodName string, signatures map[string]ir.Function, hierarchy map[string]ClassMeta) (ir.Function, string, bool) {
	curr := className
	for curr != "" {
		if fn, ok := signatures[methodImplementationName(curr, methodName)]; ok {
			return fn, methodImplementationName(curr, methodName), true
		}
		clean := curr
		if idx := strings.IndexAny(clean, "<"); idx >= 0 {
			clean = clean[:idx]
		}
		if fn, ok := signatures[methodImplementationName(clean, methodName)]; ok {
			return fn, methodImplementationName(clean, methodName), true
		}
		meta, ok := hierarchy[clean]
		if !ok {
			break
		}
		curr = meta.Extends
	}
	return ir.Function{}, "", false
}

func findMethodInHierarchy(className, methodName string, signatures map[string]ir.Function, hierarchy map[string]ClassMeta) (ir.Function, string, bool) {
	if className == "this" || className == "" {
		for cls := range hierarchy {
			if fn, ok := signatures[methodImplementationName(cls, methodName)]; ok {
				return fn, methodImplementationName(cls, methodName), true
			}
		}
	}
	curr := className
	if className != "" && className != "this" {
		dispatchName := className + "_" + methodName + "_dispatch"
		if fn, ok := signatures[dispatchName]; ok {
			return fn, dispatchName, true
		}
	}
	for curr != "" {
		mangledDirect := methodImplementationName(curr, methodName)
		if fn, ok := signatures[mangledDirect]; ok {
			return fn, mangledDirect, true
		}
		if strings.Contains(curr, "<") && strings.HasSuffix(curr, ">") {
			mangledCls := mangleGenericTypeString(curr)
			mangled := methodImplementationName(mangledCls, methodName)
			if fn, ok := signatures[mangled]; ok {
				return fn, mangled, true
			}
		}
		cleanCurr := curr
		if idx := strings.Index(curr, "<"); idx != -1 {
			cleanCurr = curr[:idx]
		} else if idx := strings.Index(curr, "__"); idx != -1 {
			cleanCurr = curr[:idx]
		}
		mangled := methodImplementationName(cleanCurr, methodName)
		if fn, ok := signatures[mangled]; ok {
			return fn, mangled, true
		}
		genMangled := "Generator_" + cleanCurr + "_" + methodName + "_impl"
		if fn, ok := signatures[genMangled]; ok {
			return fn, genMangled, true
		}
		if meta, ok := hierarchy[cleanCurr]; ok {
			curr = meta.Extends
		} else {
			break
		}
	}
	cleanCls := className
	if idx := strings.Index(className, "<"); idx != -1 {
		cleanCls = className[:idx]
	} else if idx := strings.Index(className, "__"); idx != -1 {
		cleanCls = className[:idx]
	}
	if cleanCls == "" {
		return ir.Function{}, "", false
	}
	normClass := normalizeGenericName(className)
	mangledCls := className
	if strings.Contains(className, "<") && strings.HasSuffix(className, ">") {
		mangledCls = mangleGenericTypeString(className)
	}
	var exactSubs []string
	var allSubs []string
	seenSub := map[string]bool{}
	for subName, meta := range hierarchy {
		isExact := false
		isLoose := false
		for _, imp := range meta.Implements {
			impClean := imp
			if idx := strings.Index(imp, "<"); idx != -1 {
				impClean = imp[:idx]
			} else if idx := strings.Index(imp, "__"); idx != -1 {
				impClean = imp[:idx]
			}
			impMangled := imp
			if strings.Contains(imp, "<") && strings.HasSuffix(imp, ">") {
				impMangled = mangleGenericTypeString(imp)
			}
			if imp == className || impMangled == mangledCls || normalizeGenericName(imp) == normClass {
				isExact = true
				break
			}
			if imp == cleanCls || impClean == cleanCls {
				isLoose = true
			}
		}
		if meta.Extends != "" && !isExact && !isLoose && (meta.Extends == className || meta.Extends == cleanCls || normalizeGenericName(meta.Extends) == normClass) {
			if meta.Extends == className || normalizeGenericName(meta.Extends) == normClass {
				isExact = true
			} else {
				isLoose = true
			}
		}
		subMangled := methodImplementationName(subName, methodName)
		if _, ok := signatures[subMangled]; ok && !seenSub[subName] {
			if isExact {
				exactSubs = append(exactSubs, subName)
				seenSub[subName] = true
			} else if isLoose {
				allSubs = append(allSubs, subName)
			}
		}
	}
	candidateSubs := exactSubs
	if len(candidateSubs) == 0 {
		candidateSubs = allSubs
	}
	if len(candidateSubs) == 1 {
		subMangled := methodImplementationName(candidateSubs[0], methodName)
		return signatures[subMangled], subMangled, true
	} else if len(candidateSubs) > 1 {
		slices.Sort(candidateSubs)
		dispatchName := methodDispatcherName(cleanCls, methodName)
		if fn, ok := signatures[dispatchName]; ok {
			return fn, dispatchName, true
		}
		firstSig := signatures[methodImplementationName(candidateSubs[0], methodName)]
		var params []ir.Parameter
		var forwardArgs []string
		for i, p := range firstSig.Parameters {
			if i == 0 {
				params = append(params, ir.Parameter{Name: "this", Type: ir.TypePointer})
				forwardArgs = append(forwardArgs, "this")
			} else {
				params = append(params, p)
				forwardArgs = append(forwardArgs, p.Name)
			}
		}
		dispatchFn := ir.Function{
			Name:       dispatchName,
			ReturnType: firstSig.ReturnType,
			Parameters: params,
		}
		counter := 0
		for i, subName := range candidateSubs {
			subMangled := methodImplementationName(subName, methodName)
			subSig := signatures[subMangled]
			subArgs := append([]string(nil), forwardArgs...)
			if len(subArgs) > len(subSig.Parameters) {
				subArgs = subArgs[:len(subSig.Parameters)]
			}
			if i == len(candidateSubs)-1 {
				if firstSig.ReturnType == ir.TypeVoid {
					dispatchFn.Body = append(dispatchFn.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: subMangled,
						Args:   subArgs,
					})
					dispatchFn.Body = append(dispatchFn.Body, ir.Instruction{
						Op:   ir.OpReturn,
						Type: ir.TypeVoid,
					})
				} else {
					retVal := fmt.Sprintf("ret.%d", counter)
					counter++
					dispatchFn.Body = append(dispatchFn.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   firstSig.ReturnType,
						Result: retVal,
						Callee: subMangled,
						Args:   subArgs,
					})
					dispatchFn.Body = append(dispatchFn.Body, ir.Instruction{
						Op:   ir.OpReturn,
						Type: firstSig.ReturnType,
						Args: []string{retVal},
					})
				}
			} else {
				condVar := fmt.Sprintf("is.%s.%d", subName, counter)
				counter++
				dispatchFn.Body = append(dispatchFn.Body, ir.Instruction{
					Op:     ir.OpInstanceOf,
					Type:   ir.TypeBool,
					Result: condVar,
					Value:  subName,
					Args:   []string{"this"},
				})
				var thenBody []ir.Instruction
				if firstSig.ReturnType == ir.TypeVoid {
					thenBody = append(thenBody, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: subMangled,
						Args:   subArgs,
					})
					thenBody = append(thenBody, ir.Instruction{
						Op:   ir.OpReturn,
						Type: ir.TypeVoid,
					})
				} else {
					retVal := fmt.Sprintf("ret.%d", counter)
					counter++
					thenBody = append(thenBody, ir.Instruction{
						Op:     ir.OpCall,
						Type:   firstSig.ReturnType,
						Result: retVal,
						Callee: subMangled,
						Args:   subArgs,
					})
					thenBody = append(thenBody, ir.Instruction{
						Op:   ir.OpReturn,
						Type: firstSig.ReturnType,
						Args: []string{retVal},
					})
				}
				dispatchFn.Body = append(dispatchFn.Body, ir.Instruction{
					Op:   ir.OpIf,
					Type: ir.TypeVoid,
					Args: []string{condVar},
					Then: thenBody,
				})
			}
		}
		extraFunctions = append(extraFunctions, dispatchFn)
		signatures[dispatchName] = dispatchFn
		if defaults := defaultParamsIndex[methodImplementationName(candidateSubs[0], methodName)]; defaults != nil {
			defaultParamsIndex[dispatchName] = defaults
		}
		for _, subName := range candidateSubs {
			if restParamsIndex[methodImplementationName(subName, methodName)] {
				restParamsIndex[dispatchName] = true
				break
			}
		}
		return dispatchFn, dispatchName, true
	}
	if className != "" && className != "this" {
		for sigName, fn := range signatures {
			if (strings.HasPrefix(sigName, cleanCls+"_") || strings.HasPrefix(sigName, "Generator_")) && strings.HasSuffix(sigName, "_"+methodName+"_impl") {
				return fn, sigName, true
			}
		}
	} else {
		for sigName, fn := range signatures {
			if strings.HasSuffix(sigName, "_"+methodName+"_impl") {
				return fn, sigName, true
			}
		}
	}
	return ir.Function{}, "", false
}

func findGetterInHierarchy(className, propName string, signatures map[string]ir.Function, hierarchy map[string]ClassMeta) (ir.Function, string, bool) {
	curr := className
	for curr != "" {
		mangledDirect := curr + "_get_" + propName
		if fn, ok := signatures[mangledDirect]; ok {
			return fn, mangledDirect, true
		}
		if strings.Contains(curr, "<") && strings.HasSuffix(curr, ">") {
			mangledCls := mangleGenericTypeString(curr)
			mangled := mangledCls + "_get_" + propName
			if fn, ok := signatures[mangled]; ok {
				return fn, mangled, true
			}
		}
		cleanCurr := curr
		if idx := strings.Index(curr, "<"); idx != -1 {
			cleanCurr = curr[:idx]
		}
		mangled := cleanCurr + "_get_" + propName
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
	cleanBase := strings.Split(className, "__")[0]
	fields := getInheritedFields(cleanBase, hierarchy)

	// Class metadata has a separate token kind for class names and fields. The
	// runtime needs both: class names support instanceof, while field names map
	// directly to the storage order used by the lowering shape.
	var tag strings.Builder
	tag.WriteString("__class__|")
	writeToken := func(kind byte, value string) {
		if value == "" {
			return
		}
		fmt.Fprintf(&tag, "%c%d:%s|", kind, len([]byte(value)), value)
	}
	writeToken('c', className)
	if cleanBase != className {
		writeToken('b', cleanBase)
	}
	for _, field := range fields {
		writeToken('f', field.Name)
	}

	seenClasses := map[string]bool{className: true, cleanBase: true}
	curr := cleanBase
	for {
		meta, ok := hierarchy[curr]
		if !ok || meta.Extends == "" {
			break
		}
		for _, rawBase := range strings.Split(meta.Extends, ",") {
			base := strings.TrimSpace(rawBase)
			if base == "" {
				continue
			}
			if !seenClasses[base] {
				writeToken('b', base)
				seenClasses[base] = true
			}
			baseName := base
			if idx := strings.IndexAny(baseName, "<"); idx >= 0 {
				baseName = baseName[:idx]
			} else if idx := strings.Index(baseName, "__"); idx >= 0 {
				baseName = baseName[:idx]
			}
			if baseName != "" && !seenClasses[baseName] {
				writeToken('b', baseName)
				seenClasses[baseName] = true
			}
			curr = baseName
			break
		}
	}
	return strings.TrimSuffix(tag.String(), "|")
}

func getInheritanceDepth(className string, hierarchy map[string]ClassMeta) int {
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

func isSubclassOf(subClass, baseClass string, hierarchy map[string]ClassMeta) bool {
	if subClass == "" || baseClass == "" {
		return false
	}
	curr := subClass
	for curr != "" {
		if curr == baseClass {
			return true
		}
		meta, ok := hierarchy[curr]
		if !ok {
			break
		}
		curr = meta.Extends
	}
	return false
}

func methodImplementationName(className, methodName string) string {
	return className + "_" + methodName + "_impl"
}

func methodDispatcherName(className, methodName string) string {
	return className + "_" + methodName + "_dispatch"
}

func classSatisfies(className, target string, hierarchy map[string]ClassMeta) bool {
	if isSubclassOf(className, target, hierarchy) {
		return true
	}
	meta, ok := hierarchy[className]
	if !ok {
		return false
	}
	for _, implemented := range meta.Implements {
		if implemented == target || isSubtype(implemented, target) {
			return true
		}
	}
	return false
}

func synthesizePolymorphicDispatchers(hierarchy map[string]ClassMeta, signatures map[string]ir.Function) []ir.Function {
	var dispatchers []ir.Function
	type methodKey struct {
		baseClass  string
		methodName string
	}
	seen := map[methodKey]bool{}

	for baseClass := range hierarchy {
		if baseClass == "" {
			continue
		}
		stmtClass, ok := classSyntax[baseClass]
		if !ok {
			continue
		}
		for _, m := range stmtClass.Methods {
			if m.IsStatic || m.Kind != "method" {
				continue
			}
			key := methodKey{baseClass: baseClass, methodName: m.Name}
			if seen[key] {
				continue
			}
			seen[key] = true

			var implementors []string
			for candClass := range hierarchy {
				if classSatisfies(candClass, baseClass, hierarchy) {
					candMangled := methodImplementationName(candClass, m.Name)
					if _, exists := signatures[candMangled]; exists {
						if stmtCand, ok := classSyntax[candClass]; ok {
							for _, candM := range stmtCand.Methods {
								if candM.Name == m.Name && !candM.IsStatic && candM.Kind == "method" {
									implementors = append(implementors, candClass)
									break
								}
							}
						}
					}
				}
			}

			if len(implementors) > 0 {
				sort.Slice(implementors, func(i, j int) bool {
					return getInheritanceDepth(implementors[i], hierarchy) > getInheritanceDepth(implementors[j], hierarchy)
				})

				firstMangled := methodImplementationName(implementors[0], m.Name)
				templateSig := signatures[firstMangled]

				dispatchName := methodDispatcherName(baseClass, m.Name)

				params := make([]ir.Parameter, len(templateSig.Parameters))
				copy(params, templateSig.Parameters)
				if len(params) > 0 {
					// Dispatchers use the canonical receiver name expected by method bodies.
					params[0].Name = "this"
					params[0].Type = ir.Type("object:" + baseClass)
				}

				var body []ir.Instruction
				counter := 0
				callArgs := make([]string, len(params))
				for pIdx, p := range params {
					callArgs[pIdx] = p.Name
				}

				for i := 0; i < len(implementors)-1; i++ {
					implClass := implementors[i]
					implMangled := methodImplementationName(implClass, m.Name)
					isInstVar := fmt.Sprintf("is.%s.%d", implClass, counter)
					counter++
					body = append(body, ir.Instruction{
						Op:     ir.OpInstanceOf,
						Type:   ir.TypeBool,
						Result: isInstVar,
						Value:  implClass,
						Args:   []string{params[0].Name},
					})

					var thenBlock []ir.Instruction
					if templateSig.ReturnType == ir.TypeVoid {
						thenBlock = append(thenBlock, ir.Instruction{
							Op:     ir.OpCall,
							Type:   ir.TypeVoid,
							Callee: implMangled,
							Args:   callArgs,
						})
						thenBlock = append(thenBlock, ir.Instruction{
							Op:   ir.OpReturn,
							Type: ir.TypeVoid,
						})
					} else {
						resVar := fmt.Sprintf("ret.%s.%d", implClass, counter)
						counter++
						thenBlock = append(thenBlock, ir.Instruction{
							Op:     ir.OpCall,
							Type:   templateSig.ReturnType,
							Result: resVar,
							Callee: implMangled,
							Args:   callArgs,
						})
						thenBlock = append(thenBlock, ir.Instruction{
							Op:    ir.OpReturn,
							Type:  templateSig.ReturnType,
							Value: resVar,
							Args:  []string{resVar},
						})
					}

					body = append(body, ir.Instruction{
						Op:   ir.OpIf,
						Type: ir.TypeVoid,
						Args: []string{isInstVar},
						Then: thenBlock,
					})
				}

				lastImpl := implementors[len(implementors)-1]
				lastMangled := methodImplementationName(lastImpl, m.Name)
				if templateSig.ReturnType == ir.TypeVoid {
					body = append(body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: lastMangled,
						Args:   callArgs,
					})
					body = append(body, ir.Instruction{
						Op:   ir.OpReturn,
						Type: ir.TypeVoid,
					})
				} else {
					resVar := fmt.Sprintf("ret.fallback.%d", counter)
					body = append(body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   templateSig.ReturnType,
						Result: resVar,
						Callee: lastMangled,
						Args:   callArgs,
					})
					body = append(body, ir.Instruction{
						Op:    ir.OpReturn,
						Type:  templateSig.ReturnType,
						Value: resVar,
						Args:  []string{resVar},
					})
				}

				dispatchFn := ir.Function{
					Name:       dispatchName,
					Parameters: params,
					ReturnType: templateSig.ReturnType,
					Body:       body,
				}
				dispatchers = append(dispatchers, dispatchFn)
				signatures[dispatchName] = dispatchFn
				if defaults := defaultParamsIndex[firstMangled]; defaults != nil {
					defaultParamsIndex[dispatchName] = defaults
				}
				if restParamsIndex[firstMangled] {
					restParamsIndex[dispatchName] = true
				}
			}
		}
	}
	return dispatchers
}
