package pkgmgr

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
	"strings"
	"testing"
)

func TestInstallResolvesVerifiesAndLinksPackage(t *testing.T) {
	archive := testPackageArchive(t, "index.js", "module.exports = 42;\n")
	digest := sha512.Sum512(archive)
	integrity := "sha512-" + base64.StdEncoding.EncodeToString(digest[:])
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/demo" {
			tarball := "http://" + request.Host + "/demo/-/demo.tgz"
			_ = json.NewEncoder(w).Encode(map[string]any{"versions": map[string]any{
				"1.2.3": map[string]any{"name": "demo", "version": "1.2.3", "dist": map[string]string{"tarball": tarball, "integrity": integrity}},
			}})
			return
		}
		if request.URL.Path == "/demo/-/demo.tgz" {
			_, _ = w.Write(archive)
			return
		}
		http.NotFound(w, request)
	}))
	defer server.Close()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"dependencies":{"demo":"^1.0.0"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(InstallOptions{ProjectRoot: root, Registry: Registry{BaseURL: server.URL}}); err != nil {
		t.Fatal(err)
	}
	linked, err := os.ReadFile(filepath.Join(root, "node_modules", "demo", "index.js"))
	if err != nil {
		t.Fatal(err)
	}
	if string(linked) != "module.exports = 42;\n" {
		t.Fatalf("linked package = %q", linked)
	}
	if _, err := Install(InstallOptions{ProjectRoot: root, Frozen: true, Registry: Registry{BaseURL: server.URL}}); err != nil {
		t.Fatalf("frozen install: %v", err)
	}
	if _, err := Install(InstallOptions{ProjectRoot: root, Offline: true}); err != nil {
		t.Fatalf("offline install: %v", err)
	}
}

func TestInstallLinksConflictingVersionsUnderTheirParents(t *testing.T) {
	archives := map[string][]byte{
		"a":  testPackageArchive(t, "index.js", "require('c');\n"),
		"b":  testPackageArchive(t, "index.js", "require('c');\n"),
		"c1": testPackageArchive(t, "version.js", "1\n"),
		"c2": testPackageArchive(t, "version.js", "2\n"),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		name := strings.TrimPrefix(request.URL.Path, "/")
		if strings.HasPrefix(name, "tar/") {
			name = strings.TrimPrefix(name, "tar/")
		}
		if strings.HasPrefix(request.URL.Path, "/tar/") {
			archive, ok := archives[name]
			if !ok {
				http.NotFound(w, request)
				return
			}
			_, _ = w.Write(archive)
			return
		}
		versions := map[string]any{}
		switch name {
		case "a":
			versions["1.0.0"] = registryVersion(archives, "a", "a", "^1.0.0", request.Host)
		case "b":
			versions["1.0.0"] = registryVersion(archives, "b", "b", "^2.0.0", request.Host)
		case "c":
			versions["1.0.0"] = registryVersion(archives, "c1", "c1", "", request.Host)
			versions["2.0.0"] = registryVersion(archives, "c2", "c2", "", request.Host)
		default:
			http.NotFound(w, request)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"versions": versions})
	}))
	defer server.Close()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"dependencies":{"a":"1.0.0","b":"1.0.0"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(InstallOptions{ProjectRoot: root, Registry: Registry{BaseURL: server.URL}}); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{
		"node_modules/a/node_modules/c/version.js": "1\n",
		"node_modules/b/node_modules/c/version.js": "2\n",
	} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != want {
			t.Fatalf("%s = %q, want %q", path, data, want)
		}
	}
}

func registryVersion(archives map[string][]byte, name, archiveName, dependency, host string) map[string]any {
	archive := archives[archiveName]
	digest := sha512.Sum512(archive)
	result := map[string]any{"name": name, "version": "1.0.0", "dist": map[string]string{
		"tarball":   "http://" + host + "/tar/" + archiveName,
		"integrity": "sha512-" + base64.StdEncoding.EncodeToString(digest[:]),
	}}
	if dependency != "" {
		result["dependencies"] = map[string]string{"c": dependency}
	}
	return result
}

func testPackageArchive(t *testing.T, name, content string) []byte {
	t.Helper()
	var data bytes.Buffer
	gzipWriter := gzip.NewWriter(&data)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := tarWriter.WriteHeader(&tar.Header{Name: "package/" + name, Mode: 0o644, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}
