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
	harnessPath := filepath.Join(dir, "harness.c")
	executable := filepath.Join(dir, "harness")
	if err := os.WriteFile(runtimePath, runtime.Source, 0o644); err != nil {
		t.Fatal(err)
	}
	harness := `
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
int scriptgo_async_frame_set(void *, int64_t, uint32_t, uint64_t);
int scriptgo_async_frame_get(void *, int64_t, uint32_t *, uint64_t *);
int scriptgo_async_frame_release(void *);
int scriptgo_promise_create(void **);
int scriptgo_promise_resolve_number(void *, double);
int scriptgo_promise_then(void *, void *, void *, uint32_t, void **);
int scriptgo_promise_await_number(void *, double *);
int scriptgo_promise_schedule_resume(void *, void *);
int scriptgo_event_loop_run(void);
int scriptgo_closure_create(void *, void *, void **);

static int resume_calls = 0;
static uint32_t resume_tag = 0;
static uint64_t resume_payload = 0;
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
    int32_t tag0, int32_t pad0, int64_t payload0,
    int32_t tag1, int32_t pad1, int64_t payload1,
    int32_t tag2, int32_t pad2, int64_t payload2,
    int32_t tag3, int32_t pad3, int64_t payload3) {
    (void)env; (void)pad0; (void)tag1; (void)pad1; (void)payload1;
    (void)tag2; (void)pad2; (void)payload2; (void)tag3; (void)pad3; (void)payload3;
    resume_calls++;
    resume_tag = (uint32_t)tag0;
    resume_payload = (uint64_t)payload0;
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
    uint32_t frame_tag = 0;
    uint64_t frame_payload = 0;
    if (scriptgo_async_frame_new(1, &frame) != 0) return 24;
    if (scriptgo_async_frame_set(frame, 0, 3, 0x1234) != 0) return 25;
    if (scriptgo_async_frame_get(frame, 0, &frame_tag, &frame_payload) != 0 || frame_tag != 3 || frame_payload != 0x1234) return 26;
    if (scriptgo_async_frame_release(frame) != 0) return 27;
    if (scriptgo_promise_create(&promise) != 0) return 28;
    if (scriptgo_closure_create((void *)resume_callback, NULL, &closure) != 0) return 29;
    if (scriptgo_promise_schedule_resume(promise, closure) != 0) return 30;
    if (scriptgo_promise_resolve_number(promise, 42.0) != 0) return 31;
    if (resume_calls != 0) return 32;
    if (scriptgo_event_loop_run() != 0 || resume_calls != 1 || resume_tag != 3) return 33;

    void *source = NULL, *derived = NULL, *transform = NULL;
    double source_value = 0, derived_value = 0;
    if (scriptgo_promise_create(&source) != 0) return 34;
    if (scriptgo_closure_create((void *)add_one_callback, NULL, &transform) != 0) return 35;
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
	if output, err := exec.Command(clang, harnessPath, runtimePath, "-o", executable, "-lm", "-lresolv").CombinedOutput(); err != nil {
		t.Fatalf("clang: %v\n%s", err, output)
	}
	output, err := exec.Command(executable).CombinedOutput()
	if err != nil {
		t.Fatalf("ABI harness: %v\n%s", err, output)
	}
	if strings.TrimSpace(string(output)) != "ok" {
		t.Fatalf("ABI harness output = %q", output)
	}
}
