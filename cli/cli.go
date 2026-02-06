// Package cli provides the command-line interface for sortTF.
package cli

import (
	"bytes"
	stderrors "errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"sorttf/config"
	"sorttf/hcl"
	"sorttf/internal/errors"
	"sorttf/internal/files"

	"github.com/fatih/color"
	hcllib "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Color configuration
var (
	errorColor   = color.New(color.FgRed, color.Bold)
	warningColor = color.New(color.FgYellow, color.Bold)
	successColor = color.New(color.FgGreen, color.Bold)
	infoColor    = color.New(color.FgBlue, color.Bold)
	fileColor    = color.New(color.FgCyan)
)

// RunCLI is the main entry point for CLI execution
func RunCLI(args []string) int {
	return RunCLIWithWriters(args, os.Stdout, os.Stderr)
}

// RunCLIWithWriters allows testing by providing custom writers
func RunCLIWithWriters(args []string, stdout, stderr io.Writer) int {
	config, err := config.ParseFlags(args, stderr)
	if err != nil {
		if err.Error() == "help" {
			// Help was requested, just exit with success
			return 0
		}
		errors.PrintError(err, stderr)
		return 2 // Usage error
	}

	return runMainLogic(config, stdout, stderr)
}

// runMainLogic executes the main CLI logic
func runMainLogic(config *config.Config, stdout, stderr io.Writer) int {
	// Check if the path is a file or directory
	fileInfo, err := os.Stat(config.Root)
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = errorColor.Fprintf(stderr, "❌ Path '%s' does not exist\n", fileColor.Sprint(config.Root))
		} else if os.IsPermission(err) {
			_, _ = errorColor.Fprintf(stderr, "🔒 Permission denied accessing '%s'\n", fileColor.Sprint(config.Root))
		} else {
			_, _ = errorColor.Fprintf(stderr, "❌ Error accessing '%s': %v\n", fileColor.Sprint(config.Root), err)
		}
		return 1
	}

	var filePaths []string

	if fileInfo.IsDir() {
		// It's a directory - validate and find files
		if err := files.ValidateDirectoryPath(config.Root); err != nil {
			if files.IsNotExistError(err) {
				_, _ = errorColor.Fprintf(stderr, "❌ Directory '%s' does not exist\n", fileColor.Sprint(config.Root))
			} else if files.IsPermissionError(err) {
				_, _ = errorColor.Fprintf(stderr, "🔒 Permission denied accessing directory '%s'\n", fileColor.Sprint(config.Root))
			} else {
				_, _ = errorColor.Fprintf(stderr, "❌ Error validating directory '%s': %v\n", fileColor.Sprint(config.Root), err)
			}
			return 1
		}

		// Find files to process
		filePaths, err = files.FindFiles(config.Root, config.Recursive)
		if err != nil {
			// Extract path from error if available
			var e *errors.Error
			path := config.Root
			if stderrors.As(err, &e) && e.Path != "" {
				path = e.Path
			}

			if files.IsNotExistError(err) {
				_, _ = errorColor.Fprintf(stderr, "❌ Path '%s' does not exist\n", fileColor.Sprint(path))
			} else if files.IsPermissionError(err) {
				_, _ = errorColor.Fprintf(stderr, "🔒 Permission denied accessing '%s'\n", fileColor.Sprint(path))
			} else {
				_, _ = errorColor.Fprintf(stderr, "❌ Error finding files: %v\n", err)
			}
			return 1
		}
	} else {
		// It's a file - check if it's a supported file type
		if !isSupportedFile(config.Root) {
			_, _ = errorColor.Fprintf(stderr, "❌ File '%s' is not a supported file type (.tf or .hcl)\n", fileColor.Sprint(config.Root))
			return 1
		}
		filePaths = []string{config.Root}
	}

	if len(filePaths) == 0 {
		_, _ = infoColor.Fprintf(stdout, "ℹ️  No Terraform or Terragrunt files found.\n")
		return 0
	}

	if config.Verbose {
		_, _ = infoColor.Fprintf(stdout, "📁 Found %d files:\n", len(filePaths))
		for _, f := range filePaths {
			_, _ = fmt.Fprintf(stdout, "   %s\n", fileColor.Sprint(f))
		}
	}

	// Process files
	processedCount := 0
	errorCount := 0
	noChangesCount := 0

	for _, filePath := range filePaths {
		if err := processFile(filePath, config, stdout, stderr); err != nil {
			if stderrors.Is(err, errors.ErrNoChanges) {
				noChangesCount++
			} else {
				errorCount++
				errors.PrintError(err, stderr)
				if config.Validate {
					// In validate mode, continue processing but will exit with error
					continue
				}
			}
		} else {
			processedCount++
		}
	}

	// Print summary
	if config.DryRun {
		if processedCount == 0 && errorCount == 0 {
			_, _ = successColor.Fprintf(stdout, "✅ Processed %d files, no changes needed\n", len(filePaths))
		} else {
			_, _ = infoColor.Fprintf(stdout, "📊 Processed %d files, %d would be updated\n", len(filePaths), processedCount)
		}
	} else {
		_, _ = successColor.Fprintf(stdout, "✅ Processed %d files\n", processedCount)
	}

	if errorCount > 0 {
		_, _ = errorColor.Fprintf(stderr, "❌ Encountered %d errors\n", errorCount)
		return 1 // Exit with error code if any errors occurred
	}

	return 0
}

// isSupportedFile checks if the file has a supported extension
func isSupportedFile(filePath string) bool {
	ext := filepath.Ext(filePath)
	return ext == ".tf" || ext == ".hcl"
}

// processFile handles a single file
func processFile(filePath string, config *config.Config, stdout, stderr io.Writer) error {
	if config.Verbose {
		_, _ = infoColor.Fprintf(stdout, "🔄 Processing: %s\n", fileColor.Sprint(filePath))
	}

	// Step 1: Read original file content
	origContent, err := os.ReadFile(filePath)
	if err != nil {
		return errors.New("processFile", fmt.Errorf("failed to read file: %v", err))
	}

	// Step 2: Parse and validate
	parsed, err := hcl.ParseHCLFile(filePath)
	if err != nil {
		if hcl.IsNotExistError(err) {
			return errors.New("processFile", fmt.Errorf("file not found: %s", filePath))
		} else if hcl.IsHCLParseError(err) {
			return errors.New("processFile", fmt.Errorf("syntax error in %s: %v", filePath, err))
		} else if hcl.IsParsingError(err) {
			return errors.New("processFile", fmt.Errorf("parsing error in %s: %v", filePath, err))
		}
		return errors.New("processFile", fmt.Errorf("failed to parse %s: %v", filePath, err))
	}
	if err := hcl.ValidateRequiredBlockLabels(parsed); err != nil {
		if hcl.IsValidationError(err) {
			return errors.New("processFile", fmt.Errorf("validation error in %s: %v", filePath, err))
		}
		return errors.New("processFile", fmt.Errorf("validation failed for %s: %v", filePath, err))
	}

	// Step 3: Sort and format
	hclFile, diags := hclwrite.ParseConfig(origContent, filePath, hcllib.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return errors.New("processFile", fmt.Errorf("failed to parse file as HCL: %v", diags))
	}
	formattedResult, err := hcl.SortAndFormatHCLFile(hclFile)
	if err != nil {
		if hcl.IsSortingError(err) {
			return errors.New("processFile", fmt.Errorf("sorting/formatting error in %s: %v", filePath, err))
		}
		return errors.New("processFile", fmt.Errorf("failed to sort/format %s: %v", filePath, err))
	}
	formatted := formattedResult

	// Safety check: don't write empty content
	if len(formatted) == 0 {
		return errors.New("processFile", fmt.Errorf("formatted content is empty for %s", filePath))
	}

	// Step 4: Compare
	if bytes.Equal(origContent, []byte(formatted)) {
		if config.Verbose {
			_, _ = successColor.Fprintf(stdout, "✅ No changes needed: %s\n", fileColor.Sprint(filePath))
		}
		return fmt.Errorf("%w: %s", errors.ErrNoChanges, filePath)
	}

	if config.DryRun {
		_, _ = warningColor.Fprintf(stdout, "📝 Would update: %s\n", fileColor.Sprint(filePath))
		printUnifiedDiff(string(origContent), formatted, filePath, stdout)
		return nil
	}

	if config.Validate {
		_, _ = warningColor.Fprintf(stdout, "⚠️  Needs update: %s\n", fileColor.Sprint(filePath))
		printUnifiedDiff(string(origContent), formatted, filePath, stdout)
		return errors.New("validate", fmt.Errorf("file needs update: %s", filePath))
	}

	// Step 5: Atomic write
	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(formatted), 0644); err != nil {
		return errors.New("processFile", fmt.Errorf("failed to write temp file: %v", err))
	}
	if err := os.Rename(tmpFile, filePath); err != nil {
		return errors.New("processFile", fmt.Errorf("failed to replace original file: %v", err))
	}
	_, _ = successColor.Fprintf(stdout, "✅ Updated: %s\n", fileColor.Sprint(filePath))
	return nil
}

// printUnifiedDiff prints a unified diff between two file contents
func printUnifiedDiff(a, b, filePath string, out io.Writer) {
	if a == b {
		_, _ = fmt.Fprintf(out, "(No changes)\n")
		return
	}

	// Split into lines for easier comparison
	linesA := strings.Split(a, "\n")
	linesB := strings.Split(b, "\n")

	_, _ = fmt.Fprintf(out, "--- %s (original)\n", filePath)
	_, _ = fmt.Fprintf(out, "+++ %s (formatted)\n", filePath)
	_, _ = fmt.Fprintf(out, "@@ Changes @@\n")

	// Simple line-by-line diff
	maxLines := len(linesA)
	if len(linesB) > maxLines {
		maxLines = len(linesB)
	}

	for i := 0; i < maxLines; i++ {
		lineA := ""
		lineB := ""

		if i < len(linesA) {
			lineA = linesA[i]
		}
		if i < len(linesB) {
			lineB = linesB[i]
		}

		if lineA != lineB {
			if lineA != "" {
				_, _ = fmt.Fprintf(out, "-%s\n", lineA)
			}
			if lineB != "" {
				_, _ = fmt.Fprintf(out, "+%s\n", lineB)
			}
		} else {
			// Show context (first few and last few lines)
			if i < 3 || i >= maxLines-3 {
				_, _ = fmt.Fprintf(out, " %s\n", lineA)
			} else if i == 3 && maxLines > 6 {
				_, _ = fmt.Fprintf(out, "...\n")
			}
		}
	}
}
