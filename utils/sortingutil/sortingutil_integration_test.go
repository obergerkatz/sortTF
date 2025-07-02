package sortingutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"sorttf/utils/formattingutil"
)

func TestSortHCLFileWithRealFiles(t *testing.T) {
	testCases := []struct {
		name         string
		unsortedFile string
		expectedFile string
		description  string
	}{
		{
			name:         "Complex Configuration",
			unsortedFile: "testdata/TestSortHCLFile_Complex_unsorted.tf",
			expectedFile: "testdata/TestSortHCLFile_Complex_sorted.tf",
			description:  "Test sorting of a complex Terraform configuration with multiple block types",
		},
		{
			name:         "Resource Labels",
			unsortedFile: "testdata/TestSortHCLFile_ResourceLabels_unsorted.tf",
			expectedFile: "testdata/TestSortHCLFile_ResourceLabels_sorted.tf",
			description:  "Test sorting of resources by type and then by labels",
		},
		{
			name:         "Attributes",
			unsortedFile: "testdata/TestSortHCLFile_Attributes_unsorted.tf",
			expectedFile: "testdata/TestSortHCLFile_Attributes_sorted.tf",
			description:  "Test sorting of attributes within blocks",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Read the unsorted file
			unsortedContent, err := os.ReadFile(tc.unsortedFile)
			if err != nil {
				t.Fatalf("Failed to read unsorted file %s: %v", tc.unsortedFile, err)
			}

			// Parse the unsorted file
			file, diags := hclwrite.ParseConfig(unsortedContent, tc.unsortedFile, hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				t.Fatalf("Failed to parse unsorted file %s: %v", tc.unsortedFile, diags)
			}

			// Sort and format the file using the new combined function
			actualContent, err := SortAndFormatHCLFile(file)
			if err != nil {
				if IsSortingError(err) {
					t.Logf("SortingError: %v", err)
				}
				if formattingutil.IsTerraformNotFoundError(err) {
					t.Skip("terraform command not available, skipping test")
				}
				t.Fatalf("SortAndFormatHCLFile failed: %v", err)
			}

			// Read the expected sorted file
			expectedContent, err := os.ReadFile(tc.expectedFile)
			if err != nil {
				t.Fatalf("Failed to read expected file %s: %v", tc.expectedFile, err)
			}

			// Normalize both contents for comparison
			actualNormalized := normalizeContent(actualContent)
			expectedNormalized := normalizeContent(string(expectedContent))

			if actualNormalized != expectedNormalized {
				t.Errorf("Sorting failed for %s\n\nExpected:\n%s\n\nActual:\n%s",
					tc.description, expectedNormalized, actualNormalized)
			}
		})
	}
}

func TestSortHCLFileEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		filePath    string
		description string
	}{
		{
			name:        "Empty File",
			filePath:    "testdata/TestSortHCLFile_Empty.tf",
			description: "Test sorting of an empty file",
		},
		{
			name:        "Comments Only",
			filePath:    "testdata/TestSortHCLFile_CommentsOnly.tf",
			description: "Test sorting of a file with only comments",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Read the file
			content, err := os.ReadFile(tc.filePath)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", tc.filePath, err)
			}

			// Parse the file
			file, diags := hclwrite.ParseConfig(content, tc.filePath, hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				t.Fatalf("Failed to parse file %s: %v", tc.filePath, diags)
			}

			// Sort the file
			sortedFile := SortHCLFile(file)
			actualContent := string(sortedFile.Bytes())

			// For edge cases, we expect the content to remain the same or be empty
			originalContent := string(content)
			if actualContent != originalContent && actualContent != "" {
				t.Errorf("Edge case test failed for %s\n\nOriginal:\n%s\n\nAfter sorting:\n%s",
					tc.description, originalContent, actualContent)
			}
		})
	}
}

func TestSortHCLFileRoundTrip(t *testing.T) {
	// Test that sorting a file twice produces the same result
	testFiles := []string{
		"testdata/TestSortHCLFile_Complex_unsorted.tf",
		"testdata/TestSortHCLFile_ResourceLabels_unsorted.tf",
		"testdata/TestSortHCLFile_Attributes_unsorted.tf",
	}

	for _, filePath := range testFiles {
		t.Run(filepath.Base(filePath), func(t *testing.T) {
			// Read the file
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", filePath, err)
			}

			// Parse the file
			file, diags := hclwrite.ParseConfig(content, filePath, hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				t.Fatalf("Failed to parse file %s: %v", filePath, diags)
			}

			// Sort and format once
			content1, err := SortAndFormatHCLFile(file)
			if err != nil {
				if IsSortingError(err) {
					t.Logf("SortingError: %v", err)
				}
				if formattingutil.IsTerraformNotFoundError(err) {
					t.Skip("terraform command not available, skipping test")
				}
				t.Fatalf("First SortAndFormatHCLFile failed: %v", err)
			}

			// Sort and format again (parse the formatted content first)
			file2, diags := hclwrite.ParseConfig([]byte(content1), filePath, hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				t.Fatalf("Failed to parse formatted content: %v", diags)
			}
			content2, err := SortAndFormatHCLFile(file2)
			if err != nil {
				if IsSortingError(err) {
					t.Logf("SortingError: %v", err)
				}
				t.Fatalf("Second SortAndFormatHCLFile failed: %v", err)
			}

			// Normalize both contents
			normalized1 := normalizeContent(content1)
			normalized2 := normalizeContent(content2)

			if normalized1 != normalized2 {
				t.Errorf("Round-trip sorting failed for %s\n\nFirst sort:\n%s\n\nSecond sort:\n%s",
					filePath, normalized1, normalized2)
			}
		})
	}
}

func TestSortHCLFileBasicFunctionality(t *testing.T) {
	testCases := []struct {
		name           string
		unsortedFile   string
		description    string
		expectedBlocks []string // Expected block types in order
	}{
		{
			name:         "Basic Block Order",
			unsortedFile: "testdata/TestSortHCLFile_Complex_unsorted.tf",
			description:  "Test that blocks are sorted in the correct order",
			expectedBlocks: []string{
				"terraform",
				"provider",
				"variable",
				"locals",
				"data",
				"resource",
				"module",
				"output",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Read the unsorted file
			unsortedContent, err := os.ReadFile(tc.unsortedFile)
			if err != nil {
				t.Fatalf("Failed to read unsorted file %s: %v", tc.unsortedFile, err)
			}

			// Parse the unsorted file
			file, diags := hclwrite.ParseConfig(unsortedContent, tc.unsortedFile, hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				t.Fatalf("Failed to parse unsorted file %s: %v", tc.unsortedFile, diags)
			}

			// Sort and format the file
			actualContent, err := SortAndFormatHCLFile(file)
			if err != nil {
				if IsSortingError(err) {
					t.Logf("SortingError: %v", err)
				}
				if formattingutil.IsTerraformNotFoundError(err) {
					t.Skip("terraform command not available, skipping test")
				}
				t.Fatalf("SortAndFormatHCLFile failed: %v", err)
			}

			// Log the sorted content for debugging
			t.Logf("Sorted content:\n%s", actualContent)

			// Check that blocks appear in the expected order
			content := actualContent
			lastIndex := -1
			for i, expectedBlock := range tc.expectedBlocks {
				index := strings.Index(content, expectedBlock+" {")
				if index == -1 {
					// Try alternative formats
					index = strings.Index(content, expectedBlock+" \"")
				}
				if index == -1 {
					t.Errorf("Expected block type '%s' not found in sorted content", expectedBlock)
					continue
				}
				if lastIndex != -1 && index < lastIndex {
					t.Errorf("Block type '%s' appears before '%s' in sorted content",
						expectedBlock, tc.expectedBlocks[i-1])
				}
				lastIndex = index
			}
		})
	}
}

// normalizeContent removes extra whitespace and normalizes line endings for comparison
func normalizeContent(content string) string {
	// Remove extra whitespace and normalize line endings
	lines := strings.Split(content, "\n")
	var normalizedLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			normalizedLines = append(normalizedLines, trimmed)
		}
	}

	return strings.Join(normalizedLines, "\n")
}
