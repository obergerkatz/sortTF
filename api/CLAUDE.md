# API Package Instructions

**Scope**: Applies only to `api/` directory.

This file **extends** the repository-level `.claude/CLAUDE.md`. Read that first for global standards. This file contains only what's **specific** to the api package.

---

## Purpose

The `api` package is the **public Go library interface** for sortTF, providing:

- High-level functions for sorting Terraform and Terragrunt files
- Single file, multiple file, and directory operations
- Dry-run and validation modes for CI/CD integration
- Comprehensive error handling with sentinel errors
- Concurrent file processing for performance

**Philosophy**:

- Simple, intuitive API for library consumers
- Clear error semantics with sentinel errors
- No direct output (return errors, let caller decide how to handle)
- Safe concurrent operations

---

## Package Structure

```text
api/
├── sorttf.go           # Main API functions
├── sorttf_test.go      # Comprehensive tests
└── CLAUDE.md           # This file
```

---

## Public API

### Core Functions

```go
// Single file operations
func SortFile(path string, opts Options) error
func GetSortedContent(path string) (content string, changed bool, err error)

// Batch operations
func SortFiles(paths []string, opts Options) map[string]error

// Directory operations
func SortDirectory(dir string, recursive bool, opts Options) (map[string]error, error)
```

### Configuration

```go
type Options struct {
 DryRun   bool  // Preview changes without writing
 Validate bool  // Check if files need sorting (CI/CD mode)
}
```

### Sentinel Errors

```go
var (
 // ErrNoChanges indicates the file is already correctly sorted
 ErrNoChanges = errors.New("file is already sorted")

 // ErrNeedsSorting indicates the file needs sorting (Validate mode only)
 ErrNeedsSorting = errors.New("file needs sorting")
)
```

---

## API Design Principles

### 1. Return Errors, Don't Log

Library code must NOT print to stdout/stderr. Return errors for callers to handle:

```go
// GOOD: Return error
if err != nil {
 return fmt.Errorf("failed to parse %s: %w", path, err)
}

// BAD: Print to stdout/stderr
if err != nil {
 fmt.Printf("Error: %v\n", err)  // NEVER do this in api package
 return err
}
```

### 2. Use Sentinel Errors for Expected Conditions

Not all errors are failures. Use sentinel errors for expected states:

```go
// File is already sorted (not an error, but caller needs to know)
if !changed {
 return ErrNoChanges
}

// Validate mode: file needs sorting (expected result, not a failure)
if opts.Validate && changed {
 return ErrNeedsSorting
}
```

**Callers check with `errors.Is()`**:

```go
err := api.SortFile("main.tf", api.Options{})
if errors.Is(err, api.ErrNoChanges) {
 // File already sorted (success case)
} else if err != nil {
 // Actual error
}
```

### 3. Provide Multiple Levels of Abstraction

Offer functions for different use cases:

- `SortFile()`: Simple single-file operation
- `GetSortedContent()`: Get sorted content without writing
- `SortFiles()`: Batch operation with concurrent processing
- `SortDirectory()`: High-level directory operation

### 4. Options Struct for Configuration

Use `Options` struct instead of boolean parameters:

```go
// GOOD: Extensible options
func SortFile(path string, opts Options) error

// BAD: Multiple boolean parameters
func SortFile(path string, dryRun bool, validate bool) error
```

---

## Function Specifications

### SortFile

```go
func SortFile(path string, opts Options) error
```

**Purpose**: Sort a single Terraform or Terragrunt file in place.

**Behavior**:

- Reads file from `path`
- Parses HCL content
- Sorts blocks and attributes
- Writes sorted content back (unless DryRun)
- Returns `ErrNoChanges` if already sorted
- Returns `ErrNeedsSorting` if Validate mode and file needs sorting

**Parameters**:

- `path`: File path (relative or absolute)
- `opts`: Configuration options

**Returns**:

- `nil`: File sorted successfully
- `ErrNoChanges`: File already sorted (not an error)
- `ErrNeedsSorting`: Validate mode, file needs sorting
- `error`: Parse error, I/O error, or other failure

**Error Handling**:

- Wrap errors with file path context
- Use `hcl` package for parsing
- Use `internal/files` for file operations

### GetSortedContent

```go
func GetSortedContent(path string) (content string, changed bool, err error)
```

**Purpose**: Get sorted content without writing to file.

**Behavior**:

- Reads and parses file
- Returns sorted content as string
- Returns boolean indicating if content changed
- Does NOT write to file

**Use Cases**:

- Preview mode
- Diff generation
- Testing

**Returns**:

- `content`: Sorted file content
- `changed`: True if content differs from original
- `err`: Parse or I/O error

### SortFiles

```go
func SortFiles(paths []string, opts Options) map[string]error
```

**Purpose**: Sort multiple files concurrently.

**Behavior**:

- Processes files in parallel using goroutines
- Returns map of results (path -> error)
- Errors for individual files don't stop others
- Thread-safe with mutex-protected results map

**Performance**:

- Concurrent processing for speed
- Suitable for large file sets
- Each file processed independently

**Returns**:

- Map where:
  - Key: File path
  - Value: Error (nil if success, ErrNoChanges if no changes, other error if failed)

**Example**:

```go
results := api.SortFiles(paths, api.Options{})
for path, err := range results {
 if errors.Is(err, api.ErrNoChanges) {
  fmt.Printf("%s: already sorted\n", path)
 } else if err != nil {
  fmt.Printf("%s: error: %v\n", path, err)
 } else {
  fmt.Printf("%s: sorted\n", path)
 }
}
```

### SortDirectory

```go
func SortDirectory(dir string, recursive bool, opts Options) (map[string]error, error)
```

**Purpose**: Sort all Terraform/Terragrunt files in a directory.

**Behavior**:

- Uses `internal/files.Walk()` to find files
- Filters .tf and .hcl files
- Skips excluded directories (.terraform/, .terragrunt-cache/)
- Calls `SortFiles()` for concurrent processing

**Parameters**:

- `dir`: Directory path
- `recursive`: If true, process subdirectories
- `opts`: Configuration options

**Returns**:

- `map[string]error`: Results for each file (path -> error)
- `error`: Directory walk error or validation error

**Error Handling**:

- Returns error if directory doesn't exist
- Returns error if no files found (empty directory)
- Individual file errors in map, directory errors as return value

---

## Testing the API

### Test Organization

```go
func TestSortFile(t *testing.T) {
 t.Run("already sorted", func(t *testing.T) { /* ... */ })
 t.Run("needs sorting", func(t *testing.T) { /* ... */ })
 t.Run("parse error", func(t *testing.T) { /* ... */ })
 t.Run("file not found", func(t *testing.T) { /* ... */ })
}
```

### Test Coverage Requirements

- Test all public functions
- Test all error paths
- Test sentinel errors with `errors.Is()`
- Test Options variations (DryRun, Validate)
- Test edge cases (empty files, large files)

### Test Fixtures

Use `testdata/` for test files:

```text
api/
├── sorttf.go
├── sorttf_test.go
└── testdata/
    ├── valid.tf
    ├── needs_sorting.tf
    ├── invalid.tf
    └── empty.tf
```

### Temporary Files

Use `t.TempDir()` for file I/O tests:

```go
func TestSortFile_WriteToFile(t *testing.T) {
 tmpDir := t.TempDir()
 testFile := filepath.Join(tmpDir, "test.tf")

 // Write test content
 err := os.WriteFile(testFile, []byte("..."), 0644)
 if err != nil {
  t.Fatal(err)
 }

 // Test SortFile
 err = SortFile(testFile, Options{})
 // ... assertions
}
```

### Concurrent Testing

Test `SortFiles()` with race detector:

```bash
go test -race ./api
```

---

## Error Handling Patterns

### Wrapping Errors

Always add context when wrapping:

```go
content, err := os.ReadFile(path)
if err != nil {
 return fmt.Errorf("failed to read file %s: %w", path, err)
}

sorted, err := hcl.Parse(content)
if err != nil {
 return fmt.Errorf("failed to parse HCL in %s: %w", path, err)
}
```

### Sentinel Error Checks

Use `errors.Is()` for sentinel errors:

```go
err := hcl.SortAndFormat(file)
if errors.Is(err, hcl.ErrNoChanges) {
 return ErrNoChanges
}
if err != nil {
 return fmt.Errorf("failed to sort: %w", err)
}
```

### Type Assertions

Use `errors.As()` for custom error types:

```go
var parseErr *hcl.ParseError
if errors.As(err, &parseErr) {
 return fmt.Errorf("syntax error at line %d: %w", parseErr.Line, err)
}
```

---

## Concurrency

### SortFiles Implementation

```go
func SortFiles(paths []string, opts Options) map[string]error {
 results := make(map[string]error, len(paths))
 var mu sync.Mutex
 var wg sync.WaitGroup

 for _, path := range paths {
  wg.Add(1)
  go func(p string) {
   defer wg.Done()

   // Process file
   err := SortFile(p, opts)

   // Thread-safe result storage
   mu.Lock()
   results[p] = err
   mu.Unlock()
  }(path)
 }

 wg.Wait()
 return results
}
```

**Key Points**:

- One goroutine per file
- Mutex protects shared results map
- WaitGroup ensures all complete
- No goroutine limit (assumes reasonable file count)

---

## Dependencies

**Internal**:

- `github.com/obergerkatz/sortTF/hcl` - HCL parsing and sorting
- `github.com/obergerkatz/sortTF/config` - Options type
- `github.com/obergerkatz/sortTF/internal/files` - File operations

**External**:

- None (uses standard library and internal packages)

---

## Common Patterns

### Single File Operation

```go
func processSingleFile(path string) error {
 err := api.SortFile(path, api.Options{})
 if errors.Is(err, api.ErrNoChanges) {
  fmt.Println("File already sorted")
  return nil
 }
 return err
}
```

### Dry Run Mode

```go
content, changed, err := api.GetSortedContent(path)
if err != nil {
 return err
}
if changed {
 fmt.Println("Changes would be made:")
 fmt.Println(content)
}
```

### Validation Mode (CI/CD)

```go
err := api.SortFile(path, api.Options{Validate: true})
if errors.Is(err, api.ErrNeedsSorting) {
 fmt.Fprintf(os.Stderr, "File needs sorting: %s\n", path)
 os.Exit(1)
}
```

### Batch Processing

```go
results := api.SortFiles(paths, api.Options{})

var errCount int
for path, err := range results {
 if err != nil && !errors.Is(err, api.ErrNoChanges) {
  fmt.Fprintf(os.Stderr, "Error in %s: %v\n", path, err)
  errCount++
 }
}

if errCount > 0 {
 return fmt.Errorf("%d files failed", errCount)
}
```

---

## Performance Considerations

### Concurrency

`SortFiles()` uses goroutines for parallelism:

- Faster for multiple files
- Each file processed independently
- I/O-bound workload benefits from concurrency

### Memory

- Files read entirely into memory
- Parsed AST held in memory
- Suitable for typical Terraform file sizes (<1MB)
- Very large files may impact performance

### Optimization

- Avoid reading files multiple times
- Reuse Options struct
- Consider limiting goroutines for very large file sets

---

## Breaking Changes

**The API package is public. Breaking changes affect all users.**

**Before breaking changes**:

1. Discuss in GitHub issue
2. Consider backwards-compatible alternatives
3. Document migration path
4. Bump major version (v2.0.0)

**Backwards-compatible additions**:

- New functions: OK
- New Options fields: OK (with zero-value default)
- New sentinel errors: OK

**Breaking changes**:

- Changing function signatures: AVOID
- Removing functions: AVOID
- Changing error semantics: AVOID
- Renaming types: AVOID

---

## Acceptance Checklist (API Package)

Before considering API changes complete:

- [ ] All public functions have comprehensive Godoc
- [ ] Examples in Godoc demonstrate usage
- [ ] All error returns documented
- [ ] Sentinel errors used appropriately
- [ ] No direct output (no fmt.Print*, no log.*)
- [ ] Thread-safe concurrent operations
- [ ] Tests cover all public functions
- [ ] Tests verify sentinel errors with `errors.Is()`
- [ ] Tests include edge cases (empty files, parse errors)
- [ ] Test coverage >90% (`go test -cover ./api`)
- [ ] No breaking changes (or major version bump planned)
- [ ] Integration tests updated if behavior changed
- [ ] Documentation (docs/API.md) updated if API changed
