// Package files provides file traversal and validation utilities for HCL files.
package files

import (
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sorttf/internal/errors"
)

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
	return info.IsDir() && strings.HasPrefix(info.Name(), ".terra")
}

// FindFiles recursively or non-recursively finds all valid Terraform and Terragrunt files
func FindFiles(root string, recursive bool) ([]string, error) {
	// Check if root path exists
	if _, err := os.Stat(root); err != nil {
		return nil, errors.NewWithPath("FindFiles", root, errors.Wrap(err))
	}

	var foundFiles []string
	if recursive {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return errors.NewWithPath("Walk", path, errors.Wrap(err))
			}
			if ShouldSkipDir(path, info) {
				return filepath.SkipDir
			}
			if IsValidFile(path, info) {
				foundFiles = append(foundFiles, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return foundFiles, nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, errors.NewWithPath("ReadDir", root, errors.Wrap(err))
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() == ".terraform" {
			continue
		}
		if entry.Type().IsRegular() {
			name := strings.ToLower(entry.Name())
			if (strings.HasSuffix(name, ".tf") || strings.HasSuffix(name, ".hcl")) && entry.Name() != ".terraform.lock.hcl" {
				foundFiles = append(foundFiles, filepath.Join(root, entry.Name()))
			}
		}
	}
	return foundFiles, nil
}

// ValidateFilePath checks if a file path is valid and accessible
func ValidateFilePath(path string) error {
	if path == "" {
		return errors.NewWithPath("ValidateFilePath", path, fmt.Errorf("empty path provided"))
	}

	info, err := os.Stat(path)
	if err != nil {
		return errors.NewWithPath("ValidateFilePath", path, errors.Wrap(err))
	}

	if info.IsDir() {
		return errors.NewWithPath("ValidateFilePath", path, fmt.Errorf("path is a directory, expected a file"))
	}

	return nil
}

// ValidateDirectoryPath checks if a directory path is valid and accessible
func ValidateDirectoryPath(path string) error {
	if path == "" {
		return errors.NewWithPath("ValidateDirectoryPath", path, fmt.Errorf("empty path provided"))
	}

	info, err := os.Stat(path)
	if err != nil {
		return errors.NewWithPath("ValidateDirectoryPath", path, errors.Wrap(err))
	}

	if !info.IsDir() {
		return errors.NewWithPath("ValidateDirectoryPath", path, fmt.Errorf("path is a file, expected a directory"))
	}

	return nil
}

// IsNotExistError checks if the error indicates a file/directory doesn't exist
func IsNotExistError(err error) bool {
	return stderrors.Is(err, errors.ErrFileNotFound)
}

// IsPermissionError checks if the error indicates a permission issue
func IsPermissionError(err error) bool {
	return stderrors.Is(err, errors.ErrPermissionDenied)
}

// GetFileUtilErrorPath extracts the path from an Error (for backward compatibility)
func GetFileUtilErrorPath(err error) string {
	var e *errors.Error
	if stderrors.As(err, &e) {
		return e.Path
	}
	return ""
}
