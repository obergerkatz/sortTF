# Repository-Level Claude Code Instructions

**Scope**: Applies to the entire `sortTF` repository.

---

## Repository Overview

sortTF is a **command-line tool and Go library** for sorting and formatting Terraform (.tf) and Terragrunt (.hcl) files:

- **cmd/sorttf/**: CLI entry point (main.go)
- **cli/**: CLI logic and flag parsing
- **api/**: Public Go library interface for sorting operations
- **hcl/**: HCL parsing, sorting, and formatting
- **config/**: Configuration structures and validation
- **internal/**: Private utilities (errors, files)
- **integration/**: End-to-end integration tests
- **testdata/**: Test fixtures for unit and integration tests
- **docs/**: Comprehensive documentation (ARCHITECTURE.md, USAGE.md, API.md)

**Tech Stack**:

- Go **1.23+**
- HashiCorp HCL v2 for parsing
- github.com/fatih/color for terminal output
- Native Go testing framework
- GitHub Actions for CI/CD

---

## Working in This Repository

### Prime Directive

- Work on **one file at a time**
- Explain intent and reasoning while coding
- Preserve existing conventions, tooling, and patterns unless explicitly instructed otherwise
- Follow Go idiomatic patterns and standard project layout

### Mandatory Planning Phase

**Required when**:

- Editing files >200 lines
- Refactoring core logic (hcl package, api package)
- Changing public API signatures
- Modifying sorting algorithm behavior
- Touching CI/CD workflows

**Planning format**:

1. Write a plan **before** making edits
2. Include: files involved (exact paths), functions/sections affected, dependency order, test impact, estimated edits
3. Wait for explicit approval before proceeding

**Example**:

```text
## PROPOSED EDIT PLAN
Working with: `hcl/sorter.go`
Total planned edits: 3

### Edit Sequence
1. Add support for moved blocks — Purpose: Terraform 1.5+ compatibility
2. Update block order map — Purpose: define moved block position
3. Add tests for moved blocks — Purpose: coverage

Do you approve? I will proceed with Edit 1 upon confirmation.
```

### Execution Rules

- Apply one conceptual change per edit
- After each edit: show diff snippets, explain why, confirm test/lint impact
- If unexpected complexity emerges: **STOP**, update plan, request approval
- Provide exact shell commands for any manual actions required from the user
- Report progress continuously using checkpoint format (✅ Completed X of Y, ⏭️ Next: ...)

---

## Commands & Tooling

### Development Commands

```bash
# Install dependencies
go mod download

# Build the CLI
go build -o bin/sorttf ./cmd/sorttf

# Run tests (all packages)
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./api
go test ./hcl

# Run integration tests only
go test ./integration

# Run benchmarks
go test -bench=. ./cli

# Check test coverage (detailed)
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Format code (gofmt)
go fmt ./...

# Run linter (if golangci-lint installed)
golangci-lint run

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Testing Commands

```bash
# Run all tests with race detection
go test -race ./...

# Run tests with timeout
go test -timeout 30s ./...

# Run specific test function
go test -run TestSortFile ./api

# Run tests matching pattern
go test -run Sort ./...

# Generate test coverage report
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
```

### Local Usage

```bash
# Build and install locally
go install ./cmd/sorttf

# Run without installing
go run ./cmd/sorttf <args>

# Example: Sort current directory
go run ./cmd/sorttf .

# Example: Dry run mode
go run ./cmd/sorttf --dry-run .

# Example: Validate mode (CI/CD)
go run ./cmd/sorttf --validate .
```

---

## Go Standards (All Packages)

### Language & Tooling

- Go **1.23+**
- Standard library preferred over external dependencies
- Module-aware mode (go.mod)
- Native testing framework
- Table-driven tests

### Project Layout

- Source: top-level packages (api, cli, hcl, config)
- Entry point: `cmd/sorttf/main.go`
- Internal code: `internal/` (not importable by external projects)
- Tests: `*_test.go` files alongside source
- Test data: `testdata/` directory
- Build output: `bin/` (not committed)

### Code Organization

```text
sortTF/
├── cmd/
│   └── sorttf/          # CLI entry point (main package)
├── api/                 # Public library interface
├── cli/                 # CLI logic (flag parsing, execution)
├── hcl/                 # HCL parsing and sorting
├── config/              # Configuration types
├── internal/            # Private utilities
│   ├── errors/          # Error handling
│   └── files/           # File operations
├── integration/         # Integration tests
├── testdata/            # Test fixtures
└── docs/                # Documentation
```

### Type Safety & Best Practices

**Required patterns**:

- Explicit error handling (never ignore errors)
- Use sentinel errors for expected conditions
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Prefer value receivers unless mutation required
- Use interfaces for testability
- Keep functions small and focused

**Rules**:

- No `panic()` in library code (only in tests or main)
- Prefer `errors.Is()` and `errors.As()` over type assertions
- Use context.Context for cancellation and timeouts
- Document all exported functions, types, and constants
- Package-level variables must be const or initialized once

### Documentation (Godoc)

**Required for**:

- All exported functions, types, structs, interfaces
- Package-level documentation (doc.go files)
- Non-obvious internal helpers

**Format**:

```go
// Package api provides a public interface for sorting Terraform and Terragrunt files.
//
// This package offers high-level functions for sorting individual files,
// multiple files, or entire directories. It handles HCL parsing, block sorting,
// and file writing while providing detailed error information.
package api

// SortFile sorts a single Terraform or Terragrunt file.
//
// The function reads the file, parses HCL content, sorts blocks and attributes
// according to best practices, and writes the result back to the file.
//
// Parameters:
//   - path: Absolute or relative path to the file
//   - opts: Configuration options (DryRun, Validate)
//
// Returns:
//   - nil if sorting succeeded or file already sorted (check errors.Is for ErrNoChanges)
//   - ErrNoChanges if file is already correctly sorted
//   - ErrNeedsSorting if in Validate mode and file needs sorting
//   - error for any other failure (parse error, I/O error, etc.)
//
// Example:
//
// err := api.SortFile("main.tf", api.Options{})
// if errors.Is(err, api.ErrNoChanges) {
//     fmt.Println("File already sorted")
// }
func SortFile(path string, opts Options) error {
 // ...
}
```

**Guidelines**:

- Start with a concise summary line
- Elaborate in following paragraphs if needed
- Document parameters and return values
- Provide usage examples for complex functions
- Explain edge cases and special behaviors
- Update comments when behavior changes

### Coding Standards

- Prefer small, pure functions (10-30 lines)
- No package-level mutable state
- Explicit returns (no naked returns)
- Use early returns for error cases
- Group imports: standard library, external, internal
- Constants at package level, uppercase with underscores
- Error variables prefixed with `Err` (e.g., `ErrNoChanges`)

**Import Organization**:

```go
import (
 // Standard library
 "fmt"
 "os"
 "path/filepath"

 // External dependencies
 "github.com/fatih/color"
 "github.com/hashicorp/hcl/v2"

 // Internal packages
 "github.com/obergerkatz/sortTF/internal/errors"
 "github.com/obergerkatz/sortTF/internal/files"
)
```

### Error Handling

- Always check and handle errors explicitly
- Use sentinel errors for expected conditions: `var ErrNoChanges = errors.New("no changes needed")`
- Wrap errors to add context: `fmt.Errorf("failed to parse %s: %w", path, err)`
- Use custom error types for structured errors
- Document which errors functions can return
- Never swallow errors silently

**Sentinel Errors**:

```go
// Define at package level
var (
 ErrNoChanges     = errors.New("file is already sorted")
 ErrNeedsSorting  = errors.New("file needs sorting")
)

// Check with errors.Is
if errors.Is(err, api.ErrNoChanges) {
 // Handle "already sorted" case
}
```

**Custom Error Types**:

```go
// ParseError contains HCL parsing error details
type ParseError struct {
 Path string
 Line int
 Err  error
}

func (e *ParseError) Error() string {
 return fmt.Sprintf("parse error in %s at line %d: %v", e.Path, e.Line, e.Err)
}

func (e *ParseError) Unwrap() error {
 return e.Err
}
```

### Logging and Output

- Use `fmt` package for CLI output
- Use `github.com/fatih/color` for colored terminal output
- Log to stderr for errors and warnings
- Log to stdout for normal output and results
- Library code (api, hcl) should not print directly
- Return errors instead of logging in library code

**CLI Output Patterns**:

```go
// Success message (green)
color.Green("✓ Sorted: %s", path)

// Error message (red)
color.Red("✗ Error: %s", err.Error())

// Warning message (yellow)
color.Yellow("⚠ Warning: %s", message)

// Info message (cyan)
color.Cyan("ℹ Processing: %s", path)
```

### Tests

- **Native Go testing** framework
- One behavior per test function
- **Table-driven tests** for multiple inputs
- **Prefer**:
  - Subtests with `t.Run()` for table tests
  - Test fixtures in `testdata/` directory
  - Temporary directories for file I/O tests
  - Parallel tests when safe (`t.Parallel()`)
- **Validate**:
  - Happy path and error path
  - Edge cases (empty files, malformed input)
  - Sentinel errors with `errors.Is()`
  - Error messages and types
- Tests must be deterministic and order-independent
- Clean up resources (defer cleanup functions)

**Table-Driven Test Pattern**:

```go
func TestSortBlocks(t *testing.T) {
 tests := []struct {
  name    string
  input   []Block
  want    []Block
  wantErr bool
 }{
  {
   name:  "already sorted",
   input: []Block{{Type: "provider"}, {Type: "resource"}},
   want:  []Block{{Type: "provider"}, {Type: "resource"}},
  },
  {
   name:  "needs sorting",
   input: []Block{{Type: "resource"}, {Type: "provider"}},
   want:  []Block{{Type: "provider"}, {Type: "resource"}},
  },
 }

 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   got := SortBlocks(tt.input)
   if !reflect.DeepEqual(got, tt.want) {
    t.Errorf("SortBlocks() = %v, want %v", got, tt.want)
   }
  })
 }
}
```

### Test Coverage

**Target: 90%+ coverage** for all packages:

- Statements: 90%+
- Branches: 85%+
- Functions: 95%+

**Current coverage**: ~95%

**Enforcement**: CI checks coverage on pull requests

**Run coverage locally**:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out  # Visual report
```

---

## Git and Version Control

### Branch Strategy

- **main**: Production-ready code
- **feature/**: Feature branches (e.g., `feature/add-moved-blocks`)
- **fix/**: Bug fix branches (e.g., `fix/parsing-error`)
- **docs/**: Documentation updates

### Commit Messages

Follow conventional commits format:

```text
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types**:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions or changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

**Examples**:

```bash
feat(hcl): add support for moved blocks
fix(api): handle empty file edge case
docs(api): improve SortFile documentation
test(hcl): add tests for nested block sorting
refactor(cli): simplify flag parsing logic
```

### Release Process

See [docs/RELEASING.md](../docs/RELEASING.md) for detailed release process.

**Quick summary**:

1. Update version in code (if needed)
2. Ensure all tests pass
3. Create and push git tag: `git tag -a v1.x.y -m "Release v1.x.y"`
4. Push tag: `git push origin v1.x.y`
5. GitHub Actions builds and publishes release

---

## CI/CD Pipeline

Workflows located in `.github/workflows/`:

### ci.yml - Continuous Integration

**Jobs**:

1. **Lint**: Run golangci-lint
2. **Test**: Run tests on multiple Go versions (1.23, 1.24)
3. **Test Coverage**: Check coverage thresholds
4. **Build**: Verify build succeeds on multiple platforms

**Triggers**: Pull requests, pushes to main

### release.yml - Release Automation

**Jobs**:

1. **Build Binaries**: Build for multiple platforms (Linux, macOS, Windows)
2. **Create Release**: Create GitHub release with binaries
3. **Publish**: Tag release and generate changelog

**Triggers**: Git tags matching `v*`

### dependencies.yml - Dependency Updates

**Jobs**:

1. **Update Go Modules**: Check for dependency updates
2. **Security Scan**: Scan for known vulnerabilities

**Triggers**: Weekly schedule

**Reproduce CI locally**:

```bash
# Run tests like CI
go test -v -race -cover ./...

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o bin/sorttf-linux-amd64 ./cmd/sorttf
GOOS=darwin GOARCH=amd64 go build -o bin/sorttf-darwin-amd64 ./cmd/sorttf
GOOS=windows GOARCH=amd64 go build -o bin/sorttf-windows-amd64.exe ./cmd/sorttf
```

---

## HCL and Terraform Specifics

### Supported Block Types

```go
var blockOrder = map[string]int{
 "terraform": 0,  // Terraform configuration block
 "provider":  1,  // Provider configurations
 "variable":  2,  // Input variables
 "locals":    3,  // Local values
 "data":      4,  // Data sources
 "resource":  5,  // Resources
 "module":    6,  // Module calls
 "output":    7,  // Output values
}
```

### Sorting Rules

**Block Sorting**:

1. Sort by block type (terraform, provider, variable, etc.)
2. Within same type, sort alphabetically by labels
3. Preserve block content and structure

**Attribute Sorting**:

1. Special attributes first: `for_each`, `count`
2. Remaining attributes sorted alphabetically
3. Nested blocks stay after attributes

### HCL Library Limitations

**Important**: Comments are NOT preserved during sorting. This is a limitation of the HCL write library.

**Why**: The `hashicorp/hcl/v2/hclwrite` library reconstructs the HCL file from its AST, which does not preserve comments.

**Documented**: This behavior is clearly documented in README.md and user-facing documentation.

---

## Security Requirements

- Validate file paths (prevent path traversal)
- Handle user input safely
- No arbitrary code execution
- File permissions: respect umask, don't change existing permissions
- No secrets or credentials in code or tests
- Sanitize error messages (no sensitive path information in public errors)

---

## Documentation

### Generated Documentation

None (Godoc is generated from code comments)

### Manual Documentation

- Root `README.md`: Project overview and quick start
- `docs/`:
  - `INSTALLATION.md`: Installation instructions
  - `USAGE.md`: CLI usage and examples
  - `API.md`: Library API reference
  - `ARCHITECTURE.md`: Technical design and implementation
  - `CONTRIBUTING.md`: Contribution guidelines
  - `DEVELOPMENT.md`: Development setup and workflow
  - `CI.md`: CI/CD pipeline documentation
  - `RELEASING.md`: Release process

### Documentation Style

- Use clear, concise language
- Provide code examples in Go
- Include expected output for CLI examples
- Explain "why" not just "how"
- Keep examples up to date with code changes
- Use shell code blocks for commands
- Use go code blocks for Go code

---

## PR & Change Hygiene

- Small, focused diffs (one feature or fix per PR)
- Tests included for new functionality or bug fixes
- Update documentation when behavior changes
- Ensure tests pass locally before pushing
- Run `go fmt ./...` before committing
- Run `go mod tidy` if dependencies changed
- Commit message follows conventional commits format

**Good changes**:

- Add new block type support with tests
- Fix parsing bug with regression test
- Improve error messages with updated tests
- Optimize sorting algorithm with benchmarks

**Avoid**:

- Mixing refactors with feature additions
- Large multi-file changes without a plan
- Breaking test coverage thresholds
- Changing public API without discussion
- Removing tests to make coverage pass

---

## When to Refactor vs Patch

**Patch** (quick fix):

- Single-file bug fix
- Typo correction in docs or comments
- Adding a missing test case
- Updating a dependency version
- Small performance improvement

**Refactor** (plan first):

- Extracting shared logic across multiple packages
- Changing public API in `api` package
- Modifying core sorting algorithm
- Restructuring package layout
- Adding new major features

---

## Performance Considerations

### Benchmarking

- Use Go's built-in benchmark framework
- Run benchmarks before and after optimization
- Focus on hot paths (parsing, sorting)
- Consider both CPU and memory allocations

**Run benchmarks**:

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkSortFile ./api

# With memory allocations
go test -bench=. -benchmem ./...

# Compare before/after (use benchstat)
go test -bench=. ./... > old.txt
# Make changes
go test -bench=. ./... > new.txt
benchstat old.txt new.txt
```

### Optimization Guidelines

- **Measure before optimizing** (profile first)
- Optimize hot paths only
- Prefer simplicity over premature optimization
- Consider memory allocations (reuse buffers)
- Use concurrency for I/O-bound operations

---

## Acceptance Checklist

Before considering work complete:

- [ ] All edits preserve existing conventions
- [ ] Planning phase completed for complex changes (if applicable)
- [ ] Code follows Go idioms and standard library patterns
- [ ] All exported functions have Godoc comments
- [ ] Tests written (table-driven when appropriate)
- [ ] Test coverage maintained at 90%+ (`go test -cover ./...`)
- [ ] Error handling follows repository standards (sentinel errors, wrapping)
- [ ] No direct output in library code (api, hcl, config packages)
- [ ] Security requirements met (validated paths, safe file operations)
- [ ] Code formatted (`go fmt ./...`)
- [ ] Dependencies tidied (`go mod tidy`)
- [ ] All tests pass (`go test ./...`)
- [ ] Documentation updated (if public behavior changed)
- [ ] Commit message follows conventional commits format

---

## Final Rule

**When in doubt**:

1. Stop
2. Ask
3. Plan
4. Proceed deliberately

**Consult documentation**:

- `README.md` for overview
- `docs/ARCHITECTURE.md` for design
- `docs/USAGE.md` for usage patterns
- `docs/API.md` for library API
- `docs/CONTRIBUTING.md` for guidelines
