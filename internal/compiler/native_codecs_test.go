package compiler

import "testing"

func TestNativeCodecConfigSkipsCrossTargets(t *testing.T) {
	config := nativeCodecConfigForTarget("wasm32-wasi")
	if len(config.compileFlags) != 0 || len(config.linkFlags) != 0 {
		t.Fatalf("cross-target codec config = %+v, want no host flags", config)
	}
}

func TestAppendUniquePreservesFirstOccurrence(t *testing.T) {
	got := appendUnique([]string{"-I/a", "-lz"}, "-I/a", "-L/b", "-lzstd", "-L/b")
	want := []string{"-I/a", "-lz", "-L/b", "-lzstd"}
	if len(got) != len(want) {
		t.Fatalf("appendUnique() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("appendUnique() = %v, want %v", got, want)
		}
	}
}
