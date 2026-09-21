package compiler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
		buildArgs = append(buildArgs, "-mcpu="+options.TargetCPU)
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
	h.Write(runtime.QuickJSSource())
	h.Write(runtime.QuickJSHeader())
	h.Write([]byte(options.Target))
	h.Write([]byte(options.OptLevel))
	h.Write([]byte(strings.Join(options.Sanitizers, ",")))
	h.Write([]byte("quickjs-sections-v1"))
	if options.TargetCPU != "" {
		h.Write([]byte("cpu:" + options.TargetCPU))
	}
	if options.Debug {
		h.Write([]byte("debug"))
	}
	hash := hex.EncodeToString(h.Sum(nil))
	objPath := filepath.Join(sgCache, fmt.Sprintf("quickjs-%s.o", hash[:16]))

	if info, err := os.Stat(objPath); err == nil && info.Size() > 0 {
		return objPath, nil
	}

	headerPath := filepath.Join(sgCache, "quickjs.h")
	if err := os.WriteFile(headerPath, runtime.QuickJSHeader(), 0o644); err != nil {
		return "", err
	}

	tmpSrcPath := filepath.Join(sgCache, fmt.Sprintf("quickjs-%s-%d.c", hash[:16], os.Getpid()))
	if err := os.WriteFile(tmpSrcPath, runtime.QuickJSSource(), 0o644); err != nil {
		return "", err
	}
	defer os.Remove(tmpSrcPath)

	tmpObjPath := filepath.Join(sgCache, fmt.Sprintf("quickjs-%s-%d.o", hash[:16], os.Getpid()))
	defer os.Remove(tmpObjPath)

	qjsArgs := append([]string{}, ccParts[1:]...)
	qjsArgs = append(qjsArgs, "-I", sgCache, "-ffunction-sections", "-fdata-sections")
	if options.OptLevel != "" {
		qjsArgs = append(qjsArgs, "-O"+options.OptLevel)
	} else if options.Debug {
		qjsArgs = append(qjsArgs, "-O0", "-g")
	} else {
		qjsArgs = append(qjsArgs, "-O2")
	}
	if options.TargetCPU != "" {
		qjsArgs = append(qjsArgs, "-mcpu="+options.TargetCPU)
	}
	if options.Target != "native" {
		qjsArgs = append(qjsArgs, "--target="+options.Target)
	}
	if len(options.Sanitizers) > 0 {
		qjsArgs = append(qjsArgs, "-fsanitize="+strings.Join(options.Sanitizers, ","))
	}
	qjsArgs = append(qjsArgs, "-c", tmpSrcPath, "-o", tmpObjPath)

	cmd := exec.Command(ccParts[0], qjsArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("compile QuickJS-ng: %w: %s", err, string(out))
	}
	if err := os.Rename(tmpObjPath, objPath); err != nil {
		return "", err
	}
	return objPath, nil
}
