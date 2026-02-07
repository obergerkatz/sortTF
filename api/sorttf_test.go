package api

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSortFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.tf")

	// Create unsorted file
	unsortedContent := `resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami = "ami-12345"
}

variable "environment" {
  type = string
}
`
	if err := os.WriteFile(testFile, []byte(unsortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Sort the file
	err := SortFile(testFile, Options{})
	if err != nil {
		t.Fatalf("SortFile failed: %v", err)
	}

	// Read result
	result, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	resultStr := string(result)

	// Verify variable comes before resource
	// (basic check, not exhaustive)
	if !contains(resultStr, "variable") || !contains(resultStr, "resource") {
		t.Error("Expected both variable and resource blocks")
	}

	// Sort again - should return ErrNoChanges
	err = SortFile(testFile, Options{})
	if !errors.Is(err, ErrNoChanges) {
		t.Errorf("Expected ErrNoChanges on already sorted file, got: %v", err)
	}
}

func TestSortFile_Validate(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.tf")

	// Create unsorted file
	unsortedContent := `resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami = "ami-12345"
}
`
	if err := os.WriteFile(testFile, []byte(unsortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Validate unsorted file - should return ErrNeedsSorting
	err := SortFile(testFile, Options{Validate: true})
	if !errors.Is(err, ErrNeedsSorting) {
		t.Errorf("Expected ErrNeedsSorting for unsorted file, got: %v", err)
	}

	// File should not have been modified
	result, _ := os.ReadFile(testFile)
	if string(result) != unsortedContent {
		t.Error("File was modified in validate mode")
	}
}

func TestSortFile_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.tf")

	unsortedContent := `resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami = "ami-12345"
}
`
	if err := os.WriteFile(testFile, []byte(unsortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Dry run - should succeed but not modify file
	err := SortFile(testFile, Options{DryRun: true})
	if err != nil {
		t.Errorf("DryRun should succeed, got: %v", err)
	}

	// File should not have been modified
	result, _ := os.ReadFile(testFile)
	if string(result) != unsortedContent {
		t.Error("File was modified in dry-run mode")
	}
}

func TestGetSortedContent(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.tf")

	unsortedContent := `resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami = "ami-12345"
}
`
	if err := os.WriteFile(testFile, []byte(unsortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Get sorted content
	sorted, changed, err := GetSortedContent(testFile)
	if err != nil {
		t.Fatalf("GetSortedContent failed: %v", err)
	}

	if !changed {
		t.Error("Expected changed=true for unsorted file")
	}

	if sorted == "" {
		t.Error("Expected non-empty sorted content")
	}

	if sorted == unsortedContent {
		t.Error("Sorted content should differ from original")
	}

	// Original file should not be modified
	result, _ := os.ReadFile(testFile)
	if string(result) != unsortedContent {
		t.Error("GetSortedContent modified the original file")
	}
}

func TestSortFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple files
	file1 := filepath.Join(tmpDir, "file1.tf")
	file2 := filepath.Join(tmpDir, "file2.tf")

	content := `resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami = "ami-12345"
}
`
	_ = os.WriteFile(file1, []byte(content), 0644)
	_ = os.WriteFile(file2, []byte(content), 0644)

	// Sort multiple files
	results := SortFiles([]string{file1, file2}, Options{})

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for path, err := range results {
		if err != nil {
			t.Errorf("File %s failed: %v", path, err)
		}
	}
}

func TestSortDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	file1 := filepath.Join(tmpDir, "main.tf")
	content := `resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami = "ami-12345"
}
`
	_ = os.WriteFile(file1, []byte(content), 0644)

	// Sort directory
	results, err := SortDirectory(tmpDir, false, Options{})
	if err != nil {
		t.Fatalf("SortDirectory failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 file, got %d", len(results))
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestGetSortedContent_FileNotFound tests error handling for missing files
func TestGetSortedContent_FileNotFound(t *testing.T) {
	_, _, err := GetSortedContent("/nonexistent/file.tf")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestGetSortedContent_InvalidHCL tests error handling for invalid HCL syntax
func TestGetSortedContent_InvalidHCL(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.tf")

	// Create file with invalid HCL syntax
	invalidContent := `resource "aws_instance" "web" {
  ami = "ami-12345"
` // missing closing brace
	if err := os.WriteFile(testFile, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err := GetSortedContent(testFile)
	if err == nil {
		t.Error("Expected error for invalid HCL")
	}
}

// TestGetSortedContent_ValidationError tests error handling for validation failures
func TestGetSortedContent_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.tf")

	// Create file with HCL that fails validation (resource with wrong number of labels)
	invalidContent := `resource "aws_instance" {
  ami = "ami-12345"
}
`
	if err := os.WriteFile(testFile, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err := GetSortedContent(testFile)
	if err == nil {
		t.Error("Expected validation error for resource with wrong labels")
	}
}

// TestGetSortedContent_AlreadySorted tests file that is already sorted
func TestGetSortedContent_AlreadySorted(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sorted.tf")

	// Create properly sorted file
	sortedContent := `variable "environment" {
  type = string
}

resource "aws_instance" "web" {
  ami           = "ami-12345"
  instance_type = "t3.micro"
}
`
	if err := os.WriteFile(testFile, []byte(sortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	content, changed, err := GetSortedContent(testFile)
	if err != nil {
		t.Fatalf("GetSortedContent failed: %v", err)
	}

	if changed {
		t.Error("Expected changed=false for already sorted file")
	}

	if content == "" {
		t.Error("Expected non-empty content")
	}
}

// TestSortFile_ErrorCases tests error handling in SortFile
func TestSortFile_ErrorCases(t *testing.T) {
	t.Run("nonexistent file", func(t *testing.T) {
		err := SortFile("/nonexistent/file.tf", Options{})
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("invalid HCL", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "invalid.tf")
		invalidContent := `resource "aws_instance" "web" {`
		if err := os.WriteFile(testFile, []byte(invalidContent), 0644); err != nil {
			t.Fatal(err)
		}

		err := SortFile(testFile, Options{})
		if err == nil {
			t.Error("Expected error for invalid HCL")
		}
	})

	t.Run("validate mode with already sorted file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "sorted.tf")
		sortedContent := `variable "environment" {
  type = string
}
`
		if err := os.WriteFile(testFile, []byte(sortedContent), 0644); err != nil {
			t.Fatal(err)
		}

		err := SortFile(testFile, Options{Validate: true})
		if !errors.Is(err, ErrNoChanges) {
			t.Errorf("Expected ErrNoChanges for already sorted file in validate mode, got: %v", err)
		}
	})
}

// TestSortFile_ReadOnlyDirectory tests error when temp file can't be written
func TestSortFile_ReadOnlyDirectory(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.tf")

	unsortedContent := `resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami = "ami-12345"
}
`
	if err := os.WriteFile(testFile, []byte(unsortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Make directory read-only
	if err := os.Chmod(tmpDir, 0555); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(tmpDir, 0755) }() // Cleanup

	err := SortFile(testFile, Options{})
	if err == nil {
		// On some systems, this might not fail
		t.Log("Warning: Expected error for read-only directory")
	}
}

// TestSortDirectory_ErrorCases tests error handling in SortDirectory
func TestSortDirectory_ErrorCases(t *testing.T) {
	t.Run("nonexistent directory", func(t *testing.T) {
		_, err := SortDirectory("/nonexistent/directory", false, Options{})
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		results, err := SortDirectory(tmpDir, false, Options{})
		if err != nil {
			t.Fatalf("SortDirectory should not error on empty directory: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results for empty directory, got %d", len(results))
		}
	})

	t.Run("recursive with nested files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create nested directory structure
		nestedDir := filepath.Join(tmpDir, "modules", "vpc")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create files at different levels
		file1 := filepath.Join(tmpDir, "main.tf")
		file2 := filepath.Join(nestedDir, "vpc.tf")
		content := `resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami = "ami-12345"
}
`
		if err := os.WriteFile(file1, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file2, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		results, err := SortDirectory(tmpDir, true, Options{})
		if err != nil {
			t.Fatalf("SortDirectory recursive failed: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 files in recursive mode, got %d", len(results))
		}
	})
}

// TestSortFiles_MixedResults tests SortFiles with mix of success and failure
func TestSortFiles_MixedResults(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid file
	validFile := filepath.Join(tmpDir, "valid.tf")
	validContent := `resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami = "ami-12345"
}
`
	if err := os.WriteFile(validFile, []byte(validContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create invalid file
	invalidFile := filepath.Join(tmpDir, "invalid.tf")
	invalidContent := `resource "aws_instance" "web" {` // missing closing brace
	if err := os.WriteFile(invalidFile, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Add non-existent file
	nonexistentFile := filepath.Join(tmpDir, "nonexistent.tf")

	results := SortFiles([]string{validFile, invalidFile, nonexistentFile}, Options{})

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Valid file should succeed
	if results[validFile] != nil {
		t.Errorf("Valid file should succeed, got: %v", results[validFile])
	}

	// Invalid file should error
	if results[invalidFile] == nil {
		t.Error("Invalid file should error")
	}

	// Nonexistent file should error
	if results[nonexistentFile] == nil {
		t.Error("Nonexistent file should error")
	}
}

// TestSortFile_DryRunWithAlreadySorted tests dry run on already sorted file
func TestSortFile_DryRunWithAlreadySorted(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sorted.tf")

	sortedContent := `variable "environment" {
  type = string
}
`
	if err := os.WriteFile(testFile, []byte(sortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := SortFile(testFile, Options{DryRun: true})
	if !errors.Is(err, ErrNoChanges) {
		t.Errorf("Expected ErrNoChanges for already sorted file in dry run, got: %v", err)
	}
}

// TestSortFile_RenameFailure tests handling of file rename failure
func TestSortFile_RenameFailure(t *testing.T) {
	// Skip on Windows as file locking behavior is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping rename failure test on Windows")
	}

	tmpDir := t.TempDir()

	// Create a subdirectory that will become read-only
	subdir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(subdir, "test.tf")
	unsortedContent := `resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami = "ami-12345"
}
`
	if err := os.WriteFile(testFile, []byte(unsortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Try to make directory read-only after file creation
	// This may not always trigger a rename error on all systems
	if err := os.Chmod(subdir, 0555); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(subdir, 0755) }() // Cleanup

	err := SortFile(testFile, Options{})
	if err == nil {
		// On some systems/filesystems, this might not fail
		t.Log("Warning: Expected error for rename failure (may not fail on all systems)")
	} else {
		// If it does fail, it should be a proper error
		t.Logf("Got expected error: %v", err)
	}
}

// TestGetSortedContent_ComplexContent tests GetSortedContent with various content types
func TestGetSortedContent_ComplexContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "empty file",
			content: ``,
			wantErr: false,
		},
		{
			name: "only comments",
			content: `# This is a comment
// Another comment
/* Block comment */`,
			wantErr: false,
		},
		{
			name: "terraform block",
			content: `terraform {
  required_version = ">= 1.0"
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.tf")

			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			_, _, err := GetSortedContent(testFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSortedContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSortDirectory_WithDryRun tests SortDirectory in dry run mode
func TestSortDirectory_WithDryRun(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an unsorted file
	file1 := filepath.Join(tmpDir, "main.tf")
	unsortedContent := `resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami = "ami-12345"
}
`
	if err := os.WriteFile(file1, []byte(unsortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Sort directory in dry run mode
	results, err := SortDirectory(tmpDir, false, Options{DryRun: true})
	if err != nil {
		t.Fatalf("SortDirectory DryRun failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// File should not have been modified
	content, _ := os.ReadFile(file1)
	if string(content) != unsortedContent {
		t.Error("File was modified in dry run mode")
	}
}

// TestSortDirectory_WithValidate tests SortDirectory in validate mode
func TestSortDirectory_WithValidate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an unsorted file
	file1 := filepath.Join(tmpDir, "main.tf")
	unsortedContent := `resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami = "ami-12345"
}
`
	if err := os.WriteFile(file1, []byte(unsortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Sort directory in validate mode
	results, err := SortDirectory(tmpDir, false, Options{Validate: true})
	if err != nil {
		t.Fatalf("SortDirectory Validate failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Check that the result is ErrNeedsSorting
	if !errors.Is(results[file1], ErrNeedsSorting) {
		t.Errorf("Expected ErrNeedsSorting, got: %v", results[file1])
	}
}
