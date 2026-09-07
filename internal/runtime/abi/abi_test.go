package abi

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pilotworks/scriptgo/internal/runtime"
)

func TestRuntimeABIv1(t *testing.T) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("clang is not installed")
	}
	dir := t.TempDir()
	runtimePath := filepath.Join(dir, "runtime.c")
	headerPath := filepath.Join(dir, "scriptgo_value.h")
	harnessPath := filepath.Join(dir, "harness.c")
	executable := filepath.Join(dir, "harness")
	if err := os.WriteFile(runtimePath, runtime.Source, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(headerPath, []byte(runtime.ValueHeader), 0o644); err != nil {
		t.Fatal(err)
	}
	harness := `
#include "scriptgo_value.h"
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int scriptgo_array_new(int64_t, int64_t, void **);
int scriptgo_array_set(void *, double, const void *);
int scriptgo_array_get(void *, double, void *);
int scriptgo_array_length(void *, int64_t *);
int scriptgo_array_release(void *);
int scriptgo_object_new(int64_t, void **);
int scriptgo_object_number_set(void *, int64_t, double);
int scriptgo_object_number_get(void *, int64_t, double *);
int scriptgo_object_release(void *);
int scriptgo_string_concat(const char *, const char *, char **);
int scriptgo_string_length(const char *, double *);
int scriptgo_string_last_index(const char *, const char *, double, double *);
int scriptgo_string_slice(const char *, double, double, char **);
int scriptgo_string_release(char *);
int scriptgo_async_frame_new(int64_t, void **);
int scriptgo_async_frame_set(void *, int64_t, const scriptgo_value *);
int scriptgo_async_frame_get(void *, int64_t, scriptgo_value *);
int scriptgo_async_frame_release(void *);
int scriptgo_promise_create(void **);
int scriptgo_promise_resolve_number(void *, double);
int scriptgo_promise_then(void *, void *, void *, uint32_t, void **);
int scriptgo_promise_await_number(void *, double *);
int scriptgo_promise_schedule_resume(void *, void *);
int scriptgo_event_loop_run(void);
int scriptgo_closure_create(void *, void *, void *, int32_t, void **);
int scriptgo_closure_invoke(void *, int32_t, const scriptgo_value *, const scriptgo_value *, const scriptgo_value *, const scriptgo_value *);
int scriptgo_closure_invoke_value(void *, int32_t, const scriptgo_value *, const scriptgo_value *, const scriptgo_value *, const scriptgo_value *, scriptgo_value *);

static int resume_calls = 0;
static uint32_t resume_tag = 0;
static uint64_t resume_payload = 0;
static int canonical_calls = 0;
static int void_calls = 0;
static scriptgo_value canonical_callback(void *env,
    uint32_t tag0, uint32_t flags0, uint64_t payload0,
    uint32_t tag1, uint32_t flags1, uint64_t payload1,
    uint32_t tag2, uint32_t flags2, uint64_t payload2,
    uint32_t tag3, uint32_t flags3, uint64_t payload3) {
    scriptgo_value result = {0};
    (void)env; (void)flags0; (void)tag1; (void)flags1; (void)payload1;
    (void)tag2; (void)flags2; (void)payload2; (void)tag3; (void)flags3; (void)payload3;
    canonical_calls = tag0 == SCRIPTGO_TAG_NUMBER && payload0 == 0x5678 ? 1 : -1;
    result.tag = SCRIPTGO_TAG_UNDEFINED;
    return result;
}
static void void_callback(void *env,
    uint32_t tag0, uint32_t flags0, uint64_t payload0,
    uint32_t tag1, uint32_t flags1, uint64_t payload1,
    uint32_t tag2, uint32_t flags2, uint64_t payload2,
    uint32_t tag3, uint32_t flags3, uint64_t payload3) {
    (void)env; (void)flags0; (void)tag1; (void)flags1; (void)payload1;
    (void)tag2; (void)flags2; (void)payload2; (void)tag3; (void)flags3; (void)payload3;
    void_calls = tag0 == SCRIPTGO_TAG_NUMBER && payload0 == 0x5678 ? 1 : -1;
}
static double add_one_callback(void *env,
    uint32_t tag0, uint32_t pad0, uint64_t payload0,
    uint32_t tag1, uint32_t pad1, uint64_t payload1,
    uint32_t tag2, uint32_t pad2, uint64_t payload2,
    uint32_t tag3, uint32_t pad3, uint64_t payload3) {
    double value = 0;
    (void)env; (void)pad0; (void)tag1; (void)pad1; (void)payload1;
    (void)tag2; (void)pad2; (void)payload2; (void)tag3; (void)pad3; (void)payload3;
    memcpy(&value, &payload0, sizeof(value));
    return value + 1;
}
static void resume_callback(void *env,
    uint32_t tag0, uint32_t pad0, uint64_t payload0,
    uint32_t tag1, uint32_t pad1, uint64_t payload1,
    uint32_t tag2, uint32_t pad2, uint64_t payload2,
    uint32_t tag3, uint32_t pad3, uint64_t payload3) {
    (void)env; (void)pad0; (void)tag1; (void)pad1; (void)payload1;
    (void)tag2; (void)pad2; (void)payload2; (void)tag3; (void)pad3; (void)payload3;
    resume_calls++;
    resume_tag = tag0;
    resume_payload = payload0;
}

int main(void) {
    void *array = NULL, *string_array = NULL, *object = NULL, *frame = NULL, *promise = NULL, *closure = NULL;
    int64_t length = -1;
    double number = 0, string_length = 0, index = 0;
    char *joined = NULL, *slice = NULL;
    char *word = "ok", *out_word = NULL;
    if (scriptgo_array_new(2, sizeof(double), &array) != 0) return 1;
    if (scriptgo_array_length(array, &length) != 0 || length != 2) return 2;
    number = 42;
    if (scriptgo_array_set(array, 1, &number) != 0) return 3;
    number = 0;
    if (scriptgo_array_get(array, 1, &number) != 0 || number != 42) return 4;
    if (scriptgo_array_get(array, 1.5, &number) >= 0) return 5;
    if (scriptgo_array_release(array) != 0) return 6;
    if (scriptgo_array_new(0, sizeof(void *), &string_array) != 0) return 7;
    if (scriptgo_array_length(string_array, &length) != 0 || length != 0) return 8;
    if (scriptgo_array_release(string_array) != 0) return 9;
    if (scriptgo_array_new(1, sizeof(void *), &string_array) != 0) return 10;
    if (scriptgo_array_set(string_array, 0, &word) != 0) return 11;
    if (scriptgo_array_get(string_array, 0, &out_word) != 0 || out_word != word) return 12;
    if (scriptgo_array_release(string_array) != 0) return 13;
    if (scriptgo_object_new(1, &object) != 0) return 14;
    if (scriptgo_object_number_set(object, 0, 7) != 0) return 15;
    if (scriptgo_object_number_get(object, 0, &number) != 0 || number != 7) return 16;
    if (scriptgo_object_release(object) != 0) return 17;
    if (scriptgo_string_concat("ab", "cd", &joined) != 0 || strcmp(joined, "abcd") != 0) return 18;
    if (scriptgo_string_length(joined, &string_length) != 0 || string_length != 4) return 19;
    if (scriptgo_string_last_index(joined, "b", -1, &index) != 0 || index != 1) return 20;
    if (scriptgo_string_slice(joined, 1, 3, &slice) != 0 || strcmp(slice, "bc") != 0) return 21;
    scriptgo_string_release(slice);
    if (scriptgo_string_last_index(joined, "", 1, &index) != 0 || index != 1) return 22;
    if (scriptgo_string_slice(joined, -4, 99, &slice) != 0 || strcmp(slice, "abcd") != 0) return 23;
    scriptgo_string_release(slice);
    scriptgo_string_release(joined);
    scriptgo_value frame_value;
    if (scriptgo_async_frame_new(1, &frame) != 0) return 24;
    frame_value.tag = SCRIPTGO_TAG_NUMBER;
    frame_value.flags = 0;
    frame_value.payload = 0x1234;
    frame_value.aux = 0;
    if (scriptgo_async_frame_set(frame, 0, &frame_value) != 0) return 25;
    scriptgo_value_init_undefined(&frame_value);
    if (scriptgo_async_frame_get(frame, 0, &frame_value) != 0 || frame_value.tag != SCRIPTGO_TAG_NUMBER || frame_value.payload != 0x1234) return 26;
    scriptgo_value_release(&frame_value);
    if (scriptgo_value_string_copy("frame", 5, &frame_value) != 0) return 27;
    if (scriptgo_async_frame_set(frame, 0, &frame_value) != 0) return 28;
    scriptgo_value_release(&frame_value);
    if (scriptgo_async_frame_get(frame, 0, &frame_value) != 0 || frame_value.tag != SCRIPTGO_TAG_STRING || frame_value.aux != 5 ||
        memcmp((void *)(uintptr_t)frame_value.payload, "frame", 5) != 0) return 29;
    scriptgo_value_release(&frame_value);
    if (scriptgo_async_frame_release(frame) != 0) return 30;
    void *canonical_closure = NULL;
    scriptgo_value canonical_arg = {SCRIPTGO_TAG_NUMBER, 0, 0x5678, 0};
    if (scriptgo_closure_create((void *)canonical_callback, NULL, NULL, -1, &canonical_closure) != 0) return 31;
    if (scriptgo_closure_invoke(canonical_closure, 1, &canonical_arg, NULL, NULL, NULL) != 0 || canonical_calls != 1) return 32;
    void *void_closure = NULL;
    scriptgo_value callback_result = {0};
    if (scriptgo_closure_create((void *)void_callback, NULL, NULL, SCRIPTGO_TAG_UNDEFINED, &void_closure) != 0) return 33;
    if (scriptgo_closure_invoke_value(void_closure, 1, &canonical_arg, NULL, NULL, NULL, &callback_result) != 0 ||
        void_calls != 1 || callback_result.tag != SCRIPTGO_TAG_UNDEFINED) return 34;
    void *number_closure = NULL;
    if (scriptgo_closure_create((void *)add_one_callback, NULL, NULL, SCRIPTGO_TAG_NUMBER, &number_closure) != 0) return 35;
    if (scriptgo_closure_invoke_value(number_closure, 1, &canonical_arg, NULL, NULL, NULL, &callback_result) != 0 ||
        callback_result.tag != SCRIPTGO_TAG_NUMBER) return 36;
    double callback_number = 0;
    memcpy(&callback_number, &callback_result.payload, sizeof(callback_number));
    if (callback_number != 1.0) return 37;
    if (scriptgo_promise_create(&promise) != 0) return 28;
    if (scriptgo_closure_create((void *)resume_callback, NULL, NULL, 0, &closure) != 0) return 29;
    if (scriptgo_promise_schedule_resume(promise, closure) != 0) return 30;
    if (scriptgo_promise_resolve_number(promise, 42.0) != 0) return 31;
    if (resume_calls != 0) return 32;
    if (scriptgo_event_loop_run() != 0 || resume_calls != 1 || resume_tag != 3) return 33;

    void *source = NULL, *derived = NULL, *transform = NULL;
    double source_value = 0, derived_value = 0;
    if (scriptgo_promise_create(&source) != 0) return 34;
    if (scriptgo_closure_create((void *)add_one_callback, NULL, NULL, SCRIPTGO_TAG_NUMBER, &transform) != 0) return 35;
    if (scriptgo_promise_then(source, transform, NULL, 3, &derived) != 0 || derived == source) return 36;
    if (scriptgo_promise_resolve_number(source, 41.0) != 0) return 37;
    if (scriptgo_event_loop_run() != 0) return 38;
    if (scriptgo_promise_await_number(source, &source_value) != 0 || source_value != 41.0) return 39;
    if (scriptgo_promise_await_number(derived, &derived_value) != 0 || derived_value != 42.0) return 40;
    puts("ok");
    return 0;
}
`
	if err := os.WriteFile(harnessPath, []byte(harness), 0o644); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(clang, "-fsanitize=address,undefined", "-fno-omit-frame-pointer", harnessPath, runtimePath, "-I", dir, "-o", executable, "-lm", "-lresolv").CombinedOutput(); err != nil {
		t.Fatalf("clang: %v\n%s", err, output)
	}
	command := exec.Command(executable)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("ABI harness: %v\n%s", err, output)
	}
	if strings.TrimSpace(string(output)) != "ok" {
		t.Fatalf("ABI harness output = %q", output)
	}
}

func TestBoxedValueBoundaryABIv1(t *testing.T) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("clang is not installed")
	}
	dir := t.TempDir()
	runtimePath := filepath.Join(dir, "runtime.c")
	headerPath := filepath.Join(dir, "scriptgo_value.h")
	harnessPath := filepath.Join(dir, "boxed_harness.c")
	executable := filepath.Join(dir, "boxed_harness")
	if err := os.WriteFile(runtimePath, runtime.Source, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(headerPath, []byte(runtime.ValueHeader), 0o644); err != nil {
		t.Fatal(err)
	}
	harness := `
#include "scriptgo_value.h"
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <setjmp.h>

int scriptgo_promise_create(void **);

static int retained;
static int released;
static int mode;
static uint64_t observed_callable;

static void retain_ref(scriptgo_dynamic_context *context, uint64_t handle) {
    (void)context; (void)handle; retained++;
}
static void release_ref(scriptgo_dynamic_context *context, uint64_t handle) {
    (void)context; (void)handle; released++;
}
static int32_t fake_call(scriptgo_dynamic_context *context, uint64_t callable,
    const scriptgo_value *this_value, const scriptgo_value *arguments,
    uint32_t argument_count, const scriptgo_boundary_descriptor *descriptor,
    scriptgo_value *out_result, scriptgo_value *out_exception, void *user_data) {
    double number = 42.5;
    (void)this_value; (void)arguments; (void)argument_count;
    (void)descriptor; (void)user_data;
    observed_callable = callable;
    if (mode == 1) {
        if (scriptgo_value_adopt_engine_ref(context, SCRIPTGO_TAG_OBJECT, 77, out_exception) != 0) return SCRIPTGO_CALL_FATAL;
        return SCRIPTGO_CALL_THROWN;
    }
    if (mode == 2) {
        out_result->tag = SCRIPTGO_TAG_BOOLEAN;
        out_result->payload = 1;
        return SCRIPTGO_CALL_OK;
    }
    if (mode == 3) {
        out_result->tag = 99;
        return SCRIPTGO_CALL_OK;
    }
    out_result->tag = SCRIPTGO_TAG_NUMBER;
    memcpy(&out_result->payload, &number, sizeof(number));
    return SCRIPTGO_CALL_OK;
}

static int is_error(const scriptgo_value *value, const char *code) {
    size_t i, j, length = strlen(code);
    const unsigned char *bytes = (const unsigned char *)(uintptr_t)value->payload;
    if (value->tag != SCRIPTGO_TAG_STRING || value->payload == 0 || value->aux < length) return 1;
    for (i = 0; i + length <= value->aux; i++) {
        for (j = 0; j < length && bytes[i + j] == (unsigned char)code[j]; j++) {}
        if (j == length) return 0;
    }
    return 1;
}

int main(void) {
    scriptgo_dynamic_context *context;
    scriptgo_value callable, copied, moved, this_value, argument, result, exception;
    scriptgo_value_constraint parameter = { UINT64_C(1) << SCRIPTGO_TAG_NUMBER };
    scriptgo_boundary_descriptor descriptor = {
        SCRIPTGO_BOUNDARY_FORMAT_V1, 1, &parameter,
        { UINT64_C(1) << SCRIPTGO_TAG_UNDEFINED },
        { UINT64_C(1) << SCRIPTGO_TAG_NUMBER },
        "fake", "fake.ts", 10, 3
    };
    double number = 7.0;
    uint64_t number_bits;
    const unsigned char embedded[] = {'a', 0, 'b'};
    scriptgo_value string_value;
    scriptgo_exception_frame_t *frame;
    scriptgo_value thrown, taken;
    memcpy(&number_bits, &number, sizeof(number));

    if (sizeof(scriptgo_value) != 24 || _Alignof(scriptgo_value) != 8 ||
        offsetof(scriptgo_value, tag) != 0 || offsetof(scriptgo_value, flags) != 4 ||
        offsetof(scriptgo_value, payload) != 8 || offsetof(scriptgo_value, aux) != 16) return 1;
    scriptgo_value invalid_native = {SCRIPTGO_TAG_OBJECT, SCRIPTGO_VALUE_OWNED, 1, 0};
    if (scriptgo_value_validate(&invalid_native) == 0) return 1;
    if (scriptgo_value_string_copy(embedded, sizeof(embedded), &string_value) != 0 ||
        string_value.aux != 3 || memcmp((void *)(uintptr_t)string_value.payload, embedded, 3) != 0) return 2;
    if (scriptgo_value_validate(&string_value) != 0 || scriptgo_value_clone(&copied, &string_value) != 0) return 3;
    if (copied.payload == string_value.payload || copied.aux != 3) return 4;
    if (scriptgo_value_move(&moved, &copied) != 0 || copied.tag != SCRIPTGO_TAG_UNDEFINED) return 5;
    if (scriptgo_value_release(&moved) != 0 || scriptgo_value_release(&moved) != 0) return 6;
    scriptgo_value_release(&string_value);

    frame = scriptgo_exception_frame_new();
    if (frame == NULL) return 7;
    if (setjmp(*(jmp_buf *)scriptgo_exception_buf(frame)) == 0) {
        if (scriptgo_value_string_copy("owned", 5, &thrown) != 0) return 7;
        scriptgo_exception_throw_move(&thrown);
        return 7;
    }
    scriptgo_value_init_undefined(&taken);
    scriptgo_exception_take_value(frame, &taken);
    scriptgo_exception_frame_free(frame);
    if (taken.tag != SCRIPTGO_TAG_STRING || taken.aux != 5 ||
        memcmp((void *)(uintptr_t)taken.payload, "owned", 5) != 0) return 7;
    scriptgo_value_release(&taken);

    context = scriptgo_dynamic_context_new(NULL, NULL, retain_ref, release_ref, fake_call);
    if (context == NULL) return 8;
    if (scriptgo_value_adopt_engine_ref(context, SCRIPTGO_TAG_FUNCTION, 42, &callable) != 0) return 9;
    if (scriptgo_value_clone(&copied, &callable) != 0 || retained != 1 ||
        scriptgo_dynamic_context_live_refs(context) != 2) return 9;
    if (scriptgo_value_release(&copied) != 0 || released != 1 ||
        scriptgo_dynamic_context_live_refs(context) != 1) return 10;
    scriptgo_value_init_undefined(&this_value);
    argument.tag = SCRIPTGO_TAG_NUMBER; argument.flags = 0; argument.payload = number_bits; argument.aux = 0;
    mode = 0;
    if (scriptgo_dynamic_call(context, &callable, &this_value, &argument, 1, &descriptor, &result, &exception) != 0 ||
        result.tag != SCRIPTGO_TAG_NUMBER || observed_callable != 42 || exception.tag != SCRIPTGO_TAG_UNDEFINED) return 11;
    scriptgo_value_release(&result);
    descriptor.parameters = &(scriptgo_value_constraint){ UINT64_C(1) << SCRIPTGO_TAG_BOOLEAN };
    if (scriptgo_dynamic_call(context, &callable, &this_value, &argument, 1, &descriptor, &result, &exception) != SCRIPTGO_CALL_THROWN ||
        is_error(&exception, "SG5002")) return 12;
    scriptgo_value_release(&exception);
    descriptor.parameters = &parameter;
    mode = 2;
    if (scriptgo_dynamic_call(context, &callable, &this_value, &argument, 1, &descriptor, &result, &exception) != SCRIPTGO_CALL_THROWN ||
        is_error(&exception, "SG5003")) return 13;
    scriptgo_value_release(&exception);
    descriptor.result.allowed_tags = UINT64_C(1) << SCRIPTGO_TAG_NUMBER;
    mode = 3;
    if (scriptgo_dynamic_call(context, &callable, &this_value, &argument, 1, &descriptor, &result, &exception) != SCRIPTGO_CALL_FATAL ||
        strstr(scriptgo_runtime_last_error(), "SG9001") == NULL) return 14;
    mode = 1;
    if (scriptgo_dynamic_call(context, &callable, &this_value, &argument, 1, &descriptor, &result, &exception) != SCRIPTGO_CALL_THROWN ||
        exception.tag != SCRIPTGO_TAG_OBJECT || (exception.flags & SCRIPTGO_VALUE_ENGINE_REF) == 0) return 15;
    scriptgo_value_release(&exception);
    if (scriptgo_dynamic_context_shutdown(context) == 0) return 16;
    if (scriptgo_value_release(&callable) != 0 || released != 3 ||
        scriptgo_dynamic_context_shutdown(context) != 0 || scriptgo_dynamic_context_destroy(context) != 0) return 17;
    void *canonical_promise = NULL;
    scriptgo_value promise_input, promise_output;
    promise_input.tag = SCRIPTGO_TAG_NUMBER;
    promise_input.flags = 0;
    promise_input.payload = number_bits;
    promise_input.aux = 0;
    if (scriptgo_promise_create(&canonical_promise) != 0 ||
        scriptgo_promise_resolve_value(canonical_promise, &promise_input) != 0 ||
        scriptgo_promise_await_value(canonical_promise, &promise_output) != 0 ||
        promise_output.tag != SCRIPTGO_TAG_NUMBER || promise_output.payload != number_bits) return 18;
    scriptgo_value_init_undefined(&promise_output);
    if (scriptgo_promise_await_unknown_value(&promise_input, &promise_output) != 0 ||
        promise_output.tag != SCRIPTGO_TAG_NUMBER || promise_output.payload != number_bits) return 19;
    puts("ok");
    return 0;
}
`
	if err := os.WriteFile(harnessPath, []byte(harness), 0o644); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(clang, "-fsanitize=address,undefined", "-fno-omit-frame-pointer", harnessPath, runtimePath, "-I", dir, "-o", executable, "-lm", "-lresolv").CombinedOutput(); err != nil {
		t.Fatalf("clang: %v\n%s", err, output)
	}
	command := exec.Command(executable)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("boxed ABI harness: %v\n%s", err, output)
	}
	if strings.TrimSpace(string(output)) != "ok" {
		t.Fatalf("boxed ABI harness output = %q", output)
	}
}
