package llvm

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func emitFunction(function ir.Function, functions map[string]ir.Function, stringsByValue map[string]string, debug *debugInfo, module ir.Module, options Options, loopMeta *loopMetadataRecorder) (string, error) {
	isClosure := strings.HasPrefix(function.Name, "__closure_")
	returnType := llvmType(function.ReturnType)
	if function.ReturnType == ir.TypeBool {
		returnType = "zeroext i1"
	}
	name := function.Name
	var out strings.Builder
	if name == "main" {
		out.WriteString("define i32 @main(i32 %argc, ptr %argv) nounwind \"frame-pointer\"=\"all\"")
	} else {
		out.WriteString(fmt.Sprintf("define internal %s %s(", returnType, functionSymbol(mangleFunctionName(name))))
		parameterIndex := 0
		for _, parameter := range function.Parameters {
			if parameterIndex > 0 {
				out.WriteString(", ")
			}
			if isRawCallbackParameter(parameter) {
				out.WriteString(fmt.Sprintf("i32 %%%s.tag, i32 %%%s.flags, i64 %%%s.payload", parameter.Name, parameter.Name, parameter.Name))
				parameterIndex += 3
				continue
			}
			out.WriteString(fmt.Sprintf("%s %%%s", llvmType(parameter.Type), parameter.Name))
			parameterIndex++
		}
		if strings.HasSuffix(name, "_constructor") && len(function.Body) <= 15 {
			out.WriteString(") alwaysinline nounwind \"frame-pointer\"=\"all\"")
		} else {
			out.WriteString(") nounwind \"frame-pointer\"=\"all\"")
		}
	}
	if debug != nil {
		fmt.Fprintf(&out, " !dbg !%d", debug.functions[function.Name])
	}
	out.WriteString(" {\n")
	if name == "main" {
		out.WriteString("  call void @scriptgo_process_init(i32 %argc, ptr %argv)\n")
		for _, g := range module.Globals {
			gType := llvmType(g.Type)
			if gType == "ptr" {
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_gc_add_root_slot(ptr @%s, i64 1)\n", g.Name))
			} else if gType == "{ i32, i32, i64, i64 }" {
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_gc_add_root_slot(ptr @%s, i64 3)\n", g.Name))
			}
		}
	}
	for _, parameter := range function.Parameters {
		if !isRawCallbackParameter(parameter) {
			continue
		}
		value := parameter.Name
		first := fmt.Sprintf("%s.box.0", value)
		second := fmt.Sprintf("%s.box.1", value)
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } zeroinitializer, i32 %%%s.tag, 0\n", first, value))
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i32 %%%s.flags, 1\n", second, first, value))
		out.WriteString(fmt.Sprintf("  %%%s = insertvalue { i32, i32, i64, i64 } %%%s, i64 %%%s.payload, 2\n", value, second, value))
	}

	verStr := options.CompilerVersion
	if verStr == "" || verStr == "dev" {
		verStr = "v0.1.0"
	}
	if !strings.HasPrefix(verStr, "v") {
		verStr = "v" + verStr
	}

	emitter := &functionEmitter{
		function:           function,
		functions:          functions,
		stringsByValue:     stringsByValue,
		debug:              debug,
		module:             module,
		compilerVersion:    verStr,
		target:             options.Target,
		types:              make(map[string]ir.Type, len(function.Parameters)+len(module.Globals)),
		varSlots:           make(map[string]string),
		localSSAs:          make(map[string]bool),
		hasTryCatch:        hasTryCatch(function.Body),
		hasArrayResize:     hasArrayResize(function.Body),
		integerVars:        make(map[string]bool),
		integerUpperBounds: make(map[string]float64),
		usedResults:        usedInstructionResults(function.Body),
		loopMeta:           loopMeta,
	}
	globalsMap := make(map[string]bool, len(module.Globals))
	for _, g := range module.Globals {
		emitter.types[g.Name] = g.Type
		globalsMap[g.Name] = true
	}
	for _, parameter := range function.Parameters {
		emitter.types[parameter.Name] = parameter.Type
		emitter.localSSAs[parameter.Name] = true
	}
	isLocalDecl := func(name string) bool {
		for _, l := range function.Locals {
			if l.Name == name {
				return true
			}
		}
		for _, p := range function.Parameters {
			if p.Name == name {
				return true
			}
		}
		return false
	}
	collectSSADefs(function.Body, emitter.localSSAs, function.Name == "main", globalsMap, isLocalDecl)

	if len(function.Captured) == 1 && function.Captured[0].Name == "this" {
		c := function.Captured[0]
		slot := "this.slot"
		out.WriteString(fmt.Sprintf("  %%%s = alloca ptr\n", slot))
		out.WriteString(fmt.Sprintf("  store%s ptr %%__env_ctx, ptr %%%s\n", emitter.vol(), slot))
		emitter.varSlots[c.Name] = slot
		emitter.types[c.Name] = c.Type
	} else if len(function.Captured) > 0 {
		fieldTypes := make([]string, len(function.Captured))
		for i := range function.Captured {
			fieldTypes[i] = "ptr"
		}
		structType := fmt.Sprintf("{ %s }", strings.Join(fieldTypes, ", "))
		for i, c := range function.Captured {
			fieldPtr := fmt.Sprintf("%s.field.%d", c.Name, i)
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds %s, ptr %%__env_ctx, i32 0, i32 %d\n", fieldPtr, structType, i))
			if c.Name == "this" {
				emitter.varSlots[c.Name] = fieldPtr
				emitter.types[c.Name] = c.Type
				continue
			}
			cellPtr := fmt.Sprintf("%s.cell.%d", c.Name, i)
			out.WriteString(fmt.Sprintf("  %%%s = load ptr, ptr %%%s\n", cellPtr, fieldPtr))
			emitter.varSlots[c.Name] = cellPtr
			emitter.types[c.Name] = c.Type
			if emitter.sharedEnvCells == nil {
				emitter.sharedEnvCells = make(map[string]string)
			}
			emitter.sharedEnvCells[c.Name] = cellPtr
		}
	}

	capturedInBody := findCapturedInFunction(function.Body)
	if emitter.sharedEnvCells == nil {
		emitter.sharedEnvCells = make(map[string]string)
	}
	for _, capName := range capturedInBody {
		if capName == "this" {
			continue
		}
		if _, alreadyCaptured := emitter.sharedEnvCells[capName]; alreadyCaptured {
			continue
		}
		cellSlot := fmt.Sprintf("cell.%s.%d", capName, emitter.loadCounter)
		emitter.loadCounter++
		allocSize := 8
		for _, param := range function.Parameters {
			if param.Name == capName {
				emitter.types[capName] = param.Type
				if param.Type == ir.TypeUnknown {
					allocSize = 24
				}
				break
			}
		}
		out.WriteString(fmt.Sprintf("  %%%s = call ptr @scriptgo_closure_alloc(i64 %d)\n", cellSlot, allocSize))
		emitter.sharedEnvCells[capName] = cellSlot
		for _, param := range function.Parameters {
			if param.Name == capName {
				pType := param.Type
				if pType == "" || pType == ir.TypeVoid {
					pType = ir.TypeUnknown
				}
				out.WriteString(fmt.Sprintf("  store volatile %s %%%s, ptr %%%s\n", llvmType(pType), param.Name, cellSlot))
				break
			}
		}
	}

	slotted := findSlottedVariables(function.Body)
	for _, varName := range slices.Sorted(maps.Keys(slotted)) {
		typ := slotted[varName]
		if function.Name == "main" {
			if globalsMap[varName] {
				continue
			}
		} else {
			if globalsMap[varName] && !isLocalDecl(varName) {
				continue
			}
		}
		for _, param := range function.Parameters {
			if param.Name == varName {
				typ = param.Type
				break
			}
		}
		if _, ok := emitter.varSlots[varName]; !ok {
			slotName := varName + ".slot"
			emitter.varSlots[varName] = slotName
			emitter.types[varName] = typ
			allocType := llvmType(typ)
			if allocType == "void" {
				allocType = "ptr"
			}
			out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", slotName, allocType))
			isParam := false
			vol := emitter.vol()
			for _, param := range function.Parameters {
				if param.Name == varName {
					out.WriteString(fmt.Sprintf("  store%s %s %%%s, ptr %%%s\n", vol, allocType, varName, slotName))
					isParam = true
					break
				}
			}
			if !isParam {
				switch allocType {
				case "{ i32, i32, i64, i64 }":
					out.WriteString(fmt.Sprintf("  store%s %s zeroinitializer, ptr %%%s\n", vol, allocType, slotName))
				case "i1":
					out.WriteString(fmt.Sprintf("  store%s i1 false, ptr %%%s\n", vol, slotName))
				case "double":
					out.WriteString(fmt.Sprintf("  store%s double 0.0, ptr %%%s\n", vol, slotName))
				case "i64":
					out.WriteString(fmt.Sprintf("  store%s i64 0, ptr %%%s\n", vol, slotName))
				case "i32":
					out.WriteString(fmt.Sprintf("  store%s i32 0, ptr %%%s\n", vol, slotName))
				default:
					out.WriteString(fmt.Sprintf("  store%s %s null, ptr %%%s\n", vol, allocType, slotName))
				}
			}
		}
	}
	out.WriteString("  %__slot_ptr = alloca ptr\n")
	out.WriteString("  %__slot_double = alloca double\n")
	out.WriteString("  %__slot_i32 = alloca i32\n")
	out.WriteString("  %__slot_i64 = alloca i64\n")
	out.WriteString("  %__slot_i1 = alloca i1\n")

	for _, instruction := range function.Body {
		if emitter.terminated {
			return "", fmt.Errorf("function %q contains instruction after return", function.Name)
		}
		if err := emitter.emitInstruction(&out, instruction); err != nil {
			return "", err
		}
	}

	if !emitter.terminated {
		if function.Name == "main" {
			// Async entrypoints run their synchronous prefix first; only then are
			// promise continuations drained in FIFO microtask order.
			out.WriteString("  call i32 @scriptgo_event_loop_run()\n")
			out.WriteString("  call i32 @scriptgo_timers_drain()\n")
			out.WriteString("  ret i32 0\n")
		} else if isClosure && (function.ReturnType == ir.TypeVoid || function.ReturnType == "") {
			out.WriteString("  ret void\n")
		} else if function.ReturnType == ir.TypeVoid {
			out.WriteString("  ret void\n")
		} else {
			switch function.ReturnType {
			case ir.TypeNumber:
				out.WriteString("  ret double 0.0\n")
			case ir.TypeBool:
				out.WriteString("  ret i1 false\n")
			default:
				out.WriteString("  ret ptr null\n")
			}
		}
	}
	out.WriteString("}\n\n")
	return hoistAllocas(out.String()), nil
}

func isRawCallbackParameter(parameter ir.Parameter) bool {
	return parameter.Type == ir.TypeUnknown &&
		(strings.HasSuffix(parameter.Name, "$raw") || parameter.Name == "__resume_raw")
}

func hoistAllocas(fnCode string) string {
	lines := strings.Split(fnCode, "\n")
	if len(lines) < 2 {
		return fnCode
	}
	var header string
	var allocas []string
	seenSlots := make(map[string]bool)
	var bodyLines []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i == 0 {
			header = line
			continue
		}
		if trimmed == "}" || (i == len(lines)-1 && trimmed == "") {
			continue
		}
		if strings.Contains(trimmed, " = alloca ") {
			parts := strings.SplitN(trimmed, "=", 2)
			slotName := strings.TrimSpace(parts[0])
			if !seenSlots[slotName] {
				seenSlots[slotName] = true
				allocas = append(allocas, "  "+trimmed)
			}
			continue
		}
		bodyLines = append(bodyLines, line)
	}

	var res strings.Builder
	res.WriteString(header)
	res.WriteString("\n")
	for _, a := range allocas {
		res.WriteString(a)
		res.WriteString("\n")
	}
	for _, b := range bodyLines {
		res.WriteString(b)
		res.WriteString("\n")
	}
	res.WriteString("}\n\n")
	return res.String()
}

func findSlottedVariables(instructions []ir.Instruction) map[string]ir.Type {
	slotted := make(map[string]ir.Type)
	counts := make(map[string]int)
	types := make(map[string]ir.Type)
	typeSets := make(map[string]map[ir.Type]bool)
	var scan func(list []ir.Instruction)
	scan = func(list []ir.Instruction) {
		for _, inst := range list {
			if inst.Op == ir.OpAssign {
				slotted[inst.Result] = inst.Type
			}
			if inst.CatchVar != "" {
				slotted[inst.CatchVar] = ir.TypeString
			}
			if inst.Result != "" {
				counts[inst.Result]++
				if inst.Type != "" {
					types[inst.Result] = inst.Type
					if typeSets[inst.Result] == nil {
						typeSets[inst.Result] = make(map[ir.Type]bool)
					}
					typeSets[inst.Result][inst.Type] = true
				}
			}
			scan(inst.Then)
			scan(inst.Else)
			scan(inst.Cond)
			scan(inst.Body)
			scan(inst.Step)
			scan(inst.Catch)
			scan(inst.Finally)
		}
	}
	scan(instructions)
	for name, count := range counts {
		if count > 1 {
			if typ, ok := types[name]; ok {
				// A variable that is assigned both a boxed value and a known
				// value must keep boxed storage until the checked cast runs.
				// Using the last (known) type here can allocate only 8 bytes for
				// a 24-byte unknown value and corrupt the stack.
				if typeSets[name][ir.TypeUnknown] {
					slotted[name] = ir.TypeUnknown
				} else {
					slotted[name] = typ
				}
			}
		}
	}
	return slotted
}

func findCapturedInFunction(instructions []ir.Instruction) []string {
	var captured []string
	seen := make(map[string]bool)
	var scan func(list []ir.Instruction)
	scan = func(list []ir.Instruction) {
		for _, inst := range list {
			if inst.Op == ir.OpClosure {
				for _, arg := range inst.Args {
					if !seen[arg] {
						seen[arg] = true
						captured = append(captured, arg)
					}
				}
			}
			scan(inst.Then)
			scan(inst.Else)
			scan(inst.Cond)
			scan(inst.Body)
			scan(inst.Step)
			scan(inst.Catch)
			scan(inst.Finally)
		}
	}
	scan(instructions)
	return captured
}

func collectSSADefs(instructions []ir.Instruction, defs map[string]bool, isMain bool, globals map[string]bool, isLocalDecl func(string) bool) {
	for _, inst := range instructions {
		if inst.Result != "" {
			if isMain {
				if !globals[inst.Result] {
					defs[inst.Result] = true
				}
			} else {
				if !globals[inst.Result] || isLocalDecl(inst.Result) {
					defs[inst.Result] = true
				}
			}
		}
		collectSSADefs(inst.Then, defs, isMain, globals, isLocalDecl)
		collectSSADefs(inst.Else, defs, isMain, globals, isLocalDecl)
		collectSSADefs(inst.Cond, defs, isMain, globals, isLocalDecl)
		collectSSADefs(inst.Body, defs, isMain, globals, isLocalDecl)
		collectSSADefs(inst.Step, defs, isMain, globals, isLocalDecl)
		collectSSADefs(inst.Catch, defs, isMain, globals, isLocalDecl)
		collectSSADefs(inst.Finally, defs, isMain, globals, isLocalDecl)
	}
}
