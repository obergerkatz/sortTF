package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/obergerkatz/sortTF/config"
)

// BenchmarkProcessFilesConcurrent measures concurrent file processing performance
func BenchmarkProcessFilesConcurrent(b *testing.B) {
	// Create a temporary directory with multiple files
	tmpDir := b.TempDir()

	unsortedContent := `resource "aws_instance" "test" {
  instance_type = "t2.micro"
  ami = "ami-123"
}

variable "environment" {
  type = string
}

resource "aws_s3_bucket" "data" {
  bucket_name = "test"
  acl = "private"
}
`

	// Create 50 files to process
	for i := 1; i <= 50; i++ {
		filePath := filepath.Join(tmpDir, fmt.Sprintf("file%d.tf", i))
		//nolint:gosec // G306: Benchmark test files can use 0644 permissions
		if err := os.WriteFile(filePath, []byte(unsortedContent), 0644); err != nil {
			b.Fatal(err)
		}
	}

	// Find the files
	files := []string{}
	entries, _ := os.ReadDir(tmpDir)
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(tmpDir, entry.Name()))
		}
	}

	// Create config
	config := &config.Config{
		Root:      tmpDir,
		Recursive: false,
		DryRun:    false,
		Verbose:   false,
		Validate:  false,
	}

	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		var stdout, stderr bytes.Buffer
		processFilesConcurrent(files, config, &stdout, &stderr)
	}
}

// BenchmarkProcessFilesSerial measures serial file processing performance
func BenchmarkProcessFilesSerial(b *testing.B) {
	// Create a temporary directory with multiple files
	tmpDir := b.TempDir()

	unsortedContent := `resource "aws_instance" "test" {
  instance_type = "t2.micro"
  ami = "ami-123"
}

variable "environment" {
  type = string
}

resource "aws_s3_bucket" "data" {
  bucket_name = "test"
  acl = "private"
}
`

	// Create 50 files to process
	for i := 1; i <= 50; i++ {
		filePath := filepath.Join(tmpDir, fmt.Sprintf("file%d.tf", i))
		//nolint:gosec // G306: Benchmark test files can use 0644 permissions
		if err := os.WriteFile(filePath, []byte(unsortedContent), 0644); err != nil {
			b.Fatal(err)
		}
	}

	// Find the files
	files := []string{}
	entries, _ := os.ReadDir(tmpDir)
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(tmpDir, entry.Name()))
		}
	}

	// Create config
	config := &config.Config{
		Root:      tmpDir,
		Recursive: false,
		DryRun:    false,
		Verbose:   false,
		Validate:  false,
	}

	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		var stdout, stderr bytes.Buffer
		processFilesSerial(files, config, &stdout, &stderr)
	}
}
