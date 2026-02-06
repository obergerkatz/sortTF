# Usage Guide

Comprehensive guide to using sortTF from the command line.

## Table of Contents

- [Quick Reference](#quick-reference)
- [Command-Line Options](#command-line-options)
- [Basic Usage](#basic-usage)
- [Advanced Usage](#advanced-usage)
- [Use Cases](#use-cases)
- [Exit Codes](#exit-codes)
- [Sorting Behavior](#sorting-behavior)
- [Tips and Best Practices](#tips-and-best-practices)

## Quick Reference

```bash
sorttf [OPTIONS] [PATH...]

# Most common commands
sorttf .                    # Sort files in current directory
sorttf main.tf             # Sort specific file
sorttf --recursive .       # Sort recursively
sorttf --dry-run .         # Preview changes
sorttf --validate .        # Check if sorted (CI/CD)
```

## Command-Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `--recursive`, `-r` | Process directories recursively | `false` |
| `--dry-run`, `-n` | Show changes without modifying files | `false` |
| `--validate`, `-c` | Exit with error if files need sorting | `false` |
| `--verbose`, `-v` | Print detailed processing information | `false` |
| `--help`, `-h` | Show help message | - |
| `--version` | Show version information | - |

## Basic Usage

### Sort a Single File

```bash
sorttf main.tf
```

**Output:**
```
✓ Sorted: main.tf
```

### Sort Current Directory

```bash
sorttf .
```

Processes all `.tf` and `.hcl` files in the current directory (non-recursive).

### Sort Multiple Files

```bash
sorttf main.tf variables.tf outputs.tf
```

### Sort Specific Directory

```bash
sorttf ./terraform/environments/prod
```

## Advanced Usage

### Recursive Processing

Process all subdirectories:

```bash
sorttf --recursive .
```

**Example output:**
```
✓ Sorted: ./main.tf
✓ Sorted: ./modules/vpc/main.tf
✓ Sorted: ./modules/vpc/variables.tf
✓ Sorted: ./environments/prod/main.tf
⚠ Skipped: ./environments/dev/main.tf (already sorted)
```

### Dry Run (Preview Changes)

Preview what would change without modifying files:

```bash
sorttf --dry-run main.tf
```

**Output shows a unified diff:**
```diff
--- main.tf (original)
+++ main.tf (sorted)
@@ -1,10 +1,10 @@
+provider "aws" {
+  region = "us-west-2"
+}
+
 variable "environment" {
   type = string
 }

-provider "aws" {
-  region = "us-west-2"
-}
-
 resource "aws_instance" "web" {
   ami           = "ami-123456"
   instance_type = "t3.micro"
```

### Validate Mode (CI/CD)

Check if files are sorted without modifying them:

```bash
sorttf --validate .
```

**Exit codes:**
- `0`: All files are properly sorted
- `1`: One or more files need sorting

**Example usage in CI:**
```bash
if ! sorttf --validate --recursive .; then
  echo "❌ Some files need sorting. Run: sorttf --recursive ."
  exit 1
fi
```

### Verbose Output

Get detailed information about processing:

```bash
sorttf --verbose --recursive .
```

**Output:**
```
[INFO] Processing directory: .
[INFO] Found 15 Terraform files
[INFO] Processing: main.tf
[DEBUG] Parsed 8 blocks
[DEBUG] Sorted blocks: provider, variable, data, resource
[INFO] ✓ Sorted: main.tf (8 blocks, 24 attributes)
[INFO] Processing: variables.tf
[INFO] ⚠ Skipped: variables.tf (already sorted)
...
[INFO] Summary: 12 sorted, 3 skipped, 0 errors
```

### Combining Flags

```bash
# Dry-run with verbose output, recursive
sorttf --dry-run --verbose --recursive .

# Validate with recursive processing
sorttf --validate --recursive .

# Dry-run on specific files
sorttf --dry-run main.tf variables.tf outputs.tf
```

## Use Cases

### Pre-commit Hook

Sort files before committing:

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Get staged Terraform files
FILES=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(tf|hcl)$')

if [ -n "$FILES" ]; then
  echo "Sorting Terraform files..."
  sorttf $FILES

  # Re-stage sorted files
  git add $FILES
fi
```

### CI/CD Pipeline

**GitHub Actions:**
```yaml
- name: Validate Terraform formatting
  run: |
    sorttf --validate --recursive .
```

**GitLab CI:**
```yaml
terraform-format:
  script:
    - sorttf --validate --recursive .
```

**Jenkins:**
```groovy
stage('Validate Terraform') {
  steps {
    sh 'sorttf --validate --recursive .'
  }
}
```

### Pre-push Hook

Validate before pushing:

```bash
#!/bin/bash
# .git/hooks/pre-push

echo "Validating Terraform file formatting..."
if ! sorttf --validate --recursive .; then
  echo "❌ Terraform files are not properly sorted."
  echo "Run: sorttf --recursive ."
  exit 1
fi
```

### Makefile Integration

```makefile
.PHONY: fmt fmt-check

fmt:
	@echo "Formatting Terraform files..."
	@sorttf --recursive .

fmt-check:
	@echo "Checking Terraform file formatting..."
	@sorttf --validate --recursive .
```

Usage:
```bash
make fmt        # Sort all files
make fmt-check  # Validate in CI
```

### IDE Integration

#### VS Code Task

`.vscode/tasks.json`:
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Sort Terraform Files",
      "type": "shell",
      "command": "sorttf",
      "args": ["--recursive", "."],
      "problemMatcher": [],
      "group": {
        "kind": "build",
        "isDefault": true
      }
    }
  ]
}
```

#### JetBrains IDEs (IntelliJ, GoLand)

External Tools configuration:
- **Program**: `sorttf`
- **Arguments**: `--recursive $ProjectFileDir$`
- **Working directory**: `$ProjectFileDir$`

### Monorepo with Multiple Terraform Projects

```bash
# Sort each project independently
for dir in terraform/*/; do
  echo "Processing $dir"
  sorttf --recursive "$dir"
done

# Or using find
find terraform -type d -name '*.tf' -exec dirname {} \; | sort -u | while read dir; do
  sorttf --recursive "$dir"
done
```

### Terragrunt Projects

sortTF works with `.hcl` files:

```bash
# Sort Terragrunt configuration
sorttf terragrunt.hcl

# Recursively sort all Terragrunt files
sorttf --recursive . # Processes both .tf and .hcl
```

### Docker Usage

```bash
# Sort files in current directory
docker run --rm -v $(pwd):/workspace ghcr.io/obergerkatz/sorttf:latest .

# Recursive with dry-run
docker run --rm -v $(pwd):/workspace ghcr.io/obergerkatz/sorttf:latest --recursive --dry-run .

# Validate
docker run --rm -v $(pwd):/workspace ghcr.io/obergerkatz/sorttf:latest --validate --recursive .
```

## Exit Codes

sortTF uses standard exit codes:

| Code | Meaning | When It Happens |
|------|---------|----------------|
| `0` | Success | All files processed successfully, or all files already sorted |
| `1` | Error | Parse errors, I/O errors, or files need sorting (in `--validate` mode) |

### Exit Code Examples

```bash
# Success - files were sorted
sorttf main.tf
echo $?  # 0

# Success - files already sorted
sorttf main.tf
echo $?  # 0

# Error - file needs sorting (validate mode)
sorttf --validate main.tf
echo $?  # 1 (if file needs sorting)

# Error - file doesn't exist
sorttf nonexistent.tf
echo $?  # 1
```

## Sorting Behavior

### Block Ordering

Blocks are sorted in this order:

1. `terraform` - Terraform settings
2. `provider` - Provider configurations
3. `variable` - Input variables
4. `locals` - Local values
5. `data` - Data sources
6. `resource` - Resources
7. `module` - Module calls
8. `output` - Output values

Within each type, blocks are sorted alphabetically by their labels.

### Attribute Ordering

Within blocks:
1. `for_each` is always first (if present)
2. `count` is second (if present, after `for_each`)
3. Other attributes are sorted alphabetically
4. Nested blocks come after attributes

**Example:**
```hcl
resource "aws_instance" "web" {
  for_each = var.instances  # Always first

  ami           = "ami-123456"  # Alphabetical
  instance_type = "t3.micro"
  tags          = { Name = "web" }

  lifecycle {  # Nested block after attributes
    create_before_destroy = true
  }
}
```

### Comment Preservation

Comments are preserved but may be repositioned:

```hcl
# This comment stays with the resource
resource "aws_instance" "web" {
  # This comment stays with ami
  ami = "ami-123456"
}
```

### Multiple Files

When sorting multiple files, sortTF processes them concurrently for performance:

```bash
sorttf main.tf variables.tf outputs.tf
```

Each file is processed independently.

### Ignored Files

sortTF skips:
- Non-Terraform files (only processes `.tf` and `.hcl`)
- Files in `.terraform/` directories
- Files in `.terragrunt-cache/` directories
- Hidden directories (starting with `.`)

## Tips and Best Practices

### 1. Always Preview First

Use `--dry-run` on unfamiliar code:
```bash
sorttf --dry-run --recursive .
```

### 2. Use Validate in CI

Catch unsorted files early:
```yaml
- name: Check Terraform formatting
  run: sorttf --validate --recursive .
```

### 3. Sort Before Committing

Add to your workflow:
```bash
sorttf --recursive . && git add -A
```

### 4. Combine with terraform fmt

sortTF complements `terraform fmt`:
```bash
terraform fmt -recursive .  # Format HCL
sorttf --recursive .        # Sort blocks and attributes
```

### 5. Use in Monorepos

Process multiple projects:
```bash
for dir in terraform/*/; do sorttf --recursive "$dir"; done
```

### 6. Check Exit Codes

Handle errors properly:
```bash
if ! sorttf --validate .; then
  echo "Files need sorting"
  exit 1
fi
```

### 7. Verbose for Debugging

Use verbose mode to debug issues:
```bash
sorttf --verbose --dry-run main.tf
```

### 8. Recursive by Default in CI

Always use recursive in CI:
```bash
sorttf --validate --recursive .
```

### 9. Ignore Generated Files

Don't run on generated files:
```bash
# Skip .terraform directory (automatically skipped)
sorttf --recursive .
```

### 10. Document in README

Tell your team about sortTF:
```markdown
## Code Formatting

Run before committing:
```bash
sorttf --recursive .
```

## Troubleshooting

### Files Not Being Sorted

**Check file extension:**
```bash
# Only .tf and .hcl files are processed
ls -la *.tf *.hcl
```

### "Already sorted" but looks wrong

sortTF may already be seeing the file as correct. Use `--dry-run` to see what would change:
```bash
sorttf --dry-run main.tf
```

### Parse Errors

If sortTF reports parse errors:
1. Check for syntax errors with `terraform validate`
2. Run `terraform fmt` first
3. Report the issue on GitHub if it's valid HCL

### Performance

For very large repositories, consider:
- Processing specific directories instead of entire repo
- Using CI parallelization
- Excluding unnecessary directories

## Next Steps

- Learn about using sortTF as a [library in Go programs](API.md)
- Read the [Contributing Guide](CONTRIBUTING.md) to contribute
