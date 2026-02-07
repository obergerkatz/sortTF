package hcl

import (
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// TestFormatHCLFile tests formatting of HCL files
func TestFormatHCLFile(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid simple block",
			input: `resource "aws_instance" "web" {
  ami = "ami-12345"
  instance_type = "t2.micro"
}`,
			wantErr: false,
		},
		{
			name: "unformatted input gets formatted",
			input: `resource "aws_instance" "web" {
ami="ami-12345"
instance_type="t2.micro"
}`,
			wantErr: false,
		},
		{
			name:    "empty content",
			input:   "",
			wantErr: false,
		},
		{
			name: "multiple blocks",
			input: `variable "test" {
  type = string
}

resource "aws_instance" "web" {
  ami = "ami-12345"
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, diags := hclwrite.ParseConfig([]byte(tt.input), "test.tf", hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				t.Fatalf("parse failed: %v", diags)
			}

			formatted, err := FormatHCLFile(file)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Formatted output should be valid HCL
				if formatted != "" {
					_, diags := hclwrite.ParseConfig([]byte(formatted), "test.tf", hcl.Pos{Line: 1, Column: 1})
					if diags.HasErrors() {
						t.Errorf("formatted output is invalid HCL: %v", diags)
					}
				}
			}
		})
	}
}

// TestFormatHCLFile_NilInput tests nil file input
func TestFormatHCLFile_NilInput(t *testing.T) {
	formatted, err := FormatHCLFile(nil)
	if err == nil {
		t.Error("expected error for nil input")
	}
	if formatted != "" {
		t.Errorf("expected empty string for nil input, got %q", formatted)
	}

	var hclErr *HCLError
	if !errors.As(err, &hclErr) {
		t.Errorf("expected *HCLError, got %T", err)
	} else {
		if hclErr.Kind != KindFormatting {
			t.Errorf("expected KindFormatting, got %v", hclErr.Kind)
		}
		if hclErr.Op != "FormatHCLFile" {
			t.Errorf("expected Op='FormatHCLFile', got %q", hclErr.Op)
		}
	}
}

// TestFormatHCLFile_Idempotent tests that formatting is idempotent
func TestFormatHCLFile_Idempotent(t *testing.T) {
	input := `variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

resource "aws_instance" "web" {
  ami           = "ami-12345"
  instance_type = "t2.micro"
}
`

	file, diags := hclwrite.ParseConfig([]byte(input), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse failed: %v", diags)
	}

	// Format once
	formatted1, err := FormatHCLFile(file)
	if err != nil {
		t.Fatalf("first format failed: %v", err)
	}

	// Parse and format again
	file2, diags := hclwrite.ParseConfig([]byte(formatted1), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse formatted failed: %v", diags)
	}

	formatted2, err := FormatHCLFile(file2)
	if err != nil {
		t.Fatalf("second format failed: %v", err)
	}

	// Both should be identical
	if formatted1 != formatted2 {
		t.Error("formatting is not idempotent")
		t.Logf("First:\n%s", formatted1)
		t.Logf("Second:\n%s", formatted2)
	}
}

// TestFormatHCLString tests string formatting
func TestFormatHCLString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid input",
			input: `variable "test" {
  type = string
}`,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
		},
		{
			name:    "whitespace only",
			input:   "   \n\t  ",
			wantErr: false,
		},
		{
			name:    "invalid syntax",
			input:   `variable "test" { unclosed`,
			wantErr: true,
		},
		{
			name: "comments preserved",
			input: `# Comment
variable "test" {
  type = string
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted, err := FormatHCLString(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				// On error, should return original content
				if formatted != tt.input {
					t.Error("expected original content on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Empty input should return empty output
				if tt.input == "" && formatted != "" {
					t.Error("expected empty output for empty input")
				}
			}
		})
	}
}

// TestFormatHCLString_ComplexStructures tests formatting of complex HCL
func TestFormatHCLString_ComplexStructures(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "nested blocks",
			input: `terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}`,
		},
		{
			name: "lists and maps",
			input: `variable "config" {
  type = map(object({
    name = string
    tags = list(string)
  }))
}`,
		},
		{
			name: "heredoc",
			input: `locals {
  script = <<-EOF
    #!/bin/bash
    echo "test"
  EOF
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted, err := FormatHCLString(tt.input)
			if err != nil {
				t.Fatalf("format failed: %v", err)
			}

			// Should be valid HCL
			_, diags := hclwrite.ParseConfig([]byte(formatted), "test.tf", hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				t.Errorf("formatted output is invalid: %v", diags)
			}

			// Should not be empty
			if strings.TrimSpace(formatted) == "" {
				t.Error("formatted output is empty")
			}
		})
	}
}
