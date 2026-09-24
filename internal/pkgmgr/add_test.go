package pkgmgr

import (
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

func TestParsePackageSpec(t *testing.T) {
	tests := []struct {
		raw      string
		wantName string
		wantSpec string
		wantErr  bool
	}{
		{raw: "lodash", wantName: "lodash", wantSpec: ""},
		{raw: "lodash@^4.17.21", wantName: "lodash", wantSpec: "^4.17.21"},
		{raw: "lodash@4.17.21", wantName: "lodash", wantSpec: "4.17.21"},
		{raw: "lodash@~4.17.0", wantName: "lodash", wantSpec: "~4.17.0"},
		{raw: "@types/node", wantName: "@types/node", wantSpec: ""},
		{raw: "@types/node@20.0.0", wantName: "@types/node", wantSpec: "20.0.0"},
		{raw: "@types/node@^20.0.0", wantName: "@types/node", wantSpec: "^20.0.0"},
		{raw: "shared@workspace:*", wantName: "shared", wantSpec: "workspace:*"},
		{raw: "shared@workspace:^", wantName: "shared", wantSpec: "workspace:^"},
		{raw: "", wantErr: true},
		{raw: "   ", wantErr: true},
		{raw: "foo/bar", wantErr: true},
		{raw: "@scope", wantErr: true},
		{raw: "foo\\bar", wantErr: true},
	}

	for _, tt := range tests {
		name, spec, err := ParsePackageSpec(tt.raw)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParsePackageSpec(%q) expected error, got nil", tt.raw)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParsePackageSpec(%q) unexpected error: %v", tt.raw, err)
			continue
		}
		if name != tt.wantName || spec != tt.wantSpec {
			t.Errorf("ParsePackageSpec(%q) = (%q, %q), want (%q, %q)", tt.raw, name, spec, tt.wantName, tt.wantSpec)
		}
	}
}

func TestAdd_RegistryPackage(t *testing.T) {
	archive := testPackageArchive(t, "index.js", "module.exports = 'from-registry';\n")
	digest := sha512.Sum512(archive)
	integrity := "sha512-" + base64.StdEncoding.EncodeToString(digest[:])
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/demo" {
			tarball := "http://" + request.Host + "/demo/-/demo.tgz"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"dist-tags": map[string]string{"latest": "1.2.3"},
				"versions": map[string]any{
					"1.2.3": map[string]any{"name": "demo", "version": "1.2.3", "dist": map[string]string{"tarball": tarball, "integrity": integrity}},
				},
			})
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
	initialManifest := `{
  "name": "test-app",
  "version": "1.0.0",
  "scripts": {
    "start": "scriptgo run index.ts"
  }
}`
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(initialManifest), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Add(AddOptions{
		ProjectRoot:  root,
		Dependencies: []string{"demo"},
		Registry:     Registry{BaseURL: server.URL},
	})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if result.Added["demo"] != "^1.2.3" {
		t.Fatalf("expected demo@^1.2.3, got: %s", result.Added["demo"])
	}

	manifestBytes, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest PackageManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("invalid package.json JSON: %v", err)
	}
	if manifest.Dependencies["demo"] != "^1.2.3" {
		t.Fatalf("package.json dependencies[demo] = %q, want ^1.2.3", manifest.Dependencies["demo"])
	}

	linked, err := os.ReadFile(filepath.Join(root, "node_modules", "demo", "index.js"))
	if err != nil {
		t.Fatalf("node_modules/demo/index.js missing: %v", err)
	}
	if string(linked) != "module.exports = 'from-registry';\n" {
		t.Fatalf("unexpected node_modules content: %s", string(linked))
	}

	lockBytes, err := os.ReadFile(filepath.Join(root, "scriptgo-lock.json"))
	if err != nil {
		t.Fatalf("scriptgo-lock.json missing: %v", err)
	}
	if !strings.Contains(string(lockBytes), "demo@1.2.3") {
		t.Fatalf("lockfile missing demo@1.2.3: %s", string(lockBytes))
	}
}

func TestAdd_ExactAndDevDependency(t *testing.T) {
	archive := testPackageArchive(t, "index.js", "module.exports = 'exact';\n")
	digest := sha512.Sum512(archive)
	integrity := "sha512-" + base64.StdEncoding.EncodeToString(digest[:])
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/tool" {
			tarball := "http://" + request.Host + "/tool/-/tool.tgz"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"dist-tags": map[string]string{"latest": "2.0.0"},
				"versions": map[string]any{
					"2.0.0": map[string]any{"name": "tool", "version": "2.0.0", "dist": map[string]string{"tarball": tarball, "integrity": integrity}},
				},
			})
			return
		}
		if request.URL.Path == "/tool/-/tool.tgz" {
			_, _ = w.Write(archive)
			return
		}
		http.NotFound(w, request)
	}))
	defer server.Close()

	root := t.TempDir()
	initialManifest := `{
  "name": "test-dev-app",
  "version": "1.0.0",
  "dependencies": {
    "tool": "^1.0.0"
  }
}`
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(initialManifest), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Add(AddOptions{
		ProjectRoot:  root,
		Dependencies: []string{"tool"},
		DepType:      DepTypeDev,
		Exact:        true,
		Registry:     Registry{BaseURL: server.URL},
	})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if result.Added["tool"] != "2.0.0" {
		t.Fatalf("expected exact version 2.0.0, got: %s", result.Added["tool"])
	}

	manifestBytes, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest PackageManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("invalid package.json JSON: %v", err)
	}
	if _, inDeps := manifest.Dependencies["tool"]; inDeps {
		t.Fatalf("tool should have been removed from dependencies when added to devDependencies")
	}
	if manifest.DevDependencies["tool"] != "2.0.0" {
		t.Fatalf("devDependencies[tool] = %q, want 2.0.0", manifest.DevDependencies["tool"])
	}
}

func TestAdd_WorkspacePackage(t *testing.T) {
	root := t.TempDir()
	packagesDir := filepath.Join(root, "packages")
	workspaceRoot := filepath.Join(packagesDir, "core")
	if err := os.MkdirAll(workspaceRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "package.json"), []byte(`{"name":"@local/core","version":"3.4.5","main":"index.js"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "index.js"), []byte("module.exports = 'core';\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	rootManifest := `{
  "name": "mono-root",
  "workspaces": ["packages/*"]
}`
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(rootManifest), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Add(AddOptions{
		ProjectRoot:  root,
		Dependencies: []string{"@local/core"},
	})
	if err != nil {
		t.Fatalf("Add workspace package failed: %v", err)
	}

	if result.Added["@local/core"] != "workspace:*" {
		t.Fatalf("expected workspace:*, got: %s", result.Added["@local/core"])
	}

	linked, err := os.ReadFile(filepath.Join(root, "node_modules", "@local", "core", "index.js"))
	if err != nil {
		t.Fatalf("workspace package node_modules entry missing: %v", err)
	}
	if string(linked) != "module.exports = 'core';\n" {
		t.Fatalf("unexpected content: %s", string(linked))
	}
}

func TestAdd_RollbackOnFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/failing" {
			tarball := "http://" + request.Host + "/failing/-/failing.tgz"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"dist-tags": map[string]string{"latest": "1.0.0"},
				"versions": map[string]any{
					"1.0.0": map[string]any{"name": "failing", "version": "1.0.0", "dist": map[string]string{"tarball": tarball, "integrity": "sha512-invalid"}},
				},
			})
			return
		}
		// Return 404 for tarball
		http.NotFound(w, request)
	}))
	defer server.Close()

	root := t.TempDir()
	originalManifest := `{
  "name": "rollback-test",
  "version": "1.0.0"
}`
	manifestPath := filepath.Join(root, "package.json")
	if err := os.WriteFile(manifestPath, []byte(originalManifest), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Add(AddOptions{
		ProjectRoot:  root,
		Dependencies: []string{"failing"},
		Registry:     Registry{BaseURL: server.URL},
	})
	if err == nil {
		t.Fatal("expected Add to fail due to missing tarball")
	}

	contentAfterFailure, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(contentAfterFailure) != originalManifest {
		t.Fatalf("package.json was not rolled back!\nGot:\n%s\nWant:\n%s", string(contentAfterFailure), originalManifest)
	}
}
