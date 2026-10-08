package compiler

import (
	"path/filepath"
	"testing"
)

// TestCompileIsDeterministicAcrossRuns guards reproducible builds: the typed
// IR and LLVM IR for the same input must be byte-identical on every run. The
// inputs exercise class hierarchies with inherited accessors, structural
// object shapes, and bounds-check elimination, which previously depended on
// Go map iteration order.
func TestCompileIsDeterministicAcrossRuns(t *testing.T) {
	inputs := []string{
		filepath.Join("..", "..", "examples", "classes_oop.ts"),
		filepath.Join("testdata", "corpus", "tuples", "tuple_zip_unzip.ts"),
		filepath.Join("testdata", "corpus", "language", "deep_generics_and_monomorphization.ts"),
	}
	const runs = 6
	for _, input := range inputs {
		t.Run(filepath.Base(input), func(t *testing.T) {
			firstIR, err := DumpIR(input)
			if err != nil {
				t.Fatal(err)
			}
			firstLLVM, err := Compile(input)
			if err != nil {
				t.Fatal(err)
			}
			for run := 1; run < runs; run++ {
				typedIR, err := DumpIR(input)
				if err != nil {
					t.Fatal(err)
				}
				if typedIR != firstIR {
					t.Fatalf("typed IR differs between runs (run %d)", run)
				}
				llvmIR, err := Compile(input)
				if err != nil {
					t.Fatal(err)
				}
				if llvmIR != firstLLVM {
					t.Fatalf("LLVM IR differs between runs (run %d)", run)
				}
			}
		})
	}
}
