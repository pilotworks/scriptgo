package compiler

import (
	"os"
	"strings"
)

// Version is the compiler version embedded in generated artifacts. Releases
// can override it with -ldflags -X.
var Version = "dev"

const RuntimeABIVersion = "scriptgo.runtime.v1"

// BuildOptions controls target-specific native compilation.
type BuildOptions struct {
	CC               string
	Target           string
	WorkingDir       string
	Debug            bool
	OptLevel         string
	LTO              string
	Sanitizers       []string
	WarnRuntimeCasts bool
	StrictCasts      bool
	FFIManifests     []string
	LinkFlags        []string
	ExtraSources     []string
	TSConfig         string
	Dynamic          bool
	Strip            bool
	TargetCPU        string
	Release          bool
}

func (options BuildOptions) normalized() BuildOptions {
	if options.Target == "" {
		if envTarget := os.Getenv("SCRIPTGO_TARGET"); envTarget != "" {
			options.Target = envTarget
		} else {
			options.Target = "native"
		}
	}
	if options.CC == "" {
		if envCC := os.Getenv("SCRIPTGO_CC"); envCC != "" {
			options.CC = envCC
		} else if strings.HasPrefix(options.Target, "wasm32") {
			options.CC = "zig cc"
		} else {
			options.CC = "clang"
		}
	}
	if !options.Release {
		if envRel := os.Getenv("SCRIPTGO_RELEASE"); envRel == "1" || envRel == "true" {
			options.Release = true
		}
	}
	if options.Release {
		if options.OptLevel == "" {
			options.OptLevel = "3"
		}
		if options.LTO == "" {
			options.LTO = "thin"
		}
		options.Strip = true
	}
	if options.OptLevel == "" {
		if options.Debug {
			options.OptLevel = "0"
		} else if envOpt := os.Getenv("SCRIPTGO_OPT_LEVEL"); envOpt != "" {
			options.OptLevel = envOpt
		} else {
			options.OptLevel = "2"
		}
	}
	if options.LTO == "" {
		if envLTO := os.Getenv("SCRIPTGO_LTO"); envLTO != "" {
			options.LTO = envLTO
		}
	}
	if options.TargetCPU == "" {
		if envCPU := os.Getenv("SCRIPTGO_TARGET_CPU"); envCPU != "" {
			options.TargetCPU = envCPU
		}
	}
	if !options.Strip {
		if envStrip := os.Getenv("SCRIPTGO_STRIP"); envStrip == "1" || envStrip == "true" {
			options.Strip = true
		}
	}
	return options
}
