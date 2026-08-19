// Package llvm emits LLVM IR from verified scriptgo IR.
package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

// Options controls deterministic LLVM artifact metadata.
type Options struct {
	CompilerVersion string
	RuntimeABI      string
	Target          string
	SourceHash      string
	Debug           bool
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
	var collectStrings func(list []ir.Instruction)
	collectStrings = func(list []ir.Instruction) {
		for _, instruction := range list {
			if instruction.Op == ir.OpConst && instruction.Type == ir.TypeString {
				if _, ok := stringsByValue[instruction.Value]; !ok {
					stringsByValue[instruction.Value] = fmt.Sprintf("@.str.%d", len(stringsByValue))
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
	for _, function := range module.Functions {
		functions[function.Name] = function
		collectStrings(function.Body)
	}
	if _, ok := functions["main"]; !ok {
		return "", fmt.Errorf("module has no main function")
	}

	var out strings.Builder
	out.WriteString("; ModuleID = 'scriptgo'\n")
	fmt.Fprintf(&out, "; scriptgo.compiler = %q\n", options.CompilerVersion)
	fmt.Fprintf(&out, "; scriptgo.runtime-abi = %q\n", options.RuntimeABI)
	fmt.Fprintf(&out, "; scriptgo.target = %q\n", options.Target)
	if options.SourceHash != "" {
		fmt.Fprintf(&out, "; scriptgo.source-sha256 = %q\n", options.SourceHash)
	}
	out.WriteString("declare void @scriptgo_runtime_abort_if_failed(i32)\n\n")
	for _, method := range []string{"log", "info", "warn", "error"} {
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_number(double)\n", method))
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_string(ptr)\n", method))
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_bool(i32)\n", method))
	}
	out.WriteString("\n")
	out.WriteString("declare double @llvm.fabs.f64(double)\n")
	out.WriteString("declare double @llvm.ceil.f64(double)\n")
	out.WriteString("declare double @llvm.floor.f64(double)\n")
	out.WriteString("declare double @llvm.trunc.f64(double)\n")
	out.WriteString("declare double @llvm.sqrt.f64(double)\n")
	out.WriteString("declare double @llvm.round.f64(double)\n")
	out.WriteString("declare double @llvm.sin.f64(double)\n")
	out.WriteString("declare double @llvm.cos.f64(double)\n")
	out.WriteString("declare double @llvm.minnum.f64(double, double)\n")
	out.WriteString("declare double @llvm.maxnum.f64(double, double)\n")
	out.WriteString("declare double @llvm.pow.f64(double, double)\n\n")
	out.WriteString("declare i32 @scriptgo_array_new(i64, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_get(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_set(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_length(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_push(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_pop(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_slice(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_index_of_number(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_index_of_string(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_includes_number(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_includes_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_release(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_object_new(i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_number_set(ptr, i64, double)\n")
	out.WriteString("declare i32 @scriptgo_object_number_get(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_string_set(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_string_get(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_release(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_string_concat(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_length(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_index_of(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_last_index(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_starts_with(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_ends_with(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_from_number(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_from_bool(i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_slice(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_trim(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_replace(ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_split(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_release(ptr)\n")
	out.WriteString("declare i32 @strcmp(ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_fs_read_file_sync(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_fs_write_file_sync(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_fs_exists_sync(ptr, ptr)\n\n")
	out.WriteString("declare void @scriptgo_process_init(i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_process_exit(double)\n")
	out.WriteString("declare i32 @scriptgo_process_cwd(ptr)\n")
	out.WriteString("declare i32 @scriptgo_process_argv(ptr)\n")
	out.WriteString("declare i32 @scriptgo_process_env(ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_crypto_random_uuid(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_web_btoa(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_web_atob(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_performance_now(ptr)\n\n")
	out.WriteString("declare void @scriptgo_exception_push(ptr)\n")
	out.WriteString("declare void @scriptgo_exception_pop(ptr)\n")
	out.WriteString("declare ptr @scriptgo_exception_buf(ptr)\n")
	out.WriteString("declare i32 @setjmp(ptr) returns_twice\n")
	out.WriteString("declare void @scriptgo_throw_string(ptr)\n")
	out.WriteString("declare void @scriptgo_throw_number(double)\n")
	out.WriteString("declare void @scriptgo_throw_bool(i32)\n")
	out.WriteString("declare ptr @scriptgo_exception_get_string(ptr)\n")
	out.WriteString("declare double @scriptgo_exception_get_number(ptr)\n")
	out.WriteString("declare i32 @scriptgo_exception_get_bool(ptr)\n")
	out.WriteString("declare void @scriptgo_exception_rethrow(ptr)\n\n")
	for value, name := range stringsByValue {
		encoded := escapeString(value)
		out.WriteString(fmt.Sprintf("%s = private unnamed_addr constant [%d x i8] c\"%s\\00\"\n", name, len([]byte(value))+1, encoded))
	}
	out.WriteString("\n")

	for _, function := range module.Functions {
		text, err := emitFunction(function, functions, stringsByValue, debug)
		if err != nil {
			return "", err
		}
		out.WriteString(text)
	}
	if debug != nil {
		out.WriteString(debug.metadata(module, options.CompilerVersion))
	}
	return out.String(), nil
}

func emitFunction(function ir.Function, functions map[string]ir.Function, stringsByValue map[string]string, debug *debugInfo) (string, error) {
	returnType := llvmType(function.ReturnType)
	name := function.Name
	var out strings.Builder
	if name == "main" {
		out.WriteString("define i32 @main(i32 %argc, ptr %argv)")
	} else {
		out.WriteString(fmt.Sprintf("define %s @%s(", returnType, name))
		for index, parameter := range function.Parameters {
			if index > 0 {
				out.WriteString(", ")
			}
			out.WriteString(fmt.Sprintf("%s %%%s", llvmType(parameter.Type), parameter.Name))
		}
		out.WriteString(")")
	}
	if debug != nil {
		fmt.Fprintf(&out, " !dbg !%d", debug.functions[function.Name])
	}
	out.WriteString(" {\n")
	if name == "main" {
		out.WriteString("  call void @scriptgo_process_init(i32 %argc, ptr %argv)\n")
	}


	emitter := &functionEmitter{
		function:       function,
		functions:      functions,
		stringsByValue: stringsByValue,
		debug:          debug,
		types:          make(map[string]ir.Type, len(function.Parameters)),
		varSlots:       make(map[string]string),
	}
	for _, parameter := range function.Parameters {
		emitter.types[parameter.Name] = parameter.Type
	}

	slotted := findSlottedVariables(function.Body)
	for varName, typ := range slotted {
		slotName := varName + ".slot"
		emitter.varSlots[varName] = slotName
		emitter.types[varName] = typ
		out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", slotName, llvmType(typ)))
	}

	for _, instruction := range function.Body {
		if emitter.terminated {
			return "", fmt.Errorf("function %q contains instruction after return", function.Name)
		}
		if err := emitter.emitInstruction(&out, instruction); err != nil {
			return "", err
		}
	}

	if !emitter.terminated {
		for _, arrayRef := range emitter.arrayTypes {
			out.WriteString(fmt.Sprintf("  call i32 @scriptgo_array_release(ptr %%%s)\n", arrayRef.name))
		}
		for _, object := range emitter.objects {
			out.WriteString(fmt.Sprintf("  call i32 @scriptgo_object_release(ptr %%%s)\n", object))
		}
		if function.Name == "main" {
			for _, value := range emitter.ownedStrings {
				out.WriteString(fmt.Sprintf("  call i32 @scriptgo_string_release(ptr %%%s)\n", value))
			}
		}
		if function.Name == "main" {
			out.WriteString("  ret i32 0\n")
		} else if function.ReturnType == ir.TypeVoid {
			out.WriteString("  ret void\n")
		} else {
			return "", fmt.Errorf("function %q has no return", function.Name)
		}
	}
	out.WriteString("}\n\n")
	return out.String(), nil
}

func findSlottedVariables(instructions []ir.Instruction) map[string]ir.Type {
	slotted := make(map[string]ir.Type)
	var scan func(list []ir.Instruction)
	scan = func(list []ir.Instruction) {
		for _, inst := range list {
			if inst.Op == ir.OpAssign {
				slotted[inst.Result] = inst.Type
			}
			scan(inst.Then)
			scan(inst.Else)
			scan(inst.Cond)
			scan(inst.Body)
		}
	}
	scan(instructions)
	return slotted
}
