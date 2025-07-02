package errorutil

import (
	"errors"
	"testing"
)

// Helper for error message assertion
func assertErrorMessage(t *testing.T, err error, want string) {
	t.Helper()
	got := err.Error()
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestCLIError_Behavior(t *testing.T) {
	tests := []struct {
		name string
		err  *CLIError
		want string
	}{
		{"no wrapped error", &CLIError{Op: "op", Err: nil}, "cliutil op"},
		{"with wrapped error", &CLIError{Op: "op", Err: errors.New("fail")}, "cliutil op: fail"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrorMessage(t, tt.err, tt.want)
		})
	}
}

func TestCLIError_Unwrap(t *testing.T) {
	inner := errors.New("inner")
	cliErr := &CLIError{Op: "op", Err: inner}
	if !errors.Is(cliErr, inner) {
		t.Error("errors.Is should recognize wrapped error")
	}
}

func TestNoChangesError(t *testing.T) {
	err := &NoChangesError{FilePath: "foo.tf"}
	assertErrorMessage(t, err, "no changes needed for foo.tf")
}

func TestNewCLIErrorAndNewNoChangesError(t *testing.T) {
	cli := NewCLIError("op", errors.New("fail"))
	assertErrorMessage(t, cli, "cliutil op: fail")
	nc := NewNoChangesError("bar.tf")
	assertErrorMessage(t, nc, "no changes needed for bar.tf")
}

func TestIsCLIError(t *testing.T) {
	cli := &CLIError{Op: "op"}
	if !IsCLIError(cli) {
		t.Error("IsCLIError should return true for CLIError")
	}
	if IsCLIError(nil) {
		t.Error("IsCLIError should return false for nil")
	}
	if IsCLIError(errors.New("not cli")) {
		t.Error("IsCLIError should return false for non-CLIError")
	}
}

func TestGetCLIErrorOp(t *testing.T) {
	cli := &CLIError{Op: "op"}
	if GetCLIErrorOp(cli) != "op" {
		t.Errorf("GetCLIErrorOp() = %q, want %q", GetCLIErrorOp(cli), "op")
	}
	if GetCLIErrorOp(nil) != "" {
		t.Errorf("GetCLIErrorOp(nil) = %q, want \"\"", GetCLIErrorOp(nil))
	}
}

func TestErrorTypeDetectionHelpers(t *testing.T) {
	cases := []struct {
		fn   func(error) bool
		err  error
		want bool
	}{
		{IsFileNotFoundError, errors.New("file does not exist"), true},
		{IsFileNotFoundError, errors.New("file not found"), true},
		{IsFileNotFoundError, errors.New("other error"), false},
		{IsPermissionError, errors.New("permission denied"), true},
		{IsPermissionError, errors.New("Permission denied"), true},
		{IsPermissionError, errors.New("other error"), false},
		{IsValidationError, errors.New("validation error"), true},
		{IsValidationError, errors.New("validation failed"), true},
		{IsValidationError, errors.New("other error"), false},
		{IsParsingError, errors.New("syntax error"), true},
		{IsParsingError, errors.New("parsing error"), true},
		{IsParsingError, errors.New("other error"), false},
		{IsFormattingError, errors.New("formatting error"), true},
		{IsFormattingError, errors.New("other error"), false},
		{IsSortingError, errors.New("sorting error"), true},
		{IsSortingError, errors.New("other error"), false},
	}
	for _, c := range cases {
		if got := c.fn(c.err); got != c.want {
			t.Errorf("%T(%q) = %v, want %v", c.fn, c.err, got, c.want)
		}
	}
}

func TestExtractFilePathAndErrorMessage(t *testing.T) {
	err := errors.New("failed to read file: foo.tf")
	if got := ExtractFilePath(err); got != "foo.tf" {
		t.Errorf("ExtractFilePath() = %q, want %q", got, "foo.tf")
	}
	err2 := errors.New("validation error in bar.tf: something wrong")
	if got := ExtractFilePath(err2); got != "bar.tf: something wrong" {
		t.Errorf("ExtractFilePath() = %q, want %q", got, "bar.tf: something wrong")
	}
	err3 := errors.New("some random error")
	if got := ExtractFilePath(err3); got != "" {
		t.Errorf("ExtractFilePath() = %q, want \"\"", got)
	}

	// Error message extraction
	err4 := errors.New("validation error in bar.tf: something wrong")
	if got := ExtractErrorMessage(err4); got != "bar.tf: something wrong" {
		t.Errorf("ExtractErrorMessage() = %q, want %q", got, "bar.tf: something wrong")
	}
	err5 := errors.New("other error")
	if got := ExtractErrorMessage(err5); got != "other error" {
		t.Errorf("ExtractErrorMessage() = %q, want %q", got, "other error")
	}
} 