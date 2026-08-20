package compiler

// Version is the compiler version embedded in generated artifacts. Releases
// can override it with -ldflags -X.
var Version = "dev"

const RuntimeABIVersion = "scriptgo.runtime.v1"

// BuildOptions controls target-specific native compilation.
type BuildOptions struct {
	Target           string
	Debug            bool
	Sanitizers       []string
	WarnRuntimeCasts bool
	StrictCasts      bool
}

func (options BuildOptions) normalized() BuildOptions {
	if options.Target == "" {
		options.Target = "native"
	}
	return options
}
