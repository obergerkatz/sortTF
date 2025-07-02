package cliutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sorttf/utils/fileutil"
	"sorttf/utils/parsingutil"
	"sorttf/utils/sortingutil"
	"strings"

	"sorttf/utils/argsutil"


	"github.com/fatih/color"
	"github.com/hashicorp/hcl/v2"
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
	config, err := argsutil.ParseFlags(args, stderr)
	if err != nil {
		if err.Error() == "help" {
			// Help was requested, just exit with success
			return 0
		}
		errorutil.PrintError(err, stderr)
		return 2 // Usage error
	}

	return runMainLogic(config, stdout, stderr)
}

// runMainLogic executes the main CLI logic
func runMainLogic(config *argsutil.Config, stdout, stderr io.Writer) int {
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

	var files []string

	if fileInfo.IsDir() {
		// It's a directory - validate and find files
		if err := fileutil.ValidateDirectoryPath(config.Root); err != nil {
			if fileutil.IsNotExistError(err) {
				_, _ = errorColor.Fprintf(stderr, "❌ Directory '%s' does not exist\n", fileColor.Sprint(config.Root))
			} else if fileutil.IsPermissionError(err) {
				_, _ = errorColor.Fprintf(stderr, "🔒 Permission denied accessing directory '%s'\n", fileColor.Sprint(config.Root))
			} else {
				_, _ = errorColor.Fprintf(stderr, "❌ Error validating directory '%s': %v\n", fileColor.Sprint(config.Root), err)
			}
			return 1
		}

		// Find files to process
		files, err = fileutil.FindFiles(config.Root, config.Recursive)
		if err != nil {
			if fileutil.IsNotExistError(err) {
				_, _ = errorColor.Fprintf(stderr, "❌ Path '%s' does not exist\n", fileColor.Sprint(fileutil.GetFileUtilErrorPath(err)))
			} else if fileutil.IsPermissionError(err) {
				_, _ = errorColor.Fprintf(stderr, "🔒 Permission denied accessing '%s'\n", fileColor.Sprint(fileutil.GetFileUtilErrorPath(err)))
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
		files = []string{config.Root}
	}

	if len(files) == 0 {
		_, _ = infoColor.Fprintf(stdout, "ℹ️  No Terraform or Terragrunt files found.\n")
		return 0
	}

	if config.Verbose {
		_, _ = infoColor.Fprintf(stdout, "📁 Found %d files:\n", len(files))
		for _, f := range files {
			_, _ = fmt.Fprintf(stdout, "   %s\n", fileColor.Sprint(f))
		}
	}

	// Process files
	processedCount := 0
	errorCount := 0
	noChangesCount := 0

	for _, filePath := range files {
		if err := processFile(filePath, config, stdout, stderr); err != nil {
			if _, ok := err.(*errorutil.NoChangesError); ok {
				noChangesCount++
			} else {
				errorCount++
				errorutil.PrintError(err, stderr)
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
			_, _ = successColor.Fprintf(stdout, "✅ Processed %d files, no changes needed\n", len(files))
		} else {
			_, _ = infoColor.Fprintf(stdout, "📊 Processed %d files, %d would be updated\n", len(files), processedCount)
		}
	} else {
		_, _ = successColor.Fprintf(stdout, "✅ Processed %d files\n", processedCount)
	}

	if errorCount > 0 {
		_, _ = errorColor.Fprintf(stderr, "❌ Encountered %d errors\n", errorCount)
		if config.Validate {
			return 1
		}
	}

	return 0
}

// isSupportedFile checks if the file has a supported extension
func isSupportedFile(filePath string) bool {
	ext := filepath.Ext(filePath)
	return ext == ".tf" || ext == ".hcl"
}

// processFile handles a single file
func processFile(filePath string, config *argsutil.Config, stdout, stderr io.Writer) error {
	if config.Verbose {
		_, _ = infoColor.Fprintf(stdout, "🔄 Processing: %s\n", fileColor.Sprint(filePath))
	}

	// Step 1: Read original file content
	origContent, err := os.ReadFile(filePath)
	if err != nil {
		return errorutil.NewCLIError("processFile", fmt.Errorf("failed to read file: %v", err))
	}

	// Step 2: Parse and validate
	parsed, err := parsingutil.ParseHCLFile(filePath)
	if err != nil {
		if parsingutil.IsNotExistError(err) {
			return errorutil.NewCLIError("processFile", fmt.Errorf("file not found: %s", filePath))
		} else if parsingutil.IsHCLParseError(err) {
			return errorutil.NewCLIError("processFile", fmt.Errorf("syntax error in %s: %v", filePath, err))
		} else if parsingutil.IsParsingError(err) {
			return errorutil.NewCLIError("processFile", fmt.Errorf("parsing error in %s: %v", filePath, err))
		}
		return errorutil.NewCLIError("processFile", fmt.Errorf("failed to parse %s: %v", filePath, err))
	}
	if err := parsingutil.ValidateRequiredBlockLabels(parsed); err != nil {
		if parsingutil.IsValidationError(err) {
			return errorutil.NewCLIError("processFile", fmt.Errorf("validation error in %s: %v", filePath, err))
		}
		return errorutil.NewCLIError("processFile", fmt.Errorf("validation failed for %s: %v", filePath, err))
	}

	// Step 3: Sort and format
	hclFile, diags := hclwrite.ParseConfig(origContent, filePath, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return errorutil.NewCLIError("processFile", fmt.Errorf("failed to parse file as HCL: %v", diags))
	}
	formattedResult, err := sortingutil.SortAndFormatHCLFile(hclFile)
	if err != nil {
		if sortingutil.IsSortingError(err) {
			return errorutil.NewCLIError("processFile", fmt.Errorf("sorting/formatting error in %s: %v", filePath, err))
		}
		return errorutil.NewCLIError("processFile", fmt.Errorf("failed to sort/format %s: %v", filePath, err))
	}
	formatted := formattedResult

	// Safety check: don't write empty content
	if len(formatted) == 0 {
		return errorutil.NewCLIError("processFile", fmt.Errorf("formatted content is empty for %s", filePath))
	}

	// Step 4: Compare
	if bytes.Equal(origContent, []byte(formatted)) {
		if config.Verbose {
			_, _ = successColor.Fprintf(stdout, "✅ No changes needed: %s\n", fileColor.Sprint(filePath))
		}
		return errorutil.NewNoChangesError(filePath)
	}

	if config.DryRun {
		_, _ = warningColor.Fprintf(stdout, "📝 Would update: %s\n", fileColor.Sprint(filePath))
		printUnifiedDiff(string(origContent), formatted, filePath, stdout)
		return nil
	}

	if config.Validate {
		_, _ = warningColor.Fprintf(stdout, "⚠️  Needs update: %s\n", fileColor.Sprint(filePath))
		printUnifiedDiff(string(origContent), formatted, filePath, stdout)
		return errorutil.NewCLIError("validate", fmt.Errorf("file needs update: %s", filePath))
	}

	// Step 5: Atomic write
	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(formatted), 0644); err != nil {
		return errorutil.NewCLIError("processFile", fmt.Errorf("failed to write temp file: %v", err))
	}
	if err := os.Rename(tmpFile, filePath); err != nil {
		return errorutil.NewCLIError("processFile", fmt.Errorf("failed to replace original file: %v", err))
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
