package integration

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain builds the binary before running integration tests
func TestMain(m *testing.M) {
	// Build the binary from cmd/sorttf
	// When running `go test ./...`, the build is executed from the repo root
	cmd := exec.Command("go", "build", "-o", filepath.Join("integration", "sorttf-test"), "./cmd/sorttf")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Build failed: %v\nOutput:\n%s\n", err, stderr.String())
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup - remove the binary from the integration directory
	_ = os.Remove("sorttf-test")

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
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
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
	stdout, _, exitCode = runSortTF(t, "--validate", unsortedFile)
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
	if err := os.MkdirAll(filepath.Dir(moduleMainTF), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mainTF, []byte(mainContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(variablesTF, []byte(variablesContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(moduleMainTF, []byte(moduleContent), 0644); err != nil {
		t.Fatal(err)
	}

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

// TestIntegration_FixtureSuite tests all fixture files can be processed
func TestIntegration_FixtureSuite(t *testing.T) {
	fixtureCategories := []string{"syntax", "structure", "types", "control", "realistic"}

	for _, category := range fixtureCategories {
		t.Run(category, func(t *testing.T) {
			fixtureDir := filepath.Join("..", "testdata", "fixtures", category)

			// Check if directory exists
			if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
				t.Skipf("Fixture directory %s does not exist", fixtureDir)
			}

			// Find all .tf and .hcl files
			entries, err := os.ReadDir(fixtureDir)
			if err != nil {
				t.Fatal(err)
			}

			fixtureCount := 0
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				name := entry.Name()
				if !strings.HasSuffix(name, ".tf") && !strings.HasSuffix(name, ".hcl") {
					continue
				}

				fixtureCount++
				fixturePath := filepath.Join(fixtureDir, name)

				// Create a copy in temp dir to avoid modifying fixtures
				tmpDir := t.TempDir()
				tmpFile := filepath.Join(tmpDir, name)

				content, err := os.ReadFile(fixturePath)
				if err != nil {
					t.Fatalf("failed to read fixture %s: %v", name, err)
				}

				if err := os.WriteFile(tmpFile, content, 0644); err != nil {
					t.Fatalf("failed to write temp file: %v", err)
				}

				// Run sorttf on the fixture
				stdout, stderr, exitCode := runSortTF(t, tmpFile)

				// Most fixtures should succeed (some like unclosed_brace.tf are intentionally invalid)
				if strings.Contains(name, "unclosed") || strings.Contains(name, "invalid") {
					if exitCode == 0 {
						t.Errorf("fixture %s should fail but succeeded", name)
					}
					continue
				}

				if exitCode != 0 {
					t.Errorf("fixture %s failed\nstdout: %s\nstderr: %s", name, stdout, stderr)
				}
			}

			if fixtureCount == 0 {
				t.Errorf("no fixtures found in %s", category)
			}
		})
	}
}

// TestIntegration_StructureFixtures tests structure-specific fixtures
func TestIntegration_StructureFixtures(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		wantErr  bool
		contains []string // strings that should be in output
	}{
		{
			name:     "all block types",
			fixture:  "all_block_types.tf",
			wantErr:  false,
			contains: []string{"terraform", "provider", "variable", "locals", "data", "resource", "module", "output"},
		},
		{
			name:     "nested blocks",
			fixture:  "nested_blocks.tf",
			wantErr:  false,
			contains: []string{"required_providers", "backend"},
		},
		{
			name:     "dynamic blocks",
			fixture:  "dynamic_blocks.tf",
			wantErr:  false,
			contains: []string{"dynamic", "for_each"},
		},
		{
			name:     "repeated blocks",
			fixture:  "repeated_blocks.tf",
			wantErr:  false,
			contains: []string{"web1", "web2", "data1", "data2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join("..", "testdata", "fixtures", "structure", tt.fixture)

			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, tt.fixture)

			content, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(tmpFile, content, 0644); err != nil {
				t.Fatal(err)
			}

			stdout, stderr, exitCode := runSortTF(t, tmpFile)

			if tt.wantErr && exitCode == 0 {
				t.Error("expected error but succeeded")
			}

			if !tt.wantErr && exitCode != 0 {
				t.Errorf("unexpected error\nstdout: %s\nstderr: %s", stdout, stderr)
			}

			// Read result
			result, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatal(err)
			}

			resultStr := string(result)

			// Verify expected strings are present
			for _, str := range tt.contains {
				if !strings.Contains(resultStr, str) {
					t.Errorf("output should contain %q but doesn't", str)
				}
			}
		})
	}
}

// TestIntegration_TypeFixtures tests type-specific fixtures
func TestIntegration_TypeFixtures(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		contains []string
	}{
		{
			name:     "nulls",
			fixture:  "nulls.tf",
			contains: []string{"null"},
		},
		{
			name:     "empty collections",
			fixture:  "empty_collections.tf",
			contains: []string{"[]", "{}"},
		},
		{
			name:     "numbers",
			fixture:  "numbers.tf",
			contains: []string{"42", "3.14", "-100"},
		},
		{
			name:     "booleans",
			fixture:  "booleans.tf",
			contains: []string{"true", "false"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join("..", "testdata", "fixtures", "types", tt.fixture)

			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, tt.fixture)

			content, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(tmpFile, content, 0644); err != nil {
				t.Fatal(err)
			}

			_, _, exitCode := runSortTF(t, tmpFile)

			if exitCode != 0 {
				t.Errorf("fixture %s failed with exit code %d", tt.fixture, exitCode)
			}

			// Read result
			result, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatal(err)
			}

			resultStr := string(result)

			// Verify expected strings are preserved
			for _, str := range tt.contains {
				if !strings.Contains(resultStr, str) {
					t.Errorf("output should contain %q but doesn't", str)
				}
			}
		})
	}
}

// TestIntegration_ControlFlowFixtures tests control flow fixtures
func TestIntegration_ControlFlowFixtures(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		contains []string
	}{
		{
			name:     "for_each",
			fixture:  "for_each.tf",
			contains: []string{"for_each"},
		},
		{
			name:     "count",
			fixture:  "count.tf",
			contains: []string{"count"},
		},
		{
			name:     "conditionals",
			fixture:  "conditionals.tf",
			contains: []string{"?", ":"},
		},
		{
			name:     "interpolation",
			fixture:  "interpolation.tf",
			contains: []string{"${var.env}"},
		},
		{
			name:     "functions",
			fixture:  "functions.tf",
			contains: []string{"merge", "range", "upper", "join"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join("..", "testdata", "fixtures", "control", tt.fixture)

			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, tt.fixture)

			content, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(tmpFile, content, 0644); err != nil {
				t.Fatal(err)
			}

			_, _, exitCode := runSortTF(t, tmpFile)

			if exitCode != 0 {
				t.Errorf("fixture %s failed with exit code %d", tt.fixture, exitCode)
			}

			// Read result
			result, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatal(err)
			}

			resultStr := string(result)

			// Verify expected strings are preserved
			for _, str := range tt.contains {
				if !strings.Contains(resultStr, str) {
					t.Errorf("output should contain %q but doesn't", str)
				}
			}
		})
	}
}

// TestIntegration_RealisticScenarios tests realistic fixture scenarios
func TestIntegration_RealisticScenarios(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		checks  []string // strings that should be present and in order
	}{
		{
			name:    "AWS infrastructure",
			fixture: "aws_infrastructure.tf",
			checks:  []string{"terraform", "provider", "variable", "locals", "data", "resource", "output"},
		},
		{
			name:    "Terragrunt config",
			fixture:  "terragrunt.hcl",
			checks:  []string{"include", "terraform"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join("..", "testdata", "fixtures", "realistic", tt.fixture)

			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, tt.fixture)

			content, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(tmpFile, content, 0644); err != nil {
				t.Fatal(err)
			}

			_, stderr, exitCode := runSortTF(t, tmpFile)

			if exitCode != 0 {
				t.Errorf("fixture %s failed with exit code %d\nstderr: %s", tt.fixture, exitCode, stderr)
			}

			// Read result
			result, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatal(err)
			}

			resultStr := string(result)

			// Verify all expected strings are present
			for _, check := range tt.checks {
				if !strings.Contains(resultStr, check) {
					t.Errorf("output should contain %q but doesn't", check)
				}
			}

			// Verify block ordering for .tf files
			if strings.HasSuffix(tt.fixture, ".tf") && len(tt.checks) > 1 {
				var indices []int
				for _, check := range tt.checks {
					idx := strings.Index(resultStr, check)
					if idx >= 0 {
						indices = append(indices, idx)
					}
				}

				// Check that indices are in ascending order (blocks are sorted)
				for i := 1; i < len(indices); i++ {
					if indices[i] < indices[i-1] {
						t.Errorf("blocks not in correct order: %v", tt.checks)
					}
				}
			}
		})
	}
}

// TestIntegration_SyntaxFixtures tests syntax edge case fixtures
func TestIntegration_SyntaxFixtures(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		wantErr bool
	}{
		{
			name:    "empty file",
			fixture: "empty.tf",
			wantErr: false,
		},
		{
			name:    "whitespace only",
			fixture: "whitespace.tf",
			wantErr: false,
		},
		{
			name:    "hash comments",
			fixture: "comments_hash.tf",
			wantErr: false,
		},
		{
			name:    "double slash comments",
			fixture: "comments_double_slash.tf",
			wantErr: false,
		},
		{
			name:    "multiline comments",
			fixture: "comments_multiline.tf",
			wantErr: false,
		},
		{
			name:    "mixed comments",
			fixture: "comments_mixed.tf",
			wantErr: false,
		},
		{
			name:    "heredoc standard",
			fixture: "heredoc_standard.tf",
			wantErr: false,
		},
		{
			name:    "heredoc indented",
			fixture: "heredoc_indented.tf",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join("..", "testdata", "fixtures", "syntax", tt.fixture)

			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, tt.fixture)

			content, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(tmpFile, content, 0644); err != nil {
				t.Fatal(err)
			}

			_, stderr, exitCode := runSortTF(t, tmpFile)

			if tt.wantErr && exitCode == 0 {
				t.Error("expected error but succeeded")
			}

			if !tt.wantErr && exitCode != 0 {
				t.Errorf("unexpected error\nstderr: %s", stderr)
			}
		})
	}
}

// ========================================
// Phase 5: System Tests
// ========================================

// TestSystem_CLIFlagCombinations tests various CLI flag combinations
func TestSystem_CLIFlagCombinations(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		setup    func(t *testing.T) string
		wantExit int
		check    func(t *testing.T, dir string, stdout, stderr string)
	}{
		{
			name:  "dry-run with verbose",
			flags: []string{"--dry-run", "--verbose"},
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				file := filepath.Join(tmpDir, "test.tf")
				content := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-123"
}`
				_ = os.WriteFile(file, []byte(content), 0644)
				return file
			},
			wantExit: 0,
			check: func(t *testing.T, path string, stdout, stderr string) {
				if !strings.Contains(stdout, "would be updated") && !strings.Contains(stdout, "Would update") {
					t.Errorf("Expected dry-run output, got: %s", stdout)
				}
				// File should not be modified
				content, _ := os.ReadFile(path)
				if !strings.Contains(string(content), "instance_type = \"t2.micro\"") {
					t.Error("File was modified in dry-run mode")
				}
			},
		},
		{
			name:  "validate with verbose",
			flags: []string{"--validate", "--verbose"},
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				file := filepath.Join(tmpDir, "test.tf")
				content := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-123"
}`
				_ = os.WriteFile(file, []byte(content), 0644)
				return file
			},
			wantExit: 1, // Should fail validation (file needs sorting)
			check: func(t *testing.T, path string, stdout, stderr string) {
				if !strings.Contains(stderr, "needs") && !strings.Contains(stdout, "needs") {
					t.Logf("Expected validation failure message, got stdout: %s, stderr: %s", stdout, stderr)
				}
			},
		},
		{
			name:  "recursive with verbose",
			flags: []string{"--recursive", "--verbose"},
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				subdir := filepath.Join(tmpDir, "modules")
				_ = os.Mkdir(subdir, 0755)

				file1 := filepath.Join(tmpDir, "main.tf")
				file2 := filepath.Join(subdir, "vpc.tf")
				content := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-123"
}`
				_ = os.WriteFile(file1, []byte(content), 0644)
				_ = os.WriteFile(file2, []byte(content), 0644)
				return tmpDir
			},
			wantExit: 0,
			check: func(t *testing.T, dir string, stdout, stderr string) {
				if !strings.Contains(stdout, "Processed 2 files") {
					t.Errorf("Expected 2 files processed, got: %s", stdout)
				}
			},
		},
		{
			name:  "recursive dry-run with verbose",
			flags: []string{"--recursive", "--dry-run", "--verbose"},
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				file := filepath.Join(tmpDir, "test.tf")
				content := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-123"
}`
				_ = os.WriteFile(file, []byte(content), 0644)
				return tmpDir
			},
			wantExit: 0,
			check: func(t *testing.T, dir string, stdout, stderr string) {
				if !strings.Contains(stdout, "would be updated") && !strings.Contains(stdout, "Would update") {
					t.Errorf("Expected dry-run indicator, got: %s", stdout)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			args := append(tt.flags, path)

			stdout, stderr, exitCode := runSortTF(t, args...)

			if exitCode != tt.wantExit {
				t.Errorf("expected exit code %d, got %d\nstdout: %s\nstderr: %s",
					tt.wantExit, exitCode, stdout, stderr)
			}

			if tt.check != nil {
				tt.check(t, path, stdout, stderr)
			}
		})
	}
}

// TestSystem_MultiFileWorkflows tests complex multi-file scenarios
func TestSystem_MultiFileWorkflows(t *testing.T) {
	t.Run("mixed_sorted_and_unsorted", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create sorted file
		sorted := filepath.Join(tmpDir, "sorted.tf")
		sortedContent := `variable "region" {
  type = string
}

resource "aws_instance" "web" {
  ami           = "ami-123"
  instance_type = "t2.micro"
}
`
		_ = os.WriteFile(sorted, []byte(sortedContent), 0644)

		// Create unsorted file
		unsorted := filepath.Join(tmpDir, "unsorted.tf")
		unsortedContent := `resource "aws_s3_bucket" "data" {
  bucket = "my-bucket"
}

variable "env" {
  type = string
}
`
		_ = os.WriteFile(unsorted, []byte(unsortedContent), 0644)

		stdout, _, exitCode := runSortTF(t, tmpDir)

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		// Should process at least 1 file (unsorted one), sorted file may show as already sorted
		if !strings.Contains(stdout, "Processed") {
			t.Errorf("expected files to be processed, got: %s", stdout)
		}
	})

	t.Run("nested_directory_structure", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create nested structure: root/modules/vpc/main.tf, root/modules/sg/main.tf
		_ = os.MkdirAll(filepath.Join(tmpDir, "modules", "vpc"), 0755)
		_ = os.MkdirAll(filepath.Join(tmpDir, "modules", "sg"), 0755)
		_ = os.MkdirAll(filepath.Join(tmpDir, "environments", "prod"), 0755)

		files := []string{
			filepath.Join(tmpDir, "main.tf"),
			filepath.Join(tmpDir, "modules", "vpc", "main.tf"),
			filepath.Join(tmpDir, "modules", "sg", "main.tf"),
			filepath.Join(tmpDir, "environments", "prod", "main.tf"),
		}

		content := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-123"
}`

		for _, file := range files {
			_ = os.WriteFile(file, []byte(content), 0644)
		}

		stdout, _, exitCode := runSortTF(t, "--recursive", tmpDir)

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		if !strings.Contains(stdout, "Processed 4 files") {
			t.Errorf("expected 4 files processed, got: %s", stdout)
		}
	})

	t.Run("mixed_tf_and_hcl_files", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfFile := filepath.Join(tmpDir, "main.tf")
		hclFile := filepath.Join(tmpDir, "terragrunt.hcl")

		tfContent := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-123"
}`

		hclContent := `terraform {
  source = "git::https://example.com/module"
}

include "root" {
  path = find_in_parent_folders()
}`

		_ = os.WriteFile(tfFile, []byte(tfContent), 0644)
		_ = os.WriteFile(hclFile, []byte(hclContent), 0644)

		stdout, _, exitCode := runSortTF(t, tmpDir)

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		if !strings.Contains(stdout, "Processed") {
			t.Errorf("expected files to be processed, got: %s", stdout)
		}
	})

	t.Run("with_terragrunt_cache_directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create .terragrunt-cache directory (should be skipped)
		cacheDir := filepath.Join(tmpDir, ".terragrunt-cache")
		_ = os.Mkdir(cacheDir, 0755)

		// File in cache (should be ignored)
		cachedFile := filepath.Join(cacheDir, "cached.tf")
		_ = os.WriteFile(cachedFile, []byte("# should be ignored"), 0644)

		// Regular file
		mainFile := filepath.Join(tmpDir, "main.tf")
		content := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-123"
}`
		_ = os.WriteFile(mainFile, []byte(content), 0644)

		stdout, _, exitCode := runSortTF(t, "--recursive", tmpDir)

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		// Should only process main.tf, not cached.tf
		if !strings.Contains(stdout, "Processed 1 files") {
			t.Errorf("expected 1 file processed (cache should be skipped), got: %s", stdout)
		}
	})
}

// TestSystem_ErrorScenarios tests various error conditions
func TestSystem_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) []string // returns args
		wantExit   int
		checkError func(t *testing.T, stderr string)
	}{
		{
			name: "invalid_flag",
			setup: func(t *testing.T) []string {
				return []string{"--invalid-flag", "test.tf"}
			},
			wantExit: 2,
			checkError: func(t *testing.T, stderr string) {
				// Should show usage error
			},
		},
		{
			name: "file_does_not_exist",
			setup: func(t *testing.T) []string {
				return []string{"/nonexistent/file.tf"}
			},
			wantExit: 1,
			checkError: func(t *testing.T, stderr string) {
				if !strings.Contains(stderr, "does not exist") && !strings.Contains(stderr, "not exist") {
					t.Errorf("expected 'does not exist' error, got: %s", stderr)
				}
			},
		},
		{
			name: "directory_does_not_exist",
			setup: func(t *testing.T) []string {
				return []string{"/nonexistent/directory"}
			},
			wantExit: 1,
			checkError: func(t *testing.T, stderr string) {
				if !strings.Contains(stderr, "does not exist") && !strings.Contains(stderr, "not exist") {
					t.Errorf("expected 'does not exist' error, got: %s", stderr)
				}
			},
		},
		{
			name: "invalid_hcl_syntax",
			setup: func(t *testing.T) []string {
				tmpDir := t.TempDir()
				file := filepath.Join(tmpDir, "invalid.tf")
				// Unclosed brace
				_ = os.WriteFile(file, []byte(`resource "aws_instance" "web" {
  ami = "ami-123"
`), 0644)
				return []string{file}
			},
			wantExit: 1,
			checkError: func(t *testing.T, stderr string) {
				if !strings.Contains(stderr, "error") && !strings.Contains(stderr, "Error") {
					t.Logf("Expected error message, got: %s", stderr)
				}
			},
		},
		{
			name: "empty_directory",
			setup: func(t *testing.T) []string {
				tmpDir := t.TempDir()
				return []string{tmpDir}
			},
			wantExit: 0, // Empty directory is not an error
			checkError: func(t *testing.T, stderr string) {
				// Should succeed with 0 files
			},
		},
		{
			name: "validation_failure",
			setup: func(t *testing.T) []string {
				tmpDir := t.TempDir()
				file := filepath.Join(tmpDir, "test.tf")
				// Unsorted file
				_ = os.WriteFile(file, []byte(`resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-123"
}`), 0644)
				return []string{"--validate", file}
			},
			wantExit: 1,
			checkError: func(t *testing.T, stderr string) {
				// Validation should report file needs sorting
			},
		},
		{
			name: "conflicting_flags",
			setup: func(t *testing.T) []string {
				tmpDir := t.TempDir()
				file := filepath.Join(tmpDir, "test.tf")
				_ = os.WriteFile(file, []byte(`variable "x" {}`), 0644)
				// Both validate and dry-run
				return []string{"--validate", "--dry-run", file}
			},
			wantExit: 0, // Dry-run takes precedence and returns 0
			checkError: func(t *testing.T, stderr string) {
				// Should handle conflicting modes
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.setup(t)
			_, stderr, exitCode := runSortTF(t, args...)

			if exitCode != tt.wantExit {
				t.Errorf("expected exit code %d, got %d\nstderr: %s", tt.wantExit, exitCode, stderr)
			}

			if tt.checkError != nil {
				tt.checkError(t, stderr)
			}
		})
	}
}

// TestSystem_ExitCodes validates exit code behavior
func TestSystem_ExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) []string
		wantExit int
		desc     string
	}{
		{
			name: "success",
			setup: func(t *testing.T) []string {
				tmpDir := t.TempDir()
				file := filepath.Join(tmpDir, "test.tf")
				_ = os.WriteFile(file, []byte(`resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-123"
}`), 0644)
				return []string{file}
			},
			wantExit: 0,
			desc:     "successful processing",
		},
		{
			name: "help_flag",
			setup: func(t *testing.T) []string {
				return []string{"--help"}
			},
			wantExit: 0,
			desc:     "help should exit with 0",
		},
		{
			name: "invalid_flag",
			setup: func(t *testing.T) []string {
				return []string{"--not-a-real-flag"}
			},
			wantExit: 2,
			desc:     "invalid flag should exit with 2 (usage error)",
		},
		{
			name: "file_not_found",
			setup: func(t *testing.T) []string {
				return []string{"/tmp/does-not-exist-" + t.Name() + ".tf"}
			},
			wantExit: 1,
			desc:     "file not found should exit with 1",
		},
		{
			name: "parse_error",
			setup: func(t *testing.T) []string {
				tmpDir := t.TempDir()
				file := filepath.Join(tmpDir, "invalid.tf")
				_ = os.WriteFile(file, []byte(`resource "aws_instance" "web" {`), 0644)
				return []string{file}
			},
			wantExit: 1,
			desc:     "parse error should exit with 1",
		},
		{
			name: "validate_needs_sorting",
			setup: func(t *testing.T) []string {
				tmpDir := t.TempDir()
				file := filepath.Join(tmpDir, "test.tf")
				_ = os.WriteFile(file, []byte(`resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-123"
}`), 0644)
				return []string{"--validate", file}
			},
			wantExit: 1,
			desc:     "validation failure should exit with 1",
		},
		{
			name: "validate_already_sorted",
			setup: func(t *testing.T) []string {
				tmpDir := t.TempDir()
				file := filepath.Join(tmpDir, "sorted.tf")
				// Pre-sorted content
				_ = os.WriteFile(file, []byte(`variable "region" {
  type = string
}

resource "aws_instance" "web" {
  ami           = "ami-123"
  instance_type = "t2.micro"
}
`), 0644)
				return []string{"--validate", file}
			},
			wantExit: 0,
			desc:     "validation success should exit with 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.setup(t)
			_, _, exitCode := runSortTF(t, args...)

			if exitCode != tt.wantExit {
				t.Errorf("%s: expected exit code %d, got %d", tt.desc, tt.wantExit, exitCode)
			}
		})
	}
}

// TestSystem_LargeDirectoryPerformance tests performance with many files
func TestSystem_LargeDirectoryPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Run("50_files", func(t *testing.T) {
		tmpDir := t.TempDir()

		content := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-123"
}

variable "region" {
  type = string
}`

		// Create 50 files
		for i := 0; i < 50; i++ {
			file := filepath.Join(tmpDir, fmt.Sprintf("file%03d.tf", i))
			_ = os.WriteFile(file, []byte(content), 0644)
		}

		stdout, _, exitCode := runSortTF(t, tmpDir)

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		if !strings.Contains(stdout, "Processed 50 files") {
			t.Errorf("expected 50 files processed, got: %s", stdout)
		}
	})

	t.Run("100_files_recursive", func(t *testing.T) {
		tmpDir := t.TempDir()

		content := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-123"
}`

		// Create 100 files across multiple directories
		for i := 0; i < 10; i++ {
			subdir := filepath.Join(tmpDir, fmt.Sprintf("dir%02d", i))
			_ = os.Mkdir(subdir, 0755)

			for j := 0; j < 10; j++ {
				file := filepath.Join(subdir, fmt.Sprintf("file%02d.tf", j))
				_ = os.WriteFile(file, []byte(content), 0644)
			}
		}

		stdout, _, exitCode := runSortTF(t, "--recursive", tmpDir)

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		if !strings.Contains(stdout, "Processed 100 files") {
			t.Errorf("expected 100 files processed, got: %s", stdout)
		}
	})
}

// TestSystem_SpecialCharactersInPaths tests handling of special characters
func TestSystem_SpecialCharactersInPaths(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "spaces_in_filename",
			filename: "my file.tf",
			wantErr:  false,
		},
		{
			name:     "hyphens_in_filename",
			filename: "my-file.tf",
			wantErr:  false,
		},
		{
			name:     "underscores_in_filename",
			filename: "my_file.tf",
			wantErr:  false,
		},
		{
			name:     "dots_in_filename",
			filename: "my.config.tf",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			file := filepath.Join(tmpDir, tt.filename)

			content := `resource "aws_instance" "web" {
  instance_type = "t2.micro"
  ami = "ami-123"
}`
			_ = os.WriteFile(file, []byte(content), 0644)

			_, _, exitCode := runSortTF(t, file)

			if tt.wantErr && exitCode == 0 {
				t.Error("expected error but succeeded")
			}

			if !tt.wantErr && exitCode != 0 {
				t.Errorf("unexpected error with exit code %d", exitCode)
			}
		})
	}
}
