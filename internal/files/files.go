// Package files provides file traversal and validation utilities for HCL files.
//
// This package handles discovery of Terraform (.tf) and Terragrunt (.hcl) files,
// with logic to skip common directories like .terraform and .terragrunt-cache.
// It provides both recursive and non-recursive file discovery.
package files

import (
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sorttf/internal/errors"
)

// IsValidFile checks if a file should be processed based on its name and type.
// Returns true for .tf and .hcl files (case-insensitive), excluding:
//   - Directories
//   - .terraform.lock.hcl (Terraform lock file)
//   - Files starting with .terraform
func IsValidFile(_ string, info os.FileInfo) bool {
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

// ShouldSkipDir checks if a directory should be skipped during traversal.
// Returns true for directories starting with ".terra" (e.g., .terraform, .terragrunt-cache).
// This prevents processing of Terraform/Terragrunt cache and state directories.
func ShouldSkipDir(_ string, info os.FileInfo) bool {
	if info == nil {
		return false
	}
	return info.IsDir() && strings.HasPrefix(info.Name(), ".terra")
}

// FindFiles discovers all valid Terraform and Terragrunt files in a directory.
// When recursive is true, it walks the directory tree, skipping .terraform* directories.
// When recursive is false, it only examines the immediate directory.
// Returns a slice of file paths, or an error if the root path is inaccessible.
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

// ValidateFilePath checks if a file path is valid and accessible.
// Returns an error if:
//   - The path is empty
//   - The path doesn't exist
//   - Permission is denied
//   - The path is a directory (not a file)
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

// ValidateDirectoryPath checks if a directory path is valid and accessible.
// Returns an error if:
//   - The path is empty
//   - The path doesn't exist
//   - Permission is denied
//   - The path is a file (not a directory)
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

// IsNotExistError checks if the error indicates a file or directory doesn't exist.
// Uses errors.Is to unwrap the error chain and check for ErrFileNotFound.
func IsNotExistError(err error) bool {
	return stderrors.Is(err, errors.ErrFileNotFound)
}

// IsPermissionError checks if the error indicates a permission issue.
// Uses errors.Is to unwrap the error chain and check for ErrPermissionDenied.
func IsPermissionError(err error) bool {
	return stderrors.Is(err, errors.ErrPermissionDenied)
}
