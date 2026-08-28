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
	var collectStrings func(list []ir.Instruction)
	collectStrings = func(list []ir.Instruction) {
		for _, instruction := range list {
			if (instruction.Op == ir.OpConst && (instruction.Type == ir.TypeString || instruction.Type == ir.TypeSymbol)) || (instruction.Op == ir.OpObjectNew && (instruction.Value != "" || instruction.Callee != "")) || (instruction.Op == ir.OpInstanceOf && instruction.Value != "") {
				val := instruction.Value
				if val == "" {
					val = instruction.Callee
				}
				if val != "" {
					if _, ok := stringsByValue[val]; !ok {
						stringsByValue[val] = fmt.Sprintf("@.str.%d", len(stringsByValue))
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
	fmt.Fprintf(&out, "; scriptgo.compiler = %q\n", options.CompilerVersion)
	fmt.Fprintf(&out, "; scriptgo.runtime-abi = %q\n", options.RuntimeABI)
	fmt.Fprintf(&out, "; scriptgo.target = %q\n", options.Target)
	if options.SourceHash != "" {
		fmt.Fprintf(&out, "; scriptgo.source-sha256 = %q\n", options.SourceHash)
	}
	out.WriteString("declare void @scriptgo_runtime_abort_if_failed(i32)\n")
	out.WriteString("declare void @scriptgo_debugger_break(ptr, i32)\n\n")
	for _, method := range []string{"log", "info", "debug", "warn", "error"} {
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_number(double)\n", method))
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_bigint(i64)\n", method))
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_symbol(ptr)\n", method))
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_string(ptr)\n", method))
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_bool(i32)\n", method))
		out.WriteString(fmt.Sprintf("declare i32 @scriptgo_console_%s_unknown(i32, i32, i64)\n", method))
	}
	out.WriteString("declare i32 @scriptgo_console_clear()\n")
	out.WriteString("declare i32 @scriptgo_console_group()\n")
	out.WriteString("declare i32 @scriptgo_console_group_end()\n")
	out.WriteString("declare i32 @scriptgo_console_count(ptr)\n")
	out.WriteString("declare i32 @scriptgo_console_count_reset(ptr)\n")
	out.WriteString("declare i32 @scriptgo_console_time(ptr)\n")
	out.WriteString("declare i32 @scriptgo_console_time_log(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_console_time_end(ptr)\n")
	out.WriteString("declare i32 @scriptgo_console_trace(ptr)\n\n")
	out.WriteString("declare void @__scriptgo_fail_checked_cast(i32, i32, ptr)\n")
	out.WriteString("declare ptr @__scriptgo_typeof_unknown(i32)\n")
	out.WriteString("declare i32 @scriptgo_is_truthy_unknown(i32, i64)\n")
	out.WriteString("declare i32 @scriptgo_closure_equals(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_from_unknown(i32, i32, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_from_object(ptr, ptr)\n\n")
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
	out.WriteString("declare double @asin(double)\n")
	out.WriteString("declare double @acos(double)\n")
	out.WriteString("declare double @cbrt(double)\n")
	out.WriteString("declare double @sinh(double)\n")
	out.WriteString("declare double @cosh(double)\n")
	out.WriteString("declare double @tanh(double)\n")
	out.WriteString("declare double @asinh(double)\n")
	out.WriteString("declare double @acosh(double)\n")
	out.WriteString("declare double @atanh(double)\n")
	out.WriteString("declare double @expm1(double)\n")
	out.WriteString("declare double @log1p(double)\n")
	out.WriteString("declare double @atan2(double, double)\n")
	out.WriteString("declare double @hypot(double, double)\n")
	out.WriteString("declare double @drand48()\n")
	out.WriteString("declare double @llvm.minnum.f64(double, double)\n")
	out.WriteString("declare double @llvm.maxnum.f64(double, double)\n")
	out.WriteString("declare double @llvm.pow.f64(double, double)\n")
	out.WriteString("declare i32 @llvm.ctlz.i32(i32, i1)\n\n")
	out.WriteString("declare i32 @scriptgo_number_parse_int(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_parse_float(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_is_nan(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_is_finite(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_is_integer(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_is_safe_integer(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_to_fixed(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_to_string(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_to_exponential(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_to_precision(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_number_to_locale_string(double, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_array_new(i64, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_set_tag(ptr, i64)\n")
	out.WriteString("declare i32 @scriptgo_array_get(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_set(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_length(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_set_length(ptr, double)\n")
	out.WriteString("declare i32 @scriptgo_array_push(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_pop(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_slice(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_index_of_number(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_index_of_string(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_index_of_ptr(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_includes_number(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_includes_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_includes_ptr(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_at(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_shift(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_unshift(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_reverse(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_concat(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_splice(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_join_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_join_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_join_bigint(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_join_unknown(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_keys(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_entries(ptr, ptr)\n")
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
	out.WriteString("declare i32 @scriptgo_object_unknown_set(ptr, i64, i32, i64)\n")
	out.WriteString("declare i32 @scriptgo_object_unknown_get(ptr, i64, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_type_set(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_type_get(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_instanceof(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_is_number(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_is_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_is_ptr(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_is_unknown(i32, i64, i32, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_keys(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_object_release(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_json_stringify_number(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_json_stringify_bool(i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_json_stringify_string(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_json_stringify_number_array(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_json_stringify_string_array(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_json_stringify_unknown(i32, i32, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_json_parse_string(ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_string_concat(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_length(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_index_of(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_last_index(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_starts_with(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_ends_with(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_char_at(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_char_code_at(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_code_point_at(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_from_code_point(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_is_well_formed(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_to_well_formed(ptr, ptr)\n")
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
	out.WriteString("declare i32 @scriptgo_string_from_bigint(i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_from_bool(i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_slice(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_trim(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_replace(ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_replace_regex(ptr, ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_match(ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_search(ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_split(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_substr(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_from_char_codes(ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_at(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_anchor(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_big(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_blink(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_bold(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_fixed(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_fontcolor(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_fontsize(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_italics(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_link(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_small(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_strike(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_sub(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_sup(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_encode_uri_component(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_decode_uri_component(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_encode_uri(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_decode_uri(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_release(ptr)\n")
	out.WriteString("declare i32 @scriptgo_string_compare(ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_regex_test(ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_regex_exec(ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_regex_escape(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_bigint_from_number(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_bigint_from_string(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_bigint_as_int_n(i64, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_bigint_as_uint_n(i64, i64, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_symbol_create(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_symbol_for(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_symbol_key_for(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_symbol_description(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_symbol_to_string(ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_fs_read_file_sync(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_fs_write_file_sync(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_fs_exists_sync(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_fs_unlink_sync(ptr)\n")
	out.WriteString("declare i32 @scriptgo_fs_stat_sync(ptr, ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_fs_readdir_sync(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_fs_copy_file_sync(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_fs_rename_sync(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_fs_append_file_sync(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_fs_mkdir_sync(ptr, double)\n")
	out.WriteString("declare i32 @scriptgo_fs_rm_sync(ptr, double, double)\n\n")
	out.WriteString("declare void @scriptgo_process_init(i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_process_exit(double)\n")
	out.WriteString("declare i32 @scriptgo_process_cwd(ptr)\n")
	out.WriteString("declare i32 @scriptgo_process_argv(ptr)\n")
	out.WriteString("declare i32 @scriptgo_process_env(ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_child_process_exec_sync(ptr, ptr, ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_child_process_spawn_sync(ptr, ptr, ptr, ptr, ptr, ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_crypto_random_uuid(ptr)\n")
	out.WriteString("declare i32 @scriptgo_crypto_hash_digest(ptr, ptr, ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_date_now(ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_to_iso_string(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_to_string(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_to_date_string(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_to_time_string(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_to_utc_string(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_parse(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_utc(double, double, double, double, double, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_date(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_day(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_full_year(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_hours(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_milliseconds(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_minutes(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_month(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_seconds(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_timezone_offset(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_utc_date(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_utc_day(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_utc_full_year(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_utc_hours(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_utc_milliseconds(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_utc_minutes(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_utc_month(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_get_utc_seconds(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_date(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_full_year(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_hours(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_milliseconds(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_minutes(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_month(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_seconds(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_utc_date(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_utc_full_year(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_utc_hours(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_utc_milliseconds(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_utc_minutes(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_utc_month(double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_date_set_utc_seconds(double, double, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_web_btoa(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_web_atob(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_performance_now(ptr)\n")
	out.WriteString("declare i32 @scriptgo_fetch_sync(ptr, ptr, ptr, ptr, ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_stream_get_default_high_water_mark(i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_stream_set_default_high_water_mark(i32, double)\n\n")
	out.WriteString("declare void @scriptgo_exception_push(ptr)\n")
	out.WriteString("declare void @scriptgo_exception_pop(ptr)\n")
	out.WriteString("declare ptr @scriptgo_exception_buf(ptr)\n")
	out.WriteString("declare ptr @scriptgo_exception_frame_new()\n")
	out.WriteString("declare void @scriptgo_exception_frame_free(ptr)\n")
	out.WriteString("declare i32 @setjmp(ptr) returns_twice\n")
	out.WriteString("declare void @scriptgo_throw_string(ptr)\n")
	out.WriteString("declare void @scriptgo_throw_number(double)\n")
	out.WriteString("declare void @scriptgo_throw_bool(i32)\n")
	out.WriteString("declare ptr @scriptgo_exception_get_string(ptr)\n")
	out.WriteString("declare double @scriptgo_exception_get_number(ptr)\n")
	out.WriteString("declare i32 @scriptgo_exception_get_bool(ptr)\n")
	out.WriteString("declare void @scriptgo_exception_rethrow(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_closure_create(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_get_unknown(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_map_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_map_number_from_ptr(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_map_number_from_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_map_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_map_string_from_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_map_string_from_ptr(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_map_ptr(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_map_ptr_from_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_map_ptr_from_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_filter_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_filter_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_filter_ptr(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_for_each_number(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_for_each_string(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_for_each_ptr(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_reduce_number(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_find_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_find_index_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_find_index_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_some_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_every_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_fill_number(ptr, double, double, double, i32, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_fill_string(ptr, ptr, double, double, i32, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_to_reversed(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_to_sorted_number(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_to_sorted_string(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_find_last_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_find_last_index_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_find_last_index_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_reduce_right_number(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_last_index_of_number(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_last_index_of_string(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_copy_within(ptr, double, double, double, i32, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_with_number(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_with_string(ptr, double, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_to_spliced(ptr, double, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_sort_number(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_sort_string(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_sort_closure_number(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_sort_closure_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_sort_closure_ptr(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_array_is_array(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_queue_microtask(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_promise_create(ptr)\n")
	out.WriteString("declare i32 @scriptgo_promise_resolve(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_promise_resolve_number(ptr, double)\n")
	out.WriteString("declare i32 @scriptgo_promise_reject(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_promise_then(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_promise_await_number(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_promise_await_ptr(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_event_loop_run()\n")
	out.WriteString("declare i32 @scriptgo_timer_set_timeout(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_timer_clear_timeout(double)\n")
	out.WriteString("declare i32 @scriptgo_timer_set_interval(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_timer_clear_interval(double)\n")
	out.WriteString("declare i32 @scriptgo_timer_set_immediate(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_timer_clear_immediate(double)\n")
	out.WriteString("declare i32 @scriptgo_timers_drain()\n\n")
	out.WriteString("declare i32 @scriptgo_arraybuffer_new(i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_arraybuffer_slice(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_arraybuffer_byte_length(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_arraybuffer_is_view(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_arraybuffer_to_string(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_new(i64, i64, ptr, i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_get(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_get_bigint(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_set(ptr, double, double)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_set_bigint(ptr, double, i64)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_length(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_byte_length(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_byte_offset(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_buffer(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_subarray(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_slice(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_from_array(i64, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_set_array(ptr, ptr, double)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_fill(ptr, double, double, double)\n")
	out.WriteString("declare i32 @scriptgo_typedarray_to_string(ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_buffer_alloc(double, ptr, double, i32, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_from_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_from_array(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_concat(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_is_buffer(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_byte_length(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_to_string(ptr, ptr, double, double, i32, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_copy(ptr, ptr, double, double, double, i32, i32, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_fill(ptr, ptr, double, i32, double, double, i32, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_equals(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_compare(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_index_of(ptr, ptr, double, i32, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_read_u8(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_write_u8(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_read_i8(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_write_i8(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_read_u16(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_write_u16(ptr, double, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_read_i16(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_write_i16(ptr, double, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_read_u32(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_write_u32(ptr, double, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_read_i32(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_write_i32(ptr, double, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_read_float(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_write_float(ptr, double, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_read_double(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_buffer_write_double(ptr, double, double, i32, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_dataview_new(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_byte_length(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_byte_offset(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_buffer(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_get_int8(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_set_int8(ptr, double, double)\n")
	out.WriteString("declare i32 @scriptgo_dataview_get_uint8(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_set_uint8(ptr, double, double)\n")
	out.WriteString("declare i32 @scriptgo_dataview_get_int16(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_set_int16(ptr, double, double, i32)\n")
	out.WriteString("declare i32 @scriptgo_dataview_get_uint16(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_set_uint16(ptr, double, double, i32)\n")
	out.WriteString("declare i32 @scriptgo_dataview_get_int32(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_set_int32(ptr, double, double, i32)\n")
	out.WriteString("declare i32 @scriptgo_dataview_get_uint32(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_set_uint32(ptr, double, double, i32)\n")
	out.WriteString("declare i32 @scriptgo_dataview_get_float32(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_set_float32(ptr, double, double, i32)\n")
	out.WriteString("declare i32 @scriptgo_dataview_get_float64(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_set_float64(ptr, double, double, i32)\n")
	out.WriteString("declare i32 @scriptgo_dataview_get_bigint64(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_set_bigint64(ptr, double, i64, i32)\n")
	out.WriteString("declare i32 @scriptgo_dataview_get_biguint64(ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_dataview_set_biguint64(ptr, double, i64, i32)\n")
	out.WriteString("declare i32 @scriptgo_dataview_to_string(ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_map_new(ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_new_entries(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_set_string_number(ptr, ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_set_string_string(ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_set_string_ptr(ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_set_number_number(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_set_number_string(ptr, double, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_set_number_ptr(ptr, double, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_get_number(ptr, ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_get_string(ptr, ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_get_ptr(ptr, ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_has(ptr, ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_delete(ptr, ptr, double, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_clear(ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_size(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_to_string(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_for_each(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_keys(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_values(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_map_entries(ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_set_new(ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_new_values_number(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_new_values_string(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_new_values_ptr(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_add_number(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_add_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_add_ptr(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_has_number(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_has_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_has_ptr(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_delete_number(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_delete_string(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_delete_ptr(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_clear(ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_size(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_to_string(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_for_each(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_keys(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_values(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_set_entries(ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_text_encoder_new(ptr)\n")
	out.WriteString("declare i32 @scriptgo_text_encoder_encoding(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_text_encoder_encode(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_text_encoder_encode_into(ptr, ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_text_decoder_new(ptr, i32, i32, ptr)\n")
	out.WriteString("declare i32 @scriptgo_text_decoder_encoding(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_text_decoder_fatal(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_text_decoder_ignore_bom(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_text_decoder_decode(ptr, ptr, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_console_log_object(ptr)\n")
	out.WriteString("declare i32 @scriptgo_console_info_object(ptr)\n")
	out.WriteString("declare i32 @scriptgo_console_debug_object(ptr)\n")
	out.WriteString("declare i32 @scriptgo_console_warn_object(ptr)\n")
	out.WriteString("declare i32 @scriptgo_console_error_object(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_weakref_new(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_weakref_deref(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_weakmap_new(ptr)\n")
	out.WriteString("declare i32 @scriptgo_weakmap_set(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_weakmap_get(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_weakmap_has(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_weakmap_delete(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_weakset_new(ptr)\n")
	out.WriteString("declare i32 @scriptgo_weakset_add(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_weakset_has(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_weakset_delete(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_finalization_registry_new(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_finalization_registry_register(ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_finalization_registry_unregister(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_shared_array_buffer_new(i64, ptr)\n")
	out.WriteString("declare i32 @scriptgo_atomics_is_lock_free(double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_atomics_add(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_atomics_sub(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_atomics_and(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_atomics_or(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_atomics_xor(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_atomics_load(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_atomics_store(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_atomics_exchange(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_atomics_compare_exchange(ptr, double, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_atomics_wait(ptr, double, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_atomics_notify(ptr, double, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_gc_collect(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_intl_number_format_new(ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_number_format_format(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_collator_new(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_collator_compare(ptr, ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_segmenter_new(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_segmenter_segment(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_get_canonical_locales(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_display_names_new(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_display_names_of(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_list_format_new(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_list_format_format(ptr, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_relative_time_format_new(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_relative_time_format_format(ptr, double, ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_plural_rules_new(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_plural_rules_select(ptr, double, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_date_time_format_new(ptr, ptr)\n")
	out.WriteString("declare i32 @scriptgo_intl_date_time_format_format(ptr, double, ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_os_platform(ptr)\n")
	out.WriteString("declare i32 @scriptgo_os_arch(ptr)\n")
	out.WriteString("declare i32 @scriptgo_os_homedir(ptr)\n")
	out.WriteString("declare i32 @scriptgo_os_type(ptr)\n")
	out.WriteString("declare i32 @scriptgo_os_release(ptr)\n")
	out.WriteString("declare i32 @scriptgo_os_tmpdir(ptr)\n")
	out.WriteString("declare i32 @scriptgo_os_uptime(ptr)\n")
	out.WriteString("declare i32 @scriptgo_os_totalmem(ptr)\n")
	out.WriteString("declare i32 @scriptgo_os_freemem(ptr)\n\n")
	out.WriteString("declare i32 @scriptgo_process_pid(ptr)\n")
	out.WriteString("declare i32 @scriptgo_process_ppid(ptr)\n")
	out.WriteString("declare i32 @scriptgo_process_version(ptr)\n\n")
	out.WriteString("declare ptr @malloc(i64)\n\n")

	alreadyDeclared := map[string]bool{
		"malloc": true, "setjmp": true, "tan": true, "atan": true, "atan2": true, "hypot": true, "drand48": true,
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
		out.WriteString(fmt.Sprintf("declare %s @%s(%s)\n", retType, ext.Name, strings.Join(paramTypes, ", ")))
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
		} else if gType == "{ i32, i32, i64 }" {
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
	out.WriteString("\n")

	for _, function := range module.Functions {
		text, err := emitFunction(function, functions, stringsByValue, debug, module, options)
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

func emitFunction(function ir.Function, functions map[string]ir.Function, stringsByValue map[string]string, debug *debugInfo, module ir.Module, options Options) (string, error) {
	returnType := llvmType(function.ReturnType)
	if function.ReturnType == ir.TypeBool {
		returnType = "zeroext i1"
	}
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

	verStr := options.CompilerVersion
	if verStr == "" || verStr == "dev" {
		verStr = "v0.1.0"
	}
	if !strings.HasPrefix(verStr, "v") {
		verStr = "v" + verStr
	}

	emitter := &functionEmitter{
		function:        function,
		functions:       functions,
		stringsByValue:  stringsByValue,
		debug:           debug,
		module:          module,
		compilerVersion: verStr,
		types:           make(map[string]ir.Type, len(function.Parameters)),
		varSlots:        make(map[string]string),
		localSSAs:       make(map[string]bool),
	}
	globalsMap := make(map[string]bool, len(module.Globals))
	for _, g := range module.Globals {
		globalsMap[g.Name] = true
	}
	for _, parameter := range function.Parameters {
		emitter.types[parameter.Name] = parameter.Type
		emitter.localSSAs[parameter.Name] = true
	}
	collectSSADefs(function.Body, emitter.localSSAs, function.Name == "main", globalsMap)

	if len(function.Captured) > 0 {
		fieldTypes := make([]string, len(function.Captured))
		for i := range function.Captured {
			fieldTypes[i] = "ptr"
		}
		structType := fmt.Sprintf("{ %s }", strings.Join(fieldTypes, ", "))
		for i, c := range function.Captured {
			fieldPtr := fmt.Sprintf("%s.field.%d", c.Name, i)
			out.WriteString(fmt.Sprintf("  %%%s = getelementptr inbounds %s, ptr %%__env_ctx, i32 0, i32 %d\n", fieldPtr, structType, i))
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
					allocSize = 16
				}
				break
			}
		}
		out.WriteString(fmt.Sprintf("  %%%s = call ptr @malloc(i64 %d)\n", cellSlot, allocSize))
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
	for varName, typ := range slotted {
		if globalsMap[varName] {
			continue
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
			for _, param := range function.Parameters {
				if param.Name == varName {
					out.WriteString(fmt.Sprintf("  store volatile %s %%%s, ptr %%%s\n", allocType, varName, slotName))
					break
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
			out.WriteString("  call i32 @scriptgo_timers_drain()\n")
			out.WriteString("  ret i32 0\n")
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
	return out.String(), nil
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
				slotted[name] = typ
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

func collectSSADefs(instructions []ir.Instruction, defs map[string]bool, isMain bool, globals map[string]bool) {
	for _, inst := range instructions {
		if inst.Result != "" {
			if !globals[inst.Result] {
				defs[inst.Result] = true
			}
		}
		collectSSADefs(inst.Then, defs, isMain, globals)
		collectSSADefs(inst.Else, defs, isMain, globals)
		collectSSADefs(inst.Cond, defs, isMain, globals)
		collectSSADefs(inst.Body, defs, isMain, globals)
		collectSSADefs(inst.Step, defs, isMain, globals)
		collectSSADefs(inst.Catch, defs, isMain, globals)
		collectSSADefs(inst.Finally, defs, isMain, globals)
	}
}
