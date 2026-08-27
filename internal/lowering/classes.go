package lowering

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
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
				if existing, exists := hierarchy[stmt.Class.Name]; exists {
					if stmt.Kind == "interface" {
						existing.Fields = append(existing.Fields, stmt.Class.Fields...)
						hierarchy[stmt.Class.Name] = existing
						continue
					}
					if len(existing.Fields) > 0 && len(stmt.Class.Fields) == 0 {
						continue
					}
				}
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

func isSubtype(subType, superType string) bool {
	sub := strings.TrimPrefix(subType, "object:")
	super := strings.TrimPrefix(superType, "object:")
	if sub == super {
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

func findMethodInHierarchy(className, methodName string, signatures map[string]ir.Function, hierarchy map[string]ClassMeta) (ir.Function, string, bool) {
	if className == "this" || className == "" {
		for cls := range hierarchy {
			if fn, ok := signatures[cls+"_"+methodName]; ok {
				return fn, cls + "_" + methodName, true
			}
		}
	}
	curr := className
	for curr != "" {
		mangledDirect := curr + "_" + methodName
		if fn, ok := signatures[mangledDirect]; ok {
			return fn, mangledDirect, true
		}
		if strings.Contains(curr, "<") && strings.HasSuffix(curr, ">") {
			mangledCls := mangleGenericTypeString(curr)
			mangled := mangledCls + "_" + methodName
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
		mangled := cleanCurr + "_" + methodName
		if fn, ok := signatures[mangled]; ok {
			return fn, mangled, true
		}
		genMangled := "Generator_" + cleanCurr + "_" + methodName
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
		if !isExact && !isLoose && (meta.Extends == className || meta.Extends == cleanCls || normalizeGenericName(meta.Extends) == normClass) {
			if meta.Extends == className || normalizeGenericName(meta.Extends) == normClass {
				isExact = true
			} else {
				isLoose = true
			}
		}
		subMangled := subName + "_" + methodName
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
		subMangled := candidateSubs[0] + "_" + methodName
		return signatures[subMangled], subMangled, true
	} else if len(candidateSubs) > 1 {
		slices.Sort(candidateSubs)
		dispatchName := cleanCls + "_" + methodName
		if fn, ok := signatures[dispatchName]; ok {
			return fn, dispatchName, true
		}
		firstSig := signatures[candidateSubs[0]+"_"+methodName]
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
			subMangled := subName + "_" + methodName
			if i == len(candidateSubs)-1 {
				if firstSig.ReturnType == ir.TypeVoid {
					dispatchFn.Body = append(dispatchFn.Body, ir.Instruction{
						Op:     ir.OpCall,
						Type:   ir.TypeVoid,
						Callee: subMangled,
						Args:   forwardArgs,
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
						Args:   forwardArgs,
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
						Args:   forwardArgs,
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
						Args:   forwardArgs,
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
		return dispatchFn, dispatchName, true
	}
	if className != "" && className != "this" {
		for sigName, fn := range signatures {
			if (strings.HasPrefix(sigName, cleanCls+"_") || strings.HasPrefix(sigName, "Generator_")) && strings.HasSuffix(sigName, "_"+methodName) {
				return fn, sigName, true
			}
		}
	} else {
		for sigName, fn := range signatures {
			if strings.HasSuffix(sigName, "_"+methodName) {
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
	chain := []string{className}
	curr := className
	for {
		meta, ok := hierarchy[curr]
		if ok {
			for _, f := range meta.Fields {
				if f.Name != "" {
					chain = append(chain, f.Name)
				}
			}
			if cls, okCls := classSyntax[curr]; okCls {
				for _, m := range cls.Methods {
					if m.Name != "" {
						chain = append(chain, m.Name)
					}
				}
			}
		}
		if !ok || meta.Extends == "" {
			break
		}
		chain = append(chain, meta.Extends)
		curr = meta.Extends
	}
	return ":" + strings.Join(chain, ":") + ":"
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

func synthesizePolymorphicDispatchers(hierarchy map[string]ClassMeta, signatures map[string]ir.Function) []ir.Function {
	var dispatchers []ir.Function
	type methodKey struct {
		baseClass  string
		methodName string
	}
	seen := map[methodKey]bool{}

	for baseClass := range hierarchy {
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

			hasOwnConcrete := false
			for _, baseM := range stmtClass.Methods {
				if baseM.Name == m.Name && !baseM.IsStatic && baseM.Kind == "method" && !baseM.IsAbstract && len(baseM.Body) > 0 {
					hasOwnConcrete = true
					break
				}
			}
			if hasOwnConcrete {
				continue
			}

			var implementors []string
			for candClass := range hierarchy {
				if isSubclassOf(candClass, baseClass, hierarchy) {
					candMangled := candClass + "_" + m.Name
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

			if len(implementors) > 1 {
				sort.Slice(implementors, func(i, j int) bool {
					return getInheritanceDepth(implementors[i], hierarchy) > getInheritanceDepth(implementors[j], hierarchy)
				})

				firstMangled := implementors[0] + "_" + m.Name
				templateSig := signatures[firstMangled]

				dispatchName := baseClass + "_" + m.Name

				params := make([]ir.Parameter, len(templateSig.Parameters))
				copy(params, templateSig.Parameters)
				if len(params) > 0 {
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
					implMangled := implClass + "_" + m.Name
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
				lastMangled := lastImpl + "_" + m.Name
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
			}
		}
	}
	return dispatchers
}
