package lowering

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

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
		for _, cls := range slices.Sorted(maps.Keys(hierarchy)) {
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
		cleanCurr := cleanGenericBase(curr)
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
	cleanCls := cleanGenericBase(className)
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
	for _, subName := range slices.Sorted(maps.Keys(hierarchy)) {
		meta := hierarchy[subName]
		isExact := false
		isLoose := false
		for _, imp := range meta.Implements {
			impClean := cleanGenericBase(imp)
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
		for _, sigName := range slices.Sorted(maps.Keys(signatures)) {
			fn := signatures[sigName]
			if (strings.HasPrefix(sigName, cleanCls+"_") || strings.HasPrefix(sigName, "Generator_") || strings.Contains(sigName, "_"+cleanCls+"_")) && strings.HasSuffix(sigName, "_"+methodName+"_impl") {
				return fn, sigName, true
			}
		}
		if candidates, hasCands := classCandidates[cleanCls]; hasCands {
			for _, cand := range candidates {
				if cand.Internal != "" && cand.Internal != className {
					if fn, mangled, ok := findMethodInHierarchy(cand.Internal, methodName, signatures, hierarchy); ok {
						return fn, mangled, true
					}
				}
			}
		}
	} else {
		for _, sigName := range slices.Sorted(maps.Keys(signatures)) {
			fn := signatures[sigName]
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

// staticFieldGlobal names the module global that stores a static field.
// Global names also become parts of LLVM local identifiers, which allow only
// letters, digits, $, ., and _, so any other byte of the field name (the "#"
// of a private name, non-ASCII letters) is spelled as $XX.
func staticFieldGlobal(className, field string) string {
	var b strings.Builder
	b.WriteString(className)
	b.WriteByte('_')
	for i := 0; i < len(field); i++ {
		c := field[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "$%02X", c)
		}
	}
	return b.String()
}
