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

func TestNativeCodecConfigHasDefine(t *testing.T) {
	config := nativeCodecConfig{compileFlags: []string{"-DSCRIPTGO_HAS_OPENSSL", "-DOTHER=value"}}
	if !config.hasDefine("SCRIPTGO_HAS_OPENSSL") {
		t.Fatal("hasDefine() did not find an exact define")
	}
	if !config.hasDefine("OTHER") {
		t.Fatal("hasDefine() did not find a define with a value")
	}
	if !(nativeCodecConfig{compileFlags: []string{"-DSCRIPTGO_HAS_OPENSSL=1"}}).hasDefine("SCRIPTGO_HAS_OPENSSL") {
		t.Fatal("hasDefine() did not find a define with a value")
	}
}
