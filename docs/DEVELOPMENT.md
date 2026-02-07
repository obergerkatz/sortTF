# Development Guide

Comprehensive guide for developing sortTF, including project structure, build instructions, testing strategies, and development workflows.

## Table of Contents

- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Building](#building)
- [Testing](#testing)
- [Debugging](#debugging)
- [Performance](#performance)
- [Code Quality](#code-quality)
- [CI/CD](#cicd)
- [Troubleshooting](#troubleshooting)

## Project Structure

sortTF follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout):

```text
sortTF/
├── cmd/
│   └── sorttf/              # CLI application entry point
│       └── main.go          # Main executable
│
├── api/                     # Public library API
│   ├── sorttf.go           # Core sorting functions
│   └── sorttf_test.go      # API tests
│
├── cli/                     # Command-line interface logic
│   ├── cli.go              # CLI execution and flag handling
│   └── cli_test.go         # CLI tests
│
├── config/                  # Configuration and options
│   ├── config.go           # Config structures
│   └── config_test.go      # Config tests
│
├── hcl/                     # HCL parsing and sorting
│   ├── parser.go           # HCL file parsing
│   ├── sorter.go           # Block and attribute sorting
│   ├── formatter.go        # HCL formatting
│   └── *_test.go           # HCL tests
│
├── internal/                # Private packages
│   ├── errors/             # Error handling and types
│   │   ├── errors.go       # Custom error types
│   │   └── errors_test.go
│   └── files/              # File system operations
│       ├── walker.go       # Directory traversal
│       ├── filter.go       # File filtering
│       └── *_test.go
│
├── integration/             # Integration and system tests
│   ├── integration_test.go # End-to-end CLI tests
│   ├── fixtures/           # Test fixture files
│   └── README.md           # Integration test documentation
│
├── testdata/                # Test data for unit tests
│   ├── valid/              # Valid Terraform files
│   ├── invalid/            # Invalid/edge case files
│   └── expected/           # Expected sorted outputs
│
├── docs/                    # Documentation
│   ├── INSTALLATION.md
│   ├── USAGE.md
│   ├── API.md
│   ├── CONTRIBUTING.md
│   ├── DEVELOPMENT.md      # This file
│   ├── ARCHITECTURE.md
│   └── RELEASING.md
│
├── .github/                 # GitHub configuration
│   ├── workflows/          # GitHub Actions workflows
│   │   ├── ci.yml         # Continuous integration
│   │   ├── release.yml    # Release automation
│   │   └── dependencies.yml
│   └── dependabot.yml      # Dependency updates
│
├── .golangci.yml           # Linter configuration
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── README.md               # Project overview
├── LICENSE                 # MIT License
└── .gitignore              # Git ignore rules
```

### Package Responsibilities

| Package | Purpose | Key Files |
|---------|---------|-----------|
| `cmd/sorttf` | CLI entry point | `main.go` |
| `api` | Public library API | `sorttf.go` |
| `cli` | CLI logic and execution | `cli.go` |
| `config` | Configuration management | `config.go` |
| `hcl` | HCL parsing and sorting | `parser.go`, `sorter.go` |
| `internal/errors` | Error handling | `errors.go` |
| `internal/files` | File operations | `walker.go`, `filter.go` |
| `integration` | End-to-end tests | `integration_test.go` |

## Prerequisites

### Required

- **Go 1.22+** (1.23+ recommended)

  ```bash
  go version
  ```

- **Git**

  ```bash
  git --version
  ```

### Recommended Development Tools

- **golangci-lint** - Comprehensive linter

  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```

- **gopls** - Go language server (for IDE integration)

  ```bash
  go install golang.org/x/tools/gopls@latest
  ```

- **dlv** - Delve debugger

  ```bash
  go install github.com/go-delve/delve/cmd/dlv@latest
  ```

- **Make** (optional) - Build automation

  ```bash
  # macOS
  xcode-select --install
  # Linux
  sudo apt-get install build-essential
  ```

## Building

### Development Build

```bash
# Build for current platform
go build -o sorttf ./cmd/sorttf

# Build with race detector (slower, catches race conditions)
go build -race -o sorttf ./cmd/sorttf

# Build with debugging symbols
go build -gcflags="all=-N -l" -o sorttf ./cmd/sorttf
```

### Production Build

```bash
# Optimized build with version information
VERSION="v1.0.0"
COMMIT=$(git rev-parse HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

go build \
  -trimpath \
  -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
  -o sorttf \
  ./cmd/sorttf
```

**Build flags explained:**

- `-trimpath`: Remove absolute paths for reproducible builds
- `-ldflags="-s -w"`: Strip debug info and symbol table (smaller binary)
- `-X`: Inject version information at build time

### Cross-Compilation

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o sorttf-linux-amd64 ./cmd/sorttf

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o sorttf-linux-arm64 ./cmd/sorttf

# macOS AMD64 (Intel)
GOOS=darwin GOARCH=amd64 go build -o sorttf-darwin-amd64 ./cmd/sorttf

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o sorttf-darwin-arm64 ./cmd/sorttf

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o sorttf-windows-amd64.exe ./cmd/sorttf
```

### Build All Platforms

```bash
# Script to build for all platforms
#!/bin/bash
VERSION=${1:-dev}
PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
)

for platform in "${PLATFORMS[@]}"; do
  GOOS=${platform%/*}
  GOARCH=${platform#*/}
  output="sorttf-${GOOS}-${GOARCH}"
  [ "$GOOS" = "windows" ] && output="${output}.exe"

  echo "Building for $GOOS/$GOARCH..."
  GOOS=$GOOS GOARCH=$GOARCH go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o "dist/$output" \
    ./cmd/sorttf
done
```

## Testing

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./api
go test ./hcl

# Verbose output
go test -v ./...

# With coverage
go test -cover ./...

# Detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# With race detector (important for concurrent code)
go test -race ./...

# Short mode (skip slow tests)
go test -short ./...

# Run specific test
go test ./api -run TestSortFile

# Run tests matching pattern
go test ./... -run TestSort

# Show test output even on success
go test -v ./api

# Run tests 10 times (catch flaky tests)
go test -count=10 ./...
```

### Integration Tests

```bash
# Run integration tests
go test ./integration/...

# Run integration tests with verbose output
go test -v ./integration/...

# Run specific integration test
go test ./integration -run TestCLI_SortsSingleFile

# Skip integration tests
go test -short ./...
```

### Coverage Requirements

- **Target**: 90%+ overall coverage
- **Current**: 95%
- **Per-package minimum**: 80%

```bash
# Check coverage threshold
go test -cover ./... | grep -E 'coverage: [0-9]+%' | \
  awk '{if ($2 < 90.0) print "Low coverage:", $0}'
```

### Benchmarks

```bash
# Run benchmarks
go test -bench=. ./...

# Benchmark specific function
go test -bench=BenchmarkSortFile ./api

# Benchmark with memory profiling
go test -bench=. -benchmem ./...

# Run benchmarks multiple times for accuracy
go test -bench=. -benchtime=10s ./...

# Compare benchmarks
go test -bench=. ./... > old.txt
# Make changes
go test -bench=. ./... > new.txt
benchcmp old.txt new.txt
```

### Test Organization

```go
// Unit test example
func TestSortBlocks(t *testing.T) {
    tests := []struct {
        name     string
        input    []string
        expected []string
    }{
        {"provider before resource", []string{"resource", "provider"}, []string{"provider", "resource"}},
        {"maintains terraform first", []string{"variable", "terraform"}, []string{"terraform", "variable"}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := sortBlocks(tt.input)
            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}

// Integration test example
func TestCLI_SortFile(t *testing.T) {
    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "test.tf")

    // Write test content
    os.WriteFile(testFile, []byte(`resource "x" {}`), 0644)

    // Execute binary
    cmd := exec.Command("./sorttf-test", testFile)
    output, err := cmd.CombinedOutput()

    // Assert results
    if err != nil {
        t.Fatalf("command failed: %v\nOutput: %s", err, output)
    }
}
```

## Debugging

### Using Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug tests
dlv test ./api

# Debug specific test
dlv test ./api -- -test.run TestSortFile

# Debug binary with arguments
dlv exec ./sorttf -- main.tf --dry-run

# Debug and set breakpoint
dlv debug ./cmd/sorttf
(dlv) break main.main
(dlv) continue
```

### VS Code Debugging

`.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug sortTF",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/sorttf",
      "args": ["--dry-run", "testdata/valid/simple.tf"]
    },
    {
      "name": "Debug Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/api"
    }
  ]
}
```

### Logging and Tracing

```go
// Add debug logging
import "log"

func sortFile(path string) error {
    log.Printf("Processing file: %s", path)
    // ...
}

// Use trace for performance debugging
import "runtime/trace"

func main() {
    f, _ := os.Create("trace.out")
    defer f.Close()
    trace.Start(f)
    defer trace.Stop()

    // Your code here
}

// Analyze trace
// go tool trace trace.out
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./...
go tool pprof mem.prof

# Block profiling (goroutine blocking)
go test -blockprofile=block.prof ./...
go tool pprof block.prof

# Generate call graph
go test -cpuprofile=cpu.prof -bench=. ./...
go tool pprof -web cpu.prof
```

## Performance

### Optimization Guidelines

1. **Profile first, optimize second** - Don't guess where the bottlenecks are
2. **Benchmark changes** - Always compare before/after
3. **Avoid premature optimization** - Focus on correctness first
4. **Use concurrent processing** - sortTF processes multiple files in parallel

### Common Performance Patterns

```go
// Good: Preallocate slices when size is known
blocks := make([]Block, 0, len(input))

// Good: Use strings.Builder for string concatenation
var builder strings.Builder
for _, s := range parts {
    builder.WriteString(s)
}
result := builder.String()

// Good: Reuse buffers
var buf bytes.Buffer
for _, file := range files {
    buf.Reset()
    // Process file using buf
}

// Good: Minimize allocations in hot paths
// Cache commonly used values
// Use sync.Pool for temporary objects
```

## Code Quality

### Linting

```bash
# Run golangci-lint
golangci-lint run

# Run specific linters
golangci-lint run --enable-only=gosec
golangci-lint run --enable-only=staticcheck

# Auto-fix issues
golangci-lint run --fix

# Show all linters
golangci-lint linters
```

### Static Analysis

```bash
# Go vet (included in golangci-lint)
go vet ./...

# Staticcheck
staticcheck ./...

# Check for security issues
gosec ./...

# Check for vulnerabilities
govulncheck ./...
```

### Code Formatting

```bash
# Format code
go fmt ./...

# More aggressive formatting
gofmt -s -w .

# Check if code is formatted
test -z $(gofmt -l .)
```

### Dependency Management

```bash
# Add dependency
go get github.com/some/package

# Update dependency
go get -u github.com/some/package

# Update all dependencies
go get -u ./...

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify

# Show dependency graph
go mod graph

# Why is this dependency included?
go mod why github.com/some/package
```

## CI/CD

### GitHub Actions Workflows

sortTF uses GitHub Actions for CI/CD:

- **ci.yml**: Runs on every push and PR
  - Linting
  - Testing (matrix: 3 OS × 2 Go versions)
  - Security scanning
  - Coverage reporting
  - Multi-platform builds

- **release.yml**: Runs on version tags
  - Full test suite
  - Multi-platform builds
  - GitHub Release creation

- **dependencies.yml**: Runs weekly
  - Dependency updates
  - Vulnerability scanning

### Running CI Locally

```bash
# Using act (GitHub Actions locally)
brew install act  # macOS
act -j test       # Run test job
act -j lint       # Run lint job

# Or use Docker
docker run -v $(pwd):/workspace golang:1.23 bash -c "cd /workspace && go test ./..."
```

## Troubleshooting

### Common Issues

#### "command not found: sorttf"

**Solution:**

```bash
# Ensure $GOPATH/bin is in PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

#### Tests fail with "permission denied"

**Solution:**

```bash
chmod +x ./sorttf-test
```

#### "cannot find package"

**Solution:**

```bash
go mod download
go mod tidy
```

#### Race detector reports issues

**Solution:**

```bash
# Fix race conditions, then verify
go test -race ./...
```

### Debug Environment

```bash
# Print Go environment
go env

# Check module status
go mod verify
go mod graph

# Clean cache if issues persist
go clean -cache
go clean -modcache
```

## Next Steps

- Read [Architecture Documentation](ARCHITECTURE.md) for design details
- See [Contributing Guide](CONTRIBUTING.md) for contribution workflow
- Check [Release Process](RELEASING.md) for release procedures
