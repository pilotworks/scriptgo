package typescriptgo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckResolvesLocalModules(t *testing.T) {
	dir := t.TempDir()
	dependency := filepath.Join(dir, "answer.ts")
	entry := filepath.Join(dir, "main.ts")
	if err := os.WriteFile(dependency, []byte("export const answer: number = 42;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entry, []byte("import { answer } from './answer';\nconsole.log(answer);\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 2 {
		t.Fatalf("Check resolved %d files, want 2: %+v", len(result.Files), result.Files)
	}
	if result.Files[0].FileName != dependency || result.Files[1].FileName != entry {
		t.Fatalf("Check file order = [%s, %s], want dependency before entry", result.Files[0].FileName, result.Files[1].FileName)
	}
	if len(result.Files[1].Imports) != 1 {
		t.Fatalf("entry imports = %+v, want one resolved import", result.Files[1].Imports)
	}
	if result.Files[1].Imports[0].Specifier != "./answer" || result.Files[1].Imports[0].ResolvedFileName != dependency {
		t.Fatalf("entry import = %+v, want resolved answer module", result.Files[1].Imports[0])
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Check returned unexpected diagnostics: %+v", result.Diagnostics)
	}
}

func TestCheckResolvesBuiltinPathModule(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "import * as p from 'path';\nconsole.log(p.basename('a/b.txt'));\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Check returned unexpected diagnostics: %+v", result.Diagnostics)
	}
	if len(result.Files) != 2 || len(result.Files[1].Imports) != 1 {
		t.Fatalf("resolved files/imports = %+v, want builtin and entry files", result.Files)
	}
	if !result.Files[0].Builtin || result.Files[0].BuiltinName != "path" || result.Files[1].FileName != entry {
		t.Fatalf("resolved builtin files = %+v, want builtin before entry", result.Files)
	}
	imported := result.Files[1].Imports[0]
	if imported.Specifier != "path" || imported.LocalName != "p" || !imported.Builtin {
		t.Fatalf("path import = %+v, want builtin path reference with namespace alias", imported)
	}
	foundPath := false
	for _, m := range BuiltinModuleManifest() {
		if m.Name == "path" && m.Version != "" {
			foundPath = true
			break
		}
	}
	if !foundPath {
		t.Fatalf("builtin manifest = %+v, want versioned path module", BuiltinModuleManifest())
	}
}

func TestCheckReportsSemanticDiagnostics(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, []byte("const answer: number = 'not a number';\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) == 0 {
		t.Fatal("Check returned no semantic diagnostics")
	}
	found := false
	for _, diagnostic := range result.Diagnostics {
		if strings.Contains(diagnostic.Message, "number") && strings.Contains(diagnostic.Message, "string") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Check diagnostics do not describe the type error: %+v", result.Diagnostics)
	}
}

func TestCheckNormalizesArrayLiteralAndIndexing(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(entry, []byte("const values: number[] = [10, 20];\nconsole.log(values[1]);\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Check returned unexpected diagnostics: %+v", result.Diagnostics)
	}
	if len(result.Files) != 1 || len(result.Files[0].Syntax.Statements) != 2 {
		t.Fatalf("unexpected syntax files: %+v", result.Files)
	}
	variable := result.Files[0].Syntax.Statements[0]
	if variable.Type != "number[]" || variable.Expression == nil || variable.Expression.Kind != "array" {
		t.Fatalf("array declaration = %+v, want number[] array expression", variable)
	}
	if len(variable.Expression.Arguments) != 2 {
		t.Fatalf("array elements = %+v, want two elements", variable.Expression.Arguments)
	}
	index := result.Files[0].Syntax.Statements[1].Expression
	if index == nil || index.Kind != "call" || len(index.Arguments) != 1 || index.Arguments[0].Kind != "index" {
		t.Fatalf("indexing syntax = %+v, want console.log(values[1])", index)
	}
}

func TestCheckExposesStableFrontendContract(t *testing.T) {
	dir := t.TempDir()
	dependency := filepath.Join(dir, "answer.ts")
	entry := filepath.Join(dir, "main.ts")
	if err := os.WriteFile(dependency, []byte("export const answer: number = 42;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entry, []byte("import { answer } from './answer';\nconsole.log(answer);\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(entry)
	if err != nil {
		t.Fatal(err)
	}
	if result.Options.Target != "ES2022" || result.Options.Module != "ESNext" || !result.Options.Strict {
		t.Fatalf("compiler options = %+v, want normalized strict ES2022/ESNext options", result.Options)
	}
	if len(result.Files) != 2 || result.Files[0].Source == "" {
		t.Fatalf("source files = %+v, want source text for both reachable files", result.Files)
	}
	if len(result.Files[0].Symbols) != 1 {
		t.Fatalf("dependency symbols = %+v, want one symbol", result.Files[0].Symbols)
	}
	symbol := result.Files[0].Symbols[0]
	if symbol.Name != "answer" || symbol.Kind != "variable" || symbol.Type != "number" || !symbol.Exported {
		t.Fatalf("answer symbol = %+v, want exported number variable", symbol)
	}
	if symbol.Span.Length == 0 || symbol.Span.Start < 0 {
		t.Fatalf("answer symbol span = %+v, want source span", symbol.Span)
	}
	if len(result.Files[1].Imports) != 1 || result.Files[1].Imports[0].Span.Length == 0 {
		t.Fatalf("import references = %+v, want source-anchored module edge", result.Files[1].Imports)
	}
}

func TestCheckNormalizesUnaryAndLogicalExpressions(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "const x: number = 1;\nconst y: number = 2;\nconst a: boolean = !false && (x === 1 || y !== 3);\nconst b: number = -5;\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Check returned unexpected diagnostics: %+v", result.Diagnostics)
	}
	if len(result.Files) != 1 || len(result.Files[0].Syntax.Statements) != 4 {
		t.Fatalf("unexpected syntax files: %+v", result.Files)
	}
	first := result.Files[0].Syntax.Statements[2]
	if first.Expression == nil || first.Expression.Kind != "binary" || first.Expression.Operator != "&&" {
		t.Fatalf("first statement expression = %+v, want binary &&", first.Expression)
	}
	leftUnary := first.Expression.Left
	if leftUnary == nil || leftUnary.Kind != "unary" || leftUnary.Operator != "!" {
		t.Fatalf("left unary = %+v, want unary !", leftUnary)
	}
	second := result.Files[0].Syntax.Statements[3]
	if second.Expression == nil || second.Expression.Kind != "unary" || second.Expression.Operator != "-" {
		t.Fatalf("second statement expression = %+v, want unary -", second.Expression)
	}
}

func TestCheckSupportsGlobalProcess(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "const cwd: string = process.cwd();\nconst args: string[] = process.argv;\nprocess.exit(0);\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Check returned unexpected diagnostics for global process: %+v", result.Diagnostics)
	}
	if len(result.Files) != 1 {
		t.Fatalf("Check returned %d files, want 1 entry file", len(result.Files))
	}
}

func TestCheckResolvesNodePrefixedBuiltinModules(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "import * as fs from 'node:fs';\nimport * as path from 'node:path';\nimport * as crypto from 'node:crypto';\nimport * as process from 'node:process';\nconsole.log(path.join('a', 'b'));\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Check returned unexpected diagnostics for node: imports: %+v", result.Diagnostics)
	}
	if len(result.Files[len(result.Files)-1].Imports) != 4 {
		t.Fatalf("entry imports = %+v, want 4 node: imports", result.Files[len(result.Files)-1].Imports)
	}
	for _, imp := range result.Files[len(result.Files)-1].Imports {
		if !imp.Builtin {
			t.Fatalf("import %+v expected to be marked as Builtin", imp)
		}
		if !strings.HasSuffix(imp.ResolvedFileName, "index.ts") {
			t.Fatalf("import %+v resolved filename %s does not point to builtin index.ts", imp, imp.ResolvedFileName)
		}
	}
}

func TestCheckResolvesNodePrefixedNamedImports(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "import { join, dirname } from 'node:path';\nimport { randomUUID } from 'node:crypto';\nimport { exit } from 'node:process';\nconsole.log(join('a', 'b'));\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Check(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("Check returned unexpected diagnostics for named node: imports: %+v", result.Diagnostics)
	}
}



