package hcl

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/hcl/v2"
)

// TestHCLError_Error tests HCLError string formatting
func TestHCLError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *HCLError
		expected string
	}{
		{
			name: "with path and error",
			err: &HCLError{
				Op:   "TestOp",
				Path: "/test/path.tf",
				Kind: KindParsing,
				Err:  fmt.Errorf("test error"),
			},
			expected: "TestOp /test/path.tf: test error",
		},
		{
			name: "without path",
			err: &HCLError{
				Op:   "TestOp",
				Kind: KindValidation,
				Err:  fmt.Errorf("validation failed"),
			},
			expected: "TestOp: validation failed",
		},
		{
			name: "without error",
			err: &HCLError{
				Op:   "TestOp",
				Path: "/test/path.tf",
				Kind: KindFormatting,
			},
			expected: "TestOp /test/path.tf",
		},
		{
			name: "only operation",
			err: &HCLError{
				Op:   "TestOp",
				Kind: KindSorting,
			},
			expected: "TestOp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestHCLError_Unwrap tests error unwrapping
func TestHCLError_Unwrap(t *testing.T) {
	innerErr := fmt.Errorf("inner error")
	hclErr := &HCLError{
		Op:   "TestOp",
		Kind: KindParsing,
		Err:  innerErr,
	}

	unwrapped := hclErr.Unwrap()
	if !errors.Is(unwrapped, innerErr) {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, innerErr)
	}
}

// TestHCLError_Unwrap_Nil tests unwrapping nil error
func TestHCLError_Unwrap_Nil(t *testing.T) {
	hclErr := &HCLError{
		Op:   "TestOp",
		Kind: KindParsing,
		Err:  nil,
	}

	unwrapped := hclErr.Unwrap()
	if unwrapped != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrapped)
	}
}

// TestHCLParseError_Error tests HCLParseError string formatting
func TestHCLParseError_Error(t *testing.T) {
	diags := hcl.Diagnostics{
		{
			Severity: hcl.DiagError,
			Summary:  "Test error",
			Detail:   "Test detail",
		},
	}

	parseErr := &HCLParseError{
		Path:  "/test/file.tf",
		Diags: diags,
	}

	errMsg := parseErr.Error()
	if errMsg == "" {
		t.Error("Error() returned empty string")
	}
	// Should contain path and diagnostics
	if !contains(errMsg, "/test/file.tf") {
		t.Errorf("Error() should contain path, got: %s", errMsg)
	}
}

// TestIsParsingError tests parsing error detection
func TestIsParsingError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "is parsing error",
			err: &HCLError{
				Op:   "Parse",
				Kind: KindParsing,
			},
			expected: true,
		},
		{
			name: "is validation error",
			err: &HCLError{
				Op:   "Validate",
				Kind: KindValidation,
			},
			expected: false,
		},
		{
			name: "is formatting error",
			err: &HCLError{
				Op:   "Format",
				Kind: KindFormatting,
			},
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "generic error",
			err:      fmt.Errorf("generic"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsParsingError(tt.err)
			if got != tt.expected {
				t.Errorf("IsParsingError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsHCLParseError tests HCL parse error detection
func TestIsHCLParseError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "is HCLParseError",
			err: &HCLParseError{
				Path: "/test.tf",
			},
			expected: true,
		},
		{
			name: "is HCLError",
			err: &HCLError{
				Op:   "Parse",
				Kind: KindParsing,
			},
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "generic error",
			err:      fmt.Errorf("generic"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsHCLParseError(tt.err)
			if got != tt.expected {
				t.Errorf("IsHCLParseError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsValidationError tests validation error detection
func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "is validation error",
			err: &HCLError{
				Op:   "Validate",
				Kind: KindValidation,
			},
			expected: true,
		},
		{
			name: "is parsing error",
			err: &HCLError{
				Op:   "Parse",
				Kind: KindParsing,
			},
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidationError(tt.err)
			if got != tt.expected {
				t.Errorf("IsValidationError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsFormattingError tests formatting error detection
func TestIsFormattingError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "is formatting error",
			err: &HCLError{
				Op:   "Format",
				Kind: KindFormatting,
			},
			expected: true,
		},
		{
			name: "is sorting error",
			err: &HCLError{
				Op:   "Sort",
				Kind: KindSorting,
			},
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFormattingError(tt.err)
			if got != tt.expected {
				t.Errorf("IsFormattingError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsSortingError tests sorting error detection
func TestIsSortingError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "is sorting error",
			err: &HCLError{
				Op:   "Sort",
				Kind: KindSorting,
			},
			expected: true,
		},
		{
			name: "is formatting error",
			err: &HCLError{
				Op:   "Format",
				Kind: KindFormatting,
			},
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSortingError(tt.err)
			if got != tt.expected {
				t.Errorf("IsSortingError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsNotExistError tests not exist error detection
func TestIsNotExistError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "contains 'does not exist'",
			err: &HCLError{
				Op:   "Parse",
				Kind: KindParsing,
				Err:  fmt.Errorf("file does not exist"),
			},
			expected: true,
		},
		{
			name: "contains 'no such file'",
			err: &HCLError{
				Op:   "Parse",
				Kind: KindParsing,
				Err:  fmt.Errorf("no such file or directory"),
			},
			expected: true,
		},
		{
			name: "different error message",
			err: &HCLError{
				Op:   "Parse",
				Kind: KindParsing,
				Err:  fmt.Errorf("permission denied"),
			},
			expected: false,
		},
		{
			name: "nil inner error",
			err: &HCLError{
				Op:   "Parse",
				Kind: KindParsing,
				Err:  nil,
			},
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "non-HCLError",
			err:      fmt.Errorf("does not exist"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotExistError(tt.err)
			if got != tt.expected {
				t.Errorf("IsNotExistError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsPermissionError tests permission error detection
func TestIsPermissionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "contains 'permission denied'",
			err: &HCLError{
				Op:   "Parse",
				Kind: KindParsing,
				Err:  fmt.Errorf("permission denied"),
			},
			expected: true,
		},
		{
			name: "contains 'access denied'",
			err: &HCLError{
				Op:   "Parse",
				Kind: KindParsing,
				Err:  fmt.Errorf("access denied to file"),
			},
			expected: true,
		},
		{
			name: "different error message",
			err: &HCLError{
				Op:   "Parse",
				Kind: KindParsing,
				Err:  fmt.Errorf("does not exist"),
			},
			expected: false,
		},
		{
			name: "nil inner error",
			err: &HCLError{
				Op:   "Parse",
				Kind: KindParsing,
				Err:  nil,
			},
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPermissionError(tt.err)
			if got != tt.expected {
				t.Errorf("IsPermissionError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestContainsAny tests the containsAny helper function
func TestContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substrs  []string
		expected bool
	}{
		{
			name:     "contains first substring",
			s:        "hello world",
			substrs:  []string{"hello", "goodbye"},
			expected: true,
		},
		{
			name:     "contains second substring",
			s:        "hello world",
			substrs:  []string{"goodbye", "world"},
			expected: true,
		},
		{
			name:     "contains none",
			s:        "hello world",
			substrs:  []string{"foo", "bar"},
			expected: false,
		},
		{
			name:     "empty string",
			s:        "",
			substrs:  []string{"hello"},
			expected: false,
		},
		{
			name:     "empty substrings",
			s:        "hello",
			substrs:  []string{},
			expected: false,
		},
		{
			name:     "substring longer than string",
			s:        "hi",
			substrs:  []string{"hello"},
			expected: false,
		},
		{
			name:     "exact match",
			s:        "test",
			substrs:  []string{"test"},
			expected: true,
		},
		{
			name:     "partial match",
			s:        "testing",
			substrs:  []string{"test"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsAny(tt.s, tt.substrs...)
			if got != tt.expected {
				t.Errorf("containsAny(%q, %v) = %v, want %v",
					tt.s, tt.substrs, got, tt.expected)
			}
		})
	}
}

// TestErrorKind tests ErrorKind type
func TestErrorKind(t *testing.T) {
	// Verify the constants exist and are distinct
	kinds := []ErrorKind{
		KindUnknown,
		KindParsing,
		KindValidation,
		KindFormatting,
		KindSorting,
	}

	seen := make(map[ErrorKind]bool)
	for _, kind := range kinds {
		if seen[kind] {
			t.Errorf("duplicate ErrorKind value: %v", kind)
		}
		seen[kind] = true
	}
}

// TestErrorWrapping tests error wrapping with errors.Is
func TestErrorWrapping(t *testing.T) {
	innerErr := errors.New("inner error")
	hclErr := &HCLError{
		Op:   "TestOp",
		Kind: KindParsing,
		Err:  innerErr,
	}

	// errors.Is should work through Unwrap
	if !errors.Is(hclErr, innerErr) {
		t.Error("errors.Is should work with HCLError wrapping")
	}
}

// helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
