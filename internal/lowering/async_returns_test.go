package lowering

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func TestAsyncReturnObjectUsesResolvedPromisePayloadShape(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := `
type StreamResult = { done: boolean; value: unknown };
type ByteStreamResult = { done: boolean; value: Uint8Array | undefined };

async function readChunk(queue: unknown[]): Promise<StreamResult> {
    if (queue.length > 0) {
        return { done: false, value: queue.shift() };
    }
    return { done: true, value: undefined };
}

async function readBytes(queue: unknown[]): Promise<ByteStreamResult> {
    const result = await readChunk(queue);
    return { done: result.done, value: undefined };
}
`
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}

	readChunk := program.Files[0].Syntax.Statements[2]
	originalInferredType := readChunk.Body[0].Then[0].Expression.InferredType
	module, err := Lower(program)
	if err != nil {
		t.Fatal(err)
	}
	if got := readChunk.Body[0].Then[0].Expression.InferredType; got != originalInferredType {
		t.Fatalf("Lower mutated frontend return expression type from %q to %q", originalInferredType, got)
	}

	function := findFunction(t, module, "readChunk")
	var objectShapes []string
	walkInstructions(function.Body, func(instruction ir.Instruction) {
		if instruction.Op == ir.OpObjectNew {
			objectShapes = append(objectShapes, instruction.Callee)
		}
	})
	wantShape := strings.TrimPrefix(string(toIRType("StreamResult")), "object:")
	if len(objectShapes) != 2 || objectShapes[0] != wantShape || objectShapes[1] != wantShape {
		t.Fatalf("readChunk object shapes = %v, want [%s %s]", objectShapes, wantShape, wantShape)
	}
}

func findFunction(t *testing.T, module ir.Module, name string) ir.Function {
	t.Helper()
	for _, function := range module.Functions {
		if function.Name == name {
			return function
		}
	}
	t.Fatalf("function %q not found", name)
	return ir.Function{}
}

func walkInstructions(instructions []ir.Instruction, visit func(ir.Instruction)) {
	for _, instruction := range instructions {
		visit(instruction)
		walkInstructions(instruction.Then, visit)
		walkInstructions(instruction.Else, visit)
		walkInstructions(instruction.Body, visit)
		walkInstructions(instruction.Cond, visit)
		walkInstructions(instruction.Step, visit)
	}
}
