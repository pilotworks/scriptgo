/* QuickJS-ng adapter for the first, synchronous Dynamic island. */
#include "quickjs.h"
int scriptgo_runtime_set_error(const char *message);
void scriptgo_runtime_abort_if_failed(int status);

void scriptgo_dynamic_abort_if_failed(int status) {
    if (status == SCRIPTGO_CALL_THROWN) {
        scriptgo_runtime_set_error("Dynamic JavaScript exception");
    }
    scriptgo_runtime_abort_if_failed(status);
}

static int scriptgo_dynamic_to_js(JSContext *ctx, const scriptgo_value *value, JSValue *out) {
    union { uint64_t u; double d; } number;
    if (value == NULL || out == NULL) return -1;
    switch (value->tag) {
    case SCRIPTGO_TAG_UNDEFINED: *out = JS_UNDEFINED; return 0;
    case SCRIPTGO_TAG_NULL: *out = JS_NULL; return 0;
    case SCRIPTGO_TAG_BOOLEAN: *out = JS_NewBool(ctx, value->payload != 0); return 0;
    case SCRIPTGO_TAG_NUMBER:
        number.u = value->payload; *out = JS_NewFloat64(ctx, number.d); return 0;
    case SCRIPTGO_TAG_STRING:
        *out = JS_NewStringLen(ctx, (const char *)(uintptr_t)value->payload,
                               value->aux == 0 ? strlen((const char *)(uintptr_t)value->payload) : (size_t)value->aux);
        return JS_IsException(*out) ? -1 : 0;
    default: return -1;
    }
}

static int scriptgo_dynamic_from_js(JSContext *ctx, JSValue value, scriptgo_value *out) {
    int tag = JS_VALUE_GET_TAG(value);
    double number;
    size_t length;
    const char *text;
    if (out == NULL) return -1;
    out->tag = SCRIPTGO_TAG_UNDEFINED;
    out->flags = 0;
    out->payload = 0;
    out->aux = 0;
    if (JS_IsUndefined(value)) return 0;
    if (JS_IsNull(value)) { out->tag = SCRIPTGO_TAG_NULL; return 0; }
    if (JS_IsBool(value)) { out->tag = SCRIPTGO_TAG_BOOLEAN; out->payload = (uint64_t)JS_ToBool(ctx, value); return 0; }
    if (JS_IsNumber(value)) {
        if (JS_ToFloat64(ctx, &number, value) < 0) return -1;
        { union { uint64_t u; double d; } bits; bits.d = number; out->tag = SCRIPTGO_TAG_NUMBER; out->payload = bits.u; }
        return 0;
    }
    if (JS_IsString(value)) {
        text = JS_ToCStringLen(ctx, &length, value);
        if (text == NULL) return -1;
        char *copy = (char *)malloc(length + 1);
        if (copy == NULL) { JS_FreeCString(ctx, text); return -1; }
        memcpy(copy, text, length); copy[length] = '\0'; JS_FreeCString(ctx, text);
        out->tag = SCRIPTGO_TAG_STRING; out->flags = SCRIPTGO_VALUE_OWNED;
        out->payload = (uint64_t)(uintptr_t)copy; out->aux = (uint64_t)length; return 0;
    }
    (void)tag;
    return -1;
}

static char *scriptgo_dynamic_script(const char *source, size_t length) {
    char *copy = (char *)malloc(length + 1);
    size_t i;
    if (copy == NULL) return NULL;
    memcpy(copy, source, length); copy[length] = '\0';
    /* v1 accepts named export functions; turn the declaration into a script. */
    for (i = 0; i + 7 <= length; i++) {
        if (memcmp(copy + i, "export ", 7) == 0) memcpy(copy + i, "       ", 7);
    }
    return copy;
}

int scriptgo_dynamic_call_module(const char *module_path, const char *source,
                          const char *export_name, const scriptgo_value *args,
                          int32_t argument_count, int32_t expected_arity,
                          int32_t expected_tag, scriptgo_value *out_result) {
    JSRuntime *rt = NULL; JSContext *ctx = NULL; JSValue *argv = NULL;
    JSValue global, function, result; char *script = NULL; size_t length;
    int32_t i, status = 0;
    if (module_path == NULL || source == NULL || export_name == NULL || out_result == NULL || argument_count < 0)
        return scriptgo_runtime_set_error("SG9001: invalid Dynamic call descriptor");
    if (expected_arity < 0 || argument_count != expected_arity)
        return scriptgo_runtime_set_error("SG5002: Dynamic call arity mismatch");
    length = strlen(source); script = scriptgo_dynamic_script(source, length);
    if (script == NULL) return scriptgo_runtime_set_error("Dynamic engine source allocation failed");
    rt = JS_NewRuntime(); ctx = rt == NULL ? NULL : JS_NewContext(rt);
    if (ctx == NULL) { free(script); if (rt != NULL) JS_FreeRuntime(rt); return scriptgo_runtime_set_error("Dynamic engine initialization failed"); }
    result = JS_Eval(ctx, script, length, module_path, JS_EVAL_TYPE_GLOBAL);
    free(script);
    if (JS_IsException(result)) { JS_FreeValue(ctx, result); status = 1; goto done; }
    JS_FreeValue(ctx, result);
    global = JS_GetGlobalObject(ctx);
    function = JS_GetPropertyStr(ctx, global, export_name);
    JS_FreeValue(ctx, global);
    if (!JS_IsFunction(ctx, function)) { JS_FreeValue(ctx, function); status = scriptgo_runtime_set_error("Dynamic export is not a function"); goto done; }
    argv = argument_count == 0 ? NULL : (JSValue *)calloc((size_t)argument_count, sizeof(JSValue));
    if (argument_count != 0 && argv == NULL) { JS_FreeValue(ctx, function); status = scriptgo_runtime_set_error("Dynamic argument allocation failed"); goto done; }
    for (i = 0; i < argument_count; i++) {
        if (scriptgo_dynamic_to_js(ctx, &args[i], &argv[i]) != 0) { status = scriptgo_runtime_set_error("SG5002: Dynamic input mismatch"); goto cleanup; }
    }
    result = JS_Call(ctx, function, JS_UNDEFINED, argument_count, argv);
    if (JS_IsException(result)) { JS_FreeValue(ctx, result); status = 1; goto cleanup; }
    if (scriptgo_dynamic_from_js(ctx, result, out_result) != 0) { JS_FreeValue(ctx, result); status = scriptgo_runtime_set_error("SG5003: Dynamic result mismatch"); goto cleanup; }
    if (expected_tag >= 0 && (int32_t)out_result->tag != expected_tag) { scriptgo_value_release(out_result); JS_FreeValue(ctx, result); status = scriptgo_runtime_set_error("SG5003: Dynamic result mismatch"); goto cleanup; }
    JS_FreeValue(ctx, result);
cleanup:
    for (i = 0; i < argument_count; i++) JS_FreeValue(ctx, argv[i]);
    free(argv); JS_FreeValue(ctx, function);
done:
    JS_FreeContext(ctx); JS_FreeRuntime(rt); return status;
}
