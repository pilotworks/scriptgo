# Foreign Function Interface (FFI)

`scriptgo` provides native-first interoperability with C and system libraries through **Static FFI** and **Version-Locked C Metadata Manifests** (`*.ffi.json`).

---

## 1. Static FFI (Direct C Linkage)

Because `scriptgo` compiles TypeScript directly to LLVM IR and machine code, external C functions can be called with **zero wrapper overhead** using standard TypeScript `declare function` syntax.

### Example: Calling C Standard Library

```typescript
// main.ts
declare function getpid(): number;
declare function abs(n: number): number;
declare function cos(x: number): number;
declare function sin(x: number): number;
declare function sqrt(x: number): number;
declare function puts(str: string): number;

console.log("PID:", getpid() > 0 ? "valid" : "invalid");
console.log("abs(-42):", abs(-42));
console.log("cos(0):", cos(0));
console.log("sqrt(16):", sqrt(16));
puts("Hello from native C puts!");
```

### Running and Building:

```bash
# Direct native execution (Development)
scriptgo run --native main.ts

# Build single standalone binary (Production)
scriptgo build main.ts -o myapp
./myapp
```

---

## 2. C Library JSON Metadata (`*.ffi.json`)

To link external libraries (`-l`), search directories (`-L`, `-I`), frameworks, or custom C source files (`.c`), `scriptgo` supports version-locked JSON manifests.

### Manifest Specification:

Every manifest MUST specify `"ffi_format": 1` to lock the schema specification version.

```json
{
  "ffi_format": 1,
  "name": "sqlite3",
  "link": {
    "libraries": ["sqlite3", "m"],
    "libDirs": ["/usr/local/lib", "/opt/homebrew/lib"],
    "includeDirs": ["/usr/local/include", "/opt/homebrew/include"],
    "sources": ["./native/sqlite_helper.c"],
    "cflags": ["-DSQLITE_ENABLE_JSON1=1"],
    "frameworks": []
  },
  "symbols": {
    "sqlite3_libversion": {
      "symbol": "sqlite3_libversion",
      "args": [],
      "returns": "cstring"
    },
    "sqlite3_open": {
      "symbol": "sqlite3_open",
      "args": ["cstring", "ptr"],
      "returns": "i32"
    },
    "sqlite3_close": {
      "symbol": "sqlite3_close",
      "args": ["ptr"],
      "returns": "i32"
    }
  }
}
```

### Manifest Schema Fields:

| Field | Type | Description |
|---|---|---|
| `ffi_format` | `number` (required) | Locked format version, must be `1` |
| `name` | `string` (required) | Library identifier |
| `link.libraries` | `string[]` | Library names to link with `-l` (e.g. `["m", "sqlite3"]`) |
| `link.libDirs` | `string[]` | Library search paths (`-L`) |
| `link.includeDirs` | `string[]` | Include header search paths (`-I`) |
| `link.sources` | `string[]` | C source files compiled directly with the binary |
| `link.cflags` | `string[]` | Extra C compiler flags |
| `link.frameworks` | `string[]` | macOS frameworks (e.g. `["Cocoa", "OpenGL"]`) |
| `symbols` | `object` | Map of exported C function signatures |

### Using Manifests with CLI:

```bash
# Pass manifest with --ffi-manifest (or -m)
scriptgo run --native main.ts --ffi-manifest sqlite3.ffi.json

# Compile single standalone executable
scriptgo build main.ts --ffi-manifest sqlite3.ffi.json -o myapp
./myapp
```

---

## 3. Custom C Source Interoperability

You can compile custom C source files directly alongside your TypeScript code without writing Makefiles or configuring external build tools:

### `helper.c`:
```c
int add_numbers(int a, int b) {
    return a + b;
}

double scale_float(double value, double factor) {
    return value * factor;
}
```

### `main.ts`:
```typescript
declare function add_numbers(a: number, b: number): number;
declare function scale_float(v: number, factor: number): number;

console.log("sum:", add_numbers(10, 20));
console.log("scale:", scale_float(3.5, 2.0));
```

### Run & Build:
```bash
# Pass helper.c directly to CLI
scriptgo run --native main.ts helper.c
scriptgo build main.ts helper.c -o myapp
./myapp
```

---

## 4. Type System & Calling Convention Mapping

| TypeScript Type | C Type | LLVM IR Type | Notes |
|---|---|---|---|
| `void` | `void` | `void` | No return value |
| `boolean` | `bool` / `int32_t` | `i1` | Zero-extended / truncated |
| `number` | `double` | `double` | IEEE-754 64-bit float |
| `bigint` | `int64_t` | `i64` | 64-bit signed/unsigned integer |
| `string` | `const char*` | `ptr` | Null-terminated UTF-8 byte pointer (`str.data`) |
| `Uint8Array` / `Buffer` | `uint8_t*` / `void*` | `ptr` | Raw memory byte buffer pointer |

---

## 5. Dynamic FFI Roadmap (`dlopen` / `dlsym`)

Dynamic loading of shared libraries at runtime (`dlopen`/`dlsym`) is planned for Phase 2 under the `scriptgo:ffi` module:

```typescript
// Planned Phase 2:
import { dlopen, FFIType } from "scriptgo:ffi";

const lib = dlopen("libm.dylib", {
    cos: { args: [FFIType.f64], returns: FFIType.f64 },
});
console.log(lib.symbols.cos(0));
```
