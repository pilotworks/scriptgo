package ir

import "testing"

func TestDumpIsDeterministic(t *testing.T) {
	module := Module{
		SourcePath:     "/tmp/main.ts",
		StatementCount: 2,
		Functions: []Function{{
			Name:       "main",
			Span:       SourceSpan{Path: "/tmp/main.ts", Offset: 0, Length: 42},
			ReturnType: TypeVoid,
			Body: []Instruction{
				{Op: OpConst, Type: TypeNumber, Result: "t0", Value: "20", Span: SourceSpan{Path: "/tmp/main.ts", Offset: 10, Length: 2}},
				{Op: OpPrint, Type: TypeVoid, Args: []string{"t0"}, Span: SourceSpan{Path: "/tmp/main.ts", Offset: 0, Length: 18}},
				{Op: OpReturn, Type: TypeVoid, Span: SourceSpan{Path: "/tmp/main.ts", Offset: 20, Length: 1}},
			},
		}},
	}
	got, err := module.Dump()
	if err != nil {
		t.Fatal(err)
	}
	want := "module \"/tmp/main.ts\" statements=2\n" +
		"function main() -> void @/tmp/main.ts:0+42\n" +
		"  %t0: number = const \"20\" @/tmp/main.ts:10+2\n" +
		"  print [%t0] @/tmp/main.ts:0+18\n" +
		"  return @/tmp/main.ts:20+1\n"
	if got != want {
		t.Fatalf("Dump() =\n%swant\n%s", got, want)
	}
}
