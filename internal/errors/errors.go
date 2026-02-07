// Package errors provides unified error handling for sortTF.
//
// This package defines sentinel errors for common conditions and a single
// Error type that wraps errors with context (operation, path, kind).
//
// Error checking should use errors.Is and errors.As from the standard library
// rather than string matching or type assertions.
//
//nolint:revive // var-naming: errors is an appropriate name for an error handling package
package errors

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// Sentinel errors for common conditions.
// Use these with errors.Is() for error checking.
var (
	// ErrFileNotFound indicates a file or directory doesn't exist
	ErrFileNotFound = errors.New("file not found")

	// ErrPermissionDenied indicates insufficient permissions
	ErrPermissionDenied = errors.New("permission denied")

	// ErrInvalidSyntax indicates HCL parsing failed
	ErrInvalidSyntax = errors.New("invalid HCL syntax")

	// ErrValidation indicates validation failed
	ErrValidation = errors.New("validation failed")

	// ErrTerraformNotFound indicates terraform command is not available
	ErrTerraformNotFound = errors.New("terraform command not found")

	// ErrNoChanges indicates a file doesn't need changes
	ErrNoChanges = errors.New("no changes needed")
)

// ErrorKind categorizes errors for display purposes.
// Different kinds may be displayed with different colors or formatting.
type ErrorKind int

// Error kind constants for categorizing sortTF errors.
const (
	// KindUnknown represents an uncategorized error.
	KindUnknown ErrorKind = iota
	// KindFileSystem represents file/directory access errors.
	KindFileSystem
	// KindParsing represents HCL syntax errors.
	KindParsing
	// KindValidation represents semantic validation errors.
	KindValidation
	// KindFormatting represents formatting errors.
	KindFormatting
	// KindSorting represents sorting operation errors.
	KindSorting
	// KindCLI represents command-line argument errors.
	KindCLI
)

// Error is the unified error type for sortTF.
// It wraps an underlying error with context about the operation and file path.
type Error struct {
	Op   string    // operation (e.g., "ParseHCLFile", "SortBlocks")
	Path string    // file path (optional, may be empty)
	Kind ErrorKind // error category
	Err  error     // underlying error
}

// Error implements the error interface.
func (e *Error) Error() string {
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

// Unwrap returns the underlying error, enabling errors.Is and errors.As.
func (e *Error) Unwrap() error {
	return e.Err
}

// Is enables errors.Is to match sentinel errors.
func (e *Error) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// New creates a new Error with the given operation and error.
func New(op string, err error) *Error {
	return &Error{
		Op:   op,
		Kind: inferKind(err),
		Err:  err,
	}
}

// NewWithPath creates a new Error with operation, path, and error.
func NewWithPath(op, path string, err error) *Error {
	return &Error{
		Op:   op,
		Path: path,
		Kind: inferKind(err),
		Err:  err,
	}
}

// NewWithKind creates a new Error with explicit kind.
func NewWithKind(op string, kind ErrorKind, err error) *Error {
	return &Error{
		Op:   op,
		Kind: kind,
		Err:  err,
	}
}

// inferKind attempts to infer the error kind from the underlying error.
func inferKind(err error) ErrorKind {
	switch {
	case errors.Is(err, ErrFileNotFound), errors.Is(err, ErrPermissionDenied):
		return KindFileSystem
	case errors.Is(err, ErrInvalidSyntax):
		return KindParsing
	case errors.Is(err, ErrValidation):
		return KindValidation
	case errors.Is(err, ErrTerraformNotFound):
		return KindFormatting
	case os.IsNotExist(err), os.IsPermission(err):
		return KindFileSystem
	default:
		return KindUnknown
	}
}

// Color configuration for error messages
var (
	errorColor   = color.New(color.FgRed, color.Bold)
	warningColor = color.New(color.FgYellow, color.Bold)
)

// PrintError prints an error with appropriate formatting based on its kind.
func PrintError(err error, stderr io.Writer) {
	if err == nil {
		return
	}

	// Try to extract Error type
	var e *Error
	if errors.As(err, &e) {
		printErrorWithKind(e, stderr)
		return
	}

	// Handle sentinel errors
	switch {
	case errors.Is(err, ErrFileNotFound):
		_, _ = errorColor.Fprintf(stderr, "❌ File not found: %v\n", err)
	case errors.Is(err, ErrPermissionDenied):
		_, _ = errorColor.Fprintf(stderr, "🔒 Permission denied: %v\n", err)
	case errors.Is(err, ErrInvalidSyntax):
		_, _ = errorColor.Fprintf(stderr, "🔍 Syntax error: %v\n", err)
	case errors.Is(err, ErrValidation):
		_, _ = warningColor.Fprintf(stderr, "⚠️  Validation error: %v\n", err)
	case errors.Is(err, ErrTerraformNotFound):
		_, _ = errorColor.Fprintf(stderr, "❌ Terraform not found: %v\n", err)
	case errors.Is(err, ErrNoChanges):
		// No changes is not really an error, usually silent
		return
	default:
		_, _ = errorColor.Fprintf(stderr, "❌ Error: %v\n", err)
	}
}

// printErrorWithKind prints an Error with formatting based on its kind.
func printErrorWithKind(e *Error, stderr io.Writer) {
	switch e.Kind {
	case KindFileSystem:
		if errors.Is(e.Err, ErrPermissionDenied) || os.IsPermission(e.Err) {
			_, _ = errorColor.Fprintf(stderr, "🔒 Permission denied: %v\n", e)
		} else {
			_, _ = errorColor.Fprintf(stderr, "❌ File error: %v\n", e)
		}
	case KindParsing:
		_, _ = errorColor.Fprintf(stderr, "🔍 Syntax error: %v\n", e)
	case KindValidation:
		_, _ = warningColor.Fprintf(stderr, "⚠️  Validation error: %v\n", e)
	case KindFormatting:
		_, _ = errorColor.Fprintf(stderr, "🎨 Formatting error: %v\n", e)
	case KindSorting:
		_, _ = errorColor.Fprintf(stderr, "📊 Sorting error: %v\n", e)
	case KindCLI:
		_, _ = errorColor.Fprintf(stderr, "❌ CLI error: %v\n", e)
	default:
		_, _ = errorColor.Fprintf(stderr, "❌ Error: %v\n", e)
	}
}

// Wrap wraps an os error with appropriate sentinel error.
// This is useful for converting os package errors to sortTF errors.
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return fmt.Errorf("%w: %w", ErrFileNotFound, err)
	}
	if os.IsPermission(err) {
		return fmt.Errorf("%w: %w", ErrPermissionDenied, err)
	}
	return err
}
