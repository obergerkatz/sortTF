# Contributing to sortTF

Thank you for your interest in contributing to sortTF! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)

## Code of Conduct

This project follows the standard open-source code of conduct:

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on what is best for the community
- Show empathy towards other community members

## How Can I Contribute?

### Reporting Bugs

Before creating a bug report:
1. Check the [existing issues](https://github.com/obergerkatz/sortTF/issues) to avoid duplicates
2. Ensure you're using the latest version
3. Verify the issue is reproducible

When creating a bug report, include:
- **Clear title**: Summarize the issue
- **Steps to reproduce**: Exact steps to reproduce the behavior
- **Expected behavior**: What you expected to happen
- **Actual behavior**: What actually happened
- **Environment**: OS, Go version, sortTF version
- **Sample files**: Minimal Terraform files that demonstrate the issue

**Example:**
```markdown
### Bug: Nested blocks not sorted correctly

**Environment:**
- sortTF version: v1.0.0
- OS: macOS 14.0
- Go version: 1.23.0

**Steps to Reproduce:**
1. Create a file with nested lifecycle blocks
2. Run `sorttf file.tf`
3. Observe incorrect ordering

**Expected:** Nested blocks should be sorted alphabetically

**Actual:** Nested blocks remain in original order

**Sample File:**
```hcl
resource "aws_instance" "web" {
  lifecycle { ... }
  ami = "ami-123"
}
```
```

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. Include:

- **Clear use case**: Why is this enhancement needed?
- **Proposed solution**: How should it work?
- **Alternatives considered**: What other approaches did you consider?
- **Examples**: Show examples of the proposed behavior

### Contributing Code

We welcome code contributions! Areas where contributions are especially appreciated:

- Bug fixes
- Documentation improvements
- Test coverage improvements
- Performance optimizations
- New features (discuss in an issue first)

## Development Setup

### Prerequisites

- **Go 1.22+** (1.23+ recommended)
- **Git**
- **Make** (optional, for convenience commands)

### Fork and Clone

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/sortTF.git
cd sortTF

# Add upstream remote
git remote add upstream https://github.com/obergerkatz/sortTF.git
```

### Install Dependencies

```bash
# Download Go module dependencies
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Build

```bash
# Build the binary
go build -o sorttf ./cmd/sorttf

# Or use the Makefile (if available)
make build
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./api

# Run integration tests
go test ./integration/...
```

### Run Linter

```bash
# Run golangci-lint
golangci-lint run

# Or via GitHub Actions locally (if using act)
act -j lint
```

## Development Workflow

### 1. Create a Branch

Create a descriptive branch name:

```bash
git checkout -b feature/add-sorting-for-moved-blocks
git checkout -b fix/nested-block-sorting
git checkout -b docs/improve-installation-guide
```

**Branch naming conventions:**
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `test/` - Test improvements
- `refactor/` - Code refactoring
- `perf/` - Performance improvements

### 2. Make Changes

- Write clear, self-documenting code
- Follow Go best practices and idioms
- Add tests for new functionality
- Update documentation as needed
- Keep commits focused and atomic

### 3. Test Thoroughly

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run integration tests
go test ./integration/...

# Check test coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 4. Lint Your Code

```bash
# Run linter
golangci-lint run

# Format code
go fmt ./...

# Check for common mistakes
go vet ./...

# Ensure go.mod is tidy
go mod tidy
```

### 5. Commit Changes

Write clear, descriptive commit messages following [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git commit -m "feat: add support for sorting moved blocks"
git commit -m "fix: correct nested block sorting order"
git commit -m "docs: update installation instructions"
git commit -m "test: add integration tests for terragrunt files"
```

**Commit message format:**
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions or changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

**Example:**
```
feat(hcl): add support for import blocks

Add sorting support for Terraform 1.5+ import blocks.
Import blocks are now sorted between data and resource blocks.

Closes #123
```

### 6. Push and Create Pull Request

```bash
# Push to your fork
git push origin feature/add-sorting-for-moved-blocks

# Create pull request on GitHub
```

## Coding Standards

### Go Code Style

Follow standard Go conventions:

```go
// Good: Clear, idiomatic Go
func SortFile(path string, opts Options) error {
    if path == "" {
        return fmt.Errorf("path cannot be empty")
    }

    content, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("failed to read file: %w", err)
    }

    // ... implementation
    return nil
}

// Bad: Non-idiomatic, unclear
func sortfile(p string, o Options) error {
    content, err := ioutil.ReadFile(p) // Deprecated
    if err != nil {
        return err // Lost context
    }
    // ...
}
```

### Formatting

- Use `go fmt` for formatting (or `gofmt -s`)
- Use tabs for indentation
- Keep lines under 100 characters when practical
- Group imports: stdlib, third-party, local

```go
import (
    // Standard library
    "fmt"
    "os"

    // Third-party
    "github.com/hashicorp/hcl/v2"
    "github.com/hashicorp/hcl/v2/hclwrite"

    // Local packages
    "sorttf/internal/errors"
)
```

### Naming Conventions

- **Packages**: Short, lowercase, no underscores (`api`, `hcl`, not `api_client`)
- **Exported functions**: PascalCase (`SortFile`, `GetSortedContent`)
- **Unexported functions**: camelCase (`parseBlocks`, `sortAttributes`)
- **Constants**: PascalCase or ALL_CAPS for special cases
- **Interfaces**: End with `er` suffix when appropriate (`Sorter`, `Parser`)

### Error Handling

Always provide context in errors:

```go
// Good
if err := processFile(path); err != nil {
    return fmt.Errorf("failed to process %s: %w", path, err)
}

// Bad
if err := processFile(path); err != nil {
    return err
}
```

Use sentinel errors for expected conditions:

```go
var ErrNoChanges = errors.New("file is already sorted")

// Check with errors.Is()
if errors.Is(err, ErrNoChanges) {
    // Handle expected condition
}
```

### Documentation

Document all exported functions, types, and constants:

```go
// SortFile sorts and formats a single Terraform or Terragrunt file.
//
// It reads the file, parses the HCL, sorts blocks and attributes according
// to Terraform best practices, and writes the result back to disk.
//
// Returns ErrNoChanges if the file is already sorted. Other errors indicate
// parse failures or I/O errors.
func SortFile(path string, opts Options) error {
    // ...
}
```

### Testing

- Write table-driven tests when testing multiple scenarios
- Use descriptive test names
- Test both success and failure cases
- Use testdata directory for test fixtures

```go
func TestSortFile(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name: "sorts blocks in correct order",
            input: "resource {} provider {}",
            want: "provider {} resource {}",
        },
        {
            name:    "returns error on invalid HCL",
            input:   "invalid {{{",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Testing Guidelines

### Test Coverage

Aim for high test coverage:
- **Target**: 90%+ coverage
- **Current**: 95% (let's maintain it!)
- Check coverage: `go test -cover ./...`

### Types of Tests

#### Unit Tests
Test individual functions in isolation:

```go
// api/sorttf_test.go
func TestSortBlocks(t *testing.T) {
    // Test sortBlocks function
}
```

#### Integration Tests
Test the full binary end-to-end:

```go
// integration/integration_test.go
func TestCLI_SortsSingleFile(t *testing.T) {
    // Build and execute actual binary
}
```

#### Table-Driven Tests
Use for testing multiple scenarios:

```go
func TestBlockOrder(t *testing.T) {
    tests := []struct {
        name     string
        input    []string
        expected []string
    }{
        {"provider before resource", []string{"resource", "provider"}, []string{"provider", "resource"}},
        {"variable before data", []string{"data", "variable"}, []string{"variable", "data"}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

### Test Fixtures

Place test files in `testdata/`:

```
testdata/
├── valid/
│   ├── simple.tf
│   ├── complex.tf
│   └── nested.tf
├── invalid/
│   └── syntax_error.tf
└── expected/
    ├── simple_sorted.tf
    └── complex_sorted.tf
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./api

# Specific test
go test ./api -run TestSortFile

# With coverage
go test -cover ./...

# With race detector
go test -race ./...

# Verbose output
go test -v ./...

# Integration tests only
go test ./integration/...

# Short mode (skip slow tests)
go test -short ./...
```

## Submitting Changes

### Pull Request Process

1. **Update your branch** with latest upstream:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Ensure all checks pass**:
   - Tests: `go test ./...`
   - Linting: `golangci-lint run`
   - Formatting: `go fmt ./...`
   - Coverage: Maintain or improve coverage

3. **Write a good PR description**:
   - Summarize the change
   - Reference related issues
   - Include examples if applicable
   - Note any breaking changes

**PR Template:**
```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Related Issues
Closes #123

## Testing
Describe the tests you added or how you tested the changes

## Checklist
- [ ] Tests pass locally
- [ ] Linter passes
- [ ] Added tests for new functionality
- [ ] Updated documentation
- [ ] Followed coding standards
```

### Review Process

- A maintainer will review your PR
- Address any feedback or requested changes
- Be patient and respectful during review
- Once approved, a maintainer will merge your PR

### What to Expect

- **Initial review**: Within 1-3 business days
- **Feedback**: Constructive feedback on code quality, design, testing
- **Iterations**: You may be asked to make changes
- **Approval**: Once approved, your PR will be merged
- **Credit**: You'll be credited in the changelog and commit history

## First-Time Contributors

New to open source? Welcome! Here's how to get started:

1. **Look for "good first issue" labels**
2. **Comment on the issue** to let others know you're working on it
3. **Ask questions** if anything is unclear
4. **Start small** - documentation improvements are great first contributions
5. **Don't be afraid to ask for help**

## Questions?

- **General questions**: Open a [GitHub Discussion](https://github.com/obergerkatz/sortTF/discussions)
- **Bug reports**: Open an [Issue](https://github.com/obergerkatz/sortTF/issues)
- **Security issues**: See [SECURITY.md](../SECURITY.md) if available, or email maintainers privately

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to sortTF! 🎉
