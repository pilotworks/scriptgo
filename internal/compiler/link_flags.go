package compiler

import (
	"fmt"
	"os/exec"
	goRuntime "runtime"
	"strings"
)

func isDarwinTarget(target string) bool {
	t := strings.ToLower(target)
	return strings.Contains(t, "darwin") || strings.Contains(t, "macos") || strings.Contains(t, "ios") || strings.Contains(t, "apple") || ((t == "native" || t == "") && goRuntime.GOOS == "darwin")
}

// targetCPUFlag selects the CPU to tune for: clang spells it -march on x86
// and -mcpu on ARM, AArch64 and RISC-V.
func targetCPUFlag(target, cpu string) string {
	arch := strings.ToLower(target)
	if arch == "native" || arch == "" {
		arch = goRuntime.GOARCH
	}
	if strings.HasPrefix(arch, "x86") || strings.HasPrefix(arch, "amd64") || strings.HasPrefix(arch, "i386") ||
		strings.HasPrefix(arch, "i686") || arch == "386" {
		return "-march=" + cpu
	}
	return "-mcpu=" + cpu
}

func isWindowsTarget(target string) bool {
	t := strings.ToLower(target)
	return strings.Contains(t, "windows") || strings.Contains(t, "mingw") || ((t == "native" || t == "") && goRuntime.GOOS == "windows")
}

// runtimeLibraryFlags links the native libraries the runtime may use (libm
// and the codec libraries: zlib, OpenSSL, brotli, zstd, resolv) only when the
// program references them after dead code elimination, so a program that
// never compresses, hashes or resolves names does not load those libraries
// at startup or require them on the machine it runs on. Under --as-needed,
// lld (which linkerDCEFlags selects when installed) drops a library only
// referenced by sections --gc-sections removed; GNU ld decides before garbage
// collection and keeps it, linking as before. Apple's linker drops it with
// -dead_strip_dylibs. Libraries named by FFI manifests are linked as given,
// after this group.
func runtimeLibraryFlags(target string, libraries []string) []string {
	if len(libraries) == 0 {
		return nil
	}
	if isDarwinTarget(target) {
		return append(append([]string(nil), libraries...), "-Wl,-dead_strip_dylibs")
	}
	if isWindowsTarget(target) || strings.HasPrefix(strings.ToLower(target), "wasm") {
		return append([]string(nil), libraries...)
	}
	flags := []string{"-Wl,--as-needed"}
	flags = append(flags, libraries...)
	return append(flags, "-Wl,--no-as-needed")
}

func linkerDCEFlags(target string) []string {
	if isDarwinTarget(target) {
		return []string{"-Wl,-dead_strip"}
	}
	flags := []string{"-Wl,--gc-sections"}
	if _, err := exec.LookPath("ld.lld"); err == nil {
		flags = append(flags, "-fuse-ld=lld")
	} else if _, err := exec.LookPath("lld"); err == nil {
		flags = append(flags, "-fuse-ld=lld")
	}
	return flags
}

func stripExecutable(outputPath, target string) error {
	t := strings.ToLower(target)
	if strings.HasPrefix(t, "wasm") {
		return nil
	}
	stripBin, err := exec.LookPath("strip")
	if err != nil {
		return nil
	}
	isDarwin := isDarwinTarget(target)
	var cmd *exec.Cmd
	if isDarwin {
		cmd = exec.Command(stripBin, "-x", outputPath)
	} else {
		cmd = exec.Command(stripBin, "--strip-all", outputPath)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		if !isDarwin {
			cmdFallback := exec.Command(stripBin, "-s", outputPath)
			if _, errFallback := cmdFallback.CombinedOutput(); errFallback == nil {
				return nil
			}
		}
		return fmt.Errorf("strip %s: %w: %s", outputPath, err, string(out))
	}
	return nil
}
