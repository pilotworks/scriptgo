package main

import "flag"

func registerDynamicFlag(fs *flag.FlagSet) *bool {
	return fs.Bool("dynamic", false, "classify Dynamic-compatible sites (execution runtime is not yet available)")
}
