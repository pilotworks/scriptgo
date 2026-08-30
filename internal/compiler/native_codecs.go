package compiler

import (
	"os/exec"
	"strings"
	"sync"
)

type nativeCodecConfig struct {
	compileFlags []string
	linkFlags    []string
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
		nativeCodecs.linkFlags = append(nativeCodecs.linkFlags, "-lz")
		addPkgConfigCodec(&nativeCodecs, "SCRIPTGO_HAS_BROTLI", "libbrotlienc", "libbrotlidec")
		addPkgConfigCodec(&nativeCodecs, "SCRIPTGO_HAS_ZSTD", "libzstd")
	})
	return nativeCodecConfig{
		compileFlags: append([]string(nil), nativeCodecs.compileFlags...),
		linkFlags:    append([]string(nil), nativeCodecs.linkFlags...),
	}
}

func addPkgConfigCodec(config *nativeCodecConfig, define string, packages ...string) {
	cflags, cflagsOK := pkgConfig("--cflags", packages...)
	libs, libsOK := pkgConfig("--libs", packages...)
	if !cflagsOK || !libsOK {
		return
	}
	config.compileFlags = appendUnique(config.compileFlags, "-D"+define)
	config.compileFlags = appendUnique(config.compileFlags, cflags...)
	config.linkFlags = appendUnique(config.linkFlags, libs...)
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
