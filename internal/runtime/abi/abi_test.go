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

int main(void) {
    void *array = NULL, *string_array = NULL, *object = NULL;
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
    puts("ok");
    return 0;
}
`
	if err := os.WriteFile(harnessPath, []byte(harness), 0o644); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(clang, harnessPath, runtimePath, "-o", executable).CombinedOutput(); err != nil {
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
