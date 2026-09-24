package main

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
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func createMockTarball(t *testing.T, filename, content string) ([]byte, string) {
	t.Helper()
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)
	hdr := &tar.Header{
		Name: "package/" + filename,
		Mode: 0o644,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatal(err)
	}
	archive := buf.Bytes()
	digest := sha512.Sum512(archive)
	integrity := "sha512-" + base64.StdEncoding.EncodeToString(digest[:])
	return archive, integrity
}

func TestCLI_Add(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "scriptgo")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build scriptgo: %v\noutput: %s", err, string(out))
	}

	pkgArchive, integrity := createMockTarball(t, "index.js", "module.exports = 'from-add-cli';\n")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/my-pkg" {
			tarball := "http://" + request.Host + "/my-pkg/-/my-pkg.tgz"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"dist-tags": map[string]string{"latest": "1.5.0"},
				"versions": map[string]any{
					"1.5.0": map[string]any{
						"name":    "my-pkg",
						"version": "1.5.0",
						"dist": map[string]string{
							"tarball":   tarball,
							"integrity": integrity,
						},
					},
				},
			})
			return
		}
		if request.URL.Path == "/my-pkg/-/my-pkg.tgz" {
			_, _ = w.Write(pkgArchive)
			return
		}
		http.NotFound(w, request)
	}))
	defer server.Close()

	t.Run("no packages specified returns error", func(t *testing.T) {
		cmd := exec.Command(binPath, "add")
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected error for empty add, got: %s", string(out))
		}
		if !strings.Contains(string(out), "at least one package must be specified") {
			t.Fatalf("unexpected output: %s", string(out))
		}
	})

	t.Run("mutually exclusive flags return error", func(t *testing.T) {
		cmd := exec.Command(binPath, "add", "-D", "-O", "my-pkg")
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected error for mutually exclusive flags, got: %s", string(out))
		}
		if !strings.Contains(string(out), "mutually exclusive") {
			t.Fatalf("unexpected output: %s", string(out))
		}
	})

	t.Run("add package successfully", func(t *testing.T) {
		projectDir := filepath.Join(tmpDir, "project1")
		if err := os.MkdirAll(projectDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(`{"name":"test1"}`), 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(binPath, "add", "--project", projectDir, "--registry", server.URL, "my-pkg")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("add failed: %v\noutput: %s", err, string(out))
		}
		if !strings.Contains(string(out), "added my-pkg@^1.5.0") {
			t.Fatalf("unexpected stdout: %s", string(out))
		}

		manifestBytes, err := os.ReadFile(filepath.Join(projectDir, "package.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(manifestBytes), `"my-pkg": "^1.5.0"`) {
			t.Fatalf("package.json does not contain dependency: %s", string(manifestBytes))
		}

		installedFile, err := os.ReadFile(filepath.Join(projectDir, "node_modules", "my-pkg", "index.js"))
		if err != nil {
			t.Fatalf("installed file missing: %v", err)
		}
		if string(installedFile) != "module.exports = 'from-add-cli';\n" {
			t.Fatalf("unexpected installed content: %s", string(installedFile))
		}
	})

	t.Run("add dev dependency with -D flag", func(t *testing.T) {
		projectDir := filepath.Join(tmpDir, "project2")
		if err := os.MkdirAll(projectDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(`{"name":"test2"}`), 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(binPath, "add", "--project", projectDir, "--registry", server.URL, "-D", "my-pkg")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("add -D failed: %v\noutput: %s", err, string(out))
		}

		manifestBytes, err := os.ReadFile(filepath.Join(projectDir, "package.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(manifestBytes), `"devDependencies"`) || !strings.Contains(string(manifestBytes), `"my-pkg": "^1.5.0"`) {
			t.Fatalf("package.json does not contain devDependency: %s", string(manifestBytes))
		}
	})

	t.Run("add exact version with -E flag", func(t *testing.T) {
		projectDir := filepath.Join(tmpDir, "project3")
		if err := os.MkdirAll(projectDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(`{"name":"test3"}`), 0o644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(binPath, "add", "--project", projectDir, "--registry", server.URL, "-E", "my-pkg@1.5.0")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("add -E failed: %v\noutput: %s", err, string(out))
		}
		if !strings.Contains(string(out), "added my-pkg@1.5.0") {
			t.Fatalf("unexpected stdout: %s", string(out))
		}

		manifestBytes, err := os.ReadFile(filepath.Join(projectDir, "package.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(manifestBytes), `"my-pkg": "1.5.0"`) {
			t.Fatalf("package.json does not contain exact version: %s", string(manifestBytes))
		}
	})
}
