// Package api provides a programmatic API for sorting Terraform and Terragrunt files.
//
// This package can be imported and used in other Go programs without needing
// to shell out to the CLI. It provides clean, testable functions with no I/O side effects.
//
// Example usage:
//
//	import "sorttf/api"
//
//	// Sort a single file
//	err := api.SortFile("main.tf", api.Options{})
//	if err != nil && !errors.Is(err, api.ErrNoChanges) {
//	    log.Fatal(err)
//	}
//
//	// Sort multiple files
//	results := api.SortFiles(paths, api.Options{DryRun: true})
//	for path, err := range results {
//	    if err != nil {
//	        fmt.Printf("%s: %v\n", path, err)
//	    }
//	}
package api

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"sorttf/hcl"
	"sorttf/internal/files"

	hcllib "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Options configures file sorting behavior.
type Options struct {
	// DryRun shows what would change without modifying files.
	DryRun bool

	// Validate checks if files are sorted without modifying them.
	// Returns ErrNeedsSorting if changes are needed.
	Validate bool
}

// Sentinel errors for common conditions.
var (
	// ErrNoChanges indicates a file is already sorted and formatted.
	ErrNoChanges = errors.New("file is already sorted")

	// ErrNeedsSorting indicates a file needs sorting (used in Validate mode).
	ErrNeedsSorting = errors.New("file needs sorting")
)

// GetSortedContent reads a file and returns its sorted and formatted content.
//
// This function does not modify the file; it only returns what the sorted
// content would be. Useful for computing diffs or previewing changes.
//
// Returns:
//   - content string: the sorted and formatted content
//   - changed bool: whether the content differs from the original
//   - error: parsing, validation, or I/O error
func GetSortedContent(path string) (content string, changed bool, err error) {
	// Step 1: Read file
	origContent, err := os.ReadFile(path)
	if err != nil {
		return "", false, fmt.Errorf("read file: %w", err)
	}

	// Step 2: Parse and validate
	parsed, err := hcl.ParseHCLFile(path)
	if err != nil {
		return "", false, fmt.Errorf("parse: %w", err)
	}

	if err := hcl.ValidateRequiredBlockLabels(parsed); err != nil {
		return "", false, fmt.Errorf("validate: %w", err)
	}

	// Step 3: Sort and format
	hclFile, diags := hclwrite.ParseConfig(origContent, path, hcllib.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return "", false, fmt.Errorf("parse for formatting: %w", diags)
	}

	formatted, err := hcl.SortAndFormatHCLFile(hclFile)
	if err != nil {
		return "", false, fmt.Errorf("sort/format: %w", err)
	}

	// Step 4: Check if changed
	hasChanges := !bytes.Equal(origContent, []byte(formatted))

	return formatted, hasChanges, nil
}

// SortFile sorts and formats a single Terraform or Terragrunt file.
//
// It reads the file, parses and validates the HCL, sorts blocks and attributes,
// applies formatting, and writes the result back atomically.
//
// Returns:
//   - nil: file was successfully sorted and written
//   - ErrNoChanges: file is already sorted (not an error condition)
//   - ErrNeedsSorting: file needs sorting (only in Validate mode)
//   - error: parsing, validation, or I/O error
//
// Behavior based on Options:
//   - DryRun: validates and checks for changes, but doesn't modify the file
//   - Validate: returns ErrNeedsSorting if changes needed, doesn't modify
//   - Normal: sorts and writes the file if changes are needed
func SortFile(path string, opts Options) error {
	// Get sorted content
	formatted, changed, err := GetSortedContent(path)
	if err != nil {
		return err
	}

	// No changes needed
	if !changed {
		return ErrNoChanges
	}

	// Handle modes
	if opts.Validate {
		return ErrNeedsSorting
	}

	if opts.DryRun {
		// Don't write, just indicate changes would be made
		return nil
	}

	// Write atomically (normal mode)
	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(formatted), 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmpFile, path); err != nil {
		// Try to clean up temp file
		_ = os.Remove(tmpFile)
		return fmt.Errorf("replace file: %w", err)
	}

	return nil
}

// SortFiles sorts multiple files and returns results for each.
//
// It processes files independently and continues on error, allowing you to
// see results for all files even if some fail.
//
// Returns a map of filePath -> error, where:
//   - nil: file was successfully sorted
//   - ErrNoChanges: file was already sorted (not an error)
//   - ErrNeedsSorting: file needs sorting (Validate mode)
//   - error: file failed to process
//
// Example:
//
//	results := SortFiles(paths, Options{})
//	for path, err := range results {
//	    if err != nil && !errors.Is(err, ErrNoChanges) {
//	        fmt.Printf("Error processing %s: %v\n", path, err)
//	    }
//	}
func SortFiles(paths []string, opts Options) map[string]error {
	results := make(map[string]error)
	for _, path := range paths {
		results[path] = SortFile(path, opts)
	}
	return results
}

// SortDirectory sorts all Terraform/Terragrunt files in a directory.
//
// It finds all .tf and .hcl files in the directory (optionally recursive)
// and sorts them according to the provided options.
//
// Returns a map of filePath -> error for all discovered files.
func SortDirectory(dir string, recursive bool, opts Options) (map[string]error, error) {
	// Find all files
	paths, err := files.FindFiles(dir, recursive)
	if err != nil {
		return nil, fmt.Errorf("find files: %w", err)
	}

	// Sort all files
	return SortFiles(paths, opts), nil
}
