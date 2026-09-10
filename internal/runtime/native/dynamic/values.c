/* Boxed-value conversion concatenated after the Dynamic runtime registry. */
static int scriptgo_dynamic_object_to_js(JSContext *ctx, const scriptgo_value *value, JSValue *out) {
    void *keys = NULL;
    int64_t length = 0;
    JSValue object;
    int64_t i;
    if (scriptgo_object_keys((void *)(uintptr_t)value->payload, &keys) != 0 ||
        scriptgo_array_length(keys, &length) != 0) {
        if (keys != NULL) scriptgo_array_release(keys);
        return -1;
    }
    object = JS_NewObject(ctx);
    if (JS_IsException(object)) {
        scriptgo_array_release(keys);
        return -1;
    }
    for (i = 0; i < length; i++) {
        const char *key = NULL;
        scriptgo_value field;
        JSValue field_value;
        if (scriptgo_array_get(keys, (double)i, &key) != 0 || key == NULL ||
            scriptgo_object_property_unknown_get((void *)(uintptr_t)value->payload, key, &field) != 0 ||
            scriptgo_dynamic_to_js(ctx, &field, &field_value) != 0) {
            JS_FreeValue(ctx, object);
            scriptgo_array_release(keys);
            return -1;
        }
        if (JS_SetPropertyStr(ctx, object, key, field_value) < 0) {
            JS_FreeValue(ctx, object);
            scriptgo_value_release(&field);
            scriptgo_array_release(keys);
            return -1;
        }
        scriptgo_value_release(&field);
    }
    scriptgo_array_release(keys);
    *out = object;
    return 0;
}

static int scriptgo_dynamic_array_to_js(JSContext *ctx, const scriptgo_value *value, JSValue *out) {
    int64_t length = 0;
    int64_t i;
    JSValue array;
    if (scriptgo_array_length((void *)(uintptr_t)value->payload, &length) != 0) return -1;
    array = JS_NewArray(ctx);
    if (JS_IsException(array)) return -1;
    for (i = 0; i < length; i++) {
        scriptgo_value element;
        JSValue element_value;
        if (scriptgo_array_get_unknown((void *)(uintptr_t)value->payload, (double)i, &element) != 0 ||
            scriptgo_dynamic_to_js(ctx, &element, &element_value) != 0) {
            JS_FreeValue(ctx, array);
            return -1;
        }
        if (JS_SetPropertyUint32(ctx, array, (uint32_t)i, element_value) < 0) {
            JS_FreeValue(ctx, array);
            scriptgo_value_release(&element);
            return -1;
        }
        scriptgo_value_release(&element);
    }
    *out = array;
    return 0;
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
    case SCRIPTGO_TAG_OBJECT:
        return scriptgo_dynamic_object_to_js(ctx, value, out);
    case SCRIPTGO_TAG_ARRAY:
        return scriptgo_dynamic_array_to_js(ctx, value, out);
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
    if (JS_IsFunction(ctx, value)) return -1;
    if (JS_IsArray(value)) {
        int64_t array_length;
        void *array = NULL;
        int64_t i;
        if (JS_GetLength(ctx, value, &array_length) < 0 || scriptgo_array_new(array_length, sizeof(scriptgo_value), &array) != 0) return -1;
        for (i = 0; i < array_length; i++) {
            JSValue element = JS_GetPropertyUint32(ctx, value, (uint32_t)i);
            scriptgo_value native_element;
            int status;
            if (JS_IsException(element)) { JS_FreeValue(ctx, element); scriptgo_array_release(array); return -1; }
            status = scriptgo_dynamic_from_js(ctx, element, &native_element);
            JS_FreeValue(ctx, element);
            if (status != 0 || scriptgo_array_set(array, (double)i, &native_element) != 0) {
                if (status == 0) scriptgo_value_release(&native_element);
                scriptgo_array_release(array);
                return -1;
            }
            scriptgo_value_release(&native_element);
        }
        out->tag = SCRIPTGO_TAG_ARRAY; out->flags = 0;
        out->payload = (uint64_t)(uintptr_t)array; return 0;
    }
    if (JS_IsObject(value)) {
        JSPropertyEnum *properties = NULL;
        uint32_t count = 0;
        void *object = NULL;
        uint32_t i;
        if (JS_GetOwnPropertyNames(ctx, &properties, &count, value, JS_GPN_STRING_MASK | JS_GPN_ENUM_ONLY) < 0 ||
            scriptgo_object_new(0, &object) != 0 || scriptgo_object_type_set(object, "__json__") != 0) {
            if (properties != NULL) JS_FreePropertyEnum(ctx, properties, count);
            return -1;
        }
        for (i = 0; i < count; i++) {
            const char *key = JS_AtomToCString(ctx, properties[i].atom);
            JSValue property;
            scriptgo_value native_property;
            int status;
            if (key == NULL) { scriptgo_object_release(object); JS_FreePropertyEnum(ctx, properties, count); return -1; }
            property = JS_GetProperty(ctx, value, properties[i].atom);
            status = JS_IsException(property) ? -1 : scriptgo_dynamic_from_js(ctx, property, &native_property);
            JS_FreeValue(ctx, property);
            if (status != 0 || scriptgo_object_property_unknown_set(object, key, &native_property) != 0) {
                if (status == 0) scriptgo_value_release(&native_property);
                scriptgo_object_release(object);
                JS_FreePropertyEnum(ctx, properties, count);
                return -1;
            }
            JS_FreeCString(ctx, key);
        }
        JS_FreePropertyEnum(ctx, properties, count);
        out->tag = SCRIPTGO_TAG_OBJECT; out->flags = 0;
        out->payload = (uint64_t)(uintptr_t)object; return 0;
    }
    (void)tag;
    return -1;
}
