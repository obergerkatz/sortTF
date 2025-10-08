package main

import (
	"os"
	"sorttf/utils/cliutil"
)

// Version information set at build time
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Set version info in cliutil if it supports it
	os.Exit(cliutil.RunCLI(os.Args[1:]))
}
