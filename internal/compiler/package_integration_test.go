package compiler

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pilotworks/scriptgo/internal/pkgmgr"
)

// TestInstalledDynamicPackageRunsOffline proves the product path from registry
// metadata through installation, frontend resolution, and native execution.
func TestInstalledDynamicPackageRunsOffline(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.ts"), []byte(`import { run } from "installed-esm";
console.log(run());
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{
  "dependencies": {"installed-esm": "^1.0.0"}
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	archives := map[string][]byte{
		"installed-esm-1.0.0.tgz": packageArchive(t, map[string]string{
			"package/package.json": `{"name":"installed-esm","version":"1.0.0","type":"module","dependencies":{"installed-cjs":"^1.0.0"}}`,
			"package/index.js":     `import add from "installed-cjs"; export function run() { return add(20, 22); }`,
		}),
		"installed-cjs-1.0.0.tgz": packageArchive(t, map[string]string{
			"package/package.json": `{"name":"installed-cjs","version":"1.0.0"}`,
			"package/index.js":     `module.exports = function add(left, right) { return left + right; };`,
		}),
	}
	var registryURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var name string
		switch r.URL.Path {
		case "/installed-esm":
			name = "installed-esm"
		case "/installed-cjs":
			name = "installed-cjs"
		default:
			archiveName := filepath.Base(r.URL.Path)
			data, ok := archives[archiveName]
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
			return
		}
		archiveName := name + "-1.0.0.tgz"
		payload := map[string]any{
			"name": name,
			"versions": map[string]any{
				"1.0.0": map[string]any{
					"name": name, "version": "1.0.0",
					"dependencies": map[string]string{},
					"dist": map[string]string{
						"tarball":   registryURL + "/" + archiveName,
						"integrity": integrityFor(archives[archiveName]),
					},
				},
			},
		}
		if name == "installed-esm" {
			payload["versions"].(map[string]any)["1.0.0"].(map[string]any)["dependencies"] = map[string]string{"installed-cjs": "^1.0.0"}
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()
	registryURL = server.URL

	install := func(offline, frozen bool) pkgmgr.Lockfile {
		t.Helper()
		lock, err := pkgmgr.Install(pkgmgr.InstallOptions{
			ProjectRoot: root,
			Registry:    pkgmgr.Registry{BaseURL: server.URL},
			Offline:     offline,
			Frozen:      frozen,
		})
		if err != nil {
			t.Fatal(err)
		}
		return lock
	}
	lock := install(false, false)
	if len(lock.Packages) != 2 {
		t.Fatalf("installed package graph contains %d packages, want 2", len(lock.Packages))
	}
	assertInstalledOutput(t, root, "first install")
	server.Close()
	install(true, true)
	assertInstalledOutput(t, root, "offline frozen install")
}

func assertInstalledOutput(t *testing.T, root, phase string) {
	t.Helper()
	output, err := RunWithOptions(filepath.Join(root, "main.ts"), BuildOptions{Dynamic: true, OptLevel: "0"})
	if err != nil {
		t.Fatalf("%s: run failed: %v", phase, err)
	}
	if output != "42\n" {
		t.Fatalf("%s: output = %q, want %q", phase, output, "42\n")
	}
}

func packageArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	tarWriter := tar.NewWriter(writer)
	for name, contents := range files {
		data := []byte(contents)
		if err := tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(data))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return compressed.Bytes()
}

func integrityFor(data []byte) string {
	// The installer accepts SHA-512 SRI strings; keep this fixture independent
	// from the package store implementation.
	digest := sha512.Sum512(data)
	return "sha512-" + base64.StdEncoding.EncodeToString(digest[:])
}
