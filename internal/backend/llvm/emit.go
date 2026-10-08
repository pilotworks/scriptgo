// Package llvm emits LLVM IR from verified scriptgo IR.
package llvm

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

// Options controls deterministic LLVM artifact metadata.
type Options struct {
	CompilerVersion           string
	RuntimeABI                string
	Target                    string
	SourceHash                string
	Debug                     bool
	CompatibilityMode         string
	CompatibilityReportFormat int
	StaticSites               int
	DynamicSites              int
	UnsupportedSites          int
}

// Emit converts a verified module into LLVM IR using opaque pointers.
func Emit(module ir.Module) (string, error) {
	return EmitWithOptions(module, Options{
		CompilerVersion: "dev",
		RuntimeABI:      "scriptgo.runtime.v1",
		Target:          "native",
	})
}

// EmitWithOptions converts verified IR into LLVM IR with stable build metadata.
func EmitWithOptions(module ir.Module, options Options) (string, error) {
	if options.CompilerVersion == "" {
		options.CompilerVersion = "dev"
	}
	if options.RuntimeABI == "" {
		options.RuntimeABI = "scriptgo.runtime.v1"
	}
	if options.Target == "" {
		options.Target = "native"
	}
	var debug *debugInfo
	if options.Debug {
		debug = newDebugInfo(module)
	}
	if err := module.Verify(); err != nil {
		return "", err
	}
	functions := make(map[string]ir.Function, len(module.Functions))
	stringsByValue := map[string]string{}
	verStr := options.CompilerVersion
	if verStr == "" || verStr == "dev" {
		verStr = "v0.1.0"
	}
	if !strings.HasPrefix(verStr, "v") {
		verStr = "v" + verStr
	}
	if _, ok := stringsByValue[verStr]; !ok {
		stringsByValue[verStr] = fmt.Sprintf("@.str.%d", len(stringsByValue))
	}
	for _, dynamicModule := range module.DynamicModules {
		values := []string{dynamicModule.Path, dynamicModule.Source, dynamicModule.Kind, strings.Join(dynamicModule.Exports, "\x1f")}
		for _, dependency := range dynamicModule.Imports {
			values = append(values, dependency.Specifier, dependency.Path)
		}
		for _, value := range values {
			if _, ok := stringsByValue[value]; !ok {
				stringsByValue[value] = fmt.Sprintf("@.str.%d", len(stringsByValue))
			}
		}
	}
	var collectStrings func(list []ir.Instruction)
	collectStrings = func(list []ir.Instruction) {
		for _, instruction := range list {
			if instruction.Op == ir.OpDynamicCall {
				modulePath, exportName, _ := strings.Cut(instruction.Callee, "#")
				if _, ok := stringsByValue[modulePath]; !ok {
					stringsByValue[modulePath] = fmt.Sprintf("@.str.%d", len(stringsByValue))
				}
				if _, ok := stringsByValue[exportName]; !ok {
					stringsByValue[exportName] = fmt.Sprintf("@.str.%d", len(stringsByValue))
				}
			}
			if instruction.Op == ir.OpObjectNew {
				val := instruction.Value
				if val == "" {
					for _, s := range module.Shapes {
						if s.Name == instruction.Callee && len(s.Fields) > 0 {
							var names []string
							for _, f := range s.Fields {
								names = append(names, f.Name)
							}
							val = ":" + strings.Join(names, ":") + ":"
							break
						}
					}
				}
				if val == "" {
					val = instruction.Callee
				}
				if val != "" {
					if _, ok := stringsByValue[val]; !ok {
						stringsByValue[val] = fmt.Sprintf("@.str.%d", len(stringsByValue))
					}
				}
			} else if (instruction.Op == ir.OpConst && (instruction.Type == ir.TypeString || instruction.Type == ir.TypeSymbol)) || (instruction.Op == ir.OpInstanceOf && instruction.Value != "") {
				val := instruction.Value
				if val != "" {
					if _, ok := stringsByValue[val]; !ok {
						stringsByValue[val] = fmt.Sprintf("@.str.%d", len(stringsByValue))
					}
				}
			}
			if (instruction.Op == ir.OpFieldGet || instruction.Op == ir.OpFieldSet) && instruction.Field != "" {
				if _, ok := stringsByValue[instruction.Field]; !ok {
					stringsByValue[instruction.Field] = fmt.Sprintf("@.str.%d", len(stringsByValue))
				}
			}
			if instruction.Op == ir.OpCall {
				descriptor := intrinsicObjectDescriptor(instruction.Callee)
				if descriptor != "" {
					if _, ok := stringsByValue[descriptor]; !ok {
						stringsByValue[descriptor] = fmt.Sprintf("@.str.%d", len(stringsByValue))
					}
				}
			}
			if instruction.Op == ir.OpDebugger {
				pathStr := instruction.Span.Path
				if pathStr == "" {
					pathStr = module.SourcePath
				}
				if pathStr != "" {
					if _, ok := stringsByValue[pathStr]; !ok {
						stringsByValue[pathStr] = fmt.Sprintf("@.str.%d", len(stringsByValue))
					}
				}
			}
			collectStrings(instruction.Then)
			collectStrings(instruction.Else)
			collectStrings(instruction.Cond)
			collectStrings(instruction.Body)
			collectStrings(instruction.Catch)
			collectStrings(instruction.Finally)
		}
	}
	closureCallees := make(map[string]bool)
	for _, function := range module.Functions {
		collectClosureCallees(function.Body, closureCallees)
	}
	for _, function := range module.Functions {
		functions[function.Name] = function
		if _, ok := stringsByValue[function.Name]; !ok {
			stringsByValue[function.Name] = fmt.Sprintf("@.str.%d", len(stringsByValue))
		}
		collectStrings(function.Body)
	}
	for _, typeStr := range []string{"number", "string", "boolean", "bigint", "symbol", "function", "undefined", "object", "true", "false", "null", ""} {
		if _, ok := stringsByValue[typeStr]; !ok {
			stringsByValue[typeStr] = fmt.Sprintf("@.str.%d", len(stringsByValue))
		}
	}
	if _, ok := functions["main"]; !ok {
		return "", fmt.Errorf("module has no main function")
	}

	var out strings.Builder
	out.WriteString("; ModuleID = 'scriptgo'\n")
	out.WriteString(formatArtifactMetadata(options))
	out.WriteString("@scriptgo_undefined_sentinel = external global i8\n\n")
	out.WriteString("declare void @scriptgo_runtime_abort_if_failed(i32)\n")
	if moduleHasDynamic(module) {
		out.WriteString("declare i32 @scriptgo_dynamic_register_module(ptr, ptr, ptr, ptr)\n")
		out.WriteString("declare i32 @scriptgo_dynamic_register_dependency(ptr, ptr, ptr)\n")
		out.WriteString("declare i32 @scriptgo_dynamic_call_module(ptr, ptr, ptr, i32, i32, i32, ptr)\n")
		out.WriteString("declare i32 @scriptgo_dynamic_invoke_function(ptr, ptr, i32, ptr, i32, ptr)\n")
		out.WriteString("declare i32 @scriptgo_dynamic_new_proxy(ptr, ptr, ptr)\n")
		out.WriteString("declare void @scriptgo_dynamic_abort_if_failed(i32)\n")
	}
	out.WriteString("declare void @scriptgo_debugger_break(ptr, i32)\n\n")
	for _, method := range []string{"log", "info", "debug", "warn", "error"} {
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_number(double)\n", method))
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_bigint(i64)\n", method))
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_symbol(ptr)\n", method))
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_string(ptr)\n", method))
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_bool(i32)\n", method))
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_unknown(ptr)\n", method))
	}
	writeRuntimeDeclarations(&out)

	alreadyDeclared := map[string]bool{
		"malloc": true, "setjmp": true, "tan": true, "atan": true, "atan2": true, "hypot": true, "drand48": true,
		"scriptgo_math_round": true, "scriptgo_math_pow": true, "scriptgo_bigint_pow": true,
	}

	for _, ext := range module.Externs {
		if alreadyDeclared[ext.Name] || strings.HasPrefix(ext.Name, "llvm.") {
			continue
		}
		alreadyDeclared[ext.Name] = true
		retType := llvmType(ext.ReturnType)
		if ext.ReturnType == ir.TypeBool {
			retType = "zeroext i1"
		}
		var paramTypes []string
		for _, p := range ext.Parameters {
			pType := llvmType(p.Type)
			if p.Type == ir.TypeBool {
				pType = "zeroext i1"
			}
			paramTypes = append(paramTypes, pType)
		}
		out.WriteString(fmt.Sprintf("declare %s @%s(%s)\n", retType, mangleFunctionName(ext.Name), strings.Join(paramTypes, ", ")))
	}
	out.WriteString("\n")

	type strEntry struct {
		value string
		name  string
	}
	var strEntries []strEntry
	for value, name := range stringsByValue {
		strEntries = append(strEntries, strEntry{value: value, name: name})
	}
	sort.Slice(strEntries, func(i, j int) bool {
		return strEntries[i].name < strEntries[j].name
	})
	for _, entry := range strEntries {
		encoded := escapeString(entry.value)
		out.WriteString(fmt.Sprintf("%s = private unnamed_addr constant [%d x i8] c\"%s\\00\"\n", entry.name, len([]byte(entry.value))+1, encoded))
	}
	out.WriteString("\n")

	for _, g := range module.Globals {
		gType := llvmType(g.Type)
		initVal := "null"
		if gType == "double" {
			initVal = "0.0"
		} else if gType == "i1" {
			initVal = "false"
		} else if gType == "i32" || gType == "i64" {
			initVal = "0"
		} else if gType == "{ i32, i32, i64, i64 }" {
			initVal = "zeroinitializer"
		}
		if g.Value != "" {
			if g.Type == ir.TypeNumber {
				if num, err := strconv.ParseFloat(g.Value, 64); err == nil {
					initVal = llvmNumber(num)
				}
			} else if g.Type == ir.TypeBool {
				if g.Value == "true" {
					initVal = "true"
				} else {
					initVal = "false"
				}
			}
		}
		out.WriteString(fmt.Sprintf("@%s = global %s %s\n", g.Name, gType, initVal))
	}
	out.WriteString("\ndefine internal i32 @__scriptgo_to_int32(double %val) alwaysinline nounwind readnone willreturn {\nentry:\n  %abs = call double @llvm.fabs.f64(double %val)\n  %in_range = fcmp olt double %abs, 2147483648.0\n  br i1 %in_range, label %fast, label %slow\n\nfast:\n  %i32_fast = fptosi double %val to i32\n  ret i32 %i32_fast\n\nslow:\n  %i32_slow = call i32 @scriptgo_to_int32(double %val)\n  ret i32 %i32_slow\n}\n\n")

	loopMeta := &loopMetadataRecorder{}
	for _, function := range module.Functions {
		text, err := emitFunction(function, functions, stringsByValue, debug, module, options, loopMeta)
		if err != nil {
			return "", err
		}
		out.WriteString(text)
		if closureCallees[function.Name] {
			out.WriteString(emitClosureInvokeAdapter(function))
		}
	}
	for _, def := range loopMeta.definitions {
		out.WriteString(def + "\n")
	}
	if debug != nil {
		out.WriteString(debug.metadata(module, options.CompilerVersion))
	}
	return out.String(), nil
}
