package lowering

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
)

func lowerSource(t *testing.T, source string) ir.Module {
	t.Helper()
	entry := filepath.Join(t.TempDir(), "main.ts")
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
	return module
}

func functionNames(module ir.Module) []string {
	names := make([]string, 0, len(module.Functions))
	for _, function := range module.Functions {
		names = append(names, function.Name)
	}
	return names
}

// TestLowerInheritedMembersKeepDeclarationOrder pins the order of members a
// subclass inherits: base declarations first, overrides in their base slot,
// then members the subclass adds.
func TestLowerInheritedMembersKeepDeclarationOrder(t *testing.T) {
	source := `class Animal {
  constructor(public name: string) {}
  get label(): string { return this.name; }
  speak(): string { return "..."; }
  describe(): string { return this.label + " says " + this.speak(); }
}
class Dog extends Animal {
  speak(): string { return "woof"; }
  fetch(): string { return "ball"; }
}
const dog = new Dog("Rex");
console.log(dog.describe(), dog.fetch());
`
	var dogMembers []string
	for _, name := range functionNames(lowerSource(t, source)) {
		if strings.HasPrefix(name, "Dog_") && !strings.HasSuffix(name, "_dispatch") {
			dogMembers = append(dogMembers, name)
		}
	}
	want := "Dog_constructor,Dog_get_label,Dog_speak_impl,Dog_describe_impl,Dog_fetch_impl"
	if got := strings.Join(dogMembers, ","); got != want {
		t.Fatalf("Dog members lowered as %s, want %s", got, want)
	}
}

func TestLowerIsDeterministic(t *testing.T) {
	source := `interface Shape { kind: string; size: number }
class Base { area(): number { return 0; } get name(): string { return "base"; } }
class Square extends Base { constructor(public side: number) { super(); } area(): number { return this.side * this.side; } }
class Circle extends Base { constructor(public r: number) { super(); } area(): number { return 3 * this.r * this.r; } }
const shapes: Base[] = [new Square(2), new Circle(1)];
let total = 0;
for (const s of shapes) { total += s.area(); }
const literal: Shape = { kind: "box", size: total };
console.log(literal.kind, literal.size, shapes[0].name);
`
	first := formatModule(lowerSource(t, source))
	for run := 0; run < 5; run++ {
		if got := formatModule(lowerSource(t, source)); got != first {
			t.Fatalf("lowered module differs between runs (run %d)", run)
		}
	}
}

func hasCallee(module ir.Module, callee string) bool {
	found := false
	var walk func([]ir.Instruction)
	walk = func(list []ir.Instruction) {
		for _, instruction := range list {
			if instruction.Callee == callee || instruction.Op == callee {
				found = true
			}
			walk(instruction.Body)
			walk(instruction.Then)
			walk(instruction.Else)
			walk(instruction.Cond)
			walk(instruction.Step)
			walk(instruction.Catch)
			walk(instruction.Finally)
		}
	}
	for _, function := range module.Functions {
		walk(function.Body)
	}
	return found
}

// TestLowerSplitStagesSmoke lowers one construct per extracted lowering stage
// and checks the stage produced its characteristic instruction.
func TestLowerSplitStagesSmoke(t *testing.T) {
	cases := []struct {
		name, source, want string
	}{
		{"array literal", "const xs: number[] = [1, 2, 3];\nconsole.log(xs.length);\n", ir.OpArray},
		{"template literal", "const n = 2;\nconsole.log(`n=${n}`);\n", "__string.fromNumber"},
		{"index", "const xs: number[] = [4, 5];\nconsole.log(xs[1]);\n", ir.OpIndex},
		{"typeof", "const v: unknown = 1;\nconsole.log(typeof v);\n", ir.OpTypeOf},
		{"conditional", "const a = 3;\nconsole.log(a > 2 ? \"big\" : \"small\");\n", ir.OpIf},
		{"logical and", "const a = true; const b = false;\nconsole.log(a && b);\n", ir.OpIf},
		{"nullish", "const v: number | null = null;\nconsole.log(v ?? 7);\n", ir.OpIf},
		{"assignment", "let total = 1;\ntotal += 2;\nconsole.log(total);\n", ir.OpAssign},
		{"field set", "class P { x = 0; }\nconst p = new P();\np.x = 3;\nconsole.log(p.x);\n", ir.OpFieldSet},
		{"index set", "const xs: number[] = [0];\nxs[0] = 9;\nconsole.log(xs[0]);\n", ir.OpIndexSet},
		{"throw", "function fail(): void { throw new Error(\"boom\"); }\ntry { fail(); } catch (e) { console.log(\"caught\"); }\n", ir.OpThrow},
		{"super call", "class A { constructor(public v: number) {} }\nclass B extends A { constructor() { super(1); } }\nconsole.log(new B().v);\n", "A_constructor"},
		{"closure call", "const add = (a: number, b: number): number => a + b;\nconsole.log(add(1, 2));\n", ir.OpClosureCall},
		{"string length", "const s = \"abc\";\nconsole.log(s.length);\n", "__string.length"},
		{"new map", "const m = new Map<string, number>();\nm.set(\"a\", 1);\nconsole.log(m.size);\n", "__map.new"},
		{"new error", "const e = new Error(\"bad\");\nconsole.log(e.message);\n", ir.OpObjectNew},
		{"math intrinsic", "console.log(Math.floor(2.5));\n", "__Math.floor"},
		{"number intrinsic", "console.log(Number.isInteger(4));\n", "__number.isInteger"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			module := lowerSource(t, tc.source)
			if !hasCallee(module, tc.want) {
				t.Fatalf("expected %q in lowered module:\n%s", tc.want, formatModule(module))
			}
		})
	}
}
