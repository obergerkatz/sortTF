# Architecture Documentation

Technical overview of sortTF's architecture, design decisions, and implementation details.

## Table of Contents

- [Overview](#overview)
- [Design Principles](#design-principles)
- [Architecture Diagram](#architecture-diagram)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [Sorting Algorithm](#sorting-algorithm)
- [Concurrency Model](#concurrency-model)
- [Error Handling](#error-handling)
- [Testing Strategy](#testing-strategy)
- [Performance Considerations](#performance-considerations)
- [Design Decisions](#design-decisions)
- [Future Enhancements](#future-enhancements)

## Overview

sortTF is a command-line tool and Go library designed to sort and format Terraform and Terragrunt files. The architecture emphasizes:

- **Modularity**: Clear separation of concerns
- **Testability**: High test coverage (95%)
- **Performance**: Concurrent file processing
- **Reliability**: Comprehensive error handling
- **Extensibility**: Easy to add new block types or sorting rules

## Design Principles

### 1. Separation of Concerns

Each package has a single, well-defined responsibility:

- `cmd/sorttf`: CLI entry point only
- `cli`: CLI logic and execution
- `api`: Public library interface
- `hcl`: HCL parsing and sorting
- `internal/*`: Private utilities

### 2. API-First Design

The library API (`api` package) is the primary interface. The CLI is built on top of it, ensuring:

- Users can integrate sortTF into their own tools
- CLI and library share the same code path
- Testing focuses on the API, which tests both CLI and library

### 3. Immutability Where Possible

Operations return new data structures rather than modifying in place:

```go
// Blocks are sorted into a new slice
func sortBlocks(blocks []Block) []Block {
    sorted := make([]Block, len(blocks))
    copy(sorted, blocks)
    // Sort sorted slice
    return sorted
}
```

### 4. Fail Fast

Invalid inputs are caught early:

```go
func SortFile(path string, opts Options) error {
    if path == "" {
        return fmt.Errorf("path cannot be empty")
    }
    // Continue processing
}
```

### 5. Explicit Over Implicit

Behavior is explicit, not magical:

- Options struct for configuration
- Named return values for clarity
- Sentinel errors for expected conditions

## Architecture Diagram

```text
┌─────────────────────────────────────────────────────────────┐
│                         User Input                          │
│                    (CLI Args or API Call)                   │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│                      Entry Points                           │
│  ┌──────────────┐              ┌──────────────────────┐    │
│  │ cmd/sorttf   │              │   api.SortFile()     │    │
│  │   main.go    │──────────────▶   api.SortFiles()    │    │
│  │              │              │   api.SortDirectory()│    │
│  └──────┬───────┘              └──────────┬───────────┘    │
│         │                                  │                 │
│         └───────────┬──────────────────────┘                │
│                     ▼                                        │
│         ┌───────────────────────┐                          │
│         │     cli.RunCLI()      │                          │
│         │  Flag parsing & exec  │                          │
│         └───────────┬───────────┘                          │
└─────────────────────┼───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   File Discovery                            │
│                 (internal/files)                            │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Walker: Find all .tf and .hcl files               │    │
│  │  Filter: Skip .terraform/, .terragrunt-cache/, etc │    │
│  └────────────────────┬───────────────────────────────┘    │
└─────────────────────────┼───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    For Each File                            │
└─────────────────────────────────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
┌─────────────────┐ ┌─────────────┐ ┌─────────────┐
│   Read File     │ │ Parse HCL   │ │Sort Blocks  │
│   (os.ReadFile) │→│(hcl.Parser) │→│(hcl.Sorter) │
└─────────────────┘ └─────────────┘ └──────┬──────┘
                                            │
                          ┌─────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                  Format & Write                             │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Formatter: Apply terraform fmt standards          │    │
│  │  Writer: Write back to file (or return content)    │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
                    ┌──────────┐
                    │  Result  │
                    └──────────┘
```

## Core Components

### 1. CLI Layer (`cmd/sorttf`, `cli`)

**Responsibilities:**

- Parse command-line arguments
- Validate flags
- Call appropriate API functions
- Format and display results

**Key Files:**

- `cmd/sorttf/main.go`: Entry point
- `cli/cli.go`: CLI logic

**Design:**

```go
// main.go - Minimal entry point
func main() {
    os.Exit(cli.RunCLI(os.Args[1:]))
}

// cli.go - All CLI logic
func RunCLI(args []string) int {
    // Parse flags
    // Validate inputs
    // Call API
    // Handle results
    return exitCode
}
```

### 2. Public API Layer (`api`)

**Responsibilities:**

- Public interface for sorting operations
- High-level functions for common use cases
- Error handling and sentinel errors

**Key Functions:**

```go
func SortFile(path string, opts Options) error
func GetSortedContent(path string) (string, bool, error)
func SortFiles(paths []string, opts Options) map[string]error
func SortDirectory(dir string, recursive bool, opts Options) (map[string]error, error)
```

**Design Patterns:**

- Options struct for configuration
- Sentinel errors (`ErrNoChanges`, `ErrNeedsSorting`)
- Map return for batch operations
- Concurrent processing in `SortFiles`

### 3. Configuration Layer (`config`)

**Responsibilities:**

- Define configuration structures
- Validate options
- Provide defaults

**Key Types:**

```go
type Options struct {
    DryRun   bool  // Preview mode
    Validate bool  // CI/CD validation mode
}
```

### 4. HCL Processing Layer (`hcl`)

**Responsibilities:**

- Parse HCL files using hashicorp/hcl
- Sort blocks according to Terraform best practices
- Sort attributes within blocks
- Format output

**Key Components:**

#### Parser (`hcl/parser.go`)

```go
func Parse(content []byte) (*File, error) {
    // Parse using hclwrite
    // Extract blocks
    // Build internal representation
}
```

#### Sorter (`hcl/sorter.go`)

```go
func SortBlocks(file *File) {
    // Sort top-level blocks
    // Sort nested blocks
    // Sort attributes
}
```

#### Formatter (`hcl/formatter.go`)

```go
func Format(file *File) ([]byte, error) {
    // Apply terraform fmt standards
    // Note: Comments are not preserved (HCL library limitation)
    // Generate output
}
```

**Block Order:**

```go
var blockOrder = map[string]int{
    "terraform": 0,
    "provider":  1,
    "variable":  2,
    "locals":    3,
    "data":      4,
    "resource":  5,
    "module":    6,
    "output":    7,
}
```

### 5. Internal Utilities (`internal`)

#### File Operations (`internal/files`)

**Walker:**

```go
func Walk(root string, recursive bool) ([]string, error) {
    // Traverse directory tree
    // Filter .tf and .hcl files
    // Skip excluded directories
}
```

**Filter:**

```go
func ShouldProcess(path string) bool {
    // Skip .terraform/
    // Skip .terragrunt-cache/
    // Skip hidden directories
}
```

#### Error Handling (`internal/errors`)

**Custom Error Types:**

```go
type ParseError struct {
    Path string
    Line int
    Err  error
}

type ValidationError struct {
    Path   string
    Issues []string
}
```

## Data Flow

### Single File Processing

```text
User Input (path)
    │
    ▼
Validate path exists
    │
    ▼
Read file content
    │
    ▼
Parse HCL → *hclwrite.File
    │
    ▼
Extract blocks → []Block
    │
    ▼
Sort blocks by type and name
    │
    ▼
Sort attributes within each block
    │
    ▼
Format using hclwrite
    │
    ▼
Compare with original
    │
    ├─ No changes → Return ErrNoChanges
    │
    └─ Changes exist
        │
        ├─ DryRun → Return content
        ├─ Validate → Return ErrNeedsSorting
        └─ Normal → Write file
```

### Directory Processing

```text
User Input (directory path)
    │
    ▼
Walk directory tree
    │
    ▼
Filter files (.tf, .hcl)
    │
    ▼
Skip excluded dirs (.terraform/, etc)
    │
    ▼
For each file (concurrently):
    │
    ├─ Process file
    └─ Collect results
        │
        ▼
Aggregate results → map[path]error
```

## Sorting Algorithm

### Block Sorting

#### 1: Categorize by Type**

```go
func getBlockOrder(blockType string) int {
    if order, ok := blockOrder[blockType]; ok {
        return order
    }
    return 999 // Unknown types go last
}
```

#### 2: Sort by Type**

```go
sort.SliceStable(blocks, func(i, j int) bool {
    return getBlockOrder(blocks[i].Type) < getBlockOrder(blocks[j].Type)
})
```

#### 3: Sort Within Type by Labels**

```go
// Within same type, sort alphabetically by labels
if blocks[i].Type == blocks[j].Type {
    return blocks[i].Labels < blocks[j].Labels
}
```

### Attribute Sorting

#### 1: Separate Special Attributes**

```go
var special []string  // for_each, count
var regular []string  // everything else
```

#### 2: Sort Regular Attributes**

```go
sort.Strings(regular)
```

#### 3: Combine**

```go
result := append(special, regular...)
```

**Example:**

```hcl
resource "aws_instance" "web" {
  for_each = var.instances  # Always first

  ami           = "ami-123"  # Alphabetical
  instance_type = "t3.micro"
  tags          = { Name = "web" }

  lifecycle {  # Nested blocks last
    create_before_destroy = true
  }
}
```

## Concurrency Model

### Concurrent File Processing

```go
func SortFiles(paths []string, opts Options) map[string]error {
    results := make(map[string]error)
    var mu sync.Mutex
    var wg sync.WaitGroup

    for _, path := range paths {
        wg.Add(1)
        go func(p string) {
            defer wg.Done()
            err := SortFile(p, opts)

            mu.Lock()
            results[p] = err
            mu.Unlock()
        }(path)
    }

    wg.Wait()
    return results
}
```

**Benefits:**

- Faster processing of multiple files
- Scales with CPU cores
- Safe with mutex-protected results map

**Considerations:**

- File I/O is the bottleneck, not CPU
- Limit goroutines for very large directories
- Each file processed independently

## Error Handling

### Error Hierarchy

```text
error (interface)
    │
    ├── Sentinel Errors
    │   ├── ErrNoChanges      (not really an error)
    │   └── ErrNeedsSorting   (validation mode)
    │
    ├── Wrapped Errors (fmt.Errorf with %w)
    │   ├── "failed to read file: %w"
    │   ├── "failed to parse HCL: %w"
    │   └── "failed to write file: %w"
    │
    └── Custom Error Types
        ├── ParseError        (HCL syntax errors)
        └── ValidationError   (validation failures)
```

### Error Checking Patterns

```go
// Check sentinel errors
if errors.Is(err, api.ErrNoChanges) {
    // Handle "already sorted" case
}

// Wrap errors for context
if err != nil {
    return fmt.Errorf("failed to sort %s: %w", path, err)
}

// Type assertion for custom errors
var parseErr *ParseError
if errors.As(err, &parseErr) {
    fmt.Printf("Parse error at line %d\n", parseErr.Line)
}
```

## Testing Strategy

### Test Pyramid

```text
                   ╱╲
                  ╱  ╲
                 ╱ E2E ╲         29 Integration Tests
                ╱──────╲       (integration_test.go)
               ╱        ╲
              ╱  Integration╲
             ╱──────────────╲
            ╱                ╲
           ╱   Unit Tests     ╲    155+ Unit Tests
          ╱────────────────────╲  (package_test.go files)
         ╱________________________╲
```

### Test Coverage by Package

| Package | Coverage | Test Count |
|---------|----------|------------|
| `config` | 100% | 15 |
| `internal/errors` | 100% | 10 |
| `internal/files` | 100% | 20 |
| `hcl` | 92% | 50+ |
| `api` | 90% | 30 |
| `cli` | 88% | 30 |
| **Overall** | **95%** | **155+** |

### Test Types

**Unit Tests:**

- Test individual functions in isolation
- Fast, deterministic
- Use table-driven tests

**Integration Tests:**

- Test full CLI binary end-to-end
- Real file I/O
- Multiple scenarios (flags, errors, etc.)

**System Tests:**

- Multi-file workflows
- Error handling
- Performance characteristics

## Performance Considerations

### Benchmarks

```bash
BenchmarkSortFile/small-8         10000   115000 ns/op   45000 B/op   500 allocs/op
BenchmarkSortFile/medium-8         2000   850000 ns/op  320000 B/op  3500 allocs/op
BenchmarkSortFile/large-8           500  3200000 ns/op 1280000 B/op 14000 allocs/op
```

### Optimization Strategies

1. **Concurrent Processing**: Multiple files processed in parallel
2. **Efficient Parsing**: Use hclwrite's native parser
3. **Minimal Allocations**: Preallocate slices when size known
4. **Buffered I/O**: Read/write files efficiently

### Bottlenecks

- **File I/O**: Reading and writing files (unavoidable)
- **HCL Parsing**: hashicorp/hcl parsing (library dependency)
- **Sorting**: Negligible compared to I/O

## Design Decisions

### Why hashicorp/hcl?

**Decision**: Use official HashiCorp HCL library

**Rationale:**

- Battle-tested by Terraform itself
- Correct handling of all HCL edge cases
- Standard formatting aligned with terraform fmt
- Active maintenance

**Trade-offs:**

- Adds dependency
- Slightly slower than custom parser
- Must follow library's API changes

### Why Separate CLI and API?

**Decision**: Split cmd/sorttf and api packages

**Rationale:**

- Library users don't need CLI code
- Clear API surface
- Easier testing
- Follows Go standards

### Why Concurrent File Processing?

**Decision**: Process files concurrently in `SortFiles`

**Rationale:**

- Significant speedup on multi-core systems
- Simple implementation with goroutines
- Safe with proper synchronization

**Trade-offs:**

- Added complexity
- Potential for too many goroutines (mitigated by processing in batches)

### Why Options Struct?

**Decision**: Use `Options` struct instead of function parameters

**Rationale:**

- Extensible (add options without breaking API)
- Self-documenting
- Easy to provide defaults

**Example:**

```go
// Bad: Many parameters
func SortFile(path string, dryRun bool, validate bool, verbose bool) error

// Good: Options struct
func SortFile(path string, opts Options) error
```

## Future Enhancements

### Potential Improvements

1. **Custom Sorting Rules**
   - Allow users to define custom block orders
   - Configurable attribute sorting preferences

2. **Plugin System**
   - Support custom block types
   - Extensible formatting rules

3. **LSP Integration**
   - Real-time sorting in editors
   - Format-on-save support

4. **Incremental Sorting**
   - Only sort changed blocks
   - Cache parsed results

5. **Parallel Directory Walking**
   - Concurrent directory traversal
   - Better performance on large repos

6. **Configuration Files**
   - `.sorttf.yml` for project-specific settings
   - Ignore patterns
   - Custom sorting rules

### Non-Goals

What sortTF intentionally does NOT do:

- **Semantic validation**: Use `terraform validate`
- **Dependency resolution**: Not a Terraform replacement
- **Code generation**: Only sorting/formatting
- **Auto-fixing errors**: Only sorts valid HCL

## Contributing

See [Contributing Guide](CONTRIBUTING.md) for how to contribute to sortTF's architecture and implementation.

## References

- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [HashiCorp HCL](https://github.com/hashicorp/hcl)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
