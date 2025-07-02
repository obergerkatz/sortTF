package parsingutil

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
)

// Helper: check error presence
func assertError(t *testing.T, err error, shouldErr bool, checker func(error) bool) {
	t.Helper()
	if shouldErr && err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if !shouldErr && err != nil {
		t.Fatalf("Did not expect error, got %v", err)
	}
	if checker != nil && err != nil && !checker(err) {
		t.Fatalf("Error did not match checker, got: %T", err)
	}
}

// Helper: check diagnostics
func assertDiagnostics(t *testing.T, diags hcl.Diagnostics, shouldDiag bool) {
	t.Helper()
	if shouldDiag && !diags.HasErrors() {
		t.Errorf("Expected diagnostics errors, got none")
	}
	if !shouldDiag && diags.HasErrors() {
		t.Errorf("Expected no diagnostics errors, got: %v", diags)
	}
}

func TestParseHCLFileValid(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_Valid.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	if parsed.Diags.HasErrors() {
		t.Errorf("Expected no diagnostics errors, got: %v", parsed.Diags)
	}
}

func TestParseHCLFileInvalid(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_Invalid.tf")
	parsed, err := ParseHCLFile(path)
	if err == nil {
		t.Fatalf("Expected error for invalid HCL, got none")
	}
	if !IsHCLParseError(err) {
		t.Fatalf("Expected HCLParseError, got: %T", err)
	}
	if parsed == nil {
		t.Fatalf("Expected parsed file even with errors, got nil")
	}
	if !parsed.Diags.HasErrors() {
		t.Fatalf("Expected diagnostics errors, got none")
	}
}

func TestParseHCLFileNotExist(t *testing.T) {
	_, err := ParseHCLFile("/non/existent/file.tf")
	if err == nil {
		t.Fatalf("Expected error for non-existent file, got nil")
	}
	if !IsParsingError(err) {
		t.Fatalf("Expected ParsingError, got: %T", err)
	}
	if !IsNotExistError(err) {
		t.Fatalf("Expected not exist error, got: %v", err)
	}
}

func TestParseHCLFileEmptyPath(t *testing.T) {
	_, err := ParseHCLFile("")
	if err == nil {
		t.Fatalf("Expected error for empty path, got nil")
	}
	if !IsParsingError(err) {
		t.Fatalf("Expected ParsingError, got: %T", err)
	}
}

func TestParseHCLFileEmptyFile(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_EmptyFile.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no error for empty file, got: %v", err)
	}
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	if parsed.Diags.HasErrors() {
		t.Errorf("Expected no diagnostics errors, got: %v", parsed.Diags)
	}
}

func TestParseHCLFileWhitespaceOnly(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_WhitespaceOnly.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no error for whitespace-only file, got: %v", err)
	}
	if parsed == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	if parsed.Diags.HasErrors() {
		t.Errorf("Expected no diagnostics errors for whitespace-only file, got: %v", parsed.Diags)
	}
}

func TestParseHCLFileCommentOnly(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_CommentOnly.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no error for comment-only file, got: %v", err)
	}
	if parsed == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	if parsed.Diags.HasErrors() {
		t.Errorf("Expected no diagnostics errors for comment-only file, got: %v", parsed.Diags)
	}
}

func TestParseHCLFileMultipleResources(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_MultipleResources.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no error for multiple resources, got: %v", err)
	}
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	if parsed.Diags.HasErrors() {
		t.Errorf("Expected no diagnostics errors, got: %v", parsed.Diags)
	}
}

func TestParseHCLFileUnclosedString(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_UnclosedString.tf")
	parsed, err := ParseHCLFile(path)
	if err == nil {
		t.Fatalf("Expected error for unclosed string, got none")
	}
	if !IsHCLParseError(err) {
		t.Fatalf("Expected HCLParseError, got: %T", err)
	}
	if parsed == nil {
		t.Fatalf("Expected parsed file even with errors, got nil")
	}
}

func TestParseHCLFileOnlyBraces(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_OnlyBraces.tf")
	parsed, err := ParseHCLFile(path)
	if err == nil {
		t.Fatalf("Expected error for only braces, got none")
	}
	if !IsHCLParseError(err) {
		t.Fatalf("Expected HCLParseError, got: %T", err)
	}
	if parsed == nil {
		t.Fatalf("Expected parsed file even with errors, got nil")
	}
}

func TestParseHCLFileDeeplyNestedBlocks(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_DeeplyNestedBlocks.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no error for deeply nested blocks, got: %v", err)
	}
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	if parsed.Diags.HasErrors() {
		t.Errorf("Expected no diagnostics errors, got: %v", parsed.Diags)
	}
}

func TestParseHCLFileOnlyString(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_OnlyString.tf")
	parsed, err := ParseHCLFile(path)
	if err == nil {
		t.Fatalf("Expected error for only string, got none")
	}
	if !IsHCLParseError(err) {
		t.Fatalf("Expected HCLParseError, got: %T", err)
	}
	if parsed == nil {
		t.Fatalf("Expected parsed file even with errors, got nil")
	}
}

func TestParseHCLFileOnlyNumber(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_OnlyNumber.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, true, IsHCLParseError)
	if parsed == nil {
		t.Fatalf("Expected parsed file even with errors, got nil")
	}
}

func TestParseHCLFileUnicodeCharacters(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_UnicodeCharacters.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	assertDiagnostics(t, parsed.Diags, false)
}

func TestParseHCLFileWeirdIdentifiers(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_WeirdIdentifiers.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	assertDiagnostics(t, parsed.Diags, false)
}

func TestParseHCLFileAttributeWithoutValue(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_AttributeWithoutValue.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, true, IsHCLParseError)
	if parsed == nil {
		t.Fatalf("Expected parsed file even with errors, got nil")
	}
}

func TestParseHCLFileBlockWithoutLabel(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_BlockWithoutLabel.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	err = ValidateRequiredBlockLabels(parsed)
	assertError(t, err, true, IsValidationError)
}

func TestParseHCLFileOutputBlock(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_OutputBlock.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	assertDiagnostics(t, parsed.Diags, false)
}

func TestParseHCLFileProviderBlock(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_ProviderBlock.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	assertDiagnostics(t, parsed.Diags, false)
}

func TestParseHCLFileResourceBlock(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_ResourceBlock.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	assertDiagnostics(t, parsed.Diags, false)
}

func TestParseHCLFileTerraformBlock(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_TerraformBlock.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	assertDiagnostics(t, parsed.Diags, false)
}

func TestParseHCLFileVariableBlock(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_VariableBlock.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	assertDiagnostics(t, parsed.Diags, false)
}

func TestParseHCLFileModuleBlock(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_ModuleBlock.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	assertDiagnostics(t, parsed.Diags, false)
}

func TestParseHCLFileLocalsBlock(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_LocalsBlock.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	assertDiagnostics(t, parsed.Diags, false)
}

func TestParseHCLFileDataBlock(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_DataBlock.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	assertDiagnostics(t, parsed.Diags, false)
}

func TestParseHCLFileBackendBlock(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_BackendBlock.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	assertDiagnostics(t, parsed.Diags, false)
}

func TestParseHCLFileNestedBlocks(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_NestedBlocks.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	assertDiagnostics(t, parsed.Diags, false)
	err = ValidateRequiredBlockLabels(parsed)
	assertError(t, err, false, nil)
}

func TestParseHCLFileExtraLabels(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_ExtraLabels.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	err = ValidateRequiredBlockLabels(parsed)
	assertError(t, err, true, IsValidationError)
}

func TestParseHCLFileProviderWithExtraLabels(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_ProviderWithExtraLabels.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	err = ValidateRequiredBlockLabels(parsed)
	assertError(t, err, true, nil)
}

func TestParseHCLFileInvalidBackend(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_InvalidBackend.tf")
	parsed, err := ParseHCLFile(path)
	assertError(t, err, false, nil)
	err = ValidateRequiredBlockLabels(parsed)
	assertError(t, err, true, nil)
}

func TestParseHCLFileComplexNestedBlocks(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_ComplexNestedBlocks.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no error for complex nested blocks, got: %v", err)
	}
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	if parsed.Diags.HasErrors() {
		t.Errorf("Expected no diagnostics errors, got: %v", parsed.Diags)
	}
	err = ValidateRequiredBlockLabels(parsed)
	if err != nil {
		t.Errorf("Expected no validation errors, got: %v", err)
	}
}

func TestParseHCLFileResourceWithNoBody(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_ResourceWithNoBody.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no error for resource with no body, got: %v", err)
	}
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	if parsed.Diags.HasErrors() {
		t.Errorf("Expected no diagnostics errors, got: %v", parsed.Diags)
	}
	err = ValidateRequiredBlockLabels(parsed)
	if err != nil {
		t.Errorf("Expected no validation errors, got: %v", err)
	}
}

func TestParseHCLFileBackendWithExtraLabelsInsideTerraform(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_BackendWithExtraLabelsInsideTerraform.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no parse error, got: %v", err)
	}
	err = ValidateRequiredBlockLabels(parsed)
	if err == nil {
		t.Fatalf("Expected validation error for backend with extra labels inside terraform, got none")
	}
}

func TestParseHCLFileResourceWithOneLabel(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_ResourceWithOneLabel.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no parse error, got: %v", err)
	}
	if parsed == nil || parsed.File == nil {
		t.Fatalf("Expected parsed file, got nil")
	}
	if parsed.Diags.HasErrors() {
		t.Errorf("Expected no diagnostics errors, got: %v", parsed.Diags)
	}
}

func TestParseHCLFileModuleWithNoLabels(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_ModuleWithNoLabels.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no parse error, got: %v", err)
	}
	err = ValidateRequiredBlockLabels(parsed)
	if err == nil {
		t.Fatalf("Expected validation error for module with no labels, got none")
	}
}

func TestParseHCLFileModuleWithTwoLabels(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_ModuleWithTwoLabels.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no parse error, got: %v", err)
	}
	err = ValidateRequiredBlockLabels(parsed)
	if err == nil {
		t.Fatalf("Expected validation error for module with two labels, got none")
	}
}

func TestParseHCLFileVariableWithNoLabels(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_VariableWithNoLabels.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no parse error, got: %v", err)
	}
	err = ValidateRequiredBlockLabels(parsed)
	if err == nil {
		t.Fatalf("Expected validation error for variable with no labels, got none")
	}
}

func TestParseHCLFileOutputWithNoLabels(t *testing.T) {
	path := filepath.Join("testdata", "TestParseHCLFile_OutputWithNoLabels.tf")
	parsed, err := ParseHCLFile(path)
	if err != nil {
		t.Fatalf("Expected no parse error, got: %v", err)
	}
	err = ValidateRequiredBlockLabels(parsed)
	if err == nil {
		t.Fatalf("Expected validation error for output with no labels, got none")
	}
}

// Test error types and helper functions
func TestParsingError(t *testing.T) {
	err := &ParsingError{
		Op:   "TestOp",
		Path: "/test/path",
		Err:  fmt.Errorf("test error"),
	}

	expectedMsg := "parsingutil TestOp /test/path: test error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	if err.Unwrap().Error() != "test error" {
		t.Errorf("Expected unwrapped error 'test error', got '%s'", err.Unwrap().Error())
	}

	if !IsParsingError(err) {
		t.Error("IsParsingError should return true for ParsingError")
	}
	if GetParsingErrorOp(err) != "TestOp" {
		t.Error("GetParsingErrorOp should return the operation")
	}
	if GetParsingErrorPath(err) != "/test/path" {
		t.Error("GetParsingErrorPath should return the path")
	}
}

func TestHCLParseError(t *testing.T) {
	diags := hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid syntax",
			Detail:   "Expected closing brace",
		},
	}
	err := &HCLParseError{
		Path:  "/test/path",
		Diags: diags,
	}

	expectedMsg := "HCL parsing failed for /test/path: <nil>: Invalid syntax; Expected closing brace"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	if !IsHCLParseError(err) {
		t.Error("IsHCLParseError should return true for HCLParseError")
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Op:   "TestOp",
		Path: "/test/path",
		Err:  fmt.Errorf("test error"),
	}

	expectedMsg := "validation TestOp /test/path: test error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	if err.Unwrap().Error() != "test error" {
		t.Errorf("Expected unwrapped error 'test error', got '%s'", err.Unwrap().Error())
	}

	if !IsValidationError(err) {
		t.Error("IsValidationError should return true for ValidationError")
	}
	if GetValidationErrorOp(err) != "TestOp" {
		t.Error("GetValidationErrorOp should return the operation")
	}
	if GetValidationErrorPath(err) != "/test/path" {
		t.Error("GetValidationErrorPath should return the path")
	}
}

func TestErrorHelperFunctions(t *testing.T) {
	// Test IsParsingError
	parsingErr := &ParsingError{Op: "Test", Path: "/test", Err: fmt.Errorf("test")}
	if !IsParsingError(parsingErr) {
		t.Error("IsParsingError should return true for ParsingError")
	}
	if IsParsingError(fmt.Errorf("regular error")) {
		t.Error("IsParsingError should return false for regular error")
	}

	// Test IsHCLParseError
	hclErr := &HCLParseError{Path: "/test", Diags: hcl.Diagnostics{}}
	if !IsHCLParseError(hclErr) {
		t.Error("IsHCLParseError should return true for HCLParseError")
	}

	// Test IsValidationError
	validationErr := &ValidationError{Op: "Test", Path: "/test", Err: fmt.Errorf("test")}
	if !IsValidationError(validationErr) {
		t.Error("IsValidationError should return true for ValidationError")
	}

	// Test IsNotExistError
	notExistErr := &ParsingError{Op: "Test", Err: fmt.Errorf("file does not exist")}
	if !IsNotExistError(notExistErr) {
		t.Error("IsNotExistError should return true for not exist error")
	}

	// Test IsPermissionError
	permErr := &ParsingError{Op: "Test", Err: fmt.Errorf("permission denied")}
	if !IsPermissionError(permErr) {
		t.Error("IsPermissionError should return true for permission error")
	}

	// Test GetParsingErrorPath
	if GetParsingErrorPath(parsingErr) != "/test" {
		t.Error("GetParsingErrorPath should return the path")
	}
	if GetParsingErrorPath(fmt.Errorf("regular error")) != "" {
		t.Error("GetParsingErrorPath should return empty string for non-ParsingError")
	}

	// Test GetParsingErrorOp
	if GetParsingErrorOp(parsingErr) != "Test" {
		t.Error("GetParsingErrorOp should return the operation")
	}
	if GetParsingErrorOp(fmt.Errorf("regular error")) != "" {
		t.Error("GetParsingErrorOp should return empty string for non-ParsingError")
	}

	// Test GetValidationErrorOp
	if GetValidationErrorOp(validationErr) != "Test" {
		t.Error("GetValidationErrorOp should return the operation")
	}
	if GetValidationErrorOp(fmt.Errorf("regular error")) != "" {
		t.Error("GetValidationErrorOp should return empty string for non-ValidationError")
	}

	// Test GetValidationErrorPath
	if GetValidationErrorPath(validationErr) != "/test" {
		t.Error("GetValidationErrorPath should return the path")
	}
	if GetValidationErrorPath(fmt.Errorf("regular error")) != "" {
		t.Error("GetValidationErrorPath should return empty string for non-ValidationError")
	}
}

func TestParseHCLFile_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		expectErr    bool
		expectDiag   bool
		errorChecker func(error) bool
	}{
		{"valid", "TestParseHCLFile_Valid.tf", false, false, nil},
		{"invalid", "TestParseHCLFile_Invalid.tf", true, true, IsHCLParseError},
		{"empty", "TestParseHCLFile_EmptyFile.tf", false, false, nil},
		{"whitespace only", "TestParseHCLFile_WhitespaceOnly.tf", false, false, nil},
		{"comment only", "TestParseHCLFile_CommentOnly.tf", false, false, nil},
		{"multiple resources", "TestParseHCLFile_MultipleResources.tf", false, false, nil},
		{"unclosed string", "TestParseHCLFile_UnclosedString.tf", true, true, IsHCLParseError},
		{"only braces", "TestParseHCLFile_OnlyBraces.tf", true, true, IsHCLParseError},
		{"deeply nested", "TestParseHCLFile_DeeplyNestedBlocks.tf", false, false, nil},
		{"only string", "TestParseHCLFile_OnlyString.tf", true, true, IsHCLParseError},
		{"only number", "TestParseHCLFile_OnlyNumber.tf", true, true, IsHCLParseError},
		{"unicode", "TestParseHCLFile_UnicodeCharacters.tf", false, false, nil},
		{"weird identifiers", "TestParseHCLFile_WeirdIdentifiers.tf", false, false, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.filename)
			parsed, err := ParseHCLFile(path)
			assertError(t, err, tt.expectErr, tt.errorChecker)
			if parsed == nil {
				t.Fatalf("Expected parsed file, got nil")
			}
			assertDiagnostics(t, parsed.Diags, tt.expectDiag)
		})
	}
}

func TestParseHCLFile_ErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		checkErr func(error) bool
	}{
		{"not exist", "/non/existent/file.tf", IsNotExistError},
		{"empty path", "", IsParsingError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseHCLFile(tt.path)
			if err == nil {
				t.Fatalf("Expected error, got nil")
			}
			if !tt.checkErr(err) {
				t.Fatalf("Error did not match checker, got: %T", err)
			}
		})
	}
}
