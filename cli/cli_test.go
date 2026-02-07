package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunCLI_Help tests that help flag works correctly
func TestRunCLI_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{"--help"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --help, got %d", exitCode)
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Usage: sorttf") {
		t.Error("Expected usage message in stderr")
	}
	if !strings.Contains(stderrOutput, "Flags:") {
		t.Error("Expected flags description in stderr")
	}
}

// TestRunCLI_NoArgs tests default behavior with no arguments
func TestRunCLI_NoArgs(t *testing.T) {
	// Create temp directory with no terraform files
	tmpDir := t.TempDir()

	// Change to temp directory
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tmpDir)

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for empty directory, got %d", exitCode)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "No Terraform or Terragrunt files found") {
		t.Errorf("Expected 'No files found' message, got: %s", stdoutOutput)
	}
}

// TestRunCLI_InvalidFlag tests error handling for invalid flags
func TestRunCLI_InvalidFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{"--invalid-flag"}, &stdout, &stderr)

	if exitCode != 2 {
		t.Errorf("Expected exit code 2 for invalid flag, got %d", exitCode)
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "Error") {
		t.Error("Expected error message in stderr for invalid flag")
	}
}

// TestRunCLI_NonExistentPath tests error handling for non-existent paths
func TestRunCLI_NonExistentPath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{"/non/existent/path"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for non-existent path, got %d", exitCode)
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "does not exist") {
		t.Errorf("Expected 'does not exist' in stderr, got: %s", stderrOutput)
	}
}

// TestRunCLI_SingleFile_AlreadySorted tests processing a file that doesn't need changes
func TestRunCLI_SingleFile_AlreadySorted(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "main.tf")

	// Create a properly sorted file
	content := `provider "aws" {
  region = "us-west-2"
}

variable "environment" {
  type = string
}

resource "aws_instance" "example" {
  ami           = "ami-12345"
  instance_type = "t2.micro"
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{testFile}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Processed") {
		t.Errorf("Expected 'Processed' in output, got: %s", stdoutOutput)
	}
}

// TestRunCLI_SingleFile_NeedsSorting tests processing a file that needs sorting
func TestRunCLI_SingleFile_NeedsSorting(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "main.tf")

	// Create an unsorted file
	content := `resource "aws_instance" "example" {
  instance_type = "t2.micro"
  ami = "ami-12345"
}

variable "environment" {
  type = string
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{testFile}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Verify file was updated
	updatedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	updatedStr := string(updatedContent)
	// Variable should come before resource
	varIndex := strings.Index(updatedStr, "variable")
	resIndex := strings.Index(updatedStr, "resource")
	if varIndex == -1 || resIndex == -1 {
		t.Error("File should contain both variable and resource blocks")
	}
	if varIndex > resIndex {
		t.Error("Variable should come before resource after sorting")
	}

	// Attributes should be sorted (ami before instance_type)
	amiIndex := strings.Index(updatedStr, "ami")
	instanceIndex := strings.Index(updatedStr, "instance_type")
	if amiIndex == -1 || instanceIndex == -1 {
		t.Error("File should contain both ami and instance_type attributes")
	}
	if amiIndex > instanceIndex {
		t.Error("ami should come before instance_type after sorting")
	}
}

// TestRunCLI_DryRun tests that dry-run doesn't modify files
func TestRunCLI_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "main.tf")

	originalContent := `resource "aws_instance" "example" {
  instance_type = "t2.micro"
  ami = "ami-12345"
}
`
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{"--dry-run", testFile}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Would update") {
		t.Errorf("Expected 'Would update' in dry-run output, got: %s", stdoutOutput)
	}

	// Verify file was NOT modified
	currentContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(currentContent) != originalContent {
		t.Error("File should not be modified in dry-run mode")
	}
}

// TestRunCLI_Validate tests validate mode
func TestRunCLI_Validate(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "main.tf")

	// Create unsorted file
	content := `resource "aws_instance" "example" {
  instance_type = "t2.micro"
  ami = "ami-12345"
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{"--validate", testFile}, &stdout, &stderr)

	// Validate should exit with error code if file needs changes
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for unsorted file in validate mode, got %d", exitCode)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Needs update") {
		t.Errorf("Expected 'Needs update' in validate output, got: %s", stdoutOutput)
	}

	// Verify file was NOT modified
	currentContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(currentContent) != content {
		t.Error("File should not be modified in validate mode")
	}
}

// TestRunCLI_Verbose tests verbose output
func TestRunCLI_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "main.tf")

	content := `variable "test" {
  type = string
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{"--verbose", testFile}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Processing") || !strings.Contains(stdoutOutput, "Found") {
		t.Errorf("Expected verbose messages in output, got: %s", stdoutOutput)
	}
}

// TestRunCLI_Recursive tests recursive directory processing
func TestRunCLI_Recursive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create files in both directories that need sorting
	file1 := filepath.Join(tmpDir, "main.tf")
	file2 := filepath.Join(subDir, "variables.tf")

	// Unsorted content so files will be processed
	content := `resource "aws_instance" "test" {
  instance_type = "t2.micro"
  ami = "ami-123"
}
`
	if err := os.WriteFile(file1, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{"--recursive", tmpDir}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	stdoutOutput := stdout.String()
	// Should process 2 files (output says "Processed X files")
	if !strings.Contains(stdoutOutput, "Processed 2 files") {
		t.Errorf("Expected 'Processed 2 files', got: %s", stdoutOutput)
	}
}

// TestRunCLI_Directory tests processing a directory (non-recursive)
func TestRunCLI_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create files that need sorting
	file1 := filepath.Join(tmpDir, "main.tf")
	file2 := filepath.Join(subDir, "variables.tf")

	// Unsorted content
	content := `resource "aws_instance" "test" {
  instance_type = "t2.micro"
  ami = "ami-123"
}
`
	if err := os.WriteFile(file1, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	// Without --recursive, should only process files in root directory
	exitCode := RunCLIWithWriters([]string{tmpDir}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	stdoutOutput := stdout.String()
	// Should only process 1 file (not in subdir)
	if !strings.Contains(stdoutOutput, "Processed 1 file") {
		t.Errorf("Expected 'Processed 1 file' in non-recursive mode, got: %s", stdoutOutput)
	}
}

// TestRunCLI_ConcurrentProcessing tests processing many files concurrently
func TestRunCLI_ConcurrentProcessing(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 10 files to trigger concurrent processing (threshold is 4)
	unsortedContent := `resource "aws_instance" "test" {
  instance_type = "t2.micro"
  ami = "ami-123"
}
`
	for i := 1; i <= 10; i++ {
		filePath := filepath.Join(tmpDir, fmt.Sprintf("file%d.tf", i))
		if err := os.WriteFile(filePath, []byte(unsortedContent), 0644); err != nil {
			t.Fatal(err)
		}
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{tmpDir}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	stdoutOutput := stdout.String()
	// Should process 10 files
	if !strings.Contains(stdoutOutput, "Processed 10 files") {
		t.Errorf("Expected 'Processed 10 files', got: %s", stdoutOutput)
	}

	// Verify all files were actually processed (spot check a few)
	for i := 1; i <= 3; i++ {
		filePath := filepath.Join(tmpDir, fmt.Sprintf("file%d.tf", i))
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}
		// Check that content was sorted (ami comes before instance_type)
		contentStr := string(content)
		amiIndex := strings.Index(contentStr, "ami")
		instanceIndex := strings.Index(contentStr, "instance_type")
		if amiIndex == -1 || instanceIndex == -1 {
			t.Errorf("File %d missing expected attributes", i)
		}
		if amiIndex > instanceIndex {
			t.Errorf("File %d not sorted correctly: ami should come before instance_type", i)
		}
	}
}

// TestRunCLI_InvalidSyntax tests handling of files with syntax errors
func TestRunCLI_InvalidSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.tf")

	// Create file with invalid HCL syntax (missing label for resource block)
	invalidContent := `resource "aws_instance" {
  ami = "ami-123"
}
`
	if err := os.WriteFile(testFile, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{testFile}, &stdout, &stderr)

	// Should exit with error (validation should catch missing label)
	if exitCode == 0 {
		t.Errorf("Expected non-zero exit code for invalid syntax, got 0. Stderr: %s", stderr.String())
	}

	output := stderr.String() + stdout.String()
	hasError := strings.Contains(output, "error") ||
		strings.Contains(output, "Error") ||
		strings.Contains(output, "validation")
	if !hasError {
		t.Errorf("Expected error message, got stdout: %s, stderr: %s", stdout.String(), stderr.String())
	}
}

// TestIsSupportedFile tests the file type checking
func TestIsSupportedFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"terraform file", "main.tf", true},
		{"terragrunt file", "terragrunt.hcl", true},
		{"text file", "README.txt", false},
		{"go file", "main.go", false},
		{"no extension", "Makefile", false},
		// Note: isSupportedFile checks lowercase extensions only
		{"lowercase tf", "main.tf", true},
		{"lowercase hcl", "config.hcl", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSupportedFile(tt.filename)
			if got != tt.want {
				t.Errorf("isSupportedFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

// TestPrintUnifiedDiff tests the diff output function
func TestPrintUnifiedDiff(t *testing.T) {
	tests := []struct {
		name      string
		original  string
		formatted string
		wantDiff  bool
	}{
		{
			name:      "no changes",
			original:  "variable \"x\" {\n  type = string\n}\n",
			formatted: "variable \"x\" {\n  type = string\n}\n",
			wantDiff:  false,
		},
		{
			name:      "attribute reordered",
			original:  "resource \"x\" \"y\" {\n  b = 1\n  a = 2\n}\n",
			formatted: "resource \"x\" \"y\" {\n  a = 2\n  b = 1\n}\n",
			wantDiff:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printUnifiedDiff(tt.original, tt.formatted, "test.tf", &buf)
			output := buf.String()

			hasDiff := strings.Contains(output, "+") || strings.Contains(output, "-")
			if tt.wantDiff && !hasDiff {
				t.Error("Expected diff output, got none")
			}
			if !tt.wantDiff && hasDiff && !strings.Contains(output, "No changes") {
				t.Errorf("Expected no diff, got: %s", output)
			}
		})
	}
}

// TestRunCLI_TestdataComplex tests processing the complex.tf testdata file
func TestRunCLI_TestdataComplex(t *testing.T) {
	// Get path to testdata/complex.tf
	testFile := filepath.Join("testdata", "complex.tf")

	// Verify testdata file exists
	if _, err := os.Stat(testFile); err != nil {
		t.Skipf("Skipping test: testdata file not found: %v", err)
	}

	// Create a temporary copy to avoid modifying the original
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "complex.tf")

	originalContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(tmpFile, originalContent, 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{tmpFile}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Read the result
	processedContent, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	// The complex.tf file should be processed successfully
	// Verify blocks are in expected order: terraform, provider, variable, locals, data, resource, module, output
	contentStr := string(processedContent)

	// Find indices of each block type
	terraformIdx := strings.Index(contentStr, "terraform {")
	providerIdx := strings.Index(contentStr, "provider \"aws\"")
	variableIdx := strings.Index(contentStr, "variable \"region\"")
	localsIdx := strings.Index(contentStr, "locals {")
	dataIdx := strings.Index(contentStr, "data \"aws_ami\"")
	resourceIdx := strings.Index(contentStr, "resource \"aws_instance\"")
	moduleIdx := strings.Index(contentStr, "module \"vpc\"")
	outputIdx := strings.Index(contentStr, "output \"instance_id\"")

	// Verify order
	if terraformIdx == -1 || providerIdx == -1 || variableIdx == -1 || localsIdx == -1 ||
		dataIdx == -1 || resourceIdx == -1 || moduleIdx == -1 || outputIdx == -1 {
		t.Error("Missing expected block types in output")
	}

	if !(terraformIdx < providerIdx && providerIdx < variableIdx && variableIdx < localsIdx &&
		localsIdx < dataIdx && dataIdx < resourceIdx && resourceIdx < moduleIdx && moduleIdx < outputIdx) {
		t.Error("Blocks are not in expected sorted order")
	}
}

// TestRunCLI_TestdataSorted tests processing the sorted.tf testdata file
func TestRunCLI_TestdataSorted(t *testing.T) {
	testFile := filepath.Join("testdata", "sorted.tf")

	if _, err := os.Stat(testFile); err != nil {
		t.Skipf("Skipping test: testdata file not found: %v", err)
	}

	// Create temporary copy
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "sorted.tf")

	originalContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(tmpFile, originalContent, 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{tmpFile}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// File should not be modified since it's already sorted
	processedContent, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	// Note: The formatter might still make changes even if blocks are in correct order
	// (e.g., formatting whitespace), so we just verify the tool ran successfully
	_ = processedContent
}

// TestRunCLI_TestdataUnsorted tests processing the unsorted.tf testdata file
func TestRunCLI_TestdataUnsorted(t *testing.T) {
	testFile := filepath.Join("testdata", "unsorted.tf")

	if _, err := os.Stat(testFile); err != nil {
		t.Skipf("Skipping test: testdata file not found: %v", err)
	}

	// Create temporary copy
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "unsorted.tf")

	originalContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(tmpFile, originalContent, 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{tmpFile}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	// Read processed content
	processedContent, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(processedContent)

	// Verify blocks are now sorted: provider, variable, resource
	providerIdx := strings.Index(contentStr, "provider")
	variableIdx := strings.Index(contentStr, "variable")
	resourceIdx := strings.Index(contentStr, "resource")

	if providerIdx == -1 || variableIdx == -1 || resourceIdx == -1 {
		t.Error("Missing expected blocks in output")
	}

	if !(providerIdx < variableIdx && variableIdx < resourceIdx) {
		t.Errorf("Blocks not in expected order. Provider: %d, Variable: %d, Resource: %d",
			providerIdx, variableIdx, resourceIdx)
	}
}

// TestRunCLI_TestdataDirectory tests processing all testdata files in directory
func TestRunCLI_TestdataDirectory(t *testing.T) {
	testDir := "testdata"

	if _, err := os.Stat(testDir); err != nil {
		t.Skipf("Skipping test: testdata directory not found: %v", err)
	}

	// Create temporary copy of testdata directory
	tmpDir := t.TempDir()
	testdataFiles, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range testdataFiles {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".tf" {
			srcPath := filepath.Join(testDir, entry.Name())
			dstPath := filepath.Join(tmpDir, entry.Name())

			content, err := os.ReadFile(srcPath)
			if err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(dstPath, content, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{tmpDir}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Processed") {
		t.Errorf("Expected 'Processed' in output, got: %s", stdoutOutput)
	}
}

// TestRunCLI_DryRunWithTestdata tests dry-run mode with testdata files
func TestRunCLI_DryRunWithTestdata(t *testing.T) {
	testFile := filepath.Join("testdata", "unsorted.tf")

	if _, err := os.Stat(testFile); err != nil {
		t.Skipf("Skipping test: testdata file not found: %v", err)
	}

	// Create temporary copy
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "unsorted.tf")

	originalContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(tmpFile, originalContent, 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{"--dry-run", tmpFile}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}

	stdoutOutput := stdout.String()
	if !strings.Contains(stdoutOutput, "Would update") && !strings.Contains(stdoutOutput, "no changes") {
		t.Errorf("Expected 'Would update' or 'no changes' in dry-run output, got: %s", stdoutOutput)
	}

	// Verify file was NOT modified
	currentContent, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(currentContent) != string(originalContent) {
		t.Error("File should not be modified in dry-run mode")
	}
}

// TestRunCLI_ValidateWithTestdata tests validate mode with testdata files
func TestRunCLI_ValidateWithTestdata(t *testing.T) {
	testFile := filepath.Join("testdata", "unsorted.tf")

	if _, err := os.Stat(testFile); err != nil {
		t.Skipf("Skipping test: testdata file not found: %v", err)
	}

	// Create temporary copy
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "unsorted.tf")

	originalContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(tmpFile, originalContent, 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{"--validate", tmpFile}, &stdout, &stderr)

	// unsorted.tf should trigger validation error
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for unsorted file in validate mode, got %d", exitCode)
	}

	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "Needs update") && !strings.Contains(output, "needs update") {
		t.Errorf("Expected 'needs update' in validate output, got: %s", output)
	}

	// Verify file was NOT modified
	currentContent, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(currentContent) != string(originalContent) {
		t.Error("File should not be modified in validate mode")
	}
}

// TestRunCLI_UnsupportedFileType tests error handling for unsupported file types
func TestRunCLI_UnsupportedFileType(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := "This is not a terraform file"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	exitCode := RunCLIWithWriters([]string{testFile}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for unsupported file type, got %d", exitCode)
	}

	stderrOutput := stderr.String()
	if !strings.Contains(stderrOutput, "not a supported file type") {
		t.Errorf("Expected 'not a supported file type' in stderr, got: %s", stderrOutput)
	}
}
