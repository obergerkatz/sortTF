# Using sortTF as a Library

sortTF can be imported and used as a library in your own Go programs. This allows you to integrate Terraform file sorting into tools like pre-commit hooks, CI/CD pipelines, or custom build systems.

## Installation

```bash
go get github.com/obergerkatz/sorttf/api
```

## Quick Start

```go
import "sorttf/api"

// Sort a single file
err := lib.SortFile("main.tf", lib.Options{})
if err != nil && !errors.Is(err, lib.ErrNoChanges) {
    log.Fatal(err)
}
```

## API Reference

### Functions

#### `SortFile(path string, opts Options) error`

Sorts and formats a single Terraform or Terragrunt file.

**Returns:**
- `nil`: File was successfully sorted and written
- `ErrNoChanges`: File is already sorted (not an error)
- `ErrNeedsSorting`: File needs sorting (only in Validate mode)
- `error`: Parsing, validation, or I/O error

#### `GetSortedContent(path string) (content string, changed bool, err error)`

Reads a file and returns its sorted content without modifying the file. Useful for previewing changes or computing diffs.

#### `SortFiles(paths []string, opts Options) map[string]error`

Sorts multiple files and returns results for each. Continues processing on error.

#### `SortDirectory(dir string, recursive bool, opts Options) (map[string]error, error)`

Sorts all Terraform/Terragrunt files in a directory.

### Types

#### `Options`

```go
type Options struct {
    DryRun   bool  // Don't modify files, just check
    Validate bool  // Return ErrNeedsSorting if changes needed
}
```

#### Sentinel Errors

- `lib.ErrNoChanges`: File is already sorted
- `lib.ErrNeedsSorting`: File needs sorting (Validate mode)

## Examples

See `library_usage.go` for comprehensive examples including:
- Sorting single files
- Validate mode for CI/CD
- Getting sorted content for diffs
- Batch processing multiple files
- Directory sorting
- Pre-commit hook integration

## Use Cases

### Pre-commit Hook

```go
// Sort all staged Terraform files
func preCommitHook() error {
    cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACM")
    output, _ := cmd.Output()
    files := strings.Split(string(output), "\n")

    for _, file := range files {
        if strings.HasSuffix(file, ".tf") {
            err := lib.SortFile(file, lib.Options{})
            if err != nil && !errors.Is(err, lib.ErrNoChanges) {
                return fmt.Errorf("failed to sort %s: %w", file, err)
            }
        }
    }
    return nil
}
```

### CI/CD Validation

```go
// Check if all Terraform files are sorted
func validateInCI() error {
    results, err := lib.SortDirectory("./terraform", true, lib.Options{Validate: true})
    if err != nil {
        return err
    }

    needsSorting := false
    for path, err := range results {
        if errors.Is(err, lib.ErrNeedsSorting) {
            fmt.Printf("❌ %s needs sorting\n", path)
            needsSorting = true
        }
    }

    if needsSorting {
        return fmt.Errorf("some files are not sorted - run 'sorttf .'")
    }
    return nil
}
```

### LSP/Editor Integration

```go
// Get formatted content for display in editor
func formatDocument(path string) (string, error) {
    content, _, err := lib.GetSortedContent(path)
    return content, err
}
```

## Testing

The library API is designed to be testable:

```go
func TestMyTool(t *testing.T) {
    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "test.tf")

    // Write test content
    os.WriteFile(testFile, []byte(`resource "x" {}`), 0644)

    // Sort it
    err := lib.SortFile(testFile, lib.Options{})
    if err != nil {
        t.Fatal(err)
    }

    // Verify result
    result, _ := os.ReadFile(testFile)
    // ... assertions ...
}
```

## Benefits over CLI

✅ **No subprocess overhead** - Direct function calls
✅ **Type-safe** - Compile-time errors instead of parsing CLI output
✅ **Testable** - Easy to unit test your integrations
✅ **Structured errors** - Use `errors.Is()` for error handling
✅ **Programmatic control** - Full control over behavior
✅ **Better performance** - No process spawning or argument parsing
