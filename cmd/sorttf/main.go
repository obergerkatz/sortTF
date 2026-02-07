// Command sorttf sorts and formats Terraform and Terragrunt files.
// See package doc.go for full documentation.
package main

import (
	"os"

	"github.com/obergerkatz/sortTF/cli"
)

func main() {
	os.Exit(cli.RunCLI(os.Args[1:]))
}
