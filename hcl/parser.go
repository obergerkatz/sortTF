package hcl

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ParsedFile represents a parsed HCL file with diagnostics.
// It is returned by ParseHCLFile and contains the parsed structure
// along with any parser diagnostics (warnings or errors).
type ParsedFile struct {
	File  *hcl.File       // Parsed file structure
	Body  hcl.Body        // File body for attribute/block access
	Diags hcl.Diagnostics // Parser diagnostics (may be empty)
}

// ParseHCLFile reads and parses a .tf or .hcl file.
//
// It returns a ParsedFile containing the parsed structure and any diagnostics.
// If parsing fails, the error will be of type *HCLParseError.
// The ParsedFile is always returned, even on error, to allow inspection of partial results.
func ParseHCLFile(path string) (*ParsedFile, error) {
	if path == "" {
		return nil, &HCLError{
			Op:   "ParseHCLFile",
			Kind: KindParsing,
			Err:  fmt.Errorf("empty file path provided"),
		}
	}

	// Validate file exists and is accessible
	if err := validateFilePath(path); err != nil {
		return nil, &HCLError{
			Op:   "ParseHCLFile",
			Path: path,
			Kind: KindParsing,
			Err:  err,
		}
	}

	parser := hclparse.NewParser()
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, &HCLError{
			Op:   "ParseHCLFile",
			Path: path,
			Kind: KindParsing,
			Err:  err,
		}
	}

	file, diags := parser.ParseHCL(src, path)

	// Always return a ParsedFile, but include diagnostics
	parsedFile := &ParsedFile{File: file, Body: file.Body, Diags: diags}

	// If there are parsing errors, return them as a specific error type
	if diags.HasErrors() {
		return parsedFile, &HCLParseError{
			Path:  path,
			Diags: diags,
		}
	}

	return parsedFile, nil
}

// ValidateRequiredBlockLabels checks for required labels on Terraform block types.
//
// It validates that blocks have the correct number of labels according to Terraform conventions:
//   - resource, data: require exactly 2 labels (type, name)
//   - module, provider, variable, output: require exactly 1 label (name)
//   - locals, terraform: require no labels
//   - backend: must have 1 label and appear inside a terraform block
//
// Returns an HCLError with KindValidation if any blocks have incorrect label counts.
func ValidateRequiredBlockLabels(pf *ParsedFile) error {
	if pf == nil || pf.File == nil {
		return &HCLError{
			Op:   "ValidateRequiredBlockLabels",
			Kind: KindValidation,
			Err:  fmt.Errorf("parsed file is nil"),
		}
	}

	syntaxBody, ok := pf.File.Body.(*hclsyntax.Body)
	if !ok {
		return &HCLError{
			Op:   "ValidateRequiredBlockLabels",
			Kind: KindValidation,
			Err:  fmt.Errorf("file body is not hclsyntax.Body"),
		}
	}

	for _, block := range syntaxBody.Blocks {
		switch block.Type {
		case "resource", "data":
			if len(block.Labels) != 2 {
				return &HCLError{
					Op:   "ValidateRequiredBlockLabels",
					Kind: KindValidation,
					Err:  fmt.Errorf("%s block must have exactly 2 labels, got %d", block.Type, len(block.Labels)),
				}
			}
		case "module", "provider", "variable", "output":
			if len(block.Labels) != 1 {
				return &HCLError{
					Op:   "ValidateRequiredBlockLabels",
					Kind: KindValidation,
					Err:  fmt.Errorf("%s block must have exactly 1 label, got %d", block.Type, len(block.Labels)),
				}
			}
		case "locals", "terraform":
			if len(block.Labels) != 0 {
				return &HCLError{
					Op:   "ValidateRequiredBlockLabels",
					Kind: KindValidation,
					Err:  fmt.Errorf("%s block should not have labels: got %d", block.Type, len(block.Labels)),
				}
			}
		case "backend":
			// Backend blocks should only appear inside terraform blocks
			if len(block.Labels) != 1 {
				return &HCLError{
					Op:   "ValidateRequiredBlockLabels",
					Kind: KindValidation,
					Err:  fmt.Errorf("%s block must have exactly 1 label, got %d", block.Type, len(block.Labels)),
				}
			}
			return &HCLError{
				Op:   "ValidateRequiredBlockLabels",
				Kind: KindValidation,
				Err:  fmt.Errorf("backend block must be inside a terraform block"),
			}
		}
		// Special case: backend block must be inside terraform block
		if block.Type == "terraform" {
			for _, inner := range block.Body.Blocks {
				if inner.Type == "backend" && len(inner.Labels) != 1 {
					return &HCLError{
						Op:   "ValidateRequiredBlockLabels",
						Kind: KindValidation,
						Err:  fmt.Errorf("backend block inside terraform must have exactly 1 label, got %d", len(inner.Labels)),
					}
				}
			}
		}
	}
	return nil
}

// Helper functions

// validateFilePath checks if a file path is valid and accessible.
// It returns a user-friendly error message if the path is invalid,
// doesn't exist, has permission issues, or is a directory.
func validateFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path provided")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist")
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied")
		}
		return fmt.Errorf("failed to access file: %v", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, expected a file")
	}

	return nil
}
