package main

import (
	"fmt"
	"os"
)

func printMainUsage() {
	fmt.Fprintln(os.Stderr, `ScriptGo - TypeScript Native Compiler

Usage:
  scriptgo <command> [flags] <arguments>

Commands:
  run       Compile and execute a TypeScript program or code string as a native binary
  build     Compile TypeScript into a standalone native executable
  check     Verify TypeScript syntax, types, and native subset rules
  emit      Emit LLVM IR or Typed IR
  coverage  Analyze Static/Dynamic site coverage
  install   Resolve, verify, cache, and link package.json dependencies
  version   Print compiler and runtime ABI version
  help      Show help for ScriptGo or a specific command

Global Flags:
  -v                     Verbose output
  --target <triple>      Target architecture triple (default: $SCRIPTGO_TARGET or native)
  --cc <driver>          C compiler / toolchain driver (default: $SCRIPTGO_CC or clang)
  --debug                Emit native DWARF debug symbols
  --lto <mode>           Enable link-time optimization (thin, full, none)
  --sanitize <list>      Enable Clang sanitizers (address, undefined, leak)
  --warn-runtime-casts   Warn on runtime checked casts (SG4005)
  --strict-casts         Treat cast warnings as errors
  --dynamic              Enable Dynamic-compatible JavaScript execution
  -h, --help             Show help message

Use 'scriptgo help <command>' or 'scriptgo <command> --help' for detailed command usage.`)
}

func printInstallUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo install [flags]

Description:
  Resolves production dependencies from an npm-compatible registry, verifies
  tarballs, stores immutable package content, writes scriptgo-lock.json, and
  links node_modules. Lifecycle scripts are not executed.

Flags:
  --project <dir>       Project directory (default: .)
  --manifest <path>     package.json path
  --lockfile <path>     Lockfile path
  --store <path>        Content store path
  --registry <url>      npm-compatible registry URL
  --offline             Use only the lockfile and cached tarballs
  --frozen              Use exact versions from the existing lockfile
  -h, --help            Show this help message`)
}

func printRunUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo run [flags] <entry.ts> [-- <args...>]
  scriptgo run [flags] -e "<code string>" [-- <args...>]

Description:
  Compiles a TypeScript file or inline code string to a temporary native binary
  and executes it directly on host.

Flags:
  -e <string>            Evaluate inline script string
  -m, --ffi-manifest     Path to FFI JSON metadata manifest (*.ffi.json)
  -v                     Verbose output (print compilation stages)
  --target <triple>      Target architecture triple (default: $SCRIPTGO_TARGET or native)
  --cc <driver>          C compiler / toolchain driver (default: $SCRIPTGO_CC or clang)
  --debug                Include DWARF debug symbols
  --lto <mode>           Enable link-time optimization (thin, full, none)
  --sanitize <list>      Enable Clang sanitizers (address, undefined, leak)
  --warn-runtime-casts   Warn on runtime checked casts (SG4005)
  --strict-casts         Treat cast warnings as errors
  --dynamic              Enable Dynamic-compatible JavaScript execution
  -h, --help             Show this help message

Examples:
  scriptgo run app.ts
  scriptgo run -e "console.log('hello ' + 42)"
  scriptgo run -e "console.log(100 * 20)"
  scriptgo run app.ts --ffi-manifest mylib.ffi.json
  scriptgo run app.ts helper.c
  scriptgo run --cc "zig cc" app.ts
  scriptgo run app.ts -- arg1 arg2`)
}

func printBuildUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo build [flags] <entry.ts> [sources.c...] [-o <output>]
  scriptgo build [flags] -e "<code string>" [-o <output>]

Description:
  Compiles a TypeScript program into a standalone, optimized native executable
  linked with the host C runtime.

Flags:
  -e <string>            Evaluate inline script string
  -o <path>              Output binary path (default: ./<entry_name>)
  -m, --ffi-manifest     Path to FFI JSON metadata manifest (*.ffi.json)
  -v                     Verbose output (print compilation stages)
  --target <triple>      Target architecture triple (default: $SCRIPTGO_TARGET or native)
  --cc <driver>          C compiler / toolchain driver (default: $SCRIPTGO_CC or clang)
  --debug                Include DWARF debug symbols (O0 with debug metadata)
  --lto <mode>           Enable link-time optimization (thin, full, none)
  --sanitize <list>      Enable Clang sanitizers (address, undefined, leak)
  --warn-runtime-casts   Warn on runtime checked casts (SG4005)
  --strict-casts         Treat cast warnings as errors
  --dynamic              Enable Dynamic-compatible JavaScript execution
  -h, --help             Show this help message

Examples:
  scriptgo build server.ts
  scriptgo build server.ts -o /usr/local/bin/server
  scriptgo build app.ts --ffi-manifest sqlite3.ffi.json -o myapp
  scriptgo build app.ts helper.c -o myapp
  scriptgo build cli.ts --cc "zig cc" --target x86_64-linux-gnu -o cli_linux
  scriptgo build cli.ts --debug --sanitize address -o cli_debug`)
}

func printCheckUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo check [flags] [<entry.ts> | <tsconfig.json> | <dir>]
  scriptgo check [flags] -p <path>
  scriptgo check [flags] -e "<code string>"

Description:
  Type-checks and validates the reachable source graph, tsconfig.json project,
  and native subset eligibility without invoking code generation or Clang.

Flags:
  -e <string>            Evaluate inline script string
  -p, --project <path>   Path to tsconfig.json or project directory
  -v                     Verbose output (print check stages and confirmation)
  --warn-runtime-casts   Warn on runtime checked casts (SG4005)
  --strict-casts         Treat cast warnings as errors
  --dynamic              Enable Dynamic-compatible JavaScript execution
  -h, --help             Show this help message

Examples:
  scriptgo check
  scriptgo check app.ts
  scriptgo check tsconfig.json
  scriptgo check -p ./src
  scriptgo check -e "const x: number = 42; console.log(x);"
  scriptgo check -v src/main.ts`)
}

func printEmitUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo emit [flags] <entry.ts> [--mode llvm-ir|typed-ir] [-o <output>]
  scriptgo emit [flags] -e "<code string>" [--mode llvm-ir|typed-ir] [-o <output>]

Description:
  Emits LLVM IR or Typed IR.

Flags:
  -e <string>            Evaluate inline script string
  --mode <mode>          Output mode: llvm-ir (default), typed-ir
  -o <path>              Write output to file instead of stdout
  -v                     Verbose output (print compilation stages)
  --target <triple>      Target architecture triple (default: $SCRIPTGO_TARGET, $TARGET, or native)
  --debug                Include DWARF debug symbols in LLVM IR
  --warn-runtime-casts   Warn on runtime checked casts (SG4005)
  --strict-casts         Treat cast warnings as errors
  --dynamic              Enable Dynamic-compatible JavaScript execution
  -h, --help             Show this help message

Examples:
  scriptgo emit app.ts
  scriptgo emit -e "console.log(123)" --mode typed-ir
  scriptgo emit app.ts --mode llvm-ir -o app.ll`)
}

func printCoverageUsage() {
	fmt.Fprintln(os.Stderr, `Usage:
  scriptgo coverage [flags] <entry.ts> [-o <output>]
  scriptgo coverage [flags] -e "<code string>" [-o <output>]

Description:
  Analyzes reachable source sites and summarizes Static, Dynamic, and
  Unsupported coverage. Use --format json for detailed machine-readable sites.

Flags:
  -e <string>            Evaluate inline script string
  --format <format>      Output format: summary (default), json
  -o <path>              Write output to file instead of stdout
  -v                     Verbose output (print analysis stages)
  --dynamic              Classify Dynamic-compatible sites
  -h, --help             Show this help message

Examples:
  scriptgo coverage app.ts
  scriptgo coverage app.ts --dynamic
  scriptgo coverage app.ts --dynamic --format json -o coverage.json`)
}
