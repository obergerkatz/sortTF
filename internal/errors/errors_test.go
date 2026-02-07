package errors_test

import (
	"bytes"
	stderrors "errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/obergerkatz/sortTF/internal/errors"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
	}{
		{"ErrFileNotFound", errors.ErrFileNotFound},
		{"ErrPermissionDenied", errors.ErrPermissionDenied},
		{"ErrInvalidSyntax", errors.ErrInvalidSyntax},
		{"ErrValidation", errors.ErrValidation},
		{"ErrTerraformNotFound", errors.ErrTerraformNotFound},
		{"ErrNoChanges", errors.ErrNoChanges},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.sentinel == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.sentinel.Error() == "" {
				t.Errorf("%s should have non-empty error message", tt.name)
			}
		})
	}
}

func TestError_Creation(t *testing.T) {
	baseErr := fmt.Errorf("base error")

	tests := []struct {
		name    string
		creator func() *errors.Error
		wantOp  string
		wantErr error
	}{
		{
			name:    "New",
			creator: func() *errors.Error { return errors.New("TestOp", baseErr) },
			wantOp:  "TestOp",
			wantErr: baseErr,
		},
		{
			name:    "NewWithPath",
			creator: func() *errors.Error { return errors.NewWithPath("TestOp", "/test/path", baseErr) },
			wantOp:  "TestOp",
			wantErr: baseErr,
		},
		{
			name:    "NewWithKind",
			creator: func() *errors.Error { return errors.NewWithKind("TestOp", errors.KindParsing, baseErr) },
			wantOp:  "TestOp",
			wantErr: baseErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.creator()
			if err.Op != tt.wantOp {
				t.Errorf("Op = %q, want %q", err.Op, tt.wantOp)
			}
			if tt.wantErr != nil && !stderrors.Is(err.Err, tt.wantErr) {
				t.Errorf("Err = %v, want %v", err.Err, tt.wantErr)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *errors.Error
		want    string
		wantNot string
	}{
		{
			name: "with path and error",
			err:  errors.NewWithPath("TestOp", "/test/path", fmt.Errorf("test error")),
			want: "TestOp /test/path: test error",
		},
		{
			name: "without path",
			err:  errors.New("TestOp", fmt.Errorf("test error")),
			want: "TestOp: test error",
		},
		{
			name: "without error",
			err:  &errors.Error{Op: "TestOp", Path: "/test/path"},
			want: "TestOp /test/path",
		},
		{
			name: "only operation",
			err:  &errors.Error{Op: "TestOp"},
			want: "TestOp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	err := errors.New("TestOp", baseErr)

	unwrapped := err.Unwrap()
	if !stderrors.Is(unwrapped, baseErr) {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}

func TestError_Is(t *testing.T) {
	tests := []struct {
		name   string
		err    *errors.Error
		target error
		want   bool
	}{
		{
			name:   "matches sentinel",
			err:    errors.New("TestOp", fmt.Errorf("%w", errors.ErrFileNotFound)),
			target: errors.ErrFileNotFound,
			want:   true,
		},
		{
			name:   "doesn't match different sentinel",
			err:    errors.New("TestOp", fmt.Errorf("%w", errors.ErrFileNotFound)),
			target: errors.ErrPermissionDenied,
			want:   false,
		},
		{
			name:   "deeply wrapped sentinel",
			err:    errors.New("TestOp", fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", errors.ErrValidation))),
			target: errors.ErrValidation,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stderrors.Is(tt.err, tt.target)
			if got != tt.want {
				t.Errorf("errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_As(t *testing.T) {
	baseErr := errors.New("TestOp", fmt.Errorf("test error"))
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)

	var e *errors.Error
	if !stderrors.As(wrappedErr, &e) {
		t.Error("stderrors.As should extract Error type")
	}
	if e.Op != "TestOp" {
		t.Errorf("extracted Error.Op = %q, want %q", e.Op, "TestOp")
	}
}

func TestInferKind(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want errors.ErrorKind
	}{
		{
			name: "file not found",
			err:  fmt.Errorf("%w", errors.ErrFileNotFound),
			want: errors.KindFileSystem,
		},
		{
			name: "permission denied",
			err:  fmt.Errorf("%w", errors.ErrPermissionDenied),
			want: errors.KindFileSystem,
		},
		{
			name: "invalid syntax",
			err:  fmt.Errorf("%w", errors.ErrInvalidSyntax),
			want: errors.KindParsing,
		},
		{
			name: "validation",
			err:  fmt.Errorf("%w", errors.ErrValidation),
			want: errors.KindValidation,
		},
		{
			name: "terraform not found",
			err:  fmt.Errorf("%w", errors.ErrTerraformNotFound),
			want: errors.KindFormatting,
		},
		{
			name: "os.IsNotExist",
			err:  os.ErrNotExist,
			want: errors.KindFileSystem,
		},
		{
			name: "os.IsPermission",
			err:  os.ErrPermission,
			want: errors.KindFileSystem,
		},
		{
			name: "unknown",
			err:  fmt.Errorf("unknown error"),
			want: errors.KindUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: inferKind is not exported, so we test it indirectly through New
			e := errors.New("test", tt.err)
			got := e.Kind
			if got != tt.want {
				t.Errorf("inferKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantIs    error
		wantIsNot error
	}{
		{
			name:      "nil error",
			err:       nil,
			wantIs:    nil,
			wantIsNot: errors.ErrFileNotFound,
		},
		{
			name:      "os.ErrNotExist",
			err:       os.ErrNotExist,
			wantIs:    errors.ErrFileNotFound,
			wantIsNot: errors.ErrPermissionDenied,
		},
		{
			name:      "os.ErrPermission",
			err:       os.ErrPermission,
			wantIs:    errors.ErrPermissionDenied,
			wantIsNot: errors.ErrFileNotFound,
		},
		{
			name:      "other error",
			err:       fmt.Errorf("other"),
			wantIs:    nil,
			wantIsNot: errors.ErrFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := errors.Wrap(tt.err)
			if wrapped == nil && tt.err != nil {
				t.Error("Wrap() returned nil for non-nil error")
			}
			if tt.wantIs != nil && !stderrors.Is(wrapped, tt.wantIs) {
				t.Errorf("Wrap() should wrap with %v", tt.wantIs)
			}
			if tt.wantIsNot != nil && stderrors.Is(wrapped, tt.wantIsNot) {
				t.Errorf("Wrap() should not wrap with %v", tt.wantIsNot)
			}
		})
	}
}

func TestPrintError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantOutput string // substring that should be in output
	}{
		{
			name:       "nil error",
			err:        nil,
			wantOutput: "",
		},
		{
			name:       "ErrFileNotFound",
			err:        fmt.Errorf("%w: test.tf", errors.ErrFileNotFound),
			wantOutput: "File not found",
		},
		{
			name:       "ErrPermissionDenied",
			err:        fmt.Errorf("%w: /secure", errors.ErrPermissionDenied),
			wantOutput: "Permission denied",
		},
		{
			name:       "ErrInvalidSyntax",
			err:        fmt.Errorf("%w: missing brace", errors.ErrInvalidSyntax),
			wantOutput: "Syntax error",
		},
		{
			name:       "ErrValidation",
			err:        fmt.Errorf("%w: invalid label", errors.ErrValidation),
			wantOutput: "Validation error",
		},
		{
			name:       "Error with KindFileSystem",
			err:        errors.NewWithKind("ReadFile", errors.KindFileSystem, fmt.Errorf("disk error")),
			wantOutput: "File error",
		},
		{
			name:       "Error with KindParsing",
			err:        errors.NewWithKind("ParseHCL", errors.KindParsing, fmt.Errorf("syntax")),
			wantOutput: "Syntax error",
		},
		{
			name:       "generic error",
			err:        fmt.Errorf("unknown error"),
			wantOutput: "Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			errors.PrintError(tt.err, &buf)
			got := buf.String()

			if tt.wantOutput == "" {
				if got != "" {
					t.Errorf("PrintError() output = %q, want empty", got)
				}
				return
			}

			if !strings.Contains(got, tt.wantOutput) {
				t.Errorf("PrintError() output = %q, want to contain %q", got, tt.wantOutput)
			}
		})
	}
}

func TestError_WithWrappedSentinels(t *testing.T) {
	// Test that Error properly wraps and exposes sentinel errors
	err := errors.NewWithPath("ParseFile", "test.tf", fmt.Errorf("parsing failed: %w", errors.ErrInvalidSyntax))

	// Should be able to detect with stderrors.Is
	if !stderrors.Is(err, errors.ErrInvalidSyntax) {
		t.Error("Error should wrap ErrInvalidSyntax")
	}

	// Should extract Error type with stderrors.As
	var e *errors.Error
	if !stderrors.As(err, &e) {
		t.Error("should be able to extract Error type")
	}

	// Should have correct fields
	if e.Op != "ParseFile" {
		t.Errorf("Op = %q, want %q", e.Op, "ParseFile")
	}
	if e.Path != "test.tf" {
		t.Errorf("Path = %q, want %q", e.Path, "test.tf")
	}
}

// TestPrintError_AllErrorKinds tests PrintError with all error kinds
func TestPrintError_AllErrorKinds(t *testing.T) {
	tests := []struct {
		name       string
		err        *errors.Error
		wantOutput string
	}{
		{
			name:       "KindValidation",
			err:        errors.NewWithKind("Validate", errors.KindValidation, fmt.Errorf("invalid block")),
			wantOutput: "Validation error",
		},
		{
			name:       "KindFormatting",
			err:        errors.NewWithKind("Format", errors.KindFormatting, fmt.Errorf("format failed")),
			wantOutput: "Formatting error",
		},
		{
			name:       "KindSorting",
			err:        errors.NewWithKind("Sort", errors.KindSorting, fmt.Errorf("sort failed")),
			wantOutput: "Sorting error",
		},
		{
			name:       "KindCLI",
			err:        errors.NewWithKind("ParseArgs", errors.KindCLI, fmt.Errorf("invalid args")),
			wantOutput: "CLI error",
		},
		{
			name:       "KindFileSystem with permission",
			err:        errors.NewWithKind("ReadFile", errors.KindFileSystem, fmt.Errorf("%w", errors.ErrPermissionDenied)),
			wantOutput: "Permission denied",
		},
		{
			name:       "KindFileSystem with os.Permission",
			err:        errors.NewWithKind("ReadFile", errors.KindFileSystem, os.ErrPermission),
			wantOutput: "Permission denied",
		},
		{
			name:       "KindFileSystem other",
			err:        errors.NewWithKind("ReadFile", errors.KindFileSystem, fmt.Errorf("disk error")),
			wantOutput: "File error",
		},
		{
			name:       "KindUnknown",
			err:        errors.NewWithKind("Unknown", errors.KindUnknown, fmt.Errorf("unknown")),
			wantOutput: "Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			errors.PrintError(tt.err, &buf)
			got := buf.String()

			if !strings.Contains(got, tt.wantOutput) {
				t.Errorf("PrintError() output = %q, want to contain %q", got, tt.wantOutput)
			}
		})
	}
}

// TestPrintError_SentinelErrors tests PrintError with all sentinel errors
func TestPrintError_SentinelErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantOutput string
	}{
		{
			name:       "ErrTerraformNotFound",
			err:        fmt.Errorf("%w", errors.ErrTerraformNotFound),
			wantOutput: "Terraform not found",
		},
		{
			name:       "ErrNoChanges",
			err:        fmt.Errorf("%w", errors.ErrNoChanges),
			wantOutput: "", // ErrNoChanges prints nothing (silent)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			errors.PrintError(tt.err, &buf)
			got := buf.String()

			if tt.wantOutput == "" {
				if got != "" {
					t.Errorf("PrintError() output = %q, want empty for ErrNoChanges", got)
				}
			} else if !strings.Contains(got, tt.wantOutput) {
				t.Errorf("PrintError() output = %q, want to contain %q", got, tt.wantOutput)
			}
		})
	}
}
