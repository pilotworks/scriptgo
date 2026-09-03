package lowering

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pilotworks/scriptgo/internal/frontend"
)

func TestLowerImportedTLSStringConstantProperty(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "import { DEFAULT_CIPHERS } from \"node:tls\"; console.log(DEFAULT_CIPHERS.length);\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Lower(program); err != nil {
		t.Fatalf("lower imported string constant property: %v", err)
	}
}

func TestLowerTLSSocketAcceptsNetSocket(t *testing.T) {
	entry := filepath.Join(t.TempDir(), "main.ts")
	source := "import { Socket } from \"node:net\"; import { TLSSocket } from \"node:tls\"; const socket = new Socket(); const tls = new TLSSocket(socket); console.log(tls.encrypted);\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := frontend.NewProgram(entry, source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Lower(program); err != nil {
		t.Fatalf("lower TLSSocket from net.Socket: %v", err)
	}
}
