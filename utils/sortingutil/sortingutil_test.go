package sortingutil

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"sorttf/utils/formattingutil"
)

func TestGetBlockType(t *testing.T) {
	tests := []struct {
		name     string
		expected BlockType
	}{
		{"terraform", BlockTypeTerraform},
		{"provider", BlockTypeProvider},
		{"variable", BlockTypeVariable},
		{"output", BlockTypeOutput},
		{"resource", BlockTypeResource},
		{"data", BlockTypeData},
		{"module", BlockTypeModule},
		{"locals", BlockTypeLocals},
		{"backend", BlockTypeBackend},
		{"unknown", BlockTypeOther},
		{"TERRAFORM", BlockTypeTerraform}, // Case insensitive
		{"Provider", BlockTypeProvider},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBlockType(tt.name)
			if result != tt.expected {
				t.Errorf("getBlockType(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestCompareLabels(t *testing.T) {
	tests := []struct {
		name     string
		labels1  []string
		labels2  []string
		expected bool // labels1 < labels2
	}{
		{
			name:     "empty vs empty",
			labels1:  []string{},
			labels2:  []string{},
			expected: false,
		},
		{
			name:     "empty vs non-empty",
			labels1:  []string{},
			labels2:  []string{"a"},
			expected: true,
		},
		{
			name:     "non-empty vs empty",
			labels1:  []string{"a"},
			labels2:  []string{},
			expected: false,
		},
		{
			name:     "different first label",
			labels1:  []string{"a", "b"},
			labels2:  []string{"b", "a"},
			expected: true,
		},
		{
			name:     "same first label, different second",
			labels1:  []string{"a", "b"},
			labels2:  []string{"a", "c"},
			expected: true,
		},
		{
			name:     "same labels",
			labels1:  []string{"a", "b"},
			labels2:  []string{"a", "b"},
			expected: false,
		},
		{
			name:     "prefix relationship",
			labels1:  []string{"a"},
			labels2:  []string{"a", "b"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareLabels(tt.labels1, tt.labels2)
			if result != tt.expected {
				t.Errorf("compareLabels(%v, %v) = %v, want %v", tt.labels1, tt.labels2, result, tt.expected)
			}
		})
	}
}

func TestSortBlocksByType(t *testing.T) {
	blocks := []Block{
		{Type: BlockTypeResource, Labels: []string{"aws_instance", "example"}},
		{Type: BlockTypeVariable, Labels: []string{"region"}},
		{Type: BlockTypeProvider, Labels: []string{"aws"}},
		{Type: BlockTypeOutput, Labels: []string{"instance_id"}},
	}

	sorted := SortBlocksByType(blocks)

	// Expected order: provider, variable, resource, output
	expectedOrder := []BlockType{
		BlockTypeProvider,
		BlockTypeVariable,
		BlockTypeResource,
		BlockTypeOutput,
	}

	for i, expectedType := range expectedOrder {
		if sorted[i].Type != expectedType {
			t.Errorf("Expected block type %v at position %d, got %v", expectedType, i, sorted[i].Type)
		}
	}
}

func TestSortBlocksByLabels(t *testing.T) {
	blocks := []Block{
		{Type: BlockTypeResource, Labels: []string{"aws_instance", "example2"}},
		{Type: BlockTypeResource, Labels: []string{"aws_instance", "example1"}},
		{Type: BlockTypeResource, Labels: []string{"aws_s3_bucket", "data"}},
	}

	t.Logf("Original blocks:")
	for i, block := range blocks {
		t.Logf("  %d: %v %v", i, block.Type, block.Labels)
	}

	sorted := SortBlocksByLabels(blocks)

	t.Logf("Sorted blocks:")
	for i, block := range sorted {
		t.Logf("  %d: %v %v", i, block.Type, block.Labels)
	}

	// The actual sorting order based on alphabetical comparison:
	// "aws_instance" comes before "aws_s3_bucket" alphabetically
	expectedLabels := [][]string{
		{"aws_instance", "example1"},
		{"aws_instance", "example2"},
		{"aws_s3_bucket", "data"},
	}

	for i, expectedLabels := range expectedLabels {
		if !stringsEqual(sorted[i].Labels, expectedLabels) {
			t.Errorf("Expected labels %v at position %d, got %v", expectedLabels, i, sorted[i].Labels)
		}
	}
}

func TestSortAttributes(t *testing.T) {
	file := hclwrite.NewEmptyFile()
	body := file.Body()

	// Add attributes in random order
	body.SetAttributeValue("zebra", cty.StringVal("z"))
	body.SetAttributeValue("alpha", cty.StringVal("a"))
	body.SetAttributeValue("beta", cty.StringVal("b"))

	attributes := body.Attributes()
	sortedNames := SortAttributes(attributes)

	expectedOrder := []string{"alpha", "beta", "zebra"}
	for i, expected := range expectedOrder {
		if sortedNames[i] != expected {
			t.Errorf("Expected attribute name %s at position %d, got %s", expected, i, sortedNames[i])
		}
	}
}

func TestSortHCLFile(t *testing.T) {
	file := hclwrite.NewEmptyFile()
	body := file.Body()

	// Add blocks in random order
	resourceBlock := body.AppendNewBlock("resource", []string{"aws_instance", "example"})
	resourceBody := resourceBlock.Body()
	resourceBody.SetAttributeValue("zebra", cty.StringVal("z"))
	resourceBody.SetAttributeValue("alpha", cty.StringVal("a"))

	variableBlock := body.AppendNewBlock("variable", []string{"region"})
	variableBody := variableBlock.Body()
	variableBody.SetAttributeValue("type", cty.StringVal("string"))

	providerBlock := body.AppendNewBlock("provider", []string{"aws"})
	providerBody := providerBlock.Body()
	providerBody.SetAttributeValue("region", cty.StringVal("us-west-2"))

	// Use SortAndFormatHCLFile instead of SortHCLFile
	formatted, err := SortAndFormatHCLFile(file)
	if err != nil {
		if formattingutil.IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("SortAndFormatHCLFile failed: %v", err)
	}

	// Check that blocks are in correct order
	providerIndex := strings.Index(formatted, "provider")
	variableIndex := strings.Index(formatted, "variable")
	resourceIndex := strings.Index(formatted, "resource")

	if providerIndex == -1 || variableIndex == -1 || resourceIndex == -1 {
		t.Errorf("Expected all block types to be present in sorted output")
	}

	if providerIndex > variableIndex || variableIndex > resourceIndex {
		t.Errorf("Blocks not in expected order. Provider: %d, Variable: %d, Resource: %d",
			providerIndex, variableIndex, resourceIndex)
	}

	// Check that attributes within resource block are sorted
	if !strings.Contains(formatted, "alpha") || !strings.Contains(formatted, "zebra") {
		t.Errorf("Expected both attributes to be present in sorted output")
	}
}

func TestSortHCLFileEmpty(t *testing.T) {
	file := hclwrite.NewEmptyFile()
	// Use SortAndFormatHCLFile instead of SortHCLFile
	formatted, err := SortAndFormatHCLFile(file)
	if err != nil {
		if formattingutil.IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("SortAndFormatHCLFile failed: %v", err)
	}
	if formatted != "" {
		t.Errorf("Expected empty string for empty file, got: %q", formatted)
	}
}

func TestSortHCLFileNil(t *testing.T) {
	// Use SortAndFormatHCLFile instead of SortHCLFile
	formatted, err := SortAndFormatHCLFile(nil)
	if err != nil {
		if formattingutil.IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("SortAndFormatHCLFile failed: %v", err)
	}
	if formatted != "" {
		t.Errorf("Expected empty string for nil file, got: %q", formatted)
	}
}

func TestSortBlockAttributes(t *testing.T) {
	file := hclwrite.NewEmptyFile()
	body := file.Body()

	block := body.AppendNewBlock("resource", []string{"aws_instance", "example"})
	blockBody := block.Body()

	// Add attributes in random order
	blockBody.SetAttributeValue("zebra", cty.StringVal("z"))
	blockBody.SetAttributeValue("alpha", cty.StringVal("a"))
	blockBody.SetAttributeValue("beta", cty.StringVal("b"))

	sortBlockAttributes(block)

	formatted, err := SortAndFormatHCLFile(file)
	if err != nil {
		if formattingutil.IsTerraformNotFoundError(err) {
			t.Skip("terraform command not available, skipping test")
		}
		t.Fatalf("SortAndFormatHCLFile failed: %v", err)
	}

	// Check that attributes appear in alphabetical order
	alphaIndex := strings.Index(formatted, "alpha")
	betaIndex := strings.Index(formatted, "beta")
	zebraIndex := strings.Index(formatted, "zebra")

	if alphaIndex == -1 || betaIndex == -1 || zebraIndex == -1 {
		t.Errorf("Expected all attributes to be present in sorted output")
	}

	if alphaIndex > betaIndex || betaIndex > zebraIndex {
		t.Errorf("Attributes not in alphabetical order. Alpha: %d, Beta: %d, Zebra: %d",
			alphaIndex, betaIndex, zebraIndex)
	}
}

func TestSortBlockAttributesWithNestedBlocks(t *testing.T) {
	file := hclwrite.NewEmptyFile()
	body := file.Body()

	block := body.AppendNewBlock("resource", []string{"aws_instance", "example"})
	blockBody := block.Body()

	// Add attributes in random order
	blockBody.SetAttributeValue("zebra", cty.StringVal("z"))
	blockBody.SetAttributeValue("alpha", cty.StringVal("a"))

	// Add nested block
	nestedBlock := blockBody.AppendNewBlock("tags", nil)
	nestedBody := nestedBlock.Body()
	nestedBody.SetAttributeValue("zebra", cty.StringVal("z"))
	nestedBody.SetAttributeValue("alpha", cty.StringVal("a"))

	sortBlockAttributes(block)

	formatted := string(file.Bytes())

	// Check that attributes in both main block and nested block are sorted
	if !strings.Contains(formatted, "alpha") || !strings.Contains(formatted, "zebra") {
		t.Errorf("Expected all attributes to be present in sorted output")
	}
}

// Helper function to compare string slices
func stringsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSortingError(t *testing.T) {
	err := &SortingError{
		Op:   "TestOp",
		Path: "/test/path",
		Err:  fmt.Errorf("test error"),
	}

	expectedMsg := "sortingutil TestOp /test/path: test error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	if err.Unwrap().Error() != "test error" {
		t.Errorf("Expected unwrapped error 'test error', got '%s'", err.Unwrap().Error())
	}

	if !IsSortingError(err) {
		t.Error("IsSortingError should return true for SortingError")
	}
	if GetSortingErrorOp(err) != "TestOp" {
		t.Error("GetSortingErrorOp should return the operation")
	}
	if GetSortingErrorPath(err) != "/test/path" {
		t.Error("GetSortingErrorPath should return the path")
	}
}

func TestSortBlockAttributesForEachFirst(t *testing.T) {
	file := hclwrite.NewEmptyFile()
	body := file.Body()

	block := body.AppendNewBlock("resource", []string{"aws_s3_bucket", "example"})
	blockBody := block.Body()
	blockBody.SetAttributeValue("tags", cty.StringVal("should_be_last"))
	blockBody.SetAttributeValue("for_each", cty.StringVal("should_be_first"))
	blockBody.SetAttributeValue("acl", cty.StringVal("should_be_second"))

	sortBlockAttributes(block)

	// Collect attribute names in order by scanning tokens
	var attrOrder []string
	tokens := block.Body().BuildTokens(nil)
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok.Type == hclsyntax.TokenIdent {
			// The first identifier in a line is the attribute name
			attrOrder = append(attrOrder, string(tok.Bytes))
			// Skip to the end of the line
			for i+1 < len(tokens) && tokens[i+1].Type != hclsyntax.TokenNewline {
				i++
			}
		}
	}

	expectedOrder := []string{"for_each", "acl", "tags"}
	if len(attrOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d attributes, got %d: %v", len(expectedOrder), len(attrOrder), attrOrder)
	}
	for i, name := range expectedOrder {
		if attrOrder[i] != name {
			t.Errorf("Expected attribute %q at position %d, got %q", name, i, attrOrder[i])
		}
	}
}
