package lowering

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	typescriptgo "github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func TestLowerMatchesMVPGoldenIR(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "function add(a: number, b: number): number { return a + b; }\nconsole.log(add(20, 22));\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}
	module, err := Lower(program)
	if err != nil {
		t.Fatal(err)
	}

	want, err := os.ReadFile(filepath.Join("testdata", "mvp.ir"))
	if err != nil {
		t.Fatal(err)
	}
	got := formatModule(module)
	if got != string(want) {
		t.Fatalf("lowered IR differs from golden:\n got:\n%s\nwant:\n%s", got, want)
	}
}

func formatModule(module ir.Module) string {
	var builder strings.Builder
	for _, function := range module.Functions {
		fmt.Fprintf(&builder, "function %s -> %s span=%d:%d\n", function.Name, function.ReturnType, function.Span.Offset, function.Span.Length)
		for _, instruction := range function.Body {
			fmt.Fprintf(&builder, "  %s|%s|%s|%s|%s|%s|%s|%d:%d\n",
				instruction.Op, instruction.Type, instruction.Result, instruction.Value,
				instruction.Operator, instruction.Callee, strings.Join(instruction.Args, ","),
				instruction.Span.Offset, instruction.Span.Length)
		}
	}
	return builder.String()
}

func TestLowerHelloProgram(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "const answer: number = 20 + 22;\nconsole.log(answer);\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}

	module, err := Lower(program)
	if err != nil {
		t.Fatal(err)
	}
	if err := module.Verify(); err != nil {
		t.Fatal(err)
	}
	if len(module.Functions) != 1 || module.Functions[0].Name != "main" {
		t.Fatalf("unexpected functions: %+v", module.Functions)
	}
	seenBinary := false
	seenPrint := false
	for _, instruction := range module.Functions[0].Body {
		seenBinary = seenBinary || instruction.Op == ir.OpBinary
		seenPrint = seenPrint || instruction.Op == ir.OpPrint
	}
	if !seenBinary || !seenPrint {
		t.Fatalf("lowered body has no binary and print operations: %+v", module.Functions[0].Body)
	}
	wantOffset := strings.Index(source, "20") - 1 // TypeScript-Go node.Pos includes the literal's leading trivia.
	if module.Functions[0].Body[0].Span.Path != entry || module.Functions[0].Body[0].Span.Offset != wantOffset {
		t.Fatalf("constant span = %+v, want path %q and offset %d", module.Functions[0].Body[0].Span, entry, wantOffset)
	}
}

func TestLowerRejectsUnsupportedStatementBeforeIR(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	program := frontend.Program{
		Files: []typescriptgo.SourceFile{
			{
				FileName: entry,
				Syntax: typescriptgo.SyntaxFile{
					FileName: entry,
					Statements: []typescriptgo.SyntaxStatement{
						{
							Kind: "unsupported",
							Type: "CustomUnsupportedStatement",
						},
					},
				},
			},
		},
	}
	_, err := Lower(program)
	if err == nil || !strings.Contains(err.Error(), "native subset") || !strings.Contains(err.Error(), "CustomUnsupportedStatement") || !strings.Contains(err.Error(), entry) {
		t.Fatalf("Lower error = %v, want actionable native subset diagnostic", err)
	}
}

func TestLowerRejectsAnyInStaticMode(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "const value: any = 1;\nconsole.log(value);\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(program)
	if err == nil || !strings.Contains(err.Error(), "SG1001") || !strings.Contains(err.Error(), "any type") {
		t.Fatalf("Lower error = %v, want SG1001 for any in Static mode", err)
	}
}

func TestLowerAllowsUnknownInClassFieldsAndArrays(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "class C { x: unknown; }\nconst arr: unknown[] = [1, \"a\"];\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}
	module, err := Lower(program)
	if err != nil {
		t.Fatalf("Lower error = %v, want success for unknown class field and array", err)
	}
	if err := module.Verify(); err != nil {
		t.Fatalf("Verify error = %v", err)
	}
}

func TestLowerAllowsUnknownInLocals(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "const value: unknown = 1;\nconsole.log(value);\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}
	module, err := Lower(program)
	if err != nil {
		t.Fatalf("Lower error = %v, want success for local unknown variable", err)
	}
	if err := module.Verify(); err != nil {
		t.Fatalf("Verify error = %v", err)
	}
}

func TestLowerConsoleIntrinsics(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "console.log(1); console.info(2); console.warn(3); console.error(4);\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}
	module, err := Lower(program)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"console.log", "console.info", "console.warn", "console.error"}
	var got []string
	for _, instruction := range module.Functions[0].Body {
		if instruction.Op == ir.OpPrint {
			got = append(got, instruction.Callee)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("console intrinsic callees = %v, want %v", got, want)
	}
}

func TestLowerAllowsUnionParameters(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "function printVal(value: string | number) { console.log(value); }\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}
	irModule, err := Lower(program)
	if err != nil {
		t.Fatalf("Lower error = %v, expected nil", err)
	}
	if len(irModule.Functions) == 0 {
		t.Fatalf("expected lowered functions")
	}
}

func TestLowerAllowsNullableType(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "const value: string | null = 'hello';\nconsole.log(value);\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}
	module, err := Lower(program)
	if err != nil {
		t.Fatalf("Lower error = %v, want successful lowering for nullable string", err)
	}
	if err := module.Verify(); err != nil {
		t.Fatalf("module.Verify() failed: %v", err)
	}
}

func TestLowerDoubleCastWarning(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "const n: number = 42;\nconst s: string = n as unknown as string;\nconsole.log(s);\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(program)
	if err != nil {
		t.Fatalf("Lower error = %v, want success with warning", err)
	}
	warns := GetWarnings()
	if len(warns) == 0 {
		t.Fatalf("want double cast warning, got none")
	}
	found := false
	for _, w := range warns {
		if w.Code == CodeUnsafeDoubleCast {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("want CodeUnsafeDoubleCast (SG3006) warning, got %v", warns)
	}
}

func TestLowerWarnRuntimeCastsOption(t *testing.T) {
	WarnRuntimeCasts = true
	defer func() { WarnRuntimeCasts = false }()

	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "const u: unknown = 42;\nconst n: number = u as number;\nconsole.log(n);\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(program)
	if err != nil {
		t.Fatalf("Lower error = %v, want success with warning", err)
	}
	warns := GetWarnings()
	if len(warns) == 0 {
		t.Fatalf("want runtime checked cast warning, got none")
	}
	found := false
	for _, w := range warns {
		if w.Code == CodeWarnCheckedCast {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("want CodeWarnCheckedCast (SG4005) warning, got %v", warns)
	}
}
