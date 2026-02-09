# Testdata Directory Instructions

**Scope**: Applies only to `testdata/` directory.

This file **extends** the repository-level `.claude/CLAUDE.md`. Read that first for global standards. This file contains only what's **specific** to test fixtures and data.

---

## Purpose

The `testdata/` directory contains **test fixtures** used by unit and integration tests:

- Example Terraform files (.tf)
- Example Terragrunt files (.hcl)
- Invalid/malformed files for error testing
- Expected output files for comparison
- Test directory structures

**Philosophy**:

- Real-world examples
- Cover common patterns and edge cases
- Well-organized and documented
- Committed to repository (part of tests)

---

## Directory Structure

```text
testdata/
├── valid/              # Valid, well-formed Terraform files
│   ├── simple.tf       # Minimal valid file
│   ├── complex.tf      # Complex with nested blocks
│   ├── sorted.tf       # Already sorted
│   └── unsorted.tf     # Needs sorting
├── invalid/            # Invalid files for error testing
│   ├── syntax_error.tf # HCL syntax errors
│   ├── empty.tf        # Empty file
│   └── malformed.hcl   # Malformed HCL
├── expected/           # Expected output files
│   └── sorted/
│       ├── simple.tf   # Expected result for simple.tf
│       └── complex.tf  # Expected result for complex.tf
└── CLAUDE.md           # This file
```

---

## Test Fixture Categories

### 1. Valid Files (valid/)

**Purpose**: Test correct parsing and sorting of valid Terraform/Terragrunt files.

**Files**:

- `simple.tf`: Minimal valid file (1-2 blocks)
- `complete.tf`: All block types (terraform, provider, variable, locals, data, resource, module, output)
- `sorted.tf`: Already correctly sorted
- `unsorted.tf`: Needs sorting (blocks out of order)
- `nested.tf`: Complex nested block structures
- `attributes.tf`: Various attribute types and ordering
- `terragrunt.hcl`: Terragrunt configuration file

**Example - simple.tf**:

```hcl
resource "aws_instance" "web" {
  ami           = "ami-123456"
  instance_type = "t3.micro"
}

provider "aws" {
  region = "us-west-2"
}
```

**Example - complete.tf**:

```hcl
terraform {
  required_version = ">= 1.0"
}

provider "aws" {
  region = "us-west-2"
}

variable "environment" {
  type = string
}

locals {
  tags = {
    Environment = var.environment
  }
}

data "aws_ami" "ubuntu" {
  most_recent = true
}

resource "aws_instance" "web" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"
}

module "vpc" {
  source = "./modules/vpc"
}

output "instance_id" {
  value = aws_instance.web.id
}
```

### 2. Invalid Files (invalid/)

**Purpose**: Test error handling and parsing failures.

**Files**:

- `syntax_error.tf`: HCL syntax errors (unclosed strings, missing braces)
- `empty.tf`: Completely empty file
- `whitespace_only.tf`: Only whitespace
- `invalid_block.tf`: Invalid block structure
- `malformed.hcl`: Malformed Terragrunt configuration

**Example - syntax_error.tf**:

```hcl
resource "aws_instance" "web" {
  ami = "unclosed-string
  instance_type = "t3.micro"
}
```

**Example - invalid_block.tf**:

```hcl
resource {
  # Missing resource type and name
  ami = "ami-123"
}
```

### 3. Expected Output (expected/)

**Purpose**: Reference files for comparison in tests.

**Structure**:

```text
expected/
└── sorted/
    ├── simple.tf       # Expected result for valid/simple.tf
    ├── complete.tf     # Expected result for valid/complete.tf
    └── unsorted.tf     # Expected result for valid/unsorted.tf
```

**Usage**:

```go
func TestSort_Simple(t *testing.T) {
 input, _ := os.ReadFile("testdata/valid/simple.tf")
 expected, _ := os.ReadFile("testdata/expected/sorted/simple.tf")

 result, err := SortContent(input)
 if err != nil {
  t.Fatal(err)
 }

 if !bytes.Equal(result, expected) {
  t.Errorf("output mismatch:\nGot:\n%s\nExpected:\n%s",
   result, expected)
 }
}
```

---

## Test Fixture Guidelines

### 1. Real-World Examples

Use realistic Terraform patterns:

```hcl
# GOOD: Real AWS resource configuration
resource "aws_instance" "web" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t3.micro"

  tags = {
    Name        = "web-server"
    Environment = "production"
  }
}

# BAD: Overly simplified or unrealistic
resource "foo" "bar" {
  x = 1
}
```

### 2. Cover Edge Cases

Include edge cases in fixtures:

- Empty blocks
- Blocks with many attributes
- Deeply nested blocks
- Special characters in strings
- Multi-line strings
- Comments (to verify they're removed)

### 3. Self-Documenting Names

File names should indicate their purpose:

**✅ Good names**:

- `unsorted_resources.tf` (clear: resources are unsorted)
- `already_sorted.tf` (clear: file is already correct)
- `syntax_error_unclosed_string.tf` (clear: what the error is)

**❌ Bad names**:

- `test1.tf` (unclear)
- `file.tf` (too generic)
- `abc.tf` (meaningless)

### 4. Minimal but Complete

Keep fixtures minimal while testing the specific case:

```hcl
# GOOD: Minimal file for testing block order
resource "aws_instance" "web" {}
provider "aws" {}

# BAD: Unnecessary complexity for simple test
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}
provider "aws" {
  region = "us-west-2"
}
resource "aws_instance" "web" {
  ami = "ami-123"
  # ... many more attributes
}
```

### 5. Document Special Cases

Add comments to explain non-obvious fixtures:

```hcl
# This file tests attribute ordering within a resource block.
# The 'for_each' attribute should be sorted first, then others alphabetically.
resource "aws_instance" "web" {
  for_each = var.instances  # Special attribute (should be first)

  ami           = "ami-123"  # Regular attributes (alphabetical)
  instance_type = "t3.micro"
  tags          = {}
}
```

---

## Using Test Fixtures

### Reading Fixtures

```go
func loadFixture(t *testing.T, path string) []byte {
 t.Helper()
 content, err := os.ReadFile(path)
 if err != nil {
  t.Fatalf("failed to read fixture %s: %v", path, err)
 }
 return content
}

func TestParse(t *testing.T) {
 input := loadFixture(t, "testdata/valid/simple.tf")
 // ... test with input
}
```

### Comparing with Expected Output

```go
func compareWithExpected(t *testing.T, got []byte, expectedPath string) {
 t.Helper()
 expected := loadFixture(t, expectedPath)
 if !bytes.Equal(got, expected) {
  t.Errorf("output mismatch:\nGot:\n%s\n\nExpected:\n%s",
   got, expected)
 }
}

func TestSort(t *testing.T) {
 input := loadFixture(t, "testdata/valid/unsorted.tf")
 result, err := Sort(input)
 if err != nil {
  t.Fatal(err)
 }
 compareWithExpected(t, result, "testdata/expected/sorted/unsorted.tf")
}
```

### Table-Driven Tests with Fixtures

```go
func TestSort_AllCases(t *testing.T) {
 tests := []struct {
  name     string
  input    string
  expected string
 }{
  {
   name:     "simple",
   input:    "testdata/valid/simple.tf",
   expected: "testdata/expected/sorted/simple.tf",
  },
  {
   name:     "complex",
   input:    "testdata/valid/complex.tf",
   expected: "testdata/expected/sorted/complex.tf",
  },
 }

 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   input := loadFixture(t, tt.input)
   expected := loadFixture(t, tt.expected)

   result, err := Sort(input)
   if err != nil {
    t.Fatal(err)
   }

   if !bytes.Equal(result, expected) {
    t.Errorf("mismatch for %s", tt.name)
   }
  })
 }
}
```

---

## Adding New Fixtures

### Steps

1. **Identify test case**: What are you testing?
2. **Choose category**: valid/, invalid/, or expected/?
3. **Create file**: Use descriptive name
4. **Add content**: Minimal but complete example
5. **Add expected output** (if applicable)
6. **Document** (if non-obvious)
7. **Use in tests**: Reference from test code

### Example

```bash
# Create new fixture for testing nested blocks
cat > testdata/valid/nested_blocks.tf << 'EOF'
resource "aws_instance" "web" {
  ami = "ami-123"

  lifecycle {
    create_before_destroy = true
  }

  connection {
    type = "ssh"
    user = "ubuntu"
  }
}
EOF

# Create expected output
cat > testdata/expected/sorted/nested_blocks.tf << 'EOF'
resource "aws_instance" "web" {
  ami = "ami-123"

  connection {
    type = "ssh"
    user = "ubuntu"
  }

  lifecycle {
    create_before_destroy = true
  }
}
EOF
```

---

## Maintenance

### Keep Fixtures Up to Date

When code behavior changes:

1. Update affected fixtures
2. Update expected output files
3. Verify tests still pass

### Remove Unused Fixtures

Regularly audit and remove:

- Fixtures not referenced by any test
- Duplicate fixtures
- Outdated examples

### Validate Fixtures

Occasionally validate fixtures with real Terraform:

```bash
terraform fmt -check testdata/valid/*.tf
terraform validate testdata/valid/
```

---

## Best Practices

### ✅ Do

- Use realistic Terraform examples
- Keep fixtures minimal
- Organize by category (valid/invalid/expected)
- Use descriptive names
- Document non-obvious cases
- Commit fixtures to repository
- Test fixtures in CI

### ❌ Don't

- Include sensitive data (API keys, passwords)
- Use overly complex examples when simple ones suffice
- Create fixtures without using them in tests
- Use random or meaningless content
- Include large files (keep under 100 lines)

---

## Acceptance Checklist (Testdata)

Before adding or modifying test fixtures:

- [ ] Fixture has clear, descriptive name
- [ ] File is in appropriate category (valid/invalid/expected)
- [ ] Content is minimal but complete for the test case
- [ ] Non-obvious cases are documented with comments
- [ ] Expected output file created (if applicable)
- [ ] Fixture is referenced in at least one test
- [ ] No sensitive data included
- [ ] File is properly formatted (consistent with Terraform style)
- [ ] Test using the fixture passes
