// Package cli provides the command-line interface for sortTF.
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

	"sorttf/config"
	"sorttf/internal/errors"
	"sorttf/internal/files"
	"sorttf/lib"

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

	// Process files (concurrently for better performance)
	processedCount, errorCount, _ := processFilesConcurrent(filePaths, config, stdout, stderr)

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

	// Use the library API to sort the file
	opts := lib.Options{
		DryRun:   config.DryRun,
		Validate: config.Validate,
	}

	err := lib.SortFile(filePath, opts)

	// Handle results based on error type
	if stderrors.Is(err, lib.ErrNoChanges) {
		// File is already sorted - not an error
		if config.Verbose {
			_, _ = successColor.Fprintf(stdout, "✅ No changes needed: %s\n", fileColor.Sprint(filePath))
		}
		return fmt.Errorf("%w: %s", errors.ErrNoChanges, filePath)
	}

	if stderrors.Is(err, lib.ErrNeedsSorting) {
		// Validate mode: file needs sorting, show diff
		_, _ = warningColor.Fprintf(stdout, "⚠️  Needs update: %s\n", fileColor.Sprint(filePath))

		// Get original and sorted content for diff
		origContent, _ := os.ReadFile(filePath)
		sortedContent, _, _ := lib.GetSortedContent(filePath)

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
			origContent, _ := os.ReadFile(filePath)
			sortedContent, _, _ := lib.GetSortedContent(filePath)

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

// fileResult holds the result of processing a single file
type fileResult struct {
	path          string
	err           error
	stdout        string
	stderr        string
	noChanges     bool
	processedFile bool
}

// processFilesConcurrent processes multiple files concurrently using a worker pool.
// It returns counts of processed files, errors, and files with no changes.
func processFilesConcurrent(filePaths []string, config *config.Config, stdout, stderr io.Writer) (int, int, int) {
	if len(filePaths) == 0 {
		return 0, 0, 0
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
	noChangesCount := 0

	for result := range results {
		// Write output atomically
		if result.stdout != "" {
			io.WriteString(stdout, result.stdout)
		}
		if result.stderr != "" {
			io.WriteString(stderr, result.stderr)
		}

		// Count results
		if result.noChanges {
			noChangesCount++
		} else if result.err != nil {
			errorCount++
			errors.PrintError(result.err, stderr)
		} else if result.processedFile {
			processedCount++
		}
	}

	return processedCount, errorCount, noChangesCount
}

// processFilesSerial processes files one at a time (used for small batches or verbose mode)
func processFilesSerial(filePaths []string, config *config.Config, stdout, stderr io.Writer) (int, int, int) {
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

	return processedCount, errorCount, noChangesCount
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
