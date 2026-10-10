package compiler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/pilotworks/scriptgo/internal/runtime"
)

var runtimeCacheMu sync.Mutex

func getScriptGoCacheDir() (string, error) {
	cacheDir := os.Getenv("SCRIPTGO_CACHE_DIR")
	if cacheDir == "" {
		var err error
		cacheDir, err = os.UserCacheDir()
		if err != nil {
			cacheDir = os.TempDir()
		}
		cacheDir = filepath.Join(cacheDir, "scriptgo")
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}
	return cacheDir, nil
}

func getOrBuildCachedRuntime(ccParts []string, options BuildOptions, codecConfig nativeCodecConfig, runtimeSource []byte, dynamic bool) (string, error) {
	if len(ccParts) == 0 {
		return "", fmt.Errorf("no C compiler specified")
	}

	runtimeCacheMu.Lock()
	defer runtimeCacheMu.Unlock()

	sgCache, err := getScriptGoCacheDir()
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(runtimeSource)
	if dynamic || bytes.Contains(runtimeSource, []byte("#include \"quickjs.h\"")) {
		h.Write(runtime.QuickJSSource())
		h.Write(runtime.QuickJSHeader())
	}
	h.Write([]byte(options.Target))
	h.Write([]byte(options.OptLevel))
	h.Write([]byte(strings.Join(options.Sanitizers, ",")))
	h.Write([]byte("sections-v1"))
	h.Write([]byte(codecConfig.cacheKey()))
	if options.LTO != "" && options.LTO != "none" {
		h.Write([]byte("lto:" + options.LTO))
	}
	if options.TargetCPU != "" {
		h.Write([]byte("cpu:" + options.TargetCPU))
	}
	if options.Debug {
		h.Write([]byte("debug"))
	}
	hash := hex.EncodeToString(h.Sum(nil))
	objPath := filepath.Join(sgCache, fmt.Sprintf("runtime-%s.o", hash[:16]))

	if info, err := os.Stat(objPath); err == nil && info.Size() > 0 {
		return objPath, nil
	}

	tmpSrcPath := filepath.Join(sgCache, fmt.Sprintf("runtime-%s-%d.c", hash[:16], os.Getpid()))
	if err := os.WriteFile(filepath.Join(sgCache, "scriptgo_value.h"), []byte(runtime.ValueHeader), 0o644); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(sgCache, "yyjson.h"), []byte(runtime.YYJSONHeader), 0o644); err != nil {
		return "", err
	}
	if dynamic || bytes.Contains(runtimeSource, []byte("#include \"quickjs.h\"")) {
		if err := os.WriteFile(filepath.Join(sgCache, "quickjs.h"), runtime.QuickJSHeader(), 0o644); err != nil {
			return "", err
		}
	}
	runtimeInput := append([]byte("#line 1 \"scriptgo-runtime.c\"\n"), runtimeSource...)
	if err := os.WriteFile(tmpSrcPath, runtimeInput, 0o644); err != nil {
		return "", err
	}
	defer os.Remove(tmpSrcPath)

	tmpObjPath := filepath.Join(sgCache, fmt.Sprintf("runtime-%s-%d.o", hash[:16], os.Getpid()))
	defer os.Remove(tmpObjPath)

	buildArgs := append([]string(nil), ccParts[1:]...)
	buildArgs = append(buildArgs, codecConfig.compileFlags...)
	buildArgs = append(buildArgs, "-I", sgCache, "-ffunction-sections", "-fdata-sections", "-c")
	buildArgs = append(buildArgs, tmpSrcPath, "-o", tmpObjPath)
	if options.OptLevel != "" {
		buildArgs = append(buildArgs, "-O"+options.OptLevel)
		if options.Debug {
			buildArgs = append(buildArgs, "-g")
		}
	} else if options.Debug {
		buildArgs = append(buildArgs, "-O0", "-g")
	} else {
		buildArgs = append(buildArgs, "-O2")
	}
	if options.LTO == "thin" {
		buildArgs = append(buildArgs, "-flto=thin")
	} else if options.LTO == "full" || options.LTO == "yes" || options.LTO == "true" {
		buildArgs = append(buildArgs, "-flto")
	}
	if options.TargetCPU != "" {
		buildArgs = append(buildArgs, targetCPUFlag(options.Target, options.TargetCPU))
	}
	if options.Target != "native" {
		buildArgs = append(buildArgs, "--target="+options.Target)
	}
	if len(options.Sanitizers) > 0 {
		buildArgs = append(buildArgs, "-fsanitize="+strings.Join(options.Sanitizers, ","))
	}
	cmd := exec.Command(ccParts[0], buildArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("compile runtime: %w: %s", err, string(out))
	}
	if err := os.Rename(tmpObjPath, objPath); err != nil {
		return "", err
	}
	return objPath, nil
}

func getOrBuildCachedQuickJS(ccParts []string, options BuildOptions) (string, error) {
	return getOrBuildCachedObject(ccParts, options, cachedObject{
		name:    "quickjs",
		label:   "QuickJS-ng",
		version: "quickjs-sections-v1",
		source:  runtime.QuickJSSource(),
		headers: map[string][]byte{"quickjs.h": runtime.QuickJSHeader()},
	})
}

// getOrBuildCachedRegExpEngine compiles the ECMAScript regular expression
// engine (see runtime.RegExpEngineSource), which every executable links.
func getOrBuildCachedRegExpEngine(ccParts []string, options BuildOptions) (string, error) {
	source, err := runtime.RegExpEngineSource()
	if err != nil {
		return "", err
	}
	return getOrBuildCachedObject(ccParts, options, cachedObject{
		name:    "regexp",
		label:   "regular expression engine",
		version: "regexp-sections-v1",
		source:  source,
	})
}

// regExpEngineObject is the regular expression engine object for a build:
// the cached one, or one compiled into temporaryDir when the cache is not
// writable.
func regExpEngineObject(ccParts []string, options BuildOptions, temporaryDir string) (string, error) {
	if obj, err := getOrBuildCachedRegExpEngine(ccParts, options); err == nil {
		return obj, nil
	}
	source, err := runtime.RegExpEngineSource()
	if err != nil {
		return "", err
	}
	srcPath := filepath.Join(temporaryDir, "regexp_engine.c")
	objPath := filepath.Join(temporaryDir, "regexp_engine.o")
	if err := os.WriteFile(srcPath, source, 0o644); err != nil {
		return "", fmt.Errorf("write regular expression engine source: %w", err)
	}
	if err := compileCObject(ccParts, options, "regular expression engine", srcPath, objPath, temporaryDir); err != nil {
		return "", err
	}
	return objPath, nil
}

// cachedObject is a C translation unit compiled once per source and build
// configuration into the ScriptGo cache directory.
type cachedObject struct {
	name    string
	label   string
	version string
	source  []byte
	headers map[string][]byte
}

func getOrBuildCachedObject(ccParts []string, options BuildOptions, object cachedObject) (string, error) {
	if len(ccParts) == 0 {
		return "", fmt.Errorf("no C compiler specified")
	}

	runtimeCacheMu.Lock()
	defer runtimeCacheMu.Unlock()

	sgCache, err := getScriptGoCacheDir()
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(object.source)
	for _, name := range slices.Sorted(maps.Keys(object.headers)) {
		h.Write([]byte(name))
		h.Write(object.headers[name])
	}
	h.Write([]byte(options.Target))
	h.Write([]byte(options.OptLevel))
	h.Write([]byte(strings.Join(options.Sanitizers, ",")))
	h.Write([]byte(object.version))
	if options.TargetCPU != "" {
		h.Write([]byte("cpu:" + options.TargetCPU))
	}
	if options.Debug {
		h.Write([]byte("debug"))
	}
	hash := hex.EncodeToString(h.Sum(nil))
	objPath := filepath.Join(sgCache, fmt.Sprintf("%s-%s.o", object.name, hash[:16]))

	if info, err := os.Stat(objPath); err == nil && info.Size() > 0 {
		return objPath, nil
	}

	for _, name := range slices.Sorted(maps.Keys(object.headers)) {
		if err := os.WriteFile(filepath.Join(sgCache, name), object.headers[name], 0o644); err != nil {
			return "", err
		}
	}

	tmpSrcPath := filepath.Join(sgCache, fmt.Sprintf("%s-%s-%d.c", object.name, hash[:16], os.Getpid()))
	if err := os.WriteFile(tmpSrcPath, object.source, 0o644); err != nil {
		return "", err
	}
	defer os.Remove(tmpSrcPath)

	tmpObjPath := filepath.Join(sgCache, fmt.Sprintf("%s-%s-%d.o", object.name, hash[:16], os.Getpid()))
	defer os.Remove(tmpObjPath)

	if err := compileCObject(ccParts, options, object.label, tmpSrcPath, tmpObjPath, sgCache); err != nil {
		return "", err
	}
	if err := os.Rename(tmpObjPath, objPath); err != nil {
		return "", err
	}
	return objPath, nil
}

// compileCObject compiles one C translation unit with the flags of the
// executable it is linked into, sections split for linker dead code
// elimination.
func compileCObject(ccParts []string, options BuildOptions, label, srcPath, objPath, includeDir string) error {
	args := append([]string{}, ccParts[1:]...)
	args = append(args, "-I", includeDir, "-ffunction-sections", "-fdata-sections")
	if options.OptLevel != "" {
		args = append(args, "-O"+options.OptLevel)
	} else if options.Debug {
		args = append(args, "-O0", "-g")
	} else {
		args = append(args, "-O2")
	}
	if options.TargetCPU != "" {
		args = append(args, targetCPUFlag(options.Target, options.TargetCPU))
	}
	if options.Target != "native" {
		args = append(args, "--target="+options.Target)
	}
	if len(options.Sanitizers) > 0 {
		args = append(args, "-fsanitize="+strings.Join(options.Sanitizers, ","))
	}
	args = append(args, "-c", srcPath, "-o", objPath)
	if out, err := exec.Command(ccParts[0], args...).CombinedOutput(); err != nil {
		return fmt.Errorf("compile %s: %w: %s", label, err, string(out))
	}
	return nil
}
