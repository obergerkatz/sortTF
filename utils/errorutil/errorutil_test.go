package errorutil

import (
	"fmt"
	"testing"
)

// ... tests will be added in the next step ... 

func TestCLIError(t *testing.T) {
	originalErr := &CLIError{
		Op:  "test",
		Err: nil,
	}

	if originalErr.Error() != "cliutil test" {
		t.Errorf("CLIError.Error() = %v, want 'cliutil test'", originalErr.Error())
	}

	wrappedErr := &CLIError{
		Op:  "test",
		Err: originalErr,
	}

	if wrappedErr.Error() != "cliutil test: cliutil test" {
		t.Errorf("CLIError.Error() with wrapped error = %v, want 'cliutil test: cliutil test'", wrappedErr.Error())
	}

	if wrappedErr.Unwrap() != originalErr {
		t.Errorf("CLIError.Unwrap() = %v, want %v", wrappedErr.Unwrap(), originalErr)
	}
}

func TestErrorHelpers(t *testing.T) {
	cliErr := &CLIError{
		Op:  "test",
		Err: nil,
	}

	if !IsCLIError(cliErr) {
		t.Error("IsCLIError() should return true for CLIError")
	}

	if IsCLIError(nil) {
		t.Error("IsCLIError() should return false for nil")
	}

	// Test with a regular error
	regularErr := fmt.Errorf("regular error")
	if IsCLIError(regularErr) {
		t.Error("IsCLIError() should return false for non-CLIError")
	}

	if GetCLIErrorOp(cliErr) != "test" {
		t.Errorf("GetCLIErrorOp() = %v, want 'test'", GetCLIErrorOp(cliErr))
	}

	if GetCLIErrorOp(nil) != "" {
		t.Errorf("GetCLIErrorOp() for nil = %v, want ''", GetCLIErrorOp(nil))
	}
} 