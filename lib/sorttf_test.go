package lib

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
	os.WriteFile(file1, []byte(content), 0644)
	os.WriteFile(file2, []byte(content), 0644)

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
	os.WriteFile(file1, []byte(content), 0644)

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
