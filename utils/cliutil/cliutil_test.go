package cliutil

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *Config
		wantErr bool
	}{
		{
			name: "no args",
			args: []string{},
			want: &Config{
				Root:      ".",
				Recursive: false,
				DryRun:    false,
				Verbose:   false,
				Validate:  false,
			},
			wantErr: false,
		},
		{
			name: "with directory",
			args: []string{"/path/to/dir"},
			want: &Config{
				Root:      "/path/to/dir",
				Recursive: false,
				DryRun:    false,
				Verbose:   false,
				Validate:  false,
			},
			wantErr: false,
		},
		{
			name: "with flags",
			args: []string{"--recursive", "--dry-run", "--verbose", "--validate", "/test/dir"},
			want: &Config{
				Root:      "/test/dir",
				Recursive: true,
				DryRun:    true,
				Verbose:   true,
				Validate:  true,
			},
			wantErr: false,
		},
		{
			name:    "too many args",
			args:    []string{"dir1", "dir2"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid flag",
			args:    []string{"--invalid-flag"},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			got, err := parseFlags(tt.args, &stderr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Root != tt.want.Root {
				t.Errorf("parseFlags() Root = %v, want %v", got.Root, tt.want.Root)
			}
			if got.Recursive != tt.want.Recursive {
				t.Errorf("parseFlags() Recursive = %v, want %v", got.Recursive, tt.want.Recursive)
			}
			if got.DryRun != tt.want.DryRun {
				t.Errorf("parseFlags() DryRun = %v, want %v", got.DryRun, tt.want.DryRun)
			}
			if got.Verbose != tt.want.Verbose {
				t.Errorf("parseFlags() Verbose = %v, want %v", got.Verbose, tt.want.Verbose)
			}
			if got.Validate != tt.want.Validate {
				t.Errorf("parseFlags() Validate = %v, want %v", got.Validate, tt.want.Validate)
			}
		})
	}
}

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

func TestRunCLIWithWriters_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{"--help"}, &stdout, &stderr)

	// Should exit with success code (help is not an error)
	if exitCode != 0 {
		t.Errorf("RunCLIWithWriters() exitCode = %v, want 0", exitCode)
	}

	// Should print usage to stderr
	output := stderr.String()
	if !strings.Contains(output, "Usage: sorttf") {
		t.Errorf("RunCLIWithWriters() stderr output = %v, should contain 'Usage: sorttf'", output)
	}
}

func TestRunCLIWithWriters_InvalidArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{"dir1", "dir2"}, &stdout, &stderr)

	// Should exit with usage error code
	if exitCode != 2 {
		t.Errorf("RunCLIWithWriters() exitCode = %v, want 2", exitCode)
	}

	// Should print error to stderr
	output := stderr.String()
	if !strings.Contains(output, "Error:") {
		t.Errorf("RunCLIWithWriters() stderr output = %v, should contain 'Error:'", output)
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
