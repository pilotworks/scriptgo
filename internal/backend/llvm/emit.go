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
			if (instruction.Op == ir.OpConst && instruction.Type == ir.TypeString) || (instruction.Op == ir.OpObjectNew && instruction.Value != "") || (instruction.Op == ir.OpInstanceOf && instruction.Value != "") {
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
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_unknown({ i32, i32, i64 })\n", method))
	}
	out.WriteString("declare void @__scriptgo_fail_checked_cast(i32, i32, ptr)\n")
	out.WriteString("declare ptr @__scriptgo_typeof_unknown(i32)\n\n")
	out.WriteString("declare double @llvm.fabs.f64(double)\n")
	out.WriteString("declare double @llvm.ceil.f64(double)\n")
	out.WriteString("declare double @llvm.floor.f64(double)\n")
	out.WriteString("declare double @llvm.trunc.f64(double)\n")
	out.WriteString("declare double @llvm.sqrt.f64(double)\n")
	out.WriteString("declare double @llvm.round.f64(double)\n")
	out.WriteString("declare double @llvm.sin.f64(double)\n")
	out.WriteString("declare double @llvm.cos.f64(double)\n")
	out.WriteString("declare double @llvm.log.f64(double)\n")
	out.WriteString("declare double @llvm.log2.f64(double)\n")
	out.WriteString("declare double @llvm.log10.f64(double)\n")
	out.WriteString("declare double @llvm.exp.f64(double)\n")
	out.WriteString("declare double @tan(double)\n")
	out.WriteString("declare double @atan(double)\n")
	out.WriteString("declare double @atan2(double, double)\n")
	out.WriteString("declare double @hypot(double, double)\n")
	out.WriteString("declare double @drand48()\n")
	out.WriteString("declare double @llvm.minnum.f64(double, double)\n")
	out.WriteString("declare double @llvm.maxnum.f64(double, double)\n")
	out.WriteString("declare double @llvm.pow.f64(double, double)\n\n")
	out.WriteString("declare i32 @scriptgo_number_parse_int(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_parse_float(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_is_nan(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_is_finite(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_is_integer(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_to_fixed(double, double, ptr)\n\n")
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
	out.WriteString("declare i32 @scriptgo_array_at(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_shift(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_unshift(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_reverse(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_concat(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_splice(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_join_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_join_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_release(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_object_new(i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_number_set(ptr, i64, double)\n")
	out.WriteString("declare i32 @scriptgo_object_number_get(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_string_set(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_string_get(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_bool_set(ptr, i64, i32)\n")
	out.WriteString("declare i32 @scriptgo_object_bool_get(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_ptr_set(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_ptr_get(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_type_set(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_type_get(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_instanceof(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_release(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_json_stringify_number(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_json_stringify_bool(i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_json_stringify_string(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_json_stringify_number_array(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_json_stringify_string_array(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_json_parse_string(ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_string_concat(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_length(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_index_of(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_last_index(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_starts_with(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_ends_with(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_char_at(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_char_code_at(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_includes(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_to_lower(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_to_upper(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_trim_start(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_trim_end(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_repeat(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_replace_all(ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_pad_start(ptr, double, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_pad_end(ptr, double, ptr, ptr)\n")
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
	out.WriteString("declare i32 @scriptgo_closure_create(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_map_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_map_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_filter_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_filter_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_for_each_number(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_for_each_string(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_reduce_number(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_find_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_some_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_every_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_queue_microtask(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_promise_create(ptr)\n")
	out.WriteString("declare i32 @scriptgo_promise_resolve(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_promise_resolve_number(ptr, double)\n")
	out.WriteString("declare i32 @scriptgo_promise_then(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_promise_await_number(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_event_loop_run()\n\n")
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

	if len(function.Captured) > 0 {
		fieldTypes := make([]string, len(function.Captured))
		for i, c := range function.Captured {
			fieldTypes[i] = llvmType(c.Type)
		}
		structType := fmt.Sprintf("{ %s }", strings.Join(fieldTypes, ", "))
		for i, c := range function.Captured {
			slotName := c.Name + ".slot"
			emitter.varSlots[c.Name] = slotName
			emitter.types[c.Name] = c.Type
			out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", slotName, llvmType(c.Type)))
			fieldPtr := fmt.Sprintf("%s.field.%d", c.Name, i)
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds %s, ptr %%__env_ctx, i32 0, i32 %d\n", fieldPtr, structType, i))
			loadedVal := fmt.Sprintf("%s.val.%d", c.Name, i)
			out.WriteString(fmt.Sprintf("  %%%s = load %s, ptr %%%s\n", loadedVal, llvmType(c.Type), fieldPtr))
			out.WriteString(fmt.Sprintf("  store %s %%%s, ptr %%%s\n", llvmType(c.Type), loadedVal, slotName))
		}
	}

	slotted := findSlottedVariables(function.Body)
	for varName, typ := range slotted {
		if _, ok := emitter.varSlots[varName]; !ok {
			slotName := varName + ".slot"
			emitter.varSlots[varName] = slotName
			emitter.types[varName] = typ
			out.WriteString(fmt.Sprintf("  %%%s = alloca %s\n", slotName, llvmType(typ)))
		}
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
			out.WriteString("  call i32 @scriptgo_event_loop_run()\n")
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
			scan(inst.Step)
			scan(inst.Catch)
			scan(inst.Finally)
		}
	}
	scan(instructions)
	return slotted
}
