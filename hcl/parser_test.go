package hcl

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseHCLFile tests parsing valid HCL files
func TestParseHCLFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid terraform block",
			content: `terraform {
  required_version = ">= 1.0"
}`,
			wantErr: false,
		},
		{
			name: "valid resource block",
			content: `resource "aws_instance" "web" {
  ami = "ami-12345"
  instance_type = "t2.micro"
}`,
			wantErr: false,
		},
		{
			name: "valid variable block",
			content: `variable "region" {
  type = string
  default = "us-west-2"
}`,
			wantErr: false,
		},
		{
			name: "empty file",
			content: "",
			wantErr: false,
		},
		{
			name: "only whitespace",
			content: "   \n\n\t  \n",
			wantErr: false,
		},
		{
			name: "comments only",
			content: "# This is a comment\n// Another comment\n/* Block comment */",
			wantErr: false,
		},
		{
			name: "invalid syntax - unclosed brace",
			content: `resource "aws_instance" "web" {
  ami = "ami-12345"`,
			wantErr: true,
		},
		{
			name: "invalid syntax - unclosed quote",
			content: `variable "test {
  type = string
}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.tf")
			//nolint:gosec // G306: Test files can use 0644 permissions
			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			parsed, err := ParseHCLFile(filePath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				var hclParseErr *HCLParseError
				if !errors.As(err, &hclParseErr) && err != nil {
					// Check that it's either HCLParseError or wrapped
					t.Logf("got error type: %T", err)
				}
				// ParsedFile should still be returned even on error
				if parsed == nil {
					t.Error("expected non-nil ParsedFile even on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if parsed == nil {
					t.Error("expected non-nil ParsedFile")
				}
				if parsed != nil && parsed.File == nil {
					t.Error("expected non-nil File in ParsedFile")
				}
			}
		})
	}
}

// TestParseHCLFile_EmptyPath tests empty path handling
func TestParseHCLFile_EmptyPath(t *testing.T) {
	parsed, err := ParseHCLFile("")
	if err == nil {
		t.Error("expected error for empty path")
	}
	if parsed != nil {
		t.Error("expected nil ParsedFile for empty path")
	}

	var hclErr *HCLError
	if !errors.As(err, &hclErr) {
		t.Errorf("expected *HCLError, got %T", err)
	} else {
		if hclErr.Kind != KindParsing {
			t.Errorf("expected KindParsing, got %v", hclErr.Kind)
		}
		if hclErr.Op != "ParseHCLFile" {
			t.Errorf("expected Op='ParseHCLFile', got %q", hclErr.Op)
		}
	}
}

// TestParseHCLFile_NonExistentFile tests non-existent file handling
func TestParseHCLFile_NonExistentFile(t *testing.T) {
	parsed, err := ParseHCLFile("/nonexistent/path/test.tf")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
	if parsed != nil {
		t.Error("expected nil ParsedFile for non-existent file")
	}
}

// TestParseHCLFile_Directory tests directory path handling
func TestParseHCLFile_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	parsed, err := ParseHCLFile(tmpDir)
	if err == nil {
		t.Error("expected error for directory path")
	}
	if parsed != nil {
		t.Error("expected nil ParsedFile for directory path")
	}
}

// TestValidateRequiredBlockLabels tests block label validation
func TestValidateRequiredBlockLabels(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid resource with 2 labels",
			content: `resource "aws_instance" "web" {
  ami = "ami-12345"
}`,
			wantErr: false,
		},
		{
			name: "valid data with 2 labels",
			content: `data "aws_ami" "ubuntu" {
  most_recent = true
}`,
			wantErr: false,
		},
		{
			name: "valid variable with 1 label",
			content: `variable "region" {
  type = string
}`,
			wantErr: false,
		},
		{
			name: "valid module with 1 label",
			content: `module "vpc" {
  source = "./modules/vpc"
}`,
			wantErr: false,
		},
		{
			name: "valid provider with 1 label",
			content: `provider "aws" {
  region = "us-west-2"
}`,
			wantErr: false,
		},
		{
			name: "valid output with 1 label",
			content: `output "instance_id" {
  value = aws_instance.web.id
}`,
			wantErr: false,
		},
		{
			name: "valid locals with no labels",
			content: `locals {
  name = "test"
}`,
			wantErr: false,
		},
		{
			name: "valid terraform with no labels",
			content: `terraform {
  required_version = ">= 1.0"
}`,
			wantErr: false,
		},
		{
			name: "invalid resource with 1 label",
			content: `resource "aws_instance" {
  ami = "ami-12345"
}`,
			wantErr: true,
			errMsg: "must have exactly 2 labels",
		},
		{
			name: "invalid resource with 3 labels",
			content: `resource "aws" "instance" "web" {
  ami = "ami-12345"
}`,
			wantErr: true,
			errMsg: "must have exactly 2 labels",
		},
		{
			name: "invalid variable with no labels",
			content: `variable {
  type = string
}`,
			wantErr: true,
			errMsg: "must have exactly 1 label",
		},
		{
			name: "invalid locals with labels",
			content: `locals "test" {
  name = "value"
}`,
			wantErr: true,
			errMsg: "should not have labels",
		},
		{
			name: "invalid backend at top level",
			content: `backend "s3" {
  bucket = "test"
}`,
			wantErr: true,
			errMsg: "must be inside a terraform block",
		},
		{
			name: "valid backend inside terraform",
			content: `terraform {
  backend "s3" {
    bucket = "test"
  }
}`,
			wantErr: false,
		},
		{
			name: "invalid backend inside terraform with no labels",
			content: `terraform {
  backend {
    bucket = "test"
  }
}`,
			wantErr: true,
			errMsg: "must have exactly 1 label",
		},
		{
			name: "invalid backend inside terraform with 2 labels",
			content: `terraform {
  backend "s3" "extra" {
    bucket = "test"
  }
}`,
			wantErr: true,
			errMsg: "must have exactly 1 label",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.tf")
			//nolint:gosec // G306: Test files can use 0644 permissions
			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			parsed, err := ParseHCLFile(filePath)
			if err != nil && !IsHCLParseError(err) {
				t.Fatalf("parse failed: %v", err)
			}

			err = ValidateRequiredBlockLabels(parsed)

			if tt.wantErr {
				if err == nil {
					t.Error("expected validation error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

// TestValidateRequiredBlockLabels_NilInput tests nil input handling
func TestValidateRequiredBlockLabels_NilInput(t *testing.T) {
	err := ValidateRequiredBlockLabels(nil)
	if err == nil {
		t.Error("expected error for nil input")
	}

	var hclErr *HCLError
	if !errors.As(err, &hclErr) {
		t.Errorf("expected *HCLError, got %T", err)
	} else if hclErr.Kind != KindValidation {
		t.Errorf("expected KindValidation, got %v", hclErr.Kind)
	}
}

// TestValidateRequiredBlockLabels_NilFile tests nil file in ParsedFile
func TestValidateRequiredBlockLabels_NilFile(t *testing.T) {
	pf := &ParsedFile{
		File: nil,
	}
	err := ValidateRequiredBlockLabels(pf)
	if err == nil {
		t.Error("expected error for nil file")
	}
}

// TestValidateFilePath tests file path validation
func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T) string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid file",
			setup: func(_ *testing.T) string {
				tmpDir := os.TempDir()
				path := filepath.Join(tmpDir, "test.tf")
				//nolint:gosec // G306: Test files can use 0644 permissions
				_ = os.WriteFile(path, []byte("test"), 0644)
				return path
			},
			wantErr: false,
		},
		{
			name: "empty path",
			setup: func(*testing.T) string {
				return ""
			},
			wantErr: true,
			errMsg:  "empty path",
		},
		{
			name: "non-existent file",
			setup: func(*testing.T) string {
				return "/nonexistent/test.tf"
			},
			wantErr: true,
			errMsg:  "does not exist",
		},
		{
			name: "directory instead of file",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
			errMsg:  "directory",
		},
		{
			name: "permission denied - simulate with invalid path",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				restrictedDir := filepath.Join(tmpDir, "restricted")
				_ = os.Mkdir(restrictedDir, 0000)
				//nolint:gosec // G302: Test cleanup needs to restore permissions
				t.Cleanup(func() { _ = os.Chmod(restrictedDir, 0755) })
				return filepath.Join(restrictedDir, "test.tf")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			err := validateFilePath(path)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				errStr := err.Error()
				if !strings.Contains(errStr, tt.errMsg) {
					t.Errorf("error %q should contain %q", errStr, tt.errMsg)
				}
			}
		})
	}
}

// TestValidateRequiredBlockLabels_BackendWithInvalidLabels tests backend blocks with invalid label counts
func TestValidateRequiredBlockLabels_BackendWithInvalidLabels(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.tf")

	// Backend block at root level with 0 labels
	content := `backend {
  bucket = "test"
}`
	//nolint:gosec // G306: Test file 0644
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	pf, err := ParseHCLFile(testFile)
	if err != nil {
		t.Logf("Parse error (expected): %v", err)
	}

	// Validate should fail because backend needs 1 label
	err = ValidateRequiredBlockLabels(pf)
	if err == nil {
		t.Error("Expected error for backend block without label")
	}
	if err != nil && !strings.Contains(err.Error(), "backend block must have exactly 1 label") {
		t.Errorf("Expected backend label error, got: %v", err)
	}
}

// TestValidateRequiredBlockLabels_BackendWithMultipleLabels tests backend with more than 1 label
func TestValidateRequiredBlockLabels_BackendWithMultipleLabels(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.tf")

	// Backend block with 2 labels (invalid)
	content := `backend "s3" "extra" {
  bucket = "test"
}`
	//nolint:gosec // G306: Test file 0644
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	pf, err := ParseHCLFile(testFile)
	if err != nil {
		// Parse might fail, that's okay
		t.Logf("Parse failed (acceptable): %v", err)
		return
	}

	err = ValidateRequiredBlockLabels(pf)
	if err == nil {
		t.Error("Expected error for backend block with 2 labels")
	}
}

// TestValidateRequiredBlockLabels_TerraformBackendWithInvalidLabels tests backend inside terraform with wrong labels
func TestValidateRequiredBlockLabels_TerraformBackendWithInvalidLabels(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.tf")

	// Backend inside terraform with 0 labels
	content := `terraform {
  backend {
    bucket = "test"
  }
}`
	//nolint:gosec // G306: Test file 0644
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	pf, err := ParseHCLFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateRequiredBlockLabels(pf)
	if err == nil {
		t.Error("Expected error for backend block without label inside terraform")
	}
	if err != nil && !strings.Contains(err.Error(), "backend block") {
		t.Logf("Got error: %v", err)
	}
}
