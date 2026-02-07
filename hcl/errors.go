package hcl

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// ErrorKind categorizes HCL errors for display and handling purposes.
// It allows callers to distinguish between different types of errors
// without relying on string matching.
type ErrorKind int

// Error kind constants for categorizing HCL errors.
const (
	// KindUnknown represents an unknown or uncategorized error.
	KindUnknown ErrorKind = iota
	// KindParsing represents an HCL parsing or syntax error.
	KindParsing
	// KindValidation represents a semantic validation error (e.g., wrong label count).
	KindValidation
	// KindFormatting represents an error during formatting.
	KindFormatting
	// KindSorting represents an error during sorting operation.
	KindSorting
)

//nolint:revive // exported: HCLError is intentionally named to indicate HCL-specific errors
// HCLError is the unified error type for HCL operations.
// It wraps an underlying error with context about the operation, file path, and error category.
type HCLError struct {
	Op   string    // Operation that failed (e.g., "ParseHCLFile", "SortBlocks")
	Path string    // File path (may be empty for in-memory operations)
	Kind ErrorKind // Error category
	Err  error     // Underlying error
}

// Error implements the error interface, returning a formatted error message
// that includes the operation name, file path (if available), and underlying error.
func (e *HCLError) Error() string {
	if e.Err != nil {
		if e.Path != "" {
			return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Err)
		}
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	if e.Path != "" {
		return fmt.Sprintf("%s %s", e.Op, e.Path)
	}
	return e.Op
}

// Unwrap returns the underlying error, allowing error unwrapping
// with errors.Is and errors.As.
func (e *HCLError) Unwrap() error {
	return e.Err
}

//nolint:revive // exported: HCLParseError is intentionally named to indicate HCL-specific parse errors
// HCLParseError represents a syntax error with diagnostics.
// This is kept separate from HCLError because it has a fundamentally different shape
// (holds hcl.Diagnostics instead of a simple error).
type HCLParseError struct {
	Path  string          // File path that failed to parse
	Diags hcl.Diagnostics // Parser diagnostics with error details
}

// Error implements the error interface, returning a formatted error message
// that includes the file path and the detailed diagnostics from the HCL parser.
func (e *HCLParseError) Error() string {
	return fmt.Sprintf("HCL parsing failed for %s: %s", e.Path, e.Diags.Error())
}

// Error checking functions

// IsParsingError checks if an error is a parsing error.
func IsParsingError(err error) bool {
	var hclErr *HCLError
	if errors.As(err, &hclErr) {
		return hclErr.Kind == KindParsing
	}
	return false
}

// IsHCLParseError checks if an error is an HCL syntax error.
func IsHCLParseError(err error) bool {
	var hclParseErr *HCLParseError
	return errors.As(err, &hclParseErr)
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	var hclErr *HCLError
	if errors.As(err, &hclErr) {
		return hclErr.Kind == KindValidation
	}
	return false
}

// IsFormattingError checks if an error is a formatting error.
func IsFormattingError(err error) bool {
	var hclErr *HCLError
	if errors.As(err, &hclErr) {
		return hclErr.Kind == KindFormatting
	}
	return false
}

// IsSortingError checks if an error is a sorting error.
func IsSortingError(err error) bool {
	var hclErr *HCLError
	if errors.As(err, &hclErr) {
		return hclErr.Kind == KindSorting
	}
	return false
}

// IsNotExistError checks if an error indicates a file doesn't exist.
func IsNotExistError(err error) bool {
	var hclErr *HCLError
	if errors.As(err, &hclErr) {
		if hclErr.Err != nil {
			errMsg := hclErr.Err.Error()
			return containsAny(errMsg, "does not exist", "no such file")
		}
	}
	return false
}

// IsPermissionError checks if an error indicates a permission issue.
func IsPermissionError(err error) bool {
	var hclErr *HCLError
	if errors.As(err, &hclErr) {
		if hclErr.Err != nil {
			errMsg := hclErr.Err.Error()
			return containsAny(errMsg, "permission denied", "access denied")
		}
	}
	return false
}

// containsAny checks if string s contains any of the given substrings.
// Returns true on the first match found.
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
