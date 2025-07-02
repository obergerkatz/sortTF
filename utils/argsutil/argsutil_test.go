package argsutil

import (
	"bytes"
	"strings"
	"testing"
)

func assertConfigEqual(t *testing.T, got, want *Config) {
	t.Helper()
	if got == nil && want == nil {
		return
	}
	if got == nil || want == nil {
		t.Fatalf("Config: got %v, want %v", got, want)
	}
	if got.Root != want.Root ||
		got.Recursive != want.Recursive ||
		got.DryRun != want.DryRun ||
		got.Verbose != want.Verbose ||
		got.Validate != want.Validate {
		t.Errorf("Config: got %+v, want %+v", got, want)
	}
}

func TestParseFlags_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "no args",
			args: nil,
			want: &Config{Root: ".", Recursive: false, DryRun: false, Verbose: false, Validate: false},
		},
		{
			name: "with directory",
			args: []string{"/path/to/dir"},
			want: &Config{Root: "/path/to/dir", Recursive: false, DryRun: false, Verbose: false, Validate: false},
		},
		{
			name: "with flags",
			args: []string{"--recursive", "--dry-run", "--verbose", "--validate", "/test/dir"},
			want: &Config{Root: "/test/dir", Recursive: true, DryRun: true, Verbose: true, Validate: true},
		},
		{
			name:    "too many args",
			args:    []string{"dir1", "dir2"},
			wantErr: true,
			errMsg:  "too many arguments",
		},
		{
			name:    "invalid flag",
			args:    []string{"--invalid-flag"},
			wantErr: true,
			errMsg:  "flag provided but not defined",
		},
		{
			name:    "help flag",
			args:    []string{"--help"},
			wantErr: true,
			errMsg:  "help",
		},
		{
			name: "flags in different order",
			args: []string{"--dry-run", "--verbose", "/foo"},
			want: &Config{Root: "/foo", Recursive: false, DryRun: true, Verbose: true, Validate: false},
		},
		{
			name: "path with spaces",
			args: []string{"my dir/file.tf"},
			want: &Config{Root: "my dir/file.tf", Recursive: false, DryRun: false, Verbose: false, Validate: false},
		},
		{
			name: "empty string arg",
			args: []string{""},
			want: &Config{Root: "", Recursive: false, DryRun: false, Verbose: false, Validate: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			got, err := ParseFlags(tt.args, &stderr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.errMsg != "" && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ParseFlags() error = %v, want error containing %q", err, tt.errMsg)
				}
				// Only check stderr for help and invalid flag cases
				if (tt.errMsg == "help" || tt.errMsg == "flag provided but not defined") && stderr.Len() == 0 {
					t.Error("Expected error message in stderr")
				}
				return
			}
			assertConfigEqual(t, got, tt.want)
		})
	}
}

func TestParseFlags_StderrUsage(t *testing.T) {
	var stderr bytes.Buffer
	_, err := ParseFlags([]string{"--help"}, &stderr)
	if err == nil || !strings.Contains(err.Error(), "help") {
		t.Error("Expected help error")
	}
	if !strings.Contains(stderr.String(), "Usage: sorttf") {
		t.Error("Expected usage message in stderr")
	}
}

// ... tests will be added in the next step ... 