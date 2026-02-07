// Package cli provides the command-line interface for sortTF.
//
// This package implements the main CLI execution logic including:
//   - Command-line argument parsing via config package
//   - File discovery and validation
//   - Concurrent file processing with worker pools
//   - Colorized output and error reporting
//   - Unified diff generation for dry-run mode
//
// The main entry point is RunCLI which handles all execution modes:
// normal, dry-run, validate, and verbose.
package cli

import (
	"bytes"
	stderrors "errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"sorttf/api"
	"sorttf/config"
	"sorttf/internal/errors"
	"sorttf/internal/files"

	"github.com/fatih/color"
)

// Color configuration
var (
	errorColor   = color.New(color.FgRed, color.Bold)
	warningColor = color.New(color.FgYellow, color.Bold)
	successColor = color.New(color.FgGreen, color.Bold)
	infoColor    = color.New(color.FgBlue, color.Bold)
	fileColor    = color.New(color.FgCyan)
)

// RunCLI is the main entry point for CLI execution.
// It parses command-line arguments, discovers files, and processes them
// according to the requested mode (normal, dry-run, validate, etc.).
//
// Returns an exit code:
//   - 0: Success (all files processed successfully)
//   - 1: Error during processing or validation failures
//   - 2: Invalid command-line arguments
func RunCLI(args []string) int {
	return RunCLIWithWriters(args, os.Stdout, os.Stderr)
}

// RunCLIWithWriters executes the CLI with custom output writers.
// This is primarily used for testing to capture and verify output.
// The behavior is otherwise identical to RunCLI.
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

// runMainLogic executes the main CLI logic after argument parsing.
// It validates paths, discovers files, and coordinates their processing.
// Returns an exit code suitable for os.Exit.
func runMainLogic(config *config.Config, stdout, stderr io.Writer) int {
	// Check if the path is a file or directory
	fileInfo, err := os.Stat(config.Root)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			_, _ = errorColor.Fprintf(stderr, "❌ Path '%s' does not exist\n", fileColor.Sprint(config.Root))
		case os.IsPermission(err):
			_, _ = errorColor.Fprintf(stderr, "🔒 Permission denied accessing '%s'\n", fileColor.Sprint(config.Root))
		default:
			_, _ = errorColor.Fprintf(stderr, "❌ Error accessing '%s': %v\n", fileColor.Sprint(config.Root), err)
		}
		return 1
	}

	var filePaths []string

	if fileInfo.IsDir() {
		// It's a directory - validate and find files
		if err := files.ValidateDirectoryPath(config.Root); err != nil {
			switch {
			case files.IsNotExistError(err):
				_, _ = errorColor.Fprintf(stderr, "❌ Directory '%s' does not exist\n", fileColor.Sprint(config.Root))
			case files.IsPermissionError(err):
				_, _ = errorColor.Fprintf(stderr, "🔒 Permission denied accessing directory '%s'\n", fileColor.Sprint(config.Root))
			default:
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

			switch {
			case files.IsNotExistError(err):
				_, _ = errorColor.Fprintf(stderr, "❌ Path '%s' does not exist\n", fileColor.Sprint(path))
			case files.IsPermissionError(err):
				_, _ = errorColor.Fprintf(stderr, "🔒 Permission denied accessing '%s'\n", fileColor.Sprint(path))
			default:
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

	// Process files (concurrently for better performance)
	processedCount, errorCount := processFilesConcurrent(filePaths, config, stdout, stderr)

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

// isSupportedFile checks if the file has a supported extension (.tf or .hcl).
// Returns true for Terraform and Terragrunt files, false otherwise.
func isSupportedFile(filePath string) bool {
	ext := filepath.Ext(filePath)
	return ext == ".tf" || ext == ".hcl"
}

// processFile handles sorting and formatting of a single file.
// It uses the api.SortFile API and handles different modes (normal, dry-run, validate).
// Returns nil on success, errors.ErrNoChanges if file is already sorted,
// or an error if processing fails.
func processFile(filePath string, config *config.Config, stdout, _ io.Writer) error {
	if config.Verbose {
		_, _ = infoColor.Fprintf(stdout, "🔄 Processing: %s\n", fileColor.Sprint(filePath))
	}

	// Use the library API to sort the file
	opts := api.Options{
		DryRun:   config.DryRun,
		Validate: config.Validate,
	}

	err := api.SortFile(filePath, opts)

	// Handle results based on error type
	if stderrors.Is(err, api.ErrNoChanges) {
		// File is already sorted - not an error
		if config.Verbose {
			_, _ = successColor.Fprintf(stdout, "✅ No changes needed: %s\n", fileColor.Sprint(filePath))
		}
		return fmt.Errorf("%w: %s", errors.ErrNoChanges, filePath)
	}

	if stderrors.Is(err, api.ErrNeedsSorting) {
		// Validate mode: file needs sorting, show diff
		_, _ = warningColor.Fprintf(stdout, "⚠️  Needs update: %s\n", fileColor.Sprint(filePath))

		// Get original and sorted content for diff
		//nolint:gosec // G304: User input expected for file tool
		origContent, _ := os.ReadFile(filePath)
		sortedContent, _, _ := api.GetSortedContent(filePath)

		if origContent != nil && sortedContent != "" {
			printUnifiedDiff(string(origContent), sortedContent, filePath, stdout)
		}

		return errors.New("validate", fmt.Errorf("file needs update: %s", filePath))
	}

	if err == nil {
		// Success - check which mode
		if config.DryRun {
			// Dry-run mode: show what would change
			_, _ = warningColor.Fprintf(stdout, "📝 Would update: %s\n", fileColor.Sprint(filePath))

			// Get original and sorted content for diff
			//nolint:gosec // G304: File path comes from user input, which is expected for a file processing tool
			origContent, _ := os.ReadFile(filePath)
			sortedContent, _, _ := api.GetSortedContent(filePath)

			if origContent != nil && sortedContent != "" {
				printUnifiedDiff(string(origContent), sortedContent, filePath, stdout)
			}

			return nil
		}

		// Normal mode: file was actually written
		_, _ = successColor.Fprintf(stdout, "✅ Updated: %s\n", fileColor.Sprint(filePath))
		return nil
	}

	// Some other error occurred
	return errors.New("processFile", fmt.Errorf("failed to process %s: %w", filePath, err))
}

// fileResult holds the result of processing a single file in concurrent mode.
// It contains the file path, any error that occurred, buffered output,
// and flags indicating the result type (no changes needed, successfully processed, etc.).
type fileResult struct {
	path          string // Path of the processed file
	err           error  // Error that occurred, or nil
	stdout        string // Buffered standard output
	stderr        string // Buffered standard error
	noChanges     bool   // True if file was already sorted
	processedFile bool   // True if file was successfully processed/modified
}

// processFilesConcurrent processes multiple files concurrently using a worker pool.
// It automatically determines worker count based on CPU cores (2x for I/O overlap).
// Falls back to serial processing for < 4 files or when verbose mode is enabled.
//
// Returns (processedCount, errorCount) where:
//   - processedCount: files that were successfully sorted/modified
//   - errorCount: files that encountered errors
func processFilesConcurrent(filePaths []string, config *config.Config, stdout, stderr io.Writer) (int, int) {
	if len(filePaths) == 0 {
		return 0, 0
	}

	// Determine worker count: min(numFiles, numCPU * 2)
	// Use 2x CPU count to keep CPUs busy during I/O
	numWorkers := runtime.NumCPU() * 2
	if numWorkers > len(filePaths) {
		numWorkers = len(filePaths)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	// For small numbers of files or verbose mode, process serially to maintain order
	if len(filePaths) < 4 || config.Verbose {
		return processFilesSerial(filePaths, config, stdout, stderr)
	}

	// Create channels
	jobs := make(chan string, len(filePaths))
	results := make(chan fileResult, len(filePaths))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				// Process file with buffered output
				var outBuf, errBuf bytes.Buffer
				err := processFile(path, config, &outBuf, &errBuf)

				// Send result
				result := fileResult{
					path:   path,
					err:    err,
					stdout: outBuf.String(),
					stderr: errBuf.String(),
				}

				if err == nil {
					result.processedFile = true
				} else if stderrors.Is(err, errors.ErrNoChanges) {
					result.noChanges = true
				}

				results <- result
			}
		}()
	}

	// Send jobs
	for _, path := range filePaths {
		jobs <- path
	}
	close(jobs)

	// Wait for workers in background
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results and write output in order received
	// (order doesn't matter for performance, but makes output readable)
	processedCount := 0
	errorCount := 0

	for result := range results {
		// Write output atomically
		if result.stdout != "" {
			_, _ = io.WriteString(stdout, result.stdout) // Ignore write errors to stdout
		}
		if result.stderr != "" {
			_, _ = io.WriteString(stderr, result.stderr) // Ignore write errors to stderr
		}

		// Count results
		//nolint:revive,gocritic // empty-block and ifElseChain: This structure is clearer than alternatives
		if result.noChanges {
			// Already sorted, don't count as error or processed
		} else if result.err != nil {
			errorCount++
			errors.PrintError(result.err, stderr)
		} else if result.processedFile {
			processedCount++
		}
	}

	return processedCount, errorCount
}

// processFilesSerial processes files sequentially one at a time.
// This is used for small batches (< 4 files) or when verbose mode is enabled
// to maintain readable output order. Returns the same counts as processFilesConcurrent.
func processFilesSerial(filePaths []string, config *config.Config, stdout, stderr io.Writer) (int, int) {
	processedCount := 0
	errorCount := 0

	for _, filePath := range filePaths {
		if err := processFile(filePath, config, stdout, stderr); err != nil {
			//nolint:revive // empty-block: Explicit no-op for clarity
			if stderrors.Is(err, errors.ErrNoChanges) {
				// Already sorted, don't count as error or processed
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

	return processedCount, errorCount
}

// printUnifiedDiff prints a unified diff between original and formatted content.
// It displays the changes in a readable format with +/- prefixes for added/removed lines.
// Shows context lines (first and last 3 lines) to help locate changes.
// Used in dry-run and validate modes to show what would change.
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
