package main

import (
	"os"
	"sorttf/utils/cliutil"
)

func main() {
	os.Exit(cliutil.RunCLI(os.Args[1:]))
}
