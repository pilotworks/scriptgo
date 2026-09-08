package main

import "flag"

func registerDynamicFlag(fs *flag.FlagSet) *bool {
	return fs.Bool("dynamic", false, "enable Dynamic-compatible JavaScript execution")
}
