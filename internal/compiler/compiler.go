package compiler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goRuntime "runtime"
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/scriptgo"
	"github.com/pilotworks/scriptgo/internal/backend/llvm"
	"github.com/pilotworks/scriptgo/internal/frontend"
	"github.com/pilotworks/scriptgo/internal/ir"
	"github.com/pilotworks/scriptgo/internal/lowering"
	"github.com/pilotworks/scriptgo/internal/runtime"
)

// Compile reads one TypeScript entry point and returns LLVM IR.
func Compile(entryPath string) (string, error) {
	return compileWithOptions(entryPath, BuildOptions{})
}

// CompileWithOptions emits LLVM IR with the requested target metadata.
func CompileWithOptions(entryPath string, options BuildOptions) (string, error) {
	return compileWithOptions(entryPath, options)
}

func compileWithOptions(entryPath string, options BuildOptions) (string, error) {
	options = options.normalized()
	if err := validateOptions(options); err != nil {
		return "", err
	}
	module, err := CompileModuleWithOptions(entryPath, options)
	if err != nil {
		return "", err
	}
	if options.StrictCasts {
		warns := lowering.GetWarnings()
		if len(warns) > 0 {
			return "", fmt.Errorf("strict casts: %s: %s at offset %d: %s", warns[0].Code, warns[0].FileName, warns[0].Span.Start, warns[0].Message)
		}
	}
	hash := hashSources(module.SourceFiles)
	return llvm.EmitWithOptions(module, llvm.Options{
		CompilerVersion: Version,
		RuntimeABI:      RuntimeABIVersion,
		Target:          options.Target,
		SourceHash:      hash,
		Debug:           options.Debug,
	})
}

// GetWarnings returns any compiler warnings from the most recent lowering run.
func GetWarnings() []lowering.Warning {
	return lowering.GetWarnings()
}

func validateOptions(options BuildOptions) error {
	for _, sanitizer := range options.Sanitizers {
		switch sanitizer {
		case "address", "undefined", "leak":
		default:
			return fmt.Errorf("unsupported sanitizer %q; supported sanitizers: address, undefined, leak", sanitizer)
		}
	}
	if options.OptLevel != "" {
		switch options.OptLevel {
		case "0", "1", "2", "3", "s", "z", "fast":
		default:
			return fmt.Errorf("unsupported optimization level %q; supported: 0, 1, 2, 3, s, z, fast", options.OptLevel)
		}
	}
	if options.LTO != "" {
		switch options.LTO {
		case "thin", "full", "none", "auto", "yes", "true":
		default:
			return fmt.Errorf("unsupported lto %q; supported: thin, full, none", options.LTO)
		}
	}
	return nil
}

func hashSources(sources map[string]string) string {
	paths := make([]string, 0, len(sources))
	for path := range sources {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	hasher := sha256.New()
	for _, path := range paths {
		hasher.Write([]byte{0})
		hasher.Write([]byte(sources[path]))
		hasher.Write([]byte{0})
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

// CompileModule returns the verified typed IR for an entry point.
func CompileModule(entryPath string) (ir.Module, error) {
	return CompileModuleWithOptions(entryPath, BuildOptions{})
}

// CompileModuleWithOptions returns the verified typed IR for an entry point using custom build options.
func CompileModuleWithOptions(entryPath string, options BuildOptions) (ir.Module, error) {
	source, err := os.ReadFile(entryPath)
	if err != nil {
		return ir.Module{}, fmt.Errorf("read entry point %q: %w", entryPath, err)
	}

	program, err := frontend.NewProgramWithOptions(entryPath, string(source), frontend.ProgramOptions{
		ConfigPath: options.TSConfig,
	})
	if err != nil {
		return ir.Module{}, err
	}

	module, err := lowering.LowerWithOptions(program, lowering.Options{
		WarnRuntimeCasts: options.WarnRuntimeCasts,
	})
	if err != nil {
		return ir.Module{}, err
	}
	return module, nil
}

// CheckProject typechecks an entire tsconfig.json project and validates native subset rules for project files.
func CheckProject(configPath string, options BuildOptions) ([]typescriptgo.Diagnostic, error) {
	result, err := frontend.CheckProject(configPath)
	if err != nil {
		return nil, err
	}
	if len(result.Diagnostics) > 0 {
		return result.Diagnostics, nil
	}
	for _, file := range result.Files {
		if file.Builtin {
			continue
		}
		prog := frontend.Program{
			EntryPath:      file.FileName,
			Source:         file.Source,
			StatementCount: file.StatementCount,
			Options:        result.Options,
			Files:          result.Files,
		}
		if _, err := lowering.LowerWithOptions(prog, lowering.Options{
			WarnRuntimeCasts: options.WarnRuntimeCasts,
		}); err != nil {
			return nil, err
		}
		if options.StrictCasts {
			warns := lowering.GetWarnings()
			if len(warns) > 0 {
				return nil, fmt.Errorf("strict casts: %s: %s at offset %d: %s", warns[0].Code, warns[0].FileName, warns[0].Span.Start, warns[0].Message)
			}
		}
	}
	return nil, nil
}

// DumpIR returns the stable backend-independent IR artifact for an entry point.
func DumpIR(entryPath string) (string, error) {
	module, err := CompileModule(entryPath)
	if err != nil {
		return "", err
	}
	return module.Dump()
}

// Build compiles the generated LLVM IR into a native executable with clang.
func Build(entryPath, outputPath string) error {
	return BuildWithOptions(entryPath, outputPath, BuildOptions{})
}

// BuildWithOptions compiles a native executable with deterministic inputs.
func BuildWithOptions(entryPath, outputPath string, options BuildOptions) error {
	options = options.normalized()
	ccParts, err := resolveCC(options.CC)
	if err != nil {
		return err
	}
	output, err := compileWithOptions(entryPath, options)
	if err != nil {
		return err
	}
	codecConfig := nativeCodecConfigForTarget(options.Target)
	if strings.Contains(output, "call i32 @scriptgo_tls_") && !codecConfig.hasDefine("SCRIPTGO_HAS_OPENSSL") {
		return fmt.Errorf("node:tls requires OpenSSL development files available through pkg-config")
	}
	temporaryDir, err := os.MkdirTemp("", "scriptgo-build-")
	if err != nil {
		return fmt.Errorf("create temporary build directory: %w", err)
	}
	defer os.RemoveAll(temporaryDir)
	temporaryPath := filepath.Join(temporaryDir, "module.ll")
	if err := os.WriteFile(temporaryPath, []byte(output), 0o644); err != nil {
		return fmt.Errorf("write temporary LLVM file: %w", err)
	}
	var args []string
	runtimeObj, err := getOrBuildCachedRuntime(ccParts, options, codecConfig)
	if err != nil {
		runtimePath := filepath.Join(temporaryDir, "runtime.c")
		runtimeSource := append([]byte("#line 1 \"scriptgo-runtime.c\"\n"), runtime.Source...)
		if err := os.WriteFile(runtimePath, runtimeSource, 0o644); err != nil {
			return fmt.Errorf("write temporary runtime file: %w", err)
		}
		args = []string{temporaryPath, "-x", "c", runtimePath, "-x", "none"}
		args = append(args, codecConfig.compileFlags...)
	} else {
		args = []string{temporaryPath, runtimeObj}
	}
	args = append(args, codecConfig.linkFlags...)

	// Process FFI manifests and extra native sources
	var extraLibs []string
	var extraLibDirs []string
	var extraIncludeDirs []string
	var extraCFlags []string
	var extraFrameworks []string

	manifestsToLoad := append([]string(nil), options.FFIManifests...)
	if len(manifestsToLoad) == 0 && entryPath != "" {
		dir := filepath.Dir(entryPath)
		if entries, err := os.ReadDir(dir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".ffi.json") {
					manifestsToLoad = append(manifestsToLoad, filepath.Join(dir, entry.Name()))
				}
			}
		}
	}

	for _, manifestPath := range manifestsToLoad {
		manifest, err := LoadFFIManifest(manifestPath)
		if err != nil {
			return err
		}
		for _, src := range manifest.Link.Sources {
			args = append(args, "-x", "c", src)
		}
		extraLibs = append(extraLibs, manifest.Link.Libraries...)
		extraLibDirs = append(extraLibDirs, manifest.Link.LibDirs...)
		extraIncludeDirs = append(extraIncludeDirs, manifest.Link.IncludeDirs...)
		extraCFlags = append(extraCFlags, manifest.Link.CFlags...)
		extraFrameworks = append(extraFrameworks, manifest.Link.Frameworks...)
	}

	for _, src := range options.ExtraSources {
		args = append(args, "-x", "c", src)
	}

	args = append(args, "-x", "none", "-ffunction-sections", "-fdata-sections")
	if options.OptLevel != "" {
		args = append(args, "-O"+options.OptLevel)
	} else if options.Debug {
		args = append(args, "-O0")
	} else {
		args = append(args, "-O2")
	}
	if options.LTO == "thin" {
		args = append(args, "-flto=thin")
	} else if options.LTO == "full" || options.LTO == "yes" || options.LTO == "true" {
		args = append(args, "-flto")
	}
	if options.Target != "native" {
		args = append(args, "--target="+options.Target)
	}
	if options.Debug {
		args = append(args, "-g", "-fdebug-compilation-dir=.")
	}
	if len(options.Sanitizers) > 0 {
		args = append(args, "-fsanitize="+strings.Join(options.Sanitizers, ","))
	}
	for _, inc := range extraIncludeDirs {
		args = append(args, "-I"+inc)
	}
	for _, dir := range extraLibDirs {
		args = append(args, "-L"+dir)
	}
	args = append(args, extraCFlags...)
	args = append(args, "-o", filepath.Clean(outputPath), "-lm")
	for _, lib := range extraLibs {
		if strings.HasPrefix(lib, "-l") {
			args = append(args, lib)
		} else {
			args = append(args, "-l"+lib)
		}
	}
	for _, fw := range extraFrameworks {
		args = append(args, "-framework", fw)
	}
	args = append(args, linkerDCEFlags(options.Target)...)
	args = append(args, options.LinkFlags...)
	cmdArgs := append(ccParts[1:], args...)
	command := exec.Command(ccParts[0], cmdArgs...)
	if diagnostic, err := command.CombinedOutput(); err != nil {
		driverName := filepath.Base(ccParts[0])
		return fmt.Errorf("%s: %w: %s", driverName, err, diagnostic)
	}
	return nil
}

func resolveCC(cc string) ([]string, error) {
	return resolveCCWithLookup(cc, exec.LookPath)
}

func resolveCCWithLookup(cc string, lookPath func(string) (string, error)) ([]string, error) {
	cc = strings.TrimSpace(cc)
	if cc == "" || cc == "clang" {
		if bin, err := lookPath("clang"); err == nil {
			return []string{bin}, nil
		}
		if bin, err := lookPath("zig"); err == nil {
			return []string{bin, "cc"}, nil
		}
		return nil, fmt.Errorf("native backend requires \"clang\" or \"zig\" in PATH")
	}

	parts := strings.Fields(cc)
	if len(parts) == 0 {
		return resolveCCWithLookup("", lookPath)
	}
	if parts[0] == "zigcc" {
		if bin, err := lookPath("zigcc"); err == nil {
			parts[0] = bin
			return parts, nil
		}
		if bin, err := lookPath("zig"); err == nil {
			return append([]string{bin, "cc"}, parts[1:]...), nil
		}
	}
	if parts[0] == "zig" && len(parts) == 1 {
		if bin, err := lookPath("zig"); err == nil {
			return []string{bin, "cc"}, nil
		}
	}
	bin, err := lookPath(parts[0])
	if err != nil {
		return nil, fmt.Errorf("native backend requires %q in PATH: %w", parts[0], err)
	}
	parts[0] = bin
	return parts, nil
}

// Run executes the entry point by compiling it to a temporary native binary and running it.
func Run(entryPath string) (string, error) {
	return RunWithOptions(entryPath, BuildOptions{})
}

// RunWithOptions compiles the entry point to a temporary native binary with custom build options and runs it.
func RunWithOptions(entryPath string, options BuildOptions) (string, error) {
	tempDir, err := os.MkdirTemp("", "scriptgo-run-")
	if err != nil {
		return "", fmt.Errorf("create temporary run directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	binPath := filepath.Join(tempDir, "app")
	if err := BuildWithOptions(entryPath, binPath, options); err != nil {
		return "", err
	}
	cmd := exec.Command(binPath)
	if options.WorkingDir != "" {
		cmd.Dir = options.WorkingDir
	}
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		if output.Len() > 0 {
			return output.String(), fmt.Errorf("%w: %s", err, strings.TrimSpace(output.String()))
		}
		return output.String(), err
	}
	return output.String(), nil
}

// Check parses, type checks, and validates the native subset for an entry point.
func Check(entryPath string) error {
	source, err := os.ReadFile(entryPath)
	if err != nil {
		return fmt.Errorf("read entry point %q: %w", entryPath, err)
	}
	program, err := frontend.NewProgram(entryPath, string(source))
	if err != nil {
		return err
	}
	return lowering.ValidateSubset(program)
}

var runtimeCacheMu sync.Mutex

func getOrBuildCachedRuntime(ccParts []string, options BuildOptions, codecConfig nativeCodecConfig) (string, error) {
	if len(ccParts) == 0 {
		return "", fmt.Errorf("no C compiler specified")
	}

	runtimeCacheMu.Lock()
	defer runtimeCacheMu.Unlock()

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	sgCache := filepath.Join(cacheDir, "scriptgo")
	if err := os.MkdirAll(sgCache, 0o755); err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(runtime.Source)
	h.Write([]byte(options.Target))
	h.Write([]byte(options.OptLevel))
	h.Write([]byte(strings.Join(options.Sanitizers, ",")))
	h.Write([]byte("sections-v1"))
	h.Write([]byte(codecConfig.cacheKey()))
	if options.LTO != "" && options.LTO != "none" {
		h.Write([]byte("lto:" + options.LTO))
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
	runtimeSource := append([]byte("#line 1 \"scriptgo-runtime.c\"\n"), runtime.Source...)
	if err := os.WriteFile(tmpSrcPath, runtimeSource, 0o644); err != nil {
		return "", err
	}
	defer os.Remove(tmpSrcPath)

	tmpObjPath := filepath.Join(sgCache, fmt.Sprintf("runtime-%s-%d.o", hash[:16], os.Getpid()))
	defer os.Remove(tmpObjPath)

	buildArgs := append([]string(nil), ccParts[1:]...)
	buildArgs = append(buildArgs, codecConfig.compileFlags...)
	buildArgs = append(buildArgs, "-ffunction-sections", "-fdata-sections", "-c", tmpSrcPath, "-o", tmpObjPath)
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

func linkerDCEFlags(target string) []string {
	t := strings.ToLower(target)
	if t == "native" || t == "" {
		if goRuntime.GOOS == "darwin" {
			return []string{"-Wl,-dead_strip"}
		}
		return []string{"-Wl,--gc-sections"}
	}
	if strings.Contains(t, "darwin") || strings.Contains(t, "macos") || strings.Contains(t, "ios") || strings.Contains(t, "apple") {
		return []string{"-Wl,-dead_strip"}
	}
	return []string{"-Wl,--gc-sections"}
}
