package parsingutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ParsingError represents an error during HCL parsing
type ParsingError struct {
	Op   string
	Path string
	Err  error
}

func (e *ParsingError) Error() string {
	if e.Err != nil {
		if e.Path != "" {
			return fmt.Sprintf("parsingutil %s %s: %v", e.Op, e.Path, e.Err)
		}
		return fmt.Sprintf("parsingutil %s: %v", e.Op, e.Err)
	}
	if e.Path != "" {
		return fmt.Sprintf("parsingutil %s %s", e.Op, e.Path)
	}
	return fmt.Sprintf("parsingutil %s", e.Op)
}

func (e *ParsingError) Unwrap() error {
	return e.Err
}

// HCLParseError indicates HCL parsing failed with diagnostics
type HCLParseError struct {
	Path  string
	Diags hcl.Diagnostics
}

func (e *HCLParseError) Error() string {
	return fmt.Sprintf("HCL parsing failed for %s: %s", e.Path, e.Diags.Error())
}

// ValidationError indicates validation failed
type ValidationError struct {
	Op   string
	Path string
	Err  error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		if e.Path != "" {
			return fmt.Sprintf("validation %s %s: %v", e.Op, e.Path, e.Err)
		}
		return fmt.Sprintf("validation %s: %v", e.Op, e.Err)
	}
	if e.Path != "" {
		return fmt.Sprintf("validation %s %s", e.Op, e.Path)
	}
	return fmt.Sprintf("validation %s", e.Op)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

type ParsedFile struct {
	File  *hcl.File
	Body  hcl.Body
	Diags hcl.Diagnostics
}

// ParseHCLFile reads and parses a .tf or .hcl file, returning a ParsedFile struct
func ParseHCLFile(path string) (*ParsedFile, error) {
	if path == "" {
		return nil, &ParsingError{
			Op:  "ParseHCLFile",
			Err: fmt.Errorf("empty file path provided"),
		}
	}

	// Validate file exists and is accessible
	if err := validateFilePath(path); err != nil {
		return nil, &ParsingError{
			Op:   "ParseHCLFile",
			Path: path,
			Err:  err,
		}
	}

	parser := hclparse.NewParser()
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, &ParsingError{
			Op:   "ParseHCLFile",
			Path: path,
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

// ValidateRequiredBlockLabels checks for required labels on Terraform block types
func ValidateRequiredBlockLabels(pf *ParsedFile) error {
	if pf == nil || pf.File == nil {
		return &ValidationError{
			Op:  "ValidateRequiredBlockLabels",
			Err: fmt.Errorf("parsed file is nil"),
		}
	}

	syntaxBody, ok := pf.File.Body.(*hclsyntax.Body)
	if !ok {
		return &ValidationError{
			Op:  "ValidateRequiredBlockLabels",
			Err: fmt.Errorf("file body is not hclsyntax.Body"),
		}
	}

	for _, block := range syntaxBody.Blocks {
		switch block.Type {
		case "resource", "data":
			if len(block.Labels) != 2 {
				return &ValidationError{
					Op:  "ValidateRequiredBlockLabels",
					Err: fmt.Errorf("%s block must have exactly 2 labels, got %d", block.Type, len(block.Labels)),
				}
			}
		case "module", "provider", "variable", "output":
			if len(block.Labels) != 1 {
				return &ValidationError{
					Op:  "ValidateRequiredBlockLabels",
					Err: fmt.Errorf("%s block must have exactly 1 label, got %d", block.Type, len(block.Labels)),
				}
			}
		case "locals", "terraform":
			if len(block.Labels) != 0 {
				return &ValidationError{
					Op:  "ValidateRequiredBlockLabels",
					Err: fmt.Errorf("%s block should not have labels: got %d", block.Type, len(block.Labels)),
				}
			}
		case "backend":
			// Backend blocks should only appear inside terraform blocks
			if len(block.Labels) != 1 {
				return &ValidationError{
					Op:  "ValidateRequiredBlockLabels",
					Err: fmt.Errorf("%s block must have exactly 1 label, got %d", block.Type, len(block.Labels)),
				}
			}
			return &ValidationError{
				Op:  "ValidateRequiredBlockLabels",
				Err: fmt.Errorf("backend block must be inside a terraform block"),
			}
		}
		// Special case: backend block must be inside terraform block
		if block.Type == "terraform" {
			for _, inner := range block.Body.Blocks {
				if inner.Type == "backend" && len(inner.Labels) != 1 {
					return &ValidationError{
						Op:  "ValidateRequiredBlockLabels",
						Err: fmt.Errorf("backend block inside terraform must have exactly 1 label, got %d", len(inner.Labels)),
					}
				}
			}
		}
	}
	return nil
}

// Helper functions

// validateFilePath checks if a file path is valid and accessible
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

// Error checking helper functions

// IsParsingError checks if an error is a ParsingError
func IsParsingError(err error) bool {
	_, ok := err.(*ParsingError)
	return ok
}

// IsHCLParseError checks if the error indicates HCL parsing failed
func IsHCLParseError(err error) bool {
	_, ok := err.(*HCLParseError)
	return ok
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// IsNotExistError checks if the error indicates a file doesn't exist
func IsNotExistError(err error) bool {
	if parsingErr, ok := err.(*ParsingError); ok {
		return strings.Contains(parsingErr.Err.Error(), "does not exist")
	}
	return false
}

// IsPermissionError checks if the error indicates a permission issue
func IsPermissionError(err error) bool {
	if parsingErr, ok := err.(*ParsingError); ok {
		return strings.Contains(parsingErr.Err.Error(), "permission denied")
	}
	return false
}

// GetParsingErrorPath extracts the path from a ParsingError
func GetParsingErrorPath(err error) string {
	if parsingErr, ok := err.(*ParsingError); ok {
		return parsingErr.Path
	}
	return ""
}

// GetParsingErrorOp extracts the operation from a ParsingError
func GetParsingErrorOp(err error) string {
	if parsingErr, ok := err.(*ParsingError); ok {
		return parsingErr.Op
	}
	return ""
}

// GetValidationErrorOp extracts the operation from a ValidationError
func GetValidationErrorOp(err error) string {
	if validationErr, ok := err.(*ValidationError); ok {
		return validationErr.Op
	}
	return ""
}

// GetValidationErrorPath extracts the path from a ValidationError
func GetValidationErrorPath(err error) string {
	if validationErr, ok := err.(*ValidationError); ok {
		return validationErr.Path
	}
	return ""
}
