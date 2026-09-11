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
		name = strings.TrimPrefix(name, "tar/")
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

func TestInstallCreatesBinAndRejectsFrozenManifestDrift(t *testing.T) {
	archive := testPackageArchive(t, "bin/demo.js", "#!/bin/sh\necho demo\n")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer secret-token" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}
		if request.URL.Path == "/demo" {
			_ = json.NewEncoder(w).Encode(map[string]any{"versions": map[string]any{
				"1.0.0": map[string]any{"name": "demo", "version": "1.0.0", "bin": map[string]string{"demo": "./bin/demo.js"}, "dist": map[string]string{
					"tarball": "http://" + request.Host + "/demo.tgz", "integrity": integrityFor(archive),
				}},
			}})
			return
		}
		if request.URL.Path == "/demo.tgz" {
			_, _ = w.Write(archive)
			return
		}
		http.NotFound(w, request)
	}))
	defer server.Close()
	root := t.TempDir()
	manifestPath := filepath.Join(root, "package.json")
	if err := os.WriteFile(manifestPath, []byte(`{"dependencies":{"demo":"^1.0.0"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(InstallOptions{ProjectRoot: root, Registry: Registry{BaseURL: server.URL, Token: "secret-token"}}); err != nil {
		t.Fatal(err)
	}
	bin, err := os.Readlink(filepath.Join(root, "node_modules", ".bin", "demo"))
	if err != nil {
		t.Fatal(err)
	}
	if bin != filepath.Join("..", "demo", "bin", "demo.js") {
		t.Fatalf("bin link = %q", bin)
	}
	if err := os.WriteFile(manifestPath, []byte(`{"dependencies":{"demo":"^2.0.0"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(InstallOptions{ProjectRoot: root, Frozen: true}); err == nil || !strings.Contains(err.Error(), "manifest dependencies differ") {
		t.Fatalf("frozen drift error = %v", err)
	}
}

func TestInstallSkipsUnavailableOptionalDependency(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"optionalDependencies":{"missing":"^1.0.0"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) { http.NotFound(w, request) }))
	defer server.Close()
	lock, err := Install(InstallOptions{ProjectRoot: root, Registry: Registry{BaseURL: server.URL}})
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Packages) != 0 {
		t.Fatalf("optional failure left packages in lockfile: %+v", lock.Packages)
	}
}

func TestInstallValidatesPeerDependenciesAtAncestor(t *testing.T) {
	archives := map[string][]byte{
		"plugin": testPackageArchive(t, "index.js", "module.exports = 1;\n"),
		"react":  testPackageArchive(t, "index.js", "module.exports = 1;\n"),
	}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		name := strings.TrimPrefix(request.URL.Path, "/")
		if strings.HasSuffix(name, ".tgz") {
			key := strings.TrimSuffix(name, ".tgz")
			_, _ = w.Write(archives[key])
			return
		}
		manifest := map[string]any{"name": name, "version": "1.0.0", "dist": map[string]string{
			"tarball": server.URL + "/" + name + ".tgz", "integrity": integrityFor(archives[name]),
		}}
		if name == "plugin" {
			manifest["peerDependencies"] = map[string]string{"react": "^1.0.0"}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"versions": map[string]any{"1.0.0": manifest}})
	}))
	defer server.Close()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"dependencies":{"plugin":"1.0.0","react":"1.0.0"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(InstallOptions{ProjectRoot: root, Registry: Registry{BaseURL: server.URL}}); err != nil {
		t.Fatalf("valid peer install: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"dependencies":{"plugin":"1.0.0"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(InstallOptions{ProjectRoot: root, Registry: Registry{BaseURL: server.URL}}); err == nil || !strings.Contains(err.Error(), "requires peer") {
		t.Fatalf("missing peer error = %v", err)
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

func integrityFor(data []byte) string {
	digest := sha512.Sum512(data)
	return "sha512-" + base64.StdEncoding.EncodeToString(digest[:])
}

func testPackageArchive(t *testing.T, name, content string) []byte {
	t.Helper()
	var data bytes.Buffer
	gzipWriter := gzip.NewWriter(&data)
	tarWriter := tar.NewWriter(gzipWriter)
	mode := int64(0o644)
	if strings.Contains(name, "/bin/") {
		mode = 0o755
	}
	if err := tarWriter.WriteHeader(&tar.Header{Name: "package/" + name, Mode: mode, Size: int64(len(content))}); err != nil {
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
