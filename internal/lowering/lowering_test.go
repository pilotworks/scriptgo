package lowering

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

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
	if err == nil || !strings.Contains(err.Error(), "native subset") || !strings.Contains(err.Error(), "IfStatement") {
		t.Fatalf("Lower error = %v, want actionable native subset diagnostic", err)
	}
}
