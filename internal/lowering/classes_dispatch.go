package lowering

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func getHierarchyTag(className string, hierarchy map[string]ClassMeta) string {
	if className == "" {
		return ""
	}
	cleanBase := cleanGenericBase(className)
	fields := getInheritedFields(cleanBase, hierarchy)
	if len(fields) == 0 {
		fields = getInheritedFields(className, hierarchy)
	}

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
			// Built-in error constructors are not user classes; their
			// prototype chain still reaches Error for instanceof.
			if base, builtin := builtinErrorBase[curr]; builtin && !seenClasses[base] {
				writeToken('b', base)
				seenClasses[base] = true
				curr = base
				continue
			}
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
			baseName := cleanGenericBase(base)
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

	for _, baseClass := range slices.Sorted(maps.Keys(hierarchy)) {
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
			for _, candClass := range slices.Sorted(maps.Keys(hierarchy)) {
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

			if len(implementors) > 1 {
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

// builtinErrorBase maps the ECMAScript NativeError constructors (and
// AggregateError) to their prototype parent, so instances of the built-in
// errors and of user classes extending them satisfy instanceof Error.
var builtinErrorBase = map[string]string{
	"AggregateError": "Error",
	"EvalError":      "Error",
	"RangeError":     "Error",
	"ReferenceError": "Error",
	"SyntaxError":    "Error",
	"TypeError":      "Error",
	"URIError":       "Error",
}
