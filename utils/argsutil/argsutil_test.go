package argsutil

import (
	"bytes"
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
			got, err := ParseFlags(tt.args, &stderr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Root != tt.want.Root {
				t.Errorf("ParseFlags() Root = %v, want %v", got.Root, tt.want.Root)
			}
			if got.Recursive != tt.want.Recursive {
				t.Errorf("ParseFlags() Recursive = %v, want %v", got.Recursive, tt.want.Recursive)
			}
			if got.DryRun != tt.want.DryRun {
				t.Errorf("ParseFlags() DryRun = %v, want %v", got.DryRun, tt.want.DryRun)
			}
			if got.Verbose != tt.want.Verbose {
				t.Errorf("ParseFlags() Verbose = %v, want %v", got.Verbose, tt.want.Verbose)
			}
			if got.Validate != tt.want.Validate {
				t.Errorf("ParseFlags() Validate = %v, want %v", got.Validate, tt.want.Validate)
			}
		})
	}
}

// ... tests will be added in the next step ... 