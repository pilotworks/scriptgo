package compiler

import "os"

// Version is the compiler version embedded in generated artifacts. Releases
// can override it with -ldflags -X.
var Version = "dev"

const RuntimeABIVersion = "scriptgo.runtime.v1"

// BuildOptions controls target-specific native compilation.
type BuildOptions struct {
	CC               string
	Target           string
	Debug            bool
	OptLevel         string
	Sanitizers       []string
	WarnRuntimeCasts bool
	StrictCasts      bool
	FFIManifests     []string
	LinkFlags        []string
	ExtraSources     []string
	TSConfig         string
}

func (options BuildOptions) normalized() BuildOptions {
	if options.CC == "" {
		if envCC := os.Getenv("SCRIPTGO_CC"); envCC != "" {
			options.CC = envCC
		} else {
			options.CC = "clang"
		}
	}
	if options.Target == "" {
		if envTarget := os.Getenv("SCRIPTGO_TARGET"); envTarget != "" {
			options.Target = envTarget
		} else {
			options.Target = "native"
		}
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
	return options
}
