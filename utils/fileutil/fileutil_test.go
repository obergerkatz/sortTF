package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Mock FileInfo for testing
type mockFileInfo struct {
	name  string
	isDir bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) Sys() interface{}   { return nil }

func TestIsValidFile_TableDriven(t *testing.T) {
	type args struct {
		name  string
		isDir bool
	}
	tests := []struct {
		name     string
		args     args
		expected bool
	}{
		{"valid .tf", args{"foo.tf", false}, true},
		{"valid .hcl", args{"foo.hcl", false}, true},
		{"invalid .txt", args{"foo.txt", false}, false},
		{"lock file", args{".terraform.lock.hcl", false}, false},
		{"directory", args{"foo.tf", true}, false},
		{"empty name", args{"", false}, false},
		{"hidden .tf", args{".hidden.tf", false}, true},
		{"uppercase .TF", args{"FOO.TF", false}, true},
		{"mixed case .HcL", args{"foo.HcL", false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &mockFileInfo{name: tt.args.name, isDir: tt.args.isDir}
			if got := IsValidFile(tt.args.name, info); got != tt.expected {
				t.Errorf("IsValidFile(%q, isDir=%v) = %v, want %v", tt.args.name, tt.args.isDir, got, tt.expected)
			}
		})
	}
}

func TestIsValidFile_Symlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.tf")
	if err := os.WriteFile(target, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(dir, "link.tf")
	err := os.Symlink(target, symlink)
	if err != nil {
		t.Skip("Symlink not supported on this system")
	}
	info, err := os.Lstat(symlink)
	if err != nil {
		t.Fatal(err)
	}
	if !IsValidFile(symlink, info) {
		t.Error("Symlink to .tf file should be valid")
	}
}

func TestIsValidFile(t *testing.T) {
	dir := t.TempDir()

	// .tf file
	tfPath := filepath.Join(dir, "testfile.tf")
	if err := os.WriteFile(tfPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	tfInfo, err := os.Stat(tfPath)
	if err != nil {
		t.Fatal(err)
	}
	if !IsValidFile(tfPath, tfInfo) {
		t.Errorf("Expected .tf file to be valid")
	}

	// .hcl file
	hclPath := filepath.Join(dir, "testfile.hcl")
	if err := os.WriteFile(hclPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	hclInfo, err := os.Stat(hclPath)
	if err != nil {
		t.Fatal(err)
	}
	if !IsValidFile(hclPath, hclInfo) {
		t.Errorf("Expected .hcl file to be valid")
	}

	// .terraform.lock.hcl file
	lockPath := filepath.Join(dir, ".terraform.lock.hcl")
	if err := os.WriteFile(lockPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	lockInfo, err := os.Stat(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if IsValidFile(lockPath, lockInfo) {
		t.Errorf("Expected .terraform.lock.hcl to be invalid")
	}

	// .txt file
	txtPath := filepath.Join(dir, "testfile.txt")
	if err := os.WriteFile(txtPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	txtInfo, err := os.Stat(txtPath)
	if err != nil {
		t.Fatal(err)
	}
	if IsValidFile(txtPath, txtInfo) {
		t.Errorf("Expected .txt file to be invalid")
	}
}

func TestIsValidFileEdgeCases(t *testing.T) {
	// Test with nil FileInfo
	if IsValidFile("test.tf", nil) {
		t.Error("Should return false for nil FileInfo")
	}

	// Test with empty filename
	emptyInfo := &mockFileInfo{name: "", isDir: false}
	if IsValidFile("", emptyInfo) {
		t.Error("Should return false for empty filename")
	}

	// Test with hidden .tf files (should be valid)
	hiddenTfInfo := &mockFileInfo{name: ".hidden.tf", isDir: false}
	if !IsValidFile(".hidden.tf", hiddenTfInfo) {
		t.Error("Hidden .tf files should be valid")
	}

	// Test with uppercase extensions
	upperTfInfo := &mockFileInfo{name: "test.TF", isDir: false}
	if !IsValidFile("test.TF", upperTfInfo) {
		t.Error("Uppercase .TF files should be valid")
	}

	// Test with mixed case extensions
	mixedHclInfo := &mockFileInfo{name: "test.HcL", isDir: false}
	if !IsValidFile("test.HcL", mixedHclInfo) {
		t.Error("Mixed case .HcL files should be valid")
	}
}

func TestFindFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".terraform"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.hcl"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignore.txt"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".terraform", "should_ignore.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := FindFiles(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
	for _, f := range files {
		if filepath.Base(f) != "main.tf" && filepath.Base(f) != "main.hcl" {
			t.Errorf("Unexpected file: %s", f)
		}
	}

	// Test recursive
	files, err = FindFiles(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("Expected 2 files in recursive, got %d", len(files))
	}
}

func TestFindFilesErrorHandling(t *testing.T) {
	// Test with non-existent directory
	_, err := FindFiles("/non/existent/path", false)
	if err == nil {
		t.Error("Should return error for non-existent directory")
	}

	// Test with empty directory
	emptyDir := t.TempDir()
	files, err := FindFiles(emptyDir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", len(files))
	}
}

func TestFindFilesRecursiveEdgeCases(t *testing.T) {
	dir := t.TempDir()

	// Create nested structure with various file types
	if err := os.MkdirAll(filepath.Join(dir, "subdir", "deep"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "subdir", "deep", "nested.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "subdir", "deep", "nested.hcl"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "subdir", "deep", "ignore.txt"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .terraform directory in subdirectory
	if err := os.Mkdir(filepath.Join(dir, "subdir", ".terraform"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "subdir", ".terraform", "should_ignore.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := FindFiles(dir, true)
	if err != nil {
		t.Fatal(err)
	}

	expectedCount := 2 // nested.tf and nested.hcl
	if len(files) != expectedCount {
		t.Errorf("Expected %d files, got %d", expectedCount, len(files))
	}

	// Verify no files from .terraform directory are included
	for _, f := range files {
		if strings.Contains(f, ".terraform") {
			t.Errorf("Found file from .terraform directory: %s", f)
		}
	}
}

func TestFindFilesFileNameEdgeCases(t *testing.T) {
	dir := t.TempDir()

	// Test with files that have dots in names
	if err := os.WriteFile(filepath.Join(dir, "my.config.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "terraform.backup.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.backup.hcl"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with files that start with dots
	if err := os.WriteFile(filepath.Join(dir, ".env.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".config.hcl"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with files that have spaces
	if err := os.WriteFile(filepath.Join(dir, "my file.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config file.hcl"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := FindFiles(dir, false)
	if err != nil {
		t.Fatal(err)
	}

	expectedCount := 7 // all .tf and .hcl files (including files with spaces)
	if len(files) != expectedCount {
		t.Errorf("Expected %d files, got %d", expectedCount, len(files))
	}
}

func TestFindFilesIntegration(t *testing.T) {
	// Test with a realistic Terraform project structure
	dir := t.TempDir()

	// Create typical Terraform project structure
	if err := os.MkdirAll(filepath.Join(dir, "modules", "vpc"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "environments", "dev"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "environments", "prod"), 0755); err != nil {
		t.Fatal(err)
	}

	// Main files
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "variables.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "outputs.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "terragrunt.hcl"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Module files
	if err := os.WriteFile(filepath.Join(dir, "modules", "vpc", "main.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "modules", "vpc", "variables.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Environment files
	if err := os.WriteFile(filepath.Join(dir, "environments", "dev", "main.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "environments", "dev", "terragrunt.hcl"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "environments", "prod", "main.tf"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "environments", "prod", "terragrunt.hcl"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Files that should be ignored
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".terraform.lock.hcl"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := FindFiles(dir, true)
	if err != nil {
		t.Fatal(err)
	}

	expectedCount := 10 // all .tf and .hcl files
	if len(files) != expectedCount {
		t.Errorf("Expected %d files, got %d", expectedCount, len(files))
		for _, f := range files {
			t.Logf("Found: %s", f)
		}
	}
}

func TestFindFilesPerformance(t *testing.T) {
	dir := t.TempDir()

	// Create many files to test performance
	for i := 0; i < 1000; i++ {
		if i%2 == 0 {
			if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("file%d.tf", i)), []byte(""), 0644); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("file%d.hcl", i)), []byte(""), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	// Add some non-target files
	for i := 0; i < 100; i++ {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("file%d.txt", i)), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	start := time.Now()
	files, err := FindFiles(dir, false)
	duration := time.Since(start)

	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1000 {
		t.Errorf("Expected 1000 files, got %d", len(files))
	}

	// Performance assertion (adjust threshold as needed)
	if duration > 100*time.Millisecond {
		t.Errorf("FindFiles took too long: %v", duration)
	}
}

func TestFileUtilError(t *testing.T) {
	// Test FileUtilError creation and methods
	err := &FileUtilError{
		Op:   "TestOp",
		Path: "/test/path",
		Err:  fmt.Errorf("test error"),
	}

	expectedMsg := "fileutil TestOp /test/path: test error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	if err.Unwrap().Error() != "test error" {
		t.Errorf("Expected unwrapped error 'test error', got '%s'", err.Unwrap().Error())
	}
}

func TestValidateFilePath(t *testing.T) {
	// Test with empty path
	err := ValidateFilePath("")
	if err == nil {
		t.Error("Expected error for empty path")
	}
	if !IsFileUtilError(err) {
		t.Error("Expected FileUtilError for empty path")
	}

	// Test with non-existent file
	err = ValidateFilePath("/non/existent/file.tf")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if !IsNotExistError(err) {
		t.Error("Expected not exist error for non-existent file")
	}

	// Test with directory (should fail)
	dir := t.TempDir()
	err = ValidateFilePath(dir)
	if err == nil {
		t.Error("Expected error for directory path")
	}
	if !IsFileUtilError(err) {
		t.Error("Expected FileUtilError for directory path")
	}

	// Test with valid file
	file := filepath.Join(dir, "test.tf")
	if err := os.WriteFile(file, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	err = ValidateFilePath(file)
	if err != nil {
		t.Errorf("Expected no error for valid file, got: %v", err)
	}
}

func TestValidateDirectoryPath(t *testing.T) {
	// Test with empty path
	err := ValidateDirectoryPath("")
	if err == nil {
		t.Error("Expected error for empty path")
	}
	if !IsFileUtilError(err) {
		t.Error("Expected FileUtilError for empty path")
	}

	// Test with non-existent directory
	err = ValidateDirectoryPath("/non/existent/dir")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
	if !IsNotExistError(err) {
		t.Error("Expected not exist error for non-existent directory")
	}

	// Test with file (should fail)
	dir := t.TempDir()
	file := filepath.Join(dir, "test.tf")
	if err := os.WriteFile(file, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	err = ValidateDirectoryPath(file)
	if err == nil {
		t.Error("Expected error for file path")
	}
	if !IsFileUtilError(err) {
		t.Error("Expected FileUtilError for file path")
	}

	// Test with valid directory
	err = ValidateDirectoryPath(dir)
	if err != nil {
		t.Errorf("Expected no error for valid directory, got: %v", err)
	}
}

func TestErrorHelperFunctions(t *testing.T) {
	// Test IsFileUtilError
	fileUtilErr := &FileUtilError{Op: "Test", Path: "/test", Err: fmt.Errorf("test")}
	if !IsFileUtilError(fileUtilErr) {
		t.Error("IsFileUtilError should return true for FileUtilError")
	}
	if IsFileUtilError(fmt.Errorf("regular error")) {
		t.Error("IsFileUtilError should return false for regular error")
	}

	// Test IsNotExistError
	notExistErr := &FileUtilError{Op: "Test", Path: "/test", Err: fmt.Errorf("does not exist")}
	if !IsNotExistError(notExistErr) {
		t.Error("IsNotExistError should return true for not exist error")
	}
	if !IsNotExistError(os.ErrNotExist) {
		t.Error("IsNotExistError should return true for os.ErrNotExist")
	}

	// Test IsPermissionError
	permErr := &FileUtilError{Op: "Test", Path: "/test", Err: fmt.Errorf("permission denied")}
	if !IsPermissionError(permErr) {
		t.Error("IsPermissionError should return true for permission error")
	}
	if !IsPermissionError(os.ErrPermission) {
		t.Error("IsPermissionError should return true for os.ErrPermission")
	}

	// Test GetFileUtilErrorPath
	if GetFileUtilErrorPath(fileUtilErr) != "/test" {
		t.Error("GetFileUtilErrorPath should return the path")
	}
	if GetFileUtilErrorPath(fmt.Errorf("regular error")) != "" {
		t.Error("GetFileUtilErrorPath should return empty string for non-FileUtilError")
	}

	// Test GetFileUtilErrorOp
	if GetFileUtilErrorOp(fileUtilErr) != "Test" {
		t.Error("GetFileUtilErrorOp should return the operation")
	}
	if GetFileUtilErrorOp(fmt.Errorf("regular error")) != "" {
		t.Error("GetFileUtilErrorOp should return empty string for non-FileUtilError")
	}
}

func TestFindFilesEnhancedErrorHandling(t *testing.T) {
	// Test with non-existent directory
	_, err := FindFiles("/non/existent/directory", false)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
	if !IsFileUtilError(err) {
		t.Error("Expected FileUtilError for non-existent directory")
	}
	if !IsNotExistError(err) {
		t.Error("Expected not exist error for non-existent directory")
	}

	// Test with valid directory
	dir := t.TempDir()
	files, err := FindFiles(dir, false)
	if err != nil {
		t.Errorf("Expected no error for valid directory, got: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", len(files))
	}
}
