package errorutil

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

// Color configuration for error messages
var (
	errorColor = color.New(color.FgRed, color.Bold)
)

// CLIError represents an error during CLI execution
// (Moved from cliutil)
type CLIError struct {
	Op  string
	Err error
}

func (e *CLIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("cliutil %s: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("cliutil %s", e.Op)
}

func (e *CLIError) Unwrap() error {
	return e.Err
}

// NoChangesError indicates no changes are needed for a file
// (Moved from cliutil)
type NoChangesError struct {
	FilePath string
}

func (e *NoChangesError) Error() string {
	return fmt.Sprintf("no changes needed for %s", e.FilePath)
}

// IsCLIError checks if an error is a CLIError
func IsCLIError(err error) bool {
	_, ok := err.(*CLIError)
	return ok
}

// GetCLIErrorOp extracts the operation from a CLIError
func GetCLIErrorOp(err error) string {
	if cliErr, ok := err.(*CLIError); ok {
		return cliErr.Op
	}
	return ""
}

// Error-type detection helpers
func IsFileNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "does not exist") ||
		strings.Contains(err.Error(), "file not found")
}

func IsPermissionError(err error) bool {
	return strings.Contains(err.Error(), "permission denied") ||
		strings.Contains(err.Error(), "Permission denied")
}

func IsValidationError(err error) bool {
	return strings.Contains(err.Error(), "validation error") ||
		strings.Contains(err.Error(), "validation failed")
}

func IsParsingError(err error) bool {
	return strings.Contains(err.Error(), "syntax error") ||
		strings.Contains(err.Error(), "parsing error")
}

func IsFormattingError(err error) bool {
	return strings.Contains(err.Error(), "formatting error")
}

func IsSortingError(err error) bool {
	return strings.Contains(err.Error(), "sorting error")
}

// Error printing functions
func PrintError(err error, stderr io.Writer) {
	if err == nil {
		return
	}

	switch {
	case IsFileNotFoundError(err):
		PrintFileNotFoundError(err, stderr)
	case IsPermissionError(err):
		PrintPermissionError(err, stderr)
	case IsValidationError(err):
		PrintValidationError(err, stderr)
	case IsParsingError(err):
		PrintParsingError(err, stderr)
	case IsFormattingError(err):
		PrintFormattingError(err, stderr)
	case IsSortingError(err):
		PrintSortingError(err, stderr)
	default:
		PrintGenericError(err, stderr)
	}
}

func PrintFileNotFoundError(err error, stderr io.Writer) {
	_, _ = errorColor.Fprintf(stderr, "❌ File not found: %v\n", err)
}

func PrintPermissionError(err error, stderr io.Writer) {
	_, _ = errorColor.Fprintf(stderr, "🔒 Permission denied: %v\n", err)
}

func PrintValidationError(err error, stderr io.Writer) {
	_, _ = errorColor.Fprintf(stderr, "⚠️  Validation error: %v\n", err)
}

func PrintParsingError(err error, stderr io.Writer) {
	_, _ = errorColor.Fprintf(stderr, "🔍 Syntax error: %v\n", err)
}

func PrintFormattingError(err error, stderr io.Writer) {
	_, _ = errorColor.Fprintf(stderr, "🎨 Formatting error: %v\n", err)
}

func PrintSortingError(err error, stderr io.Writer) {
	_, _ = errorColor.Fprintf(stderr, "📊 Sorting error: %v\n", err)
}

func PrintGenericError(err error, stderr io.Writer) {
	_, _ = errorColor.Fprintf(stderr, "❌ Error: %v\n", err)
}

// Error extraction helpers
func ExtractFilePath(err error) string {
	errStr := err.Error()
	patterns := []string{
		"failed to read file:",
		"failed to parse",
		"error in",
		"validation error in",
		"syntax error in",
		"formatting error in",
		"sorting error in",
	}
	for _, pattern := range patterns {
		if idx := strings.Index(errStr, pattern); idx != -1 {
			afterPattern := errStr[idx+len(pattern):]
			path := strings.TrimSpace(afterPattern)
			path = strings.Trim(path, " :")
			path = strings.Trim(path, `"'`)
			return path
		}
	}
	return ""
}

func ExtractErrorMessage(err error) string {
	errStr := err.Error()
	prefixes := []string{
		"failed to read file:",
		"failed to parse",
		"error in",
		"validation error in",
		"syntax error in",
		"formatting error in",
		"sorting error in",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(errStr, prefix) {
			msg := strings.TrimPrefix(errStr, prefix)
			msg = strings.TrimSpace(msg)
			msg = strings.Trim(msg, " :")
			return msg
		}
	}
	return errStr
}

// NewCLIError constructs a new CLIError
func NewCLIError(op string, err error) *CLIError {
	return &CLIError{Op: op, Err: err}
}

// NewNoChangesError constructs a new NoChangesError
func NewNoChangesError(filePath string) *NoChangesError {
	return &NoChangesError{FilePath: filePath}
}
