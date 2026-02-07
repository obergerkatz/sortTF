package hcl

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// TestGetBlockType tests block type identification
func TestGetBlockType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected BlockType
	}{
		{"terraform", "terraform", BlockTypeTerraform},
		{"provider", "provider", BlockTypeProvider},
		{"variable", "variable", BlockTypeVariable},
		{"output", "output", BlockTypeOutput},
		{"resource", "resource", BlockTypeResource},
		{"data", "data", BlockTypeData},
		{"module", "module", BlockTypeModule},
		{"locals", "locals", BlockTypeLocals},
		{"backend", "backend", BlockTypeBackend},
		{"unknown", "unknown_type", BlockTypeOther},
		{"case insensitive - TERRAFORM", "TERRAFORM", BlockTypeTerraform},
		{"case insensitive - Variable", "Variable", BlockTypeVariable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getBlockType(tt.input)
			if got != tt.expected {
				t.Errorf("getBlockType(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// TestSortHCLFile_BlockOrder tests that blocks are sorted in correct order
func TestSortHCLFile_BlockOrder(t *testing.T) {
	input := `output "id" {
  value = aws_instance.web.id
}

resource "aws_instance" "web" {
  ami = "ami-12345"
}

variable "region" {
  type = string
}

provider "aws" {
  region = var.region
}

terraform {
  required_version = ">= 1.0"
}
`

	file, diags := hclwrite.ParseConfig([]byte(input), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse failed: %v", diags)
	}

	sorted := SortHCLFile(file)
	output := string(sorted.Bytes())

	// Check order: terraform, provider, variable, resource, output
	terraformIdx := strings.Index(output, "terraform {")
	providerIdx := strings.Index(output, "provider \"aws\"")
	variableIdx := strings.Index(output, "variable \"region\"")
	resourceIdx := strings.Index(output, "resource \"aws_instance\"")
	outputIdx := strings.Index(output, "output \"id\"")

	if terraformIdx == -1 || providerIdx == -1 || variableIdx == -1 || resourceIdx == -1 || outputIdx == -1 {
		t.Fatal("not all blocks found in output")
	}

	if terraformIdx >= providerIdx || providerIdx >= variableIdx || variableIdx >= resourceIdx || resourceIdx >= outputIdx {
		t.Errorf("blocks not in correct order\nterraform:%d provider:%d variable:%d resource:%d output:%d",
			terraformIdx, providerIdx, variableIdx, resourceIdx, outputIdx)
		t.Logf("Output:\n%s", output)
	}
}

// TestSortHCLFile_SameTypeAlphabetical tests alphabetical sorting within same type
func TestSortHCLFile_SameTypeAlphabetical(t *testing.T) {
	input := `variable "zebra" {
  type = string
}

variable "alpha" {
  type = string
}

variable "beta" {
  type = string
}
`

	file, diags := hclwrite.ParseConfig([]byte(input), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse failed: %v", diags)
	}

	sorted := SortHCLFile(file)
	output := string(sorted.Bytes())

	alphaIdx := strings.Index(output, "variable \"alpha\"")
	betaIdx := strings.Index(output, "variable \"beta\"")
	zebraIdx := strings.Index(output, "variable \"zebra\"")

	if alphaIdx >= betaIdx || betaIdx >= zebraIdx {
		t.Errorf("variables not sorted alphabetically: alpha:%d beta:%d zebra:%d",
			alphaIdx, betaIdx, zebraIdx)
		t.Logf("Output:\n%s", output)
	}
}

// TestSortHCLFile_ResourcesSortedAlphabetically tests resource sorting
func TestSortHCLFile_ResourcesSortedAlphabetically(t *testing.T) {
	input := `resource "aws_s3_bucket" "data" {
  bucket = "test"
}

resource "aws_instance" "web" {
  ami = "ami-12345"
}

resource "aws_security_group" "app" {
  name = "app-sg"
}
`

	file, diags := hclwrite.ParseConfig([]byte(input), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse failed: %v", diags)
	}

	sorted := SortHCLFile(file)
	output := string(sorted.Bytes())

	// Should be: aws_instance, aws_s3_bucket, aws_security_group (by first label)
	instanceIdx := strings.Index(output, "resource \"aws_instance\" \"web\"")
	bucketIdx := strings.Index(output, "resource \"aws_s3_bucket\" \"data\"")
	sgIdx := strings.Index(output, "resource \"aws_security_group\" \"app\"")

	if instanceIdx >= bucketIdx || bucketIdx >= sgIdx {
		t.Errorf("resources not sorted alphabetically: instance:%d bucket:%d sg:%d",
			instanceIdx, bucketIdx, sgIdx)
		t.Logf("Output:\n%s", output)
	}
}

// TestSortHCLFile_AttributesSorted tests that attributes within blocks are sorted
func TestSortHCLFile_AttributesSorted(t *testing.T) {
	input := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-12345"
  availability_zone = "us-west-2a"
  tags = {
    Name = "web"
  }
}
`

	file, diags := hclwrite.ParseConfig([]byte(input), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse failed: %v", diags)
	}

	sorted := SortHCLFile(file)
	output := string(sorted.Bytes())

	// Attributes should be: ami, availability_zone, instance_type, tags (alphabetical)
	lines := strings.Split(output, "\n")
	var attrs []string
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "resource") {
			inBlock = true
			continue
		}
		if inBlock && strings.Contains(trimmed, "=") && !strings.Contains(trimmed, "Name") {
			// Extract attribute name
			parts := strings.Split(trimmed, "=")
			if len(parts) > 0 {
				attrs = append(attrs, strings.TrimSpace(parts[0]))
			}
		}
	}

	expected := []string{"ami", "availability_zone", "instance_type", "tags"}
	if len(attrs) != len(expected) {
		t.Errorf("expected %d attributes, got %d: %v", len(expected), len(attrs), attrs)
	}

	for i, attr := range attrs {
		if i < len(expected) && attr != expected[i] {
			t.Errorf("attribute %d: expected %q, got %q", i, expected[i], attr)
		}
	}
}

// TestSortHCLFile_ForEachFirst tests that for_each is always first
func TestSortHCLFile_ForEachFirst(t *testing.T) {
	input := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-12345"
  for_each = var.instances
  tags = {
    Name = "web"
  }
}
`

	file, diags := hclwrite.ParseConfig([]byte(input), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse failed: %v", diags)
	}

	sorted := SortHCLFile(file)
	output := string(sorted.Bytes())

	lines := strings.Split(output, "\n")
	foundForEach := false
	foundOtherAttr := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "for_each") {
			foundForEach = true
			if foundOtherAttr {
				t.Error("for_each is not first attribute")
			}
		} else if strings.Contains(trimmed, "=") && !strings.Contains(trimmed, "{") && !strings.Contains(trimmed, "Name") {
			if strings.HasPrefix(trimmed, "ami") || strings.HasPrefix(trimmed, "instance_type") || strings.HasPrefix(trimmed, "tags") {
				foundOtherAttr = true
			}
		}
	}

	if !foundForEach {
		t.Error("for_each not found in output")
	}
}

// TestSortHCLFile_EmptyFile tests handling of empty files
func TestSortHCLFile_EmptyFile(t *testing.T) {
	file := hclwrite.NewEmptyFile()
	sorted := SortHCLFile(file)

	if sorted == nil {
		t.Error("expected non-nil result for empty file")
	}

	output := string(sorted.Bytes())
	if strings.TrimSpace(output) != "" {
		t.Errorf("expected empty output, got: %q", output)
	}
}

// TestSortHCLFile_NilFile tests handling of nil file
func TestSortHCLFile_NilFile(t *testing.T) {
	sorted := SortHCLFile(nil)

	if sorted == nil {
		t.Error("expected non-nil result for nil file")
	}

	output := string(sorted.Bytes())
	if strings.TrimSpace(output) != "" {
		t.Errorf("expected empty output, got: %q", output)
	}
}

// TestCompareLabels tests label comparison
func TestCompareLabels(t *testing.T) {
	tests := []struct {
		name     string
		labels1  []string
		labels2  []string
		expected bool
	}{
		{
			name:     "first less than second",
			labels1:  []string{"aaa"},
			labels2:  []string{"bbb"},
			expected: true,
		},
		{
			name:     "first greater than second",
			labels1:  []string{"bbb"},
			labels2:  []string{"aaa"},
			expected: false,
		},
		{
			name:     "equal single labels",
			labels1:  []string{"aaa"},
			labels2:  []string{"aaa"},
			expected: false,
		},
		{
			name:     "shorter list comes first",
			labels1:  []string{"aaa"},
			labels2:  []string{"aaa", "bbb"},
			expected: true,
		},
		{
			name:     "longer list comes second",
			labels1:  []string{"aaa", "bbb"},
			labels2:  []string{"aaa"},
			expected: false,
		},
		{
			name:     "compare second label when first equal",
			labels1:  []string{"aaa", "xxx"},
			labels2:  []string{"aaa", "yyy"},
			expected: true,
		},
		{
			name:     "empty slices",
			labels1:  []string{},
			labels2:  []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareLabels(tt.labels1, tt.labels2)
			if got != tt.expected {
				t.Errorf("compareLabels(%v, %v) = %v, want %v",
					tt.labels1, tt.labels2, got, tt.expected)
			}
		})
	}
}

// TestSortBlocksByType tests block type sorting function
func TestSortBlocksByType(t *testing.T) {
	blocks := []Block{
		{Type: BlockTypeOutput, Labels: []string{"id"}},
		{Type: BlockTypeResource, Labels: []string{"aws_instance", "web"}},
		{Type: BlockTypeVariable, Labels: []string{"region"}},
		{Type: BlockTypeTerraform},
	}

	sorted := SortBlocksByType(blocks)

	// Should be: terraform, variable, resource, output
	if sorted[0].Type != BlockTypeTerraform {
		t.Errorf("first block should be terraform, got %v", sorted[0].Type)
	}
	if sorted[1].Type != BlockTypeVariable {
		t.Errorf("second block should be variable, got %v", sorted[1].Type)
	}
	if sorted[2].Type != BlockTypeResource {
		t.Errorf("third block should be resource, got %v", sorted[2].Type)
	}
	if sorted[3].Type != BlockTypeOutput {
		t.Errorf("fourth block should be output, got %v", sorted[3].Type)
	}
}

// TestSortBlocksByLabels tests label sorting function
func TestSortBlocksByLabels(t *testing.T) {
	blocks := []Block{
		{Type: BlockTypeVariable, Labels: []string{"zebra"}},
		{Type: BlockTypeVariable, Labels: []string{"alpha"}},
		{Type: BlockTypeVariable, Labels: []string{"beta"}},
	}

	sorted := SortBlocksByLabels(blocks)

	if sorted[0].Labels[0] != "alpha" {
		t.Errorf("first should be alpha, got %v", sorted[0].Labels)
	}
	if sorted[1].Labels[0] != "beta" {
		t.Errorf("second should be beta, got %v", sorted[1].Labels)
	}
	if sorted[2].Labels[0] != "zebra" {
		t.Errorf("third should be zebra, got %v", sorted[2].Labels)
	}
}

// TestSortAttributes tests attribute sorting helper
func TestSortAttributes(t *testing.T) {
	// Create a simple block to get attributes
	input := `resource "test" "test" {
  zebra = "z"
  alpha = "a"
  beta = "b"
}
`
	file, diags := hclwrite.ParseConfig([]byte(input), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse failed: %v", diags)
	}

	blocks := file.Body().Blocks()
	if len(blocks) == 0 {
		t.Fatal("no blocks found")
	}

	attrs := blocks[0].Body().Attributes()
	sorted := SortAttributes(attrs)

	expected := []string{"alpha", "beta", "zebra"}
	for i, name := range sorted {
		if i < len(expected) && name != expected[i] {
			t.Errorf("position %d: expected %q, got %q", i, expected[i], name)
		}
	}
}

// TestSortAndFormatHCLFile tests the combined operation
func TestSortAndFormatHCLFile(t *testing.T) {
	input := `output "id" {
  value = aws_instance.web.id
}

resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-12345"
}

variable "region" {
  type = string
}
`

	file, diags := hclwrite.ParseConfig([]byte(input), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse failed: %v", diags)
	}

	formatted, err := SortAndFormatHCLFile(file)
	if err != nil {
		t.Fatalf("SortAndFormatHCLFile failed: %v", err)
	}

	// Check it's valid HCL
	_, diags = hclwrite.ParseConfig([]byte(formatted), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Errorf("output is invalid HCL: %v", diags)
	}

	// Check order
	variableIdx := strings.Index(formatted, "variable")
	resourceIdx := strings.Index(formatted, "resource")
	outputIdx := strings.Index(formatted, "output")

	if variableIdx >= resourceIdx || resourceIdx >= outputIdx {
		t.Error("blocks not in correct order after SortAndFormatHCLFile")
		t.Logf("Output:\n%s", formatted)
	}
}

// TestSortAndFormatHCLFile_WithError tests error handling
func TestSortAndFormatHCLFile_WithError(t *testing.T) {
	// SortHCLFile handles nil by returning empty file, which then formats successfully
	// So we need a different approach - test with a valid file

	// The function actually doesn't error in most cases since SortHCLFile handles nil
	// Let's test that it works correctly
	formatted, err := SortAndFormatHCLFile(nil)

	// Should not error (handles nil gracefully)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should return empty string for nil file
	if strings.TrimSpace(formatted) != "" {
		t.Errorf("expected empty string for nil file, got: %q", formatted)
	}
}

// TestSortHCLFile_NestedBlocks tests sorting of nested blocks
func TestSortHCLFile_NestedBlocks(t *testing.T) {
	input := `terraform {
  required_providers {
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
    aws = {
      source = "hashicorp/aws"
    }
  }
}
`

	file, diags := hclwrite.ParseConfig([]byte(input), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse failed: %v", diags)
	}

	sorted := SortHCLFile(file)
	output := string(sorted.Bytes())

	// Nested blocks should also be sorted (aws before kubernetes)
	awsIdx := strings.Index(output, "aws")
	k8sIdx := strings.Index(output, "kubernetes")

	if awsIdx > k8sIdx && awsIdx != -1 && k8sIdx != -1 {
		t.Error("nested blocks not sorted alphabetically")
		t.Logf("Output:\n%s", output)
	}
}

// TestSortHCLFile_DeeplyNestedBlocks tests deeply nested block sorting
func TestSortHCLFile_DeeplyNestedBlocks(t *testing.T) {
	input := `resource "aws_instance" "web" {
  ebs_block_device {
    volume_type = "gp2"
    device_name = "/dev/sda1"
    volume_size = 10
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 20
  }

  root_block_device {
    volume_type = "gp2"
    volume_size = 8
  }
}
`

	file, diags := hclwrite.ParseConfig([]byte(input), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse failed: %v", diags)
	}

	sorted := SortHCLFile(file)
	output := string(sorted.Bytes())

	// Should have all blocks
	if !strings.Contains(output, "ebs_block_device") {
		t.Error("ebs_block_device missing from output")
	}
	if !strings.Contains(output, "root_block_device") {
		t.Error("root_block_device missing from output")
	}

	// Attributes within nested blocks should be sorted
	lines := strings.Split(output, "\n")
	inEBSBlock := false
	var attrs []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "ebs_block_device") {
			inEBSBlock = true
			attrs = attrs[:0] // reset
			continue
		}
		if inEBSBlock && (strings.HasPrefix(trimmed, "}") || trimmed == "") {
			// Check if we collected any attributes
			if len(attrs) > 1 {
				// Verify they're sorted
				for i := 1; i < len(attrs); i++ {
					if attrs[i-1] > attrs[i] {
						t.Errorf("attributes not sorted in nested block: %v", attrs)
						break
					}
				}
			}
			inEBSBlock = false
		}
		if inEBSBlock && strings.Contains(trimmed, "=") {
			parts := strings.Split(trimmed, "=")
			if len(parts) > 0 {
				attrs = append(attrs, strings.TrimSpace(parts[0]))
			}
		}
	}
}

// TestSortBlockAttributes_Nil tests nil block handling
func TestSortBlockAttributes_Nil(_ *testing.T) {
	// Should not panic
	sortBlockAttributes(nil)
}

// TestSortBlockAttributes_EmptyBlock tests empty block handling
func TestSortBlockAttributes_EmptyBlock(t *testing.T) {
	input := `resource "test" "test" {
}
`
	file, diags := hclwrite.ParseConfig([]byte(input), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse failed: %v", diags)
	}

	blocks := file.Body().Blocks()
	if len(blocks) == 0 {
		t.Fatal("no blocks found")
	}

	// Should not panic
	sortBlockAttributes(blocks[0])
}

// TestParseBlocks_BackendSkipped tests that top-level backend blocks are skipped
func TestParseBlocks_BackendSkipped(t *testing.T) {
	input := `backend "s3" {
  bucket = "test"
}

variable "test" {
  type = string
}
`

	file, diags := hclwrite.ParseConfig([]byte(input), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse failed: %v", diags)
	}

	blocks := parseBlocks(file.Body())

	// Backend should be skipped, only variable should be present
	if len(blocks) != 1 {
		t.Errorf("expected 1 block (backend skipped), got %d", len(blocks))
	}

	if len(blocks) > 0 && blocks[0].Type != BlockTypeVariable {
		t.Errorf("expected variable block, got %v", blocks[0].Type)
	}
}

// TestSortHCLFile_ComplexRealWorld tests a realistic complex file
func TestSortHCLFile_ComplexRealWorld(t *testing.T) {
	input := `output "vpc_id" {
  value = module.vpc.id
}

module "vpc" {
  source = "./modules/vpc"
  cidr   = "10.0.0.0/16"
}

data "aws_ami" "ubuntu" {
  most_recent = true
}

resource "aws_security_group" "app" {
  name = "app"
}

resource "aws_instance" "web" {
  ami = data.aws_ami.ubuntu.id
}

locals {
  tags = {
    Environment = "prod"
  }
}

variable "region" {
  type = string
}

variable "env" {
  type = string
}

provider "aws" {
  region = var.region
}

terraform {
  required_version = ">= 1.0"
}
`

	file, diags := hclwrite.ParseConfig([]byte(input), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse failed: %v", diags)
	}

	sorted := SortHCLFile(file)
	output := string(sorted.Bytes())

	// Check correct order: terraform, provider, variable, locals, data, resource, module, output
	indices := map[string]int{
		"terraform": strings.Index(output, "terraform {"),
		"provider":  strings.Index(output, "provider \"aws\""),
		"variable":  strings.Index(output, "variable \"env\""), // First alphabetically
		"locals":    strings.Index(output, "locals {"),
		"data":      strings.Index(output, "data \"aws_ami\""),
		"resource":  strings.Index(output, "resource \"aws_instance\""), // First alphabetically
		"module":    strings.Index(output, "module \"vpc\""),
		"output":    strings.Index(output, "output \"vpc_id\""),
	}

	// Verify all blocks exist
	for name, idx := range indices {
		if idx == -1 {
			t.Errorf("block %s not found in output", name)
		}
	}

	// Verify ordering
	if indices["terraform"] >= indices["provider"] ||
		indices["provider"] >= indices["variable"] ||
		indices["variable"] >= indices["locals"] ||
		indices["locals"] >= indices["data"] ||
		indices["data"] >= indices["resource"] ||
		indices["resource"] >= indices["module"] ||
		indices["module"] >= indices["output"] {
		t.Error("blocks not in correct canonical order")
		t.Logf("Indices: %+v", indices)
		t.Logf("Output:\n%s", output)
	}
}

// TestSortHCLFile_NestedBlocksOfDifferentTypes tests sorting of nested blocks with different types
func TestSortHCLFile_NestedBlocksOfDifferentTypes(t *testing.T) {
	// Create a terraform block with different nested block types (backend and required_providers)
	input := `terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
  backend "s3" {
    bucket = "mybucket"
  }
  required_version = ">= 1.0"
}`

	file, diags := hclwrite.ParseConfig([]byte(input), "test.hcl", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse error: %s", diags)
	}

	sorted := SortHCLFile(file)
	output := string(sorted.Bytes())

	// Verify output contains both nested blocks
	if !strings.Contains(output, "backend") {
		t.Error("sorted output should contain backend block")
	}
	if !strings.Contains(output, "required_providers") {
		t.Error("sorted output should contain required_providers block")
	}
	if !strings.Contains(output, "required_version") {
		t.Error("sorted output should contain required_version attribute")
	}

	t.Logf("Sorted output:\n%s", output)
}
