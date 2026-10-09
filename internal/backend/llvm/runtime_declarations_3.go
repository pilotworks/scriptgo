package llvm

// Native runtime ABI declarations, emitted verbatim in module order by
// writeRuntimeDeclarations. Keep each block in sync with internal/runtime.

const declarationsText = `declare i32 @scriptgo_text_encoder_new(ptr)
declare i32 @scriptgo_text_encoder_encoding(ptr, ptr)
declare i32 @scriptgo_text_encoder_encode(ptr, ptr, ptr)
declare i32 @scriptgo_text_encoder_encode_into(ptr, ptr, ptr, ptr, ptr)
declare i32 @scriptgo_text_decoder_new(ptr, i32, i32, ptr)
declare i32 @scriptgo_text_decoder_encoding(ptr, ptr)
declare i32 @scriptgo_text_decoder_fatal(ptr, ptr)
declare i32 @scriptgo_text_decoder_ignore_bom(ptr, ptr)
declare i32 @scriptgo_text_decoder_decode(ptr, ptr, ptr)

`

const declarationsConsole2 = `declare i32 @scriptgo_console_log_object(ptr)
declare i32 @scriptgo_console_info_object(ptr)
declare i32 @scriptgo_console_debug_object(ptr)
declare i32 @scriptgo_console_warn_object(ptr)
declare i32 @scriptgo_console_error_object(ptr)

declare i32 @scriptgo_console_log_buffer(ptr)
declare i32 @scriptgo_console_info_buffer(ptr)
declare i32 @scriptgo_console_debug_buffer(ptr)
declare i32 @scriptgo_console_warn_buffer(ptr)
declare i32 @scriptgo_console_error_buffer(ptr)

`

const declarationsJson2 = `declare i32 @scriptgo_json_inspect_object(ptr, ptr)
declare i32 @scriptgo_console_inspect_buffer(ptr, ptr)

`

const declarationsConsole3 = `declare i32 @scriptgo_console_inspect_array(ptr, ptr)
declare i32 @scriptgo_console_inspect_number(double, ptr)
declare i32 @scriptgo_weakref_new(ptr, ptr)
declare i32 @scriptgo_weakref_deref(ptr, ptr)
declare i32 @scriptgo_weakmap_new(ptr)
declare i32 @scriptgo_weakmap_set(ptr, ptr, ptr, i32)
declare i32 @scriptgo_weakmap_get(ptr, ptr, ptr, ptr)
declare i32 @scriptgo_weakmap_has(ptr, ptr, ptr)
declare i32 @scriptgo_weakmap_delete(ptr, ptr, ptr)
declare i32 @scriptgo_weakset_new(ptr)
declare i32 @scriptgo_weakset_add(ptr, ptr)
declare i32 @scriptgo_weakset_has(ptr, ptr, ptr)
declare i32 @scriptgo_weakset_delete(ptr, ptr, ptr)
declare i32 @scriptgo_finalization_registry_new(ptr, ptr)
declare i32 @scriptgo_finalization_registry_register(ptr, ptr, ptr, ptr)
declare i32 @scriptgo_finalization_registry_unregister(ptr, ptr, ptr)
declare i32 @scriptgo_shared_array_buffer_new(i64, ptr)
declare i32 @scriptgo_atomics_is_lock_free(double, ptr)
declare i32 @scriptgo_atomics_add(ptr, double, double, ptr)
declare i32 @scriptgo_atomics_sub(ptr, double, double, ptr)
declare i32 @scriptgo_atomics_and(ptr, double, double, ptr)
declare i32 @scriptgo_atomics_or(ptr, double, double, ptr)
declare i32 @scriptgo_atomics_xor(ptr, double, double, ptr)
declare i32 @scriptgo_atomics_load(ptr, double, ptr)
declare i32 @scriptgo_atomics_store(ptr, double, double, ptr)
declare i32 @scriptgo_atomics_exchange(ptr, double, double, ptr)
declare i32 @scriptgo_atomics_compare_exchange(ptr, double, double, double, ptr)
declare i32 @scriptgo_atomics_wait(ptr, double, double, double, ptr)
declare i32 @scriptgo_atomics_notify(ptr, double, double, ptr)
declare i32 @scriptgo_gc_collect(ptr)
declare i32 @scriptgo_gc_get_tag(ptr)
declare i32 @scriptgo_gc_add_root_slot(ptr, i64)

`

const declarationsWebsocket = `declare i32 @scriptgo_websocket_connect(ptr, ptr, ptr)
declare i32 @scriptgo_websocket_send_text(double, ptr, ptr)
declare i32 @scriptgo_websocket_send_binary(double, ptr, double, ptr)
declare i32 @scriptgo_websocket_close(double, double, ptr)
declare i32 @scriptgo_websocket_poll(double, ptr, ptr, ptr, ptr)
declare i32 @scriptgo_websocket_ready_state(double, ptr)

`

const declarationsIntl = `declare i32 @scriptgo_intl_number_format_new(ptr, ptr, ptr, ptr)
declare i32 @scriptgo_intl_number_format_format(ptr, double, ptr)
declare i32 @scriptgo_intl_collator_new(ptr, ptr)
declare i32 @scriptgo_intl_collator_compare(ptr, ptr, ptr, ptr)
declare i32 @scriptgo_intl_segmenter_new(ptr, ptr)
declare i32 @scriptgo_intl_segmenter_segment(ptr, ptr, ptr)
declare i32 @scriptgo_intl_segments_length(ptr, ptr)
declare i32 @scriptgo_intl_segments_get(ptr, double, ptr)
declare i32 @scriptgo_intl_segments_containing(ptr, double, ptr)
declare i32 @scriptgo_intl_get_canonical_locales(ptr, ptr)
declare i32 @scriptgo_intl_display_names_new(ptr, ptr, ptr)
declare i32 @scriptgo_intl_display_names_of(ptr, ptr, ptr)
declare i32 @scriptgo_intl_list_format_new(ptr, ptr)
declare i32 @scriptgo_intl_list_format_format(ptr, ptr, ptr)
declare i32 @scriptgo_intl_relative_time_format_new(ptr, ptr)
declare i32 @scriptgo_intl_relative_time_format_format(ptr, double, ptr, ptr)
declare i32 @scriptgo_intl_plural_rules_new(ptr, ptr)
declare i32 @scriptgo_intl_plural_rules_select(ptr, double, ptr)
declare i32 @scriptgo_intl_date_time_format_new(ptr, ptr)
declare i32 @scriptgo_intl_date_time_format_format(ptr, double, ptr)

`

const declarationsOs = `declare i32 @scriptgo_os_platform(ptr)
declare i32 @scriptgo_os_arch(ptr)
declare i32 @scriptgo_os_homedir(ptr)
declare i32 @scriptgo_os_type(ptr)
declare i32 @scriptgo_os_release(ptr)
declare i32 @scriptgo_os_tmpdir(ptr)
declare i32 @scriptgo_os_uptime(ptr)
declare i32 @scriptgo_os_totalmem(ptr)
declare i32 @scriptgo_os_freemem(ptr)
declare i32 @scriptgo_os_available_parallelism(ptr)
declare i32 @scriptgo_os_hostname(ptr)
declare i32 @scriptgo_os_loadavg(ptr)
declare i32 @scriptgo_os_cpus(ptr)
declare i32 @scriptgo_os_network_interfaces(ptr)
declare i32 @scriptgo_os_user_info(ptr)
declare i32 @scriptgo_os_machine(ptr)
declare i32 @scriptgo_os_version(ptr)
declare i32 @scriptgo_os_get_priority(double, ptr)
declare i32 @scriptgo_os_set_priority(double, double)

`

const declarationsProcess2 = `declare i32 @scriptgo_process_pid(ptr)
declare i32 @scriptgo_process_ppid(ptr)
declare i32 @scriptgo_process_version(ptr)

`

const declarationsTty = `declare i32 @scriptgo_tty_isatty(double, ptr)
declare i32 @scriptgo_tty_get_window_size(double, ptr, ptr)
declare i32 @scriptgo_tty_set_raw_mode(double, double, ptr)
declare i32 @scriptgo_tty_read(double, double, ptr, ptr)
declare i32 @scriptgo_tty_read_line(double, ptr)
declare i32 @scriptgo_tty_write(double, ptr, double, ptr)

`

const declarationsClosure2 = `declare ptr @scriptgo_closure_alloc(i64)

`

const declarationsObject2 = `declare void @scriptgo_object_region_begin()
declare void @scriptgo_object_region_end()

`
