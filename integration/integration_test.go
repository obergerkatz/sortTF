package integration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain builds the binary before running integration tests
func TestMain(m *testing.M) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "sorttf-test", "..")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		panic("failed to build binary: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.Remove("sorttf-test")

	os.Exit(code)
}

// runSortTF executes the sorttf binary with the given arguments
func runSortTF(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command("./sorttf-test", args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run sorttf: %v", err)
		}
	}

	return stdout, stderr, exitCode
}

// TestIntegration_Help tests the help command
func TestIntegration_Help(t *testing.T) {
	stdout, stderr, exitCode := runSortTF(t, "--help")

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stderr, "Usage: sorttf") {
		t.Errorf("expected usage message in stderr, got: %s", stderr)
	}

	if stdout != "" {
		t.Errorf("expected empty stdout, got: %s", stdout)
	}
}

// TestIntegration_SingleFile tests sorting a single file end-to-end
func TestIntegration_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "main.tf")

	// Create an unsorted file
	unsortedContent := `resource "aws_s3_bucket" "data" {
  bucket = "my-bucket"
  acl    = "private"
}

variable "region" {
  type = string
}

resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami           = "ami-12345"
}
`

	if err := os.WriteFile(testFile, []byte(unsortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Run sorttf
	stdout, stderr, exitCode := runSortTF(t, testFile)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstdout: %s\nstderr: %s", exitCode, stdout, stderr)
	}

	if !strings.Contains(stdout, "Updated") || !strings.Contains(stdout, "main.tf") {
		t.Errorf("expected success message in stdout, got: %s", stdout)
	}

	// Verify file was sorted correctly
	sortedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	content := string(sortedContent)

	// Verify variable comes before resources
	variableIdx := strings.Index(content, "variable \"region\"")
	instanceIdx := strings.Index(content, "resource \"aws_instance\" \"web\"")
	bucketIdx := strings.Index(content, "resource \"aws_s3_bucket\" \"data\"")

	if variableIdx == -1 || instanceIdx == -1 || bucketIdx == -1 {
		t.Errorf("missing expected blocks in sorted content: %s", content)
	}

	if variableIdx > instanceIdx {
		t.Error("variable should come before resource")
	}

	if instanceIdx > bucketIdx {
		t.Error("resources should be sorted alphabetically (aws_instance before aws_s3_bucket)")
	}
}

// TestIntegration_DryRun tests dry-run mode doesn't modify files
func TestIntegration_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "main.tf")

	unsortedContent := `resource "aws_instance" "web" {
  ami = "ami-12345"
}

variable "region" {
  type = string
}
`

	if err := os.WriteFile(testFile, []byte(unsortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Run with --dry-run
	stdout, stderr, exitCode := runSortTF(t, "--dry-run", testFile)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "Would update") {
		t.Errorf("expected 'Would update' in stdout, got: %s", stdout)
	}

	// Verify file wasn't modified
	currentContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(currentContent) != unsortedContent {
		t.Error("file was modified in dry-run mode")
	}
}

// TestIntegration_Validate tests validate mode
func TestIntegration_Validate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a sorted file
	sortedFile := filepath.Join(tmpDir, "sorted.tf")
	sortedContent := `variable "region" {
  type = string
}

resource "aws_instance" "web" {
  ami           = "ami-12345"
  instance_type = "t2.micro"
}
`
	if err := os.WriteFile(sortedFile, []byte(sortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create an unsorted file
	unsortedFile := filepath.Join(tmpDir, "unsorted.tf")
	unsortedContent := `resource "aws_instance" "web" {
  ami = "ami-12345"
}

variable "region" {
  type = string
}
`
	if err := os.WriteFile(unsortedFile, []byte(unsortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test sorted file passes validation
	stdout, stderr, exitCode := runSortTF(t, "--validate", sortedFile)
	if exitCode != 0 {
		t.Errorf("sorted file should pass validation, got exit code %d\nstdout: %s\nstderr: %s",
			exitCode, stdout, stderr)
	}

	// Test unsorted file fails validation
	stdout, stderr, exitCode = runSortTF(t, "--validate", unsortedFile)
	if exitCode != 1 {
		t.Errorf("unsorted file should fail validation, got exit code %d", exitCode)
	}

	if !strings.Contains(stdout, "Needs update") {
		t.Errorf("expected 'Needs update' in stdout, got: %s", stdout)
	}
}

// TestIntegration_Recursive tests recursive directory processing
func TestIntegration_Recursive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	// tmpDir/
	//   main.tf
	//   modules/
	//     network/
	//       main.tf
	//     compute/
	//       main.tf

	unsortedContent := `resource "aws_instance" "web" {
  ami = "ami-12345"
}

variable "name" {
  type = string
}
`

	files := []string{
		filepath.Join(tmpDir, "main.tf"),
		filepath.Join(tmpDir, "modules", "network", "main.tf"),
		filepath.Join(tmpDir, "modules", "compute", "main.tf"),
	}

	for _, file := range files {
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(unsortedContent), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Run with --recursive
	stdout, _, exitCode := runSortTF(t, "--recursive", tmpDir)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	// Verify all files were processed
	if !strings.Contains(stdout, "Processed 3 files") {
		t.Errorf("expected 'Processed 3 files' in stdout, got: %s", stdout)
	}

	// Verify each file was sorted
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		// Variable should come before resource
		if strings.Index(string(content), "variable") > strings.Index(string(content), "resource") {
			t.Errorf("file %s was not sorted correctly", file)
		}
	}
}

// TestIntegration_InvalidSyntax tests handling of files with syntax errors
func TestIntegration_InvalidSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.tf")

	invalidContent := `resource "aws_instance" "web" {
  ami = "ami-12345"
  # Missing closing brace
`

	if err := os.WriteFile(testFile, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, exitCode := runSortTF(t, testFile)

	if exitCode != 1 {
		t.Errorf("expected exit code 1 for invalid syntax, got %d", exitCode)
	}

	output := stdout + stderr
	if !strings.Contains(output, "❌") || !strings.Contains(output, "invalid.tf") {
		t.Errorf("expected error message for invalid.tf, got stdout: %s, stderr: %s", stdout, stderr)
	}
}

// TestIntegration_NonExistentPath tests error handling for missing paths
func TestIntegration_NonExistentPath(t *testing.T) {
	stdout, stderr, exitCode := runSortTF(t, "/nonexistent/path/to/file.tf")

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	output := stdout + stderr
	if !strings.Contains(output, "does not exist") {
		t.Errorf("expected 'does not exist' in output, got: %s", output)
	}
}

// TestIntegration_MixedFileTypes tests processing directory with mixed file types
func TestIntegration_MixedFileTypes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .tf file
	tfFile := filepath.Join(tmpDir, "main.tf")
	tfContent := `variable "name" { type = string }
resource "aws_instance" "web" { ami = "ami-12345" }`
	if err := os.WriteFile(tfFile, []byte(tfContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .hcl file (Terragrunt)
	hclFile := filepath.Join(tmpDir, "terragrunt.hcl")
	hclContent := `terraform {
  source = "../modules/vpc"
}

inputs = {
  region = "us-west-2"
}
`
	if err := os.WriteFile(hclFile, []byte(hclContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create non-Terraform file (should be ignored)
	txtFile := filepath.Join(tmpDir, "README.txt")
	if err := os.WriteFile(txtFile, []byte("readme"), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _, exitCode := runSortTF(t, tmpDir)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	// Should process both .tf and .hcl files (2 files)
	if !strings.Contains(stdout, "Processed 2 files") {
		t.Errorf("expected 'Processed 2 files', got: %s", stdout)
	}
}

// TestIntegration_VerboseMode tests verbose output
func TestIntegration_VerboseMode(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "main.tf")

	content := `variable "name" { type = string }`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _, exitCode := runSortTF(t, "--verbose", testFile)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	// Verbose mode should show more details
	if !strings.Contains(stdout, "Processing") {
		t.Errorf("expected verbose 'Processing' message, got: %s", stdout)
	}
}

// TestIntegration_ComplexRealWorld tests a realistic complex Terraform configuration
func TestIntegration_ComplexRealWorld(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "complex.tf")

	// Realistic but unsorted Terraform configuration
	complexContent := `resource "aws_security_group" "app" {
  name        = "app-sg"
  description = "Security group for application"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

terraform {
  required_version = ">= 1.0"

  backend "s3" {
    bucket = "terraform-state"
    key    = "app/terraform.tfstate"
    region = "us-west-2"
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

output "security_group_id" {
  description = "Security group ID"
  value       = aws_security_group.app.id
}

provider "aws" {
  region = var.region
}

variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "environment" {
  description = "Environment name"
  type        = string
}

resource "aws_instance" "app" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type

  vpc_security_group_ids = [aws_security_group.app.id]
  subnet_id              = module.vpc.private_subnets[0]

  tags = merge(
    local.common_tags,
    {
      Name = "app-server"
    }
  )
}

locals {
  common_tags = {
    Environment = var.environment
    ManagedBy   = "Terraform"
    Project     = "MyApp"
  }
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
}

module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  name = "${var.environment}-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["us-west-2a", "us-west-2b"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24"]

  enable_nat_gateway = true
  enable_vpn_gateway = false

  tags = local.common_tags
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

output "instance_id" {
  description = "Application instance ID"
  value       = aws_instance.app.id
}
`

	if err := os.WriteFile(testFile, []byte(complexContent), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, exitCode := runSortTF(t, testFile)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstdout: %s\nstderr: %s", exitCode, stdout, stderr)
	}

	// Read sorted content
	sortedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	content := string(sortedContent)

	// Verify block order: terraform, provider, variable, locals, data, resource, module, output
	indices := map[string]int{
		"terraform":    strings.Index(content, "terraform {"),
		"provider":     strings.Index(content, "provider \"aws\""),
		"variable":     strings.Index(content, "variable \"environment\""), // First alphabetically
		"locals":       strings.Index(content, "locals {"),
		"data":         strings.Index(content, "data \"aws_ami\""),
		"resource_app": strings.Index(content, "resource \"aws_instance\""),
		"resource_sg":  strings.Index(content, "resource \"aws_security_group\""),
		"module":       strings.Index(content, "module \"vpc\""),
		"output_inst":  strings.Index(content, "output \"instance_id\""),
		"output_sg":    strings.Index(content, "output \"security_group_id\""),
	}

	// Verify all blocks exist
	for name, idx := range indices {
		if idx == -1 {
			t.Errorf("missing block: %s", name)
		}
	}

	// Verify correct ordering
	if indices["terraform"] > indices["provider"] {
		t.Error("terraform block should come before provider")
	}
	if indices["provider"] > indices["variable"] {
		t.Error("provider block should come before variable")
	}
	if indices["variable"] > indices["locals"] {
		t.Error("variable blocks should come before locals")
	}
	if indices["locals"] > indices["data"] {
		t.Error("locals should come before data")
	}
	if indices["data"] > indices["resource_app"] {
		t.Error("data blocks should come before resource")
	}
	if indices["resource_app"] > indices["resource_sg"] {
		t.Error("resources should be sorted alphabetically (aws_instance before aws_security_group)")
	}
	if indices["resource_sg"] > indices["module"] {
		t.Error("resource blocks should come before module")
	}
	if indices["module"] > indices["output_inst"] {
		t.Error("module blocks should come before output")
	}
	if indices["output_inst"] > indices["output_sg"] {
		t.Error("outputs should be sorted alphabetically (instance_id before security_group_id)")
	}
}

// TestIntegration_EmptyDirectory tests handling of empty directories
func TestIntegration_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, _, exitCode := runSortTF(t, tmpDir)

	if exitCode != 0 {
		t.Errorf("expected exit code 0 for empty directory, got %d", exitCode)
	}

	if !strings.Contains(stdout, "No Terraform or Terragrunt files found") {
		t.Errorf("expected 'No files found' message, got: %s", stdout)
	}
}

// TestIntegration_AlreadySorted tests that already sorted files are not modified
func TestIntegration_AlreadySorted(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sorted.tf")

	sortedContent := `variable "name" {
  type = string
}

resource "aws_instance" "web" {
  ami           = "ami-12345"
  instance_type = "t2.micro"
}
`

	if err := os.WriteFile(testFile, []byte(sortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _, exitCode := runSortTF(t, testFile)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Processed 0 files") {
		t.Errorf("expected 'Processed 0 files' message, got: %s", stdout)
	}

	// Verify content is unchanged
	currentContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(currentContent) != sortedContent {
		t.Error("file content was modified")
	}
}

// TestIntegration_CICDPipeline simulates a CI/CD pipeline validation workflow
func TestIntegration_CICDPipeline(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup: Create a project directory structure
	// project/
	//   main.tf (sorted)
	//   variables.tf (unsorted)
	//   modules/
	//     network/
	//       main.tf (sorted)

	mainTF := filepath.Join(tmpDir, "main.tf")
	mainContent := `variable "region" {
  type = string
}

resource "aws_instance" "app" {
  ami = "ami-12345"
}
`

	variablesTF := filepath.Join(tmpDir, "variables.tf")
	variablesContent := `variable "instance_type" {
  default = "t2.micro"
}

variable "environment" {
  type = string
}
`

	moduleMainTF := filepath.Join(tmpDir, "modules", "network", "main.tf")
	moduleContent := `variable "vpc_cidr" {
  type = string
}

resource "aws_vpc" "main" {
  cidr_block = var.vpc_cidr
}
`

	// Create all files
	os.MkdirAll(filepath.Dir(moduleMainTF), 0755)
	os.WriteFile(mainTF, []byte(mainContent), 0644)
	os.WriteFile(variablesTF, []byte(variablesContent), 0644)
	os.WriteFile(moduleMainTF, []byte(moduleContent), 0644)

	// Scenario 1: Run validate on the entire project (should fail)
	stdout, _, exitCode := runSortTF(t, "--validate", "--recursive", tmpDir)

	if exitCode != 1 {
		t.Errorf("CI validation should fail for unsorted files, got exit code %d", exitCode)
	}

	if !strings.Contains(stdout, "Needs update") {
		t.Errorf("expected 'Needs update' in output for unsorted files, got: %s", stdout)
	}

	// Scenario 2: Developer fixes the issues locally (non-validate mode)
	stdout, _, exitCode = runSortTF(t, "--recursive", tmpDir)

	if exitCode != 0 {
		t.Errorf("expected exit code 0 after sorting, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Updated") {
		t.Errorf("expected 'Updated' in output after sorting, got: %s", stdout)
	}

	// Scenario 3: Run validate again (should pass now)
	stdout, _, exitCode = runSortTF(t, "--validate", "--recursive", tmpDir)

	if exitCode != 0 {
		t.Errorf("CI validation should pass after sorting, got exit code %d\noutput: %s", exitCode, stdout)
	}

	if strings.Contains(stdout, "Needs update") {
		t.Errorf("should not have unsorted files after fixing, got: %s", stdout)
	}

	// Scenario 4: Verify all files are correctly sorted
	files := []string{mainTF, variablesTF, moduleMainTF}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		// Variables should come before resources in all files
		variableIdx := strings.Index(string(content), "variable")
		resourceIdx := strings.Index(string(content), "resource")

		// If both exist, variable should come first
		if variableIdx != -1 && resourceIdx != -1 && variableIdx > resourceIdx {
			t.Errorf("file %s: variable should come before resource", file)
		}
	}
}

// TestIntegration_AttributeSorting tests that attributes within blocks are sorted
func TestIntegration_AttributeSorting(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "attrs.tf")

	// Resource with unsorted attributes
	unsortedContent := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami           = "ami-12345"
  tags = {
    Name = "web-server"
  }
  availability_zone = "us-west-2a"
}
`

	if err := os.WriteFile(testFile, []byte(unsortedContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, stderr, exitCode := runSortTF(t, testFile)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	// Read sorted content
	sortedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	content := string(sortedContent)

	// Verify attributes are sorted alphabetically
	amiIdx := strings.Index(content, "ami")
	availabilityZoneIdx := strings.Index(content, "availability_zone")
	instanceTypeIdx := strings.Index(content, "instance_type")
	tagsIdx := strings.Index(content, "tags")

	if amiIdx == -1 || availabilityZoneIdx == -1 || instanceTypeIdx == -1 || tagsIdx == -1 {
		t.Fatalf("missing expected attributes in sorted content: %s", content)
	}

	// Attributes should be in alphabetical order: ami, availability_zone, instance_type, tags
	if !(amiIdx < availabilityZoneIdx && availabilityZoneIdx < instanceTypeIdx && instanceTypeIdx < tagsIdx) {
		t.Errorf("attributes not sorted alphabetically:\nami: %d\navailability_zone: %d\ninstance_type: %d\ntags: %d\n\nContent:\n%s",
			amiIdx, availabilityZoneIdx, instanceTypeIdx, tagsIdx, content)
	}
}

// TestIntegration_ForEachFirst tests that for_each attribute comes first
func TestIntegration_ForEachFirst(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "foreach.tf")

	// Resource with for_each that should be first
	content := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami           = "ami-12345"
  for_each      = var.instances
  tags = {
    Name = "web-${each.key}"
  }
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, stderr, exitCode := runSortTF(t, testFile)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	// Read sorted content
	sortedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	sortedStr := string(sortedContent)

	// for_each should be the first attribute after the resource declaration
	lines := strings.Split(sortedStr, "\n")
	foundResource := false
	foundForEach := false

	for i, line := range lines {
		if strings.Contains(line, `resource "aws_instance" "web"`) {
			foundResource = true
			// Next non-empty line should contain for_each
			for j := i + 1; j < len(lines); j++ {
				nextLine := strings.TrimSpace(lines[j])
				if nextLine != "" && nextLine != "{" {
					if strings.HasPrefix(nextLine, "for_each") {
						foundForEach = true
					}
					break
				}
			}
			break
		}
	}

	if !foundResource {
		t.Error("resource declaration not found")
	}

	if !foundForEach {
		t.Errorf("for_each should be the first attribute, got:\n%s", sortedStr)
	}
}

// TestIntegration_MultipleFilesParallel tests concurrent processing of multiple files
func TestIntegration_MultipleFilesParallel(t *testing.T) {
	tmpDir := t.TempDir()

	unsortedContent := `resource "aws_instance" "web" {
  ami = "ami-12345"
}

variable "name" {
  type = string
}
`

	// Create 20 files to trigger parallel processing
	var files []string
	for i := 1; i <= 20; i++ {
		file := filepath.Join(tmpDir, fmt.Sprintf("file%02d.tf", i))
		if err := os.WriteFile(file, []byte(unsortedContent), 0644); err != nil {
			t.Fatal(err)
		}
		files = append(files, file)
	}

	stdout, stderr, exitCode := runSortTF(t, tmpDir)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "Processed 20 files") {
		t.Errorf("expected 'Processed 20 files', got: %s", stdout)
	}

	// Verify all files were sorted
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		// Variable should come before resource
		if strings.Index(string(content), "variable") > strings.Index(string(content), "resource") {
			t.Errorf("file %s was not sorted correctly", file)
		}
	}
}
