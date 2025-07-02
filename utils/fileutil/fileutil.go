package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Custom error types for better error handling
type FileUtilError struct {
	Op   string
	Path string
	Err  error
}

func (e *FileUtilError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("fileutil %s %s: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("fileutil %s %s", e.Op, e.Path)
}

func (e *FileUtilError) Unwrap() error {
	return e.Err
}

// IsValidFile checks if a file should be processed based on its name and type
func IsValidFile(path string, info os.FileInfo) bool {
	if info == nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	if strings.HasPrefix(info.Name(), ".terraform") || info.Name() == ".terraform.lock.hcl" {
		return false
	}
	name := strings.ToLower(info.Name())
	if strings.HasSuffix(name, ".tf") || strings.HasSuffix(name, ".hcl") {
		return true
	}
	return false
}

// ShouldSkipDir checks if a directory should be skipped during traversal
func ShouldSkipDir(path string, info os.FileInfo) bool {
	if info == nil {
		return false
	}
	return info.IsDir() && info.Name() == ".terraform"
}

// FindFiles recursively or non-recursively finds all valid Terraform and Terragrunt files
func FindFiles(root string, recursive bool) ([]string, error) {
	// Check if root path exists
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, &FileUtilError{
			Op:   "FindFiles",
			Path: root,
			Err:  fmt.Errorf("path does not exist: %s", root),
		}
	}

	var files []string
	if recursive {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				// Return a wrapped error with context
				return &FileUtilError{
					Op:   "Walk",
					Path: path,
					Err:  err,
				}
			}
			if ShouldSkipDir(path, info) {
				return filepath.SkipDir
			}
			if IsValidFile(path, info) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return files, nil
	} else {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, &FileUtilError{
				Op:   "ReadDir",
				Path: root,
				Err:  err,
			}
		}
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() == ".terraform" {
				continue
			}
			if entry.Type().IsRegular() {
				name := strings.ToLower(entry.Name())
				if (strings.HasSuffix(name, ".tf") || strings.HasSuffix(name, ".hcl")) && entry.Name() != ".terraform.lock.hcl" {
					files = append(files, filepath.Join(root, entry.Name()))
				}
			}
		}
		return files, nil
	}
}

// ValidateFilePath checks if a file path is valid and accessible
func ValidateFilePath(path string) error {
	if path == "" {
		return &FileUtilError{
			Op:   "ValidateFilePath",
			Path: path,
			Err:  fmt.Errorf("empty path provided"),
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &FileUtilError{
				Op:   "ValidateFilePath",
				Path: path,
				Err:  fmt.Errorf("file does not exist"),
			}
		}
		if os.IsPermission(err) {
			return &FileUtilError{
				Op:   "ValidateFilePath",
				Path: path,
				Err:  fmt.Errorf("permission denied"),
			}
		}
		return &FileUtilError{
			Op:   "ValidateFilePath",
			Path: path,
			Err:  err,
		}
	}

	if info.IsDir() {
		return &FileUtilError{
			Op:   "ValidateFilePath",
			Path: path,
			Err:  fmt.Errorf("path is a directory, expected a file"),
		}
	}

	return nil
}

// ValidateDirectoryPath checks if a directory path is valid and accessible
func ValidateDirectoryPath(path string) error {
	if path == "" {
		return &FileUtilError{
			Op:   "ValidateDirectoryPath",
			Path: path,
			Err:  fmt.Errorf("empty path provided"),
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &FileUtilError{
				Op:   "ValidateDirectoryPath",
				Path: path,
				Err:  fmt.Errorf("directory does not exist"),
			}
		}
		if os.IsPermission(err) {
			return &FileUtilError{
				Op:   "ValidateDirectoryPath",
				Path: path,
				Err:  fmt.Errorf("permission denied"),
			}
		}
		return &FileUtilError{
			Op:   "ValidateDirectoryPath",
			Path: path,
			Err:  err,
		}
	}

	if !info.IsDir() {
		return &FileUtilError{
			Op:   "ValidateDirectoryPath",
			Path: path,
			Err:  fmt.Errorf("path is a file, expected a directory"),
		}
	}

	return nil
}

// IsFileUtilError checks if an error is a FileUtilError
func IsFileUtilError(err error) bool {
	_, ok := err.(*FileUtilError)
	return ok
}

// IsNotExistError checks if the error indicates a file/directory doesn't exist
func IsNotExistError(err error) bool {
	if fileUtilErr, ok := err.(*FileUtilError); ok {
		return strings.Contains(fileUtilErr.Err.Error(), "does not exist")
	}
	return os.IsNotExist(err)
}

// IsPermissionError checks if the error indicates a permission issue
func IsPermissionError(err error) bool {
	if fileUtilErr, ok := err.(*FileUtilError); ok {
		return strings.Contains(fileUtilErr.Err.Error(), "permission denied")
	}
	return os.IsPermission(err)
}

// GetFileUtilErrorPath extracts the path from a FileUtilError
func GetFileUtilErrorPath(err error) string {
	if fileUtilErr, ok := err.(*FileUtilError); ok {
		return fileUtilErr.Path
	}
	return ""
}

// GetFileUtilErrorOp extracts the operation from a FileUtilError
func GetFileUtilErrorOp(err error) string {
	if fileUtilErr, ok := err.(*FileUtilError); ok {
		return fileUtilErr.Op
	}
	return ""
}
