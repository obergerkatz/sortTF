package cliutil

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sorttf/utils/fileutil"
	"sorttf/utils/formattingutil"
	"sorttf/utils/parsingutil"
	"sorttf/utils/sortingutil"
	"strings"

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

// CLIError represents an error during CLI execution
type CLIError struct {
	Op  string
	Err error
}

func (e *CLIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("cliutil %s: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("cliutil %s", e.Op)
}

func (e *CLIError) Unwrap() error {
	return e.Err
}

// Config holds all CLI configuration
type Config struct {
	Root      string
	Recursive bool
	DryRun    bool
	Verbose   bool
	Validate  bool
}

// NoChangesError indicates no changes are needed for a file
type NoChangesError struct {
	FilePath string
}

func (e *NoChangesError) Error() string {
	return fmt.Sprintf("no changes needed for %s", e.FilePath)
}

// RunCLI is the main entry point for CLI execution
func RunCLI(args []string) int {
	return RunCLIWithWriters(args, os.Stdout, os.Stderr)
}

// RunCLIWithWriters allows testing by providing custom writers
func RunCLIWithWriters(args []string, stdout, stderr io.Writer) int {
	config, err := parseFlags(args, stderr)
	if err != nil {
		if cliErr, ok := err.(*CLIError); ok && cliErr.Op == "help" {
			// Help was requested, just exit with success
			return 0
		}
		printError(err, stderr)
		return 2 // Usage error
	}

	return runMainLogic(config, stdout, stderr)
}

// parseFlags parses command line arguments and returns a Config
func parseFlags(args []string, stderr io.Writer) (*Config, error) {
	fs := flag.NewFlagSet("sorttf", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Suppress default error output

	var config Config

	fs.BoolVar(&config.Recursive, "recursive", false, "Scan directories recursively")
	fs.BoolVar(&config.DryRun, "dry-run", false, "Show what would be changed without writing (shows a unified diff)")
	fs.BoolVar(&config.Verbose, "verbose", false, "Print detailed logs about which files were parsed, sorted, and formatted")
	fs.BoolVar(&config.Validate, "validate", false, "Exit with a non-zero code if any files are not sorted/formatted")

	// Custom usage function
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: sorttf [flags] [path]\n")
		fmt.Fprintf(stderr, "\nSort and format Terraform (.tf) and Terragrunt (.hcl) files for consistency and readability.\n")
		fmt.Fprintf(stderr, "\nPath can be a file or directory. If no path is provided, the current directory is used.\n")
		fmt.Fprintf(stderr, "\nFlags:\n")

		// Create a temporary buffer to capture flag output
		var flagOutput bytes.Buffer
		fs.SetOutput(&flagOutput)
		fs.PrintDefaults()
		fs.SetOutput(io.Discard) // Reset to discard

		fmt.Fprintf(stderr, "%s", flagOutput.String())
		fmt.Fprintf(stderr, "\nExamples:\n")
		fmt.Fprintf(stderr, "  sorttf .                    # Sort and format files in current directory\n")
		fmt.Fprintf(stderr, "  sorttf main.tf              # Sort and format a specific file\n")
		fmt.Fprintf(stderr, "  sorttf --recursive .        # Recursively process subdirectories\n")
		fmt.Fprintf(stderr, "  sorttf --validate .         # Check if files are properly sorted/formatted\n")
		fmt.Fprintf(stderr, "  sorttf --dry-run .          # Show what would change, with a unified diff\n")
	}

	if err := fs.Parse(args); err != nil {
		// Don't treat help request as an error
		if err.Error() == "flag: help requested" {
			return nil, &CLIError{
				Op:  "help",
				Err: nil,
			}
		}
		return nil, &CLIError{
			Op:  "parseFlags",
			Err: err,
		}
	}

	// Get positional arguments
	positionalArgs := fs.Args()
	if len(positionalArgs) > 1 {
		return nil, &CLIError{
			Op:  "parseFlags",
			Err: fmt.Errorf("too many arguments provided"),
		}
	}

	// Set root directory
	if len(positionalArgs) == 0 {
		config.Root = "."
	} else {
		config.Root = positionalArgs[0]
	}

	return &config, nil
}

// runMainLogic executes the main CLI logic
func runMainLogic(config *Config, stdout, stderr io.Writer) int {
	// Check if the path is a file or directory
	fileInfo, err := os.Stat(config.Root)
	if err != nil {
		if os.IsNotExist(err) {
			errorColor.Fprintf(stderr, "❌ Path '%s' does not exist\n", fileColor.Sprint(config.Root))
		} else if os.IsPermission(err) {
			errorColor.Fprintf(stderr, "🔒 Permission denied accessing '%s'\n", fileColor.Sprint(config.Root))
		} else {
			errorColor.Fprintf(stderr, "❌ Error accessing '%s': %v\n", fileColor.Sprint(config.Root), err)
		}
		return 1
	}

	var files []string

	if fileInfo.IsDir() {
		// It's a directory - validate and find files
		if err := fileutil.ValidateDirectoryPath(config.Root); err != nil {
			if fileutil.IsNotExistError(err) {
				errorColor.Fprintf(stderr, "❌ Directory '%s' does not exist\n", fileColor.Sprint(config.Root))
			} else if fileutil.IsPermissionError(err) {
				errorColor.Fprintf(stderr, "🔒 Permission denied accessing directory '%s'\n", fileColor.Sprint(config.Root))
			} else {
				errorColor.Fprintf(stderr, "❌ Error validating directory '%s': %v\n", fileColor.Sprint(config.Root), err)
			}
			return 1
		}

		// Find files to process
		files, err = fileutil.FindFiles(config.Root, config.Recursive)
		if err != nil {
			if fileutil.IsNotExistError(err) {
				errorColor.Fprintf(stderr, "❌ Path '%s' does not exist\n", fileColor.Sprint(fileutil.GetFileUtilErrorPath(err)))
			} else if fileutil.IsPermissionError(err) {
				errorColor.Fprintf(stderr, "🔒 Permission denied accessing '%s'\n", fileColor.Sprint(fileutil.GetFileUtilErrorPath(err)))
			} else {
				errorColor.Fprintf(stderr, "❌ Error finding files: %v\n", err)
			}
			return 1
		}
	} else {
		// It's a file - check if it's a supported file type
		if !isSupportedFile(config.Root) {
			errorColor.Fprintf(stderr, "❌ File '%s' is not a supported file type (.tf or .hcl)\n", fileColor.Sprint(config.Root))
			return 1
		}
		files = []string{config.Root}
	}

	if len(files) == 0 {
		infoColor.Fprintf(stdout, "ℹ️  No Terraform or Terragrunt files found.\n")
		return 0
	}

	if config.Verbose {
		infoColor.Fprintf(stdout, "📁 Found %d files:\n", len(files))
		for _, f := range files {
			fmt.Fprintf(stdout, "   %s\n", fileColor.Sprint(f))
		}
	}

	// Process files
	processedCount := 0
	errorCount := 0
	noChangesCount := 0

	for _, filePath := range files {
		if err := processFile(filePath, config, stdout, stderr); err != nil {
			if _, ok := err.(*NoChangesError); ok {
				noChangesCount++
			} else {
				errorCount++
				printError(err, stderr)
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
			successColor.Fprintf(stdout, "✅ Processed %d files, no changes needed\n", len(files))
		} else {
			infoColor.Fprintf(stdout, "📊 Processed %d files, %d would be updated\n", len(files), processedCount)
		}
	} else {
		successColor.Fprintf(stdout, "✅ Processed %d files\n", processedCount)
	}

	if errorCount > 0 {
		errorColor.Fprintf(stderr, "❌ Encountered %d errors\n", errorCount)
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
func processFile(filePath string, config *Config, stdout, stderr io.Writer) error {
	if config.Verbose {
		infoColor.Fprintf(stdout, "🔄 Processing: %s\n", fileColor.Sprint(filePath))
	}

	// Step 1: Read original file content
	origContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return &CLIError{
			Op:  "processFile",
			Err: fmt.Errorf("failed to read file: %v", err),
		}
	}

	// Step 2: Parse and validate
	parsed, err := parsingutil.ParseHCLFile(filePath)
	if err != nil {
		if parsingutil.IsNotExistError(err) {
			return &CLIError{
				Op:  "processFile",
				Err: fmt.Errorf("file not found: %s", filePath),
			}
		} else if parsingutil.IsHCLParseError(err) {
			return &CLIError{
				Op:  "processFile",
				Err: fmt.Errorf("syntax error in %s: %v", filePath, err),
			}
		} else if parsingutil.IsParsingError(err) {
			return &CLIError{
				Op:  "processFile",
				Err: fmt.Errorf("parsing error in %s: %v", filePath, err),
			}
		}
		return &CLIError{
			Op:  "processFile",
			Err: fmt.Errorf("failed to parse %s: %v", filePath, err),
		}
	}
	if err := parsingutil.ValidateRequiredBlockLabels(parsed); err != nil {
		if parsingutil.IsValidationError(err) {
			return &CLIError{
				Op:  "processFile",
				Err: fmt.Errorf("validation error in %s: %v", filePath, err),
			}
		}
		return &CLIError{
			Op:  "processFile",
			Err: fmt.Errorf("validation failed for %s: %v", filePath, err),
		}
	}

	// Step 3: Sort and format
	hclFile, diags := hclwrite.ParseConfig(origContent, filePath, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return &CLIError{
			Op:  "processFile",
			Err: fmt.Errorf("failed to parse file as HCL: %v", diags),
		}
	}
	formattedResult, err := sortingutil.SortAndFormatHCLFile(hclFile)
	if err != nil {
		if sortingutil.IsSortingError(err) {
			return &CLIError{
				Op:  "processFile",
				Err: fmt.Errorf("sorting/formatting error in %s: %v", filePath, err),
			}
		}
		return &CLIError{
			Op:  "processFile",
			Err: fmt.Errorf("failed to sort/format %s: %v", filePath, err),
		}
	}
	formatted := formattedResult

	// Safety check: don't write empty content
	if len(formatted) == 0 {
		return &CLIError{
			Op:  "processFile",
			Err: fmt.Errorf("formatted content is empty for %s", filePath),
		}
	}

	// Step 4: Compare
	if bytes.Equal(origContent, []byte(formatted)) {
		if config.Verbose {
			successColor.Fprintf(stdout, "✅ No changes needed: %s\n", fileColor.Sprint(filePath))
		}
		return &NoChangesError{FilePath: filePath}
	}

	if config.DryRun {
		warningColor.Fprintf(stdout, "📝 Would update: %s\n", fileColor.Sprint(filePath))
		printUnifiedDiff(string(origContent), formatted, filePath, stdout)
		return nil
	}

	if config.Validate {
		warningColor.Fprintf(stdout, "⚠️  Needs update: %s\n", fileColor.Sprint(filePath))
		printUnifiedDiff(string(origContent), formatted, filePath, stdout)
		return &CLIError{Op: "validate", Err: fmt.Errorf("file needs update: %s", filePath)}
	}

	// Step 5: Atomic write
	tmpFile := filePath + ".tmp"
	if err := ioutil.WriteFile(tmpFile, []byte(formatted), 0644); err != nil {
		return &CLIError{
			Op:  "processFile",
			Err: fmt.Errorf("failed to write temp file: %v", err),
		}
	}
	if err := os.Rename(tmpFile, filePath); err != nil {
		return &CLIError{
			Op:  "processFile",
			Err: fmt.Errorf("failed to replace original file: %v", err),
		}
	}
	successColor.Fprintf(stdout, "✅ Updated: %s\n", fileColor.Sprint(filePath))
	return nil
}

// printError prints a formatted error message with color and context
func printError(err error, stderr io.Writer) {
	if err == nil {
		return
	}

	// Determine error type and format accordingly
	switch {
	case isFileNotFoundError(err):
		printFileNotFoundError(err, stderr)
	case isPermissionError(err):
		printPermissionError(err, stderr)
	case isValidationError(err):
		printValidationError(err, stderr)
	case isParsingError(err):
		printParsingError(err, stderr)
	case isFormattingError(err):
		printFormattingError(err, stderr)
	case isSortingError(err):
		printSortingError(err, stderr)
	default:
		printGenericError(err, stderr)
	}
}

// Helper functions to identify error types
func isFileNotFoundError(err error) bool {
	return fileutil.IsNotExistError(err) ||
		strings.Contains(err.Error(), "does not exist") ||
		strings.Contains(err.Error(), "file not found")
}

func isPermissionError(err error) bool {
	return fileutil.IsPermissionError(err) ||
		strings.Contains(err.Error(), "permission denied") ||
		strings.Contains(err.Error(), "Permission denied")
}

func isValidationError(err error) bool {
	return parsingutil.IsValidationError(err) ||
		strings.Contains(err.Error(), "validation error") ||
		strings.Contains(err.Error(), "validation failed")
}

func isParsingError(err error) bool {
	return parsingutil.IsParsingError(err) ||
		parsingutil.IsHCLParseError(err) ||
		strings.Contains(err.Error(), "syntax error") ||
		strings.Contains(err.Error(), "parsing error")
}

func isFormattingError(err error) bool {
	return formattingutil.IsFormattingError(err) ||
		formattingutil.IsTerraformNotFoundError(err) ||
		strings.Contains(err.Error(), "formatting error")
}

func isSortingError(err error) bool {
	return sortingutil.IsSortingError(err) ||
		strings.Contains(err.Error(), "sorting error")
}

// Specific error printing functions
func printFileNotFoundError(err error, stderr io.Writer) {
	filePath := extractFilePath(err)
	if filePath != "" {
		errorColor.Fprintf(stderr, "❌ File not found: %s\n", fileColor.Sprint(filePath))
		infoColor.Fprintf(stderr, "   Make sure the file exists and the path is correct.\n")
	} else {
		errorColor.Fprintf(stderr, "❌ File not found: %v\n", err)
	}
}

func printPermissionError(err error, stderr io.Writer) {
	filePath := extractFilePath(err)
	if filePath != "" {
		errorColor.Fprintf(stderr, "🔒 Permission denied: %s\n", fileColor.Sprint(filePath))
		infoColor.Fprintf(stderr, "   Check file permissions or run with appropriate privileges.\n")
	} else {
		errorColor.Fprintf(stderr, "🔒 Permission denied: %v\n", err)
	}
}

func printValidationError(err error, stderr io.Writer) {
	filePath := extractFilePath(err)
	if filePath != "" {
		errorColor.Fprintf(stderr, "⚠️  Validation error in %s:\n", fileColor.Sprint(filePath))
	} else {
		errorColor.Fprintf(stderr, "⚠️  Validation error:\n")
	}

	// Extract and format the validation message
	msg := extractErrorMessage(err)
	if msg != "" {
		fmt.Fprintf(stderr, "   %s\n", msg)
	} else {
		fmt.Fprintf(stderr, "   %v\n", err)
	}
}

func printParsingError(err error, stderr io.Writer) {
	filePath := extractFilePath(err)
	if filePath != "" {
		errorColor.Fprintf(stderr, "🔍 Syntax error in %s:\n", fileColor.Sprint(filePath))
	} else {
		errorColor.Fprintf(stderr, "🔍 Syntax error:\n")
	}

	msg := extractErrorMessage(err)
	if msg != "" {
		fmt.Fprintf(stderr, "   %s\n", msg)
	} else {
		fmt.Fprintf(stderr, "   %v\n", err)
	}

	infoColor.Fprintf(stderr, "   Check for missing quotes, brackets, or invalid HCL syntax.\n")
}

func printFormattingError(err error, stderr io.Writer) {
	filePath := extractFilePath(err)
	if filePath != "" {
		errorColor.Fprintf(stderr, "🎨 Formatting error in %s:\n", fileColor.Sprint(filePath))
	} else {
		errorColor.Fprintf(stderr, "🎨 Formatting error:\n")
	}

	msg := extractErrorMessage(err)
	if msg != "" {
		fmt.Fprintf(stderr, "   %s\n", msg)
	} else {
		fmt.Fprintf(stderr, "   %v\n", err)
	}

	// Check if it's a terraform not found error
	if formattingutil.IsTerraformNotFoundError(err) {
		infoColor.Fprintf(stderr, "   Make sure 'terraform' is installed and available in your PATH.\n")
	}
}

func printSortingError(err error, stderr io.Writer) {
	filePath := extractFilePath(err)
	if filePath != "" {
		errorColor.Fprintf(stderr, "📊 Sorting error in %s:\n", fileColor.Sprint(filePath))
	} else {
		errorColor.Fprintf(stderr, "📊 Sorting error:\n")
	}

	msg := extractErrorMessage(err)
	if msg != "" {
		fmt.Fprintf(stderr, "   %s\n", msg)
	} else {
		fmt.Fprintf(stderr, "   %v\n", err)
	}
}

func printGenericError(err error, stderr io.Writer) {
	errorColor.Fprintf(stderr, "❌ Error: %v\n", err)
}

// Helper functions to extract information from errors
func extractFilePath(err error) string {
	errStr := err.Error()

	// Look for file paths in common error patterns
	patterns := []string{
		"failed to read file:",
		"failed to parse",
		"error in",
		"validation error in",
		"syntax error in",
		"formatting error in",
		"sorting error in",
	}

	for _, pattern := range patterns {
		if idx := strings.Index(errStr, pattern); idx != -1 {
			// Extract the file path after the pattern
			afterPattern := errStr[idx+len(pattern):]
			// Clean up the path (remove extra text, quotes, etc.)
			path := strings.TrimSpace(afterPattern)
			path = strings.Trim(path, " :")
			path = strings.Trim(path, `"'`)
			return path
		}
	}

	// Try to extract from specific error types
	if fileutil.IsNotExistError(err) {
		return fileutil.GetFileUtilErrorPath(err)
	}

	return ""
}

func extractErrorMessage(err error) string {
	errStr := err.Error()

	// Remove common prefixes to get the core message
	prefixes := []string{
		"failed to read file:",
		"failed to parse",
		"error in",
		"validation error in",
		"syntax error in",
		"formatting error in",
		"sorting error in",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(errStr, prefix) {
			msg := strings.TrimPrefix(errStr, prefix)
			msg = strings.TrimSpace(msg)
			msg = strings.Trim(msg, " :")
			return msg
		}
	}

	return errStr
}

// printSuccess prints a success message with color
func printSuccess(format string, args ...interface{}) {
	successColor.Printf(format, args...)
}

// printInfo prints an info message with color
func printInfo(format string, args ...interface{}) {
	infoColor.Printf(format, args...)
}

// printWarning prints a warning message with color
func printWarning(format string, args ...interface{}) {
	warningColor.Printf(format, args...)
}

// printFile prints a file path with color
func printFile(format string, args ...interface{}) {
	fileColor.Printf(format, args...)
}

// Error helper functions

// IsCLIError checks if an error is a CLIError
func IsCLIError(err error) bool {
	_, ok := err.(*CLIError)
	return ok
}

// GetCLIErrorOp extracts the operation from a CLIError
func GetCLIErrorOp(err error) string {
	if cliErr, ok := err.(*CLIError); ok {
		return cliErr.Op
	}
	return ""
}

func printUnifiedDiff(a, b, filePath string, out io.Writer) {
	if a == b {
		fmt.Fprintf(out, "(No changes)\n")
		return
	}

	// Split into lines for easier comparison
	linesA := strings.Split(a, "\n")
	linesB := strings.Split(b, "\n")

	fmt.Fprintf(out, "--- %s (original)\n", filePath)
	fmt.Fprintf(out, "+++ %s (formatted)\n", filePath)
	fmt.Fprintf(out, "@@ Changes @@\n")

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
				fmt.Fprintf(out, "-%s\n", lineA)
			}
			if lineB != "" {
				fmt.Fprintf(out, "+%s\n", lineB)
			}
		} else {
			// Show context (first few and last few lines)
			if i < 3 || i >= maxLines-3 {
				fmt.Fprintf(out, " %s\n", lineA)
			} else if i == 3 && maxLines > 6 {
				fmt.Fprintf(out, "...\n")
			}
		}
	}
}
