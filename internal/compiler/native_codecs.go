package compiler

import (
	"bytes"
	"os/exec"
	goRuntime "runtime"
	"strings"
	"sync"
)

type nativeCodecConfig struct {
	compileFlags []string
	linkFlags    []string
}

func (config nativeCodecConfig) hasDefine(name string) bool {
	prefix := "-D" + name
	for _, flag := range config.compileFlags {
		if flag == prefix || strings.HasPrefix(flag, prefix+"=") {
			return true
		}
	}
	return false
}

var (
	nativeCodecsOnce sync.Once
	nativeCodecs     nativeCodecConfig
)

func nativeCodecConfigForTarget(target string) nativeCodecConfig {
	if !nativeTarget(target) {
		return nativeCodecConfig{}
	}
	nativeCodecsOnce.Do(func() {
		nativeCodecs.compileFlags = append(nativeCodecs.compileFlags, "-DSCRIPTGO_HAS_ZLIB")
		nativeCodecs.linkFlags = append(nativeCodecs.linkFlags, "-lz", "-lresolv")
		if goRuntime.GOOS == "darwin" {
			nativeCodecs.linkFlags = append(nativeCodecs.linkFlags, "-framework", "Security", "-framework", "CoreFoundation")
		}
		addPkgConfigCodec(&nativeCodecs, "SCRIPTGO_HAS_OPENSSL", "openssl")
		if !addPkgConfigCodec(&nativeCodecs, "SCRIPTGO_HAS_BROTLI", "libbrotlienc", "libbrotlidec") && linkAvailableFiles("libbrotlienc.so.1", "libbrotlidec.so.1", "libbrotlicommon.so.1") {
			nativeCodecs.compileFlags = appendUnique(nativeCodecs.compileFlags, "-DSCRIPTGO_HAS_BROTLI")
			nativeCodecs.linkFlags = appendUnique(nativeCodecs.linkFlags, "-Wl,-l:libbrotlienc.so.1", "-Wl,-l:libbrotlidec.so.1", "-Wl,-l:libbrotlicommon.so.1")
		}
		addPkgConfigCodec(&nativeCodecs, "SCRIPTGO_HAS_ZSTD", "libzstd")
		if !nativeCodecs.hasDefine("SCRIPTGO_HAS_ZSTD") && linkAvailableFiles("libzstd.so.1") {
			nativeCodecs.compileFlags = appendUnique(nativeCodecs.compileFlags, "-DSCRIPTGO_HAS_ZSTD")
			nativeCodecs.linkFlags = appendUnique(nativeCodecs.linkFlags, "-Wl,-l:libzstd.so.1")
		}
	})
	return nativeCodecConfig{
		compileFlags: append([]string(nil), nativeCodecs.compileFlags...),
		linkFlags:    append([]string(nil), nativeCodecs.linkFlags...),
	}
}

func addPkgConfigCodec(config *nativeCodecConfig, define string, packages ...string) bool {
	cflags, cflagsOK := pkgConfig("--cflags", packages...)
	libs, libsOK := pkgConfig("--libs", packages...)
	if !cflagsOK || !libsOK {
		return false
	}
	config.compileFlags = appendUnique(config.compileFlags, "-D"+define)
	config.compileFlags = appendUnique(config.compileFlags, cflags...)
	config.linkFlags = appendUnique(config.linkFlags, libs...)
	return true
}

func linkAvailable(libraries ...string) bool {
	cc, err := exec.LookPath("cc")
	if err != nil {
		cc, err = exec.LookPath("clang")
		if err != nil {
			return false
		}
	}
	args := []string{"-x", "c", "-", "-o", "/dev/null"}
	for _, library := range libraries {
		args = append(args, "-l"+library)
	}
	cmd := exec.Command(cc, args...)
	cmd.Stdin = bytes.NewBufferString("int main(void) { return 0; }\n")
	return cmd.Run() == nil
}

func linkAvailableFiles(files ...string) bool {
	cc, err := exec.LookPath("cc")
	if err != nil {
		cc, err = exec.LookPath("clang")
		if err != nil {
			return false
		}
	}
	args := []string{"-x", "c", "-", "-o", "/dev/null"}
	for _, file := range files {
		args = append(args, "-Wl,-l:"+file)
	}
	cmd := exec.Command(cc, args...)
	cmd.Stdin = bytes.NewBufferString("int main(void) { return 0; }\n")
	return cmd.Run() == nil
}

func pkgConfig(mode string, packages ...string) ([]string, bool) {
	path, err := exec.LookPath("pkg-config")
	if err != nil {
		return nil, false
	}
	args := append([]string{mode}, packages...)
	output, err := exec.Command(path, args...).Output()
	if err != nil {
		return nil, false
	}
	return strings.Fields(string(output)), true
}

func appendUnique(dst []string, values ...string) []string {
	seen := make(map[string]struct{}, len(dst)+len(values))
	for _, value := range dst {
		seen[value] = struct{}{}
	}
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		dst = append(dst, value)
	}
	return dst
}

func (config nativeCodecConfig) cacheKey() string {
	return strings.Join(config.compileFlags, "\x00") + "\x01" + strings.Join(config.linkFlags, "\x00")
}

func nativeTarget(target string) bool {
	t := strings.TrimSpace(strings.ToLower(target))
	return t == "" || t == "native"
}
