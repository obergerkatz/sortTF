package errors

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
	}{
		{"ErrFileNotFound", ErrFileNotFound},
		{"ErrPermissionDenied", ErrPermissionDenied},
		{"ErrInvalidSyntax", ErrInvalidSyntax},
		{"ErrValidation", ErrValidation},
		{"ErrTerraformNotFound", ErrTerraformNotFound},
		{"ErrNoChanges", ErrNoChanges},
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
		creator func() *Error
		wantOp  string
		wantErr error
	}{
		{
			name:    "New",
			creator: func() *Error { return New("TestOp", baseErr) },
			wantOp:  "TestOp",
			wantErr: baseErr,
		},
		{
			name:    "NewWithPath",
			creator: func() *Error { return NewWithPath("TestOp", "/test/path", baseErr) },
			wantOp:  "TestOp",
			wantErr: baseErr,
		},
		{
			name:    "NewWithKind",
			creator: func() *Error { return NewWithKind("TestOp", KindParsing, baseErr) },
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
			if err.Err != tt.wantErr {
				t.Errorf("Err = %v, want %v", err.Err, tt.wantErr)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *Error
		want    string
		wantNot string
	}{
		{
			name: "with path and error",
			err:  NewWithPath("TestOp", "/test/path", fmt.Errorf("test error")),
			want: "TestOp /test/path: test error",
		},
		{
			name: "without path",
			err:  New("TestOp", fmt.Errorf("test error")),
			want: "TestOp: test error",
		},
		{
			name: "without error",
			err:  &Error{Op: "TestOp", Path: "/test/path"},
			want: "TestOp /test/path",
		},
		{
			name: "only operation",
			err:  &Error{Op: "TestOp"},
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
	err := New("TestOp", baseErr)

	unwrapped := err.Unwrap()
	if unwrapped != baseErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}

func TestError_Is(t *testing.T) {
	tests := []struct {
		name   string
		err    *Error
		target error
		want   bool
	}{
		{
			name:   "matches sentinel",
			err:    New("TestOp", fmt.Errorf("%w", ErrFileNotFound)),
			target: ErrFileNotFound,
			want:   true,
		},
		{
			name:   "doesn't match different sentinel",
			err:    New("TestOp", fmt.Errorf("%w", ErrFileNotFound)),
			target: ErrPermissionDenied,
			want:   false,
		},
		{
			name:   "deeply wrapped sentinel",
			err:    New("TestOp", fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrValidation))),
			target: ErrValidation,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errors.Is(tt.err, tt.target)
			if got != tt.want {
				t.Errorf("errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_As(t *testing.T) {
	baseErr := New("TestOp", fmt.Errorf("test error"))
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)

	var e *Error
	if !errors.As(wrappedErr, &e) {
		t.Error("errors.As should extract Error type")
	}
	if e.Op != "TestOp" {
		t.Errorf("extracted Error.Op = %q, want %q", e.Op, "TestOp")
	}
}

func TestInferKind(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorKind
	}{
		{
			name: "file not found",
			err:  fmt.Errorf("%w", ErrFileNotFound),
			want: KindFileSystem,
		},
		{
			name: "permission denied",
			err:  fmt.Errorf("%w", ErrPermissionDenied),
			want: KindFileSystem,
		},
		{
			name: "invalid syntax",
			err:  fmt.Errorf("%w", ErrInvalidSyntax),
			want: KindParsing,
		},
		{
			name: "validation",
			err:  fmt.Errorf("%w", ErrValidation),
			want: KindValidation,
		},
		{
			name: "terraform not found",
			err:  fmt.Errorf("%w", ErrTerraformNotFound),
			want: KindFormatting,
		},
		{
			name: "os.IsNotExist",
			err:  os.ErrNotExist,
			want: KindFileSystem,
		},
		{
			name: "os.IsPermission",
			err:  os.ErrPermission,
			want: KindFileSystem,
		},
		{
			name: "unknown",
			err:  fmt.Errorf("unknown error"),
			want: KindUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferKind(tt.err)
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
			wantIsNot: ErrFileNotFound,
		},
		{
			name:      "os.ErrNotExist",
			err:       os.ErrNotExist,
			wantIs:    ErrFileNotFound,
			wantIsNot: ErrPermissionDenied,
		},
		{
			name:      "os.ErrPermission",
			err:       os.ErrPermission,
			wantIs:    ErrPermissionDenied,
			wantIsNot: ErrFileNotFound,
		},
		{
			name:      "other error",
			err:       fmt.Errorf("other"),
			wantIs:    nil,
			wantIsNot: ErrFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := Wrap(tt.err)
			if wrapped == nil && tt.err != nil {
				t.Error("Wrap() returned nil for non-nil error")
			}
			if tt.wantIs != nil && !errors.Is(wrapped, tt.wantIs) {
				t.Errorf("Wrap() should wrap with %v", tt.wantIs)
			}
			if tt.wantIsNot != nil && errors.Is(wrapped, tt.wantIsNot) {
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
			err:        fmt.Errorf("%w: test.tf", ErrFileNotFound),
			wantOutput: "File not found",
		},
		{
			name:       "ErrPermissionDenied",
			err:        fmt.Errorf("%w: /secure", ErrPermissionDenied),
			wantOutput: "Permission denied",
		},
		{
			name:       "ErrInvalidSyntax",
			err:        fmt.Errorf("%w: missing brace", ErrInvalidSyntax),
			wantOutput: "Syntax error",
		},
		{
			name:       "ErrValidation",
			err:        fmt.Errorf("%w: invalid label", ErrValidation),
			wantOutput: "Validation error",
		},
		{
			name:       "Error with KindFileSystem",
			err:        NewWithKind("ReadFile", KindFileSystem, fmt.Errorf("disk error")),
			wantOutput: "File error",
		},
		{
			name:       "Error with KindParsing",
			err:        NewWithKind("ParseHCL", KindParsing, fmt.Errorf("syntax")),
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
			PrintError(tt.err, &buf)
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
	err := NewWithPath("ParseFile", "test.tf", fmt.Errorf("parsing failed: %w", ErrInvalidSyntax))

	// Should be able to detect with errors.Is
	if !errors.Is(err, ErrInvalidSyntax) {
		t.Error("Error should wrap ErrInvalidSyntax")
	}

	// Should extract Error type with errors.As
	var e *Error
	if !errors.As(err, &e) {
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
		err        *Error
		wantOutput string
	}{
		{
			name:       "KindValidation",
			err:        NewWithKind("Validate", KindValidation, fmt.Errorf("invalid block")),
			wantOutput: "Validation error",
		},
		{
			name:       "KindFormatting",
			err:        NewWithKind("Format", KindFormatting, fmt.Errorf("format failed")),
			wantOutput: "Formatting error",
		},
		{
			name:       "KindSorting",
			err:        NewWithKind("Sort", KindSorting, fmt.Errorf("sort failed")),
			wantOutput: "Sorting error",
		},
		{
			name:       "KindCLI",
			err:        NewWithKind("ParseArgs", KindCLI, fmt.Errorf("invalid args")),
			wantOutput: "CLI error",
		},
		{
			name:       "KindFileSystem with permission",
			err:        NewWithKind("ReadFile", KindFileSystem, fmt.Errorf("%w", ErrPermissionDenied)),
			wantOutput: "Permission denied",
		},
		{
			name:       "KindFileSystem with os.Permission",
			err:        NewWithKind("ReadFile", KindFileSystem, os.ErrPermission),
			wantOutput: "Permission denied",
		},
		{
			name:       "KindFileSystem other",
			err:        NewWithKind("ReadFile", KindFileSystem, fmt.Errorf("disk error")),
			wantOutput: "File error",
		},
		{
			name:       "KindUnknown",
			err:        NewWithKind("Unknown", KindUnknown, fmt.Errorf("unknown")),
			wantOutput: "Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintError(tt.err, &buf)
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
			err:        fmt.Errorf("%w", ErrTerraformNotFound),
			wantOutput: "Terraform not found",
		},
		{
			name:       "ErrNoChanges",
			err:        fmt.Errorf("%w", ErrNoChanges),
			wantOutput: "", // ErrNoChanges prints nothing (silent)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintError(tt.err, &buf)
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
