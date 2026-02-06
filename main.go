package main

import (
	"os"

	"sorttf/cli"
)

func main() {
	os.Exit(cli.RunCLI(os.Args[1:]))
}
