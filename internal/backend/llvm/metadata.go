package llvm

import (
	"fmt"
	"strings"
)

func formatArtifactMetadata(options Options) string {
	mode := options.CompatibilityMode
	if mode == "" {
		mode = "static"
	}
	format := options.CompatibilityReportFormat
	if format == 0 {
		format = 1
	}
	var out strings.Builder
	fmt.Fprintf(&out, "; scriptgo.compiler = %q\n", options.CompilerVersion)
	fmt.Fprintf(&out, "; scriptgo.runtime-abi = %q\n", options.RuntimeABI)
	fmt.Fprintf(&out, "; scriptgo.target = %q\n", options.Target)
	if options.SourceHash != "" {
		fmt.Fprintf(&out, "; scriptgo.source-sha256 = %q\n", options.SourceHash)
	}
	fmt.Fprintf(&out, "; scriptgo.compatibility-mode = %q\n", mode)
	fmt.Fprintf(&out, "; scriptgo.compatibility-report-format = %q\n", fmt.Sprint(format))
	fmt.Fprintf(&out, "; scriptgo.static-sites = %q\n", fmt.Sprint(options.StaticSites))
	fmt.Fprintf(&out, "; scriptgo.dynamic-sites = %q\n", fmt.Sprint(options.DynamicSites))
	fmt.Fprintf(&out, "; scriptgo.unsupported-sites = %q\n", fmt.Sprint(options.UnsupportedSites))
	return out.String()
}
