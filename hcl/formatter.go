package hcl

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// FormatHCLFile takes an hclwrite.File and returns the formatted string.
//
// It uses the canonical hclwrite formatting, which produces correctly formatted HCL
// according to the HashiCorp HCL specification. This is the same formatter used
// internally by terraform fmt.
//
// Returns the formatted content as a string, or an HCLError with KindFormatting if the file is nil.
func FormatHCLFile(file *hclwrite.File) (string, error) {
	if file == nil {
		return "", &HCLError{
			Op:   "FormatHCLFile",
			Kind: KindFormatting,
			Err:  fmt.Errorf("nil file provided"),
		}
	}

	// hclwrite.Bytes() returns canonically formatted HCL
	return string(file.Bytes()), nil
}

// FormatHCLString formats a raw HCL string.
//
// It first parses the string to validate syntax, then returns the canonically
// formatted output using hclwrite. Useful for formatting HCL content that is
// not yet written to a file.
//
// Returns the formatted string, or an error if parsing fails.
func FormatHCLString(content string) (string, error) {
	if content == "" {
		return "", nil
	}

	// Parse the content first to validate it
	file, diags := hclwrite.ParseConfig([]byte(content), "input", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return content, &HCLParseError{
			Path:  "input",
			Diags: diags,
		}
	}

	return FormatHCLFile(file)
}
