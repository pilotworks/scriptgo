package lowering

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	source := "if (true) { console.log(42); }\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Lower(program)
	if err == nil || !strings.Contains(err.Error(), "native subset") || !strings.Contains(err.Error(), "IfStatement") || !strings.Contains(err.Error(), "at offset 0") {
		t.Fatalf("Lower error = %v, want actionable native subset diagnostic", err)
	}
}
