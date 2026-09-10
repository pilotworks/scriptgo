# Package Manager Architecture

This document specifies the technical design, protocol contracts, and storage
engine for a high-performance, native JavaScript/TypeScript package manager
compatible with the npm registry, designed along the principles of **Bun** and
**pnpm**.

## Design Principles

- **Zero-Copy Installation**: Maximize file-system level sharing via Copy-on-Write
  (`clonefile` on macOS APFS, `FICLONE` on Linux) or Content-Addressable
  Hardlinks, eliminating redundant I/O copy operations.
- **Deterministic Resolution**: Produce identical dependency graphs and lockfile
  manifests for any given dependency specification across platforms.
- **Decoupled Engine Pipeline**: Keep network, resolution, caching, and
  filesystem operations strictly layered.
- **Phantom Dependency Prevention**: Isolate dependencies to avoid accidental
  imports of unlisted transitive packages.
- **Direct Streaming Unpack**: Extract compressed archives directly from network
  streams into content-addressable storage without intermediary disk temporary
  files.

## Pipeline Architecture

The package manager operates as a 6-stage execution pipeline:

```text
package.json / CLI
      │
      ▼
1. Metadata Resolution ────► npm Registry API (abbreviated JSON)
      │
      ▼
2. Dependency Graph ───────► SemVer Solver & Cycle Detection
      │
      ▼
3. CAS Cache Check ────────► ~/.scriptgo/store/ (Content-Addressable Storage)
      │ (on miss)
      ▼
4. Network Fetcher ────────► Concurrent HTTP/2 Pool + Streaming SHA-512
      │
      ▼
5. Storage & Link Engine ──► OS-level clonefile / FICLONE / hardlinks
      │
      ▼
6. Bins & Lifecycle ───────► node_modules/.bin links & postinstall sandbox
      │
      ▼
node_modules/ + scriptgo-lock.json
```

## Registry Protocol Specification

The package manager communicates with standard npm registries
(`https://registry.npmjs.org`) using public HTTP REST endpoints.

### Endpoints and Headers

| Endpoint | Method | Purpose | Key Payload Fields |
| :--- | :--- | :--- | :--- |
| `/<pkg>` | `GET` | Fetch all versions & dist-tags | `dist-tags`, `versions` |
| `/<pkg>/<version>` | `GET` | Fetch manifest of specific version | `dependencies`, `dist.tarball`, `dist.integrity` |
| `/<pkg>/-/<tarball>.tgz` | `GET` | Download package tarball | `.tar.gz` archive |

### Abbreviated Metadata Negotiation

To minimize network latency and bandwidth overhead, requests send the npm
install header:

```http
GET /lodash HTTP/1.1
Host: registry.npmjs.org
Accept: application/vnd.npm.install-v1+json; q=1.0, application/json; q=0.8, */*
```

This strips package descriptions, readmes, versions' changelogs, and maintainer
lists, reducing network payload sizes by ~70%.

### Subresource Integrity (SRI) Verification

Every package version provides an `integrity` string formatted as
`<algorithm>-<base64-hash>` (typically `sha512-...`):

1. Stream the tarball bytes into an incremental hash digest simultaneously while
   reading.
2. Verify that `base64(digest) == expected_hash`.
3. Reject corrupted or tampered payloads immediately before committing artifacts
   to the global store.

## Dependency Resolution & Graph Solver

Resolution maps SemVer ranges (`^1.2.0`, `~2.1.0`, `>=1.0.0 <3.0.0`) into a
concrete, conflict-free dependency tree.

### Graph Constraints

- **Semantic Version Ranges**:
  - `^x.y.z`: `>=x.y.z <(x+1).0.0` (for `x > 0`).
  - `~x.y.z`: `>=x.y.z <x.(y+1).0`.
- **Diamond Dependencies**:
  When package `A` depends on `C@^1.0.0` and package `B` depends on `C@^1.2.0`,
  the resolver selects the highest common compatible version `C@1.2.x` to
  prevent duplicate installation.
- **Version Incompatibilities**:
  When requirements conflict (e.g. `C@^1.0.0` vs `C@^2.0.0`), the store
  retains both versions concurrently and scopes their links per importing package.

### Lockfile Schema (`scriptgo-lock.json`)

```json
{
  "lockfileVersion": 1,
  "packages": {
    "chalk@4.1.2": {
      "resolved": "https://registry.npmjs.org/chalk/-/chalk-4.1.2.tgz",
      "integrity": "sha512-oKnbhFyRIXpUuez8iBMmyEa4nbj4IOQyuhc/wy9kY7/WVPcwIO9VAU70SbNCHrUTAgGcgq0ee7EajR2m7G129==",
      "dependencies": {
        "ansi-styles": "^4.1.0",
        "supports-color": "^7.1.0"
      }
    }
  }
}
```

## Storage Engine & `node_modules` Layout

### Comparison of Storage Strategies

| Mechanism | npm / Yarn Classic | pnpm | Bun / scriptgo Design |
| :--- | :--- | :--- | :--- |
| **Directory Model** | Hoisted (Flat) | Hardlinks + Symlink Graph | Copy-on-Write / Hardlinks |
| **Phantom Dependencies** | Vulnerable | Fully Isolated | Isolated / Configurable |
| **Disk Overhead** | High (Duplicated) | Minimal (Shared 1 copy) | Minimal (Shared extents) |
| **I/O Speed** | Slow (Userspace Copy) | Fast (VFS Hardlinks) | Instant ($O(1)$ Syscalls) |

### Global Content-Addressable Storage (CAS)

The global store is maintained in `~/.scriptgo/store/v1/`:

```text
~/.scriptgo/store/v1/
├── index/                     # Manifest and resolution index
└── packages/
    ├── lodash@4.17.21/
    │   ├── package.json
    │   └── lodash.js
    └── chalk@4.1.2/
```

### OS-Level Linking Primitives

Instead of recursive userspace byte-copying, file allocation delegates directly
to kernel capabilities:

1. **macOS (APFS)**:
   Call `clonefile(const char *src, const char *dst, int flags)`.
   APFS creates a CoW metadata entry referencing identical storage blocks
   instantly in $O(1)$ time.
2. **Linux (Btrfs, XFS, ext4)**:
   - For Btrfs/XFS: Issue `ioctl(dst_fd, FICLONE, src_fd)` (reflink).
   - For ext4/generic filesystems: Fallback to `copy_file_range(2)` or
     hardlinking `link(src, dst)`.
3. **Windows (NTFS)**:
   Invoke `CreateHardLinkW` or directory junction points.

## Binary Linking (`.bin`) and Executable Shims

When a package manifest contains a `"bin"` declaration:

```json
{
  "name": "typescript",
  "bin": {
    "tsc": "./bin/tsc",
    "tsserver": "./bin/tsserver"
  }
}
```

The linker performs two actions:
1. Creates target directory `node_modules/.bin/`.
2. Generates a relative symbolic link:
   `node_modules/.bin/tsc` $\rightarrow$ `../typescript/bin/tsc`.
3. Ensures execution permissions are granted (`chmod +x`).

On Windows platforms, the linker creates `.cmd` and PowerShell `.ps1` wrapper
shims directing invocation to the target script via `node` / `scriptgo`.

## Reference Architecture Implementation

Below is a reference Go implementation demonstrating stream decompression,
SHA-512 SRI verification, and concurrent untar directly into the package directory:

```go
package pkgmgr

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type PackageManifest struct {
	Dist struct {
		Tarball   string `json:"tarball"`
		Integrity string `json:"integrity"`
	} `json:"dist"`
}

// FetchAndExtract downloads, validates integrity, and unpacks a package archive.
func FetchAndExtract(pkg, version, targetDir string) error {
	reqURL := fmt.Sprintf("https://registry.npmjs.org/%s/%s", pkg, version)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.npm.install-v1+json; q=1.0, application/json; q=0.8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("registry query failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry returned status %d for %s", resp.StatusCode, reqURL)
	}

	var manifest PackageManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return fmt.Errorf("invalid manifest JSON: %w", err)
	}

	tarResp, err := http.Get(manifest.Dist.Tarball)
	if err != nil {
		return fmt.Errorf("tarball fetch error: %w", err)
	}
	defer tarResp.Body.Close()

	// Compute SHA-512 while uncompressing concurrently
	hasher := sha512.New()
	pipeReader, pipeWriter := io.Pipe()
	tee := io.MultiWriter(pipeWriter, hasher)

	go func() {
		defer pipeWriter.Close()
		_, _ = io.Copy(tee, tarResp.Body)
	}()

	destPkgDir := filepath.Join(targetDir, pkg)
	if err := os.MkdirAll(destPkgDir, 0755); err != nil {
		return err
	}

	if err := extractTarGz(pipeReader, destPkgDir); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Validate integrity digest
	if strings.HasPrefix(manifest.Dist.Integrity, "sha512-") {
		computed := "sha512-" + base64.StdEncoding.EncodeToString(hasher.Sum(nil))
		if computed != manifest.Dist.Integrity {
			_ = os.RemoveAll(destPkgDir)
			return fmt.Errorf("integrity violation: expected %s, got %s", manifest.Dist.Integrity, computed)
		}
	}

	return nil
}

func extractTarGz(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(header.Name, "package/")
		target := filepath.Join(dest, relPath)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}
```

## Scriptgo-Specific Architectural Optimizations

Integrating a bespoke package manager directly into `scriptgo` unlocks several
deep compiler-level synergies that standard, external managers (npm, yarn, pnpm)
cannot provide:

### 1. Pre-Compiled Typed IR & Native Object Caching

Standard package managers only cache raw `.js` or `.ts` source files. On every
subsequent application build, the compiler must re-parse, re-typecheck, and
re-emit LLVM IR for every third-party package.

A dedicated `scriptgo` package manager can pre-compile imported packages into
**Typed IR** (`internal/ir`) artifacts or native object files (`.o`) directly
within the global Content-Addressable Storage (`~/.scriptgo/store/`). During
`scriptgo build`, transitive dependencies are linked directly as pre-verified
objects, yielding near-instant incremental builds.

### 2. Pre-Flight Compatibility Triaging (Static vs. Dynamic)

`scriptgo` enforces a three-tiered compilation policy: **Static** (AOT native),
**Dynamic** (QuickJS-ng execution island), and **Unsupported**.

During `scriptgo add <pkg>` or `scriptgo install`, the package manager can
traverse the package AST before compilation begins:
- Identify if the package conforms 100% to the AOT native subset.
- Flag dynamic idioms requiring `--dynamic` classification early in the
  developer workflow.
- Report unsupported APIs (e.g., incompatible runtime reflection or unmapped
  host primitives) at installation time rather than failing late during backend
  code generation.

### 3. In-Memory Graph Sharing (Zero-I/O Resolution)

Because both `cmd/scriptgo` and the package manager engine are implemented in
Go, dependency resolution and program checking can share the same in-memory data
structures. `internal/frontend` does not need to traverse filesystem directories
to re-resolve imports through classic Node.js probing algorithms; it can ingest
the resolved dependency graph directly from memory.

### 4. Direct C / FFI Native Linking (Bypassing `node-gyp`)

Packages that bundle native C/C++ addons typically rely on complex build tooling
(`node-gyp`, Python, makefiles) targeting Node's V8 C++ ABI.

For `scriptgo`, native package manifests can declare standard C sources or
static archives that link directly into the `scriptgo` runtime ABI
(`internal/runtime/native`) and Clang/LLVM pipeline (`internal/backend/llvm`),
producing fully self-contained static binaries without runtime addon loaders.

### 5. Unified Single-Binary Toolchain

Users do not need to install Node.js, npm, or external runtimes. The single
`scriptgo` binary provides an end-to-end toolchain:

```bash
scriptgo add <pkg>     # Package management & CAS linking
scriptgo check         # TypeScript-Go frontend verification
scriptgo build         # AOT native compilation to Mach-O / ELF
scriptgo run           # Immediate execution
```

## Integration with scriptgo Roadmap

ScriptGo now compiles frontend-resolved local ESM/CommonJS package graphs into
persistent Dynamic islands. Package-manager work must feed that existing graph
contract rather than add a second resolver:

1. **Installation Contract**: Materialize deterministic CAS-linked
   `node_modules` trees that TypeScript-Go can resolve normally.
2. **Package Service**: Own registry, SemVer, lockfile, integrity, CAS, and link
   behavior in a focused internal package; `cmd/scriptgo` remains a thin caller.
3. **Builtin Commands**: Expose `scriptgo add` and `scriptgo install` only after
   the service contract and offline lockfile behavior are tested.
