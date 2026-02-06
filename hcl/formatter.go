package hcl

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Custom error types for better error handling
type FormattingError struct {
	Op      string
	Path    string
	Content string
	Err     error
}

func (e *FormattingError) Error() string {
	if e.Err != nil {
		if e.Path != "" {
			return fmt.Sprintf("formattingutil %s %s: %v", e.Op, e.Path, e.Err)
		}
		return fmt.Sprintf("formattingutil %s: %v", e.Op, e.Err)
	}
	if e.Path != "" {
		return fmt.Sprintf("formattingutil %s %s", e.Op, e.Path)
	}
	return fmt.Sprintf("formattingutil %s", e.Op)
}

func (e *FormattingError) Unwrap() error {
	return e.Err
}

// TerraformNotFoundError indicates terraform command is not available
type TerraformNotFoundError struct {
	Err error
}

func (e *TerraformNotFoundError) Error() string {
	return fmt.Sprintf("terraform command not found: %v", e.Err)
}

func (e *TerraformNotFoundError) Unwrap() error {
	return e.Err
}

// Note: HCLParseError is defined in parser.go

// FormatHCLFile takes an hclwrite.File and returns the formatted string
// using terraform fmt standards
func FormatHCLFile(file *hclwrite.File) (string, error) {
	if file == nil {
		return "", &FormattingError{
			Op:  "FormatHCLFile",
			Err: fmt.Errorf("nil file provided"),
		}
	}

	// Get the raw formatted bytes from hclwrite
	rawBytes := file.Bytes()

	// Apply terraform fmt formatting
	formatted, err := applyTerraformFmt(string(rawBytes))
	if err != nil {
		return string(rawBytes), &FormattingError{
			Op:      "FormatHCLFile",
			Content: string(rawBytes),
			Err:     err,
		}
	}

	return formatted, nil
}

// applyTerraformFmt applies terraform fmt formatting to HCL content
func applyTerraformFmt(content string) (string, error) {
	if content == "" {
		return "", nil
	}

	// Check if terraform is available
	if err := checkTerraformAvailable(); err != nil {
		return content, err
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "sorttf-*.tf")
	if err != nil {
		return content, &FormattingError{
			Op:      "applyTerraformFmt",
			Content: content,
			Err:     fmt.Errorf("failed to create temporary file: %v", err),
		}
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	defer func() { _ = tmpFile.Close() }()

	// Write content to temp file
	_, err = tmpFile.WriteString(content)
	if err != nil {
		return content, &FormattingError{
			Op:      "applyTerraformFmt",
			Content: content,
			Err:     fmt.Errorf("failed to write to temporary file: %v", err),
		}
	}

	// Run terraform fmt on the temp file
	cmd := exec.Command("terraform", "fmt", tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return content, &FormattingError{
			Op:      "applyTerraformFmt",
			Content: content,
			Err:     fmt.Errorf("terraform fmt failed: %v\nOutput: %s", err, string(output)),
		}
	}

	// Read the formatted content back
	formattedBytes, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return content, &FormattingError{
			Op:      "applyTerraformFmt",
			Content: content,
			Err:     fmt.Errorf("failed to read formatted file: %v", err),
		}
	}

	return string(formattedBytes), nil
}

// FormatHCLString formats a raw HCL string using terraform fmt
func FormatHCLString(content string) (string, error) {
	if content == "" {
		return "", nil
	}

	// Parse the content first to validate it
	file, diags := hclwrite.ParseConfig([]byte(content), "input", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return content, &HCLParseError{
			Path:  "input",
			Diags: diags,
		}
	}

	return FormatHCLFile(file)
}

// FormatFile formats an existing file using terraform fmt
func FormatFile(filePath string) error {
	if filePath == "" {
		return &FormattingError{
			Op:  "FormatFile",
			Err: fmt.Errorf("empty file path provided"),
		}
	}

	// Validate file exists and is accessible
	if err := validateFilePath(filePath); err != nil {
		return &FormattingError{
			Op:   "FormatFile",
			Path: filePath,
			Err:  err,
		}
	}

	// Check if terraform is available
	if err := checkTerraformAvailable(); err != nil {
		return &FormattingError{
			Op:   "FormatFile",
			Path: filePath,
			Err:  err,
		}
	}

	// Run terraform fmt on the file
	cmd := exec.Command("terraform", "fmt", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &FormattingError{
			Op:   "FormatFile",
			Path: filePath,
			Err:  fmt.Errorf("terraform fmt failed: %v\nOutput: %s", err, string(output)),
		}
	}

	return nil
}

// FormatDirectory formats all .tf files in a directory using terraform fmt
func FormatDirectory(dirPath string) error {
	if dirPath == "" {
		return &FormattingError{
			Op:  "FormatDirectory",
			Err: fmt.Errorf("empty directory path provided"),
		}
	}

	// Validate directory exists and is accessible
	if err := validateDirectoryPath(dirPath); err != nil {
		return &FormattingError{
			Op:   "FormatDirectory",
			Path: dirPath,
			Err:  err,
		}
	}

	// Check if terraform is available
	if err := checkTerraformAvailable(); err != nil {
		return &FormattingError{
			Op:   "FormatDirectory",
			Path: dirPath,
			Err:  err,
		}
	}

	// Run terraform fmt on the directory
	cmd := exec.Command("terraform", "fmt", dirPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &FormattingError{
			Op:   "FormatDirectory",
			Path: dirPath,
			Err:  fmt.Errorf("terraform fmt failed: %v\nOutput: %s", err, string(output)),
		}
	}

	return nil
}

// Helper functions

// checkTerraformAvailable checks if terraform command is available
func checkTerraformAvailable() error {
	_, err := exec.LookPath("terraform")
	if err != nil {
		return &TerraformNotFoundError{Err: err}
	}
	return nil
}

// Note: validateFilePath is defined in parser.go

// validateDirectoryPath checks if a directory path is valid and accessible
func validateDirectoryPath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path provided")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist")
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied")
		}
		return fmt.Errorf("failed to access directory: %v", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is a file, expected a directory")
	}

	return nil
}

// Error checking helper functions

// IsFormattingError checks if an error is a FormattingError
func IsFormattingError(err error) bool {
	_, ok := err.(*FormattingError)
	return ok
}

// IsTerraformNotFoundError checks if the error indicates terraform command is not found
func IsTerraformNotFoundError(err error) bool {
	if _, ok := err.(*TerraformNotFoundError); ok {
		return true
	}
	if formattingErr, ok := err.(*FormattingError); ok {
		_, ok = formattingErr.Err.(*TerraformNotFoundError)
		return ok
	}
	return false
}

// Note: IsHCLParseError, IsNotExistError, IsPermissionError are defined in parser.go

// GetFormattingErrorPath extracts the path from a FormattingError
func GetFormattingErrorPath(err error) string {
	if formattingErr, ok := err.(*FormattingError); ok {
		return formattingErr.Path
	}
	return ""
}

// GetFormattingErrorOp extracts the operation from a FormattingError
func GetFormattingErrorOp(err error) string {
	if formattingErr, ok := err.(*FormattingError); ok {
		return formattingErr.Op
	}
	return ""
}

// GetFormattingErrorContent extracts the content from a FormattingError
func GetFormattingErrorContent(err error) string {
	if formattingErr, ok := err.(*FormattingError); ok {
		return formattingErr.Content
	}
	return ""
}
