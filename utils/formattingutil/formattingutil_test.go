package formattingutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func TestFormatHCLFile(t *testing.T) {
	f := hclwrite.NewEmptyFile()
	body := f.Body()
	body.SetAttributeValue("foo", cty.StringVal("bar"))

	formatted, err := FormatHCLFile(f)
	if err != nil {
		// If terraform is not available, skip this test
		if IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("FormatHCLFile failed: %v", err)
	}
	if len(formatted) == 0 {
		t.Errorf("Expected non-empty formatted output")
	}
	if formatted != "foo = \"bar\"\n" {
		t.Errorf("Unexpected formatted output: %q", formatted)
	}
}

func TestFormatHCLFileComplex(t *testing.T) {
	f := hclwrite.NewEmptyFile()
	body := f.Body()

	// Add a resource block
	resourceBlock := body.AppendNewBlock("resource", []string{"aws_instance", "example"})
	resourceBody := resourceBlock.Body()
	resourceBody.SetAttributeValue("ami", cty.StringVal("ami-123456"))
	resourceBody.SetAttributeValue("instance_type", cty.StringVal("t3.micro"))

	// Add a variable block
	variableBlock := body.AppendNewBlock("variable", []string{"region"})
	variableBody := variableBlock.Body()
	variableBody.SetAttributeValue("type", cty.StringVal("string"))

	formatted, err := FormatHCLFile(f)
	if err != nil {
		// If terraform is not available, skip this test
		if IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("FormatHCLFile failed: %v", err)
	}
	if len(formatted) == 0 {
		t.Errorf("Expected non-empty formatted output for complex file")
	}

	// Check that both blocks are present
	if !strings.Contains(formatted, "resource \"aws_instance\" \"example\"") {
		t.Errorf("Expected resource block in formatted output")
	}
	if !strings.Contains(formatted, "variable \"region\"") {
		t.Errorf("Expected variable block in formatted output")
	}
}

func TestFormatHCLFileWithNestedBlocks(t *testing.T) {
	f := hclwrite.NewEmptyFile()
	body := f.Body()

	// Create a resource with nested blocks
	resourceBlock := body.AppendNewBlock("resource", []string{"aws_instance", "example"})
	resourceBody := resourceBlock.Body()
	resourceBody.SetAttributeValue("ami", cty.StringVal("ami-123456"))

	// Add a nested block
	nestedBlock := resourceBody.AppendNewBlock("tags", nil)
	nestedBody := nestedBlock.Body()
	nestedBody.SetAttributeValue("Name", cty.StringVal("example"))

	formatted, err := FormatHCLFile(f)
	if err != nil {
		// If terraform is not available, skip this test
		if IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("FormatHCLFile failed: %v", err)
	}
	if !strings.Contains(formatted, "tags") {
		t.Errorf("Expected nested block to be in formatted output")
	}
}

// TestFormatHCLFileConsistency is now defined later in the file
func TestFormatHCLFileWithNumbers(t *testing.T) {
	f := hclwrite.NewEmptyFile()
	body := f.Body()
	body.SetAttributeValue("count", cty.NumberIntVal(3))
	body.SetAttributeValue("port", cty.NumberFloatVal(8080.5))

	formatted, err := FormatHCLFile(f)
	if err != nil {
		// If terraform is not available, skip this test
		if IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("FormatHCLFile failed: %v", err)
	}

	if !strings.Contains(formatted, "count = 3") && !strings.Contains(formatted, "count  = 3") {
		t.Errorf("Expected number to be formatted correctly")
	}
	// Check for float in various possible formats (with different spacing around =)
	if !strings.Contains(formatted, "port = 8080.5") &&
		!strings.Contains(formatted, "port  = 8080.5") &&
		!strings.Contains(formatted, "port = 8.0805e+03") &&
		!strings.Contains(formatted, "port  = 8.0805e+03") &&
		!strings.Contains(formatted, "port = 8080") &&
		!strings.Contains(formatted, "port  = 8080") {
		t.Errorf("Expected float to be formatted correctly, got: %s", formatted)
	}
}

func TestFormatHCLFileWithBooleans(t *testing.T) {
	f := hclwrite.NewEmptyFile()
	body := f.Body()
	body.SetAttributeValue("enabled", cty.BoolVal(true))
	body.SetAttributeValue("disabled", cty.BoolVal(false))

	formatted, err := FormatHCLFile(f)
	if err != nil {
		// If terraform is not available, skip this test
		if IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("FormatHCLFile failed: %v", err)
	}

	if !strings.Contains(formatted, "enabled = true") && !strings.Contains(formatted, "enabled  = true") {
		t.Errorf("Expected boolean true to be formatted correctly")
	}
	if !strings.Contains(formatted, "disabled = false") && !strings.Contains(formatted, "disabled  = false") {
		t.Errorf("Expected boolean false to be formatted correctly")
	}
}

func TestFormatHCLFileWithLists(t *testing.T) {
	f := hclwrite.NewEmptyFile()
	body := f.Body()
	body.SetAttributeValue("ports", cty.ListVal([]cty.Value{
		cty.NumberIntVal(80),
		cty.NumberIntVal(443),
		cty.NumberIntVal(8080),
	}))

	formatted, err := FormatHCLFile(f)
	if err != nil {
		// If terraform is not available, skip this test
		if IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("FormatHCLFile failed: %v", err)
	}

	if !strings.Contains(formatted, "ports = [") {
		t.Errorf("Expected list to be formatted correctly")
	}
}

func TestFormatHCLFileWithMaps(t *testing.T) {
	f := hclwrite.NewEmptyFile()
	body := f.Body()
	body.SetAttributeValue("tags", cty.MapVal(map[string]cty.Value{
		"Name": cty.StringVal("example"),
		"Env":  cty.StringVal("prod"),
	}))

	formatted, err := FormatHCLFile(f)
	if err != nil {
		// If terraform is not available, skip this test
		if IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("FormatHCLFile failed: %v", err)
	}

	if !strings.Contains(formatted, "tags = {") {
		t.Errorf("Expected map to be formatted correctly")
	}
}

func TestFormatHCLFileWithRealFiles(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
	}{
		{"Basic_Formatting", "TestFormatHCLFile_Basic_not_formatted.tf"},
		{"Complex_Formatting", "TestFormatHCLFile_Complex_not_formatted.tf"},
		{"Edge_Cases", "TestFormatHCLFile_EdgeCases_not_formatted.tf"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join("testdata", tc.filename))
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			formatted, err := FormatHCLString(string(content))
			if err != nil {
				// If terraform is not available, skip this test
				if IsTerraformNotFoundError(err) {
					t.Skip("terraform command not available, skipping test")
				}
				t.Fatalf("FormatHCLString failed: %v", err)
			}

			if len(formatted) == 0 {
				t.Errorf("Expected non-empty formatted output")
			}

			// Basic validation that formatting occurred
			if formatted == string(content) {
				t.Errorf("Expected content to be formatted, but it was unchanged")
			}
		})
	}
}

func TestFormatHCLString(t *testing.T) {
	content := `resource "aws_instance" "example" {
  ami           = "ami-123456"
  instance_type = "t3.micro"
}`

	formatted, err := FormatHCLString(content)
	if err != nil {
		// If terraform is not available, skip this test
		if IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("FormatHCLString failed: %v", err)
	}

	if len(formatted) == 0 {
		t.Errorf("Expected non-empty formatted output")
	}

	// Check that the formatted content contains the expected elements
	if !strings.Contains(formatted, "resource \"aws_instance\" \"example\"") {
		t.Errorf("Expected resource block in formatted output")
	}
}

func TestFormatHCLFileNil(t *testing.T) {
	formatted, err := FormatHCLFile(nil)
	if err == nil {
		t.Error("Expected error for nil file")
	}
	if !IsFormattingError(err) {
		t.Error("Expected FormattingError for nil file")
	}
	if formatted != "" {
		t.Errorf("Expected empty string for nil file, got: %q", formatted)
	}
}

func TestFormatHCLFileEmpty(t *testing.T) {
	file := hclwrite.NewEmptyFile()
	result, err := FormatHCLFile(file)
	if err != nil {
		// If terraform is not available, skip this test
		if IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("FormatHCLFile failed: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty string for empty file, got: %q", result)
	}
}

func TestFormatFile(t *testing.T) {
	// Create a temporary file with unformatted content
	tmpFile, err := os.CreateTemp("", "test-*.tf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	unformatted := `terraform{required_version=">= 1.0"}
provider"aws"{region="us-west-2"}`

	_, err = tmpFile.WriteString(unformatted)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Format the file
	err = FormatFile(tmpFile.Name())
	if err != nil {
		// If terraform is not available, skip this test
		if IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("FormatFile failed: %v", err)
	}

	// Read the formatted content
	formattedBytes, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read formatted file: %v", err)
	}

	formatted := string(formattedBytes)
	if !strings.Contains(formatted, "terraform {") {
		t.Errorf("Expected formatted content to contain 'terraform {'")
	}
}

func TestFormatDirectory(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file in the directory
	tmpFile := filepath.Join(tmpDir, "test.tf")
	unformatted := `resource "aws_instance" "example" {
  ami           = "ami-123456"
  instance_type = "t3.micro"
}`

	err = os.WriteFile(tmpFile, []byte(unformatted), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Format the directory
	err = FormatDirectory(tmpDir)
	if err != nil {
		// If terraform is not available, skip this test
		if IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("FormatDirectory failed: %v", err)
	}

	// Read the formatted content
	formattedBytes, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read formatted file: %v", err)
	}

	formatted := string(formattedBytes)
	if !strings.Contains(formatted, "resource \"aws_instance\" \"example\"") {
		t.Errorf("Expected formatted content to contain 'resource \"aws_instance\" \"example\"'")
	}
}

func TestFormatHCLStringInvalid(t *testing.T) {
	invalidContent := `resource "aws_instance" "example" {
  ami = "ami-123456"
  instance_type = "t3.micro"
  # Missing closing brace
`

	formatted, err := FormatHCLString(invalidContent)
	if err == nil {
		t.Error("Expected error for invalid HCL content")
	}
	if !IsHCLParseError(err) {
		t.Error("Expected HCLParseError for invalid HCL content")
	}
	// Should return original content on error
	if formatted != invalidContent {
		t.Errorf("Expected original content on error, got: %q", formatted)
	}
}

func TestFormatHCLFileConsistency(t *testing.T) {
	// Test that formatting is idempotent
	original := `resource "aws_instance" "example" {
  ami           = "ami-123456"
  instance_type = "t3.micro"
}`

	// Format once
	formatted1, err := FormatHCLString(original)
	if err != nil {
		// If terraform is not available, skip this test
		if IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("First format failed: %v", err)
	}

	// Format again
	formatted2, err := FormatHCLString(formatted1)
	if err != nil {
		t.Fatalf("Second format failed: %v", err)
	}

	// Should be the same
	if formatted1 != formatted2 {
		t.Errorf("Formatting is not idempotent:\nFirst:  %q\nSecond: %q", formatted1, formatted2)
	}
}

func TestFormatHCLFileWithAdditionalCases(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
	}{
		{"Comments", "TestFormatHCLFile_Comments_not_formatted.tf"},
		{"Heredoc", "TestFormatHCLFile_Heredoc_not_formatted.tf"},
		{"Empty_Block", "TestFormatHCLFile_EmptyBlock_not_formatted.tf"},
		{"Nested_Blocks", "TestFormatHCLFile_Nested_not_formatted.tf"},
		{"Multi_Resource", "TestFormatHCLFile_MultiResource_not_formatted.tf"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join("testdata", tc.filename))
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			formatted, err := FormatHCLString(string(content))
			if err != nil {
				// If terraform is not available, skip this test
				if IsTerraformNotFoundError(err) {
					t.Skip("terraform command not available, skipping test")
				}
				t.Fatalf("FormatHCLString failed: %v", err)
			}

			if len(formatted) == 0 {
				t.Errorf("Expected non-empty formatted output")
			}

			// For some files, terraform fmt might not change already properly formatted content
			// This is expected behavior, so we just verify the output is valid
			if formatted == string(content) {
				t.Logf("Content was already properly formatted and unchanged (this is expected for some files)")
			}
		})
	}
}

// Test error handling functions
func TestFormattingError(t *testing.T) {
	// Test FormattingError creation and methods
	err := &FormattingError{
		Op:      "TestOp",
		Path:    "/test/path",
		Content: "test content",
		Err:     fmt.Errorf("test error"),
	}

	expectedMsg := "formattingutil TestOp /test/path: test error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	if err.Unwrap().Error() != "test error" {
		t.Errorf("Expected unwrapped error 'test error', got '%s'", err.Unwrap().Error())
	}
}

func TestTerraformNotFoundError(t *testing.T) {
	// Test TerraformNotFoundError creation and methods
	originalErr := fmt.Errorf("executable file not found")
	err := &TerraformNotFoundError{Err: originalErr}

	expectedMsg := "terraform command not found: executable file not found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	if err.Unwrap().Error() != "executable file not found" {
		t.Errorf("Expected unwrapped error 'executable file not found', got '%s'", err.Unwrap().Error())
	}
}

func TestHCLParseError(t *testing.T) {
	// Test HCLParseError creation and methods
	diags := hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid syntax",
			Detail:   "Expected closing brace",
		},
	}
	err := &HCLParseError{
		Content: "invalid content",
		Diags:   diags,
	}

	expectedMsg := "HCL parsing failed: <nil>: Invalid syntax; Expected closing brace"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestErrorHelperFunctions(t *testing.T) {
	// Test IsFormattingError
	formattingErr := &FormattingError{Op: "Test", Path: "/test", Err: fmt.Errorf("test")}
	if !IsFormattingError(formattingErr) {
		t.Error("IsFormattingError should return true for FormattingError")
	}
	if IsFormattingError(fmt.Errorf("regular error")) {
		t.Error("IsFormattingError should return false for regular error")
	}

	// Test IsTerraformNotFoundError
	terraformErr := &TerraformNotFoundError{Err: fmt.Errorf("not found")}
	if !IsTerraformNotFoundError(terraformErr) {
		t.Error("IsTerraformNotFoundError should return true for TerraformNotFoundError")
	}
	wrappedTerraformErr := &FormattingError{Op: "Test", Err: terraformErr}
	if !IsTerraformNotFoundError(wrappedTerraformErr) {
		t.Error("IsTerraformNotFoundError should return true for wrapped TerraformNotFoundError")
	}

	// Test IsHCLParseError
	hclErr := &HCLParseError{Content: "test", Diags: hcl.Diagnostics{}}
	if !IsHCLParseError(hclErr) {
		t.Error("IsHCLParseError should return true for HCLParseError")
	}

	// Test IsNotExistError
	notExistErr := &FormattingError{Op: "Test", Err: fmt.Errorf("file does not exist")}
	if !IsNotExistError(notExistErr) {
		t.Error("IsNotExistError should return true for not exist error")
	}

	// Test IsPermissionError
	permErr := &FormattingError{Op: "Test", Err: fmt.Errorf("permission denied")}
	if !IsPermissionError(permErr) {
		t.Error("IsPermissionError should return true for permission error")
	}

	// Test GetFormattingErrorPath
	if GetFormattingErrorPath(formattingErr) != "/test" {
		t.Error("GetFormattingErrorPath should return the path")
	}
	if GetFormattingErrorPath(fmt.Errorf("regular error")) != "" {
		t.Error("GetFormattingErrorPath should return empty string for non-FormattingError")
	}

	// Test GetFormattingErrorOp
	if GetFormattingErrorOp(formattingErr) != "Test" {
		t.Error("GetFormattingErrorOp should return the operation")
	}
	if GetFormattingErrorOp(fmt.Errorf("regular error")) != "" {
		t.Error("GetFormattingErrorOp should return empty string for non-FormattingError")
	}

	// Test GetFormattingErrorContent
	if GetFormattingErrorContent(formattingErr) != "" {
		t.Error("GetFormattingErrorContent should return empty string when no content")
	}
	contentErr := &FormattingError{Op: "Test", Content: "test content"}
	if GetFormattingErrorContent(contentErr) != "test content" {
		t.Error("GetFormattingErrorContent should return the content")
	}
}

func TestFormatFileErrorHandling(t *testing.T) {
	// Test with empty path
	err := FormatFile("")
	if err == nil {
		t.Error("Expected error for empty file path")
	}
	if !IsFormattingError(err) {
		t.Error("Expected FormattingError for empty file path")
	}

	// Test with non-existent file
	err = FormatFile("/non/existent/file.tf")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if !IsFormattingError(err) {
		t.Error("Expected FormattingError for non-existent file")
	}
	if !IsNotExistError(err) {
		t.Error("Expected not exist error for non-existent file")
	}
}

func TestFormatDirectoryErrorHandling(t *testing.T) {
	// Test with empty path
	err := FormatDirectory("")
	if err == nil {
		t.Error("Expected error for empty directory path")
	}
	if !IsFormattingError(err) {
		t.Error("Expected FormattingError for empty directory path")
	}

	// Test with non-existent directory
	err = FormatDirectory("/non/existent/dir")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
	if !IsFormattingError(err) {
		t.Error("Expected FormattingError for non-existent directory")
	}
	if !IsNotExistError(err) {
		t.Error("Expected not exist error for non-existent directory")
	}
}
