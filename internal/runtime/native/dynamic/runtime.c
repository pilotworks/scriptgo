/* Persistent QuickJS-ng runtime and closed Dynamic module graph. */
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "quickjs.h"
int scriptgo_runtime_set_error(const char *message);
void scriptgo_runtime_abort_if_failed(int status);

void scriptgo_dynamic_abort_if_failed(int status) {
    if (status == SCRIPTGO_CALL_THROWN) {
        scriptgo_runtime_set_error("Dynamic JavaScript exception");
    }
    scriptgo_runtime_abort_if_failed(status);
}

int scriptgo_object_keys(void *handle, void **out_array);
int scriptgo_object_property_unknown_get(void *handle, const char *property, scriptgo_value *out_value);
int scriptgo_object_new(int64_t field_count, void **out_object);
int scriptgo_object_type_set(void *handle, const char *type_name);
int scriptgo_object_property_unknown_set(void *handle, const char *property, const scriptgo_value *value);
int scriptgo_array_length(void *handle, int64_t *out_length);
int scriptgo_array_get(void *handle, double index, void *out_value);
int scriptgo_array_get_unknown(void *handle, double index, void *out_value);
int scriptgo_array_new(int64_t length, int64_t element_size, void **out_array);
int scriptgo_array_set(void *handle, double index, const void *value);
int scriptgo_array_release(void *handle);
int scriptgo_object_release(void *handle);

static int scriptgo_dynamic_to_js(JSContext *ctx, const scriptgo_value *value, JSValue *out);
static int scriptgo_dynamic_from_js(JSContext *ctx, JSValue value, scriptgo_value *out);

typedef struct {
    char *specifier;
    char *path;
} scriptgo_dynamic_dependency;

typedef struct {
    char *path;
    char *source;
    char *kind;
    char *exports;
    scriptgo_dynamic_dependency *dependencies;
    size_t dependency_count;
    JSValue cjs_exports;
    int cjs_state;
    JSModuleDef *esm_module;
} scriptgo_dynamic_module;

static scriptgo_dynamic_module *scriptgo_dynamic_modules = NULL;
static size_t scriptgo_dynamic_module_count = 0;
static JSRuntime *scriptgo_dynamic_runtime = NULL;
static JSContext *scriptgo_dynamic_js_context = NULL;
static scriptgo_dynamic_context *scriptgo_dynamic_value_context = NULL;
static int scriptgo_dynamic_cleanup_registered = 0;

typedef struct scriptgo_dynamic_function_ref {
    JSValue function;
    uint32_t refs;
} scriptgo_dynamic_function_ref;

static scriptgo_dynamic_module *scriptgo_dynamic_find_module(const char *path);
static JSValue scriptgo_dynamic_require_module(JSContext *ctx, scriptgo_dynamic_module *module);
static JSValue scriptgo_dynamic_esm_namespace(JSContext *ctx, scriptgo_dynamic_module *module);
static int scriptgo_dynamic_function_call(scriptgo_dynamic_context *context, uint64_t handle,
                                          const scriptgo_value *this_value, const scriptgo_value *arguments,
                                          uint32_t argument_count, const scriptgo_boundary_descriptor *descriptor,
                                          scriptgo_value *out_result, scriptgo_value *out_exception, void *user_data);

static void scriptgo_dynamic_function_retain(scriptgo_dynamic_context *context, uint64_t handle) {
    scriptgo_dynamic_function_ref *ref = (scriptgo_dynamic_function_ref *)(uintptr_t)handle;
    (void)context;
    if (ref != NULL) ref->refs++;
}

static void scriptgo_dynamic_function_release(scriptgo_dynamic_context *context, uint64_t handle) {
    scriptgo_dynamic_function_ref *ref = (scriptgo_dynamic_function_ref *)(uintptr_t)handle;
    (void)context;
    if (ref == NULL || ref->refs == 0) return;
    ref->refs--;
    if (ref->refs == 0) {
        JS_FreeValue(scriptgo_dynamic_js_context, ref->function);
        free(ref);
    }
}

int scriptgo_dynamic_adopt_function(JSValue value, scriptgo_value *out) {
    scriptgo_dynamic_function_ref *ref;
    if (out == NULL || scriptgo_dynamic_js_context == NULL || scriptgo_dynamic_value_context == NULL)
        return -1;
    ref = (scriptgo_dynamic_function_ref *)calloc(1, sizeof(*ref));
    if (ref == NULL) return -1;
    ref->function = JS_DupValue(scriptgo_dynamic_js_context, value);
    ref->refs = 1;
    if (scriptgo_value_adopt_engine_ref(scriptgo_dynamic_value_context, SCRIPTGO_TAG_FUNCTION,
                                        (uint64_t)(uintptr_t)ref, out) != 0) {
        JS_FreeValue(scriptgo_dynamic_js_context, ref->function);
        free(ref);
        return -1;
    }
    return 0;
}

static char *scriptgo_dynamic_copy(const char *value) {
    size_t length;
    char *copy;
    if (value == NULL) return NULL;
    length = strlen(value);
    copy = (char *)malloc(length + 1);
    if (copy != NULL) memcpy(copy, value, length + 1);
    return copy;
}

static scriptgo_dynamic_module *scriptgo_dynamic_find_module(const char *path) {
    size_t i;
    if (path == NULL) return NULL;
    for (i = 0; i < scriptgo_dynamic_module_count; i++) {
        if (strcmp(scriptgo_dynamic_modules[i].path, path) == 0) return &scriptgo_dynamic_modules[i];
    }
    return NULL;
}

int scriptgo_dynamic_register_module(const char *path, const char *source,
                                     const char *kind, const char *exports) {
    scriptgo_dynamic_module *module;
    scriptgo_dynamic_module *grown;
    if (path == NULL || source == NULL || kind == NULL || exports == NULL)
        return scriptgo_runtime_set_error("SG9001: invalid Dynamic module descriptor");
    module = scriptgo_dynamic_find_module(path);
    if (module != NULL) {
        if (strcmp(module->source, source) != 0 || strcmp(module->kind, kind) != 0 || strcmp(module->exports, exports) != 0)
            return scriptgo_runtime_set_error("SG9001: conflicting Dynamic module descriptor");
        return 0;
    }
    grown = (scriptgo_dynamic_module *)realloc(scriptgo_dynamic_modules,
        (scriptgo_dynamic_module_count + 1) * sizeof(*scriptgo_dynamic_modules));
    if (grown == NULL) return scriptgo_runtime_set_error("Dynamic module registry allocation failed");
    scriptgo_dynamic_modules = grown;
    module = &scriptgo_dynamic_modules[scriptgo_dynamic_module_count];
    memset(module, 0, sizeof(*module));
    module->path = scriptgo_dynamic_copy(path);
    module->source = scriptgo_dynamic_copy(source);
    module->kind = scriptgo_dynamic_copy(kind);
    module->exports = scriptgo_dynamic_copy(exports);
    if (module->path == NULL || module->source == NULL || module->kind == NULL || module->exports == NULL) {
        free(module->path); free(module->source); free(module->kind); free(module->exports);
        memset(module, 0, sizeof(*module));
        return scriptgo_runtime_set_error("Dynamic module descriptor allocation failed");
    }
    scriptgo_dynamic_module_count++;
    return 0;
}

int scriptgo_dynamic_register_dependency(const char *module_path,
                                         const char *specifier,
                                         const char *resolved_path) {
    scriptgo_dynamic_module *module = scriptgo_dynamic_find_module(module_path);
    scriptgo_dynamic_dependency *grown;
    size_t i;
    if (module == NULL || specifier == NULL || resolved_path == NULL)
        return scriptgo_runtime_set_error("SG9001: invalid Dynamic dependency descriptor");
    for (i = 0; i < module->dependency_count; i++) {
        if (strcmp(module->dependencies[i].specifier, specifier) == 0) {
            if (strcmp(module->dependencies[i].path, resolved_path) != 0)
                return scriptgo_runtime_set_error("SG9001: conflicting Dynamic dependency descriptor");
            return 0;
        }
    }
    grown = (scriptgo_dynamic_dependency *)realloc(module->dependencies,
        (module->dependency_count + 1) * sizeof(*module->dependencies));
    if (grown == NULL) return scriptgo_runtime_set_error("Dynamic dependency registry allocation failed");
    module->dependencies = grown;
    module->dependencies[module->dependency_count].specifier = scriptgo_dynamic_copy(specifier);
    module->dependencies[module->dependency_count].path = scriptgo_dynamic_copy(resolved_path);
    if (module->dependencies[module->dependency_count].specifier == NULL || module->dependencies[module->dependency_count].path == NULL) {
        free(module->dependencies[module->dependency_count].specifier);
        free(module->dependencies[module->dependency_count].path);
        memset(&module->dependencies[module->dependency_count], 0, sizeof(*module->dependencies));
        return scriptgo_runtime_set_error("Dynamic dependency descriptor allocation failed");
    }
    module->dependency_count++;
    return 0;
}

static const char *scriptgo_dynamic_resolve_dependency(scriptgo_dynamic_module *module,
                                                       const char *specifier) {
    size_t i;
    if (module == NULL || specifier == NULL) return NULL;
    for (i = 0; i < module->dependency_count; i++) {
        if (strcmp(module->dependencies[i].specifier, specifier) == 0)
            return module->dependencies[i].path;
    }
    return NULL;
}

static char *scriptgo_dynamic_normalize(JSContext *ctx, const char *base,
                                        const char *name, void *opaque) {
    scriptgo_dynamic_module *module;
    const char *resolved;
    (void)opaque;
    if (scriptgo_dynamic_find_module(name) != NULL) return js_strdup(ctx, name);
    module = scriptgo_dynamic_find_module(base);
    resolved = scriptgo_dynamic_resolve_dependency(module, name);
    if (resolved != NULL) return js_strdup(ctx, resolved);
    JS_ThrowReferenceError(ctx, "unresolved Dynamic module '%s' imported from '%s'", name, base == NULL ? "" : base);
    return NULL;
}

static int scriptgo_dynamic_append(char **buffer, size_t *length, size_t *capacity,
                                   const char *text) {
    size_t addition = strlen(text);
    char *grown;
    if (*length + addition + 1 > *capacity) {
        size_t next = (*length + addition + 1) * 2;
        grown = (char *)realloc(*buffer, next);
        if (grown == NULL) return -1;
        *buffer = grown;
        *capacity = next;
    }
    memcpy(*buffer + *length, text, addition);
    *length += addition;
    (*buffer)[*length] = '\0';
    return 0;
}

static char *scriptgo_dynamic_quote(const char *text) {
    size_t i, length = 0, capacity = strlen(text) * 2 + 3;
    char *quoted = (char *)malloc(capacity);
    if (quoted == NULL) return NULL;
    quoted[length++] = '"';
    for (i = 0; text[i] != '\0'; i++) {
        char ch = text[i];
        if (ch == '\\' || ch == '"') quoted[length++] = '\\';
        if (ch == '\n') { quoted[length++] = '\\'; ch = 'n'; }
        else if (ch == '\r') { quoted[length++] = '\\'; ch = 'r'; }
        quoted[length++] = ch;
    }
    quoted[length++] = '"'; quoted[length] = '\0';
    return quoted;
}

static char *scriptgo_dynamic_cjs_bridge(scriptgo_dynamic_module *module) {
    char *quoted = scriptgo_dynamic_quote(module->path);
    char *buffer = NULL;
    char *exports = NULL;
    char *name;
    size_t length = 0, capacity = 0;
    int index = 0;
    if (quoted == NULL || scriptgo_dynamic_append(&buffer, &length, &capacity, "const __scriptgo_cjs = globalThis.__scriptgo_require_module(") != 0 ||
        scriptgo_dynamic_append(&buffer, &length, &capacity, quoted) != 0 ||
        scriptgo_dynamic_append(&buffer, &length, &capacity, ");\nexport default __scriptgo_cjs;\n") != 0) {
        free(quoted); free(buffer); return NULL;
    }
    free(quoted);
    exports = scriptgo_dynamic_copy(module->exports);
    if (exports == NULL) { free(buffer); return NULL; }
    name = exports;
    while (name != NULL && *name != '\0') {
        char *separator = strchr(name, 0x1f);
        char declaration[96];
        if (separator != NULL) *separator = '\0';
        if (strcmp(name, "default") != 0) {
            snprintf(declaration, sizeof(declaration), "const __scriptgo_export_%d = __scriptgo_cjs[", index);
            quoted = scriptgo_dynamic_quote(name);
            if (quoted == NULL || scriptgo_dynamic_append(&buffer, &length, &capacity, declaration) != 0 ||
                scriptgo_dynamic_append(&buffer, &length, &capacity, quoted) != 0 ||
                scriptgo_dynamic_append(&buffer, &length, &capacity, "]; export { __scriptgo_export_") != 0) {
                free(quoted); free(exports); free(buffer); return NULL;
            }
            free(quoted);
            quoted = scriptgo_dynamic_quote(name);
            snprintf(declaration, sizeof(declaration), "%d as ", index);
            if (quoted == NULL || scriptgo_dynamic_append(&buffer, &length, &capacity, declaration) != 0 ||
                scriptgo_dynamic_append(&buffer, &length, &capacity, quoted) != 0 ||
                scriptgo_dynamic_append(&buffer, &length, &capacity, " };\n") != 0) {
                free(quoted); free(exports); free(buffer); return NULL;
            }
            free(quoted);
            index++;
        }
        name = separator == NULL ? NULL : separator + 1;
    }
    free(exports);
    return buffer;
}

static JSModuleDef *scriptgo_dynamic_load_module(JSContext *ctx,
                                                  const char *module_name,
                                                  void *opaque) {
    scriptgo_dynamic_module *module = scriptgo_dynamic_find_module(module_name);
    const char *source;
    char *bridge = NULL;
    JSValue compiled;
    (void)opaque;
    if (module == NULL) {
        JS_ThrowReferenceError(ctx, "Dynamic module '%s' is not bundled", module_name);
        return NULL;
    }
    source = module->source;
    if (strcmp(module->kind, "commonjs") == 0) {
        bridge = scriptgo_dynamic_cjs_bridge(module);
        if (bridge == NULL) { JS_ThrowInternalError(ctx, "Dynamic CommonJS bridge allocation failed"); return NULL; }
        source = bridge;
    }
    compiled = JS_Eval(ctx, source, strlen(source), module->path,
                       JS_EVAL_TYPE_MODULE | JS_EVAL_FLAG_COMPILE_ONLY);
    free(bridge);
    if (JS_IsException(compiled)) return NULL;
    module->esm_module = (JSModuleDef *)JS_VALUE_GET_PTR(compiled);
    return module->esm_module;
}

static JSValue scriptgo_dynamic_cjs_require(JSContext *ctx, JSValueConst this_value,
                                            int argc, JSValueConst *argv, int magic,
                                            JSValueConst *data) {
    const char *base = JS_ToCString(ctx, data[0]);
    const char *specifier = argc > 0 ? JS_ToCString(ctx, argv[0]) : NULL;
    scriptgo_dynamic_module *module = scriptgo_dynamic_find_module(base);
    const char *resolved = scriptgo_dynamic_resolve_dependency(module, specifier);
    scriptgo_dynamic_module *dependency = scriptgo_dynamic_find_module(resolved);
    JSValue result;
    (void)this_value; (void)magic;
    if (base == NULL || specifier == NULL || dependency == NULL) {
        if (base != NULL) JS_FreeCString(ctx, base);
        if (specifier != NULL) JS_FreeCString(ctx, specifier);
        return JS_ThrowReferenceError(ctx, "unresolved CommonJS dependency");
    }
    if (strcmp(dependency->kind, "commonjs") == 0)
        result = scriptgo_dynamic_require_module(ctx, dependency);
    else
        result = scriptgo_dynamic_esm_namespace(ctx, dependency);
    JS_FreeCString(ctx, base); JS_FreeCString(ctx, specifier);
    return result;
}

static JSValue scriptgo_dynamic_require_by_path(JSContext *ctx, JSValueConst this_value,
                                                int argc, JSValueConst *argv) {
    const char *path = argc > 0 ? JS_ToCString(ctx, argv[0]) : NULL;
    scriptgo_dynamic_module *module = scriptgo_dynamic_find_module(path);
    JSValue result;
    (void)this_value;
    if (path == NULL || module == NULL || strcmp(module->kind, "commonjs") != 0) {
        if (path != NULL) JS_FreeCString(ctx, path);
        return JS_ThrowReferenceError(ctx, "Dynamic CommonJS module is not bundled");
    }
    result = scriptgo_dynamic_require_module(ctx, module);
    JS_FreeCString(ctx, path);
    return result;
}

static char *scriptgo_dynamic_dirname(const char *path) {
    const char *slash = strrchr(path, '/');
    size_t length = slash == NULL ? 1 : (size_t)(slash - path);
    char *result = (char *)malloc(length + 1);
    if (result == NULL) return NULL;
    if (slash == NULL) result[0] = '.';
    else memcpy(result, path, length);
    result[length] = '\0';
    return result;
}

static JSValue scriptgo_dynamic_require_module(JSContext *ctx, scriptgo_dynamic_module *module) {
    static const char prefix[] = "(function(module,exports,require,__filename,__dirname){\n";
    static const char suffix[] = "\n})";
    size_t source_length = strlen(module->source);
    char *wrapper;
    char *dirname;
    JSValue function, module_object, exports_object, require_function, path_value, dirname_value, result;
    JSValue data[1], args[5];
    if (module->cjs_state != 0) return JS_DupValue(ctx, module->cjs_exports);
    module->cjs_state = 1;
    module->cjs_exports = JS_NewObject(ctx);
    if (JS_IsException(module->cjs_exports)) { module->cjs_state = 0; return module->cjs_exports; }
    wrapper = (char *)malloc(sizeof(prefix) - 1 + source_length + sizeof(suffix));
    if (wrapper == NULL) { module->cjs_state = 0; return JS_ThrowInternalError(ctx, "Dynamic CommonJS wrapper allocation failed"); }
    memcpy(wrapper, prefix, sizeof(prefix) - 1);
    memcpy(wrapper + sizeof(prefix) - 1, module->source, source_length);
    memcpy(wrapper + sizeof(prefix) - 1 + source_length, suffix, sizeof(suffix));
    function = JS_Eval(ctx, wrapper, sizeof(prefix) - 1 + source_length + sizeof(suffix) - 1,
                       module->path, JS_EVAL_TYPE_GLOBAL);
    free(wrapper);
    if (JS_IsException(function)) { module->cjs_state = 0; return function; }
    module_object = JS_NewObject(ctx);
    exports_object = JS_DupValue(ctx, module->cjs_exports);
    JS_SetPropertyStr(ctx, module_object, "exports", JS_DupValue(ctx, exports_object));
    data[0] = JS_NewString(ctx, module->path);
    require_function = JS_NewCFunctionData(ctx, scriptgo_dynamic_cjs_require, 1, 0, 1, data);
    JS_FreeValue(ctx, data[0]);
    path_value = JS_NewString(ctx, module->path);
    dirname = scriptgo_dynamic_dirname(module->path);
    dirname_value = JS_NewString(ctx, dirname == NULL ? "." : dirname);
    free(dirname);
    args[0] = module_object; args[1] = exports_object; args[2] = require_function;
    args[3] = path_value; args[4] = dirname_value;
    result = JS_Call(ctx, function, JS_UNDEFINED, 5, args);
    JS_FreeValue(ctx, function);
    JS_FreeValue(ctx, exports_object); JS_FreeValue(ctx, require_function);
    JS_FreeValue(ctx, path_value); JS_FreeValue(ctx, dirname_value);
    if (JS_IsException(result)) { JS_FreeValue(ctx, module_object); module->cjs_state = 0; return result; }
    JS_FreeValue(ctx, result);
    result = JS_GetPropertyStr(ctx, module_object, "exports");
    JS_FreeValue(ctx, module_object);
    if (JS_IsException(result)) { module->cjs_state = 0; return result; }
    JS_FreeValue(ctx, module->cjs_exports);
    module->cjs_exports = JS_DupValue(ctx, result);
    module->cjs_state = 2;
    return result;
}

static void scriptgo_dynamic_cleanup(void) {
    size_t i, j;
    if (scriptgo_dynamic_value_context != NULL)
        scriptgo_dynamic_context_release_all(scriptgo_dynamic_value_context);
    if (scriptgo_dynamic_js_context != NULL) {
        for (i = 0; i < scriptgo_dynamic_module_count; i++) {
            if (scriptgo_dynamic_modules[i].cjs_state != 0)
                JS_FreeValue(scriptgo_dynamic_js_context, scriptgo_dynamic_modules[i].cjs_exports);
        }
        JS_FreeContext(scriptgo_dynamic_js_context);
    }
    if (scriptgo_dynamic_runtime != NULL) JS_FreeRuntime(scriptgo_dynamic_runtime);
    scriptgo_dynamic_value_context = NULL;
    for (i = 0; i < scriptgo_dynamic_module_count; i++) {
        free(scriptgo_dynamic_modules[i].path); free(scriptgo_dynamic_modules[i].source);
        free(scriptgo_dynamic_modules[i].kind); free(scriptgo_dynamic_modules[i].exports);
        for (j = 0; j < scriptgo_dynamic_modules[i].dependency_count; j++) {
            free(scriptgo_dynamic_modules[i].dependencies[j].specifier);
            free(scriptgo_dynamic_modules[i].dependencies[j].path);
        }
        free(scriptgo_dynamic_modules[i].dependencies);
    }
    free(scriptgo_dynamic_modules);
}

static int scriptgo_dynamic_ensure_context(void) {
    JSValue global, require_function;
    if (scriptgo_dynamic_js_context != NULL) return 0;
    scriptgo_dynamic_runtime = JS_NewRuntime();
    scriptgo_dynamic_js_context = scriptgo_dynamic_runtime == NULL ? NULL : JS_NewContext(scriptgo_dynamic_runtime);
    if (scriptgo_dynamic_js_context == NULL) return scriptgo_runtime_set_error("Dynamic engine initialization failed");
    scriptgo_dynamic_value_context = scriptgo_dynamic_context_new(
        scriptgo_dynamic_js_context, NULL, scriptgo_dynamic_function_retain,
        scriptgo_dynamic_function_release, scriptgo_dynamic_function_call);
    if (scriptgo_dynamic_value_context == NULL) return scriptgo_runtime_set_error("Dynamic value context initialization failed");
    JS_SetModuleLoaderFunc(scriptgo_dynamic_runtime, scriptgo_dynamic_normalize,
                           scriptgo_dynamic_load_module, NULL);
    global = JS_GetGlobalObject(scriptgo_dynamic_js_context);
    require_function = JS_NewCFunction(scriptgo_dynamic_js_context, scriptgo_dynamic_require_by_path,
                                       "__scriptgo_require_module", 1);
    JS_SetPropertyStr(scriptgo_dynamic_js_context, global, "__scriptgo_require_module", require_function);
    JS_FreeValue(scriptgo_dynamic_js_context, global);
    if (!scriptgo_dynamic_cleanup_registered) {
        atexit(scriptgo_dynamic_cleanup);
        scriptgo_dynamic_cleanup_registered = 1;
    }
    return 0;
}

static int scriptgo_dynamic_function_call(scriptgo_dynamic_context *context, uint64_t handle,
                                          const scriptgo_value *this_value, const scriptgo_value *arguments,
                                          uint32_t argument_count, const scriptgo_boundary_descriptor *descriptor,
                                          scriptgo_value *out_result, scriptgo_value *out_exception, void *user_data) {
    scriptgo_dynamic_function_ref *ref = (scriptgo_dynamic_function_ref *)(uintptr_t)handle;
    JSValue *argv = NULL;
    JSValue js_this = JS_UNDEFINED;
    JSValue result;
    uint32_t i;
    (void)context; (void)descriptor; (void)user_data;
    if (ref == NULL || out_result == NULL || out_exception == NULL || argument_count > 4)
        return SCRIPTGO_CALL_FATAL;
    scriptgo_value_init_undefined(out_result);
    scriptgo_value_init_undefined(out_exception);
    argv = argument_count == 0 ? NULL : calloc(argument_count, sizeof(*argv));
    if (argument_count != 0 && argv == NULL) return SCRIPTGO_CALL_FATAL;
    for (i = 0; i < argument_count; i++) {
        if (scriptgo_dynamic_to_js(scriptgo_dynamic_js_context, &arguments[i], &argv[i]) != 0) {
            for (i = 0; i < argument_count; i++) JS_FreeValue(scriptgo_dynamic_js_context, argv[i]);
            free(argv);
            return SCRIPTGO_CALL_FATAL;
        }
    }
    if (scriptgo_dynamic_to_js(scriptgo_dynamic_js_context, this_value, &js_this) != 0) {
        for (i = 0; i < argument_count; i++) JS_FreeValue(scriptgo_dynamic_js_context, argv[i]);
        free(argv);
        return SCRIPTGO_CALL_FATAL;
    }
    result = JS_Call(scriptgo_dynamic_js_context, ref->function, js_this, argument_count, argv);
    for (i = 0; i < argument_count; i++) JS_FreeValue(scriptgo_dynamic_js_context, argv[i]);
    free(argv);
    JS_FreeValue(scriptgo_dynamic_js_context, js_this);
    if (JS_IsException(result)) {
        JSValue exception = JS_GetException(scriptgo_dynamic_js_context);
        if (scriptgo_dynamic_from_js(scriptgo_dynamic_js_context, exception, out_exception) != 0) {
            JS_FreeValue(scriptgo_dynamic_js_context, exception);
            return SCRIPTGO_CALL_FATAL;
        }
        JS_FreeValue(scriptgo_dynamic_js_context, exception);
        return SCRIPTGO_CALL_THROWN;
    }
    if (scriptgo_dynamic_from_js(scriptgo_dynamic_js_context, result, out_result) != 0) {
        JS_FreeValue(scriptgo_dynamic_js_context, result);
        return SCRIPTGO_CALL_FATAL;
    }
    JS_FreeValue(scriptgo_dynamic_js_context, result);
    return SCRIPTGO_CALL_OK;
}

int scriptgo_dynamic_engine_call_js(uint64_t handle, JSContext *ctx, JSValueConst this_val,
                                    int argc, JSValueConst *argv, JSValue *out) {
    scriptgo_value this_native = {0};
    scriptgo_value arguments[4] = {{0}};
    scriptgo_value result = {0};
    scriptgo_value exception = {0};
    scriptgo_value_constraint any = {UINT64_C(0x3ff)};
    scriptgo_boundary_descriptor descriptor = {SCRIPTGO_BOUNDARY_FORMAT_V1, 0, NULL, any, any,
        "dynamic-function", "<dynamic>", 0, 1};
    scriptgo_value callable = {SCRIPTGO_TAG_FUNCTION,
        SCRIPTGO_VALUE_OWNED | SCRIPTGO_VALUE_ENGINE_REF, handle, 0};
    int status;
    int i;
    if (out == NULL || argc < 0 || argc > 4) return -1;
    if (scriptgo_dynamic_from_js(ctx, this_val, &this_native) != 0) return -1;
    for (i = 0; i < argc; i++) {
        if (scriptgo_dynamic_from_js(ctx, argv[i], &arguments[i]) != 0) {
            while (i > 0) scriptgo_value_release(&arguments[--i]);
            scriptgo_value_release(&this_native);
            return -1;
        }
    }
    descriptor.parameter_count = (uint32_t)argc;
    descriptor.parameters = argc == 0 ? NULL : calloc((size_t)argc, sizeof(*descriptor.parameters));
    if (argc != 0 && descriptor.parameters == NULL) {
        for (i = 0; i < argc; i++) scriptgo_value_release(&arguments[i]);
        scriptgo_value_release(&this_native);
        return -1;
    }
    if (argc != 0) {
        for (i = 0; i < argc; i++) ((scriptgo_value_constraint *)descriptor.parameters)[i] = any;
    }
    status = scriptgo_dynamic_call(scriptgo_dynamic_value_context, &callable,
        &this_native, arguments, (uint32_t)argc, &descriptor, &result, &exception);
    free((void *)descriptor.parameters);
    for (i = 0; i < argc; i++) scriptgo_value_release(&arguments[i]);
    scriptgo_value_release(&this_native);
    if (status != SCRIPTGO_CALL_OK) {
        scriptgo_value_release(&result);
        scriptgo_value_release(&exception);
        return -1;
    }
    if (scriptgo_dynamic_to_js(ctx, &result, out) != 0) {
        scriptgo_value_release(&result);
        return -1;
    }
    scriptgo_value_release(&result);
    return 0;
}

int scriptgo_dynamic_invoke_function(void *callable, const scriptgo_value *this_value,
                                     int32_t argument_count,
                                     const scriptgo_value *a1, const scriptgo_value *a2,
                                     const scriptgo_value *a3, const scriptgo_value *a4,
                                     int32_t expected_tag, scriptgo_value *out_result) {
    scriptgo_value callable_value = {SCRIPTGO_TAG_FUNCTION,
        SCRIPTGO_VALUE_OWNED | SCRIPTGO_VALUE_ENGINE_REF,
        (uint64_t)(uintptr_t)callable, 0};
    scriptgo_value arguments[4] = {{0}};
    scriptgo_value exception = {0};
    scriptgo_value_constraint any = {UINT64_C(0x3ff)};
    scriptgo_boundary_descriptor descriptor = {SCRIPTGO_BOUNDARY_FORMAT_V1, 0, NULL, any, any,
        "dynamic-function", "<dynamic>", 0, 1};
    uint32_t count = argument_count < 0 ? 0 : (uint32_t)argument_count;
    int32_t status;
    if (count > 4 || out_result == NULL) return scriptgo_runtime_set_error("SG5002: Dynamic call arity mismatch");
    if (count > 0 && a1 != NULL) arguments[0] = *a1;
    if (count > 1 && a2 != NULL) arguments[1] = *a2;
    if (count > 2 && a3 != NULL) arguments[2] = *a3;
    if (count > 3 && a4 != NULL) arguments[3] = *a4;
    descriptor.parameter_count = count;
    descriptor.parameters = count == 0 ? NULL : calloc(count, sizeof(*descriptor.parameters));
    if (count != 0 && descriptor.parameters == NULL) return scriptgo_runtime_set_error("Dynamic call descriptor allocation failed");
    if (count != 0) {
        uint32_t i;
        for (i = 0; i < count; i++) ((scriptgo_value_constraint *)descriptor.parameters)[i] = any;
    }
    status = scriptgo_dynamic_call(scriptgo_dynamic_value_context, &callable_value,
        this_value == NULL ? &(scriptgo_value){SCRIPTGO_TAG_UNDEFINED, 0, 0, 0} : this_value, arguments, count,
        &descriptor, out_result, &exception);
    free((void *)descriptor.parameters);
    if (status != SCRIPTGO_CALL_OK) {
        if (status == SCRIPTGO_CALL_THROWN && exception.tag == SCRIPTGO_TAG_STRING && exception.payload != 0)
            scriptgo_runtime_set_error((const char *)(uintptr_t)exception.payload);
        if (status == SCRIPTGO_CALL_THROWN) {
            scriptgo_value_release(&exception);
            return -1;
        }
        return scriptgo_runtime_set_error("Dynamic function call fatal");
    }
    if (expected_tag >= 0 && (int32_t)out_result->tag != expected_tag) {
        scriptgo_value_release(out_result);
        return scriptgo_runtime_set_error("SG5003: Dynamic result mismatch");
    }
    return 0;
}

static JSValue scriptgo_dynamic_module_export(JSContext *ctx,
                                              scriptgo_dynamic_module *module,
                                              const char *export_name) {
    JSValue namespace_object, exports_object, value;
    if (strcmp(module->kind, "commonjs") == 0) {
        exports_object = scriptgo_dynamic_require_module(ctx, module);
        if (JS_IsException(exports_object)) return exports_object;
        if (strcmp(export_name, "default") == 0 && JS_IsFunction(ctx, exports_object)) return exports_object;
        value = JS_GetPropertyStr(ctx, exports_object, export_name);
        JS_FreeValue(ctx, exports_object);
        return value;
    }
    namespace_object = scriptgo_dynamic_esm_namespace(ctx, module);
    if (JS_IsException(namespace_object)) return namespace_object;
    value = JS_GetPropertyStr(ctx, namespace_object, export_name);
    JS_FreeValue(ctx, namespace_object);
    return value;
}

static JSValue scriptgo_dynamic_esm_namespace(JSContext *ctx, scriptgo_dynamic_module *module) {
    JSValue compiled, evaluated;
    if (module->esm_module == NULL) {
        compiled = JS_Eval(ctx, module->source, strlen(module->source), module->path,
                           JS_EVAL_TYPE_MODULE | JS_EVAL_FLAG_COMPILE_ONLY);
        if (JS_IsException(compiled)) return compiled;
        module->esm_module = (JSModuleDef *)JS_VALUE_GET_PTR(compiled);
        evaluated = JS_EvalFunction(ctx, compiled);
        if (JS_IsException(evaluated)) return evaluated;
        JS_FreeValue(ctx, evaluated);
    }
    return JS_GetModuleNamespace(ctx, module->esm_module);
}

int scriptgo_dynamic_call_module(const char *module_path,
                          const char *export_name, const scriptgo_value *args,
                          int32_t argument_count, int32_t expected_arity,
                          int32_t expected_tag, scriptgo_value *out_result) {
    JSContext *ctx; JSValue *argv = NULL;
	JSValue function, result;
    scriptgo_dynamic_module *module;
    int32_t i, status = 0;
    if (module_path == NULL || export_name == NULL || out_result == NULL || argument_count < 0)
        return scriptgo_runtime_set_error("SG9001: invalid Dynamic call descriptor");
    if (expected_arity < 0 || argument_count != expected_arity)
        return scriptgo_runtime_set_error("SG5002: Dynamic call arity mismatch");
    status = scriptgo_dynamic_ensure_context();
    if (status != 0) return status;
    ctx = scriptgo_dynamic_js_context;
    module = scriptgo_dynamic_find_module(module_path);
    if (module == NULL) return scriptgo_runtime_set_error("SG9001: Dynamic module is not registered");
	function = scriptgo_dynamic_module_export(ctx, module, export_name);
    if (!JS_IsFunction(ctx, function)) {
        JS_FreeValue(ctx, function);
        status = scriptgo_runtime_set_error("Dynamic export is not a function"); goto done;
    }
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
    return status;
}
